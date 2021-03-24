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
