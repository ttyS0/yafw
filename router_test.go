package yafw

import (
	"github.com/vishvananda/netns"
)

func newTestRouter() *Router {
	_ = netns.DeleteNamed("yafw-ns")
	ns, err := netns.NewNamed("yafw-ns")

	if err != nil {
		panic(err)
	}

	router, err := NewRouterNS(ns)

	if err != nil {
		panic(err)
	}

	return router
}
