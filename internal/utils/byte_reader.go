package utils

import (
	"io"
)

type ByteReader struct {
	reader  io.Reader
	buffer  []byte
	counter int
}

func NewByteReader(reader io.Reader, buffer []byte) ByteReader {
	return ByteReader{
		reader: reader,
		buffer: buffer,
	}
}

func (b *ByteReader) ReadByte() (byte, error) {
	if _, err := io.ReadFull(b.reader, b.buffer[b.counter:b.counter+1]); err != nil {
		return 0, err
	}
	result := b.buffer[b.counter]
	b.counter++
	return result, nil
}

func (b *ByteReader) BytesRead() int {
	return b.counter
}
