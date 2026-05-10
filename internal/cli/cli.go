package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/pendig/kelompok/internal/config"
	"github.com/pendig/kelompok/internal/database"
	"github.com/pendig/kelompok/internal/httpapi"
)

func Run(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 || args[0] == "help" || args[0] == "--help" || args[0] == "-h" {
		printHelp(stdout)
		return nil
	}

	switch args[0] {
	case "serve":
		return serve(ctx, stdout)
	case "health":
		return health(ctx, stdout)
	case "migrate":
		return migrate(ctx, stdout)
	case "db":
		return runDB(ctx, args[1:], stdout)
	default:
		fmt.Fprintf(stderr, "unknown command: %s\n\n", args[0])
		printHelp(stderr)
		return errors.New("unknown command")
	}
}

func runDB(ctx context.Context, args []string, stdout io.Writer) error {
	if len(args) == 0 {
		return errors.New("db command requires a subcommand")
	}

	switch args[0] {
	case "ping":
		return health(ctx, stdout)
	case "migrate":
		return migrate(ctx, stdout)
	default:
		return fmt.Errorf("unknown db subcommand: %s", args[0])
	}
}

func serve(ctx context.Context, stdout io.Writer) error {
	cfg := config.Load()
	pool, err := database.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	server := httpapi.New(cfg, pool).HTTPServer()
	errs := make(chan error, 1)

	go func() {
		fmt.Fprintf(stdout, "kelompok-api listening on %s\n", cfg.APIAddr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errs <- err
			return
		}
		errs <- nil
	}()

	stopCtx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	select {
	case <-stopCtx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return server.Shutdown(shutdownCtx)
	case err := <-errs:
		return err
	}
}

func health(ctx context.Context, stdout io.Writer) error {
	cfg := config.Load()
	pool, err := database.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := database.Ping(pingCtx, pool); err != nil {
		return err
	}

	fmt.Fprintln(stdout, "database: ok")
	return nil
}

func migrate(ctx context.Context, stdout io.Writer) error {
	cfg := config.Load()
	pool, err := database.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	if err := database.Migrate(ctx, pool, "migrations"); err != nil {
		return err
	}

	fmt.Fprintln(stdout, "migrations: ok")
	return nil
}

func printHelp(stdout io.Writer) {
	fmt.Fprint(stdout, `Kelompok CLI

Usage:
  kelompok serve        Start the API server
  kelompok health       Check database connectivity
  kelompok migrate      Apply pending SQL migrations
  kelompok db ping      Check database connectivity
  kelompok db migrate   Apply pending SQL migrations
  kelompok help         Show this help

Environment:
  KELOMPOK_ENV
  KELOMPOK_API_ADDR
  KELOMPOK_DATABASE_URL
`)
}
