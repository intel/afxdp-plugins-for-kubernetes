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

package deviceplugin

import (
	"context"
	"encoding/json"
	"github.com/intel/afxdp-plugins-for-kubernetes/internal/bpf"
	"github.com/intel/afxdp-plugins-for-kubernetes/internal/cndp"
	"github.com/stretchr/testify/assert"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
	"testing"
)

func TestAllocate(t *testing.T) {
	pm := &PoolManager{Mode: "cdq", Timeout: 30}
	pm.ServerFactory = cndp.NewFakeServerFactory()
	pm.BpfHandler = bpf.NewFakeHandler()

	testCases := []struct {
		name                  string
		containerRequests     []*pluginapi.ContainerAllocateRequest
		expContainerResponses []*pluginapi.ContainerAllocateResponse
	}{
		{
			name: "Single Container Single Device",
			containerRequests: []*pluginapi.ContainerAllocateRequest{
				{DevicesIDs: []string{"dev_1"}},
			},
			expContainerResponses: []*pluginapi.ContainerAllocateResponse{
				{
					Envs: map[string]string{"CNDP_DEVICES": "dev_1"},
					Mounts: []*pluginapi.Mount{
						{
							ContainerPath: "/tmp/cndp.sock",
							HostPath:      "/tmp/fake-socket.sock",
							ReadOnly:      false,
						},
					},
					Devices:     []*pluginapi.DeviceSpec{},
					Annotations: map[string]string{},
				},
			},
		},

		{
			name: "Single Container Multiple Devices",
			containerRequests: []*pluginapi.ContainerAllocateRequest{
				{DevicesIDs: []string{"dev_1", "dev_2", "dev_3"}},
			},
			expContainerResponses: []*pluginapi.ContainerAllocateResponse{
				{
					Envs: map[string]string{"CNDP_DEVICES": "dev_1 dev_2 dev_3"},
					Mounts: []*pluginapi.Mount{
						{
							ContainerPath: "/tmp/cndp.sock",
							HostPath:      "/tmp/fake-socket.sock",
							ReadOnly:      false,
						},
					},
					Devices:     []*pluginapi.DeviceSpec{},
					Annotations: map[string]string{},
				},
			},
		},

		{
			name: "Multiple Containers Single Device",
			containerRequests: []*pluginapi.ContainerAllocateRequest{
				{DevicesIDs: []string{"dev_1"}},
				{DevicesIDs: []string{"dev_2"}},
			},
			expContainerResponses: []*pluginapi.ContainerAllocateResponse{
				{
					Envs: map[string]string{"CNDP_DEVICES": "dev_1"},
					Mounts: []*pluginapi.Mount{
						{
							ContainerPath: "/tmp/cndp.sock",
							HostPath:      "/tmp/fake-socket.sock",
							ReadOnly:      false,
						},
					},
					Devices:     []*pluginapi.DeviceSpec{},
					Annotations: map[string]string{},
				},
				{
					Envs: map[string]string{"CNDP_DEVICES": "dev_2"},
					Mounts: []*pluginapi.Mount{
						{
							ContainerPath: "/tmp/cndp.sock",
							HostPath:      "/tmp/fake-socket.sock",
							ReadOnly:      false,
						},
					},
					Devices:     []*pluginapi.DeviceSpec{},
					Annotations: map[string]string{},
				},
			},
		},

		{
			name: "Multiple Containers Multiple Devices",
			containerRequests: []*pluginapi.ContainerAllocateRequest{
				{DevicesIDs: []string{"dev_1", "dev_2", "dev_3"}},
				{DevicesIDs: []string{"dev_4", "dev_5", "dev_6"}},
			},
			expContainerResponses: []*pluginapi.ContainerAllocateResponse{
				{
					Envs: map[string]string{"CNDP_DEVICES": "dev_1 dev_2 dev_3"},
					Mounts: []*pluginapi.Mount{
						{
							ContainerPath: "/tmp/cndp.sock",
							HostPath:      "/tmp/fake-socket.sock",
							ReadOnly:      false,
						},
					},
					Devices:     []*pluginapi.DeviceSpec{},
					Annotations: map[string]string{},
				},
				{
					Envs: map[string]string{"CNDP_DEVICES": "dev_4 dev_5 dev_6"},
					Mounts: []*pluginapi.Mount{
						{
							ContainerPath: "/tmp/cndp.sock",
							HostPath:      "/tmp/fake-socket.sock",
							ReadOnly:      false,
						},
					},
					Devices:     []*pluginapi.DeviceSpec{},
					Annotations: map[string]string{},
				},
			},
		},

		{
			name: "No Device",
			containerRequests: []*pluginapi.ContainerAllocateRequest{
				{DevicesIDs: []string{}},
			},
			expContainerResponses: []*pluginapi.ContainerAllocateResponse{
				{
					Envs: map[string]string{"CNDP_DEVICES": ""},
					Mounts: []*pluginapi.Mount{
						{
							ContainerPath: "/tmp/cndp.sock",
							HostPath:      "/tmp/fake-socket.sock",
							ReadOnly:      false,
						},
					},
					Devices:     []*pluginapi.DeviceSpec{},
					Annotations: map[string]string{},
				},
			},
		},
		{
			name:                  "No Container",
			containerRequests:     []*pluginapi.ContainerAllocateRequest{},
			expContainerResponses: []*pluginapi.ContainerAllocateResponse{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			allocateRequest := &pluginapi.AllocateRequest{
				ContainerRequests: tc.containerRequests,
			}

			expectedResponse := &pluginapi.AllocateResponse{
				ContainerResponses: tc.expContainerResponses,
			}

			response, err := pm.Allocate(context.Background(), allocateRequest)

			if err != nil {
				assert.FailNow(t, "Unexpected error during Allocate %v", err)
			}

			//TODO error
			expectedJSON, _ := json.Marshal(expectedResponse)
			responseJSON, _ := json.Marshal(response)

			assert.Equal(t, string(expectedJSON), string(responseJSON), "Unexpected AllocateResponse")

		})
	}
}
