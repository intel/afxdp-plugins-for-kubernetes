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
	"time"
)

/*
PoolManager represents an manages the pool of devices.
Each PoolManager registers with Kubernetes as a different device type.
*/
type PoolManager struct {
	Name         string
	Devices      map[string]*pluginapi.Device
	Server       *grpc.Server
	Socket       string
	Endpoint     string
	UpdateSignal chan bool
	Cndp         cndp.Interface
}

/*
Init is called it initialise the PoolManager.
*/
func (pm *PoolManager) Init(config poolConfig) error {
	pm.Cndp = cndp.NewCndp()

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

	err = pm.discoverResources(config)
	if err != nil {
		return err
	}

	return nil
}

/*
Terminate is called it terminate the PoolManager.
*/
func (pm *PoolManager) Terminate() error {
	pm.stopGRPC()
	pm.cleanup()
	glog.Infof(pm.Name + " terminated")

	return nil
}

/*
ListAndWatch is part of the device plugin API.
Returns a stream list of Devices. Whenever a device state changes,
ListAndWatch should return the new list.
*/
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

/*
Allocate is part of the device plugin API.
Called during container creation so that the Device Plugin can run
device specific operations and instruct Kubelet of the steps to make
the Device available in the container.
*/
func (pm *PoolManager) Allocate(ctx context.Context,
	rqt *pluginapi.AllocateRequest) (*pluginapi.AllocateResponse, error) {
	response := pluginapi.AllocateResponse{}
	sockAddr := pm.Cndp.CreateUdsSocket()

	//loop each container
	for _, crqt := range rqt.ContainerRequests {
		cresp := new(pluginapi.ContainerAllocateResponse)

		cresp.Mounts = append(cresp.Mounts, &pluginapi.Mount{
			HostPath:      sockAddr,
			ContainerPath: "/tmp/cndp.sock",
			ReadOnly:      false,
		})

		//loop each device request per container
		for _, dev := range crqt.DevicesIDs {
			glog.Info("Allocating device " + dev)
		 bpf.Load_bpf_send_xsk_map("ens786f3") //TODO - temporary dummy call to CGo
		}
		response.ContainerResponses = append(response.ContainerResponses, cresp)
	}

	go pm.Cndp.StartSocketServer(sockAddr)
	return &response, nil
}

/*
GetDevicePluginOptions is part of the device plugin API.
Unused.
*/
func (pm *PoolManager) GetDevicePluginOptions(context.Context, *pluginapi.Empty) (*pluginapi.DevicePluginOptions, error) {
	return &pluginapi.DevicePluginOptions{}, nil
}

/*
PreStartContainer is part of the device plugin API.
Unused.
*/
func (pm *PoolManager) PreStartContainer(context.Context, *pluginapi.PreStartContainerRequest) (*pluginapi.PreStartContainerResponse, error) {
	return &pluginapi.PreStartContainerResponse{}, nil
}

/*
GetPreferredAllocation is part of the device plugin API.
Unused.
*/
func (pm *PoolManager) GetPreferredAllocation(context.Context, *pluginapi.PreferredAllocationRequest) (*pluginapi.PreferredAllocationResponse, error) {
	return &pluginapi.PreferredAllocationResponse{}, nil
}

func (pm *PoolManager) discoverResources(config poolConfig) error {

	for _, device := range config.Devices {
		newdev := pluginapi.Device{ID: device, Health: pluginapi.Healthy}
		pm.Devices[device] = &newdev
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
		ResourceName: devicePrefix + "/" + pm.Name,
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
