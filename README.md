# CNDP Kubernetes Plugins

A proof of concept Kubernetes device plugin and CNI plugin to provide AF_XDP networking to Kubernetes pods using Intel's Cloud Native Data Plane framework.

## Prerequisites
### Required
The following prerequisites are required to build and deploy the plugins:

- **Docker**
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
- **Kubernetes**
 	- All recent versions should work. Tested on `1.20.2`, `1.21.1`.
- **A CNI network**
	- To serve as the [default network](https://github.com/k8snetworkplumbingwg/multus-cni/blob/master/docs/quickstart.md#key-concepts) to the Kubernetes pods.
	- Any CNI should work. Tested with [Flannel](https://github.com/flannel-io/flannel).
- **Multus CNI**
	- To enable attaching multiple network interfaces to pods.
	- [Multus quickstart guide.](https://github.com/k8snetworkplumbingwg/multus-cni/blob/master/docs/quickstart.md)
- **GoLang**
	- To build the plugin binaries.
	- All recent versions should work. Tested on `1.13.8`, TODO - more versions
	- [Download and install](https://golang.org/doc/install).
- **BPF Library**
	- To load and unload the XDP program onto the network device.
	- TODO - link to install steps

### Development
The following static analysis, linting and formatting tools are not required for building and deploying, but are built into some of the Make targets and enforced by CI. It is recommended to have these installed on your development system.

- **[GoFmt](https://pkg.go.dev/cmd/gofmt)**
	- Applies standard formatting to Go code.
	- Supplied with GoLang.
- **[Go Vet](https://pkg.go.dev/cmd/vet)**
	- Examines Go source code and reports suspicious constructs.
	- Supplied with GoLang.
- **[Go Lint](https://github.com/golang/lint)**
	- A linter for Go source code.
	- Install: `go get -u golang.org/x/lint/golint`
	- *Note: Deprecated, but still useful in day to day development as a quick check*
- **[GolangCI-Lint](https://golangci-lint.run/)**
	- A Go linters aggregator.
	- Install: `curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.42.1`
- **[Hadolint](https://github.com/hadolint/hadolint)**
	- A Dockerfile linter that helps build best practice into Docker images.
	- Runs in Docker container.
- **[Shellcheck](https://github.com/koalaman/shellcheck)**
	- A static analysis tool for shell scripts.
	- Install on Ubuntu: `apt install shellcheck`
- **[Clang Format](https://clang.llvm.org/docs/ClangFormat.html)**
	- Applies standard formatting to C code.
	- Install on Ubuntu: `apt install clang-format`
- **[CLOC](https://github.com/AlDanial/cloc)**
	- Count Lines Of Code, counts lines of code in many programming languages.
	- Install on Ubuntu `apt install cloc`

## Build and Deploy

 - Clone this repo and `cd` into it.
 - Run `make deploy`

This following steps happen **automatically**:

 1. `make build` is executed, resulting in CNI and Device Plugin
    binaries in `./bin`.
2. `make image` is executed, resulting in the creation of a new Docker image that includes the Device Plugin binary.
	- ***Note:** if testing on a multi-node cluster. In the current absence of a Docker registry, this image will need to be manually copied to all nodes (or rebuilt on all nodes using: `make image`).*
3. The damenonset running the Device Plugin image will then be deployed on all nodes in the cluster.

**Note:** The daemonset deploys the Device Plugin but does not yet deploy the CNI plugin.
`./bin/cndp` should be copied to `/opt/cni/bin/` on all nodes. 

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
Go                              19            513            748           4319
Markdown                         2             57              0            227
YAML                             6             11             25            213
Bourne Shell                     1             14              0            170
C                                2             48             25            159
make                             1             12              2             62
C/C++ Header                     2             11             24             15
JSON                             1              0              0             10
Dockerfile                       1              2              0              7
-------------------------------------------------------------------------------
SUM:                            35            668            824           5182
-------------------------------------------------------------------------------

```
<!---clocend--->

## Sequence Diagram
![](./docs/sequence.png)
