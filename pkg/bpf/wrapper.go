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

//import (

//	"fmt"
//)

func Load_bpf_send_xsk_map(ifname string) {
	cs := C.CString(ifname)
	C.load_bpf_send_xsk_map(cs)
	//afterwards the FD should be returned here to be used by dp
	//	return C.fd;
}

//func main (){
//	fmt.Println("hello");
//	load_bpf_send_xsk_map("ens786f3");
//}
