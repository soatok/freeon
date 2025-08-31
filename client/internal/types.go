package internal

// Shares from keygen ceremonies are stored (encrypted)
type Shares struct {
	Host           string            `json:"host"`
	GroupID        string            `json:"group-id"`
	PublicKey      string            `json:"public-key"`
	MyPartyID      uint16            `json:"my-party-id"`
	EncryptedShare string            `json:"encrypted-share"`
	PublicShares   map[string]string `json:"public-shares"`
}

// This may expand in future versions
type FreonConfig struct {
	Shares []Shares `json:"shares"`
}

//------- Request/Response --------//
type InitKeyGenRequest struct {
	Participants uint16 `json:"n"`
	Threshold    uint16 `json:"t"`
}
type InitKeyGenResponse struct {
	GroupID string `json:"group-id"`
}

type PollKeyGenRequest struct {
	GroupID string  `json:"group-id"`
	PartyID *uint16 `json:"party-id,omitempty"`
}
type PollKeyGenResponse struct {
	GroupID      string   `json:"group-id"`
	MyPartyID    *uint16  `json:"party-id"`
	OtherParties []uint16 `json:"parties"`
	Threshold    uint16   `json:"t"`
	PartySize    uint16   `json:"n"`
}

type InitSignRequest struct {
	GroupID     string `json:"group-id"`
	MessageHash string `json:"hash"`
	OpenSSH     bool   `json:"openssh"`
	Namespace   string `json:"openssh-namespace"`
}
type InitSignResponse struct {
	CeremonyID string `json:"ceremony-id"`
}

type PollSignRequest struct {
	CeremonyID string  `json:"ceremony-id"`
	PartyID    *uint16 `json:"party-id"`
}
type PollSignResponse struct {
	GroupID      string   `json:"group-id"`
	MyPartyID    uint16   `json:"party-id"`
	Threshold    uint16   `json:"t"`
	OtherParties []uint16 `json:"parties"`
}

type JoinKeyGenRequest struct {
	GroupID string `json:"group-id"`
}
type JoinKeyGenResponse struct {
	Status    bool   `json:"status"`
	MyPartyID uint16 `json:"my-party-id"`
}

type JoinSignRequest struct {
	CeremonyID  string `json:"ceremony-id"`
	MessageHash string `json:"hash"`
	MyPartyID   uint16 `json:"party-id"`
}
type JoinSignResponse struct {
	Status    bool   `json:"status"`
	OpenSSH   bool   `json:"openssh"`
	Namespace string `json:"openssh-namespace"`
}

type SendKeyGenRequest struct {
	GroupID    string `json:"group-id"`
	MyPartyID  uint16 `json:"party-id"`
	LastIDSeen int64  `json:"last-seen-id"`
	Message    string `json:"message"`
}
type SendKeyGenResponse struct {
	Status   bool     `json:"status"`
	Messages []string `json:"messages"`
}

type KeyGenMessageRequest struct {
	GroupID   string
	Message   string
	MyPartyID uint16
	LastSeen  int64
}
type KeyGenMessageResponse struct {
	LatestMessageID int64
	Messages        []string
}

type SignMessageRequest struct {
	CeremonyID string `json:"ceremony-id"`
	MyPartyID  uint16 `json:"party-id"`
	Message    string `json:"message"`
	LastSeen   int64  `json:"last-seen"`
}
type SignMessageResponse struct {
	LatestMessageID int64    `json:"last-seen"`
	Messages        []string `json:"messages"`
}

type KeygenFinalRequest struct {
	GroupID   string `json:"group-id"`
	MyPartyID uint16 `json:"party-id"`
	PublicKey string `json:"public-key"`
}

type SignFinalRequest struct {
	CeremonyID string `json:"ceremony-id"`
	MyPartyID  uint16 `json:"party-id"`
	Signature  string `json:"signature"`
}

type GetSignRequest struct {
	CeremonyID string `json:"ceremony"`
}

type GetSignResponse struct {
	Signature string `json:"signature"`
}

type TerminateRequest struct {
	CeremonyID string `json:"ceremony-id"`
}

type ResponseErrorPage struct {
	Error string `json:"message"`
}

type VapidResponse struct {
	Status string `json:"status"`
}

type FreonCeremonySummary struct {
	Uid              string
	Active           bool
	Hash             string
	Signature        *string
	OpenSSH          bool
	OpenSSHNamespace string
}
type ListSignRequest struct {
	GroupID string `json:"group-id"`
	Limit   int64  `json:"limit"`
	Offset  int64  `json:"offset"`
}
type ListSignResponse struct {
	Ceremonies []FreonCeremonySummary
}
