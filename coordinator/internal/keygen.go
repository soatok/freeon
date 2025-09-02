package internal

import (
	"database/sql"
	"errors"
)

// Create a new DKG group
func NewKeyGroup(db *sql.DB, partySize, threshold uint16) (string, error) {
	// Unique ID (192 bits entropy)
	uid, err := UniqueID()
	if err != nil {
		return "", err
	}
	uid = "g_" + uid

	stmt, err := db.Prepare(`INSERT INTO keygroups (uid, participants, threshold) VALUES (?, ?, ?)`)
	if err != nil {
		return "", err
	}
	_, err = stmt.Exec(uid, partySize, threshold)
	if err != nil {
		return "", err
	}

	return uid, nil
}

// Create a blank slate participant ID
func AddParticipant(db *sql.DB, groupUid string) (FreeonParticipant, error) {
	tx, err := db.Begin()
	if err != nil {
		return FreeonParticipant{}, err
	}
	defer tx.Rollback() // Rollback on error

	groupData, err := GetGroupData(tx, groupUid)
	if err != nil {
		return FreeonParticipant{}, err
	}
	participants, err := GetGroupParticipants(tx, groupUid)
	if err != nil {
		return FreeonParticipant{}, err
	}
	if len(participants) >= int(groupData.Participants) {
		return FreeonParticipant{}, errors.New("cannot add participant: group is full")
	}

	// Figure out the maximum party ID for existing participants
	var max uint16 = 0
	for _, p := range participants {
		if p.PartyID > max {
			max = p.PartyID
		}
	}
	if max == 0xFFFF {
		return FreeonParticipant{}, errors.New("cannot add participant: party ID would overflow")
	}
	nextMaxId := max + 1

	// Get a unique participant ID
	uid, err := UniqueID()
	if err != nil {
		return FreeonParticipant{}, err
	}
	uid = "p_" + uid

	p := FreeonParticipant{
		DbId:    int64(0),
		GroupID: groupData.DbId,
		Uid:     uid,
		PartyID: nextMaxId,
		State:   []byte{},
	}
	id, err := InsertParticipant(tx, p)
	if err != nil {
		return FreeonParticipant{}, err
	}
	p.DbId = id

	if err = tx.Commit(); err != nil {
		return FreeonParticipant{}, err
	}

	return p, nil
}

// Add a keygen message to the queue
func AddKeyGenMessage(db *sql.DB, groupUid string, myPartyID uint16, message []byte) (FreeonKeygenMessage, error) {
	group, err := GetGroupData(db, groupUid)
	if err != nil {
		return FreeonKeygenMessage{}, err
	}
	participant, err := GetParticipantID(db, groupUid, myPartyID)
	if err != nil {
		return FreeonKeygenMessage{}, err
	}
	msg := FreeonKeygenMessage{
		DbId:    int64(0),
		GroupID: group.DbId,
		Sender:  participant,
		Message: message,
	}
	id, err := InsertKeygenMessage(db, msg)
	if err != nil {
		return FreeonKeygenMessage{}, err
	}
	msg.DbId = id
	return msg, nil
}

func SetGroupPublicKey(db *sql.DB, groupUid string, publicKey string) error {
	group, err := GetGroupData(db, groupUid)
	if err != nil {
		return err
	}
	if group.PublicKey != nil {
		return errors.New("public key is already defined")
	}
	group.PublicKey = &publicKey
	return FinalizeGroup(db, group)
}
