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

package uds

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestCtrlBufHasValue(t *testing.T) {

	testCases := []struct {
		name      string
		data      []byte
		expReturn bool
	}{
		{
			name:      "first_test_1111",
			data:      []byte{1, 1, 1, 1},
			expReturn: true,
		},

		{
			name:      "second_test_0000",
			data:      []byte{0, 0, 0, 0},
			expReturn: false,
		},

		{
			name:      "third_test_1010",
			data:      []byte{1, 0, 1, 0},
			expReturn: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			actualReturn := ctrlBufHasValue(tc.data)

			assert.Equal(t, tc.expReturn, actualReturn, "Returned value does not match expected value")

		})
	}
}

func TestInit(t *testing.T) {
	myUDSHandler := NewHandler()

	testCases := []struct {
		testName   string
		socketPath string
		protocol   string
		msgBufSize int
		ctlBufSize int
		timeout    time.Duration
		expErr     error
	}{
		{
			testName:   "socket does not exist",
			socketPath: "/file/does/not/exist.sock",
			protocol:   "unixpacket",
			msgBufSize: 64,
			ctlBufSize: 4,
			timeout:    20,
			expErr:     errors.New("unknown network /file/does/not/exist.sock"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {

			err := myUDSHandler.Init(tc.socketPath, tc.protocol, tc.msgBufSize, tc.ctlBufSize, tc.timeout)

			if err != nil {
				require.Error(t, tc.expErr, err, "Error was expected")
				assert.Contains(t, err.Error(), tc.expErr.Error(), "Unexpected error returned")
			}
		})
	}
}
