package main

import (
	"context"
	"fmt"
	"os"

	"github.com/pendig/kelompok/internal/cli"
)

func main() {
	ctx := context.Background()
	if err := cli.Run(ctx, os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
