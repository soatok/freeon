package internal_test

import (
	"testing"

	"github.com/soatok/freon/client/internal"
	"github.com/stretchr/testify/assert"
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
