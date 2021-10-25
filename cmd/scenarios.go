package cmd

import (
	concon "github.com/ipochi/konnscen/pkg/scenarios/concurrent-connections"
	conportforwards "github.com/ipochi/konnscen/pkg/scenarios/concurrent-portforwards"
	"github.com/spf13/cobra"
)

// scenariosCmd represents the scenarios command
var scenariosCmd = &cobra.Command{
	Use:   "scenarios",
	Short: "Test scenarios for Konnectivity.",
}

func AvailableScenarios() map[string]string {
	s := map[string]string{}
	s[concon.Name] = concon.Name
	s[conportforwards.Name] = conportforwards.Name

	return s
}

func init() {
	rootCmd.AddCommand(scenariosCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// scenariosCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// scenariosCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
