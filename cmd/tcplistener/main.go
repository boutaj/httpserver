package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
)

func getLinesChannel(connection io.ReadCloser) <-chan string {
	read := make(chan string)

	go func ()  {

		defer close(read)
		defer connection.Close()

		currentLine := ""
		for {
			buffer := make([]byte, 8);
			n, err := connection.Read(buffer)
			data := buffer[:n]

			if err == io.EOF {
				break
			}

			if i := bytes.IndexByte(data, '\n'); i != -1 {
				currentLine += string(data[:i])
				data = data[i+1:]
				read <- currentLine
				currentLine = ""
			}
			currentLine += string(data)
		}

	}()

	return read
}

func main() {

	listener, _ := net.Listen("tcp", ":42069")

	for {
		connection, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}

		for out := range getLinesChannel(connection) {
			fmt.Println(out)
		}
	}
}