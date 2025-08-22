package internal

import (
	"database/sql"
	"encoding/hex"
	"errors"
)

func DbEnsureTablesExist(db *sql.DB) error {
	createTable := `
    CREATE TABLE IF NOT EXISTS keygroups (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        uid TEXT NOT NULL,
		participants INTEGER,
		threshold INTEGER,
		publickey TEXT NULL
    );
	CREATE TABLE IF NOT EXISTS participants (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
        groupid INTEGER REFERENCES keygroups(id), 
		uid TEXT NOT NULL,
		partyid INTEGER
	);
	CREATE TABLE IF NOT EXISTS ceremonies (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		groupid INTEGER REFERENCES keygroups(id),
		uid TEXT NOT NULL,
		active BOOLEAN DEFAULT TRUE,
		openssh BOOLEAN DEFAULT FALSE,
		opensshnamespace TEXT NULL,
		hash TEXT,
		signature TEXT NULL
	);
	CREATE TABLE IF NOT EXISTS players (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		ceremonyid INTEGER REFERENCES ceremonies(id),
		participantid INTEGER REFERENCES participants(id)
	);
	CREATE TABLE IF NOT EXISTS keygenmsg (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
        groupid INTEGER REFERENCES keygroups(id),
		sender INTEGER REFERENCES participants(id),
		message TEXT
	);
	CREATE TABLE IF NOT EXISTS signmsg (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		ceremonyid INTEGER REFERENCES ceremonies(id),
		sender INTEGER REFERENCES participants(id),
		message TEXT
	);
	`

	_, err := db.Exec(createTable)
	if err != nil {
		return err
	}
	return nil
}

// Get the row ID for a given group
func GetGroupRowId(db *sql.DB, groupUid string) (int, error) {
	stmt, err := db.Prepare("SELECT id FROM keygroups WHERE uid = ?")
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	var id int
	err = stmt.QueryRow(groupUid).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// Get the row ID for a given group
func GetGroupData(db *sql.DB, groupUid string) (FreonGroup, error) {
	stmt, err := db.Prepare("SELECT id, threshold, participants, publicKey FROM keygroups WHERE uid = ?")
	if err != nil {
		return FreonGroup{}, err
	}
	defer stmt.Close()

	var id int64
	var threshold uint16
	var participants uint16
	var publicKey *string
	err = stmt.QueryRow(groupUid).Scan(&id, &threshold, &participants, &publicKey)
	if err != nil {
		return FreonGroup{}, err
	}
	return FreonGroup{
		DbId:         id,
		Uid:          groupUid,
		Participants: participants,
		Threshold:    threshold,
		PublicKey:    publicKey,
	}, nil
}
func GetGroupByID(db *sql.DB, groupID int64) (FreonGroup, error) {
	stmt, err := db.Prepare("SELECT id, uid, threshold, participants, publicKey FROM keygroups WHERE id = ?")
	if err != nil {
		return FreonGroup{}, err
	}
	defer stmt.Close()

	var id int64
	var uid string
	var threshold uint16
	var participants uint16
	var publicKey *string
	err = stmt.QueryRow(groupID).Scan(&id, &uid, &threshold, &participants, &publicKey)
	if err != nil {
		return FreonGroup{}, err
	}
	return FreonGroup{
		DbId:         id,
		Uid:          uid,
		Participants: participants,
		Threshold:    threshold,
		PublicKey:    publicKey,
	}, nil
}

// Get all of the participants for a group
func GetGroupParticipants(db *sql.DB, groupUid string) ([]FreonParticipant, error) {
	stmt, err := db.Prepare(`
		SELECT
			p.id,
			g.id AS groupid,
			p.uid,
			p.partyid
		FROM keygroups g 
		JOIN participants p ON p.groupid = g.id
		WHERE g.uid = ?
	`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	rows, err := stmt.Query(groupUid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var participants []FreonParticipant
	for rows.Next() {
		var dbId int64
		var groupId int64
		var uid string
		var partyid uint16
		if err := rows.Scan(&dbId, &groupId, &uid, &partyid); err != nil {
			return nil, err
		}
		p := FreonParticipant{
			DbId:    dbId,
			GroupID: groupId,
			Uid:     uid,
			PartyID: partyid,
		}
		participants = append(participants, p)
	}
	return participants, nil
}

func GetParticipantID(db *sql.DB, groupUid string, myPartyID uint16) (int64, error) {
	stmt, err := db.Prepare(`
		SELECT p.id 
		FROM participants p 
		JOIN keygroups g ON p.groupid = g.id 
		WHERE g.uid = ? AND p.partyid = ?
		`)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	var id int64
	err = stmt.QueryRow(groupUid, myPartyID).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func GetCeremonyData(db *sql.DB, ceremonyID string) (FreonCeremonies, error) {
	stmt, err := db.Prepare(`SELECT
		id, groupid, active, hash, signature, openssh, opensshnamespace
		FROM ceremonies
		WHERE uid = ?`)
	if err != nil {
		return FreonCeremonies{}, err
	}
	defer stmt.Close()

	var id int64
	var groupid int64
	var active bool
	var hash string
	var signature *string
	var openssh bool
	var opensshnamespace *string
	err = stmt.QueryRow(ceremonyID).Scan(&id, &groupid, &active, &hash, &signature, &openssh, &opensshnamespace)
	if err != nil {
		return FreonCeremonies{}, err
	}
	return FreonCeremonies{
		DbId:             id,
		GroupID:          groupid,
		Uid:              ceremonyID,
		Active:           active,
		Hash:             hash,
		Signature:        signature,
		OpenSSH:          openssh,
		OpenSSHNamespace: opensshnamespace,
	}, nil
}

func GetRecentCeremonies(db *sql.DB, groupID string, limit, offset int64) ([]FreonCeremonySummary, error) {
	stmt, err := db.Prepare(`SELECT
		c.uid, c.hash, c.signature, c.openssh, c.opensshnamespace, c.active
		FROM ceremonies c
		JOIN keygroups g ON c.groupid = g.id
		WHERE c.active AND g.uid = ?
		ORDER BY c.id DESC
		LIMIT ? OFFSET ?`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(groupID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var results []FreonCeremonySummary
	for rows.Next() {
		var ceremonyID string
		var hash string
		var signature *string
		var openssh bool
		var opensshnamespace *string
		var active bool
		if err := rows.Scan(&ceremonyID, &hash, &signature, &openssh, &opensshnamespace, &active); err != nil {
			return nil, err
		}
		var ns string
		if opensshnamespace == nil {
			ns = ""
		} else {
			ns = *opensshnamespace
		}
		row := FreonCeremonySummary{
			Uid:              ceremonyID,
			Active:           active,
			Hash:             hash,
			Signature:        signature,
			OpenSSH:          openssh,
			OpenSSHNamespace: ns,
		}
		results = append(results, row)
	}
	return results, nil
}

func GetCeremonyPlayers(db *sql.DB, ceremonyID string) ([]FreonPlayers, error) {
	stmt, err := db.Prepare(`
		SELECT
			x.id,
			x.ceremonyid,
			x.participantid,
			p.partyid
		FROM players x 
		JOIN participants p ON x.participantid = p.id
		JOIN ceremonies c ON x.ceremonyid = c.id
		WHERE c.uid = ?`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	rows, err := stmt.Query(ceremonyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var players []FreonPlayers
	for rows.Next() {
		var dbId int64
		var ceremonyId int64
		var participantId int64
		var partyId uint16
		if err := rows.Scan(&dbId, &ceremonyId, &participantId, &partyId); err != nil {
			return nil, err
		}
		p := FreonPlayers{
			DbId:          dbId,
			CeremonyID:    ceremonyId,
			ParticipantID: participantId,
			PartyID:       partyId,
		}
		players = append(players, p)
	}
	return players, nil
}

func GetKeygenMessagesSince(db *sql.DB, groupUid string, lastSeen int64) ([]FreonKeygenMessage, error) {
	stmt, err := db.Prepare(`
		SELECT
			msg.id,
			msg.groupid,
			msg.sender,
			msg.message
		FROM keygroups g 
		JOIN keygenmsg msg ON msg.groupid = g.id
		JOIN participants p ON msg.sender = p.id
		WHERE g.uid = ? AND msg.id > ?
	`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	rows, err := stmt.Query(groupUid, lastSeen)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []FreonKeygenMessage
	for rows.Next() {
		var id int64
		var group int64
		var sender int64
		var messageHex string
		if err := rows.Scan(&id, &group, &sender, &messageHex); err != nil {
			return nil, err
		}
		messageBody, err := hex.DecodeString(messageHex)
		if err != nil {
			return nil, err
		}
		msg := FreonKeygenMessage{
			DbId:    id,
			GroupID: group,
			Sender:  sender,
			Message: messageBody,
		}
		messages = append(messages, msg)
	}
	return messages, nil
}

func GetSignMessagesSince(db *sql.DB, ceremonyUid string, lastSeen int64) ([]FreonSignMessage, error) {
	stmt, err := db.Prepare(`
		SELECT
			msg.id,
			msg.ceremonyid,
			msg.sender,
			msg.message
		FROM ceremonies c
		JOIN signmsg msg ON msg.ceremonyid = c.id
		JOIN participants p ON msg.sender = p.id
		WHERE c.uid = ? AND msg.id > ?
	`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	rows, err := stmt.Query(ceremonyUid, lastSeen)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []FreonSignMessage
	for rows.Next() {
		var id int64
		var ceremony int64
		var sender int64
		var messageHex string
		if err := rows.Scan(&id, &ceremony, &sender, &messageHex); err != nil {
			return nil, err
		}
		messageBody, err := hex.DecodeString(messageHex)
		if err != nil {
			return nil, err
		}
		msg := FreonSignMessage{
			DbId:       id,
			CeremonyID: ceremony,
			Sender:     sender,
			Message:    messageBody,
		}
		messages = append(messages, msg)
	}
	return messages, nil
}

func InsertGroup(db *sql.DB, g FreonGroup) (int64, error) {
	stmt, err := db.Prepare(`INSERT INTO keygroups (uid, participants, threshold) VALUES (?, ?, ?)`)
	if err != nil {
		return 0, err
	}
	res, err := stmt.Exec(g.Uid, g.Participants, g.Threshold)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

func InsertCeremony(db *sql.DB, c FreonCeremonies) (int64, error) {
	stmt, err := db.Prepare(`INSERT INTO ceremonies (groupid, uid, active, hash) VALUES (?, ?, ?, ?)`)
	if err != nil {
		return 0, err
	}
	res, err := stmt.Exec(c.GroupID, c.Uid, c.Active, c.Hash)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

func InsertParticipant(db *sql.DB, p FreonParticipant) (int64, error) {
	stmt, err := db.Prepare(`INSERT INTO participants (groupid, uid, partyid) VALUES (?, ?, ?)`)
	if err != nil {
		return 0, err
	}
	res, err := stmt.Exec(p.GroupID, p.Uid, p.PartyID)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

func InsertPlayer(db *sql.DB, p FreonPlayers) (int64, error) {
	stmt, err := db.Prepare(`INSERT INTO players (ceremonyid, participantid) VALUES (?, ?)`)
	if err != nil {
		return 0, err
	}
	res, err := stmt.Exec(p.CeremonyID, p.ParticipantID)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

func InsertKeygenMessage(db *sql.DB, m FreonKeygenMessage) (int64, error) {
	stmt, err := db.Prepare(`INSERT INTO keygenmsg (groupid, sender, message) VALUES (?, ?, ?)`)
	if err != nil {
		return 0, err
	}
	stateHex := hex.EncodeToString(m.Message)
	res, err := stmt.Exec(m.GroupID, m.Sender, stateHex)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

func InsertSignMessage(db *sql.DB, m FreonSignMessage) (int64, error) {
	stmt, err := db.Prepare(`INSERT INTO signmsg (ceremonyid, sender, message) VALUES (?, ?, ?)`)
	if err != nil {
		return 0, err
	}
	stateHex := hex.EncodeToString(m.Message)
	res, err := stmt.Exec(m.CeremonyID, m.Sender, stateHex)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

func FinalizeGroup(db *sql.DB, g FreonGroup) error {
	if g.PublicKey == nil {
		return errors.New("public key is not stored in FreonGroup struct")
	}
	stmt, err := db.Prepare(`UPDATE keygroups SET publicKey = ? WHERE id = ?`)
	if err != nil {
		return err
	}
	_, err = stmt.Exec(g.PublicKey, g.DbId)
	if err != nil {
		return err
	}
	return nil
}

func FinalizeSignature(db *sql.DB, c FreonCeremonies, sig string) error {
	if !c.Active {
		return errors.New("group is already finalized")
	}
	stmt, err := db.Prepare(`UPDATE ceremonies SET signature = ?, active = FALSE WHERE id = ?`)
	if err != nil {
		return err
	}
	_, err = stmt.Exec(sig, c.DbId)
	return err
}

func TerminateCeremony(db *sql.DB, ceremonyUid string) error {
	stmt, err := db.Prepare(`UPDATE ceremonies SET active = FALSE WHERE uid = ?`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	res, err := stmt.Exec(ceremonyUid)
	if err != nil {
		return err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count < 1 {
		return errors.New("no ceremony found with that UID")
	}
	return nil
}
