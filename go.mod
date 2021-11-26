module github.com/intel/cndp_device_plugin

go 1.13

require (
	github.com/containernetworking/cni v1.0.1
	github.com/containernetworking/plugins v1.0.1
	github.com/fsnotify/fsnotify v1.5.1 // indirect
	github.com/go-cmd/cmd v1.3.0
	github.com/go-ozzo/ozzo-validation/v4 v4.3.0
	github.com/google/go-cmp v0.5.6 // indirect
	github.com/google/uuid v1.2.0
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.7.0
	github.com/vishvananda/netlink v1.1.1-0.20210330154013-f5de75959ad5
	golang.org/x/net v0.0.0-20210813160813-60bc85c4be6d
	golang.org/x/sys v0.0.0-20210915083310-ed5796bab164 // indirect
	golang.org/x/text v0.3.7 // indirect
	google.golang.org/grpc v1.40.0
	google.golang.org/protobuf v1.27.1 // indirect
	gotest.tools v2.2.0+incompatible
	k8s.io/kubelet v0.22.1
)
