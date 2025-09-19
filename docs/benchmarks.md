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
BenchmarkEntryChecksumWriter/crc32_on_0_KB-32           205598827                5.836 ns/op           0 B/op          0 allocs/op
BenchmarkEntryChecksumWriter/crc32_on_1_KB-32           47105160                25.54 ns/op            0 B/op          0 allocs/op
BenchmarkEntryChecksumWriter/crc32_on_2_KB-32           21886632                50.30 ns/op            0 B/op          0 allocs/op
BenchmarkEntryChecksumWriter/crc32_on_4_KB-32           12359979                97.08 ns/op            0 B/op          0 allocs/op
BenchmarkEntryChecksumWriter/crc32_on_8_KB-32            6267903               191.1 ns/op             0 B/op          0 allocs/op
BenchmarkEntryChecksumWriter/crc32_on_16_KB-32           3154522               380.6 ns/op             0 B/op          0 allocs/op

BenchmarkEntryChecksumWriter/crc64_on_0_KB-32           249316920                4.763 ns/op           0 B/op          0 allocs/op
BenchmarkEntryChecksumWriter/crc64_on_1_KB-32            3420583               349.8 ns/op             0 B/op          0 allocs/op
BenchmarkEntryChecksumWriter/crc64_on_2_KB-32            1726636               694.6 ns/op             0 B/op          0 allocs/op
BenchmarkEntryChecksumWriter/crc64_on_4_KB-32             781408              1397 ns/op               0 B/op          0 allocs/op
BenchmarkEntryChecksumWriter/crc64_on_8_KB-32             425374              2765 ns/op               0 B/op          0 allocs/op
BenchmarkEntryChecksumWriter/crc64_on_16_KB-32            217086              5524 ns/op               0 B/op          0 allocs/op

BenchmarkEntryChecksumReader/crc32_on_0_KB-32           135496183                8.846 ns/op           0 B/op          0 allocs/op
BenchmarkEntryChecksumReader/crc32_on_1_KB-32           42284120                28.44 ns/op            0 B/op          0 allocs/op
BenchmarkEntryChecksumReader/crc32_on_2_KB-32           21688291                53.23 ns/op            0 B/op          0 allocs/op
BenchmarkEntryChecksumReader/crc32_on_4_KB-32           11935336               100.7 ns/op             0 B/op          0 allocs/op
BenchmarkEntryChecksumReader/crc32_on_8_KB-32            6183378               194.1 ns/op             0 B/op          0 allocs/op
BenchmarkEntryChecksumReader/crc32_on_16_KB-32           3114247               383.5 ns/op             0 B/op          0 allocs/op

BenchmarkEntryChecksumReader/crc64_on_0_KB-32           155132996                7.737 ns/op           0 B/op          0 allocs/op
BenchmarkEntryChecksumReader/crc64_on_1_KB-32            3378812               353.1 ns/op             0 B/op          0 allocs/op
BenchmarkEntryChecksumReader/crc64_on_2_KB-32            1720148               697.2 ns/op             0 B/op          0 allocs/op
BenchmarkEntryChecksumReader/crc64_on_4_KB-32             859754              1387 ns/op               0 B/op          0 allocs/op
BenchmarkEntryChecksumReader/crc64_on_8_KB-32             433384              2769 ns/op               0 B/op          0 allocs/op
BenchmarkEntryChecksumReader/crc64_on_16_KB-32            216812              5526 ns/op               0 B/op          0 allocs/op

BenchmarkEntryLengthWriter/uint16-32                    609504109                1.950 ns/op           0 B/op          0 allocs/op
BenchmarkEntryLengthWriter/uint32-32                    619112922                1.945 ns/op           0 B/op          0 allocs/op
BenchmarkEntryLengthWriter/uint64-32                    614008844                1.943 ns/op           0 B/op          0 allocs/op
BenchmarkEntryLengthWriter/uvarint-32                   476988631                2.505 ns/op           0 B/op          0 allocs/op

BenchmarkEntryLengthReader/uint16-32                    187735140                6.379 ns/op           0 B/op          0 allocs/op
BenchmarkEntryLengthReader/uint32-32                    240415224                5.013 ns/op           0 B/op          0 allocs/op
BenchmarkEntryLengthReader/uint64-32                    245308364                4.881 ns/op           0 B/op          0 allocs/op
BenchmarkEntryLengthReader/uvarint-32                   122647065                9.764 ns/op           0 B/op          0 allocs/op

BenchmarkWriteHeader-32                                 548512419                2.189 ns/op           0 B/op          0 allocs/op
BenchmarkReadHeader-32                                  154525996                7.761 ns/op           0 B/op          0 allocs/op
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
BenchmarkSegmentReader_Next/uint16_crc32_0_KB-32                47161341                25.35 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uint16_crc32_1_KB-32                20667250                57.93 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uint16_crc32_2_KB-32                13981471                85.36 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uint16_crc32_4_KB-32                 8586624               140.1 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uint16_crc32_8_KB-32                 4757185               251.3 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uint16_crc32_16_KB-32                2549586               471.8 ns/op             0 B/op          0 allocs/op

BenchmarkSegmentReader_Next/uint16_crc64_0_KB-32                49389055                23.84 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uint16_crc64_1_KB-32                 3189102               375.4 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uint16_crc64_2_KB-32                 1660845               725.2 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uint16_crc64_4_KB-32                  830443              1423 ns/op               0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uint16_crc64_8_KB-32                  424890              2819 ns/op               0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uint16_crc64_16_KB-32                 207048              5786 ns/op               0 B/op          0 allocs/op

BenchmarkSegmentReader_Next/uint32_crc32_0_KB-32                45630573                25.78 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uint32_crc32_1_KB-32                20818885                57.47 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uint32_crc32_2_KB-32                14204276                84.52 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uint32_crc32_4_KB-32                 8544732               140.2 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uint32_crc32_8_KB-32                 4776418               252.2 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uint32_crc32_16_KB-32                2540028               472.2 ns/op             0 B/op          0 allocs/op

BenchmarkSegmentReader_Next/uint32_crc64_0_KB-32                45147618                24.96 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uint32_crc64_1_KB-32                 3204740               374.2 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uint32_crc64_2_KB-32                 1661263               722.4 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uint32_crc64_4_KB-32                  830768              1424 ns/op               0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uint32_crc64_8_KB-32                  423915              2821 ns/op               0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uint32_crc64_16_KB-32                 207391              5776 ns/op               0 B/op          0 allocs/op

BenchmarkSegmentReader_Next/uint64_crc32_0_KB-32                43697334                27.42 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uint64_crc32_1_KB-32                18371131                63.23 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uint64_crc32_2_KB-32                13305404                90.31 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uint64_crc32_4_KB-32                 8236240               145.8 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uint64_crc32_8_KB-32                 4665849               256.8 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uint64_crc32_16_KB-32                2507650               477.9 ns/op             0 B/op          0 allocs/op

BenchmarkSegmentReader_Next/uint64_crc64_0_KB-32                47914659                25.05 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uint64_crc64_1_KB-32                 3172028               378.3 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uint64_crc64_2_KB-32                 1649588               727.2 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uint64_crc64_4_KB-32                  761125              1433 ns/op               0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uint64_crc64_8_KB-32                  412868              2828 ns/op               0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uint64_crc64_16_KB-32                 205164              5768 ns/op               0 B/op          0 allocs/op

BenchmarkSegmentReader_Next/uvarint_crc32_0_KB-32               46763612                24.09 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uvarint_crc32_1_KB-32               20832184                57.29 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uvarint_crc32_2_KB-32               14046723                84.24 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uvarint_crc32_4_KB-32                8486071               140.1 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uvarint_crc32_8_KB-32                4777819               251.6 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uvarint_crc32_16_KB-32               2521687               475.5 ns/op             0 B/op          0 allocs/op

BenchmarkSegmentReader_Next/uvarint_crc64_0_KB-32               54617047                21.75 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uvarint_crc64_1_KB-32                3208736               374.2 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uvarint_crc64_2_KB-32                1661653               721.8 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uvarint_crc64_4_KB-32                 842624              1422 ns/op               0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uvarint_crc64_8_KB-32                 393436              2826 ns/op               0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uvarint_crc64_16_KB-32                207181              5790 ns/op               0 B/op          0 allocs/op

BenchmarkSegmentWriter_AppendEntry/uint16_crc32_0_KB-32         64297947                17.48 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint16_crc32_1_KB-32         22572460                52.39 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint16_crc32_2_KB-32         15049900                79.27 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint16_crc32_4_KB-32          8573284               139.8 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint16_crc32_8_KB-32          4638420               257.7 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint16_crc32_16_KB-32         2389704               501.9 ns/op             0 B/op          0 allocs/op

BenchmarkSegmentWriter_AppendEntry/uint16_crc64_0_KB-32         72588672                16.56 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint16_crc64_1_KB-32          3236221               370.7 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint16_crc64_2_KB-32          1665110               719.9 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint16_crc64_4_KB-32           838732              1430 ns/op               0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint16_crc64_8_KB-32           424620              2828 ns/op               0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint16_crc64_16_KB-32          208348              5751 ns/op               0 B/op          0 allocs/op

BenchmarkSegmentWriter_AppendEntry/uint32_crc32_0_KB-32         63029655                17.95 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint32_crc32_1_KB-32         21562850                55.39 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint32_crc32_2_KB-32         14708571                81.41 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint32_crc32_4_KB-32          8530479               140.1 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint32_crc32_8_KB-32          4710543               255.1 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint32_crc32_16_KB-32         2441170               490.7 ns/op             0 B/op          0 allocs/op

BenchmarkSegmentWriter_AppendEntry/uint32_crc64_0_KB-32         69494883                17.26 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint32_crc64_1_KB-32          3208680               374.1 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint32_crc64_2_KB-32          1661685               721.7 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint32_crc64_4_KB-32           839284              1429 ns/op               0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint32_crc64_8_KB-32           422901              2824 ns/op               0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint32_crc64_16_KB-32          209340              5728 ns/op               0 B/op          0 allocs/op

BenchmarkSegmentWriter_AppendEntry/uint64_crc32_0_KB-32         64535881                18.35 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint64_crc32_1_KB-32         19593387                61.31 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint64_crc32_2_KB-32         13404836                88.49 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint64_crc32_4_KB-32          8290268               145.4 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint64_crc32_8_KB-32          4646379               260.5 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint64_crc32_16_KB-32         2433942               492.4 ns/op             0 B/op          0 allocs/op

BenchmarkSegmentWriter_AppendEntry/uint64_crc64_0_KB-32         67880812                17.62 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint64_crc64_1_KB-32          3253912               368.2 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint64_crc64_2_KB-32          1655116               723.7 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint64_crc64_4_KB-32           831103              1434 ns/op               0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint64_crc64_8_KB-32           424234              2829 ns/op               0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint64_crc64_16_KB-32          208339              5741 ns/op               0 B/op          0 allocs/op

BenchmarkSegmentWriter_AppendEntry/uvarint_crc32_0_KB-32        69577647                17.24 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uvarint_crc32_1_KB-32        22806454                52.52 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uvarint_crc32_2_KB-32        15095666                79.47 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uvarint_crc32_4_KB-32         8607867               139.4 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uvarint_crc32_8_KB-32         4640164               257.7 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uvarint_crc32_16_KB-32        2383491               503.2 ns/op             0 B/op          0 allocs/op

BenchmarkSegmentWriter_AppendEntry/uvarint_crc64_0_KB-32        73572508                16.41 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uvarint_crc64_1_KB-32         3229232               371.1 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uvarint_crc64_2_KB-32         1664958               717.0 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uvarint_crc64_4_KB-32          839149              1431 ns/op               0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uvarint_crc64_8_KB-32          410739              2830 ns/op               0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uvarint_crc64_16_KB-32         208875              5742 ns/op               0 B/op          0 allocs/op
```

## WAL

The wal package is responsible for abstracting away multiple segment files behind a uniform interface.
The benchmark is running against a disk which influences performance.

```
goos: linux
goarch: amd64
pkg: github.com/backbone81/write-ahead-log/internal/wal
cpu: Intel(R) Core(TM) i9-14900K
BenchmarkReader_Next/uint32_crc32_0_KB-32        2619058               437.6 ns/op            14 B/op          0 allocs/op
BenchmarkReader_Next/uint32_crc32_1_KB-32        1648290               706.9 ns/op      1380.98 MB/s          14 B/op          0 allocs/op
BenchmarkReader_Next/uint32_crc32_2_KB-32        1552500               768.4 ns/op      2541.46 MB/s          14 B/op          0 allocs/op
BenchmarkReader_Next/uint32_crc32_4_KB-32        1417668               843.4 ns/op      4631.05 MB/s          23 B/op          0 allocs/op
BenchmarkReader_Next/uint32_crc32_8_KB-32        1000000              1058 ns/op        7383.70 MB/s          31 B/op          0 allocs/op
BenchmarkReader_Next/uint32_crc32_16_KB-32        822733              1459 ns/op        10712.52 MB/s         43 B/op          0 allocs/op

BenchmarkWriter_AppendEntry_Serial/uint32_crc32_periodic_0_KB-32                   15925             73026 ns/op               0 B/op          0 allocs/op
BenchmarkWriter_AppendEntry_Serial/uint32_crc32_periodic_1_KB-32                   15469             78973 ns/op          12.28 MB/s           0 B/op          0 allocs/op
BenchmarkWriter_AppendEntry_Serial/uint32_crc32_periodic_2_KB-32                   13120             92240 ns/op          20.66 MB/s           0 B/op          0 allocs/op
BenchmarkWriter_AppendEntry_Serial/uint32_crc32_periodic_4_KB-32                   14372             87425 ns/op          44.57 MB/s           1 B/op          0 allocs/op
BenchmarkWriter_AppendEntry_Serial/uint32_crc32_periodic_8_KB-32                   12844             88912 ns/op          87.57 MB/s           1 B/op          0 allocs/op
BenchmarkWriter_AppendEntry_Serial/uint32_crc32_periodic_16_KB-32                  10000            111485 ns/op         139.93 MB/s           5 B/op          0 allocs/op

BenchmarkWriter_AppendEntry_Serial/uint32_crc32_grouped_0_KB-32                      100          15952154 ns/op               4 B/op          0 allocs/op
BenchmarkWriter_AppendEntry_Serial/uint32_crc32_grouped_1_KB-32                      100          15602721 ns/op               2 B/op          0 allocs/op
BenchmarkWriter_AppendEntry_Serial/uint32_crc32_grouped_2_KB-32                      100          15470882 ns/op               7 B/op          0 allocs/op
BenchmarkWriter_AppendEntry_Serial/uint32_crc32_grouped_4_KB-32                      100          16319880 ns/op               0 B/op          0 allocs/op
BenchmarkWriter_AppendEntry_Serial/uint32_crc32_grouped_8_KB-32                      100          16543850 ns/op               0 B/op          0 allocs/op
BenchmarkWriter_AppendEntry_Serial/uint32_crc32_grouped_16_KB-32                     100          16226694 ns/op           0.62 MB/s          16 B/op          0 allocs/op

BenchmarkWriter_AppendEntry_Serial/uint32_crc32_none_0_KB-32                     3822621               283.7 ns/op             0 B/op          0 allocs/op
BenchmarkWriter_AppendEntry_Serial/uint32_crc32_none_1_KB-32                     1675518               689.5 ns/op      1416.21 MB/s           0 B/op          0 allocs/op
BenchmarkWriter_AppendEntry_Serial/uint32_crc32_none_2_KB-32                     1000000              1071 ns/op        1822.93 MB/s           0 B/op          0 allocs/op
BenchmarkWriter_AppendEntry_Serial/uint32_crc32_none_4_KB-32                     1000000              1815 ns/op        2152.10 MB/s           0 B/op          0 allocs/op
BenchmarkWriter_AppendEntry_Serial/uint32_crc32_none_8_KB-32                      340712              3486 ns/op        2240.70 MB/s           1 B/op          0 allocs/op
BenchmarkWriter_AppendEntry_Serial/uint32_crc32_none_16_KB-32                     180058              6531 ns/op        2392.11 MB/s           5 B/op          0 allocs/op

BenchmarkWriter_AppendEntry_Serial/uint32_crc32_immediate_0_KB-32                    184           6735988 ns/op               0 B/op          0 allocs/op
BenchmarkWriter_AppendEntry_Serial/uint32_crc32_immediate_1_KB-32                    176           6369914 ns/op               0 B/op          0 allocs/op
BenchmarkWriter_AppendEntry_Serial/uint32_crc32_immediate_2_KB-32                    228           5525752 ns/op               0 B/op          0 allocs/op
BenchmarkWriter_AppendEntry_Serial/uint32_crc32_immediate_4_KB-32                    182           6565794 ns/op               0 B/op          0 allocs/op
BenchmarkWriter_AppendEntry_Serial/uint32_crc32_immediate_8_KB-32                    176           6696168 ns/op           0.85 MB/s           0 B/op          0 allocs/op
BenchmarkWriter_AppendEntry_Serial/uint32_crc32_immediate_16_KB-32                   171           6847013 ns/op           1.71 MB/s           0 B/op          0 allocs/op

BenchmarkWriter_AppendEntry_Concurrently/uint32_crc32_none_0_KB-32               1000000              1067 ns/op             372 B/op          2 allocs/op
BenchmarkWriter_AppendEntry_Concurrently/uint32_crc32_none_1_KB-32               1000000              1451 ns/op         672.41 MB/s         244 B/op          1 allocs/op
BenchmarkWriter_AppendEntry_Concurrently/uint32_crc32_none_2_KB-32                581934              1774 ns/op        1100.18 MB/s         147 B/op          1 allocs/op
BenchmarkWriter_AppendEntry_Concurrently/uint32_crc32_none_4_KB-32                591079              2575 ns/op        1516.40 MB/s         150 B/op          1 allocs/op
BenchmarkWriter_AppendEntry_Concurrently/uint32_crc32_none_8_KB-32                302647              4191 ns/op        1863.71 MB/s         155 B/op          1 allocs/op
BenchmarkWriter_AppendEntry_Concurrently/uint32_crc32_none_16_KB-32               155649              7430 ns/op        2102.93 MB/s         160 B/op          1 allocs/op

BenchmarkWriter_AppendEntry_Concurrently/uint32_crc32_immediate_0_KB-32             9644            125048 ns/op             139 B/op          1 allocs/op
BenchmarkWriter_AppendEntry_Concurrently/uint32_crc32_immediate_1_KB-32            10195            109904 ns/op           8.03 MB/s         139 B/op          1 allocs/op
BenchmarkWriter_AppendEntry_Concurrently/uint32_crc32_immediate_2_KB-32            12246             90589 ns/op          20.73 MB/s         184 B/op          1 allocs/op
BenchmarkWriter_AppendEntry_Concurrently/uint32_crc32_immediate_4_KB-32            15260             85259 ns/op          45.35 MB/s         376 B/op          2 allocs/op
BenchmarkWriter_AppendEntry_Concurrently/uint32_crc32_immediate_8_KB-32            13036             77780 ns/op          99.61 MB/s         195 B/op          1 allocs/op
BenchmarkWriter_AppendEntry_Concurrently/uint32_crc32_immediate_16_KB-32            4970            268525 ns/op          57.70 MB/s        1101 B/op          3 allocs/op

BenchmarkWriter_AppendEntry_Concurrently/uint32_crc32_periodic_0_KB-32            148626             14138 ns/op             158 B/op          1 allocs/op
BenchmarkWriter_AppendEntry_Concurrently/uint32_crc32_periodic_1_KB-32             85177             14752 ns/op          66.06 MB/s         158 B/op          1 allocs/op
BenchmarkWriter_AppendEntry_Concurrently/uint32_crc32_periodic_2_KB-32             79036             16326 ns/op         119.35 MB/s         159 B/op          1 allocs/op
BenchmarkWriter_AppendEntry_Concurrently/uint32_crc32_periodic_4_KB-32             75775             17902 ns/op         217.47 MB/s         159 B/op          1 allocs/op
BenchmarkWriter_AppendEntry_Concurrently/uint32_crc32_periodic_8_KB-32             56041             22228 ns/op         350.82 MB/s         160 B/op          1 allocs/op
BenchmarkWriter_AppendEntry_Concurrently/uint32_crc32_periodic_16_KB-32            40548             31422 ns/op         496.82 MB/s         164 B/op          1 allocs/op

BenchmarkWriter_AppendEntry_Concurrently/uint32_crc32_grouped_0_KB-32            1096074              2473 ns/op             123 B/op          1 allocs/op
BenchmarkWriter_AppendEntry_Concurrently/uint32_crc32_grouped_1_KB-32             526350              2242 ns/op         435.49 MB/s         140 B/op          1 allocs/op
BenchmarkWriter_AppendEntry_Concurrently/uint32_crc32_grouped_2_KB-32             397734              2792 ns/op         698.71 MB/s         145 B/op          1 allocs/op
BenchmarkWriter_AppendEntry_Concurrently/uint32_crc32_grouped_4_KB-32             254162              5159 ns/op         756.62 MB/s         152 B/op          1 allocs/op
BenchmarkWriter_AppendEntry_Concurrently/uint32_crc32_grouped_8_KB-32             152569              8641 ns/op         903.42 MB/s         156 B/op          1 allocs/op
BenchmarkWriter_AppendEntry_Concurrently/uint32_crc32_grouped_16_KB-32             61689             16407 ns/op         951.47 MB/s         158 B/op          1 allocs/op
```
