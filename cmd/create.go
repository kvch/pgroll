// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strconv"
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
		outputFormat, _ := pterm.DefaultInteractiveSelect.WithDefaultText("File format").WithMaxHeight(7).WithOptions([]string{"yaml", "json"}).WithDefaultOption("yaml").Show()
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
				case reflect.Bool:
					selected, _ := pterm.DefaultInteractiveTextInput.WithDefaultText(strings.ToLower(opSetting.Name)).Show()
					selectedBool, _ := strconv.ParseBool(selected)
					value.Field(i).SetBool(selectedBool)
				case reflect.Slice:
					addElems, _ := pterm.DefaultInteractiveConfirm.WithDefaultValue(true).WithDefaultText(fmt.Sprintf("Add %s", strings.ToLower(opSetting.Name))).Show()
					if addElems {
						elemSlice := reflect.MakeSlice(opSetting.Type, 0, 1)
						for addElems {
							elemType := opSetting.Type.Elem()
							newElem := reflect.New(elemType)
							for j := 0; j < elemType.NumField(); j++ {
								elemAttr := elemType.Field(j)
								switch elemAttr.Type.Kind() {
								case reflect.String:
									selected, _ := pterm.DefaultInteractiveTextInput.WithDefaultText(strings.ToLower(elemAttr.Name)).Show()
									reflect.Indirect(newElem).Field(j).SetString(selected)
								case reflect.Bool:
									selected, _ := pterm.DefaultInteractiveTextInput.WithDefaultText(strings.ToLower(elemAttr.Name)).Show()
									selectedBool, _ := strconv.ParseBool(selected)
									reflect.Indirect(newElem).Field(j).SetBool(selectedBool)
								case reflect.Pointer:
									//fmt.Println(elemAttr.Type.Elem())
									//newVal := reflect.New(elemAttr.Type.Elem())
									if elemAttr.Type.Elem().Kind() != reflect.String {
										continue
									}
									selected, _ := pterm.DefaultInteractiveTextInput.WithDefaultText(strings.ToLower(elemAttr.Name)).Show()
									if selected != "" {
										reflect.Indirect(newElem).Field(j).Set(reflect.ValueOf(&selected))
									}
								default:
									continue
									//fmt.Println(elemAttr.Type.Kind())
								}
							}
							elemSlice = reflect.Append(elemSlice, reflect.Indirect(newElem))
							addElems, _ = pterm.DefaultInteractiveConfirm.WithDefaultValue(true).WithDefaultText(fmt.Sprintf("Add %s", strings.ToLower(opSetting.Name))).Show()
						}
						value.Field(i).Set(elemSlice)
					}
				default:
					//fmt.Println(opSetting.Type.Kind())
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
