package internal_test

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/soatok/freon/coordinator/internal"
	"github.com/stretchr/testify/assert"
)

func setupTestDBForKeygen(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", "file::memory:?cache=shared")
	assert.NoError(t, err)
	err = internal.DbEnsureTablesExist(db)
	assert.NoError(t, err)
	t.Cleanup(func() {
		db.Close()
	})
	return db
}

func TestNewKeyGroup(t *testing.T) {
	db := setupTestDBForKeygen(t)
	uid, err := internal.NewKeyGroup(db, 3, 2)
	assert.NoError(t, err)
	assert.NotEmpty(t, uid)

	group, err := internal.GetGroupData(db, uid)
	assert.NoError(t, err)
	assert.Equal(t, uint16(3), group.Participants)
	assert.Equal(t, uint16(2), group.Threshold)
}

func TestAddParticipant(t *testing.T) {
	db := setupTestDBForKeygen(t)
	uid, err := internal.NewKeyGroup(db, 2, 2)
	assert.NoError(t, err)

	p1, err := internal.AddParticipant(db, uid)
	assert.NoError(t, err)
	assert.Equal(t, uint16(1), p1.PartyID)

	p2, err := internal.AddParticipant(db, uid)
	assert.NoError(t, err)
	assert.Equal(t, uint16(2), p2.PartyID)

	// Group is full
	_, err = internal.AddParticipant(db, uid)
	assert.Error(t, err)
}

func TestAddKeyGenMessage(t *testing.T) {
	db := setupTestDBForKeygen(t)
	g_uid, err := internal.NewKeyGroup(db, 2, 2)
	assert.NoError(t, err)
	p, err := internal.AddParticipant(db, g_uid)
	assert.NoError(t, err)

	msg, err := internal.AddKeyGenMessage(db, g_uid, p.PartyID, []byte("test message"))
	assert.NoError(t, err)
	assert.NotZero(t, msg.DbId)

	msgs, err := internal.GetKeygenMessagesSince(db, g_uid, 0)
	assert.NoError(t, err)
	assert.Len(t, msgs, 1)
	assert.Equal(t, []byte("test message"), msgs[0].Message)
}

func TestSetGroupPublicKey(t *testing.T) {
	db := setupTestDBForKeygen(t)
	g_uid, err := internal.NewKeyGroup(db, 2, 2)
	assert.NoError(t, err)

	err = internal.SetGroupPublicKey(db, g_uid, "test_pk")
	assert.NoError(t, err)

	group, err := internal.GetGroupData(db, g_uid)
	assert.NoError(t, err)
	assert.Equal(t, "test_pk", *group.PublicKey)

	// Cannot set it again
	err = internal.SetGroupPublicKey(db, g_uid, "test_pk_2")
	assert.Error(t, err)
}
