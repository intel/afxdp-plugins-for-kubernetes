package main

import (
	"errors"
	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
	"github.com/containernetworking/plugins/pkg/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGetConfig(t *testing.T) {
	netConf := types.NetConf{
		CNIVersion: "0.3.0",
		Name:       "test-network",
		Type:       "cndp", Capabilities: map[string]bool(nil),
		IPAM: types.IPAM{Type: ""},
		DNS: types.DNS{Nameservers: []string(nil), Domain: "",
			Search:  []string(nil),
			Options: []string(nil)},
		RawPrevResult: map[string]interface{}(nil),
		PrevResult:    types.Result(nil),
	}

	testCases := []struct {
		name      string
		config    string
		expConfig *netConfig
		expErr    error
	}{
		{
			name:      "load good config 1",
			config:    `{"cniVersion":"0.3.0","deviceID":"dev_1","name":"test-network","pciBusID":"","type":"cndp"}`,
			expConfig: &netConfig{NetConf: netConf, Device: "dev_1"},
		},

		{
			name:      "load no config",
			config:    `{ }`,
			expConfig: nil,
			expErr:    errors.New("loadConf(): no device specified"),
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
			expErr:    errors.New("loadConf(): no device specified"),
		},
		{
			name:      "load bad config 4 - incorrect JSON format",
			config:    `{"cniVersion":"0.3.0","deviceID":" },`,
			expConfig: nil,
			expErr:    errors.New("loadConf(): failed to load network configuration: unexpected end of JSON input"),
		},
		{
			name:      "load bad config 5 - invalid character",
			config:    `{"cniVersion":"0.3.0","deviceID":"dev_1","name":"test-network","pciBusID":"","type":"cndp"}}`,
			expConfig: nil,
			expErr:    errors.New("loadConf(): failed to load network configuration: invalid character '}' after top-level value"),
		},
		{
			name:      "load bad config 6 - invalid character 2",
			config:    `{"cniVersion":"0.3.0",%"deviceID":"dev_1","name":"test-network",%"pciBusID":"","type":"cndp"}}`,
			expConfig: nil,
			expErr:    errors.New("loadConf(): failed to load network configuration: invalid character '%' looking for beginning of object key string"),
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
			expError:   "loadConf(): no device specified",
		},

		{
			name:       "fail to parse netConf - missing brace",
			netConfStr: "{ ",
			netNS:      "",
			expError:   "loadConf(): failed to load network configuration: unexpected end of JSON input",
		},
		{
			name:       "fail to open netns -  no device name",
			netConfStr: `{"cniVersion":"0.3.0","deviceID":" ","name":"test-network","pciBusID":"","type":"cndp"}`,
			netNS:      "",
			expError:   "cmdAdd(): failed to open container netns \"\": failed to Statfs \"\": no such file or directory",
		},

		{
			name:       "fail to open netns -  bad netns",
			netConfStr: `{"cniVersion":"0.3.0","deviceID":"dev_1","name":"test-network","pciBusID":"","type":"cndp"}`,
			netNS:      "B@dN%eTNS",
			expError:   "cmdAdd(): failed to open container netns \"B@dN%eTNS\": failed to Statfs \"B@dN%eTNS\": no such file or directory",
		},
		{
			name:       "fail to parse netConf - generate netns",
			netConfStr: `{"cniVersion":"0.3.0","deviceID":"dev_1","name":"test-network","pciBusID":"","type":"cndp"}`,
			netNS:      "generate",
			expError:   "cmdAdd(): failed to open container netns \"generate\": failed to Statfs \"generate\": no such file or directory",
		},
	}

	for _, tc := range testCases {

		t.Run(tc.name, func(t *testing.T) {

			args.StdinData = []byte(tc.netConfStr)
			args.Netns = tc.netNS
			err := cmdAdd(args)

			if tc.netNS == "generate" {
				netNS, nsErr := testutils.NewNS()
				require.NoError(t, nsErr, "Can't create NewNS")
				defer testutils.UnmountNS(netNS)
				args.Netns = netNS.Path()
			} else {
				args.Netns = tc.netNS
			}

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
			expError:   "loadConf(): no device specified",
		},

		{
			name:       "bad load configuration - inncorrect JSON Formatting",
			netConfStr: "{ ",
			expError:   "loadConf(): failed to load network configuration: unexpected end of JSON input",
		},
	}

	for _, tc := range testCases {

		t.Run(tc.name, func(t *testing.T) {

			args.StdinData = []byte(tc.netConfStr)
			args.Netns = tc.netNS
			err := cmdDel(args)

			netNS, nsErr := testutils.NewNS()
			require.NoError(t, nsErr, "Can't create NewNS")
			defer testutils.UnmountNS(netNS)
			args.Netns = netNS.Path()
			args.Netns = tc.netNS

			if tc.expError == " " {
				require.Error(t, err, "Unexpected error")
			} else {
				require.Error(t, err, "Unexpected error")
				assert.Contains(t, err.Error(), tc.expError, "Unexpected error")
			}

		})
	}
}
