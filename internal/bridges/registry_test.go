package bridges

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegistryLoaded(t *testing.T) {
	// Registry should be loaded automatically
	assert.NotEmpty(t, registry)
}

func TestGet(t *testing.T) {
	whatsapp := Get("whatsapp")
	require.NotNil(t, whatsapp)
	assert.Equal(t, "whatsapp", whatsapp.Name)
	assert.Equal(t, 29318, whatsapp.Port)

	nonexistent := Get("nonexistent")
	assert.Nil(t, nonexistent)
}

func TestList(t *testing.T) {
	all := List()
	assert.Len(t, all, 13)

	// Should be sorted by name
	names := make([]string, len(all))
	for i, b := range all {
		names[i] = b.Name
	}
	assert.Equal(t, []string{"bluesky", "discord", "gmessages", "googlechat", "gvoice", "irc", "linkedin", "meta", "signal", "slack", "telegram", "twitter", "whatsapp"}, names)
}

func TestNames(t *testing.T) {
	names := Names()
	assert.Len(t, names, 13)
	assert.Equal(t, []string{"bluesky", "discord", "gmessages", "googlechat", "gvoice", "irc", "linkedin", "meta", "signal", "slack", "telegram", "twitter", "whatsapp"}, names)
}

func TestExists(t *testing.T) {
	assert.True(t, Exists("whatsapp"))
	assert.True(t, Exists("discord"))
	assert.False(t, Exists("nonexistent"))
	assert.False(t, Exists(""))
}

func TestBridgeDetails(t *testing.T) {
	tests := []struct {
		name    string
		port    int
		hasNote bool
	}{
		{"whatsapp", 29318, false},
		{"signal", 29313, false},
		{"discord", 29316, false},
		{"gmessages", 29314, false},
		{"bluesky", 29325, false},
		{"irc", 29326, false},
		{"googlechat", 29320, true},
		{"gvoice", 29321, true},
		{"slack", 29315, true},
		{"meta", 29323, true},
		{"twitter", 29324, true},
		{"linkedin", 29322, true},
		{"telegram", 29317, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := Get(tt.name)
			require.NotNil(t, b)
			assert.Equal(t, tt.name, b.Name)
			assert.Equal(t, tt.port, b.Port)
			assert.Equal(t, tt.hasNote, b.HasNote())
			assert.NotEmpty(t, b.Description)
			assert.NotEmpty(t, b.LoginInstructions)
		})
	}
}
