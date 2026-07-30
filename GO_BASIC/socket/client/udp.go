package client

import (
	transport "go_basic/socket"
	"log"

	"net"
	"time"
)

// 连接UDP服务端
func connect2UdpServer(serverAddr string) net.Conn {
	conn, err := net.DialTimeout("udp", serverAddr, 3*time.Minute)
	transport.CheckError(err)

	log.Printf("establish connection to server %s myself %s\n",
		conn.RemoteAddr().String(),
		conn.LocalAddr().String())
	return conn
}

// 发送UDP数据
func sendUdpServer(conn net.Conn) {
	n, err := conn.Write([]byte("hello"))
	// 注释：即使Server还未启动，建立连接和发送数据都不会返回error，Server启动后也收不到这个数据
	transport.CheckError(err)
	log.Printf("send %d bytes\n", n)
}

func UdpClient() {
	conn := connect2UdpServer("127.0.0.1:5678")
	sendUdpServer(conn)
	conn.Close()
	log.Println("close connection")
}

func UdpLongConnection() {
	conn := connect2UdpServer("127.0.0.1:5678")
	for range 3 {
		sendUdpServer(conn)
	}
	conn.Close()
	log.Println("close connection")
}
