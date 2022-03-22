module github.com/intel/afxdp-plugins-for-kubernetes

go 1.13

require (
	github.com/containernetworking/cni v1.0.1
	github.com/containernetworking/plugins v1.0.1
	github.com/fsnotify/fsnotify v1.5.1 // indirect
	github.com/go-ozzo/ozzo-validation/v4 v4.3.0
	github.com/google/go-cmp v0.5.6 // indirect
	github.com/google/uuid v1.3.0
	github.com/google/gofuzz v1.1.0
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.0
	github.com/vishvananda/netlink v1.1.1-0.20210330154013-f5de75959ad5
	golang.org/x/net v0.0.0-20220225172249-27dd8689420f
	google.golang.org/grpc v1.45.0
	gotest.tools v2.2.0+incompatible
	k8s.io/apimachinery v0.23.4
	k8s.io/kubelet v0.23.4

)
