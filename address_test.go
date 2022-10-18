package yafw

import (
	"net"
	"os/exec"
	"testing"
)

func TestIPNetLast(t *testing.T) {
	want := [][2]string{
		{"0.0.0.0/0", "255.255.255.255"},
		{"192.168.1.0/24", "192.168.1.255"},
		{"10.255.255.0/24", "10.255.255.255"},
	}

	for _, w := range want {
		_, ipnet, err := net.ParseCIDR(w[0])
		if err != nil {
			t.Fatalf("error parse CIDR %s: %v", w[0], err)
		}
		wanted := net.ParseIP(w[1])
		if err != nil {
			t.Fatalf("error parse IP %s: %v", w[1], err)
		}

		test := IPNetLast(ipnet)
		if !test.Equal(wanted) {
			t.Fatalf("test assert error: IPNetLast(%v) = %v (expecting %v)", ipnet, test, wanted)
		}
	}
}

func TestIPNetEnd(t *testing.T) {
	want := [][2]string{
		{"0.0.0.0/0", "0.0.0.0"},
		{"192.168.1.0/24", "192.168.2.0"},
		{"10.255.255.0/24", "11.0.0.0"},
	}

	for _, w := range want {
		_, ipnet, err := net.ParseCIDR(w[0])
		if err != nil {
			t.Fatalf("error parse CIDR %s: %v", w[0], err)
		}
		wanted := net.ParseIP(w[1])
		if err != nil {
			t.Fatalf("error parse IP %s: %v", w[1], err)
		}

		test := IPNetEnd(ipnet)
		if !test.Equal(wanted) {
			t.Fatalf("test assert error: IPNetEnd(%v) = %v (expecting %v)", ipnet, test, wanted)
		}
	}
}

func TestIPSet(t *testing.T) {
	err := error(nil)
	router := newTestRouter()

	ipset := router.NewIPSet("test-ipset")

	members := []string{
		"192.168.10.1",
		"192.168.11.1",
		"192.168.1.0/24",
		"192.168.100.1/32",
		"192.168.233.1/32",
		"192.168.233.0/24",
		"10.255.255.0/24",
		"11.255.255.254/32",
		"192.168.6.0-192.168.6.120",
	}

	for _, member := range members {
		r := NewIPRangeString(member)
		ipset.AddIPRange(r)
	}

	t.Logf("Current members: %v", ipset.Members())

	err = router.UpdateIPSet(ipset)
	if err != nil {
		t.Fatal(err)
	}

	elements, err := router.nft.GetSetElements(ipset.set)
	if err != nil {
		t.Fatal(err)
	}

	for _, element := range elements {
		t.Logf("%v", element)
	}

	out, err := exec.Command("nft", "list", "set", "ip", "yafw", "ipset-test-ipset").CombinedOutput()
	t.Logf("NFT Output:\n%s\n", out)
	if err != nil {
		t.Fatal(err)
	}
}
