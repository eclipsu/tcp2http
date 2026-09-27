package main

import (
	"fmt"
	"log"
	"net"

	"github.com/eclipsu/tcp2http/internal/request"
)

func main() {
	listener, err := net.Listen("tcp", ":42069")
	if err != nil {
		log.Fatal(err)
	}

	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("Connection accepted")
		req, err := request.RequestFromReader(conn)

		if err != nil {
			fmt.Printf("Something went wrong: %v\n", err)
			conn.Close()
			continue
		}

		fmt.Println("Request line:")
		fmt.Printf("- Method: %s\n", req.RequestLine.Method)
		fmt.Printf("- Target: %s\n", req.RequestLine.RequestTarget)
		fmt.Printf("- Version: %s\n", req.RequestLine.HttpVersion)

		fmt.Println("Headers:")
		for key, value := range req.Headers {
			fmt.Printf("- %s: %s\n", key, value)
		}

		fmt.Println("Body:")
		fmt.Printf("%s\n", req.Body)

		fmt.Println("Connection closed\n")
		conn.Close()
	}
}
