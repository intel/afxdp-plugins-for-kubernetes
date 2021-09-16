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

package main

import (
	"encoding/json"
	"fmt"
	"runtime"

	"github.com/vishvananda/netlink"

	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
	current "github.com/containernetworking/cni/pkg/types/100"
	"github.com/containernetworking/cni/pkg/version"
	"github.com/intel/cndp_device_plugin/pkg/bpf"
	"github.com/intel/cndp_device_plugin/pkg/logging"

	"github.com/containernetworking/plugins/pkg/ipam"
	"github.com/containernetworking/plugins/pkg/ns"
	"github.com/containernetworking/plugins/pkg/utils/buildversion"
)

var bpfHanfler = bpf.NewHandler()

type netConfig struct {
	types.NetConf
	Device   string `json:"deviceID"`
	LogFile  string `json:"logFile,omitempty"`
	LogLevel string `json:"logLevel,omitempty"`
}

func init() {
	runtime.LockOSThread()
}

func loadConf(bytes []byte) (*netConfig, error) {
	n := &netConfig{}
	if err := json.Unmarshal(bytes, n); err != nil {
		return nil, fmt.Errorf("loadConf(): failed to load network configuration: %v", err)
	}

	if n.LogFile != "" {
		logging.SetLogFile(n.LogFile)
	}

	if n.LogLevel != "" {
		logging.SetLogLevel(n.LogLevel)
	}

	logging.SetPluginName("CNDP-CNI")

	if n.Device == "" {
		return nil, fmt.Errorf("loadConf(): no device specified")
	}

	return n, nil
}

func cmdAdd(args *skel.CmdArgs) error {
	cfg, err := loadConf(args.StdinData)
	if err != nil {
		return err
	}
	logging.Debugf("cmdAdd(): loaded config: %v", cfg)

	logging.Infof("cmdAdd(): getting container network namespace")
	containerNs, err := ns.GetNS(args.Netns)
	if err != nil {
		err = fmt.Errorf("cmdAdd(): failed to open container netns %q: %v", args.Netns, err)
		logging.Errorf(err.Error())

		return err
	}
	defer containerNs.Close()

	logging.Infof("cmdAdd(): getting device from name")
	device, err := netlink.LinkByName(cfg.Device)
	if err != nil {
		err = fmt.Errorf("cmdAdd(): failed to find device: %v", err)
		logging.Errorf(err.Error())

		return err
	}

	logging.Infof("cmdAdd(): moving device from default to container network namespace")
	if err := netlink.LinkSetNsFd(device, int(containerNs.Fd())); err != nil {
		err = fmt.Errorf("cmdAdd(): failed to move device %q to container netns: %v", device.Attrs().Name, err)
		logging.Errorf(err.Error())

		return err
	}

	logging.Infof("cmdAdd(): executing within container network namespace:")
	if err := containerNs.Do(func(_ ns.NetNS) error {

		logging.Infof("cmdAdd(): set device to UP state")
		if err := netlink.LinkSetUp(device); err != nil {
			err = fmt.Errorf("cmdAdd(): failed to set device %q to UP state: %v", device.Attrs().Name, err)
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
			err = fmt.Errorf("cmdAdd(): error configuring IPAM on device %q: %v", device.Attrs().Name, err)
			logging.Errorf(err.Error())

			return err
		}
		return types.PrintResult(result, cfg.CNIVersion)
	}

	return printLink(device, cfg.CNIVersion, containerNs)
}

func cmdDel(args *skel.CmdArgs) error {
	cfg, err := loadConf(args.StdinData)
	if err != nil {
		err = fmt.Errorf("error loading config data: %v", err)
		logging.Errorf(err.Error())

		return err
	}

	logging.Infof("cmdDel(): getting container network namespace")
	containerNs, err := ns.GetNS(args.Netns)
	if err != nil {
		err = fmt.Errorf("cmdDel(): failed to open container netns %q: %v", args.Netns, err)
		logging.Errorf(err.Error())

		return err
	}
	defer containerNs.Close()

	logging.Infof("cmdDel(): getting default network namespace")
	defaultNs, err := ns.GetCurrentNS()
	if err != nil {
		err = fmt.Errorf("cmdDel(): failed to open default netns %q: %v", args.Netns, err)
		logging.Errorf(err.Error())

		return err
	}
	defer defaultNs.Close()

	logging.Infof("cmdDel(): executing within container network namespace:")
	if err := containerNs.Do(func(_ ns.NetNS) error {

		logging.Infof("cmdDel(): getting device from name")
		device, err := netlink.LinkByName(cfg.Device)
		if err != nil {
			err = fmt.Errorf("cmdDel(): failed to find device %q in containerNS: %v", cfg.Device, err)
			logging.Errorf(err.Error())

			return err
		}

		logging.Infof("cmdDel(): moving device from container to default network namespace")
		if err = netlink.LinkSetNsFd(device, int(defaultNs.Fd())); err != nil {
			err = fmt.Errorf("cmdDel(): failed to move %q to host netns: %v", device.Attrs().Alias, err)
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

	logging.Infof("cmdDel(): removing BPF program from device")
	if err := bpfHanfler.Cleanbpf(cfg.Device); err != nil {
		err = fmt.Errorf("cmdDel(): error removing BPF program from device: %v", err)
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

func configureIPAM(args *skel.CmdArgs, cfg *netConfig, device netlink.Link, netns ns.NetNS) (*current.Result, error) {
	var result *current.Result

	logging.Infof("configureIPAM(): running IPAM plugin: " + cfg.IPAM.Type)
	ipamResult, err := ipam.ExecAdd(cfg.IPAM.Type, args.StdinData)
	if err != nil {
		err = fmt.Errorf("configureIPAM(): failed to get IPAM: %v", err)
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
		err = fmt.Errorf("configureIPAM(): Failed to convert IPAM result into current result type: %v", err)
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
			err = fmt.Errorf("configureIPAM(): Error setting IPAM on device %q: %v", device.Attrs().Name, err)
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

func cmdCheck(args *skel.CmdArgs) error {
	return nil
}

func main() {
	skel.PluginMain(cmdAdd, cmdCheck, cmdDel, version.All, buildversion.BuildString("cndp-cni"))
}
