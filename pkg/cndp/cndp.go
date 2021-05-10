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

package cndp

import (
	"github.com/golang/glog"
	"github.com/nu7hatch/gouuid"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	podresourcesapi "k8s.io/kubelet/pkg/apis/podresources/v1alpha1"
	"net"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"
)

/*CNDP UDS*/
const (
	handshakeVersion = "0.1"
	requestVersion = "/version"

	requestConnect  = "/connect"
	responseHostOk  = "/host_ok"
	responseHostNak = "/host_nak"

	requestFd     = "/xsk_map_fd"
	responseFdAck = "/fd_ack"
	responseFdNak = "/fd_nak"

	requestFin     = "/fin"
	responseFinAck = "/fin_ack"

	responseBadRequest     = "/nak"
	responseNotImplemented = "/nak"
	responseError          = "/error"

	udsProtocol    = "unixpacket" // "unix"=SOCK_STREAM, "unixdomain"=SOCK_DGRAM, "unixpacket"=SOCK_SEQPACKET
	udsBufSize     = 64
	usdSockDir     = "/tmp/"
	udsIdleTimeout = 60 * time.Second
)

/*Pod Resources API*/
const (
	podResSockDir  = "/var/lib/kubelet/pod-resources"
	podResSockPath = podResSockDir + "/kubelet.sock"
	podResTimeout  = 10 * time.Second
)

/*
Cndp is the interface to the cndp package.
Mainly exists for testing purposes, allowing the unit tests to
test device plugin code against a non-functioning fake cndp.
*/
type Cndp interface {
	CreateUdsServer(deviceType string) (UdsServer, string)
}

/*
UdsServer is the interface for the Unix domain socket server.
Defines the public facing functions of the server.
*/
type UdsServer interface {
	Start()
	AddDevice(dev string, fd int)
}

/*
cndp implements the Cndp interface.
*/
type cndp struct {
	Cndp
}

/*
udsServer implements the UdsServer interface.
*/
type udsServer struct {
	UdsServer
	socket     string
	conn       *net.UnixConn
	udsFD      int
	timeout    bool
	deviceType string
	devices    map[string]int
}

/*
NewCndp returns a struct implementing the Cndp interface.
*/
func NewCndp() Cndp {
	return &cndp{}
}

/*
CreateUdsServer initialises and returns a struct implementing the UdsServer interface.
Also returns the filepath of the UDS.
*/
func (c *cndp) CreateUdsServer(deviceType string) (UdsServer, string) {
	socket := generateSocketPath()

	server := &udsServer{
		socket:     socket,
		timeout:    false, // TODO enable, make configurable
		deviceType: deviceType,
		devices:    make(map[string]int),
	}

	return server, socket
}

/*
Start is the public facing function for starting the udsServer.
It runs the servers private start() function on a Go routine.
*/
func (server *udsServer) Start() {
	go server.start()
}

/*
AddDevice appends a netdev name and its file descriptor to the map of devices in the udsServer.
*/
func (server *udsServer) AddDevice(dev string, fd int) {
	server.devices[dev] = fd
}

/*
start is the main loop of the udsServer. It listens for and serves a single connection.
Across this connection it validates the pod hostname and serves file descriptors to the CNDP app.
*/
func (server *udsServer) start() {
	glog.Info("Initialising UDS server on socket " + server.socket)

	// resolve UDS address
	addr, err := net.ResolveUnixAddr(udsProtocol, server.socket)
	if err != nil {
		glog.Error("Error resolving Unix address "+server.socket+": ", err)
		return
	}

	// create UDS listener
	listener, err := net.ListenUnix(udsProtocol, addr)
	if err != nil {
		glog.Error("Error creating Unix listener for "+server.socket+": ", err)
		return
	}
	defer func() {
		glog.Info("Closing Unix listener")
		listener.Close()
	}()

	// set UDS listener timeout
	if server.timeout {
		err = listener.SetDeadline(time.Now().Add(udsIdleTimeout))
		if err != nil {
			glog.Error("Error setting listener timeout: ", err)
			return
		}
	}

	glog.Info("UDS server initialised. Listening for new connection.")

	// listen for new connection
	server.conn, err = listener.AcceptUnix()
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			glog.Error("Listener timed out: ", err)
			return
		}
		glog.Error("Listener Accept error: ", err)
		return
	}
	defer func() {
		glog.Info("Closing connection")
		server.conn.Close()
	}()

	glog.Info("New connection. Waiting for requests.")

	// get the UDS socket file descriptor, required for syscall.Recvmsg/Sendmsg
	socketFile, err := server.conn.File()
	if err != nil {
		glog.Error("Error getting socket file descriptor : ", err)
		return
	}
	defer socketFile.Close()
	server.udsFD = int(socketFile.Fd())

	// read incomming request
	request, err := server.read()
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			glog.Error("Connection timed out: ", err)
			return
		}
		glog.Error("Connection read error: ", err)
		return
	}

	// first request should validate hostname
	connected := false
	if strings.Contains(request, requestConnect) {
		s := strings.Split(request, ",")
		hostname := strings.ReplaceAll(s[1], " ", "")

		valid, err := server.validateHost(hostname)
		if err != nil {
			glog.Error("Error validating host "+hostname+": ", err)
			server.write(responseError)
		}

		if valid {
			server.write(responseHostOk)
			connected = true
		} else {
			server.write(responseHostNak)
		}
	}

	// once valid maintain connection, loop for remaining requests
	for connected {
		// read incoming request
		request, err := server.read()
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				glog.Error("Connection timed out: ", err)
				return
			}
			glog.Error("Connection read error: ", err)
			return
		}

		// handle request
		switch {
		case strings.Contains(request, requestFd):
			err = server.handleXskRequest(request)

		case request == requestVersion:
			err = server.write(handshakeVersion)

		case request == requestFin:
			err = server.write(responseFinAck)
			connected = false

		default:
			err = server.write(responseBadRequest)
		}

		if err != nil {
			glog.Error("Error handling request: ", err)
			return
		}
	}
}

func (server *udsServer) read() (string, error) {
	msgBuf := make([]byte, udsBufSize)

	// set connection timeout
	if server.timeout {
		err := server.conn.SetDeadline(time.Now().Add(udsIdleTimeout))
		if err != nil {
			glog.Error("Error setting connection timeout: ", err)
			return "", err
		}
	}

	// read request message
	n, _, _, _, err := syscall.Recvmsg(server.udsFD, msgBuf, nil, 0)
	if err != nil {
		glog.Error("Recvmsg error: ", err)
		return "", err
	}

	request := string(msgBuf[0:n])
	glog.Info("Request: " + request)
	return request, nil
}

func (server *udsServer) write(response string) error {
	if err := server.writeWithFD(response, -1); err != nil {
		return err
	}
	return nil
}

func (server *udsServer) writeWithFD(response string, fd int) error {
	// write response with or without file descriptor
	if fd > 0 {
		glog.Info("Response: " + response + ", FD: " + strconv.Itoa(fd))
		rights := syscall.UnixRights(fd)
		if err := syscall.Sendmsg(server.udsFD, []byte(response), rights, nil, 0); err != nil {
			glog.Error("Sendmsg error: ", err)
			return err
		}
	} else {
		glog.Info("Response: " + response)
		if err := syscall.Sendmsg(server.udsFD, []byte(response), nil, nil, 0); err != nil {
			glog.Error("Sendmsg error: ", err)
			return err
		}
	}
	return nil
}

func (server *udsServer) handleXskRequest(request string) error {
	s := strings.Split(request, ",")
	iface := strings.ReplaceAll(s[1], " ", "")

	if fd, ok := server.devices[iface]; ok {
		glog.Info("Device " + iface + " recognised")
		if err := server.writeWithFD(responseFdAck, fd); err != nil {
			return err
		}
	} else {
		glog.Error("Device " + iface + " not recognised")
		if err := server.write(responseFdNak); err != nil {
			return err
		}
	}
	return nil
}

func (server *udsServer) validateHost(hostname string) (bool, error) {
	glog.Info("Validating pod hostname: " + hostname)

	resp, err := getPodResources(podResSockPath)
	if err != nil {
		glog.Error("Error Getting pod resources: ", err)
		return false, err
	}

	podResourceMap := make(map[string]podresourcesapi.PodResources)

	for _, pod := range resp.GetPodResources() {
		podResourceMap[pod.GetName()] = *pod
	}

	if _, ok := podResourceMap[hostname]; ok {
		glog.Info(hostname + " found on node")
	} else {
		glog.Error(hostname + " not found on node")
		return false, nil
	}

	pod := podResourceMap[hostname]
	valid := false

	for _, container := range pod.GetContainers() {
		for _, device := range container.GetDevices() {

			if device.GetResourceName() != server.deviceType ||
				len(device.GetDeviceIds()) != len(server.devices) {
				// not the resource we're interested in
				// or this container has a different number of the resource
				continue
			}

			// compare known devices (from Allocate) vs devices from resource api
			for _, dev := range device.GetDeviceIds() {
				if _, exists := server.devices[dev]; exists {
					valid = true // valid while devices match
				} else {
					valid = false
					continue // not valid if any device does not match
				}
			}

			if valid {
				glog.Info(hostname + " is valid for this UDS connection")
				return true, nil
			}
		}
	}

	glog.Info(hostname + " could not be validated for this UDS connection")
	return false, nil
}

func getPodResources(socket string) (*podresourcesapi.ListPodResourcesResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), podResTimeout)
	defer cancel()

	glog.Info("Opening Pod Resource API connection")
	conn, err := grpc.DialContext(ctx, socket, grpc.WithInsecure(), grpc.WithBlock(),
		grpc.WithDialer(func(addr string, timeout time.Duration) (net.Conn, error) {
			return net.DialTimeout("unix", addr, timeout)
		}),
	)
	if err != nil {
		glog.Error("Error connecting to Pod Resource API: ", err)
		return nil, err
	}
	defer func() {
		glog.Info("Closing Pod Resource API connection")
		conn.Close()
	}()

	glog.Info("Requesting pod resource list")
	client := podresourcesapi.NewPodResourcesListerClient(conn)

	resp, err := client.List(ctx, &podresourcesapi.ListPodResourcesRequest{})
	if err != nil {
		glog.Error("Error getting Pod Resource list: ", err)
		return nil, err
	}

	return resp, nil
}

func generateSocketPath() string {
	var sockPath string

	for {
		sockName, err := uuid.NewV4()
		if err != nil {
			glog.Error(err)
		}

		sockPath = usdSockDir + sockName.String() + ".sock"
		if _, err := os.Stat(sockPath); os.IsNotExist(err) {
			break
		}
		glog.Info(sockPath + " already exists. Regenerating.")
	}

	return sockPath
}
