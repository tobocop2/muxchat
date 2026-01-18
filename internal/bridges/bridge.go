package bridges

import "fmt"

// BridgeInfo contains metadata about a messaging bridge
type BridgeInfo struct {
	Name                   string `yaml:"-"` // Set from map key
	Description            string `yaml:"description"`
	Port                   int    `yaml:"port"`
	Note                   string `yaml:"note,omitempty"`                     // Optional note about the bridge
	RequiresAPICredentials bool   `yaml:"requires_api_credentials,omitempty"` // Requires user to provide API credentials
	LoginInstructions      string `yaml:"login_instructions"`
}

// Image returns the Docker image for this bridge
func (b BridgeInfo) Image() string {
	return fmt.Sprintf("dock.mau.dev/mautrix/%s:latest", b.Name)
}

// BotUsername returns the Matrix username for the bridge bot
func (b BridgeInfo) BotUsername() string {
	return b.Name + "bot"
}

// NamespacePrefix returns the namespace prefix for bridge users/rooms
func (b BridgeInfo) NamespacePrefix() string {
	return b.Name + "_"
}

// ServiceName returns the Docker Compose service name
func (b BridgeInfo) ServiceName() string {
	return "mautrix-" + b.Name
}

// HasNote returns true if this bridge has a note
func (b BridgeInfo) HasNote() bool {
	return b.Note != ""
}
