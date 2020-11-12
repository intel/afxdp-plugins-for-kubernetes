# End-to-End Test

Bash script to run an e2e test of deceive plugin and CNI.

## Assumptions 
- Kubernetes installed and running
- Multus installed
## Run Test
- cd to this directory, script must be run from here
- `./e2e-test.sh`
## What Happens
- CNI and DP are build from scratch. The build script in the root of the repo is used.
- The CNI binary is moved to /opt/cni/bin. The binary is uniquely named for this test so it will not impact an existing install.
- A network attachment definition is created. The NAD is uniquely named for this test so it will not impact an existing NAD.
- A sample app is built.
	- Currently an echo server to test the UDS.
	- This will be updated to xdpsock and later CNDP.
- Sample app is built into an Docker image. The image is uniquely tagged for this test so it will not overwrite existing images.
- The Device Plugin is started.
- A pod is created, a single container running the Docker image. The pod is uniquely named for this test.
- `ip a` is run in the pod to show attached interfaces.
- The sample app in the pod is run:
	- Currently the sample app writes a message to the DP over the UDS.
	- If all works correctly the DP will echo back the message.
	- Later we'll test xdpsock and CNDP here.
- Everything above is cleaned up at the beginning and end of each run.
## Test Output
- A successful test will show three netdevs in the pod.
```
*****************************************************
*              Netdevs attached to pod              *
*****************************************************
1: lo: <LOOPBACK,UP,LOWER_UP> mtu 65536 qdisc noqueue state UNKNOWN group default qlen 1000
    link/loopback 00:00:00:00:00:00 brd 00:00:00:00:00:00
    inet 127.0.0.1/8 scope host lo
       valid_lft forever preferred_lft forever
3: eth0@if303: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1450 qdisc noqueue state UP group default
    link/ether a2:3c:12:e2:06:a7 brd ff:ff:ff:ff:ff:ff link-netnsid 0
    inet 10.244.0.30/24 scope global eth0
       valid_lft forever preferred_lft forever
11: ens785f0: <BROADCAST,MULTICAST,PROMISC> mtu 1500 qdisc noop state DOWN group default qlen 1000
    link/ether 98:03:9b:6a:b4:ee brd ff:ff:ff:ff:ff:ff

```
- A successful test will also show the sample app worked correctly, verifying DP functionality.
	- Currently success means the DP echoes back a message sent from the sample app. This will evolve.
```
*****************************************************
*                     UDS Test                      *
*****************************************************
I1112 11:33:50.963770   86982 cndp.go:49] Client connected
I1112 11:33:50.963859   86982 cndp.go:58] Received: end to end test
I1112 11:33:50.964076   86982 cndp.go:58] Received: exit
I1112 11:33:50.964105   86982 cndp.go:71] Closing connection
-> Received: Hello from DP, you said: end to end test
```