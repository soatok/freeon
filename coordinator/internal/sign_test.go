package internal_test

import (
	"database/sql"
	"testing"

	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
	"github.com/soatok/freeon/coordinator/internal"
	"github.com/stretchr/testify/assert"
)

func setupTestDBForSign(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", "file::memory:?cache=shared")
	assert.NoError(t, err)
	err = internal.DbEnsureTablesExist(db)
	assert.NoError(t, err)
	t.Cleanup(func() {
		db.Close()
	})
	return db
}

func TestNewSignGroup(t *testing.T) {
	db := setupTestDBForSign(t)
	g_uid, err := internal.NewKeyGroup(db, 2, 2)
	assert.NoError(t, err)

	c_uid, err := internal.NewSignGroup(db, g_uid, "hash", false, "")
	assert.NoError(t, err)
	assert.NotEmpty(t, c_uid)

	c, err := internal.GetCeremonyData(db, c_uid)
	assert.NoError(t, err)
	assert.Equal(t, "hash", c.Hash)
}

func TestJoinSignCeremony(t *testing.T) {
	db := setupTestDBForSign(t)
	g_uid, err := internal.NewKeyGroup(db, 2, 2)
	assert.NoError(t, err)
	p, err := internal.AddParticipant(db, g_uid)
	assert.NoError(t, err)
	c_uid, err := internal.NewSignGroup(db, g_uid, "hash", false, "")
	assert.NoError(t, err)

	pid, err := internal.JoinSignCeremony(db, c_uid, "hash", p.PartyID)
	assert.NoError(t, err)
	assert.Equal(t, p.DbId, pid)

	// Wrong hash
	_, err = internal.JoinSignCeremony(db, c_uid, "wrong_hash", p.PartyID)
	assert.Error(t, err)
}

func TestPollSignCeremony(t *testing.T) {
	db := setupTestDBForSign(t)
	g_uid, err := internal.NewKeyGroup(db, 2, 2)
	assert.NoError(t, err)
	p1, err := internal.AddParticipant(db, g_uid)
	assert.NoError(t, err)
	p2, err := internal.AddParticipant(db, g_uid)
	assert.NoError(t, err)
	c_uid, err := internal.NewSignGroup(db, g_uid, "hash", false, "")
	assert.NoError(t, err)
	_, err = internal.JoinSignCeremony(db, c_uid, "hash", p1.PartyID)
	assert.NoError(t, err)
	_, err = internal.JoinSignCeremony(db, c_uid, "hash", p2.PartyID)
	assert.NoError(t, err)

	poll, err := internal.PollSignCeremony(db, c_uid, p1.PartyID)
	assert.NoError(t, err)
	assert.Equal(t, g_uid, poll.GroupID)

	// Ensure party 1 sees party 2
	assert.Len(t, poll.OtherParties, 1)
	assert.Equal(t, p2.PartyID, poll.OtherParties[0])
}

func TestAddSignMessage(t *testing.T) {
	db := setupTestDBForSign(t)
	g_uid, err := internal.NewKeyGroup(db, 2, 2)
	assert.NoError(t, err)
	p, err := internal.AddParticipant(db, g_uid)
	assert.NoError(t, err)
	c_uid, err := internal.NewSignGroup(db, g_uid, "hash", false, "")
	assert.NoError(t, err)

	msg, err := internal.AddSignMessage(db, c_uid, p.PartyID, []byte("test message"))
	assert.NoError(t, err)
	assert.NotZero(t, msg.DbId)

	msgs, err := internal.GetSignMessagesSince(db, c_uid, 0)
	assert.NoError(t, err)
	assert.Len(t, msgs, 1)
	assert.Equal(t, []byte("test message"), msgs[0].Message)
}

func TestSetSignature(t *testing.T) {
	db := setupTestDBForSign(t)
	g_uid, err := internal.NewKeyGroup(db, 2, 2)
	assert.NoError(t, err)
	c_uid, err := internal.NewSignGroup(db, g_uid, "hash", false, "")
	assert.NoError(t, err)

	err = internal.SetSignature(db, c_uid, "sig")
	assert.NoError(t, err)

	c, err := internal.GetCeremonyData(db, c_uid)
	assert.NoError(t, err)
	assert.Equal(t, "sig", *c.Signature)

	// Cannot set it again
	err = internal.SetSignature(db, c_uid, "sig2")
	assert.Error(t, err)
}
