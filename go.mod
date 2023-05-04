module github.com/intel/afxdp-plugins-for-kubernetes

go 1.13

require (
	github.com/containernetworking/cni v1.1.2
	github.com/containernetworking/plugins v1.1.1
	github.com/go-ozzo/ozzo-validation/v4 v4.3.0
	github.com/golang/protobuf v1.5.3
	github.com/google/gofuzz v1.1.0
	github.com/google/uuid v1.3.0
	github.com/intel/afxdp-plugins-for-kubernetes/pkg/subfunctions v0.0.0
	github.com/moby/sys/mount v0.3.3
	github.com/pkg/errors v0.9.1
	github.com/safchain/ethtool v0.0.0-20210803160452-9aa261dae9b1
	github.com/sirupsen/logrus v1.9.0
	github.com/stretchr/testify v1.7.0
	github.com/vishvananda/netlink v1.1.1-0.20210330154013-f5de75959ad5
	golang.org/x/net v0.7.0
	google.golang.org/grpc v1.49.0
	google.golang.org/protobuf v1.30.0 // indirect
	gotest.tools v2.2.0+incompatible
	k8s.io/apimachinery v0.25.2
	k8s.io/kubelet v0.25.2
)

replace github.com/intel/afxdp-plugins-for-kubernetes/pkg/subfunctions => ./pkg/subfunctions
