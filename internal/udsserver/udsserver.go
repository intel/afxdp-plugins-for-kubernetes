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

package udsserver

import (
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/intel/afxdp-plugins-for-kubernetes/constants"
	"github.com/intel/afxdp-plugins-for-kubernetes/internal/bpf"
	"github.com/intel/afxdp-plugins-for-kubernetes/internal/resourcesapi"
	"github.com/intel/afxdp-plugins-for-kubernetes/internal/uds"
	logging "github.com/sirupsen/logrus"
)

/*
Server is the interface defining the Unix domain socket server.
Implementations of this interface are the main type of this UDSServer package.
*/
type Server interface {
	AddDevice(dev string, fd int)
	Start()
}

/*
ServerFactory is the interface defining a factory that creates and returns Servers.
Each device plugin poolManager will have its own ServerFactory and each time a
UDSServer container is created the factory will create a Server to serve the
associated Unix domain socket.
*/
type ServerFactory interface {
	CreateServer(deviceType, user string, timeout int, udsFuzz bool) (Server, string, error)
}

/*
server implements the Server interface. It is the main type for this package.
*/
type server struct {
	podName        string
	deviceType     string
	devices        map[string]int
	udsPath        string
	uds            uds.Handler
	bpf            bpf.Handler
	podRes         resourcesapi.Handler
	udsIdleTimeout time.Duration
	uid            string
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
func (f *serverFactory) CreateServer(deviceType, user string, timeout int, udsFuzz bool) (Server, string, error) {
	var udsHandler uds.Handler

	if udsFuzz {
		logging.Warningf("UDS Server Fuzzing enabled: Please see fuzzing logs")
		udsHandler = uds.NewFuzzHandler()
	} else {
		udsHandler = uds.NewHandler()
	}

	subDir := strings.ReplaceAll(deviceType, "/", "_")
	udsPath, err := uds.GenerateRandomSocketName(constants.Uds.SockDir+subDir+"/", os.FileMode(constants.Uds.DirFileMode))
	if err != nil {
		logging.Errorf("Error generating socket file path: %v", err)
		return &server{}, "", err
	}

	timeoutUds := time.Duration(timeout) * time.Second

	server := &server{
		podName:        "unvalidated",
		deviceType:     deviceType,
		devices:        make(map[string]int),
		udsPath:        udsPath,
		uds:            udsHandler,
		bpf:            bpf.NewHandler(),
		podRes:         resourcesapi.NewHandler(),
		udsIdleTimeout: timeoutUds,
		uid:            user,
	}

	return server, udsPath, nil
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
and serves XSK file descriptors to the UDS Server app within the pod.
*/
func (s *server) start() {
	logging.Debugf("Initialising Unix domain socket: " + s.udsPath)

	// init
	if err := s.uds.Init(s.udsPath, constants.Uds.Protocol, constants.Uds.MsgBufSize, constants.Uds.CtlBufSize, s.udsIdleTimeout, s.uid); err != nil {
		logging.Errorf("Error Initialising UDS: %v", err)
		return
	}

	logging.Infof("Unix domain socket initialised. Listening for new connection.")

	cleanup, err := s.uds.Listen()
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			logging.Errorf("Listener timed out: %v", err)
			cleanup()
			return
		}
		logging.Errorf("Listener Accept error: %v", err)
		cleanup()
		return
	}
	defer cleanup()

	logging.Infof("New connection accepted. Waiting for requests.")

	// read incoming request
	request, _, err := s.read()
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			logging.Errorf("Connection timed out: %v", err)
			return
		}
		logging.Errorf("Connection read error: %v", err)
		return
	}

	// first request should validate hostname/podname
	connected := false
	var podName string
	if strings.Contains(request, constants.Uds.Handshake.RequestConnect) {
		words := strings.Split(request, ",")
		if len(words) == 2 && words[0] == constants.Uds.Handshake.RequestConnect {
			podName = strings.ReplaceAll(words[1], " ", "")
			connected, err = s.validatePod(podName)
			if err != nil {
				logging.Errorf("Error validating host %s: %v", podName, err)
				if err := s.write(constants.Uds.Handshake.ResponseError); err != nil {
					logging.Errorf("Connection write error: %v", err)
				}
			}
		}
		if connected {
			s.podName = podName
			if err := s.write(constants.Uds.Handshake.ResponseHostOk); err != nil {
				logging.Errorf("Connection write error: %v", err)
			}
		} else {
			if err := s.write(constants.Uds.Handshake.ResponseHostNak); err != nil {
				logging.Errorf("Connection write error: %v", err)
			}
		}
	}

	// once valid, maintain connection and loop for remaining requests
	for connected {
		// read incoming request
		request, fd, err := s.read()
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				logging.Errorf("Pod "+s.podName+" - Connection timed out: %v", err)
				return
			}
			logging.Errorf("Pod "+s.podName+" - Connection read error: %v", err)
			return
		}

		// process request
		switch {
		case strings.Contains(request, constants.Uds.Handshake.RequestFd):
			err = s.handleFdRequest(request)

		case request == constants.Uds.Handshake.RequestVersion:
			err = s.write(constants.Uds.Handshake.Version)

		case strings.Contains(request, constants.Uds.Handshake.RequestBusyPoll):
			err = s.handleBusyPollRequest(request, fd)

		case request == constants.Uds.Handshake.RequestFin:
			err = s.write(constants.Uds.Handshake.ResponseFinAck)
			connected = false

		default:
			err = s.write(constants.Uds.Handshake.ResponseBadRequest)
		}

		if err != nil {
			logging.Errorf("Pod "+s.podName+" - Error handling request: %v", err)
			return
		}
	}
}

func (s *server) read() (string, int, error) {
	request, fd, err := s.uds.Read()
	if err != nil {
		logging.Errorf("Pod "+s.podName+" - Read error: %v", err)
		return "", 0, err
	}

	logging.Infof("Pod " + s.podName + " - Request: " + request)
	return request, fd, nil
}

func (s *server) write(response string) error {
	logging.Infof("Pod " + s.podName + " - Response: " + response)
	if err := s.uds.Write(response, -1); err != nil {
		return err
	}
	return nil
}

func (s *server) writeWithFD(response string, fd int) error {
	logging.Infof("Pod " + s.podName + " - Response: " + response + ", FD: " + strconv.Itoa(fd))
	if err := s.uds.Write(response, fd); err != nil {
		return err
	}
	return nil
}

func (s *server) handleFdRequest(request string) error {
	words := strings.Split(request, ",")
	if len(words) != 2 || words[0] != constants.Uds.Handshake.RequestFd {
		if err := s.write(constants.Uds.Handshake.ResponseBadRequest); err != nil {
			return err
		}
		return nil
	}

	iface := strings.ReplaceAll(words[1], " ", "")

	if fd, ok := s.devices[iface]; ok {
		logging.Debugf("Pod " + s.podName + " - Device " + iface + " recognised")
		if err := s.writeWithFD(constants.Uds.Handshake.ResponseFdAck, fd); err != nil {
			return err
		}
	} else {
		logging.Warningf("Pod " + s.podName + " - Device " + iface + " not recognised")
		if err := s.write(constants.Uds.Handshake.ResponseFdNak); err != nil {
			return err
		}
	}
	return nil
}

func (s *server) handleBusyPollRequest(request string, fd int) error {
	if fd <= 0 {
		logging.Errorf("Pod " + s.podName + " - Invalid file descriptor")
		if err := s.write(constants.Uds.Handshake.ResponseBusyPollNak); err != nil {
			return err
		}
	}

	words := strings.Split(request, ",")
	if len(words) != 3 || words[0] != constants.Uds.Handshake.RequestBusyPoll {
		if err := s.write(constants.Uds.Handshake.ResponseBadRequest); err != nil {
			return err
		}
		return nil
	}

	timeoutString := strings.ReplaceAll(words[1], " ", "")
	budgetString := strings.ReplaceAll(words[2], " ", "")

	timeout, err := strconv.Atoi(timeoutString)
	if err != nil {
		logging.Errorf("Pod "+s.podName+" - Error converting busy timeout to int: %v", err)
		return err
	}

	budget, err := strconv.Atoi(budgetString)
	if err != nil {
		logging.Errorf("Pod "+s.podName+" - Error converting busy budget to int: %v", err)
		return err
	}

	logging.Infof("Pod " + s.podName + " - Configuring busy poll, FD: " + strconv.Itoa(fd) + ", Timeout: " + timeoutString + ", Budget: " + budgetString)

	if err := s.bpf.ConfigureBusyPoll(fd, timeout, budget); err != nil {
		logging.Errorf("Error configuring busy poll: %v", err)
		if err := s.write(constants.Uds.Handshake.ResponseBusyPollNak); err != nil {
			logging.Errorf("Connection write error: %v", err)
		}
		return err
	}
	if err := s.write(constants.Uds.Handshake.ResponseBusyPollAck); err != nil {
		logging.Errorf("Connection write error: %v", err)
	}

	return nil
}

func (s *server) validatePod(podName string) (bool, error) {
	logging.Debugf("Pod " + podName + " - Validating pod hostname")

	podResourceMap, err := s.podRes.GetPodResources()
	if err != nil {
		logging.Errorf("Error getting pod resources: %v", err)
		return false, err
	}

	if _, ok := podResourceMap[podName]; ok {
		logging.Debugf("Pod " + podName + " - Found on node")
	} else {
		logging.Warningf("Pod " + podName + " - Not found on node")
		return false, nil
	}

	pod := podResourceMap[podName]
	valid := false

	for _, container := range pod.GetContainers() {
		var contDevs []string

		for _, devType := range container.GetDevices() {
			if devType.GetResourceName() == s.deviceType {
				contDevs = append(contDevs, devType.GetDeviceIds()...)

			}
		}

		if len(contDevs) == len(s.devices) {
			// compare known devices (from Allocate) vs devices from resource api
			for _, dev := range contDevs {
				if _, exists := s.devices[dev]; exists {
					valid = true // valid while devices match
				} else {
					valid = false
					break // not valid if any device does not match
				}
			}
		}

		if valid {
			logging.Infof("Pod " + podName + " is valid for this UDS connection")
			return true, nil
		}
	}

	logging.Warningf("Pod " + podName + " could not be validated for this UDS connection")
	return false, nil
}
