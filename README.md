# Write Ahead Log

This Go library provides a high-performance, zero-allocation general purpose write-ahead log (WAL) implementation. It is
designed for systems that require durability, crash recovery, and efficient sequential writes, such as databases,
distributed systems, and event sourcing architectures.

## What is a Write-Ahead Log?

A write-ahead log is a persistent, append-only log that records changes before they are applied to the main data store.
By writing changes to the log first, systems can recover from crashes or power failures by replaying the log and
restoring consistency.

## Use Cases

- **Databases**: Ensure durability and atomicity of transactions.
- **Distributed Systems**: Achieve consensus and replicate state changes.
- **Event Sourcing**: Store immutable event streams for system state reconstruction.
- **Message Queues**: Persist messages for reliable delivery.

## Why Use This Library?

- **Transparent Segmentation**: Automatically splits the log into segment files with configurable rollover.
- **Concurrent Writes**: Thread-safe writer for high-throughput, multi-goroutine environments.
- **Configurable Checksums**: Choose between different algorithms for data integrity.
- **Flexible Sync Policies**: Select from different policies to balance durability and performance.
- **Custom Metrics**: Integrate with your monitoring stack for operational insights.
- **Zero Allocations**: Engineered for minimal GC pressure and maximum throughput.

## Quick Start

Install:

```sh
go get github.com/backbone81/write-ahead-log
```

```go
package main

import (
	"log"

	"github.com/backbone81/write-ahead-log/pkg/wal"
)

func main() {
	if err := runDemo(); err != nil {
		log.Fatalln(err)
	}
}

func runDemo() error {
	walDir := "."

	// Make sure the write-ahead log is initialized. This means that
	// at least one wal file is in the wal directory.
	if err := wal.InitIfRequired(walDir); err != nil {
		return err
	}

	// This is the sequence number which you store somewhere else. For
	// this example, we always assume 0 but in reality this should
	// change whenever you flush your own state to disk.
	var sequenceNumber uint64 = 0

	// Before we can write to the write-ahead log, we need to read all
	// entries starting from our well-known sequence number.
	log.Println("Opening write-ahead log for reading")
	reader, err := wal.NewReader(walDir, sequenceNumber)
	if err != nil {
		return err
	}
	for reader.Next() {
		// We would use reader.Value().Data and apply it to our saved
		// state. reader.Next will return false when we have reached
		// the end of the write-ahead log.
	}

	// We convert the reader to a writer. Note that this is only
	// possible after having read all entries. Afterward the reader
	// cannot be used anymore.
	log.Println("Opening write-ahead log for writing")
	writer, err := reader.ToWriter()
	if err != nil {
		return err
	}

	// We can now append entries. This can also happen concurrently
	// from multiple go routines.
	log.Println("Appending entry")
	if _, err := writer.AppendEntry([]byte("foo")); err != nil {
		return err
	}

	// At the end we should close the writer to make sure that all
	// data was flushed to stable storage.
	if err := writer.Close(); err != nil {
		return err
	}
	return nil
}
```

See the [examples](examples) folder for more examples.

## CLI

You can also use the CLI for interacting with the write-ahead log. To install:

```
go install github.com/backbone81/write-ahead-log/cmd/wal-cli@latest
```

Use the `--help` flag to get an overview of all options:

```
A tool for interacting with write-ahead logs.

Usage:
  wal-cli [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  describe    Provides detailed information about the write-ahead log.
  help        Help about any command
  init        Initializes a new write-ahead log.

Flags:
  -d, --directory string   The directory the write-ahead log is located in. (default ".")
  -h, --help               help for wal-cli

Use "wal-cli [command] --help" for more information about a command.
```


## Length Encoding

The following entry length encodings are currently supported:

- **uint16**: For situations where you know that your WAL entries will never exceed the length of 65,535 bytes. This
  will save space when you have lots of very small entries.
- **uint32**: A good fit for most situations as it allows for WAL entries with a length of up to 4,294,967,295 bytes.
- **uint64**: For situations where 32 bits are not enough, and you need all 64 bits for the length. This can be quite
  wasteful when your entries are small.
- **uvarint**: Variable-length encoding where small values use fewer bytes and larger bylues use more.

## Checksum Type

The following checksum types are currently supported:

- **crc32**: Provides a fast and simple checksum for small entry sizes.
- **crc64**: Provides more reliability for bigger entry sizes.

## Sync Policies

The following sync policies are currently supported:

- **none**: For situations where it is not necessary to flush entries of the write-ahead log to stable
  storage at all. This might be helpful for tests.
- **immediate**: For flushing every single entry to stable storage immediately. This provides the most
  reliability but incurs the highest cost with regard to latency.
- **periodic**: For flushing multiple entries asynchronously in a regular interval or after some number of
  entries to stable storage. This provides a middle ground between "none" and immediate. Your code is not blocked until
  a flush occurs, so there is still a small time window where data might get lost.
- **grouped**: For flushing all entries synchronously which are written within a defined time window after
  the first pending entry. This amortizes the cost of flushing data to stable storage over multiple concurrent writes.
  It guarantees that the entry was flushed after the call to the writer returns.

## Metrics

Several metrics are provided to gain insights into the operation of the write-ahead log. You can register those metrics
with your prometheus registerer with `wal.RegisterMetrics()`.

## Benchmarks

All parts of this library are covered with extensive benchmarks. See [docs/benchmarks](docs/benchmarks.md) for details.

## TODOs

There are several points still open:

- Reading from disk might be improved by introducing a bufio.Reader. This might result in less read calls to disk and
  improve read performance. Measurements need to be done to gain insights.
- Extend the CLI with functionality to print details for every entry (sequence number, file offset, length).
- Extend the CLI with functionality for YAML and JSON output.
- Extend the CLI with functionality for rewriting an existing wal with different settings (entry length encoding, entry
  checksum type, max segment size)
- Extend the CLI with functionality to run read and write benchmarks to disk.
