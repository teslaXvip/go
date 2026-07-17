package io

import (
	"os"
)

type BufferedFileWriter struct {
	fout           *os.File
	buffer         []byte
	bufferEndIndex int
}

func NewBufferedFileWriter(fout *os.File, bufferSize int) *BufferedFileWriter {
	return &BufferedFileWriter{
		fout:           fout,
		buffer:         make([]byte, bufferSize),
		bufferEndIndex: 0,
	}
}

func (b *BufferedFileWriter) Flush() {
	b.fout.Write(b.buffer[:b.bufferEndIndex])
	b.bufferEndIndex = 0
}

func (b *BufferedFileWriter) Write(cont []byte) {
	if len(cont) > len(b.buffer) {
		b.Flush()
		b.fout.Write(cont)
	} else {
		if len(cont)+b.bufferEndIndex > len(b.buffer) {
			b.Flush()
		}
		copy(b.buffer[b.bufferEndIndex:], cont)
		b.bufferEndIndex += len(cont)
	}
}

func (b *BufferedFileWriter) WriteString(cont string) {
	b.Write([]byte(cont))
}
