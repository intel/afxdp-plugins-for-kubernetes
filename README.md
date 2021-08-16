# CNDP Device Plugin

A proof of concept Kubernetes device plugin and CNI plugin to provide AF_XDP networking to Kubernetes pods using Intel's Cloud Native Data Plane framework.

## Clone and Build

```bash
git clone ssh://git@gitlab.devtools.intel.com:29418/OrchSW/CNO/containers_afxdp_network_device_plugin-.git
cd containers_afxdp_network_device_plugin-/
./build.sh
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
Go                              20            491            647           3997
Markdown                         3            104              0            352
YAML                             6             13              2            232
Bourne Shell                     2             18              3            219
C                                2             48             24            159
make                             1             11              2             58
JSON                             3              0              0             26
C/C++ Header                     2             11             24             15
Dockerfile                       1              0              0              3
-------------------------------------------------------------------------------
SUM:                            40            696            702           5061
-------------------------------------------------------------------------------

```
<!---clocend--->

## Sequence Diagram
```plantuml
@startuml

skinparam sequenceGroupBorderThickness 1

actor "User"
participant "Linux"
participant "Kubelet"
box "Device Plugin"
	participant "Device Plugin"
	participant "UDS server"
end box
participant "CNI"
participant "Pod/CNDP App"

== Initialization ==
autonumber

"User" -> "Kubelet": network attachment definition (CNI config)
"User" -> "Device Plugin": deploy
activate "Device Plugin"

"Device Plugin" -> "Device Plugin": config.json
"Device Plugin" -> "Linux": net.Interfaces()
activate "Device Plugin" #DarkRed
note right #DarkRed: <color #White>discover resources</color>

"Linux" --> "Device Plugin": interface list

"Device Plugin" -> "Linux": ethtool.DriverName()
activate "Device Plugin" #FireBrick
note right #FireBrick: <color #White>loop interfaces, build device list</color>
"Linux" --> "Device Plugin": driver
deactivate "Device Plugin"

"Device Plugin" [hidden]-> "Device Plugin"


deactivate "Device Plugin"

"Device Plugin" -> "Device Plugin": start GRPC
"Device Plugin" -> "Kubelet": GRPC: register
"Device Plugin" -> "Kubelet": GRPC: device list

deactivate "Device Plugin"

== Pod Creation ==
autonumber

"User" -> "Kubelet": create pod
"Kubelet" -> "Device Plugin": GRPC: Allocate(ifname)
activate "Device Plugin"

"Device Plugin" -> "Linux": net.if_nametoindex(ifname)
"Linux" --> "Device Plugin": if_index
"Device Plugin" -> "Linux": bpf.xsk_setup_xdp_prog(if_index)
"Linux" --> "Device Plugin": XSK file descriptor

"Device Plugin" -> "Device Plugin": create UDS
"Device Plugin" -> "UDS server" ** : create & start UDS server
"Device Plugin" -> "UDS server" : ifname, XSK FD, UDS filepath
hnote over "UDS server"
	  listen for
	  connection
endhnote
"Device Plugin" --> "Kubelet": GRPC: AllocateResponse(UDS mount path, pod env vars)

deactivate "Device Plugin"

autonumber stop
"Kubelet" -[#deepSkyBlue]>> "Pod/CNDP App" : <color:deepSkyBlue>Kubelet starts creating the pod around now
autonumber resume

"Kubelet" -> "CNI" : cmdAdd(ifname, namespace, config)
activate "CNI"
"CNI" -> "CNI" : cni.IPAM
"CNI" -> "Linux" : ethtool.filters(ifname)
"Linux" --> "CNI" : return 0
"CNI" -> "Pod/CNDP App" : place interface in pod netns
"CNI" --> "Kubelet" : return 0
deactivate "CNI"

autonumber stop

== Pod Running ==

"Kubelet" -> "Pod/CNDP App" : start pod
"Pod/CNDP App" -> "Pod/CNDP App" : application start
activate "Pod/CNDP App"

"Pod/CNDP App" -> "UDS server": /connect,hostname
note right
	CNDP application starts
	UDS handshake begins
end note
activate "UDS server"
"UDS server" -> "Kubelet": podresources.ListPodResourcesRequest()
"Kubelet" --> "UDS server": podresources.ListPodResourcesResponse()
"UDS server" -> "UDS server": validate hostname against ifname
"UDS server" --> "Pod/CNDP App": /host_ok
deactivate "UDS server"
hnote over "UDS server"
	listen for
	request
end note

"Pod/CNDP App" -> "UDS server": /xsk_map_fd,ifname
activate "UDS server"
"UDS server" -> "UDS server": FD for ifname
"UDS server" --> "Pod/CNDP App": /fd_ack,FD
deactivate "UDS server"

hnote over "UDS server"
	listen for
	request
endhnote

"Pod/CNDP App" -> "UDS server": /config_busy_poll,FD
activate "UDS server"
"UDS server" -> "Linux" : bpf.setsockopt(FD, BUSY_POLL)
"Linux" --> "UDS server" : return 0
"UDS server" --> "Pod/CNDP App": /config_busy_poll_ack
deactivate "UDS server"

hnote over "UDS server"
	listen for
	request
endhnote

"Pod/CNDP App" -> "UDS server": /fin
activate "UDS server"
"UDS server" --> "Pod/CNDP App": /fin_ack
"Pod/CNDP App" [hidden]-> "Pod/CNDP App"
note left: UDS handshake ends
deactivate "UDS server"

destroy "UDS server"

== Pod Deletion ==
autonumber

"User" -> "Kubelet": delete pod
"Kubelet" -> "Pod/CNDP App" : stop pod
deactivate "Pod/CNDP App"
"Kubelet" -> "CNI" : cmdDel(ifname, config)
activate "CNI"
"CNI" -> "Pod/CNDP App" : retrieve interface from pod netns
autonumber stop
"Pod/CNDP App" -->> "CNI" : //interface//
autonumber resume
"CNI" -> "CNI" : clear IPAM
"CNI" -> "Linux": net.if_nametoindex(ifname)
"Linux" --> "CNI": if_index
"CNI" -> "Linux": bpf.set_link_xdp_fd(if_index, -1)
"Linux" --> "CNI": return 0
"CNI" -> "Linux" : ethtool.clearFilters(ifname)
"Linux" --> "CNI" : return 0
"CNI" --> "Kubelet": return 0
deactivate "CNI"

"Kubelet" -> "Pod/CNDP App" : delete pod
deactivate "Pod/CNDP App"


@enduml
```
