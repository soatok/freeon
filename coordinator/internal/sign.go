package internal

import (
	"database/sql"
)

func NewSignGroup(db *sql.DB, groupUid string, hash string) (string, error) {
	// Unique ID (192 bits entropy)
	uid, err := UniqueID()
	if err != nil {
		return "", err
	}
	uid = "c_" + uid

	groupId, err := GetGroupData(db, groupUid)
	if err != nil {
		return "", err
	}

	stmt, err := db.Prepare("INSERT INTO ceremonies (uid, groupid, hash) VALUES (?, ?, ?)")
	if err != nil {
		return "", err
	}
	_, err = stmt.Exec(uid, groupId, hash)
	if err != nil {
		return "", err
	}

	return uid, nil
}
