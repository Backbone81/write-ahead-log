// Package wal provides an implementation of a write-ahead log.
//
// The on-disk structure looks like this:
//
//   - The write-ahead log is made up of multiple segment files. All segment files are assumed to be located in the same
//     directory. Every segment file has the sequence number of its first entry as its file name, padded with leading
//     zeros to be 20 characters in length with a `.wal` file extension.
//   - Each segment file consists of a file header describing some details of the entries stored in the segment. After
//     the file header, the entries follow one after the other. Each entry is made up of a length, the data itself and
//     a checksum.
//   - Sequence numbers uniquely identify every entry. They are unsigned 64-bit integers which are monotonically
//     increasing.
package wal
