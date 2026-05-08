package main

import (
	"sort"

	"github.com/spf13/cobra"
)

func newAccountsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "accounts",
		Short: "List account IDs found in the Flex statement",
		RunE: func(cmd *cobra.Command, _ []string) error {
			stmt, err := loadStatement(cmd.Context())
			if err != nil {
				return err
			}
			ids := stmt.AccountIDs()
			sort.Strings(ids)
			return renderAccounts(ids)
		},
	}
}
