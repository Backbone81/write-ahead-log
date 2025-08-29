package wal

import (
	"encoding/binary"
	"unsafe"
)

var (
	Endian = detectNativeEndian()
)

// detectNativeEndian determines the native endianness of the system the application is running on.
func detectNativeEndian() binary.ByteOrder {
	var i uint16 = 0x1
	b := (*[2]byte)(unsafe.Pointer(&i))
	if b[0] == 0x1 {
		return binary.LittleEndian
	}
	return binary.BigEndian
}

type SyncPolicy interface {
	EntryAppended(sequenceNumber uint64) error
	Close() error
}
