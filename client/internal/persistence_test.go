package internal_test

import (
	"os"
	"testing"

	"github.com/soatok/freon/client/internal"
	"github.com/stretchr/testify/assert"
)

func TestPersistence(t *testing.T) {
	// Create a temporary directory for the config file
	tmpdir, err := os.MkdirTemp("", "freon-test-")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpdir)

	// Set the user home directory to the temporary directory
	os.Setenv("HOME", tmpdir)

	// Test NewUserConfig
	cfg, err := internal.NewUserConfig()
	assert.NoError(t, err)
	assert.NotNil(t, cfg)

	// Test LoadUserConfig
	loadedCfg, err := internal.LoadUserConfig()
	assert.NoError(t, err)
	assert.Equal(t, cfg, loadedCfg)

	// Test AddShare
	err = loadedCfg.AddShare("localhost", "group1", "pk1", "share1", nil, 1)
	assert.NoError(t, err)

	// Load the config again to check if the share was added
	reloadedCfg, err := internal.LoadUserConfig()
	assert.NoError(t, err)
	assert.Len(t, reloadedCfg.Shares, 1)
	assert.Equal(t, "localhost", reloadedCfg.Shares[0].Host)
	assert.Equal(t, "group1", reloadedCfg.Shares[0].GroupID)
	assert.Equal(t, "pk1", reloadedCfg.Shares[0].PublicKey)
	assert.Equal(t, "share1", reloadedCfg.Shares[0].EncryptedShare)
}
