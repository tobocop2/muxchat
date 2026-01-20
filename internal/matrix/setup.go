package matrix

import (
	"fmt"

	"github.com/tobias/muxbee/internal/bridges"
	"github.com/tobias/muxbee/internal/config"
)

// BotWelcomeMessages contains welcome messages for each bridge bot
var BotWelcomeMessages = map[string]string{
	"whatsapp": `👋 Welcome to the WhatsApp bridge!

To connect your WhatsApp account:

Option 1: QR Code
1. Send: login qr
2. Scan the QR code with WhatsApp on your phone
   (Settings → Linked Devices → Link a Device)

Option 2: Pairing Code
1. Send: login phone
2. Enter the 8-letter code in WhatsApp on your phone
   (Settings → Linked Devices → Link a Device → Link with phone number instead)

Note: Your phone must stay online. After 14 days offline, you'll need to relink.

Commands:
• login qr - Link via QR code
• login phone - Link via pairing code
• logout - Disconnect your WhatsApp
• ping - Check connection status
• help - Show all commands

📖 Docs: https://docs.mau.fi/bridges/go/whatsapp/authentication.html`,

	"telegram": `👋 Welcome to the Telegram bridge!

This bridge requires API credentials from https://my.telegram.org
(You provided these when enabling the bridge via muxbee)

To connect your Telegram account:
1. Send: login
2. Enter your phone number when prompted
3. Enter the verification code from Telegram
4. If you have 2FA enabled, enter your password

Commands:
• login - Start linking your Telegram
• logout - Disconnect your Telegram
• ping - Check connection status
• help - Show all commands

📖 Docs: https://docs.mau.fi/bridges/python/telegram/authentication.html`,

	"signal": `👋 Welcome to the Signal bridge!

To connect your Signal account:
1. Open Signal on your phone
2. Go to Settings → Linked Devices → Add New Device
3. Send: login
4. Scan the QR code shown here with Signal

Note: Message history is not available - Signal doesn't support syncing history to linked devices.

Commands:
• login - Start linking your Signal
• logout - Disconnect your Signal
• ping - Check connection status
• help - Show all commands

📖 Docs: https://docs.mau.fi/bridges/go/signal/authentication.html`,

	"gmessages": `👋 Welcome to the Google Messages bridge!

To connect your Google Messages:
1. Send: login qr
2. Open Google Messages on your phone
3. Tap ⋮ menu → Device pairing → QR code scanner
4. Scan the QR code shown here

Note: Your phone must be connected to the internet for the bridge to work.

Commands:
• login qr - Link via QR code
• logout - Disconnect Google Messages
• ping - Check connection status
• help - Show all commands

📖 Docs: https://docs.mau.fi/bridges/go/gmessages/authentication.html`,

	"googlechat": `👋 Welcome to the Google Chat bridge!

This bridge requires cookie extraction from your browser.

You need these cookies from https://chat.google.com:
  COMPASS, SSID, SID, OSID, HSID

Then send them as JSON:
  login-cookie {"compass":"...","ssid":"...","sid":"...","osid":"...","hsid":"..."}

Commands:
• login-cookie <json> - Login with cookies
• logout - Disconnect Google Chat
• ping - Check connection status
• help - Show all commands

📖 Auth docs: https://docs.mau.fi/bridges/python/googlechat/authentication.html
📖 Helper scripts: https://github.com/tobocop2/muxbee#helper-scripts`,

	"discord": `👋 Welcome to the Discord bridge!

⚠️ This bridge uses token auth which may violate Discord ToS.
Discord may ban accounts that appear suspicious. Consider using a bot account.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Option 1: QR Login (easiest, may hit CAPTCHA)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
1. Send: login-qr
2. Scan the QR code with Discord mobile app
3. Approve the login on your phone

If you hit a CAPTCHA, use token login instead.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Option 2: Token Login
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
1. Open Discord in a private/incognito browser window
2. Log in to Discord
3. Open DevTools (F12 or Cmd+Shift+I)
4. Go to Network tab, filter for "api"
5. Reload the page (Ctrl+R or Cmd+R)
6. Click any successful request (status 200)
7. Find "Authorization" header in Request Headers
8. Copy the token value

Send the token here:
login-token user YOUR_TOKEN

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Option 3: Bot Token (safest for your account)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
1. Go to https://discord.com/developers/applications
2. Create a new application
3. Go to Bot section, create bot, copy token
4. Enable "Server Members Intent" and "Message Content Intent"
5. Send: login-token bot YOUR_BOT_TOKEN
6. Use OAuth2 URL Generator to add bot to your servers

Commands:
• login-qr - Login via QR code
• login-token user <token> - Login with user token
• login-token bot <token> - Login with bot token
• logout - Disconnect Discord
• guilds - List joined servers
• ping - Check connection status
• help - Show all commands

📖 Docs: https://docs.mau.fi/bridges/go/discord/authentication.html`,

	"slack": `👋 Welcome to the Slack bridge!

⚠️ This bridge uses token auth which may violate Slack ToS.

This bridge requires token and cookie extraction from your browser.

You need:
• Token (starts with xoxc-) from localStorage
• Cookie (starts with xoxd-) named "d" from app.slack.com

Then send both:
  login token xoxc-your-token xoxd-your-cookie

Commands:
• login token <token> <cookie> - Connect with your credentials
• logout - Disconnect Slack
• ping - Check connection status
• help - Show all commands

📖 Auth docs: https://docs.mau.fi/bridges/go/slack/authentication.html
📖 Helper scripts: https://github.com/tobocop2/muxbee#helper-scripts`,

	"gvoice": `👋 Welcome to the Google Voice bridge!

This bridge requires cookie extraction from your browser.

You need these cookies from https://voice.google.com:
  SID, HSID, SSID, OSID, APISID, SAPISID

Then send them as JSON:
  login-cookie {"sid":"...","hsid":"...","ssid":"...","osid":"...","apisid":"...","sapisid":"..."}

Commands:
• login-cookie <json> - Login with cookies
• logout - Disconnect Google Voice
• ping - Check connection status
• help - Show all commands

📖 Auth docs: https://docs.mau.fi/bridges/go/gvoice/authentication.html
📖 Helper scripts: https://github.com/tobocop2/muxbee#helper-scripts`,

	"linkedin": `👋 Welcome to the LinkedIn bridge!

⚠️ This bridge uses unofficial API access which may violate LinkedIn ToS.

To connect your LinkedIn account:
1. Send: login
2. Enter your LinkedIn email or phone number
3. Enter your password or verification code

Commands:
• login - Start linking LinkedIn
• logout - Disconnect LinkedIn
• ping - Check connection status
• help - Show all commands

📖 Docs: https://github.com/mautrix/linkedin`,
}

// SetupBotsForUser creates DM rooms with all enabled bridge bots and sends welcome messages
func SetupBotsForUser(cfg *config.Config) error {
	if len(cfg.EnabledBridges) == 0 {
		return nil
	}

	// Connect to homeserver
	client := NewClient(fmt.Sprintf("http://localhost:%d", cfg.SynapsePort()))

	// Login as admin
	if err := client.Login(cfg.Admin.Username, cfg.Admin.Password); err != nil {
		return fmt.Errorf("failed to login as admin: %w", err)
	}

	// Create DM with each enabled bot
	for _, bridgeName := range cfg.EnabledBridges {
		bridge := bridges.Get(bridgeName)
		if bridge == nil {
			continue
		}

		botUserID := fmt.Sprintf("@%s:%s", bridge.BotUsername(), cfg.ServerName)

		// Get or create DM room
		roomID, isNew, err := client.GetOrCreateDirectMessage(botUserID)
		if err != nil {
			fmt.Printf("  Note: Could not create room with %s: %v\n", bridge.BotUsername(), err)
			continue
		}

		// Send welcome message if we have one (only for new rooms to avoid spam)
		if isNew {
			if welcomeMsg, ok := BotWelcomeMessages[bridgeName]; ok {
				if err := client.SendMessage(roomID, welcomeMsg); err != nil {
					fmt.Printf("  Note: Could not send welcome message to %s\n", bridgeName)
				}
			}
			fmt.Printf("  ✓ Created chat with %s bot\n", bridgeName)
		} else {
			fmt.Printf("  ✓ Found existing chat with %s bot\n", bridgeName)
		}
	}

	return nil
}

// SetupBotForBridge creates a DM room with a specific bridge bot and sends a welcome message.
// This is called automatically when a bridge is enabled via the TUI.
// Returns nil on success or error. Errors are non-fatal - the bridge still works,
// the user just needs to find the bot manually.
func SetupBotForBridge(cfg *config.Config, bridgeName string) error {
	bridge := bridges.Get(bridgeName)
	if bridge == nil {
		return fmt.Errorf("unknown bridge: %s", bridgeName)
	}

	// Connect to homeserver
	client := NewClient(fmt.Sprintf("http://localhost:%d", cfg.SynapsePort()))

	// Login as admin
	if err := client.Login(cfg.Admin.Username, cfg.Admin.Password); err != nil {
		return fmt.Errorf("failed to login: %w", err)
	}

	botUserID := fmt.Sprintf("@%s:%s", bridge.BotUsername(), cfg.ServerName)

	// Get or create DM room
	roomID, isNew, err := client.GetOrCreateDirectMessage(botUserID)
	if err != nil {
		return fmt.Errorf("could not create room with %s: %w", bridge.BotUsername(), err)
	}

	// Send welcome message if we have one (only for new rooms to avoid spam)
	if isNew {
		if welcomeMsg, ok := BotWelcomeMessages[bridgeName]; ok {
			// Ignore error - welcome message is nice-to-have
			client.SendMessage(roomID, welcomeMsg)
		}
	}

	return nil
}

// CleanupBotForBridge leaves and forgets the DM room with a bridge bot.
// This is called automatically when a bridge is disabled via the TUI.
// Returns nil on success or error. Errors are non-fatal.
func CleanupBotForBridge(cfg *config.Config, bridgeName string) error {
	bridge := bridges.Get(bridgeName)
	if bridge == nil {
		return fmt.Errorf("unknown bridge: %s", bridgeName)
	}

	// Connect to homeserver
	client := NewClient(fmt.Sprintf("http://localhost:%d", cfg.SynapsePort()))

	// Login as admin
	if err := client.Login(cfg.Admin.Username, cfg.Admin.Password); err != nil {
		return fmt.Errorf("failed to login: %w", err)
	}

	botUserID := fmt.Sprintf("@%s:%s", bridge.BotUsername(), cfg.ServerName)

	// Find the DM room
	roomID, err := client.FindDirectMessageRoom(botUserID)
	if err != nil || roomID == "" {
		// No room to leave
		return nil
	}

	// Leave and forget the room
	return client.LeaveRoom(roomID)
}
