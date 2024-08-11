package main

import (
	"os"

	"github.com/heartandu/easyrpc/pkg/app"
)

func main() {
	if err := app.NewApp().Run(); err != nil {
		os.Exit(1)
	}
}
