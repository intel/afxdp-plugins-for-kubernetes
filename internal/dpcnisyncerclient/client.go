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
package dpcnisyncerclient

import (
	"context"
	"net"

	"github.com/intel/afxdp-plugins-for-kubernetes/constants"
	pb "github.com/intel/afxdp-plugins-for-kubernetes/internal/dpcnisyncer"
	logging "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

const (
	_proto = "unix"
)

var (
	sock = pluginapi.DevicePluginPath + constants.Plugins.DevicePlugin.DevicePrefix + "-" + "syncer.sock"
)

func DeleteNetDev(name string) error {
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, sock, grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(ctx context.Context, addr string) (net.Conn, error) {
			return (&net.Dialer{}).DialContext(ctx, _proto, addr)
		}))
	if err != nil {
		logging.Errorf("error connecting to Server")
		return err
	}
	defer conn.Close()

	c := pb.NewNetDevClient(conn)
	r, err := c.DelNetDev(ctx, &pb.DeleteNetDevReq{Name: name})
	if err != nil || r.Ret == -1 {
		logging.Errorf("error deleting netdev resources for netdev %s", name)
		return err
	}
	logging.Infof("Server response:%v", r)

	return nil
}
