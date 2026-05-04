package main

import (
	"os"

	"github.com/largeoliu/redmine-cli/internal/app"
)

var osExit = os.Exit

func run() int {
	return app.Execute()
}

func main() {
	osExit(run())
}
