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
	"net"
)

/*
FakeHandler interface extends the Handler interface to provide additional testing methods.
*/
type FakeHandler interface {
	Handler
	SetHostDevices(interfaceNames map[string][]string)
}

/*
fakeHandler implements the FakeHandler interface.
*/
type fakeHandler struct{}

/*
interfaceList holds a map of drivers and net.Interface objects, representing fake netdev objects.
*/
var interfaceList map[string][]net.Interface

/*
NewFakeHandler returns an implementation of the FakeHandler interface.
*/
func NewFakeHandler() FakeHandler {
	return &fakeHandler{}
}

/*
GetHostDevices returns a list of fake net.Interface, representing the devices on the host.
*/
func (r *fakeHandler) GetHostDevices() ([]net.Interface, error) {
	var returnList []net.Interface

	for _, interfaceNames := range interfaceList {
		returnList = append(returnList, interfaceNames...)
	}

	return returnList, nil
}

/*
SetHostDevices is a function used to dynamically setup mock devices and drivers
*/
func (r *fakeHandler) SetHostDevices(interfaceMap map[string][]string) {
	interfaceList = make(map[string][]net.Interface)

	for driver, interfaceNames := range interfaceMap {
		for _, name := range interfaceNames {
			netDev := net.Interface{
				Index:        1,              // positive integer that starts at one, zero is never used
				MTU:          1,              // maximum transmission unit
				Name:         name,           // e.g., "en0", "lo0", "eth0.100"
				HardwareAddr: []byte("1234"), // IEEE MAC-48, EUI-48 and EUI-64 form
				Flags:        net.FlagUp,     // e.g., FlagUp, FlagLoopback, FlagMulticast
			}
			interfaceList[driver] = append(interfaceList[driver], netDev)
		}
	}
}

/*
GetDriverName takes a netdave name and returns the driver type.
In this fakeHandler it returns the driver of the fake netdev.
*/
func (r *fakeHandler) GetDeviceDriver(interfaceName string) (string, error) {
	for driver, devices := range interfaceList {
		for _, device := range devices {
			if device.Name == interfaceName {
				return driver, nil
			}
		}
	}

	return "defaultDriver", nil
}

/*
GetDevicePci takes a netdave name and returns the pci address.
In this fakeHandler it returns a dummy pci address.
*/
func (r *fakeHandler) GetDevicePci(interfaceName string) (string, error) {
	return "0000:18:00.3", nil
}

/*
GetAddresses takes a net.Interface and returns its IP addresses.
In this fakeHandler it returns the IP of the fake netdev.
*/
func (r *fakeHandler) GetAddresses(interfaceName net.Interface) ([]net.Addr, error) {
	var addrs []net.Addr
	return addrs, nil
}

/*
CycleDevice takes a netdave name and sets the device 'UP', then 'DOWN'
In this fake handler it does nothing.
*/
func (r *fakeHandler) CycleDevice(interfaceName string) error {
	return nil
}

/*
SetQueueSize sets the queue size for the netdev.
In this fake handler it does nothing.
*/
func (r *fakeHandler) SetQueueSize(interfaceName string, size string) error {
	return nil
}

/*
SetDefaultQueueSize sets the netdev queue size back to default.
In this fake handler it does nothing.
*/
func (r *fakeHandler) SetDefaultQueueSize(interfaceName string) error {
	return nil
}
