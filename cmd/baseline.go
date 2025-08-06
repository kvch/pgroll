// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/xataio/pgroll/cmd/flags"
	"github.com/xataio/pgroll/pkg/migrations"
)

func baselineCmd() *cobra.Command {
	var useJSON bool
	var yes bool

	baselineCmd := &cobra.Command{
		Use:       "baseline <version> <target directory>",
		Short:     "Create a baseline migration for an existing database schema",
		Args:      cobra.ExactArgs(2),
		ValidArgs: []string{"version", "directory"},
		RunE: func(cmd *cobra.Command, args []string) error {
			version := args[0]
			targetDir := args[1]

			ctx := cmd.Context()

			// Create a roll instance
			m, err := NewRollWithInitCheck(ctx)
			if err != nil {
				return err
			}
			defer m.Close()

			// Ensure that the target directory exists
			if err := ensureDirectoryExists(targetDir); err != nil {
				return err
			}

			isDryRun := flags.DryRun()
			
			// Prompt for confirmation unless --yes flag is set or in dry-run mode
			if !yes && !isDryRun {
				fmt.Println("Creating a baseline migration will restart the migration history.")
				ok, _ := pterm.DefaultInteractiveConfirm.Show()
				if !ok {
					return nil
				}
			}

			// Create a placeholder baseline migration
			ops := migrations.Operations{&migrations.OpRawSQL{Up: ""}}
			opsJSON, err := json.Marshal(ops)
			if err != nil {
				return fmt.Errorf("failed to marshal operations: %w", err)
			}
			mig := &migrations.RawMigration{
				Name:       version,
				Operations: opsJSON,
			}

			// Write the placeholder migration to disk (unless dry-run)
			var filePath string
			if !isDryRun {
				filePath, err = writeMigrationToFile(mig, targetDir, "", useJSON)
				if err != nil {
					return fmt.Errorf("failed to write placeholder baseline migration: %w", err)
				}
			}

			spinnerText := fmt.Sprintf("Creating baseline migration %q...", version)
			if isDryRun {
				spinnerText = fmt.Sprintf("[DRY RUN] Creating baseline migration %q...", version)
			}
			sp, _ := pterm.DefaultSpinner.WithText(spinnerText).Start()

			// Create the baseline in the target database
			err = m.CreateBaseline(ctx, version)
			if err != nil {
				sp.Fail(fmt.Sprintf("Failed to create baseline: %s", err))
				if !isDryRun && filePath != "" {
					err = errors.Join(err, os.Remove(filePath))
				}
				return err
			}

			msg := fmt.Sprintf("Baseline created successfully. Placeholder migration %q written", filePath)
			if isDryRun {
				msg = fmt.Sprintf("[DRY RUN] Baseline would be created for version %q (no changes made)", version)
			}
			sp.Success(msg)
			return nil
		},
	}

	baselineCmd.Flags().BoolVarP(&useJSON, "json", "j", false, "output in JSON format instead of YAML")
	baselineCmd.Flags().BoolVarP(&yes, "yes", "y", false, "skip confirmation prompt")

	return baselineCmd
}
