/*
 * Copyright(c) 2022 Intel Corporation.
 * Copyright(c) Red Hat Inc.
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

package bpf

/*
fakeHandler implements the Handler interface.
*/
type fakeHandler struct{}

/*
NewFakeHandler returns a fake implementation of the Handler interface.
*/
func NewFakeHandler() Handler {
	return &fakeHandler{}
}

/*
LoadBpfSendXskMap is the GoLang wrapper for the C function Load_bpf_send_xsk_map
In this fakeHandler it returns a hardcoded file descriptor.
*/
func (f *fakeHandler) LoadBpfSendXskMap(ifname string) (int, error) {
	var fakeFileDescriptor int = 7
	return fakeFileDescriptor, nil
}

/*
LoadAttachBpfXdpPass is the GoLang wrapper for the C function Load_attach_bpf_xdp_pass
In this fakeHandler it does nothing.
*/
func (f *fakeHandler) LoadAttachBpfXdpPass(ifname string) error {
	return nil
}

/*
ConfigureBusyPoll is the GoLang wrapper for the C function Configure_busy_poll
In this fakeHandler it does nothing.
*/
func (f *fakeHandler) ConfigureBusyPoll(fd int, busyTimeout int, busyBudget int) error {
	return nil
}

/*
Cleanbpf is the GoLang wrapper for the C function Clean_bpf
In this fakeHandler it does nothing.
*/
func (f *fakeHandler) Cleanbpf(ifname string) error {
	return nil
}
