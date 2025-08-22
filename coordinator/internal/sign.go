package internal

import (
	"crypto/subtle"
	"database/sql"
	"errors"
)

func NewSignGroup(db *sql.DB, groupUid string, hash string, openssh bool, namespace string) (string, error) {
	// Unique ID (192 bits entropy)
	uid, err := UniqueID()
	if err != nil {
		return "", err
	}
	uid = "c_" + uid

	groupData, err := GetGroupData(db, groupUid)
	if err != nil {
		return "", err
	}

	stmt, err := db.Prepare("INSERT INTO ceremonies (uid, groupid, hash, openssh, opensshnamespace) VALUES (?, ?, ?, ?, ?)")
	if err != nil {
		return "", err
	}
	_, err = stmt.Exec(uid, groupData.DbId, hash, openssh)
	if err != nil {
		return "", err
	}

	return uid, nil
}

func JoinSignCeremony(db *sql.DB, ceremonyID, hash string, myPartyID uint16) (int64, error) {
	ceremonyData, err := GetCeremonyData(db, ceremonyID)
	if err != nil {
		return 0, err
	}

	stmt, err := db.Prepare(`
		SELECT par.id
		FROM participants par
		JOIN groups g ON par.grroupid = g.id
		JOIN ceremonies c ON c.groupid = g.id 
		WHERE c.id = ? AND par.partyid = ?
		`)
	if err != nil {
		return 0, err
	}
	var participantId int64
	err = stmt.QueryRow(ceremonyID, myPartyID).Scan(&participantId)
	if err != nil {
		return 0, err
	}
	if subtle.ConstantTimeCompare([]byte(ceremonyData.Hash), []byte(hash)) != 1 {
		return 0, errors.New("hash mismatch")
	}
	return participantId, nil
}

func PollSignCeremony(db *sql.DB, ceremonyID string, myPartyID uint16) (PollSignResponse, error) {
	ceremonyData, err := GetCeremonyData(db, ceremonyID)
	if err != nil {
		return PollSignResponse{}, err
	}

	stmt, err := db.Prepare(`
		SELECT par.id, g.id, g.uid, g.threshold
		FROM participants par
		JOIN groups g ON par.grroupid = g.id
		JOIN ceremonies c ON c.groupid = g.id 
		WHERE c.id = ? AND par.partyid = ?
		`)
	if err != nil {
		return PollSignResponse{}, err
	}
	defer stmt.Close()

	var participantId int64
	var groupId int64
	var groupUid string
	var threshold uint16
	err = stmt.QueryRow(ceremonyID, ceremonyData.GroupID).Scan(&participantId, &groupId, &groupUid, &threshold)
	if err != nil {
		return PollSignResponse{}, err
	}

	stmt2, err := db.Prepare(`
		SELECT par.partyid, g.uid
		FROM participants par
		JOIN groups g ON par.grroupid = g.id
		JOIN ceremonies c ON c.groupid = g.id 
		WHERE c.id = ? AND par.partyid != ?
	`)
	if err != nil {
		return PollSignResponse{}, err
	}
	defer stmt2.Close()
	rows, err := stmt.Query(ceremonyID)
	if err != nil {
		return PollSignResponse{}, err
	}
	defer rows.Close()

	var otherParties []uint16
	for rows.Next() {
		var otherPartyId uint16
		if err := rows.Scan(&otherPartyId); err != nil {
			return PollSignResponse{}, err
		}
		otherParties = append(otherParties, otherPartyId)
	}

	return PollSignResponse{
		GroupID:      groupUid,
		MyPartyID:    myPartyID,
		Threshold:    threshold,
		OtherParties: otherParties,
	}, nil
}

func AddSignMessage(db *sql.DB, ceremonyUid string, myPartyID uint16, message []byte) (FreonSignMessage, error) {
	ceremony, err := GetCeremonyData(db, ceremonyUid)
	if err != nil {
		return FreonSignMessage{}, err
	}

	group, err := GetGroupByID(db, ceremony.GroupID)
	if err != nil {
		return FreonSignMessage{}, err
	}

	participant, err := GetParticipantID(db, group.Uid, myPartyID)
	if err != nil {
		return FreonSignMessage{}, err
	}
	msg := FreonSignMessage{
		DbId:       int64(0),
		CeremonyID: ceremony.DbId,
		Sender:     participant,
		Message:    message,
	}
	id, err := InsertSignMessage(db, msg)
	if err != nil {
		return FreonSignMessage{}, err
	}
	msg.DbId = id
	return msg, nil
}

func SetSignature(db *sql.DB, ceremonyUid, sig string) error {
	ceremony, err := GetCeremonyData(db, ceremonyUid)
	if err != nil {
		return err
	}

	if ceremony.Signature != nil {
		return errors.New("signature is already defined")
	}

	return FinalizeSignature(db, ceremony, sig)
}
