/*
 * Copyright(c) 2022 Intel Corporation.
 * Copyright(c) Red Hat Inc.
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *	 http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package deviceplugin

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/intel/afxdp-plugins-for-kubernetes/constants"
	"github.com/intel/afxdp-plugins-for-kubernetes/internal/bpf"
	dpcnisyncer "github.com/intel/afxdp-plugins-for-kubernetes/internal/dpcnisyncerserver"
	"github.com/intel/afxdp-plugins-for-kubernetes/internal/networking"
	"github.com/intel/afxdp-plugins-for-kubernetes/internal/tools"
	"github.com/intel/afxdp-plugins-for-kubernetes/internal/udsserver"
	"github.com/pkg/errors"
	logging "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

/*
PoolManager represents an manages the pool of devices.
Each PoolManager registers with Kubernetes as a different device type.
*/
type PoolManager struct {
	Name                string
	Mode                string
	Devices             map[string]*networking.Device
	UpdateSignal        chan bool
	DpAPISocket         string
	DpAPIEndpoint       string
	UdsServerDisable    bool
	BpfMapPinningEnable bool
	UdsTimeout          int
	DevicePrefix        string
	UdsFuzz             bool
	UID                 string
	EthtoolFilters      []string
	DpAPIServer         *grpc.Server
	ServerFactory       udsserver.ServerFactory
	MapManagerFactory   bpf.MapManagerFactory
	BpfHandler          bpf.Handler
	NetHandler          networking.Handler
	DpCniSyncerServer   *grpc.Server
	SyncerActive        bool
	Pbm                 bpf.PoolBpfMapManager
}

func NewPoolManager(config PoolConfig) PoolManager {
	return PoolManager{
		Name:                config.Name,
		Mode:                config.Mode,
		Devices:             config.Devices,
		UpdateSignal:        make(chan bool),
		DpAPISocket:         pluginapi.DevicePluginPath + constants.Plugins.DevicePlugin.DevicePrefix + "-" + config.Name + ".sock",
		DpAPIEndpoint:       constants.Plugins.DevicePlugin.DevicePrefix + "-" + config.Name + ".sock",
		UdsServerDisable:    config.UdsServerDisable,
		BpfMapPinningEnable: config.BpfMapPinningEnable,
		UdsTimeout:          config.UdsTimeout,
		DevicePrefix:        constants.Plugins.DevicePlugin.DevicePrefix,
		UdsFuzz:             config.UdsFuzz,
		UID:                 strconv.Itoa(config.UID),
		EthtoolFilters:      config.EthtoolCmds,
	}
}

/*
Init is called it initialise the PoolManager.
*/
func (pm *PoolManager) Init(config PoolConfig) error {
	pm.ServerFactory = udsserver.NewServerFactory()
	pm.MapManagerFactory = bpf.NewMapMangerFactory()
	pm.BpfHandler = bpf.NewHandler()
	pm.NetHandler = networking.NewHandler()

	if err := pm.startGRPC(); err != nil {
		return err
	}
	logging.Infof("Pool "+pm.DevicePrefix+"/%s started serving", pm.Name)

	if err := pm.registerWithKubelet(); err != nil {
		return err
	}
	logging.Infof("Pool "+pm.DevicePrefix+"/%s registered with Kubelet", pm.Name)

	if pm.BpfMapPinningEnable {
		var err error

		logging.Infof("Creating new BPF Map manager %s %s", pm.DevicePrefix+"-maps/", pm.UID)
		pm.Pbm.Manager, pm.Pbm.Path, err = pm.MapManagerFactory.CreateMapManager(pm.DevicePrefix+"-maps/", pm.UID)
		if err != nil {
			logging.Errorf("Error new BPF Map manager: %v", err)
			return err
		}

		logging.Debug("Creating new DP<=>CNI gRPC Syncer")
		if err := pm.startGRPCSyncer(pm.Pbm); err != nil {
			logging.Error("Pool "+pm.DevicePrefix+"/%s syncer error %v", err)
			return err
		}
		logging.Infof("Pool "+pm.DevicePrefix+"/%s syncer started serving", pm.Name)
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
	pm.stopGRPCSyncer()
	if err := pm.cleanup(); err != nil {
		logging.Infof("Cleanup error: %v", err)
	}
	logging.Infof(pm.DevicePrefix + "/" + pm.Name + " terminated")

	return nil
}

/*
ListAndWatch is part of the device plugin API.
Returns a stream list of Devices. Whenever a device state changes,
ListAndWatch should return the new list.
*/
func (pm *PoolManager) ListAndWatch(empty *pluginapi.Empty,
	stream pluginapi.DevicePlugin_ListAndWatchServer) error {

	logging.Debugf("Pool "+pm.DevicePrefix+"/%s ListAndWatch started", pm.Name)

	for {
		<-pm.UpdateSignal
		resp := new(pluginapi.ListAndWatchResponse)

		for devName := range pm.Devices {
			resp.Devices = append(resp.Devices, &pluginapi.Device{ID: devName, Health: pluginapi.Healthy})
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
	response := pluginapi.AllocateResponse{}
	var udsServer udsserver.Server
	var udsPath string
	var err error

	logging.Debugf("New allocate request on pool %s", pm.Name)

	if !pm.UdsServerDisable {
		logging.Infof("Creating new UDS server")
		udsServer, udsPath, err = pm.ServerFactory.CreateServer(pm.DevicePrefix+"/"+pm.Name, pm.UID, pm.UdsTimeout, pm.UdsFuzz)
		if err != nil {
			logging.Errorf("Error Creating new UDS server: %v", err)
			return &response, err
		}
	}

	//loop each container request
	for _, crqt := range rqt.ContainerRequests {
		cresp := new(pluginapi.ContainerAllocateResponse)
		envs := make(map[string]string)

		if !pm.UdsServerDisable {
			cresp.Mounts = append(cresp.Mounts, &pluginapi.Mount{
				HostPath:      udsPath,
				ContainerPath: constants.Uds.PodPath,
				ReadOnly:      false,
			})
		}

		//loop each device request per container
		for _, devName := range crqt.DevicesIDs {
			device := pm.Devices[devName]
			pretty, _ := tools.PrettyString(device.Public())
			logging.Debugf("Device: %s", pretty)

			if device.Mode() != pm.Mode {
				err := fmt.Errorf("pool mode %s does not match device mode %s", pm.Mode, device.Mode())
				logging.Errorf("%v", err)
				return &response, err
			}

			switch pm.Mode {
			case "primary":
				logging.Debugf("Primary mode")
			case "cdq":
				if err := device.ActivateCdqSubfunction(); err != nil {
					logging.Errorf("Error creating CDQ subfunction: %v", err)
					return &response, err
				}
			default:
				err := fmt.Errorf("unsupported pool mode: %s", pm.Mode)
				logging.Errorf("%v", err)
				return &response, err
			}

			logging.Debugf("Cycling state of device %s", device.Name())
			if err := device.Cycle(); err != nil {
				logging.Errorf("Error cycling the state of device %s: %v", device.Name(), err)
				continue
			}

			if !pm.UdsServerDisable {
				logging.Infof("Loading BPF program on device: %s", device.Name())
				fd, err := pm.BpfHandler.LoadBpfSendXskMap(device.Name())
				if err != nil {
					logging.Errorf("Error loading BPF Program on interface %s: %v", device.Name(), err)
					return &response, err
				}
				logging.Infof("BPF program loaded on: %s File descriptor: %s", device.Name(), strconv.Itoa(fd))
				udsServer.AddDevice(device.Name(), fd)
			}

			if pm.BpfMapPinningEnable {
				logging.Infof("Loading BPF program on device: %s and pinning the map", device.Name())
				pinPath, err := pm.Pbm.Manager.CreateBPFFS(device.Name(), pm.Pbm.Path)
				if err != nil {
					logging.Errorf("Error Creating the BPFFS: %v", err)
					return &response, err
				}

				err = pm.BpfHandler.LoadBpfPinXskMap(device.Name(), pinPath)
				if err != nil {
					logging.Errorf("Error loading BPF Program on interface %s and pinning the map: %v", device.Name(), err)
					return &response, err
				}

				pm.Pbm.Manager.AddMap(device.Name(), pinPath)

				//FULL PATH WILL INCLUDE THE XSKMAP...
				fullPath := pinPath + constants.Bpf.Xsk_map
				logging.Debugf("mapping %s to %s", fullPath, constants.Bpf.BpfMapPodPath)
				cresp.Mounts = append(cresp.Mounts, &pluginapi.Mount{
					HostPath:      fullPath,
					ContainerPath: constants.Bpf.BpfMapPodPath,
					ReadOnly:      false,
				})
			}

			if pm.EthtoolFilters != nil {
				device.SetEthtoolFilter(pm.EthtoolFilters)
				if err = pm.NetHandler.WriteDeviceFile(device, constants.DeviceFile.Directory+constants.DeviceFile.Name); err != nil {
					logging.Debugf("Error writing to device file %v", err)
					return &response, err
				}
			}
		}

		envs[constants.Devices.EnvVarList] = strings.Join(crqt.DevicesIDs, " ")
		envsPrint, err := tools.PrettyString(envs)
		if err != nil {
			logging.Errorf("Error printing container environment variables: %v", err)
		} else {
			logging.Debugf("Container environment variables: %s", envsPrint)
		}
		cresp.Envs = envs
		response.ContainerResponses = append(response.ContainerResponses, cresp)

	}

	if !pm.UdsServerDisable {
		udsServer.Start()
	}

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
	conn, err := grpc.DialContext(ctx, pluginapi.KubeletSocket, grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(ctx context.Context, addr string) (net.Conn, error) {
			return (&net.Dialer{}).DialContext(ctx, "unix", addr)
		}))
	if err != nil {
		return fmt.Errorf("error connecting to Kubelet: %w", err)
	}
	defer conn.Close()

	client := pluginapi.NewRegistrationClient(conn)

	reqt := &pluginapi.RegisterRequest{
		Version:      pluginapi.Version,
		Endpoint:     pm.DpAPIEndpoint,
		ResourceName: pm.DevicePrefix + "/" + pm.Name,
	}

	_, err = client.Register(context.Background(), reqt)
	if err != nil {
		return fmt.Errorf("error registering with Kubelet: %w", err)
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
	conn, err := grpc.DialContext(ctx, pm.DpAPISocket, grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(), grpc.WithContextDialer(func(ctx context.Context, addr string) (net.Conn, error) {
			return (&net.Dialer{}).DialContext(ctx, "unix", addr)
		}),
	)
	if err != nil {
		logging.Errorf("Unable to establish test connection with %s gRPC server: %v", pm.Name, err)
		return err
	}
	conn.Close()
	logging.Debugf(pm.DevicePrefix+"/%s started serving on %s", pm.Name, pm.DpAPISocket)

	return nil
}

func (pm *PoolManager) startGRPCSyncer(mm bpf.PoolBpfMapManager) error {

	var err error

	pm.DpCniSyncerServer, err = dpcnisyncer.NewSyncerServer(mm)
	if err != nil {
		return errors.Wrap(err, "Error creating the DpCniSyncerServer")
	}
	logging.Debugf(pm.DevicePrefix + " Syncer started serving")

	return nil
}

func (pm *PoolManager) stopGRPCSyncer() {
	if pm.DpCniSyncerServer != nil {
		pm.DpCniSyncerServer.Stop()
		pm.DpCniSyncerServer = nil
	}
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

	for dev := range pm.Pbm.Manager.GetMaps() {
		logging.Debugf("Deleting BPFFS for dev %s", dev)
		err := pm.Pbm.Manager.DeleteBPFFS(dev)
		if err != nil {
			return errors.Wrapf(err, "Could NOT delete BPFFS for %s", dev, err.Error())
		}
	}
	return nil
}
