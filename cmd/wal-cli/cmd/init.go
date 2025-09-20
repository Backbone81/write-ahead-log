package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/backbone81/write-ahead-log/pkg/wal"
)

var (
	initEntryLengthEncoding string
	initEntryChecksumType   string
)

// initCmd represents the init command.
var initCmd = &cobra.Command{
	Use:          "init",
	Short:        "Initializes a new write-ahead log.",
	Long:         `Initializes a new write-ahead log.`,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		segments, err := wal.GetSegments(directory)
		if err != nil {
			return err
		}
		if len(segments) != 0 {
			return fmt.Errorf("WAL already initialized at %q", directory)
		}

		var withEntryLengthEncoding wal.WriterOption
		switch initEntryLengthEncoding {
		case "uint16":
			withEntryLengthEncoding = wal.WithEntryLengthEncoding(wal.EntryLengthEncodingUint16)
		case "uint32":
			withEntryLengthEncoding = wal.WithEntryLengthEncoding(wal.EntryLengthEncodingUint32)
		case "uint64":
			withEntryLengthEncoding = wal.WithEntryLengthEncoding(wal.EntryLengthEncodingUint64)
		case "uvarint":
			withEntryLengthEncoding = wal.WithEntryLengthEncoding(wal.EntryLengthEncodingUvarint)
		default:
			return fmt.Errorf("unsupported entry length encoding %q", initEntryLengthEncoding)
		}

		var withEntryChecksumType wal.WriterOption
		switch initEntryChecksumType {
		case "crc32":
			withEntryChecksumType = wal.WithEntryChecksumType(wal.EntryChecksumTypeCrc32)
		case "crc64":
			withEntryChecksumType = wal.WithEntryChecksumType(wal.EntryChecksumTypeCrc64)
		default:
			return fmt.Errorf("unsupported entry checksum type %q", initEntryChecksumType)
		}

		if err := wal.Init(directory, withEntryLengthEncoding, withEntryChecksumType); err != nil {
			return err
		}
		fmt.Printf("WAL initialized at %q.\n", directory)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)

	initCmd.Flags().StringVarP(
		&initEntryLengthEncoding,
		"entry-length-encoding",
		"l",
		"uint32",
		"The entry length encoding to use. Valid values are uint16, uint32, uint64, uvarint.",
	)

	initCmd.Flags().StringVarP(
		&initEntryChecksumType,
		"entry-checksum-type",
		"c",
		"crc32",
		"The entry checksum type to use. Valid values are crc32, crc64.",
	)
}
