package internal_test

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/soatok/freon/coordinator/internal"
	"github.com/stretchr/testify/assert"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", "file::memory:?cache=shared")
	assert.NoError(t, err)
	t.Cleanup(func() {
		db.Close()
	})
	return db
}

func TestDbEnsureTablesExist(t *testing.T) {
	db := setupTestDB(t)
	err := internal.DbEnsureTablesExist(db)
	assert.NoError(t, err)

	// Check if tables were created
	tables := []string{
		"keygroups",
		"participants",
		"ceremonies",
		"players",
		"keygenmsg",
		"signmsg",
	}
	for _, table := range tables {
		var name string
		err := db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&name)
		assert.NoError(t, err)
		assert.Equal(t, table, name)
	}
}

func TestParticipantFunctions(t *testing.T) {
	db := setupTestDB(t)
	err := internal.DbEnsureTablesExist(db)
	assert.NoError(t, err)

	// Insert a group
	group := internal.FreonGroup{
		Uid:          "test_group",
		Participants: 3,
		Threshold:    2,
	}
	gid, err := internal.InsertGroup(db, group)
	assert.NoError(t, err)

	// Insert a participant
	p := internal.FreonParticipant{
		GroupID: gid,
		Uid:     "test_participant",
		PartyID: 1,
		State:   []byte("state1"),
	}
	pid, err := internal.InsertParticipant(db, p)
	assert.NoError(t, err)
	assert.NotZero(t, pid)
	p.DbId = pid

	// GetGroupParticipants
	participants, err := internal.GetGroupParticipants(db, "test_group")
	assert.NoError(t, err)
	assert.Len(t, participants, 1)
	assert.Equal(t, p.Uid, participants[0].Uid)

	// GetParticipantID
	participantID, err := internal.GetParticipantID(db, "test_group", 1)
	assert.NoError(t, err)
	assert.Equal(t, pid, participantID)

	// UpdateParticipantState
	p.State = []byte("state2")
	err = internal.UpdateParticipantState(db, p)
	assert.NoError(t, err)

	participants, err = internal.GetGroupParticipants(db, "test_group")
	assert.NoError(t, err)
	assert.Len(t, participants, 1)
	assert.Equal(t, p.State, participants[0].State)
}

func TestCeremonyAndMessageFunctions(t *testing.T) {
	db := setupTestDB(t)
	err := internal.DbEnsureTablesExist(db)
	assert.NoError(t, err)

	// Insert a group and participant
	group := internal.FreonGroup{Uid: "g", Participants: 2, Threshold: 2}
	gid, err := internal.InsertGroup(db, group)
	assert.NoError(t, err)
	p := internal.FreonParticipant{GroupID: gid, Uid: "p", PartyID: 1}
	pid, err := internal.InsertParticipant(db, p)
	assert.NoError(t, err)

	// Insert a ceremony
	c := internal.FreonCeremonies{
		GroupID: gid,
		Uid:     "c",
		Active:  true,
		Hash:    "hash",
	}
	cid, err := internal.InsertCeremony(db, c)
	assert.NoError(t, err)
	c.DbId = cid

	// GetCeremonyData
	cData, err := internal.GetCeremonyData(db, "c")
	assert.NoError(t, err)
	assert.Equal(t, c.Uid, cData.Uid)

	// Insert a player
	player := internal.FreonPlayers{
		CeremonyID:    cid,
		ParticipantID: pid,
	}
	_, err = internal.InsertPlayer(db, player)
	assert.NoError(t, err)

	// GetCeremonyPlayers
	players, err := internal.GetCeremonyPlayers(db, "c")
	assert.NoError(t, err)
	assert.Len(t, players, 1)
	assert.Equal(t, pid, players[0].ParticipantID)

	// GetRecentCeremonies
	recent, err := internal.GetRecentCeremonies(db, "g", 10, 0)
	assert.NoError(t, err)
	assert.Len(t, recent, 1)
	assert.Equal(t, "c", recent[0].Uid)

	// FinalizeSignature
	err = internal.FinalizeSignature(db, c, "sig")
	assert.NoError(t, err)
	cData, err = internal.GetCeremonyData(db, "c")
	assert.NoError(t, err)
	assert.False(t, cData.Active)
	assert.Equal(t, "sig", *cData.Signature)

	// InsertKeygenMessage
	km := internal.FreonKeygenMessage{
		GroupID: gid,
		Sender:  pid,
		Message: []byte("keygen message"),
	}
	kmid, err := internal.InsertKeygenMessage(db, km)
	assert.NoError(t, err)
	assert.NotZero(t, kmid)

	// GetKeygenMessagesSince
	kms, err := internal.GetKeygenMessagesSince(db, "g", 0)
	assert.NoError(t, err)
	assert.Len(t, kms, 1)
	assert.Equal(t, km.Message, kms[0].Message)

	// InsertSignMessage
	sm := internal.FreonSignMessage{
		CeremonyID: cid,
		Sender:     pid,
		Message:    []byte("sign message"),
	}
	smid, err := internal.InsertSignMessage(db, sm)
	assert.NoError(t, err)
	assert.NotZero(t, smid)

	// GetSignMessagesSince
	sms, err := internal.GetSignMessagesSince(db, "c", 0)
	assert.NoError(t, err)
	assert.Len(t, sms, 1)
	assert.Equal(t, sm.Message, sms[0].Message)
}

func TestGroupFunctions(t *testing.T) {
	db := setupTestDB(t)
	err := internal.DbEnsureTablesExist(db)
	assert.NoError(t, err)

	// Insert a group
	group := internal.FreonGroup{
		Uid:          "test_group",
		Participants: 3,
		Threshold:    2,
	}
	id, err := internal.InsertGroup(db, group)
	assert.NoError(t, err)
	assert.NotZero(t, id)
	group.DbId = id

	// GetGroupRowId
	rowId, err := internal.GetGroupRowId(db, "test_group")
	assert.NoError(t, err)
	assert.Equal(t, id, int64(rowId))

	// GetGroupData
	groupData, err := internal.GetGroupData(db, "test_group")
	assert.NoError(t, err)
	assert.Equal(t, group.Uid, groupData.Uid)
	assert.Equal(t, group.Participants, groupData.Participants)
	assert.Equal(t, group.Threshold, groupData.Threshold)

	// GetGroupByID
	groupByID, err := internal.GetGroupByID(db, id)
	assert.NoError(t, err)
	assert.Equal(t, group.Uid, groupByID.Uid)

	// FinalizeGroup
	publicKey := "test_public_key"
	group.PublicKey = &publicKey
	err = internal.FinalizeGroup(db, group)
	assert.NoError(t, err)

	finalizedGroup, err := internal.GetGroupData(db, "test_group")
	assert.NoError(t, err)
	assert.NotNil(t, finalizedGroup.PublicKey)
	assert.Equal(t, publicKey, *finalizedGroup.PublicKey)
}
