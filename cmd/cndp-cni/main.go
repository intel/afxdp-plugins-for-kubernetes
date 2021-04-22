/*
 Copyright(c) 2021 Intel Corporation.
 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package main

import (
	"encoding/json"
	"fmt"
	"runtime"

	"github.com/vishvananda/netlink"

	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
	"github.com/containernetworking/cni/pkg/types/current"
	"github.com/containernetworking/cni/pkg/version"
	"github.com/intel/cndp_device_plugin/pkg/bpf"

	"github.com/containernetworking/plugins/pkg/ipam"
	"github.com/containernetworking/plugins/pkg/ns"
	"github.com/containernetworking/plugins/pkg/utils/buildversion"
)

type NetConf struct {
	types.NetConf
	Device string `json:"deviceID"`
}

func init() {
	runtime.LockOSThread()
}

func loadConf(bytes []byte) (*NetConf, error) {
	n := &NetConf{}
	if err := json.Unmarshal(bytes, n); err != nil {
		return nil, fmt.Errorf("loadConf(): failed to load network configuration: %v", err)
	}

	if n.Device == "" {
		return nil, fmt.Errorf("loadConf(): no device specified")
	}

	return n, nil
}

func cmdAdd(args *skel.CmdArgs) error {
	//load config
	cfg, err := loadConf(args.StdinData)
	if err != nil {
		return err
	}

	//get container netns
	containerNs, err := ns.GetNS(args.Netns)
	if err != nil {
		return fmt.Errorf("cmdAdd(): failed to open container netns %q: %v", args.Netns, err)
	}
	defer containerNs.Close()

	//get netdev from name
	device, err := netlink.LinkByName(cfg.Device)
	if err != nil {
		return fmt.Errorf("cmdAdd(): failed to find device: %v", err)
	}

	//move netdev from default to container netns
	if err := netlink.LinkSetNsFd(device, int(containerNs.Fd())); err != nil {
		return fmt.Errorf("cmdAdd(): failed to move device %q to container netns: %v", device.Attrs().Name, err)
	}

	//execute within container netns:
	if err := containerNs.Do(func(_ ns.NetNS) error {

		//set netdev to UP state
		if err := netlink.LinkSetUp(device); err != nil {
			return fmt.Errorf("cmdAdd(): failed to set device %q to UP state: %v", device.Attrs().Name, err)
		}

		return nil
	}); err != nil {
		return err
	}

	//configure IPAM if required
	if cfg.IPAM.Type != "" {
		result, err := configureIPAM(args, cfg, device, containerNs)
		if err != nil {
			return fmt.Errorf("cmdAdd(): error configuring IPAM on device %q: %v", device.Attrs().Name, err)
		}
		return types.PrintResult(result, cfg.CNIVersion)
	}

	return printLink(device, cfg.CNIVersion, containerNs)
}

func cmdDel(args *skel.CmdArgs) error {
	//load config
	cfg, err := loadConf(args.StdinData)
	if err != nil {
		return err
	}

	//get container netns
	containerNs, err := ns.GetNS(args.Netns)
	if err != nil {
		return fmt.Errorf("cmdDel(): failed to open container netns %q: %v", args.Netns, err)
	}
	defer containerNs.Close()

	//get default netns
	defaultNs, err := ns.GetCurrentNS()
	if err != nil {
		return fmt.Errorf("cmdDel(): failed to open default netns %q: %v", args.Netns, err)
	}
	defer defaultNs.Close()

	//execute within container netns:
	if err := containerNs.Do(func(_ ns.NetNS) error {

		//get netdev from name
		device, err := netlink.LinkByName(cfg.Device)
		if err != nil {
			return fmt.Errorf("cmdDel(): failed to find device %q in containerNS: %v", cfg.Device, err)
		}

		//move netdev from container to default netns
		if err = netlink.LinkSetNsFd(device, int(defaultNs.Fd())); err != nil {
			return fmt.Errorf("cmdDel(): failed to move %q to host netns: %v", device.Attrs().Alias, err)
		}

		return nil
	}); err != nil {
		return err
	}

	//cleanup IPAM config on netdev
	if cfg.IPAM.Type != "" {
		if err := ipam.ExecDel(cfg.IPAM.Type, args.StdinData); err != nil {
			return err
		}
	}

	//cleanup BPF config on netdev
	bpf.Cleanbpf(cfg.Device) //TODO BPF should return error
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

func configureIPAM(args *skel.CmdArgs, cfg *NetConf, device netlink.Link, netns ns.NetNS) (*current.Result, error) {
	var result *current.Result

	//get IPAM
	ipamResult, err := ipam.ExecAdd(cfg.IPAM.Type, args.StdinData)
	if err != nil {
		return result, err
	}

	//delete IPAM incase of error, prevent IP leak
	defer func() {
		if err != nil {
			ipam.ExecDel(cfg.IPAM.Type, args.StdinData)
		}
	}()

	//convert IPAM result into current result type
	result, err = current.NewResultFromResult(ipamResult)
	if err != nil {
		return result, err
	}

	if len(result.IPs) == 0 {
		return result, fmt.Errorf("ipamConfig(): IPAM plugin returned no IPs")
	}

	result.Interfaces = []*current.Interface{{
		Name:    device.Attrs().Name,
		Mac:     device.Attrs().HardwareAddr.String(),
		Sandbox: netns.Path(),
	}}
	for _, ipc := range result.IPs {
		ipc.Interface = current.Int(0)
	}

	//execute within container netns:
	if err := netns.Do(func(_ ns.NetNS) error {

		//set device IP
		if err := ipam.ConfigureIface(device.Attrs().Name, result); err != nil {
			return fmt.Errorf("ipamConfig(): Error setting IPAM on device %q: %v", device.Attrs().Name, err)
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
