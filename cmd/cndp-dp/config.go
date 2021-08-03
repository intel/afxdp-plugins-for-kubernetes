/*
 Copyright(c) 2021 Intel Corporation.
 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package main

import (
	"encoding/json"
	"github.com/intel/cndp_device_plugin/pkg/logging"
	"github.com/safchain/ethtool"
	"io/ioutil"
	"net"
	"strings"
)

var driversTypes = []string{"i40e", "E810"}
var excludedInfs = []string{"eno", "eth", "lo", "docker", "flannel", "cni"}
var assignedInfs []string

/*
PoolConfig is  contains the pool name and device list
*/
type PoolConfig struct {
	Name    string   `json:"name"`
	Devices []string `json:"devices"`
	Drivers []string `json:"drivers"`
}

/*
Config contains the overall configuration for the device plugin
*/
type Config struct {
	Pools    []*PoolConfig `json:"pools"`
	LogFile  string        `json:"logFile,omitempty"`
	LogLevel string        `json:"logLevel,omitempty"`
}

/*
GetConfig returns the overall config for the device plugin. Host devices are discovered if not explicitly set in the config file
*/
func GetConfig(configFile string) (Config, error) {
	var cfg Config

	logging.Infof("Reading config file: %s", configFile)
	raw, err := ioutil.ReadFile(configFile)
	if err != nil {
		logging.Errorf("Error reading config file: %v", err)
	} else {
		err = json.Unmarshal(raw, &cfg)
		if err != nil {
			logging.Errorf("Error unmarshalling config data: %v", err)
			return cfg, err
		}
	}

	if cfg.LogFile != "" {
		logging.SetLogFile(cfg.LogFile)
	}

	if cfg.LogLevel != "" {
		logging.SetLogLevel(cfg.LogLevel)
	}

	logging.Infof("Checking pools for manually assigned devices")
	for _, pool := range cfg.Pools {
		for _, device := range pool.Devices {
			logging.Infof("Device " + device + " has been manually assigned to pool " + pool.Name)

			if contains(assignedInfs, device) {
				logging.Warningf("Device " + device + " is already assigned to another pool, removing from " + pool.Name)
				pool.Devices = remove(pool.Devices, device)
				continue
			}

			assignedInfs = append(assignedInfs, device)
		}
	}

	logging.Infof("Checking pools for assigned drivers")
	for _, pool := range cfg.Pools {
		if len(pool.Drivers) > 0 {
			logging.Infof("Pool " + pool.Name + " has drivers assigned")

			for _, driver := range pool.Drivers {
				logging.Infof("Pool " + pool.Name + " discovering devices of type " + driver)
				devices, err := deviceDiscovery(driver)
				if err != nil {
					logging.Errorf("Error discovering devices: %v", err.Error())
					return cfg, err
				}

				if len(devices) > 0 {
					logging.Infof("Pool "+pool.Name+" discovered "+driver+" devices: %s", devices)

					for _, device := range devices {
						pool.Devices = append(pool.Devices, device)
						assignedInfs = append(assignedInfs, device)
					}
				}
			}
		}
	}

	if len(cfg.Pools) == 0 {
		logging.Infof("No pools configured, defaulting to pool per driver")
		for _, driver := range driversTypes {
			pool := new(PoolConfig)
			pool.Name = driver

			logging.Infof("Pool " + pool.Name + " discovering devices")
			devices, err := deviceDiscovery(driver)
			if err != nil {
				logging.Errorf("Error discovering devices: %v", err.Error())
				return cfg, err
			}

			if len(devices) > 0 {
				logging.Infof("Pool "+pool.Name+" discovered devices: %s", devices)

				for _, device := range devices {
					pool.Devices = append(pool.Devices, device)
					assignedInfs = append(assignedInfs, device)
				}

				cfg.Pools = append(cfg.Pools, pool)
			}
		}
	}
	return cfg, nil
}

func deviceDiscovery(requiredDriver string) ([]string, error) {
	var devices []string

	interfaces, err := net.Interfaces()
	if err != nil {
		logging.Warningf("Error setting up Interfaces: %v", err)
		return devices, err
	}

	ethtool, err := ethtool.NewEthtool()
	if err != nil {
		logging.Errorf("Error setting up Ethtool: %v", err)
		return devices, err
	}
	defer ethtool.Close()

	for _, intf := range interfaces {
		if containsPrefix(excludedInfs, intf.Name) {
			logging.Debugf("%s is an excluded device, skipping", intf.Name)
			continue
		}

		deviceDriver, err := ethtool.DriverName(intf.Name)
		if err != nil {
			logging.Errorf("Error getting driver name: %v", err.Error())
			return devices, err
		}

		if deviceDriver == requiredDriver {
			logging.Infof("Device %s is type %s", intf.Name, requiredDriver)

			if contains(assignedInfs, intf.Name) {
				logging.Infof("Device %s is an already assigned to a pool, skipping", intf.Name)
				continue
			}

			addrs, err := intf.Addrs()
			if err != nil {
				logging.Errorf("Error getting device IP: %v", err.Error())
				return devices, err
			}

			if len(addrs) > 0 {
				logging.Infof("Device %s has an assigned IP address, skipping", intf.Name)
				continue
			}

			devices = append(devices, intf.Name)
			logging.Infof("Device %s appended to the device list", intf.Name)
		} else {
			logging.Debugf("%s has the wrong driver type: %s", intf.Name, deviceDriver)
		}

	}
	return devices, nil
}

func contains(array []string, str string) bool {
	for _, s := range array {
		if s == str {
			return true
		}
	}
	return false
}

func containsPrefix(array []string, str string) bool {
	for _, s := range array {
		if strings.HasPrefix(str, s) {
			return true
		}
	}
	return false
}

func remove(array []string, rem string) []string {
	for i, elm := range array {
		if elm == rem {
			return append(array[:i], array[i+1:]...)
		}
	}
	return array
}
