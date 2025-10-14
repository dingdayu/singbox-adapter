// Package cmd wires Cobra commands for the application.
package cmd

//revive:disable:unused-parameter

import (
	"fmt"

	"github.com/dingdayu/go-project-template/model/entity"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the app version information",
	Long:  "Print the app version information for the current command.",
	Args: func(cmd *cobra.Command, args []string) error {
		return nil
	},
	PreRun: func(cmd *cobra.Command, args []string) {
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Version: %s\n", entity.BuildVersion)
		fmt.Printf("BuildTime: %s\n", entity.BuildTime)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
