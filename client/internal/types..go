package internal

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
	Status bool `json:"status"`
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
