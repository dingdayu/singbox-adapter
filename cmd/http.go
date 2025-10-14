// Package cmd wires Cobra commands for the application.
package cmd

//revive:disable:unused-parameter

import (
	"github.com/dingdayu/go-project-template/api"
	"github.com/dingdayu/go-project-template/model/dao"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	httpAsync bool
	httpCron  bool
)

var httpCmd = &cobra.Command{
	Use:   "http",
	Short: "Start Singbox Adapter http server",
	Long:  "Start Singbox Adapter http server",
	Args: func(cmd *cobra.Command, args []string) error {
		return nil
	},
	PreRun: func(cmd *cobra.Command, args []string) {
		// redis.Init()
		if viper.GetString("db") != "" {
			dao.Setup()
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		if httpAsync {
			go api.AsyncRun(cmd.Context())
		}
		if httpCron {
			go api.CronRun(cmd.Context())
		}

		api.Run(cmd.Context())
	},
}

func init() {
	rootCmd.AddCommand(httpCmd)

	// add --async flag to control whether to start async processing
	httpCmd.Flags().BoolVar(&httpAsync, "async", false, "Start async processing at server start")
	_ = viper.BindPFlag("http.async", httpCmd.Flags().Lookup("async"))

	httpCmd.Flags().BoolVar(&httpCron, "cron", false, "Start cron processing at server start")
	_ = viper.BindPFlag("http.cron", httpCmd.Flags().Lookup("cron"))
}
