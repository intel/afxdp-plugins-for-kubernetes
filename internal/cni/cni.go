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
	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
	current "github.com/containernetworking/cni/pkg/types/100"
	"github.com/containernetworking/plugins/pkg/ipam"
	"github.com/containernetworking/plugins/pkg/ns"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/intel/afxdp-plugins-for-kubernetes/internal/bpf"
	"github.com/intel/afxdp-plugins-for-kubernetes/internal/logformats"
	"github.com/intel/afxdp-plugins-for-kubernetes/internal/networking"
	logging "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
	"os"
	"regexp"
	"runtime"
)

var (
	logLevels  = []string{"debug", "info", "warning", "error"} // acceptable log levels
	modes      = []string{"cndp"}                              // acceptable modes
	logDir     = "/var/log/afxdp-k8s-plugins/"                 // acceptable log directory
	bpfHanfler = bpf.NewHandler()
)

/*
NetConfig holds the config passed via stdin
*/
type NetConfig struct {
	types.NetConf
	Device   string `json:"deviceID"`
	Mode     string `json:"mode"`
	Queues   string `json:"queues,omitempty"`
	LogFile  string `json:"logFile,omitempty"`
	LogLevel string `json:"logLevel,omitempty"`
}

func init() {
	runtime.LockOSThread()
}

/*
Validate validates the contents of the Config struct
*/
func (n NetConfig) Validate() error {
	var iLogLevels []interface{} = make([]interface{}, len(logLevels))
	var iModes []interface{} = make([]interface{}, len(modes))

	for i, logLevel := range logLevels {
		iLogLevels[i] = logLevel
	}

	for i, mode := range modes {
		iModes[i] = mode
	}

	return validation.ValidateStruct(&n,
		validation.Field(
			&n.Device,
			validation.Required.Error("validate(): no device specified"),
			is.Alphanumeric.Error("validate(): device names can only contain letters and numbers"),
		),
		validation.Field(
			&n.Queues,
			validation.Required.When(n.Mode == "cndp").Error("Queues setting is required for CNDP mode"),
		),
		validation.Field(
			&n.LogFile,
			validation.Match(regexp.MustCompile("^/$|^(/[a-zA-Z0-9._-]+)+$")).Error("must be a valid filepath"),
			validation.Match(regexp.MustCompile("^"+logDir)).Error("validate(): must in directory "+logDir),
		),
		validation.Field(
			&n.LogLevel,
			validation.In(iLogLevels...).Error("validate(): must be "+fmt.Sprintf("%v", iLogLevels)),
		),
		validation.Field(
			&n.Mode,
			//validation.Required.Error("validate(): Mode is required"), // TODO make required once more modes available
			validation.In(iModes...).Error("validate(): must be "+fmt.Sprintf("%v", iModes)),
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
		fp, err := os.OpenFile(n.LogFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
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
CmdAdd is called by kublet during pod create
*/
func CmdAdd(args *skel.CmdArgs) error {
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

	if cfg.Queues != "" {
		logging.Infof("cmdAdd(): setting queue size %s for device %s ", cfg.Queues, device.Attrs().Name)
		if err := netHandler.SetQueueSize(device.Attrs().Name, cfg.Queues); err != nil {
			err = fmt.Errorf("cmdAdd(): failed to set queue size: %w", err)
			logging.Errorf(err.Error())

			return err
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

	logging.Infof("cmdAdd(): checking if IPAM is required")
	if cfg.IPAM.Type != "" {
		result, err := configureIPAM(args, cfg, device, containerNs)
		if err != nil {
			err = fmt.Errorf("cmdAdd(): error configuring IPAM on device %q: %w", device.Attrs().Name, err)
			logging.Errorf(err.Error())

			return err
		}
		return types.PrintResult(result, cfg.CNIVersion)
	}

	return printLink(device, cfg.CNIVersion, containerNs)
}

/*
CmdDel is called by kublet during pod delete
*/
func CmdDel(args *skel.CmdArgs) error {
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

	if cfg.Queues != "" {
		logging.Infof("cmdDel(): setting queue size to default for device %s", cfg.Device)
		if err := netHandler.SetDefaultQueueSize(cfg.Device); err != nil {
			err = fmt.Errorf("cmdDel(): failed to set queue size to default: %w", err)
			logging.Errorf(err.Error())

			return err
		}
	}

	logging.Infof("cmdDel(): cleaning IPAM config on device")
	if cfg.IPAM.Type != "" {
		if err := ipam.ExecDel(cfg.IPAM.Type, args.StdinData); err != nil {
			return err
		}
	}

	logging.Infof("cmdDel(): removing BPF program from device")
	if err := bpfHanfler.Cleanbpf(cfg.Device); err != nil {
		err = fmt.Errorf("cmdDel(): error removing BPF program from device: %w", err)
		logging.Errorf(err.Error())

		return err
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

func configureIPAM(args *skel.CmdArgs, cfg *NetConfig, device netlink.Link, netns ns.NetNS) (*current.Result, error) {
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

	logging.Infof("configureIPAM(): executing within container netns")
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
