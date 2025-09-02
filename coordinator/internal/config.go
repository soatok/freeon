package internal

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// This may expand in future versions
type CoordinatorConfig struct {
	Hostname string `json:"hostname"`
	Database string `json:"database"`
}

func getConfigFile() (string, error) {
	if path := os.Getenv("FREON_COORDINATOR_CONFIG"); path != "" {
		return path, nil
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".freeon-coordinator.json"), nil
}

// Default user config
func NewServerConfig() (CoordinatorConfig, error) {
	config := CoordinatorConfig{
		Hostname: "localhost:8462",
		Database: "./database.sqlite",
	}
	err := config.Save()
	if err != nil {
		return CoordinatorConfig{}, err
	}
	return config, nil
}

// Load the user config from a saved file
func LoadServerConfig() (CoordinatorConfig, error) {
	configPath, err := getConfigFile()
	if err != nil {
		return CoordinatorConfig{}, err
	}

	file, err := os.Open(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return NewServerConfig()
		}
		return CoordinatorConfig{}, err
	}
	defer file.Close()

	var conf CoordinatorConfig
	if err := json.NewDecoder(file).Decode(&conf); err != nil {
		return CoordinatorConfig{}, err
	}
	return conf, err
}

func (cfg CoordinatorConfig) Save() error {
	configPath, err := getConfigFile()
	if err != nil {
		return err
	}

	file, err := os.Create(configPath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "    ") // pretty-print
	return encoder.Encode(cfg)
}
