package client

import (
	transport "go_basic/socket"
	"log"
	"net"
)

func connect2TcpServer(serverAddr string) net.Conn {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", serverAddr)
	transport.CheckError(err)
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	transport.CheckError(err)
	return conn
}

func sendTcpServer(conn net.Conn) {
	n, err := conn.Write([]byte("hello"))
	transport.CheckError(err)
	log.Printf("send %d bytes\n", n)
}

func TcpClient() {
	conn := connect2TcpServer("127.0.0.1:5678")
	sendTcpServer(conn)

}

func TcpLongConnection() {
	conn := connect2TcpServer("127.0.0.1:5678")
	for range 300 {
		sendTcpServer(conn)
	}
	conn.Close()
	log.Println("close connection")
}
