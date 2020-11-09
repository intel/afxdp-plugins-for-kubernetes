package cndp

import (
	"github.com/golang/glog"
	"github.com/nu7hatch/gouuid"
	"net"
	"strings"
)

type CndpInterface interface {
	StartSocketServer(SockAddr string)
	CreateUdsSocket() string
}

type CNDP struct {
	CndpInterface
}

//TODO currently rough sample code, update this to provide the FD to xdpsock, proper error and socket handeling
//TODO later update with protocol for interacting with cndp
func (c *CNDP) StartSocketServer(SockAddr string) {

	glog.Info("Listening on socket " + SockAddr)

	l, err := net.Listen("unix", SockAddr)
	if err != nil {
		glog.Fatal("listen error:", err)
	}

	// Accept new connections
	conn, err := l.Accept()
	if err != nil {
		glog.Fatal("accept error:", err)
	}

	glog.Info("Client connected")

	buf := make([]byte, 1024)
	for {
		n, err := conn.Read(buf[:])
		if err != nil {
			glog.Fatal("read error:", err)
		}

		glog.Info("Received: " + string(buf[0:n]))

		if strings.Compare("exit", string(buf[0:n])) == 0 {
			break
		}

		_, err = conn.Write([]byte("Hello from DP, you said: " + string(buf[0:n])))
		if err != nil {
			glog.Fatal("write error:", err)
		}

	}

	glog.Info("Closing connection")
	conn.Close()
	l.Close()
}

func (c *CNDP) CreateUdsSocket() string {

	sockName, err := uuid.NewV4() //TODO check if it exists

	if err != nil {
		glog.Fatal(err)
	}

	return "/tmp/" + sockName.String() + ".sock"
}

//TODO also return error?
func NewCndp() CndpInterface {
	return &CNDP{}
}
