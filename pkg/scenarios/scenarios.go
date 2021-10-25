package scenarios

import (
	"fmt"

	"github.com/ipochi/konnscen/pkg/config"
	concon "github.com/ipochi/konnscen/pkg/scenarios/concurrent-connections"
	conportforwards "github.com/ipochi/konnscen/pkg/scenarios/concurrent-portforwards"
)

var (
	scenariosMap map[string]Scenario
)

type Scenario interface {
	Run() error
	Cleanup() error
}

func initializeMap(cfg *config.Config) {
	scenariosMap = map[string]Scenario{}

	scenariosMap[concon.Name] = cfg.ConcurrentConnections
	scenariosMap[conportforwards.Name] = cfg.ConcurrentPortForwards
}

func Run(cfg *config.Config, sc []string) error {
	initializeMap(cfg)

	for _, s := range sc {
		err := scenariosMap[s].Run()
		if err != nil {
			return fmt.Errorf("running scenario %q: %w", s, err)
		}
	}

	return nil
}
