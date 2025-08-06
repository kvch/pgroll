// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	
	"github.com/xataio/pgroll/cmd/flags"
)

var completeCmd = &cobra.Command{
	Use:   "complete <file>",
	Short: "Complete an ongoing migration with the operations present in the given file",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Create a roll instance and check if pgroll is initialized
		m, err := NewRollWithInitCheck(cmd.Context())
		if err != nil {
			return err
		}
		defer m.Close()

		isDryRun := flags.DryRun()
		spinnerText := "Completing migration..."
		if isDryRun {
			spinnerText = "[DRY RUN] Completing migration..."
		}

		sp, _ := pterm.DefaultSpinner.WithText(spinnerText).Start()
		err = m.Complete(cmd.Context())
		if err != nil {
			sp.Fail(fmt.Sprintf("Failed to complete migration: %s", err))
			return err
		}

		msg := "Migration successful!"
		if isDryRun {
			msg = "[DRY RUN] Migration would be completed (no changes made)"
		}
		sp.Success(msg)
		return nil
	},
}
