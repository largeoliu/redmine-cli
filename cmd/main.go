package main

import (
	"os"

	"github.com/largeoliu/redmine-cli/internal/app"
)

func run() int {
	return app.Execute()
}

func main() {
	os.Exit(run())
}
