package bridges

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBridgeInfo_Image(t *testing.T) {
	b := BridgeInfo{Name: "whatsapp"}
	assert.Equal(t, "dock.mau.dev/mautrix/whatsapp:latest", b.Image())
}

func TestBridgeInfo_BotUsername(t *testing.T) {
	b := BridgeInfo{Name: "telegram"}
	assert.Equal(t, "telegrambot", b.BotUsername())
}

func TestBridgeInfo_NamespacePrefix(t *testing.T) {
	b := BridgeInfo{Name: "signal"}
	assert.Equal(t, "signal_", b.NamespacePrefix())
}

func TestBridgeInfo_ServiceName(t *testing.T) {
	b := BridgeInfo{Name: "discord"}
	assert.Equal(t, "mautrix-discord", b.ServiceName())
}

func TestBridgeInfo_HasNote(t *testing.T) {
	withNote := BridgeInfo{Note: "Uses unofficial API"}
	assert.True(t, withNote.HasNote())

	withoutNote := BridgeInfo{}
	assert.False(t, withoutNote.HasNote())
}
