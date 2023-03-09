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

func CreateVeth(name, PeerName string) (*netlink.Veth, error) {

	v := &netlink.Veth{
		LinkAttrs: netlink.LinkAttrs{
			Name:  name,
			Flags: net.FlagUp,
		},
		PeerName: PeerName,
	}

	err := netlink.LinkAdd(v)
	if errors.Is(err, syscall.EEXIST) {
		logging.Infof("veth interface already exists. It will be recreated")

		// Delete the existing interface and re-create it.
		if err = netlink.LinkDel(v); err != nil {
			return nil, errors.Wrap(err, "failed to delete the existing veth interface")
		}

		if err = netlink.LinkAdd(v); err != nil {
			return nil, errors.Wrap(err, "failed to re-create the the veth interface")
		}
	}

	if err != nil {
		return nil, errors.Wrap(err, "failed to create the the veth interface")
	}

	logging.Infof("Successfully created veth pair %s %s", v.Name, v.PeerName)

	return v, nil
}

func DeleteVeth(v *netlink.Veth) error {
	_, err := netlink.LinkByName(v.Attrs().Name)
	if err != nil {
		return errors.Wrap(err, "failed to find veth interface to delete")
	}

	return netlink.LinkDel(v)
}

func GetVethByName(n string) (*netlink.Veth, error) {
	link, err := netlink.LinkByName(n)
	if err != nil {
		return nil, errors.Wrapf(err, "Didn't find the veth %s", n)
	}
	veth, ok := link.(*netlink.Veth)
	if !ok {
		return nil, errors.Wrapf(err, "Interface %s is not a veth", n)
	}
	return veth, nil

}

func GetPeer(v *netlink.Veth) (*netlink.Link, error) {
	p, err := netlink.LinkByName(v.PeerName)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to find veth peer %s", v.PeerName)
	}

	return &p, nil
}

func CheckVethExists(name string) (bool, error) {
	_, err := netlink.LinkByName(name)
	if err != nil {
		return false, errors.Wrapf(err, "failed to find veth %s", name)
	}

	return true, nil
}
