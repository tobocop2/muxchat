#!/usr/bin/env node
/**
 * Google Chat Cookie Extractor for mautrix-googlechat
 *
 * Opens Chrome, lets you log in, then extracts the required cookies.
 * Output is copied to clipboard and printed to stdout.
 *
 * Usage: node googlechat.js
 * Requires: npm install puppeteer-core
 */
const { spawn, execSync } = require("child_process");
const fs = require("fs");
const os = require("os");
const path = require("path");
const puppeteer = require("puppeteer-core");

const CHROME = "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome";
const PORT = 9222;
const URL = "https://chat.google.com";
const NEED = ["COMPASS", "SSID", "SID", "OSID", "HSID"];

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
  throw new Error("DevTools not reachable on 127.0.0.1:9222");
}

function haveAll(out) {
  return NEED.every((n) => out[n.toLowerCase()]);
}

(async () => {
  console.log("Google Chat Cookie Extractor");
  console.log("============================");
  console.log("1. A Chrome window will open to chat.google.com");
  console.log("2. Log in with your Google Workspace account");
  console.log("3. Cookies will be extracted and copied to clipboard\n");

  const profile = fs.mkdtempSync(path.join(os.tmpdir(), "gchat-"));
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
        const hits = cookies.filter((c) => c.name === name);
        const pick = name === "COMPASS" ? (hits.find((c) => c.path === "/") || hits[0]) : hits[0];
        if (pick) out[name.toLowerCase()] = pick.value;
      }
      if (haveAll(out)) break;
      await sleep(500);
    }

    if (!haveAll(out)) {
      throw new Error("Timed out waiting for cookies. Make sure you completed login in the Chrome window.");
    }

    const json = JSON.stringify(out);
    console.log("\n✓ Cookies extracted successfully!\n");
    console.log("Send this to @googlechatbot:\n");
    console.log(json);

    // Copy to clipboard (macOS)
    try {
      execSync("pbcopy", { input: json });
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
