#!/usr/bin/env node
/**
 * Meta (Facebook/Instagram) Cookie Extractor for mautrix-meta
 *
 * Opens Chrome, lets you log in to Facebook, then extracts the required cookies.
 * Output is copied to clipboard and printed to stdout.
 *
 * Usage: node meta.js
 * Requires: npm install puppeteer-core
 *
 * Docs: https://docs.mau.fi/bridges/go/meta/authentication.html
 */
const { spawn, execSync } = require("child_process");
const fs = require("fs");
const os = require("os");
const path = require("path");
const puppeteer = require("puppeteer-core");

const CHROME = "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome";
const PORT = 9225;
const URL = "https://www.facebook.com";
const NEED = ["c_user", "xs", "datr"];

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
  console.log("Meta (Facebook/Instagram) Cookie Extractor");
  console.log("==========================================");
  console.log("1. A Chrome window will open to Facebook");
  console.log("2. Log in with your Facebook account");
  console.log("3. Cookies will be extracted and copied to clipboard\n");

  const profile = fs.mkdtempSync(path.join(os.tmpdir(), "meta-"));
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
        const cookie = cookies.find((c) => c.name === name && c.domain.includes("facebook.com"));
        if (cookie) out[name] = cookie.value;
      }
      if (haveAll(out)) break;
      await sleep(500);
    }

    if (!haveAll(out)) {
      throw new Error("Timed out waiting for cookies. Make sure you completed login in the Chrome window.");
    }

    // Format as cURL header for mautrix-meta
    const cookieStr = NEED.map(n => `${n}=${out[n]}`).join("; ");
    const curlCmd = `curl 'https://www.facebook.com/' -H 'cookie: ${cookieStr}'`;

    console.log("\n✓ Cookies extracted successfully!\n");
    console.log("Send this to @metabot:\n");
    console.log("login-cookies");
    console.log("\nThen paste this cURL command when prompted:\n");
    console.log(curlCmd);

    // Copy to clipboard (macOS)
    try {
      execSync("pbcopy", { input: curlCmd });
      console.log("\n(cURL command copied to clipboard)");
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
