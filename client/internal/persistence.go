package internal

import (
	"encoding/json"
	"os"
	"path/filepath"
)

func getConfigFile() (string, error) {
	homeDir := os.Getenv("FREEON_HOME")
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
	return filepath.Join(homeDir, ".freeon.json"), nil
}

// Default user config
func NewUserConfig() (FreeonConfig, error) {
	config := FreeonConfig{
		Shares: []Shares{},
	}
	err := config.Save()
	if err != nil {
		return FreeonConfig{}, err
	}
	return config, nil
}

// Load the user config from a saved file
func LoadUserConfig() (FreeonConfig, error) {
	configPath, err := getConfigFile()
	if err != nil {
		return FreeonConfig{}, err
	}

	file, err := os.Open(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return NewUserConfig()
		}
		return FreeonConfig{}, err
	}
	defer file.Close()

	var conf FreeonConfig
	if err := json.NewDecoder(file).Decode(&conf); err != nil {
		return FreeonConfig{}, err
	}
	return conf, err
}

// This API felt more natural for me to implement than `func SaveUserConfig(cfg FreeonConfig) error`
func (cfg FreeonConfig) Save() error {
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

func (cfg FreeonConfig) AddShare(host, groupID, publicKey, share string, otherShares map[string]string, myPartyID uint16) error {
	s := Shares{
		Host:           host,
		GroupID:        groupID,
		PublicKey:      publicKey,
		EncryptedShare: share,
		PublicShares:   otherShares,
		MyPartyID:      myPartyID,
	}
	cfg.Shares = append(cfg.Shares, s)
	return cfg.Save()
}
