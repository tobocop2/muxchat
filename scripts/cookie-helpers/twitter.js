#!/usr/bin/env node
/**
 * Twitter/X Cookie Extractor for mautrix-twitter
 *
 * Opens Chrome, lets you log in to Twitter, then extracts the required cookies.
 * Output is copied to clipboard and printed to stdout.
 *
 * Usage: node twitter.js
 * Requires: npm install puppeteer-core
 *
 * Docs: https://docs.mau.fi/bridges/go/twitter/authentication.html
 */
const { spawn, execSync } = require("child_process");
const fs = require("fs");
const os = require("os");
const path = require("path");
const puppeteer = require("puppeteer-core");

const CHROME = "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome";
const PORT = 9226;
const URL = "https://twitter.com/login";
const NEED = ["ct0", "auth_token"];

const sleep = (ms) => new Promise((r) => setTimeout(r, ms));

async function waitForDevtools(timeoutMs = 20000) {
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

function haveAll(cookies) {
  return NEED.every((n) => cookies[n]);
}

(async () => {
  console.log("Twitter/X Cookie Extractor");
  console.log("==========================");
  console.log("1. A Chrome window will open to Twitter");
  console.log("2. Log in with your Twitter account");
  console.log("3. Cookies will be extracted and copied to clipboard\n");

  const profile = fs.mkdtempSync(path.join(os.tmpdir(), "twitter-"));
  const chrome = spawn(
    CHROME,
    [
      `--remote-debugging-port=${PORT}`,
      `--user-data-dir=${profile}`,
      "--no-first-run",
      "--no-default-browser-check",
      URL,
    ],
    { stdio: "ignore" }
  );

  const killChrome = () => {
    try { process.kill(chrome.pid, "SIGTERM"); } catch {}
  };

  try {
    await waitForDevtools();

    const browser = await puppeteer.connect({ browserURL: `http://127.0.0.1:${PORT}` });
    const pages = await browser.pages();
    const page = pages[0] || (await browser.newPage());

    const cdp = await page.target().createCDPSession();

    console.log("Waiting for login (5 minute timeout)...");

    const t0 = Date.now();
    let out = {};
    while (Date.now() - t0 < 5 * 60 * 1000) {
      const { cookies } = await cdp.send("Network.getAllCookies");
      out = {};
      for (const name of NEED) {
        // Twitter cookies can be on twitter.com or x.com
        const cookie = cookies.find((c) => c.name === name && (c.domain.includes("twitter.com") || c.domain.includes("x.com")));
        if (cookie) out[name] = cookie.value;
      }
      if (haveAll(out)) break;
      await sleep(500);
    }

    if (!haveAll(out)) {
      throw new Error("Timed out waiting for cookies. Make sure you completed login in the Chrome window.");
    }

    // Format login command for mautrix-twitter
    const loginCmd = `login ${out.ct0} ${out.auth_token}`;

    console.log("\n✓ Cookies extracted successfully!\n");
    console.log("Send this to @twitterbot:\n");
    console.log(loginCmd);

    // Copy to clipboard (macOS)
    try {
      execSync("pbcopy", { input: loginCmd });
      console.log("\n(Copied to clipboard)");
    } catch {}

    try { await browser.close(); } catch {}

  } finally {
    killChrome();
    try { fs.rmSync(profile, { recursive: true, force: true }); } catch {}
  }
})().catch((e) => {
  console.error(String(e?.message || e));
  process.exit(1);
});
