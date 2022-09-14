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
	"fmt"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/intel/afxdp-plugins-for-kubernetes/constants"
	"regexp"
)

const (
	// device errors
	deviceValidNameError  = "Device name must only contain letters, numbers, hyphen and underscore"
	deviceNameLengthError = "Device name must be between 1 and 50 characters"
	deviceValidPciError   = "Device PCI must be a valid BDF PCI address"
	deviceValidMacError   = "Device MAC must be a valid MAC address"
	deviceMustHaveIdError = "Device must be identified by either name, pci, or mac"
	deviceOnlyOneIdError  = "Only one form of device identification can be used: name, pci, or mac"
	deviceSecondaryError  = "Number of secondary devices must be between 1 and 1000"

	// driver errors
	driverValidError      = "Driver name must only contain letters, numbers, hyphen and underscore"
	driverNameLengthError = "Driver name must be between 1 and 50 characters"
	driverMustHaveIdError = "Driver must have a name"
	driverPrimaryError    = "Number of primary devices must be between 1 and 100"

	// node errors
	nodeValidHostError    = "Node hostname must be a valid Linux hostname"
	nodeHostLengthError   = "Node hostname must be between 1 and 63 characters"
	nodeMustHaveIdError   = "Node must have a hostname"
	nodeMustHaveDevsError = "Node must contain devices or drivers"

	// pools errors
	poolValidlNameError   = "Pool name must only contain letters and numbers"
	poolNameRequiredError = "Pool must have a name"
	poolNameLengthError   = "Pool name must be between 1 and 20 characters"
	poolMustHaveDevsError = "Pool must contain devices, drivers or nodes"
	poolUdsTimeoutError   = "UDS socket timeout must be between 30 and 300 seconds"
	poolModeRequiredError = "Plugin must have a mode"
	poolModeMustBeError   = "Plugin mode must be one of "
	poolEthtoolNotEmpty   = "Ethtool commands cannot be empty"
	poolEthtoolCharacters = "Ethtool commands must be alphanumeric or contain only approved charaters"

	// logging errors
	filenameValidError = "must be a valid .log or .txt filename"
)

type configFile_Device struct {
	Name      string `json:"Name"`
	Pci       string `json:"Pci"`
	Mac       string `json:"Mac"`
	Secondary int    `json:"Secondary"`
}

type configFile_Driver struct {
	Name             string               `json:"Name"`
	Primary          int                  `json:"Primary"`
	Secondary        int                  `json:"Secondary"`
	ExcludeDevices   []*configFile_Device `json:"ExcludeDevices"`
	ExcludeAddressed bool                 `json:"ExcludeAddressed"`
}

type configFile_Node struct {
	Hostname string               `json:"Hostname"`
	Drivers  []*configFile_Driver `json:"Drivers"`
	Devices  []*configFile_Device `json:"Devices"`
}

type configFile_Pool struct {
	Name                    string               `json:"Name"`
	Mode                    string               `json:"Mode"`
	Drivers                 []*configFile_Driver `json:"Drivers"`
	Devices                 []*configFile_Device `json:"Devices"`
	Nodes                   []*configFile_Node   `json:"Nodes"`
	UdsServerDisable        bool                 `json:"UdsServerDisable"`
	UdsTimeout              int                  `json:"UdsTimeout"`
	UdsFuzz                 bool                 `json:"UdsFuzz"`
	RequiresUnprivilegedBpf bool                 `json:"RequiresUnprivilegedBpf"`
	UID                     int                  `json:"uid"`
	EthtoolCmds             []string             `json:"ethtoolCmds"`
}

type configFile struct {
	Pools    []*configFile_Pool `json:"Pools"`
	LogFile  string             `json:"LogFile"`
	LogLevel string             `json:"LogLevel"`
}

func (c configFile_Device) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(
			&c.Name,
			validation.Match(regexp.MustCompile(constants.Devices.ValidNameRegex)).Error(deviceValidNameError),
			validation.Length(constants.Devices.ValidNameMin, constants.Devices.ValidNameMax).Error(deviceNameLengthError),
			validation.Required.When(len(c.Pci) == 0 && len(c.Mac) == 0).Error(deviceMustHaveIdError),
			validation.Empty.When(len(c.Pci) > 0 || len(c.Mac) > 0).Error(deviceOnlyOneIdError),
		),
		validation.Field(
			&c.Pci,
			validation.Match(regexp.MustCompile(constants.Devices.ValidPciRegex)).Error(deviceValidPciError),
			validation.Required.When(len(c.Name) == 0 && len(c.Mac) == 0).Error(deviceMustHaveIdError),
			validation.Empty.When(len(c.Name) > 0 || len(c.Mac) > 0).Error(deviceOnlyOneIdError),
		),
		validation.Field(
			&c.Mac,
			is.MAC.Error(deviceValidMacError),
			validation.Required.When(len(c.Name) == 0 && len(c.Pci) == 0).Error(deviceMustHaveIdError),
			validation.Empty.When(len(c.Name) > 0 || len(c.Pci) > 0).Error(deviceOnlyOneIdError),
		),
		validation.Field(
			&c.Secondary,
			validation.When(
				c.Secondary != 0,
				validation.Min(constants.Devices.SecondaryMin).Error(deviceSecondaryError),
				validation.Max(constants.Devices.SecondaryMax).Error(deviceSecondaryError),
			),
		),
	)
}

func (c configFile_Driver) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(
			&c.Name,
			validation.Match(regexp.MustCompile(constants.Drivers.ValidNameRegex)).Error(driverValidError),
			validation.Length(constants.Drivers.ValidNameMin, constants.Drivers.ValidNameMax).Error(driverNameLengthError),
			validation.Required.Error(driverMustHaveIdError),
		),
		validation.Field(
			&c.Primary,
			validation.When(
				c.Primary != 0,
				validation.Min(constants.Drivers.PrimaryMin).Error(driverPrimaryError),
				validation.Max(constants.Drivers.PrimaryMax).Error(driverPrimaryError),
			),
		),
		validation.Field(
			&c.Secondary,
			validation.When(
				c.Secondary != 0,
				validation.Min(constants.Devices.SecondaryMin).Error(deviceSecondaryError),
				validation.Max(constants.Devices.SecondaryMax).Error(deviceSecondaryError),
			),
		),
		validation.Field(
			&c.ExcludeDevices,
		),
	)
}

func (c configFile_Node) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(
			&c.Hostname,
			validation.Match(regexp.MustCompile(constants.Nodes.ValidNameRegex)).Error(nodeValidHostError),
			validation.Length(constants.Nodes.ValidNameMin, constants.Nodes.ValidNameMax).Error(nodeHostLengthError),
			validation.Required.Error(nodeMustHaveIdError),
		),
		validation.Field(
			&c.Devices,
			validation.Required.When(len(c.Drivers) == 0).Error(nodeMustHaveDevsError),
		),
		validation.Field(
			&c.Drivers,
			validation.Required.When(len(c.Devices) == 0).Error(nodeMustHaveDevsError),
		),
	)
}

func (c configFile_Pool) Validate() error {
	var iModes []interface{} = make([]interface{}, len(constants.Plugins.Modes))

	for i, mode := range constants.Plugins.Modes {
		iModes[i] = mode
	}

	return validation.ValidateStruct(&c,
		validation.Field(
			&c.Name,
			validation.Required.Error(poolNameRequiredError),
			is.Alphanumeric.Error(poolValidlNameError),
			validation.Length(constants.Pools.ValidNameMin, constants.Pools.ValidNameMax).Error(poolNameLengthError),
		),
		validation.Field(
			&c.Mode,
			validation.Required.Error(poolModeRequiredError),
			validation.In(iModes...).Error(poolModeMustBeError+fmt.Sprintf("%v", iModes)),
		),
		validation.Field(
			&c.Drivers,
			validation.Required.When(len(c.Devices) == 0 && len(c.Nodes) == 0).Error(poolMustHaveDevsError),
		),
		validation.Field(
			&c.Devices,
			validation.Required.When(len(c.Drivers) == 0 && len(c.Nodes) == 0).Error(poolMustHaveDevsError),
		),
		validation.Field(
			&c.Nodes,
			validation.Required.When(len(c.Drivers) == 0 && len(c.Devices) == 0).Error(poolMustHaveDevsError),
		),
		validation.Field(
			&c.UdsTimeout,
			validation.When(
				c.UdsTimeout != 0,
				validation.Min(constants.Uds.MinTimeout).Error(poolUdsTimeoutError),
				validation.Max(constants.Uds.MaxTimeout).Error(poolUdsTimeoutError),
			),
		),
		validation.Field(
			&c.UID,
			validation.When(!(c.UID == 0), validation.Max(constants.UID.Maximum)),
			validation.When(!(c.UID == 0), validation.Min(constants.UID.Minimum)),
		),
		validation.Field(
			&c.EthtoolCmds,
			validation.Each(
				validation.Required.When(len(c.EthtoolCmds) > 0).Error(poolEthtoolNotEmpty),
				validation.Match(regexp.MustCompile(constants.EthtoolFilter.EthtoolFilterRegex)).Error(poolEthtoolCharacters),
			),
		),
	)
}

func (c configFile) Validate() error {
	var iLogLevels []interface{} = make([]interface{}, len(constants.Logging.Levels))

	for i, logLevel := range constants.Logging.Levels {
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
			validation.Match(regexp.MustCompile(constants.Logging.ValidFileRegex)).Error(filenameValidError),
		),
		validation.Field(
			&c.LogLevel,
			validation.In(iLogLevels...).Error("must be "+fmt.Sprintf("%v", iLogLevels)),
		),
	)
}

func (c configFile_Pool) GetDeviceList() []string {
	var list []string
	for _, dev := range c.Devices {
		list = append(list, dev.Name) //TODO needs to also figure pci and mac
	}
	return list
}

func (c configFile_Driver) GetExcludedDeviceList() []string {
	var list []string
	for _, dev := range c.ExcludeDevices {
		list = append(list, dev.Name) //TODO needs to also figure pci and mac
	}
	return list
}
