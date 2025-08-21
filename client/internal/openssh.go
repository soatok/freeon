package internal

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
)

func putString(buf *bytes.Buffer, s []byte) {
	binary.Write(buf, binary.BigEndian, uint32(len(s)))
	buf.Write(s)
}

func makeOpenSSHSignature(pubKey []byte, rawSig []byte, namespace string) string {
	if len(pubKey) != 32 {
		panic("Ed25519 public key must be 32 bytes")
	}
	if len(rawSig) != 64 {
		panic("Ed25519 signature must be 64 bytes")
	}

	var buf bytes.Buffer
	// "SSHSIG" v1
	putString(&buf, []byte("SSHSIG"))
	binary.Write(&buf, binary.BigEndian, uint32(1))

	// public key blob
	var pkBlob bytes.Buffer
	putString(&pkBlob, []byte("ssh-ed25519"))
	putString(&pkBlob, pubKey)
	putString(&buf, pkBlob.Bytes())

	// namespace
	putString(&buf, []byte(namespace))

	// reserved (empty string)
	putString(&buf, []byte{})

	// signature algorithm
	putString(&buf, []byte("ssh-ed25519"))

	// signature blob
	var sigBlob bytes.Buffer
	putString(&sigBlob, []byte("ssh-ed25519"))
	putString(&sigBlob, rawSig)
	putString(&buf, sigBlob.Bytes())

	// Base64 encode
	b64 := base64.StdEncoding.EncodeToString(buf.Bytes())

	// Wrap at 70 chars per line
	var wrapped bytes.Buffer
	for i := 0; i < len(b64); i += 70 {
		end := i + 70
		if end > len(b64) {
			end = len(b64)
		}
		wrapped.WriteString(b64[i:end] + "\n")
	}

	return "-----BEGIN SSH SIGNATURE-----\n" +
		wrapped.String() +
		"-----END SSH SIGNATURE-----\n"
}
