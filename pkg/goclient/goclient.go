package goclient

import (
	"fmt"
	"time"

	"github.com/intel/afxdp-plugins-for-kubernetes/constants"
	"github.com/intel/afxdp-plugins-for-kubernetes/internal/host"
	"github.com/intel/afxdp-plugins-for-kubernetes/internal/uds"
)

var (
	hostUds       uds.Handler
	hostPod       host.Handler
	cleanupGlobal uds.CleanupFunc
	connected     bool = false
)

/*
GetClientVersion returns the version of our Handshake from the client
*/
func GetClientVersion() string {
	return constants.Uds.Handshake.Version
}

/*
GetServerVersion returns the version of our Handshake from the server as a response
*/
func GetServerVersion() (string, uds.CleanupFunc, error) {
	if !connected {
		err := initFunc()
		if err != nil {
			return "", cleanupGlobal, err
		}
	}

	if err := hostUds.Write(constants.Uds.Handshake.RequestVersion, -1); err != nil {
		return "", cleanupGlobal, fmt.Errorf("Library Error: Writing Error: %v", err)
	}

	response, _, err := hostUds.Read()
	if err != nil {
		return "", cleanupGlobal, fmt.Errorf("Library Error: Reading Error: %v", err)
	}

	return response,
		cleanupGlobal,
		nil
}

/*
RequestXSKmapFD requires a device name and returns a fds the device, a cleanup function to close the connection, and an error
*/
func RequestXSKmapFD(device string) (int, uds.CleanupFunc, error) {
	if !connected {
		err := initFunc()

		if err != nil {
			return 0, cleanupGlobal, fmt.Errorf("Library Error: Initializing Error: %v", err)
		}
	}

	if err := hostUds.Write(constants.Uds.Handshake.RequestFd+", "+device, -1); err != nil {
		return 0, cleanupGlobal, fmt.Errorf("Library Error: UDS Write error: %v", err)

	}

	response, fd, err := hostUds.Read()
	if err != nil {
		return 0, cleanupGlobal, fmt.Errorf("Library Error: UDS Read error: %v", err)

	}

	if response == constants.Uds.Handshake.ResponseFdAck {
		return fd, cleanupGlobal, nil
	} else {
		return 0, cleanupGlobal, fmt.Errorf("Library Error: Request for FD was not acknowledged")
	}

}

/*
RequestBusyPoll takes a timeout, budget and a fd to request the busypoll for a specific device, and returns an fd, response, cleanup function and error
*/
func RequestBusyPoll(busyTimeout, busyBudget, fd int) (uds.CleanupFunc, error) {
	if !connected {
		err := initFunc()
		if err != nil {
			return cleanupGlobal, fmt.Errorf("Library Error: Failed to initialize UDS error: %v", err)
		}
	}

	pollString := fmt.Sprintf("%s, %d, %d", constants.Uds.Handshake.RequestBusyPoll, busyTimeout, busyBudget)

	if err := hostUds.Write(pollString, fd); err != nil {
		return cleanupGlobal, fmt.Errorf("Library Error: Failed to write to UDS error: %v", err)
	}

	response, _, err := hostUds.Read()
	if err != nil {
		return cleanupGlobal, fmt.Errorf("Library Error: Failed to read UDS error: %v", err)
	}

	if response == constants.Uds.Handshake.ResponseBusyPollNak {
		return cleanupGlobal, fmt.Errorf("Library Error: Device plugin error configuring busy poll")
	}

	return cleanupGlobal, nil
}

/*
initFunc initializes the library, returns a cleanup function and an error
*/
func initFunc() error {
	hostUds = uds.NewHandler()
	hostPod = host.NewHandler()
	var response string

	// init uds Handler for reading and writing
	if err := hostUds.Init(constants.Uds.PodPath, constants.Uds.Protocol, constants.Uds.MsgBufSize, constants.Uds.CtlBufSize, 0*time.Second, ""); err != nil {
		return fmt.Errorf("Library Error: Error Initialising UDS server: %v", err)
	}

	cleanup, err := hostUds.Dial()
	cleanupGlobal = cleanup
	if err != nil {
		return fmt.Errorf("Library Error: UDS Dial error: %v", err)
	}

	hostname, err := hostPod.Hostname()
	if err != nil {
		return fmt.Errorf("Library Error: Failed to initialize host: %v", err)
	}

	if err = hostUds.Write(constants.Uds.Handshake.RequestConnect+", "+hostname, -1); err != nil {
		return fmt.Errorf("Library Error: UDS Write error: %v", err)
	}

	if response, _, err = hostUds.Read(); err != nil {
		return fmt.Errorf("Library Error: UDS Read error : %v", err)
	}

	if response == constants.Uds.Handshake.ResponseHostOk {
		connected = true
	}

	return nil
}
