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
}

/*
fakeHandler implements the FakeHandler interface.
*/
type fakeHandler struct{}

/*
NewFakeHandler returns an implementation of the FakeHandler interface.
*/
func NewFakeHandler() FakeHandler {
	return &fakeHandler{}
}

/*
KernelVersion is a function used to mock and retrieve the KernelVersion
*/
func (r *fakeHandler) KernelVersion() (string, error) {

	return "5.4.1", nil
}

/*
HasEthtool is a function used to mock and search for ethtool
*/
func (r *fakeHandler) HasEthtool() (bool, error) {

	return true, nil
}

/*
HasEthtool is a function used to mock and search for Libbpf
*/
func (r *fakeHandler) HasLibbpf() (bool, error) {
	return true, nil
}

/*
 AllowsUnprivilegedBpf is a function used to mock and check that the AllowsUnprivilegedBpf is set to true
*/
func (r *fakeHandler) AllowsUnprivilegedBpf() (bool, error) {
	boolValue := false
	return !boolValue, nil
}
