package main

import (
	"github.com/spf13/cobra"

	"github.com/Azero-NG/ibkr-flex/internal/flex"
)

func newTradesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "trades",
		Short: "List trade executions",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := requireAccount(); err != nil {
				return err
			}
			stmt, err := loadStatement(cmd.Context())
			if err != nil {
				return err
			}
			out := make([]flex.Trade, 0, len(stmt.Trades))
			for _, t := range stmt.Trades {
				if t.Account != globals.account {
					continue
				}
				if !dateInRange(t.TradeDate) {
					continue
				}
				out = append(out, t)
			}
			return renderTrades(out)
		},
	}
}
