package main_test

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"database/sql"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"filippo.io/age"
	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
	"github.com/soatok/freon/coordinator/internal"
	"github.com/stretchr/testify/require"
)

var (
	clientBinPath      string
	coordinatorBinPath string
)

func TestMain(m *testing.M) {
	tempDir, err := os.MkdirTemp("", "freon-bins-")
	if err != nil {
		fmt.Printf("could not create temp dir: %v", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tempDir)

	// Build client
	clientBinPath = filepath.Join(tempDir, "client")
	cmdClient := exec.Command("go", "build", "-o", getExePath(clientBinPath), "../../client")
	if err := cmdClient.Run(); err != nil {
		fmt.Printf("could not build client: %v", err)
		os.Exit(1)
	}
	clientBinPath = getExePath(clientBinPath)

	// Build coordinator
	coordinatorBinPath = filepath.Join(tempDir, "coordinator")
	cmdCoord := exec.Command("go", "build", "-o", getExePath(coordinatorBinPath), "..")
	if err := cmdCoord.Run(); err != nil {
		fmt.Printf("could not build coordinator: %v", err)
		os.Exit(1)
	}
	coordinatorBinPath = getExePath(coordinatorBinPath)

	exitCode := m.Run()

	os.Exit(exitCode)
}

// Kludge to make Windows testing succeed
func getExePath(path string) string {
	if runtime.GOOS == "windows" {
		return path + ".exe"
	}
	return path
}

// coordinator holds the state of a coordinator instance
type coordinator struct {
	db       *sql.DB
	proc     *os.Process
	hostname string
}

// client holds the state of a client instance
type client struct {
	homeDir      string
	ageKey       *age.X25519Identity
	agePubKey    string
	identityFile string
}

// startCoordinator starts a new coordinator instance on a random port
func startCoordinator(t *testing.T) *coordinator {
	t.Helper()

	// Create a temporary directory for the coordinator's database
	dir, err := os.MkdirTemp("", "freon-coordinator-test")
	require.NoError(t, err)
	t.Cleanup(func() { os.RemoveAll(dir) })

	dbFile := filepath.Join(dir, "database.sqlite")
	db, err := sql.Open("sqlite3", dbFile)
	require.NoError(t, err)

	// Ensure foreign keys
	_, err = db.Exec("PRAGMA foreign_keys = ON")
	require.NoError(t, err)
	internal.DbEnsureTablesExist(db)

	// Find an available port
	port, err := findAvailablePort()
	require.NoError(t, err)
	hostname := fmt.Sprintf("localhost:%d", port)

	// Create a temporary config file for the coordinator
	configFile, err := os.CreateTemp("", "freon-coordinator-config-*.json")
	require.NoError(t, err)
	defer os.Remove(configFile.Name())

	f, err := os.OpenFile(configFile.Name(), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	require.NoError(t, err)
	dbFileForJson := strings.ReplaceAll(dbFile, `\`, `\\`)
	_, err = f.WriteString(fmt.Sprintf(`{"hostname": "%s", "database": "%s"}`, hostname, dbFileForJson))
	require.NoError(t, err)
	f.Close()

	// Start the coordinator
	cmd := exec.Command(coordinatorBinPath)
	cmd.Env = append(os.Environ(), "FREON_COORDINATOR_CONFIG="+configFile.Name())
	var coordOutput bytes.Buffer
	cmd.Stdout = &coordOutput
	cmd.Stderr = &coordOutput
	err = cmd.Start()
	require.NoError(t, err)

	// Wait for the coordinator to be ready
	waitForCoordinator(t, hostname, &coordOutput)

	return &coordinator{
		db:       db,
		proc:     cmd.Process,
		hostname: hostname,
	}
}

// stop stops the coordinator instance
func (c *coordinator) stop(t *testing.T) {
	t.Helper()
	err := c.proc.Kill()
	require.NoError(t, err)
	c.db.Close()
}

// newClient creates a new client instance with its own home directory
func newClient(t *testing.T) *client {
	t.Helper()

	homeDir, err := os.MkdirTemp("", "freon-client-test")
	require.NoError(t, err)
	t.Cleanup(func() { os.RemoveAll(homeDir) })

	ageKey, err := age.GenerateX25519Identity()
	require.NoError(t, err)

	identityFile := filepath.Join(homeDir, "keys.age")
	err = os.WriteFile(identityFile, []byte(ageKey.String()), 0600)
	require.NoError(t, err)

	return &client{
		homeDir:      homeDir,
		ageKey:       ageKey,
		agePubKey:    ageKey.Recipient().String(),
		identityFile: identityFile,
	}
}

// run runs a freon client command
func (c *client) run(t *testing.T, args ...string) (string, error) {
	t.Helper()

	cmd := exec.Command(clientBinPath, args...)
	cmd.Env = append(os.Environ(), "FREON_HOME="+c.homeDir)

	output, err := cmd.CombinedOutput()
	return string(output), err
}

// waitForCoordinator waits for the coordinator to be ready to accept connections
func waitForCoordinator(t *testing.T, host string, coordOutput *bytes.Buffer) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			t.Logf("Coordinator logs:\n%s", coordOutput.String())
			t.Fatal("coordinator did not start in time")
		default:
			conn, err := net.DialTimeout("tcp", host, 1*time.Second)
			if err == nil {
				conn.Close()
				return
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
}

// findAvailablePort finds an available TCP port on the local machine
func findAvailablePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}

func TestIntegration(t *testing.T) {
	// Start coordinator
	coord := startCoordinator(t)
	defer coord.stop(t)

	// Create clients
	numClients := 4
	threshold := 3
	clients := make([]*client, numClients)
	for i := 0; i < numClients; i++ {
		clients[i] = newClient(t)
	}

	// DKG ceremony
	var groupID string
	t.Run("DKG", func(t *testing.T) {
		// Client 0 creates the DKG
		output, err := clients[0].run(t, "keygen", "create", "-h", coord.hostname, "-n", fmt.Sprintf("%d", numClients), "-t", fmt.Sprintf("%d", threshold))
		require.NoError(t, err, output)
		re := regexp.MustCompile(`Group ID:\s*(\S+)`)
		matches := re.FindStringSubmatch(output)
		require.Len(t, matches, 2)
		groupID = matches[1]

		// Other clients join the DKG
		var wg sync.WaitGroup
		for i := 0; i < numClients; i++ {
			wg.Add(1)
			time.Sleep(100 * time.Millisecond)
			go func(i int) {
				defer wg.Done()
				out, err := clients[i].run(t, "keygen", "join", "-h", coord.hostname, "-g", groupID, "-r", clients[i].agePubKey)
				require.NoError(t, err, out)
			}(i)
		}
		wg.Wait()
	})

	// Signing ceremony
	t.Run("SignAndVerify", func(t *testing.T) {
		message := "test message"
		messageFile := filepath.Join(clients[0].homeDir, "message.txt")
		err := os.WriteFile(messageFile, []byte(message), 0644)
		require.NoError(t, err)

		// Client 0 creates the signing ceremony
		output, err := clients[0].run(t, "sign", "create", "-h", coord.hostname, "-g", groupID, messageFile)
		require.NoError(t, err, output)
		re := regexp.MustCompile(`created!\s*(\S+)`)
		matches := re.FindStringSubmatch(output)
		require.Len(t, matches, 2)
		ceremonyID := matches[1]

		// First `threshold` clients join the signing ceremony
		var wg sync.WaitGroup
		for i := 0; i < threshold; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				output, err := clients[i].run(t, "sign", "join", "-h", coord.hostname, "-c", ceremonyID, "-i", clients[i].identityFile, messageFile)
				require.NoError(t, err, output)
			}(i)
		}
		wg.Wait()

		// Get the signature
		output, err = clients[0].run(t, "sign", "get", "-h", coord.hostname, "-c", ceremonyID)
		require.NoError(t, err, output)

		// Extract signature from output
		re = regexp.MustCompile(`Signature:\s*(\S+)`)
		matches = re.FindStringSubmatch(output)
		require.Len(t, matches, 2)
		sigHex := matches[1]
		signature, err := hex.DecodeString(sigHex)
		require.NoError(t, err)

		// Get the group public key
		output, err = clients[0].run(t, "keygen", "list")
		require.NoError(t, err, output)
		re = regexp.MustCompile(groupID + `\s+([a-f0-9]+)`)
		matches = re.FindStringSubmatch(output)
		require.Len(t, matches, 2, "could not find public key for group in client 0 output")
		pubKeyHex := matches[1]
		pubKey, err := hex.DecodeString(pubKeyHex)
		require.NoError(t, err)

		// Verify the Ed25519 signature
		verified := ed25519.Verify(pubKey, []byte(message), signature)
		require.True(t, verified, "Ed25519 signature verification failed")
	})
}
