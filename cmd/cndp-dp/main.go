package main

import (
	"flag"
	"github.com/golang/glog"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
	"os"
	"os/signal"
	"syscall"
)

/*
DP represents the overall device plugin.
It contains a list of poolManagers.
*/
type DP struct {
	pools map[string]PoolManager
}

func main() {

	//TODO log to stderr for now, change later
	flag.Parse()
	flag.Lookup("logtostderr").Value.Set("true")

	dp := DP{
		pools: make(map[string]PoolManager),
	}

	//TODO this needs to go in a loop when we add multiple pools
	pm := PoolManager{
		Name:         "cndp/poc",
		Devices:      make(map[string]*pluginapi.Device),
		Socket:       pluginapi.DevicePluginPath + "cndp-poc.sock",
		Endpoint:     "cndp-poc.sock",
		UpdateSignal: make(chan bool),
	}

	err := pm.Init()
	if err != nil {
		glog.Error("Error initializing pool: " + pm.Name)
		glog.Fatal(err)
	}

	dp.pools["cndp-poc"] = pm

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	select {
	case s := <-sigs:
		glog.Infof("Received signal \"%v\"", s)
		for _, pm := range dp.pools {
			glog.Infof("Terminating " + pm.Name)
			pm.Terminate()
		}
		return
	}
}
