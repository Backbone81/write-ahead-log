# Write Ahead Log

This library provides an implementation of a write-ahead log. It can be useful for database systems or other situations
where a loss of in-memory data needs to be restored by from disk.

It supports the following features:

- Transparent segmentation of the write-ahead log into multiple segment files with a configurable size for rollover.
- The writer is safe to use for concurrent writes without external synchronization.
- Several hash policies to configure the hash calculated for every entry.
- Hash policy "crc32" provides a fast and simple checksum for small entry sizes.
- TODO: Hash policy "crc64" provides more reliability for bigger entry sizes.
- Several sync policies to adjust the way entries are flushed to stable storage.
- Sync policy "none" for situations where it is not necessary to flush entries of the write-ahead log to stable
  storage at all. This might be helpful for tests.
- Sync policy "immediate" for flushing every single entry to stable storage immediately. This provides the most
  reliability but incurs the highest cost with regard to latency.
- Sync policy "periodic" for flushing multiple entries asynchronously in a regular interval or after some number of
  entries to stable storage. This provides a middle ground between "none" and immediate. Your code is not blocked until
  a flush occurs, so there is still a small time window where data might get lost.
- Sync policy "grouped" for flushing all entries synchronously which are written within a defined time window after
  the first pending entry. This amortizes the cost of flushing data to stable storage over multiple concurrent writes.
  It guarantees that the entry was flushed after the call to the writer returns.

## TODOs

- Implement segment creation which works on windows and linux
- Drop dynamic endianness and have an inline implementation with a fixed implementation.
- Replace the varint implementation with a custom one which does not use a bytereader.
- Collect metrics about what is happening with the wal
- Provide a CLI for inspecting the WAL and to do maintenance
- Add checksum to the segment header
