package main

import (
	"context"
	"fmt"
	"os"

	"github.com/pendig/kelompok/internal/cli"
)

func main() {
	if err := cli.Run(context.Background(), []string{"serve"}, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
