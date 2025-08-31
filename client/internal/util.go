package internal

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha512"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"slices"
)

// This is just a consistency check for the message, so we can abort early if something mismatches
func HashMessageForSanity(data []byte, groupID string) string {
	key := sha512.Sum384([]byte(groupID))
	mac := hmac.New(sha512.New384, key[:])
	mac.Write(data)
	return hex.EncodeToString(mac.Sum(nil))
}

func Uint16ToHexBE(n uint16) string {
	bytes := []byte{byte(n >> 8), byte(n)}
	return hex.EncodeToString(bytes)
}

func HexBEToUint16(s string) (uint16, error) {
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

// Determine index from hash + party size.
//
// We sample 64 bits and reduce by x bits (where x <= 16).
// This gives us n+48 bits reduced modulo an n-bit number, which minimizes modulo bias.
func SelectIndex(ch []byte, partySize uint64) uint64 {
	if len(ch) < 8 {
		panic("ch must be at least 8 bytes long")
	}
	buf := ch[len(ch)-8:]
	unbiased := binary.BigEndian.Uint64(buf)
	return unbiased % partySize
}

// Given a hash, party ID, and list of all party IDs:
// Select a pseudorandom element of party.
// The last 8 bytes of hash are converted to a uint64, then reduced modulo the size of the party
// If the result of this reduction equals your party ID, you are elected
func AmIElected(ch []byte, me uint16, party []uint16) bool {
	if len(party) < 2 {
		return true
	}
	if !slices.Contains(party, me) {
		// If you are not a member, you are not chosen
		return false
	}
	slices.Sort(party)
	ps := uint64(len(party))
	index := SelectIndex(ch, ps)
	return party[index] == me
}
