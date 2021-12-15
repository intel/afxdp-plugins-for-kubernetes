/*
 * Copyright(c) 2021 Intel Corporation.
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

import (
	"github.com/intel/cndp_device_plugin/internal/resourcesapi"
	"github.com/intel/cndp_device_plugin/internal/uds"
	"gotest.tools/assert"
	"testing"
)

func TestCreateNewServer(t *testing.T) {
	//cndp := NewServerFactory() //TODO

	testCases := []struct {
		testName       string
		deviceType     string
		expectedPath   string
		expectedServer *server
	}{
		{
			testName:   "Create UDS Server",
			deviceType: "cndp/device",
			expectedServer: &server{
				deviceType: "cndp/device",
				devices:    make(map[string]int),
				uds:        uds.NewFakeHandler(),
				podRes:     resourcesapi.NewFakeHandler(),
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			//TODO compare individual elements
			//receivedServer, _ := cndp.CreateServer(tc.deviceType)
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
		testName          string
		fakePodName       string
		fakePodNamespace  string
		fakeResourceName  string
		cndpServerDevType string
		fakePodDevices    []string
		cndpServerDevices []string
		fakeRequests      map[int]string
		expectedResponse  map[int]string
	}{
		/***************************************
		Positive Tests - validate and disconnect
		***************************************/
		{
			//Try connect good podA which has 1 good device - devA
			testName:          "Connect successfully and disconnect, 1 device",
			fakePodName:       "podA",
			fakePodNamespace:  "default",
			fakeResourceName:  "cndp/testing",
			cndpServerDevType: "cndp/testing",
			fakePodDevices:    []string{"devA"},
			cndpServerDevices: []string{"devA"},
			fakeRequests: map[int]string{
				0: requestConnect + ", podA",
				1: requestFin,
			},
			expectedResponse: map[int]string{
				0: responseHostOk,
				1: responseFinAck,
			},
		},
		{
			//Try connect good podA which has 2 good devices - devA and devB
			testName:          "Connect successfully and disconnect, 2 device",
			fakePodName:       "podA",
			fakePodNamespace:  "default",
			fakeResourceName:  "cndp/testing",
			cndpServerDevType: "cndp/testing",
			fakePodDevices:    []string{"devA", "devB"},
			cndpServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: requestConnect + ", podA",
				1: requestFin,
			},
			expectedResponse: map[int]string{
				0: responseHostOk,
				1: responseFinAck,
			},
		},
		{
			//Try connect good podA which has 10 good devices - devA to devJ
			testName:          "Connect successfully and disconnect, 2 device",
			fakePodName:       "podA",
			fakePodNamespace:  "default",
			fakeResourceName:  "cndp/testing",
			cndpServerDevType: "cndp/testing",
			fakePodDevices:    []string{"devA", "devB", "devC", "devD", "devE", "devF", "devG", "devH", "devI", "devJ"},
			cndpServerDevices: []string{"devA", "devB", "devC", "devD", "devE", "devF", "devG", "devH", "devI", "devJ"},
			fakeRequests: map[int]string{
				0: requestConnect + ", podA",
				1: requestFin,
			},
			expectedResponse: map[int]string{
				0: responseHostOk,
				1: responseFinAck,
			},
		},
		{
			//Try connect good podA, in non-default namespace, which has 2 good devices - devA and devB
			testName:          "Connect a pod in non-default NS",
			fakePodName:       "podA",
			fakePodNamespace:  "someOtherNamespace",
			fakeResourceName:  "cndp/testing",
			cndpServerDevType: "cndp/testing",
			fakePodDevices:    []string{"devA", "devB"},
			cndpServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: requestConnect + ", podA",
				1: requestFin,
			},
			expectedResponse: map[int]string{
				0: responseHostOk,
				1: responseFinAck,
			},
		},
		/*********************************************************
		Positive Tests - validate, request good FDs and disconnect
		*********************************************************/
		{
			//Connect podA, request FD for it's single device - devA
			testName:          "Connect and request good FD, 1 device",
			fakePodName:       "podA",
			fakePodNamespace:  "default",
			fakeResourceName:  "cndp/testing",
			cndpServerDevType: "cndp/testing",
			fakePodDevices:    []string{"devA"},
			cndpServerDevices: []string{"devA"},
			fakeRequests: map[int]string{
				0: requestConnect + ", podA",
				1: requestFd + ", devA",
				2: requestFin,
			},
			expectedResponse: map[int]string{
				0: responseHostOk,
				1: responseFdAck,
				2: responseFinAck,
			},
		},
		{
			//Connect podA, request FDs for it's 3 devices - devA devB devC
			testName:          "Connect and request good FDs, 3 devices",
			fakePodName:       "podA",
			fakePodNamespace:  "default",
			fakeResourceName:  "cndp/testing",
			cndpServerDevType: "cndp/testing",
			fakePodDevices:    []string{"devA", "devB", "devC"},
			cndpServerDevices: []string{"devA", "devB", "devC"},
			fakeRequests: map[int]string{
				0: requestConnect + ", podA",
				1: requestFd + ", devA",
				2: requestFd + ", devB",
				3: requestFd + ", devC",
				4: requestFin,
			},
			expectedResponse: map[int]string{
				0: responseHostOk,
				1: responseFdAck,
				2: responseFdAck,
				3: responseFdAck,
				4: responseFinAck,
			},
		},
		{
			//Connect podA, request version and disconnect
			testName:          "Connect and request version",
			fakePodName:       "podA",
			fakePodNamespace:  "default",
			fakeResourceName:  "cndp/testing",
			cndpServerDevType: "cndp/testing",
			fakePodDevices:    []string{"devA", "devB"},
			cndpServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: requestConnect + ", podA",
				1: requestVersion,
				2: requestFin,
			},
			expectedResponse: map[int]string{
				0: responseHostOk,
				1: handshakeVersion,
				2: responseFinAck,
			},
		},
		{
			//Connect and test full handshake
			testName:          "Full handshake",
			fakePodName:       "podA",
			fakePodNamespace:  "default",
			fakeResourceName:  "cndp/testing",
			cndpServerDevType: "cndp/testing",
			fakePodDevices:    []string{"devA", "devB", "devC"},
			cndpServerDevices: []string{"devA", "devB", "devC"},
			fakeRequests: map[int]string{
				0: requestConnect + ", podA",
				1: requestVersion,
				2: requestFd + ", devA",
				3: requestFd + ", devB",
				4: requestFd + ", devC",
				5: requestFin,
			},
			expectedResponse: map[int]string{
				0: responseHostOk,
				1: handshakeVersion,
				2: responseFdAck,
				3: responseFdAck,
				4: responseFdAck,
				5: responseFinAck,
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
			testName:          "No hostname (1)",
			fakePodName:       "podA",
			fakePodNamespace:  "default",
			fakeResourceName:  "cndp/testing",
			cndpServerDevType: "cndp/testing",
			fakePodDevices:    []string{"devA", "devB"},
			cndpServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: requestConnect,
				1: requestFin,
			},
			expectedResponse: map[int]string{
				0: responseHostNak,
				1: "should not get " + responseFinAck + " as should not have connected",
			},
		},
		{
			//Try connect without passing any pod, include the comma
			testName:          "No hostname (2)",
			fakePodName:       "podA",
			fakePodNamespace:  "default",
			fakeResourceName:  "cndp/testing",
			cndpServerDevType: "cndp/testing",
			fakePodDevices:    []string{"devA", "devB"},
			cndpServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: requestConnect + ",",
				1: requestFin,
			},
			expectedResponse: map[int]string{
				0: responseHostNak,
				1: "should not get " + responseFinAck + " as should not have connected",
			},
		},
		{
			//Try connect 2 hostnames
			testName:          "Two hostnames",
			fakePodName:       "podA",
			fakePodNamespace:  "default",
			fakeResourceName:  "cndp/testing",
			cndpServerDevType: "cndp/testing",
			fakePodDevices:    []string{"devA", "devB"},
			cndpServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: requestConnect + ", podA, podB",
				1: requestFin,
			},
			expectedResponse: map[int]string{
				0: responseHostNak,
				1: "should not get " + responseFinAck + " as should not have connected",
			},
		},
		{
			//Try connect 5 hostnames
			testName:          "Many hostnames",
			fakePodName:       "podA",
			fakePodNamespace:  "default",
			fakeResourceName:  "cndp/testing",
			cndpServerDevType: "cndp/testing",
			fakePodDevices:    []string{"devA", "devB"},
			cndpServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: requestConnect + ", podA, podB, podC, podD, podE",
				1: requestFin,
			},
			expectedResponse: map[int]string{
				0: responseHostNak,
				1: "should not get " + responseFinAck + " as should not have connected",
			},
		},
		{
			//Put the podname before connect request
			testName:          "Hostname before request",
			fakePodName:       "podA",
			fakePodNamespace:  "default",
			fakeResourceName:  "cndp/testing",
			cndpServerDevType: "cndp/testing",
			fakePodDevices:    []string{"devA", "devB"},
			cndpServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: "podA, " + requestConnect,
				1: requestFin,
			},
			expectedResponse: map[int]string{
				0: responseHostNak,
				1: "should not get " + responseFinAck + " as should not have connected",
			},
		},
		{
			//Put the podname before and after connect request
			testName:          "Hostname before and after request",
			fakePodName:       "podA",
			fakePodNamespace:  "default",
			fakeResourceName:  "cndp/testing",
			cndpServerDevType: "cndp/testing",
			fakePodDevices:    []string{"devA", "devB"},
			cndpServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: "podA, " + requestConnect + ", podA",
				1: requestFin,
			},
			expectedResponse: map[int]string{
				0: responseHostNak,
				1: "should not get " + responseFinAck + " as should not have connected",
			},
		},
		{
			//Put a null byte before an otherwise good connect request
			testName:          "Null byte before good connect request (1)",
			fakePodName:       "podA",
			fakePodNamespace:  "default",
			fakeResourceName:  "cndp/testing",
			cndpServerDevType: "cndp/testing",
			fakePodDevices:    []string{"devA", "devB"},
			cndpServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: `\0` + requestConnect + ", podA",
				1: requestFin,
			},
			expectedResponse: map[int]string{
				0: responseHostNak,
				1: "should not get " + responseFinAck + " as should not have connected",
			},
		},
		{
			//Put a null byte an otherwise good connect request
			testName:          "Null byte before good connect request (2)",
			fakePodName:       "podA",
			fakePodNamespace:  "default",
			fakeResourceName:  "cndp/testing",
			cndpServerDevType: "cndp/testing",
			fakePodDevices:    []string{"devA", "devB"},
			cndpServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: "\\0" + requestConnect + ", podA",
				1: requestFin,
			},
			expectedResponse: map[int]string{
				0: responseHostNak,
				1: "should not get " + responseFinAck + " as should not have connected",
			},
		},
		{
			//Put garbage before an otherwise good connect request
			testName:          "Garbage before good connect request (1)",
			fakePodName:       "podA",
			fakePodNamespace:  "default",
			fakeResourceName:  "cndp/testing",
			cndpServerDevType: "cndp/testing",
			fakePodDevices:    []string{"devA", "devB"},
			cndpServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: "dkjfhgkjdfs" + requestConnect + ", podA",
				1: requestFin,
			},
			expectedResponse: map[int]string{
				0: responseHostNak,
				1: "should not get " + responseFinAck + " as should not have connected",
			},
		},
		{
			//Put garbage before an otherwise good connect request
			testName:          "Garbage before good connect request (2)",
			fakePodName:       "podA",
			fakePodNamespace:  "default",
			fakeResourceName:  "cndp/testing",
			cndpServerDevType: "cndp/testing",
			fakePodDevices:    []string{"devA", "devB"},
			cndpServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: " " + requestConnect + ", podA",
				1: requestFin,
			},
			expectedResponse: map[int]string{
				0: responseHostNak,
				1: "should not get " + responseFinAck + " as should not have connected",
			},
		},
		{
			//Put garbage before an otherwise good connect request
			testName:          "Garbage before good connect request (3)",
			fakePodName:       "podA",
			fakePodNamespace:  "default",
			fakeResourceName:  "cndp/testing",
			cndpServerDevType: "cndp/testing",
			fakePodDevices:    []string{"devA", "devB"},
			cndpServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: `\` + requestConnect + ", podA",
				1: requestFin,
			},
			expectedResponse: map[int]string{
				0: responseHostNak,
				1: "should not get " + responseFinAck + " as should not have connected",
			},
		},
		{
			//Put garbage before an otherwise good connect request
			testName:          "Garbage before good connect request (4)",
			fakePodName:       "podA",
			fakePodNamespace:  "default",
			fakeResourceName:  "cndp/testing",
			cndpServerDevType: "cndp/testing",
			fakePodDevices:    []string{"devA", "devB"},
			cndpServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: "\\\\" + requestConnect + ", podA",
				1: requestFin,
			},
			expectedResponse: map[int]string{
				0: responseHostNak,
				1: "should not get " + responseFinAck + " as should not have connected",
			},
		},
		{
			//Put garbage before an otherwise good connect request
			testName:          "Garbage before good connect request (5)",
			fakePodName:       "podA",
			fakePodNamespace:  "default",
			fakeResourceName:  "cndp/testing",
			cndpServerDevType: "cndp/testing",
			fakePodDevices:    []string{"devA", "devB"},
			cndpServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: "," + requestConnect + ", podA",
				1: requestFin,
			},
			expectedResponse: map[int]string{
				0: responseHostNak,
				1: "should not get " + responseFinAck + " as should not have connected",
			},
		},
		{
			//Put a null byte after an otherwise good connect request
			testName:          "Null byte after good connect request (1)",
			fakePodName:       "podA",
			fakePodNamespace:  "default",
			fakeResourceName:  "cndp/testing",
			cndpServerDevType: "cndp/testing",
			fakePodDevices:    []string{"devA", "devB"},
			cndpServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: requestConnect + ", podA" + `\0`,
				1: requestFin,
			},
			expectedResponse: map[int]string{
				0: responseHostNak,
				1: "should not get " + responseFinAck + " as should not have connected",
			},
		},
		{
			//Put a null byte after an otherwise good connect request
			testName:          "Null byte after good connect request (2)",
			fakePodName:       "podA",
			fakePodNamespace:  "default",
			fakeResourceName:  "cndp/testing",
			cndpServerDevType: "cndp/testing",
			fakePodDevices:    []string{"devA", "devB"},
			cndpServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: requestConnect + ", podA" + "\\0",
				1: requestFin,
			},
			expectedResponse: map[int]string{
				0: responseHostNak,
				1: "should not get " + responseFinAck + " as should not have connected",
			},
		},
		{
			//Put garbage after an otherwise good connect request
			testName:          "Garbage after good connect request (1)",
			fakePodName:       "podA",
			fakePodNamespace:  "default",
			fakeResourceName:  "cndp/testing",
			cndpServerDevType: "cndp/testing",
			fakePodDevices:    []string{"devA", "devB"},
			cndpServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: requestConnect + ", podAdkjfhgkjdfs",
				1: requestFin,
			},
			expectedResponse: map[int]string{
				0: responseHostNak,
				1: "should not get " + responseFinAck + " as should not have connected",
			},
		},
		{
			//Put garbage after an otherwise good connect request
			testName:          "Garbage after good connect request (2)",
			fakePodName:       "podA",
			fakePodNamespace:  "default",
			fakeResourceName:  "cndp/testing",
			cndpServerDevType: "cndp/testing",
			fakePodDevices:    []string{"devA", "devB"},
			cndpServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: requestConnect + ", podA,dkjfhgkjdfs",
				1: requestFin,
			},
			expectedResponse: map[int]string{
				0: responseHostNak,
				1: "should not get " + responseFinAck + " as should not have connected",
			},
		},
		{
			//Put garbage after an otherwise good connect request
			testName:          "Garbage after good connect request (3)",
			fakePodName:       "podA",
			fakePodNamespace:  "default",
			fakeResourceName:  "cndp/testing",
			cndpServerDevType: "cndp/testing",
			fakePodDevices:    []string{"devA", "devB"},
			cndpServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: requestConnect + ", podA" + `\`,
				1: requestFin,
			},
			expectedResponse: map[int]string{
				0: responseHostNak,
				1: "should not get " + responseFinAck + " as should not have connected",
			},
		},
		{
			//Put garbage after an otherwise good connect request
			testName:          "Garbage after good connect request (4)",
			fakePodName:       "podA",
			fakePodNamespace:  "default",
			fakeResourceName:  "cndp/testing",
			cndpServerDevType: "cndp/testing",
			fakePodDevices:    []string{"devA", "devB"},
			cndpServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: requestConnect + ", podA\\\\",
				1: requestFin,
			},
			expectedResponse: map[int]string{
				0: responseHostNak,
				1: "should not get " + responseFinAck + " as should not have connected",
			},
		},
		{
			//Put garbage after an otherwise good connect request
			testName:          "Garbage after good connect request (5)",
			fakePodName:       "podA",
			fakePodNamespace:  "default",
			fakeResourceName:  "cndp/testing",
			cndpServerDevType: "cndp/testing",
			fakePodDevices:    []string{"devA", "devB"},
			cndpServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: requestConnect + ", podA,",
				1: requestFin,
			},
			expectedResponse: map[int]string{
				0: responseHostNak,
				1: "should not get " + responseFinAck + " as should not have connected",
			},
		},
		{
			//Put null byte within otherwise good connect request
			testName:          "Null byte within good connect request (1)",
			fakePodName:       "podA",
			fakePodNamespace:  "default",
			fakeResourceName:  "cndp/testing",
			cndpServerDevType: "cndp/testing",
			fakePodDevices:    []string{"devA", "devB"},
			cndpServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: requestConnect + `\0` + ", podA",
				1: requestFin,
			},
			expectedResponse: map[int]string{
				0: responseHostNak,
				1: "should not get " + responseFinAck + " as should not have connected",
			},
		},
		{
			//Put null byte within otherwise good connect request
			testName:          "Null byte within good connect request (2)",
			fakePodName:       "podA",
			fakePodNamespace:  "default",
			fakeResourceName:  "cndp/testing",
			cndpServerDevType: "cndp/testing",
			fakePodDevices:    []string{"devA", "devB"},
			cndpServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: requestConnect + "\\0" + ", podA",
				1: requestFin,
			},
			expectedResponse: map[int]string{
				0: responseHostNak,
				1: "should not get " + responseFinAck + " as should not have connected",
			},
		},
		{
			//Null byte before connect
			testName:          "Send a null byte (1)",
			fakePodName:       "podA",
			fakePodNamespace:  "default",
			fakeResourceName:  "cndp/testing",
			cndpServerDevType: "cndp/testing",
			fakePodDevices:    []string{"devA", "devB"},
			cndpServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: `\0`,
				1: requestFin,
			},
			expectedResponse: map[int]string{
				0: responseHostNak,
				1: "should not get " + responseFinAck + " as should not have connected",
			},
		},
		{
			//Null byte before connect
			testName:          "Send a null byte (2)",
			fakePodName:       "podA",
			fakePodNamespace:  "default",
			fakeResourceName:  "cndp/testing",
			cndpServerDevType: "cndp/testing",
			fakePodDevices:    []string{"devA", "devB"},
			cndpServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: "\\0",
				1: requestFin,
			},
			expectedResponse: map[int]string{
				0: responseHostNak,
				1: "should not get " + responseFinAck + " as should not have connected",
			},
		},
		{
			//Garbage request before connect
			testName:          "Garbage request, before connect (1)",
			fakePodName:       "podA",
			fakePodNamespace:  "default",
			fakeResourceName:  "cndp/testing",
			cndpServerDevType: "cndp/testing",
			fakePodDevices:    []string{"devA", "devB"},
			cndpServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: "asdfxc",
				1: requestFin,
			},
			expectedResponse: map[int]string{
				0: responseBadRequest,
				1: "should not get " + responseFinAck + " as should not have connected",
			},
		},
		{
			//Garbage request before connect
			testName:          "Garbage request, before connect (2)",
			fakePodName:       "podA",
			fakePodNamespace:  "default",
			fakeResourceName:  "cndp/testing",
			cndpServerDevType: "cndp/testing",
			fakePodDevices:    []string{"devA", "devB"},
			cndpServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: "asdfxc,",
				1: requestFin,
			},
			expectedResponse: map[int]string{
				0: responseBadRequest,
				1: "should not get " + responseFinAck + " as should not have connected",
			},
		},
		{
			//Garbage request before connect
			testName:          "Garbage request, before connect (3)",
			fakePodName:       "podA",
			fakePodNamespace:  "default",
			fakeResourceName:  "cndp/testing",
			cndpServerDevType: "cndp/testing",
			fakePodDevices:    []string{"devA", "devB"},
			cndpServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: "asdfxc, podA,",
				1: requestFin,
			},
			expectedResponse: map[int]string{
				0: responseBadRequest,
				1: "should not get " + responseFinAck + " as should not have connected",
			},
		},
		{
			//Garbage request before connect
			testName:          "Garbage request, before connect (4)",
			fakePodName:       "podA",
			fakePodNamespace:  "default",
			fakeResourceName:  "cndp/testing",
			cndpServerDevType: "cndp/testing",
			fakePodDevices:    []string{"devA", "devB"},
			cndpServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: `\`,
				1: requestFin,
			},
			expectedResponse: map[int]string{
				0: responseBadRequest,
				1: "should not get " + responseFinAck + " as should not have connected",
			},
		},
		{
			//Garbage request before connect
			testName:          "Garbage request, before connect (5)",
			fakePodName:       "podA",
			fakePodNamespace:  "default",
			fakeResourceName:  "cndp/testing",
			cndpServerDevType: "cndp/testing",
			fakePodDevices:    []string{"devA", "devB"},
			cndpServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: "\\\\",
				1: requestFin,
			},
			expectedResponse: map[int]string{
				0: responseBadRequest,
				1: "should not get " + responseFinAck + " as should not have connected",
			},
		},
		{
			//Garbage request before connect
			testName:          "Garbage request, before connect (6)",
			fakePodName:       "podA",
			fakePodNamespace:  "default",
			fakeResourceName:  "cndp/testing",
			cndpServerDevType: "cndp/testing",
			fakePodDevices:    []string{"devA", "devB"},
			cndpServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: "*",
				1: requestFin,
			},
			expectedResponse: map[int]string{
				0: responseBadRequest,
				1: "should not get " + responseFinAck + " as should not have connected",
			},
		},
		{
			//Send varient of the accepted /connect request
			testName:          "Bad connect request (1)",
			fakePodName:       "podA",
			fakePodNamespace:  "default",
			fakeResourceName:  "cndp/testing",
			cndpServerDevType: "cndp/testing",
			fakePodDevices:    []string{"devA", "devB"},
			cndpServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: "connect",
				1: requestFin,
			},
			expectedResponse: map[int]string{
				0: responseBadRequest,
				1: "should not get " + responseFinAck + " as should not have connected",
			},
		},
		{
			//Send varient of the accepted /connect request
			testName:          "Bad connect request (2)",
			fakePodName:       "podA",
			fakePodNamespace:  "default",
			fakeResourceName:  "cndp/testing",
			cndpServerDevType: "cndp/testing",
			fakePodDevices:    []string{"devA", "devB"},
			cndpServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: "connect/",
				1: requestFin,
			},
			expectedResponse: map[int]string{
				0: responseBadRequest,
				1: "should not get " + responseFinAck + " as should not have connected",
			},
		},
		{
			//Send varient of the accepted /connect request
			testName:          "Bad connect request (3)",
			fakePodName:       "podA",
			fakePodNamespace:  "default",
			fakeResourceName:  "cndp/testing",
			cndpServerDevType: "cndp/testing",
			fakePodDevices:    []string{"devA", "devB"},
			cndpServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: "/Connect",
				1: requestFin,
			},
			expectedResponse: map[int]string{
				0: responseBadRequest,
				1: "should not get " + responseFinAck + " as should not have connected",
			},
		},
		{
			//Send varient of the accepted /connect request
			testName:          "Bad connect request (4)",
			fakePodName:       "podA",
			fakePodNamespace:  "default",
			fakeResourceName:  "cndp/testing",
			cndpServerDevType: "cndp/testing",
			fakePodDevices:    []string{"devA", "devB"},
			cndpServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: "/connect*",
				1: requestFin,
			},
			expectedResponse: map[int]string{
				0: responseHostNak,
				1: "should not get " + responseFinAck + " as should not have connected",
			},
		},
		{
			//Send varient of the accepted /connect request
			testName:          "Bad connect request (5)",
			fakePodName:       "podA",
			fakePodNamespace:  "default",
			fakeResourceName:  "cndp/testing",
			cndpServerDevType: "cndp/testing",
			fakePodDevices:    []string{"devA", "devB"},
			cndpServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: `\/connect*`,
				1: requestFin,
			},
			expectedResponse: map[int]string{
				0: responseHostNak,
				1: "should not get " + responseFinAck + " as should not have connected",
			},
		},
		{
			//Send varient of the accepted /connect request
			testName:          "Bad connect request (6)",
			fakePodName:       "podA",
			fakePodNamespace:  "default",
			fakeResourceName:  "cndp/testing",
			cndpServerDevType: "cndp/testing",
			fakePodDevices:    []string{"devA", "devB"},
			cndpServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: "*/connect*",
				1: requestFin,
			},
			expectedResponse: map[int]string{
				0: responseHostNak,
				1: "should not get " + responseFinAck + " as should not have connected",
			},
		},
		{
			//Try connect podX but we only know podA. podX has 2 good devices - devA and devB
			testName:          "Bad hostname, Good Devices",
			fakePodName:       "podA",
			fakePodNamespace:  "default",
			fakeResourceName:  "cndp/testing",
			cndpServerDevType: "cndp/testing",
			fakePodDevices:    []string{"devA", "devB"},
			cndpServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: requestConnect + ", podX",
				1: requestFin,
			},
			expectedResponse: map[int]string{
				0: responseHostNak,
				1: "should not get " + responseFinAck + " as should not have connected",
			},
		},
		{
			//Try connect good podA, but we have 2 bad devices - devX and devY
			testName:          "Good hostname, Bad Devices",
			fakePodName:       "podA",
			fakePodNamespace:  "default",
			fakeResourceName:  "cndp/testing",
			cndpServerDevType: "cndp/testing",
			fakePodDevices:    []string{"devA", "devB"},
			cndpServerDevices: []string{"devX", "devY"},
			fakeRequests: map[int]string{
				0: requestConnect + ", podA",
				1: requestFin,
			},
			expectedResponse: map[int]string{
				0: responseHostNak,
				1: "should not get " + responseFinAck + " as should not have connected",
			},
		},
		{
			//Try connect good podA, but 1 of 4 devices is bad - devX
			testName:          "Good hostname, Single bad device (1)",
			fakePodName:       "podA",
			fakePodNamespace:  "default",
			fakeResourceName:  "cndp/testing",
			cndpServerDevType: "cndp/testing",
			fakePodDevices:    []string{"devA", "devB", "devC", "devD"},
			cndpServerDevices: []string{"devA", "devB", "devC", "devX"},
			fakeRequests: map[int]string{
				0: requestConnect + ", podA",
				1: requestFin,
			},
			expectedResponse: map[int]string{
				0: responseHostNak,
				1: "should not get " + responseFinAck + " as should not have connected",
			},
		},
		{
			//Try connect good podA, but 1 of 4 devices is bad - devX
			testName:          "Good hostname, Single bad device (2)",
			fakePodName:       "podA",
			fakePodNamespace:  "default",
			fakeResourceName:  "cndp/testing",
			cndpServerDevType: "cndp/testing",
			fakePodDevices:    []string{"devA", "devB", "devC", "devD"},
			cndpServerDevices: []string{"devX", "devB", "devC", "devD"},
			fakeRequests: map[int]string{
				0: requestConnect + ", podA",
				1: requestFin,
			},
			expectedResponse: map[int]string{
				0: responseHostNak,
				1: "should not get " + responseFinAck + " as should not have connected",
			},
		},
		{
			//Try connect good podA, but 1 of 4 devices is bad - devX
			testName:          "Good hostname, Single bad device (3)",
			fakePodName:       "podA",
			fakePodNamespace:  "default",
			fakeResourceName:  "cndp/testing",
			cndpServerDevType: "cndp/testing",
			fakePodDevices:    []string{"devA", "devB", "devC", "devD"},
			cndpServerDevices: []string{"devA", "devB", "devX", "devD"},
			fakeRequests: map[int]string{
				0: requestConnect + ", podA",
				1: requestFin,
			},
			expectedResponse: map[int]string{
				0: responseHostNak,
				1: "should not get " + responseFinAck + " as should not have connected",
			},
		},
		{
			//Try connect good podA, all 4 devices are good, but we have one extra - devX
			testName:          "Good hostname, good devices, 1 extra device (1)",
			fakePodName:       "podA",
			fakePodNamespace:  "default",
			fakeResourceName:  "cndp/testing",
			cndpServerDevType: "cndp/testing",
			fakePodDevices:    []string{"devA", "devB", "devC", "devD"},
			cndpServerDevices: []string{"devA", "devB", "devC", "devD", "devX"},
			fakeRequests: map[int]string{
				0: requestConnect + ", podA",
				1: requestFin,
			},
			expectedResponse: map[int]string{
				0: responseHostNak,
				1: "should not get " + responseFinAck + " as should not have connected",
			},
		},
		{
			//Try connect good podA, all devices are good, but we are missing one - devD
			testName:          "Good hostname, good devices, 1 missing device",
			fakePodName:       "podA",
			fakePodNamespace:  "default",
			fakeResourceName:  "cndp/testing",
			cndpServerDevType: "cndp/testing",
			fakePodDevices:    []string{"devA", "devB", "devC", "devD"},
			cndpServerDevices: []string{"devA", "devB", "devC"},
			fakeRequests: map[int]string{
				0: requestConnect + ", podA",
				1: requestFin,
			},
			expectedResponse: map[int]string{
				0: responseHostNak,
				1: "should not get " + responseFinAck + " as should not have connected",
			},
		},
		{
			//Try connect good podA, both devices are good, but they are of the wrong type - cndp/badType
			testName:          "Good hostname, good device, bad device type",
			fakePodName:       "podA",
			fakePodNamespace:  "default",
			fakeResourceName:  "cndp/testing",
			cndpServerDevType: "cndp/badType",
			fakePodDevices:    []string{"devA", "devB"},
			cndpServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: requestConnect + ", podA",
				1: requestFin,
			},
			expectedResponse: map[int]string{
				0: responseHostNak,
				1: "should not get " + responseFinAck + " as should not have connected",
			},
		},
		/****************************************************************
		Negative Tests - validate, but request FDs for devs we don't have
		****************************************************************/
		{
			//Connect podA, request FD device it does not have - devX
			testName:          "Connect and request bad FD, 1 device",
			fakePodName:       "podA",
			fakePodNamespace:  "default",
			fakeResourceName:  "cndp/testing",
			cndpServerDevType: "cndp/testing",
			fakePodDevices:    []string{"devA"},
			cndpServerDevices: []string{"devA"},
			fakeRequests: map[int]string{
				0: requestConnect + ", podA",
				1: requestFd + ", devX",
				2: requestFin,
			},
			expectedResponse: map[int]string{
				0: responseHostOk,
				1: responseFdNak,
				2: responseFinAck,
			},
		},
		{
			//Connect podA, request one bad FD out of 3 - devX
			testName:          "Connect and request 2 good and 1 bad FD",
			fakePodName:       "podA",
			fakePodNamespace:  "default",
			fakeResourceName:  "cndp/testing",
			cndpServerDevType: "cndp/testing",
			fakePodDevices:    []string{"devA", "devB", "devC"},
			cndpServerDevices: []string{"devA", "devB", "devC"},
			fakeRequests: map[int]string{
				0: requestConnect + ", podA",
				1: requestFd + ", devA",
				2: requestFd + ", devX",
				3: requestFd + ", devC",
				4: requestFin,
			},
			expectedResponse: map[int]string{
				0: responseHostOk,
				1: responseFdAck,
				2: responseFdNak,
				3: responseFdAck,
				4: responseFinAck,
			},
		},
		/******************************************************
		Negative Tests - validate, but send corrupt FD requests
		******************************************************/
		{
			//Send FD request without any device
			testName:          "Request no FD (1)",
			fakePodName:       "podA",
			fakePodNamespace:  "default",
			fakeResourceName:  "cndp/testing",
			cndpServerDevType: "cndp/testing",
			fakePodDevices:    []string{"devA"},
			cndpServerDevices: []string{"devA"},
			fakeRequests: map[int]string{
				0: requestConnect + ", podA",
				1: requestFd,
				2: requestFin,
			},
			expectedResponse: map[int]string{
				0: responseHostOk,
				1: responseBadRequest,
				2: responseFinAck,
			},
		},
		{
			//Send FD request without any device, include comma
			testName:          "Request no FD (2)",
			fakePodName:       "podA",
			fakePodNamespace:  "default",
			fakeResourceName:  "cndp/testing",
			cndpServerDevType: "cndp/testing",
			fakePodDevices:    []string{"devA"},
			cndpServerDevices: []string{"devA"},
			fakeRequests: map[int]string{
				0: requestConnect + ", podA",
				1: requestFd + ",",
				2: requestFin,
			},
			expectedResponse: map[int]string{
				0: responseHostOk,
				1: responseFdNak,
				2: responseFinAck,
			},
		},
		{
			//Send FD request for 2 devs
			testName:          "Request 2 FD",
			fakePodName:       "podA",
			fakePodNamespace:  "default",
			fakeResourceName:  "cndp/testing",
			cndpServerDevType: "cndp/testing",
			fakePodDevices:    []string{"devA", "devB"},
			cndpServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: requestConnect + ", podA",
				1: requestFd + ", devA, devB",
				2: requestFin,
			},
			expectedResponse: map[int]string{
				0: responseHostOk,
				1: responseBadRequest,
				2: responseFinAck,
			},
		},
		{
			//Send dev before FD request
			testName:          "Device before FD request",
			fakePodName:       "podA",
			fakePodNamespace:  "default",
			fakeResourceName:  "cndp/testing",
			cndpServerDevType: "cndp/testing",
			fakePodDevices:    []string{"devA", "devB"},
			cndpServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: requestConnect + ", podA",
				1: "devA, " + requestFd,
				2: requestFin,
			},
			expectedResponse: map[int]string{
				0: responseHostOk,
				1: responseBadRequest,
				2: responseFinAck,
			},
		},
		{
			//Send dev before and after FD request
			testName:          "Device before and after FD request",
			fakePodName:       "podA",
			fakePodNamespace:  "default",
			fakeResourceName:  "cndp/testing",
			cndpServerDevType: "cndp/testing",
			fakePodDevices:    []string{"devA", "devB"},
			cndpServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: requestConnect + ", podA",
				1: "devA, " + requestFd + ", devA",
				2: requestFin,
			},
			expectedResponse: map[int]string{
				0: responseHostOk,
				1: responseBadRequest,
				2: responseFinAck,
			},
		},
		{
			//Put null byte before an otherwise good FD request
			testName:          "Null byte before good FD request (1)",
			fakePodName:       "podA",
			fakePodNamespace:  "default",
			fakeResourceName:  "cndp/testing",
			cndpServerDevType: "cndp/testing",
			fakePodDevices:    []string{"devA", "devB"},
			cndpServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: requestConnect + ", podA",
				1: `\0` + requestFd + ", devA",
				2: requestFin,
			},
			expectedResponse: map[int]string{
				0: responseHostOk,
				1: responseBadRequest,
				2: responseFinAck,
			},
		},
		{
			//Put null byte before an otherwise good FD request
			testName:          "Null byte before good FD request (2)",
			fakePodName:       "podA",
			fakePodNamespace:  "default",
			fakeResourceName:  "cndp/testing",
			cndpServerDevType: "cndp/testing",
			fakePodDevices:    []string{"devA", "devB"},
			cndpServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: requestConnect + ", podA",
				1: "\\0" + requestFd + ", devA",
				2: requestFin,
			},
			expectedResponse: map[int]string{
				0: responseHostOk,
				1: responseBadRequest,
				2: responseFinAck,
			},
		},
		{
			//Put garbage before an otherwise good FD request
			testName:          "Garbage before good FD request (1)",
			fakePodName:       "podA",
			fakePodNamespace:  "default",
			fakeResourceName:  "cndp/testing",
			cndpServerDevType: "cndp/testing",
			fakePodDevices:    []string{"devA", "devB"},
			cndpServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: requestConnect + ", podA",
				1: "dkjfhgkjdfs" + requestFd + ", devA",
				2: requestFin,
			},
			expectedResponse: map[int]string{
				0: responseHostOk,
				1: responseBadRequest,
				2: responseFinAck,
			},
		},
		{
			//Put garbage before an otherwise good FD request
			testName:          "Garbage before good FD request (2)",
			fakePodName:       "podA",
			fakePodNamespace:  "default",
			fakeResourceName:  "cndp/testing",
			cndpServerDevType: "cndp/testing",
			fakePodDevices:    []string{"devA", "devB"},
			cndpServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: requestConnect + ", podA",
				1: " " + requestFd + ", devA",
				2: requestFin,
			},
			expectedResponse: map[int]string{
				0: responseHostOk,
				1: responseBadRequest,
				2: responseFinAck,
			},
		},
		{
			//Put garbage before an otherwise good FD request
			testName:          "Garbage before good FD request (3)",
			fakePodName:       "podA",
			fakePodNamespace:  "default",
			fakeResourceName:  "cndp/testing",
			cndpServerDevType: "cndp/testing",
			fakePodDevices:    []string{"devA", "devB"},
			cndpServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: requestConnect + ", podA",
				1: `\` + requestFd + ", devA",
				2: requestFin,
			},
			expectedResponse: map[int]string{
				0: responseHostOk,
				1: responseBadRequest,
				2: responseFinAck,
			},
		},
		{
			//Put garbage before an otherwise good FD request
			testName:          "Garbage before good FD request (4)",
			fakePodName:       "podA",
			fakePodNamespace:  "default",
			fakeResourceName:  "cndp/testing",
			cndpServerDevType: "cndp/testing",
			fakePodDevices:    []string{"devA", "devB"},
			cndpServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: requestConnect + ", podA",
				1: "\\\\" + requestFd + ", devA",
				2: requestFin,
			},
			expectedResponse: map[int]string{
				0: responseHostOk,
				1: responseBadRequest,
				2: responseFinAck,
			},
		},
		{
			//Put garbage before an otherwise good FD request
			testName:          "Garbage before good FD request (5)",
			fakePodName:       "podA",
			fakePodNamespace:  "default",
			fakeResourceName:  "cndp/testing",
			cndpServerDevType: "cndp/testing",
			fakePodDevices:    []string{"devA", "devB"},
			cndpServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: requestConnect + ", podA",
				1: "," + requestFd + ", devA",
				2: requestFin,
			},
			expectedResponse: map[int]string{
				0: responseHostOk,
				1: responseBadRequest,
				2: responseFinAck,
			},
		},
		{
			//Put null byte after an otherwise good FD request
			testName:          "Null byte after good FD request (1)",
			fakePodName:       "podA",
			fakePodNamespace:  "default",
			fakeResourceName:  "cndp/testing",
			cndpServerDevType: "cndp/testing",
			fakePodDevices:    []string{"devA", "devB"},
			cndpServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: requestConnect + ", podA",
				1: requestFd + "," + `\0`,
				2: requestFin,
			},
			expectedResponse: map[int]string{
				0: responseHostOk,
				1: responseFdNak,
				2: responseFinAck,
			},
		},
		{
			//Put null byte after an otherwise good FD request
			testName:          "Null byte after good FD request (2)",
			fakePodName:       "podA",
			fakePodNamespace:  "default",
			fakeResourceName:  "cndp/testing",
			cndpServerDevType: "cndp/testing",
			fakePodDevices:    []string{"devA", "devB"},
			cndpServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: requestConnect + ", podA",
				1: requestFd + "," + "\\0",
				2: requestFin,
			},
			expectedResponse: map[int]string{
				0: responseHostOk,
				1: responseFdNak,
				2: responseFinAck,
			},
		},
		{
			//Put garbage after an otherwise good FD request
			testName:          "Garbage after good FD request (1)",
			fakePodName:       "podA",
			fakePodNamespace:  "default",
			fakeResourceName:  "cndp/testing",
			cndpServerDevType: "cndp/testing",
			fakePodDevices:    []string{"devA", "devB"},
			cndpServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: requestConnect + ", podA",
				1: requestFd + ", devAdkjfhgkjdfs",
				2: requestFin,
			},
			expectedResponse: map[int]string{
				0: responseHostOk,
				1: responseFdNak,
				2: responseFinAck,
			},
		},
		{
			//Put garbage after an otherwise good FD request
			testName:          "Garbage after good FD request (2)",
			fakePodName:       "podA",
			fakePodNamespace:  "default",
			fakeResourceName:  "cndp/testing",
			cndpServerDevType: "cndp/testing",
			fakePodDevices:    []string{"devA", "devB"},
			cndpServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: requestConnect + ", podA",
				1: requestFd + ", devA,dkjfhgkjdfs",
				2: requestFin,
			},
			expectedResponse: map[int]string{
				0: responseHostOk,
				1: responseBadRequest,
				2: responseFinAck,
			},
		},
		{
			//Put garbage after an otherwise good FD request
			testName:          "Garbage after good FD request (3)",
			fakePodName:       "podA",
			fakePodNamespace:  "default",
			fakeResourceName:  "cndp/testing",
			cndpServerDevType: "cndp/testing",
			fakePodDevices:    []string{"devA", "devB"},
			cndpServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: requestConnect + ", podA",
				1: requestFd + ", devA" + `\`,
				2: requestFin,
			},
			expectedResponse: map[int]string{
				0: responseHostOk,
				1: responseFdNak,
				2: responseFinAck,
			},
		},
		{
			//Put garbage after an otherwise good FD request
			testName:          "Garbage after good FD request (4)",
			fakePodName:       "podA",
			fakePodNamespace:  "default",
			fakeResourceName:  "cndp/testing",
			cndpServerDevType: "cndp/testing",
			fakePodDevices:    []string{"devA", "devB"},
			cndpServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: requestConnect + ", podA",
				1: requestFd + ", devA\\\\",
				2: requestFin,
			},
			expectedResponse: map[int]string{
				0: responseHostOk,
				1: responseFdNak,
				2: responseFinAck,
			},
		},
		{
			//Put garbage after an otherwise good FD request
			testName:          "Garbage after good FD request (5)",
			fakePodName:       "podA",
			fakePodNamespace:  "default",
			fakeResourceName:  "cndp/testing",
			cndpServerDevType: "cndp/testing",
			fakePodDevices:    []string{"devA", "devB"},
			cndpServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: requestConnect + ", podA",
				1: requestFd + ", devA,",
				2: requestFin,
			},
			expectedResponse: map[int]string{
				0: responseHostOk,
				1: responseBadRequest,
				2: responseFinAck,
			},
		},
		/***************************************************
		Negative Tests - validate, but send garbage requests
		***************************************************/
		{
			//Validate but then send null byte
			testName:          "Validate and send null byte (1)",
			fakePodName:       "podA",
			fakePodNamespace:  "default",
			fakeResourceName:  "cndp/testing",
			cndpServerDevType: "cndp/testing",
			fakePodDevices:    []string{"devA", "devB"},
			cndpServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: requestConnect + ", podA",
				1: `\0`,
				2: requestFin,
			},
			expectedResponse: map[int]string{
				0: responseHostOk,
				1: responseBadRequest,
				2: responseFinAck,
			},
		},
		{
			//Validate but then send null byte
			testName:          "Validate and send null byte (2)",
			fakePodName:       "podA",
			fakePodNamespace:  "default",
			fakeResourceName:  "cndp/testing",
			cndpServerDevType: "cndp/testing",
			fakePodDevices:    []string{"devA", "devB"},
			cndpServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: requestConnect + ", podA",
				1: "\\0",
				2: requestFin,
			},
			expectedResponse: map[int]string{
				0: responseHostOk,
				1: responseBadRequest,
				2: responseFinAck,
			},
		},
		{
			//Validate but then send garbage request
			testName:          "Validate and send no request (1)",
			fakePodName:       "podA",
			fakePodNamespace:  "default",
			fakeResourceName:  "cndp/testing",
			cndpServerDevType: "cndp/testing",
			fakePodDevices:    []string{"devA", "devB"},
			cndpServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: requestConnect + ", podA",
				1: "",
				2: requestFin,
			},
			expectedResponse: map[int]string{
				0: responseHostOk,
				1: responseBadRequest,
				2: responseFinAck,
			},
		},
		{
			//Validate but then send garbage request
			testName:          "Validate and send no request (2)",
			fakePodName:       "podA",
			fakePodNamespace:  "default",
			fakeResourceName:  "cndp/testing",
			cndpServerDevType: "cndp/testing",
			fakePodDevices:    []string{"devA", "devB"},
			cndpServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: requestConnect + ", podA",
				1: " ",
				2: requestFin,
			},
			expectedResponse: map[int]string{
				0: responseHostOk,
				1: responseBadRequest,
				2: responseFinAck,
			},
		},
		{
			//Validate but then send garbage request
			testName:          "Validate and send garbage request (1)",
			fakePodName:       "podA",
			fakePodNamespace:  "default",
			fakeResourceName:  "cndp/testing",
			cndpServerDevType: "cndp/testing",
			fakePodDevices:    []string{"devA", "devB"},
			cndpServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: requestConnect + ", podA",
				1: ",",
				2: requestFin,
			},
			expectedResponse: map[int]string{
				0: responseHostOk,
				1: responseBadRequest,
				2: responseFinAck,
			},
		},
		{
			//Validate but then send garbage request
			testName:          "Validate and send garbage request (2)",
			fakePodName:       "podA",
			fakePodNamespace:  "default",
			fakeResourceName:  "cndp/testing",
			cndpServerDevType: "cndp/testing",
			fakePodDevices:    []string{"devA", "devB"},
			cndpServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: requestConnect + ", podA",
				1: "/",
				2: requestFin,
			},
			expectedResponse: map[int]string{
				0: responseHostOk,
				1: responseBadRequest,
				2: responseFinAck,
			},
		},
		{
			//Validate but then send garbage request
			testName:          "Validate and send garbage request (3)",
			fakePodName:       "podA",
			fakePodNamespace:  "default",
			fakeResourceName:  "cndp/testing",
			cndpServerDevType: "cndp/testing",
			fakePodDevices:    []string{"devA", "devB"},
			cndpServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: requestConnect + ", podA",
				1: `\`,
				2: requestFin,
			},
			expectedResponse: map[int]string{
				0: responseHostOk,
				1: responseBadRequest,
				2: responseFinAck,
			},
		},
		{
			//Validate but then send garbage request
			testName:          "Validate and send garbage request (4)",
			fakePodName:       "podA",
			fakePodNamespace:  "default",
			fakeResourceName:  "cndp/testing",
			cndpServerDevType: "cndp/testing",
			fakePodDevices:    []string{"devA", "devB"},
			cndpServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: requestConnect + ", podA",
				1: "dkjfhgkjdfs",
				2: requestFin,
			},
			expectedResponse: map[int]string{
				0: responseHostOk,
				1: responseBadRequest,
				2: responseFinAck,
			},
		},
		{
			//Validate but then send garbage request
			testName:          "Validate and send garbage request (5)",
			fakePodName:       "podA",
			fakePodNamespace:  "default",
			fakeResourceName:  "cndp/testing",
			cndpServerDevType: "cndp/testing",
			fakePodDevices:    []string{"devA", "devB"},
			cndpServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: requestConnect + ", podA",
				1: "*",
				2: requestFin,
			},
			expectedResponse: map[int]string{
				0: responseHostOk,
				1: responseBadRequest,
				2: responseFinAck,
			},
		},
		{
			//Validate but then send garbage request
			testName:          "Validate and send garbage request (6)",
			fakePodName:       "podA",
			fakePodNamespace:  "default",
			fakeResourceName:  "cndp/testing",
			cndpServerDevType: "cndp/testing",
			fakePodDevices:    []string{"devA", "devB"},
			cndpServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: requestConnect + ", podA",
				1: ";",
				2: requestFin,
			},
			expectedResponse: map[int]string{
				0: responseHostOk,
				1: responseBadRequest,
				2: responseFinAck,
			},
		},
		{
			//Validate but then send garbage request
			testName:          "Validate and send garbage request (7)",
			fakePodName:       "podA",
			fakePodNamespace:  "default",
			fakeResourceName:  "cndp/testing",
			cndpServerDevType: "cndp/testing",
			fakePodDevices:    []string{"devA", "devB"},
			cndpServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: requestConnect + ", podA",
				1: "\n",
				2: requestFin,
			},
			expectedResponse: map[int]string{
				0: responseHostOk,
				1: responseBadRequest,
				2: responseFinAck,
			},
		},
		{
			//Validate but then send garbage request
			testName:          "Validate and send garbage request (8)",
			fakePodName:       "podA",
			fakePodNamespace:  "default",
			fakeResourceName:  "cndp/testing",
			cndpServerDevType: "cndp/testing",
			fakePodDevices:    []string{"devA", "devB"},
			cndpServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: requestConnect + ", podA",
				1: requestVersion + requestFin,
				2: requestFin,
			},
			expectedResponse: map[int]string{
				0: responseHostOk,
				1: responseBadRequest,
				2: responseFinAck,
			},
		},
		{
			//Validate but then send garbage request
			testName:          "Validate and send garbage request (9)",
			fakePodName:       "podA",
			fakePodNamespace:  "default",
			fakeResourceName:  "cndp/testing",
			cndpServerDevType: "cndp/testing",
			fakePodDevices:    []string{"devA", "devB"},
			cndpServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: requestConnect + ", podA",
				1: requestVersion + " " + requestFin,
				2: requestFin,
			},
			expectedResponse: map[int]string{
				0: responseHostOk,
				1: responseBadRequest,
				2: responseFinAck,
			},
		},
		{
			//Validate but then send garbage request
			testName:          "Validate and send garbage request (10)",
			fakePodName:       "podA",
			fakePodNamespace:  "default",
			fakeResourceName:  "cndp/testing",
			cndpServerDevType: "cndp/testing",
			fakePodDevices:    []string{"devA", "devB"},
			cndpServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: requestConnect + ", podA",
				1: requestVersion + "," + requestFin,
				2: requestFin,
			},
			expectedResponse: map[int]string{
				0: responseHostOk,
				1: responseBadRequest,
				2: responseFinAck,
			},
		},
		{
			//Validate but then send garbage request
			testName:          "Validate and send garbage request (10)",
			fakePodName:       "podA",
			fakePodNamespace:  "default",
			fakeResourceName:  "cndp/testing",
			cndpServerDevType: "cndp/testing",
			fakePodDevices:    []string{"devA", "devB"},
			cndpServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: requestConnect + ", podA",
				1: requestVersion + "\n" + requestFin,
				2: requestFin,
			},
			expectedResponse: map[int]string{
				0: responseHostOk,
				1: responseBadRequest,
				2: responseFinAck,
			},
		},
		{
			//Validate but then send garbage request
			testName:          "Validate and send garbage request (11)",
			fakePodName:       "podA",
			fakePodNamespace:  "default",
			fakeResourceName:  "cndp/testing",
			cndpServerDevType: "cndp/testing",
			fakePodDevices:    []string{"devA", "devB"},
			cndpServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: requestConnect + ", podA",
				1: requestFd + requestFin,
				2: requestFin,
			},
			expectedResponse: map[int]string{
				0: responseHostOk,
				1: responseBadRequest,
				2: responseFinAck,
			},
		},
		{
			//Validate but then send garbage request
			testName:          "Validate and send garbage request (12)",
			fakePodName:       "podA",
			fakePodNamespace:  "default",
			fakeResourceName:  "cndp/testing",
			cndpServerDevType: "cndp/testing",
			fakePodDevices:    []string{"devA", "devB"},
			cndpServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: requestConnect + ", podA",
				1: requestFd + " " + requestFin,
				2: requestFin,
			},
			expectedResponse: map[int]string{
				0: responseHostOk,
				1: responseBadRequest,
				2: responseFinAck,
			},
		},
		{
			//Validate but then send garbage request
			testName:          "Validate and send garbage request (13)",
			fakePodName:       "podA",
			fakePodNamespace:  "default",
			fakeResourceName:  "cndp/testing",
			cndpServerDevType: "cndp/testing",
			fakePodDevices:    []string{"devA", "devB"},
			cndpServerDevices: []string{"devA", "devB"},
			fakeRequests: map[int]string{
				0: requestConnect + ", podA",
				1: requestFd + ", " + requestFin,
				2: requestFin,
			},
			expectedResponse: map[int]string{
				0: responseHostOk,
				1: responseFdNak,
				2: responseFinAck,
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			// make a new server each time to clear things like device list
			server := &server{
				deviceType: tc.cndpServerDevType,
				devices:    make(map[string]int),
				uds:        fakeUDS,
				podRes:     fakeResAPI,
			}

			fakeResAPI.CreateFakePod(tc.fakePodName, tc.fakePodNamespace, tc.fakeResourceName, tc.fakePodDevices)
			fakeUDS.SetRequests(tc.fakeRequests)

			for fd, device := range tc.cndpServerDevices {
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
