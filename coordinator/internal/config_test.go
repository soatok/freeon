package internal_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/soatok/freeon/coordinator/internal"
	"github.com/stretchr/testify/assert"
)

func TestServerConfig(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("HOME", tempDir)
	// Windows uses this environment variable instead
	t.Setenv("USERPROFILE", tempDir)
	os.Mkdir(tempDir, 0700)

	// Test NewServerConfig
	config, err := internal.NewServerConfig()
	assert.NoError(t, err)
	assert.Equal(t, config.Hostname, "localhost:8462")
	assert.Equal(t, config.Database, "./database.sqlite")

	if err := config.Save(); err != nil {
		panic(err)
	}

	// Check if the file was created
	configPath := filepath.Join(tempDir, ".freeon-coordinator.json")
	_, err = os.Stat(configPath)
	assert.NoError(t, err)

	// Test LoadServerConfig
	loadedConfig, err := internal.LoadServerConfig()
	assert.NoError(t, err)
	assert.Equal(t, config, loadedConfig)

	// Test Save
	loadedConfig.Hostname = "example.com:1234"
	err = loadedConfig.Save()
	assert.NoError(t, err)

	savedConfig, err := internal.LoadServerConfig()
	assert.NoError(t, err)
	assert.Equal(t, loadedConfig, savedConfig)
}
