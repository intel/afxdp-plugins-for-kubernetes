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

package uds

import "time"

/*
FakeHandler interface extends the Handler interface to provide additional testing methods.
*/
type FakeHandler interface {
	Handler
	SetRequests(requests map[int]string)
	GetResponses() map[int]string
}

/*
fakeHandler implements the Handler interface.
*/
type fakeHandler struct {
	counter         int
	fakeRequests    map[int]string
	actualResponses map[int]string
}

/*
NewFakeHandler returns a fake implementation of the Handler interface.
*/
func NewFakeHandler() FakeHandler {
	return &fakeHandler{}
}

/*
Init should initialises the Unix domain socket.
In this fakeHandler it resets some counters and inits a map for recording calls to the Write() function.
*/
func (f *fakeHandler) Init(socketPath string, protocol string, msgbufSize int, ctlBufSize int, timeout time.Duration) error {
	f.actualResponses = make(map[int]string)
	f.counter = 0
	return nil
}

/*
Listen listens for and accepts new connections.
In this fakeHandler it does nothing.
*/
func (f *fakeHandler) Listen() (CleanupFunc, error) {
	return func() {}, nil
}

/*
Dial is TODO
*/
func (f *fakeHandler) Dial() (CleanupFunc, error) {
	return func() {}, nil
}

/*
Read should read the incoming message from the UDS.
In this fakeHandler it will sequentially return a set of predetermined strings.
*/
func (f *fakeHandler) Read() (string, int, error) {
	request := f.fakeRequests[f.counter]
	return request, 0, nil
}

/*
Write should write a string to the UDS.
In this fakeHandler, the string is stored in a map so we can later compare each response to each request.
*/
func (f *fakeHandler) Write(response string, fd int) error {
	f.actualResponses[f.counter] = response
	f.counter = f.counter + 1
	return nil
}

/*
SetRequests takes a map of strings. These strings will be sequentially returned
each time the Read function is called. This allows us to build a list of fake
requests we want to make over the UDS.
*/
func (f *fakeHandler) SetRequests(requests map[int]string) {
	f.fakeRequests = requests
	f.counter = 0
}

/*
GetResponses returns the list of responses that were made via the Write function.
*/
func (f *fakeHandler) GetResponses() map[int]string {
	return f.actualResponses
}
