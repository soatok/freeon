package internal_test

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"strings"
	"testing"

	"github.com/soatok/freeon/client/internal"
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

	// Some deterministic values which should produce an expected key format:
	h := sha512.Sum384([]byte("Soatok Dreamseeker"))
	pk2 := h[0:32][:]
	var sig2 []byte
	h2 := sha256.Sum256([]byte("Signature Format"))
	sig2 = append(sig2, h2[:]...)
	h2 = sha256.Sum256([]byte("Freeon - OpenSSH"))
	sig2 = append(sig2, h2[:]...)

	encoded2 := internal.OpenSSHEncode(pk2, sig2, namespace)
	assert.NotEmpty(t, encoded2)
	const expected = `-----BEGIN SSH SIGNATURE-----
AAAABlNTSFNJRwAAAAEAAAAzAAAAC3NzaC1lZDI1NTE5AAAAIEWbPXw3NFqPht+qbUzQeU
ot2rnHXclITN0UivggnYz5AAAABHRlc3QAAAAAAAAAC3NzaC1lZDI1NTE5AAAAUwAAAAtz
c2gtZWQyNTUxOQAAAEAn5PrscAKy4X4bzwdTN19iOi+Tb3UJYRJU9z/U6Jb+qtX3kF5ZYH
6eVkXFIipre97XzH+lojn92vOx7elXLe/Y
-----END SSH SIGNATURE-----
`
    left := strings.ReplaceAll(expected, "\r", "")
    left = strings.ReplaceAll(left, "\\n", "\n")
    right := strings.ReplaceAll(encoded2, "\r", "")
    right = strings.ReplaceAll(right, "\\n", "\n")
	assert.Equal(t, left, right)
}
