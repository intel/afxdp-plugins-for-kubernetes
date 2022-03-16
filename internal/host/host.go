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
	"github.com/go-cmd/cmd"
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
	HasDevLink() (bool, error)
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

func (r *handler) KernelVersion() (string, error) {
	cmd := cmd.NewCmd("uname", "-r")
	status := <-cmd.Start()

	linuxVer := string(status.Stdout[0])

	return linuxVer, nil
}

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

func (r *handler) AllowsUnprivilegedBpf() (bool, error) {
	cmd := cmd.NewCmd("sysctl", "kernel.unprivileged_bpf_disabled")
	status := <-cmd.Start()

	bpfStatus := strings.Split(status.Stdout[0], "=")[1]
	bpfStatus = strings.TrimSpace(bpfStatus)

	if bpfStatus != "0" {
		return false, nil
	}

	return true, nil
}

func (r *handler) HasEthtool() (bool, string, error) {
	path, err := exec.LookPath("ethtool")
	if err != nil {
		logging.Errorf("Error checking ethtool: %v", err)
		return false, "", err
	}
	if path == "" {
		return false, "", nil
	}

	cmd := cmd.NewCmd("ethtool", "--version")
	status := <-cmd.Start()
	version := string(status.Stdout[0])

	return true, version, nil
}

func (r *handler) HasDevLink() (bool, error) {
	path, err := exec.LookPath("devlink")
	if err != nil {
		logging.Errorf("Error checking devlink: %v", err)
		return false, err
	}
	if path == "" {
		return false, nil
	}
	return true, nil
}
