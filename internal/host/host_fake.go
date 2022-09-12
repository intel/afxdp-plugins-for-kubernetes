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

package host

/*
FakeHandler interface extends the Handler interface to provide additional testing methods.
*/
type FakeHandler interface {
	Handler
	SetKernalVersion(version string)
	SetAllowsUnprivilegedBpf(allowed bool)
}

/*
fakeHandler implements the FakeHandler interface.
*/
type fakeHandler struct{}

var (
	kernalVersion        string
	privilegedBpfAllowed bool
)

/*
NewFakeHandler returns an implementation of the FakeHandler interface.
*/
func NewFakeHandler() FakeHandler {
	return &fakeHandler{}
}

/*
KernelVersion checks the host kernel version and returns it as a string.
In this FakeHandler it returns a dummy version for testing purposes.
*/
func (r *fakeHandler) KernelVersion() (string, error) {
	return kernalVersion, nil
}

func (r *fakeHandler) SetKernalVersion(version string) {
	kernalVersion = version
}

/*
HasEthtool checks if the host has ethtool installed and returns a boolean.
In this FakeHandler it returns a dummy version for testing purposes.
*/
func (r *fakeHandler) HasEthtool() (bool, string, error) {
	return true, "ethtool version 5.4", nil
}

//setter

/*
HasLibbpf checks if the host has libbpf installed and returns a boolean.
In this FakeHandler it returns a dummy value.
*/
func (r *fakeHandler) HasLibbpf() (bool, []string, error) {
	return true, nil, nil
}

//set setter as SetHasLibBpf

/*
AllowsUnprivilegedBpf checks if the host allows unpriviliged bpf calls and
returns a boolean. In this FakeHandler it returns a dummy value.
*/
func (r *fakeHandler) AllowsUnprivilegedBpf() (bool, error) {
	return privilegedBpfAllowed, nil
}

func (r *fakeHandler) SetAllowsUnprivilegedBpf(allowed bool) {
	privilegedBpfAllowed = allowed
}

/*
HasDevlink checks if the host has devlink installed and returns a boolean.
In this FakeHandler it returns a dummy value.
*/
func (r *fakeHandler) HasDevlink() (bool, string, error) {
	return true, "devlink utility, iproute2-ss200127", nil
}

/*
Hostname is a wrapper function for unit testing that calls os.Hostname.
*/
func (r *fakeHandler) Hostname() (string, error) {
	return "k8sNode1", nil
}

//set setter for setDevLink
