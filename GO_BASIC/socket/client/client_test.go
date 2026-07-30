package client_test

import (
	"go_basic/socket/client"
	"testing"
)

func TestTcpClient(t *testing.T) {
	client.TcpClient()
}

func TestUdpClient(t *testing.T) {
	client.UdpClient()
}
