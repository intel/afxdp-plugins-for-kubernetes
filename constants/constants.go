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

package constants

var (
	/* Plugins */
	pluginModes                   = []string{"cndp"} // accepted plugin modes
	devicePluginDefaultConfigFile = "./config.json"  // device plugin default config file if none explicitly provided
	devicePluginDevicePrefix      = "afxdp"          // default name prefix that the device plugin gives to its devices
	devicePluginExitNormal        = 0                // device plugin normal exit code
	devicePluginExitConfigError   = 1                // device plugin config error exit code, problem with the provided config
	devicePluginExitLogError      = 2                // device plugin logging error exit code, error creating log file, bad log level, etc.
	devicePluginExitHostError     = 3                // device plugin host check exit code, error occurred checking some atribute of the host
	devicePluginExitPoolError     = 4                // device plugin device pool exit code, error occurred while building a device pool

	/* Logging */
	logLevels          = []string{"debug", "info", "warning", "error"} // accepted log levels
	logDirectory       = "/var/log/afxdp-k8s-plugins/"                 // log file directory
	logDirPermissions  = 0744                                          // permissions for log directory
	logFilePermissions = 0644                                          // permissions for log file
	regexValidFile     = `^/$|^(/[a-zA-Z0-9._-]+)+$`                   // regex to check if a string is a valid filepath
	regexLogDir        = `^` + logDirectory                            // regex to check if the logfile is in the correct directory

	/* Devices */
	prohibitedDevices = []string{"eno", "eth", "lo", "docker", "flannel", "cni"} // interfaces we never add to a pool
	regexValidDevice  = `^[a-zA-Z0-9_-]+$`                                       // regex to check if a string is a valid device name

	/* Drivers */
	zeroCopyDrivers  = []string{"i40e", "E810", "ice"} // drivers that support zero copy AF_XDP
	cdqDrivers       = []string{"ice"}                 // drivers that support CDQ subfunctions
	regexValidDriver = `^[a-zA-Z0-9_-]+$`              // regex to check if a string is a valid driver name

	/* AF_XDP */
	afxdpMinimumLinux = "4.18.0" // minimum Linux version for AF_XDP support

	/* CNDP UDS*/
	maxUdsTimeout     = 300              // maximum configurable uds timeout in seconds
	defaultUdsTimeout = 30               // default uds timeout in seconds
	udsMsgBufSize     = 64               // uds message buffer size
	udsCtlBufSize     = 4                // uds control buffer size
	udsProtocol       = "unixpacket"     // uds protocol: "unix"=SOCK_STREAM, "unixdomain"=SOCK_DGRAM, "unixpacket"=SOCK_SEQPACKET
	udsSockDir        = "/tmp/afxdp_dp/" // host location where we place our uds sockets. If changing location remember to update daemonset mount point
	udsDirFileMode    = 0700             // permissions for the directory in which we create our uds sockets

	/* CNDP Handshake*/
	cndpHandshakeVersion    = "0.1"                   // increase this version if changes are made to the protocol below
	cndpRequestVersion      = "/version"              // used to request the handshake version
	cndpRequestConnect      = "/connect"              // used to request a new connection, this request will be combined with the podname
	cndpResponseHostOk      = "/host_ok"              // the response given if a valid podname was sent along with the connection request
	cndpResponseHostNak     = "/host_nak"             // the response given if an invalid podname was sent with the connection request
	cndpRequestFd           = "/xsk_map_fd"           // used to request the xsk map file descriptor for a network device, this request will be combined with the device name
	cndpResponseFdAck       = "/fd_ack"               // the response given if the xsk map file descriptor for a device can be provided, the file descriptor will be in the response control buffer
	cndpResponseFdNak       = "/fd_nak"               // the response given if there was a problem providing the xsk map file descriptor for a device, there will be no file descriptor included
	cndpRequestBusyPoll     = "/config_busy_poll"     // used to request configuration of busy poll, this request will be combined with busy budget and timeout values and a file descriptor in the rerquest control buffer
	cndpResponseBusyPollAck = "/config_busy_poll_ack" // the response given if busy poll was successfully configured
	cndpResponseBusyPollNak = "/config_busy_poll_nak" // the response given if there was a problem configuring busy poll
	cndpRequestFin          = "/fin"                  // used to request connection termination
	cndpResponseFinAck      = "/fin_ack"              // the response given to acknowledge the connection termination request
	cndpResponseBadRequest  = "/nak"                  // general non-acknowledgement response, usually indicates a bad request
	cndpResponseError       = "/error"                // general error occurred response, indicates an error occurred on the device plugin end
)

/* Public variables and types */
var (
	/* Plugins contains constants related to both CNI and device plugin */
	Plugins plugins
	/* Afxdp contains constants related to the AF_XDP technology */
	Afxdp afxdp
	/* Logging contains constants related logging */
	Logging logging
	/* Drivers contains constants related netdev drivers */
	Drivers drivers
	/* Devices contains constants related to netdevs */
	Devices devices
	/* Cndp contains constants related to intel's Cloud Native Dataplane */
	Cndp cndp
)

type cni struct {
}

type devicePlugin struct {
	DefaultConfigFile string
	DevicePrefix      string
	ExitNormal        int
	ExitConfigError   int
	ExitLogError      int
	ExitHostError     int
	ExitPoolError     int
}

type plugins struct {
	Modes        []string
	Cni          cni
	DevicePlugin devicePlugin
}

type afxdp struct {
	MinumumKernel string
}

type drivers struct {
	ZeroCopy       []string
	Cdq            []string
	RegexValidName string
}

type devices struct {
	Prohibited     []string
	RegexValidName string
}

type logging struct {
	Levels               []string
	Directory            string
	DirectoryPermissions int
	FilePermissions      int
	RegexValidFile       string
	RegexCorrectDir      string
}

type uds struct {
	MaxTimeout     int
	DefaultTimeout int
	MsgBufSize     int
	CtlBufSize     int
	Protocol       string
	SockDir        string
	DirFileMode    int
}

type cndpHandshake struct {
	Version             string
	RequestVersion      string
	RequestConnect      string
	ResponseHostOk      string
	ResponseHostNak     string
	RequestFd           string
	ResponseFdAck       string
	ResponseFdNak       string
	RequestBusyPoll     string
	ResponseBusyPollAck string
	ResponseBusyPollNak string
	RequestFin          string
	ResponseFinAck      string
	ResponseBadRequest  string
	ResponseError       string
}

type cndp struct {
	Uds       uds
	Handshake cndpHandshake
}

func init() {
	Plugins = plugins{
		Modes: pluginModes,
		DevicePlugin: devicePlugin{
			DefaultConfigFile: devicePluginDefaultConfigFile,
			DevicePrefix:      devicePluginDevicePrefix,
			ExitNormal:        devicePluginExitNormal,
			ExitConfigError:   devicePluginExitConfigError,
			ExitLogError:      devicePluginExitLogError,
			ExitHostError:     devicePluginExitHostError,
			ExitPoolError:     devicePluginExitPoolError,
		},
	}

	Afxdp = afxdp{
		MinumumKernel: afxdpMinimumLinux,
	}

	Drivers = drivers{
		ZeroCopy:       zeroCopyDrivers,
		Cdq:            cdqDrivers,
		RegexValidName: regexValidDriver,
	}

	Devices = devices{
		Prohibited:     prohibitedDevices,
		RegexValidName: regexValidDevice,
	}

	Logging = logging{
		Levels:               logLevels,
		Directory:            logDirectory,
		DirectoryPermissions: logDirPermissions,
		FilePermissions:      logFilePermissions,
		RegexValidFile:       regexValidFile,
		RegexCorrectDir:      regexLogDir,
	}

	Cndp = cndp{
		Uds: uds{
			MaxTimeout:     maxUdsTimeout,
			DefaultTimeout: defaultUdsTimeout,

			MsgBufSize:  udsMsgBufSize,
			CtlBufSize:  udsCtlBufSize,
			Protocol:    udsProtocol,
			SockDir:     udsSockDir,
			DirFileMode: udsDirFileMode,
		},
		Handshake: cndpHandshake{
			Version:             cndpHandshakeVersion,
			RequestVersion:      cndpRequestVersion,
			RequestConnect:      cndpRequestConnect,
			ResponseHostOk:      cndpResponseHostOk,
			ResponseHostNak:     cndpResponseHostNak,
			RequestFd:           cndpRequestFd,
			ResponseFdAck:       cndpResponseFdAck,
			ResponseFdNak:       cndpResponseFdNak,
			RequestBusyPoll:     cndpRequestBusyPoll,
			ResponseBusyPollAck: cndpResponseBusyPollAck,
			ResponseBusyPollNak: cndpResponseBusyPollNak,
			RequestFin:          cndpRequestFin,
			ResponseFinAck:      cndpResponseFinAck,
			ResponseBadRequest:  cndpResponseBadRequest,
			ResponseError:       cndpResponseError,
		},
	}
}
