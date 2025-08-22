package internal_test

import (
	"crypto/sha512"
	"fmt"
	"testing"

	"github.com/soatok/freon/client/internal"
	"github.com/stretchr/testify/assert"
	"github.com/taurusgroup/frost-ed25519/pkg/frost/party"
)

func TestHashMessageForSanity(t *testing.T) {
	groupID, err := internal.UniqueID()
	if err != nil {
		panic(err)
	}
	msgA := []byte("Hello World")
	msgB := []byte("UwU")
	hash1 := internal.HashMessageForSanity(msgA, "g_"+groupID)
	hash2 := internal.HashMessageForSanity(msgA, "c_"+groupID)
	hash3 := internal.HashMessageForSanity(msgB, "g_"+groupID)
	hash4 := internal.HashMessageForSanity(msgB, "c_"+groupID)
	assert.NotEqual(t, hash1, hash2)
	assert.NotEqual(t, hash1, hash3)
	assert.NotEqual(t, hash1, hash4)
	assert.NotEqual(t, hash2, hash3)
	assert.NotEqual(t, hash2, hash4)
	assert.NotEqual(t, hash3, hash4)
}

func TestAmIElected(t *testing.T) {
	hash := sha512.Sum384([]byte("freon testing"))
	tests := []struct {
		Elected bool
		MyID    uint16
		Party   []uint16
	}{
		{true, 1, []uint16{1}},

		{true, 1, []uint16{1, 2}},
		{false, 2, []uint16{1, 2}},
		{false, 0xffff, []uint16{1, 2}},

		{false, 1, []uint16{1, 2, 3}},
		{true, 2, []uint16{1, 2, 3}},
		{false, 3, []uint16{1, 2, 3}},

		{true, 1, []uint16{1, 2, 3, 4}},
		{false, 2, []uint16{1, 2, 3, 4}},
		{false, 3, []uint16{1, 2, 3, 4}},
		{false, 4, []uint16{1, 2, 3, 4}},

		{false, 1, []uint16{1, 2, 3, 4, 5}},
		{false, 2, []uint16{1, 2, 3, 4, 5}},
		{true, 3, []uint16{1, 2, 3, 4, 5}},
		{false, 4, []uint16{1, 2, 3, 4, 5}},
		{false, 5, []uint16{1, 2, 3, 4, 5}},

		{false, 1, []uint16{1, 2, 3, 4, 5, 6}},
		{false, 2, []uint16{1, 2, 3, 4, 5, 6}},
		{false, 3, []uint16{1, 2, 3, 4, 5, 6}},
		{false, 4, []uint16{1, 2, 3, 4, 5, 6}},
		{true, 5, []uint16{1, 2, 3, 4, 5, 6}},
		{false, 6, []uint16{1, 2, 3, 4, 5, 6}},

		{false, 1, []uint16{1, 2, 3, 4, 5, 6, 7}},
		{false, 2, []uint16{1, 2, 3, 4, 5, 6, 7}},
		{false, 3, []uint16{1, 2, 3, 4, 5, 6, 7}},
		{false, 4, []uint16{1, 2, 3, 4, 5, 6, 7}},
		{false, 5, []uint16{1, 2, 3, 4, 5, 6, 7}},
		{true, 6, []uint16{1, 2, 3, 4, 5, 6, 7}},
		{false, 7, []uint16{1, 2, 3, 4, 5, 6, 7}},

		{false, 1, []uint16{1, 2, 3, 4, 5, 6, 7, 8}},
		{false, 2, []uint16{1, 2, 3, 4, 5, 6, 7, 8}},
		{false, 3, []uint16{1, 2, 3, 4, 5, 6, 7, 8}},
		{false, 4, []uint16{1, 2, 3, 4, 5, 6, 7, 8}},
		{true, 5, []uint16{1, 2, 3, 4, 5, 6, 7, 8}},
		{false, 6, []uint16{1, 2, 3, 4, 5, 6, 7, 8}},
		{false, 7, []uint16{1, 2, 3, 4, 5, 6, 7, 8}},
		{false, 8, []uint16{1, 2, 3, 4, 5, 6, 7, 8}},
	}
	for _, tt := range tests {
		res := internal.AmIElected(hash[:], tt.MyID, tt.Party)
		if tt.Elected != res {
			fmt.Printf("id = %d, ids = %d\n", tt.MyID, tt.Party)
		}
		assert.Equal(t, res, tt.Elected)
	}
}

func TestUniqueID(t *testing.T) {
	id1, err := internal.UniqueID()
	assert.NoError(t, err)
	assert.Len(t, id1, 48)

	id2, err := internal.UniqueID()
	assert.NoError(t, err)
	assert.Len(t, id2, 48)

	assert.NotEqual(t, id1, id2)
}

func TestSelectIndex(t *testing.T) {
	hash := sha512.Sum384([]byte("freon testing"))
	index := internal.SelectIndex(hash[:], 8)
	assert.Equal(t, uint64(4), index)
}

func TestPartyToUint16(t *testing.T) {
	party := party.IDSlice{5, 7, 6}
	slice := internal.PartyToUint16(party)
	assert.Equal(t, []uint16{5, 6, 7}, slice)
}

func TestHexBEToUint16(t *testing.T) {
	val, err := internal.HexBEToUint16("0100")
	assert.NoError(t, err)
	assert.Equal(t, uint16(256), val)

	val, err = internal.HexBEToUint16("ffff")
	assert.NoError(t, err)
	assert.Equal(t, uint16(65535), val)
}

func TestUint16ToHexBE(t *testing.T) {
	hex := internal.Uint16ToHexBE(256)
	assert.Equal(t, "0100", hex)

	hex = internal.Uint16ToHexBE(65535)
	assert.Equal(t, "ffff", hex)
}
