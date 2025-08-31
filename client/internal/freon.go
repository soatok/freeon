package internal

import (
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"hash"
	"os"
	"time"

	"github.com/bytemare/dkg"
	"github.com/bytemare/ecc"
	"github.com/bytemare/frost"
	"github.com/bytemare/secret-sharing/keys"
)

// The default timeout for the FROST protocol.
// 1 hour is eventually to allow complex key ceremonies involving airgapped machines.
var timeout time.Duration = time.Hour

// The ID of the last message seen. Sent with HTTP requests to fetch more messages.
var lastMessageIdSeen int64

// Used for determining which party should report the final result to the ceremony
var ceremonyHash hash.Hash

// Two prefixes for initializing the ceremony Hash state
var ceremonyKeyGen = []byte("FREON KeyGen Ceremony v1")
var ceremonySign = []byte("FREON Sign Ceremony v1")

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

func joinCeremonyAndPoll(host, groupID string) (uint16, uint16, uint16, []uint16, error) {
	pollRequest := PollKeyGenRequest{
		GroupID: groupID,
		PartyID: nil,
	}
	pollResponse, err := DuctPollKeyGenCeremony(host, pollRequest)
	if err != nil {
		return 0, 0, 0, nil, err
	}

	joinRequest := JoinKeyGenRequest{
		GroupID: groupID,
	}
	joinResponse, err := DuctJoinKeyGenCeremony(host, joinRequest)
	if err != nil {
		return 0, 0, 0, nil, err
	}
	ceremonyHash = sha512.New384()
	ceremonyHash.Write(ceremonyKeyGen)

	myPartyID := joinResponse.MyPartyID
	threshold := pollResponse.Threshold
	partySize := pollResponse.PartySize
	pollRequest.PartyID = &myPartyID

	for {
		pollResponse, err = DuctPollKeyGenCeremony(host, pollRequest)
		if err != nil {
			return 0, 0, 0, nil, err
		}
		found := uint16(len(pollResponse.OtherParties))
		if found+1 == partySize {
			break
		}
		time.Sleep(time.Second)
	}

	partyMembers := []uint16{myPartyID}
	partyMembers = append(partyMembers, pollResponse.OtherParties...)
	return myPartyID, threshold, partySize, partyMembers, nil
}

func performDKGRound1(host, groupID string, myPartyID, threshold, partySize uint16, partyMembers []uint16) (*dkg.Participant, []*dkg.Round1Data, error) {
	participant, err := dkg.Edwards25519Sha512.NewParticipant(myPartyID, threshold, partySize)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to start dkg: %w", err)
	}

	r1Message := participant.Start()
	r1Bytes := r1Message.Encode()
	_, err = DuctKeygenProtocolMessage(host, KeyGenMessageRequest{
		GroupID:   groupID,
		Message:   hex.EncodeToString(r1Bytes),
		MyPartyID: myPartyID,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to send r1 message: %w", err)
	}

	r1Messages := make(map[uint16]*dkg.Round1Data)
	r1Messages[myPartyID] = r1Message
	for len(r1Messages) < len(partyMembers) {
		resp, err := DuctKeygenProtocolMessage(host, KeyGenMessageRequest{
			GroupID:   groupID,
			MyPartyID: myPartyID,
			LastSeen:  lastMessageIdSeen,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("failed to poll for r1 messages: %w", err)
		}
		for _, msgStr := range resp.Messages {
			msgBytes, err := hex.DecodeString(msgStr)
			if err != nil {
				continue
			}
			ceremonyHash.Write(msgBytes)
			msg := &dkg.Round1Data{}
			if err := msg.Decode(msgBytes); err == nil {
				if _, ok := r1Messages[msg.SenderIdentifier]; !ok {
					r1Messages[msg.SenderIdentifier] = msg
				}
			}
		}
		lastMessageIdSeen = resp.LatestMessageID
		time.Sleep(time.Second)
	}
	var r1Data []*dkg.Round1Data
	for _, m := range r1Messages {
		r1Data = append(r1Data, m)
	}
	return participant, r1Data, nil
}

func performDKGRound2(host, groupID string, myPartyID, partySize uint16, participant *dkg.Participant, r1Data []*dkg.Round1Data) ([]*dkg.Round2Data, error) {
	r2Messages, err := participant.Continue(r1Data)
	if err != nil {
		return nil, fmt.Errorf("failed to continue dkg: %w", err)
	}
	for _, msg := range r2Messages {
		msgBytes := msg.Encode()
		ceremonyHash.Write(msgBytes)
		_, err = DuctKeygenProtocolMessage(host, KeyGenMessageRequest{
			GroupID:   groupID,
			Message:   hex.EncodeToString(msgBytes),
			MyPartyID: myPartyID,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to send r2 message: %w", err)
		}
	}

	myR2Messages := make(map[uint16]*dkg.Round2Data)
	for len(myR2Messages) < int(partySize)-1 {
		resp, err := DuctKeygenProtocolMessage(host, KeyGenMessageRequest{
			GroupID:   groupID,
			MyPartyID: myPartyID,
			LastSeen:  lastMessageIdSeen,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to poll for r2 messages: %w", err)
		}
		for _, msgStr := range resp.Messages {
			msgBytes, err := hex.DecodeString(msgStr)
			if err != nil {
				continue
			}
			ceremonyHash.Write(msgBytes)
			msg := &dkg.Round2Data{}
			if err := msg.Decode(msgBytes); err == nil {
				if msg.RecipientIdentifier == myPartyID {
					if _, ok := myR2Messages[msg.SenderIdentifier]; !ok {
						myR2Messages[msg.SenderIdentifier] = msg
					}
				}
			}
		}
		lastMessageIdSeen = resp.LatestMessageID
		time.Sleep(time.Second)
	}
	var r2Data []*dkg.Round2Data
	for _, m := range myR2Messages {
		r2Data = append(r2Data, m)
	}
	return r2Data, nil
}

func finalizeAndStoreKeys(host, groupID, recipient string, myPartyID uint16, partyMembers []uint16, participant *dkg.Participant, r1Data []*dkg.Round1Data, r2Data []*dkg.Round2Data) error {
	keyShare, err := participant.Finalize(r1Data, r2Data)
	if err != nil {
		return fmt.Errorf("failed to finalize dkg: %w", err)
	}

	var allCommitments [][]*ecc.Element
	for _, d := range r1Data {
		allCommitments = append(allCommitments, d.Commitment)
	}

	publicShares := make(map[string]string)
	for _, pID := range partyMembers {
		pubKey, err := dkg.ComputeParticipantPublicKey(dkg.Edwards25519Sha512, pID, allCommitments)
		if err != nil {
			return fmt.Errorf("failed to compute public key for party %d: %w", pID, err)
		}
		pubKeyBytes := pubKey.Encode()
		publicShares[Uint16ToHexBE(pID)] = hex.EncodeToString(pubKeyBytes)
	}

	groupKeyBytes := keyShare.VerificationKey.Encode()
	groupKeyHex := hex.EncodeToString(groupKeyBytes)

	secretShareBytes := keyShare.Secret.Encode()
	encryptedShare, err := EncryptShare(recipient, secretShareBytes)
	if err != nil {
		return fmt.Errorf("failed to encrypt share: %w", err)
	}

	config, err := LoadUserConfig()
	if err != nil {
		return err
	}

	err = config.AddShare(host, groupID, groupKeyHex, encryptedShare, publicShares, myPartyID)
	if err != nil {
		return err
	}
	ch := ceremonyHash.Sum(nil)
	if AmIElected(ch, myPartyID, partyMembers) {
		report := KeygenFinalRequest{
			GroupID:   groupID,
			MyPartyID: myPartyID,
			PublicKey: groupKeyHex,
		}
		err := DuctKeygenFinalize(host, report)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		}
	}
	fmt.Printf("Group public key:\n%s\n", groupKeyHex)
	os.Exit(0)
	return nil
}

// Join a keygen ceremony
func JoinKeyGenCeremony(host, groupID, recipient string) {
	// This function is getting long. Let's break it down into smaller pieces.
	// 1. Join the ceremony and get participant info.
	// 2. Perform DKG Round 1.
	// 3. Perform DKG Round 2.
	// 4. Finalize and store keys.

	// 1. Join the ceremony and get participant info.
	myPartyID, threshold, partySize, partyMembers, err := joinCeremonyAndPoll(host, groupID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to join ceremony: %s\n", err.Error())
		os.Exit(1)
	}

	// 2. Perform DKG Round 1.
	participant, r1Data, err := performDKGRound1(host, groupID, myPartyID, threshold, partySize, partyMembers)
	if err != nil {
		fmt.Fprintf(os.Stderr, "DKG round 1 failed: %s\n", err.Error())
		os.Exit(1)
	}

	// 3. Perform DKG Round 2.
	r2Data, err := performDKGRound2(host, groupID, myPartyID, partySize, participant, r1Data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "DKG round 2 failed: %s\n", err.Error())
		os.Exit(1)
	}

	// 4. Finalize and store keys.
	err = finalizeAndStoreKeys(host, groupID, recipient, myPartyID, partyMembers, participant, r1Data, r2Data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to finalize and store keys: %s\n", err.Error())
		os.Exit(1)
	}
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
	// Let's pull in the data from the local config:
	config, err := LoadUserConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}

	// Before we can do anything, we need to get the GroupID from the ceremony.
	// The only way to do that is to poll.
	pollRequest := PollSignRequest{
		CeremonyID: ceremonyID,
		PartyID:    nil, // We don't know our party ID yet.
	}
	pollResponse, err := DuctPollSignCeremony(host, pollRequest)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}
	groupID := pollResponse.GroupID
	threshold := pollResponse.Threshold

	var encryptedShare string
	var publicSharesHex map[string]string
	var publicKeyHex string
	var myPartyID uint16
	for _, s := range config.Shares {
		if s.GroupID == groupID {
			encryptedShare = s.EncryptedShare
			publicSharesHex = s.PublicShares
			publicKeyHex = s.PublicKey
			myPartyID = s.MyPartyID
			break
		}
	}
	if encryptedShare == "" {
		fmt.Fprintf(os.Stderr, "could not find encrypted share for group %s\n", groupID)
		os.Exit(1)
	}
	if myPartyID == 0 {
		fmt.Fprintf(os.Stderr, "could not find party ID for group %s\n", groupID)
		os.Exit(1)
	}

	// Next, we need to formally join the party
	hash := HashMessageForSanity(message, groupID)
	joinRequest := JoinSignRequest{
		CeremonyID:  ceremonyID,
		MessageHash: hash,
		MyPartyID:   myPartyID,
	}

	// Enlist ourselves before we begin polling
	res, err := DuctJoinSignCeremony(host, joinRequest)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}
	if !res.Status {
		fmt.Fprintf(os.Stderr, "An unexpected error has occurred.\n")
		os.Exit(1)
	}
	openssh := res.OpenSSH
	opensshNamespace := res.Namespace

	// Now let's begin polling the server until enough parties join
	for {
		// We need to use our actual party ID for polling now
		pollRequest.PartyID = &myPartyID
		pollResponse, err = DuctPollSignCeremony(host, pollRequest)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err.Error())
			os.Exit(1)
		}
		others := uint16(len(pollResponse.OtherParties))
		if others+1 >= threshold {
			break
		}
		time.Sleep(time.Second)
	}

	// Let's decrypt the local share with age
	secretBytes, err := DecryptShareFor(encryptedShare, identityFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}
	secretKey := dkg.Edwards25519Sha512.Group().NewScalar()
	if err := secretKey.Decode(secretBytes); err != nil {
		fmt.Fprintf(os.Stderr, "failed to decode secret key: %s\n", err.Error())
		os.Exit(1)
	}

	// Let's decode the public key and public shares
	groupKeyBytes, err := hex.DecodeString(publicKeyHex)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to decode group key: %s\n", err.Error())
		os.Exit(1)
	}
	groupKey := dkg.Edwards25519Sha512.Group().NewElement()
	if err := groupKey.Decode(groupKeyBytes); err != nil {
		fmt.Fprintf(os.Stderr, "failed to decode group key: %s\n", err.Error())
		os.Exit(1)
	}

	// Great, let's process the party members now that we're full
	partyMembers := []uint16{myPartyID}
	partyMembers = append(partyMembers, pollResponse.OtherParties...)

	// Create a map of party members for quick lookup
	partyMemberSet := make(map[uint16]struct{})
	for _, p := range partyMembers {
		partyMemberSet[p] = struct{}{}
	}

	// Let's make sure we have all parties' public shares setup locally
	publicShares := make([]*keys.PublicKeyShare, 0, len(publicSharesHex))
	for k, v := range publicSharesHex {
		p16, err := HexBEToUint16(k)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err.Error())
			os.Exit(1)
		}
		if _, ok := partyMemberSet[p16]; !ok {
			continue
		}
		rawEl, err := hex.DecodeString(v)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err.Error())
			os.Exit(1)
		}
		el := dkg.Edwards25519Sha512.Group().NewElement()
		if err := el.Decode(rawEl); err != nil {
			fmt.Fprintf(os.Stderr, "failed to decode public share for party %d: %s\n", p16, err.Error())
			os.Exit(1)
		}
		ps := &keys.PublicKeyShare{
			ID:        p16,
			PublicKey: el,
			Group:     dkg.Edwards25519Sha512.Group(),
		}
		publicShares = append(publicShares, ps)
	}

	conf := &frost.Configuration{
		Ciphersuite:           frost.Ed25519,
		Threshold:             threshold,
		MaxSigners:            uint16(len(partyMembers)),
		VerificationKey:       groupKey,
		SignerPublicKeyShares: publicShares,
	}
	if err := conf.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialize frost config: %s\n", err.Error())
		os.Exit(1)
	}

	myPublicKey := dkg.Edwards25519Sha512.Group().Base().Multiply(secretKey)
	myPkShare := &keys.PublicKeyShare{
		ID:        myPartyID,
		PublicKey: myPublicKey,
		Group:     dkg.Edwards25519Sha512.Group(),
	}
	myKeyShare := &keys.KeyShare{
		Secret:          secretKey,
		PublicKeyShare:  *myPkShare,
		VerificationKey: groupKey,
	}

	signer, err := conf.Signer(myKeyShare)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create signer: %s\n", err.Error())
		os.Exit(1)
	}

	// Round 1: Commitment
	ceremonyHash = sha512.New384()
	ceremonyHash.Write(ceremonySign)
	commitment := signer.Commit()
	commitBytes := commitment.Encode()
	_, err = DuctSignProtocolMessage(host, SignMessageRequest{
		CeremonyID: ceremonyID,
		Message:    hex.EncodeToString(commitBytes),
		MyPartyID:  myPartyID,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to send commitment: %s\n", err.Error())
		os.Exit(1)
	}

	// Poll for commitments from other participants
	commitments := make(map[uint16]*frost.Commitment)
	commitments[myPartyID] = commitment
	for len(commitments) < len(partyMembers) {
		resp, err := DuctSignProtocolMessage(host, SignMessageRequest{
			CeremonyID: ceremonyID,
			MyPartyID:  myPartyID,
			LastSeen:   lastMessageIdSeen,
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to poll for commitments: %s\n", err.Error())
			os.Exit(1)
		}
		for _, msgStr := range resp.Messages {
			msgBytes, err := hex.DecodeString(msgStr)
			if err != nil {
				continue // Ignore invalid messages
			}
			ceremonyHash.Write(msgBytes)
			c := &frost.Commitment{}
			if err := c.Decode(msgBytes); err == nil {
				if _, ok := commitments[c.SignerID]; !ok {
					commitments[c.SignerID] = c
				}
			}
		}
		lastMessageIdSeen = resp.LatestMessageID
		time.Sleep(time.Second)
	}

	var commitmentList []*frost.Commitment
	for _, c := range commitments {
		commitmentList = append(commitmentList, c)
	}

	// Round 2: Sign
	sigShare, err := signer.Sign(message, commitmentList)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to sign: %s\n", err.Error())
		os.Exit(1)
	}
	shareBytes := sigShare.Encode()
	_, err = DuctSignProtocolMessage(host, SignMessageRequest{
		CeremonyID: ceremonyID,
		Message:    hex.EncodeToString(shareBytes),
		MyPartyID:  myPartyID,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to send signature share: %s\n", err.Error())
		os.Exit(1)
	}

	// Poll for signature shares from other participants
	sigShares := make(map[uint16]*frost.SignatureShare)
	sigShares[myPartyID] = sigShare
	for len(sigShares) < len(partyMembers) {
		resp, err := DuctSignProtocolMessage(host, SignMessageRequest{
			CeremonyID: ceremonyID,
			MyPartyID:  myPartyID,
			LastSeen:   lastMessageIdSeen,
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to poll for signature shares: %s\n", err.Error())
			os.Exit(1)
		}
		for _, msgStr := range resp.Messages {
			msgBytes, err := hex.DecodeString(msgStr)
			if err != nil {
				continue // Ignore invalid messages
			}
			ceremonyHash.Write(msgBytes)
			s := &frost.SignatureShare{}
			if err := s.Decode(msgBytes); err == nil {
				if _, ok := sigShares[s.SignerIdentifier]; !ok {
					sigShares[s.SignerIdentifier] = s
				}
			}
		}
		lastMessageIdSeen = resp.LatestMessageID
		time.Sleep(time.Second)
	}
	var signatureShares []*frost.SignatureShare
	for _, s := range sigShares {
		signatureShares = append(signatureShares, s)
	}

	// Aggregate signatures
	finalSignature, err := conf.AggregateSignatures(message, signatureShares, commitmentList, true)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to aggregate signatures: %s\n", err.Error())
		os.Exit(1)
	}

	finalSignatureBytes := append(finalSignature.R.Encode(), finalSignature.Z.Encode()...)
	var groupSig string
	if openssh {
		groupSig = OpenSSHEncode(groupKeyBytes, finalSignatureBytes, opensshNamespace)
	} else {
		groupSig = hex.EncodeToString(finalSignatureBytes)
	}
	ch := ceremonyHash.Sum(nil)
	if AmIElected(ch, myPartyID, partyMembers) {
		report := SignFinalRequest{
			CeremonyID: ceremonyID,
			MyPartyID:  myPartyID,
			Signature:  groupSig,
		}
		err := DuctSignFinalize(host, report)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err.Error())
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

// Tell the coordinator to pull the plug on a signing ceremony
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
