package internal

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"

	"filippo.io/age"
)

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

	return hex.EncodeToString(encryptedBuf.Bytes()), nil
}

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
