package main

import (
	"github.com/spf13/cobra"

	"github.com/azero/ibkr-flex/internal/flex"
)

func newDividendsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "dividends",
		Short: "List dividend accruals and payments",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := requireAccount(); err != nil {
				return err
			}
			stmt, err := loadStatement(cmd.Context())
			if err != nil {
				return err
			}
			out := make([]flex.Dividend, 0, len(stmt.Dividends))
			for _, d := range stmt.Dividends {
				if d.Account != globals.account {
					continue
				}
				if !dateInRange(d.Date) {
					continue
				}
				out = append(out, d)
			}
			return renderDividends(out)
		},
	}
}
