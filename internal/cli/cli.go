package cli

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/pendig/kelompok/internal/config"
	"github.com/pendig/kelompok/internal/database"
	"github.com/pendig/kelompok/internal/httpapi"
	"github.com/pendig/kelompok/internal/members"
	"github.com/pendig/kelompok/internal/organizations"
	"github.com/pendig/kelompok/internal/seed"
	migrationfiles "github.com/pendig/kelompok/migrations"
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
	case "seed":
		return runSeed(ctx, args[1:], stdout)
	case "db":
		return runDB(ctx, args[1:], stdout)
	case "organization", "org":
		return runOrganization(ctx, args[1:], stdout, stderr)
	case "member":
		return runMember(ctx, args[1:], stdout, stderr)
	default:
		printHelp(stderr)
		return fmt.Errorf("unknown command: %s", args[0])
	}
}

func runOrganization(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		return errors.New("organization command requires a subcommand")
	}

	switch args[0] {
	case "list":
		flags := flag.NewFlagSet("organization list", flag.ContinueOnError)
		flags.SetOutput(stderr)
		limit := flags.Int("limit", 50, "maximum number of organizations")
		jsonOut := flags.Bool("json", false, "print JSON output")
		if err := flags.Parse(args[1:]); err != nil {
			return err
		}
		return withOrganizationRepository(ctx, func(repo *organizations.Repository) error {
			items, err := repo.ListPublic(ctx, *limit)
			if err != nil {
				return err
			}
			if *jsonOut {
				return writeJSONOutput(stdout, items)
			}
			for _, item := range items {
				fmt.Fprintf(stdout, "%s\t%s\t%s\n", item.Slug, item.Name, item.ClaimStatus)
			}
			return nil
		})
	case "create":
		flags := flag.NewFlagSet("organization create", flag.ContinueOnError)
		flags.SetOutput(stderr)
		input := organizations.AdminInput{}
		flags.StringVar(&input.Slug, "slug", "", "organization slug")
		flags.StringVar(&input.Name, "name", "", "organization name")
		flags.StringVar(&input.LegalName, "legal-name", "", "legal name")
		flags.StringVar(&input.Description, "description", "", "description")
		flags.StringVar(&input.Country, "country", "", "country")
		flags.StringVar(&input.Region, "region", "", "region")
		flags.StringVar(&input.City, "city", "", "city")
		flags.StringVar(&input.WebsiteURL, "website-url", "", "website URL")
		flags.StringVar(&input.OfficialEmail, "official-email", "", "official email")
		flags.StringVar(&input.ClaimStatus, "claim-status", "unclaimed", "claim status")
		jsonOut := flags.Bool("json", false, "print JSON output")
		if err := flags.Parse(args[1:]); err != nil {
			return err
		}
		return withOrganizationRepository(ctx, func(repo *organizations.Repository) error {
			item, err := repo.Create(ctx, input)
			if err != nil {
				return err
			}
			if *jsonOut {
				return writeJSONOutput(stdout, item)
			}
			fmt.Fprintf(stdout, "organization: created slug=%s name=%s\n", item.Slug, item.Name)
			return nil
		})
	default:
		return fmt.Errorf("unknown organization subcommand: %s", args[0])
	}
}

func runMember(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		return errors.New("member command requires a subcommand")
	}

	switch args[0] {
	case "list":
		flags := flag.NewFlagSet("member list", flag.ContinueOnError)
		flags.SetOutput(stderr)
		organizationSlug := flags.String("organization", "", "organization slug")
		limit := flags.Int("limit", 50, "maximum number of members")
		jsonOut := flags.Bool("json", false, "print JSON output")
		if err := flags.Parse(args[1:]); err != nil {
			return err
		}
		if strings.TrimSpace(*organizationSlug) == "" {
			return errors.New("member list requires --organization")
		}
		return withMemberRepository(ctx, func(repo *members.Repository) error {
			items, err := repo.ListByOrganizationSlug(ctx, *organizationSlug, *limit)
			if err != nil {
				return err
			}
			if *jsonOut {
				return writeJSONOutput(stdout, items)
			}
			for _, item := range items {
				fmt.Fprintf(stdout, "%s\t%s\t%s\n", item.ID, item.Name, item.Position)
			}
			return nil
		})
	case "create":
		flags := flag.NewFlagSet("member create", flag.ContinueOnError)
		flags.SetOutput(stderr)
		organizationSlug := flags.String("organization", "", "organization slug")
		input := members.Input{}
		flags.StringVar(&input.Name, "name", "", "member name")
		flags.StringVar(&input.Position, "position", "", "position")
		flags.StringVar(&input.Email, "email", "", "email")
		flags.StringVar(&input.Phone, "phone", "", "phone")
		flags.StringVar(&input.Bio, "bio", "", "bio")
		jsonOut := flags.Bool("json", false, "print JSON output")
		if err := flags.Parse(args[1:]); err != nil {
			return err
		}
		if strings.TrimSpace(*organizationSlug) == "" {
			return errors.New("member create requires --organization")
		}
		return withMemberRepository(ctx, func(repo *members.Repository) error {
			item, err := repo.Create(ctx, *organizationSlug, input)
			if err != nil {
				return err
			}
			if *jsonOut {
				return writeJSONOutput(stdout, item)
			}
			fmt.Fprintf(stdout, "member: created id=%s name=%s organization=%s\n", item.ID, item.Name, item.OrganizationSlug)
			return nil
		})
	default:
		return fmt.Errorf("unknown member subcommand: %s", args[0])
	}
}

func runSeed(ctx context.Context, args []string, stdout io.Writer) error {
	if len(args) == 0 {
		return errors.New("seed command requires a subcommand")
	}

	switch args[0] {
	case "demo":
		return seedDemo(ctx, stdout)
	default:
		return fmt.Errorf("unknown seed subcommand: %s", args[0])
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
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	pool, err := database.Open(ctx, cfg.DatabaseURL, poolSettings(cfg))
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
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	pool, err := database.Open(ctx, cfg.DatabaseURL, poolSettings(cfg))
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
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	pool, err := database.Open(ctx, cfg.DatabaseURL, poolSettings(cfg))
	if err != nil {
		return err
	}
	defer pool.Close()

	if err := database.Migrate(ctx, pool, migrationfiles.FS, "."); err != nil {
		return err
	}

	fmt.Fprintln(stdout, "migrations: ok")
	return nil
}

func seedDemo(ctx context.Context, stdout io.Writer) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	pool, err := database.Open(ctx, cfg.DatabaseURL, poolSettings(cfg))
	if err != nil {
		return err
	}
	defer pool.Close()

	result, err := seed.Demo(ctx, pool)
	if err != nil {
		return err
	}

	fmt.Fprintf(
		stdout,
		"demo seed: ok organization=%s posts=%d impact_reports=%d sdgs_signals=%d\n",
		result.OrganizationSlug,
		result.Posts,
		result.ImpactReports,
		result.SDGSSignals,
	)
	return nil
}

func withOrganizationRepository(ctx context.Context, fn func(*organizations.Repository) error) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	pool, err := database.Open(ctx, cfg.DatabaseURL, poolSettings(cfg))
	if err != nil {
		return err
	}
	defer pool.Close()
	return fn(organizations.NewRepository(pool))
}

func withMemberRepository(ctx context.Context, fn func(*members.Repository) error) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	pool, err := database.Open(ctx, cfg.DatabaseURL, poolSettings(cfg))
	if err != nil {
		return err
	}
	defer pool.Close()
	return fn(members.NewRepository(pool))
}

func writeJSONOutput(stdout io.Writer, value any) error {
	encoder := json.NewEncoder(stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}

func printHelp(stdout io.Writer) {
	fmt.Fprint(stdout, `Kelompok CLI

Usage:
  kelompok serve        Start the API server
  kelompok health       Check database connectivity
  kelompok migrate      Apply pending SQL migrations
  kelompok seed demo    Insert or update demo public MVP data
  kelompok db ping      Check database connectivity
  kelompok db migrate   Apply pending SQL migrations
  kelompok org list      List organizations
  kelompok org create    Create an organization
  kelompok member list   List organization members
  kelompok member create Create an organization member
  kelompok help         Show this help

Environment:
  KELOMPOK_ENV
  KELOMPOK_API_ADDR
  KELOMPOK_DATABASE_URL
  KELOMPOK_DB_MAX_CONNS
  KELOMPOK_DB_MIN_CONNS
  KELOMPOK_DB_MAX_CONN_LIFETIME
  KELOMPOK_DB_MAX_CONN_IDLE_TIME
  KELOMPOK_DB_HEALTH_CHECK_PERIOD
`)
}

func poolSettings(cfg config.Config) database.PoolSettings {
	return database.PoolSettings{
		MaxConns:          cfg.DatabaseMaxConns,
		MinConns:          cfg.DatabaseMinConns,
		MaxConnLifetime:   cfg.DatabaseMaxConnLifetime,
		MaxConnIdleTime:   cfg.DatabaseMaxConnIdleTime,
		HealthCheckPeriod: cfg.DatabaseHealthCheckPeriod,
	}
}
