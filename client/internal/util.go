package internal

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
)

// This is just a consistency check for the message, so we can abort early if something mismatches
func HashMessageForSanity(data []byte, groupID string) string {
	key := sha512.Sum384([]byte(groupID))
	mac := hmac.New(sha512.New384, key[:])
	mac.Write(data)
	return hex.EncodeToString(mac.Sum(nil))
}

func uint16ToHexBE(n uint16) string {
	bytes := []byte{byte(n >> 8), byte(n)}
	return hex.EncodeToString(bytes)
}

func hexBEToUint16(s string) (uint16, error) {
	bytes, err := hex.DecodeString(s)
	if err != nil {
		return 0, err
	}
	if len(bytes) != 2 {
		return 0, fmt.Errorf("expected 2 bytes, got %d", len(bytes))
	}
	return (uint16(bytes[0]) << 8) | uint16(bytes[1]), nil
}

// We aren't using UUIDs here because they only have 126 bits of entropy
func UniqueID() (string, error) {
	b := make([]byte, 24)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
