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

type ZoneTable struct {
	r            *Router
	zoneMap      map[string]*Zone  // zone name to zone data
	interfaceMap map[string]string // interface name to zone name
}

func NewZoneTable(r *Router) *ZoneTable {
	return &ZoneTable{
		r:            r,
		zoneMap:      make(map[string]*Zone),
		interfaceMap: make(map[string]string),
	}
}

func (t *ZoneTable) Update(zone *Zone) error {
	r := t.r
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

	zone, ok := t.zoneMap[zone.Name]
	if !ok {
		t.zoneMap[zone.Name] = zone
	}

	zone.oldM = zone.m

	return nil
}

func (t *ZoneTable) FindZone(name string) *Zone {
	if zone, ok := t.zoneMap[name]; ok {
		return zone
	} else {
		return nil
	}
}

func (r *ZoneTable) All() []*Zone {
	ret := []*Zone{}
	for _, z := range r.zoneMap {
		ret = append(ret, z)
		return ret
	}
	return ret
}

func (t *ZoneTable) AssignInterfaceToZone(iface *net.Interface, zone string) error {
	if oldZone, ok := t.interfaceMap[iface.Name]; ok {
		if oldZone != zone {
			t.zoneMap[oldZone].RemoveInterface(iface)
			t.Update(t.zoneMap[oldZone])
		}
	}

	t.zoneMap[zone].AddInterface(iface)
	t.interfaceMap[iface.Name] = zone
	t.Update(t.zoneMap[zone])

	return nil
}

func (t *ZoneTable) DeleteZone(name string) {
	if zone, ok := t.zoneMap[name]; ok {
		for _, iface := range zone.Members() {
			delete(t.interfaceMap, iface.Name)
		}

		delete(t.zoneMap, name)
	}
}

func (t *ZoneTable) AddZone(name string) *Zone {
	z := &Zone{
		Name:        name,
		Description: "",
		oldM:        make(map[string]*net.Interface),
		m:           make(map[string]*net.Interface),
		set:         nil,
	}

	t.zoneMap[name] = z

	return z
}
