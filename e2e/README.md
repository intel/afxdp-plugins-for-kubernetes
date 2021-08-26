# End-to-End Test

Bash script to run an e2e test of deceive plugin and CNI.

## Assumptions 
- Kubernetes installed and running
- Multus installed
- Availability of at least 2 AF_XDP capable netdevs
## Run Test
- Run `build.sh` in the root directory of the repo.
- Navigate to this directory (`examples/e2e-test/`). Script must be run from here.
- Place a `config.json` file here that includes the name of the netdevs that you would like to make available to the test pod.
	- A sample config can be found in `examples/sample-config/`
	- For this test there must be a pool named `e2e` with at least 2 real netdevs.
- Run the test script `./e2e-test.sh`
## What Happens
- The CNI binary is moved to /opt/cni/bin. The binary is uniquely named for this test so it will not impact an existing install.
- A network attachment definition is created. The NAD is uniquely named for this test so it will not impact an existing NAD.
- A sample app is built. This sample app tests the handshake over the UDS.
- Sample app is built into a Docker image. The image is uniquely tagged so it will not overwrite existing images.
- The device plugin is started and devices listed in `config.json` are advertised to Kubernetes.
- A pod is created, a single container running the Docker image. The pod is uniquely named for this test.
- `ip a` is run in the pod to show attached interfaces.
- The sample app in the pod is run. The app makes requests via the UDS.
	- Cycles through and tests the full UDS handshake protocol.
	- Some bad requests are sent to generate expected errors.
	- All requests and responses are printed to screen.
- Everything above is cleaned up at the end of each run.
## Test Output
A successful test will show three netdevs in the pod:
- lo - the loopback interface
- eth0 - the default network interface
- The CNDP interface, in this example named ens785f0
 
```
***** Netdevs attached to pod *****

1: lo: <LOOPBACK,UP,LOWER_UP> mtu 65536 qdisc noqueue state UNKNOWN group default qlen 1000
    link/loopback 00:00:00:00:00:00 brd 00:00:00:00:00:00
    inet 127.0.0.1/8 scope host lo
       valid_lft forever preferred_lft forever
3: eth0@if430: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1450 qdisc noqueue state UP group default
    link/ether a6:76:2a:39:48:98 brd ff:ff:ff:ff:ff:ff link-netnsid 0
    inet 10.244.0.146/24 brd 10.244.0.255 scope global eth0
       valid_lft forever preferred_lft forever
13: ens785f0: <NO-CARRIER,BROADCAST,MULTICAST,UP> mtu 1500 xdp/id:1234 qdisc mq state DOWN group default qlen 1000
    link/ether 98:03:9b:6a:b4:ef brd ff:ff:ff:ff:ff:ff
    inet 192.168.1.202/24 brd 192.168.1.255 scope global ens785f0
       valid_lft forever preferred_lft forever

```
A successful test will also show the output from the test app as it runs the handshake protocol over the UDS. Some responses containing errors are expected here as the app intentionally sends bad requests.

It should also be noted that the test app will request the file descriptor for every device known to the device plugin. Every request will result in an error response, except for requests associated with devices that are actually assigned to the current pod.
```
***** UDS Test *****

I0422 13:13:48.988962  317356 cndp.go:141] Client connected
I0422 13:13:48.989044  317356 cndp.go:153] Request: /connect
I0422 13:13:48.989062  317356 cndp.go:199] Response: /host

Request: /connect
Response: /host

I0422 13:13:50.989440  317356 cndp.go:153] Request: "hostname": "cndp-e2e-test"
I0422 13:13:50.989470  317356 cndp.go:217] Validating pod hostname: cndp-e2e-test
I0422 13:13:50.990405  317356 cndp.go:231] cndp-e2e-test found on node
I0422 13:13:50.990437  317356 cndp.go:261] cndp-e2e-test is valid for this UDS connection
I0422 13:13:50.990451  317356 cndp.go:199] Response: /host_ok

Request: "hostname": "cndp-e2e-test"
Response: /host_ok

I0422 13:13:52.990889  317356 cndp.go:153] Request: "hostname": "bad-hostname"
I0422 13:13:52.990916  317356 cndp.go:217] Validating pod hostname: bad-hostname
I0422 13:13:52.991603  317356 cndp.go:233] bad-hostname not found on node
I0422 13:13:52.991632  317356 cndp.go:199] Response: Error: Invalid Host

Request: "hostname": "bad-hostname"
Response: Error: Invalid Host

I0422 13:13:54.992114  317356 cndp.go:153] Request: /xsk_map_ens785f0
I0422 13:13:54.992156  317356 cndp.go:199] Response: 6

Request: /xsk_map_ens785f0
Response: 6

I0422 13:13:56.992557  317356 cndp.go:153] Request: /xsk_map_ens785f1
I0422 13:13:56.992600  317356 cndp.go:199] Response: Error: Unknown Interface ens785f1

Request: /xsk_map_ens785f1
Response: Error: Unknown Interface ens785f1

I0422 13:13:58.993017  317356 cndp.go:153] Request: /fin
I0422 13:13:58.993056  317356 cndp.go:199] Response: /fin_ack
I0422 13:13:58.993081  317356 cndp.go:211] Closing connection

Request: /fin
Response: /fin_ack

```

## Extended Test
The e2e test script can also do an extended run. In addition to the single container single device test, the script will go on to create:
- A pod with a single container requesting 2 devices
- A pod with 2 containers, each requesting a single device

To do the full extended run, add the flag -f or --full when calling the script:
`./e2e-test.sh --full`

## Manual  Test
It can be useful to run the test manually. For instance, allowing the user time to log into the pod, check the environment and debug. Start the following steps from the root directory of the repo.
```
# build the plugins
./build.sh

# move to the e2e test dir
cd examples/e2e-test/

# install the cni plugin
cp ../../bin/cndp /opt/cni/bin/cndp-e2e

# create the NAD
kubectl create -f ./nad.yaml

# build the sample apps
go build -o uds-client-auto ./autoTest/main.go
go build -o uds-client-manual ./manualTest/main.go

# build the docker image (these 8 lines are a single command)
docker build \
--build-arg http_proxy=${http_proxy} \
--build-arg HTTP_PROXY=${HTTP_PROXY} \
--build-arg https_proxy=${https_proxy} \
--build-arg HTTPS_PROXY=${HTTPS_PROXY} \
--build-arg no_proxy=${no_proxy} \
--build-arg NO_PROXY=${NO_PROXY} \
-t cndp-e2e-test -f Dockerfile .

# start the device plugin
./../../bin/cndp-dp
```
At this point the device plugin will consume your terminal. Open a new terminal and navigate to this directory (`examples/e2e-test/`) to continue.

```
# create the pod
kubectl create -f pod-1c1d.yaml

# open a terminal into the pod
kubectl exec -it cndp-e2e-test -- bash

# sample apps can be found in /cndp
cd /cndp/
```
There are 2 sample apps in `/cndp`:
 - uds-client-auto
 - uds-client-manual

The auto app will run through the full UDS handshake protocol. This is the app used in the automated e2e tests.
The manual app gives you a prompt and allows you to manually write messages to the UDS. Type `exit` to exit out of this app.

**Note:** The device plugin is designed to give a single session over the UDS. After this it closes the connection and stops listening. This means the test apps can only be run once in a pods lifecycle. Attempting to run the app a second time will fail.

Example usage of the manual app:
```
root@cndp-e2e-test:/cndp# ./uds-client-manual

-> /connect
Received: /host

-> "hostname": "cndp-e2e-test"
Received: /host_ok

-> "hostname": "bad-host"
Received: Error: Invalid Host

-> /xsk_map_ens785f0
Received: 6

-> /xsk_map_bad-device
Received: Error: Unknown Interface bad-device

-> foobar
Received: Error: Bad Request

-> /foobar
Received: Error: Bad Request

-> /fin
Received: /fin_ack

-> exit

root@cndp-e2e-test:/cndp#
```
