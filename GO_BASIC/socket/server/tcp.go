package server

import (
	transport "go_basic/socket"
	"log"
	"net"
	"time"
)

func TcpServer() {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", "127.0.0.1:5678")
	transport.CheckError(err)
	listener, err := net.ListenTCP("tcp4", tcpAddr)
	transport.CheckError(err)
	log.Println("waiting for client connection")
	conn, err := listener.Accept()
	transport.CheckError(err)
	log.Printf("establish connection to client %s \n", conn.RemoteAddr().String())
	conn.SetReadDeadline(time.Now().Add(time.Second * 5))
	defer conn.Close()

	request := make([]byte, 1024)
	n, err := conn.Read(request)
	transport.CheckError(err)
	log.Printf("recevier %s\n", string(request[:n]))
}

func TcpLongConnection() {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", "127.0.0.1:5678")
	transport.CheckError(err)
	listener, err := net.ListenTCP("tcp4", tcpAddr)
	transport.CheckError(err)
	log.Println("waiting for client connection")
	conn, err := listener.Accept()
	transport.CheckError(err)
	log.Printf("establish connection to client %s \n", conn.RemoteAddr().String())
	conn.SetReadDeadline(time.Now().Add(time.Second * 5))
	defer conn.Close()

	request := make([]byte, 1024)
	for {
		n, err := conn.Read(request)
		transport.CheckError(err)
		log.Printf("recevier %s\n", string(request[:n]))
		conn.SetReadDeadline(time.Now().Add(time.Second * 5))
	}
}
