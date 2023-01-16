/*
Copyright © 2023 Blair McMillan

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
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