// Package main is the application entrypoint.
package main

import (
	"fmt"

	"github.com/dingdayu/go-project-template/cmd"

	_ "github.com/joho/godotenv/autoload"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Recovered from panic: %v", r)
		}
	}()

	cmd.Execute()
}
