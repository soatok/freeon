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
	_, err = stmt.Exec(uid, groupData.DbId, hash, openssh, namespace)
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
	if !ceremonyData.Active {
		return 0, errors.New("ceremony is not active or does not exist")
	}

	stmt, err := db.Prepare(`
		SELECT par.id
		FROM participants par
		JOIN keygroups g ON par.groupid = g.id
		JOIN ceremonies c ON c.groupid = g.id 
		WHERE c.uid = ? AND par.partyid = ?
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

	// Now insert the player
	insert, err := db.Prepare(`INSERT INTO players (ceremonyid, participantid) VALUES (?, ?)`)
	if err != nil {
		return 0, err
	}
	_, err = insert.Exec(ceremonyData.DbId, participantId)
	if err != nil {
		return 0, err
	}
	return participantId, nil
}

func PollSignCeremony(db *sql.DB, ceremonyID string, myPartyID uint16) (PollSignResponse, error) {
	ceremonyData, err := GetCeremonyData(db, ceremonyID)
	if err != nil {
		return PollSignResponse{}, err
	}

	groupData, err := GetGroupByID(db, ceremonyData.GroupID)
	if err != nil {
		return PollSignResponse{}, err
	}

	players, err := GetCeremonyPlayers(db, ceremonyID)
	if err != nil {
		return PollSignResponse{}, err
	}

	var otherParties []uint16
	for _, player := range players {
		if player.PartyID != myPartyID {
			otherParties = append(otherParties, player.PartyID)
		}
	}

	return PollSignResponse{
		GroupID:      groupData.Uid,
		MyPartyID:    myPartyID,
		Threshold:    groupData.Threshold,
		OtherParties: otherParties,
	}, nil
}

func AddSignMessage(db *sql.DB, ceremonyUid string, myPartyID uint16, message []byte) (FreeonSignMessage, error) {
	ceremony, err := GetCeremonyData(db, ceremonyUid)
	if err != nil {
		return FreeonSignMessage{}, err
	}
	if !ceremony.Active {
		return FreeonSignMessage{}, errors.New("ceremony is not active or does not exist")
	}

	group, err := GetGroupByID(db, ceremony.GroupID)
	if err != nil {
		return FreeonSignMessage{}, err
	}

	participant, err := GetParticipantID(db, group.Uid, myPartyID)
	if err != nil {
		return FreeonSignMessage{}, err
	}
	msg := FreeonSignMessage{
		DbId:       int64(0),
		CeremonyID: ceremony.DbId,
		Sender:     participant,
		Message:    message,
	}
	id, err := InsertSignMessage(db, msg)
	if err != nil {
		return FreeonSignMessage{}, err
	}
	msg.DbId = id
	return msg, nil
}

func SetSignature(db *sql.DB, ceremonyUid, sig string) error {
	ceremony, err := GetCeremonyData(db, ceremonyUid)
	if err != nil {
		return err
	}

	if !ceremony.Active {
		return errors.New("ceremony is not active or does not exist")
	}
	if ceremony.Signature != nil {
		return errors.New("signature is already defined")
	}

	return FinalizeSignature(db, ceremony, sig)
}

func GetSignature(db *sql.DB, ceremonyUid string) (string, error) {
	ceremony, err := GetCeremonyData(db, ceremonyUid)
	if err != nil {
		return "", err
	}
	if ceremony.Signature != nil {
		return *ceremony.Signature, nil
	}
	return "", errors.New("signature not found")
}
