package bpf

//#include <bpf/bpf.h>
//#include <bpf/libbpf.h>
//#include <bpf/xsk.h>
//#include <bpf/xsk.h>
//#include <stdlib.h>
//#cgo CFLAGS: -I.
//#cgo LDFLAGS: -L. -lwrapper -lbpf
//#include "wrapper.h"
import "C"

import (
	"github.com/golang/glog"
)

var LOG_INFO = int(C.GET_LOG_INFO())
var LOG_WARN = int(C.GET_LOG_WARN())
var LOG_ERROR = int(C.GET_LOG_ERROR())

/*
LoadBpfSendXskMap is the GoLang wrapper for the C function load_bpf_send_xsk_map
*/
func LoadBpfSendXskMap(ifname string) {
	cs := C.CString(ifname)
	C.load_bpf_send_xsk_map(cs)
}

/*
Cleanbpf is the GoLang wrapper for the C function cleanbpf
*/
func Cleanbpf(ifname string) {
	cs := C.CString(ifname)
	C.cleanbpf(cs)
}

//export cLogger
func cLogger(cString *C.char, level int) {
	goString := C.GoString(cString)

	switch level {
	case LOG_INFO:
		glog.Info("INFO: " + goString)
	case LOG_WARN:
		glog.Warning("WARNING: " + goString)
	case LOG_ERROR:
		glog.Error("ERROR: " + goString)
	default:
		glog.Error("ERROR: Unrecognised log level")
	}
}
