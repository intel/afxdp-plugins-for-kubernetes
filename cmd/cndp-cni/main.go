// Copyright 2020 CNI authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"runtime"

	"github.com/vishvananda/netlink"

	"github.com/golang/glog"

	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
	"github.com/containernetworking/cni/pkg/types/current"
	"github.com/containernetworking/cni/pkg/version"
	"github.com/intel/cndp_device_plugin/pkg/bpf"

	"github.com/containernetworking/plugins/pkg/ip"
	"github.com/containernetworking/plugins/pkg/ipam"
	"github.com/containernetworking/plugins/pkg/ns"
	bv "github.com/containernetworking/plugins/pkg/utils/buildversion"
)

const (
	sysBusPCI = "/sys/bus/pci/devices"
)

//NetConf for host-device config
type NetConf struct {
	types.NetConf
	Device string `json:"deviceID"`
}

func init() {
	// this ensures that main runs only on main thread (thread group leader).
	// since namespace ops (unshare, setns) are done for a single thread, we
	// must ensure that the goroutine does not jump from OS thread to thread
	runtime.LockOSThread()
	glog.Info("Initializing CNDP CNI device plugin...")
}

func loadConf(bytes []byte) (*NetConf, error) {
	n := &NetConf{}
	if err := json.Unmarshal(bytes, n); err != nil {
		return nil, fmt.Errorf("loadConf(): loading network configuration unsuccessful : %v", err)
	}

	if n.Device == "" {
		return nil, fmt.Errorf(`loadConf(): specify a "device" - field blank`)
	}

	glog.Info("loadConf(): Device Name was successfully given")
	return n, nil

}

func cmdAdd(args *skel.CmdArgs) error {
	cfg, err := loadConf(args.StdinData)
	if err != nil {
		return err
	}
	containerNs, err := ns.GetNS(args.Netns)
	if err != nil {
		return fmt.Errorf("cmdAdd(): failed to open netns %q: %v and load configuration file ", args.Netns, err)
	}
	defer containerNs.Close()

	hostDev, err := getLink(cfg.Device)
	if err != nil {
		return fmt.Errorf("cmdAdd(): failed to find host device: %v", err)
	}

	contDev, err := moveLinkIn(hostDev, containerNs, args.IfName)
	if err != nil && args.IfName == " " {
		return fmt.Errorf("cmdAdd(): failed to move link %v to host container namespace from host namespace", err)
	}

	var result *current.Result
	// run the IPAM plugin and get back the config to apply
	if cfg.IPAM.Type != "" {
		r, err := ipam.ExecAdd(cfg.IPAM.Type, args.StdinData)
		if err != nil {
			return err
		}

		// Invoke ipam del if err to avoid ip leak
		defer func() {
			if err != nil {
				ipam.ExecDel(cfg.IPAM.Type, args.StdinData)
			}
		}()

		// Convert whatever the IPAM result was into the current Result type
		result, err = current.NewResultFromResult(r)
		if err != nil {
			return err
		}

		if len(result.IPs) == 0 {
			return errors.New("IPAM plugin returned missing IP config")
		}

		result.Interfaces = []*current.Interface{{
			Name:    contDev.Attrs().Name,
			Mac:     contDev.Attrs().HardwareAddr.String(),
			Sandbox: containerNs.Path(),
		}}
		for _, ipc := range result.IPs {
			// All addresses apply to the container interface (move from host)
			ipc.Interface = current.Int(0)
		}

		err = containerNs.Do(func(_ ns.NetNS) error {
			if err := ipam.ConfigureIface(args.IfName, result); err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return err
		}

		result.DNS = cfg.DNS

		return types.PrintResult(result, cfg.CNIVersion)
	}

	return printLink(contDev, cfg.CNIVersion, containerNs)
}

func cmdDel(args *skel.CmdArgs) error {
	cfg, err := loadConf(args.StdinData)
	if err != nil {
		return err
	}
	if args.Netns == "" {
		return nil
	}
	containerNs, err := ns.GetNS(args.Netns)
	if err != nil {
		return fmt.Errorf("cmdDel(): failed to open netns %q: %v", args.Netns, err)
	}
	defer containerNs.Close()

	if err := moveLinkOut(containerNs, args.IfName); err != nil {
		return err
	}

	if cfg.IPAM.Type != "" {
		if err := ipam.ExecDel(cfg.IPAM.Type, args.StdinData); err != nil {
			return err
		}
	}

	bpf.Cleanbpf(args.IfName)
	return nil
}

func moveLinkIn(hostDev netlink.Link, containerNs ns.NetNS, ifName string) (netlink.Link, error) {
	if err := netlink.LinkSetNsFd(hostDev, int(containerNs.Fd())); err != nil {
		return nil, err
	}

	var contDev netlink.Link
	if err := containerNs.Do(func(_ ns.NetNS) error {
		var err error
		contDev, err = netlink.LinkByName(hostDev.Attrs().Name)
		if err != nil {
			return fmt.Errorf("moveLinkIn(): failed to find %q: %v", hostDev.Attrs().Name, err)
		}
		// Devices can be renamed only when down
		if err = netlink.LinkSetDown(contDev); err != nil {
			return fmt.Errorf("moveLinkIn(): failed to set %q down: %v", hostDev.Attrs().Name, err)
		}
		// Save host device name into the container device's alias property
		if err := netlink.LinkSetAlias(contDev, hostDev.Attrs().Name); err != nil {
			return fmt.Errorf("moveLinkIn(): failed to set alias to %q: %v", hostDev.Attrs().Name, err)
		}
		// Rename container device to respect args.IfName
		if err := netlink.LinkSetName(contDev, ifName); err != nil {
			return fmt.Errorf("moveLinkIn(): failed to rename device %q to %q: %v", hostDev.Attrs().Name, ifName, err)
		}
		// Retrieve link again to get up-to-date name and attributes
		contDev, err = netlink.LinkByName(ifName)
		if err != nil {
			return fmt.Errorf("moveLinkIn(): failed to find %q: %v", ifName, err)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return contDev, nil
}

func moveLinkOut(containerNs ns.NetNS, ifName string) error {
	defaultNs, err := ns.GetCurrentNS()
	if err != nil {
		return err
	}
	defer defaultNs.Close()

	return containerNs.Do(func(_ ns.NetNS) error {
		dev, err := netlink.LinkByName(ifName)
		if err != nil {
			return fmt.Errorf("moveLinkOut(): failed to find %q: %v", ifName, err)
		}

		// Devices can be renamed only when down
		if err = netlink.LinkSetDown(dev); err != nil {
			return fmt.Errorf("moveLinkOut(): failed to set %q down: %v", ifName, err)
		}

		// Rename device to it's original name
		if err = netlink.LinkSetName(dev, dev.Attrs().Alias); err != nil {
			return fmt.Errorf("moveLinkOut(): failed to restore %q to original name %q: %v", ifName, dev.Attrs().Alias, err)
		}
		defer func() {
			if err != nil {

				_ = netlink.LinkSetName(dev, ifName)
			}
		}()

		if err = netlink.LinkSetNsFd(dev, int(defaultNs.Fd())); err != nil {
			return fmt.Errorf("moveLinkOut(): failed to move %q to host netns: %v", dev.Attrs().Alias, err)
		}
		return nil
	})
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

func getLink(devname string) (netlink.Link, error) {

	if len(devname) > 0 {
		return netlink.LinkByName(devname)
	}

	return nil, fmt.Errorf("getLink(): failed to find physical interface")
}

func main() {
	skel.PluginMain(cmdAdd, cmdCheck, cmdDel, version.All, bv.BuildString("host-device1"))
}

func cmdCheck(args *skel.CmdArgs) error {

	cfg, err := loadConf(args.StdinData)
	if err != nil {
		return err
	}
	netns, err := ns.GetNS(args.Netns)
	if err != nil {
		return fmt.Errorf("cmdCheck(): failed to open netns %q: %v", args.Netns, err)
	}
	defer netns.Close()

	// run the IPAM plugin and get back the config to apply
	if cfg.IPAM.Type != "" {
		err = ipam.ExecCheck(cfg.IPAM.Type, args.StdinData)
		if err != nil {
			return err
		}
	}

	// Parse previous result.
	if cfg.NetConf.RawPrevResult == nil {
		return fmt.Errorf("cmdCheck(): Required prevResult missing")
	}

	if err := version.ParsePrevResult(&cfg.NetConf); err != nil {
		return err
	}

	result, err := current.NewResultFromResult(cfg.PrevResult)
	if err != nil {
		return err
	}

	var contMap current.Interface
	// Find interfaces for name we know, that of host-device inside container
	for _, intf := range result.Interfaces {
		if args.IfName == intf.Name {
			if args.Netns == intf.Sandbox {
				contMap = *intf
				continue
			}
		}
	}

	// The namespace must be the same as what was configured
	if args.Netns != contMap.Sandbox {
		return fmt.Errorf("cmdCheck(): Sandbox in prevResult %s doesn't match configured netns: %s",
			contMap.Sandbox, args.Netns)
	}

	//
	// Check prevResults for ips, routes and dns against values found in the container
	if err := netns.Do(func(_ ns.NetNS) error {

		// Check interface against values found in the container
		err := validateCniContainerInterface(contMap)
		if err != nil {
			return err
		}

		err = ip.ValidateExpectedInterfaceIPs(args.IfName, result.IPs)
		if err != nil {
			return err
		}

		err = ip.ValidateExpectedRoute(result.Routes)
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}

	//
	return nil
}

func validateCniContainerInterface(intf current.Interface) error {

	var link netlink.Link
	var err error

	if intf.Name == "" {
		return fmt.Errorf("validateCniContainerInterface(): Container interface name missing in prevResult: %v", intf.Name)
	}
	link, err = netlink.LinkByName(intf.Name)
	if err != nil {
		return fmt.Errorf("validateCniContainerInterface(): Container Interface name in prevResult: %s not found", intf.Name)
	}
	if intf.Sandbox == "" {
		return fmt.Errorf("Error: validateCniContainerInterface(): Container interface %s should not be in host namespace", link.Attrs().Name)
	}

	if intf.Mac != "" {
		if intf.Mac != link.Attrs().HardwareAddr.String() {
			return fmt.Errorf("validateCniContainerInterface(): Interface %s Mac %s doesn't match container Mac: %s", intf.Name, intf.Mac, link.Attrs().HardwareAddr)
		}
	}

	return nil
}
