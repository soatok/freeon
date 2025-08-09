package internal

import (
	"crypto/rand"
	"encoding/hex"
)

// We aren't using UUIDs here because they only have 126 bits of entropy
func UniqueID() (string, error) {
	b := make([]byte, 24)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
