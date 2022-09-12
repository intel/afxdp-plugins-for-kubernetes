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

package networking

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/intel/afxdp-plugins-for-kubernetes/constants"
	"github.com/intel/afxdp-plugins-for-kubernetes/internal/tools"
	logging "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
)

var (
	sysClassNet = "/sys/class/net"
	driverLink  = "device/driver"
	pciLink     = "device"
	pciDir      = "/sys/bus/pci/devices"
)

/*
Handler is the CNI and device plugins interface to the host networking.
The interface exists for testing purposes, allowing unit tests to test
against a fake API.
*/
type Handler interface {
	GetHostDevices() (map[string]*Device, error)
	GetDeviceDriver(interfaceName string) (string, error)
	GetDevicePci(interfaceName string) (string, error)
	GetIPAddresses(interfaceName string) ([]string, error)
	GetMacAddress(device string) (string, error)
	GetDeviceByMAC(mac string) (string, error)
	GetDeviceByPCI(pci string) (string, error)
	CycleDevice(interfaceName string) error
	NetDevExists(device string) (bool, error)
	GetDeviceFromFile(deviceName string, filepath string) (*Device, error)
	WriteDeviceFile(device *Device, filepath string) error
	CreateCdqSubfunction(parentPci string, sfnum string) error                   // see cdq.go
	DeleteCdqSubfunction(portIndex string) error                                 // see cdq.go
	IsCdqSubfunction(name string) (bool, error)                                  // see cdq.go
	NumAvailableCdqSubfunctions(interfaceName string) (int, error)               // see cdq.go
	GetCdqPortIndex(netdev string) (string, error)                               // see cdq.go
	SetEthtool(ethtoolCmd []string, interfaceName string, ipResult string) error // see ethtool.go
	DeleteEthtool(interfaceName string) error                                    // see ethtool.go
	IsPhysicalPort(name string) (bool, error)
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
GetHostDevices returns information relating to host network devices.
Device information is initialised and returned as a map of devices.
*/
func (r *handler) GetHostDevices() (map[string]*Device, error) {
	devices := make(map[string]*Device)

	interfaces, err := net.Interfaces()
	if err != nil {
		logging.Errorf("Error getting host devices: %v", err)
		return devices, err
	}
	for _, intf := range interfaces {
		pciAdd, err := r.GetDevicePci(intf.Name)
		if err != nil {
			return devices, err
		}
		driver, err := r.GetDeviceDriver(intf.Name)
		if err != nil {
			return devices, err
		}
		macAddr, err := r.GetMacAddress(intf.Name)
		if err != nil {
			return devices, err
		}
		newDev, err := newPrimaryDevice(intf.Name, driver, pciAdd, macAddr, r)
		if err != nil {
			return devices, err
		}
		devices[intf.Name] = newDev
	}
	return devices, nil
}

/*
IPAddresses takes a netdev name and returns its IP addresses
*/
func (r *handler) GetIPAddresses(interfaceName string) ([]string, error) {
	IPAddrs := []string{}
	Addrs, err := net.InterfaceByName(interfaceName)
	if err != nil {
		return IPAddrs, err
	}
	add, err := Addrs.Addrs()
	if err != nil {
		logging.Errorf("Error with GetAddress")
	}
	for _, addr := range add {
		IPAddrs = append(IPAddrs, addr.String())
	}

	return IPAddrs, nil
}

/*
CycleDevice takes a netdave name and sets the device 'UP', then 'DOWN'
Primerally used to workaround error - 22 of loading bpf prog onto a device
that was never in 'UP' state, e.g. after a reboot
Equivalent to 'ip link set <interface_name> down' and 'ip link set <interface_name> up'
*/
func (r *handler) CycleDevice(interfaceName string) error {
	device, err := netlink.LinkByName(interfaceName)
	if err != nil {
		return err
	}

	if err := netlink.LinkSetUp(device); err != nil {
		return err
	}

	if err := netlink.LinkSetDown(device); err != nil {
		return err
	}

	return nil
}

/*
GetDeviceDriver takes a netdave name and returns the driver type.
*/
func (r *handler) GetDeviceDriver(interfaceName string) (string, error) {
	link := filepath.Join(sysClassNet, interfaceName, driverLink)
	driverInfo, err := os.Readlink(link)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		logging.Errorf("Error getting driver for device %s: %v", interfaceName, err.Error())
		return "", err
	}
	return filepath.Base(driverInfo), nil
}

/*
GetDevicePci takes a netdave name and returns the pci address.
*/
func (r *handler) GetDevicePci(interfaceName string) (string, error) {
	link := filepath.Join(sysClassNet, interfaceName, pciLink)
	pciInfo, err := os.Readlink(link)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		logging.Errorf("Error getting PCI for device %s: %v", interfaceName, err.Error())
		return "", err
	}
	return filepath.Base(pciInfo), nil
}

/*
MacAddress takes a device name and returns the MAC-address.
*/
func (r *handler) GetMacAddress(device string) (string, error) {
	hwAddr, err := net.InterfaceByName(device)
	if err != nil {
		return "", err
	}
	macAddr := hwAddr.HardwareAddr.String()
	if true {
		return macAddr, nil
	}
	err = errors.New("device name cannot be empty")
	return macAddr, err

}

/*
NetDevExists takes a device name and verifies if device exists on host.
*/
func (r *handler) NetDevExists(device string) (bool, error) {
	_, err := netlink.LinkByName(device)
	if err != nil {
		if fmt.Sprint(err) == "Link not found" {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

/*
GetDeviceFromFile extracts device map fields from the device file (device.json).
It creates and populates a new instance of the device map with the device file field values
and returns the device object.
*/
func (r *handler) GetDeviceFromFile(deviceName string, filepath string) (*Device, error) {
	deviceDetailsMap := make(map[string]*DeviceDetails)

	raw, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal(raw, &deviceDetailsMap); err != nil {
		return nil, err
	}

	if deviceDetails, ok := deviceDetailsMap[deviceName]; ok {
		return &Device{
			name:           deviceDetails.Name,
			mode:           deviceDetails.Mode,
			driver:         deviceDetails.Driver,
			pci:            deviceDetails.Pci,
			macAddress:     deviceDetails.MacAddress,
			fullyAssigned:  deviceDetails.FullyAssigned,
			ethtoolFilters: deviceDetails.EthtoolFilters,
			netHandler:     r,
			primary: &Device{
				name:          deviceDetails.Primary.Name,
				mode:          deviceDetails.Primary.Mode,
				driver:        deviceDetails.Primary.Driver,
				pci:           deviceDetails.Primary.Pci,
				macAddress:    deviceDetails.Primary.MacAddress,
				fullyAssigned: deviceDetails.Primary.FullyAssigned,
			},
		}, nil
	}
	return nil, nil
}

/*
WriteDeviceFile creates and writes the device map fields to file, enabling the
CNI to read device information.
*/
func (r *handler) WriteDeviceFile(device *Device, filepath string) error {
	fileMap := make(map[string]DeviceDetails)

	raw, err := ioutil.ReadFile(filepath)
	if err != nil {
		logging.Warningf("Error reading device file: %v", err)
	}

	if err := json.Unmarshal(raw, &fileMap); err != nil {
		logging.Warningf("Error unmarshalling device map: %v", err)
	}

	fileMap[device.Name()] = device.Public()

	jsonStr, err := json.MarshalIndent(fileMap, "", " ")
	if err != nil {
		logging.Errorf("Error marshalling device map to json format: %v", err)
		return err
	}

	err = ioutil.WriteFile(filepath, jsonStr, os.FileMode(constants.DeviceFile.DeviceFilePermissions))
	if err != nil {
		logging.Errorf("Error writing to device file: %v", err)
		return err
	}

	return nil
}

/*
GetDeviceByMAC returns the device name assiciated with a MAC address. Returns "" if it does not exist.
*/
func (r *handler) GetDeviceByMAC(mac string) (string, error) {
	devList, err := netlink.LinkList()
	if err != nil {
		logging.Errorf("Unable to list devices")
		return "", err
	}

	for _, dev := range devList {
		if dev.Attrs().HardwareAddr.String() == mac {
			return dev.Attrs().Name, nil
		}
	}
	logging.Warnf("Device with MAC address %s not found", mac)
	return "", nil
}

/*
GetDeviceByPCI returns the device name associated with a PCI address. Returns "" if it does not exist.
*/
func (r *handler) GetDeviceByPCI(pci string) (string, error) {
	path := filepath.Join(pciDir, pci, "/net/")
	exists, err := tools.FilePathExists(path)
	if !exists || err != nil {
		logging.Errorf("Directory %s does not exist", path)
		return "", err
	}

	list, err := os.ReadDir(path)
	if err != nil {
		logging.Errorf("Unable to read directory names")
		return "", err
	}
	if len(list) <= 0 {
		logging.Warnf("Unable to get device name; no subdirectories found")
		return "", nil
	}

	return list[0].Name(), nil
}

/*
IsPhysicalPort takes in a device name. It returns true if it is a physical port, and false otherwise.
*/
func (r *handler) IsPhysicalPort(name string) (bool, error) {
	path := filepath.Join(sysClassNet, name, pciLink)
	physical, err := tools.FilePathExists(path)
	if err != nil {
		return false, err
	}

	if physical {
		// CDQ subfunctions need a further check
		driver, err := r.GetDeviceDriver(name)
		if err != nil {
			return false, err
		}
		if tools.ArrayContains(constants.Drivers.Cdq, driver) {
			subfunction, err := r.IsCdqSubfunction(name)
			if err != nil {
				return false, err
			}
			if subfunction {
				return false, nil
			}
		}
		return true, nil
	} else {
		return false, nil
	}
}
