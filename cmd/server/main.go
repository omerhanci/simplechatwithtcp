package main

import (
	"fmt"
	"net"

	"github.com/Applifier/golang-backend-assignment/internal/server"
)

func main() {
	fmt.Println("Hello from server!")
	server := server.New()
	tcpAddress, _ := net.ResolveTCPAddr("tcp", ":2525")
	server.Start(tcpAddress)
}
