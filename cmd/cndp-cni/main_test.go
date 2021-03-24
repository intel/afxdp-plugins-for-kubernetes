package main

import (
	"errors"
	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
	"github.com/containernetworking/cni/pkg/types/current"
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
		expConfig *NetConf
		expErr    error
	}{
		{
			name:      "load good config 1",
			config:    `{"cniVersion":"0.3.0","deviceID":"dev_1","name":"test-network","pciBusID":"","type":"cndp"}`,
			expConfig: &NetConf{NetConf: netConf, Device: "dev_1"},
		},

		{
			name:      "load no config",
			config:    `{ }`,
			expConfig: nil,
			expErr:    errors.New("loadConf(): specify a \"device\" - field blank"),
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
			expErr:    errors.New("loadConf(): loading network configuration unsuccessful : unexpected end of JSON input"),
		},
		{
			name:      "load bad config 3 - empty braces",
			config:    `{}`,
			expConfig: nil,
			expErr:    errors.New("loadConf(): specify a \"device\" - field blank"),
		},
		{
			name:      "load bad config 4 - incorrect JSON format",
			config:    `{"cniVersion":"0.3.0","deviceID":" },`,
			expConfig: nil,
			expErr:    errors.New("loadConf(): loading network configuration unsuccessful : unexpected end of JSON input"),
		},
		{
			name:      "load bad config 5 - invalid character",
			config:    `{"cniVersion":"0.3.0","deviceID":"dev_1","name":"test-network","pciBusID":"","type":"cndp"}}`,
			expConfig: nil,
			expErr:    errors.New("loadConf(): loading network configuration unsuccessful : invalid character '}' after top-level value"),
		},
		{
			name:      "load bad config 6 - invalid character 2",
			config:    `{"cniVersion":"0.3.0",%"deviceID":"dev_1","name":"test-network",%"pciBusID":"","type":"cndp"}}`,
			expConfig: nil,
			expErr:    errors.New("loadConf(): loading network configuration unsuccessful : invalid character '%' looking for beginning of object key string"),
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
			expError:   "loadConf(): loading network configuration unsuccessful : unexpected end of JSON input",
		},
		{
			name:       "fail to parse netConf - no arguments",
			netConfStr: "{}",
			netNS:      "",
			expError:   "loadConf(): specify a \"device\" - field blank",
		},

		{
			name:       "fail to parse netConf - missing brace",
			netConfStr: "{ ",
			netNS:      "",
			expError:   "loadConf(): loading network configuration unsuccessful : unexpected end of JSON input",
		},
		{
			name:       "fail to open netns -  no device name",
			netConfStr: `{"cniVersion":"0.3.0","deviceID":" ","name":"test-network","pciBusID":"","type":"cndp"}`,
			netNS:      "",
			expError:   "cmdAdd(): failed to open netns \"\": failed to Statfs \"\": no such file or directory",
		},

		{
			name:       "fail to open netns -  bad netns",
			netConfStr: `{"cniVersion":"0.3.0","deviceID":"dev_1","name":"test-network","pciBusID":"","type":"cndp"}`,
			netNS:      "B@dN%eTNS",
			expError:   "cmdAdd(): failed to open netns \"B@dN%eTNS\": failed to Statfs \"B@dN%eTNS\": no such file or directory",
		},
		{
			name:       "fail to parse netConf - generate netns",
			netConfStr: `{"cniVersion":"0.3.0","deviceID":"dev_1","name":"test-network","pciBusID":"","type":"cndp"}`,
			netNS:      "generate",
			expError:   "cmdAdd(): failed to open netns \"generate\": failed to Statfs \"generate\": no such file or directory",
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
			expError:   "loadConf(): loading network configuration unsuccessful : unexpected end of JSON input",
		},
		{
			name:       "bad load configuration - no arguments and no no device specified",
			netConfStr: "{} ",
			expError:   "loadConf(): specify a \"device\" - field blank",
		},

		{
			name:       "bad load configuration - inncorrect JSON Formatting",
			netConfStr: "{ ",
			expError:   "loadConf(): loading network configuration unsuccessful : unexpected end of JSON input",
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

func TestCmdCheck(t *testing.T) {
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
			expError:   "loadConf(): loading network configuration unsuccessful : unexpected end of JSON input",
		},
		{
			name:       "bad load configuration - empty device name",
			netConfStr: "{} ",
			expError:   "loadConf(): specify a \"device\" - field blank",
		},

		{
			name:       "bad load configuration - JSON Format incorrect",
			netConfStr: "{ ",
			expError:   "loadConf(): loading network configuration unsuccessful : unexpected end of JSON input",
		},
	}

	for _, tc := range testCases {

		t.Run(tc.name, func(t *testing.T) {

			args.StdinData = []byte(tc.netConfStr)
			args.Netns = tc.netNS
			err := cmdCheck(args)

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

func TestGetLink(t *testing.T) {

	testCases := []struct {
		name    string
		devname string
		expErr  error
	}{
		{
			name:    "fail to find physical interface - empty device name",
			devname: "",
			expErr:  errors.New("getLink(): failed to find physical interface"),
		},
		{
			name:    "fail to find physical interface - full device name",
			devname: "enp1s0f3",
			expErr:  errors.New("Link not found"),
		},
	}

	for _, tc := range testCases {

		t.Run(tc.name, func(t *testing.T) {
			_, err := getLink(tc.devname)

			if err == nil {
				assert.Equal(t, tc.expErr, err, "Error was expected")
			} else {
				require.Error(t, tc.expErr, "Unexpected error returned")
				assert.Contains(t, err.Error(), tc.expErr.Error(), "Unexpected error returned")
			}

		})
	}
}

func TestValidateCniContainerInterface(t *testing.T) {

	intf := current.Interface{}

	testCases := []struct {
		name        string
		intfname    string
		intfMac     string
		intfSandbox string
		expintf     current.Interface
		expErr      error
	}{

		{
			name:        "failure to find container interface - missing intf",
			intfname:    "",
			intfMac:     "Mac",
			intfSandbox: "Sandbox",
			expintf:     intf,
			expErr:      errors.New("validateCniContainerInterface(): Container interface name missing in prevResult"),
		},
		{
			name:        "failure to find container interface - missing intf: incomplete brace",
			intfname:    "{",
			intfMac:     "Mac",
			intfSandbox: "Sandbox",
			expintf:     intf,
			expErr:      errors.New("validateCniContainerInterface(): Container Interface name in prevResult: { not found"),
		},
		{
			name:        "test 3 - full intf",
			intfname:    "enp1s0f3",
			intfMac:     "Mac",
			intfSandbox: "Sandbox",
			expintf:     intf,
			expErr:      errors.New("validateCniContainerInterface(): Container Interface name in prevResult: enp1s0f3 not found"),
		},
	}

	for _, tc := range testCases {

		t.Run(tc.name, func(t *testing.T) {

			intf.Name = tc.intfname
			intf.Mac = tc.intfMac
			intf.Sandbox = tc.intfSandbox

			err := validateCniContainerInterface(intf)

			if err == nil {
				assert.Equal(t, tc.expErr, err, "Error was expected")
			} else {
				require.Error(t, tc.expErr, "Unexpected error returned")
				assert.Contains(t, err.Error(), tc.expErr.Error(), "Unexpected error returned")
			}

		})
	}
}
