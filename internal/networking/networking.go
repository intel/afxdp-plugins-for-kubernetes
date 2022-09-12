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
	"errors"
	"fmt"
	"github.com/intel/afxdp-plugins-for-kubernetes/constants"
	"github.com/intel/afxdp-plugins-for-kubernetes/internal/tools"
	logging "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
	"net"
	"os"
	"os/exec"
	"path/filepath"
)

var (
	sysClassNet = "/sys/class/net"
	driverLink  = "device/driver"
	pciLink     = "device"
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
	CycleDevice(interfaceName string) error
	SetQueueSize(interfaceName string, size string) error
	SetDefaultQueueSize(interfaceName string) error
	NetDevExists(device string) (bool, error)
	IsPhysicalPort(name string) (bool, error)
	CreateCdqSubfunction(parentPci string, sfnum string) error     // see cdq.go
	DeleteCdqSubfunction(portIndex string) error                   // see cdq.go
	IsCdqSubfunction(name string) (bool, error)                    // see cdq.go
	NumAvailableCdqSubfunctions(interfaceName string) (int, error) // see cdq.go
	GetCdqPortIndex(netdev string) (string, error)                 // see cdq.go
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
SetQueueSize sets the queue size for the netdev.
It executes the command: ethtool -X <interface_name> equal <num_of_queues> start <queue_id>
*/
func (r *handler) SetQueueSize(interfaceName string, size string) error {
	app := "ethtool"
	args := "-X"
	startQID := "4"

	_, err := exec.Command(app, args, interfaceName, "equal", size, "start", startQID).Output()
	if err != nil {
		logging.Errorf("Error setting queue for device %s: %v", interfaceName, err.Error())
		return err
	}
	return nil
}

/*
SetDefaultQueueSize sets the netdev queue size back to default.
It executes the command: ethtool -X <interface_name> default
*/
func (r *handler) SetDefaultQueueSize(interfaceName string) error {
	app := "ethtool"
	args := "-X"

	_, err := exec.Command(app, args, interfaceName, "default").Output()
	if err != nil {
		logging.Errorf("Error setting default queue for device %s: %v", interfaceName, err.Error())
		return err
	}
	return nil
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
