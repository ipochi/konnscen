package config

import (
	"io/ioutil"
	"log"
	"os"

	concon "github.com/ipochi/konnscen/pkg/scenarios/concurrent-connections"
	conportforwards "github.com/ipochi/konnscen/pkg/scenarios/concurrent-portforwards"
	"gopkg.in/yaml.v3"
)

type Config struct {
	ConcurrentConnections  *concon.ConcurrentConnections           `yaml:"concurrent_connections,omitempty"`
	ConcurrentPortForwards *conportforwards.ConcurrentPortForwards `yaml:"concurrent_portforwards,omitempty"`
}

func NewConfig() *Config {
	return &Config{
		ConcurrentConnections:  concon.NewConcurrentConnections(),
		ConcurrentPortForwards: conportforwards.NewConcurrentPortForwards(),
	}
}

func LoadConfig(path string) *Config {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return NewConfig()
	}

	yfile, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalf("could not find config file named `q`: %q", err)
	}

	cfg := NewConfig()
	err = yaml.Unmarshal(yfile, cfg)
	if err != nil {
		log.Fatalf("failed to unmarshal config file`: %q", err)
	}

	return cfg
}
