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
	logging "github.com/sirupsen/logrus"
	"os/exec"
	"strings"
)

var ethtool = "ethtool"

/*
SetEthtool applies ethtool filters on the physical device during cmdAdd().
Ethtool filters are set via the DP config.json file.
*/
func (r *handler) SetEthtool(ethtoolFilters []string, interfaceName string, ipAddr string) error {
	fd := "on"
	err := flowDirector(interfaceName, fd)
	if err != nil {
		logging.Errorf("Failed to enable flow director: %s", err.Error())
		return err
	}
	for _, ethtoolFilter := range ethtoolFilters {
		ethtoolFilter = strings.Replace(ethtoolFilter, "-device-", interfaceName, -1)

		ethtoolFilter = strings.Replace(ethtoolFilter, "-ip-", ipAddr, -1)

		cmd := exec.Command(ethtool, strings.Split(ethtoolFilter, " ")...)
		stdout, err := cmd.CombinedOutput()
		if err != nil {
			logging.Errorf("Error setting ethtool filter [%s]: %s", ethtoolFilter, string(stdout))
			return err
		}

		logging.Debugf("Ethtool filters [%s] successfully executed", ethtoolFilter)
	}

	return nil
}

/*
DeleteEthtool sets the default queue size ethtool filter.
It also removes perfect-flow ethtool filter entries during cmdDel()
*/
func (r *handler) DeleteEthtool(interfaceName string) error {
	defaultArg := "-X"
	fd := "off"

	ethtoolFilter := []string{defaultArg, interfaceName, "default"}
	cmd := exec.Command(ethtool, ethtoolFilter...)
	stdout, err := cmd.CombinedOutput()
	if err != nil {
		logging.Errorf("Error setting default ethtool queue size [%s]: %v", ethtoolFilter, string(stdout))
		return err
	}

	err = flowDirector(interfaceName, fd)
	if err != nil {
		logging.Errorf("Error removing perfect flow entries: %v", err.Error())
		return err
	}

	logging.Debugf("Ethtool filters removed on device: %s", interfaceName)

	return nil
}

/*
flowDirector enables and disables the Ethernet Flow Director. It must be enabled
for filter flow entries. Disabling, enables entries to be removed from device.
*/
func flowDirector(interfaceName string, fdStatus string) error {
	final := exec.Command(ethtool, "--features", interfaceName, "ntuple", fdStatus)
	_, err := final.CombinedOutput()
	if err != nil {
		return err
	}
	return nil
}
