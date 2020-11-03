# CNDP Device Plugin

A proof of concept Kubernetes device plugin to provide AF_XDP networking to Kubernetes pods using Intel's Cloud Native Data Plane framework.

## Clone and Build

```bash
git clone ssh://git@gitlab.devtools.intel.com:29418/OrchSW/CNO/containers_afxdp_network_device_plugin-.git
cd containers_afxdp_network_device_plugin-/
git checkout -b cndp_q4_2020_poc
./build.sh
```
Two binaries will be placed in ./bin directory:
- **cndp-dp** is the device plugin
- **cndp** is the CNI plugin. This needs to be placed in `/opt/cni/bin/`
