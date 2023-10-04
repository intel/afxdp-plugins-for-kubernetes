package main

import (
	"C"
	"fmt"
	"os"

	"github.com/intel/afxdp-plugins-for-kubernetes/internal/uds"
	"github.com/intel/afxdp-plugins-for-kubernetes/pkg/goclient"
)

func main() {
	// Needed for cgo to generate the .h
}

var cleaner uds.CleanupFunc

/*
GetClientVersion is an exported version for c of the goclient GetClientVersion()
*/
//export GetUdsClientVersion
func GetUdsClientVersion() *C.char {
	return C.CString(goclient.GetClientVersion())
}

/*
ServerVersion is an exported version for c of the goclient GetServerVersion()
*/
//export GetUdsServerVersion
func GetUdsServerVersion() *C.char {
	response, function, err := goclient.GetServerVersion()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		function()
		return C.CString("-1")
	}

	cleaner = function

	return C.CString(response)
}

/*
GetXskMapFd is an exported version for c of the goclient XskMapFd()
*/
//export RequestXskMapFd
func RequestXskMapFd(device *C.char) (fd C.int) {
	if device != nil {
		fdVal, function, err := goclient.RequestXSKmapFD(C.GoString(device))
		fd = C.int(fdVal)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			function()
			return -1
		}

		cleaner = function
		return fd
	}

	return -1
}

/*
RequestBusyPoll is an exported version for c of the goclient RequestBusyPoll()
*/
//export RequestBusyPoll
func RequestBusyPoll(busyTimeout, busyBudget, fd C.int) C.int {
	timeout, budget, fdInt := int(busyTimeout), int(busyBudget), int(fd)
	if timeout > -1 && budget > -1 && fdInt > -1 {
		function, err := goclient.RequestBusyPoll(timeout, budget, fdInt)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			function()
			return -1
		}
		cleaner = function
		return 0
	}
	return -1
}

/*
CleanUpConnection an explicit exported cgo function to cleanup a connection after calling any of the other functions.
Pass in one of the available function names to clean up the connection after use.
*/
//export CleanUpConnection
func CleanUpConnection() {
	if cleaner == nil {
		fmt.Println("No cleanup function available to call")
	} else {
		cleaner()
	}
}
