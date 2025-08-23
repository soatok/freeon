package internal

import (
	"context"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"hash"
	"os"
	"sync"
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
var lastMessageMutex sync.Mutex

// Used for goroutines that process FROST protocol messages
// See ProcessKeygenMessages() and ProcessSignMessages() below.
var messagesIn chan *messages.Message

// Used for determining which party should report the final result to the ceremony
var ceremonyHash hash.Hash

// Two prefixes for initializing the ceremony Hash state
var ceremonyKeyGen = []byte("FREON KeyGen Ceremony v1")
var ceremonySign = []byte("FREON Sign Ceremony v1")

// Initialize a keygen ceremony with the coordinator
func InitKeyGenCeremony(host string, participants, threshold uint16) {
	if threshold > participants {
		fmt.Printf("t > n: t = %d, n = %d\n", threshold, participants)
		os.Exit(1)
	}
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

// Kicking off a key-signing ceremony
func InitSignCeremony(host, groupID string, message []byte, openssh bool, namespace string) {
	req := InitSignRequest{
		GroupID:     groupID,
		MessageHash: HashMessageForSanity(message, groupID),
		OpenSSH:     openssh,
		Namespace:   namespace,
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
				lastMessageMutex.Lock()
				request := KeyGenMessageRequest{
					GroupID:   groupID,
					Message:   hex.EncodeToString(msgBytes),
					MyPartyID: myPartyID,
					LastSeen:  lastMessageIdSeen,
				}
				response, err := DuctKeygenProtocolMessage(host, request)
				if err != nil {
					fmt.Println("failed to parse response", err)
					lastMessageMutex.Unlock()
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
				lastMessageMutex.Unlock()
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
				lastMessageMutex.Lock()
				request := SignMessageRequest{
					CeremonyID: ceremonyID,
					MyPartyID:  myPartyID,
					Message:    hex.EncodeToString(msgBytes),
					LastSeen:   lastMessageIdSeen,
				}
				response, err := DuctSignProtocolMessage(host, request)
				if err != nil {
					fmt.Println("failed to parse response", err)
					lastMessageMutex.Unlock()
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
				lastMessageMutex.Unlock()
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

func PollKeyGenMessages(ctx context.Context, host, groupID string, myPartyID uint16) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			lastMessageMutex.Lock()
			seen := lastMessageIdSeen
			lastMessageMutex.Unlock()
			response, err := DuctKeygenGetMessages(host, groupID, myPartyID, seen)
			if err != nil {
				// Don't spam errors for polling
				continue
			}
			if len(response.Messages) > 0 {
				for _, m := range response.Messages {
					raw, err := hex.DecodeString(m)
					if err != nil {
						continue
					}
					newMsg := messages.Message{}
					newMsg.UnmarshalBinary(raw)
					messagesIn <- &newMsg
				}
			}
			lastMessageMutex.Lock()
			lastMessageIdSeen = response.LatestMessageID
			lastMessageMutex.Unlock()
		}
	}
}

func PollSignMessages(ctx context.Context, host, ceremonyID string, myPartyID uint16) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			lastMessageMutex.Lock()
			seen := lastMessageIdSeen
			lastMessageMutex.Unlock()
			response, err := DuctSignGetMessages(host, ceremonyID, myPartyID, seen)
			if err != nil {
				// Don't spam errors
				continue
			}
			if len(response.Messages) > 0 {
				for _, m := range response.Messages {
					raw, err := hex.DecodeString(m)
					if err != nil {
						continue
					}
					newMsg := messages.Message{}
					newMsg.UnmarshalBinary(raw)
					messagesIn <- &newMsg
				}
			}
			lastMessageMutex.Lock()
			lastMessageIdSeen = response.LatestMessageID
			lastMessageMutex.Unlock()
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
	if pollResponse.Threshold > pollResponse.PartySize {
		fmt.Fprintf(os.Stderr, "Threshold is larger than Party Size")
		os.Exit(1)
	}

	// Load the properties from this threshold
	myPartyID := party.ID(joinResponse.MyPartyID)
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
	messagesIn = make(chan *messages.Message, len(partyMembers)*2)
	lastMessageIdSeen = 0
	ceremonyHash = sha512.New384()
	ceremonyHash.Write(ceremonyKeyGen)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go PollKeyGenMessages(ctx, host, groupID, joinResponse.MyPartyID)
	go ProcessKeygenMessages(messagesIn, state, host, groupID, joinResponse.MyPartyID)

	// Kick off the ceremony
	for _, msgOut := range state.ProcessAll() {
		msgBytes, err := msgOut.MarshalBinary()
		if err != nil {
			fmt.Println("failed to serialize", err)
			continue
		}
		lastMessageMutex.Lock()
		request := KeyGenMessageRequest{
			GroupID:   groupID,
			Message:   hex.EncodeToString(msgBytes),
			MyPartyID: joinResponse.MyPartyID,
			LastSeen:  lastMessageIdSeen,
		}
		response, err := DuctKeygenProtocolMessage(host, request)
		if err != nil {
			fmt.Println("failed to parse response", err)
			lastMessageMutex.Unlock()
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
		lastMessageMutex.Unlock()
	}

	err = state.WaitForError()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err.Error())
		os.Exit(1)
	}

	// If we've gotten here without an error, a group key has been established!
	public := output.Public
	groupKeyStore, err := public.GroupKey.MarshalJSON()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err.Error())
		os.Exit(1)
	}
	groupKey := hex.EncodeToString(public.GroupKey.ToEd25519())
	plaintextShare, err := output.SecretKey.MarshalJSON()
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
	for index, sh := range output.Public.Shares {
		i := Uint16ToHexBE(uint16(index))
		shEncoded, err := sh.MarshalText()
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s", err.Error())
			os.Exit(1)
		}
		publicShares[i] = string(shEncoded)
	}

	// Okay, finally, we add the share data to the local config
	err = config.AddShare(host, groupID, uint16(myPartyID), string(groupKeyStore), secretShare, publicShares)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err.Error())
		os.Exit(1)
	}
	ch := ceremonyHash.Sum(nil)
	if AmIElected(ch, uint16(myPartyID), PartyToUint16(partyMembers)) {
		report := KeygenFinalRequest{
			GroupID:   groupID,
			MyPartyID: uint16(myPartyID),
			PublicKey: groupKey,
		}
		err := DuctKeygenFinalize(host, report)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s", err.Error())
			// This is only a reporting error, so do not error out.
		}
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

// Join a signing ceremony
func JoinSignCeremony(ceremonyID, host, identityFile string, message []byte) {
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
	groupID := pollResponse.GroupID
	threshold := pollResponse.Threshold

	// Let's pull in the data from the local config to find our party ID
	config, err := LoadUserConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err.Error())
		os.Exit(1)
	}
	var encryptedShare string = ""
	var publicSharesEncoded map[string]string
	var publicKeyEncoded string
	var myPartyIDfromConfig uint16
	for _, s := range config.Shares {
		if s.GroupID == groupID {
			encryptedShare = s.EncryptedShare
			publicSharesEncoded = s.PublicShares
			publicKeyEncoded = s.PublicKey
			myPartyIDfromConfig = s.PartyID
			break
		}
	}
	if encryptedShare == "" {
		fmt.Fprintf(os.Stderr, "could not find encrypted share for group %s", groupID)
		os.Exit(1)
	}
	myPartyID := party.ID(myPartyIDfromConfig)

	// Next, we need to formally join the party
	hash := HashMessageForSanity(message, groupID)
	joinRequest := JoinSignRequest{
		CeremonyID:  ceremonyID,
		MessageHash: hash,
		MyPartyID:   myPartyIDfromConfig,
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
	openssh := res.OpenSSH
	opensshNamespace := res.Namespace

	// Now let's begin polling the server until enough parties join
	pollRequest.PartyID = &myPartyIDfromConfig
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

	publicKey := new(eddsa.PublicKey)
	err = publicKey.UnmarshalJSON([]byte(publicKeyEncoded))
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not decode encrypted share for group %s\n%s", groupID, err.Error())
		os.Exit(1)
	}

	// Let's deserialize the public shares
	publicShares := make(map[party.ID]*ristretto.Element, len(publicSharesEncoded))
	for k, v := range publicSharesEncoded {
		p16, err := HexBEToUint16(k)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s", err.Error())
			os.Exit(1)
		}
		pid := party.ID(p16)
		el := new(ristretto.Element)
		el.UnmarshalText([]byte(v))
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s", err.Error())
			os.Exit(1)
		}
		publicShares[pid] = el
	}

	// Let's decrypt the local share with age
	secretBytes, err := DecryptShareFor(encryptedShare, identityFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err.Error())
		os.Exit(1)
	}
	var secret eddsa.SecretShare
	err = secret.UnmarshalJSON(secretBytes)
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

	publicData := eddsa.Public{
		PartyIDs:  set,
		Threshold: party.Size(threshold),
		Shares:    publicShares,
		GroupKey:  publicKey,
	}

	// Initilize the Sign ceremony state
	state, signOutput, err := frost.NewSignState(set, &secret, &publicData, message, timeout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err.Error())
		os.Exit(1)
	}

	// Use a goroutine for processing messages (which can append more messages)
	messagesIn = make(chan *messages.Message, len(partyMembers)*2)
	lastMessageIdSeen = 0
	ceremonyHash = sha512.New384()
	ceremonyHash.Write(ceremonySign)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go PollSignMessages(ctx, host, ceremonyID, uint16(myPartyID))
	go ProcessSignMessages(messagesIn, state, host, ceremonyID, uint16(myPartyID))

	// Kick off the ceremony
	for _, msgOut := range state.ProcessAll() {
		msgBytes, err := msgOut.MarshalBinary()
		if err != nil {
			fmt.Println("failed to serialize", err)
			continue
		}
		lastMessageMutex.Lock()
		seen := lastMessageIdSeen
		lastMessageMutex.Unlock()
		request := SignMessageRequest{
			CeremonyID: ceremonyID,
			Message:    hex.EncodeToString(msgBytes),
			MyPartyID:  uint16(myPartyID),
			LastSeen:   seen,
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
		lastMessageMutex.Lock()
		lastMessageIdSeen = response.LatestMessageID
		lastMessageMutex.Unlock()
	}

	err = state.WaitForError()
	if err != nil {
		fmt.Fprintf(os.Stderr, "WaitForError\n")
		fmt.Fprintf(os.Stderr, "%s", err.Error())
		os.Exit(1)
	}

	// Final signature aggregation
	var groupSig string
	if openssh {
		rawPk := publicKey.ToEd25519()
		groupSig = OpenSSHEncode(rawPk, signOutput.Signature.ToEd25519(), opensshNamespace)
	} else {
		groupSig = hex.EncodeToString(signOutput.Signature.ToEd25519())
	}
	ch := ceremonyHash.Sum(nil)

	if AmIElected(ch, uint16(myPartyID), PartyToUint16(partyMembers)) {
		report := SignFinalRequest{
			CeremonyID: ceremonyID,
			MyPartyID:  uint16(myPartyID),
			Signature:  groupSig,
		}
		err := DuctSignFinalize(host, report)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s", err.Error())
			// We do not abort here, since the only error was with reporting upstream
		}
	}
	fmt.Printf("Signature:\n%s\n", groupSig)
}

// List the most recent signing ceremonies
func ListSign(host, groupID string, limit, offset int64) {
	req := ListSignRequest{
		GroupID: groupID,
		Limit:   limit,
		Offset:  offset,
	}
	res, err := DuctSignList(host, req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err.Error())
		os.Exit(1)
	}
	count := len(res.Ceremonies)
	if count < 1 {
		fmt.Printf("No ceremonies found.")
		os.Exit(1)
	}

	// Loop over the list and print the output to the console.
	fmt.Printf("Listing the most recent %d ceremonies:\n\n", count)
	fmt.Printf("\tCeremony ID\tHash\tFormat\tOpen?\n")
	fmt.Printf("\t-------------------------------------------------------------------------------\n")
	for _, ceremony := range res.Ceremonies {
		var format string
		var status string

		if ceremony.OpenSSH {
			format = "OpeenSSH"
		} else {
			format = "Raw"
		}
		if ceremony.Active {
			status = "Open"
		} else {
			status = " -- "
		}
		fmt.Printf("\t%s\t%s\t%s\t%s\n", ceremony.Uid, ceremony.Hash, format, status)
	}
	fmt.Printf("\n")
	os.Exit(0)
}

// Fetch a signature from the coordinator for a given ceremony
func GetSignSignature(ceremonyID, host string) {
	req := GetSignRequest{
		CeremonyID: ceremonyID,
	}
	res, err := DuctGetSignature(host, req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err.Error())
		os.Exit(1)
	}
	fmt.Printf("Signature:\n%s\n", res.Signature)
	os.Exit(0)
}

func TerminateSignCeremony(host, ceremonyID string) {
	req := TerminateRequest{
		CeremonyID: ceremonyID,
	}
	err := DuctTerminateSignCeremony(host, req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		os.Exit(1)
	}
	fmt.Println("Ceremony terminated.")
	os.Exit(0)
}
