package yafw

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"strings"

	"github.com/google/nftables"
)

type AddressType int

const (
	// immediate addresses:
	// either single IPRange (immediate in nftables expressions)
	// or multiple IPRanges (anonymous set in nftables)
	AddressImmediate AddressType = iota

	// a reference to one of commonly defined IPSets
	AddressIPSet
)

type Address struct {
	t         AddressType
	Immediate []*IPRange `json:"immediate"`
	IPSet     string     `json:"ipset"`
}

func NewAddressImmediate(immediate []*IPRange) *Address {
	return &Address{
		t:         AddressImmediate,
		Immediate: immediate,
	}
}

func NewAddressIPSet(ipset string) *Address {
	return &Address{
		t:     AddressIPSet,
		IPSet: ipset,
	}
}

func (ad *Address) Type() AddressType {
	return ad.t
}

func (r *Address) UnmarshalJSON(data []byte) error {
	ipranges := []string{}
	err := json.Unmarshal(data, &ipranges)
	if err != nil {
		ipset := ""
		err := json.Unmarshal(data, &ipset)
		if err != nil {
			return fmt.Errorf("cannot convert to addresss")
		}

		r.t = AddressIPSet
		r.IPSet = ipset
		return nil
	}

	r.t = AddressImmediate
	for _, iprange := range ipranges {
		r.Immediate = append(r.Immediate, NewIPRangeString(iprange))
	}

	return nil
}

func (r *Address) MarshalJSON() ([]byte, error) {
	switch r.Type() {
	case AddressIPSet:
		return json.Marshal(r.IPSet)
	case AddressImmediate:
		ipranges := []string{}
		for _, iprange := range r.Immediate {
			ipranges = append(ipranges, iprange.String())
		}
		return json.Marshal(ipranges)
	default:
		return nil, fmt.Errorf("unknown address type")
	}
}

func (r *Address) String() string {
	switch r.Type() {
	case AddressIPSet:
		return fmt.Sprintf("ipset:%s", r.IPSet)
	case AddressImmediate:
		ipranges := []string{}
		for _, iprange := range r.Immediate {
			ipranges = append(ipranges, iprange.String())
		}
		return fmt.Sprintf("[%s]", strings.Join(ipranges, ","))
	default:
		return "(unknown)"
	}
}

type IPSet struct {
	set *nftables.Set

	// incremental update
	willAdd    []*IPRange
	willDelete []*IPRange

	name    string
	members []*IPRange
}

func IPMaskedLast(ip net.IP, mask net.IPMask) net.IP {
	n := len(ip)
	out := make(net.IP, n)
	for i := 0; i < n; i++ {
		out[i] = ip[i] | ^mask[i]
	}
	return out
}

func IPMaskedEnd(ip net.IP, mask net.IPMask) net.IP {
	n := len(ip)
	out := make(net.IP, n)
	copy(out, ip)

	// Get the lowbit in a byte. lowbit(x) = the lowest non-zero bit mask
	//
	// e.g. lowbit(11100100b) = 00000100b
	//                  ^ lowest non-zero bit
	lowbit := func(x byte) byte { return x & (x ^ (x - 1)) }

	previousLowbit := byte(0)
	for i := n - 1; i >= 0; i-- {
		if mask[i] == 0 {
			continue
		}
		currentLowbit := lowbit(mask[i])
		if previousLowbit == 0 {
			// We are now at the last non-zero segment of mask.
			//
			// e.g. 255.255.224.0
			//               ^ we are here

			out[i] += currentLowbit
		} else if previousLowbit > out[i+1] {
			// Indicates a carry at the previous segment.

			out[i] += currentLowbit
		}

		previousLowbit = currentLowbit
	}
	return out
}

// Get the last IP in a range represented by IPNet.
//
// e.g. IPNetLast of "192.168.1.0/24" is "192.168.1.255"
func IPNetLast(ip *net.IPNet) net.IP {
	return IPMaskedLast(ip.IP, ip.Mask)
}

// Get the end IP in a range represented by IPNet.
//
// The end IP is the start of the next range.
//
// e.g. IPNetEnd of "192.168.1.0/24" is "192.168.2.0"
func IPNetEnd(ip *net.IPNet) net.IP {
	return IPMaskedEnd(ip.IP, ip.Mask)
}

func IPNetEqual(a *net.IPNet, b *net.IPNet) bool {
	return a.IP.Equal(b.IP) && bytes.Equal(a.Mask, b.Mask)
}

func IPNext(ip net.IP) net.IP {
	n := len(ip)
	out := make(net.IP, n)
	copy(out, ip)

	carry := byte(1)
	for i := n - 1; i >= 0; i-- {
		if carry != 0 {
			// out[i] = ip[i] + carry
			out[i] += carry

			if out[i] < ip[i] {
				carry = 1
			} else {
				carry = 0
			}
		}
	}

	return out
}

// func removeIPNet(nets []*net.IPNet, index int) []*net.IPNet {
// 	return append(nets[:index], nets[index+1:]...)
// }

// func findIPNet(nets []*net.IPNet, ipnet *net.IPNet) int {
// 	ret := -1
// 	for i, member := range nets {
// 		if IPNetEqual(member, ipnet) {
// 			return i
// 		}
// 	}
// 	return ret
// }

// remove specific index from a slice of IPRange
func removeIPRange(ranges []*IPRange, index int) []*IPRange {
	return append(ranges[:index], ranges[index+1:]...)
}

// find in a slice of IPRange
func findIPRange(ranges []*IPRange, find *IPRange) int {
	ret := -1
	for i, r := range ranges {
		if r.Equal(find) {
			return i
		}
	}

	return ret
}

func setElementsFromIPRanges(nets []*IPRange) []nftables.SetElement {
	ret := []nftables.SetElement{}
	for _, member := range nets {
		ret = append(ret, nftables.SetElement{
			Key:    member.First(),
			KeyEnd: member.End(),
		})
	}
	return ret
}

func (r *Router) MakeImmediateAddress(address *Address) (*nftables.Set, error) {
	nft := r.nft

	set := &nftables.Set{
		Table:         r.table,
		Concatenation: true,
		Interval:      true,
		Anonymous:     true,
		Constant:      true,
		KeyType:       nftables.TypeIPAddr,
	}

	if err := nft.AddSet(set, setElementsFromIPRanges(address.Immediate)); err != nil {
		return nil, err
	}

	return set, nil
}

func (r *Router) addressToSet(address *Address) (*nftables.Set, error) {
	switch address.Type() {
	case AddressIPSet:
		ipset := r.FindIPSet(address.IPSet)
		return ipset.set, nil
	case AddressImmediate:
		return r.MakeImmediateAddress(address)
	default:
		return nil, fmt.Errorf("unsupported address type")
	}
}

func (r *Router) NewIPSet(name string) *IPSet {
	found := r.FindIPSet(name)

	if found != nil {
		return nil
	}

	ret := &IPSet{
		name: name,
	}
	r.ipsets[name] = ret

	return ret
}

func (r *Router) FindIPSet(name string) *IPSet {
	ipset, ok := r.ipsets[name]
	if ok {
		return ipset
	} else {
		return nil
	}
}

func (r *Router) UpdateIPSet(ipset *IPSet) error {
	defer func() {
		ipset.willAdd = nil
		ipset.willDelete = nil
	}()

	nft := r.nft

	if ipset.set != nil {
		// incremental update
		willAdd := setElementsFromIPRanges(ipset.willAdd)
		willDelete := setElementsFromIPRanges(ipset.willDelete)

		if len(willAdd) > 0 {
			err := nft.SetAddElements(ipset.set, willAdd)
			if err != nil {
				return err
			}
		}
		if len(willDelete) > 0 {
			err := nft.SetDeleteElements(ipset.set, willDelete)
			if err != nil {
				return err
			}
		}
	} else {
		// create a new set
		elements := setElementsFromIPRanges(ipset.members)

		ipset.set = &nftables.Set{
			Table:    r.table,
			Name:     fmt.Sprintf("ipset-%s", ipset.name),
			Interval: true,
			KeyType:  nftables.TypeIPAddr,
		}

		err := nft.AddSet(ipset.set, elements)
		if err != nil {
			return err
		}
	}

	if err := r.Update(); err != nil {
		return err
	}

	ipset, ok := r.ipsets[ipset.name]
	if !ok {
		r.ipsets[ipset.name] = ipset
	}

	return nil
}

func (s *IPSet) Name() string {
	return s.name
}

func (s *IPSet) Members() []*IPRange {
	return s.members
}

func (s *IPSet) AddIPRange(r *IPRange) *IPSet {
	if r == nil {
		return s
	}

	if findIPRange(s.willAdd, r) >= 0 {
		// already in the incremental addition list, do nothing
		return s
	}

	index := findIPRange(s.members, r)

	if index < 0 {
		// not found in our members, add it
		s.members = append(s.members, r)

		// append it in the incremental addition list
		s.willAdd = append(s.willAdd, r)

		// lookup in the incremental deletion list, remove if found
		if index := findIPRange(s.willDelete, r); index != -1 {
			s.willDelete = removeIPRange(s.willDelete, index)
		}
	}

	return s
}

func (s *IPSet) DeleteIPRange(r *IPRange) *IPSet {
	if r == nil {
		return s
	}

	if findIPRange(s.willDelete, r) >= 0 {
		// already in the incremental deletion list, do nothing
		return s
	}

	index := findIPRange(s.members, r)

	if index >= 0 {
		// found in our members, remove it
		s.members = removeIPRange(s.members, index)

		// append it in the incremental deletion list
		s.willDelete = append(s.willDelete, r)

		// lookup in the incremental addition list, remove if found
		if index := findIPRange(s.willAdd, r); index != -1 {
			s.willAdd = removeIPRange(s.willAdd, index)
		}
	}

	return s
}

type IPRangeType int

const (
	IPRangeHost IPRangeType = iota
	IPRangeNet
	IPRangeInterval
)

type IPRange struct {
	t IPRangeType

	net   *net.IPNet
	first net.IP
	last  net.IP
}

func NewIPRangeHost(host net.IP) *IPRange {
	return &IPRange{
		t:     IPRangeHost,
		first: host,
	}
}

func NewIPRangeNet(net *net.IPNet) *IPRange {
	return &IPRange{
		t:   IPRangeNet,
		net: net,
	}
}

func NewIPRange(first net.IP, last net.IP) *IPRange {
	return &IPRange{
		t:     IPRangeInterval,
		first: first,
		last:  last,
	}
}

func NewIPRangeString(ip string) *IPRange {
	trial := net.ParseIP(ip).To4()
	if trial != nil && trial.To4() != nil {
		return &IPRange{
			t:     IPRangeHost,
			first: trial.To4(),
		}
	}

	trial, trialNet, err := net.ParseCIDR(ip)
	if trialNet != nil && trial.To4() != nil {
		return &IPRange{
			t:   IPRangeNet,
			net: trialNet,

			// unused
			first: trial.To4(),
		}
	}

	if err != nil && strings.Contains(ip, "-") {
		interval := strings.Split(ip, "-")
		if len(interval) == 2 {
			first := net.ParseIP(strings.TrimSpace(interval[0]))
			last := net.ParseIP(strings.TrimSpace(interval[1]))

			if first != nil && last != nil {
				first = first.To4()
				last = last.To4()
				if first != nil && last != nil {
					return &IPRange{
						t:     IPRangeInterval,
						first: first,
						last:  last,
					}
				}
			}
		}
	}

	return nil
}

func (r *IPRange) Type() IPRangeType {
	return r.t
}

func (r *IPRange) Start() net.IP {
	switch r.t {
	case IPRangeHost:
		return r.first
	case IPRangeNet:
		return r.net.IP
	case IPRangeInterval:
		return r.first
	default:
		return nil
	}
}

func (r *IPRange) First() net.IP {
	return r.Start()
}

func (r *IPRange) Last() net.IP {
	switch r.t {
	case IPRangeHost:
		return r.first
	case IPRangeNet:
		return IPNetLast(r.net)
	case IPRangeInterval:
		return r.last
	default:
		return nil
	}
}

func (r *IPRange) End() net.IP {
	switch r.t {
	case IPRangeHost:
		return IPNext(r.first)
	case IPRangeNet:
		return IPNetEnd(r.net)
	case IPRangeInterval:
		return IPNext(r.last)
	default:
		return nil
	}
}

func (r *IPRange) Equal(e *IPRange) bool {
	return r.Type() == e.Type() &&
		r.First().Equal(e.First()) &&
		r.Last().Equal(e.Last())
}

func (r *IPRange) String() string {
	switch r.t {
	case IPRangeHost:
		return r.first.String()
	case IPRangeNet:
		return r.net.String()
	case IPRangeInterval:
		return fmt.Sprintf("%s-%s", r.first.String(), r.last.String())
	default:
		return "(invalid)"
	}
}

func (r *IPRange) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	temp := NewIPRangeString(s)

	if temp != nil {
		return fmt.Errorf("cannot unmarshal %s to IPRange", string(data))
	}

	r.t = temp.t
	r.net = temp.net
	r.first = temp.first
	r.last = temp.last

	return nil
}

func (r *IPRange) MarshalJSON() ([]byte, error) {
	return []byte("\"" + r.String() + "\""), nil
}
