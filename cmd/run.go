package cmd

import (
	"fmt"
	"log"

	"github.com/ipochi/konnscen/pkg/config"
	"github.com/ipochi/konnscen/pkg/scenarios"
	"github.com/spf13/cobra"
)

var (

	// runCmd represents the run command
	runCmd = &cobra.Command{
		Use:   "run",
		Short: "Run specified Konnectivity test scenario.",
		Run:   runScenario,
	}

	cfg        *config.Config
	configFile string
)

func init() {
	scenariosCmd.AddCommand(runCmd)

	runCmd.Flags().StringVarP(&configFile, "config-file", "c", "config.yaml", "Config file for scenarios")
}

func runScenario(cmd *cobra.Command, args []string) {
	if err := validateArgs(args); err != nil {
		log.Fatal(err)
	}

	cfg = config.LoadConfig(configFile)
	if err := scenarios.Run(cfg, args); err != nil {
		log.Fatal(err)
	}
}

func validateArgs(args []string) error {
	sc := AvailableScenarios()
	for _, arg := range args {
		if _, ok := sc[arg]; !ok {
			return fmt.Errorf("scenario %q is not valid", arg)
		}
	}

	return nil
}
