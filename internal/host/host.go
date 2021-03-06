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

package host

import (
	"errors"
	logging "github.com/sirupsen/logrus"
	"io/ioutil"
	"os/exec"
	"strings"
)

/*
Handler is the CNI and device plugins interface to the host.
The interface exists for testing purposes, allowing unit tests to test
against a fake API.
*/
type Handler interface {
	AllowsUnprivilegedBpf() (bool, error)
	KernelVersion() (string, error)
	HasEthtool() (bool, string, error)
	HasLibbpf() (bool, []string, error)
	HasDevlink() (bool, string, error)
}

/*
handler implements the Handler interface.
*/
type handler struct{}

/*
NewHandler returns an implementation of the Handler interface.
*/
func NewHandler() Handler {
	return &handler{}
}

/*
KernelVersion checks the host kernel version and returns it as a string.
It executes the command: uname -r and returns the output.
*/
func (r *handler) KernelVersion() (string, error) {
	app := "uname"
	args := "-r"

	output, err := exec.Command(app, args).Output()
	if err != nil {
		logging.Errorf("Error getting host kernel version: %v", err.Error())
		return "", err
	}

	kernel := strings.Split(string(output), "\n")[0]

	return kernel, nil
}

/*
HasLibbpf checks if the host has libbpf installed and returns a boolean.
It also returns a string array of libbpf libraries found under /usr/lib(64)/
*/
func (r *handler) HasLibbpf() (bool, []string, error) {
	libPaths := []string{"/usr/lib/", "/usr/lib64/"}
	foundLibbpf := false
	var foundLibs []string

	for _, path := range libPaths {
		files, err := ioutil.ReadDir(path)
		if err != nil {
			if strings.Contains(err.Error(), "no such file or directory") {
				logging.Debugf("Directory " + path + " does not exist")
			} else {
				logging.Errorf("Error checking path "+path+": %v", err)
				return false, nil, err
			}
		}

		for _, file := range files {
			if strings.Contains(file.Name(), "libbpf.so") {
				foundLibbpf = true
				foundLibs = append(foundLibs, path+file.Name())
			}
		}
	}

	if foundLibbpf {
		return true, foundLibs, nil
	}
	return false, nil, nil
}

/*
AllowsUnprivilegedBpf checks if the host allows unpriviliged bpf calls and
returns a boolean. It executes the command: sysctl kernel.unprivileged_bpf_disabled
and returns a boolean value based on the output ("0", "1", "2").
*/
func (r *handler) AllowsUnprivilegedBpf() (bool, error) {
	app := "sysctl"
	args := "kernel.unprivileged_bpf_disabled"

	output, err := exec.Command(app, args).Output()
	if err != nil {
		logging.Errorf("Error checking if host allows unprivileged BPF: %v", err.Error())
		return false, err
	}

	unprivBpfInfo := strings.Split(string(output), "\n")[0]
	unprivBpfStatus := strings.Split(unprivBpfInfo, " = ")[1]

	if unprivBpfStatus != "0" {
		return false, nil
	}

	return true, nil
}

/*
HasEthtool checks if the host has ethtool installed and returns a boolean.
It also executes the command: ethtool --version and returns the version as
a string.
*/
func (r *handler) HasEthtool() (bool, string, error) {
	app := "ethtool"
	args := "--version"

	path, err := exec.LookPath(app)
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return false, "", nil
		}
		logging.Errorf("Error checking if ethtool is present: %v", err)
		return false, "", err
	}
	if path == "" {
		return false, "", nil
	}

	output, err := exec.Command(app, args).Output()
	if err != nil {
		logging.Errorf("Error getting ethtool version: %v", err.Error())
		return false, "", err
	}

	version := strings.Split(string(output), "\n")[0]
	return true, version, nil
}

/*
HasDevlink checks if the host has devlink installed and returns a boolean.
It also executes the command: devlink -Version and returns the version as
a string.
*/
func (r *handler) HasDevlink() (bool, string, error) {
	app := "devlink"
	args := "-Version"

	path, err := exec.LookPath(app)
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return false, "", nil
		}
		logging.Errorf("Error checking if devlink is present: %v", err)
		return false, "", err
	}
	if path == "" {
		return false, "", nil
	}

	output, err := exec.Command(app, args).Output()
	if err != nil {
		logging.Errorf("Error getting devlink version: %v", err.Error())
		return false, "", err
	}

	version := strings.Split(string(output), "\n")[0]
	return true, version, nil
}
