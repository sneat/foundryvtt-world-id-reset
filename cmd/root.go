package cmd

import (
	"os"

	"github.com/sneat/foundryvtt-world-id-reset/parser"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "foundryvtt-world-id-reset",
	Short: "Resets the IDs of all of the documents in a Foundry VTT world",
	Long: `Resets the IDs of all of the documents in a Foundry VTT world.

This is likely only something you will do if you have duplicated a world
to use as a base, and would like to be able to merge them afterwards.`,
	RunE: parser.Run,
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
	path, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Whether to output verbose logging.")
	rootCmd.PersistentFlags().StringVarP(&path, "path", "p", path, "The path to the \"worlds/\" folder you wish to parse. It should contain the \"world.json\" file. Defaults to the current directory.")

}
