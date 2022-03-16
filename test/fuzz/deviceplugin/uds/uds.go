package uds

import (
	"github.com/intel/afxdp_k8s_plugins/internal/logformats"
	"github.com/intel/afxdp_k8s_plugins/internal/uds"
	logging "github.com/sirupsen/logrus"
	"os"
	"time"
)

const (
	udsMsgBufSize  = 64
	udsCtlBufSize  = 4
	udsProtocol    = "unixpacket" // "unix"=SOCK_STREAM, "unixdomain"=SOCK_DGRAM, "unixpacket"=SOCK_SEQPACKET
	udsIdleTimeout = 10 * time.Second
	interesting    = 1
	uninteresting  = 0
	discard        = -1

	logLevel       = "error"
	udsDirFileMode = os.FileMode(0700) // drwx------
)

var ch = make(chan string)

/*
Fuzz seeds fuzzed data to UDS write function.
The input data is considered:
 - uninteresting if is caught by an existing error
 - interesting if it does not result in an error, input priority increases for subsequent fuzzing
*/
func Fuzz(data []byte) int {
	if len(data) == 0 {
		return discard
	}

	fp, _ := os.OpenFile("./fuzz_"+logLevel+".log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	logging.SetOutput(fp)
	level, _ := logging.ParseLevel(logLevel)
	logging.SetLevel(level)
	logging.SetFormatter(logformats.Fuzz)

	udsPath, _ := uds.GenerateRandomSocketName("/tmp/afxdp/", udsDirFileMode)
	go reader(udsPath, data)
	time.Sleep(10 * time.Millisecond)

	uds := uds.NewHandler()
	err := uds.Init(udsPath, udsProtocol, udsMsgBufSize, udsCtlBufSize, udsIdleTimeout)
	if err != nil {
		logging.Errorf("Error Initialising UDS: %v", err)
	}
	cleanup, _ := uds.Dial()
	defer cleanup()

	err = uds.Write(string(data), -1)
	if err != nil {
		logging.Errorf("Connection write error: %v", err)
	}

	returned := <-ch
	logging.Infof("Wrote: %s", string(data))
	logging.Infof("Read: %s", returned)
	if returned != string(data) {
		return interesting
	}

	return uninteresting

}

func reader(udsPath string, data []byte) {
	uds := uds.NewHandler()
	err := uds.Init(udsPath, udsProtocol, udsMsgBufSize, udsCtlBufSize, udsIdleTimeout)
	if err != nil {
		logging.Errorf("Error Initialising UDS: %v", err)
	}
	cleanup, _ := uds.Listen()
	defer cleanup()
	msg, _, err := uds.Read()
	if err != nil {
		logging.Errorf("Data at time of error: %s", string(data))
	}
	ch <- msg
}
