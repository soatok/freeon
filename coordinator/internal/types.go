package internal

type FreeonGroup struct {
	DbId         int64
	Uid          string
	Participants uint16
	Threshold    uint16
	PublicKey    *string
}

type FreeonParticipant struct {
	DbId    int64
	GroupID int64
	Uid     string
	PartyID uint16
	State   []byte
}

type FreeonKeygenMessage struct {
	DbId    int64
	GroupID int64
	Sender  int64
	Message []byte
}

type FreeonCeremonies struct {
	DbId             int64
	GroupID          int64
	Uid              string
	Active           bool
	Hash             string
	Signature        *string
	OpenSSH          bool
	OpenSSHNamespace *string
}

// For public lists of signing ceremonies
type FreeonCeremonySummary struct {
	Uid              string
	Active           bool
	Hash             string
	Signature        *string
	OpenSSH          bool
	OpenSSHNamespace string
}

type FreeonPlayers struct {
	DbId          int64
	CeremonyID    int64
	ParticipantID int64
	PartyID       uint16
}

type FreeonSignMessage struct {
	DbId       int64
	CeremonyID int64
	Sender     int64
	Message    []byte
}

type PollSignResponse struct {
	GroupID      string   `json:"group-id"`
	MyPartyID    uint16   `json:"party-id"`
	Threshold    uint16   `json:"t"`
	OtherParties []uint16 `json:"parties"`
}
