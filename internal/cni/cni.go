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

package cni

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"runtime"
	"strings"

	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
	current "github.com/containernetworking/cni/pkg/types/100"
	"github.com/containernetworking/plugins/pkg/ipam"
	"github.com/containernetworking/plugins/pkg/ns"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/intel/afxdp-plugins-for-kubernetes/constants"
	"github.com/intel/afxdp-plugins-for-kubernetes/internal/bpf"
	dpcnisyncer "github.com/intel/afxdp-plugins-for-kubernetes/internal/dpcnisyncerclient"
	"github.com/intel/afxdp-plugins-for-kubernetes/internal/host"
	"github.com/intel/afxdp-plugins-for-kubernetes/internal/logformats"
	"github.com/intel/afxdp-plugins-for-kubernetes/internal/networking"
	"github.com/intel/afxdp-plugins-for-kubernetes/internal/tools"
	logging "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
)

var bpfHandler = bpf.NewHandler()

/*
NetConfig holds the config passed via stdin
*/
type NetConfig struct {
	types.NetConf
	Device        string `json:"deviceID"`
	Mode          string `json:"mode"`
	SkipUnloadBpf bool   `json:"skipUnloadBpf,omitempty"`
	Queues        string `json:"queues,omitempty"`
	LogFile       string `json:"logFile,omitempty"`
	LogLevel      string `json:"logLevel,omitempty"`
	DPSyncer      bool   `json:"dpSyncer,omitempty"`
}

func init() {
	runtime.LockOSThread()
}

/*
Validate validates the contents of the Config struct
*/
func (n NetConfig) Validate() error {
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

	return validation.ValidateStruct(&n,
		validation.Field(
			&n.Device,
			validation.Required.Error("validate(): no device specified"),
			validation.Match(regexp.MustCompile(constants.Devices.ValidNameRegex)).Error("device names must only contain letters, numbers and selected symbols"),
		),
		validation.Field(
			&n.LogFile,
			validation.Match(regexp.MustCompile(constants.Logging.ValidFileRegex)).Error("must be a valid filename"),
		),
		validation.Field(
			&n.LogLevel,
			validation.In(logLevels...).Error("validate(): must be "+fmt.Sprintf("%v", logLevels)),
		),
		validation.Field(
			&n.Mode,
			validation.In(modes...).Error("validate(): must be "+fmt.Sprintf("%v", modes)),
		),
	)
}

func loadConf(bytes []byte) (*NetConfig, error) {
	n := &NetConfig{}
	logging.SetReportCaller(true)
	logging.SetFormatter(logformats.Default)

	if err := json.Unmarshal(bytes, n); err != nil {
		return nil, fmt.Errorf("loadConf(): failed to load network configuration: %w", err)
	}

	if err := n.Validate(); err != nil {
		return nil, fmt.Errorf("loadConf(): Config validation error: %v", err)
	}

	if n.LogFile != "" {
		fp, err := os.OpenFile(constants.Logging.Directory+n.LogFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, os.FileMode(constants.Logging.FilePermissions))
		if err != nil {
			return nil, fmt.Errorf("loadConf(): cannot open logfile %s: %w", n.LogFile, err)
		}
		logging.SetOutput(fp)
	}

	if n.LogLevel != "" {
		level, err := logging.ParseLevel(n.LogLevel)
		if err != nil {
			return nil, fmt.Errorf("loadConf(): cannot set log level: %w", err)
		}
		logging.SetLevel(level)

		if n.LogLevel == "debug" {
			logging.SetFormatter(logformats.Debug)
		}
	}

	if n.Mode != "" {
		logging.Debugf("loadConf(): Mode is set to: %s", n.Mode)
	}

	return n, nil
}

/*
CmdAdd is called by kubelet during pod create
*/
func CmdAdd(args *skel.CmdArgs) error {
	host := host.NewHandler()
	var result *current.Result
	var deviceDetails *networking.Device
	netHandler := networking.NewHandler()

	cfg, err := loadConf(args.StdinData)
	if err != nil {
		err = fmt.Errorf("cmdAdd(): error loading config data: %w", err)
		logging.Errorf(err.Error())

		return err
	}

	logging.Debugf("cmdAdd(): loaded config: %+v", cfg)
	logging.Infof("cmdAdd(): getting container network namespace")
	containerNs, err := ns.GetNS(args.Netns)
	if err != nil {
		err = fmt.Errorf("cmdAdd(): failed to open container netns %q: %w", args.Netns, err)
		logging.Errorf(err.Error())

		return err
	}
	defer containerNs.Close()

	logging.Infof("cmdAdd(): getting device from name")
	device, err := netlink.LinkByName(cfg.Device)
	if err != nil {
		err = fmt.Errorf("cmdAdd(): failed to find device: %w", err)
		logging.Errorf(err.Error())

		return err
	}

	logging.Infof("cmdAdd(): getting default network namespace")
	defaultNs, err := ns.GetCurrentNS()
	if err != nil {
		err = fmt.Errorf("cmdDel(): failed to open default netns %q: %w", args.Netns, err)
		logging.Errorf(err.Error())

		return err
	}
	defer defaultNs.Close()

	logging.Infof("cmdAdd(): checking if IPAM is required")
	if cfg.IPAM.Type != "" {
		result, err = getIPAM(args, cfg, device, defaultNs)
		if err != nil {
			err = fmt.Errorf("cmdAdd(): error configuring IPAM on device %q: %w", device.Attrs().Name, err)
			logging.Errorf(err.Error())

			return err
		}
	}

	if cfg.Mode == "primary" {
		deviceFile, err := tools.FilePathExists(constants.DeviceFile.Directory + constants.DeviceFile.Name)
		if err != nil {
			logging.Errorf("cmdAdd(): Failed to locate deviceFile: %v", err)
		}

		if deviceFile {
			deviceDetails, err = netHandler.GetDeviceFromFile(cfg.Device, constants.DeviceFile.Directory+constants.DeviceFile.Name)
			if err != nil {
				logging.Errorf("cmdAdd():- Failed to extract device map values: %v", err)
				return err
			}

			ethInstalled, version, err := host.HasEthtool()
			if err != nil {
				logging.Warningf("cmdAdd(): failed to discover ethtool on host: %v", err)
			}

			if ethInstalled {
				logging.Debugf("cmdAdd(): ethtool found on host")
				logging.Debugf("\t" + version)
				if deviceDetails != nil {
					if deviceDetails.GetEthtoolFilters() != nil {
						logging.Infof("cmdAdd(): applying ethtool filters on device: %s", cfg.Device)
						ethtoolCommand := deviceDetails.GetEthtoolFilters()
						iPAddr, err := extractIP(result)
						if err != nil {
							logging.Errorf("cmdAdd(): Error extracting IP from result interface %v", err)
							return err
						}
						err = netHandler.SetEthtool(ethtoolCommand, cfg.Device, iPAddr)
						if err != nil {
							logging.Errorf("cmdAdd(): unable to executed ethtool filter: %v", err)
							return err
						}
					} else {
						logging.Debugf("cmdAdd(): ethtool filters have not been specified")
					}
				}
			}
		}
	}

	logging.Infof("cmdAdd(): moving device from default to container network namespace")
	if err := netlink.LinkSetNsFd(device, int(containerNs.Fd())); err != nil {
		err = fmt.Errorf("cmdAdd(): failed to move device %q to container netns: %w", device.Attrs().Name, err)
		logging.Errorf(err.Error())

		return err
	}

	logging.Infof("cmdAdd(): executing within container network namespace:")
	if err := containerNs.Do(func(_ ns.NetNS) error {

		logging.Infof("cmdAdd(): set device to UP state")
		if err := netlink.LinkSetUp(device); err != nil {
			err = fmt.Errorf("cmdAdd(): failed to set device %q to UP state: %w", device.Attrs().Name, err)
			logging.Errorf(err.Error())

			return err
		}

		return nil
	}); err != nil {
		return err
	}

	if cfg.IPAM.Type != "" {
		result, err = setIPAM(cfg, result, device, containerNs)
		if err != nil {
			err = fmt.Errorf("cmdAdd(): error configuring IPAM on device netns %q: %w", device.Attrs().Name, err)
			logging.Errorf(err.Error())

			return err
		}
	}

	if result == nil {
		return printLink(device, cfg.CNIVersion, containerNs)
	}

	return types.PrintResult(result, cfg.CNIVersion)
}

/*
CmdDel is called by kublet during pod delete
*/
func CmdDel(args *skel.CmdArgs) error {
	host := host.NewHandler()
	netHandler := networking.NewHandler()

	cfg, err := loadConf(args.StdinData)
	if err != nil {
		err = fmt.Errorf("cmdDel(): error loading config data: %w", err)
		logging.Errorf(err.Error())

		return err
	}

	logging.Infof("cmdDel(): getting container network namespace")
	containerNs, err := ns.GetNS(args.Netns)
	if err != nil {
		err = fmt.Errorf("cmdDel(): failed to open container netns %q: %w", args.Netns, err)
		logging.Errorf(err.Error())

		return err
	}
	defer containerNs.Close()

	logging.Infof("cmdDel(): getting default network namespace")
	defaultNs, err := ns.GetCurrentNS()
	if err != nil {
		err = fmt.Errorf("cmdDel(): failed to open default netns %q: %w", args.Netns, err)
		logging.Errorf(err.Error())

		return err
	}
	defer defaultNs.Close()

	logging.Infof("cmdDel(): executing within container network namespace:")
	if err := containerNs.Do(func(_ ns.NetNS) error {

		logging.Infof("cmdDel(): getting device from name")
		device, err := netlink.LinkByName(cfg.Device)
		if err != nil {
			err = fmt.Errorf("cmdDel(): failed to find device %q in containerNS: %w", cfg.Device, err)
			logging.Errorf(err.Error())

			return err
		}

		logging.Infof("cmdDel(): moving device from container to default network namespace")
		if err = netlink.LinkSetNsFd(device, int(defaultNs.Fd())); err != nil {
			err = fmt.Errorf("cmdDel(): failed to move %q to host netns: %w", device.Attrs().Alias, err)
			logging.Errorf(err.Error())

			return err
		}

		return nil
	}); err != nil {
		return err
	}

	logging.Infof("cmdDel(): cleaning IPAM config on device")
	if cfg.IPAM.Type != "" {
		if err := ipam.ExecDel(cfg.IPAM.Type, args.StdinData); err != nil {
			return err
		}
	}

	if !cfg.SkipUnloadBpf {
		logging.Infof("cmdDel(): removing BPF program from device")
		if err := bpfHandler.Cleanbpf(cfg.Device); err != nil {
			err = fmt.Errorf("cmdDel(): error removing BPF program from device: %w", err)
			logging.Errorf(err.Error())

			return err
		}
	}

	if cfg.Mode == "primary" {
		logging.Debugf("cmdDel: checking host for Ethtool")
		ethInstalled, _, err := host.HasEthtool()
		if err != nil {
			logging.Errorf("cmdDel(): error checking if Ethtool is present on host: %v", err)
			return err
		}
		if ethInstalled {
			logging.Infof("cmdDel(): Removing ethtool filters on device: %s", cfg.Device)
			err := netHandler.DeleteEthtool(cfg.Device)
			if err != nil {
				logging.Warningf("cmdDel(): failed to remove ethtool filter: %v", err)
			}
		}
	}

	if cfg.DPSyncer == true {
		logging.Infof("cmdDel(): Asking Device Plugin to delete any BPF maps for %s", cfg.Device)
		err := dpcnisyncer.DeleteNetDev(cfg.Device)
		if err != nil {
			logging.Errorf("cmdDel(): DeleteNetDev from Syncer Server Failed for %s: %v", cfg.Device, err)
		}
	}

	if cfg.Mode == "cdq" {
		isSf, err := netHandler.IsCdqSubfunction(cfg.Device)
		if err != nil {
			logging.Errorf("cmdDel(): error determining if %s is a CDQ subfunction: %v", cfg.Device, err)
			isSf = false
		}
		if isSf {
			logging.Debugf("cmdDel(): deleting subfunction %s", cfg.Device)
			portIndex, err := netHandler.GetCdqPortIndex(cfg.Device)
			if err != nil {
				logging.Errorf("cmdDel(): error getting port index of device %s: %v", cfg.Device, err)
			} else {
				if err := netHandler.DeleteCdqSubfunction(portIndex); err != nil {
					logging.Errorf("cmdDel(): error deleting CDQ subfunction %s: %v", cfg.Device, err)
				} else {
					logging.Infof("cmdDel(): subfunction %s deleted", cfg.Device)
				}
			}
		}
	}

	return nil
}

func printLink(dev netlink.Link, cniVersion string, containerNs ns.NetNS) error {
	result := current.Result{
		CNIVersion: current.ImplementedSpecVersion,
		Interfaces: []*current.Interface{
			{
				Name:    dev.Attrs().Name,
				Mac:     dev.Attrs().HardwareAddr.String(),
				Sandbox: containerNs.Path(),
			},
		},
	}
	return types.PrintResult(&result, cniVersion)
}

func getIPAM(args *skel.CmdArgs, cfg *NetConfig, device netlink.Link, netns ns.NetNS) (*current.Result, error) {
	var result *current.Result

	logging.Infof("configureIPAM(): running IPAM plugin: " + cfg.IPAM.Type)
	ipamResult, err := ipam.ExecAdd(cfg.IPAM.Type, args.StdinData)
	if err != nil {
		err = fmt.Errorf("configureIPAM(): failed to get IPAM: %w", err)
		logging.Errorf(err.Error())

		return result, err
	}

	defer func() {
		if err != nil {
			logging.Debugf("configureIPAM(): An error occurred. Deleting IPAM to prevent IP leak.")
			err = ipam.ExecDel(cfg.IPAM.Type, args.StdinData)
			if err != nil {
				logging.Errorf("error while executing IPAM addition: %v", err)
			}
		}
	}()

	logging.Infof("configureIPAM(): converting IPAM result into current result type")
	result, err = current.NewResultFromResult(ipamResult)
	if err != nil {
		err = fmt.Errorf("configureIPAM(): Failed to convert IPAM result into current result type: %w", err)
		logging.Errorf(err.Error())

		return result, err
	}
	logging.Infof("configureIPAM(): checking IPAM plugin returned IP")
	if len(result.IPs) == 0 {
		err = fmt.Errorf("configureIPAM(): IPAM plugin returned no IPs")
		logging.Errorf(err.Error())

		return result, err
	}

	result.Interfaces = []*current.Interface{{
		Name:    device.Attrs().Name,
		Mac:     device.Attrs().HardwareAddr.String(),
		Sandbox: netns.Path(),
	}}
	for _, ipc := range result.IPs {
		logging.Debugf("configureIPAM(): setting IPConfig interface")
		ipc.Interface = current.Int(0)
	}

	return result, nil
}

func setIPAM(cfg *NetConfig, result *current.Result, device netlink.Link, netns ns.NetNS) (*current.Result, error) {
	logging.Infof("configureIPAM(): executing within host netns")
	if err := netns.Do(func(_ ns.NetNS) error {

		logging.Infof("configureIPAM(): setting device IP")
		if err := ipam.ConfigureIface(device.Attrs().Name, result); err != nil {
			err = fmt.Errorf("configureIPAM(): Error setting IPAM on device %q: %w", device.Attrs().Name, err)
			logging.Errorf(err.Error())

			return err
		}

		return nil
	}); err != nil {
		return result, err
	}

	result.DNS = cfg.DNS

	return result, nil
}

/*
CmdCheck is currently unused
*/
func CmdCheck(args *skel.CmdArgs) error {
	return nil
}

/*
extractIP extracts the IP address from the Result interface
and returns the IP as type string
*/
func extractIP(result *current.Result) (string, error) {
	resultIP := fmt.Sprintf("%s", result.IPs)
	if len(resultIP) != 0 {
		ip := strings.Split(resultIP, " ")[1]
		ip = strings.Split(ip, "IP:")[1]
		return ip, nil
	}
	err := fmt.Errorf("extractIP(): ip is an empty string")

	return resultIP, err
}
