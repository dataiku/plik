package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/docopt/docopt-go"
	"github.com/stretchr/testify/require"
)

// makeOpts returns a docopt.Opts map with sensible zero-value defaults
// for every flag/option the CLI declares. Callers can override specific
// keys before passing to UnmarshalArgs.
func makeOpts() docopt.Opts {
	return docopt.Opts{
		"FILE":              []string{},
		"--debug":           false,
		"--quiet":           false,
		"--json":            false,
		"--server":          nil,
		"--name":            nil,
		"--oneshot":         false,
		"--removable":       false,
		"--stream":          false,
		"--ttl":             nil,
		"--extend-ttl":      false,
		"--comments":        nil,
		"-p":                false,
		"--password":        nil,
		"-a":                false,
		"--archive":         nil,
		"--compress":        nil,
		"--archive-options": nil,
		"-s":                false,
		"--not-secure":      false,
		"--secure":          nil,
		"--cipher":          nil,
		"--passphrase":      nil,
		"--recipient":       nil,
		"--secure-options":  nil,
		"--insecure":        false,
		"--update":          false,
		"--login":           false,
		"--mcp":             false,
		"--version":         false,
		"--info":            false,
		"--help":            false,
		"--token":           nil,
		"--stdin":           false,
	}
}

// --- TTL parsing ---

func TestUnmarshalArgs_TTL_Minutes(t *testing.T) {
	config := NewUploadConfig()
	opts := makeOpts()
	opts["--ttl"] = "5m"

	err := config.UnmarshalArgs(opts)
	require.NoError(t, err)
	require.Equal(t, 300, config.TTL) // 5 * 60
}

func TestUnmarshalArgs_TTL_Hours(t *testing.T) {
	config := NewUploadConfig()
	opts := makeOpts()
	opts["--ttl"] = "2h"

	err := config.UnmarshalArgs(opts)
	require.NoError(t, err)
	require.Equal(t, 7200, config.TTL) // 2 * 3600
}

func TestUnmarshalArgs_TTL_Days(t *testing.T) {
	config := NewUploadConfig()
	opts := makeOpts()
	opts["--ttl"] = "1d"

	err := config.UnmarshalArgs(opts)
	require.NoError(t, err)
	require.Equal(t, 86400, config.TTL) // 1 * 86400
}

func TestUnmarshalArgs_TTL_Seconds(t *testing.T) {
	config := NewUploadConfig()
	opts := makeOpts()
	opts["--ttl"] = "3600"

	err := config.UnmarshalArgs(opts)
	require.NoError(t, err)
	require.Equal(t, 3600, config.TTL) // raw seconds
}

func TestUnmarshalArgs_TTL_Invalid(t *testing.T) {
	config := NewUploadConfig()
	opts := makeOpts()
	opts["--ttl"] = "abc"

	err := config.UnmarshalArgs(opts)
	require.Error(t, err)
	require.Contains(t, err.Error(), "Invalid TTL")
}

// --- Password parsing ---

func TestUnmarshalArgs_Password_LoginPassword(t *testing.T) {
	config := NewUploadConfig()
	opts := makeOpts()
	opts["--password"] = "admin:secret"

	err := config.UnmarshalArgs(opts)
	require.NoError(t, err)
	require.Equal(t, "admin", config.Login)
	require.Equal(t, "secret", config.Password)
}

func TestUnmarshalArgs_Password_DefaultLogin(t *testing.T) {
	config := NewUploadConfig()
	opts := makeOpts()
	opts["--password"] = "mysecret"

	err := config.UnmarshalArgs(opts)
	require.NoError(t, err)
	require.Equal(t, "plik", config.Login)
	require.Equal(t, "mysecret", config.Password)
}

func TestUnmarshalArgs_Password_ColonInPassword(t *testing.T) {
	config := NewUploadConfig()
	opts := makeOpts()
	opts["--password"] = "user:pass:word"

	err := config.UnmarshalArgs(opts)
	require.NoError(t, err)
	require.Equal(t, "user", config.Login)
	require.Equal(t, "pass:word", config.Password)
}

// --- Boolean flags ---

func TestUnmarshalArgs_Flags(t *testing.T) {
	tests := []struct {
		flag  string
		field func(c *CliConfig) bool
		name  string
	}{
		{"--oneshot", func(c *CliConfig) bool { return c.OneShot }, "OneShot"},
		{"--removable", func(c *CliConfig) bool { return c.Removable }, "Removable"},
		{"--stream", func(c *CliConfig) bool { return c.Stream }, "Stream"},
		{"--quiet", func(c *CliConfig) bool { return c.Quiet }, "Quiet"},
		{"--debug", func(c *CliConfig) bool { return c.Debug }, "Debug"},
		{"--extend-ttl", func(c *CliConfig) bool { return c.ExtendTTL }, "ExtendTTL"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NewUploadConfig()
			opts := makeOpts()
			opts[tt.flag] = true

			err := config.UnmarshalArgs(opts)
			require.NoError(t, err)
			require.True(t, tt.field(config))
		})
	}
}

func TestUnmarshalArgs_JSON_ImpliesQuiet(t *testing.T) {
	config := NewUploadConfig()
	opts := makeOpts()
	opts["--json"] = true

	err := config.UnmarshalArgs(opts)
	require.NoError(t, err)
	require.True(t, config.JSON)
	require.True(t, config.Quiet)
}

// --- Server override ---

func TestUnmarshalArgs_ServerOverride(t *testing.T) {
	config := NewUploadConfig()
	opts := makeOpts()
	opts["--server"] = "https://plik.example.com"

	err := config.UnmarshalArgs(opts)
	require.NoError(t, err)
	require.Equal(t, "https://plik.example.com", config.URL)
}

// --- Secure mode ---

func TestUnmarshalArgs_SecureEnabled(t *testing.T) {
	config := NewUploadConfig()
	opts := makeOpts()
	opts["-s"] = true

	err := config.UnmarshalArgs(opts)
	require.NoError(t, err)
	require.True(t, config.Secure)
}

func TestUnmarshalArgs_SecureExplicitBackend(t *testing.T) {
	config := NewUploadConfig()
	opts := makeOpts()
	opts["--secure"] = "age"

	err := config.UnmarshalArgs(opts)
	require.NoError(t, err)
	require.True(t, config.Secure)
	require.Equal(t, "age", config.SecureMethod)
}

func TestUnmarshalArgs_NotSecure(t *testing.T) {
	config := NewUploadConfig()
	config.Secure = true // pre-set from config file
	opts := makeOpts()
	opts["--not-secure"] = true

	err := config.UnmarshalArgs(opts)
	require.NoError(t, err)
	require.False(t, config.Secure)
}

// --- Archive mode ---

func TestUnmarshalArgs_ArchiveShortFlag(t *testing.T) {
	config := NewUploadConfig()
	opts := makeOpts()
	opts["-a"] = true

	err := config.UnmarshalArgs(opts)
	require.NoError(t, err)
	require.True(t, config.Archive)
}

func TestUnmarshalArgs_ArchiveExplicitBackend(t *testing.T) {
	config := NewUploadConfig()
	opts := makeOpts()
	opts["--archive"] = "zip"

	err := config.UnmarshalArgs(opts)
	require.NoError(t, err)
	require.True(t, config.Archive)
	require.Equal(t, "zip", config.ArchiveMethod)
}

// --- Token handling ---

func TestUnmarshalArgs_Token(t *testing.T) {
	config := NewUploadConfig()
	opts := makeOpts()
	opts["--token"] = "my-upload-token"

	err := config.UnmarshalArgs(opts)
	require.NoError(t, err)
	require.Equal(t, "my-upload-token", config.Token)
}

func TestUnmarshalArgs_StdinOverride(t *testing.T) {
	config := NewUploadConfig()
	config.DisableStdin = true
	opts := makeOpts()
	opts["--stdin"] = true

	err := config.UnmarshalArgs(opts)
	require.NoError(t, err)
	require.False(t, config.DisableStdin)
}

// --- Config file loading ---

func TestLoadConfigFromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".plikrc")
	err := os.WriteFile(path, []byte(`
URL = "https://upload.example.com"
OneShot = true
TTL = 3600
DownloadBinary = "wget"
`), 0600)
	require.NoError(t, err)

	config, err := LoadConfigFromFile(path)
	require.NoError(t, err)
	require.Equal(t, "https://upload.example.com", config.URL)
	require.True(t, config.OneShot)
	require.Equal(t, 3600, config.TTL)
	require.Equal(t, "wget", config.DownloadBinary)
}

func TestLoadConfigFromFile_MissingFile(t *testing.T) {
	_, err := LoadConfigFromFile("/nonexistent/plikrc")
	require.Error(t, err)
}

func TestLoadConfigFromFile_URLTrailingSlash(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".plikrc")
	err := os.WriteFile(path, []byte(`
URL = "https://upload.example.com/"
`), 0600)
	require.NoError(t, err)

	config, err := LoadConfigFromFile(path)
	require.NoError(t, err)
	require.Equal(t, "https://upload.example.com", config.URL, "trailing slash should be stripped")
}

// --- NewUploadConfig defaults ---

func TestNewUploadConfig_Defaults(t *testing.T) {
	config := NewUploadConfig()
	require.Equal(t, "http://127.0.0.1:8080", config.URL)
	require.Equal(t, "tar", config.ArchiveMethod)
	require.Equal(t, "openssl", config.SecureMethod)
	require.Equal(t, "curl", config.DownloadBinary)
	require.False(t, config.Debug)
	require.False(t, config.Quiet)
	require.False(t, config.OneShot)
}

// --- Comments ---

func TestUnmarshalArgs_Comments(t *testing.T) {
	config := NewUploadConfig()
	opts := makeOpts()
	opts["--comments"] = "This is a test upload"

	err := config.UnmarshalArgs(opts)
	require.NoError(t, err)
	require.Equal(t, "This is a test upload", config.Comments)
}

// --- Filename override ---

func TestUnmarshalArgs_FilenameOverride(t *testing.T) {
	config := NewUploadConfig()
	opts := makeOpts()
	opts["--name"] = "custom-name.txt"

	err := config.UnmarshalArgs(opts)
	require.NoError(t, err)
	require.Equal(t, "custom-name.txt", config.filenameOverride)
}
