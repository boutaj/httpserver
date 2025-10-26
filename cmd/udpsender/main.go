package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
)

func main()  {
	resolve, _ := net.ResolveUDPAddr("udp", "localhost:42069")

	connection, _ := net.DialUDP("udp", nil, resolve)
	defer connection.Close()

	reader := bufio.NewReader(os.Stdin) 
	for {
		fmt.Print("> ")
		line, err := reader.ReadString('\n')
		if len(line) == 0 && err == io.EOF {
            break
        }
		
		connection.Write([]byte(line))
	}
}