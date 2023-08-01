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
GetClientVersion is an exported version for c of goclient's GetClientVersion()
*/
//export GetUdsClientVersion
func GetUdsClientVersion() *C.char {
	return C.CString(goclient.GetClientVersion())
}

/*
ServerVersion is an exported version for c of goclient's GetServerVersion()
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
GetXskMapFd is an exported version for c of goclient's XskMapFd()
*/
//export RequestXskMapFd
func RequestXskMapFd(device *C.char) (fd C.int) {
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

/*
RequestBusyPoll is an exported version for c of goclient's RequestBusyPoll()
*/
//export RequestBusyPoll
func RequestBusyPoll(busyTimeout, busyBudget, fd C.int) C.int {
	function, err := goclient.RequestBusyPoll(int(busyTimeout), int(busyBudget), int(fd))
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		function()
		return -1
	}
	cleaner = function
	return 0
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
