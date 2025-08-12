package internal

import (
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"os"
	"time"

	"github.com/taurusgroup/frost-ed25519/pkg/eddsa"
	"github.com/taurusgroup/frost-ed25519/pkg/frost"
	"github.com/taurusgroup/frost-ed25519/pkg/frost/party"
	"github.com/taurusgroup/frost-ed25519/pkg/messages"
	"github.com/taurusgroup/frost-ed25519/pkg/ristretto"
	"github.com/taurusgroup/frost-ed25519/pkg/state"
)

// The default timeout for the FROST protocol.
// 1 hour is eventually to allow complex key ceremonies involving airgapped machines.
var timeout time.Duration = time.Hour

// The ID of the last message seen. Sent with HTTP requests to fetch more messages.
var lastMessageIdSeen int64

// Used for goroutines that process FROST protocol messages
// See ProcessKeygenMessages() and ProcessSignMessages() below.
var messagesIn chan *messages.Message

// Initialize a keygen ceremony with the coordinator
func InitKeyGenCeremony(host string, participants uint16, threshold uint16) {
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

func HashMessageForSanity(data []byte) string {
	hash := sha512.Sum384(data)
	return hex.EncodeToString(hash[:])
}

// Kicking off a key-signing ceremony
func InitSignCeremony(host, groupID string, message []byte) {
	req := InitSignRequest{
		GroupID:     groupID,
		MessageHash: HashMessageForSanity(message),
	}
	res, err := DuctInitSignCeremony(host, req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err.Error())
		os.Exit(1)
	}
	fmt.Printf("Key signing ceremony created!\n%s\n", res.CeremonyID)
	os.Exit(0)
}

// Goroutine for processing the Keygen protocol messages
func ProcessKeygenMessages(msgsIn chan *messages.Message, s *state.State, host, groupID string, myPartyID uint16) {
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
					GroupID:   groupID,
					Message:   hex.EncodeToString(msgBytes),
					MyPartyID: myPartyID,
					LastSeen:  lastMessageIdSeen,
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

// Goroutine for processing the Sign protocol messages
func ProcessSignMessages(msgsIn chan *messages.Message, s *state.State, host, ceremonyID string, myPartyID uint16) {
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
					MyPartyID:  myPartyID,
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
	// First, poll the server to make sure it exists
	pollRequest := PollKeyGenRequest{
		GroupID: groupID,
		PartyID: nil,
	}
	pollResponse, err := DuctPollKeyGenCeremony(host, pollRequest)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err.Error())
		os.Exit(1)
	}

	// Next, we need to formally join the party and get your ID
	joinRequest := JoinKeyGenRequest{
		GroupID: groupID,
	}
	joinResponse, err := DuctJoinKeyGenCeremony(host, joinRequest)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err.Error())
		os.Exit(1)
	}

	// Load the properties from this threshold
	myPartyID := party.ID(joinResponse.MyPartyID)
	// partySize := party.Size(pollResponse.PartySize)
	threshold := party.Size(pollResponse.Threshold)
	pollRequest.PartyID = &joinResponse.MyPartyID

	// Now let's begin polling the server until enough parties join
	for {
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
		time.Sleep(time.Second)
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
	go ProcessKeygenMessages(messagesIn, state, host, groupID, joinResponse.MyPartyID)

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
	// Let's build the list of public shares
	publicShares := make(map[string]string)
	for index, sh := range public.Shares {
		i := uint16ToHexBE(uint16(index))
		shh := hex.EncodeToString(sh.BytesEd25519())
		publicShares[i] = shh
	}

	// Okay, finally, we add the share data to the local config
	err = config.AddShare(host, groupID, groupKey, secretShare, publicShares)
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

func JoinSignCeremony(ceremonyID, host, identityFile string, message []byte) {
	//, autoConfirm bool
	// First, poll the server to get metadata
	pollRequest := PollSignRequest{
		CeremonyID: ceremonyID,
		PartyID:    nil,
	}
	pollResponse, err := DuctPollSignCeremony(host, pollRequest)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err.Error())
		os.Exit(1)
	}
	myPartyID := party.ID(pollResponse.MyPartyID)
	groupID := pollResponse.GroupID
	threshold := pollResponse.Threshold

	// Next, we need to actually join the signing quorum

	// Next, we need to formally join the party and get your ID
	hash := HashMessageForSanity(message)
	joinRequest := JoinSignRequest{
		CeremonyID:  ceremonyID,
		MessageHash: hash,
		MyPartyID:   pollResponse.MyPartyID,
	}

	// Enlist ourselves before we begin polling
	res, err := DuctJoinSignCeremony(host, joinRequest)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err.Error())
		os.Exit(1)
	}
	if !res.Status {
		fmt.Fprintf(os.Stderr, "An unexpected error has occurred.\n")
		os.Exit(1)
	}

	// Now let's begin polling the server until enough parties join
	for {
		pollResponse, err = DuctPollSignCeremony(host, pollRequest)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s", err.Error())
			os.Exit(1)
		}
		others := uint16(len(pollResponse.OtherParties))
		if others+1 >= threshold {
			break
		}
		time.Sleep(time.Second)
	}

	config, err := LoadUserConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err.Error())
		os.Exit(1)
	}
	var encryptedShare string = ""
	var publicSharesHex map[string]string
	var publicKeyHex string
	for _, s := range config.Shares {
		if s.GroupID == groupID {
			encryptedShare = s.EncryptedShare
			publicSharesHex = s.PublicShares
			publicKeyHex = s.PublicKey
		}
	}
	if encryptedShare == "" {
		fmt.Fprintf(os.Stderr, "could not find encrypted share for group %s", groupID)
		os.Exit(1)
	}
	rawPk, err := hex.DecodeString(publicKeyHex)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not find encrypted share for group %s", groupID)
		os.Exit(1)
	}

	// Let's deserialize the public shares
	publicShares := make(map[party.ID]*ristretto.Element, len(publicSharesHex))
	for k, v := range publicSharesHex {
		p16, err := hexBEToUint16(k)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s", err.Error())
			os.Exit(1)
		}
		rawEl, err := hex.DecodeString(v)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s", err.Error())
			os.Exit(1)
		}
		pid := party.ID(p16)
		var el *ristretto.Element
		el.SetCanonicalBytes(rawEl)
		publicShares[pid] = el
	}

	// Let's decrypt with age
	secretBytes, err := DecryptShareFor(encryptedShare, identityFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err.Error())
		os.Exit(1)
	}
	var secret eddsa.SecretShare
	err = secret.UnmarshalBinary(secretBytes)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err.Error())
		os.Exit(1)
	}

	// Great, let's process the party members now that we're full
	partyMembers := []party.ID{myPartyID}
	for _, p := range pollResponse.OtherParties {
		partyMembers = append(partyMembers, party.ID(p))
	}
	set := party.NewIDSlice(partyMembers)

	var pkEl *ristretto.Element
	pkEl.SetCanonicalBytes(rawPk)
	pk := eddsa.NewPublicKeyFromPoint(pkEl)

	publicData := eddsa.Public{
		PartyIDs:  set,
		Threshold: party.Size(threshold),
		Shares:    publicShares,
		GroupKey:  pk,
	}

	state, signOutput, err := frost.NewSignState(set, &secret, &publicData, message, timeout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err.Error())
		os.Exit(1)
	}

	// Use a goroutine for processing messages (which can append more messages)
	lastMessageIdSeen = 0
	go ProcessSignMessages(messagesIn, state, host, groupID, uint16(myPartyID))

	err = state.WaitForError()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err.Error())
		os.Exit(1)
	}

	groupSig := hex.EncodeToString(signOutput.Signature.ToEd25519())
	fmt.Printf("Signature:\n%s\n", groupSig)
}

func ListSign(groupID string) {
	// TODO - soatok
}

func TerminateSignCeremony(ceremonyID string) {
	// TODO - soatok
}
