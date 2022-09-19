# End-to-End Test

Bash script to run an e2e test of device plugin and CNI.

## Assumptions 
- Kubernetes installed and running
- Multus installed
- Availability of at least 2 AF_XDP capable netdevs on the node
## Run Test
- Run `make build` in the root directory of the repo.
- Navigate to this directory (`e2e/`). E2e script must be run from here.
- Modify `config.json` so there is a pool named `e2e` that has at least 2 compatible netdevs.
	- The e2e pool can configured to pick up a driver type or can be given a list of netdev names:  
		- Driver type: `"drivers" : ["i40e"]`
		- Netdev list: `"devices" : ["ens785f0", "ens785f1"]`
- Run the test script `./e2e-test.sh`
## What Happens
- The CNI binary is moved to /opt/cni/bin. The binary is uniquely named for this test so it will not impact an existing install.
- A network attachment definition is created. The NAD is uniquely named for this test so it will not impact an existing NAD.
- The udsTest app is built into a Docker image. The image is uniquely tagged for this test so it will not overwrite existing images.
- The device plugin is started and netdevs (determined by `config.json`) are advertised to Kubernetes.
- A pod is created with a single container running the Docker image.
- `ip a` is run in the pod to show attached interfaces.
- `ip l` is run in the pod to show attached interfaces and the XDP program ID of the AF_XDP netdev.
- Pod environment variables are printed, showing the CNDP_DEVICES variable.
- The udsTest app is run within the pod. This app tests the handshake over the UDS.
	- Cycles through and tests the full UDS handshake protocol.
	- Some bad requests are sent to generate expected errors.
	- All requests and responses are printed to screen.
- Everything above is cleaned up at the end of each run.
## Test Output
A successful test will show three netdevs in the pod:
- lo - the loopback interface
- eth0 - the default network interface
- The CNDP netdev, in this example named ens801f2

```
***** Netdevs attached to pod (ip a) *****

1: lo: <LOOPBACK,UP,LOWER_UP> mtu 65536 qdisc noqueue state UNKNOWN group default qlen 1000
    link/loopback 00:00:00:00:00:00 brd 00:00:00:00:00:00
    inet 127.0.0.1/8 scope host lo
       valid_lft forever preferred_lft forever
3: eth0@if353: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1450 qdisc noqueue state UP group default
    link/ether fe:fd:aa:ec:47:b4 brd ff:ff:ff:ff:ff:ff link-netnsid 0
    inet 10.244.0.134/24 brd 10.244.0.255 scope global eth0
       valid_lft forever preferred_lft forever
5: ens801f2: <NO-CARRIER,BROADCAST,MULTICAST,UP> mtu 1500 xdp/id:800 qdisc mq state DOWN group default qlen 1000
    link/ether 68:05:ca:2d:e9:92 brd ff:ff:ff:ff:ff:ff
    inet 192.168.1.207/24 brd 192.168.1.255 scope global ens801f2
       valid_lft forever preferred_lft forever

***** Netdevs attached to pod (ip l) *****

1: lo: <LOOPBACK,UP,LOWER_UP> mtu 65536 qdisc noqueue state UNKNOWN mode DEFAULT group default qlen 1000
    link/loopback 00:00:00:00:00:00 brd 00:00:00:00:00:00
3: eth0@if353: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1450 qdisc noqueue state UP mode DEFAULT group default
    link/ether fe:fd:aa:ec:47:b4 brd ff:ff:ff:ff:ff:ff link-netnsid 0
5: ens801f2: <NO-CARRIER,BROADCAST,MULTICAST,UP> mtu 1500 xdp qdisc mq state DOWN mode DEFAULT group default qlen 1000
    link/ether 68:05:ca:2d:e9:92 brd ff:ff:ff:ff:ff:ff
    prog/xdp id 800
```

A successful test will also show the environment variables of the pod. Note that the CNDP_DEVICES variables lists our netdev ens801f2 in this example.

```
***** Pod Env Vars *****

PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
HOSTNAME=afxdp-e2e-test
CNDP_DEVICES=ens801f2
KUBERNETES_SERVICE_PORT_HTTPS=443
KUBERNETES_PORT=tcp://10.96.0.1:443
KUBERNETES_PORT_443_TCP=tcp://10.96.0.1:443
KUBERNETES_PORT_443_TCP_PROTO=tcp
KUBERNETES_PORT_443_TCP_PORT=443
KUBERNETES_PORT_443_TCP_ADDR=10.96.0.1
KUBERNETES_SERVICE_HOST=10.96.0.1
KUBERNETES_SERVICE_PORT=443
HOME=/root

```

A successful test will also show the output from the udsTest app as it tests the handshake over the UDS. Some responses containing errors are expected here as the test app intentionally sends bad requests.

```
***** UDS Test *****

2021-08-27T11:53:46Z [info] udsserver.go:168 New connection accepted. Waiting for requests.
2021-08-27T11:53:46Z [debug] uds.go:187 Request: /connect, afxdp-e2e-test
2021-08-27T11:53:46Z [debug] uds.go:218 Request contains no file descriptor
2021-08-27T11:53:46Z [info] udsserver.go:248 Pod unvalidated - Request: /connect, afxdp-e2e-test
2021-08-27T11:53:46Z [debug] udsserver.go:339 Pod afxdp-e2e-test - Validating pod hostname
2021-08-27T11:53:46Z [debug] resources_api.go:77 Opening Pod Resource API connection
2021-08-27T11:53:46Z [debug] resources_api.go:92 Requesting pod resource list
2021-08-27T11:53:46Z [debug] resources_api.go:88 Closing Pod Resource API connection
2021-08-27T11:53:46Z [debug] udsserver.go:343 Pod afxdp-e2e-test - Found on node
2021-08-27T11:53:46Z [info] udsserver.go:373 Pod afxdp-e2e-test - valid for this UDS connection
2021-08-27T11:53:46Z [info] udsserver.go:253 Pod afxdp-e2e-test - Response: /host_ok
2021-08-27T11:53:46Z [debug] uds.go:238 Response: /host_ok

Request: /connect, afxdp-e2e-test
Response: /host_ok

2021-08-27T11:53:48Z [debug] uds.go:187 Request: /version
2021-08-27T11:53:48Z [debug] uds.go:218 Request contains no file descriptor
2021-08-27T11:53:48Z [info] udsserver.go:248 Pod afxdp-e2e-test - Request: /version
2021-08-27T11:53:48Z [info] udsserver.go:253 Pod afxdp-e2e-test - Response: 0.1
2021-08-27T11:53:48Z [debug] uds.go:238 Response: 0.1

Request: /version
Response: 0.1

2021-08-27T11:53:50Z [debug] uds.go:187 Request: /xsk_map_fd, ens801f2
2021-08-27T11:53:50Z [debug] uds.go:218 Request contains no file descriptor
2021-08-27T11:53:50Z [info] udsserver.go:248 Pod afxdp-e2e-test - Request: /xsk_map_fd, ens801f2
2021-08-27T11:53:50Z [debug] udsserver.go:280 Pod afxdp-e2e-test - Device ens801f2 recognised
2021-08-27T11:53:50Z [info] udsserver.go:261 Pod afxdp-e2e-test - Response: /fd_ack, FD: 7
2021-08-27T11:53:50Z [debug] uds.go:231 Response: /fd_ack, FD: 7

Request: /xsk_map_fd, ens801f2
Response: /fd_ack
File Descriptor: 6

2021-08-27T11:53:52Z [debug] uds.go:187 Request: /xsk_map_fd, bad-device
2021-08-27T11:53:52Z [debug] uds.go:218 Request contains no file descriptor
2021-08-27T11:53:52Z [info] udsserver.go:248 Pod afxdp-e2e-test - Request: /xsk_map_fd, bad-device
2021-08-27T11:53:52Z [warning] udsserver.go:285 Pod afxdp-e2e-test - Device bad-device not recognised
2021-08-27T11:53:52Z [info] udsserver.go:253 Pod afxdp-e2e-test - Response: /fd_nak
2021-08-27T11:53:52Z [debug] uds.go:238 Response: /fd_nak

Request: /xsk_map_fd, bad-device
Response: /fd_nak
File Descriptor: NA

2021-08-27T11:53:54Z [debug] uds.go:187 Request: /bad-request
2021-08-27T11:53:54Z [debug] uds.go:218 Request contains no file descriptor
2021-08-27T11:53:54Z [info] udsserver.go:248 Pod afxdp-e2e-test - Request: /bad-request
2021-08-27T11:53:54Z [info] udsserver.go:253 Pod afxdp-e2e-test - Response: /nak
2021-08-27T11:53:54Z [debug] uds.go:238 Response: /nak

Request: /bad-request
Response: /nak

2021-08-27T11:53:56Z [debug] uds.go:187 Request: /fin
2021-08-27T11:53:56Z [debug] uds.go:218 Request contains no file descriptor
2021-08-27T11:53:56Z [info] udsserver.go:248 Pod afxdp-e2e-test - Request: /fin
2021-08-27T11:53:56Z [info] udsserver.go:253 Pod afxdp-e2e-test - Response: /fin_ack
2021-08-27T11:53:56Z [debug] uds.go:238 Response: /fin_ack
2021-08-27T11:53:56Z [debug] uds.go:151 Closing connection
2021-08-27T11:53:56Z [debug] uds.go:153 Closing socket file
2021-08-27T11:53:56Z [debug] uds.go:114 Closing Unix listener

Request: /fin
Response: /fin_ack

```

### Timeout
The device plugin includes a timeout action for the unix domain sockets(UDS).
Once the timeout is invoked, the UDS is closed and disconnected.

The timeout can be set to a minimum of 30 seconds and a maximum of 300 seconds. If no timeout is configured, the plugin will default to the minimum 30.

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


## Extended Test
The e2e test script can also do an extended run. In addition to the single container single device test, the script will go on to create:
- A pod with a single container requesting 2 devices
- A pod with 2 containers, each requesting a single device
- Timeout before the UDS connection
- Timeout after the UDS connection

To do the full extended run, add the flag -f or --full when calling the script:
`./e2e-test.sh --full`
