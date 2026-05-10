package main

import (
	"github.com/spf13/cobra"

	"github.com/Azero-NG/ibkr-flex/internal/flex"
)

func newNAVCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "nav",
		Short: "List net asset value daily snapshots",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := requireAccount(); err != nil {
				return err
			}
			stmt, err := loadStatement(cmd.Context())
			if err != nil {
				return err
			}
			out := make([]flex.NAVEntry, 0, len(stmt.NAVEntries))
			for _, n := range stmt.NAVEntries {
				if n.Account != globals.account {
					continue
				}
				if !dateInRange(n.Date) {
					continue
				}
				out = append(out, n)
			}
			return renderNAV(out)
		},
	}
}
