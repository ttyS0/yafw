package yafw

import (
	"github.com/google/nftables/expr"
)

type ServiceGroup struct {
	Name     string     `json:"name"`
	Services []*Service `json:"services"`
}

type Service struct {
	Name               string `json:"name"`
	Protocol           uint8  `json:"protocol"`
	SourcePortMin      uint16 `json:"source_port_min"`
	SourcePortMax      uint16 `json:"source_port_max"`
	DestinationPortMin uint16 `json:"destination_port_min"`
	DestinationPortMax uint16 `json:"destination_port_max"`
}

func (s *Service) Exprs() []expr.Any {
	builder := &ExprBuilder{}
	builder.MetaL4Protocol(1).CompareL4Protocol(1, s.Protocol)

	if s.SourcePortMin != 0 && s.SourcePortMax != 0 {
		builder.LoadSourcePort(1).ComparePortRange(1, s.SourcePortMin, s.SourcePortMax)
	}

	if s.DestinationPortMin != 0 && s.DestinationPortMax != 0 {
		builder.LoadDestinationPort(1).ComparePortRange(1, s.DestinationPortMin, s.DestinationPortMax)
	}

	return builder.Exprs()
}

// func (r *Router) ServiceGroups() []*ServiceGroup {
// 	ret := make([]*ServiceGroup, 0)
// 	for _, sg := range r.serviceGroups {
// 		ret = append(ret, sg)
// 	}
// 	return ret
// }

// func (r *Router) AddService(sg *ServiceGroup) error {
// 	if _, ok := r.serviceGroups[sg.Name]; ok {
// 		return fmt.Errorf("service group name duplicated")
// 	}

// 	r.serviceGroups[sg.Name] = sg

// 	return nil
// }

// func (r *Router) UpdateService(sg *ServiceGroup) {
// 	r.serviceGroups[sg.Name] = sg
// }

// func (r *Router) DeleteService(sg *ServiceGroup) error {
// 	if _, ok := r.serviceGroups[sg.Name]; ok {
// 		delete(r.serviceGroups, sg.Name)
// 	} else {
// 		return fmt.Errorf("service group name not found")
// 	}

// 	return nil
// }
