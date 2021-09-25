//Package cmd contains commands / sub commands for go-swagger-diff
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var debug bool

var rootCmd = &cobra.Command{
	Use:   "go-swagger-diff",
	Short: "Go Swagger Diff will let you know if your api is breaking",
}

//Execute is responsible for initiating the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "debug output")
	rootCmd.Run = rootCmd.HelpFunc()
}
