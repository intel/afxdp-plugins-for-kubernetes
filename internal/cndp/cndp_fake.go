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

package cndp

/*
fakeServer is a fake implementation the Server interface.
*/
type fakeServer struct{}

/*
fakeServerFactory is a fake implementation the ServerFactory interface.
*/
type fakeServerFactory struct{}

/*
NewFakeServerFactory returns a fake implementation of the ServerFactory interface.
*/
func NewFakeServerFactory() ServerFactory {
	return &fakeServerFactory{}
}

/*
CreateServer creates, initialises, and returns an implementation of the Server interface.
In this fakeServerFactory it returnss an empty fakeServer implementation and a hardcoded
fake UDS filepath.
*/
func (f *fakeServerFactory) CreateServer(deviceType string, timeout int, cndpFuzzTest bool) (Server, string, error) {
	return &fakeServer{}, "/tmp/fake-socket.sock", nil
}

/*
Start is the public facing method for starting a Server.
In this fakeServer it does nothing.
*/
func (s *fakeServer) Start() {
}

/*
AddDevice appends a netdev and its associated XSK file descriptor to the Servers map of devices.
In this fakeServer it does nothing.
*/
func (s *fakeServer) AddDevice(dev string, fd int) {
}
