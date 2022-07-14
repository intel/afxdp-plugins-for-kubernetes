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
	"errors"
	"github.com/intel/afxdp-plugins-for-kubernetes/internal/networking"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestGetConfig(t *testing.T) {
	testCases := []struct {
		name         string
		configFile   string
		expGetCfgErr error
		expBldPlsErr error
		expcfg       Config
		hostNetDev   map[string][]string
	}{
		{
			name: "get config : one pool two manually set devices",
			configFile: `{
							"mode": "cndp",
							"timeout": 30,
							"logLevel": "debug",
							"logFile": "/var/log/afxdp-k8s-plugins/file.log",
							"pools": [{
								"name": "pool1",
								"devices": ["dev1", "dev2"]
							}]
						}`,
			hostNetDev: map[string][]string{
				"i40e": []string{"dev1", "dev2", "dev3", "dev4"},
			},
			expcfg: Config{
				Pools: []*PoolConfig{
					{
						Name:    "pool1",
						Devices: []string{"dev1", "dev2"},
					},
				},
				Mode:                   "cndp",
				UdsTimeout:             30,
				RequireUnprivilegedBpf: false,
				LogDir:                 "/var/log/afxdp-k8s-plugins/",
				LogDirPermission:       0x1e4,
				LogFile:                "/var/log/afxdp-k8s-plugins/file.log",
				LogFilePermission:      0x1a4,
				LogLevel:               "debug",
				MinLinuxVersion:        "4.18.0",
			},
			expGetCfgErr: nil,
			expBldPlsErr: nil,
		},
		{
			name: "get config : one pool two manually set devices, the rest in pool 2",
			configFile: `{
							"mode": "cndp",
							"timeout": 30,
							"logLevel": "debug",
							"logFile": "/var/log/afxdp-k8s-plugins/file.log",
							"pools": [{
								"name": "pool1",
								"devices": ["dev1", "dev2"]
							},
							{
								"name": "pool2",
								"drivers": ["i40e"]
							}]
						}`,
			hostNetDev: map[string][]string{
				"i40e": []string{"dev1", "dev2", "dev3", "dev4"},
			},
			expcfg: Config{
				Pools: []*PoolConfig{
					{
						Name:    "pool1",
						Devices: []string{"dev1", "dev2"},
					},
					{
						Name:    "pool2",
						Devices: []string{"dev3", "dev4"},
						Drivers: []string{"i40e"},
					},
				},
				Mode:                   "cndp",
				UdsTimeout:             30,
				RequireUnprivilegedBpf: false,
				LogDir:                 "/var/log/afxdp-k8s-plugins/",
				LogDirPermission:       0x1e4,
				LogFile:                "/var/log/afxdp-k8s-plugins/file.log",
				LogFilePermission:      0x1a4,
				LogLevel:               "debug",
				MinLinuxVersion:        "4.18.0",
			},
			expGetCfgErr: nil,
			expBldPlsErr: nil,
		},
		{
			name: "get config : mix of devices and drivers",
			configFile: `{
							"mode": "cndp",
							"timeout": 30,
							"logLevel": "debug",
							"logFile": "/var/log/afxdp-k8s-plugins/file.log",
							"pools": [{
								"name": "pool1",
								"devices": ["dev5", "dev6"],
								"drivers": ["i40e"]
							},
							{
								"name": "pool2",
								"drivers": ["E810"]
							}]
						}`,
			hostNetDev: map[string][]string{
				"i40e": []string{"dev1", "dev2", "dev3", "dev4"},
				"E810": []string{"dev5", "dev6", "dev7", "dev8"},
			},
			expcfg: Config{
				Pools: []*PoolConfig{
					{
						Name:    "pool1",
						Devices: []string{"dev5", "dev6", "dev1", "dev2", "dev3", "dev4"},
						Drivers: []string{"i40e"},
					},
					{
						Name:    "pool2",
						Devices: []string{"dev7", "dev8"},
						Drivers: []string{"E810"},
					},
				},
				Mode:                   "cndp",
				UdsTimeout:             30,
				RequireUnprivilegedBpf: false,
				LogDir:                 "/var/log/afxdp-k8s-plugins/",
				LogDirPermission:       0x1e4,
				LogFile:                "/var/log/afxdp-k8s-plugins/file.log",
				LogFilePermission:      0x1a4,
				LogLevel:               "debug",
				MinLinuxVersion:        "4.18.0",
			},
			expGetCfgErr: nil,
			expBldPlsErr: nil,
		},
		{
			name: "get config : one_pool three_devices",
			configFile: `{
							"mode": "cndp",
							"timeout": 30,
							"logLevel": "debug",
							"logFile": "/var/log/afxdp-k8s-plugins/file.log",
							"pools": [{
								"name": "pool1",
								"devices": ["dev1", "dev2","dev3"]
							}]
						}`,
			hostNetDev: map[string][]string{
				"i40e": []string{"dev1", "dev2", "dev3", "dev4"},
			},
			expcfg: Config{
				Pools: []*PoolConfig{
					{
						Name:    "pool1",
						Devices: []string{"dev1", "dev2", "dev3"},
					},
				},
				Mode:                   "cndp",
				UdsTimeout:             30,
				RequireUnprivilegedBpf: false,
				LogDir:                 "/var/log/afxdp-k8s-plugins/",
				LogDirPermission:       0x1e4,
				LogFile:                "/var/log/afxdp-k8s-plugins/file.log",
				LogFilePermission:      0x1a4,
				LogLevel:               "debug",
				MinLinuxVersion:        "4.18.0",
			},
			expGetCfgErr: nil,
			expBldPlsErr: nil,
		},
		{
			name: "get config : two_pools four_devices",
			configFile: `{
								"mode": "cndp",
								"timeout": 30,
								"logLevel": "debug",
								"logFile": "/var/log/afxdp-k8s-plugins/file.log",
								"pools": [{
									"name": "pool1",
									"drivers": ["i40e"]
								}, {
									"name": "pool2",
									"drivers": ["E810"]
								}]
							}`,
			hostNetDev: map[string][]string{
				"i40e": []string{"dev1", "dev2"},
				"E810": []string{"dev3", "dev4"},
			},
			expcfg: Config{
				Pools: []*PoolConfig{
					{
						Name:    "pool1",
						Devices: []string{"dev1", "dev2"},
						Drivers: []string{"i40e"},
					},
					{
						Name:    "pool2",
						Devices: []string{"dev3", "dev4"},
						Drivers: []string{"E810"},
					},
				},
				Mode:                   "cndp",
				UdsTimeout:             30,
				RequireUnprivilegedBpf: false,
				LogDir:                 "/var/log/afxdp-k8s-plugins/",
				LogDirPermission:       0x1e4,
				LogFile:                "/var/log/afxdp-k8s-plugins/file.log",
				LogFilePermission:      0x1a4,
				LogLevel:               "debug",
				MinLinuxVersion:        "4.18.0",
			},
			expGetCfgErr: nil,
			expBldPlsErr: nil,
		},
		{
			name: "get config : two_pools six_devices",
			configFile: `{
							"mode": "cndp",
							"timeout": 30,	
							"logLevel": "debug",
							"logFile": "/var/log/afxdp-k8s-plugins/file.log",
							"pools": [{
								"name": "pool1",
								"drivers": ["i40e"]
							}, {
								"name": "pool2",
								"drivers": ["E810"]
							}]
						}`,
			hostNetDev: map[string][]string{
				"i40e": []string{"dev1", "dev2", "dev3"},
				"E810": []string{"dev4", "dev5", "dev6"},
			},
			expcfg: Config{
				Pools: []*PoolConfig{
					{
						Name:    "pool1",
						Devices: []string{"dev1", "dev2", "dev3"},
						Drivers: []string{"i40e"},
					},
					{
						Name:    "pool2",
						Devices: []string{"dev4", "dev5", "dev6"},
						Drivers: []string{"E810"},
					},
				},
				Mode:                   "cndp",
				UdsTimeout:             30,
				RequireUnprivilegedBpf: false,
				LogDir:                 "/var/log/afxdp-k8s-plugins/",
				LogDirPermission:       0x1e4,
				LogFile:                "/var/log/afxdp-k8s-plugins/file.log",
				LogFilePermission:      0x1a4,
				LogLevel:               "debug",
				MinLinuxVersion:        "4.18.0",
			},
			expGetCfgErr: nil,
			expBldPlsErr: nil,
		},

		{
			name: "get config : one pool with hyphenated driver",
			configFile: `{
							"mode": "cndp",
							"timeout": 30,
							"logLevel": "debug",
							"logFile": "/var/log/afxdp-k8s-plugins/file.log",
							"pools": [{
								"name": "pool1",
								"drivers": ["vfio-pci"]
							}]
						}`,
			hostNetDev: map[string][]string{
				"vfio-pci": []string{"dev1", "dev2", "dev3", "dev4"},
			},
			expcfg: Config{
				Pools: []*PoolConfig{
					{
						Name:    "pool1",
						Devices: []string{"dev1", "dev2", "dev3", "dev4"},
						Drivers: []string{"vfio-pci"},
					},
				},
				Mode:                   "cndp",
				UdsTimeout:             30,
				RequireUnprivilegedBpf: false,
				LogDir:                 "/var/log/afxdp-k8s-plugins/",
				LogDirPermission:       0x1e4,
				LogFile:                "/var/log/afxdp-k8s-plugins/file.log",
				LogFilePermission:      0x1a4,
				LogLevel:               "debug",
				MinLinuxVersion:        "4.18.0",
			},
			expGetCfgErr: nil,
			expBldPlsErr: nil,
		},

		{
			name: "get config : one pool with underscore driver",
			configFile: `{
							"mode": "cndp",
							"timeout": 30,
							"logLevel": "debug",
							"logFile": "/var/log/afxdp-k8s-plugins/file.log",
							"pools": [{
								"name": "pool1",
								"drivers": ["cool_driver"]
							}]
						}`,
			hostNetDev: map[string][]string{
				"cool_driver": []string{"dev1", "dev2", "dev3"},
			},
			expcfg: Config{
				Pools: []*PoolConfig{
					{
						Name:    "pool1",
						Devices: []string{"dev1", "dev2", "dev3"},
						Drivers: []string{"cool_driver"},
					},
				},
				Mode:                   "cndp",
				UdsTimeout:             30,
				RequireUnprivilegedBpf: false,
				LogDir:                 "/var/log/afxdp-k8s-plugins/",
				LogDirPermission:       0x1e4,
				LogFile:                "/var/log/afxdp-k8s-plugins/file.log",
				LogFilePermission:      0x1a4,
				LogLevel:               "debug",
				MinLinuxVersion:        "4.18.0",
			},
			expGetCfgErr: nil,
			expBldPlsErr: nil,
		},

		{
			name:         "load bad config : device field missing",
			configFile:   `{"mode": "cndp","timeout": 30,"logLevel": "debug","logFile": "/tmp/file.log","pools":[{"name":"pool1",:["dev1","dev2","dev3"],"drivers":["i40e"]}]}`,
			expGetCfgErr: errors.New("invalid character ':' looking for beginning of object key string"),
			expBldPlsErr: nil,
			expcfg:       Config{LogDir: "/var/log/afxdp-k8s-plugins/", LogDirPermission: 0x1e4, LogFilePermission: 0x1a4, MinLinuxVersion: "4.18.0"},
		},

		{
			name:         "load bad config : invalid JSON",
			configFile:   `{"mode": "cndp","timeout": 30,"logLevel": "debug","logFile": "/tmp/file.log","pools":[{"name":" "["dev1","dev2","dev3"],"drivers":["i40e"]}]}`,
			expGetCfgErr: errors.New("invalid character '[' after object key:value pair"),
			expBldPlsErr: nil,
			expcfg:       Config{LogDir: "/var/log/afxdp-k8s-plugins/", LogDirPermission: 0x1e4, LogFilePermission: 0x1a4, MinLinuxVersion: "4.18.0"},
		},

		{
			name:         "load bad config : no pools",
			configFile:   `{"mode": "cndp","timeout": 30,"logLevel": "debug","logFile": "/tmp/file.log", :[{"name: ["dev1","dev2","dev3"],"drivers":["i40e"]}]}`,
			expGetCfgErr: errors.New("invalid character ':' looking for beginning of object key string"),
			expBldPlsErr: nil,
			expcfg:       Config{LogDir: "/var/log/afxdp-k8s-plugins/", LogDirPermission: 0x1e4, LogFilePermission: 0x1a4, MinLinuxVersion: "4.18.0"},
		},

		{
			name:         "load bad config : empty pool ",
			configFile:   ` `,
			expGetCfgErr: errors.New("unexpected end of JSON input"),
			expBldPlsErr: nil,
			expcfg:       Config{LogDir: "/var/log/afxdp-k8s-plugins/", LogDirPermission: 0x1e4, LogFilePermission: 0x1a4, MinLinuxVersion: "4.18.0"},
		},

		{
			name:         "load bad config : invalid character ",
			configFile:   "?",
			expGetCfgErr: errors.New("invalid character '?' looking for beginning of value"),
			expBldPlsErr: nil,
			expcfg:       Config{LogDir: "/var/log/afxdp-k8s-plugins/", LogDirPermission: 0x1e4, LogFilePermission: 0x1a4, MinLinuxVersion: "4.18.0"},
		},

		{
			name: "load bad config : bad driver name",
			configFile: `{
								"mode": "cndp",
								"timeout": 30,
								"logLevel": "debug",
								"logFile": "/var/log/afxdp-k8s-plugins/file.log",
								"pools": [
									{
									"name": "pool1",
									"drivers": [
										"ice+"
									]
									}
								]
							}`,
			expGetCfgErr: errors.New("driver names must only contain letters, numbers and selected symbols"),
			expBldPlsErr: nil,
			expcfg: Config{
				Pools: []*PoolConfig{
					{
						Name:    "pool1",
						Drivers: []string{"ice+"},
					},
				},
				Mode:                   "cndp",
				UdsTimeout:             30,
				RequireUnprivilegedBpf: false,
				LogDir:                 "/var/log/afxdp-k8s-plugins/",
				LogDirPermission:       0x1e4,
				LogFile:                "/var/log/afxdp-k8s-plugins/file.log",
				LogFilePermission:      0x1a4,
				LogLevel:               "debug",
				MinLinuxVersion:        "4.18.0",
			},
		},

		{
			name: "load bad config : bad device name",
			configFile: `{
								"mode": "cndp",
								"timeout": 30,
								"logLevel": "debug",
								"logFile": "/var/log/afxdp-k8s-plugins/file.log",
								"pools": [
									{
									"name": "pool1",
									"devices": [
										"dev1^"
									]
									}
								]
							}`,
			expGetCfgErr: errors.New("device names must only contain letters, numbers and selected symbols"),
			expBldPlsErr: nil,
			expcfg: Config{
				Pools: []*PoolConfig{
					{
						Name:    "pool1",
						Devices: []string{"dev1^"},
					},
				},
				Mode:                   "cndp",
				UdsTimeout:             30,
				RequireUnprivilegedBpf: false,
				LogDir:                 "/var/log/afxdp-k8s-plugins/",
				LogDirPermission:       0x1e4,
				LogFile:                "/var/log/afxdp-k8s-plugins/file.log",
				LogFilePermission:      0x1a4,
				LogLevel:               "debug",
				MinLinuxVersion:        "4.18.0",
			},
		},
	}

	for _, tc := range testCases {

		t.Run(tc.name, func(t *testing.T) {
			assignedInfs = nil //clear assignedInfs at beginning of each run
			fakeNetHandler := networking.NewFakeHandler()
			fakeNetHandler.SetHostDevices(tc.hostNetDev)
			content := []byte(tc.configFile)
			dir, dirErr := ioutil.TempDir("/tmp", "test-cndp-")
			require.NoError(t, dirErr, "Can't create temporary directory")
			testDir := filepath.Join(dir, "tmpfile")
			err := ioutil.WriteFile(testDir, content, 0666)
			require.NoError(t, err, "Can't create temporary file")

			defer os.RemoveAll(dir)

			cfg, err := GetConfig(testDir, fakeNetHandler)
			if err == nil {
				assert.Equal(t, tc.expGetCfgErr, err, "Error was expected")
			} else {
				require.Error(t, tc.expGetCfgErr, "Unexpected error returned")
				assert.Contains(t, err.Error(), tc.expGetCfgErr.Error(), "Unexpected error returned")
			}

			err = cfg.BuildPools()
			if err == nil {
				assert.Equal(t, tc.expBldPlsErr, err, "Error was expected")
			} else {
				require.Error(t, tc.expBldPlsErr, "Unexpected error returned")
				assert.Contains(t, err.Error(), tc.expBldPlsErr.Error(), "Unexpected error returned")
			}

			assert.Equal(t, tc.expcfg, cfg, "Error was expected: configs do not match")
		})
	}
}
