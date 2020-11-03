package main


import (

	"github.com/intel/cndp_device_plugin/pkg/bpf"
)


func main() {
	bpf.LoadBpfProgram()
}

//TODO main should read some config to determine the devices and device pools
//	The main device plugin code should go into another file, because we may need to run multiple instances for different pools
//	Sililar to SRIOV-DP or the userspace device plugin POC where we have different pools - requirement from Maryam
