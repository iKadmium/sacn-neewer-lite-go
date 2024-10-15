package main

import (
	"encoding/json"
	"io"
	"os"
)

type LightConfig struct {
	ID       string `json:"id"`
	Universe uint16 `json:"universe"`
	Address  uint16 `json:"address"`
}

type Config struct {
	Lights []LightConfig `json:"lights"`
}

func NewConfig() Config {
	return Config{
		Lights: make([]LightConfig, 0),
	}
}

func (c *Config) GetUniverses() []uint16 {
	universeSet := make(map[uint16]struct{})
	for _, light := range c.Lights {
		universeSet[light.Universe] = struct{}{}
	}

	universes := make([]uint16, 0, len(universeSet))
	for universe := range universeSet {
		universes = append(universes, universe)
	}
	return universes
}

func ConfigFromFile(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func ConfigFromFileOrNew(path string) *Config {
	config, err := ConfigFromFile(path)
	if err != nil {
		return &Config{}
	}
	return config
}

func (c *Config) Save(path string) error {
	data, err := json.Marshal(c)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
