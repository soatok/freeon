package internal_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/soatok/freon/client/internal"
	"github.com/stretchr/testify/assert"
)

func TestDuct(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/keygen/create":
			var req internal.InitKeyGenRequest
			json.NewDecoder(r.Body).Decode(&req)
			resp := internal.InitKeyGenResponse{GroupID: "test-group"}
			json.NewEncoder(w).Encode(resp)
		case "/keygen/join":
			var req internal.JoinKeyGenRequest
			json.NewDecoder(r.Body).Decode(&req)
			resp := internal.JoinKeyGenResponse{Status: true, MyPartyID: 1}
			json.NewEncoder(w).Encode(resp)
		case "/keygen/poll":
			var req internal.PollKeyGenRequest
			json.NewDecoder(r.Body).Decode(&req)
			resp := internal.PollKeyGenResponse{GroupID: "test-group", PartySize: 2, Threshold: 2, OtherParties: []uint16{2}}
			json.NewEncoder(w).Encode(resp)
		case "/sign/create":
			var req internal.InitSignRequest
			json.NewDecoder(r.Body).Decode(&req)
			resp := internal.InitSignResponse{CeremonyID: "test-ceremony"}
			json.NewEncoder(w).Encode(resp)
		case "/sign/join":
			var req internal.JoinSignRequest
			json.NewDecoder(r.Body).Decode(&req)
			resp := internal.JoinSignResponse{Status: true}
			json.NewEncoder(w).Encode(resp)
		case "/sign/poll":
			var req internal.PollSignRequest
			json.NewDecoder(r.Body).Decode(&req)
			resp := internal.PollSignResponse{GroupID: "test-group", Threshold: 2, OtherParties: []uint16{2}}
			json.NewEncoder(w).Encode(resp)
		case "/sign/list":
			var req internal.ListSignRequest
			json.NewDecoder(r.Body).Decode(&req)
			resp := internal.ListSignResponse{Ceremonies: []internal.FreonCeremonySummary{{Uid: "test-ceremony"}}}
			json.NewEncoder(w).Encode(resp)
		case "/keygen/send":
			var req internal.KeyGenMessageRequest
			json.NewDecoder(r.Body).Decode(&req)
			resp := internal.KeyGenMessageResponse{LatestMessageID: 1, Messages: []string{"test-message"}}
			json.NewEncoder(w).Encode(resp)
		case "/sign/send":
			var req internal.SignMessageRequest
			json.NewDecoder(r.Body).Decode(&req)
			resp := internal.SignMessageResponse{LatestMessageID: 1, Messages: []string{"test-message"}}
			json.NewEncoder(w).Encode(resp)
		case "/keygen/finalize":
			var req internal.KeygenFinalRequest
			json.NewDecoder(r.Body).Decode(&req)
			w.WriteHeader(http.StatusOK)
		case "/sign/finalize":
			var req internal.SignFinalRequest
			json.NewDecoder(r.Body).Decode(&req)
			w.WriteHeader(http.StatusOK)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	// Test DuctInitKeyGenCeremony
	initKeyGenReq := internal.InitKeyGenRequest{Participants: 2, Threshold: 2}
	initKeyGenResp, err := internal.DuctInitKeyGenCeremony(server.URL, initKeyGenReq)
	assert.NoError(t, err)
	assert.Equal(t, "test-group", initKeyGenResp.GroupID)

	// Test DuctJoinKeyGenCeremony
	joinKeyGenReq := internal.JoinKeyGenRequest{GroupID: "test-group"}
	joinKeyGenResp, err := internal.DuctJoinKeyGenCeremony(server.URL, joinKeyGenReq)
	assert.NoError(t, err)
	assert.True(t, joinKeyGenResp.Status)
	assert.Equal(t, uint16(1), joinKeyGenResp.MyPartyID)

	// Test DuctPollKeyGenCeremony
	var partyID uint16 = 1
	pollKeyGenReq := internal.PollKeyGenRequest{GroupID: "test-group", PartyID: &partyID}
	pollKeyGenResp, err := internal.DuctPollKeyGenCeremony(server.URL, pollKeyGenReq)
	assert.NoError(t, err)
	assert.Equal(t, "test-group", pollKeyGenResp.GroupID)
	assert.Equal(t, uint16(2), pollKeyGenResp.PartySize)
	assert.Equal(t, uint16(2), pollKeyGenResp.Threshold)
	assert.Equal(t, []uint16{2}, pollKeyGenResp.OtherParties)

	// Test DuctInitSignCeremony
	initSignReq := internal.InitSignRequest{GroupID: "test-group", MessageHash: "test-hash"}
	initSignResp, err := internal.DuctInitSignCeremony(server.URL, initSignReq)
	assert.NoError(t, err)
	assert.Equal(t, "test-ceremony", initSignResp.CeremonyID)

	// Test DuctJoinSignCeremony
	joinSignReq := internal.JoinSignRequest{CeremonyID: "test-ceremony", MessageHash: "test-hash", MyPartyID: 1}
	joinSignResp, err := internal.DuctJoinSignCeremony(server.URL, joinSignReq)
	assert.NoError(t, err)
	assert.True(t, joinSignResp.Status)

	// Test DuctPollSignCeremony
	pollSignReq := internal.PollSignRequest{CeremonyID: "test-ceremony", PartyID: &partyID}
	pollSignResp, err := internal.DuctPollSignCeremony(server.URL, pollSignReq)
	assert.NoError(t, err)
	assert.Equal(t, "test-group", pollSignResp.GroupID)
	assert.Equal(t, uint16(2), pollSignResp.Threshold)
	assert.Equal(t, []uint16{2}, pollSignResp.OtherParties)

	// Test DuctSignList
	listSignReq := internal.ListSignRequest{GroupID: "test-group"}
	listSignResp, err := internal.DuctSignList(server.URL, listSignReq)
	assert.NoError(t, err)
	assert.Len(t, listSignResp.Ceremonies, 1)
	assert.Equal(t, "test-ceremony", listSignResp.Ceremonies[0].Uid)

	// Test DuctKeygenProtocolMessage
	keygenMsgReq := internal.KeyGenMessageRequest{GroupID: "test-group", MyPartyID: 1, Message: "test-message"}
	keygenMsgResp, err := internal.DuctKeygenProtocolMessage(server.URL, keygenMsgReq)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), keygenMsgResp.LatestMessageID)
	assert.Equal(t, []string{"test-message"}, keygenMsgResp.Messages)

	// Test DuctSignProtocolMessage
	signMsgReq := internal.SignMessageRequest{CeremonyID: "test-ceremony", MyPartyID: 1, Message: "test-message"}
	signMsgResp, err := internal.DuctSignProtocolMessage(server.URL, signMsgReq)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), signMsgResp.LatestMessageID)
	assert.Equal(t, []string{"test-message"}, signMsgResp.Messages)

	// Test DuctKeygenFinalize
	keygenFinalReq := internal.KeygenFinalRequest{GroupID: "test-group", MyPartyID: 1, PublicKey: "test-pk"}
	err = internal.DuctKeygenFinalize(server.URL, keygenFinalReq)
	assert.NoError(t, err)

	// Test DuctSignFinalize
	signFinalReq := internal.SignFinalRequest{CeremonyID: "test-ceremony", MyPartyID: 1, Signature: "test-sig"}
	err = internal.DuctSignFinalize(server.URL, signFinalReq)
	assert.NoError(t, err)
}
