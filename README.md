# Write Ahead Log

This library provides a write-ahead log. It can be useful for database systems, distributed systems or other situations
where a loss of in-memory data needs to be restored from stable storage. It was carefully engineered with performance
and zero memory allocations in mind.

It supports the following features:

- Transparent segmentation of the write-ahead log into multiple segment files with a configurable size for rollover.
- The writer is safe to use for concurrent writes without external synchronization.
- Several hash policies to configure the hash calculated for every entry.
  - Hash policy "crc32" provides a fast and simple checksum for small entry sizes.
  - Hash policy "crc64" provides more reliability for bigger entry sizes.
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

## TODOs

- Implement segment creation which works on windows and linux
- Introduce read buffer for better read performance from disk
- Provide a CLI for inspecting the WAL and to do maintenance

## Benchmarks

See the results of the benchmark suite and notice that there are no memory allocations across the board.

```
goos: linux
goarch: amd64
pkg: write-ahead-log/internal/wal
cpu: Intel(R) Core(TM) i9-14900K
BenchmarkEntryChecksumWriter/crc32_on_0_KB-32           197650897                5.798 ns/op           0 B/op          0 allocs/op
BenchmarkEntryChecksumWriter/crc32_on_1_KB-32           45509235                25.87 ns/op            0 B/op          0 allocs/op
BenchmarkEntryChecksumWriter/crc32_on_2_KB-32           23550248                49.83 ns/op            0 B/op          0 allocs/op
BenchmarkEntryChecksumWriter/crc32_on_4_KB-32           12670674                96.38 ns/op            0 B/op          0 allocs/op
BenchmarkEntryChecksumWriter/crc32_on_8_KB-32            6218697               189.1 ns/op             0 B/op          0 allocs/op
BenchmarkEntryChecksumWriter/crc32_on_16_KB-32           3198892               372.4 ns/op             0 B/op          0 allocs/op

BenchmarkEntryChecksumWriter/crc64_on_0_KB-32           245508477                4.844 ns/op           0 B/op          0 allocs/op
BenchmarkEntryChecksumWriter/crc64_on_1_KB-32            3489908               343.3 ns/op             0 B/op          0 allocs/op
BenchmarkEntryChecksumWriter/crc64_on_2_KB-32            1737457               685.1 ns/op             0 B/op          0 allocs/op
BenchmarkEntryChecksumWriter/crc64_on_4_KB-32             875552              1383 ns/op               0 B/op          0 allocs/op
BenchmarkEntryChecksumWriter/crc64_on_8_KB-32             433854              2739 ns/op               0 B/op          0 allocs/op
BenchmarkEntryChecksumWriter/crc64_on_16_KB-32            222968              5347 ns/op               0 B/op          0 allocs/op

BenchmarkEntryChecksumReader/crc32_on_0_KB-32           136545807                8.814 ns/op           0 B/op          0 allocs/op
BenchmarkEntryChecksumReader/crc32_on_1_KB-32           42273093                28.38 ns/op            0 B/op          0 allocs/op
BenchmarkEntryChecksumReader/crc32_on_2_KB-32           22696653                51.66 ns/op            0 B/op          0 allocs/op
BenchmarkEntryChecksumReader/crc32_on_4_KB-32           11995602                97.21 ns/op            0 B/op          0 allocs/op
BenchmarkEntryChecksumReader/crc32_on_8_KB-32            6231278               188.4 ns/op             0 B/op          0 allocs/op
BenchmarkEntryChecksumReader/crc32_on_16_KB-32           3217704               370.8 ns/op             0 B/op          0 allocs/op

BenchmarkEntryChecksumReader/crc64_on_0_KB-32           159712017                7.470 ns/op           0 B/op          0 allocs/op
BenchmarkEntryChecksumReader/crc64_on_1_KB-32            3460966               349.7 ns/op             0 B/op          0 allocs/op
BenchmarkEntryChecksumReader/crc64_on_2_KB-32            1745211               685.0 ns/op             0 B/op          0 allocs/op
BenchmarkEntryChecksumReader/crc64_on_4_KB-32             770263              1362 ns/op               0 B/op          0 allocs/op
BenchmarkEntryChecksumReader/crc64_on_8_KB-32             433927              2756 ns/op               0 B/op          0 allocs/op
BenchmarkEntryChecksumReader/crc64_on_16_KB-32            216703              5464 ns/op               0 B/op          0 allocs/op

BenchmarkEntryLengthWriter/uint16-32                    533813410                2.261 ns/op           0 B/op          0 allocs/op
BenchmarkEntryLengthWriter/uint32-32                    529550756                2.263 ns/op           0 B/op          0 allocs/op
BenchmarkEntryLengthWriter/uint64-32                    524395750                2.248 ns/op           0 B/op          0 allocs/op
BenchmarkEntryLengthWriter/uvarint-32                   418645252                2.910 ns/op           0 B/op          0 allocs/op

BenchmarkEntryLengthReader/uint16-32                    182216800                6.535 ns/op           0 B/op          0 allocs/op
BenchmarkEntryLengthReader/uint32-32                    224711335                5.374 ns/op           0 B/op          0 allocs/op
BenchmarkEntryLengthReader/uint64-32                    233841888                5.108 ns/op           0 B/op          0 allocs/op
BenchmarkEntryLengthReader/uvarint-32                   124727624                9.611 ns/op           0 B/op          0 allocs/op

BenchmarkWriteHeader-32                                 532468887                2.285 ns/op           0 B/op          0 allocs/op
BenchmarkReadHeader-32                                  154639329                7.740 ns/op           0 B/op          0 allocs/op

BenchmarkSegmentReader_Next/uint16_crc32_0_KB-32        50568260                23.56 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uint16_crc32_1_KB-32        23731480                50.68 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uint16_crc32_2_KB-32        15372093                77.81 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uint16_crc32_4_KB-32         9021279               132.8 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uint16_crc32_8_KB-32         4938338               242.6 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uint16_crc32_16_KB-32        2571808               463.1 ns/op             0 B/op          0 allocs/op

BenchmarkSegmentReader_Next/uint16_crc64_0_KB-32        55713228                20.68 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uint16_crc64_1_KB-32         3261276               362.4 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uint16_crc64_2_KB-32         1708870               696.7 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uint16_crc64_4_KB-32          833712              1401 ns/op               0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uint16_crc64_8_KB-32          422709              2804 ns/op               0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uint16_crc64_16_KB-32         208527              5728 ns/op               0 B/op          0 allocs/op

BenchmarkSegmentReader_Next/uint32_crc32_0_KB-32        45618482                24.83 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uint32_crc32_1_KB-32        23800244                50.33 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uint32_crc32_2_KB-32        14691026                77.58 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uint32_crc32_4_KB-32         8896959               134.5 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uint32_crc32_8_KB-32         4895478               244.2 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uint32_crc32_16_KB-32        2564528               465.1 ns/op             0 B/op          0 allocs/op

BenchmarkSegmentReader_Next/uint32_crc64_0_KB-32        55588066                21.91 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uint32_crc64_1_KB-32         3237874               364.9 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uint32_crc64_2_KB-32         1672164               718.3 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uint32_crc64_4_KB-32          837904              1413 ns/op               0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uint32_crc64_8_KB-32          425424              2814 ns/op               0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uint32_crc64_16_KB-32         207166              5761 ns/op               0 B/op          0 allocs/op

BenchmarkSegmentReader_Next/uint64_crc32_0_KB-32        48171414                24.84 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uint64_crc32_1_KB-32        21436998                54.86 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uint64_crc32_2_KB-32        14404341                81.84 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uint64_crc32_4_KB-32         8601595               138.0 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uint64_crc32_8_KB-32         4751832               250.1 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uint64_crc32_16_KB-32        2533447               471.1 ns/op             0 B/op          0 allocs/op

BenchmarkSegmentReader_Next/uint64_crc64_0_KB-32        52987934                22.28 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uint64_crc64_1_KB-32         3232362               363.5 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uint64_crc64_2_KB-32         1671200               715.5 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uint64_crc64_4_KB-32          855993              1413 ns/op               0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uint64_crc64_8_KB-32          418981              2817 ns/op               0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uint64_crc64_16_KB-32         208161              5719 ns/op               0 B/op          0 allocs/op

BenchmarkSegmentReader_Next/uvarint_crc32_0_KB-32       50783358                22.47 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uvarint_crc32_1_KB-32       22985659                52.10 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uvarint_crc32_2_KB-32       15071208                78.60 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uvarint_crc32_4_KB-32        8951845               134.0 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uvarint_crc32_8_KB-32        4901199               243.5 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uvarint_crc32_16_KB-32       2551641               465.1 ns/op             0 B/op          0 allocs/op

BenchmarkSegmentReader_Next/uvarint_crc64_0_KB-32       61744434                18.86 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uvarint_crc64_1_KB-32        3291063               364.4 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uvarint_crc64_2_KB-32        1696827               701.9 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uvarint_crc64_4_KB-32         838482              1396 ns/op               0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uvarint_crc64_8_KB-32         419983              2803 ns/op               0 B/op          0 allocs/op
BenchmarkSegmentReader_Next/uvarint_crc64_16_KB-32        207486              5775 ns/op               0 B/op          0 allocs/op

BenchmarkSegmentWriter_AppendEntry/uint16_crc32_none_0_KB-32            83060565                14.77 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint16_crc32_none_1_KB-32            26355826                45.84 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint16_crc32_none_2_KB-32            16059964                73.86 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint16_crc32_none_4_KB-32             8818561               135.4 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint16_crc32_none_8_KB-32             4720334               254.1 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint16_crc32_none_16_KB-32            2448052               494.0 ns/op             0 B/op          0 allocs/op

BenchmarkSegmentWriter_AppendEntry/uint16_crc32_immediate_0_KB-32       78404980                15.53 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint16_crc32_immediate_1_KB-32       25897888                45.67 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint16_crc32_immediate_2_KB-32       15987676                75.08 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint16_crc32_immediate_4_KB-32        8824845               134.5 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint16_crc32_immediate_8_KB-32        4709973               251.1 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint16_crc32_immediate_16_KB-32       2452797               487.9 ns/op             0 B/op          0 allocs/op

BenchmarkSegmentWriter_AppendEntry/uint16_crc32_periodic_0_KB-32        83366390                15.11 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint16_crc32_periodic_1_KB-32        24764890                45.75 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint16_crc32_periodic_2_KB-32        15918196                74.96 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint16_crc32_periodic_4_KB-32         8807365               135.6 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint16_crc32_periodic_8_KB-32         4681340               254.4 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint16_crc32_periodic_16_KB-32        2375294               503.9 ns/op             0 B/op          0 allocs/op

BenchmarkSegmentWriter_AppendEntry/uint16_crc32_grouped_0_KB-32             1080           1136857 ns/op               0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint16_crc32_grouped_1_KB-32             1041           1163923 ns/op               0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint16_crc32_grouped_2_KB-32             1051           1159651 ns/op               0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint16_crc32_grouped_4_KB-32             1057           1138316 ns/op               7 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint16_crc32_grouped_8_KB-32             1078           1145164 ns/op               8 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint16_crc32_grouped_16_KB-32            1032           1163571 ns/op              17 B/op          0 allocs/op

BenchmarkSegmentWriter_AppendEntry/uint16_crc64_none_0_KB-32            92233732                12.61 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint16_crc64_none_1_KB-32             3284920               364.9 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint16_crc64_none_2_KB-32             1668255               719.0 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint16_crc64_none_4_KB-32              839739              1428 ns/op               0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint16_crc64_none_8_KB-32              409359              2832 ns/op               0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint16_crc64_none_16_KB-32             208890              5742 ns/op               0 B/op          0 allocs/op

BenchmarkSegmentWriter_AppendEntry/uint16_crc64_immediate_0_KB-32       90685084                13.09 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint16_crc64_immediate_1_KB-32        3281059               367.1 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint16_crc64_immediate_2_KB-32        1658491               721.5 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint16_crc64_immediate_4_KB-32         839215              1429 ns/op               0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint16_crc64_immediate_8_KB-32         412508              2831 ns/op               0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint16_crc64_immediate_16_KB-32        209334              5757 ns/op               0 B/op          0 allocs/op

BenchmarkSegmentWriter_AppendEntry/uint16_crc64_periodic_0_KB-32        89556381                13.38 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint16_crc64_periodic_1_KB-32         3276560               366.1 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint16_crc64_periodic_2_KB-32         1667719               719.1 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint16_crc64_periodic_4_KB-32          838134              1437 ns/op               0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint16_crc64_periodic_8_KB-32          414399              2874 ns/op               0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint16_crc64_periodic_16_KB-32         208064              5735 ns/op               0 B/op          0 allocs/op

BenchmarkSegmentWriter_AppendEntry/uint16_crc64_grouped_0_KB-32             1059           1134700 ns/op               0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint16_crc64_grouped_1_KB-32             1059           1143220 ns/op               0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint16_crc64_grouped_2_KB-32             1044           1157050 ns/op               0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint16_crc64_grouped_4_KB-32             1062           1159979 ns/op               7 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint16_crc64_grouped_8_KB-32             1054           1146659 ns/op               9 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint16_crc64_grouped_16_KB-32             976           1167245 ns/op              27 B/op          0 allocs/op

BenchmarkSegmentWriter_AppendEntry/uint32_crc32_none_0_KB-32            70592764                16.45 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint32_crc32_none_1_KB-32            23747880                48.55 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint32_crc32_none_2_KB-32            14846974                75.63 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint32_crc32_none_4_KB-32             8053273               141.3 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint32_crc32_none_8_KB-32             4804032               248.4 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint32_crc32_none_16_KB-32            2491477               483.1 ns/op             0 B/op          0 allocs/op

BenchmarkSegmentWriter_AppendEntry/uint32_crc32_immediate_0_KB-32       70531274                16.91 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint32_crc32_immediate_1_KB-32       24380190                48.82 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint32_crc32_immediate_2_KB-32       15839584                76.71 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint32_crc32_immediate_4_KB-32        8906503               134.8 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint32_crc32_immediate_8_KB-32        4778172               251.4 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint32_crc32_immediate_16_KB-32       2433014               492.2 ns/op             0 B/op          0 allocs/op

BenchmarkSegmentWriter_AppendEntry/uint32_crc32_periodic_0_KB-32        71987196                16.18 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint32_crc32_periodic_1_KB-32        24614028                48.87 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint32_crc32_periodic_2_KB-32        15769934                75.94 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint32_crc32_periodic_4_KB-32         8830735               135.4 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint32_crc32_periodic_8_KB-32         4790215               250.5 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint32_crc32_periodic_16_KB-32        2483241               483.9 ns/op             0 B/op          0 allocs/op

BenchmarkSegmentWriter_AppendEntry/uint32_crc32_grouped_0_KB-32             1074           1148408 ns/op               0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint32_crc32_grouped_1_KB-32             1017           1162376 ns/op               0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint32_crc32_grouped_2_KB-32             1040           1166312 ns/op               0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint32_crc32_grouped_4_KB-32             1004           1182434 ns/op               8 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint32_crc32_grouped_8_KB-32             1018           1168204 ns/op               9 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint32_crc32_grouped_16_KB-32            1030           1157718 ns/op              17 B/op          0 allocs/op

BenchmarkSegmentWriter_AppendEntry/uint32_crc64_none_0_KB-32            85141328                13.81 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint32_crc64_none_1_KB-32             3273315               366.4 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint32_crc64_none_2_KB-32             1667647               719.6 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint32_crc64_none_4_KB-32              830680              1427 ns/op               0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint32_crc64_none_8_KB-32              422390              2825 ns/op               0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint32_crc64_none_16_KB-32             209006              5734 ns/op               0 B/op          0 allocs/op

BenchmarkSegmentWriter_AppendEntry/uint32_crc64_immediate_0_KB-32       84892717                14.23 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint32_crc64_immediate_1_KB-32        3234577               368.7 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint32_crc64_immediate_2_KB-32        1667047               719.2 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint32_crc64_immediate_4_KB-32         839608              1432 ns/op               0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint32_crc64_immediate_8_KB-32         424220              2829 ns/op               0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint32_crc64_immediate_16_KB-32        205821              5745 ns/op               0 B/op          0 allocs/op

BenchmarkSegmentWriter_AppendEntry/uint32_crc64_periodic_0_KB-32        80034387                14.37 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint32_crc64_periodic_1_KB-32         3250376               367.7 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint32_crc64_periodic_2_KB-32         1668445               718.7 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint32_crc64_periodic_4_KB-32          840634              1425 ns/op               0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint32_crc64_periodic_8_KB-32          424158              2828 ns/op               0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint32_crc64_periodic_16_KB-32         206160              5753 ns/op               0 B/op          0 allocs/op

BenchmarkSegmentWriter_AppendEntry/uint32_crc64_grouped_0_KB-32             1076           1159132 ns/op               0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint32_crc64_grouped_1_KB-32             1026           1165536 ns/op               0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint32_crc64_grouped_2_KB-32             1016           1185771 ns/op               0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint32_crc64_grouped_4_KB-32             1017           1180560 ns/op              12 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint32_crc64_grouped_8_KB-32             1028           1163095 ns/op              14 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint32_crc64_grouped_16_KB-32            1000           1197106 ns/op              24 B/op          0 allocs/op

BenchmarkSegmentWriter_AppendEntry/uint64_crc32_none_0_KB-32            71752882                16.47 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint64_crc32_none_1_KB-32            20843886                54.16 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint64_crc32_none_2_KB-32            14838645                81.33 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint64_crc32_none_4_KB-32             8662338               138.7 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint64_crc32_none_8_KB-32             4734474               253.6 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint64_crc32_none_16_KB-32            2454456               487.6 ns/op             0 B/op          0 allocs/op

BenchmarkSegmentWriter_AppendEntry/uint64_crc32_immediate_0_KB-32       69557361                17.09 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint64_crc32_immediate_1_KB-32       21529057                55.00 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint64_crc32_immediate_2_KB-32       14715685                81.73 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint64_crc32_immediate_4_KB-32        8504695               139.2 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint64_crc32_immediate_8_KB-32        4723768               254.2 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint64_crc32_immediate_16_KB-32       2411980               495.2 ns/op             0 B/op          0 allocs/op

BenchmarkSegmentWriter_AppendEntry/uint64_crc32_periodic_0_KB-32        71784154                16.41 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint64_crc32_periodic_1_KB-32        21806209                54.41 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint64_crc32_periodic_2_KB-32        14700078                81.55 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint64_crc32_periodic_4_KB-32         8581093               139.6 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint64_crc32_periodic_8_KB-32         4727545               253.9 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint64_crc32_periodic_16_KB-32        2413800               496.0 ns/op             0 B/op          0 allocs/op

BenchmarkSegmentWriter_AppendEntry/uint64_crc32_grouped_0_KB-32             1050           1190239 ns/op               0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint64_crc32_grouped_1_KB-32              986           1201306 ns/op               0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint64_crc32_grouped_2_KB-32             1014           1189746 ns/op               0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint64_crc32_grouped_4_KB-32              987           1199500 ns/op               8 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint64_crc32_grouped_8_KB-32             1010           1197855 ns/op               9 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint64_crc32_grouped_16_KB-32            1006           1182343 ns/op              18 B/op          0 allocs/op

BenchmarkSegmentWriter_AppendEntry/uint64_crc64_none_0_KB-32            76756837                14.68 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint64_crc64_none_1_KB-32             3276566               367.2 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint64_crc64_none_2_KB-32             1673880               716.8 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint64_crc64_none_4_KB-32              837943              1428 ns/op               0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint64_crc64_none_8_KB-32              421998              2827 ns/op               0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint64_crc64_none_16_KB-32             209656              5723 ns/op               0 B/op          0 allocs/op

BenchmarkSegmentWriter_AppendEntry/uint64_crc64_immediate_0_KB-32       78346910                15.13 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint64_crc64_immediate_1_KB-32        3268696               366.9 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint64_crc64_immediate_2_KB-32        1671747               717.7 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint64_crc64_immediate_4_KB-32         839372              1428 ns/op               0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint64_crc64_immediate_8_KB-32         418128              2860 ns/op               0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint64_crc64_immediate_16_KB-32        196731              5768 ns/op               0 B/op          0 allocs/op

BenchmarkSegmentWriter_AppendEntry/uint64_crc64_periodic_0_KB-32        75879294                15.37 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint64_crc64_periodic_1_KB-32         3206901               369.1 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint64_crc64_periodic_2_KB-32         1672117               717.3 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint64_crc64_periodic_4_KB-32          839570              1428 ns/op               0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint64_crc64_periodic_8_KB-32          423870              2829 ns/op               0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint64_crc64_periodic_16_KB-32         208473              5759 ns/op               0 B/op          0 allocs/op

BenchmarkSegmentWriter_AppendEntry/uint64_crc64_grouped_0_KB-32             1027           1176019 ns/op               0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint64_crc64_grouped_1_KB-32              998           1191122 ns/op               1 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint64_crc64_grouped_2_KB-32             1006           1164526 ns/op               4 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint64_crc64_grouped_4_KB-32             1012           1166205 ns/op               8 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint64_crc64_grouped_8_KB-32             1039           1177512 ns/op              10 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uint64_crc64_grouped_16_KB-32            1010           1186412 ns/op              21 B/op          0 allocs/op

BenchmarkSegmentWriter_AppendEntry/uvarint_crc32_none_0_KB-32           80451661                14.89 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uvarint_crc32_none_1_KB-32           23962300                45.61 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uvarint_crc32_none_2_KB-32           14718678                75.20 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uvarint_crc32_none_4_KB-32            8826146               136.1 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uvarint_crc32_none_8_KB-32            4713891               254.9 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uvarint_crc32_none_16_KB-32           2397129               501.7 ns/op             0 B/op          0 allocs/op

BenchmarkSegmentWriter_AppendEntry/uvarint_crc32_immediate_0_KB-32      83618378                14.92 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uvarint_crc32_immediate_1_KB-32      25344649                46.78 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uvarint_crc32_immediate_2_KB-32      15745483                75.94 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uvarint_crc32_immediate_4_KB-32       8194045               144.9 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uvarint_crc32_immediate_8_KB-32       4686007               255.5 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uvarint_crc32_immediate_16_KB-32      2383507               502.4 ns/op             0 B/op          0 allocs/op

BenchmarkSegmentWriter_AppendEntry/uvarint_crc32_periodic_0_KB-32       77023804                14.66 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uvarint_crc32_periodic_1_KB-32       26001138                46.23 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uvarint_crc32_periodic_2_KB-32       15580549                75.83 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uvarint_crc32_periodic_4_KB-32        8674860               139.0 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uvarint_crc32_periodic_8_KB-32        4676968               256.3 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uvarint_crc32_periodic_16_KB-32       2393383               501.3 ns/op             0 B/op          0 allocs/op

BenchmarkSegmentWriter_AppendEntry/uvarint_crc32_grouped_0_KB-32            1042           1173468 ns/op               0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uvarint_crc32_grouped_1_KB-32            1036           1182938 ns/op               0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uvarint_crc32_grouped_2_KB-32            1002           1190891 ns/op               0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uvarint_crc32_grouped_4_KB-32             978           1196690 ns/op               8 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uvarint_crc32_grouped_8_KB-32            1017           1196403 ns/op               9 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uvarint_crc32_grouped_16_KB-32            999           1199073 ns/op              19 B/op          0 allocs/op

BenchmarkSegmentWriter_AppendEntry/uvarint_crc64_none_0_KB-32           93809200                12.21 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uvarint_crc64_none_1_KB-32            3278786               365.4 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uvarint_crc64_none_2_KB-32            1669815               718.4 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uvarint_crc64_none_4_KB-32             838837              1444 ns/op               0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uvarint_crc64_none_8_KB-32             409252              2835 ns/op               0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uvarint_crc64_none_16_KB-32            192200              5769 ns/op               0 B/op          0 allocs/op

BenchmarkSegmentWriter_AppendEntry/uvarint_crc64_immediate_0_KB-32      95898638                12.68 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uvarint_crc64_immediate_1_KB-32       3278607               365.6 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uvarint_crc64_immediate_2_KB-32       1663624               720.0 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uvarint_crc64_immediate_4_KB-32        838652              1434 ns/op               0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uvarint_crc64_immediate_8_KB-32        422770              2842 ns/op               0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uvarint_crc64_immediate_16_KB-32       208128              5749 ns/op               0 B/op          0 allocs/op

BenchmarkSegmentWriter_AppendEntry/uvarint_crc64_periodic_0_KB-32       79611337                12.78 ns/op            0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uvarint_crc64_periodic_1_KB-32        3271425               366.2 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uvarint_crc64_periodic_2_KB-32        1667041               719.8 ns/op             0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uvarint_crc64_periodic_4_KB-32         836542              1432 ns/op               0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uvarint_crc64_periodic_8_KB-32         375048              2839 ns/op               0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uvarint_crc64_periodic_16_KB-32        209146              5746 ns/op               0 B/op          0 allocs/op

BenchmarkSegmentWriter_AppendEntry/uvarint_crc64_grouped_0_KB-32            1039           1172373 ns/op               0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uvarint_crc64_grouped_1_KB-32            1010           1174378 ns/op               0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uvarint_crc64_grouped_2_KB-32            1038           1190186 ns/op               0 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uvarint_crc64_grouped_4_KB-32            1009           1188392 ns/op               8 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uvarint_crc64_grouped_8_KB-32             993           1196691 ns/op              15 B/op          0 allocs/op
BenchmarkSegmentWriter_AppendEntry/uvarint_crc64_grouped_16_KB-32           1026           1190667 ns/op              20 B/op          0 allocs/op
```
