package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

func main() {
	c, err := net.Dial("unix", "/tmp/cndp.sock")
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()
	
	reader := bufio.NewReader(os.Stdin)
	buf := make([]byte, 1024)

	for {		
		fmt.Print("-> ")
		text, _ := reader.ReadString('\n')
		text = strings.Replace(text, "\n", "", -1)
		
		_, err = c.Write([]byte(text))
		if err != nil {
			println("Write error: ", err)
		}
		
		if strings.Compare("exit", text) == 0 {
			c.Close()
			break
		}
		
		n, err := c.Read(buf[:])
		if err != nil {
			println("Read error: ", err)
		}
		println("Received:", string(buf[0:n]))
		
	}

}
