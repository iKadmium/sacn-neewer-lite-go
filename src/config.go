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

func (c *Config) GetUniverses() []uint16 {
	universes := make([]uint16, len(c.Lights))
	for i, light := range c.Lights {
		universes[i] = light.Universe
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

// func main() {
// 	config, err := ConfigFromFile("path/to/config.json")
// 	if err != nil {
// 		fmt.Println("Error:", err)
// 		return
// 	}

// 	fmt.Println("Universes:", config.GetUniverses())
// }
