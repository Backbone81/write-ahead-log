# Write Ahead Log

This library provides a write-ahead log. It can be useful for database systems, distributed systems or other situations
where a loss of in-memory data needs to be restored from stable storage. It was carefully engineered with performance
and zero memory allocations in mind.

It supports the following features:

- Transparent segmentation of the write-ahead log into multiple segment files with a configurable size for rollover.
- The writer is safe to use for concurrent writes without external synchronization.
- Several checksum policies to configure the checksum calculated for every entry.
  - Checksum policy "crc32" provides a fast and simple checksum for small entry sizes.
  - Checksum policy "crc64" provides more reliability for bigger entry sizes.
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
- Custom metrics for insights into the WAL.

## Examples

See the [examples](examples) folder for how to use this library.

## Benchmarks

All parts of this library are covered with extensive benchmarks. See [docs/benchmarks](docs/benchmarks.md) for details.

## TODOs

- The creation of segments currently only works on Linux but not on Windows. This is because Linux allows for renaming
  files which are open, which Windows does not support. For Windows support, dedicated code for Windows needs to be
  added.
- Reading from disk might be improved by introducing a bufio.Reader. This might result in less read calls to disk and
  improve read performance. Measurements need to be done to gain insights.
- A CLI for inspecting a write-ahead log and for doing some maintenance like consolidating multiple segment files into
  one or changing the configuration might prove valuable for production grade usage.
