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

	// Replace with the actual proto package
	pb "github.com/intel/afxdp-plugins-for-kubernetes/internal/dpcnisyncer"
	logging "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	protocol = "unix"
	sockAddr = "/tmp/afxdp_dp/syncer.sock"
)

type SyncerServer struct {
	pb.UnimplementedNetDevServer
}

func (s *SyncerServer) DelNetDev(ctx context.Context, in *pb.DeleteNetDevReq) (*pb.DeleteNetDevResp, error) {
	netDevName := in.GetName()
	// Delete the network interface called netdev using system calls or appropriate libraries
	logging.Infof("Network interface %s deleted", netDevName)
	return &pb.DeleteNetDevResp{}, nil
}

func NewSyncerServer() (*grpc.Server, error) {
	if _, err := os.Stat(sockAddr); !os.IsNotExist(err) {
		if err := os.RemoveAll(sockAddr); err != nil {
			logging.Errorf("sockAddr %s does not exist", sockAddr)
			return nil, err
		}
	}

	lis, err := net.Listen(protocol, sockAddr)
	if err != nil {
		logging.Errorf("Could not listen to %s", sockAddr)
		return nil, err
	}

	s := grpc.NewServer()
	pb.RegisterNetDevServer(s, &SyncerServer{})
	go func() {
		if err := s.Serve(lis); err != nil {
			logging.Error("Could not RegisterNetDevServer: %v", err)
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
	return s, nil
}

func Cleanup() error {
	if err := os.Remove(sockAddr); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// func StopGRPCSyncer() {
// 	//TODO
// }
