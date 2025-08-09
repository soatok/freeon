package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/alexedwards/scs/v2"
	_ "github.com/mattn/go-sqlite3" // SQLite driver
	"github.com/soatok/freon/coordinator/internal"
	_ "github.com/taurusgroup/frost-ed25519/pkg/frost"
)

type ResponseMainPage struct {
	Message string `json:"message"`
}

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
	db, err = sql.Open("sqlite3", "./example.db")
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

	mux := http.NewServeMux()

	http.HandleFunc("/", indexPage)

	http.HandleFunc("/keygen/create", createKeygen)
	http.HandleFunc("/keygen/join", joinKeygen)
	http.HandleFunc("/keygen/poll", pollKeygen)
	http.HandleFunc("/keygen/send", sendKeygen)

	http.HandleFunc("/sign/create", createSign)
	http.HandleFunc("/sign/join", joinSign)
	http.HandleFunc("/sign/poll", pollSign)
	http.HandleFunc("/sign/send", sendSign)

	http.HandleFunc("/terminate", terminateSign)
	http.ListenAndServe(serverConfig.Hostname, sessionManager.LoadAndSave(mux))
}

func sendError(w http.ResponseWriter, e error) {
	response := ResponseMainPage{Message: e.Error()}
	w.WriteHeader(http.StatusInternalServerError)
	h := w.Header()
	h.Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func indexPage(w http.ResponseWriter, r *http.Request) {
	response := ResponseMainPage{Message: "Freon Coordinator v0.0.0"}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

type InitKeyGenRequest struct {
	Participants uint16 `json:"n"`
	Threshold    uint16 `json:"t"`
}
type InitKeyGenResponse struct {
	GroupID string `json:"group-id"`
}

// Initialize a key generation ceremony
func createKeygen(w http.ResponseWriter, r *http.Request) {
	var req InitKeyGenRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		sendError(w, err)
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
	json.NewEncoder(w).Encode(response)
}

func joinKeygen(w http.ResponseWriter, r *http.Request) {

}

func pollKeygen(w http.ResponseWriter, r *http.Request) {

}

func sendKeygen(w http.ResponseWriter, r *http.Request) {

}

func createSign(w http.ResponseWriter, r *http.Request) {

}

func joinSign(w http.ResponseWriter, r *http.Request) {

}

func pollSign(w http.ResponseWriter, r *http.Request) {

}

func sendSign(w http.ResponseWriter, r *http.Request) {

}

func terminateSign(w http.ResponseWriter, r *http.Request) {

}
