package main

import (
	"github.com/spf13/cobra"

	"github.com/Azero-NG/ibkr-flex/internal/flex"
)

func newPositionsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "positions",
		Short: "List open positions snapshot",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := requireAccount(); err != nil {
				return err
			}
			stmt, err := loadStatement(cmd.Context())
			if err != nil {
				return err
			}
			out := make([]flex.Position, 0, len(stmt.Positions))
			for _, p := range stmt.Positions {
				if p.Account != globals.account {
					continue
				}
				if !dateInRange(p.ReportDate) {
					continue
				}
				out = append(out, p)
			}
			return renderPositions(out)
		},
	}
}
