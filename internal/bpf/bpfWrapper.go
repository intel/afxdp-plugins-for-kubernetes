/*
 * Copyright(c) 2022 Intel Corporation.
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

package bpf

//#include <xdp/libxdp.h>
//#include <xdp/xsk.h>
//#cgo CFLAGS: -I.
//#cgo LDFLAGS: -L. -lxdp -lbpf -lelf -lz
//#include "bpfWrapper.h"
//#include "log.h"
import "C"

import (
	"os/exec"
	"strings"

	"github.com/pkg/errors"
	logging "github.com/sirupsen/logrus"
)

/*
Handler is the interface to the BPF package.
The interface exists for testing purposes, allowing unit tests to run
without making actual BPF calls.
*/
type Handler interface {
	LoadBpfSendXskMap(ifname string) (int, error)
	LoadAttachBpfXdpPass(ifname string) error
	ConfigureBusyPoll(fd int, busyTimeout int, busyBudget int) error
	LoadBpfPinXskMap(ifname, pin_path string) error
	Cleanbpf(ifname string) error
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
LoadBpfSendXskMap is the GoLang wrapper for the C function Load_bpf_send_xsk_map
*/
func (r *handler) LoadBpfSendXskMap(ifname string) (int, error) {
	fd := int(C.Load_bpf_send_xsk_map(C.CString(ifname)))

	if fd <= 0 {
		return fd, errors.New("error loading BPF program onto interface")
	}

	return fd, nil
}

/*
LoadBpfXdpPass is the GoLang wrapper for the C function Load_bpf_send_xsk_map
*/
func (r *handler) LoadAttachBpfXdpPass(ifname string) error {
	bpfProg := "/afxdp/xdp_pass.o"

	if err := XdpLoaderCmd(ifname, "load", bpfProg, ""); err != nil {
		return errors.Wrapf(err, "Couldn't Load %s to interface %s", bpfProg, ifname)
	}

	return nil
}

/*
LoadBpfPinXskMap is the GoLang wrapper for the C function Load_bpf_send_xsk_map
*/
func (r *handler) LoadBpfPinXskMap(ifname, pin_path string) error {

	bpfProg := "/afxdp/xdp_afxdp_redirect.o"

	if err := XdpLoaderCmd(ifname, "load", bpfProg, pin_path); err != nil {
		return errors.Wrapf(err, "Couldn't Load and pin %s to interface %s", bpfProg, ifname)
	}

	return nil
}

func XdpLoaderCmd(ifname, action, bpfProg, pin_path string) error {

	cmd := exec.Command("xdp-loader", "unload", ifname, "--all")

	if err := cmd.Run(); err != nil && err.Error() != "exit status 1" { // exit status 1 means no prog to unload
		logging.Errorf("Error removing BPF program from device: %v", err)
		return errors.New("Error removing BPF program from device")
	}

	if action == "load" {
		loaderArgs := action + " " + ifname + " " + bpfProg
		if pin_path != "" {
			loaderArgs += " -p " + pin_path
		}
		logging.Infof("Loading XDP program using: xdp-loader %s", loaderArgs)

		cmd := exec.Command("xdp-loader", strings.Split(loaderArgs, " ")...)
		if err := cmd.Run(); err != nil {
			logging.Errorf("error loading and pinning BPF program onto interface %v", err)
			return errors.New("error loading and pinning BPF program onto interface")
		}
	}

	return nil
}

/*
ConfigureBusyPoll is the GoLang wrapper for the C function Configure_busy_poll
*/
func (r *handler) ConfigureBusyPoll(fd int, busyTimeout int, busyBudget int) error {
	ret := C.Configure_busy_poll(C.int(fd), C.int(busyTimeout), C.int(busyBudget))

	if ret != 0 {
		return errors.New("error configuring busy poll on interface")
	}

	return nil
}

/*
Cleanbpf is the GoLang wrapper for the C function Clean_bpf
*/
func (r *handler) Cleanbpf(ifname string) error {

	ret := C.Clean_bpf(C.CString(ifname))

	if ret != 0 {
		return errors.New("error removing BPF program from interface")
	}

	return nil
}

// Debugf is exported to C, so C code can write logs to the Golang logging package
//
//export Debugf
func Debugf(msg *C.char) {
	logging.Debugf(C.GoString(msg))
}

// Infof is exported to C, so C code can write logs to the Golang logging package
//
//export Infof
func Infof(msg *C.char) {
	logging.Infof(C.GoString(msg))
}

// Warningf is exported to C, so C code can write logs to the Golang logging package
//
//export Warningf
func Warningf(msg *C.char) {
	logging.Warningf(C.GoString(msg))
}

// Errorf is exported to C, so C code can write logs to the Golang logging package
//
//export Errorf
func Errorf(msg *C.char) {
	logging.Errorf(C.GoString(msg))
}

// Panicf is exported to C, so C code can write logs to the Golang logging package
//
//export Panicf
func Panicf(msg *C.char) {
	logging.Panicf(C.GoString(msg))
}
