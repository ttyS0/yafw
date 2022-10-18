package yafw

import (
	"fmt"
	"net"

	"github.com/google/nftables"
)

type Zone struct {
	Name        string `json:"name"`
	Description string `json:"description"`

	set  *nftables.Set
	oldM map[string]*net.Interface
	m    map[string]*net.Interface
}

func (r *Router) NewZone(name string) *Zone {
	found := r.FindZone(name)

	if found != nil {
		return nil
	}

	ret := &Zone{
		Name: name,
		m:    make(map[string]*net.Interface),
	}
	r.zones[name] = ret

	return ret
}

func (r *Router) FindZone(name string) *Zone {
	zone, ok := r.zones[name]
	if ok {
		return zone
	} else {
		return nil
	}
}

func (r *Router) Zones(name string) []*Zone {
	ret := []*Zone{}
	for _, z := range r.zones {
		ret = append(ret, z)
		return ret
	}
	return ret
}

func (r *Router) DeleteZone(name string) {
	delete(r.zones, name)
}

func (z *Zone) element(member *net.Interface) []nftables.SetElement {
	ret := []nftables.SetElement{
		{
			Key: InterfaceName(member.Name),
		},
	}

	return ret
}

func (z *Zone) elements() []nftables.SetElement {
	ret := []nftables.SetElement(nil)
	for _, member := range z.m {
		if member != nil {
			ret = append(ret, z.element(member)...)
		}
	}

	return ret
}

func (r *Router) UpdateZone(zone *Zone) error {
	nft := r.nft

	if zone.set != nil {
		// incremental update

		for key, member := range zone.m {
			if _, ok := zone.oldM[key]; !ok {
				// incremental addition
				nft.SetAddElements(zone.set, zone.element(member))
			}
		}

		for key, member := range zone.oldM {
			if _, ok := zone.m[key]; !ok {
				// incremental deletion
				nft.SetDeleteElements(zone.set, zone.element(member))
			}
		}
	} else {
		// create a new set
		elements := zone.elements()

		zone.set = &nftables.Set{
			Table:   r.table,
			Name:    fmt.Sprintf("zone-%s", zone.Name),
			KeyType: nftables.TypeIFName,
		}

		err := nft.AddSet(zone.set, elements)
		if err != nil {
			return err
		}
	}

	if err := r.Update(); err != nil {
		return err
	}

	zone, ok := r.zones[zone.Name]
	if !ok {
		r.zones[zone.Name] = zone
	}

	zone.oldM = zone.m

	return nil
}

func (z *Zone) AddInterface(iface *net.Interface) *Zone {
	_, ok := z.m[iface.Name]
	if !ok {
		z.m[iface.Name] = iface
	}
	return z
}

func (z *Zone) RemoveInterface(iface *net.Interface) *Zone {
	delete(z.m, iface.Name)
	return z
}

func (z *Zone) Members() []*net.Interface {
	ret := make([]*net.Interface, 0)
	for _, member := range z.m {
		ret = append(ret, member)
	}
	return ret
}
