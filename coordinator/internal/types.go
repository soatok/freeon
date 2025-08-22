package internal

type FreonGroup struct {
	DbId         int64
	Uid          string
	Participants uint16
	Threshold    uint16
	PublicKey    *string
}

type FreonParticipant struct {
	DbId    int64
	GroupID int64
	Uid     string
	PartyID uint16
	State   []byte
}

type FreonKeygenMessage struct {
	DbId    int64
	GroupID int64
	Sender  int64
	Message []byte
}

type FreonCeremonies struct {
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
type FreonCeremonySummary struct {
	Uid              string
	Active           bool
	Hash             string
	Signature        *string
	OpenSSH          bool
	OpenSSHNamespace string
}

type FreonPlayers struct {
	DbId          int64
	CeremonyID    int64
	ParticipantID int64
	PartyID       uint16
	State         []byte
}

type FreonSignMessage struct {
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
