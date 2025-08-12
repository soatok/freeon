package internal

import (
	"encoding/hex"
	"fmt"
)

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
