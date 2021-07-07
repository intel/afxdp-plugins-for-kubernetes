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
var excludedInfs = []string{"eno", "eth", "lo"}

/*
PoolConfig is  contains the pool name and device list
*/
type PoolConfig struct {
	Name    string   `json:"name"`
	Devices []string `json:"devices"`
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

	if len(cfg.Pools) == 0 {
		logging.Errorf("No pools configured, discovering devices on node")
		e, err := ethtool.NewEthtool()
		if err != nil {
			logging.Errorf("Error setting up Ethtool: %v", err)
			return cfg, err
		}
		defer e.Close()

		interfaces, err := net.Interfaces()
		if err != nil {
			logging.Errorf("Error setting up Interfaces: %v", err)
			return cfg, err
		}

		logging.Infof("Searching for devices on node...")
		poolConfigs := make(map[string]*PoolConfig)

		for _, driver := range driversTypes {
			poolConfigs[driver] = new(PoolConfig)
			poolConfigs[driver].Name = driver
		}

		for _, intf := range interfaces {
			if containsPrefix(excludedInfs, intf.Name) {
				logging.Infof("%s is an excluded device, skipping", intf.Name)
				continue
			}
			driver, err := e.DriverName(intf.Name)
			if err != nil {
				logging.Errorf("%v", err.Error())
			}
			if contains(driversTypes, driver) {
				addrs, err := intf.Addrs()
				if err != nil {
					logging.Errorf("%v", err.Error())
				}

				if len(addrs) > 0 {
					logging.Infof("%s has an assigned IP address, skipping", intf.Name)
					continue
				}
				poolConfigs[driver].Devices = append(poolConfigs[driver].Devices, intf.Name)
			}

		}
		for _, poolConfig := range poolConfigs {
			if len(poolConfig.Devices) != 0 {
				cfg.Pools = append(cfg.Pools, poolConfig)
			}
		}
	}
	return cfg, nil
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
