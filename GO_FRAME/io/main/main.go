package main

import (
	"github.com/teslaXvip/go/go_frame/io"
)

func main() {
	io.InitArray()

	io.WriteRamSequentially()
	io.ReadRamSequentially()
	io.ReadRamRandomly()
	io.WriteRamRandomly()
	io.WriteDiskSequentially()
	io.ReadDiskSequentially()
}
