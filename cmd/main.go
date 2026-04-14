// cmd/main.go
package main

import (
	"os"

	"github.com/largeoliu/redmine-cli/internal/app"
)

func main() {
	os.Exit(app.Execute())
}
