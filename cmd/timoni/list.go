/*
Copyright 2023 Stefan Prodan

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"context"
	"io"

	apiv1 "github.com/stefanprodan/timoni/api/v1alpha1"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"

	"github.com/stefanprodan/timoni/internal/runtime"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "Prints a table of instances and their module version",
	Example: ` # List all instances in a namespace
  timoni list --namespace default

  # List all instances on a cluster
  timoni ls -A

  # List all instances on a cluster subject to a certain bundle
  timoni ls -A --bundle podinfo
`,
	RunE: runListCmd,
}

type listFlags struct {
	allNamespaces bool
	bundleName    string
}

var listArgs listFlags

func init() {
	listCmd.Flags().BoolVarP(&listArgs.allNamespaces, "all-namespaces", "A", false,
		"List the requested object(s) across all namespaces.")
	listCmd.Flags().StringVarP(&listArgs.bundleName, "bundle", "", "",
		"List the requested object(s) subject to a certain bundle.")

	rootCmd.AddCommand(listCmd)
}

func runListCmd(cmd *cobra.Command, args []string) error {
	instances, err := listInstancesFromFlags()
	if err != nil {
		return err
	}

	var rows [][]string
	for _, inv := range instances {
		row := []string{}
		if listArgs.allNamespaces {
			row = []string{
				inv.Name,
				inv.Namespace,
				inv.Module.Repository,
				inv.Module.Version,
				inv.LastTransitionTime,
				printOrPass(inv.Labels[apiv1.BundleNameLabelKey]),
			}
		} else {
			row = []string{
				inv.Name,
				inv.Module.Repository,
				inv.Module.Version,
				inv.LastTransitionTime,
				printOrPass(inv.Labels[apiv1.BundleNameLabelKey]),
			}
		}
		rows = append(rows, row)
	}

	if listArgs.allNamespaces {
		printTable(rootCmd.OutOrStdout(), []string{"name", "namespace", "module", "version", "last applied", "bundle"}, rows)
	} else {
		printTable(rootCmd.OutOrStdout(), []string{"name", "module", "version", "last applied", "bundle"}, rows)
	}

	return nil
}

func listInstancesFromFlags() ([]*apiv1.Instance, error) {
	sm, err := runtime.NewResourceManager(kubeconfigArgs)
	if err != nil {
		return nil, err
	}

	iStorage := runtime.NewStorageManager(sm)

	ctx, cancel := context.WithTimeout(context.Background(), rootArgs.timeout)
	defer cancel()

	ns := *kubeconfigArgs.Namespace
	if listArgs.allNamespaces {
		ns = ""
	}

	return iStorage.List(ctx, ns, listArgs.bundleName)
}

func printTable(writer io.Writer, header []string, rows [][]string) {
	table := tablewriter.NewWriter(writer)
	table.SetHeader(header)
	table.SetAutoWrapText(false)
	table.SetAutoFormatHeaders(true)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetHeaderLine(false)
	table.SetBorder(false)
	table.SetTablePadding("\t")
	table.SetNoWhiteSpace(true)
	table.AppendBulk(rows)
	table.Render()
}

func printOrPass(value string) string {
	if value == "" {
		return "-"
	}
	return value
}
