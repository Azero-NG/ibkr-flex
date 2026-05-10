package main

import (
	"github.com/spf13/cobra"

	"github.com/Azero-NG/ibkr-flex/internal/flex"
)

func newMTMCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "mtm",
		Short: "List mark-to-market performance summary entries",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := requireAccount(); err != nil {
				return err
			}
			stmt, err := loadStatement(cmd.Context())
			if err != nil {
				return err
			}
			out := make([]flex.MTMEntry, 0, len(stmt.MTMEntries))
			for _, m := range stmt.MTMEntries {
				if m.Account != globals.account {
					continue
				}
				if !dateInRange(m.Date) {
					continue
				}
				out = append(out, m)
			}
			return renderMTM(out)
		},
	}
}
