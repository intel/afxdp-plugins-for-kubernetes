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
	"github.com/intel/afxdp-plugins-for-kubernetes/internal/deviceplugin"
	"github.com/intel/afxdp-plugins-for-kubernetes/internal/host"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCheckHost(t *testing.T) {
	testCases := []struct {
		testName          string
		hostKernal        string
		hostPrivilegedBpf bool
		cfg               deviceplugin.Config
		expResult         bool
		expErr            error
	}{
		{
			testName:          "Test checkhost for correct linuxVersion version",
			hostKernal:        "5.11.0",
			hostPrivilegedBpf: true,
			cfg: deviceplugin.Config{
				RequireUnprivilegedBpf: false,
				MinLinuxVersion:        "4.18.0",
			},
			expResult: true,
			expErr:    nil,
		},
		{
			testName:          " Test Checkhost linuxVersion does not meet minimum version requirement",
			hostKernal:        "4.18.0",
			hostPrivilegedBpf: true,
			cfg: deviceplugin.Config{
				RequireUnprivilegedBpf: false,
				MinLinuxVersion:        "7.18.0",
			},
			expResult: false,
			expErr:    nil,
		},
		{
			testName:          "Test checkhost is passing an empty string as LinuxVersion",
			hostKernal:        "",
			hostPrivilegedBpf: true,
			cfg: deviceplugin.Config{
				RequireUnprivilegedBpf: false,
				MinLinuxVersion:        "4.18.0",
			},
			expResult: false,
			expErr:    nil,
		},
		{
			testName:          "Test checkhost is passing of whole number for LinuxVersion",
			hostKernal:        "6",
			hostPrivilegedBpf: true,
			cfg: deviceplugin.Config{
				RequireUnprivilegedBpf: false,
				MinLinuxVersion:        "4.18.0",
			},
			expResult: false,
			expErr:    nil,
		},
		{
			testName:          "Test checkhost is passing false hostPrivilegedBpf with RequiredUnprivilegedBpf set as false",
			hostKernal:        "5.11.0",
			hostPrivilegedBpf: false,
			cfg: deviceplugin.Config{
				RequireUnprivilegedBpf: false,
				MinLinuxVersion:        "4.18.0",
			},
			expResult: true,
			expErr:    nil,
		},
		{
			testName:          "Test checkhost is passing false hostPrivilegedBpf with RequiredUnprivilegedBpf set as true",
			hostKernal:        "5.11.0",
			hostPrivilegedBpf: false,
			cfg: deviceplugin.Config{
				RequireUnprivilegedBpf: true,
				MinLinuxVersion:        "4.18.0",
			},
			expResult: false,
			expErr:    nil,
		},
		{
			testName:          "\"Test checkhost is passing true hostPrivilegedBpf with RequiredUnprivilegedBpf set as true",
			hostKernal:        "5.11.0",
			hostPrivilegedBpf: true,
			cfg: deviceplugin.Config{
				RequireUnprivilegedBpf: true,
				MinLinuxVersion:        "4.18.0",
			},
			expResult: true,
			expErr:    nil,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {

			handler := host.NewFakeHandler()

			handler.SetKernalVersion(tc.hostKernal)
			handler.SetAllowsUnprivilegedBpf(tc.hostPrivilegedBpf)
			actualReturn, err := checkHost(handler, tc.cfg)

			assert.Equal(t, tc.expResult, actualReturn, "Returned error on test")

			if err != nil {
				require.Error(t, tc.expErr, err, "Error was expected")
				assert.Contains(t, err.Error(), tc.expErr.Error(), "Unexpected error returned")
			}
		})
	}
}

func TestIntVersion(t *testing.T) {
	testCases := []struct {
		name      string
		version   string
		expResult int64
		expError  error
	}{
		{
			name:      "first_test_5.4.0-89-generic",
			version:   "5.4.0-89-generic",
			expResult: 500040000,
			expError:  nil,
		},
		{
			name:      "second_test_5.4.0-89",
			version:   "5.4.0-89",
			expResult: 500040000,
			expError:  nil,
		},
		{
			name:      "third_test_5.4.0-generic",
			version:   "5.4.0-generic",
			expResult: 500040000,
			expError:  nil,
		},
		{
			name:      "fourth_test_5.4.0",
			version:   "5.4.0",
			expResult: 500040000,
			expError:  nil,
		},
		{
			name:      "fifth_test_5.4.0--generic",
			version:   "5.4.0--generic",
			expResult: 500040000,
			expError:  nil,
		},
		{
			name:      "sixth_test_5.4.0-43-.generic",
			version:   "5.4.0-43-.generic",
			expResult: 500040000,
			expError:  nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			actualReturn, err := intVersion(tc.version)

			assert.Equal(t, tc.expResult, actualReturn, "Returned value does not match expected value")

			if err != nil {
				require.Error(t, tc.expError, err, "Error was expected")
				assert.Contains(t, err.Error(), tc.expError.Error(), "Unexpected error returned")
			}
		})
	}

}
