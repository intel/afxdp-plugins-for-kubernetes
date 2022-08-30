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
	"encoding/json"
	"fmt"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/intel/afxdp-plugins-for-kubernetes/constants"
	"github.com/intel/afxdp-plugins-for-kubernetes/internal/networking"
	"github.com/intel/afxdp-plugins-for-kubernetes/internal/tools"
	logging "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"regexp"
)

const (
	maxUdsTimeout     = 300                           // Maximum timeout set to seconds
	defaultUdsTimeout = 30                            // Default timeout for unset timeout value in config.json
	logDirPermission  = 0744                          // Permissions for setting log directory
	logFilePermission = 0644                          // Permissions for setting log file
	logDir            = "/var/log/afxdp-k8s-plugins/" // Acceptable log directory
	minLinuxVersion   = "4.18.0"                      // Minimum Linux version for AF_XDP support
)

var (
	assignedInfs []string
	netHandler   networking.Handler
)

/*
PoolConfig contains the pool name and device list
*/
type PoolConfig struct {
	Name    string   `json:"name"`
	Devices []string `json:"devices"`
	Drivers []string `json:"drivers"`
	UID     int      `json:"uid"`
}

/*
Config contains the overall configuration for the device plugin
*/
type Config struct {
	LogDir                 string
	LogDirPermission       os.FileMode
	LogFilePermission      os.FileMode
	MinLinuxVersion        string
	Pools                  []*PoolConfig `json:"pools"`
	Mode                   string        `json:"mode"`
	LogFile                string        `json:"logFile"`
	LogLevel               string        `json:"logLevel"`
	UdsTimeout             int           `json:"timeout"`
	RequireUnprivilegedBpf bool          `json:"requireUnprivilegedBpf"`
	CndpFuzzing            bool          `json:"cndpFuzz"`
}

/*
GetConfig returns the overall config for the device plugin
*/
func GetConfig(configFile string, networking networking.Handler) (Config, error) {
	cfg := Config{
		LogDir:            logDir,
		LogDirPermission:  logDirPermission,
		LogFilePermission: logFilePermission,
		MinLinuxVersion:   minLinuxVersion,
	}
	netHandler = networking

	logging.Infof("Reading config file: %s", configFile)
	raw, err := ioutil.ReadFile(configFile)
	if err != nil {
		logging.Errorf("Error reading config file: %v", err)
		return cfg, err
	}

	logging.Infof("Unmarshalling config data")
	if err := json.Unmarshal(raw, &cfg); err != nil {
		logging.Errorf("Error unmarshalling config data: %v", err)
		return cfg, err
	}

	logging.Infof("Validating config data")
	if err := cfg.Validate(); err != nil {
		logging.Errorf("Config validation error: %v", err)
		return cfg, err
	}

	if cfg.UdsTimeout == -1 {
		cfg.UdsTimeout = 0
		logging.Debugf("Timeout is set to: %d seconds", cfg.UdsTimeout)
	} else if cfg.UdsTimeout == 0 {
		cfg.UdsTimeout = defaultUdsTimeout
		logging.Debugf("Using default value, timeout set to: %d seconds", cfg.UdsTimeout)
	} else {
		logging.Debugf("Timeout is set to: %d seconds", cfg.UdsTimeout)
	}

	return cfg, nil
}

/*
BuildPools builds up the device list in each of the configured pools
*/
func (c *Config) BuildPools() error {
	logging.Debugf("Checking pools for manually assigned devices")
	for _, pool := range c.Pools {
		for _, device := range pool.Devices {
			logging.Debugf("Device " + device + " has been manually assigned to pool " + pool.Name)

			if tools.ArrayContains(assignedInfs, device) {
				logging.Warningf("Device " + device + " is already assigned to another pool, removing from " + pool.Name)
				pool.Devices = tools.RemoveFromArray(pool.Devices, device)
				continue
			}

			assignedInfs = append(assignedInfs, device)
		}
	}

	logging.Debugf("Checking pools for assigned drivers")
	for _, pool := range c.Pools {
		if len(pool.Drivers) > 0 {
			logging.Debugf("Pool " + pool.Name + " has drivers assigned")

			for _, driver := range pool.Drivers {
				logging.Debugf("Pool " + pool.Name + " discovering devices of type " + driver)
				devices, err := deviceDiscovery(driver)
				if err != nil {
					logging.Errorf("Error discovering devices: %v", err.Error())
					return err
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

	return nil
}

/*
Validate validates the contents of the PoolConfig struct
*/
func (p PoolConfig) Validate() error {
	return validation.ValidateStruct(&p,
		validation.Field(
			&p.Name,
			validation.Required.Error("pools must have a name"),
			is.Alphanumeric.Error("pool names must only contain letters and numbers"),
		),
		validation.Field(
			&p.Devices,
			validation.Each(
				validation.Required.Error("devices must have a name"),
				validation.Match(regexp.MustCompile(constants.Devices.ValidNameRegex)).Error("device names must only contain letters, numbers and selected symbols"),
			),
			validation.Required.When(len(p.Drivers) == 0).Error("pools must contain devices or drivers"),
		),
		validation.Field(
			&p.Drivers,
			validation.Each(
				validation.Required.Error("drivers must have a name"),
				validation.Match(regexp.MustCompile(constants.Drivers.ValidNameRegex)).Error("driver names must only contain letters, numbers and selected symbols"),
			),
		),
		validation.Field(
			&p.UID,
			validation.When(!(p.UID == 0), validation.Max(constants.UID.Maximum)),
			validation.When(!(p.UID == 0), validation.Min(constants.UID.Minimum)),
		),
	)
}

/*
Validate validates the contents of the Config struct
*/
func (c Config) Validate() error {
	var (
		allowedLogLevels               = constants.Logging.Levels
		allowedModes                   = constants.Plugins.Modes
		logLevels        []interface{} = make([]interface{}, len(allowedLogLevels))
		modes            []interface{} = make([]interface{}, len(allowedModes))
	)

	for i, logLevel := range allowedLogLevels {
		logLevels[i] = logLevel
	}
	for i, mode := range allowedModes {
		modes[i] = mode
	}

	return validation.ValidateStruct(&c,
		validation.Field(
			&c.Pools,
			validation.Each(
				validation.NotNil.Error("cannot be null"),
			),
		),
		validation.Field(
			&c.LogLevel,
			validation.In(logLevels...).Error("must be "+fmt.Sprintf("%v", logLevels)),
		),
		validation.Field(
			&c.UdsTimeout,
			validation.When(!(c.UdsTimeout == -1 || c.UdsTimeout == 0), validation.Min(defaultUdsTimeout)),
			validation.When(!(c.UdsTimeout == -1 || c.UdsTimeout == 0), validation.Max(maxUdsTimeout)),
		),
		validation.Field(
			&c.Mode,
			//validation.Required.Error("validate(): Mode is required"), // TODO make required once more modes available
			validation.In(modes...).Error("validate(): must be "+fmt.Sprintf("%v", modes)),
		),
	)
}

func deviceDiscovery(requiredDriver string) ([]string, error) {
	var poolDevices []string

	hostDevices, err := netHandler.GetHostDevices()
	if err != nil {
		logging.Warningf("Error setting up Interfaces: %v", err)
		return poolDevices, err
	}

	for _, hostDevice := range hostDevices {
		if tools.ArrayContainsPrefix(constants.Devices.Prohibited, hostDevice.Name) {
			logging.Debugf("%s is an excluded device, skipping", hostDevice.Name)
			continue
		}

		deviceDriver, err := netHandler.GetDeviceDriver(hostDevice.Name)
		if err != nil {
			logging.Errorf("Error getting driver name: %v", err.Error())
			return poolDevices, err
		}

		if deviceDriver == requiredDriver {
			logging.Debugf("Device %s is type %s", hostDevice.Name, requiredDriver)

			if tools.ArrayContains(assignedInfs, hostDevice.Name) {
				logging.Infof("Device %s is already assigned to a pool, skipping", hostDevice.Name)
				continue
			}

			addrs, err := netHandler.GetAddresses(hostDevice)
			if err != nil {
				logging.Errorf("Error getting device IP: %v", err.Error())
				return poolDevices, err
			}

			if len(addrs) > 0 {
				logging.Infof("Device %s has an assigned IP address, skipping", hostDevice.Name)
				continue
			}

			poolDevices = append(poolDevices, hostDevice.Name)
			logging.Debugf("Device %s appended to the device list", hostDevice.Name)
		} else {
			logging.Debugf("%s has the wrong driver type: %s", hostDevice.Name, deviceDriver)
		}

	}
	return poolDevices, nil
}
