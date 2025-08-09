package internal

import (
	"encoding/hex"
	"fmt"
	"os"
	"time"

	"github.com/taurusgroup/frost-ed25519/pkg/frost"
	"github.com/taurusgroup/frost-ed25519/pkg/frost/party"
	"github.com/taurusgroup/frost-ed25519/pkg/messages"
	"github.com/taurusgroup/frost-ed25519/pkg/state"
)

type InitKeyGenRequest struct {
	Participants int `json:"n"`
	Threshold    int `json:"t"`
}
type InitKeyGenResponse struct {
	GroupID string `json:"group-id"`
}

// Initialize a keygen ceremony with the coordinator
func InitKeyGenCeremony(host string, participants int, threshold int) {
	req := InitKeyGenRequest{
		Participants: participants,
		Threshold:    threshold,
	}
	res, err := DuctInitKeyGenCeremony(host, req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err.Error())
		os.Exit(1)
	}
	fmt.Printf("Distributed key generation ceremony created! Group ID:\n%s\n", res.GroupID)
	os.Exit(0)
}

type PollKeyGenRequest struct {
	GroupID string  `json:"group-id"`
	PartyID *uint16 `json:"party-id,omitempty"`
}
type PollKeyGenResponse struct {
	GroupID      string   `json:"group-id"`
	MyPartyID    uint16   `json:"party-id"`
	OtherParties []uint16 `json:"parties"`
	Threshold    uint16   `json:"t"`
	PartySize    uint16   `json:"n"`
}

type PollSignRequest struct {
	GroupID string `json:"group-id"`
	PartyID uint16 `json:"party-id"`
}
type PollSignResponse struct {
	GroupID      string   `json:"group-id"`
	MyPartyID    uint16   `json:"party-id"`
	OtherParties []uint16 `json:"parties"`
}

type JoinKeyGenRequest struct {
	GroupID string `json:"group-id"`
}
type JoinKeyGenResponse struct {
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

var timeout time.Duration = time.Hour

var lastMessageIdSeen int64
var messagesIn chan *messages.Message

func ProcessKeygenMessages(msgsIn chan *messages.Message, s *state.State, host, groupID string) {
	for {
		select {
		case msg := <-msgsIn:
			// The State performs some verification to check that the message is relevant for this protocol
			if err := s.HandleMessage(msg); err != nil {
				// An error here may not be too bad, it is not necessary to abort.
				fmt.Println("failed to handle message", err)
				continue
			}

			// We ask the State for the next round of messages, and must handle them here.
			// If an abort has occurred, then no messages are returned.
			for _, msgOut := range s.ProcessAll() {
				// Transport layer
				msgBytes, err := msgOut.MarshalBinary()
				if err != nil {
					fmt.Println("failed to serialize", err)
					continue
				}
				request := KeyGenMessageRequest{
					GroupID:  groupID,
					Message:  hex.EncodeToString(msgBytes),
					LastSeen: lastMessageIdSeen,
				}
				response, err := DuctKeygenProtocolMessage(host, request)
				if err != nil {
					fmt.Println("failed to parse response", err)
					continue
				}

				// Did we get new messages to process?
				for _, m := range response.Messages {
					raw, err := hex.DecodeString(m)
					if err != nil {
						fmt.Println("failed to parse message", err)
						continue
					}
					newMsg := messages.Message{}
					newMsg.UnmarshalBinary(raw)
					// Append to messagesIn
					messagesIn <- &newMsg
				}
				lastMessageIdSeen = response.LatestMessageID
			}

		case <-s.Done():
			// s.Done() closes either when an abort has been called, or when the output has successfully been computed.
			// If an error did occur, we can handle it here
			err := s.WaitForError()
			if err != nil {
				fmt.Println("protocol aborted: ", err)
			}
			// In the main thread, it is safe to use the Output.
			return
		}
	}
}

func ProcessSignMessages(msgsIn chan *messages.Message, s *state.State, host, ceremonyID string) {
	for {
		select {
		case msg := <-msgsIn:
			// The State performs some verification to check that the message is relevant for this protocol
			if err := s.HandleMessage(msg); err != nil {
				// An error here may not be too bad, it is not necessary to abort.
				fmt.Println("failed to handle message", err)
				continue
			}

			// We ask the State for the next round of messages, and must handle them here.
			// If an abort has occurred, then no messages are returned.
			for _, msgOut := range s.ProcessAll() {
				// Transport layer
				msgBytes, err := msgOut.MarshalBinary()
				if err != nil {
					fmt.Println("failed to serialize", err)
					continue
				}
				request := SignMessageRequest{
					CeremonyID: ceremonyID,
					Message:    hex.EncodeToString(msgBytes),
					LastSeen:   lastMessageIdSeen,
				}
				response, err := DuctSignProtocolMessage(host, request)
				if err != nil {
					fmt.Println("failed to parse response", err)
					continue
				}

				// Did we get new messages to process?
				for _, m := range response.Messages {
					raw, err := hex.DecodeString(m)
					if err != nil {
						fmt.Println("failed to parse message", err)
						continue
					}
					newMsg := messages.Message{}
					newMsg.UnmarshalBinary(raw)
					// Append to messagesIn
					messagesIn <- &newMsg
				}
				lastMessageIdSeen = response.LatestMessageID
			}

		case <-s.Done():
			// s.Done() closes either when an abort has been called, or when the output has successfully been computed.
			// If an error did occur, we can handle it here
			err := s.WaitForError()
			if err != nil {
				fmt.Println("protocol aborted: ", err)
			}
			// In the main thread, it is safe to use the Output.
			return
		}
	}
}

// Join a keygen ceremony
func JoinKeyGenCeremony(host, groupID, recipient string) {
	// First, poll the server to get our party ID
	pollRequest := PollKeyGenRequest{
		GroupID: groupID,
		PartyID: nil,
	}
	pollResponse, err := DuctPollKeyGenCeremony(host, pollRequest)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err.Error())
		os.Exit(1)
	}

	// Load the properties from this threshold
	myPartyID := party.ID(pollResponse.MyPartyID)
	// partySize := party.Size(pollResponse.PartySize)
	threshold := party.Size(pollResponse.Threshold)
	pollRequest.PartyID = &pollResponse.MyPartyID

	// Now let's begin polling the server until enough parties join
	for {
		time.Sleep(time.Second)
		pollResponse, err = DuctPollKeyGenCeremony(host, pollRequest)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s", err.Error())
			os.Exit(1)
		}
		found := uint16(len(pollResponse.OtherParties))
		if found+1 == pollResponse.PartySize {
			// We can stop polling
			break
		}
	}

	// Great, let's process the party members now that we're full
	partyMembers := []party.ID{myPartyID}
	for _, p := range pollResponse.OtherParties {
		partyMembers = append(partyMembers, party.ID(p))
	}
	set := party.NewIDSlice(partyMembers)
	state, output, err := frost.NewKeygenState(myPartyID, set, threshold, timeout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err.Error())
		os.Exit(1)
	}

	// Use a goroutine for processing messages (which can append more messages)
	lastMessageIdSeen = 0
	go ProcessKeygenMessages(messagesIn, state, host, groupID)

	err = state.WaitForError()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err.Error())
		os.Exit(1)
	}

	// If we've gotten here without an error, a group key has been established!
	public := output.Public
	groupKey := hex.EncodeToString(public.GroupKey.ToEd25519())
	plaintextShare, err := output.SecretKey.MarshalBinary()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err.Error())
		os.Exit(1)
	}
	secretShare, err := EncryptShare(recipient, plaintextShare)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err.Error())
		os.Exit(1)
	}
	config, err := LoadUserConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err.Error())
		os.Exit(1)
	}
	err = config.AddShare(host, groupID, groupKey, secretShare)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err.Error())
		os.Exit(1)
	}
	fmt.Printf("Group public key:\n%s\n", groupKey)
	// OK
	os.Exit(0)
}

// List local key shares and groups
func ListKeyGen() {
	config, err := LoadUserConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err.Error())
		os.Exit(1)
	}
	if len(config.Shares) < 1 {
		fmt.Printf("No local key shares/groups found")
		os.Exit(0)
	}
	// List shares
	fmt.Printf("Group ID\tPublic Key\n")
	for _, share := range config.Shares {
		fmt.Printf("%s\t%s\n", share.GroupID, share.PublicKey)
	}
}

func InitSignCeremony(groupID, host string, message []byte, openssh bool) {

}

func JoinSignCeremony(ceremonyID, host, identityFile string, message []byte, autoConfirm bool) {

}

func ListSign(groupID string) {

}

func TerminateSignCeremony(ceremonyID string) {

}
