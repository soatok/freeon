// The naming convention here might provoke a deep sigh, but "duct" is an HVAC thing that delivers cold air.
// So this is, consequently, the part of the Client code that talks to the Coordinator.
//
// ...
//
// Look, you all know what you signed up for when you saw my dumb puns on Fedi.
//
// The code here is still part of the internal package, but I like logically separating files.
package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
)

var httpClient *http.Client = nil

func InitializeHttpClient() error {
	if httpClient == nil {
		jar, err := cookiejar.New(nil)
		if err != nil {
			return err
		}
		httpClient = &http.Client{
			Jar: jar,
		}
	}
	return nil
}

// If we change the backend API, we will change this function to accomodate it
func GetApiEndpoint(host string, feature string) (string, error) {
	u, err := url.Parse(host)
	if err != nil {
		return "", err
	}

	switch feature {
	case "InitKeyGenCeremony":
		u.Path = "/keygen/create"
	case "JoinKeyGenCeremony":
		u.Path = "/keygen/join"
	case "PollKeyGenCeremony":
		u.Path = "/keygen/poll"
	case "SendKeygenMessage":
		u.Path = "/keygen/send"
	case "InitSignCeremony":
		u.Path = "/sign/create"
	case "PollSignCeremony":
		u.Path = "/sign/poll"
	case "JoinSignCeremony":
		u.Path = "/sign/join"
	case "SendSignMessage":
		u.Path = "/sign/send"
	case "TerminateSignCeremony":
		u.Path = "/sign/terminate"
	default:
		return "", fmt.Errorf("unknown feature: %s", feature)
	}

	return u.String(), nil
}

// The network handler for creating a key ceremony
func DuctInitKeyGenCeremony(host string, req InitKeyGenRequest) (InitKeyGenResponse, error) {
	err := InitializeHttpClient()
	if err != nil {
		return InitKeyGenResponse{}, err
	}
	uri, err := GetApiEndpoint(host, "InitKeyGenCeremony")
	if err != nil {
		return InitKeyGenResponse{}, err
	}
	body, _ := json.Marshal(req)
	resp, err := httpClient.Post(uri, "application/json", bytes.NewReader(body))
	if err != nil {
		return InitKeyGenResponse{}, err
	}
	defer resp.Body.Close()

	var response InitKeyGenResponse
	json.NewDecoder(resp.Body).Decode(&response)
	return response, nil
}

// The network handler for joining a key ceremony
func DuctJoinKeyGenCeremony(host string, req JoinKeyGenRequest) (JoinKeyGenResponse, error) {
	err := InitializeHttpClient()
	if err != nil {
		return JoinKeyGenResponse{}, err
	}
	uri, err := GetApiEndpoint(host, "JoinKeyGenCeremony")
	if err != nil {
		return JoinKeyGenResponse{}, err
	}
	body, _ := json.Marshal(req)
	resp, err := httpClient.Post(uri, "application/json", bytes.NewReader(body))
	if err != nil {
		return JoinKeyGenResponse{}, err
	}
	defer resp.Body.Close()

	var response JoinKeyGenResponse
	json.NewDecoder(resp.Body).Decode(&response)
	return response, nil
}

// Poll a keygen ceremony until enough participants have joined
func DuctPollKeyGenCeremony(host string, req PollKeyGenRequest) (PollKeyGenResponse, error) {
	err := InitializeHttpClient()
	if err != nil {
		return PollKeyGenResponse{}, err
	}
	uri, err := GetApiEndpoint(host, "PollKeyGenCeremony")
	if err != nil {
		return PollKeyGenResponse{}, err
	}
	body, _ := json.Marshal(req)
	resp, err := httpClient.Post(uri, "application/json", bytes.NewReader(body))
	if err != nil {
		return PollKeyGenResponse{}, err
	}
	defer resp.Body.Close()

	var response PollKeyGenResponse
	json.NewDecoder(resp.Body).Decode(&response)
	return response, nil
}

// We're kicking off a signing ceremony
func DuctInitSignCeremony(host string, req InitSignRequest) (InitSignResponse, error) {
	err := InitializeHttpClient()
	if err != nil {
		return InitSignResponse{}, err
	}
	uri, err := GetApiEndpoint(host, "InitSignCeremony")
	if err != nil {
		return InitSignResponse{}, err
	}
	body, _ := json.Marshal(req)
	resp, err := httpClient.Post(uri, "application/json", bytes.NewReader(body))
	if err != nil {
		return InitSignResponse{}, err
	}
	defer resp.Body.Close()

	var response InitSignResponse
	json.NewDecoder(resp.Body).Decode(&response)
	return response, nil
}

func DuctJoinSignCeremony(host string, req JoinSignRequest) (JoinSignResponse, error) {
	err := InitializeHttpClient()
	if err != nil {
		return JoinSignResponse{}, err
	}
	uri, err := GetApiEndpoint(host, "JoinSignCeremony")
	if err != nil {
		return JoinSignResponse{}, err
	}
	body, _ := json.Marshal(req)

	resp, err := httpClient.Post(uri, "application/json", bytes.NewReader(body))
	if err != nil {
		return JoinSignResponse{}, err
	}
	defer resp.Body.Close()

	var response JoinSignResponse
	json.NewDecoder(resp.Body).Decode(&response)
	return response, nil
}

func DuctPollSignCeremony(host string, req PollSignRequest) (PollSignResponse, error) {
	err := InitializeHttpClient()
	if err != nil {
		return PollSignResponse{}, err
	}
	uri, err := GetApiEndpoint(host, "PollSignCeremony")
	if err != nil {
		return PollSignResponse{}, err
	}
	body, _ := json.Marshal(req)
	resp, err := httpClient.Post(uri, "application/json", bytes.NewReader(body))
	if err != nil {
		return PollSignResponse{}, err
	}
	defer resp.Body.Close()

	var response PollSignResponse
	json.NewDecoder(resp.Body).Decode(&response)
	return response, nil
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

// Send keygen protocol messages
func DuctKeygenProtocolMessage(host string, req KeyGenMessageRequest) (KeyGenMessageResponse, error) {
	err := InitializeHttpClient()
	if err != nil {
		return KeyGenMessageResponse{}, err
	}
	uri, err := GetApiEndpoint(host, "SendKeygenMessage")
	if err != nil {
		return KeyGenMessageResponse{}, err
	}
	body, _ := json.Marshal(req)
	resp, err := httpClient.Post(uri, "application/json", bytes.NewReader(body))
	if err != nil {
		return KeyGenMessageResponse{}, err
	}
	defer resp.Body.Close()

	var response KeyGenMessageResponse
	json.NewDecoder(resp.Body).Decode(&response)
	return response, nil
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

// Send sign protocol messages
func DuctSignProtocolMessage(host string, req SignMessageRequest) (SignMessageResponse, error) {
	err := InitializeHttpClient()
	if err != nil {
		return SignMessageResponse{}, err
	}
	uri, err := GetApiEndpoint(host, "SendSignMessage")
	if err != nil {
		return SignMessageResponse{}, err
	}
	body, _ := json.Marshal(req)
	resp, err := httpClient.Post(uri, "application/json", bytes.NewReader(body))
	if err != nil {
		return SignMessageResponse{}, err
	}
	defer resp.Body.Close()

	var response SignMessageResponse
	json.NewDecoder(resp.Body).Decode(&response)
	return response, nil
}
