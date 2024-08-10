package main

import (
	"fmt"
	"os"

	"github.com/heartandu/easyrpc/pkg/app"
)

func main() {
	if err := app.NewApp().Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}
