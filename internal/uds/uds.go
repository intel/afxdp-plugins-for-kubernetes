/*
 * Copyright(c) 2022 Intel Corporation.
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package uds

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/intel/afxdp-plugins-for-kubernetes/internal/host"
	logging "github.com/sirupsen/logrus"
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
	Init(socketPath string, protocol string, msgBufSize int, ctlBufSize int, timeout time.Duration, uid string) error
	Listen() (CleanupFunc, error)
	Dial() (CleanupFunc, error)
	Read() (string, int, error)
	Write(response string, fd int) error
}

/*
handler implements the Handler interface.
*/
type handler struct {
	socketPath string
	socketFile *os.File
	addr       *net.UnixAddr
	listener   *net.UnixListener
	conn       *net.UnixConn
	msgBufSize int
	ctlBufSize int
	timeout    time.Duration
	protocol   string
	uid        string
}

/*
NewHandler returns an implementation of the Handler interface.
*/
func NewHandler() Handler {
	return &handler{}
}

/*
CleanupFunc defines a function that we return from the other functions.
This function is responsible for proper cleanup of the socket files.
*/
type CleanupFunc func()

/*
Init initialises the UDS Handler.
A CleanupFunc function is returned. This function should be deferred by the calling code
to ensure proper socket cleanup.
*/
func (h *handler) Init(socketPath string, protocol string, msgBufSize int, ctlBufSize int, timeout time.Duration, uid string) error {
	var err error

	h.socketPath = socketPath
	h.protocol = protocol
	h.msgBufSize = msgBufSize
	h.ctlBufSize = ctlBufSize
	h.timeout = timeout
	h.uid = uid

	// resolve UDS address
	h.addr, err = net.ResolveUnixAddr(h.protocol, h.socketPath)
	if err != nil {
		logging.Errorf("Error resolving Unix address %s: %v", h.socketPath, err)
		return err
	}

	return nil
}

/*
Listen listens for and accepts new connections.
A CleanupFunc function is returned. This function should be deferred by the calling code
to ensure proper socket cleanup.
*/
func (h *handler) Listen() (CleanupFunc, error) {
	var err error

	// create UDS listener
	h.listener, err = net.ListenUnix(h.protocol, h.addr)
	if err != nil {
		logging.Errorf("Error creating Unix listener for %s: %v", h.socketPath, err)
		return func() { h.cleanup() }, err
	}

	//ACL Permissions
	if h.uid != "0" {
		logging.Infof("Giving permissions to UID %s", h.uid)
		err = host.GivePermissions(h.socketPath, h.uid, "rwx")
		if err != nil {
			logging.Errorf("Error giving permissions to socket file path: %v", err)
			return func() { h.cleanup() }, err
		} else {
			logging.Infof("User %s has access to %s", h.uid, h.socketPath)
		}
	}

	if h.timeout > 0 {
		if err := h.listener.SetDeadline(time.Now().Add(h.timeout)); err != nil {
			logging.Errorf("Error setting listener timeout: %v", err)
			return func() { h.cleanup() }, err
		}
	}

	h.conn, err = h.listener.AcceptUnix()
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			logging.Errorf("Listener timed out: %v", err)
			return func() { h.cleanup() }, err
		}
		logging.Errorf("Listener Accept error: %v", err)
		return func() { h.cleanup() }, err
	}

	return func() { h.cleanup() }, nil
}

/*
Dial creates a new connection
A CleanupFunc function is returned. This function should be deferred by the calling code
to ensure proper socket cleanup.
*/
func (h *handler) Dial() (CleanupFunc, error) {
	var err error

	// create UDS dialer
	h.conn, err = net.DialUnix(h.protocol, nil, h.addr)
	if err != nil {
		logging.Errorf("Error dialling Unix connection on %s: %v", h.socketPath, err)
		return func() { h.cleanup() }, err
	}

	return func() { h.cleanup() }, nil
}

/*
Read will read the incoming message from the UDS.
Message byte array is converted and returned as a string.
The control messages are also checked and returns the FD as an int, if present.
*/
func (h *handler) Read() (string, int, error) {
	var request = ""
	var fd int = 0
	msgBuf := make([]byte, h.msgBufSize)
	ctrlBuf := make([]byte, syscall.CmsgSpace(h.ctlBufSize))

	if h.timeout > 0 {
		if err := h.conn.SetDeadline(time.Now().Add(h.timeout)); err != nil {
			logging.Errorf("Error setting connection timeout: %v", err)
			return request, fd, err
		}
	}

	n, _, _, _, err := h.conn.ReadMsgUnix(msgBuf, ctrlBuf)
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			logging.Errorf("Connection timed out: %v", err)
			return request, fd, err
		}
		logging.Errorf("ReadMsgUnix error: %v", err)
		return request, fd, err
	}

	request = string(msgBuf[0:n])
	logging.Debugf("Read: %s", request)

	if ctrlBufHasValue(ctrlBuf) {
		ctrlMsgs, err := syscall.ParseSocketControlMessage(ctrlBuf)
		if err != nil {
			logging.Errorf("Control messages parse error: %v", err)
			return request, fd, err
		}

		if len(ctrlMsgs) > 0 {
			//Typically code would loop through ctrlMsgs and fds
			//We're handling a single msg and single fd here, so it's msg[0] fds[0]
			fds, _ := syscall.ParseUnixRights(&ctrlMsgs[0])
			fd = fds[0]
			logging.Debugf("Request contains file descriptor: %s", strconv.Itoa(fd))
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

	if fd > 0 {
		logging.Debugf("Write: %s, FD: %s", response, strconv.Itoa(fd))
		rights := syscall.UnixRights(fd)

		if _, _, err := h.conn.WriteMsgUnix([]byte(response), rights, nil); err != nil {
			logging.Errorf("WriteMsgUnix error: %v", err)
			return err
		}
	} else {
		logging.Debugf("Write: %s", response)

		if _, _, err := h.conn.WriteMsgUnix([]byte(response), nil, nil); err != nil {
			logging.Errorf("WriteMsgUnix error: %v", err)
			return err
		}
	}
	return nil
}

/*
GenerateRandomSocketName will take the file directory path, and apply a unique name per each
UDS socket file created.
*/
func GenerateRandomSocketName(directory string, udsDirFileMode os.FileMode) (string, error) {
	//create directory if not exists, with correct file permissions
	if err := os.MkdirAll(directory, udsDirFileMode); err != nil {
		logging.Errorf("Error creating socket file directory %s: %v", directory, err)
		return "", err
	}

	//get directory info
	fileInfo, err := os.Stat(directory)
	if err != nil {
		logging.Errorf("Error getting directory info %s: %v", directory, err)
		return "", err
	}

	//verify it is a directory, in case of pre existing file
	if !fileInfo.IsDir() {
		err = fmt.Errorf("%s is not a directory", directory)
		logging.Errorf(err.Error())
		return "", err
	}

	//verify the permissions are correct, in case of pre existing dir
	if fileInfo.Mode().Perm() != udsDirFileMode {
		err = fmt.Errorf("incorrect permissions on directory %s", directory)
		logging.Errorf(err.Error())
		return "", err
	}

	var sockPath string
	var count int = 0
	for {
		if count >= 5 {
			err = fmt.Errorf("error generating a unique UDS filepath")
			logging.Errorf(err.Error())
			return "", err
		}

		sockName, err := uuid.NewRandom()
		if err != nil {
			logging.Errorf("Error generating random UDS filename: %v", err)
		}

		sockPath = directory + sockName.String() + ".sock"
		if _, err := os.Stat(sockPath); os.IsNotExist(err) {
			break
		}

		logging.Debugf("%s already exists. Regenerating.", sockPath)
		count++
	}

	return sockPath, nil
}

func (h *handler) cleanup() {
	logging.Debugf("Closing Unix listener")
	h.listener.Close()
	if h.conn != nil {
		logging.Debugf("Closing connection")
		h.conn.Close()
	}
	logging.Debugf("Closing socket file")
	h.socketFile.Close()
	logging.Debugf("Removing socket file")
	os.Remove(h.socketPath)
}

func ctrlBufHasValue(s []byte) bool {
	for _, v := range s {
		if v != 0 {
			return true
		}
	}
	return false
}
