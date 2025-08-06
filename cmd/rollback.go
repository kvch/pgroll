// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	
	"github.com/xataio/pgroll/cmd/flags"
)

var rollbackCmd = &cobra.Command{
	Use:   "rollback",
	Short: "Roll back an ongoing migration",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Create a roll instance and check if pgroll is initialized
		m, err := NewRollWithInitCheck(cmd.Context())
		if err != nil {
			return err
		}
		defer m.Close()

		isDryRun := flags.DryRun()
		spinnerText := "Rolling back migration..."
		if isDryRun {
			spinnerText = "[DRY RUN] Rolling back migration..."
		}

		sp, _ := pterm.DefaultSpinner.WithText(spinnerText).Start()
		err = m.Rollback(cmd.Context())
		if err != nil {
			sp.Fail(fmt.Sprintf("Failed to roll back migration: %s", err))
			return err
		}

		msg := "Migration rolled back. Changes made since the last version have been reverted"
		if isDryRun {
			msg = "[DRY RUN] Migration would be rolled back (no changes made)"
		}
		sp.Success(msg)
		return nil
	},
}
