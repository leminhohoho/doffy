package runner

import (
	"errors"
	"os"
	"path"

	"dario.cat/mergo"
	"github.com/pelletier/go-toml/v2"
)

type Config struct {
	Files struct {
		Exclude []string
	}
}

var defaultCfg = Config{
	Files: struct{ Exclude []string }{
		Exclude: []string{".git"},
	},
}

var configFileName = ".doffy.toml"

func NewConfig(dotfilesPath string) (*Config, error) {
	cfg := defaultCfg

	configPath := path.Join(dotfilesPath, configFileName)

	data, err := os.ReadFile(configPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &cfg, nil
		}

		return nil, err
	}

	var overrideCfg Config

	if err := toml.Unmarshal(data, &overrideCfg); err != nil {
		return nil, err
	}

	if err := mergo.Merge(&cfg, overrideCfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
