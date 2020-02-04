package main

import (
	"fmt"
	"net"
	"time"

	"github.com/Applifier/golang-backend-assignment/internal/client"
)

func main() {
	fmt.Println("Hello from client!")

	client := client.New()
	tcpAddress, _ := net.ResolveTCPAddr("tcp", ":2525")
	go client.Connect(tcpAddress)
	time.Sleep(5 * time.Second)
	client.WhoAmI()
}
