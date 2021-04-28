package main

import (
	"log"
	"net"
	"os"
	"strings"
	"time"
)

func main() {
	devices := os.Args[1:]

	c, err := net.Dial("unix", "/tmp/cndp.sock")
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	makeRequest("/connect", c)
	time.Sleep(2 * time.Second)

	makeRequest("\"hostname\": \"cndp-e2e-test\"", c)
	time.Sleep(2 * time.Second)

	makeRequest("\"hostname\": \"bad-hostname\"", c)
	time.Sleep(2 * time.Second)

	for _, dev := range devices {
		dev = strings.Replace(dev, "\"", "", -1)
		devReq := "/xsk_map_" + dev
		makeRequest(devReq, c)
		time.Sleep(2 * time.Second)
	}

	makeRequest("/xsk_map_bad-device", c)
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
