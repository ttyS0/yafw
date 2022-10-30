package yafw

import (
	"net"
	"os/exec"
	"testing"
)

func TestZone(t *testing.T) {
	router := newTestRouter()

	zone := router.zones.AddZone("trust")

	// other interfaces will not be available since the test is now under a
	// new network namespace.
	ifnames := []string{
		"lo",
	}

	for _, ifname := range ifnames {
		iface, err := net.InterfaceByName(ifname)
		if err != nil {
			t.Fatalf("fatal error find interface %s: %v", ifname, err)
		}
		zone.AddInterface(iface)
	}

	router.zones.Update(zone)

	out, err := exec.Command("nft", "-j", "list", "set", "ip", "yafw", zone.set.Name).CombinedOutput()
	t.Logf("NFT Output:\n%s\n", out)
	if err != nil {
		t.Fatal(err)
	}
}
