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
	"strings"
	"time"
)

/*CNDP UDS*/
const (
	requestConnect    = "/connect"
	requestHostname   = "/host"
	requestXskMap     = "/xsk_map_"
	requestXskSock    = "/xsk_sock_"
	requestFdReceived = "/fd_recvd"
	requestFin        = "/fin"

	responseFinAck = "/fin_ack"
	responseHostKv = "\"hostname\":"
	responseHostOk = "/host_ok"

	errorBadRequest     = "Error: Bad Request"
	errorNotImplemented = "Error: Not Implemented"
	errorBadHost        = "Error: Invalid Host"
	errorHostError      = "Error: Error occurred during host validation"

	udsBufSize     = 1024
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
Interface is the interface to the cndp package, representing CNDP to the rest of device plugin.
All interactions with CNDP should be done with an object implementing this interface.
*/
type Interface interface {
	StartUdsServer(server UdsServer)
	CreateUdsSocketPath() string
}

/*
CNDP implements cndp.Interface.
We use this object to interact with CNDP over a Unix Domain Socket.
*/
type CNDP struct {
	Interface
}

/*
NewCndp returns a CNPD object of type cndp.Interface.
*/
func NewCndp() Interface {
	return &CNDP{}
}

/*
UdsServer is the struct representing the UDS Server.
The server runs on a Go routine and serves XDP info to the pod.
*/
type UdsServer struct {
	Socket     string
	DeviceType string
	Devices    map[string]string
}

/*
CreateUdsSocketPath generates a unique filename/filepath for the Unix Domain Socket (UDS).
This UDS will be mounted into the pod and is used for communicating with CNDP.
*/
func (c *CNDP) CreateUdsSocketPath() string {
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

/*
StartUdsServer is the public facing function for starting a UDS server
It runs a private startUdsServer function on a Go Routine
*/
func (c *CNDP) StartUdsServer(server UdsServer) {
	go startUdsServer(server)
}

/*
startUdsServer is the a private function, run on a Go Routine from the public StartUdsServer.
This is the main loop of the UDS server. It listens for and serves only a single connection.
Across this single connection we validate host and serve file descriptors to the CNDP app.
*/
func startUdsServer(server UdsServer) {
	glog.Info("Initialising UDS server on socket " + server.Socket)

	//create UDS
	addr, err := net.ResolveUnixAddr("unix", server.Socket)
	if err != nil {
		glog.Error("Error resolving Unix address "+server.Socket+": ", err)
		return
	}

	//create UDS listener
	listener, err := net.ListenUnix("unix", addr)
	if err != nil {
		glog.Error("Error creating Unix listener for "+server.Socket+": ", err)
		return
	}
	defer func() {
		glog.Info("Closing Unix listener")
		listener.Close()
	}()

	//set UDS listener timeout
	err = listener.SetDeadline(time.Now().Add(udsIdleTimeout))
	if err != nil {
		glog.Error("Error setting listener timeout: ", err)
		return
	}

	glog.Info("UDS server initialised. Listening for new connection.")

	//listen for new connection
	conn, err := listener.Accept()
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
		conn.Close()
	}()

	glog.Info("New connection. Waiting for requests.")
	connected := true
	buf := make([]byte, udsBufSize)

	for connected {
		//(re)set connection timeout
		err = conn.SetDeadline(time.Now().Add(udsIdleTimeout))
		if err != nil {
			glog.Error("Error setting connection timeout: ", err)
			return
		}

		//read incomming requests
		n, err := conn.Read(buf[:])
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				glog.Error("Connection timed out: ", err)
				return
			}
			glog.Error("Connection read error: ", err)
		}

		request := string(buf[0:n])
		glog.Info("Request: " + request)

		var response string

		//handle request
		switch {
		case request == requestConnect:
			response = requestHostname

		case strings.Contains(request, responseHostKv):
			response = server.handleHostValidation(request)

		case strings.Contains(request, requestXskMap):
			response = server.handleXskRequest(request)

		case strings.Contains(request, requestXskSock):
			response = errorNotImplemented

		case request == requestFdReceived:
			response = ""

		case request == requestFin:
			response = responseFinAck
			connected = false

		default:
			response = errorBadRequest
		}

		//send response
		if response != "" {
			glog.Info("Response: " + response)
			_, err = conn.Write([]byte(response))
			if err != nil {
				glog.Error("Connection write error: ", err)
				return
			}
		}
	}
}

func (s *UdsServer) handleHostValidation(request string) string {
	hostname := strings.Trim(strings.TrimPrefix(request, responseHostKv), " \"")
	valid, err := s.validateHost(hostname)
	if err != nil {
		glog.Error("Error validating host "+hostname+": ", err)
		return errorHostError
	}

	if valid {
		return responseHostOk
	}

	return errorBadHost
}

func (s *UdsServer) handleXskRequest(request string) string {
	iface := strings.ReplaceAll(request, requestXskMap, "")
	if descriptor, ok := s.Devices[iface]; ok {
		return descriptor
	}

	return "Error: Interface not valid for this pod: " + iface
}

func (s *UdsServer) validateHost(hostname string) (bool, error) {
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
		glog.Info(hostname + " not found on node")
		return false, nil
	}

	pod := podResourceMap[hostname]
	valid := false

	for _, container := range pod.GetContainers() {
		for _, device := range container.GetDevices() {

			if device.GetResourceName() != s.DeviceType ||
				len(device.GetDeviceIds()) != len(s.Devices) {
				//not the resource we're interested in
				//or this container has a different number of the resource
				continue
			}

			//compare known devices (from Allocate) vs devices from resource api
			for _, dev := range device.GetDeviceIds() {
				if _, exists := s.Devices[dev]; exists {
					valid = true //valid while devices match
				} else {
					valid = false
					continue //not valid if any device does not match
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
