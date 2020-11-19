package bpf

// #include "bpf.h"
import "C"

/*
LoadBpfProgram is the GoLang wrapper the equivalent C function
*/
func LoadBpfProgram() {
	C.LoadBpfProgram()
}
