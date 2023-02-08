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
	pluginModes                   = []string{"primary", "cdq"} // accepted plugin modes
	devicePluginDefaultConfigFile = "./config.json"            // device plugin default config file if none explicitly provided
	devicePluginDevicePrefix      = "afxdp"                    // devive name prefix that the device plugin gives to devices, devices will be of type prefix/poolName
	devicePluginExitNormal        = 0                          // device plugin normal exit code
	devicePluginExitConfigError   = 1                          // device plugin config error exit code, problem with the provided config
	devicePluginExitLogError      = 2                          // device plugin logging error exit code, error creating log file, bad log level, etc.
	devicePluginExitHostError     = 3                          // device plugin host check exit code, error occurred checking some attribute of the host
	devicePluginExitPoolError     = 4                          // device plugin device pool exit code, error occurred while building a device pool

	/* Logging */
	logLevels          = []string{"debug", "info", "warning", "error"} // accepted log levels
	logDirectory       = "/var/log/afxdp-k8s-plugins/"                 // log file directory
	logDirPermissions  = 0744                                          // permissions for log directory
	logFilePermissions = 0644                                          // permissions for log file
	logValidFileRegex  = `^[a-zA-Z0-9_-]+(\.log|\.txt)$`               // regex to check if a string is a valid log filename

	/* Devices */
	devicesProhibited    = []string{"eno", "eth", "lo", "docker", "flannel", "cni"} // interfaces we never add to a pool
	devicesEnvVar        = "AFXDP_DEVICES"                                          // env var set in the end user application pod, lists AF_XDP devices attached
	deviceValidNameRegex = `^[a-zA-Z0-9_-]+$`                                       // regex to check if a string is a valid device name
	deviceValidNameMin   = 1                                                        // minimum length of a device name
	deviceValidNameMax   = 50                                                       // maximum length of a device name
	deviceValidPciRegex  = `[0-9a-f]{4}:[0-9a-f]{2,4}:[0-9a-f]{2}\.[0-9a-f]`        // regex to check if a string is a valid pci address
	deviceSecondaryMin   = 1                                                        // minimum number of secondary devices that can be created on top of a primary device
	deviceSecondaryMax   = 64                                                       // maximum number of secondary devices that can be created on top of a primary device

	/* Drivers */
	driversZeroCopy      = []string{"i40e", "E810", "ice"} // drivers that support zero copy AF_XDP
	driversCdq           = []string{"ice"}                 // drivers that support CDQ subfunctions
	driverValidNameRegex = `^[a-zA-Z0-9_-]+$`              // regex to check if a string is a valid driver name
	driverValidNameMin   = 1                               // minimum length of a driver name
	driverValidNameMax   = 50                              // maximum length of a deiver name
	driverPrimaryMin     = 1                               // minimum number of primary devices a driver can take from a node
	driverPrimaryMax     = 10                              // maximum number of primary devices a driver can take from a node

	/* Nodes */
	nodeValidHostRegex = `^[a-zA-Z0-9-]+$` // regex to check if a string is a valid node name
	nodeValidNameMin   = 1                 // minimum length of a node name
	nodeValidNameMax   = 63                // maximum length of a node name

	/* Pools */
	poolValidNameMin = 1  // minimum length of a pool name
	poolValidNameMax = 20 // maximum length of a pool name

	/* UID */
	uidMaximum = 256000 // maximum UID supported by BusyBox adduser
	uidMinimum = 1000   // minimum non-reserved UID in Alpine

	/* AF_XDP */
	afxdpMinimumLinux = "4.18.0" // minimum Linux version for AF_XDP support

	/* UDS*/
	udsMaxTimeout = 300               // maximum configurable uds timeout in seconds
	udsMinTimeout = 30                // minimum (and default) uds timeout in seconds
	udsMsgBufSize = 64                // uds message buffer size
	udsCtlBufSize = 4                 // uds control buffer size
	udsProtocol   = "unixpacket"      // uds protocol: "unix"=SOCK_STREAM, "unixdomain"=SOCK_DGRAM, "unixpacket"=SOCK_SEQPACKET
	udsSockDir    = "/tmp/afxdp_dp/"  // host location where we place our uds sockets. If changing location remember to update daemonset mount point
	udsPodPath    = "/tmp/afxdp.sock" // the uds filepath as it will appear in the end user application pod

	udsDirFileMode = 0700 // permissions for the directory in which we create our uds sockets

	/* Handshake*/
	handshakeHandshakeVersion    = "0.1"                   // increase this version if changes are made to the protocol below
	handshakeRequestVersion      = "/version"              // used to request the handshake version
	handshakeRequestConnect      = "/connect"              // used to request a new connection, this request will be combined with the podname
	handshakeResponseHostOk      = "/host_ok"              // the response given if a valid podname was sent along with the connection request
	handshakeResponseHostNak     = "/host_nak"             // the response given if an invalid podname was sent with the connection request
	handshakeRequestFd           = "/xsk_map_fd"           // used to request the xsk map file descriptor for a network device, this request will be combined with the device name
	handshakeResponseFdAck       = "/fd_ack"               // the response given if the xsk map file descriptor for a device can be provided, the file descriptor will be in the response control buffer
	handshakeResponseFdNak       = "/fd_nak"               // the response given if there was a problem providing the xsk map file descriptor for a device, there will be no file descriptor included
	handshakeRequestBusyPoll     = "/config_busy_poll"     // used to request configuration of busy poll, this request will be combined with busy budget and timeout values and a file descriptor in the rerquest control buffer
	handshakeResponseBusyPollAck = "/config_busy_poll_ack" // the response given if busy poll was successfully configured
	handshakeResponseBusyPollNak = "/config_busy_poll_nak" // the response given if there was a problem configuring busy poll
	handshakeRequestFin          = "/fin"                  // used to request connection termination
	handshakeResponseFinAck      = "/fin_ack"              // the response given to acknowledge the connection termination request
	handshakeResponseBadRequest  = "/nak"                  // general non-acknowledgement response, usually indicates a bad request
	handshakeResponseError       = "/error"                // general error occurred response, indicates an error occurred on the device plugin end

	/*DeviceFile*/
	name            = "device.json"    // file which enables passing of device information from device plugin to CNI in the form of device map object.
	directory       = "/tmp/afxdp_dp/" // host location where deviceFile file is placed.
	filePermissions = 0600             // permissions for device file.

	/*EthtoolFilters*/
	ethtoolFilterRegex = `^[a-zA-Z0-9-:.-/\s/g]+$` // regex to validate ethtool filter commands.
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
	/* UID contains contains constants related to user ids */
	UID uid
	/* Nodes contains constants related to kubernetes nodes */
	Nodes nodes
	/* Pools contains constants related to device pools */
	Pools pools
	/* Uds contains constants related to the Unix domain sockets */
	Uds uds
	/* DeviceFile contains constants related to the devicefile */
	DeviceFile deviceFile
	/* DeviceFile contains constants related to the devicefile */
	EthtoolFilter ethtoolFilter
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
	ValidNameRegex string
	ValidNameMin   int
	ValidNameMax   int
	PrimaryMin     int
	PrimaryMax     int
}

type devices struct {
	Prohibited     []string
	EnvVarList     string
	ValidNameRegex string
	ValidNameMin   int
	ValidNameMax   int
	ValidPciRegex  string
	SecondaryMin   int
	SecondaryMax   int
}

type nodes struct {
	ValidNameRegex string
	ValidNameMin   int
	ValidNameMax   int
}

type pools struct {
	ValidNameMin int
	ValidNameMax int
}

type uid struct {
	Maximum int
	Minimum int
}

type logging struct {
	Levels               []string
	Directory            string
	DirectoryPermissions int
	FilePermissions      int
	ValidFileRegex       string
}

type uds struct {
	MaxTimeout  int
	MinTimeout  int
	MsgBufSize  int
	CtlBufSize  int
	Protocol    string
	SockDir     string
	DirFileMode int
	PodPath     string
	Handshake   handshake
}

type handshake struct {
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

type deviceFile struct {
	Name            string
	FilePermissions int
	Directory       string
}

type ethtoolFilter struct {
	EthtoolFilterRegex string
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
		ZeroCopy:       driversZeroCopy,
		Cdq:            driversCdq,
		ValidNameRegex: driverValidNameRegex,
		ValidNameMin:   driverValidNameMin,
		ValidNameMax:   driverValidNameMax,
		PrimaryMin:     driverPrimaryMin,
		PrimaryMax:     driverPrimaryMax,
	}

	Devices = devices{
		Prohibited:     devicesProhibited,
		EnvVarList:     devicesEnvVar,
		ValidNameRegex: deviceValidNameRegex,
		ValidNameMin:   deviceValidNameMin,
		ValidNameMax:   deviceValidNameMax,
		ValidPciRegex:  deviceValidPciRegex,
		SecondaryMin:   deviceSecondaryMin,
		SecondaryMax:   deviceSecondaryMax,
	}

	Nodes = nodes{
		ValidNameRegex: nodeValidHostRegex,
		ValidNameMin:   nodeValidNameMin,
		ValidNameMax:   nodeValidNameMax,
	}

	Pools = pools{
		ValidNameMin: poolValidNameMin,
		ValidNameMax: poolValidNameMax,
	}

	UID = uid{
		Maximum: uidMaximum,
		Minimum: uidMinimum,
	}

	Logging = logging{
		Levels:               logLevels,
		Directory:            logDirectory,
		DirectoryPermissions: logDirPermissions,
		FilePermissions:      logFilePermissions,
		ValidFileRegex:       logValidFileRegex,
	}

	Uds = uds{
		MaxTimeout:  udsMaxTimeout,
		MinTimeout:  udsMinTimeout,
		MsgBufSize:  udsMsgBufSize,
		CtlBufSize:  udsCtlBufSize,
		Protocol:    udsProtocol,
		SockDir:     udsSockDir,
		DirFileMode: udsDirFileMode,
		PodPath:     udsPodPath,
		Handshake: handshake{
			Version:             handshakeHandshakeVersion,
			RequestVersion:      handshakeRequestVersion,
			RequestConnect:      handshakeRequestConnect,
			ResponseHostOk:      handshakeResponseHostOk,
			ResponseHostNak:     handshakeResponseHostNak,
			RequestFd:           handshakeRequestFd,
			ResponseFdAck:       handshakeResponseFdAck,
			ResponseFdNak:       handshakeResponseFdNak,
			RequestBusyPoll:     handshakeRequestBusyPoll,
			ResponseBusyPollAck: handshakeResponseBusyPollAck,
			ResponseBusyPollNak: handshakeResponseBusyPollNak,
			RequestFin:          handshakeRequestFin,
			ResponseFinAck:      handshakeResponseFinAck,
			ResponseBadRequest:  handshakeResponseBadRequest,
			ResponseError:       handshakeResponseError,
		},
	}

	DeviceFile = deviceFile{
		Name:            name,
		FilePermissions: filePermissions,
		Directory:       directory,
	}

	EthtoolFilter = ethtoolFilter{
		EthtoolFilterRegex: ethtoolFilterRegex,
	}
}
