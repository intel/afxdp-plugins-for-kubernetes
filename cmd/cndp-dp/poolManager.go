/*
 * Copyright(c) 2022 Intel Corporation.
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/intel/cndp_device_plugin/internal/bpf"
	"github.com/intel/cndp_device_plugin/internal/cndp"
	"github.com/intel/cndp_device_plugin/internal/networking"
	logging "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

const (
	envVarDevs = "CNDP_DEVICES"
)

/*
PoolManager represents an manages the pool of devices.
Each PoolManager registers with Kubernetes as a different device type.
*/
type PoolManager struct {
	Name          string
	Mode          string
	Devices       map[string]*pluginapi.Device
	UpdateSignal  chan bool
	DpAPISocket   string
	DpAPIEndpoint string
	DpAPIServer   *grpc.Server
	ServerFactory cndp.ServerFactory
	BpfHandler    bpf.Handler
	Timeout       int
}

/*
Init is called it initialise the PoolManager.
*/
func (pm *PoolManager) Init(config *PoolConfig) error {
	pm.ServerFactory = cndp.NewServerFactory()
	pm.BpfHandler = bpf.NewHandler()
	netHandler := networking.NewHandler()

	if err := pm.registerWithKubelet(); err != nil {
		return err
	}
	logging.Infof("Pool "+devicePrefix+"/%s registered with Kubelet", pm.Name)

	if err := pm.startGRPC(); err != nil {
		return err
	}
	logging.Infof("Pool "+devicePrefix+"/%s started serving", pm.Name)

	for _, device := range config.Devices {
		logging.Debugf("Cycling state of device %s", device)
		if err := netHandler.CycleDevice(device); err != nil {
			logging.Errorf("Error cycling the state of device %s: %v", device, err)
			logging.Errorf("Device %s was not added to pool %s", device, pm.Name)
			continue
		}

		newdev := pluginapi.Device{ID: device, Health: pluginapi.Healthy}
		pm.Devices[device] = &newdev
	}

	if len(pm.Devices) > 0 {
		pm.UpdateSignal <- true
	}

	return nil
}

/*
Terminate is called it terminate the PoolManager.
*/
func (pm *PoolManager) Terminate() error {
	pm.stopGRPC()
	if err := pm.cleanup(); err != nil {
		logging.Infof("Cleanup error: %v", err)
	}
	logging.Infof(devicePrefix + "/" + pm.Name + " terminated")

	return nil
}

/*
ListAndWatch is part of the device plugin API.
Returns a stream list of Devices. Whenever a device state changes,
ListAndWatch should return the new list.
*/
func (pm *PoolManager) ListAndWatch(empty *pluginapi.Empty,
	stream pluginapi.DevicePlugin_ListAndWatchServer) error {

	logging.Debugf("Pool "+devicePrefix+"/%s ListAndWatch started", pm.Name)

	for {
		<-pm.UpdateSignal
		resp := new(pluginapi.ListAndWatchResponse)
		logging.Debugf("Pool "+devicePrefix+"/%s device list:", pm.Name)

		for _, dev := range pm.Devices {
			logging.Debugf("      " + dev.ID + ", " + dev.Health)
			resp.Devices = append(resp.Devices, dev)
		}

		if err := stream.Send(resp); err != nil {
			logging.Errorf("Failed to send stream to kubelet: %v", err)
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
	var err error
	var response *pluginapi.AllocateResponse

	logging.Debugf("New allocate request")
	if pm.Mode == "cndp" {
		response, err = pm.allocateCndp(ctx, rqt)
		if err != nil {
			logging.Errorf("Error during CNDP pod allocation: %v", err)
			return response, err
		}

		return response, nil
	}

	err = fmt.Errorf("Unrecognised device plugin mode")
	logging.Errorf("Allocate error: %v", err)

	return nil, err
}

func (pm *PoolManager) allocateCndp(ctx context.Context,
	rqt *pluginapi.AllocateRequest) (*pluginapi.AllocateResponse, error) {
	response := pluginapi.AllocateResponse{}

	logging.Infof("New CNDP allocate request. Creating new UDS server")
	cndpServer, udsPath, err := pm.ServerFactory.CreateServer(devicePrefix+"/"+pm.Name, pm.Timeout)
	if err != nil {
		logging.Errorf("Error Creating new UDS server: %v", err)
	}

	logging.Infof("UDS socket path: %s", udsPath)

	//loop each container
	for _, crqt := range rqt.ContainerRequests {
		cresp := new(pluginapi.ContainerAllocateResponse)
		envs := make(map[string]string)

		cresp.Mounts = append(cresp.Mounts, &pluginapi.Mount{
			HostPath:      udsPath,
			ContainerPath: "/tmp/cndp.sock",
			ReadOnly:      false,
		})

		//loop each device request per container
		for _, dev := range crqt.DevicesIDs {
			logging.Infof("Loading BPF program on device: %s", dev)
			fd, err := pm.BpfHandler.LoadBpfSendXskMap(dev)
			if err != nil {
				logging.Errorf("Error loading BPF Program on interface %s: %v", dev, err)
				return &response, err
			}
			logging.Infof("BPF program loaded on: %s File descriptor: %s", dev, strconv.Itoa(fd))

			cndpServer.AddDevice(dev, fd)
		}
		envs[envVarDevs] = strings.Join(crqt.DevicesIDs, " ")
		envsJSON, err := json.MarshalIndent(envs, "", " ")
		if err != nil {
			logging.Errorf("error while marshalling envs: %v", err)
		}
		logging.Infof("Setting environment variables: %s", string(envsJSON))

		cresp.Envs = envs

		response.ContainerResponses = append(response.ContainerResponses, cresp)
	}

	cndpServer.Start()

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

func (pm *PoolManager) registerWithKubelet() error {
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, pluginapi.KubeletSocket, grpc.WithInsecure(),
		grpc.WithContextDialer(func(ctx context.Context, addr string) (net.Conn, error) {
			return (&net.Dialer{}).DialContext(ctx, "unix", addr)
		}))
	if err != nil {
		return fmt.Errorf("Error connecting to Kubelet: %w", err)
	}
	defer conn.Close()

	client := pluginapi.NewRegistrationClient(conn)

	reqt := &pluginapi.RegisterRequest{
		Version:      pluginapi.Version,
		Endpoint:     pm.DpAPIEndpoint,
		ResourceName: devicePrefix + "/" + pm.Name,
	}

	_, err = client.Register(context.Background(), reqt)
	if err != nil {
		return fmt.Errorf("Error registering with Kubelet: %w", err)
	}

	return nil
}

func (pm *PoolManager) startGRPC() error {
	if err := pm.cleanup(); err != nil {
		return err
	}

	sock, err := net.Listen("unix", pm.DpAPISocket)
	if err != nil {
		return err
	}

	pm.DpAPIServer = grpc.NewServer([]grpc.ServerOption{}...)
	pluginapi.RegisterDevicePluginServer(pm.DpAPIServer, pm)
	go func() {
		if err := pm.DpAPIServer.Serve(sock); err != nil {
			logging.Errorf("API Server socket error: %v", err)
		}
	}()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	conn, err := grpc.DialContext(ctx, pm.DpAPISocket, grpc.WithInsecure(), grpc.WithBlock(),
		grpc.WithContextDialer(func(ctx context.Context, addr string) (net.Conn, error) {
			return (&net.Dialer{}).DialContext(ctx, "unix", addr)
		}),
	)
	if err != nil {
		logging.Errorf("Unable to establish test connection with %s gRPC server: %v", pm.Name, err)
		return err
	}
	conn.Close()
	logging.Debugf(devicePrefix+"/%s started serving on %s", pm.Name, pm.DpAPISocket)

	return nil
}

func (pm *PoolManager) stopGRPC() {
	if pm.DpAPIServer != nil {
		pm.DpAPIServer.Stop()
		pm.DpAPIServer = nil
	}
}

func (pm *PoolManager) cleanup() error {
	if err := os.Remove(pm.DpAPISocket); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
