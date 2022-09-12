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
	"fmt"
	"github.com/intel/afxdp-plugins-for-kubernetes/constants"
	"github.com/intel/afxdp-plugins-for-kubernetes/internal/tools"
	logging "github.com/sirupsen/logrus"
	"strconv"
	"strings"
)

/*
Device object represents networking devices, primary and secondary
*/
type Device struct {
	name           string
	mode           string
	driver         string
	pci            string
	macAddress     string
	fullyAssigned  bool
	ethtoolFilters []string
	primary        *Device
	secondaries    []*Device
	netHandler     Handler
}

/*
DeviceDetails is a representation of Device above, but with public fields
This object has no functionality, methods or uses other than debug logging
and writing the device to a JSON file.
*/
type DeviceDetails struct {
	Name           string
	Mode           string
	Driver         string
	Pci            string
	MacAddress     string
	FullyAssigned  bool
	EthtoolFilters []string
	Primary        *DeviceDetails
}

/*
AssignAsPrimary means this device is assigned to a pool as a primary device
The device is put into primary mode and is set fully assigned, ensuring it will not be assigned again
*/
func (d *Device) AssignAsPrimary() (*Device, error) {
	if (d.mode == "") || (d.mode == "primary") {
		d.SetFullyAssigned()
		d.mode = "primary"
	} else {
		return nil, fmt.Errorf("Device is in an incompatible mode. %s is not compatible with primary mode", d.mode)
	}

	return d, nil
}

/*
AssignCdqSecondaries takes an integer and, if available, returns that number of CDQ subfunctions (secondary devs)
The primary device is put into CDQ mode. If the primary does not have yet have secondaries, they are now created
The function loops through the primary device's subfunctions and assigns any unassigned subfunctions.
An array of these newly assigned subfunctions is then returned.
*/
func (d *Device) AssignCdqSecondaries(limit int) ([]*Device, error) {
	var subFunctions []*Device
	var counting bool
	var deviceCount = 0

	if !tools.ArrayContains(constants.Drivers.Cdq, d.driver) {
		return nil, fmt.Errorf("Device has an incompatible driver, %s does not support CDQ", d.driver)
	}

	if (d.mode == "") || (d.mode == "cdq") {
		d.mode = "cdq"
	} else {
		return nil, fmt.Errorf("Device is in an incompatible mode. %s is not compatible with cdq mode", d.mode)
	}

	if limit > 0 {
		counting = true
	}

	if d.secondaries == nil {
		pci, err := d.Pci()
		if err != nil {
			return nil, fmt.Errorf("Error getting PCI address of device %s: %v", d.name, err)
		}

		numSF, err := d.netHandler.NumAvailableCdqSubfunctions(pci)
		if err != nil {
			return nil, fmt.Errorf("Error finding the number of available subfunctions on device %s: %v", d.name, err)
		}

		for i := 1; i <= numSF; i++ {
			newSF, err := newSecondaryDevice(d.name+"sf"+strconv.Itoa(i), d)
			if err != nil {
				continue
			}
			d.secondaries = append(d.secondaries, newSF)
		}
	}

	for _, sf := range d.secondaries {
		if !sf.IsFullyAssigned() {
			subFunctions = append(subFunctions, sf)
			sf.SetFullyAssigned()
			if counting {
				deviceCount++
			}
		}
		if counting && deviceCount >= limit {
			break
		}
	}

	return subFunctions, nil
}

/*
ActivateCdqSubfunction converts our device object in code into an actual CDQ subfunction on the host
*/
func (d *Device) ActivateCdqSubfunction() error {
	if d.IsPrimary() {
		return fmt.Errorf("Cannot activate CDQ subfunction %s. This is a primary device $s", d.name)
	}

	if !tools.ArrayContains(constants.Drivers.Cdq, d.driver) {
		return fmt.Errorf("Cannot activate CDQ subfunction %s. Driver %s is not CDQ compatible", d.name, d.driver)
	}

	if d.mode != "cdq" {
		return fmt.Errorf("Cannot activate CDQ subfunction %s. Device is not in CDQ mode $s", d.name)
	}

	exists, err := d.netHandler.NetDevExists(d.name)
	if err != nil {
		logging.Errorf("Error determining if subfunction %s already exists: %v", d.name, err)
		exists = false
	}

	if exists {
		logging.Warningf("Subfunction %s already exists", d.name)
		return nil
	}

	pci, err := d.primary.Pci()
	if err != nil {
		return fmt.Errorf("Error getting primary device PCI while activating subfunction %s", d.name)
	}

	sfNum := strings.Split(d.name, "sf")[1]

	err = d.netHandler.CreateCdqSubfunction(pci, sfNum)
	if err != nil {
		return fmt.Errorf("Error creating CDQ subfunction %s: %v", d.name, err)
	}

	return nil
}

/*
Name returns the name of the device
*/
func (d *Device) Name() string {
	return d.name
}

/*
Mode returns the mode of the device
*/
func (d *Device) Mode() string {
	return d.mode
}

/*
Driver will check Device object for its driver and return the result
If driver is not stored it will be discovered through the netHandler
Driver is then stored for subsequent calls
*/
func (d *Device) Driver() (string, error) {
	if d.driver != "" {
		return d.driver, nil
	}
	driver, err := d.netHandler.GetDeviceDriver(d.name)
	if err != nil {
		return "", err
	}

	d.driver = driver
	return d.driver, nil
}

/*
Pci will check Device object for its pci and return the result
If pci is not stored it will be discovered through the netHandler
Pci is then stored for subsequent calls
*/
func (d *Device) Pci() (string, error) {
	if d.pci != "" {
		return d.pci, nil
	}
	pci, err := d.netHandler.GetDevicePci(d.name)
	if err != nil {
		return pci, err
	}

	d.pci = pci
	return d.pci, nil
}

/*
Mac will check Device object for its mac and return the result
If mac is not stored it will be discovered through the netHandler
Mac is then stored for subsequent calls
For secondary devices, which tend to be created and deleted regularly,
we always recheck the mac.
*/
func (d *Device) Mac() (string, error) {
	if d.IsSecondary() {
		mac, err := d.netHandler.GetMacAddress(d.name)
		if err != nil {
			return "", err
		}
		d.macAddress = mac
	}

	if d.macAddress != "" {
		return d.macAddress, nil
	}
	mac, err := d.netHandler.GetMacAddress(d.name)
	if err != nil {
		return mac, err
	}

	d.macAddress = mac
	return d.macAddress, nil
}

/*
Ips are discovered through the netHandler
Ips are not stored as they can change frequently
*/
func (d *Device) Ips() ([]string, error) {
	ips, err := d.netHandler.GetIPAddresses(d.name)
	if err != nil {
		return nil, err
	}

	return ips, nil
}

/*
Primary returns a pointer to this device's primary device
Primary devices will return a pointer to themselves
*/
func (d *Device) Primary() *Device {
	return d.primary
}

/*
IsFullyAssigned returns the assignment status of the device
A fully assigned device will return true and should not be assigned to subsequent pools
*/
func (d *Device) IsFullyAssigned() bool {
	return d.fullyAssigned
}

/*
SetFullyAssigned is used to specify that the device is fully assigned
All primary mode devices should be automatically fully assigned or
in other modes a device is considered fully assigned when all secondaries
are assigned.
*/
func (d *Device) SetFullyAssigned() {
	d.fullyAssigned = true
}

/*
IsPrimary returns true if this is a primary device
Primary devices point to themselves in the primary field of the device object
*/
func (d *Device) IsPrimary() bool {
	return d.name == d.primary.name
}

/*
IsSecondary returns true if this is a secondary device
It simply returns the opposite of IsPrimary
*/
func (d *Device) IsSecondary() bool {
	return !d.IsPrimary()
}

/*
Cycle cycles the device state, up and then down
It uses the net handler CycleDevice function
*/
func (d *Device) Cycle() error {
	if err := d.netHandler.CycleDevice(d.name); err != nil {
		return err
	}
	return nil
}

/*
GetEthtoolFilters returns a string array of ethtool filters from
the device object
*/
func (d *Device) GetEthtoolFilters() []string {
	return d.ethtoolFilters
}

/*
UnassignedSecondaries returns the number of unassigned secondary devices available on this primary
*/
func (d *Device) UnassignedSecondaries() int {
	var count int
	for _, dev := range d.secondaries {
		if !dev.IsFullyAssigned() {
			count++
		}
	}
	return count
}

/*
Exists returns true if the device exists on the host (if the device has been created)
It uses the netHandler NetDevExists function. This "exists" status is not stored, as
secondary devices tend to be created and deleted frequently. We always check.
*/
func (d *Device) Exists() (bool, error) {
	deviceExists, err := d.netHandler.NetDevExists(d.name)
	if err != nil {
		return false, err
	}

	return deviceExists, nil
}

/*
Public returns a representation of Device, but with public fields
To be used in debug logging and writing the device to a JSON file.
*/
func (d *Device) Public() DeviceDetails {
	return DeviceDetails{
		Name:           d.name,
		Mode:           d.mode,
		Driver:         d.driver,
		Pci:            d.pci,
		MacAddress:     d.macAddress,
		FullyAssigned:  d.fullyAssigned,
		EthtoolFilters: d.ethtoolFilters,

		Primary: &DeviceDetails{
			Name:          d.primary.name,
			Mode:          d.primary.mode,
			Driver:        d.primary.driver,
			Pci:           d.primary.pci,
			MacAddress:    d.primary.macAddress,
			FullyAssigned: d.primary.fullyAssigned,
		},
	}
}

/*
newPrimaryDevice creates, initialises, and returns a primary device
Primary devices must have a name and a netHandler
*/
func newPrimaryDevice(name string, driver string, pci string, macAddress string,
	netHandler Handler) (*Device, error) {

	if name == "" {
		return nil, fmt.Errorf("device name cannot be empty")
	}

	if netHandler == nil {
		return nil, fmt.Errorf("device requires a network handler")
	}

	dev := &Device{
		name:       name,
		driver:     driver,
		pci:        pci,
		macAddress: macAddress,
		netHandler: netHandler,
	}
	dev.primary = dev

	return dev, nil
}

/*
newSecondaryDevice creates, initialises, and returns a secondary device
Secondary devices must have a name and be associated with a primary device
*/
func newSecondaryDevice(name string, primary *Device) (*Device, error) {
	if name == "" {
		return nil, fmt.Errorf("device name cannot be empty")
	}

	if primary == nil {
		return nil, fmt.Errorf("secondary devices must have a primary")
	}

	driver, err := primary.Driver()
	if err != nil {
		return nil, fmt.Errorf("error creating secondary device %s: %v", name, err)
	}

	dev := &Device{
		name:       name,
		mode:       primary.Mode(),
		driver:     driver,
		primary:    primary,
		netHandler: primary.netHandler,
	}

	return dev, nil
}

/*
CreateTestDevice returns a device object and is intended for unit testing purposes only
This function should not be used outside of testing
Devices should always be created via a net handler
*/
func CreateTestDevice(name string, driver string, pci string, macAddress string,
	netHandler Handler) *Device {

	dev := &Device{
		name:       name,
		driver:     driver,
		pci:        pci,
		macAddress: macAddress,
		netHandler: netHandler,
	}
	dev.primary = dev

	return dev
}

/*
SetEthtoolFilter assigns ethtool filters to the ethtoolFilters
field in the device object.
*/
func (d *Device) SetEthtoolFilter(ethtool []string) {
	d.ethtoolFilters = ethtool
}
