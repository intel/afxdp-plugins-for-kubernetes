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

package udshandler

import (
	"github.com/golang/glog"
	"github.com/nu7hatch/gouuid"
	"net"
	"os"
	"strconv"
	"syscall"
	"time"
)

/*
Handler is the device plugins interface for reading and writing to a Unix domain socket.
The interface exists for testing purposes, allowing unit tests to run without making calls
on a real socket.
*/
type Handler interface {
	Init(protocol string, bufSize int, timeout time.Duration) (CancelFunc, error)
	Listen() (CancelFunc, error)
	GetSocketPath() string
	Read() (string, error)
	Write(response string, fd int) error
}

/*
handler implements the Handler interface.
*/
type handler struct {
	Handler
	socket   string
	listener *net.UnixListener
	conn     *net.UnixConn
	bufSize  int
	udsFD    int
	timeout  time.Duration
}

/*
NewHandler returns an implementation of the Handler interface.
*/
func NewHandler(directory string) Handler {
	socket := generateSocketPath(directory)
	handler := &handler{
		socket: socket,
	}
	return handler
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
func (h *handler) Init(protocol string, bufSize int, timeout time.Duration) (CancelFunc, error) {
	h.bufSize = bufSize

	if timeout > 0 { //TODO test and comment //TODO is this if needed? no?
		h.timeout = timeout
	}

	// resolve UDS address
	addr, err := net.ResolveUnixAddr(protocol, h.socket)
	if err != nil {
		glog.Error("Error resolving Unix address "+h.socket+": ", err)
		return func() {}, err
	}

	// create UDS listener
	h.listener, err = net.ListenUnix(protocol, addr)
	if err != nil {
		glog.Error("Error creating Unix listener for "+h.socket+": ", err)
		return func() { glog.Info("Closing Unix listener"); h.listener.Close() }, err
	}

	if h.timeout > 0 {
		err = h.listener.SetDeadline(time.Now().Add(h.timeout))
		if err != nil {
			glog.Error("Error setting listener timeout: ", err)
			return func() { glog.Info("Closing Unix listener"); h.listener.Close() }, err
		}
	}

	return func() {
		glog.Info("Closing Unix listener")
		h.listener.Close()
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
			glog.Error("Listener timed out: ", err)
			return func() { glog.Info("Closing connection"); h.conn.Close() }, err
		}
		glog.Error("Listener Accept error: ", err)
		return func() { glog.Info("Closing connection"); h.conn.Close() }, err
	}

	// get the UDS socket file descriptor, required for syscall.Recvmsg/Sendmsg
	socketFile, err := h.conn.File()
	if err != nil {
		glog.Error("Error getting socket file descriptor : ", err)
		return func() { glog.Info("Closing connection"); h.conn.Close() }, err
	}
	h.udsFD = int(socketFile.Fd())

	return func() { glog.Info("Closing connection"); h.conn.Close() }, nil
}

/*
Read will read the incoming message from the UDS
Byte arrey is converted and returned as a string
*/
func (h *handler) Read() (string, error) {
	msgBuf := make([]byte, h.bufSize)

	// set connection timeout
	if h.timeout > 0 {
		err := h.conn.SetDeadline(time.Now().Add(h.timeout))
		if err != nil {
			glog.Error("Error setting connection timeout: ", err)
			return "", err
		}
	}

	// read request message
	n, _, _, _, err := syscall.Recvmsg(h.udsFD, msgBuf, nil, 0)
	if err != nil {
		glog.Error("Recvmsg error: ", err)
		return "", err
	}

	return string(msgBuf[0:n]), nil
}

/*
Write will take a string, convert it to byte array and write to UDS
If a file descriptor is included, Write will configure and include it
*/
func (h *handler) Write(response string, fd int) error {
	// write response with or without file descriptor
	if fd > 0 {
		glog.Info("Response: " + response + ", FD: " + strconv.Itoa(fd))
		rights := syscall.UnixRights(fd)
		if err := syscall.Sendmsg(h.udsFD, []byte(response), rights, nil, 0); err != nil {
			glog.Error("Sendmsg error: ", err)
			return err
		}
	} else {
		glog.Info("Response: " + response)
		if err := syscall.Sendmsg(h.udsFD, []byte(response), nil, nil, 0); err != nil {
			glog.Error("Sendmsg error: ", err)
			return err
		}
	}
	return nil
}

func generateSocketPath(directory string) string {
	var sockPath string

	for {
		sockName, err := uuid.NewV4()
		if err != nil {
			glog.Error(err)
		}

		sockPath = directory + sockName.String() + ".sock"
		if _, err := os.Stat(sockPath); os.IsNotExist(err) {
			break
		}
		glog.Info(sockPath + " already exists. Regenerating.")
	}

	return sockPath
}
