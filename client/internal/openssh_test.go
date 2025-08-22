package internal_test

import (
	"crypto/rand"
	"testing"

	"github.com/soatok/freon/client/internal"
	"github.com/stretchr/testify/assert"
)

func TestOpenSSHEncode(t *testing.T) {
	pk := make([]byte, 32)
	rand.Read(pk)

	sig := make([]byte, 64)
	rand.Read(sig)

	namespace := "test"

	encoded := internal.OpenSSHEncode(pk, sig, namespace)
	assert.NotEmpty(t, encoded)
	assert.Contains(t, encoded, "-----BEGIN SSH SIGNATURE-----")
	assert.Contains(t, encoded, "-----END SSH SIGNATURE-----")
}
