package age

import (
	"filippo.io/age"
	"github.com/root-gg/utils"
)

// Config object for the age crypto backend
type Config struct {
	Passphrase string
	Recipient  string

	// Recipients holds resolved age.Recipient objects (SSH or X25519).
	// Populated during Configure() when Recipient is set.
	Recipients []age.Recipient `json:"-"`

	// DecryptHint stores the decrypt command hint for Comments().
	// Set during Configure() based on the recipient type.
	DecryptHint string `json:"-"`
}

// NewAgeBackendConfig instantiate a new Backend Configuration
// from config map passed as argument
func NewAgeBackendConfig(params map[string]any) (config *Config) {
	config = new(Config)
	utils.Assign(config, params)
	return
}
