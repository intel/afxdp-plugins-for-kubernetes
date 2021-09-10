# CNDP Device Plugin

A proof of concept Kubernetes device plugin and CNI plugin to provide AF_XDP networking to Kubernetes pods using Intel's Cloud Native Data Plane framework.

## Prerequisites

- Docker installed and running.
	- All recent versions should work. Tested on `20.10.5`, `20.10.7`.
	- **Note:** You may need to disable memlock on Docker.
		Add the following section to `/etc/docker/daemon.json`:
		```
		"default-ulimits": {
		"memlock": {
			"Name": "memlock",
			"Hard": -1,
			"Soft": -1
			}
		}
		```
		Restart the Docker service: `systemctl restart docker.service`
- Kubernetes installed an running.
 	- All recent versions should work. Tested on `1.20.2`, `1.21.1`.
- A CNI network to serve as the [default network](https://github.com/k8snetworkplumbingwg/multus-cni/blob/master/docs/quickstart.md#key-concepts) to the Kubernetes pods.
	- Any CNI should work. Tested with [Flannel](https://github.com/flannel-io/flannel).
- [Multus CNI](https://github.com/k8snetworkplumbingwg/multus-cni), to enable attaching multiple network interfaces to pods.
	- [Multus quickstart guide.](https://github.com/k8snetworkplumbingwg/multus-cni/blob/master/docs/quickstart.md)

## Clone and Build

```bash
git clone ssh://git@gitlab.devtools.intel.com:29418/OrchSW/CNO/containers_afxdp_network_device_plugin-.git
cd containers_afxdp_network_device_plugin-/
make build

Common dependencies and packages that are requried:

 - Go ($ wget -c https://dl.google.com/go/go1.17.linux-amd64.tar.gz -O - | sudo tar -xz -C /usr/local)
 - Clang Format - C Formatter (apt instal clang-format)
 - CLOC - cloc counts blank lines, comment lines, and physical lines of source code in many programming languages. (sudo apt-get install -y cloc)

```
Two binaries will be placed in ./bin directory:
- **cndp-dp** is the device plugin
- **cndp** is the CNI plugin. This needs to be placed in `/opt/cni/bin/`

## Device Plugin Config
The device plugin currently reads a list of devices from a config file, rather than actual device discovery. A sample config file can be found in `examples/sample-config/`
For actual testing with the CNI, the config file should be updated to include only the names of real netdevs that exist on the node.
By default the device plugin will search for `config.json` in the current directory. An alternative path can be provided using the `-config` flag, e.g.
```bash
./bin/cndp-dp -config ./examples/sample-config/config.json
```


# Logging Overview

This logging framework consists of several customisable features. This is particularly helpful for debugging and helps support learning of code processes and its configurations. 



Logging configurations for CNI and Device Plugin are devised in separate files:
- **cndp-dp:**  DP Config                   `/examples/e2e-test/config.json`
- **cndp-cni:** networkAttachmentDefinition `/examples/e2e-test/nad.yaml`

## Logging Level

Specifying a logging level enables filtering and differentiation of log events based on severity.
```bash
"logLevel": "debug",
```


As illustrated above, the default logging level has been set to ```debug```, this is the maximum level of verbosity. Increasing the logging level reduces filtering and severity of log entries, meaning more basic indications are captured in the log output.  

•	```panic ```: The application has encountered a severe problem and will exit immediately.

•	```error ```: A defect has occurred i.e., invalid input or inability to access a particular service. The application will eventually exit the code.

•	```warning ```: A defect has occurred i.e., invalid input or inability to access a particular service. Application will continue irrespective of the unusual event.

•	```info ```: Basic information, indication of major code paths.

•	```debug ```: Additional information, indication of minor code branches.

## Writing to a Log File
There is also an option to log to a file on the file system. 
The file will be created on the node executing the CNI or Device Plugin.

```bash
"logFile": "debug",
````
The log file path can be configured via the logFile field. The default log file path for both the CNI and Device Plugin are set to ```/var/log/```.


## CLOC
Output from CLOC (count lines of code) - github.com/AlDanial/cloc 
<!---clocstart--->
```
-------------------------------------------------------------------------------
Language                     files          blank        comment           code
-------------------------------------------------------------------------------
Go                              19            506            730           4033
Markdown                         2             86              0            318
YAML                             6             11             26            203
Bourne Shell                     1             14              0            170
C                                2             48             25            159
make                             1             12              2             62
C/C++ Header                     2             11             24             15
JSON                             1              0              0             10
Dockerfile                       1              2              0              7
-------------------------------------------------------------------------------
SUM:                            35            690            807           4977
-------------------------------------------------------------------------------

```
<!---clocend--->

## Sequence Diagram
![](./docs/sequence.png)
