package main

import (
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const udsProtocol = "unixpacket"

func main() {
	devices := os.Args[1:]

	c, err := net.Dial(udsProtocol, "/tmp/cndp.sock")
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	makeRequest("/connect, cndp-e2e-test", c)
	time.Sleep(2 * time.Second)

	makeRequest("/version", c)
	time.Sleep(2 * time.Second)

	for _, dev := range devices {
		dev = strings.Replace(dev, "\"", "", -1)
		devReq := "/xsk_map_fd, " + dev
		makeRequestFD(devReq, c)
		time.Sleep(2 * time.Second)
	}

	makeRequestFD("/xsk_map_fd, bad-device", c)
	time.Sleep(2 * time.Second)

	makeRequest("/bad-request", c)
	time.Sleep(2 * time.Second)

	makeRequest("/fin", c)
	time.Sleep(2 * time.Second)
}

func makeRequest(request string, c net.Conn) {

	buf := make([]byte, 1024)

	_, err := c.Write([]byte(request))
	if err != nil {
		println("Write error: ", err)
	}

	n, err := c.Read(buf[:])
	if err != nil {
		println("Read error: ", err)
	}

	println()
	println("Request: " + request)
	println("Response:", string(buf[0:n]))
	println()
}

func makeRequestFD(request string, c net.Conn) {

	_, err := c.Write([]byte(request))
	if err != nil {
		println("Write error: ", err)
	}

	conn := c.(*net.UnixConn)

	fd, response, err := getFD(conn)
	if err != nil {
		log.Fatal(err)
	}

	println()
	println("Request: " + request)
	println("Response:", response)
	if fd > 0 {
		println("File Descriptor:", strconv.Itoa(fd))
	} else {
		println("File Descriptor: NA")
	}
	println()
}

func getFD(conn *net.UnixConn) (int, string, error) {

	// get the underlying socket
	socketFile, err := conn.File()
	if err != nil {
		println("Socket error: ", err)
		return 0, "", err
	}
	defer socketFile.Close()
	socketFD := int(socketFile.Fd())

	// recvmsg
	rights := make([]byte, syscall.CmsgSpace(4))
	msgBuf := make([]byte, 128)

	n, _, _, _, err := syscall.Recvmsg(socketFD, msgBuf, rights, 0)
	if err != nil {
		println("Recvmsg error: ", err)
		return 0, "", err
	}

	msg := string(msgBuf[0:n])

	if msg == "/fd_ack" {
		// parse control msgs
		var msgs []syscall.SocketControlMessage
		msgs, err = syscall.ParseSocketControlMessage(rights)
		//should be looping through msgs and FDs
		//but I know it's a single FD, so it's msg[0] fds[0]
		fds, _ := syscall.ParseUnixRights(&msgs[0])
		fd := fds[0]
		return fd, msg, err
	} else if msg == "/fd_nak" {
		return -1, msg, err
	}

	return -1, "unexpected reply", err
}
