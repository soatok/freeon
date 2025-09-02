package internal_test

import (
	"testing"

	"github.com/soatok/freeon/coordinator/internal"
	"github.com/stretchr/testify/assert"
)

func TestUniqueID(t *testing.T) {
	id1, err := internal.UniqueID()
	assert.NoError(t, err)
	assert.Len(t, id1, 48)

	id2, err := internal.UniqueID()
	assert.NoError(t, err)
	assert.Len(t, id2, 48)

	// These should never be equal
	assert.NotEqual(t, id1, id2)
}
