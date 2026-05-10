package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/Azero-NG/ibkr-flex/internal/cache"
	"github.com/Azero-NG/ibkr-flex/internal/config"
	"github.com/Azero-NG/ibkr-flex/internal/flex"
)

type globalFlags struct {
	account string
	format  string
	refresh bool
	since   string
	until   string
	verbose bool
}

var globals globalFlags

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "ibkr-flex",
		Short: "Read-only IBKR Flex Web Service CLI",
		Long: `ibkr-flex pulls a bundled Activity Flex Query from IBKR and slices it by
data dimension (trades, positions, cash, dividends, nav, mtm).
JSON output by default; --format=table or --format=csv for human views.

Required environment:
  IBKR_FLEX_TOKEN    16-char Flex Web Service token
  IBKR_FLEX_QUERY_ID Activity Flex Query ID

See docs/flex-setup.md for IBKR backend setup.`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	pf := root.PersistentFlags()
	pf.StringVar(&globals.account, "account", "", "filter by IBKR account ID (required for data commands)")
	pf.StringVar(&globals.format, "format", "json", "output format: json | table | csv")
	pf.BoolVar(&globals.refresh, "refresh", false, "skip cache and re-fetch from IBKR")
	pf.StringVar(&globals.since, "since", "", "filter records on or after this date (YYYY-MM-DD)")
	pf.StringVar(&globals.until, "until", "", "filter records on or before this date (YYYY-MM-DD)")
	pf.BoolVarP(&globals.verbose, "verbose", "v", false, "print fetch progress to stderr")

	root.AddCommand(
		newAccountsCmd(),
		newTradesCmd(),
		newPositionsCmd(),
		newCashCmd(),
		newDividendsCmd(),
		newNAVCmd(),
		newMTMCmd(),
		newPnLCmd(),
	)
	return root
}

func loadStatement(ctx context.Context) (*flex.Statement, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}
	client := flex.NewClient()
	if globals.verbose {
		client.OnRetry = func(elapsed time.Duration) {
			fmt.Fprintf(os.Stderr, "[ibkr-flex] generating, elapsed=%s\n", elapsed.Round(time.Second))
		}
	}
	xmlBytes, err := cache.Get(cfg.QueryID, globals.refresh, func() ([]byte, error) {
		if globals.verbose {
			fmt.Fprintln(os.Stderr, "[ibkr-flex] fetching from IBKR...")
		}
		return client.Fetch(ctx, cfg.Token, cfg.QueryID)
	})
	if err != nil {
		return nil, err
	}
	return flex.Parse(xmlBytes)
}

func requireAccount() error {
	if globals.account == "" {
		return errors.New("--account is required (run `ibkr-flex accounts` to list available)")
	}
	return nil
}

func dateInRange(d string) bool {
	// Compare only the date portion so "YYYY-MM-DDTHH:MM:SS" vs "YYYY-MM-DD" works.
	cmp := d
	if len(cmp) > 10 {
		cmp = cmp[:10]
	}
	if globals.since != "" && cmp < globals.since {
		return false
	}
	if globals.until != "" && cmp > globals.until {
		return false
	}
	return true
}

func exitCodeFor(err error) int {
	if err == nil {
		return 0
	}
	if errors.Is(err, config.ErrMissing) {
		return 1
	}
	if errors.Is(err, flex.ErrInvalidToken) || errors.Is(err, flex.ErrInvalidQueryID) {
		return 2
	}
	var fe *flex.FlexError
	if errors.As(err, &fe) {
		return 2
	}
	if errors.Is(err, flex.ErrStatementTimeout) {
		return 3
	}
	return 4
}
