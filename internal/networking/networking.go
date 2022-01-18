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

package networking

import (
	"net"
	"strings"

	"github.com/go-cmd/cmd"
	logging "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
)

/*
Handler is the CNI and device plugins interface to the host networking.
The interface exists for testing purposes, allowing unit tests to test
against a fake API.
*/
type Handler interface {
	GetHostDevices() ([]net.Interface, error)
	GetDeviceDriver(interfaceName string) (string, error)
	GetAddresses(interfaceName net.Interface) ([]net.Addr, error)
	CycleDevice(interfaceName string) error
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
GetHostDevices returns a list of net.Interface, representing the devices
on the host.
*/
func (r *handler) GetHostDevices() ([]net.Interface, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		logging.Errorf("Error getting host devices: %v", err)
		return interfaces, err
	}
	return interfaces, nil
}

/*
GetDriverName takes a netdave name and returns the driver type
It executes the command: ethtool -i <interface_name>
*/
func (r *handler) GetDeviceDriver(interfaceName string) (string, error) {
	// build the command
	cmd := cmd.NewCmd("ethtool", "-i", interfaceName)

	// run and wait for cmd to return status
	status := <-cmd.Start()

	// take first line of Stdout - Stdout[0]
	// split the line on colon - ":"
	// driver is 2nd half of split string - [1]
	driver := strings.Split(status.Stdout[0], ":")[1]

	// trim whitespace and return
	return strings.TrimSpace(driver), nil
}

func (r *handler) GetAddresses(interfaceName net.Interface) ([]net.Addr, error) {
	return interfaceName.Addrs()
}

/*
CycleDevice takes a netdave name and sets the device 'UP', then 'DOWN'
Primerally used to workaround error - 22 of loading bpf prog onto a device
that was never 'UP', e.g. after a reboot
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
