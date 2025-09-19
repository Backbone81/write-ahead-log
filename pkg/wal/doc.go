// Package wal provides the implementation of a general purpose write-ahead log.
//
//   - The write-ahead log is made up of individual entries, which contain arbitrary data as a payload and metadata
//     in the form of payload length and a checksum.
//   - Entries are stored in segment files. Each segment file consists of a file header describing some details of the
//     entries stored in the segment. After the file header, the entries follow one after the other.All segment files
//     are assumed to be located in the same directory. Every segment file has the sequence number of its first entry
//     as its file name, padded with leading zeros to be 20 characters in length with a `.wal` file extension.
//   - The write-ahead log abstracts away the fact that entries are stored in segment files and provides a uniform
//     interface for reading and writing entries without knowing those details.
//   - Sequence numbers uniquely identify every entry. They are unsigned 64-bit integers which are monotonically
//     increasing.
package wal
