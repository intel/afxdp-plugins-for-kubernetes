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

package udsserver

import (
	"testing"

	"github.com/intel/afxdp-plugins-for-kubernetes/constants"
	"github.com/intel/afxdp-plugins-for-kubernetes/internal/resourcesapi"
	"github.com/intel/afxdp-plugins-for-kubernetes/internal/uds"
	"gotest.tools/assert"
)

func TestCreateNewServer(t *testing.T) {
	//uds := NewServerFactory() //TODO

	testCases := []struct {
		testName       string
		deviceType     string
		expectedPath   string
		expectedServer *server
	}{
		{
			testName:   "Create UDS Server",
			deviceType: "uds/device",
			expectedServer: &server{
				deviceType: "uds/device",
				devices:    make(map[string]int),
				uds:        uds.NewFakeHandler(),
				podRes:     resourcesapi.NewFakeHandler(),
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			//TODO compare individual elements
			//receivedServer, _ := udsserver.CreateServer(tc.deviceType)
			//assert.DeepEqual(t, tc.expectedServer, receivedServer, cmp.AllowUnexported(server{}, uds.udsHandler{}))
		})
	}
}

func TestAddDevice(t *testing.T) {
	server := &server{
		devices: make(map[string]int),
	}

	testCases := []struct {
		testName string
		devices  map[string]int
	}{
		{
			testName: "Add device",
			devices: map[string]int{
				"dev1": 123,
			},
		},
		{
			testName: "Add devices",
			devices: map[string]int{
				"dev1": 1,
				"dev2": 23,
				"dev3": 456,
				"dev4": 78910,
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {

			for device, fd := range tc.devices {
				server.AddDevice(device, fd)
			}
			assert.DeepEqual(t, server.devices, tc.devices)
		})
	}
}

func TestStart(t *testing.T) {
	fakeUDS := uds.NewFakeHandler()
	fakeResAPI := resourcesapi.NewFakeHandler()

	testCases := []struct {
		testName         string
		fakePodName      string
		fakePodNamespace string
		fakeResourceName string
		udsServerDevType string
		fakePodDevices   []string
		udsServerDevices []string
		fakeRequests     map[int]string
		expectedResponse map[int]string
	}{
		/***************************************
		Positive Tests - validate and disconnect
		***************************************/
		{
			//Try connect good podA which has 1 good device - devA
			testName:         "Connect successfully and disconnect, 1 device",
			fakePodName:      "podA",
			fakePodNamespace: "default",
			fakeResourceName: "uds/testing",
			udsServerDevType: "uds/testing",
			fakePodDevices:   []string{"devA"},
			udsServerDevices: []string{"devA"},
			fakeRequests: map[int]string{
				0: constants.Uds.Handshake.RequestConnect + ", podA",
				1: constants.Uds.Handshake.RequestFin,
			},
			expectedResponse: map[int]string{
				0: constants.Uds.Handshake.ResponseHostOk,
				1: constants.Uds.Handshake.ResponseFinAck,
			},
		},
		{
			//Try connect good podA which has 2 good devices - devA and devB
			testName:         "Connect successfully and disconnect, 2 device",
			fakePodName:      "podA",
			fakePodNamespace: "default",
			fakeResourceName: "uds/testing",
			udsServerDevType: "uds/testing",
			fakePodDevices:   []string{"devA", "devB"},
			udsServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: constants.Uds.Handshake.RequestConnect + ", podA",
				1: constants.Uds.Handshake.RequestFin,
			},
			expectedResponse: map[int]string{
				0: constants.Uds.Handshake.ResponseHostOk,
				1: constants.Uds.Handshake.ResponseFinAck,
			},
		},
		{
			//Try connect good podA which has 10 good devices - devA to devJ
			testName:         "Connect successfully and disconnect, 2 device",
			fakePodName:      "podA",
			fakePodNamespace: "default",
			fakeResourceName: "uds/testing",
			udsServerDevType: "uds/testing",
			fakePodDevices:   []string{"devA", "devB", "devC", "devD", "devE", "devF", "devG", "devH", "devI", "devJ"},
			udsServerDevices: []string{"devA", "devB", "devC", "devD", "devE", "devF", "devG", "devH", "devI", "devJ"},
			fakeRequests: map[int]string{
				0: constants.Uds.Handshake.RequestConnect + ", podA",
				1: constants.Uds.Handshake.RequestFin,
			},
			expectedResponse: map[int]string{
				0: constants.Uds.Handshake.ResponseHostOk,
				1: constants.Uds.Handshake.ResponseFinAck,
			},
		},
		{
			//Try connect good podA, in non-default namespace, which has 2 good devices - devA and devB
			testName:         "Connect a pod in non-default NS",
			fakePodName:      "podA",
			fakePodNamespace: "someOtherNamespace",
			fakeResourceName: "uds/testing",
			udsServerDevType: "uds/testing",
			fakePodDevices:   []string{"devA", "devB"},
			udsServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: constants.Uds.Handshake.RequestConnect + ", podA",
				1: constants.Uds.Handshake.RequestFin,
			},
			expectedResponse: map[int]string{
				0: constants.Uds.Handshake.ResponseHostOk,
				1: constants.Uds.Handshake.ResponseFinAck,
			},
		},
		/*********************************************************
		Positive Tests - validate, request good FDs and disconnect
		*********************************************************/
		{
			//Connect podA, request FD for it's single device - devA
			testName:         "Connect and request good FD, 1 device",
			fakePodName:      "podA",
			fakePodNamespace: "default",
			fakeResourceName: "uds/testing",
			udsServerDevType: "uds/testing",
			fakePodDevices:   []string{"devA"},
			udsServerDevices: []string{"devA"},
			fakeRequests: map[int]string{
				0: constants.Uds.Handshake.RequestConnect + ", podA",
				1: constants.Uds.Handshake.RequestFd + ", devA",
				2: constants.Uds.Handshake.RequestFin,
			},
			expectedResponse: map[int]string{
				0: constants.Uds.Handshake.ResponseHostOk,
				1: constants.Uds.Handshake.ResponseFdAck,
				2: constants.Uds.Handshake.ResponseFinAck,
			},
		},
		{
			//Connect podA, request FDs for it's 3 devices - devA devB devC
			testName:         "Connect and request good FDs, 3 devices",
			fakePodName:      "podA",
			fakePodNamespace: "default",
			fakeResourceName: "uds/testing",
			udsServerDevType: "uds/testing",
			fakePodDevices:   []string{"devA", "devB", "devC"},
			udsServerDevices: []string{"devA", "devB", "devC"},
			fakeRequests: map[int]string{
				0: constants.Uds.Handshake.RequestConnect + ", podA",
				1: constants.Uds.Handshake.RequestFd + ", devA",
				2: constants.Uds.Handshake.RequestFd + ", devB",
				3: constants.Uds.Handshake.RequestFd + ", devC",
				4: constants.Uds.Handshake.RequestFin,
			},
			expectedResponse: map[int]string{
				0: constants.Uds.Handshake.ResponseHostOk,
				1: constants.Uds.Handshake.ResponseFdAck,
				2: constants.Uds.Handshake.ResponseFdAck,
				3: constants.Uds.Handshake.ResponseFdAck,
				4: constants.Uds.Handshake.ResponseFinAck,
			},
		},
		{
			//Connect podA, request version and disconnect
			testName:         "Connect and request version",
			fakePodName:      "podA",
			fakePodNamespace: "default",
			fakeResourceName: "uds/testing",
			udsServerDevType: "uds/testing",
			fakePodDevices:   []string{"devA", "devB"},
			udsServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: constants.Uds.Handshake.RequestConnect + ", podA",
				1: constants.Uds.Handshake.RequestVersion,
				2: constants.Uds.Handshake.RequestFin,
			},
			expectedResponse: map[int]string{
				0: constants.Uds.Handshake.ResponseHostOk,
				1: constants.Uds.Handshake.Version,
				2: constants.Uds.Handshake.ResponseFinAck,
			},
		},
		{
			//Connect and test full handshake
			testName:         "Full handshake",
			fakePodName:      "podA",
			fakePodNamespace: "default",
			fakeResourceName: "uds/testing",
			udsServerDevType: "uds/testing",
			fakePodDevices:   []string{"devA", "devB", "devC"},
			udsServerDevices: []string{"devA", "devB", "devC"},
			fakeRequests: map[int]string{
				0: constants.Uds.Handshake.RequestConnect + ", podA",
				1: constants.Uds.Handshake.RequestVersion,
				2: constants.Uds.Handshake.RequestFd + ", devA",
				3: constants.Uds.Handshake.RequestFd + ", devB",
				4: constants.Uds.Handshake.RequestFd + ", devC",
				5: constants.Uds.Handshake.RequestFin,
			},
			expectedResponse: map[int]string{
				0: constants.Uds.Handshake.ResponseHostOk,
				1: constants.Uds.Handshake.Version,
				2: constants.Uds.Handshake.ResponseFdAck,
				3: constants.Uds.Handshake.ResponseFdAck,
				4: constants.Uds.Handshake.ResponseFdAck,
				5: constants.Uds.Handshake.ResponseFinAck,
			},
		},
		/*************************************************************************************
		Negative Tests - do not validate
		NOTE: we shouldn't need to call /fin in any of these as we should never connect
		Including it anyway to prevent hanging in case we manage to make an unexpected connect
		i.e. in case of bad code!
		*************************************************************************************/
		{
			//Try connect without passing any pod
			testName:         "No hostname (1)",
			fakePodName:      "podA",
			fakePodNamespace: "default",
			fakeResourceName: "uds/testing",
			udsServerDevType: "uds/testing",
			fakePodDevices:   []string{"devA", "devB"},
			udsServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: constants.Uds.Handshake.RequestConnect,
				1: constants.Uds.Handshake.RequestFin,
			},
			expectedResponse: map[int]string{
				0: constants.Uds.Handshake.ResponseHostNak,
				1: "should not get " + constants.Uds.Handshake.ResponseFinAck + " as should not have connected",
			},
		},
		{
			//Try connect without passing any pod, include the comma
			testName:         "No hostname (2)",
			fakePodName:      "podA",
			fakePodNamespace: "default",
			fakeResourceName: "uds/testing",
			udsServerDevType: "uds/testing",
			fakePodDevices:   []string{"devA", "devB"},
			udsServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: constants.Uds.Handshake.RequestConnect + ",",
				1: constants.Uds.Handshake.RequestFin,
			},
			expectedResponse: map[int]string{
				0: constants.Uds.Handshake.ResponseHostNak,
				1: "should not get " + constants.Uds.Handshake.ResponseFinAck + " as should not have connected",
			},
		},
		{
			//Try connect 2 hostnames
			testName:         "Two hostnames",
			fakePodName:      "podA",
			fakePodNamespace: "default",
			fakeResourceName: "uds/testing",
			udsServerDevType: "uds/testing",
			fakePodDevices:   []string{"devA", "devB"},
			udsServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: constants.Uds.Handshake.RequestConnect + ", podA, podB",
				1: constants.Uds.Handshake.RequestFin,
			},
			expectedResponse: map[int]string{
				0: constants.Uds.Handshake.ResponseHostNak,
				1: "should not get " + constants.Uds.Handshake.ResponseFinAck + " as should not have connected",
			},
		},
		{
			//Try connect 5 hostnames
			testName:         "Many hostnames",
			fakePodName:      "podA",
			fakePodNamespace: "default",
			fakeResourceName: "uds/testing",
			udsServerDevType: "uds/testing",
			fakePodDevices:   []string{"devA", "devB"},
			udsServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: constants.Uds.Handshake.RequestConnect + ", podA, podB, podC, podD, podE",
				1: constants.Uds.Handshake.RequestFin,
			},
			expectedResponse: map[int]string{
				0: constants.Uds.Handshake.ResponseHostNak,
				1: "should not get " + constants.Uds.Handshake.ResponseFinAck + " as should not have connected",
			},
		},
		{
			//Put the podname before connect request
			testName:         "Hostname before request",
			fakePodName:      "podA",
			fakePodNamespace: "default",
			fakeResourceName: "uds/testing",
			udsServerDevType: "uds/testing",
			fakePodDevices:   []string{"devA", "devB"},
			udsServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: "podA, " + constants.Uds.Handshake.RequestConnect,
				1: constants.Uds.Handshake.RequestFin,
			},
			expectedResponse: map[int]string{
				0: constants.Uds.Handshake.ResponseHostNak,
				1: "should not get " + constants.Uds.Handshake.ResponseFinAck + " as should not have connected",
			},
		},
		{
			//Put the podname before and after connect request
			testName:         "Hostname before and after request",
			fakePodName:      "podA",
			fakePodNamespace: "default",
			fakeResourceName: "uds/testing",
			udsServerDevType: "uds/testing",
			fakePodDevices:   []string{"devA", "devB"},
			udsServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: "podA, " + constants.Uds.Handshake.RequestConnect + ", podA",
				1: constants.Uds.Handshake.RequestFin,
			},
			expectedResponse: map[int]string{
				0: constants.Uds.Handshake.ResponseHostNak,
				1: "should not get " + constants.Uds.Handshake.ResponseFinAck + " as should not have connected",
			},
		},
		{
			//Put a null byte before an otherwise good connect request
			testName:         "Null byte before good connect request (1)",
			fakePodName:      "podA",
			fakePodNamespace: "default",
			fakeResourceName: "uds/testing",
			udsServerDevType: "uds/testing",
			fakePodDevices:   []string{"devA", "devB"},
			udsServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: `\0` + constants.Uds.Handshake.RequestConnect + ", podA",
				1: constants.Uds.Handshake.RequestFin,
			},
			expectedResponse: map[int]string{
				0: constants.Uds.Handshake.ResponseHostNak,
				1: "should not get " + constants.Uds.Handshake.ResponseFinAck + " as should not have connected",
			},
		},
		{
			//Put a null byte an otherwise good connect request
			testName:         "Null byte before good connect request (2)",
			fakePodName:      "podA",
			fakePodNamespace: "default",
			fakeResourceName: "uds/testing",
			udsServerDevType: "uds/testing",
			fakePodDevices:   []string{"devA", "devB"},
			udsServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: "\\0" + constants.Uds.Handshake.RequestConnect + ", podA",
				1: constants.Uds.Handshake.RequestFin,
			},
			expectedResponse: map[int]string{
				0: constants.Uds.Handshake.ResponseHostNak,
				1: "should not get " + constants.Uds.Handshake.ResponseFinAck + " as should not have connected",
			},
		},
		{
			//Put garbage before an otherwise good connect request
			testName:         "Garbage before good connect request (1)",
			fakePodName:      "podA",
			fakePodNamespace: "default",
			fakeResourceName: "uds/testing",
			udsServerDevType: "uds/testing",
			fakePodDevices:   []string{"devA", "devB"},
			udsServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: "dkjfhgkjdfs" + constants.Uds.Handshake.RequestConnect + ", podA",
				1: constants.Uds.Handshake.RequestFin,
			},
			expectedResponse: map[int]string{
				0: constants.Uds.Handshake.ResponseHostNak,
				1: "should not get " + constants.Uds.Handshake.ResponseFinAck + " as should not have connected",
			},
		},
		{
			//Put garbage before an otherwise good connect request
			testName:         "Garbage before good connect request (2)",
			fakePodName:      "podA",
			fakePodNamespace: "default",
			fakeResourceName: "uds/testing",
			udsServerDevType: "uds/testing",
			fakePodDevices:   []string{"devA", "devB"},
			udsServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: " " + constants.Uds.Handshake.RequestConnect + ", podA",
				1: constants.Uds.Handshake.RequestFin,
			},
			expectedResponse: map[int]string{
				0: constants.Uds.Handshake.ResponseHostNak,
				1: "should not get " + constants.Uds.Handshake.ResponseFinAck + " as should not have connected",
			},
		},
		{
			//Put garbage before an otherwise good connect request
			testName:         "Garbage before good connect request (3)",
			fakePodName:      "podA",
			fakePodNamespace: "default",
			fakeResourceName: "uds/testing",
			udsServerDevType: "uds/testing",
			fakePodDevices:   []string{"devA", "devB"},
			udsServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: `\` + constants.Uds.Handshake.RequestConnect + ", podA",
				1: constants.Uds.Handshake.RequestFin,
			},
			expectedResponse: map[int]string{
				0: constants.Uds.Handshake.ResponseHostNak,
				1: "should not get " + constants.Uds.Handshake.ResponseFinAck + " as should not have connected",
			},
		},
		{
			//Put garbage before an otherwise good connect request
			testName:         "Garbage before good connect request (4)",
			fakePodName:      "podA",
			fakePodNamespace: "default",
			fakeResourceName: "uds/testing",
			udsServerDevType: "uds/testing",
			fakePodDevices:   []string{"devA", "devB"},
			udsServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: "\\\\" + constants.Uds.Handshake.RequestConnect + ", podA",
				1: constants.Uds.Handshake.RequestFin,
			},
			expectedResponse: map[int]string{
				0: constants.Uds.Handshake.ResponseHostNak,
				1: "should not get " + constants.Uds.Handshake.ResponseFinAck + " as should not have connected",
			},
		},
		{
			//Put garbage before an otherwise good connect request
			testName:         "Garbage before good connect request (5)",
			fakePodName:      "podA",
			fakePodNamespace: "default",
			fakeResourceName: "uds/testing",
			udsServerDevType: "uds/testing",
			fakePodDevices:   []string{"devA", "devB"},
			udsServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: "," + constants.Uds.Handshake.RequestConnect + ", podA",
				1: constants.Uds.Handshake.RequestFin,
			},
			expectedResponse: map[int]string{
				0: constants.Uds.Handshake.ResponseHostNak,
				1: "should not get " + constants.Uds.Handshake.ResponseFinAck + " as should not have connected",
			},
		},
		{
			//Put a null byte after an otherwise good connect request
			testName:         "Null byte after good connect request (1)",
			fakePodName:      "podA",
			fakePodNamespace: "default",
			fakeResourceName: "uds/testing",
			udsServerDevType: "uds/testing",
			fakePodDevices:   []string{"devA", "devB"},
			udsServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: constants.Uds.Handshake.RequestConnect + ", podA" + `\0`,
				1: constants.Uds.Handshake.RequestFin,
			},
			expectedResponse: map[int]string{
				0: constants.Uds.Handshake.ResponseHostNak,
				1: "should not get " + constants.Uds.Handshake.ResponseFinAck + " as should not have connected",
			},
		},
		{
			//Put a null byte after an otherwise good connect request
			testName:         "Null byte after good connect request (2)",
			fakePodName:      "podA",
			fakePodNamespace: "default",
			fakeResourceName: "uds/testing",
			udsServerDevType: "uds/testing",
			fakePodDevices:   []string{"devA", "devB"},
			udsServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: constants.Uds.Handshake.RequestConnect + ", podA" + "\\0",
				1: constants.Uds.Handshake.RequestFin,
			},
			expectedResponse: map[int]string{
				0: constants.Uds.Handshake.ResponseHostNak,
				1: "should not get " + constants.Uds.Handshake.ResponseFinAck + " as should not have connected",
			},
		},
		{
			//Put garbage after an otherwise good connect request
			testName:         "Garbage after good connect request (1)",
			fakePodName:      "podA",
			fakePodNamespace: "default",
			fakeResourceName: "uds/testing",
			udsServerDevType: "uds/testing",
			fakePodDevices:   []string{"devA", "devB"},
			udsServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: constants.Uds.Handshake.RequestConnect + ", podAdkjfhgkjdfs",
				1: constants.Uds.Handshake.RequestFin,
			},
			expectedResponse: map[int]string{
				0: constants.Uds.Handshake.ResponseHostNak,
				1: "should not get " + constants.Uds.Handshake.ResponseFinAck + " as should not have connected",
			},
		},
		{
			//Put garbage after an otherwise good connect request
			testName:         "Garbage after good connect request (2)",
			fakePodName:      "podA",
			fakePodNamespace: "default",
			fakeResourceName: "uds/testing",
			udsServerDevType: "uds/testing",
			fakePodDevices:   []string{"devA", "devB"},
			udsServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: constants.Uds.Handshake.RequestConnect + ", podA,dkjfhgkjdfs",
				1: constants.Uds.Handshake.RequestFin,
			},
			expectedResponse: map[int]string{
				0: constants.Uds.Handshake.ResponseHostNak,
				1: "should not get " + constants.Uds.Handshake.ResponseFinAck + " as should not have connected",
			},
		},
		{
			//Put garbage after an otherwise good connect request
			testName:         "Garbage after good connect request (3)",
			fakePodName:      "podA",
			fakePodNamespace: "default",
			fakeResourceName: "uds/testing",
			udsServerDevType: "uds/testing",
			fakePodDevices:   []string{"devA", "devB"},
			udsServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: constants.Uds.Handshake.RequestConnect + ", podA" + `\`,
				1: constants.Uds.Handshake.RequestFin,
			},
			expectedResponse: map[int]string{
				0: constants.Uds.Handshake.ResponseHostNak,
				1: "should not get " + constants.Uds.Handshake.ResponseFinAck + " as should not have connected",
			},
		},
		{
			//Put garbage after an otherwise good connect request
			testName:         "Garbage after good connect request (4)",
			fakePodName:      "podA",
			fakePodNamespace: "default",
			fakeResourceName: "uds/testing",
			udsServerDevType: "uds/testing",
			fakePodDevices:   []string{"devA", "devB"},
			udsServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: constants.Uds.Handshake.RequestConnect + ", podA\\\\",
				1: constants.Uds.Handshake.RequestFin,
			},
			expectedResponse: map[int]string{
				0: constants.Uds.Handshake.ResponseHostNak,
				1: "should not get " + constants.Uds.Handshake.ResponseFinAck + " as should not have connected",
			},
		},
		{
			//Put garbage after an otherwise good connect request
			testName:         "Garbage after good connect request (5)",
			fakePodName:      "podA",
			fakePodNamespace: "default",
			fakeResourceName: "uds/testing",
			udsServerDevType: "uds/testing",
			fakePodDevices:   []string{"devA", "devB"},
			udsServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: constants.Uds.Handshake.RequestConnect + ", podA,",
				1: constants.Uds.Handshake.RequestFin,
			},
			expectedResponse: map[int]string{
				0: constants.Uds.Handshake.ResponseHostNak,
				1: "should not get " + constants.Uds.Handshake.ResponseFinAck + " as should not have connected",
			},
		},
		{
			//Put null byte within otherwise good connect request
			testName:         "Null byte within good connect request (1)",
			fakePodName:      "podA",
			fakePodNamespace: "default",
			fakeResourceName: "uds/testing",
			udsServerDevType: "uds/testing",
			fakePodDevices:   []string{"devA", "devB"},
			udsServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: constants.Uds.Handshake.RequestConnect + `\0` + ", podA",
				1: constants.Uds.Handshake.RequestFin,
			},
			expectedResponse: map[int]string{
				0: constants.Uds.Handshake.ResponseHostNak,
				1: "should not get " + constants.Uds.Handshake.ResponseFinAck + " as should not have connected",
			},
		},
		{
			//Put null byte within otherwise good connect request
			testName:         "Null byte within good connect request (2)",
			fakePodName:      "podA",
			fakePodNamespace: "default",
			fakeResourceName: "uds/testing",
			udsServerDevType: "uds/testing",
			fakePodDevices:   []string{"devA", "devB"},
			udsServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: constants.Uds.Handshake.RequestConnect + "\\0" + ", podA",
				1: constants.Uds.Handshake.RequestFin,
			},
			expectedResponse: map[int]string{
				0: constants.Uds.Handshake.ResponseHostNak,
				1: "should not get " + constants.Uds.Handshake.ResponseFinAck + " as should not have connected",
			},
		},
		{
			//Null byte before connect
			testName:         "Send a null byte (1)",
			fakePodName:      "podA",
			fakePodNamespace: "default",
			fakeResourceName: "uds/testing",
			udsServerDevType: "uds/testing",
			fakePodDevices:   []string{"devA", "devB"},
			udsServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: `\0`,
				1: constants.Uds.Handshake.RequestFin,
			},
			expectedResponse: map[int]string{
				0: constants.Uds.Handshake.ResponseHostNak,
				1: "should not get " + constants.Uds.Handshake.ResponseFinAck + " as should not have connected",
			},
		},
		{
			//Null byte before connect
			testName:         "Send a null byte (2)",
			fakePodName:      "podA",
			fakePodNamespace: "default",
			fakeResourceName: "uds/testing",
			udsServerDevType: "uds/testing",
			fakePodDevices:   []string{"devA", "devB"},
			udsServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: "\\0",
				1: constants.Uds.Handshake.RequestFin,
			},
			expectedResponse: map[int]string{
				0: constants.Uds.Handshake.ResponseHostNak,
				1: "should not get " + constants.Uds.Handshake.ResponseFinAck + " as should not have connected",
			},
		},
		{
			//Garbage request before connect
			testName:         "Garbage request, before connect (1)",
			fakePodName:      "podA",
			fakePodNamespace: "default",
			fakeResourceName: "uds/testing",
			udsServerDevType: "uds/testing",
			fakePodDevices:   []string{"devA", "devB"},
			udsServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: "asdfxc",
				1: constants.Uds.Handshake.RequestFin,
			},
			expectedResponse: map[int]string{
				0: constants.Uds.Handshake.ResponseBadRequest,
				1: "should not get " + constants.Uds.Handshake.ResponseFinAck + " as should not have connected",
			},
		},
		{
			//Garbage request before connect
			testName:         "Garbage request, before connect (2)",
			fakePodName:      "podA",
			fakePodNamespace: "default",
			fakeResourceName: "uds/testing",
			udsServerDevType: "uds/testing",
			fakePodDevices:   []string{"devA", "devB"},
			udsServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: "asdfxc,",
				1: constants.Uds.Handshake.RequestFin,
			},
			expectedResponse: map[int]string{
				0: constants.Uds.Handshake.ResponseBadRequest,
				1: "should not get " + constants.Uds.Handshake.ResponseFinAck + " as should not have connected",
			},
		},
		{
			//Garbage request before connect
			testName:         "Garbage request, before connect (3)",
			fakePodName:      "podA",
			fakePodNamespace: "default",
			fakeResourceName: "uds/testing",
			udsServerDevType: "uds/testing",
			fakePodDevices:   []string{"devA", "devB"},
			udsServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: "asdfxc, podA,",
				1: constants.Uds.Handshake.RequestFin,
			},
			expectedResponse: map[int]string{
				0: constants.Uds.Handshake.ResponseBadRequest,
				1: "should not get " + constants.Uds.Handshake.ResponseFinAck + " as should not have connected",
			},
		},
		{
			//Garbage request before connect
			testName:         "Garbage request, before connect (4)",
			fakePodName:      "podA",
			fakePodNamespace: "default",
			fakeResourceName: "uds/testing",
			udsServerDevType: "uds/testing",
			fakePodDevices:   []string{"devA", "devB"},
			udsServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: `\`,
				1: constants.Uds.Handshake.RequestFin,
			},
			expectedResponse: map[int]string{
				0: constants.Uds.Handshake.ResponseBadRequest,
				1: "should not get " + constants.Uds.Handshake.ResponseFinAck + " as should not have connected",
			},
		},
		{
			//Garbage request before connect
			testName:         "Garbage request, before connect (5)",
			fakePodName:      "podA",
			fakePodNamespace: "default",
			fakeResourceName: "uds/testing",
			udsServerDevType: "uds/testing",
			fakePodDevices:   []string{"devA", "devB"},
			udsServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: "\\\\",
				1: constants.Uds.Handshake.RequestFin,
			},
			expectedResponse: map[int]string{
				0: constants.Uds.Handshake.ResponseBadRequest,
				1: "should not get " + constants.Uds.Handshake.ResponseFinAck + " as should not have connected",
			},
		},
		{
			//Garbage request before connect
			testName:         "Garbage request, before connect (6)",
			fakePodName:      "podA",
			fakePodNamespace: "default",
			fakeResourceName: "uds/testing",
			udsServerDevType: "uds/testing",
			fakePodDevices:   []string{"devA", "devB"},
			udsServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: "*",
				1: constants.Uds.Handshake.RequestFin,
			},
			expectedResponse: map[int]string{
				0: constants.Uds.Handshake.ResponseBadRequest,
				1: "should not get " + constants.Uds.Handshake.ResponseFinAck + " as should not have connected",
			},
		},
		{
			//Send variant of the accepted /connect request
			testName:         "Bad connect request (1)",
			fakePodName:      "podA",
			fakePodNamespace: "default",
			fakeResourceName: "uds/testing",
			udsServerDevType: "uds/testing",
			fakePodDevices:   []string{"devA", "devB"},
			udsServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: "connect",
				1: constants.Uds.Handshake.RequestFin,
			},
			expectedResponse: map[int]string{
				0: constants.Uds.Handshake.ResponseBadRequest,
				1: "should not get " + constants.Uds.Handshake.ResponseFinAck + " as should not have connected",
			},
		},
		{
			//Send variant of the accepted /connect request
			testName:         "Bad connect request (2)",
			fakePodName:      "podA",
			fakePodNamespace: "default",
			fakeResourceName: "uds/testing",
			udsServerDevType: "uds/testing",
			fakePodDevices:   []string{"devA", "devB"},
			udsServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: "connect/",
				1: constants.Uds.Handshake.RequestFin,
			},
			expectedResponse: map[int]string{
				0: constants.Uds.Handshake.ResponseBadRequest,
				1: "should not get " + constants.Uds.Handshake.ResponseFinAck + " as should not have connected",
			},
		},
		{
			//Send variant of the accepted /connect request
			testName:         "Bad connect request (3)",
			fakePodName:      "podA",
			fakePodNamespace: "default",
			fakeResourceName: "uds/testing",
			udsServerDevType: "uds/testing",
			fakePodDevices:   []string{"devA", "devB"},
			udsServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: "/Connect",
				1: constants.Uds.Handshake.RequestFin,
			},
			expectedResponse: map[int]string{
				0: constants.Uds.Handshake.ResponseBadRequest,
				1: "should not get " + constants.Uds.Handshake.ResponseFinAck + " as should not have connected",
			},
		},
		{
			//Send variant of the accepted /connect request
			testName:         "Bad connect request (4)",
			fakePodName:      "podA",
			fakePodNamespace: "default",
			fakeResourceName: "uds/testing",
			udsServerDevType: "uds/testing",
			fakePodDevices:   []string{"devA", "devB"},
			udsServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: "/connect*",
				1: constants.Uds.Handshake.RequestFin,
			},
			expectedResponse: map[int]string{
				0: constants.Uds.Handshake.ResponseHostNak,
				1: "should not get " + constants.Uds.Handshake.ResponseFinAck + " as should not have connected",
			},
		},
		{
			//Send variant of the accepted /connect request
			testName:         "Bad connect request (5)",
			fakePodName:      "podA",
			fakePodNamespace: "default",
			fakeResourceName: "uds/testing",
			udsServerDevType: "uds/testing",
			fakePodDevices:   []string{"devA", "devB"},
			udsServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: `\/connect*`,
				1: constants.Uds.Handshake.RequestFin,
			},
			expectedResponse: map[int]string{
				0: constants.Uds.Handshake.ResponseHostNak,
				1: "should not get " + constants.Uds.Handshake.ResponseFinAck + " as should not have connected",
			},
		},
		{
			//Send variant of the accepted /connect request
			testName:         "Bad connect request (6)",
			fakePodName:      "podA",
			fakePodNamespace: "default",
			fakeResourceName: "uds/testing",
			udsServerDevType: "uds/testing",
			fakePodDevices:   []string{"devA", "devB"},
			udsServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: "*/connect*",
				1: constants.Uds.Handshake.RequestFin,
			},
			expectedResponse: map[int]string{
				0: constants.Uds.Handshake.ResponseHostNak,
				1: "should not get " + constants.Uds.Handshake.ResponseFinAck + " as should not have connected",
			},
		},
		{
			//Try connect podX but we only know podA. podX has 2 good devices - devA and devB
			testName:         "Bad hostname, Good Devices",
			fakePodName:      "podA",
			fakePodNamespace: "default",
			fakeResourceName: "uds/testing",
			udsServerDevType: "uds/testing",
			fakePodDevices:   []string{"devA", "devB"},
			udsServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: constants.Uds.Handshake.RequestConnect + ", podX",
				1: constants.Uds.Handshake.RequestFin,
			},
			expectedResponse: map[int]string{
				0: constants.Uds.Handshake.ResponseHostNak,
				1: "should not get " + constants.Uds.Handshake.ResponseFinAck + " as should not have connected",
			},
		},
		{
			//Try connect good podA, but we have 2 bad devices - devX and devY
			testName:         "Good hostname, Bad Devices",
			fakePodName:      "podA",
			fakePodNamespace: "default",
			fakeResourceName: "uds/testing",
			udsServerDevType: "uds/testing",
			fakePodDevices:   []string{"devA", "devB"},
			udsServerDevices: []string{"devX", "devY"},
			fakeRequests: map[int]string{
				0: constants.Uds.Handshake.RequestConnect + ", podA",
				1: constants.Uds.Handshake.RequestFin,
			},
			expectedResponse: map[int]string{
				0: constants.Uds.Handshake.ResponseHostNak,
				1: "should not get " + constants.Uds.Handshake.ResponseFinAck + " as should not have connected",
			},
		},
		{
			//Try connect good podA, but 1 of 4 devices is bad - devX
			testName:         "Good hostname, Single bad device (1)",
			fakePodName:      "podA",
			fakePodNamespace: "default",
			fakeResourceName: "uds/testing",
			udsServerDevType: "uds/testing",
			fakePodDevices:   []string{"devA", "devB", "devC", "devD"},
			udsServerDevices: []string{"devA", "devB", "devC", "devX"},
			fakeRequests: map[int]string{
				0: constants.Uds.Handshake.RequestConnect + ", podA",
				1: constants.Uds.Handshake.RequestFin,
			},
			expectedResponse: map[int]string{
				0: constants.Uds.Handshake.ResponseHostNak,
				1: "should not get " + constants.Uds.Handshake.ResponseFinAck + " as should not have connected",
			},
		},
		{
			//Try connect good podA, but 1 of 4 devices is bad - devX
			testName:         "Good hostname, Single bad device (2)",
			fakePodName:      "podA",
			fakePodNamespace: "default",
			fakeResourceName: "uds/testing",
			udsServerDevType: "uds/testing",
			fakePodDevices:   []string{"devA", "devB", "devC", "devD"},
			udsServerDevices: []string{"devX", "devB", "devC", "devD"},
			fakeRequests: map[int]string{
				0: constants.Uds.Handshake.RequestConnect + ", podA",
				1: constants.Uds.Handshake.RequestFin,
			},
			expectedResponse: map[int]string{
				0: constants.Uds.Handshake.ResponseHostNak,
				1: "should not get " + constants.Uds.Handshake.ResponseFinAck + " as should not have connected",
			},
		},
		{
			//Try connect good podA, but 1 of 4 devices is bad - devX
			testName:         "Good hostname, Single bad device (3)",
			fakePodName:      "podA",
			fakePodNamespace: "default",
			fakeResourceName: "uds/testing",
			udsServerDevType: "uds/testing",
			fakePodDevices:   []string{"devA", "devB", "devC", "devD"},
			udsServerDevices: []string{"devA", "devB", "devX", "devD"},
			fakeRequests: map[int]string{
				0: constants.Uds.Handshake.RequestConnect + ", podA",
				1: constants.Uds.Handshake.RequestFin,
			},
			expectedResponse: map[int]string{
				0: constants.Uds.Handshake.ResponseHostNak,
				1: "should not get " + constants.Uds.Handshake.ResponseFinAck + " as should not have connected",
			},
		},
		{
			//Try connect good podA, all 4 devices are good, but we have one extra - devX
			testName:         "Good hostname, good devices, 1 extra device (1)",
			fakePodName:      "podA",
			fakePodNamespace: "default",
			fakeResourceName: "uds/testing",
			udsServerDevType: "uds/testing",
			fakePodDevices:   []string{"devA", "devB", "devC", "devD"},
			udsServerDevices: []string{"devA", "devB", "devC", "devD", "devX"},
			fakeRequests: map[int]string{
				0: constants.Uds.Handshake.RequestConnect + ", podA",
				1: constants.Uds.Handshake.RequestFin,
			},
			expectedResponse: map[int]string{
				0: constants.Uds.Handshake.ResponseHostNak,
				1: "should not get " + constants.Uds.Handshake.ResponseFinAck + " as should not have connected",
			},
		},
		{
			//Try connect good podA, all devices are good, but we are missing one - devD
			testName:         "Good hostname, good devices, 1 missing device",
			fakePodName:      "podA",
			fakePodNamespace: "default",
			fakeResourceName: "uds/testing",
			udsServerDevType: "uds/testing",
			fakePodDevices:   []string{"devA", "devB", "devC", "devD"},
			udsServerDevices: []string{"devA", "devB", "devC"},
			fakeRequests: map[int]string{
				0: constants.Uds.Handshake.RequestConnect + ", podA",
				1: constants.Uds.Handshake.RequestFin,
			},
			expectedResponse: map[int]string{
				0: constants.Uds.Handshake.ResponseHostNak,
				1: "should not get " + constants.Uds.Handshake.ResponseFinAck + " as should not have connected",
			},
		},
		{
			//Try connect good podA, both devices are good, but they are of the wrong type - uds/badType
			testName:         "Good hostname, good device, bad device type",
			fakePodName:      "podA",
			fakePodNamespace: "default",
			fakeResourceName: "uds/testing",
			udsServerDevType: "uds/badType",
			fakePodDevices:   []string{"devA", "devB"},
			udsServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: constants.Uds.Handshake.RequestConnect + ", podA",
				1: constants.Uds.Handshake.RequestFin,
			},
			expectedResponse: map[int]string{
				0: constants.Uds.Handshake.ResponseHostNak,
				1: "should not get " + constants.Uds.Handshake.ResponseFinAck + " as should not have connected",
			},
		},
		/****************************************************************
		Negative Tests - validate, but request FDs for devs we don't have
		****************************************************************/
		{
			//Connect podA, request FD device it does not have - devX
			testName:         "Connect and request bad FD, 1 device",
			fakePodName:      "podA",
			fakePodNamespace: "default",
			fakeResourceName: "uds/testing",
			udsServerDevType: "uds/testing",
			fakePodDevices:   []string{"devA"},
			udsServerDevices: []string{"devA"},
			fakeRequests: map[int]string{
				0: constants.Uds.Handshake.RequestConnect + ", podA",
				1: constants.Uds.Handshake.RequestFd + ", devX",
				2: constants.Uds.Handshake.RequestFin,
			},
			expectedResponse: map[int]string{
				0: constants.Uds.Handshake.ResponseHostOk,
				1: constants.Uds.Handshake.ResponseFdNak,
				2: constants.Uds.Handshake.ResponseFinAck,
			},
		},
		{
			//Connect podA, request one bad FD out of 3 - devX
			testName:         "Connect and request 2 good and 1 bad FD",
			fakePodName:      "podA",
			fakePodNamespace: "default",
			fakeResourceName: "uds/testing",
			udsServerDevType: "uds/testing",
			fakePodDevices:   []string{"devA", "devB", "devC"},
			udsServerDevices: []string{"devA", "devB", "devC"},
			fakeRequests: map[int]string{
				0: constants.Uds.Handshake.RequestConnect + ", podA",
				1: constants.Uds.Handshake.RequestFd + ", devA",
				2: constants.Uds.Handshake.RequestFd + ", devX",
				3: constants.Uds.Handshake.RequestFd + ", devC",
				4: constants.Uds.Handshake.RequestFin,
			},
			expectedResponse: map[int]string{
				0: constants.Uds.Handshake.ResponseHostOk,
				1: constants.Uds.Handshake.ResponseFdAck,
				2: constants.Uds.Handshake.ResponseFdNak,
				3: constants.Uds.Handshake.ResponseFdAck,
				4: constants.Uds.Handshake.ResponseFinAck,
			},
		},
		/******************************************************
		Negative Tests - validate, but send corrupt FD requests
		******************************************************/
		{
			//Send FD request without any device
			testName:         "Request no FD (1)",
			fakePodName:      "podA",
			fakePodNamespace: "default",
			fakeResourceName: "uds/testing",
			udsServerDevType: "uds/testing",
			fakePodDevices:   []string{"devA"},
			udsServerDevices: []string{"devA"},
			fakeRequests: map[int]string{
				0: constants.Uds.Handshake.RequestConnect + ", podA",
				1: constants.Uds.Handshake.RequestFd,
				2: constants.Uds.Handshake.RequestFin,
			},
			expectedResponse: map[int]string{
				0: constants.Uds.Handshake.ResponseHostOk,
				1: constants.Uds.Handshake.ResponseBadRequest,
				2: constants.Uds.Handshake.ResponseFinAck,
			},
		},
		{
			//Send FD request without any device, include comma
			testName:         "Request no FD (2)",
			fakePodName:      "podA",
			fakePodNamespace: "default",
			fakeResourceName: "uds/testing",
			udsServerDevType: "uds/testing",
			fakePodDevices:   []string{"devA"},
			udsServerDevices: []string{"devA"},
			fakeRequests: map[int]string{
				0: constants.Uds.Handshake.RequestConnect + ", podA",
				1: constants.Uds.Handshake.RequestFd + ",",
				2: constants.Uds.Handshake.RequestFin,
			},
			expectedResponse: map[int]string{
				0: constants.Uds.Handshake.ResponseHostOk,
				1: constants.Uds.Handshake.ResponseFdNak,
				2: constants.Uds.Handshake.ResponseFinAck,
			},
		},
		{
			//Send FD request for 2 devs
			testName:         "Request 2 FD",
			fakePodName:      "podA",
			fakePodNamespace: "default",
			fakeResourceName: "uds/testing",
			udsServerDevType: "uds/testing",
			fakePodDevices:   []string{"devA", "devB"},
			udsServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: constants.Uds.Handshake.RequestConnect + ", podA",
				1: constants.Uds.Handshake.RequestFd + ", devA, devB",
				2: constants.Uds.Handshake.RequestFin,
			},
			expectedResponse: map[int]string{
				0: constants.Uds.Handshake.ResponseHostOk,
				1: constants.Uds.Handshake.ResponseBadRequest,
				2: constants.Uds.Handshake.ResponseFinAck,
			},
		},
		{
			//Send dev before FD request
			testName:         "Device before FD request",
			fakePodName:      "podA",
			fakePodNamespace: "default",
			fakeResourceName: "uds/testing",
			udsServerDevType: "uds/testing",
			fakePodDevices:   []string{"devA", "devB"},
			udsServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: constants.Uds.Handshake.RequestConnect + ", podA",
				1: "devA, " + constants.Uds.Handshake.RequestFd,
				2: constants.Uds.Handshake.RequestFin,
			},
			expectedResponse: map[int]string{
				0: constants.Uds.Handshake.ResponseHostOk,
				1: constants.Uds.Handshake.ResponseBadRequest,
				2: constants.Uds.Handshake.ResponseFinAck,
			},
		},
		{
			//Send dev before and after FD request
			testName:         "Device before and after FD request",
			fakePodName:      "podA",
			fakePodNamespace: "default",
			fakeResourceName: "uds/testing",
			udsServerDevType: "uds/testing",
			fakePodDevices:   []string{"devA", "devB"},
			udsServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: constants.Uds.Handshake.RequestConnect + ", podA",
				1: "devA, " + constants.Uds.Handshake.RequestFd + ", devA",
				2: constants.Uds.Handshake.RequestFin,
			},
			expectedResponse: map[int]string{
				0: constants.Uds.Handshake.ResponseHostOk,
				1: constants.Uds.Handshake.ResponseBadRequest,
				2: constants.Uds.Handshake.ResponseFinAck,
			},
		},
		{
			//Put null byte before an otherwise good FD request
			testName:         "Null byte before good FD request (1)",
			fakePodName:      "podA",
			fakePodNamespace: "default",
			fakeResourceName: "uds/testing",
			udsServerDevType: "uds/testing",
			fakePodDevices:   []string{"devA", "devB"},
			udsServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: constants.Uds.Handshake.RequestConnect + ", podA",
				1: `\0` + constants.Uds.Handshake.RequestFd + ", devA",
				2: constants.Uds.Handshake.RequestFin,
			},
			expectedResponse: map[int]string{
				0: constants.Uds.Handshake.ResponseHostOk,
				1: constants.Uds.Handshake.ResponseBadRequest,
				2: constants.Uds.Handshake.ResponseFinAck,
			},
		},
		{
			//Put null byte before an otherwise good FD request
			testName:         "Null byte before good FD request (2)",
			fakePodName:      "podA",
			fakePodNamespace: "default",
			fakeResourceName: "uds/testing",
			udsServerDevType: "uds/testing",
			fakePodDevices:   []string{"devA", "devB"},
			udsServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: constants.Uds.Handshake.RequestConnect + ", podA",
				1: "\\0" + constants.Uds.Handshake.RequestFd + ", devA",
				2: constants.Uds.Handshake.RequestFin,
			},
			expectedResponse: map[int]string{
				0: constants.Uds.Handshake.ResponseHostOk,
				1: constants.Uds.Handshake.ResponseBadRequest,
				2: constants.Uds.Handshake.ResponseFinAck,
			},
		},
		{
			//Put garbage before an otherwise good FD request
			testName:         "Garbage before good FD request (1)",
			fakePodName:      "podA",
			fakePodNamespace: "default",
			fakeResourceName: "uds/testing",
			udsServerDevType: "uds/testing",
			fakePodDevices:   []string{"devA", "devB"},
			udsServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: constants.Uds.Handshake.RequestConnect + ", podA",
				1: "dkjfhgkjdfs" + constants.Uds.Handshake.RequestFd + ", devA",
				2: constants.Uds.Handshake.RequestFin,
			},
			expectedResponse: map[int]string{
				0: constants.Uds.Handshake.ResponseHostOk,
				1: constants.Uds.Handshake.ResponseBadRequest,
				2: constants.Uds.Handshake.ResponseFinAck,
			},
		},
		{
			//Put garbage before an otherwise good FD request
			testName:         "Garbage before good FD request (2)",
			fakePodName:      "podA",
			fakePodNamespace: "default",
			fakeResourceName: "uds/testing",
			udsServerDevType: "uds/testing",
			fakePodDevices:   []string{"devA", "devB"},
			udsServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: constants.Uds.Handshake.RequestConnect + ", podA",
				1: " " + constants.Uds.Handshake.RequestFd + ", devA",
				2: constants.Uds.Handshake.RequestFin,
			},
			expectedResponse: map[int]string{
				0: constants.Uds.Handshake.ResponseHostOk,
				1: constants.Uds.Handshake.ResponseBadRequest,
				2: constants.Uds.Handshake.ResponseFinAck,
			},
		},
		{
			//Put garbage before an otherwise good FD request
			testName:         "Garbage before good FD request (3)",
			fakePodName:      "podA",
			fakePodNamespace: "default",
			fakeResourceName: "uds/testing",
			udsServerDevType: "uds/testing",
			fakePodDevices:   []string{"devA", "devB"},
			udsServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: constants.Uds.Handshake.RequestConnect + ", podA",
				1: `\` + constants.Uds.Handshake.RequestFd + ", devA",
				2: constants.Uds.Handshake.RequestFin,
			},
			expectedResponse: map[int]string{
				0: constants.Uds.Handshake.ResponseHostOk,
				1: constants.Uds.Handshake.ResponseBadRequest,
				2: constants.Uds.Handshake.ResponseFinAck,
			},
		},
		{
			//Put garbage before an otherwise good FD request
			testName:         "Garbage before good FD request (4)",
			fakePodName:      "podA",
			fakePodNamespace: "default",
			fakeResourceName: "uds/testing",
			udsServerDevType: "uds/testing",
			fakePodDevices:   []string{"devA", "devB"},
			udsServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: constants.Uds.Handshake.RequestConnect + ", podA",
				1: "\\\\" + constants.Uds.Handshake.RequestFd + ", devA",
				2: constants.Uds.Handshake.RequestFin,
			},
			expectedResponse: map[int]string{
				0: constants.Uds.Handshake.ResponseHostOk,
				1: constants.Uds.Handshake.ResponseBadRequest,
				2: constants.Uds.Handshake.ResponseFinAck,
			},
		},
		{
			//Put garbage before an otherwise good FD request
			testName:         "Garbage before good FD request (5)",
			fakePodName:      "podA",
			fakePodNamespace: "default",
			fakeResourceName: "uds/testing",
			udsServerDevType: "uds/testing",
			fakePodDevices:   []string{"devA", "devB"},
			udsServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: constants.Uds.Handshake.RequestConnect + ", podA",
				1: "," + constants.Uds.Handshake.RequestFd + ", devA",
				2: constants.Uds.Handshake.RequestFin,
			},
			expectedResponse: map[int]string{
				0: constants.Uds.Handshake.ResponseHostOk,
				1: constants.Uds.Handshake.ResponseBadRequest,
				2: constants.Uds.Handshake.ResponseFinAck,
			},
		},
		{
			//Put null byte after an otherwise good FD request
			testName:         "Null byte after good FD request (1)",
			fakePodName:      "podA",
			fakePodNamespace: "default",
			fakeResourceName: "uds/testing",
			udsServerDevType: "uds/testing",
			fakePodDevices:   []string{"devA", "devB"},
			udsServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: constants.Uds.Handshake.RequestConnect + ", podA",
				1: constants.Uds.Handshake.RequestFd + "," + `\0`,
				2: constants.Uds.Handshake.RequestFin,
			},
			expectedResponse: map[int]string{
				0: constants.Uds.Handshake.ResponseHostOk,
				1: constants.Uds.Handshake.ResponseFdNak,
				2: constants.Uds.Handshake.ResponseFinAck,
			},
		},
		{
			//Put null byte after an otherwise good FD request
			testName:         "Null byte after good FD request (2)",
			fakePodName:      "podA",
			fakePodNamespace: "default",
			fakeResourceName: "uds/testing",
			udsServerDevType: "uds/testing",
			fakePodDevices:   []string{"devA", "devB"},
			udsServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: constants.Uds.Handshake.RequestConnect + ", podA",
				1: constants.Uds.Handshake.RequestFd + "," + "\\0",
				2: constants.Uds.Handshake.RequestFin,
			},
			expectedResponse: map[int]string{
				0: constants.Uds.Handshake.ResponseHostOk,
				1: constants.Uds.Handshake.ResponseFdNak,
				2: constants.Uds.Handshake.ResponseFinAck,
			},
		},
		{
			//Put garbage after an otherwise good FD request
			testName:         "Garbage after good FD request (1)",
			fakePodName:      "podA",
			fakePodNamespace: "default",
			fakeResourceName: "uds/testing",
			udsServerDevType: "uds/testing",
			fakePodDevices:   []string{"devA", "devB"},
			udsServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: constants.Uds.Handshake.RequestConnect + ", podA",
				1: constants.Uds.Handshake.RequestFd + ", devAdkjfhgkjdfs",
				2: constants.Uds.Handshake.RequestFin,
			},
			expectedResponse: map[int]string{
				0: constants.Uds.Handshake.ResponseHostOk,
				1: constants.Uds.Handshake.ResponseFdNak,
				2: constants.Uds.Handshake.ResponseFinAck,
			},
		},
		{
			//Put garbage after an otherwise good FD request
			testName:         "Garbage after good FD request (2)",
			fakePodName:      "podA",
			fakePodNamespace: "default",
			fakeResourceName: "uds/testing",
			udsServerDevType: "uds/testing",
			fakePodDevices:   []string{"devA", "devB"},
			udsServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: constants.Uds.Handshake.RequestConnect + ", podA",
				1: constants.Uds.Handshake.RequestFd + ", devA,dkjfhgkjdfs",
				2: constants.Uds.Handshake.RequestFin,
			},
			expectedResponse: map[int]string{
				0: constants.Uds.Handshake.ResponseHostOk,
				1: constants.Uds.Handshake.ResponseBadRequest,
				2: constants.Uds.Handshake.ResponseFinAck,
			},
		},
		{
			//Put garbage after an otherwise good FD request
			testName:         "Garbage after good FD request (3)",
			fakePodName:      "podA",
			fakePodNamespace: "default",
			fakeResourceName: "uds/testing",
			udsServerDevType: "uds/testing",
			fakePodDevices:   []string{"devA", "devB"},
			udsServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: constants.Uds.Handshake.RequestConnect + ", podA",
				1: constants.Uds.Handshake.RequestFd + ", devA" + `\`,
				2: constants.Uds.Handshake.RequestFin,
			},
			expectedResponse: map[int]string{
				0: constants.Uds.Handshake.ResponseHostOk,
				1: constants.Uds.Handshake.ResponseFdNak,
				2: constants.Uds.Handshake.ResponseFinAck,
			},
		},
		{
			//Put garbage after an otherwise good FD request
			testName:         "Garbage after good FD request (4)",
			fakePodName:      "podA",
			fakePodNamespace: "default",
			fakeResourceName: "uds/testing",
			udsServerDevType: "uds/testing",
			fakePodDevices:   []string{"devA", "devB"},
			udsServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: constants.Uds.Handshake.RequestConnect + ", podA",
				1: constants.Uds.Handshake.RequestFd + ", devA\\\\",
				2: constants.Uds.Handshake.RequestFin,
			},
			expectedResponse: map[int]string{
				0: constants.Uds.Handshake.ResponseHostOk,
				1: constants.Uds.Handshake.ResponseFdNak,
				2: constants.Uds.Handshake.ResponseFinAck,
			},
		},
		{
			//Put garbage after an otherwise good FD request
			testName:         "Garbage after good FD request (5)",
			fakePodName:      "podA",
			fakePodNamespace: "default",
			fakeResourceName: "uds/testing",
			udsServerDevType: "uds/testing",
			fakePodDevices:   []string{"devA", "devB"},
			udsServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: constants.Uds.Handshake.RequestConnect + ", podA",
				1: constants.Uds.Handshake.RequestFd + ", devA,",
				2: constants.Uds.Handshake.RequestFin,
			},
			expectedResponse: map[int]string{
				0: constants.Uds.Handshake.ResponseHostOk,
				1: constants.Uds.Handshake.ResponseBadRequest,
				2: constants.Uds.Handshake.ResponseFinAck,
			},
		},
		/***************************************************
		Negative Tests - validate, but send garbage requests
		***************************************************/
		{
			//Validate but then send null byte
			testName:         "Validate and send null byte (1)",
			fakePodName:      "podA",
			fakePodNamespace: "default",
			fakeResourceName: "uds/testing",
			udsServerDevType: "uds/testing",
			fakePodDevices:   []string{"devA", "devB"},
			udsServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: constants.Uds.Handshake.RequestConnect + ", podA",
				1: `\0`,
				2: constants.Uds.Handshake.RequestFin,
			},
			expectedResponse: map[int]string{
				0: constants.Uds.Handshake.ResponseHostOk,
				1: constants.Uds.Handshake.ResponseBadRequest,
				2: constants.Uds.Handshake.ResponseFinAck,
			},
		},
		{
			//Validate but then send null byte
			testName:         "Validate and send null byte (2)",
			fakePodName:      "podA",
			fakePodNamespace: "default",
			fakeResourceName: "uds/testing",
			udsServerDevType: "uds/testing",
			fakePodDevices:   []string{"devA", "devB"},
			udsServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: constants.Uds.Handshake.RequestConnect + ", podA",
				1: "\\0",
				2: constants.Uds.Handshake.RequestFin,
			},
			expectedResponse: map[int]string{
				0: constants.Uds.Handshake.ResponseHostOk,
				1: constants.Uds.Handshake.ResponseBadRequest,
				2: constants.Uds.Handshake.ResponseFinAck,
			},
		},
		{
			//Validate but then send garbage request
			testName:         "Validate and send no request (1)",
			fakePodName:      "podA",
			fakePodNamespace: "default",
			fakeResourceName: "uds/testing",
			udsServerDevType: "uds/testing",
			fakePodDevices:   []string{"devA", "devB"},
			udsServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: constants.Uds.Handshake.RequestConnect + ", podA",
				1: "",
				2: constants.Uds.Handshake.RequestFin,
			},
			expectedResponse: map[int]string{
				0: constants.Uds.Handshake.ResponseHostOk,
				1: constants.Uds.Handshake.ResponseBadRequest,
				2: constants.Uds.Handshake.ResponseFinAck,
			},
		},
		{
			//Validate but then send garbage request
			testName:         "Validate and send no request (2)",
			fakePodName:      "podA",
			fakePodNamespace: "default",
			fakeResourceName: "uds/testing",
			udsServerDevType: "uds/testing",
			fakePodDevices:   []string{"devA", "devB"},
			udsServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: constants.Uds.Handshake.RequestConnect + ", podA",
				1: " ",
				2: constants.Uds.Handshake.RequestFin,
			},
			expectedResponse: map[int]string{
				0: constants.Uds.Handshake.ResponseHostOk,
				1: constants.Uds.Handshake.ResponseBadRequest,
				2: constants.Uds.Handshake.ResponseFinAck,
			},
		},
		{
			//Validate but then send garbage request
			testName:         "Validate and send garbage request (1)",
			fakePodName:      "podA",
			fakePodNamespace: "default",
			fakeResourceName: "uds/testing",
			udsServerDevType: "uds/testing",
			fakePodDevices:   []string{"devA", "devB"},
			udsServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: constants.Uds.Handshake.RequestConnect + ", podA",
				1: ",",
				2: constants.Uds.Handshake.RequestFin,
			},
			expectedResponse: map[int]string{
				0: constants.Uds.Handshake.ResponseHostOk,
				1: constants.Uds.Handshake.ResponseBadRequest,
				2: constants.Uds.Handshake.ResponseFinAck,
			},
		},
		{
			//Validate but then send garbage request
			testName:         "Validate and send garbage request (2)",
			fakePodName:      "podA",
			fakePodNamespace: "default",
			fakeResourceName: "uds/testing",
			udsServerDevType: "uds/testing",
			fakePodDevices:   []string{"devA", "devB"},
			udsServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: constants.Uds.Handshake.RequestConnect + ", podA",
				1: "/",
				2: constants.Uds.Handshake.RequestFin,
			},
			expectedResponse: map[int]string{
				0: constants.Uds.Handshake.ResponseHostOk,
				1: constants.Uds.Handshake.ResponseBadRequest,
				2: constants.Uds.Handshake.ResponseFinAck,
			},
		},
		{
			//Validate but then send garbage request
			testName:         "Validate and send garbage request (3)",
			fakePodName:      "podA",
			fakePodNamespace: "default",
			fakeResourceName: "uds/testing",
			udsServerDevType: "uds/testing",
			fakePodDevices:   []string{"devA", "devB"},
			udsServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: constants.Uds.Handshake.RequestConnect + ", podA",
				1: `\`,
				2: constants.Uds.Handshake.RequestFin,
			},
			expectedResponse: map[int]string{
				0: constants.Uds.Handshake.ResponseHostOk,
				1: constants.Uds.Handshake.ResponseBadRequest,
				2: constants.Uds.Handshake.ResponseFinAck,
			},
		},
		{
			//Validate but then send garbage request
			testName:         "Validate and send garbage request (4)",
			fakePodName:      "podA",
			fakePodNamespace: "default",
			fakeResourceName: "uds/testing",
			udsServerDevType: "uds/testing",
			fakePodDevices:   []string{"devA", "devB"},
			udsServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: constants.Uds.Handshake.RequestConnect + ", podA",
				1: "dkjfhgkjdfs",
				2: constants.Uds.Handshake.RequestFin,
			},
			expectedResponse: map[int]string{
				0: constants.Uds.Handshake.ResponseHostOk,
				1: constants.Uds.Handshake.ResponseBadRequest,
				2: constants.Uds.Handshake.ResponseFinAck,
			},
		},
		{
			//Validate but then send garbage request
			testName:         "Validate and send garbage request (5)",
			fakePodName:      "podA",
			fakePodNamespace: "default",
			fakeResourceName: "uds/testing",
			udsServerDevType: "uds/testing",
			fakePodDevices:   []string{"devA", "devB"},
			udsServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: constants.Uds.Handshake.RequestConnect + ", podA",
				1: "*",
				2: constants.Uds.Handshake.RequestFin,
			},
			expectedResponse: map[int]string{
				0: constants.Uds.Handshake.ResponseHostOk,
				1: constants.Uds.Handshake.ResponseBadRequest,
				2: constants.Uds.Handshake.ResponseFinAck,
			},
		},
		{
			//Validate but then send garbage request
			testName:         "Validate and send garbage request (6)",
			fakePodName:      "podA",
			fakePodNamespace: "default",
			fakeResourceName: "uds/testing",
			udsServerDevType: "uds/testing",
			fakePodDevices:   []string{"devA", "devB"},
			udsServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: constants.Uds.Handshake.RequestConnect + ", podA",
				1: ";",
				2: constants.Uds.Handshake.RequestFin,
			},
			expectedResponse: map[int]string{
				0: constants.Uds.Handshake.ResponseHostOk,
				1: constants.Uds.Handshake.ResponseBadRequest,
				2: constants.Uds.Handshake.ResponseFinAck,
			},
		},
		{
			//Validate but then send garbage request
			testName:         "Validate and send garbage request (7)",
			fakePodName:      "podA",
			fakePodNamespace: "default",
			fakeResourceName: "uds/testing",
			udsServerDevType: "uds/testing",
			fakePodDevices:   []string{"devA", "devB"},
			udsServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: constants.Uds.Handshake.RequestConnect + ", podA",
				1: "\n",
				2: constants.Uds.Handshake.RequestFin,
			},
			expectedResponse: map[int]string{
				0: constants.Uds.Handshake.ResponseHostOk,
				1: constants.Uds.Handshake.ResponseBadRequest,
				2: constants.Uds.Handshake.ResponseFinAck,
			},
		},
		{
			//Validate but then send garbage request
			testName:         "Validate and send garbage request (8)",
			fakePodName:      "podA",
			fakePodNamespace: "default",
			fakeResourceName: "uds/testing",
			udsServerDevType: "uds/testing",
			fakePodDevices:   []string{"devA", "devB"},
			udsServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: constants.Uds.Handshake.RequestConnect + ", podA",
				1: constants.Uds.Handshake.RequestVersion + constants.Uds.Handshake.RequestFin,
				2: constants.Uds.Handshake.RequestFin,
			},
			expectedResponse: map[int]string{
				0: constants.Uds.Handshake.ResponseHostOk,
				1: constants.Uds.Handshake.ResponseBadRequest,
				2: constants.Uds.Handshake.ResponseFinAck,
			},
		},
		{
			//Validate but then send garbage request
			testName:         "Validate and send garbage request (9)",
			fakePodName:      "podA",
			fakePodNamespace: "default",
			fakeResourceName: "uds/testing",
			udsServerDevType: "uds/testing",
			fakePodDevices:   []string{"devA", "devB"},
			udsServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: constants.Uds.Handshake.RequestConnect + ", podA",
				1: constants.Uds.Handshake.RequestVersion + " " + constants.Uds.Handshake.RequestFin,
				2: constants.Uds.Handshake.RequestFin,
			},
			expectedResponse: map[int]string{
				0: constants.Uds.Handshake.ResponseHostOk,
				1: constants.Uds.Handshake.ResponseBadRequest,
				2: constants.Uds.Handshake.ResponseFinAck,
			},
		},
		{
			//Validate but then send garbage request
			testName:         "Validate and send garbage request (10)",
			fakePodName:      "podA",
			fakePodNamespace: "default",
			fakeResourceName: "uds/testing",
			udsServerDevType: "uds/testing",
			fakePodDevices:   []string{"devA", "devB"},
			udsServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: constants.Uds.Handshake.RequestConnect + ", podA",
				1: constants.Uds.Handshake.RequestVersion + "," + constants.Uds.Handshake.RequestFin,
				2: constants.Uds.Handshake.RequestFin,
			},
			expectedResponse: map[int]string{
				0: constants.Uds.Handshake.ResponseHostOk,
				1: constants.Uds.Handshake.ResponseBadRequest,
				2: constants.Uds.Handshake.ResponseFinAck,
			},
		},
		{
			//Validate but then send garbage request
			testName:         "Validate and send garbage request (10)",
			fakePodName:      "podA",
			fakePodNamespace: "default",
			fakeResourceName: "uds/testing",
			udsServerDevType: "uds/testing",
			fakePodDevices:   []string{"devA", "devB"},
			udsServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: constants.Uds.Handshake.RequestConnect + ", podA",
				1: constants.Uds.Handshake.RequestVersion + "\n" + constants.Uds.Handshake.RequestFin,
				2: constants.Uds.Handshake.RequestFin,
			},
			expectedResponse: map[int]string{
				0: constants.Uds.Handshake.ResponseHostOk,
				1: constants.Uds.Handshake.ResponseBadRequest,
				2: constants.Uds.Handshake.ResponseFinAck,
			},
		},
		{
			//Validate but then send garbage request
			testName:         "Validate and send garbage request (11)",
			fakePodName:      "podA",
			fakePodNamespace: "default",
			fakeResourceName: "uds/testing",
			udsServerDevType: "uds/testing",
			fakePodDevices:   []string{"devA", "devB"},
			udsServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: constants.Uds.Handshake.RequestConnect + ", podA",
				1: constants.Uds.Handshake.RequestFd + constants.Uds.Handshake.RequestFin,
				2: constants.Uds.Handshake.RequestFin,
			},
			expectedResponse: map[int]string{
				0: constants.Uds.Handshake.ResponseHostOk,
				1: constants.Uds.Handshake.ResponseBadRequest,
				2: constants.Uds.Handshake.ResponseFinAck,
			},
		},
		{
			//Validate but then send garbage request
			testName:         "Validate and send garbage request (12)",
			fakePodName:      "podA",
			fakePodNamespace: "default",
			fakeResourceName: "uds/testing",
			udsServerDevType: "uds/testing",
			fakePodDevices:   []string{"devA", "devB"},
			udsServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: constants.Uds.Handshake.RequestConnect + ", podA",
				1: constants.Uds.Handshake.RequestFd + " " + constants.Uds.Handshake.RequestFin,
				2: constants.Uds.Handshake.RequestFin,
			},
			expectedResponse: map[int]string{
				0: constants.Uds.Handshake.ResponseHostOk,
				1: constants.Uds.Handshake.ResponseBadRequest,
				2: constants.Uds.Handshake.ResponseFinAck,
			},
		},
		{
			//Validate but then send garbage request
			testName:         "Validate and send garbage request (13)",
			fakePodName:      "podA",
			fakePodNamespace: "default",
			fakeResourceName: "uds/testing",
			udsServerDevType: "uds/testing",
			fakePodDevices:   []string{"devA", "devB"},
			udsServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: constants.Uds.Handshake.RequestConnect + ", podA",
				1: constants.Uds.Handshake.RequestFd + ", " + constants.Uds.Handshake.RequestFin,
				2: constants.Uds.Handshake.RequestFin,
			},
			expectedResponse: map[int]string{
				0: constants.Uds.Handshake.ResponseHostOk,
				1: constants.Uds.Handshake.ResponseFdNak,
				2: constants.Uds.Handshake.ResponseFinAck,
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			// make a new server each time to clear things like device list
			server := &server{
				deviceType: tc.udsServerDevType,
				devices:    make(map[string]int),
				uds:        fakeUDS,
				podRes:     fakeResAPI,
			}

			fakeResAPI.CreateFakePod(tc.fakePodName, tc.fakePodNamespace, tc.fakeResourceName, tc.fakePodDevices)
			fakeUDS.SetRequests(tc.fakeRequests)

			for fd, device := range tc.udsServerDevices {
				server.AddDevice(device, fd)
			}

			server.start()

			responses := fakeUDS.GetResponses()

			for i, response := range responses {
				assert.Equal(t, response, tc.expectedResponse[i])
			}
		})
	}
}

func TestRead(t *testing.T) {
	fakeUDS := uds.NewFakeHandler()

	server := &server{
		devices: make(map[string]int),
		uds:     fakeUDS,
	}

	testCases := []struct {
		testName        string
		fakeRequests    map[int]string
		expectedRequest string
		expectedError   error
	}{
		{
			testName: "Short string",
			fakeRequests: map[int]string{
				0: "hello world",
			},
			expectedRequest: "hello world",
			expectedError:   nil,
		},
		{
			testName: "Long string, single line",
			fakeRequests: map[int]string{
				0: "Lorem ipsum dolor sit amet, consectetur adipiscing elit, " +
					"sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. " +
					"Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. " +
					"Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. " +
					"Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.",
			},
			expectedRequest: "Lorem ipsum dolor sit amet, consectetur adipiscing elit, " +
				"sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. " +
				"Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. " +
				"Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. " +
				"Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.",
			expectedError: nil,
		},
		{
			testName: "Long string, multi line",
			fakeRequests: map[int]string{
				0: "Lorem ipsum dolor sit amet, consectetur adipiscing elit, \n" +
					"sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. \n" +
					"Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. \n" +
					"Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. \n" +
					"Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.",
			},
			expectedRequest: "Lorem ipsum dolor sit amet, consectetur adipiscing elit, \n" +
				"sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. \n" +
				"Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. \n" +
				"Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. \n" +
				"Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.",
			expectedError: nil,
		},
		{
			testName: "Numbers",
			fakeRequests: map[int]string{
				0: "1234567890",
			},
			expectedRequest: "1234567890",
			expectedError:   nil,
		},
		{
			testName: "Symbols",
			fakeRequests: map[int]string{
				0: "!\"£$%^&*()",
			},
			expectedRequest: "!\"£$%^&*()",
			expectedError:   nil,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			fakeUDS.SetRequests(tc.fakeRequests)
			request, _, err := server.read()
			assert.Equal(t, err, tc.expectedError)
			assert.Equal(t, request, tc.expectedRequest)
		})
	}
}
