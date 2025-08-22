package internal

import (
	"encoding/json"
	"os"
	"path/filepath"
)

func getConfigFile() (string, error) {
	homeDir := os.Getenv("FREON_HOME")
	var err error
	if homeDir == "" {
		homeDir, err = os.UserHomeDir()
		if err != nil {
			return "", err
		}
	}
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".freon.json"), nil
}

// Default user config
func NewUserConfig() (FreonConfig, error) {
	config := FreonConfig{
		Shares: []Shares{},
	}
	err := config.Save()
	if err != nil {
		return FreonConfig{}, err
	}
	return config, nil
}

// Load the user config from a saved file
func LoadUserConfig() (FreonConfig, error) {
	configPath, err := getConfigFile()
	if err != nil {
		return FreonConfig{}, err
	}

	file, err := os.Open(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return NewUserConfig()
		}
		return FreonConfig{}, err
	}
	defer file.Close()

	var conf FreonConfig
	if err := json.NewDecoder(file).Decode(&conf); err != nil {
		return FreonConfig{}, err
	}
	return conf, err
}

// This API felt more natural for me to implement than `func SaveUserConfig(cfg FreonConfig) error`
func (cfg FreonConfig) Save() error {
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

func (cfg FreonConfig) AddShare(host, groupID, publicKey, share string, otherShares map[string]string) error {
	s := Shares{
		Host:           host,
		GroupID:        groupID,
		PublicKey:      publicKey,
		EncryptedShare: share,
		PublicShares:   otherShares,
	}
	cfg.Shares = append(cfg.Shares, s)
	return cfg.Save()
}
