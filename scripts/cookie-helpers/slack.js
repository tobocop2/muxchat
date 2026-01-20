#!/usr/bin/env node
/**
 * Slack Token Extractor for mautrix-slack
 *
 * Opens Chrome, lets you log in to Slack, then extracts the required token and cookie.
 * Output is copied to clipboard and printed to stdout.
 *
 * Usage: node slack.js
 * Requires: npm install puppeteer-core
 */
const { spawn, execSync } = require("child_process");
const fs = require("fs");
const os = require("os");
const path = require("path");
const puppeteer = require("puppeteer-core");

const CHROME = "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome";
const PORT = 9224;
const URL = "https://slack.com/signin";

const sleep = (ms) => new Promise(r => setTimeout(r, ms));

async function waitDevtools(timeoutMs = 20000) {
  const t0 = Date.now();
  while (Date.now() - t0 < timeoutMs) {
    try {
      const r = await fetch(`http://127.0.0.1:${PORT}/json/version`);
      if (r.ok) return;
    } catch {}
    await sleep(200);
  }
  throw new Error(`DevTools not reachable on 127.0.0.1:${PORT}`);
}

(async () => {
  console.log("Slack Token Extractor");
  console.log("=====================");
  console.log("1. A Chrome window will open to Slack");
  console.log("2. Log in to your Slack workspace");
  console.log("3. Make sure you're IN a workspace (seeing channels/messages)");
  console.log("4. Token and cookie will be extracted automatically\n");

  const profile = fs.mkdtempSync(path.join(os.tmpdir(), "slack-"));
  const chrome = spawn(CHROME, [
    `--remote-debugging-port=${PORT}`,
    `--user-data-dir=${profile}`,
    "--no-first-run",
    "--no-default-browser-check",
    URL,
  ], { stdio: "ignore" });

  const killChrome = () => { try { process.kill(chrome.pid, "SIGTERM"); } catch {} };

  try {
    await waitDevtools();

    const b = await puppeteer.connect({ browserURL: `http://127.0.0.1:${PORT}` });

    console.log("Waiting for login (5 minute timeout)...");
    console.log("Looking for token and cookie...\n");

    let token = null;
    let cookie = null;
    let pollCount = 0;
    let lastUrl = "";

    const deadline = Date.now() + 5 * 60 * 1000;

    while (Date.now() < deadline) {
      pollCount++;

      // Show progress every 10 polls (5 seconds)
      if (pollCount % 10 === 0) {
        process.stdout.write(`  Still looking... (${Math.floor(pollCount/2)}s)\r`);
      }

      // Get all pages and check each one
      const pages = await b.pages();

      for (const page of pages) {
        try {
          const url = page.url();

          // Log URL changes
          if (url !== lastUrl && url.includes('slack.com')) {
            lastUrl = url;
            console.log(`  Page: ${url.substring(0, 60)}...`);
          }

          // Check cookies via CDP
          if (!cookie) {
            try {
              const cdp = await page.target().createCDPSession();
              const res = await cdp.send("Network.getAllCookies");
              const cookies = res.cookies || [];
              const dCookie = cookies.find(c => c.name === 'd' && c.value?.startsWith('xoxd-'));
              if (dCookie) {
                cookie = dCookie.value;
                console.log("✓ Found cookie (d)");
              }
              await cdp.detach();
            } catch {}
          }

          // Only check localStorage on app.slack.com pages that look like a workspace
          if (!token && url.includes('app.slack.com/client/')) {
            const result = await page.evaluate(() => {
              const config = localStorage.getItem('localConfig_v2');
              if (!config) {
                return { token: null, debug: { hasConfig: false, url: location.href } };
              }

              // Try to parse and find token
              try {
                const parsed = JSON.parse(config);
                if (parsed.teams) {
                  const teamIds = Object.keys(parsed.teams);
                  for (const teamId of teamIds) {
                    const team = parsed.teams[teamId];
                    if (team.token && team.token.startsWith('xoxc-')) {
                      return { token: team.token, source: `teams.${teamId}` };
                    }
                  }
                }
              } catch {}

              // Fallback: regex match
              const match = config.match(/xoxc-[a-zA-Z0-9-]+/);
              if (match) {
                return { token: match[0], source: 'regex' };
              }

              return { token: null, debug: { hasConfig: true, length: config.length, hasXoxc: config.includes('xoxc-') } };
            });

            if (result?.token) {
              token = result.token;
              console.log(`✓ Found token (from ${result.source})`);
            } else if (result?.debug && pollCount % 20 === 0) {
              console.log(`\n  Debug: hasConfig=${result.debug.hasConfig}, length=${result.debug.length || 0}, hasXoxc=${result.debug.hasXoxc}`);
            }
          }
        } catch {}
      }

      if (token && cookie) {
        console.log(""); // Clear progress line
        const cmd = `login token ${token} ${cookie}`;
        console.log("\n✓ Credentials extracted successfully!\n");
        console.log("Send this to @slackbot:\n");
        console.log(cmd);

        try {
          execSync("pbcopy", { input: cmd });
          console.log("\n(Copied to clipboard)");
        } catch {}

        await b.close();
        killChrome();
        try { fs.rmSync(profile, { recursive: true, force: true }); } catch {}
        process.exit(0);
      }

      await sleep(500);
    }

    // Timeout
    console.log("");
    const missing = [];
    if (!token) missing.push("token (xoxc-)");
    if (!cookie) missing.push("cookie (xoxd-)");
    throw new Error(`Timeout. Missing: ${missing.join(", ")}.\n\nMake sure you:\n1. Logged into a workspace\n2. Can see channels and messages\n3. URL contains 'app.slack.com/client/'`);

  } catch (e) {
    killChrome();
    try { fs.rmSync(profile, { recursive: true, force: true }); } catch {}
    console.error("\n" + String(e?.message || e));
    process.exit(1);
  }
})();
