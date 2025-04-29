// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/xataio/pgroll/pkg/migrations"
	"sigs.k8s.io/yaml"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new migration interactively",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := pterm.DefaultInteractiveTextInput.WithDefaultText("Set the name of your migration").Show()
		outputFormat, _ := pterm.DefaultInteractiveSelect.WithDefaultText("File format").WithOptions([]string{"yaml", "json"}).WithDefaultOption("yaml").Show()
		addMoreOperations := true

		mig := &migrations.Migration{}
		for addMoreOperations {
			selectedOption, _ := pterm.DefaultInteractiveSelect.WithDefaultText("Select operation").WithOptions(migrations.AllOperations).Show()
			pterm.Info.Printfln("Selected option: %s", pterm.Green(selectedOption))
			op, _ := migrations.OperationFromName(migrations.OpName(selectedOption))
			mig.Operations = append(mig.Operations, op)
			for i := 0; i < reflect.Indirect(reflect.ValueOf(op)).Type().NumField(); i++ {
				value := reflect.Indirect(reflect.ValueOf(op))
				opSetting := value.Type().Field(i)
				switch opSetting.Type.Kind() {
				case reflect.String:
					selected, _ := pterm.DefaultInteractiveTextInput.WithDefaultText(strings.ToLower(opSetting.Name)).Show()
					value.Field(i).SetString(selected)
				default:
					continue
				}

			}
			addMoreOperations, _ = pterm.DefaultInteractiveConfirm.WithDefaultText("Add more operations").Show()
		}

		file, _ := os.Create(fmt.Sprintf("%s.%s", name, outputFormat))
		defer file.Close()

		switch outputFormat {
		case "json":
			enc := json.NewEncoder(file)
			enc.SetIndent("", "  ")
			if err := enc.Encode(mig); err != nil {
				return fmt.Errorf("encode migration: %w", err)
			}
		case "yaml":
			out, err := yaml.Marshal(mig)
			if err != nil {
				return fmt.Errorf("encode migration: %w", err)
			}
			_, err = file.Write(out)
			if err != nil {
				return fmt.Errorf("write migration: %w", err)
			}
		}

		return nil
	},
}
