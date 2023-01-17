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
var interfaceList map[string]*Device

/*
NewFakeHandler returns an implementation of the FakeHandler interface.
*/
func NewFakeHandler() FakeHandler {
	return &fakeHandler{}
}

/*
GetHostDevices returns a map of devices on the host
*/
func (r *fakeHandler) GetHostDevices() (map[string]*Device, error) {

	return interfaceList, nil
}

/*
SetHostDevices is a function used to dynamically setup mock devices and drivers
*/
func (r *fakeHandler) SetHostDevices(interfaceMap map[string][]string) {
	interfaceList = make(map[string]*Device)

	for driver, interfaceNames := range interfaceMap {
		for _, name := range interfaceNames {
			netDev, _ := newPrimaryDevice(name, driver, "1234", "1234", r)
			interfaceList[name] = netDev
		}
	}
}

/*
GetDeviceDriver takes a device name and returns the driver type.
In this fakeHandler it returns the driver of the fake netdev.
*/
func (r *fakeHandler) GetDeviceDriver(interfaceName string) (string, error) {
	return interfaceList[interfaceName].Driver()
}

/*
GetDevicePci takes a device name and returns the pci address.
In this fakeHandler it returns a dummy pci address.
*/
func (r *fakeHandler) GetDevicePci(interfaceName string) (string, error) {
	return "0000:18:00.3", nil
}

/*
IPAddresses takes a netdev name and returns its IP addresses
In this fakeHandler it returns the IP of the fake netdev.
*/
func (r *fakeHandler) GetIPAddresses(interfaceName string) ([]string, error) {
	var addrs []string
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

/*
GetMacAddress takes a device name and returns the MAC-address.
This function uses fake handler, its purpose is for unit-testing only
*/
func (r *fakeHandler) GetMacAddress(device string) (string, error) {
	return "", nil
}

/*
NetDevExists takes a device name and verifies if device exists on host.
This function uses fake handler, its purpose is for unit-testing
*/
func (r *fakeHandler) NetDevExists(device string) (bool, error) {
	return true, nil
}

/*
CreateCdqSubfunction takes the device name, PCI address of a port and a subfunction number
It creates that subfunction on top of that port and activates it
In this fake handler it does nothing
*/
func (r *fakeHandler) CreateCdqSubfunction(device string, parentPci string, sfnum string) error {
	return nil
}

/*
DeleteCdqSubfunction takes the port index of a subfunction, deactivates and deletes it
In this fake handler it does nothing
*/
func (r *fakeHandler) DeleteCdqSubfunction(portIndex string) error {
	return nil
}

/*
IsCdqSubfunction takes a netdev name and returns true if is a CDQ subfunction
In this fake handler it currently always returns true
*/
func (r *fakeHandler) IsCdqSubfunction(name string) (bool, error) {
	return true, nil
}

/*
GetCdqPortInfo takes a netdev name and returns the port number and index (pci/sfnum)
Note this function only works on physical devices and CDQ subfunctions
Other netdevs will return a "device not found by devlink" error
In this fake handler it currently returns an empty string
*/
func (r *fakeHandler) GetCdqPortInfo(netdev string) (string, string, error) {
	return "", "", nil
}

/*
NumAvailableCdqSubfunctions takes the PCI of a physical port and returns how
many unused CDQ subfunctions are available
In this fake handler it currently returns 0
*/
func (r *fakeHandler) NumAvailableCdqSubfunctions(interfaceName string) (int, error) {
	return 0, nil
}

/*
SetEthtool applies ethtool filters on the physical device during cmdAdd().
Ethtool filters are set via the DP config.json file. This function uses fake handler,
its purpose is for unit-testing only.
*/
func (r *fakeHandler) SetEthtool(ethtoolCmd []string, interfaceName string, ipResult string) error {
	return nil
}

/*
DeleteEthtool sets the default queue size ethtool filter.
It also removes perfect-flow ethtool filter entries during cmdDel()
This function uses fake handler, its purpose is for unit-testing
*/
func (r *fakeHandler) DeleteEthtool(interfaceName string) error {
	return nil
}

/*
GetDeviceFromFile extracts device map fields from the device file (device.json).
It creates and populates a new instance of the device map with the device file field values
and returns the device object.This function uses fake handler, its purpose is for unit-testing
*/
func (r *fakeHandler) GetDeviceFromFile(deviceName string, filepath string) (*Device, error) {
	return &Device{name: "fakeDevice", netHandler: r}, nil
}

/*
WriteDeviceFile creates and writes the device map fields to file, enabling the
CNI to read device information.This function uses fake handler, its purpose is for unit-testing
*/
func (r *fakeHandler) WriteDeviceFile(device *Device, filepath string) error {
	return nil
}

func (r *fakeHandler) GetDeviceByMAC(mac string) (string, error) {
	return "", nil
}

func (r *fakeHandler) GetDeviceByPCI(pci string) (string, error) {
	return "", nil
}

func (r *fakeHandler) IsPhysicalPort(name string) (bool, error) {
	return false, nil
}
