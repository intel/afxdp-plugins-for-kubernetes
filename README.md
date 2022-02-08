# AF_XDP Kubernetes Plugins

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
	- [Multus quickstart guide](https://github.com/k8snetworkplumbingwg/multus-cni/blob/master/docs/quickstart.md).
- **GoLang**
	- To build the plugin binaries.
	- All recent versions should work. Tested on `1.13.8`, `1.15.15` and `1.17.1`.
	- [Download and install](https://golang.org/doc/install).
- **Libbpf**
	- To load and unload the XDP program onto the network device.
	- Install on Ubuntu 20.10+: `apt install libbpf-dev`
	- Older versions: [Install from source](https://github.com/libbpf/libbpf#build).
- **GCC Compiler**
	- To compile the C code used to call on the BPF Library.
	- Install on Ubuntu: `apt install build-essential`
- **Binutils**
	- Used in archiving of C code object file.
	- Install on Ubuntu: `apt install binutils`

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
	- Install on Ubuntu: `apt install cloc`

## Build and Deploy

 - Clone this repo and `cd` into it.
 - Optional: Update configuration. See [Device Plugin Config](#device-plugin-config).
 - Run `make deploy`.

The following steps happen **automatically**:

1. `make build` is executed, resulting in CNI and Device Plugin binaries in `./bin`.
2. `make image` is executed, resulting in the creation of a new Docker image that includes the CNI and Device Plugin binaries.
	- ***Note:** if testing on a multi-node cluster. The current absence of a Docker registry means this image will need to be manually copied to all nodes (or rebuilt on all nodes using: `make image`).*
3. The damenonset will run on all nodes, installing the CNI and starting the Device Plugin running on each node.

The CNI and Device Plugin are now deployed.

#### Running Pods

- Create a network attachment definition file. This is the config for the CNI plugin.
	- An example file can be found under [examples/network-attachment-definition.yaml](./examples/network-attachment-definition.yaml)
	- Change the config if necessary. See comments in the example file.
	- `kubectl create -f network-attachment-definition.yaml`
- Create a pod spec:
	- An example pod spec can be found under [examples/pod-spec.yaml](./examples/pod-spec.yaml)
	- Configure the pod spec to use a suitable Docker image and to reference the network attachment definition as well as the resource type from the Device Plugin. See comments in the example file.
	- `kubectl create -f pod-spec.yaml`


## Device Plugin Config
Under normal circumstances the device plugin config is set as part of a config map at the top of the [daemonset.yml](./deployments/daemonset.yml) file.

The device plugin binary can also be run manually on the host for development and testing purposes. In these scenarios the device plugin will search for a `config.json` file in its current directory, or the device plugin can be pointed to a config file using the `-config` flag followed by a filepath. 

### Default Behaviour
If no config is given, the default behaviour of the device plugin is to discover AF_XDP capable devices on the node and create device pools based on driver type.
For example, a host with a single 4-port X710 NIC will result in 4 devices being added to the `cndp/i40e` pool.

### Driver Pools
It is possible to have multiple driver types in a single device pool. The example below will result in a pool named `cndp/intel` that contains all the x710 and all E810 devices on the node.

```
{
    "pools" : [
        {
            "name" : "intel",
            "drivers" : ["i40e", "E810"]
        }
    ]
}
```

### Device Pools
It is possible to assign individual devices to a pool. The example below will generate a pool named `cndp/test` with the two listed devices.
This is not scalable over many nodes and is intended only for development and testing purposes.

```
{
    "pools" : [
        {
            "name" : "test",
            "devices" : ["ens801f0", "ens801f1"],
        }
    ]
}
```

### Logging
A log file and log level can be configured for the device plugin. As above, these are set in the config map at the top of the [daemonset.yml](./deployments/daemonset.yml) file. Or, as above, a `config.json` file.
- The log file is set using `logFile`. This file should be placed under `/var/log/afxdp-k8s-plugins/`. 
- The log level is set using `logLevel`. Available options are:
	- `error` - Only logs errors.
	- `warning` - Logs errors and warnings.
	- `info` - Logs errors, warnings and basic info about the operation of the device plugin.
	- `debug` - Logs all of the above along with additional in-depth info about the operation of the device plugin.
- Example config including log settings:

```
{
    "logLevel": "debug",
    "logFile": "/var/log/afxdp-k8s-plugins/cndp-dp.log",
    "timeout": 30,
    "pools" : [
        {
            "name" : "i40e",
            "drivers" : ["i40e"]
        }
    ]
}
```

### Mode

The device plugin allows for different modes of operation. CNDP is the only mode at present, with additional modes to be implemented in due course.
Mode type must be configured for both device plugin and CNI. 

Mode setting for device plugin is set via the `config.json` file. Please see example below:

```
{
    "mode": "cndp"
    "pools" : [
        {
            "name" : "i40e",
            "drivers" : ["i40e"]
        }
    ]
}
```

Mode setting for CNI is set via the network-attachment-definition(NAD) file `NAD.yml`. Please see example below:

```
apiVersion: "k8s.cni.cncf.io/v1"
kind: NetworkAttachmentDefinition
metadata:
  name: cndp-e2e-test
  annotations:
    k8s.v1.cni.cncf.io/resourceName: cndp/e2e
spec:
  config: '{
      "cniVersion": "0.3.0",
      "type": "cndp-e2e",
      "mode": "cndp",
      "logFile": "/var/log/afxdp-k8s-plugins/cndp-cni-e2e.log",
      "logLevel": "debug",
      "ipam": {
        "type": "host-local",
        "subnet": "192.168.1.0/24",
        "rangeStart": "192.168.1.200",
        "rangeEnd": "192.168.1.216",
        "routes": [
          { "dst": "0.0.0.0/0" }
        ],
        "gateway": "192.168.1.1"
      }
    }
```

### Timeout 
The device plugin includes a timeout action for the unix domain sockets(UDS). 
Once the timeout is invoked, the UDS is closed and disconnected.

The timeout is configured to a default of 90 seconds. 
Setting the timeout to zero will disable the timeout action, which will allow the UDS to remain idle.

The timeout value is set in the `config.json` file. Please see example below.

```
{
    "timeout": 30,
    "pools" : [
        {
            "name" : "i40e",
            "drivers" : ["i40e"]
        }
    ]
}
```

## CLOC
Output from CLOC (count lines of code) - github.com/AlDanial/cloc 
<!---clocstart--->
```
-------------------------------------------------------------------------------
Language                     files          blank        comment           code
-------------------------------------------------------------------------------
Go                              25            546            858           4521
Markdown                         3             72              0            431
YAML                            10             16             17            420
Bourne Shell                     3             18              2            249
C                                2             34             29            151
make                             1             17              0            125
C/C++ Header                     2             10             28             28
JSON                             1              0              0             12
Dockerfile                       1              0              0              3
-------------------------------------------------------------------------------
SUM:                            48            713            934           5940
-------------------------------------------------------------------------------

```
<!---clocend--->

## Sequence Diagram
![](./docs/sequence.png)
