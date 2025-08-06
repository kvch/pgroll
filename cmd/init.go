// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	
	"github.com/xataio/pgroll/cmd/flags"
)

var initCmd = &cobra.Command{
	Use:   "init <file>",
	Short: "Initialize pgroll in the target database",
	RunE: func(cmd *cobra.Command, args []string) error {
		m, err := NewRoll(cmd.Context())
		if err != nil {
			return err
		}
		defer m.Close()

		isDryRun := flags.DryRun()
		spinnerText := "Initializing pgroll..."
		if isDryRun {
			spinnerText = "[DRY RUN] Initializing pgroll..."
		}

		sp, _ := pterm.DefaultSpinner.WithText(spinnerText).Start()
		err = m.Init(cmd.Context())
		if err != nil {
			sp.Fail(fmt.Sprintf("Failed to initialize pgroll: %s", err))
			return err
		}

		msg := "Initialization complete"
		if isDryRun {
			msg = "[DRY RUN] Initialization would be performed (no changes made)"
		}
		sp.Success(msg)
		return nil
	},
}
