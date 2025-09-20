package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/backbone81/write-ahead-log/pkg/wal"
)

// describeCmd represents the describe command.
var describeCmd = &cobra.Command{
	Use:          "describe",
	Short:        "Provides detailed information about the write-ahead log.",
	Long:         `Provides detailed information about the write-ahead log.`,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		segments, err := wal.GetSegments(directory)
		if err != nil {
			return err
		}
		if len(segments) == 0 {
			return fmt.Errorf("no segment found in %q", directory)
		}

		reader, err := wal.NewReader(directory, segments[0])
		if err != nil {
			return err
		}
		defer func() {
			if err := reader.Close(); err != nil {
				fmt.Println(err)
			}
		}()

		filePath := ""
		for {
			if filePath != reader.FilePath() {
				filePath = reader.FilePath()
				fmt.Printf("Segment:               %s\n", reader.FilePath())
				fmt.Printf("Magic:                 %s\n", reader.Header().Magic)
				fmt.Printf("Version:               %d\n", reader.Header().Version)
				fmt.Printf("Entry Length Encoding: %s\n", reader.Header().EntryLengthEncoding)
				fmt.Printf("Entry Checksum Type:   %s\n", reader.Header().EntryChecksumType)
				fmt.Printf("First Sequence Number: %d\n", reader.Header().FirstSequenceNumber)
				fmt.Println()
			}

			if !reader.Next() {
				break
			}
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(describeCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// describeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// describeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
