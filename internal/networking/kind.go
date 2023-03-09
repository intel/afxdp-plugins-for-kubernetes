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
	"strconv"
	"net"

	"github.com/pkg/errors"
	logging "github.com/sirupsen/logrus"
)

var (
	BridgeName     = "afxdp-kind-br"
	vEthNamePrefix = "veth"
	bridgeIP       = "192.168.1.1"
)

// Create a Kind Network
func CreateKindNetwork(numVeths, offset int) error {

	kindNetworkExists, _ := CheckKindNetworkExists()
	if kindNetworkExists == true {
		logging.Infof("Bridge %s already exists skipping recreating it in case it's already in use", BridgeName)
		return nil
	}

	b, err := NewBridge(BridgeName)
	if err != nil {
		return errors.Wrapf(err, "Error Creating bridge %s", err.Error())
	}
	logging.Infof("Created bridge %s", b.Attrs().Name)

	for i := offset; i < ((numVeths * 2) + offset); i = i+2 {
		vName := vEthNamePrefix + strconv.Itoa(i)
		vPeer := vEthNamePrefix + strconv.Itoa(i+1)
		veth, err := CreateVeth(vName, vPeer)
		if err != nil {
			return errors.Wrapf(err, "Error Creating veth pair %s <===> %s", vName, vPeer)
		}

		err = Attach(b, vPeer)
		if err != nil {
			return errors.Wrapf(err, "Error attaching veth %s peer %s to bridge %s", veth.Attrs().Name, b.Attrs().Name)
		}
		logging.Infof("Attached veth %s to bridge %s", veth.Attrs().Name, b.Attrs().Name)
	}

	//TODO configure Bridge IP addr
	var ipNet *net.IPNet
	ipNet.IP = net.ParseIP(bridgeIP)

	err = IPAddrAdd(b, ipNet)
	if err != nil {
		return errors.Wrapf(err, "Setting the ip address for %s", b.Attrs().Name)
	}

	return nil
}

func CheckKindNetworkExists() (bool, error) {
	return CheckBridgeExists(BridgeName)
}

// DeleteKindNetwork
func DeleteKindNetwork(numVeths, offset int) error {
	err := DelBridge(BridgeName)
	if err != nil {
		return errors.Wrap(err, "failed to delete bridge")
	}

	for i := offset; i < numVeths; i = i+2 {
		vName := vEthNamePrefix + strconv.Itoa(i)
		err := DelBridge(vName)
		if err != nil {
			return errors.Wrapf(err, "Error Deleting veth %s", vName)
		}
	}

	logging.Info("Deleted kind secondary network")

	return nil
}
