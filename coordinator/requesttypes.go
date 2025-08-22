package main

import "github.com/soatok/freon/coordinator/internal"

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
	Namespace   string `json:"openssh-namespace"`
}
type InitSignResponse struct {
	CeremonyID string `json:"ceremony-id"`
}

type ListSignRequest struct {
	GroupID string `json:"group-id"`
	Limit   *int64 `json:"limit"`
	Offset  *int64 `json:"offset"`
}
type ListSignResponse struct {
	Ceremonies []internal.FreonCeremonySummary
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

type PollSignRequest struct {
	CeremonyID string  `json:"ceremony-id"`
	PartyID    *uint16 `json:"party-id"`
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

type VapidResponse struct {
	Status string `json:"status"`
}
