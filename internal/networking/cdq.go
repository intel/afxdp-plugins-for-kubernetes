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

package networking

import (
	"fmt"
	logging "github.com/sirupsen/logrus"
	"os/exec"
	"strconv"
	"strings"
)

/*
CreateCdqSubfunction takes the PCI address of a port and a subfunction number
It creates that subfunction on top of that port and activates it
*/
func (r *handler) CreateCdqSubfunction(parentPci string, sfnum string) error {
	app := "devlink"
	args := []string{"port", "add", "pci/" + parentPci, "flavour", "pcisf", "pfnum", "0", "sfnum", sfnum}

	output, err := exec.Command(app, args...).Output()
	if err != nil {
		logging.Errorf("Error creating sub-function %s on pci %s: %v", sfnum, parentPci, err.Error())
		return err
	}

	portIndex := strings.Split(string(output), ": ")[0]
	args = []string{"port", "function", "set", portIndex, "state", "active"}

	_, err = exec.Command(app, args...).Output()
	if err != nil {
		logging.Errorf("Error activating sub-function %s on pci %s: %v", sfnum, parentPci, err.Error())
		return err
	}

	return nil
}

/*
DeleteCdqSubfunction takes the port index of a subfunction, deactivates and deletes it
*/
func (r *handler) DeleteCdqSubfunction(portIndex string) error {
	app := "devlink"
	args := []string{"port", "function", "set", "pci/" + portIndex, "state", "inactive"}

	_, err := exec.Command(app, args...).Output()
	if err != nil {
		logging.Errorf("Error setting sub-function inactive %s: %v", portIndex, err.Error())
		return err
	}

	args = []string{"port", "del", "pci/" + portIndex}

	_, err = exec.Command(app, args...).Output()
	if err != nil {
		logging.Errorf("Error deleting sub-function %s: %v", portIndex, err.Error())
		return err
	}

	return nil
}

/*
IsCdqSubfunction takes a netdev name and returns true if is a CDQ subfunction.
*/
func (r *handler) IsCdqSubfunction(name string) (bool, error) {
	portIndex, err := r.GetCdqPortIndex(name)
	if err != nil {
		return false, err
	}

	split := strings.Split(portIndex, "/")
	index := split[1]

	if index == "0" {
		return false, nil
	}
	return true, nil
}

/*
GetCdqPortIndex takes a netdev name and returns the port index (pci/sfnum)
Note this function only works on physical devices and CDQ subfunctions
Other netdevs will return a "device not found by devlink" error
*/
func (r *handler) GetCdqPortIndex(netdev string) (string, error) {
	devlinkList := "devlink port list | grep " + `"\b` + netdev + `\b"`

	devList, err := exec.Command("sh", "-c", devlinkList).CombinedOutput()
	if err != nil {
		if strings.Contains(err.Error(), "exit status 1") {
			return "", fmt.Errorf("device %s not found by devlink (1)", netdev)
		}
		return "", err
	}

	if devList != nil {
		portIndex := strings.Fields(string(devList))[0]

		pciSplit := strings.Split(portIndex, "pci/")
		portIndexAddress := pciSplit[1]

		lastInd := strings.LastIndex(portIndexAddress, ":")
		portAddrIndex := portIndexAddress[:lastInd]
		return portAddrIndex, nil
	}

	return "", fmt.Errorf("device %s not found by devlink (2)", netdev)
}

/*
NumAvailableCdqSubfunctions takes the PCI of a physical port and returns how
many unused CDQ subfunctions are available
*/
func (r *handler) NumAvailableCdqSubfunctions(pci string) (int, error) {
	app := "devlink"
	args := []string{"resource", "show", "pci/" + pci}

	resourceInfo, err := exec.Command(app, args...).CombinedOutput()
	if err != nil {
		logging.Errorf("Error getting devlink resource for pci %s: %v", pci, err)
		return 0, err
	}

	lines := strings.Split(string(resourceInfo), "\n")
	totalSFs, err := strconv.Atoi(strings.Fields(lines[3])[3]) //line 3, word 3 - "size"
	if err != nil {
		logging.Errorf("Error converting total available SFs to int %s", err)
		return 0, err
	}
	inUseSFs, err := strconv.Atoi(strings.Fields(lines[3])[5]) //line 3, word 5 - "occ"
	if err != nil {
		logging.Errorf("Error converting in use SFs to int %s", err)
		return 0, err
	}
	return totalSFs - inUseSFs, nil
}
