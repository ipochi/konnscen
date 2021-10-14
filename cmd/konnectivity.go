package cmd

import (
	"github.com/spf13/cobra"
)

var (
	patchAPIserver bool

	// konnectivityCmd represents the konnectivity command
	konnectivityCmd = &cobra.Command{
		Use:   "konnectivity",
		Short: "List and Run various Konnectivity test scenarios.",
	}
)

func init() {
	rootCmd.AddCommand(konnectivityCmd)
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// konnectivityCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// konnectivityCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
