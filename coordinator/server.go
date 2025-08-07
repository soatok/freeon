package coordinator

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/alexedwards/scs/v2"
)

type ResponseMainPage struct {
	Message string `json:"message"`
}

var sessionManager *scs.SessionManager

// The Coordinator starts here
func main() {
	sessionManager = scs.New()
	sessionManager.Lifetime = 12 * time.Hour

	mux := http.NewServeMux()

	http.HandleFunc("/", indexPage)
	http.ListenAndServe(":8462", sessionManager.LoadAndSave(mux))
}

func indexPage(w http.ResponseWriter, r *http.Request) {
	response := ResponseMainPage{Message: "Freon Coordinator v0,0.0"}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
