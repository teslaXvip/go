package main

import (
	"go_frame/io"
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
