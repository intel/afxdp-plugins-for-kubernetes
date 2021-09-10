/*
 Copyright(c) 2021 Intel Corporation.
 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package uds

import (
	"fmt"
	"github.com/intel/cndp_device_plugin/pkg/logging"
	"github.com/nu7hatch/gouuid"
	"net"
	"os"
	"strconv"
	"syscall"
	"time"
)

const (
	fileMode = os.FileMode(0600) // drw-------
)

/*
Handler is the device plugins interface for reading and writing to a Unix domain socket.
The interface exists for testing purposes, allowing unit tests to run without making calls
on a real socket.
*/
type Handler interface {
	Init(protocol string, msgBufSize int, ctlBufSize int, timeout time.Duration) (CancelFunc, error)
	Listen() (CancelFunc, error)
	GetSocketPath() string
	Read() (string, int, error)
	Write(response string, fd int) error
}

/*
handler implements the Handler interface.
*/
type handler struct {
	socket     string
	listener   *net.UnixListener
	conn       *net.UnixConn
	msgBufSize int
	ctlBufSize int
	udsFD      int
	timeout    time.Duration
}

/*
NewHandler returns an implementation of the Handler interface.
*/
func NewHandler(directory string) (Handler, error) {
	socket, err := generateSocketPath(directory)
	if err != nil {
		logging.Errorf("Error generating socket file path: %v", err)
		return &handler{}, err
	}

	handler := &handler{
		socket: socket,
	}
	return handler, nil
}

/*
GetSocketPath returns the socket path that this Handler is serving
*/
func (h *handler) GetSocketPath() string {
	return h.socket
}

/*
CancelFunc defines a function that we return from the Init function.
This function is responsible for proper cleanup of the socket file.
*/
type CancelFunc func()

/*
Init initialises the Unix domain socket and creates a Unix listener
A CancelFunc function is returned. This function should be deferred by the calling code
to ensure proper socket cleanup.
*/
func (h *handler) Init(protocol string, msgBufSize int, ctlBufSize int, timeout time.Duration) (CancelFunc, error) {
	h.msgBufSize = msgBufSize
	h.ctlBufSize = ctlBufSize

	if timeout > 0 { //TODO test and comment //TODO is this if needed? no?
		h.timeout = timeout
	}

	// resolve UDS address
	addr, err := net.ResolveUnixAddr(protocol, h.socket)
	if err != nil {
		logging.Errorf("Error resolving Unix address %s: %v", h.socket, err)
		return func() {}, err
	}

	// create UDS listener
	h.listener, err = net.ListenUnix(protocol, addr)
	if err != nil {
		logging.Errorf("Error creating Unix listener for %s: %v", h.socket, err)
		return func() {
			logging.Debugf("Closing Unix listener")
			h.listener.Close()
			logging.Debugf("Removing socket file")
			os.Remove(h.socket)
		}, err
	}

	if h.timeout > 0 {
		if err := h.listener.SetDeadline(time.Now().Add(h.timeout)); err != nil {
			logging.Errorf("Error setting listener timeout: %v", err)
			return func() {
				logging.Debugf("Closing Unix listener")
				h.listener.Close()
				logging.Debugf("Removing socket file")
				os.Remove(h.socket)
			}, err
		}
	}

	return func() {
		logging.Debugf("Closing Unix listener")
		h.listener.Close()
		logging.Debugf("Removing socket file")
		os.Remove(h.socket)
	}, nil

}

/*
Listen listens for and accepts new connections
A CancelFunc function is returned. This function should be deferred by the calling code
to ensure proper socket cleanup.
*/
func (h *handler) Listen() (CancelFunc, error) {
	var err error
	h.conn, err = h.listener.AcceptUnix()
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			logging.Errorf("Listener timed out: %v", err)
			return func() { logging.Debugf("Closing connection"); h.conn.Close() }, err
		}
		logging.Errorf("Listener Accept error: %v", err)
		return func() { logging.Debugf("Closing connection"); h.conn.Close() }, err
	}

	// get the UDS socket file descriptor, required for syscall.Recvmsg/Sendmsg
	socketFile, err := h.conn.File()
	if err != nil {
		logging.Errorf("Error getting socket file descriptor: %v", err)
		return func() {
			logging.Debugf("Closing connection")
			h.conn.Close()
			logging.Debugf("Closing socket file")
			socketFile.Close()
		}, err
	}
	h.udsFD = int(socketFile.Fd())

	return func() {
		logging.Debugf("Closing connection")
		h.conn.Close()
		logging.Debugf("Closing socket file")
		socketFile.Close()
	}, nil

}

/*
Read will read the incoming message from the UDS
Message byte array is converted and returned as a string
The control messages are also checked and returns the FD as an int, if present
*/
func (h *handler) Read() (string, int, error) {
	var request = ""
	var fd int = 0
	msgBuf := make([]byte, h.msgBufSize)
	ctrlBuf := make([]byte, syscall.CmsgSpace(h.ctlBufSize))

	// set connection timeout
	if h.timeout > 0 {
		if err := h.conn.SetDeadline(time.Now().Add(h.timeout)); err != nil {
			logging.Errorf("Error setting connection timeout: %v", err)
			return request, fd, err
		}
	}

	// read request message
	n, _, _, _, err := syscall.Recvmsg(h.udsFD, msgBuf, ctrlBuf, 0)
	if err != nil {
		logging.Errorf("Recvmsg error: %v", err)
		return request, fd, err
	}

	request = string(msgBuf[0:n])
	logging.Debugf("Request: %s", request)

	if ctrlBufHasValue(ctrlBuf) {
		ctrlMsgs, err := syscall.ParseSocketControlMessage(ctrlBuf)
		if err != nil {
			logging.Errorf("Control messages parse error: %v", err)
			return request, fd, err
		}

		//TODO fmts should be a debug log
		//TODO can new logging package handle %08b

		//fmt.Println("ctrlMsgs:")
		//fmt.Printf("%08b", ctrlMsgs)
		//fmt.Println()

		if len(ctrlMsgs) > 0 {
			//Typically code would loop through ctrlMsgs and fds
			//We're handling a single msg and single fd here, so it's msg[0] fds[0]
			fds, _ := syscall.ParseUnixRights(&ctrlMsgs[0])
			fd = fds[0]
			logging.Debugf("Request contains file descriptor: %s", strconv.Itoa(fd))

			//TODO fmt prints should be a debug log
			//TODO can new logging package handle %08b

			//fmt.Println("FD:")
			//fmt.Printf("%08b", fd)
			//fmt.Println()
		}
	} else {
		logging.Debugf("Request contains no file descriptor")
	}

	return request, fd, err
}

/*
Write will take a string, convert it to byte array and write to UDS
If a file descriptor is included, Write will configure and include it
*/
func (h *handler) Write(response string, fd int) error {
	// write response with or without file descriptor
	if fd > 0 {
		logging.Debugf("Response: %s, FD: %s", response, strconv.Itoa(fd))
		rights := syscall.UnixRights(fd)
		if err := syscall.Sendmsg(h.udsFD, []byte(response), rights, nil, 0); err != nil {
			logging.Errorf("Sendmsg error: %v", err)
			return err
		}
	} else {
		logging.Debugf("Response: %s", response)
		if err := syscall.Sendmsg(h.udsFD, []byte(response), nil, nil, 0); err != nil {
			logging.Errorf("Sendmsg error: %v", err)
			return err
		}
	}
	return nil
}

func generateSocketPath(directory string) (string, error) {
	var sockPath string

	//create directory if not exists, with correct file permissions
	if err := os.MkdirAll(directory, fileMode); err != nil {
		logging.Errorf("Error creating socket file directory %s: %v", directory, err)
		return sockPath, err
	}

	//get directory info
	fileInfo, err := os.Stat(directory)
	if err != nil {
		logging.Errorf("Error getting directory info %s: %v", directory, err)
		return sockPath, err
	}

	//verify it is a directory, incase of pre existing file
	if fileInfo.IsDir() != true {
		err = fmt.Errorf("%s is not a directory", directory)
		logging.Errorf(err.Error())
		return sockPath, err
	}

	//verify the permissions are correct, incase of pre existing dir
	if fileInfo.Mode().Perm() != fileMode {
		err = fmt.Errorf("Incorrect permissions on directory %s", directory)
		logging.Errorf(err.Error())
		return sockPath, err
	}

	for {
		sockName, err := uuid.NewV4()
		if err != nil {
			logging.Errorf("%s", err)
		}

		sockPath = directory + sockName.String() + ".sock"
		if _, err := os.Stat(sockPath); os.IsNotExist(err) {
			break
		}
		logging.Debugf("%s already exists. Regenerating.", sockPath)
	}

	return sockPath, nil
}

func ctrlBufHasValue(s []byte) bool {
	for _, v := range s {
		if v != 0 {
			return true
		}
	}
	return false
}
