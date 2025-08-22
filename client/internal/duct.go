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
	"strings"
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
	if !strings.HasPrefix(host, "http://") && !strings.HasPrefix(host, "https://") {
		host = "http://" + host
	}
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
	case "GetKeygenMessages":
		u.Path = "/keygen/get-messages"
	case "FinalizeKeygenMessage":
		u.Path = "/keygen/finalize"
	case "InitSignCeremony":
		u.Path = "/sign/create"
	case "PollSignCeremony":
		u.Path = "/sign/poll"
	case "JoinSignCeremony":
		u.Path = "/sign/join"
	case "ListSignCeremony":
		u.Path = "/sign/list"
	case "SendSignMessage":
		u.Path = "/sign/send"
	case "GetSignMessages":
		u.Path = "/sign/get-messages"
	case "FinalizeSignMessage":
		u.Path = "/sign/finalize"
	case "GetSignature":
		u.Path = "/sign/get"
	case "TerminateSignCeremony":
		u.Path = "/terminate"
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

	if resp.StatusCode != http.StatusOK {
		var errResp ResponseErrorPage
		if json.NewDecoder(resp.Body).Decode(&errResp) == nil {
			return InitKeyGenResponse{}, fmt.Errorf("request failed: %s", errResp.Error)
		}
		return InitKeyGenResponse{}, fmt.Errorf("request failed with status code: %d", resp.StatusCode)
	}

	var response InitKeyGenResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return InitKeyGenResponse{}, err
	}
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

	if resp.StatusCode != http.StatusOK {
		var errResp ResponseErrorPage
		if json.NewDecoder(resp.Body).Decode(&errResp) == nil {
			return JoinKeyGenResponse{}, fmt.Errorf("request failed: %s", errResp.Error)
		}
		return JoinKeyGenResponse{}, fmt.Errorf("request failed with status code: %d", resp.StatusCode)
	}

	var response JoinKeyGenResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return JoinKeyGenResponse{}, err
	}
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

	if resp.StatusCode != http.StatusOK {
		var errResp ResponseErrorPage
		if json.NewDecoder(resp.Body).Decode(&errResp) == nil {
			return PollKeyGenResponse{}, fmt.Errorf("request failed: %s", errResp.Error)
		}
		return PollKeyGenResponse{}, fmt.Errorf("request failed with status code: %d", resp.StatusCode)
	}

	var response PollKeyGenResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return PollKeyGenResponse{}, err
	}
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

	if resp.StatusCode != http.StatusOK {
		var errResp ResponseErrorPage
		if json.NewDecoder(resp.Body).Decode(&errResp) == nil {
			return InitSignResponse{}, fmt.Errorf("request failed: %s", errResp.Error)
		}
		return InitSignResponse{}, fmt.Errorf("request failed with status code: %d", resp.StatusCode)
	}
	var response InitSignResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return InitSignResponse{}, err
	}
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

	if resp.StatusCode != http.StatusOK {
		var errResp ResponseErrorPage
		if json.NewDecoder(resp.Body).Decode(&errResp) == nil {
			return JoinSignResponse{}, fmt.Errorf("request failed: %s", errResp.Error)
		}
		return JoinSignResponse{}, fmt.Errorf("request failed with status code: %d", resp.StatusCode)
	}
	var response JoinSignResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return JoinSignResponse{}, err
	}
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

	if resp.StatusCode != http.StatusOK {
		var errResp ResponseErrorPage
		if json.NewDecoder(resp.Body).Decode(&errResp) == nil {
			return PollSignResponse{}, fmt.Errorf("request failed: %s", errResp.Error)
		}
		return PollSignResponse{}, fmt.Errorf("request failed with status code: %d", resp.StatusCode)
	}
	var response PollSignResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return PollSignResponse{}, err
	}
	return response, nil
}

func DuctSignList(host string, req ListSignRequest) (ListSignResponse, error) {
	err := InitializeHttpClient()
	if err != nil {
		return ListSignResponse{}, err
	}
	uri, err := GetApiEndpoint(host, "ListSignCeremony")
	if err != nil {
		return ListSignResponse{}, err
	}
	body, err := json.Marshal(req)
	if err != nil {
		return ListSignResponse{}, err
	}
	resp, err := httpClient.Post(uri, "application/json", bytes.NewReader(body))
	if err != nil {
		return ListSignResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp ResponseErrorPage
		if json.NewDecoder(resp.Body).Decode(&errResp) == nil {
			return ListSignResponse{}, fmt.Errorf("request failed: %s", errResp.Error)
		}
		return ListSignResponse{}, fmt.Errorf("request failed with status code: %d", resp.StatusCode)
	}
	var response ListSignResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return ListSignResponse{}, err
	}
	return response, nil
}

// Get keygen protocol messages
func DuctKeygenGetMessages(host string, groupID string, myPartyID uint16, lastSeen int64) (KeyGenMessageResponse, error) {
	err := InitializeHttpClient()
	if err != nil {
		return KeyGenMessageResponse{}, err
	}
	uri, err := GetApiEndpoint(host, "GetKeygenMessages")
	if err != nil {
		return KeyGenMessageResponse{}, err
	}
	req := KeyGenMessageRequest{
		GroupID:   groupID,
		MyPartyID: myPartyID,
		LastSeen:  lastSeen,
	}
	body, _ := json.Marshal(req)
	resp, err := httpClient.Post(uri, "application/json", bytes.NewReader(body))
	if err != nil {
		return KeyGenMessageResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp ResponseErrorPage
		if json.NewDecoder(resp.Body).Decode(&errResp) == nil {
			return KeyGenMessageResponse{}, fmt.Errorf("request failed: %s", errResp.Error)
		}
		return KeyGenMessageResponse{}, fmt.Errorf("request failed with status code: %d", resp.StatusCode)
	}
	var response KeyGenMessageResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return KeyGenMessageResponse{}, err
	}
	return response, nil
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

	if resp.StatusCode != http.StatusOK {
		var errResp ResponseErrorPage
		if json.NewDecoder(resp.Body).Decode(&errResp) == nil {
			return KeyGenMessageResponse{}, fmt.Errorf("request failed: %s", errResp.Error)
		}
		return KeyGenMessageResponse{}, fmt.Errorf("request failed with status code: %d", resp.StatusCode)
	}
	var response KeyGenMessageResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return KeyGenMessageResponse{}, err
	}
	return response, nil
}

// Get sign protocol messages
func DuctSignGetMessages(host string, ceremonyID string, myPartyID uint16, lastSeen int64) (SignMessageResponse, error) {
	err := InitializeHttpClient()
	if err != nil {
		return SignMessageResponse{}, err
	}
	uri, err := GetApiEndpoint(host, "GetSignMessages")
	if err != nil {
		return SignMessageResponse{}, err
	}
	req := SignMessageRequest{
		CeremonyID: ceremonyID,
		MyPartyID:  myPartyID,
		LastSeen:   lastSeen,
	}
	body, _ := json.Marshal(req)
	resp, err := httpClient.Post(uri, "application/json", bytes.NewReader(body))
	if err != nil {
		return SignMessageResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp ResponseErrorPage
		if json.NewDecoder(resp.Body).Decode(&errResp) == nil {
			return SignMessageResponse{}, fmt.Errorf("request failed: %s", errResp.Error)
		}
		return SignMessageResponse{}, fmt.Errorf("request failed with status code: %d", resp.StatusCode)
	}
	var response SignMessageResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return SignMessageResponse{}, err
	}
	return response, nil
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

	if resp.StatusCode != http.StatusOK {
		var errResp ResponseErrorPage
		if json.NewDecoder(resp.Body).Decode(&errResp) == nil {
			return SignMessageResponse{}, fmt.Errorf("request failed: %s", errResp.Error)
		}
		return SignMessageResponse{}, fmt.Errorf("request failed with status code: %d", resp.StatusCode)
	}
	var response SignMessageResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return SignMessageResponse{}, err
	}
	return response, nil
}

func DuctKeygenFinalize(host string, req KeygenFinalRequest) error {
	err := InitializeHttpClient()
	if err != nil {
		return err
	}
	uri, err := GetApiEndpoint(host, "FinalizeKeygenMessage")
	if err != nil {
		return err
	}
	body, err := json.Marshal(req)
	if err != nil {
		return err
	}
	resp, err := httpClient.Post(uri, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp ResponseErrorPage
		if json.NewDecoder(resp.Body).Decode(&errResp) == nil {
			return fmt.Errorf("request failed: %s", errResp.Error)
		}
		return fmt.Errorf("request failed with status code: %d", resp.StatusCode)
	}
	return nil
}

func DuctSignFinalize(host string, req SignFinalRequest) error {
	err := InitializeHttpClient()
	if err != nil {
		return err
	}
	uri, err := GetApiEndpoint(host, "FinalizeSignMessage")
	if err != nil {
		return err
	}
	body, err := json.Marshal(req)
	if err != nil {
		return err
	}
	resp, err := httpClient.Post(uri, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp ResponseErrorPage
		if json.NewDecoder(resp.Body).Decode(&errResp) == nil {
			return fmt.Errorf("request failed: %s", errResp.Error)
		}
		return fmt.Errorf("request failed with status code: %d", resp.StatusCode)
	}
	return nil
}

func DuctGetSignature(host string, req GetSignRequest) (GetSignResponse, error) {
	err := InitializeHttpClient()
	if err != nil {
		return GetSignResponse{}, err
	}
	uri, err := GetApiEndpoint(host, "GetSignature")
	if err != nil {
		return GetSignResponse{}, err
	}
	body, _ := json.Marshal(req)
	resp, err := httpClient.Post(uri, "application/json", bytes.NewReader(body))
	if err != nil {
		return GetSignResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp ResponseErrorPage
		if json.NewDecoder(resp.Body).Decode(&errResp) == nil {
			return GetSignResponse{}, fmt.Errorf("request failed: %s", errResp.Error)
		}
		return GetSignResponse{}, fmt.Errorf("request failed with status code: %d", resp.StatusCode)
	}

	var response GetSignResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return GetSignResponse{}, err
	}
	return response, nil
}

func DuctTerminateSignCeremony(host string, req TerminateRequest) error {
	err := InitializeHttpClient()
	if err != nil {
		return err
	}
	uri, err := GetApiEndpoint(host, "TerminateSignCeremony")
	if err != nil {
		return err
	}
	body, err := json.Marshal(req)
	if err != nil {
		return err
	}
	resp, err := httpClient.Post(uri, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp ResponseErrorPage
		if json.NewDecoder(resp.Body).Decode(&errResp) == nil {
			return fmt.Errorf("request failed: %s", errResp.Error)
		}
		return fmt.Errorf("request failed with status code: %d", resp.StatusCode)
	}

	var response VapidResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return err
	}
	if response.Status != "OK" {
		return fmt.Errorf("termination failed: %s", response.Status)
	}
	return nil
}
