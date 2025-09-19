package encoding

import "encoding/binary"

// Endian is the endianness the write-ahead log uses for serializing/deserializing integers to file.
var Endian = binary.LittleEndian
