package config

import (
	"gopkg.in/yaml.v3"
	"os"
)

type ElasticsearchConfig struct {
	Addresses []string
	Username  *string `yaml:"username"`
	Password  *string
	Index     string
}

type Rule struct {
	Record string
	Labels map[string]string
	Query  string
}

type Group struct {
	Name          string
	Interval      string
	Rules         []Rule
	Elasticsearch ElasticsearchConfig
}

type Configuration struct {
	Groups []Group
}

func Load(filePath string) (Configuration, error) {
	cfg := Configuration{}
	dat, err := os.ReadFile(filePath)

	if err != nil {
		return cfg, err
	}

	err = yaml.Unmarshal(dat, &cfg)

	if err != nil {
		return cfg, err
	}

	return cfg, nil
}
