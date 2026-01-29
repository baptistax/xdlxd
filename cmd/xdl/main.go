package main

import (
	"fmt"
	"os"

	"github.com/yourname/xdl2/internal/app"
)

func main() {
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
	os.Exit(0)
}
