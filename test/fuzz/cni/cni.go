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

package cni

import (
	"encoding/json"
	"fmt"
	"github.com/containernetworking/cni/pkg/skel"
	"github.com/intel/afxdp-plugins-for-kubernetes/internal/cni"
	"k8s.io/apimachinery/pkg/util/uuid"
)

const (
	fuzzNs        = "fuzznet"
	interesting   = 1
	uninteresting = 0
	discard       = -1
)

/*
FuzzAdd sends fuzzed data into the cni CmdAdd function
The input data is considered:
 - uninteresting if is caught by an existing error
 - interesting if it does not result in an error, input priority increases for subsequent fuzzing
 - discard if it will not unmarshall, so we don't just end up testing the json.Unmarshall function
*/
func FuzzAdd(data []byte) int {
	tmp := &cni.NetConfig{}
	if err := json.Unmarshal(data, tmp); err != nil {
		return discard
	}

	fakeID := uuid.NewUUID()
	a := &skel.CmdArgs{
		StdinData:   data, // data prepared by go-fuzz
		Netns:       "/var/run/netns/" + fuzzNs,
		ContainerID: string(fakeID),
		IfName:      fmt.Sprintf("eth%v", int(fakeID[7])),
	}

	if err := cni.CmdAdd(a); err != nil {
		return uninteresting
	}
	return interesting
}

/*
FuzzDel sends fuzzed data into the cni CmdDel function
The input data is considered:
 - uninteresting if is caught by an exesting error
 - interesting if it does not result in an error, input priority increases for subsequent fuzzing
 - discard if it will not unmarshall, so we don't just end up testing the json.Unmarshall function
*/
func FuzzDel(data []byte) int {
	tmp := &cni.NetConfig{}
	if err := json.Unmarshal(data, tmp); err != nil {
		return discard
	}

	fakeID := uuid.NewUUID()
	a := &skel.CmdArgs{
		StdinData:   data, // data prepared by go-fuzz
		Netns:       "/var/run/netns/" + fuzzNs,
		ContainerID: string(fakeID),
		IfName:      fmt.Sprintf("eth%v", int(fakeID[7])),
	}

	if err := cni.CmdDel(a); err != nil {
		return uninteresting
	}
	return interesting
}
