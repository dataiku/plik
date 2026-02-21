package age

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolveRecipients_X25519(t *testing.T) {
	// Valid native age recipient
	recipients, _, err := resolveRecipients("age1ql3z7hjy54pw3hyww5ayyfg7zqgvc7w3j2elw8zmrj2kg5sfn9aqmcac8p")
	require.NoError(t, err)
	require.Len(t, recipients, 1)
}

func TestResolveRecipients_InvalidX25519(t *testing.T) {
	_, _, err := resolveRecipients("not-a-valid-recipient")
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid age recipient")
}

func TestResolveRecipients_SSHKey(t *testing.T) {
	// Valid SSH ed25519 key
	key := "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIN6urz7ksxlezuqXw40WL7AK++XKSGAFZG95NZFgpzev test@example"
	recipients, _, err := resolveRecipients(key)
	require.NoError(t, err)
	require.Len(t, recipients, 1)
}

func TestResolveRecipients_InvalidSSHKey(t *testing.T) {
	_, _, err := resolveRecipients("ssh-ed25519 invalid-key-data")
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid SSH key")
}

func TestResolveRecipients_URL(t *testing.T) {
	// Mock server returning SSH keys
	key := "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIN6urz7ksxlezuqXw40WL7AK++XKSGAFZG95NZFgpzev test@example"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(key + "\n"))
	}))
	defer server.Close()

	recipients, _, err := resolveRecipients(server.URL)
	require.NoError(t, err)
	require.Len(t, recipients, 1)
}

func TestResolveRecipients_URLNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	_, _, err := resolveRecipients(server.URL)
	require.Error(t, err)
	require.Contains(t, err.Error(), "404 Not Found")
}

func TestResolveRecipients_URLNoSupportedKeys(t *testing.T) {
	// Server returns only unsupported key types
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ecdsa-sha2-nistp256 AAAA...\n"))
	}))
	defer server.Close()

	_, _, err := resolveRecipients(server.URL)
	require.Error(t, err)
	require.Contains(t, err.Error(), "no supported SSH keys")
}

func TestResolveRecipients_URLMultipleKeys(t *testing.T) {
	// Server returns multiple keys, some supported, some not
	keys := "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIN6urz7ksxlezuqXw40WL7AK++XKSGAFZG95NZFgpzev key1\n" +
		"ecdsa-sha2-nistp256 not-supported\n" +
		"ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIN6urz7ksxlezuqXw40WL7AK++XKSGAFZG95NZFgpzev key2\n"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(keys))
	}))
	defer server.Close()

	recipients, _, err := resolveRecipients(server.URL)
	require.NoError(t, err)
	require.Len(t, recipients, 2)
}

func TestResolveRecipients_SSHHostEmpty(t *testing.T) {
	_, _, err := resolveRecipients("ssh://")
	require.Error(t, err)
	require.Contains(t, err.Error(), "empty SSH hostname")
}

func TestResolveRecipients_SSHHostUnreachable(t *testing.T) {
	// Port 1 should not have an SSH server
	_, _, err := resolveRecipients("ssh://127.0.0.1:1")
	require.Error(t, err)
}

func TestResolveRecipients_GitHubShorthand(t *testing.T) {
	_, _, err := resolveRecipients("@")
	require.Error(t, err)
	require.Contains(t, err.Error(), "empty GitHub username")
}

func TestEncryptWithSSHRecipient(t *testing.T) {
	key := "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIN6urz7ksxlezuqXw40WL7AK++XKSGAFZG95NZFgpzev test"
	recipients, _, err := resolveRecipients(key)
	require.NoError(t, err)

	backend := &Backend{Config: &Config{Recipients: recipients}}
	plaintext := "hello world"
	reader, err := backend.Encrypt(bytes.NewBufferString(plaintext))
	require.NoError(t, err)

	encrypted, err := io.ReadAll(reader)
	require.NoError(t, err)
	require.NotEmpty(t, encrypted)
	require.NotEqual(t, plaintext, string(encrypted))
	// Encrypted data should start with the age header
	require.Contains(t, string(encrypted), "age-encryption.org/v1")
}

func TestEncryptWithPassphrase(t *testing.T) {
	backend := &Backend{Config: &Config{Passphrase: "test-passphrase"}}
	plaintext := "hello world"
	reader, err := backend.Encrypt(bytes.NewBufferString(plaintext))
	require.NoError(t, err)

	encrypted, err := io.ReadAll(reader)
	require.NoError(t, err)
	require.NotEmpty(t, encrypted)
	require.Contains(t, string(encrypted), "age-encryption.org/v1")
}
