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

package resourcesapi

import (
	"net"
	"time"

	logging "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	api "k8s.io/kubelet/pkg/apis/podresources/v1"
)

const (
	podResSockDir  = "/var/lib/kubelet/pod-resources"
	podResSockPath = podResSockDir + "/kubelet.sock"
	grpcTimeout    = 5 * time.Second
)

/*
Handler is the device plugins interface to the K8s pod resources API.
The interface exists for testing purposes, allowing unit tests to test
against a fake API.
*/
type Handler interface {
	GetPodResources() (map[string]api.PodResources, error)
}

/*
handler implements the Handler interface.
*/
type handler struct{}

/*
NewHandler returns an implementation of the Handler interface.
*/
func NewHandler() Handler {
	return &handler{}
}

/*
GetPodResources calls the pod resources api and returns a map of pods and associated devices
*/
func (r *handler) GetPodResources() (map[string]api.PodResources, error) {
	podResourceMap := make(map[string]api.PodResources)

	resp, err := getPodResources(podResSockPath)
	if err != nil {
		logging.Errorf("Error Getting pod resources: %v", err)
		return podResourceMap, err
	}

	for _, pod := range resp.GetPodResources() {
		podResourceMap[pod.GetName()] = *pod
	}

	return podResourceMap, nil
}

func getPodResources(socket string) (*api.ListPodResourcesResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), grpcTimeout)
	defer cancel()

	logging.Debugf("Opening Pod Resource API connection")
	conn, err := grpc.DialContext(ctx, socket, grpc.WithInsecure(), grpc.WithBlock(),
		grpc.WithContextDialer(func(ctx context.Context, addr string) (net.Conn, error) {
			return (&net.Dialer{}).DialContext(ctx, "unix", addr)
		}),
	)
	if err != nil {
		logging.Errorf("Error connecting to Pod Resource API: %v", err)
		conn.Close()
		return nil, err
	}
	defer func() {
		logging.Debugf("Closing Pod Resource API connection")
		conn.Close()
	}()

	logging.Debugf("Requesting pod resource list")
	client := api.NewPodResourcesListerClient(conn)

	resp, err := client.List(ctx, &api.ListPodResourcesRequest{})
	if err != nil {
		logging.Errorf("Error getting Pod Resource list: %v", err)
		return nil, err
	}

	return resp, nil
}
