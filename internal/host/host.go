package host

import (
	"github.com/go-cmd/cmd"
	logging "github.com/sirupsen/logrus"
	"io/ioutil"
	"os/exec"
	"strconv"
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
	HasEthtool() (bool, error)
	HasLibbpf() (bool, error)
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

func (r *handler) HasEthtool() (bool, error) {
	path, err := exec.LookPath("ethtool")
	if err != nil {
		logging.Errorf("Error checking ethtool: %v", err)
		return false, err
	}
	if path == "" {
		return false, nil
	}
	return true, nil
}

func (r *handler) HasLibbpf() (bool, error) {
	libPaths := []string{"/usr/lib/", "/usr/lib64/"}

	for _, path := range libPaths {
		files, err := ioutil.ReadDir(path)
		if err != nil {
			logging.Errorf("Error checking path "+path+": %v", err)
			return false, err
		}

		for _, file := range files {
			if strings.Contains(file.Name(), "libbpf.so") {
				return true, nil
			}
		}
	}

	return false, nil
}

func (r *handler) AllowsUnprivilegedBpf() (bool, error) {
	cmd := cmd.NewCmd("sysctl", "kernel.unprivileged_bpf_disabled")
	status := <-cmd.Start()

	bpfStatus := strings.Split(status.Stdout[0], "=")[1]
	bpfStatus = strings.TrimSpace(bpfStatus)
	boolValue, err := strconv.ParseBool(bpfStatus)

	return !boolValue, err
}
