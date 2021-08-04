package ethtool

import (
	"github.com/intel/cndp_device_plugin/pkg/logging"
	"os/exec"
	"regexp"
)

/*
GetDriverName function is a wrapper function to exeute ethtool command
ethtool -i <interface_name>
*/
func GetDriverName(interfaceName string) (string, error) {
	var result string
	var regxDriver = regexp.MustCompile(`(?s)(driver: )(.*?)(?:[\n\r])`)

	cmd := "ethtool"
	arg0 := "-i"
	execCommand := exec.Command(cmd, arg0, interfaceName)
	outputBytes, _ := execCommand.Output() //returns byte array
	outputString := string(outputBytes)
	match := regxDriver.FindStringSubmatch(outputString)
	if len(match) > 0 {
		result = match[2]
	}
	return result, nil
}

/*
RXFlow function is a wrapper function to exeute ethtool command:
ethtool -N <interface_name> rx-flow-hash udp4
*/
func RXFlow(interfaceName string) error {
	cmd := "ethtool"
	args := []string{"-N", interfaceName, "rx-flow-hash", "udp4"}
	_, err := execCommand(cmd, args)
	logging.Errorf("RXFlow(): failed to execute command: %v", err)
	return err
}

/*
FlowType function is a wrapper function to exeute ethtool command:
ethtool -N <name_of_interface> flow-type src-port <src_port_id> dst-port <dst_port_id> action <queue_id>
*/
func FlowType(interfaceName string, srcPortID int, dstPortID int, QueueID int) error {
	cmd := "ethtool"
	args := []string{"-N", interfaceName, "flow-type", "src-port", string(srcPortID), "dst-port", string(dstPortID), "action", string(QueueID)}
	_, err := execCommand(cmd, args)
	logging.Errorf("FlowType(): failed to execute command: %v", err)
	return err
}

func execCommand(cmd string, args []string) ([]byte, error) {
	return exec.Command(cmd, args...).Output()
}