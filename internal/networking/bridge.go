/*
 * Copyright(c) Red Hat Inc.
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
	"syscall"

	"github.com/pkg/errors"
	logging "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
)

type Bridge interface {
	Attach(*netlink.Link) error
	GetPorts() []*netlink.Link
	IPAddrAdd(ip *net.IPNet) error
}

func NewBridge(name string) (*netlink.Bridge, error) {

	b := &netlink.Bridge{
		LinkAttrs: netlink.LinkAttrs{
			Name:  name,
			Flags: net.FlagUp,
		},
	}

	err := netlink.LinkAdd(b)
	if errors.Is(err, syscall.EEXIST) {
		logging.Infof("Bridge already exists. It will be recreated")

		// Delete the existing bridge and re-create it.
		if err = netlink.LinkDel(b); err != nil {
			return nil, errors.Wrap(err, "failed to delete the existing Bridge interface")
		}

		if err = netlink.LinkAdd(b); err != nil {
			return nil, errors.Wrap(err, "failed to re-create the the Bridge interface")
		}
	}

	if err != nil {
		return nil, errors.Wrap(err, "failed to create the the Bridge interface")
	}

	logging.Infof("Successfully created Bridge pair %s", b.Attrs().Name)

	return b, nil
}

func DelBridge(BridgeName string) error {
	b,  err  := netlink.LinkByName(BridgeName)
	if err != nil {
		return errors.Wrap(err, "failed to find bridge")
	}

	return netlink.LinkDel(b)
}

func GetBridgeByName(n string) (*netlink.Bridge, error) {
	link, err := netlink.LinkByName(n)
	if err != nil {
		return nil, errors.Wrapf(err, "Didn't find the Bridge %s", n)
	}
	b, ok := link.(*netlink.Bridge)
	if !ok {
		return nil, errors.Wrapf(err, "Interface %s is not a Bridge", n)
	}
	return b, nil

}

func CheckBridgeExists(name string) (bool, error) {
	_, err := netlink.LinkByName(name)
	if err != nil {
		return false, errors.Wrap(err, "failed to find bridge")
	}

	return true, nil
}

func Attach(b *netlink.Bridge, name string) error {
	l, err := netlink.LinkByName(name)
	if err != nil {
		return errors.Wrap(err, "failed to find interface to attach to bridge")
	}

	return netlink.LinkSetMaster(l, b)
}

func IPAddrAdd(b *netlink.Bridge, ip *net.IPNet) error {

	var addr *netlink.Addr
	addr.IPNet = ip

	return netlink.AddrAdd(b, addr)
}
