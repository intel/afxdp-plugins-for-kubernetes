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
	"errors"
	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
	"github.com/intel/afxdp-plugins-for-kubernetes/internal/bpf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGetConfig(t *testing.T) {
	netConf := types.NetConf{
		CNIVersion:   "0.3.0",
		Name:         "test-network",
		Type:         "afxdp",
		Capabilities: map[string]bool(nil),
		IPAM:         types.IPAM{Type: ""},
		DNS: types.DNS{Nameservers: []string(nil), Domain: "",
			Search:  []string(nil),
			Options: []string(nil)},
		RawPrevResult: map[string]interface{}(nil),
		PrevResult:    types.Result(nil),
	}

	testCases := []struct {
		name      string
		config    string
		expConfig *NetConfig
		expErr    error
	}{
		{
			name:      "load good config 1",
			config:    `{"cniVersion":"0.3.0","deviceID":"dev1","name":"test-network","pciBusID":"","type":"afxdp","mode":"cdq","Queues":"4"}`,
			expConfig: &NetConfig{NetConf: netConf, Device: "dev1", Mode: "cdq", Queues: "4"},
		},
		{
			name:      "load no config",
			config:    `{ }`,
			expConfig: nil,
			expErr:    errors.New("validate(): no device specified"),
		},
		{
			name:      "load bad config 1 - no JSON Format",
			config:    ``,
			expConfig: nil,
			expErr:    errors.New("unexpected end of JSON input"),
		},

		{
			name:      "load bad config 2 - Missing Brace",
			config:    `{`,
			expConfig: nil,
			expErr:    errors.New("loadConf(): failed to load network configuration: unexpected end of JSON input"),
		},
		{
			name:      "load bad config 3 - empty braces",
			config:    `{}`,
			expConfig: nil,
			expErr:    errors.New("validate(): no device specified"),
		},
		{
			name:      "load bad config 4 - incorrect JSON format",
			config:    `{"cniVersion":"0.3.0","deviceID":" },`,
			expConfig: nil,
			expErr:    errors.New("loadConf(): failed to load network configuration: unexpected end of JSON input"),
		},
		{
			name:      "load bad config 5 - invalid character",
			config:    `{"cniVersion":"0.3.0","deviceID":"dev_1","name":"test-network","pciBusID":"","type":"afxdp"}}`,
			expConfig: nil,
			expErr:    errors.New("loadConf(): failed to load network configuration: invalid character '}' after top-level value"),
		},
		{
			name:      "load bad config 6 - invalid character 2",
			config:    `{"cniVersion":"0.3.0",%"deviceID":"dev_1","name":"test-network",%"pciBusID":"","type":"afxdp"}}`,
			expConfig: nil,
			expErr:    errors.New("loadConf(): failed to load network configuration: invalid character '%' looking for beginning of object key string"),
		},
		{
			name:      "load good config 7 - bad device name",
			config:    `{"cniVersion":"0.3.0","deviceID":"dev1^","name":"test-network","pciBusID":"","type":"afxdp","mode":"primary","Queues":"4"}`,
			expConfig: nil,
			expErr:    errors.New("loadConf(): Config validation error: deviceID: device names must only contain letters, numbers and selected symbols"),
		},
	}

	for _, tc := range testCases {

		t.Run(tc.name, func(t *testing.T) {

			rawConfig := []byte(tc.config)
			cfg, err := loadConf(rawConfig)

			if err == nil {
				assert.Equal(t, tc.expErr, err, "Error was expected")
			} else {
				require.Error(t, tc.expErr, "Unexpected error returned")
				assert.Contains(t, err.Error(), tc.expErr.Error(), "Unexpected error returned")
			}
			assert.Equal(t, tc.expConfig, cfg, "Returned unexpected config")

		})
	}
}

func TestCmdAdd(t *testing.T) {
	args := &skel.CmdArgs{}

	testCases := []struct {
		name       string
		netConfStr string
		netNS      string
		expError   string
		fakeErr    error
	}{
		{
			name:       "fail to parse netConf - no braces",
			netConfStr: "",
			netNS:      "",
			expError:   "loadConf(): failed to load network configuration: unexpected end of JSON input",
		},
		{
			name:       "fail to parse netConf - no arguments",
			netConfStr: "{}",
			netNS:      "",
			expError:   "validate(): no device specified",
		},

		{
			name:       "fail to parse netConf - missing brace",
			netConfStr: "{ ",
			netNS:      "",
			expError:   "loadConf(): failed to load network configuration: unexpected end of JSON input",
		},

		{
			name:       "no device name",
			netConfStr: `{"cniVersion":"0.3.0","deviceID":"","name":"test-network","pciBusID":"","type":"afxdp"}`,
			netNS:      "",
			expError:   "validate(): no device specified",
		},

		{
			name:       "fail to open netns - bad netns",
			netConfStr: `{"cniVersion":"0.3.0","deviceID":"dev1","name":"test-network","pciBusID":"","type":"afxdp"}`,
			netNS:      "B@dN%eTNS",
			expError:   "cmdAdd(): failed to open container netns \"B@dN%eTNS\": failed to Statfs \"B@dN%eTNS\": no such file or directory",
		},
	}

	for _, tc := range testCases {

		t.Run(tc.name, func(t *testing.T) {

			args.StdinData = []byte(tc.netConfStr)
			args.Netns = tc.netNS
			err := CmdAdd(args)

			if tc.expError == " " {
				require.Error(t, err, "Unexpected error")
			} else {
				require.Error(t, err, "Unexpected error")
				assert.Contains(t, err.Error(), tc.expError, "Unexpected error")
			}

		})
	}
}

func TestCmdDel(t *testing.T) {
	args := &skel.CmdArgs{}

	testCases := []struct {
		name       string
		netConfStr string
		netNS      string
		expError   string
		fakeErr    error
	}{
		{
			name:       "bad load configuration - empty configuration",
			netConfStr: "",
			expError:   "loadConf(): failed to load network configuration: unexpected end of JSON input",
		},
		{
			name:       "bad load configuration - no arguments and no no device specified",
			netConfStr: "{} ",
			expError:   "validate(): no device specified",
		},

		{
			name:       "bad load configuration - inncorrect JSON Formatting",
			netConfStr: "{ ",
			expError:   "loadConf(): failed to load network configuration: unexpected end of JSON input",
		},
	}

	for _, tc := range testCases {

		t.Run(tc.name, func(t *testing.T) {
			bpfHandler = bpf.NewFakeHandler()
			args.StdinData = []byte(tc.netConfStr)
			args.Netns = tc.netNS
			err := CmdDel(args)

			if tc.expError == " " {
				require.Error(t, err, "Unexpected error")
			} else {
				require.Error(t, err, "Unexpected error")
				assert.Contains(t, err.Error(), tc.expError, "Unexpected error")
			}
		})
	}
}
