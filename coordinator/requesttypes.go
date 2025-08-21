package main

type ResponseMainPage struct {
	Message string `json:"message"`
}
type ResponseErrorPage struct {
	Error string `json:"message"`
}

type InitKeyGenRequest struct {
	Participants uint16 `json:"n"`
	Threshold    uint16 `json:"t"`
}
type InitKeyGenResponse struct {
	GroupID string `json:"group-id"`
}

type JoinKeyGenRequest struct {
	GroupID string `json:"group-id"`
}
type JoinKeyGenResponse struct {
	Status    bool   `json:"status"`
	MyPartyID uint16 `json:"my-party-id"`
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

type InitSignRequest struct {
	GroupID     string `json:"group-id"`
	MessageHash string `json:"hash"`
	OpenSSH     bool   `json:"openssh"`
}
type InitSignResponse struct {
	CeremonyID string `json:"ceremony-id"`
}

type JoinSignRequest struct {
	CeremonyID  string `json:"ceremony-id"`
	MessageHash string `json:"hash"`
	MyPartyID   uint16 `json:"party-id"`
}
type JoinSignResponse struct {
	Status bool `json:"status"`
}

type PollSignRequest struct {
	CeremonyID string  `json:"ceremony-id"`
	PartyID    *uint16 `json:"party-id"`
}

type SignMessageRequest struct {
	CeremonyID string
	MyPartyID  uint16
	Message    string
	LastSeen   int64
}
type SignMessageResponse struct {
	LatestMessageID int64
	Messages        []string
}
