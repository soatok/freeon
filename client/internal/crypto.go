package internal

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"

	"filippo.io/age"
)

// Encrypt a Shamir share to a public key, using age.
func EncryptShare(recipientStr string, share []byte) (string, error) {
	// Parse the recipient string (could be a public key or identity)
	recipient, err := age.ParseX25519Recipient(recipientStr)
	if err != nil {
		return "", fmt.Errorf("failed to parse recipient: %w", err)
	}

	var encryptedBuf bytes.Buffer
	// The writer will write encrypted output to encryptedBuf
	w, err := age.Encrypt(&encryptedBuf, recipient)
	if err != nil {
		return "", fmt.Errorf("failed to create age writer: %w", err)
	}

	if _, err := w.Write(share); err != nil {
		return "", fmt.Errorf("failed to write data: %w", err)
	}
	if err = w.Close(); err != nil {
		return "", fmt.Errorf("failed to close age writer: %w", err)
	}

	return hex.EncodeToString(encryptedBuf.Bytes()), nil
}

// Decrypt an arbitrary hex-encoded string with a specific age identity.
// If you do not have an age.Identity on hand, you probably want DecryptShareFor() instead.
func DecryptShare(encryptedShareHex string, identity age.Identity) ([]byte, error) {
	encryptedShare, err := hex.DecodeString(encryptedShareHex)
	if err != nil {
		return nil, err
	}
	r, err := age.Decrypt(bytes.NewReader(encryptedShare), identity)
	if err != nil {
		return nil, err
	}

	// Read all decrypted data
	decryptedData, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read decrypted data: %w", err)
	}

	return decryptedData, nil
}

// We need a list of age.Identity structs to pass to age. This helper function
// wraps age's existing API.
func ParseAgeIdentityFile(filePath string) ([]age.Identity, error) {
	// Open the identity file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open identity file %s: %w", filePath, err)
	}
	defer file.Close()

	// Parse identities from the file
	identities, err := age.ParseIdentities(file)
	if err != nil {
		return nil, fmt.Errorf("failed to parse identities from %s: %w", filePath, err)
	}

	if len(identities) == 0 {
		return nil, fmt.Errorf("no valid identities found in %s", filePath)
	}

	return identities, nil
}

// This is the high-level API used for decryption. The inputs are sourced from the
// Config (eencryptedShareHex) and CLI arguments (filePath) respectively.
func DecryptShareFor(encryptedShareHex, filePath string) ([]byte, error) {
	idents, err := ParseAgeIdentityFile(filePath)
	if err != nil {
		return nil, err
	}
	for _, id := range idents {
		decrypted, err := DecryptShare(encryptedShareHex, id)
		if err == nil {
			return decrypted, nil
		}
	}
	return nil, errors.New("could not decrypt share")
}
