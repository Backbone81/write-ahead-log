package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var directory string

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "wal-cli",
	Short: "A tool for interacting with write-ahead logs.",
	Long:  `A tool for interacting with write-ahead logs.`,
	// RunE: func(cmd *cobra.Command, args []string) error {
	//	return nil
	// },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(
		&directory,
		"directory",
		"d",
		".",
		"The directory the write-ahead log is located in.",
	)
}
