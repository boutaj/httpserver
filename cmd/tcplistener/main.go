package main

import (
	"fmt"
	"httpfromtcp/internal/request"
	"log"
	"net"
)

func main() {

	listener, _ := net.Listen("tcp", ":42069")

	for {
		connection, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}

		request, err := request.RequestFromReader(connection);
		if err != nil {
			log.Fatal("Could not read the request properly")
		}

		fmt.Printf("Request line:\n- Method: %s\n- Target: %s\n- Version: %s", request.RequestLine.Method, request.RequestLine.RequestTarget, request.RequestLine.HttpVersion)
	}
}
