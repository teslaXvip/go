package server

import (
	transport "go_basic/socket"
	"log"
	"net"
	"time"
)

func UdpServer() {
	udpAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:5678")
	transport.CheckError(err)

	conn, err := net.ListenUDP("udp", udpAddr)
	transport.CheckError(err)

	log.Println("return conn")
	defer conn.Close()

	request := make([]byte, 256)
	n, remoteAddr, err := conn.ReadFromUDP(request)
	transport.CheckError(err)

	log.Printf("receive request %s from %s\n", string(request[:n]), remoteAddr.String())
	// 向客户端回写数据
	conn.WriteToUDP([]byte("hello"), remoteAddr)
}

func UdpLongConnection() {
	udpAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:5678")
	transport.CheckError(err)

	conn, err := net.ListenUDP("udp", udpAddr)
	transport.CheckError(err)

	log.Println("return conn")
	defer conn.Close()

	time.Sleep(time.Second * 5)

	request := make([]byte, 256)
	for {
		conn.SetReadDeadline(time.Now().Add(2 * time.Minute))
		n, remoteAddr, err := conn.ReadFromUDP(request)
		transport.CheckError(err)
		log.Printf("receive request %s from %s\n", string(request[:n]), remoteAddr.String())
	}
}
