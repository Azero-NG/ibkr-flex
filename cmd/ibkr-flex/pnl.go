package main

import (
	"github.com/spf13/cobra"

	"github.com/Azero-NG/ibkr-flex/internal/flex"
)

func newPnLCmd() *cobra.Command {
	var summary bool
	cmd := &cobra.Command{
		Use:   "pnl",
		Short: "Daily P&L attribution (mtm, fx, dividends, fees, ...) from ChangeInNAV",
		Long: `Daily P&L attribution per ChangeInNAV records — each row is one day's
breakdown of NAV change into trading MTM, FX translation, dividends,
commissions, withholding tax, deposits/withdrawals, etc.

With --summary, components are summed across the date range and a single
record is returned representing cumulative P&L attribution. The
canonical "actual profit" answer is then:

  endingValue - startingValue - depositsWithdrawals
  = mtm + realized + changeInUnrealized + fxTranslation + netFxTrading
    + dividends + changeInDividendAccruals + interest + commissions
    + otherFees + withholdingTax + grantActivity`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := requireAccount(); err != nil {
				return err
			}
			stmt, err := loadStatement(cmd.Context())
			if err != nil {
				return err
			}
			items := make([]flex.NAVChange, 0, len(stmt.NAVChanges))
			for _, n := range stmt.NAVChanges {
				if n.Account != globals.account {
					continue
				}
				if !dateInRange(n.Date) {
					continue
				}
				items = append(items, n)
			}
			if summary {
				items = []flex.NAVChange{aggregateNAVChanges(items)}
			}
			return renderNAVChanges(items)
		},
	}
	cmd.Flags().BoolVar(&summary, "summary", false, "aggregate components across the period into a single cumulative record")
	return cmd
}

func aggregateNAVChanges(items []flex.NAVChange) flex.NAVChange {
	if len(items) == 0 {
		return flex.NAVChange{}
	}
	out := flex.NAVChange{
		Currency: items[0].Currency,
		Account:  items[0].Account,
	}
	earliest, latest := items[0], items[0]
	for _, n := range items {
		if n.Date < earliest.Date {
			earliest = n
		}
		if n.Date > latest.Date {
			latest = n
		}
		out.MTM += n.MTM
		out.Realized += n.Realized
		out.ChangeInUnrealized += n.ChangeInUnrealized
		out.FxTranslation += n.FxTranslation
		out.NetFxTrading += n.NetFxTrading
		out.Dividends += n.Dividends
		out.ChangeInDividendAccruals += n.ChangeInDividendAccruals
		out.Interest += n.Interest
		out.Commissions += n.Commissions
		out.OtherFees += n.OtherFees
		out.WithholdingTax += n.WithholdingTax
		out.DepositsWithdrawals += n.DepositsWithdrawals
		out.GrantActivity += n.GrantActivity
	}
	out.Date = earliest.Date + ".." + latest.Date
	out.StartingValue = earliest.StartingValue
	out.EndingValue = latest.EndingValue
	return out
}
