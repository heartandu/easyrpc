package main

import (
	"os"

	"github.com/heartandu/easyrpc/internal/app"
)

var version = "dev"

func main() {
	if err := app.NewApp(version).Run(); err != nil {
		os.Exit(1)
	}
}
