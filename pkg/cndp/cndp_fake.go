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

type fakeCNDP struct {
	Interface
}

func (c *fakeCNDP) StartUdsServer(server UdsServer) {

}

func (c *fakeCNDP) CreateUdsSocketPath() string {
	return "/tmp/fake-socket.sock"
}

/*
NewFakeCndp returns a fake CNDP object of type cndp.Interface.
The functions of the fake CNDP have little functionality. Returns are static and predictable.
This is used while testing in other areas of the device plugin.
*/
func NewFakeCndp() Interface {
	return &fakeCNDP{}
}
