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
	"github.com/intel/cndp_device_plugin/pkg/bpf"
	"github.com/intel/cndp_device_plugin/pkg/logging"
	"github.com/intel/cndp_device_plugin/pkg/resourcesapi"
	"github.com/intel/cndp_device_plugin/pkg/uds"
	"net"
	"strconv"
	"strings"
	"time"
)

const (
	handshakeVersion = "0.1"
	requestVersion   = "/version"

	requestConnect  = "/connect"
	responseHostOk  = "/host_ok"
	responseHostNak = "/host_nak"

	requestFd     = "/xsk_map_fd"
	responseFdAck = "/fd_ack"
	responseFdNak = "/fd_nak"

	requestBusyPoll     = "/config_busy_poll"
	responseBusyPollAck = "/config_busy_poll_ack"
	responseBusyPollNak = "/config_busy_poll_nak"

	requestFin     = "/fin"
	responseFinAck = "/fin_ack"

	responseBadRequest     = "/nak"
	responseNotImplemented = "/nak"
	responseError          = "/error"

	udsProtocol    = "unixpacket" // "unix"=SOCK_STREAM, "unixdomain"=SOCK_DGRAM, "unixpacket"=SOCK_SEQPACKET
	udsMsgBufSize  = 64
	udsCtlBufSize  = 4
	usdSockDir     = "/tmp/"
	udsIdleTimeout = 0 * time.Second //TODO make configurable
)

/*
Server is the interface defining the CNDP Unix domain socket server.
Implementations of this interface are the main type of this CNDP package.
*/
type Server interface {
	AddDevice(dev string, fd int)
	Start()
}

/*
ServerFactory is the interface defining a factory that creates and returns Servers.
Each device plugin poolManager will have its own ServerFactory and each time a CNDP
container is created the factory will create a Server to serve the associated Unix
domain socket.
*/
type ServerFactory interface {
	CreateServer(deviceType string) (Server, string)
}

/*
server implements the Server interface. It is the main type for this package.
*/
type server struct {
	deviceType string
	devices    map[string]int
	uds        uds.Handler
	bpf        bpf.Handler
	podRes     resourcesapi.Handler
}

/*
serverFactory implements the ServerFactory interface.
*/
type serverFactory struct {
	ServerFactory
}

/*
NewServerFactory returns an implementation of the ServerFactory interface.
*/
func NewServerFactory() ServerFactory {
	return &serverFactory{}
}

/*
CreateServer creates, initialises, and returns an implementation of the Server interface.
It also returns the filepath of the UDS being served.
*/
func (f *serverFactory) CreateServer(deviceType string) (Server, string) {
	server := &server{
		deviceType: deviceType,
		devices:    make(map[string]int),
		uds:        uds.NewHandler(usdSockDir),
		bpf:        bpf.NewHandler(),
		podRes:     resourcesapi.NewHandler(),
	}

	return server, server.uds.GetSocketPath()
}

/*
Start is the public facing method for starting a Server.
It runs the servers private start method on a Go routine.
*/
func (s *server) Start() {
	go s.start()
}

/*
AddDevice appends a netdev and its associated XSK file descriptor to the Servers map of devices.
*/
func (s *server) AddDevice(dev string, fd int) {
	s.devices[dev] = fd
}

/*
start is a private method and the main loop of the Server.
It listens for and serves a single connection. Across this connection it validates the pod hostname
and serves XSK file descriptors to the CNDP app within the pod.
*/
func (s *server) start() {
	logging.Infof("Initialising Unix domain socket: " + s.uds.GetSocketPath())

	// init
	closeListener, err := s.uds.Init(udsProtocol, udsMsgBufSize, udsCtlBufSize, udsIdleTimeout)
	if err != nil {
		logging.Errorf("Error Initialising UDS: %v", err)
		closeListener()
		return
	}
	defer closeListener()

	logging.Infof("Unix domain socket initialised. Listening for new connection.")

	closeConnection, err := s.uds.Listen()
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			logging.Errorf("Listener timed out: %v", err)
			closeConnection()
			return
		}
		logging.Errorf("Listener Accept error: %v", err)
		closeConnection()
		return
	}
	defer closeConnection()

	logging.Infof("New connection accepted. Waiting for requests.")

	// read incomming request
	request, _, err := s.read()
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			logging.Errorf("Connection timed out: %v", err)
			return
		}
		logging.Errorf("Connection read error: %v", err)
		return
	}

	// first request should validate hostname
	connected := false
	if strings.Contains(request, requestConnect) {
		words := strings.Split(request, ",")
		if len(words) == 2 && words[0] == requestConnect {
			hostname := strings.ReplaceAll(words[1], " ", "")
			connected, err = s.validateHost(hostname)
			if err != nil {
				logging.Errorf("Error validating host "+hostname+": ", err)
				s.write(responseError)
			}
		}
		if connected {
			s.write(responseHostOk)
		} else {
			s.write(responseHostNak)
		}
	}

	// once valid, maintain connection and loop for remaining requests
	for connected {
		// read incoming request
		request, fd, err := s.read()
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				logging.Errorf("Connection timed out: %v", err)
				return
			}
			logging.Errorf("Connection read error: %v", err)
			return
		}

		// process request
		switch {
		case strings.Contains(request, requestFd):
			err = s.handleFdRequest(request)

		case request == requestVersion:
			err = s.write(handshakeVersion)

		case strings.Contains(request, requestBusyPoll):
			err = s.handleBusyPollRequest(request, fd)

		case request == requestFin:
			err = s.write(responseFinAck)
			connected = false

		default:
			err = s.write(responseBadRequest)
		}

		if err != nil {
			logging.Errorf("Error handling request: %v", err)
			return
		}
	}
}

func (s *server) read() (string, int, error) {
	request, fd, err := s.uds.Read()
	if err != nil {
		logging.Errorf("Read error: %v", err)
		return "", 0, err
	}

	logging.Infof("Request: " + request)
	return request, fd, nil
}

func (s *server) write(response string) error {
	if err := s.uds.Write(response, -1); err != nil {
		return err
	}
	return nil
}

func (s *server) writeWithFD(response string, fd int) error {
	if err := s.uds.Write(response, fd); err != nil {
		return err
	}
	return nil
}

func (s *server) handleFdRequest(request string) error {
	words := strings.Split(request, ",")
	if len(words) != 2 || words[0] != requestFd {
		if err := s.write(responseBadRequest); err != nil {
			return err
		}
		return nil
	}

	iface := strings.ReplaceAll(words[1], " ", "")

	if fd, ok := s.devices[iface]; ok {
		logging.Infof("Device " + iface + " recognised")
		if err := s.writeWithFD(responseFdAck, fd); err != nil {
			return err
		}
	} else {
		logging.Warningf("Device " + iface + " not recognised")
		if err := s.write(responseFdNak); err != nil {
			return err
		}
	}
	return nil
}

func (s *server) handleBusyPollRequest(request string, fd int) error {
	if fd <= 0 {
		logging.Errorf("Invalid file descriptor")
		if err := s.write(responseBusyPollNak); err != nil {
			return err
		}
	}

	words := strings.Split(request, ",")
	if len(words) != 3 || words[0] != requestBusyPoll {
		if err := s.write(responseBadRequest); err != nil {
			return err
		}
		return nil
	}

	timeoutString := strings.ReplaceAll(words[1], " ", "")
	budgetString := strings.ReplaceAll(words[2], " ", "")

	timeout, err := strconv.Atoi(timeoutString)
	if err != nil {
		logging.Errorf("Error converting busy timeout to int: %v", err)
		return err
	}

	budget, err := strconv.Atoi(budgetString)
	if err != nil {
		logging.Errorf("Error converting busy budget to int: %v", err)
		return err
	}

	logging.Infof("Configuring busy poll, FD: " + strconv.Itoa(fd) + ", Timeout: " + timeoutString + ", Budget: " + budgetString)

	err = s.bpf.ConfigureBusyPoll(fd, timeout, budget)
	if err != nil {
		logging.Errorf("Error configuring busy poll: %v", err)
		s.write(responseBusyPollNak)
		return err
	}

	s.write(responseBusyPollAck)

	return nil
}

func (s *server) validateHost(hostname string) (bool, error) {
	logging.Infof("Validating pod hostname: " + hostname)
	podResourceMap, _ := s.podRes.GetPodResources() //TODO error

	if _, ok := podResourceMap[hostname]; ok {
		logging.Infof("Pod " + hostname + " found on node")
	} else {
		logging.Errorf(hostname + " not found on node")
		return false, nil
	}

	pod := podResourceMap[hostname]
	valid := false

	for _, container := range pod.GetContainers() {
		for _, device := range container.GetDevices() {

			if device.GetResourceName() != s.deviceType ||
				len(device.GetDeviceIds()) != len(s.devices) {
				// not the resource we're interested in
				// or this container has a different number of the resource
				continue
			}

			// compare known devices (from Allocate) vs devices from resource api
			for _, dev := range device.GetDeviceIds() {
				if _, exists := s.devices[dev]; exists {
					valid = true // valid while devices match
				} else {
					valid = false
					break // not valid if any device does not match
				}
			}

			if valid {
				logging.Infof("Pod " + hostname + " is valid for this UDS connection")
				return true, nil
			}
		}
	}

	logging.Infof("Pod " + hostname + " could not be validated for this UDS connection")
	return false, nil
}
