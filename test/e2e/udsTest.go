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

package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/intel/afxdp-plugins-for-kubernetes/constants"
	"github.com/intel/afxdp-plugins-for-kubernetes/internal/uds"
	"github.com/intel/afxdp-plugins-for-kubernetes/pkg/goclient"
)

const (
	udsIdleTimeout  = 0 * time.Second
	requestDelay    = 100 * time.Millisecond // not required but keeps things in nice order when DP and this test app are both printing to screen
	timeoutDuration = 40                     // For UDS timeout test - timeoutDuration must exceed timeout value set in config.json.
)

var udsHandler uds.Handler

func main() {
	if os.Args[1] == "uds" {
		udsTest()
	} else if os.Args[1] == "golang" {
		clientTest()
	} else {
		println("Unrecognized testing parameters.")
	}
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

func udsTest() {
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
	devicesVar, exists := os.LookupEnv(constants.Devices.EnvVarList)
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
	if err := udsHandler.Init(constants.Uds.PodPath, constants.Uds.Protocol, constants.Uds.MsgBufSize, constants.Uds.CtlBufSize, udsIdleTimeout, ""); err != nil {
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
	cleanup()
}

func clientTest() {

	var clean2 uds.CleanupFunc
	var fd int

	//Get environment variable device values
	devicesVar, exists := os.LookupEnv(constants.Devices.EnvVarList)
	if !exists {
		println("Test App Error: Devices env var does not exist")
		os.Exit(1)
	}
	devices := strings.Split(devicesVar, " ")

	// Request Client Version
	println("GO Library: Requesting client version from GO library")
	ver := goclient.GetClientVersion()
	fmt.Printf("GO Library: Client Version: %s \n \n", ver)
	time.Sleep(requestDelay)

	// Request Server Version
	println("GO Library: Requesting server version from GO library")
	ver, clean1, _ := goclient.GetServerVersion()
	fmt.Printf("GO Library: Server Version: %s \n \n", ver)
	time.Sleep(requestDelay)

	// Request XSK map FD for all devices
	println("GO Library: Requesting XSK map FD")
	for _, dev := range devices {
		fd, clean2, _ = goclient.RequestXSKmapFD(dev)
		fmt.Printf("GO Library: XSK map FD request succeded for %s with fd %d \n \n", dev, fd)
		// time.Sleep(requestDelay)
	}
	time.Sleep(requestDelay)

	// Request XSK map FD for an unknown device
	println("GO Library: Request XSk map FD for an unknown device")
	fd, clean3, err := goclient.RequestXSKmapFD("bad-device")
	if err != nil {
		fmt.Printf("GO Library: Returned value from unknown device: %d \n \n", fd)
	}
	time.Sleep(requestDelay)

	println("Cleaning up... \n")
	clean1()
	clean2()
	clean3()
}
