package networking

import (
	"github.com/go-cmd/cmd"
	"github.com/intel/cndp_device_plugin/pkg/logging"
	"net"
	"strings"
)

/*
Handler is the CNI and device plugins interface to the host networking.
The interface exists for testing purposes, allowing unit tests to test
against a fake API.
*/
type Handler interface {
	GetHostDevices() ([]net.Interface, error)
	GetDeviceDriver(interfaceName string) (string, error)
}

/*
handler implements the Handler interface.
*/
type handler struct{}

/*
NewHandler returns an implementation of the Handler interface.
*/
func NewHandler() Handler {
	return &handler{}
}

/*
GetHostDevices returns a list of net.Interface, representing the devices
on the host.
*/
func (r *handler) GetHostDevices() ([]net.Interface, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		logging.Errorf("Error getting host devices: %v", err)
		return interfaces, err
	}
	return interfaces, nil
}

/*
GetDriverName takes a netdave name and returns the driver type
It executes the command: ethtool -i <interface_name>
*/
func (r *handler) GetDeviceDriver(interfaceName string) (string, error) {
	// build the command
	cmd := cmd.NewCmd("ethtool", "-i", interfaceName)

	// run and wait for cmd to return status
	status := <-cmd.Start()

	// take first line of Stdout - Stdout[0]
	// split the line on colon - ":"
	// driver is 2nd half of split string - [1]
	driver := strings.Split(status.Stdout[0], ":")[1]

	// trim whitespace and return
	return strings.TrimSpace(driver), nil
}

/*
RXFlow function is a wrapper function to exeute ethtool command:
ethtool -N <interface_name> rx-flow-hash udp4
*/
//func RXFlow(interfaceName string) error {
//	cmd := "ethtool"
//	args := []string{"-N", interfaceName, "rx-flow-hash", "udp4"}
//	_, err := execCommand(cmd, args)
//	logging.Errorf("RXFlow(): failed to execute command: %v", err)
//	return err
//}

/*
FlowType function is a wrapper function to exeute ethtool command:
ethtool -N <name_of_interface> flow-type src-port <src_port_id> dst-port <dst_port_id> action <queue_id>
*/
//func FlowType(interfaceName string, srcPortID int, dstPortID int, QueueID int) error {
//	cmd := "ethtool"
//	args := []string{"-N", interfaceName, "flow-type", "src-port", string(srcPortID), "dst-port", string(dstPortID), "action", string(QueueID)}
//	_, err := execCommand(cmd, args)
//	logging.Errorf("FlowType(): failed to execute command: %v", err)
//	return err
//}
