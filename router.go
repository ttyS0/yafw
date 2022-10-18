package yafw

import (
	"encoding/binary"
	"errors"
	"log"
	"reflect"
	"syscall"

	"github.com/google/nftables"
	"github.com/google/nftables/expr"
	"github.com/ti-mo/conntrack"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
)

type Router struct {
	// kernel network interfaces
	ns  netns.NsHandle
	nft *nftables.Conn
	nl  *netlink.Handle
	ct  *conntrack.Conn

	// main netfilter table
	table *nftables.Table

	forward     *nftables.Chain
	postrouting *nftables.Chain
	prerouting  *nftables.Chain

	zones  map[string]*Zone
	ipsets map[string]*IPSet
	// serviceGroups map[string]*ServiceGroup

	snatEntries   *EntryTable
	dnatEntries   *EntryTable
	policyEntries *EntryTable
}

var (
	ErrEntryTypeMismatch    = errors.New("entry type mismatch")
	ErrEntryIndexDuplicated = errors.New("entry index duplicated")
	ErrEntryIndexNotFound   = errors.New("entry index not found")
)

// A general entry in nftables chains, which stands for a bunch of rules.
type Entry interface {
	buildArtifact(router *Router) error

	Index() int
	SetIndex(int)
	ToRules() []*nftables.Rule
}

type EntryTable struct {
	r         *Router
	entryType reflect.Type
	list      []Entry
	ruleMap   map[int][]*nftables.Rule
	counter   int

	chain *nftables.Chain
}

func NewEntryTable(router *Router, chain *nftables.Chain, v any) *EntryTable {
	return &EntryTable{
		r:         router,
		chain:     chain,
		entryType: reflect.TypeOf(v),
		ruleMap:   make(map[int][]*nftables.Rule),
	}
}

func (t *EntryTable) Type() reflect.Type {
	return t.entryType
}

func (t *EntryTable) All() []Entry {
	return t.list
}

func (t *EntryTable) Append(e Entry) error {
	return t.Update(e, nil)
}

func (t *EntryTable) InsertBefore(e Entry, beforeIndex int) error {
	return t.Update(e, &beforeIndex)
}

func (t *EntryTable) Update(e Entry, beforeIndex *int) error {
	if reflect.TypeOf(e) != t.entryType {
		return ErrEntryTypeMismatch
	}

	update := true
	if _, ok := t.ruleMap[e.Index()]; !ok {
		t.counter++
		e.SetIndex(t.counter)
		t.ruleMap[e.Index()] = make([]*nftables.Rule, 0)

		update = false
	}

	if beforeIndex != nil {
		if _, ok := t.ruleMap[*beforeIndex]; !ok {
			beforeIndex = nil
		}
	}

	beforeHandle := (*uint64)(nil)
	if update {
		for i, entry := range t.list {
			if entry.Index() == e.Index() {
				if i+1 < len(t.list) {
					index := t.list[i+1].Index()
					if beforeIndex == nil {
						beforeIndex = &index
					}
					handle := t.ruleMap[index][0].Handle
					beforeHandle = &handle
				}
				t.list = append(t.list[:i], t.list[i+1:]...)
				break
			}
		}
	}

	if beforeIndex != nil {
		for i, entry := range t.list {
			if entry.Index() == *beforeIndex {
				rule := t.ruleMap[entry.Index()][0]
				beforeHandle = &rule.Handle
				t.list = append(t.list[:i+1], t.list[i:]...)
				t.list[i] = e
				break
			}
		}
	} else {
		t.list = append(t.list, e)
	}

	err := e.buildArtifact(t.r)
	if err != nil {
		return err
	}

	if update {
		err := t.removeRules(t.ruleMap[e.Index()])
		if err != nil {
			return err
		}
	}

	t.addRules(e.Index(), beforeHandle, e.ToRules())
	if err := t.r.Update(); err != nil {
		return err
	}
	rules, err := t.findRulesByTag(e.Index())
	if err != nil {
		return err
	}
	t.ruleMap[e.Index()] = rules

	return nil
}

func (t *EntryTable) Remove(index int) error {
	if t.ruleMap[index] != nil {
		{
			if err := t.removeRules(t.ruleMap[index]); err != nil {
				return err
			}
			if err := t.r.Update(); err != nil {
				return err
			}

			t.ruleMap[index] = nil
		}

		for i, entry := range t.list {
			if entry.Index() == index {
				t.list = append(t.list[:i], t.list[i+1:]...)
				break
			}
		}
	} else {
		return ErrEntryIndexNotFound
	}

	return nil
}

func (t *EntryTable) findRulesByTag(tag int) ([]*nftables.Rule, error) {
	r := t.r

	allRules, err := r.nft.GetRules(r.table, t.chain)
	if err != nil {
		return nil, err
	}

	ret := make([]*nftables.Rule, 0)
	for _, rule := range allRules {
		if len(rule.UserData) == 8 && binary.BigEndian.Uint64(rule.UserData) == uint64(tag) {
			ret = append(ret, rule)
		}
	}

	return ret, nil
}

func (t *EntryTable) addRules(tag int, beforeHandle *uint64, rules []*nftables.Rule) {
	r := t.r

	for _, rule := range rules {
		rule.Table = r.table
		rule.Chain = t.chain

		userdata := make([]byte, 8)
		binary.BigEndian.PutUint64(userdata, uint64(tag))
		rule.UserData = userdata

		if beforeHandle != nil {
			rule.Position = *beforeHandle
			r.nft.InsertRule(rule)
		} else {
			r.nft.AddRule(rule)
		}
	}
}

func (t *EntryTable) removeRules(rules []*nftables.Rule) error {
	r := t.r

	for _, rule := range rules {
		err := r.nft.DelRule(rule)
		if err != nil {
			return err
		}
	}

	return nil
}

func InterfaceName(str string) []byte {
	ret := make([]byte, syscall.IFNAMSIZ)
	copy(ret, []byte(str+"\x00"))
	return ret
}

func (r *Router) initNftables() {
	tables, _ := r.nft.ListTables()
	for _, table := range tables {
		if table.Name == "yafw" {
			r.nft.FlushTable(table)
			r.nft.DelTable(table)

			err := r.Update()
			if err != nil {
				log.Printf("update error: %v", err)
			}
		}
	}
	r.table = r.nft.AddTable(&nftables.Table{
		Name:   "yafw",
		Family: nftables.TableFamilyIPv4,
	})

	table := r.table
	nft := r.nft

	defaultPolicy := nftables.ChainPolicyDrop
	// defaultPolicy := nftables.ChainPolicyAccept
	r.forward = nft.AddChain(&nftables.Chain{
		Name:     "forward",
		Table:    table,
		Type:     nftables.ChainTypeFilter,
		Hooknum:  nftables.ChainHookForward,
		Priority: nftables.ChainPriorityFilter,
		Policy:   &defaultPolicy,
	})

	r.postrouting = nft.AddChain(&nftables.Chain{
		Name:     "postrouting",
		Table:    table,
		Type:     nftables.ChainTypeNAT,
		Hooknum:  nftables.ChainHookPostrouting,
		Priority: nftables.ChainPriorityNATSource,
	})

	r.prerouting = nft.AddChain(&nftables.Chain{
		Name:     "prerouting",
		Table:    table,
		Type:     nftables.ChainTypeNAT,
		Hooknum:  nftables.ChainHookPrerouting,
		Priority: nftables.ChainPriorityNATDest,
	})

	builder := &ExprBuilder{}
	builder.ConntrackState(expr.CtStateBitESTABLISHED | expr.CtStateBitRELATED).VerdictAccept()
	r.nft.AddRule(&nftables.Rule{
		Table: r.table,
		Chain: r.forward,
		Exprs: builder.Exprs(),
	})

	err := r.Update()
	if err != nil {
		log.Printf("update error: %v", err)
	}
}

func NewRouterNS(ns netns.NsHandle) (*Router, error) {
	nft, err := nftables.New(nftables.WithNetNSFd(int(ns)))

	if err != nil {
		return nil, err
	}

	nl, err := netlink.NewHandleAt(ns)

	if err != nil {
		return nil, err
	}

	ct, err := conntrack.Dial(nil)
	if err != nil {
		return nil, err
	}

	ret := &Router{
		ns:     ns,
		nft:    nft,
		nl:     nl,
		ct:     ct,
		ipsets: make(map[string]*IPSet),
		zones:  make(map[string]*Zone),
	}

	ret.initNftables()

	ret.snatEntries = NewEntryTable(ret, ret.postrouting, &SNATRule{})
	ret.dnatEntries = NewEntryTable(ret, ret.prerouting, &DNATRule{})
	ret.policyEntries = NewEntryTable(ret, ret.forward, &Policy{})

	return ret, nil
}

func NewRouter() (*Router, error) {
	ns, err := netns.Get()

	if err != nil {
		return nil, err
	}

	return NewRouterNS(ns)
}

func (r *Router) Stop() {
	sets, _ := r.nft.GetSets(r.table)
	r.nft.FlushTable(r.table)
	for _, set := range sets {
		r.nft.DelSet(set)
	}
}

func (r *Router) Update() error {
	if err := r.nft.Flush(); err != nil {
		return err
	}

	return nil
}

func (r *Router) SNATRuleTable() *EntryTable {
	return r.snatEntries
}

func (r *Router) PolicyTable() *EntryTable {
	return r.policyEntries
}
