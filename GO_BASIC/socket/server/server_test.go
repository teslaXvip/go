package server_test

import (
	"go_basic/socket/server"
	"testing"
)

func TestTcpServer(t *testing.T) {
	server.TcpServer()
}

func TestUdpServer(t *testing.T) {
	server.UdpServer()
}
