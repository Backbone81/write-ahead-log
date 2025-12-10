# Benchmarks

This document provides results of the benchmarks provided with the sources for reference.

## Encoding

The encoding package is responsible for the low level encoding and decoding of data to file.
The benchmark is running against a memory buffer to not have measurements influenced by disk performance.
Note that no memory allocations happen for reading and writing which is important for keeping performance up.

```
goos: linux
goarch: amd64
pkg: github.com/backbone81/write-ahead-log/internal/encoding
cpu: Intel(R) Core(TM) i9-14900K
BenchmarkEntryChecksumWriter/crc32_on_0_KB-32           213347302                5.623 ns/op           0 B/op          0 allocs/op
BenchmarkEntryChecksumWriter/crc32_on_1_KB-32           46428430                25.57 ns/op            0 B/op          0 allocs/op
BenchmarkEntryChecksumWriter/crc32_on_2_KB-32           23520754                49.00 ns/op            0 B/op          0 allocs/op
BenchmarkEntryChecksumWriter/crc32_on_4_KB-32           12500259                95.15 ns/op            0 B/op          0 allocs/op
BenchmarkEntryChecksumWriter/crc32_on_8_KB-32            5990434               191.6 ns/op             0 B/op          0 allocs/op
BenchmarkEntryChecksumWriter/crc32_on_16_KB-32           3061810               384.9 ns/op             0 B/op          0 allocs/op

BenchmarkEntryChecksumWriter/crc64_on_0_KB-32           246808075                4.824 ns/op           0 B/op          0 allocs/op
BenchmarkEntryChecksumWriter/crc64_on_1_KB-32            3433406               347.0 ns/op             0 B/op          0 allocs/op
BenchmarkEntryChecksumWriter/crc64_on_2_KB-32            1748809               686.7 ns/op             0 B/op          0 allocs/op
BenchmarkEntryChecksumWriter/crc64_on_4_KB-32             858184              1396 ns/op               0 B/op          0 allocs/op
BenchmarkEntryChecksumWriter/crc64_on_8_KB-32             399157              2791 ns/op               0 B/op          0 allocs/op
BenchmarkEntryChecksumWriter/crc64_on_16_KB-32            216918              5539 ns/op               0 B/op          0 allocs/op

BenchmarkEntryChecksumReader/crc32_on_0_KB-32           143847985                8.272 ns/op           0 B/op          0 allocs/op
BenchmarkEntryChecksumReader/crc32_on_1_KB-32           41926746                28.46 ns/op            0 B/op          0 allocs/op
BenchmarkEntryChecksumReader/crc32_on_2_KB-32           23160171                52.03 ns/op            0 B/op          0 allocs/op
BenchmarkEntryChecksumReader/crc32_on_4_KB-32           12166480                98.78 ns/op            0 B/op          0 allocs/op
BenchmarkEntryChecksumReader/crc32_on_8_KB-32            6235137               191.8 ns/op             0 B/op          0 allocs/op
BenchmarkEntryChecksumReader/crc32_on_16_KB-32           3111595               383.3 ns/op             0 B/op          0 allocs/op

BenchmarkEntryChecksumReader/crc64_on_0_KB-32           152380088                7.892 ns/op           0 B/op          0 allocs/op
BenchmarkEntryChecksumReader/crc64_on_1_KB-32            3414534               351.0 ns/op             0 B/op          0 allocs/op
BenchmarkEntryChecksumReader/crc64_on_2_KB-32            1723486               691.1 ns/op             0 B/op          0 allocs/op
BenchmarkEntryChecksumReader/crc64_on_4_KB-32             764466              1381 ns/op               0 B/op          0 allocs/op
BenchmarkEntryChecksumReader/crc64_on_8_KB-32             430243              2758 ns/op               0 B/op          0 allocs/op
BenchmarkEntryChecksumReader/crc64_on_16_KB-32            216915              5504 ns/op               0 B/op          0 allocs/op

BenchmarkEntryLengthWriter/uint16-32                    614146473                1.942 ns/op           0 B/op          0 allocs/op
BenchmarkEntryLengthWriter/uint32-32                    619950919                1.932 ns/op           0 B/op          0 allocs/op
BenchmarkEntryLengthWriter/uint64-32                    620832824                1.937 ns/op           0 B/op          0 allocs/op
BenchmarkEntryLengthWriter/uvarint-32                   442404314                2.725 ns/op           0 B/op          0 allocs/op

BenchmarkEntryLengthReader/uint16-32                    189283162                6.338 ns/op           0 B/op          0 allocs/op
BenchmarkEntryLengthReader/uint32-32                    239519172                4.978 ns/op           0 B/op          0 allocs/op
BenchmarkEntryLengthReader/uint64-32                    245305549                4.867 ns/op           0 B/op          0 allocs/op
BenchmarkEntryLengthReader/uvarint-32                   135220516                8.879 ns/op           0 B/op          0 allocs/op

BenchmarkWriteHeader-32                                 510349974                2.216 ns/op           0 B/op          0 allocs/op
BenchmarkReadHeader-32                                  155079242                7.732 ns/op           0 B/op          0 allocs/op
```

## Segment

The segment package is responsible for reading and writing entries to a single file.
The benchmark is running against a memory buffer to not have measurements influenced by disk performance.
Note that no memory allocations happen for reading and writing which is important for keeping performance up.

```
goos: linux
goarch: amd64
pkg: github.com/backbone81/write-ahead-log/internal/segment
cpu: Intel(R) Core(TM) i9-14900K
BenchmarkSegmentReader_Next/uint16_crc32_0_KB-32                        45895807                25.03 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uint16_crc32_1_KB-32                        20668233                57.75 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uint16_crc32_2_KB-32                        14072797                84.99 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uint16_crc32_4_KB-32                         8535564               141.4 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uint16_crc32_8_KB-32                         4787287               250.1 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uint16_crc32_16_KB-32                        2569746               469.4 ns/op             0 B/op          0 allocs/op

BenchmarkSegmentReader_Next/uint16_crc64_0_KB-32                        46907268                23.81 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uint16_crc64_1_KB-32                         3193688               375.0 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uint16_crc64_2_KB-32                         1681652               713.9 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uint16_crc64_4_KB-32                          756452              1424 ns/op               0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uint16_crc64_8_KB-32                          414782              2802 ns/op               0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uint16_crc64_16_KB-32                         207378              5782 ns/op               0 B/op          0 allocs/op

BenchmarkSegmentReader_Next/uint32_crc32_0_KB-32                        46702518                25.23 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uint32_crc32_1_KB-32                        21019095                56.97 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uint32_crc32_2_KB-32                        14226814                83.79 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uint32_crc32_4_KB-32                         8555097               138.2 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uint32_crc32_8_KB-32                         4882568               249.4 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uint32_crc32_16_KB-32                        2536965               472.3 ns/op             0 B/op          0 allocs/op

BenchmarkSegmentReader_Next/uint32_crc64_0_KB-32                        42192098                24.44 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uint32_crc64_1_KB-32                         3260257               367.6 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uint32_crc64_2_KB-32                         1690485               709.5 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uint32_crc64_4_KB-32                          835752              1387 ns/op               0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uint32_crc64_8_KB-32                          425394              2775 ns/op               0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uint32_crc64_16_KB-32                         208369              5754 ns/op               0 B/op          0 allocs/op

BenchmarkSegmentReader_Next/uint64_crc32_0_KB-32                        43875032                26.58 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uint64_crc32_1_KB-32                        18898818                62.83 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uint64_crc32_2_KB-32                        13282293                90.15 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uint64_crc32_4_KB-32                         8223826               144.6 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uint64_crc32_8_KB-32                         4733448               255.8 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uint64_crc32_16_KB-32                        2525044               475.9 ns/op             0 B/op          0 allocs/op

BenchmarkSegmentReader_Next/uint64_crc64_0_KB-32                        43411516                24.19 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uint64_crc64_1_KB-32                         3164664               377.8 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uint64_crc64_2_KB-32                         1661656               724.0 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uint64_crc64_4_KB-32                          827866              1419 ns/op               0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uint64_crc64_8_KB-32                          423261              2828 ns/op               0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uint64_crc64_16_KB-32                         205608              5804 ns/op               0 B/op          0 allocs/op

BenchmarkSegmentReader_Next/uvarint_crc32_0_KB-32                       46157852                23.68 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uvarint_crc32_1_KB-32                       19019785                57.55 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uvarint_crc32_2_KB-32                       14171289                83.27 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uvarint_crc32_4_KB-32                        8584347               139.3 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uvarint_crc32_8_KB-32                        4796556               249.1 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uvarint_crc32_16_KB-32                       2522905               475.2 ns/op             0 B/op          0 allocs/op

BenchmarkSegmentReader_Next/uvarint_crc64_0_KB-32                       54034370                21.82 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uvarint_crc64_1_KB-32                        3207948               373.5 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uvarint_crc64_2_KB-32                        1660592               722.0 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uvarint_crc64_4_KB-32                         840100              1424 ns/op               0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uvarint_crc64_8_KB-32                         424735              2817 ns/op               0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uvarint_crc64_16_KB-32                        205326              5783 ns/op               0 B/op          0 allocs/op

BenchmarkSegmentWriter_AppendEntry/uint16_crc32_0_KB-32                 68703928                17.40 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint16_crc32_1_KB-32                 22918502                52.38 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint16_crc32_2_KB-32                 15176880                78.27 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint16_crc32_4_KB-32                  8543607               138.8 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint16_crc32_8_KB-32                  4640926               258.1 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint16_crc32_16_KB-32                 2408066               497.7 ns/op             0 B/op          0 allocs/op

BenchmarkSegmentWriter_AppendEntry/uint16_crc64_0_KB-32                 72741366                16.49 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint16_crc64_1_KB-32                  3269170               368.1 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint16_crc64_2_KB-32                  1667859               717.7 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint16_crc64_4_KB-32                   813744              1429 ns/op               0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint16_crc64_8_KB-32                   422606              2796 ns/op               0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint16_crc64_16_KB-32                  207418              5741 ns/op               0 B/op          0 allocs/op

BenchmarkSegmentWriter_AppendEntry/uint32_crc32_0_KB-32                 65454710                18.04 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint32_crc32_1_KB-32                 21679323                55.26 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint32_crc32_2_KB-32                 14590278                81.66 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint32_crc32_4_KB-32                  8578970               140.0 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint32_crc32_8_KB-32                  4698958               254.4 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint32_crc32_16_KB-32                 2416170               492.4 ns/op             0 B/op          0 allocs/op

BenchmarkSegmentWriter_AppendEntry/uint32_crc64_0_KB-32                 69640263                17.11 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint32_crc64_1_KB-32                  3298201               369.6 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint32_crc64_2_KB-32                  1668832               719.5 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint32_crc64_4_KB-32                   834564              1414 ns/op               0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint32_crc64_8_KB-32                   423424              2804 ns/op               0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint32_crc64_16_KB-32                  208594              5744 ns/op               0 B/op          0 allocs/op

BenchmarkSegmentWriter_AppendEntry/uint64_crc32_0_KB-32                 64031096                18.39 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint64_crc32_1_KB-32                 19530834                61.07 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint64_crc32_2_KB-32                 13381654                87.76 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint64_crc32_4_KB-32                  8301666               143.0 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint64_crc32_8_KB-32                  4451118               261.9 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint64_crc32_16_KB-32                 2416978               491.8 ns/op             0 B/op          0 allocs/op

BenchmarkSegmentWriter_AppendEntry/uint64_crc64_0_KB-32                 66843457                17.81 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint64_crc64_1_KB-32                  3274482               364.2 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint64_crc64_2_KB-32                  1670439               717.8 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint64_crc64_4_KB-32                   833280              1429 ns/op               0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint64_crc64_8_KB-32                   423667              2828 ns/op               0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint64_crc64_16_KB-32                  208161              5756 ns/op               0 B/op          0 allocs/op

BenchmarkSegmentWriter_AppendEntry/uvarint_crc32_0_KB-32                69399318                17.25 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uvarint_crc32_1_KB-32                21201792                52.51 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uvarint_crc32_2_KB-32                13605531                79.18 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uvarint_crc32_4_KB-32                 8731087               138.0 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uvarint_crc32_8_KB-32                 4732102               254.6 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uvarint_crc32_16_KB-32                2381695               502.1 ns/op             0 B/op          0 allocs/op

BenchmarkSegmentWriter_AppendEntry/uvarint_crc64_0_KB-32                73579068                16.38 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uvarint_crc64_1_KB-32                 3287173               368.2 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uvarint_crc64_2_KB-32                 1691874               705.5 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uvarint_crc64_4_KB-32                  849597              1425 ns/op               0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uvarint_crc64_8_KB-32                  423432              2812 ns/op               0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uvarint_crc64_16_KB-32                 207691              5742 ns/op               0 B/op          0 allocs/op
```

## WAL

The wal package is responsible for abstracting away multiple segment files behind a uniform interface.
The benchmark is running against a disk which influences performance.

```
goos: linux
goarch: amd64
pkg: github.com/backbone81/write-ahead-log/internal/wal
cpu: Intel(R) Core(TM) i9-14900K
BenchmarkReader_Next/uint32_crc32_0_KB-32                                        2615650               436.4 ns/op                            14 B/op          0 allocs/op
BenchmarkReader_Next/uint32_crc32_1_KB-32                                        1681372               713.0 ns/op      1368.89 MB/s          14 B/op          0 allocs/op
BenchmarkReader_Next/uint32_crc32_2_KB-32                                        1553118               785.7 ns/op      2485.36 MB/s          14 B/op          0 allocs/op
BenchmarkReader_Next/uint32_crc32_4_KB-32                                        1390792               860.0 ns/op      4541.63 MB/s          22 B/op          0 allocs/op
BenchmarkReader_Next/uint32_crc32_8_KB-32                                        1000000              1093 ns/op        7148.90 MB/s          31 B/op          0 allocs/op
BenchmarkReader_Next/uint32_crc32_16_KB-32                                        730272              1592 ns/op        9813.42 MB/s          44 B/op          0 allocs/op

BenchmarkWriter_AppendEntry_Serial/uint32_crc32_periodic_0_KB-32                   16467             74553 ns/op                               0 B/op          0 allocs/op
BenchmarkWriter_AppendEntry_Serial/uint32_crc32_periodic_1_KB-32                   14802             76084 ns/op          12.43 MB/s           0 B/op          0 allocs/op
BenchmarkWriter_AppendEntry_Serial/uint32_crc32_periodic_2_KB-32                   14689             89093 ns/op          21.40 MB/s           0 B/op          0 allocs/op
BenchmarkWriter_AppendEntry_Serial/uint32_crc32_periodic_4_KB-32                   15500             79498 ns/op          48.69 MB/s           1 B/op          0 allocs/op
BenchmarkWriter_AppendEntry_Serial/uint32_crc32_periodic_8_KB-32                   14118             98869 ns/op          78.81 MB/s           1 B/op          0 allocs/op
BenchmarkWriter_AppendEntry_Serial/uint32_crc32_periodic_16_KB-32                  10000            115832 ns/op         134.68 MB/s           5 B/op          0 allocs/op

BenchmarkWriter_AppendEntry_Serial/uint32_crc32_grouped_0_KB-32                      100          16392326 ns/op                               0 B/op          0 allocs/op
BenchmarkWriter_AppendEntry_Serial/uint32_crc32_grouped_1_KB-32                      100          15568101 ns/op                              50 B/op          0 allocs/op
BenchmarkWriter_AppendEntry_Serial/uint32_crc32_grouped_2_KB-32                      100          15693179 ns/op                               0 B/op          0 allocs/op
BenchmarkWriter_AppendEntry_Serial/uint32_crc32_grouped_4_KB-32                      100          15593445 ns/op                              31 B/op          0 allocs/op
BenchmarkWriter_AppendEntry_Serial/uint32_crc32_grouped_8_KB-32                      100          15522970 ns/op                              18 B/op          0 allocs/op
BenchmarkWriter_AppendEntry_Serial/uint32_crc32_grouped_16_KB-32                     100          15460634 ns/op           0.65 MB/s           0 B/op          0 allocs/op

BenchmarkWriter_AppendEntry_Serial/uint32_crc32_none_0_KB-32                     4041283               283.1 ns/op                             0 B/op          0 allocs/op
BenchmarkWriter_AppendEntry_Serial/uint32_crc32_none_1_KB-32                     1610323               694.2 ns/op      1406.29 MB/s           0 B/op          0 allocs/op
BenchmarkWriter_AppendEntry_Serial/uint32_crc32_none_2_KB-32                     1000000              1064 ns/op        1836.32 MB/s           0 B/op          0 allocs/op
BenchmarkWriter_AppendEntry_Serial/uint32_crc32_none_4_KB-32                      591164              1816 ns/op        2150.58 MB/s           0 B/op          0 allocs/op
BenchmarkWriter_AppendEntry_Serial/uint32_crc32_none_8_KB-32                      384589              3343 ns/op        2336.82 MB/s           1 B/op          0 allocs/op
BenchmarkWriter_AppendEntry_Serial/uint32_crc32_none_16_KB-32                     172198              6566 ns/op        2379.14 MB/s           5 B/op          0 allocs/op

BenchmarkWriter_AppendEntry_Serial/uint32_crc32_immediate_0_KB-32                    214           5597603 ns/op                               0 B/op          0 allocs/op
BenchmarkWriter_AppendEntry_Serial/uint32_crc32_immediate_1_KB-32                    189           5948959 ns/op                               0 B/op          0 allocs/op
BenchmarkWriter_AppendEntry_Serial/uint32_crc32_immediate_2_KB-32                    177           5765668 ns/op                               0 B/op          0 allocs/op
BenchmarkWriter_AppendEntry_Serial/uint32_crc32_immediate_4_KB-32                    225           5218694 ns/op                               0 B/op          0 allocs/op
BenchmarkWriter_AppendEntry_Serial/uint32_crc32_immediate_8_KB-32                    230           5673873 ns/op           0.77 MB/s           0 B/op          0 allocs/op
BenchmarkWriter_AppendEntry_Serial/uint32_crc32_immediate_16_KB-32                   177           5974674 ns/op           1.89 MB/s           0 B/op          0 allocs/op

BenchmarkWriter_AppendEntry_Concurrently/uint32_crc32_none_0_KB-32               1153914              1108 ns/op                               0 B/op          0 allocs/op
BenchmarkWriter_AppendEntry_Concurrently/uint32_crc32_none_1_KB-32               1000000              1427 ns/op         683.73 MB/s           0 B/op          0 allocs/op
BenchmarkWriter_AppendEntry_Concurrently/uint32_crc32_none_2_KB-32                679923              1839 ns/op        1061.48 MB/s           0 B/op          0 allocs/op
BenchmarkWriter_AppendEntry_Concurrently/uint32_crc32_none_4_KB-32                532069              2860 ns/op        1365.33 MB/s           0 B/op          0 allocs/op
BenchmarkWriter_AppendEntry_Concurrently/uint32_crc32_none_8_KB-32                268378              4717 ns/op        1655.58 MB/s           1 B/op          0 allocs/op
BenchmarkWriter_AppendEntry_Concurrently/uint32_crc32_none_16_KB-32               139796              7897 ns/op        1978.23 MB/s           5 B/op          0 allocs/op

BenchmarkWriter_AppendEntry_Concurrently/uint32_crc32_immediate_0_KB-32             8752            125994 ns/op                               3 B/op          0 allocs/op
BenchmarkWriter_AppendEntry_Concurrently/uint32_crc32_immediate_1_KB-32            12942             97724 ns/op           9.49 MB/s          20 B/op          0 allocs/op
BenchmarkWriter_AppendEntry_Concurrently/uint32_crc32_immediate_2_KB-32            13040             90753 ns/op          21.13 MB/s           0 B/op          0 allocs/op
BenchmarkWriter_AppendEntry_Concurrently/uint32_crc32_immediate_4_KB-32            14280             75944 ns/op          50.72 MB/s         152 B/op          0 allocs/op
BenchmarkWriter_AppendEntry_Concurrently/uint32_crc32_immediate_8_KB-32            14719             78176 ns/op          99.07 MB/s         127 B/op          0 allocs/op
BenchmarkWriter_AppendEntry_Concurrently/uint32_crc32_immediate_16_KB-32            8954            200193 ns/op          77.54 MB/s        1234 B/op          3 allocs/op

BenchmarkWriter_AppendEntry_Concurrently/uint32_crc32_periodic_0_KB-32            150778              7527 ns/op                               0 B/op          0 allocs/op
BenchmarkWriter_AppendEntry_Concurrently/uint32_crc32_periodic_1_KB-32            145207              9811 ns/op          98.97 MB/s           0 B/op          0 allocs/op
BenchmarkWriter_AppendEntry_Concurrently/uint32_crc32_periodic_2_KB-32            156138              9301 ns/op         209.33 MB/s           0 B/op          0 allocs/op
BenchmarkWriter_AppendEntry_Concurrently/uint32_crc32_periodic_4_KB-32            122088             10798 ns/op         361.07 MB/s           1 B/op          0 allocs/op
BenchmarkWriter_AppendEntry_Concurrently/uint32_crc32_periodic_8_KB-32             59461             19189 ns/op         406.61 MB/s           1 B/op          0 allocs/op
BenchmarkWriter_AppendEntry_Concurrently/uint32_crc32_periodic_16_KB-32            44600             30250 ns/op         515.87 MB/s           6 B/op          0 allocs/op

BenchmarkWriter_AppendEntry_Concurrently/uint32_crc32_grouped_0_KB-32             642428              2847 ns/op                               0 B/op          0 allocs/op
BenchmarkWriter_AppendEntry_Concurrently/uint32_crc32_grouped_1_KB-32             394404              3075 ns/op         317.40 MB/s           0 B/op          0 allocs/op
BenchmarkWriter_AppendEntry_Concurrently/uint32_crc32_grouped_2_KB-32             368478              4187 ns/op         465.97 MB/s           0 B/op          0 allocs/op
BenchmarkWriter_AppendEntry_Concurrently/uint32_crc32_grouped_4_KB-32             206770              5356 ns/op         728.66 MB/s           0 B/op          0 allocs/op
BenchmarkWriter_AppendEntry_Concurrently/uint32_crc32_grouped_8_KB-32             130074              9198 ns/op         849.22 MB/s           2 B/op          0 allocs/op
BenchmarkWriter_AppendEntry_Concurrently/uint32_crc32_grouped_16_KB-32             80832             17276 ns/op         904.42 MB/s           6 B/op          0 allocs/op
```
