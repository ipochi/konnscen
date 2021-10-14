package cmd

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List Konnectivity test scenarios.",
	Run:   runList,
}

func init() {
	scenariosCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) {
	scenarios := AvailableScenarios()

	keys := make([]string, 0, len(scenarios))
	for k := range scenarios {
		keys = append(keys, k)
	}

	sort.Strings(keys)
	for _, name := range keys {
		fmt.Println("\t", name)
	}
}
