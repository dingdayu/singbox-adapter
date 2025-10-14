// Package cmd wires Cobra commands for the application.
package cmd

//revive:disable:unused-parameter

import (
	"fmt"

	"github.com/dingdayu/go-project-template/model/dao"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Run database migrations",
	Long:  "Run database migrations for the current environment.",
	Args: func(cmd *cobra.Command, args []string) error {
		return nil
	},
	PreRun: func(cmd *cobra.Command, args []string) {
		if viper.GetString("db") == "" {
			fmt.Println("❌ Database DSN is not configured. Please set it in your config file or via the DB_DSN environment variable.")
			return
		}
		dao.Setup()
	},
	Run: func(cmd *cobra.Command, args []string) {
		// AutoMigrate will create tables, missing foreign keys, constraints, columns and indexes.
		// It will change existing column's type if it's size, precision, nullable changed.
		// AutoMigrate will not delete unused columns to protect your data.
		// 参考：https://gorm.io/docs/migration.html#Auto-Migration

		err := dao.GetContextDB(cmd.Context()).AutoMigrate(dao.User{})
		if err != nil {
			fmt.Printf("❌ Database migration failed: %v\n", err)
			return
		}

		fmt.Println("✅ Database migration completed successfully!")
	},
}

func init() {
	rootCmd.AddCommand(migrateCmd)
}
