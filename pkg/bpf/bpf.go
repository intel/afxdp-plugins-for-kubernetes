package bpf

// #include "bpf.h"
import "C"

func LoadBpfProgram() {
	C.LoadBpfProgram()
}
