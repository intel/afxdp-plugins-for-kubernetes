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
//	"net"

	"github.com/pkg/errors"
	logging "github.com/sirupsen/logrus"
	"github.com/intel/afxdp-plugins-for-kubernetes/internal/bpf"
)

var (
	BridgeName     = "afxdp-kind-br"
	vEthNamePrefix = "veth"
	bridgeIP       = "192.168.1.1/24"
)

// Create a Kind Network
func CreateKindNetwork(numVeths, offset int) error {

	kindNetworkExists, _ := CheckKindNetworkExists()
	if kindNetworkExists == true {
		logging.Infof("Bridge %s already exists deleting at as we don't know it's current state", BridgeName)
		err := DeleteKindNetwork(numVeths, offset)
		if err != nil {
			return errors.Wrapf(err, "Error Creating bridge %s", err.Error())
		}
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

		/* Attach one end of the veth pair to the bridge */
		err = Attach(b, vPeer)
		if err != nil {
			return errors.Wrapf(err, "Error attaching veth %s peer %s to bridge %s", veth.Attrs().Name, b.Attrs().Name)
		}
		logging.Infof("Attached veth %s to bridge %s", veth.Attrs().Name, b.Attrs().Name)

		/* Load the xdp-pass program on that peer */
		bh := bpf.NewHandler()
		err = bh.LoadAttachBpfXdpPass(vPeer)
		if err != nil {
			return errors.Wrapf(err, "Error loading xdp-pass Program on interface %s", vPeer)
		}
		logging.Infof("xdp-pass program loaded on: %s", vPeer)
	}

	err = IPAddrAdd(b, bridgeIP)
	if err != nil {
		return errors.Wrapf(err, "Setting the ip address for %s", b.Attrs().Name)
	}

	// TODO SET BRIDGE to up state

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
		v, err := GetVethByName(vName)
		if err != nil {
			return errors.Wrapf(err, "failed to find veth %s", vName)
		}
		err = DeleteVeth(v)
		if err != nil {
			return errors.Wrapf(err, "Error Deleting veth %s", vName)
		}
	}

	logging.Info("Deleted kind secondary network")

	return nil
}
