package yafw

import (
	"encoding/json"

	"github.com/google/nftables"
	"github.com/google/nftables/expr"
)

type PolicyAction int

const (
	PolicyAccept PolicyAction = iota
	PolicyDrop
)

func (action PolicyAction) MarshalJSON() ([]byte, error) {
	switch action {
	case PolicyAccept:
		return json.Marshal("accept")
	case PolicyDrop:
		return json.Marshal("drop")
	default:
		return json.Marshal("(unknown)")
	}
}

func (action *PolicyAction) UnmarshalJSON(data []byte) error {
	text := ""
	if err := json.Unmarshal(data, &text); err != nil {
		return err
	}
	switch text {
	case "accept":
		*action = PolicyAccept
	case "drop":
		*action = PolicyDrop
	}

	return nil
}

type Policy struct {
	ID          int          `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Log         bool         `json:"log"`
	Action      PolicyAction `json:"action"`

	Source          *Address `json:"source"`
	SourceZone      string   `json:"source_zone"`
	Destination     *Address `json:"destination"`
	DestinationZone string   `json:"destination_zone"`
	Service         *Service `json:"service"`

	artifact *PolicyArtifact
}

type PolicyArtifact struct {
	Source          *nftables.Set
	SourceZone      *nftables.Set
	Destination     *nftables.Set
	DestinationZone *nftables.Set
}

func (r *Router) Policies() (ret []*Policy) {
	ret = make([]*Policy, 0)

	for _, entry := range r.policyEntries.All() {
		ret = append(ret, entry.(*Policy))
	}

	return ret
}

// the following contents implement Entry in entry.go

func (policy *Policy) buildArtifact(router *Router) error {
	artifact := &PolicyArtifact{}

	// if policy.SourceZone != "" {
	// 	zone := router.zones.FindZone(policy.SourceZone)
	// 	if zone != nil {
	// 		artifact.SourceZone = zone.set
	// 	} else {
	// 		return fmt.Errorf("zone \"%s\" not found", policy.SourceZone)
	// 	}
	// }

	// if policy.DestinationZone != "" {
	// 	zone := router.zones.FindZone(policy.DestinationZone)
	// 	if zone != nil {
	// 		artifact.DestinationZone = zone.set
	// 	} else {
	// 		return fmt.Errorf("zone \"%s\" not found", policy.DestinationZone)
	// 	}
	// }

	if policy.Source != nil {
		set, err := router.addressToSet(policy.Source)
		if err != nil {
			return err
		}
		artifact.Source = set
	}

	if policy.Destination != nil {
		set, err := router.addressToSet(policy.Destination)
		if err != nil {
			return err
		}
		artifact.Destination = set
	}

	policy.artifact = artifact

	return nil
}

func (policy *Policy) Index() int {
	return policy.ID
}

func (policy *Policy) SetIndex(index int) {
	policy.ID = index
}

func (policy *Policy) ToRules() []*nftables.Rule {
	builder := &ExprBuilder{}
	artifact := policy.artifact

	if artifact != nil {
		if policy.SourceZone != "" && artifact.SourceZone != nil {
			builder.MetaIngressInterface(1).LookupSet(1, artifact.SourceZone)
		}

		if policy.DestinationZone != "" && artifact.DestinationZone != nil {
			builder.MetaIngressInterface(1).LookupSet(1, artifact.DestinationZone)
		}

		if policy.Source != nil && artifact.Source != nil {
			builder.PayloadIPSource(1).LookupSet(1, artifact.Source)
		}

		if policy.Destination != nil && artifact.Destination != nil {
			builder.PayloadIPDestination(1).LookupSet(1, artifact.Destination)
		}
	}

	if policy.Service != nil {
		builder.AppendGroup(policy.Service.Exprs())
	}

	if policy.Log {
		builder.Append(&expr.Log{
			Data:  []byte("yafw-policy"),
			Flags: expr.LogFlagsIPOpt | expr.LogFlagsTCPOpt,
		})
	}

	switch policy.Action {
	case PolicyAccept:
		builder.VerdictAccept()
	case PolicyDrop:
		builder.VerdictDrop()
	}

	return []*nftables.Rule{
		{
			Exprs: builder.Exprs(),
		},
	}
}
