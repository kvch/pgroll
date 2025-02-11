// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/xataio/pgroll/cmd/flags"
	"github.com/xataio/pgroll/pkg/backfill"
)

func migrateCmd() *cobra.Command {
	var complete bool

	migrateCmd := &cobra.Command{
		Use:       "migrate <directory>",
		Short:     "Apply outstanding migrations from a directory to a database",
		Example:   "migrate ./migrations",
		Args:      cobra.ExactArgs(1),
		ValidArgs: []string{"directory"},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			migrationsDir := args[0]

			m, err := NewRoll(ctx)
			if err != nil {
				return err
			}
			defer m.Close()

			latestVersion, err := m.State().LatestVersion(ctx, m.Schema())
			if err != nil {
				return fmt.Errorf("unable to determine latest version: %w", err)
			}

			active, err := m.State().IsActiveMigrationPeriod(ctx, m.Schema())
			if err != nil {
				return fmt.Errorf("unable to determine active migration period: %w", err)
			}
			if active {
				return fmt.Errorf("migration %q is active and must be completed first", *latestVersion)
			}

			info, err := os.Stat(migrationsDir)
			if err != nil {
				return fmt.Errorf("failed to stat directory: %w", err)
			}
			if !info.IsDir() {
				return fmt.Errorf("migrations directory %q is not a directory", migrationsDir)
			}

			migs, err := m.UnappliedMigrations(ctx, os.DirFS(migrationsDir))
			if err != nil {
				return fmt.Errorf("failed to get migrations to apply: %w", err)
			}

			if len(migs) == 0 {
				fmt.Println("database is up to date; no migrations to apply")
				return nil
			}

			backfillConfig := backfill.NewConfig(
				backfill.WithBatchSize(flags.BackfillBatchSize()),
				backfill.WithBatchDelay(flags.BackfillBatchDelay()),
			)

			// Run all migrations after the latest version up to the final migration,
			// completing each one.
			for _, mig := range migs[:len(migs)-1] {
				if err := runMigration(ctx, m, mig, true, backfillConfig); err != nil {
					return fmt.Errorf("failed to run migration file %q: %w", mig.Name, err)
				}
			}

			// Run the final migration, completing it only if requested.
			return runMigration(ctx, m, migs[len(migs)-1], complete, backfillConfig)
		},
	}

	migrateCmd.Flags().Int("backfill-batch-size", backfill.DefaultBatchSize, "Number of rows backfilled in each batch")
	migrateCmd.Flags().Duration("backfill-batch-delay", backfill.DefaultDelay, "Duration of delay between batch backfills (eg. 1s, 1000ms)")
	migrateCmd.Flags().BoolVarP(&complete, "complete", "c", false, "complete the final migration rather than leaving it active")

	viper.BindPFlag("BACKFILL_BATCH_SIZE", migrateCmd.Flags().Lookup("backfill-batch-size"))
	viper.BindPFlag("BACKFILL_BATCH_DELAY", migrateCmd.Flags().Lookup("backfill-batch-delay"))

	return migrateCmd
}
