package internal_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"filippo.io/age"
	"github.com/soatok/freon/client/internal"
)

func TestEncryptDecryptShare(t *testing.T) {
	identity, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatal(err)
	}

	recipient := identity.Recipient().String()
	share := []byte("test share")

	encrypted, err := internal.EncryptShare(recipient, share)
	if err != nil {
		t.Fatal(err)
	}

	decrypted, err := internal.DecryptShare(encrypted, identity)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(share, decrypted) {
		t.Fatal("decrypted share does not match original share")
	}
}

func TestParseAgeIdentityFile(t *testing.T) {
	identity, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatal(err)
	}

	tempDir := t.TempDir()
	identityFile := filepath.Join(tempDir, "key.txt")

	f, err := os.Create(identityFile)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	_, err = f.WriteString(identity.String())
	if err != nil {
		t.Fatal(err)
	}
	f.Close()

	identities, err := internal.ParseAgeIdentityFile(identityFile)
	if err != nil {
		t.Fatal(err)
	}

	if len(identities) != 1 {
		t.Fatalf("expected 1 identity, got %d", len(identities))
	}

	parsedIdentity, ok := identities[0].(*age.X25519Identity)
	if !ok {
		t.Fatal("identity is not of type X25519Identity")
	}

	if parsedIdentity.String() != identity.String() {
		t.Fatal("parsed identity does not match original identity")
	}
}

func TestDecryptShareFor(t *testing.T) {
	identity, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatal(err)
	}

	recipient := identity.Recipient().String()
	share := []byte("test share")

	encrypted, err := internal.EncryptShare(recipient, share)
	if err != nil {
		t.Fatal(err)
	}

	tempDir := t.TempDir()
	identityFile := filepath.Join(tempDir, "key.txt")

	f, err := os.Create(identityFile)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	_, err = f.WriteString(identity.String())
	if err != nil {
		t.Fatal(err)
	}
	f.Close()

	decrypted, err := internal.DecryptShareFor(encrypted, identityFile)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(share, decrypted) {
		t.Fatal("decrypted share does not match original share")
	}
}
