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
	"github.com/intel/afxdp-plugins-for-kubernetes/constants"
	"github.com/intel/afxdp-plugins-for-kubernetes/internal/host"
	"github.com/intel/afxdp-plugins-for-kubernetes/internal/networking"
	"github.com/intel/afxdp-plugins-for-kubernetes/internal/tools"
	logging "github.com/sirupsen/logrus"
	"io/ioutil"
)

var (
	network     networking.Handler
	node        host.Handler
	cfgFile     *configFile
	hostDevices map[string]*networking.Device
)

/*
PluginConfig is the object that represents the overall plugin config.
Global configurations such as log levels are contained here.
*/
type PluginConfig struct {
	LogFile  string
	LogLevel string
}

/*
PoolConfig is the object representing the config of an individual device pool.
It contains pool specific details, such as mode and the device list.
This object is passed into the PoolManager.
*/
type PoolConfig struct {
	Name                    string                        // the name of the pool, used for logging and advertising resource to K8s. Pods will request this resource
	Mode                    string                        // the mode that this pool operates in
	Devices                 map[string]*networking.Device // a map of devices that the pool will manage
	Filters                 []string                      // the ethtool/rss filters we apply to devices from this pool
	UdsServerDisable        bool                          // a boolean to say if pods in this pool require BPF loading the UDS server
	UdsTimeout              int                           // timeout value in seconds for the UDS sockets, user provided or defaults to value from constants package
	UdsFuzz                 bool                          // a boolean to turn on fuzz testing within the UDS server, has no use outside of development and testing
	RequiresUnprivilegedBpf bool                          // a boolean to say if this pool requires unprivileged BPF
	UID                     int                           // the id of the pod user, we give this user ACL access to the UDS socket
	EthtoolCmds             []string                      // list of ethtool filters to apply to the netdev
}

/*
GetPluginConfig returns the global config for the device plugin.
This config is returned in a PluginConfig object
*/
func GetPluginConfig(configFile string) (PluginConfig, error) {
	var pluginConfig PluginConfig

	if cfgFile == nil {
		if err := readConfigFile(configFile); err != nil {
			logging.Errorf("Error reading config file: %v", err)
			return pluginConfig, err
		}
	}

	pluginConfig = PluginConfig{
		LogFile:  cfgFile.LogFile,
		LogLevel: cfgFile.LogLevel,
	}

	return pluginConfig, nil
}

/*
GetPoolConfigs returns a slice of PoolConfig objects.
Each object containing the config and device list for one pool.
*/
func GetPoolConfigs(configFile string, net networking.Handler, host host.Handler) ([]PoolConfig, error) {
	var poolConfigs []PoolConfig
	network = net
	node = host

	if cfgFile == nil {
		if err := readConfigFile(configFile); err != nil {
			logging.Errorf("Error reading config file: %v", err)
			return poolConfigs, err
		}
	}

	hostname, err := node.Hostname()
	if err != nil {
		logging.Errorf("Error getting node hostname: %v", err)
		return poolConfigs, err
	}

	unprivBpfAllowed, err := node.AllowsUnprivilegedBpf()
	if err != nil {
		logging.Errorf("Error checking if host allows unprivileged BPF operations: %v", err)
	}
	if unprivBpfAllowed {
		logging.Debugf("Unprivileged BPF is allowed on this host")
	} else {
		logging.Warningf("Unprivileged BPF is disabled on this host")
	}

	hostDevices, err = network.GetHostDevices()
	if err != nil {
		logging.Errorf("Error getting host devices: %v", err)
		return poolConfigs, err
	}

	for device := range hostDevices {
		physical, err := network.IsPhysicalPort(device)
		if err != nil {
			logging.Errorf("Error determining if %s is a physical device: %v", device, err)
			delete(hostDevices, device)
			continue
		}
		if !physical {
			logging.Debugf("%s is not a physical device, removing from list of host devices", device)
			delete(hostDevices, device)
			continue
		}
		if tools.ArrayContainsPrefix(constants.Devices.Prohibited, device) {
			logging.Debugf("%s a globally prohibited device, removing from list of host devices", device)
			delete(hostDevices, device)
			continue
		}
	}

	prettyDevices, err := tools.PrettyString(hostDevices)
	if err != nil {
		logging.Errorf("Error printing host devices: %v", err)
	} else {
		logging.Debugf("Host devices:\n%s", prettyDevices)
	}

	for _, pool := range cfgFile.Pools {
		logging.Infof("Processing Pool: %s", pool.Name)

		// check if pool requires unprivileged BPF and if the host allows it
		if pool.RequiresUnprivilegedBpf && !unprivBpfAllowed {
			logging.Warningf("Pool %s requires unprivileged BPF which is not allowed on this node", pool.Name)
			continue
		}

		// uds timeout - user disabled, user did not set, user set
		if pool.UdsTimeout == -1 {
			pool.UdsTimeout = 0
			logging.Debugf("UDS timeout is disabled: %d seconds", pool.UdsTimeout)
		} else if pool.UdsTimeout == 0 {
			pool.UdsTimeout = constants.Uds.MinTimeout
			logging.Debugf("Using default UDS timeout: %d seconds", pool.UdsTimeout)
		} else {
			logging.Debugf("UDS timeout is set to: %d seconds", pool.UdsTimeout)
		}

		// check if we have specific config for this node
		for _, node := range pool.Nodes {
			if node.Hostname == hostname {
				logging.Debugf("Pool %s has specific config for this node - %s", pool.Name, hostname)
				pool.Devices = node.Devices
				pool.Drivers = node.Drivers
				logging.Debugf("Devices and drivers updated to specific node config")
				break
			}
		}

		// if devices are configured check that they exist, are in a valid mode, etc.
		if pool.Devices != nil {
			var validDevices []*configFile_Device
			for _, device := range pool.Devices {
				if hostDev, ok := hostDevices[device.Name]; ok {
					if !validateDevice(hostDev, nil, pool) {
						continue
					}
					validDevices = append(validDevices, device)
				} else {
					logging.Warningf("Device %s does not exist on this node", device.Name)
				}
			}
			pool.Devices = validDevices
		}

		// if drivers are configured, get devices of that type
		if pool.Drivers != nil {
			for _, driver := range pool.Drivers {
				devices := getDeviceListOfDriverType(driver, pool)
				pool.Devices = append(pool.Devices, devices...)
			}
		}

		/*
			up until this point we have been building, configuring and validating our pool devices
			these devices have been of type configFile_Device, a basic object identifying a device
			getSecondaryDevices will take these objects and process them
			what is returned is a map of fully functional device objects from the networking package
			our devices become "real" at this point
		*/
		devices := getSecondaryDevices(pool)

		if len(devices) != 0 {
			poolConfigs = append(poolConfigs, PoolConfig{
				Name:                    pool.Name,
				Mode:                    pool.Mode,
				Devices:                 devices,
				UdsServerDisable:        pool.UdsServerDisable,
				UdsTimeout:              pool.UdsTimeout,
				UdsFuzz:                 pool.UdsFuzz,
				RequiresUnprivilegedBpf: pool.RequiresUnprivilegedBpf,
				UID:                     pool.UID,
			})
		}

	}

	return poolConfigs, nil
}

func getDeviceListOfDriverType(driver *configFile_Driver, pool *configFile_Pool) []*configFile_Device {
	var devices []*configFile_Device
	var counting bool

	deviceLimit := driver.Primary
	deviceCount := 0
	if deviceLimit > 0 {
		counting = true
	} else {
		counting = false
	}

	for _, hostDev := range hostDevices {
		hostDevDriver, err := hostDev.Driver()
		if err != nil {
			logging.Errorf("Error determining driver of device %s: %v", hostDev.Name(), err)
		}
		if hostDevDriver != driver.Name {
			logging.Debugf("%s is the wrong driver type: %s ", hostDev.Name(), hostDevDriver)
			continue
		}

		if !validateDevice(hostDev, driver, pool) {
			continue
		}

		device := configFile_Device{Name: hostDev.Name(), Secondary: driver.Secondary} // the device inherits the secondary limit from its driver
		devices = append(devices, &device)
		logging.Infof("%s added to pool", hostDev.Name())
		deviceCount++

		if counting && deviceCount >= deviceLimit {
			logging.Debugf("Pool %s has filled primary device allocation for %s driver", pool.Name, driver.Name)
			break
		}
	}

	logging.Debugf("Exit discovery.")
	return devices
}

func getSecondaryDevices(pool *configFile_Pool) map[string]*networking.Device {
	secondaryDevices := make(map[string]*networking.Device)

	for _, configDevice := range pool.Devices {
		if hostDevice, ok := hostDevices[configDevice.Name]; ok {
			switch pool.Mode {
			case "primary":
				dev, err := hostDevice.AssignAsPrimary()
				if err != nil {
					logging.Errorf("Error assigning device %s as primary: %v", hostDevice.Name(), err)
					continue
				}
				secondaryDevices[dev.Name()] = dev
			case "cdq":
				sfs, err := hostDevice.AssignCdqSecondaries(configDevice.Secondary)
				if err != nil {
					logging.Errorf("Error assigning subfunctions from device %s: %v", hostDevice.Name(), err)
					continue
				}
				for _, sf := range sfs {
					secondaryDevices[sf.Name()] = sf
				}
			default:
				logging.Errorf("Unsupported Mode: %s", pool.Mode)
			}
		} else {
			logging.Errorf("Device %s is not available on this host", configDevice.Name)
		}
	}
	return secondaryDevices
}

func validateDevice(device *networking.Device, driver *configFile_Driver, pool *configFile_Pool) bool {
	if _, ok := hostDevices[device.Name()]; !ok {
		logging.Debugf("Device %s does not exist on this node", device.Name())
		return false
	}

	if device.IsFullyAssigned() {
		logging.Debugf("Device %s is fully assigned", device.Name())
		return false
	}

	if tools.ArrayContainsPrefix(constants.Devices.Prohibited, device.Name()) {
		logging.Debugf("%s a globally prohibited device", device.Name())
		return false
	}

	if driver != nil {
		// if passed a driver, check that this device was not already manually configured
		if tools.ArrayContains(pool.GetDeviceList(), device.Name()) {
			logging.Debugf("%s is already in this pool", device.Name())
			return false
		}
		if tools.ArrayContains(driver.GetExcludedDeviceList(), device.Name()) {
			logging.Debugf("%s is an excluded device for %s driver", device.Name(), driver.Name)
			return false
		}
	}

	if (device.Mode() != "") && (device.Mode() != pool.Mode) {
		logging.Warningf("Device %s in the wrong mode: %s", device.Name(), device.Mode())
		return false
	}

	return true
}

func readConfigFile(file string) error {
	cfgFile = &configFile{}

	logging.Infof("Reading config file: %s", file)
	raw, err := ioutil.ReadFile(file)
	if err != nil {
		logging.Errorf("Error reading config file: %v", err)
		return err
	}

	logging.Infof("Unmarshalling config data")
	if err := json.Unmarshal(raw, &cfgFile); err != nil {
		logging.Errorf("Error unmarshalling config data: %v", err)
		return err
	}

	if cfgFile.LogLevel == "debug" {
		pretty, err := tools.PrettyString(cfgFile)
		if err != nil {
			logging.Errorf("Error printing config data: %v", err)
		} else {
			logging.Infof("Config Data:\n%s", pretty)
		}
	}

	logging.Infof("Validating config data")
	if err := cfgFile.Validate(); err != nil {
		logging.Errorf("Config validation error: %v", err)
		return err
	}
	return nil
}
