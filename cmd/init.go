// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

func initCmd() *cobra.Command {
	var force bool
	initCmd := &cobra.Command{
		Use:   "init <file>",
		Short: "Initialize pgroll in the target database",
		RunE: func(cmd *cobra.Command, args []string) error {
			m, err := NewRoll(cmd.Context())
			if err != nil {
				return err
			}
			defer m.Close()

			if force {
				err := m.State().Clean(cmd.Context())
				if err != nil {
					pterm.Fatal.Println(fmt.Sprintf("Failed to clean pgroll state before initialization: %s", err))
					return err
				}
			}

			sp, _ := pterm.DefaultSpinner.WithText("Initializing pgroll...").Start()
			err = m.Init(cmd.Context())
			if err != nil {
				sp.Fail(fmt.Sprintf("Failed to initialize pgroll: %s", err))
				return err
			}

			sp.Success("Initialization complete")
			return nil
		},
	}

	initCmd.Flags().BoolVarP(&force, "force", "f", false, "Initialize pgroll state again, delete complete migration history")

	return initCmd
}
