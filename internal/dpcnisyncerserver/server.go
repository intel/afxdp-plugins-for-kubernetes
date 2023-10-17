/*
 * Copyright(c) 2023 Intel Corporation.
 * Copyright(c) Red Hat Inc.
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
package dpcnisyncerserver

import (
	"context"
	"net"
	"os"
	"time"

	"github.com/intel/afxdp-plugins-for-kubernetes/constants"
	"github.com/intel/afxdp-plugins-for-kubernetes/internal/bpf"
	pb "github.com/intel/afxdp-plugins-for-kubernetes/internal/dpcnisyncer"
	"github.com/pkg/errors"
	logging "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

const (
	protocol = "unix"
)

var (
	sockAddr = pluginapi.DevicePluginPath + constants.Plugins.DevicePlugin.DevicePrefix + "-" + "syncer.sock"
)

type SyncerServer struct {
	pb.UnimplementedNetDevServer
	mapManagers     []bpf.PoolBpfMapManager
	grpcServer      *grpc.Server
	BpfMapPinEnable bool
}

func (s *SyncerServer) RegisterMapManager(b bpf.PoolBpfMapManager) {

	if s.mapManagers != nil {
		for _, v := range s.mapManagers {
			if v.Manager.GetName() == b.Manager.GetName() {
				logging.Infof("%s is already registered", b.Manager.GetName())
				return
			}
		}
	}

	s.mapManagers = append(s.mapManagers, b)
}

func (s *SyncerServer) DelNetDev(ctx context.Context, in *pb.DeleteNetDevReq) (*pb.DeleteNetDevResp, error) {

	if s.BpfMapPinEnable {
		netDevName := in.GetName()

		logging.Infof("Looking up Map Manager for %s", netDevName)
		found := false
		var pm bpf.PoolBpfMapManager
		for _, mm := range s.mapManagers {
			_, err := mm.Manager.GetBPFFS(netDevName)
			if err == nil {
				found = true
				pm = mm
				break
			}
		}

		if !found {
			logging.Errorf("Could NOT find the map manager for device %s", netDevName)
			return &pb.DeleteNetDevResp{Ret: -1}, errors.New("Could NOT find the map manager for device")
		}

		logging.Infof("Map Manager found, deleting BPFFS for %s", netDevName)
		err := pm.Manager.DeleteBPFFS(netDevName)
		if err != nil {
			logging.Errorf("Could NOT delete BPFFS for %s", netDevName)
			return &pb.DeleteNetDevResp{Ret: -1}, errors.Wrapf(err, "Could NOT delete BPFFS for %s: %v", netDevName, err.Error())
		}

		logging.Infof("Network interface %s deleted", netDevName)
		return &pb.DeleteNetDevResp{Ret: 0}, nil
	}

	return &pb.DeleteNetDevResp{Ret: -1}, errors.New("BPF Map pinning is not enabled")
}

func (s *SyncerServer) StopGRPCSyncer() {
	if s.grpcServer != nil {
		s.grpcServer.Stop()
		s.grpcServer = nil
	}
	s.cleanup()
}

func NewSyncerServer() (*SyncerServer, error) {
	if _, err := os.Stat(sockAddr); !os.IsNotExist(err) {
		if err := os.RemoveAll(sockAddr); err != nil {
			logging.Errorf("sockAddr %s does not exist", sockAddr)
			return nil, err
		}
	}

	server := &SyncerServer{
		grpcServer:      grpc.NewServer(),
		BpfMapPinEnable: false,
	}

	lis, err := net.Listen(protocol, sockAddr)
	if err != nil {
		logging.Errorf("Could not listen to %s", sockAddr)
		return nil, err
	}

	pb.RegisterNetDevServer(server.grpcServer, server)
	go func() {
		if err := server.grpcServer.Serve(lis); err != nil {
			logging.Errorf("Could not RegisterNetDevServer: %v", err)
		}
	}()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	conn, err := grpc.DialContext(ctx, sockAddr, grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(), grpc.WithContextDialer(func(ctx context.Context, addr string) (net.Conn, error) {
			return (&net.Dialer{}).DialContext(ctx, "unix", addr)
		}),
	)
	if err != nil {
		logging.Errorf("Unable to establish test connection with gRPC server: %v", err)
		return nil, err
	}
	conn.Close()
	logging.Debugf("NewSyncerServer up and Running")
	return server, nil
}

func (s *SyncerServer) cleanup() error {
	if err := os.Remove(sockAddr); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
