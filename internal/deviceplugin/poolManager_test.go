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
	"strings"
	"testing"

	"github.com/intel/afxdp-plugins-for-kubernetes/constants"
	"github.com/intel/afxdp-plugins-for-kubernetes/internal/bpf"
	"github.com/intel/afxdp-plugins-for-kubernetes/internal/networking"
	"github.com/intel/afxdp-plugins-for-kubernetes/internal/udsserver"
	"github.com/stretchr/testify/assert"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

func TestAllocate(t *testing.T) {
	netHandler := networking.NewFakeHandler()

	config := PoolConfig{
		Name: "myPool",
		Mode: "primary",
		Devices: map[string]*networking.Device{
			"dev_1": networking.CreateTestDevice("dev_1", "primary", "ice", "0000:81:00.1", "68:05:ca:2d:e9:01", netHandler),
			"dev_2": networking.CreateTestDevice("dev_2", "primary", "ice", "0000:81:00.2", "68:05:ca:2d:e9:02", netHandler),
			"dev_3": networking.CreateTestDevice("dev_3", "primary", "ice", "0000:81:00.3", "68:05:ca:2d:e9:03", netHandler),
			"dev_4": networking.CreateTestDevice("dev_4", "primary", "ice", "0000:81:00.4", "68:05:ca:2d:e9:04", netHandler),
			"dev_5": networking.CreateTestDevice("dev_5", "primary", "ice", "0000:81:00.5", "68:05:ca:2d:e9:05", netHandler),
			"dev_6": networking.CreateTestDevice("dev_6", "primary", "ice", "0000:81:00.6", "68:05:ca:2d:e9:06", netHandler),
			"dev_7": networking.CreateTestDevice("dev_7", "primary", "ice", "0000:81:00.7", "68:05:ca:2d:e9:07", netHandler),
			"dev_8": networking.CreateTestDevice("dev_8", "primary", "ice", "0000:81:00.8", "68:05:ca:2d:e9:08", netHandler),
			"dev_9": networking.CreateTestDevice("dev_9", "primary", "ice", "0000:81:00.9", "68:05:ca:2d:e9:09", netHandler),
		},
		UdsServerDisable:        false,
		BpfMapPinningEnable:     false,
		UdsTimeout:              0,
		UdsFuzz:                 false,
		RequiresUnprivilegedBpf: false,
		UID:                     1500,
	}

	pm := NewPoolManager(config)
	pm.ServerFactory = udsserver.NewFakeServerFactory()
	pm.BpfHandler = bpf.NewFakeHandler()

	envVar := constants.Devices.EnvVarList + strings.ToUpper(pm.Name)

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
					Envs: map[string]string{envVar: "dev_1"},
					Mounts: []*pluginapi.Mount{
						{
							ContainerPath: constants.Uds.PodPath + "dev_1" + constants.Uds.SockName,
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
					Envs: map[string]string{envVar: "dev_1 dev_2 dev_3"},
					Mounts: []*pluginapi.Mount{
						{
							ContainerPath: constants.Uds.PodPath + "dev_1" + constants.Uds.SockName,
							HostPath:      "/tmp/fake-socket.sock",
							ReadOnly:      false,
						},
						{
							ContainerPath: constants.Uds.PodPath + "dev_2" + constants.Uds.SockName,
							HostPath:      "/tmp/fake-socket.sock",
							ReadOnly:      false,
						},
						{
							ContainerPath: constants.Uds.PodPath + "dev_3" + constants.Uds.SockName,
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
					Envs: map[string]string{envVar: "dev_1"},
					Mounts: []*pluginapi.Mount{
						{
							ContainerPath: constants.Uds.PodPath + "dev_1" + constants.Uds.SockName,
							HostPath:      "/tmp/fake-socket.sock",
							ReadOnly:      false,
						},
					},
					Devices:     []*pluginapi.DeviceSpec{},
					Annotations: map[string]string{},
				},
				{
					Envs: map[string]string{envVar: "dev_2"},
					Mounts: []*pluginapi.Mount{
						{
							ContainerPath: constants.Uds.PodPath + "dev_2" + constants.Uds.SockName,
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
					Envs: map[string]string{envVar: "dev_1 dev_2 dev_3"},
					Mounts: []*pluginapi.Mount{
						{
							ContainerPath: constants.Uds.PodPath + "dev_1" + constants.Uds.SockName,
							HostPath:      "/tmp/fake-socket.sock",
							ReadOnly:      false,
						},
						{
							ContainerPath: constants.Uds.PodPath + "dev_2" + constants.Uds.SockName,
							HostPath:      "/tmp/fake-socket.sock",
							ReadOnly:      false,
						},
						{
							ContainerPath: constants.Uds.PodPath + "dev_3" + constants.Uds.SockName,
							HostPath:      "/tmp/fake-socket.sock",
							ReadOnly:      false,
						},
					},
					Devices:     []*pluginapi.DeviceSpec{},
					Annotations: map[string]string{},
				},
				{
					Envs: map[string]string{envVar: "dev_4 dev_5 dev_6"},
					Mounts: []*pluginapi.Mount{
						{
							ContainerPath: constants.Uds.PodPath + "dev_4" + constants.Uds.SockName,
							HostPath:      "/tmp/fake-socket.sock",
							ReadOnly:      false,
						},
						{
							ContainerPath: constants.Uds.PodPath + "dev_5" + constants.Uds.SockName,
							HostPath:      "/tmp/fake-socket.sock",
							ReadOnly:      false,
						},
						{
							ContainerPath: constants.Uds.PodPath + "dev_6" + constants.Uds.SockName,
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
					Envs:        map[string]string{envVar: ""},
					Mounts:      []*pluginapi.Mount{},
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
