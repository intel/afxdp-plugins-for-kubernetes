/*
 Copyright(c) 2021 Intel Corporation.
 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package bpf

//#include <bpf/bpf.h>
//#include <bpf/libbpf.h>
//#include <bpf/xsk.h>
//#include <bpf/xsk.h>
//#include <stdlib.h>
//#cgo CFLAGS: -I.
//#cgo LDFLAGS: -L. -lwrapper -lbpf
//#include "wrapper.h"
//#include "log.h"
import "C"

import (
	"github.com/intel/cndp_device_plugin/pkg/logging"
)

var logDebug = int(C.Get_log_debug())
var logInfo = int(C.Get_log_info())
var logWarn = int(C.Get_log_warn())
var logError = int(C.Get_log_error())
var logPanic = int(C.Get_log_panic())

/*
LoadBpfSendXskMap is the GoLang wrapper for the C function Load_bpf_send_xsk_map
*/
func LoadBpfSendXskMap(ifname string) int {
	cs := C.CString(ifname)
	fd := int(C.Load_bpf_send_xsk_map(cs))
	return fd
}

/*
Cleanbpf is the GoLang wrapper for the C function Clean_bpf
*/
func Cleanbpf(ifname string) {
	cs := C.CString(ifname)
	C.Clean_bpf(cs)
}

/*
GoLogger is exported to C, so C code can write logs back to our main log
*/
//export GoLogger
func GoLogger(cString *C.char, level int) {
	goString := C.GoString(cString)

	switch level {
	case logDebug:
		logging.Debugf(goString)
	case logInfo:
		logging.Infof(goString)
	case logWarn:
		logging.Warningf(goString)
	case logError:
		logging.Errorf(goString)
	case logPanic:
		logging.Panicf(goString)
	default:
		logging.Errorf("ERROR: Unrecognised log level")
	}
}
