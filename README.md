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

## CLOC
Output from CLOC (count lines of code) - github.com/AlDanial/cloc 
<!---clocstart--->
```
-------------------------------------------------------------------------------
Language                     files          blank        comment           code
-------------------------------------------------------------------------------
Go                              18            411            570           3689
Markdown                         3             52              0            196
Bourne Shell                     2             21              3            194
YAML                             5              3              0            100
C                                2             21             26             91
make                             1             10              4             38
C/C++ Header                     2             14             28             24
JSON                             3              0              0             24
Dockerfile                       1              0              0              3
-------------------------------------------------------------------------------
SUM:                            37            532            631           4359
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
