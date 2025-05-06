package loader

import (
	"encoding/json"
	"os"

	"github.com/CTAPblY/yadro/structures"
)

func LoadConfig(filename string) (*structures.Config, error) {
	file, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config structures.Config
	if err := json.Unmarshal(file, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
