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

package main

import (
	"github.com/intel/cndp_device_plugin/internal/uds"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	udsProtocol     = "unixpacket"
	udsPath         = "/tmp/cndp.sock"
	udsMsgBufSize   = 64
	udsCtlBufSize   = 4
	udsIdleTimeout  = 0 * time.Second
	requestDelay    = 100 * time.Millisecond // not required but keeps things in nice order when DP and this test app are both printing to screen
	timeoutDuration = 40                     // For UDS timeout test - timeoutDuration must exceed timeout value set in config.json.
)

var udsHandler uds.Handler

func main() {
	timeoutAfterConnect := false
	timeoutBeforeConnect := false
	// Command line argument to set timeout test
	timeoutArgs := os.Args[1:]
	for _, arg := range timeoutArgs {
		if arg == "-timeout-before-connect" {
			timeoutBeforeConnect = true
		}
		if arg == "-timeout-after-connect" {
			timeoutAfterConnect = true
		}
	}

	//Get environment variable device values
	devicesVar, exists := os.LookupEnv("CNDP_DEVICES")
	if !exists {
		println("Test App Error: Devices env var does not exist")
		os.Exit(1)
	}
	devices := strings.Split(devicesVar, " ")

	hostname, exists := os.LookupEnv("HOSTNAME")
	if !exists {
		println("Test App Error: Hostname env var does not exist")
		os.Exit(1)
	}

	udsHandler = uds.NewHandler()

	// init
	if err := udsHandler.Init(udsPath, udsProtocol, udsMsgBufSize, udsCtlBufSize, udsIdleTimeout); err != nil {
		println("Test App Error: Error Initialising UDS server: ", err)
		os.Exit(1)
	}

	// Execute timeoutBeforeConnect when set to true
	if timeoutBeforeConnect {
		println("Test App - Executing timeout before connect")
		timeout()
	}

	cleanup, err := udsHandler.Dial()
	if err != nil {
		println("Test App Error: UDS Dial error:: ", err)
		cleanup()
		os.Exit(1)
	}
	defer cleanup()

	// connect and verify pod hostname
	makeRequest("/connect, " + hostname)
	time.Sleep(requestDelay)

	// Execute timeoutAfterConnect when set to true
	if timeoutAfterConnect {
		println("Test App - Executing timeout after connect")
		timeout()
	}

	// request version
	makeRequest("/version")
	time.Sleep(requestDelay)

	// request XSK map FD for all devices
	for _, dev := range devices {
		request := "/xsk_map_fd, " + dev
		makeRequest(request)
		time.Sleep(requestDelay)
	}

	// request an unknown device
	makeRequest("/xsk_map_fd, bad-device")
	time.Sleep(requestDelay)

	// send a bad request
	makeRequest("/bad-request")
	time.Sleep(requestDelay)

	// finish
	makeRequest("/fin")
	time.Sleep(requestDelay)
}

func timeout() {
	println("Test App - Pausing for", timeoutDuration, "seconds to force timeout")
	println("Test App - Expecting timeout error to occur")
	time.Sleep(timeoutDuration * time.Second)
	println("Test App - Exiting")
	os.Exit(0)
}

func makeRequest(request string) {
	println()
	println("Test App - Request: " + request)

	if err := udsHandler.Write(request, -1); err != nil {
		println("Test App - Write error: ", err)
	}

	response, fd, err := udsHandler.Read()
	if err != nil {
		println("Test App - Read error: ", err)
	}

	println("Test App - Response: " + response)
	if fd > 0 {
		println("Test App - File Descriptor:", strconv.Itoa(fd))
	}
	println()
}
