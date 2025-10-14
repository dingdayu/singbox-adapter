package cmd

import (
	"fmt"
	"os"

	"github.com/dingdayu/go-project-template/model/entity"
	"github.com/dingdayu/go-project-template/pkg/config"
	"github.com/dingdayu/go-project-template/pkg/logger"
	"github.com/dingdayu/go-project-template/pkg/otel"

	"github.com/common-nighthawk/go-figure"
	"github.com/spf13/cobra"
)

var appName = otel.GetServiceName()

// RootCmd RootCmd
var rootCmd = &cobra.Command{
	Use:              appName,
	Short:            appName + " command.",
	Long:             appName + " command.",
	TraverseChildren: true,
}

func init() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Recovered from panic: %v", r)
		}
	}()

	myFigure := figure.NewFigure(appName, "slant", true)
	myFigure.Print()

	fmt.Printf("Version: %s\n", entity.BuildVersion)
	fmt.Printf("BuildTime: %s\n", entity.BuildTime)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().StringVar(&config.CfgFile, "c", "config.yaml", "config file (default is config.yaml)")

	cobra.OnInitialize(onInitialize)
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println("rootCmd.Execute error:", err)
		os.Exit(1)
	}
}

func onInitialize() {
	//
	config.Init()
	logger.Init()
}
