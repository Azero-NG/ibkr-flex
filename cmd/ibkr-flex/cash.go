package main

import (
	"github.com/spf13/cobra"

	"github.com/Azero-NG/ibkr-flex/internal/flex"
)

func newCashCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "cash",
		Short: "List cash transactions (deposits, withdrawals, fees, interest)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := requireAccount(); err != nil {
				return err
			}
			stmt, err := loadStatement(cmd.Context())
			if err != nil {
				return err
			}
			out := make([]flex.CashTx, 0, len(stmt.CashTxs))
			for _, c := range stmt.CashTxs {
				if c.Account != globals.account {
					continue
				}
				if !dateInRange(c.Date) {
					continue
				}
				out = append(out, c)
			}
			return renderCashTxs(out)
		},
	}
}
