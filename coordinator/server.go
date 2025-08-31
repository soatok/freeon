package main

import (
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/alexedwards/scs/v2"
	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
	"github.com/soatok/freon/coordinator/internal"
)

var sessionManager *scs.SessionManager
var db *sql.DB

// The Coordinator starts here
func main() {
	serverConfig, err := internal.LoadServerConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err.Error())
		os.Exit(1)
	}

	// Database
	// Open database (creates file if it doesn't exist)
	db, err = sql.Open("sqlite3", serverConfig.Database)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err.Error())
		os.Exit(1)
	}
	defer db.Close()

	// Ensure foreign keys
	_, err = db.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err.Error())
		os.Exit(1)
	}
	internal.DbEnsureTablesExist(db)

	// Session storage
	sessionManager = scs.New()
	sessionManager.Lifetime = 12 * time.Hour

	http.HandleFunc("/", indexPage)

	http.HandleFunc("/keygen/create", createKeygen)
	http.HandleFunc("/keygen/join", joinKeygen)
	http.HandleFunc("/keygen/poll", pollKeygen)
	http.HandleFunc("/keygen/send", sendKeygen)
	http.HandleFunc("/keygen/get-messages", getKeygenMessages)
	http.HandleFunc("/keygen/finalize", finalizeKeygen)

	http.HandleFunc("/sign/create", createSign)
	http.HandleFunc("/sign/list", listSign)
	http.HandleFunc("/sign/join", joinSign)
	http.HandleFunc("/sign/poll", pollSign)
	http.HandleFunc("/sign/send", sendSign)
	http.HandleFunc("/sign/get-messages", getSignMessages)
	http.HandleFunc("/sign/finalize", finalizeSign)
	http.HandleFunc("/sign/get", getSign)

	http.HandleFunc("/terminate", terminateSign)
	http.ListenAndServe(serverConfig.Hostname, sessionManager.LoadAndSave(http.DefaultServeMux))
}

// Handler for error pages
func sendError(w http.ResponseWriter, e error) {
	// TODO - not disclose this once the code is stable!
	response := ResponseErrorPage{Error: e.Error()}
	w.WriteHeader(http.StatusInternalServerError)
	h := w.Header()
	h.Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Handler for index page
func indexPage(w http.ResponseWriter, r *http.Request) {
	response := ResponseMainPage{Message: "Freon Coordinator v0.0.0"}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Initialize a key generation ceremony
func createKeygen(w http.ResponseWriter, r *http.Request) {
	var req InitKeyGenRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		sendError(w, err)
		return
	}
	if req.Threshold > req.Participants {
		sendError(w, errors.New("threshold cannot exceeed party size"))
		return
	}
	uid, err := internal.NewKeyGroup(db, req.Participants, req.Threshold)
	if err != nil {
		sendError(w, err)
		return
	}
	response := InitKeyGenResponse{
		GroupID: uid,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Join a key ceremony as a participant
func joinKeygen(w http.ResponseWriter, r *http.Request) {
	var req JoinKeyGenRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		sendError(w, err)
		return
	}
	participant, err := internal.AddParticipant(db, req.GroupID)
	if err != nil {
		sendError(w, err)
		return
	}
	response := JoinKeyGenResponse{
		Status:    true,
		MyPartyID: participant.PartyID,
	}
	// TODO: use session storage
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Poll a keygen ceremony to get the status
func pollKeygen(w http.ResponseWriter, r *http.Request) {
	var req PollKeyGenRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		sendError(w, err)
		return
	}
	group, err := internal.GetGroupData(db, req.GroupID)
	if err != nil {
		sendError(w, err)
		return
	}
	participants, err := internal.GetGroupParticipants(db, req.GroupID)
	if err != nil {
		sendError(w, err)
		return
	}

	// Assemble list of "others"
	var others []uint16
	if req.PartyID == nil {
		for _, p := range participants {
			others = append(others, p.PartyID)
		}
	} else {
		for _, p := range participants {
			if p.PartyID != *req.PartyID {
				others = append(others, p.PartyID)
			}
		}
	}

	response := PollKeyGenResponse{
		GroupID:      group.Uid,
		MyPartyID:    req.PartyID,
		OtherParties: others,
		Threshold:    group.Threshold,
		PartySize:    group.Participants,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Get messages for a keygen ceremony
func getKeygenMessages(w http.ResponseWriter, r *http.Request) {
	var req KeyGenMessageRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		sendError(w, err)
		return
	}
	inbox, err := internal.GetKeygenMessagesSince(db, req.GroupID, req.LastSeen)
	if err != nil {
		sendError(w, err)
		return
	}
	// Get a new maximum
	var latestID = req.LastSeen
	var messages []string
	for _, m := range inbox {
		messages = append(messages, hex.EncodeToString(m.Message))
		if m.DbId > latestID {
			latestID = m.DbId
		}
	}

	// Let's queue up the messages
	response := KeyGenMessageResponse{
		LatestMessageID: latestID,
		Messages:        messages,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&response)
}

// Send a message to participate in a keygen ceremony
func sendKeygen(w http.ResponseWriter, r *http.Request) {
	var req KeyGenMessageRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		sendError(w, err)
		return
	}
	msg, err := hex.DecodeString(req.Message)
	if err != nil {
		sendError(w, err)
		return
	}

	// First, add the new message to the database.
	_, err = internal.AddKeyGenMessage(db, req.GroupID, req.MyPartyID, msg)
	if err != nil {
		sendError(w, err)
		return
	}

	// Now, get all messages since the client's last seen ID.
	// This will include the message we just added, and any from other clients.
	inbox, err := internal.GetKeygenMessagesSince(db, req.GroupID, req.LastSeen)
	if err != nil {
		sendError(w, err)
		return
	}

	// Build the response
	var latestID = req.LastSeen
	var messages []string
	for _, m := range inbox {
		messages = append(messages, hex.EncodeToString(m.Message))
		if m.DbId > latestID {
			latestID = m.DbId
		}
	}

	response := KeyGenMessageResponse{
		LatestMessageID: latestID,
		Messages:        messages,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&response)
}

// Create a signing ceremony
func createSign(w http.ResponseWriter, r *http.Request) {
	var req InitSignRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		sendError(w, err)
		return
	}
	uid, err := internal.NewSignGroup(db, req.GroupID, req.MessageHash, req.OpenSSH, req.Namespace)
	if err != nil {
		sendError(w, err)
		return
	}
	response := InitSignResponse{
		CeremonyID: uid,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Join an existing signing ceremony
func joinSign(w http.ResponseWriter, r *http.Request) {
	var req JoinSignRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		sendError(w, err)
		return
	}
	_, err = internal.JoinSignCeremony(db, req.CeremonyID, req.MessageHash, req.MyPartyID)
	if err != nil {
		sendError(w, err)
		return
	}
	response := JoinSignResponse{
		Status: true,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Poll the status of a signing ceremony
func pollSign(w http.ResponseWriter, r *http.Request) {
	var req PollSignRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		sendError(w, err)
		return
	}
	var partyID uint16 = 0
	if req.PartyID != nil {
		partyID = *req.PartyID
	}
	response, err := internal.PollSignCeremony(db, req.CeremonyID, partyID)
	if err != nil {
		sendError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&response)

}

// Get messages for a signing ceremony
func getSignMessages(w http.ResponseWriter, r *http.Request) {
	var req SignMessageRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		sendError(w, err)
		return
	}
	inbox, err := internal.GetSignMessagesSince(db, req.CeremonyID, req.LastSeen)
	if err != nil {
		sendError(w, err)
		return
	}
	// Get a new maximum
	var max = req.LastSeen
	var messages []string
	for _, m := range inbox {
		if m.DbId >= max {
			max = m.DbId
		}
		messages = append(messages, hex.EncodeToString(m.Message))
	}

	// Let's queue up the messages
	response := SignMessageResponse{
		LatestMessageID: max,
		Messages:        messages,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Send a message to a signing ceremony
func sendSign(w http.ResponseWriter, r *http.Request) {
	var req SignMessageRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		sendError(w, err)
		return
	}
	msg, err := hex.DecodeString(req.Message)
	if err != nil {
		sendError(w, err)
		return
	}

	// First, add the new message to the database.
	_, err = internal.AddSignMessage(db, req.CeremonyID, req.MyPartyID, msg)
	if err != nil {
		sendError(w, err)
		return
	}

	// Now, get all messages since the client's last seen ID.
	inbox, err := internal.GetSignMessagesSince(db, req.CeremonyID, req.LastSeen)
	if err != nil {
		sendError(w, err)
		return
	}

	// Build the response
	var latestID = req.LastSeen
	var messages []string
	for _, m := range inbox {
		messages = append(messages, hex.EncodeToString(m.Message))
		if m.DbId > latestID {
			latestID = m.DbId
		}
	}

	response := SignMessageResponse{
		LatestMessageID: latestID,
		Messages:        messages,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&response)
}

// Store the final public key for the group
func finalizeKeygen(w http.ResponseWriter, r *http.Request) {
	var req KeygenFinalRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		sendError(w, err)
		return
	}
	err = internal.SetGroupPublicKey(db, req.GroupID, req.PublicKey)
	if err != nil {
		sendError(w, err)
		return
	}

	// Return a vapid response.
	response := VapidResponse{
		Status: "OK",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Store the final signature for the ceremony
func finalizeSign(w http.ResponseWriter, r *http.Request) {
	var req SignFinalRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		sendError(w, err)
		return
	}
	err = internal.SetSignature(db, req.CeremonyID, req.Signature)
	if err != nil {
		sendError(w, err)
		return
	}

	// Return a vapid response.
	response := VapidResponse{
		Status: "OK",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

}

// List the receent signing ceremonies for a given key group
func listSign(w http.ResponseWriter, r *http.Request) {
	var req ListSignRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		sendError(w, err)
		return
	}

	var limit int64
	var offset int64

	// Default values
	if req.Limit == nil {
		limit = 10
	} else {
		limit = *req.Limit
	}
	if req.Offset == nil {
		offset = 0
	} else {
		offset = *req.Offset
	}

	list, err := internal.GetRecentCeremonies(db, req.GroupID, limit, offset)
	if err != nil {
		sendError(w, err)
		return
	}

	// Return a vapid response.
	response := ListSignResponse{
		Ceremonies: list,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func getSign(w http.ResponseWriter, r *http.Request) {
	var req GetSignRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		sendError(w, err)
		return
	}
	signature, err := internal.GetSignature(db, req.CeremonyID)
	if err != nil {
		sendError(w, err)
		return
	}

	response := GetSignResponse{
		Signature: signature,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func terminateSign(w http.ResponseWriter, r *http.Request) {
	var req TerminateRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		sendError(w, err)
		return
	}

	err = internal.TerminateCeremony(db, req.CeremonyID)
	if err != nil {
		sendError(w, err)
		return
	}

	response := VapidResponse{
		Status: "OK",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
