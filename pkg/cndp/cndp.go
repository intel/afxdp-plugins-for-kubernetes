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
	"net"
	"strings"
)

/*
Interface is the interface to the cndp package, representing CNDP to the rest of device plugin.
All interactions with CNDP should be done with an object implementing this interface.
*/
type Interface interface {
	StartSocketServer(SockAddr string)
	CreateUdsSocket() string
}

/*
CNDP implements cndp.Interface.
We use this object to interact with CNDP over a Unix Domain Socket.
*/
type CNDP struct {
	Interface
}

/*
StartSocketServer starts listening on the UDS socket for calls from CNDP.
*/
func (c *CNDP) StartSocketServer(SockAddr string) {

	//TODO currently rough sample code, update this to provide the FD to xdpsock, proper error and socket handeling
	//TODO later update with protocol for interacting with cndp
	//TODO the go routine should be in here, not out in poolManager

	glog.Info("Listening on socket " + SockAddr)

	l, err := net.Listen("unix", SockAddr)
	if err != nil {
		glog.Fatal("listen error:", err)
	}

	// Accept new connections
	conn, err := l.Accept()
	if err != nil {
		glog.Fatal("accept error:", err)
	}

	glog.Info("Client connected")

	buf := make([]byte, 1024)
	for {
		n, err := conn.Read(buf[:])
		if err != nil {
			glog.Fatal("read error:", err)
		}

		glog.Info("Received: " + string(buf[0:n]))

		if strings.Compare("exit", string(buf[0:n])) == 0 {
			break
		}

		_, err = conn.Write([]byte("Hello from DP, you said: " + string(buf[0:n])))
		if err != nil {
			glog.Fatal("write error:", err)
		}

	}

	glog.Info("Closing connection")
	conn.Close()
	l.Close()
}

/*
CreateUdsSocket creates a Unix Domain Socket that will be mounted into the pod.
This UDS is used for interacting with CNDP.
*/
func (c *CNDP) CreateUdsSocket() string {

	sockName, err := uuid.NewV4() //TODO check if it exists

	if err != nil {
		glog.Fatal(err)
	}

	return "/tmp/" + sockName.String() + ".sock"
}

/*
NewCndp returns a CNPD object of type cndp.Interface.
*/
func NewCndp() Interface {
	//TODO also return error?
	return &CNDP{}
}
