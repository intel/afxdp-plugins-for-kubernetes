package main

import (
	"fmt"
	"github.com/golang/glog"
	"github.com/intel/cndp_device_plugin/pkg/bpf"
	"github.com/intel/cndp_device_plugin/pkg/cndp"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
	"net"
	"os"
	"strconv"
	"time"
)

type PoolManager struct {
	Name         string
	Devices      map[string]*pluginapi.Device
	Server       *grpc.Server
	Socket       string
	Endpoint     string
	UpdateSignal chan bool
}

func (pm *PoolManager) Init() error {
	err := pm.registerWithKubelet()
	if err != nil {
		return err
	}
	glog.Infof(pm.Name + " registered with Kubelet")

	err = pm.startGRPC()
	if err != nil {
		return err
	}
	glog.Infof("Starting to serve on %s", pm.Socket)

	err = pm.discoverResources()
	if err != nil {
		return err
	}

	return nil
}

func (pm *PoolManager) Terminate() error {
	pm.stopGRPC()
	pm.cleanup()
	glog.Infof(pm.Name + " terminated")

	return nil
}

func (pm *PoolManager) ListAndWatch(emtpy *pluginapi.Empty,
	stream pluginapi.DevicePlugin_ListAndWatchServer) error {

	glog.Info(pm.Name + " ListAndWatch started")

	for {
		select {
		case <-pm.UpdateSignal:
			resp := new(pluginapi.ListAndWatchResponse)
			glog.Infof(pm.Name + " device list:")

			for _, dev := range pm.Devices {
				glog.Infof("\t" + dev.ID + ", " + dev.Health)
				resp.Devices = append(resp.Devices, dev)
			}

			if err := stream.Send(resp); err != nil {
				glog.Errorf("Failed to send stream to kubelet: %v", err)
			}
		}
	}
}

func (pm *PoolManager) Allocate(ctx context.Context,
	rqt *pluginapi.AllocateRequest) (*pluginapi.AllocateResponse, error) {

	response := pluginapi.AllocateResponse{}
	sockAddr := cndp.CreateUdsSocket()

	//loop each container
	for _, crqt := range rqt.ContainerRequests {
		cresp := new(pluginapi.ContainerAllocateResponse)
		//loop each device request per container
		for _, dev := range crqt.DevicesIDs {
			cresp.Mounts = append(cresp.Mounts, &pluginapi.Mount{
				HostPath:      sockAddr,
				ContainerPath: "/tmp/cndp.sock",
				ReadOnly:      false,
			})
			glog.Info("Allocating device " + dev)
			bpf.LoadBpfProgram() //TODO - temporary dummy call to CGo
		}
		response.ContainerResponses = append(response.ContainerResponses, cresp)
	}

	go cndp.StartSocketServer(sockAddr)
	return &response, nil
}

func (pm *PoolManager) GetDevicePluginOptions(context.Context, *pluginapi.Empty) (*pluginapi.DevicePluginOptions, error) {
	return &pluginapi.DevicePluginOptions{}, nil
}

func (pm *PoolManager) PreStartContainer(context.Context, *pluginapi.PreStartContainerRequest) (*pluginapi.PreStartContainerResponse, error) {
	return &pluginapi.PreStartContainerResponse{}, nil
}

func (pm *PoolManager) GetPreferredAllocation(context.Context, *pluginapi.PreferredAllocationRequest) (*pluginapi.PreferredAllocationResponse, error) {
	return &pluginapi.PreferredAllocationResponse{}, nil
}

func (pm *PoolManager) discoverResources() error {

	for i := 1; i <= 5; i++ {
		devName := "dev_" + strconv.Itoa(i)
		glog.Info("Discovered Resource " + devName)
		newdev := pluginapi.Device{ID: devName, Health: pluginapi.Healthy}
		pm.Devices[devName] = &newdev
	}

	if len(pm.Devices) > 0 {
		pm.UpdateSignal <- true
	}

	return nil
}

func (pm *PoolManager) registerWithKubelet() error {
	conn, err := grpc.Dial(pluginapi.KubeletSocket, grpc.WithInsecure(),
		grpc.WithDialer(func(addr string, timeout time.Duration) (net.Conn, error) {
			return net.DialTimeout("unix", addr, timeout)
		}))
	defer conn.Close()
	if err != nil {
		return fmt.Errorf("Error registering with Kubelet: %v", err)
	}
	client := pluginapi.NewRegistrationClient(conn)

	reqt := &pluginapi.RegisterRequest{
		Version:      pluginapi.Version,
		Endpoint:     pm.Endpoint,
		ResourceName: pm.Name,
	}

	_, err = client.Register(context.Background(), reqt)
	if err != nil {
		return fmt.Errorf("Error registering with Kubelet: %v", err)
	}

	return nil
}

func (pm *PoolManager) startGRPC() error {
	err := pm.cleanup()
	if err != nil {
		return err
	}

	sock, err := net.Listen("unix", pm.Socket)
	if err != nil {
		return err
	}

	pm.Server = grpc.NewServer([]grpc.ServerOption{}...)
	pluginapi.RegisterDevicePluginServer(pm.Server, pm)
	go pm.Server.Serve(sock)

	conn, err := grpc.Dial(pm.Socket, grpc.WithInsecure(), grpc.WithBlock(),
		grpc.WithTimeout(5*time.Second),
		grpc.WithDialer(func(addr string, timeout time.Duration) (net.Conn, error) {
			return net.DialTimeout("unix", addr, timeout)
		}),
	)
	conn.Close()

	if err != nil {
		return err
	}

	return nil
}

func (pm *PoolManager) stopGRPC() {
	if pm.Server != nil {
		pm.Server.Stop()
		pm.Server = nil
	}
}

func (pm *PoolManager) cleanup() error {
	if err := os.Remove(pm.Socket); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
