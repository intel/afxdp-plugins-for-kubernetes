/*
 * Copyright(c) 2021 Intel Corporation.
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

//#include <bpf/bpf.h>
//#include <bpf/libbpf.h>
//#include <bpf/xsk.h>
//#include <bpf/xsk.h>
//#include <stdlib.h>
//#cgo CFLAGS: -I.
//#cgo LDFLAGS: -L. -lwrapper -lbpf
//#include "bpfWrapper.h"
//#include "log.h"
import "C"

import (
	"errors"
	logging "github.com/sirupsen/logrus"
)

/*
Handler is the interface to the BPF package.
The interface exists for testing purposes, allowing unit tests to run
without making actual BPF calls.
*/
type Handler interface {
	LoadBpfSendXskMap(ifname string) (int, error)
	ConfigureBusyPoll(fd int, busyTimeout int, busyBudget int) error
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
		return fd, errors.New("Error loading BPF program onto interface")
	}

	return fd, nil
}

/*
ConfigureBusyPoll is the GoLang wrapper for the C function Configure_busy_poll
*/
func (r *handler) ConfigureBusyPoll(fd int, busyTimeout int, busyBudget int) error {
	ret := C.Configure_busy_poll(C.int(fd), C.int(busyTimeout), C.int(busyBudget))

	if ret != 0 {
		return errors.New("Error configuring busy poll on interface")
	}

	return nil
}

/*
Cleanbpf is the GoLang wrapper for the C function Clean_bpf
*/
func (r *handler) Cleanbpf(ifname string) error {
	ret := C.Clean_bpf(C.CString(ifname))

	if ret != 0 {
		return errors.New("Error removing BPF program from interface")
	}

	return nil
}

//Debugf is exported to C, so C code can write logs to the Golang logging package
//export Debugf
func Debugf(msg *C.char) {
	logging.Debugf(C.GoString(msg))
}

//Infof is exported to C, so C code can write logs to the Golang logging package
//export Infof
func Infof(msg *C.char) {
	logging.Infof(C.GoString(msg))
}

//Warningf is exported to C, so C code can write logs to the Golang logging package
//export Warningf
func Warningf(msg *C.char) {
	logging.Warningf(C.GoString(msg))
}

//Errorf is exported to C, so C code can write logs to the Golang logging package
//export Errorf
func Errorf(msg *C.char) {
	logging.Errorf(C.GoString(msg))
}

//Panicf is exported to C, so C code can write logs to the Golang logging package
//export Panicf
func Panicf(msg *C.char) {
	logging.Panicf(C.GoString(msg))
}
