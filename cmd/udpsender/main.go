package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	serverAddr, err := net.ResolveUDPAddr("udp", "localhost:42069")
	if err != nil {
		log.Fatal(err)
	}

	conn, err := net.DialUDP("udp", nil, serverAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")
		line, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("error reading from stdin: %v", err)
			continue
		}

		_, err = conn.Write([]byte(line))
		if err != nil {
			log.Printf("error writing to udp connection: %v", err)
		}
	}
}
