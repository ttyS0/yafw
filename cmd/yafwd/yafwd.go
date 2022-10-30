package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/sigeryang/yafw"
)

var socket = "/var/run/yafw.socket"
var logger *log.Logger

type Interface struct {
	Name string `json:"name"`
	MTU  int    `json:"mtu"`
	MAC  string `json:"mac"`
	Up   bool   `json:"up"`
	Zone string `json:"zone"`
}

func APIGetInterfaces(c *gin.Context) {
	ifaces, err := net.Interfaces()
	if err != nil {
		APIError(c, http.StatusInternalServerError, err)
		return
	}

	result := make([]*Interface, 0)
	for _, iface := range ifaces {
		if iface.Flags&net.FlagLoopback != net.FlagLoopback {
			result = append(result, &Interface{
				Name: iface.Name,
				MTU:  iface.MTU,
				MAC:  iface.HardwareAddr.String(),
				Up:   iface.Flags&net.FlagUp == net.FlagUp,
			})
		}
	}

	c.JSON(http.StatusOK, result)
}

func APIPutInterface(c *gin.Context) {
	// TODO: ...
}

func APIGetPolicies(c *gin.Context) {
	policies := router.Policies()

	c.JSON(http.StatusOK, policies)
}

func APIError(c *gin.Context, code int, err error) {
	c.JSON(code, gin.H{
		"ok":      false,
		"message": err.Error(),
	})
}

func APIPostPolicies(c *gin.Context) {
	var p yafw.Policy
	if err := c.BindJSON(&p); err != nil {
		APIError(c, http.StatusBadRequest, err)
		return
	}

	before := c.Query("before")
	beforeIndex := (*int)(nil)
	if before != "" {
		index, err := strconv.Atoi(before)
		if err != nil {
			APIError(c, http.StatusBadRequest, err)
		} else {
			beforeIndex = &index
		}
	}

	var err error
	if beforeIndex != nil {
		err = router.PolicyTable().InsertBefore(&p, *beforeIndex)
	} else {
		err = router.PolicyTable().Append(&p)
	}
	if err != nil {
		APIError(c, http.StatusInternalServerError, err)
	} else {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	}
}

func APIPutPolicy(c *gin.Context) {
	index, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		APIError(c, http.StatusBadRequest, err)
	}

	before := c.Query("before")
	beforeIndex := (*int)(nil)
	if before != "" {
		index, err := strconv.Atoi(before)
		if err != nil {
			APIError(c, http.StatusBadRequest, err)
		} else {
			beforeIndex = &index
		}
	}

	var p yafw.Policy
	if err := c.BindJSON(&p); err != nil {
		APIError(c, http.StatusBadRequest, err)
		return
	}
	p.SetIndex(index)

	if err := router.PolicyTable().Update(&p, beforeIndex); err != nil {
		APIError(c, http.StatusInternalServerError, err)
	} else {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	}
}

func APIDeletePolicy(c *gin.Context) {
	index, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		APIError(c, http.StatusBadRequest, err)
	}

	if err := router.PolicyTable().Remove(index); err != nil {
		APIError(c, http.StatusInternalServerError, err)
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func APIGetNAT(c *gin.Context) {
	c.JSON(http.StatusOK, router.SNATRules())
}

func APIExport(c *gin.Context) {
	cmd := exec.Command("nft", "--json", "list", "ruleset")
	var message json.RawMessage
	output, err := cmd.CombinedOutput()
	if err != nil {
		c.String(http.StatusInternalServerError, "")
	}
	if err := json.Unmarshal(output, &message); err != nil {
		c.String(http.StatusInternalServerError, "")
	}
	c.JSON(http.StatusOK, message)
	// router.AddPolicy(p)
}

func APIGetConnections(c *gin.Context) {
	data, err := os.ReadFile("/proc/net/nf_conntrack")
	if err != nil {
		c.String(http.StatusInternalServerError, "")
	}
	c.JSON(http.StatusOK, gin.H{
		"raw": data,
	})
}

func StartHTTP(wg *sync.WaitGroup) {
	defer wg.Done()

	syscall.Unlink(socket)
	server := gin.Default()

	api := server.Group("/api/v1")
	{
		api.GET("/interfaces", APIGetInterfaces)
		api.GET("/policies", APIGetPolicies)
		api.POST("/policies", APIPostPolicies)
		api.PUT("/policies/:id", APIPutPolicy)
		api.DELETE("/policies/:id", APIDeletePolicy)
		api.GET("/nat", APIGetNAT)
		api.GET("/export", APIExport)
		api.GET("/connections", APIGetConnections)
	}

	err := server.Run(":9085")
	if err != nil {
		logger.Fatalf("cannot start http server: %v", err)
		return
	}
}

var router *yafw.Router

type Config struct {
	Policies []*yafw.Policy   `json:"policies"`
	NAT      []*yafw.SNATRule `json:"nat"`
}

var configFile = flag.String("config", "/app/config.json", "configuration file")

func main() {
	flag.Parse()

	logger = log.New(os.Stdout, "yafwd", log.Ltime|log.Lmsgprefix)
	wg := new(sync.WaitGroup)
	wg.Add(1)
	go StartHTTP(wg)

	var err error
	router, err = yafw.NewRouter()
	if err != nil {
		logger.Fatalf("init router error: %v", err)
		return
	}
	defer router.Stop()

	// ipset := router.NewIPSet("servers")
	// if ipset == nil {
	// 	logger.Fatalf("error init ipset: %v", ipset)
	// 	return
	// }
	// ipset.AddIPRange(yafw.NewIPRangeString("192.168.234.0/24"))
	// router.UpdateIPSet(ipset)

	var config Config
	data, err := os.ReadFile(*configFile)
	if err != nil {
		fmt.Printf("read config error: %v", err)
		return
	}
	if err := json.Unmarshal(data, &config); err != nil {
		fmt.Printf("read config error: %v", err)
		return
	}

	for _, nat := range config.NAT {
		err := router.SNATRuleTable().Append(nat)
		if err != nil {
			fmt.Println(err)
		}
	}

	for _, policy := range config.Policies {
		err := router.PolicyTable().Append(policy)
		if err != nil {
			fmt.Println(err)
		}
	}

	// router.DeletePolicy(1)
	// router.Update()

	// _, ipnet, _ := net.ParseCIDR("0.0.0.0/0")
	// ipset.AddIPNet()

	// _, ipnet, _ := net.ParseCIDR("0.0.0.0/0")
	// router.NewIPSet("Any").AddIPNet(ipnet).RequestUpdate()

	// for _, str := range []string{"10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"} {
	// 	_, ipnet, _ := net.ParseCIDR(str)
	// 	router.NewIPSet("private").AddIPNet(ipnet).RequestUpdate()
	// }
	// router.Update()

	wg.Wait()
}
