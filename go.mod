module github.com/intel/cndp_device_plugin

go 1.13

require (
	github.com/containernetworking/cni v0.8.0
	github.com/containernetworking/plugins v0.8.7
	github.com/fsnotify/fsnotify v1.4.9 // indirect
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/nu7hatch/gouuid v0.0.0-20131221200532-179d4d0c4d8d
	github.com/pkg/errors v0.9.1
	github.com/safchain/ethtool v0.0.0-20190326074333-42ed695e3de8
	github.com/stretchr/testify v1.6.1
	github.com/vishvananda/netlink v1.1.0
	golang.org/x/lint v0.0.0-20210508222113-6edffad5e616 // indirect
	golang.org/x/net v0.0.0-20210405180319-a5a99cb37ef4
	golang.org/x/tools v0.1.5 // indirect
	google.golang.org/grpc v1.27.1
	gotest.tools v2.2.0+incompatible
	k8s.io/kubelet v0.21.0
)
