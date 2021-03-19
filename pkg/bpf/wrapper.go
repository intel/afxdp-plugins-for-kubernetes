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

func Load_bpf_send_xsk_map(ifname string) {
	cs := C.CString(ifname)
	C.load_bpf_send_xsk_map(cs)
}

func Cleanbpf(ifname string) {
	cs := C.CString(ifname)
	C.cleanbpf(cs)
}

