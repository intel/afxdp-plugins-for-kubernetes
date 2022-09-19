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

package uds

import (
	"fmt"
	fuzz "github.com/google/gofuzz"
	"github.com/intel/afxdp-plugins-for-kubernetes/internal/logformats"
	logging "github.com/sirupsen/logrus"
	"io"
	"os"
	"time"
)

const (
	fuzzFile       = "/var/log/afxdp-k8s-plugins/fuzz.log"
	filePermission = 0644
)

var firstCall bool = true

/*
FuzzHandler interface extends the Handler interface to provide additional testing methods.
*/
type FuzzHandler interface {
	Handler
}

/*
fuzzHandler implements the Handler interface.
*/
type fuzzHandler struct {
}

/*
NewFuzzHandler returns a fuzz implementation of the Handler interface.
*/
func NewFuzzHandler() Handler {
	return &fuzzHandler{}
}

/*
Init should initialises the Unix domain socket. The fuzzlogging() function is called which creates a separate
file for fuzzing logs.
*/
func (f *fuzzHandler) Init(socketPath string, protocol string, msgbufSize int, ctlBufSize int, timeout time.Duration, uid string) error {
	if err := fuzzLogging(); err != nil {
		return err
	}

	return nil
}

/*
Listen listens for and accepts new connections.
fuzzHandler returns nil as it's functionality isn't required for fuzz testing.
*/
func (f *fuzzHandler) Listen() (CleanupFunc, error) {
	return func() {}, nil
}

/*
Dial creates a new connection.
fuzzHandler returns nil as it's functionality isn't required for fuzz testing.
*/
func (f *fuzzHandler) Dial() (CleanupFunc, error) {
	return func() {}, nil
}

/*
Read should read the incoming message from the UDS.
FuzzHandler seeds malformed fuzzing data to CNDP read().
*/
func (f *fuzzHandler) Read() (string, int, error) {
	var fuzzResponse string
	var fd int = 0

	if firstCall {
		fuzzResponse = "/connect, afxdp-fuzz-test"
	} else {

		f := fuzz.New()
		f.Fuzz(&fuzzResponse)
	}

	firstCall = false
	return fuzzResponse, fd, nil
}

/*
Write should write a string to the UDS.
fuzzHandler returns nil as it's functionality isn't required for fuzz testing.
*/
func (f *fuzzHandler) Write(response string, fd int) error {
	return nil
}

func fuzzLogging() error {

	logging.SetReportCaller(true)

	fp, err := os.OpenFile(fuzzFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, filePermission)
	if err != nil {
		err = fmt.Errorf("fuzzlogging(): cannot open logfile %s: %w", fuzzFile, err)
		return err
	}
	logging.SetOutput(io.MultiWriter(fp, os.Stdout))
	logging.SetFormatter(logformats.Debug)

	return nil
}
