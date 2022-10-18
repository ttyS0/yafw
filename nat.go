package yafw

import (
	"fmt"
	"net"

	"github.com/google/nftables"
	"github.com/google/nftables/expr"
)

type SNATTarget int

const (
	// Masquerade to the egress IP
	SNATEgress SNATTarget = iota
	// SNAT to specific IP
	SNATSpecific
)

// TODO: SNAT mode support

// type SNATMode int
// const (
// 	SNATStatic SNATMode = iota
// 	SNATDynamic
// 	SNATPort
// )

type SNATRule struct {
	ID            int        `json:"id"`
	Description   string     `json:"description"`
	Enabled       bool       `json:"enabled"`
	Source        *Address   `json:"source"`
	Destination   *Address   `json:"destination"`
	Egress        string     `json:"egress"`
	Target        SNATTarget `json:"target"`
	TargetAddress *Address   `json:"target_address"`
	Log           bool       `json:"log"`
	// Mode        SNATMode

	artifact *SNATRuleArtifact
}

type SNATRuleArtifact struct {
	Source      *nftables.Set
	Destination *nftables.Set
	Egress      *net.Interface
}

type DNATRule struct {
	Source      *net.IPNet
	Destination *net.IPNet
	Egress      *net.Interface
	// Target      DNATTarget
	IP net.IPNet
}

func (r *Router) SNATRules() (ret []*SNATRule) {
	ret = make([]*SNATRule, 0)

	for _, entry := range r.snatEntries.All() {
		ret = append(ret, entry.(*SNATRule))
	}

	return ret
}

// the following contents implement Entry in entry.go

func (snat *SNATRule) buildArtifact(router *Router) error {
	artifact := &SNATRuleArtifact{}

	if snat.Source != nil {
		set, err := router.addressToSet(snat.Source)
		if err != nil {
			return err
		}
		artifact.Source = set
	}

	if snat.Destination != nil {
		set, err := router.addressToSet(snat.Destination)
		if err != nil {
			return err
		}
		artifact.Destination = set
	}

	if snat.Egress != "" {
		iface, err := net.InterfaceByName(snat.Egress)
		if err != nil {
			return err
		}
		artifact.Egress = iface
	}

	snat.artifact = artifact

	return nil
}

func (snat *SNATRule) Index() int {
	return snat.ID
}

func (snat *SNATRule) SetIndex(index int) {
	snat.ID = index
}

func (snat *SNATRule) ToRules() []*nftables.Rule {
	builder := &ExprBuilder{}
	artifact := snat.artifact

	if artifact != nil {
		if snat.Egress != "" {
			builder.MetaEgressInterface(1).CompareInterfaceName(1, artifact.Egress.Name)
		}

		if snat.Source != nil {
			builder.PayloadIPSource(1).LookupSet(1, artifact.Source)
		}

		if snat.Destination != nil {
			builder.PayloadIPSource(1).LookupSet(1, artifact.Destination)
		}
	}

	builder.Append(&expr.Log{
		Data:  []byte("yafw-snat"),
		Flags: expr.LogFlagsIPOpt | expr.LogFlagsTCPOpt,
	})

	switch snat.Target {
	case SNATEgress:
		builder.Masquerade()
		// case SNATSpecific:
		// 	builder.SourceNATIP(snat.TargetAddress)
	}

	fmt.Printf("nat builder: %v\n", len(builder.Exprs()))

	return []*nftables.Rule{
		{
			Exprs: builder.Exprs(),
		},
	}
}
