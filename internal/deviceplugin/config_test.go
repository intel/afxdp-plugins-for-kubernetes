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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestReadConfigFile(t *testing.T) {
	testCases := []struct {
		name       string
		configFile string
		expErr     error
		//		expcfg     Config
		hostNetDev map[string][]string
	}{
		/*********************** Device Validation ***********************/
		{
			name: "device name must only use certain characters 1",
			configFile: `{
							"pools":[
								{
									"name":"testPool",
									"mode":"cdq",
									"devices":[
										{
											"name":"dev1"
										},
										{
											"name":"dev~2"
										}
									]
								}
							]
						}`,
			expErr: errors.New(deviceValidNameError),
		},
		{
			name: "device name must only use certain characters 2",
			configFile: `{
							"pools":[
								{
									"name":"testPool",
									"mode":"cdq",
									"devices":[
										{
											"name":"dev1"
										},
										{
											"name":"dev 2"
										}
									]
								}
							]
						}`,
			expErr: errors.New(deviceValidNameError),
		},
		{
			name: "device name must not be too long",
			configFile: `{
							"pools":[
								{
									"name":"testPool",
									"mode":"cdq",
									"devices":[
										{
											"name":"dev1"
										},
										{
											"name":"abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyz"
										}
									]
								}
							]
						}`,
			expErr: errors.New(deviceNameLengthError),
		},
		{
			name: "device mac must be valid 1",
			configFile: `{
							"pools":[
								{
									"name":"testPool",
									"mode":"cdq",
									"devices":[
										{
											"name":"dev1"
										},
										{
											"mac":"dev2"
										}
									]
								}
							]
						}`,
			expErr: errors.New(deviceValidMacError),
		},
		{
			name: "device mac must be valid 2",
			configFile: `{
							"pools":[
								{
									"name":"testPool",
									"mode":"cdq",
									"devices":[
										{
											"name":"dev1"
										},
										{
											"mac":"98:03:9b:6a:b4"
										}
									]
								}
							]
						}`,
			expErr: errors.New(deviceValidMacError),
		},
		{
			name: "device mac must be valid 3",
			configFile: `{
							"pools":[
								{
									"name":"testPool",
									"mode":"cdq",
									"devices":[
										{
											"name":"dev1"
										},
										{
											"mac":"98:03:9b:6a:b4:ez"
										}
									]
								}
							]
						}`,
			expErr: errors.New(deviceValidMacError),
		},
		{
			name: "device pci must be valid 1",
			configFile: `{
							"pools":[
								{
									"name":"testPool",
									"mode":"cdq",
									"devices":[
										{
											"name":"dev1"
										},
										{
											"pci":"0000:81:00"
										}
									]
								}
							]
						}`,
			expErr: errors.New(deviceValidPciError),
		},
		{
			name: "device pci must be valid 2",
			configFile: `{
							"pools":[
								{
									"name":"testPool",
									"mode":"cdq",
									"devices":[
										{
											"name":"dev1"
										},
										{
											"pci":"0000:81:00.z"
										}
									]
								}
							]
						}`,
			expErr: errors.New(deviceValidPciError),
		},
		{
			name: "device must have ID 1",
			configFile: `{
							"pools":[
								{
									"name":"testPool",
									"mode":"cdq",
									"devices":[
										{
											"name":"dev1"
										},
										{
										}
									]
								}
							]
						}`,
			expErr: errors.New(deviceMustHaveIdError),
		},
		{
			name: "device must have ID 2",
			configFile: `{
							"pools":[
								{
									"name":"testPool",
									"mode":"cdq",
									"devices":[
										{
											"name":"dev1"
										},
										{
											"name":""
										}
									]
								}
							]
						}`,
			expErr: errors.New(deviceMustHaveIdError),
		},
		{
			name: "device must have ID 3",
			configFile: `{
							"pools":[
								{
									"name":"testPool",
									"mode":"cdq",
									"devices":[
										{
											"name":"dev1"
										},
										{
											"mac":""
										}
									]
								}
							]
						}`,
			expErr: errors.New(deviceMustHaveIdError),
		},
		{
			name: "device must have ID 4",
			configFile: `{
							"pools":[
								{
									"name":"testPool",
									"mode":"cdq",
									"devices":[
										{
											"name":"dev1"
										},
										{
											"pci":""
										}
									]
								}
							]
						}`,
			expErr: errors.New(deviceMustHaveIdError),
		},
		{
			name: "device must only have one ID 1",
			configFile: `{
							"pools":[
								{
									"name":"testPool",
									"mode":"cdq",
									"devices":[
										{
											"name":"dev1"
										},
										{
											"name":"dev2",
											"mac":"98:03:9b:6a:b4:ef"
										}
									]
								}
							]
						}`,
			expErr: errors.New(deviceOnlyOneIdError),
		},
		{
			name: "device must only have one ID 2",
			configFile: `{
							"pools":[
								{
									"name":"testPool",
									"mode":"cdq",
									"devices":[
										{
											"name":"dev1"
										},
										{
											"name":"dev2",
											"pci":"0000:81:00.1"
										}
									]
								}
							]
						}`,
			expErr: errors.New(deviceOnlyOneIdError),
		},
		{
			name: "device must only have one ID 3",
			configFile: `{
							"pools":[
								{
									"name":"testPool",
									"mode":"cdq",
									"devices":[
										{
											"name":"dev1"
										},
										{
											"mac":"98:03:9b:6a:b4:ef",
											"name":"dev2"
										}
									]
								}
							]
						}`,
			expErr: errors.New(deviceOnlyOneIdError),
		},
		{
			name: "device must only have one ID 4",
			configFile: `{
							"pools":[
								{
									"name":"testPool",
									"mode":"cdq",
									"devices":[
										{
											"name":"dev1"
										},
										{
											"mac":"98:03:9b:6a:b4:ef",
											"pci":"0000:81:00.1"
										}
									]
								}
							]
						}`,
			expErr: errors.New(deviceOnlyOneIdError),
		},
		{
			name: "device must only have one ID 5",
			configFile: `{
							"pools":[
								{
									"name":"testPool",
									"mode":"cdq",
									"devices":[
										{
											"name":"dev1"
										},
										{
											"pci":"0000:81:00.1",
											"name":"dev2"
										}
									]
								}
							]
						}`,
			expErr: errors.New(deviceOnlyOneIdError),
		},
		{
			name: "device must only have one ID 6",
			configFile: `{
							"pools":[
								{
									"name":"testPool",
									"mode":"cdq",
									"devices":[
										{
											"name":"dev1"
										},
										{
											"pci":"0000:81:00.1",
											"mac":"98:03:9b:6a:b4:ef"
										}
									]
								}
							]
						}`,
			expErr: errors.New(deviceOnlyOneIdError),
		},
		{
			name: "device must only have one ID 7",
			configFile: `{
							"pools":[
								{
									"name":"testPool",
									"mode":"cdq",
									"devices":[
										{
											"name":"dev1"
										},
										{
											"name":"dev2",
											"mac":"98:03:9b:6a:b4:ef",
											"pci":"0000:81:00.1"
										}
									]
								}
							]
						}`,
			expErr: errors.New(deviceOnlyOneIdError),
		},
		{
			name: "secondary devices must not be below below min",
			configFile: `{
							"pools":[
								{
									"name":"testPool",
									"mode":"cdq",
									"devices":[
										{
											"name":"dev1"
										},
										{
											"name":"dev2",
											"secondary":-10
										}
									]
								}
							]
						}`,
			expErr: errors.New(deviceSecondaryError),
		},
		{
			name: "secondary devices must not be above max",
			configFile: `{
							"pools":[
								{
									"name":"testPool",
									"mode":"cdq",
									"devices":[
										{
											"name":"dev1"
										},
										{
											"name":"dev2",
											"secondary":9999
										}
									]
								}
							]
						}`,
			expErr: errors.New(deviceSecondaryError),
		},
		/*********************** Driver Validation ***********************/
		{
			name: "driver name must only use certain characters 1",
			configFile: `{
							"pools":[
								{
									"name":"testPool",
									"mode":"cdq",
									"drivers":[
										{
											"name":"i40e"
										},
										{
											"name":"ice$"
										}
									]
								}
							]
						}`,
			expErr: errors.New(driverValidError),
		},
		{
			name: "driver name must only use certain characters 2",
			configFile: `{
							"pools":[
								{
									"name":"testPool",
									"mode":"cdq",
									"drivers":[
										{
											"name":"i40e"
										},
										{
											"name":"ice driver"
										}
									]
								}
							]
						}`,
			expErr: errors.New(driverValidError),
		},
		{
			name: "driver name must not be too long",
			configFile: `{
							"pools":[
								{
									"name":"testPool",
									"mode":"cdq",
									"drivers":[
										{
											"name":"abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyz"
										}
									]
								}
							]
						}`,
			expErr: errors.New(driverNameLengthError),
		},
		{
			name: "driver must have a name 1",
			configFile: `{
							"pools":[
								{
									"name":"testPool",
									"mode":"cdq",
									"drivers":[
										{
										}
									]
								}
							]
						}`,
			expErr: errors.New(driverMustHaveIdError),
		},
		{
			name: "driver must have a name 2",
			configFile: `{
							"pools":[
								{
									"name":"testPool",
									"mode":"cdq",
									"drivers":[
										{
											"name":""
										}
									]
								}
							]
						}`,
			expErr: errors.New(driverMustHaveIdError),
		},
		{
			name: "number of primary devices must not be below min",
			configFile: `{
							"pools":[
								{
									"name":"testPool",
									"mode":"cdq",
									"drivers":[
										{
											"name":"ice",
											"primary":-5
										}
									]
								}
							]
						}`,
			expErr: errors.New(driverPrimaryError),
		},
		{
			name: "number of primary devices must not be above max",
			configFile: `{
							"pools":[
								{
									"name":"testPool",
									"mode":"cdq",
									"drivers":[
										{
											"name":"ice",
											"primary":105
										}
									]
								}
							]
						}`,
			expErr: errors.New(driverPrimaryError),
		},
		{
			name: "number of secondary devices must not be below min",
			configFile: `{
							"pools":[
								{
									"name":"testPool",
									"mode":"cdq",
									"drivers":[
										{
											"name":"ice",
											"secondary":-10
										}
									]
								}
							]
						}`,
			expErr: errors.New(deviceSecondaryError),
		},
		{
			name: "number of secondary devices must not be above max",
			configFile: `{
							"pools":[
								{
									"name":"testPool",
									"mode":"cdq",
									"drivers":[
										{
											"name":"ice",
											"secondary":9999
										}
									]
								}
							]
						}`,
			expErr: errors.New(deviceSecondaryError),
		},
		{
			name: "exclude devices must havs id",
			configFile: `{
							"pools":[
								{
									"name":"testPool",
									"mode":"cdq",
									"drivers":[
										{
											"name":"ice",
											"excludeDevices":[
												{}
											]
										}
									]
								}
							]
						}`,
			expErr: errors.New(deviceMustHaveIdError),
		},
		{
			name: "exclude devices must use only one form of id",
			configFile: `{
							"pools":[
								{
									"name":"testPool",
									"mode":"cdq",
									"drivers":[
										{
											"name":"ice",
											"excludeDevices":[
												{
													"name":"dev1"
												},
												{
													"pci":"0000:81:00.1",
													"mac":"98:03:9b:6a:b4:ef"
												}
											]
										}
									]
								}
							]
						}`,
			expErr: errors.New(deviceOnlyOneIdError),
		},
		/*********************** Node Validation ***********************/
		{
			name: "node must use a valid hostname 1",
			configFile: `{
							"pools":[
								{
									"name":"testPool",
									"mode":"cdq",
									"nodes":[
										{
											"hostname":"k8_node1"
										}
									]
								}
							]
						}`,
			expErr: errors.New(nodeValidHostError),
		},
		{
			name: "node must use a valid hostname 2",
			configFile: `{
							"pools":[
								{
									"name":"testPool",
									"mode":"cdq",
									"nodes":[
										{
											"hostname":"k8s node1"
										}
									]
								}
							]
						}`,
			expErr: errors.New(nodeValidHostError),
		},
		{
			name: "node hostname must not be too long",
			configFile: `{
							"pools":[
								{
									"name":"testPool",
									"mode":"cdq",
									"nodes":[
										{
											"hostname":"abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyz"
										}
									]
								}
							]
						}`,
			expErr: errors.New(nodeHostLengthError),
		},
		{
			name: "node must have a hostname 1",
			configFile: `{
							"pools":[
								{
									"name":"testPool",
									"mode":"cdq",
									"nodes":[
										{}
									]
								}
							]
						}`,
			expErr: errors.New(nodeMustHaveIdError),
		},
		{
			name: "node must have a hostname 2",
			configFile: `{
							"pools":[
								{
									"name":"testPool",
									"mode":"cdq",
									"nodes":[
										{
											"hostname":""
										}
									]
								}
							]
						}`,
			expErr: errors.New(nodeMustHaveIdError),
		},
		{
			name: "node must have devices or drivers",
			configFile: `{
							"pools":[
								{
									"name":"testPool",
									"mode":"cdq",
									"nodes":[
										{
											"hostname":"k8s-node1"
										}
									]
								}
							]
						}`,
			expErr: errors.New(nodeMustHaveDevsError),
		},
		{
			name: "node devices must have id",
			configFile: `{
							"pools":[
								{
									"name":"testPool",
									"mode":"cdq",
									"nodes":[
										{
											"hostname":"k8s-node1",
											"devices":[
												{}
											]
										}
									]
								}
							]
						}`,
			expErr: errors.New(deviceMustHaveIdError),
		},
		{
			name: "node devices must use only one form of id",
			configFile: `{
							"pools":[
								{
									"name":"testPool",
									"mode":"cdq",
									"nodes":[
										{
											"hostname":"k8s-node1",
											"devices":[
												{
													"name":"dev1",
													"mac":"98:03:9b:6a:b4:ef"
												}
											]
										}
									]
								}
							]
						}`,
			expErr: errors.New(deviceOnlyOneIdError),
		},
		/*********************** Pool Validation ***********************/
		{
			name: "pool must have a name 1",
			configFile: `{
							"pools":[
								{
									"mode":"cdq",
									"nodes":[
										{
											"hostname":"k8s-node1",
											"devices":[
												{
													"name":"dev1"
												}
											]
										}
									]
								}
							]
						}`,
			expErr: errors.New(poolNameRequiredError),
		},
		{
			name: "pool must have a name 2",
			configFile: `{
							"pools":[
								{
									"name":"",
									"mode":"cdq",
									"nodes":[
										{
											"hostname":"k8s-node1",
											"devices":[
												{
													"name":"dev1"
												}
											]
										}
									]
								}
							]
						}`,
			expErr: errors.New(poolNameRequiredError),
		},
		{
			name: "pool name must only contain letters and numbers",
			configFile: `{
							"pools":[
								{
									"name":"test-pool-1",
									"mode":"cdq",
									"nodes":[
										{
											"hostname":"k8s-node1",
											"devices":[
												{
													"name":"dev1"
												}
											]
										}
									]
								}
							]
						}`,
			expErr: errors.New(poolValidlNameError),
		},
		{
			name: "pool name must not be longer than max",
			configFile: `{
							"pools":[
								{
									"name":"abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyz",
									"mode":"cdq",
									"nodes":[
										{
											"hostname":"k8s-node1",
											"devices":[
												{
													"name":"dev1"
												}
											]
										}
									]
								}
							]
						}`,
			expErr: errors.New(poolNameLengthError),
		},
		{
			name: "pool must have a mode 1",
			configFile: `{
							"pools":[
								{
									"name":"testPool",
									"nodes":[
										{
											"hostname":"k8s-node1",
											"devices":[
												{
													"name":"dev1"
												}
											]
										}
									]
								}
							]
						}`,
			expErr: errors.New(poolModeRequiredError),
		},
		{
			name: "pool must have a mode 2",
			configFile: `{
							"pools":[
								{
									"name":"testPool",
									"mode":"",
									"nodes":[
										{
											"hostname":"k8s-node1",
											"devices":[
												{
													"name":"dev1"
												}
											]
										}
									]
								}
							]
						}`,
			expErr: errors.New(poolModeRequiredError),
		},
		{
			name: "pool must use a valid mode",
			configFile: `{
							"pools":[
								{
									"name":"testPool",
									"mode":"invalidMode",
									"nodes":[
										{
											"hostname":"k8s-node1",
											"devices":[
												{
													"name":"dev1"
												}
											]
										}
									]
								}
							]
						}`,
			expErr: errors.New(poolModeMustBeError),
		},
		{
			name: "pool must contain devices, drivers or nodes",
			configFile: `{
							"pools":[
								{
									"name":"testPool",
									"mode":"cdq"
								}
							]
						}`,
			expErr: errors.New(poolMustHaveDevsError),
		},
		{
			name: "pool devices must have id",
			configFile: `{
							"pools":[
								{
									"name":"testPool",
									"mode":"cdq",
									"hostname":"k8s-node1",
									"devices":[
										{}
									]
								}
							]
						}`,
			expErr: errors.New(deviceMustHaveIdError),
		},
		{
			name: "pool drivers must have id",
			configFile: `{
							"pools":[
								{
									"name":"testPool",
									"mode":"cdq",
									"hostname":"k8s-node1",
									"drivers":[
										{}
									]
								}
							]
						}`,
			expErr: errors.New(driverMustHaveIdError),
		},
		{
			name: "pool node devices must have id",
			configFile: `{
							"pools":[
								{
									"name":"testPool",
									"mode":"cdq",
									"nodes":[
										{
											"hostname":"k8s-node1",
											"devices":[
												{}
											]
										}
									]
								}
							]
						}`,
			expErr: errors.New(deviceMustHaveIdError),
		},
		{
			name: "pool node drivers must have id",
			configFile: `{
							"pools":[
								{
									"name":"testPool",
									"mode":"cdq",
									"nodes":[
										{
											"hostname":"k8s-node1",
											"drivers":[
												{}
											]
										}
									]
								}
							]
						}`,
			expErr: errors.New(driverMustHaveIdError),
		},
		{
			name: "uds timeout must not be below min 1",
			configFile: `{
							"pools":[
								{
									"name":"testPool",
									"mode":"cdq",
									"udsTimeout":20,
									"devices":[
										{
											"name":"dev1"
										},
										{
											"name":"dev2"
										}
									],
									"drivers":[
										{
											"name":"ice"
										}
									]
								}
							]
						}`,
			expErr: errors.New(poolUdsTimeoutError),
		},
		{
			name: "uds timeout must not be below min 2",
			configFile: `{
							"pools":[
								{
									"name":"testPool",
									"mode":"cdq",
									"udsTimeout":-40,
									"devices":[
										{
											"name":"dev1"
										},
										{
											"name":"dev2"
										}
									],
									"drivers":[
										{
											"name":"ice"
										}
									]
								}
							]
						}`,
			expErr: errors.New(poolUdsTimeoutError),
		},
		{
			name: "uds timeout must not be above max",
			configFile: `{
							"pools":[
								{
									"name":"testPool",
									"mode":"cdq",
									"udsTimeout":999,
									"devices":[
										{
											"name":"dev1"
										},
										{
											"name":"dev2"
										}
									],
									"drivers":[
										{
											"name":"ice"
										}
									]
								}
							]
						}`,
			expErr: errors.New(poolUdsTimeoutError),
		},
		{
			name: "ethtool cannot be empty",
			configFile: `{
							"pools":[
								{
									"name":"testPool",
									"mode":"cdq",
									"ethtoolCmds" : ["-X -device- equal 5 start 3", ""],
									"devices":[
										{
											"name":"dev1"
										},
										{
											"name":"dev2"
										}
									],
									"drivers":[
										{
											"name":"ice"
										}
									]
								}
							]
						}`,
			expErr: errors.New(poolEthtoolNotEmpty),
		},
		{
			name: "ethtool cannot be empty",
			configFile: `{
							"pools":[
								{
									"name":"testPool",
									"mode":"cdq",
									"ethtoolCmds" : ["-X -device- equal 5 start 3","--config-ntuple _device_ flow-type udp4 dst-ip -ip- action"],
									"devices":[
										{
											"name":"dev1"
										},
										{
											"name":"dev2"
										}
									],
									"drivers":[
										{
											"name":"ice"
										}
									]
								}
							]
						}`,
			expErr: errors.New(poolEthtoolCharacters),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfgFile = nil
			content := []byte(tc.configFile)
			dir, dirErr := ioutil.TempDir("/tmp", "test-cndp-")
			require.NoError(t, dirErr, "Can't create temporary directory")
			testDir := filepath.Join(dir, "tmpfile")
			err := ioutil.WriteFile(testDir, content, 0666)
			require.NoError(t, err, "Can't create temporary file")

			defer os.RemoveAll(dir)

			err = readConfigFile(testDir)
			if err == nil {
				assert.Equal(t, tc.expErr, err, "Error was expected")
			} else {
				require.Error(t, tc.expErr, "Unexpected error returned")
				assert.Contains(t, err.Error(), tc.expErr.Error(), "Unexpected error returned")
			}
		})
	}
}
