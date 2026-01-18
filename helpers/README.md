# Muxchat Helpers

Cookie extraction helpers for bridges that require manual authentication.

**Most bridges support QR code login** - you don't need these helpers for:
- WhatsApp (`login qr`)
- Signal (`login`)
- Discord (`login-qr`)
- Google Messages (`login`)

These helpers are only needed for bridges that require cookie/token extraction:
- **Google Chat** - requires Google Workspace cookies
- **Google Voice** - requires Google account cookies
- **Slack** - requires xoxc token + xoxd cookie

## Alternative: mautrix-manager

Instead of using these scripts, you can use [mautrix-manager](https://github.com/mautrix/manager) which provides a web UI for automated cookie extraction.

## Setup

```bash
cd helpers
npm install
```

Requires Chrome installed at `/Applications/Google Chrome.app` (macOS).

## Usage

### Google Chat
```bash
npm run googlechat
# or: node googlechat.js
```
Opens Chrome to chat.google.com. Log in with your Google Workspace account. Cookies are extracted and copied to clipboard. Paste the JSON to @googlechatbot.

### Google Voice
```bash
npm run gvoice
# or: node gvoice.js
```
Opens Chrome to voice.google.com. Log in with your Google account. Cookies are extracted and copied to clipboard. Paste the JSON to @gvoicebot.

### Slack
```bash
npm run slack
# or: node slack.js
```
Opens Chrome to app.slack.com. Log in to your workspace. Token and cookie are extracted. The full `login token <token> <cookie>` command is copied to clipboard. Paste it to @slackbot.

## How it works

These scripts:
1. Launch Chrome with remote debugging enabled
2. Open the service's login page
3. Wait for you to complete login
4. Extract the required cookies/tokens via Chrome DevTools Protocol
5. Output the credentials and copy to clipboard
6. Clean up the temporary Chrome profile
