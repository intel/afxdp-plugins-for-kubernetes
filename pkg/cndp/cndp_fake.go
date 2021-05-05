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

/*
fakeCndp implements the Cndp interface.
*/
type fakeCndp struct {
	Cndp
}

/*
fakeServer implements the UdsServer interface.
*/
type fakeServer struct {
	UdsServer
}

/*
NewFakeCndp returns a struct implementing the Cndp interface.
*/
func NewFakeCndp() Cndp {
	return &fakeCndp{}
}

/*
CreateUdsServer returns an empty struct implementing the UdsServer interface.
Also returns a hardcoded UDS filepath.
*/
func (c *fakeCndp) CreateUdsServer(deviceType string) (UdsServer, string) {
	return &fakeServer{}, "/tmp/fake-socket.sock"
}

/*
Start is the public facing function for starting the udsServer.
In this fakeServer it does nothing.
*/
func (server *fakeServer) Start() {
	return
}

/*
AddDevice appends a netdev name and its file descriptor to the map of devices in the udsServer.
In this fakeServer it does nothing.
*/
func (server *fakeServer) AddDevice(dev string, fd int) {
	return
}
