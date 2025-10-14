// Package config loads application configuration and supports hot reload.
package config

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

// changeEventHandle contains registered config change handlers.
var (
	changeEventHandle []func(e fsnotify.Event)
	eventLock         sync.Mutex
	once              sync.Once
)

// CfgFile is the config file path, can be set before Init.
var CfgFile string

// Init initializes configuration from file and environment variables.
func Init() {
	once.Do(func() {
		// set config file name and search paths
		viper.SetConfigName("config")                 // name of config file (without extension)
		viper.SetConfigType("yaml")                   // REQUIRED if the config file does not have the extension in the name
		viper.AddConfigPath("/etc/singbox-adapter/")  // path to look for the config file in
		viper.AddConfigPath("$HOME/.singbox-adapter") // call multiple times to add many search paths
		viper.AddConfigPath(".")                      // optionally look for config in the working directory

		if len(CfgFile) > 0 {
			viper.SetConfigFile(CfgFile)
		}

		// read in matching environment variables
		viper.SetEnvPrefix("GO")
		viper.AutomaticEnv()

		if err := viper.BindEnv("app.service_name", "OTEL_SERVICE_NAME"); err != nil {
			fmt.Printf("\u001B[1;30;41m[error]\u001B[0m bind env app.service_name failed: %v\n", err)
		}
		if err := viper.BindEnv("app.port", "HTTP_PORT"); err != nil {
			fmt.Printf("\u001B[1;30;41m[error]\u001B[0m bind env app.port failed: %v\n", err)
		}
		if err := viper.BindEnv("app.environment", "ENVIRONMENT"); err != nil {
			fmt.Printf("\u001B[1;30;41m[error]\u001B[0m bind env app.environment failed: %v\n", err)
		}
		if err := viper.BindEnv("jwt.secret", "JWT_SECRET"); err != nil {
			fmt.Printf("\u001B[1;30;41m[error]\u001B[0m bind env jwt.secret failed: %v\n", err)
		}
		if err := viper.BindEnv("db", "DB"); err != nil {
			fmt.Printf("\u001B[1;30;41m[error]\u001B[0m bind env db failed: %v\n", err)
		}

		viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
		if err := viper.ReadInConfig(); err == nil {
			fmt.Printf("\033[1;30;42m[info]\033[0m using config file %s\n", viper.ConfigFileUsed())
		} else {
			fmt.Printf("\033[1;30;41m[error]\033[0m using config file error %s\n", err.Error())
			os.Exit(1)
		}

		// watch and handle config changes
		// viper.WatchConfig()
		// viper.OnConfigChange(onConfigChange)

		fmt.Printf("\033[1;30;42m[info]\033[0m config init %s\n", viper.ConfigFileUsed())
	})
}

// RegisterChangeEvent registers a config change event callback.
func RegisterChangeEvent(f func(e fsnotify.Event)) {
	eventLock.Lock()
	defer eventLock.Unlock()

	changeEventHandle = append(changeEventHandle, f)
}

// onConfigChange runs registered handlers for config changes.
//
//nolint:unused
func onConfigChange(e fsnotify.Event) {
	for _, f := range changeEventHandle {
		f(e)
	}
}
