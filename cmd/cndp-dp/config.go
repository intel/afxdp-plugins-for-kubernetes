/*
 * Copyright(c) 2021 Intel Corporation.
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

package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/intel/cndp_device_plugin/pkg/logging"
	"github.com/intel/cndp_device_plugin/pkg/networking"
	"io/ioutil"
	"regexp"
	"strings"
)

var (
	driversTypes = []string{"i40e", "E810"}                                 // drivers we search for by default if none configured
	excludedInfs = []string{"eno", "eth", "lo", "docker", "flannel", "cni"} // interfaces we never add to a pool
	logLevels    = []string{"debug", "info", "warning", "error"}            // acceptable log levels
	logDir       = "/var/log/cndp/"                                         // acceptable log directory
	assignedInfs []string                                                   // keeps track of devices that are assigned to pools
)

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
	LogFile  string        `json:"logFile"`
	LogLevel string        `json:"logLevel"`
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
		return cfg, err
	}

	if err := json.Unmarshal(raw, &cfg); err != nil {
		logging.Errorf("Error unmarshalling config data: %v", err)
		return cfg, err
	}

	if err := cfg.Validate(); err != nil {
		logging.Errorf("Config validation error: %v", err)
		return cfg, err
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

/*
Validate validates the contents of the PoolConfig struct
*/
func (c PoolConfig) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(
			&c.Name,
			validation.Required.Error("pools must have a name"),
			is.Alphanumeric.Error("pool names can only contain letters and numbers"),
		),
		validation.Field(
			&c.Devices,
			validation.Each(
				validation.Required.Error("devices must have a name"),
				is.Alphanumeric.Error("device names can only contain letters and numbers"),
			),
			validation.Required.When(len(c.Drivers) == 0).Error("pools must contain devices or drivers"),
		),
		validation.Field(
			&c.Drivers,
			validation.Each(
				validation.Required.Error("drivers must have a name"),
				is.Alphanumeric.Error("driver names must only contain letters and numbers"),
			),
		),
	)
}

/*
Validate validates the contents of the Config struct
*/
func (c Config) Validate() error {
	var iLogLevels []interface{} = make([]interface{}, len(logLevels))
	for i, logLevel := range logLevels {
		iLogLevels[i] = logLevel
	}

	return validation.ValidateStruct(&c,
		validation.Field(
			&c.Pools,
			validation.Each(
				validation.NotNil.Error("cannot be null"),
			),
		),
		validation.Field(
			&c.LogFile,
			validation.Match(regexp.MustCompile("^/$|^(/[a-zA-Z0-9._-]+)+$")).Error("must be a valid filepath"),
		),
		validation.Field(
			&c.LogFile,
			validation.Match(regexp.MustCompile("^"+logDir)).Error("must in directory "+logDir),
		),
		validation.Field(
			&c.LogLevel,
			validation.In(iLogLevels...).Error("must be "+fmt.Sprintf("%v", iLogLevels)),
		),
	)
}

func deviceDiscovery(requiredDriver string) ([]string, error) {
	var poolDevices []string
	netHandler := networking.NewHandler()

	hostDevices, err := netHandler.GetHostDevices()
	if err != nil {
		logging.Warningf("Error setting up Interfaces: %v", err)
		return poolDevices, err
	}

	for _, hostDevice := range hostDevices {
		if containsPrefix(excludedInfs, hostDevice.Name) {
			logging.Debugf("%s is an excluded device, skipping", hostDevice.Name)
			continue
		}

		deviceDriver, err := netHandler.GetDeviceDriver(hostDevice.Name)
		if err != nil {
			logging.Errorf("Error getting driver name: %v", err.Error())
			return poolDevices, err
		}

		if deviceDriver == requiredDriver {
			logging.Infof("Device %s is type %s", hostDevice.Name, requiredDriver)

			if contains(assignedInfs, hostDevice.Name) {
				logging.Infof("Device %s is an already assigned to a pool, skipping", hostDevice.Name)
				continue
			}

			addrs, err := hostDevice.Addrs()
			if err != nil {
				logging.Errorf("Error getting device IP: %v", err.Error())
				return poolDevices, err
			}

			if len(addrs) > 0 {
				logging.Infof("Device %s has an assigned IP address, skipping", hostDevice.Name)
				continue
			}

			poolDevices = append(poolDevices, hostDevice.Name)
			logging.Infof("Device %s appended to the device list", hostDevice.Name)
		} else {
			logging.Debugf("%s has the wrong driver type: %s", hostDevice.Name, deviceDriver)
		}

	}
	return poolDevices, nil
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
