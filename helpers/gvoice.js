#!/usr/bin/env node
/**
 * Google Voice Cookie Extractor for mautrix-gvoice
 *
 * Opens Chrome, lets you log in, then extracts the required cookies.
 * Output is copied to clipboard and printed to stdout.
 *
 * Usage: node gvoice.js
 * Requires: npm install puppeteer-core
 */
const { spawn, execSync } = require("child_process");
const fs = require("fs");
const os = require("os");
const path = require("path");
const puppeteer = require("puppeteer-core");

const CHROME = "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome";
const PORT = 9223;

const URLS = [
  "https://voice.google.com",
  "https://accounts.google.com",
  "https://myaccount.google.com",
];

const REQUIRED = ["SID","HSID","SSID","OSID","APISID","SAPISID"];
const OPTIONAL = ["__Secure-1PSIDTS"];

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

function pickBest(name, cookies) {
  const hits = cookies.filter(c => c.name === name && c.value);
  if (!hits.length) return null;
  const score = (c) => {
    const d = (c.domain || "").toLowerCase();
    let s = 0;
    if (d === ".google.com" || d === "google.com") s += 50;
    if (d.endsWith(".google.com")) s += 40;
    if (d.includes("voice.google.com")) s += 30;
    if (d.includes("accounts.google.com")) s += 20;
    if (c.path === "/") s += 5;
    return s;
  };
  hits.sort((a,b) => score(b) - score(a));
  return hits[0];
}

(async () => {
  console.log("Google Voice Cookie Extractor");
  console.log("=============================");
  console.log("1. A Chrome window will open to voice.google.com");
  console.log("2. Log in with your Google account");
  console.log("3. Cookies will be extracted and copied to clipboard\n");

  const profile = fs.mkdtempSync(path.join(os.tmpdir(), "gvoice-"));
  const chrome = spawn(CHROME, [
    `--remote-debugging-port=${PORT}`,
    `--user-data-dir=${profile}`,
    "--no-first-run",
    "--no-default-browser-check",
    URLS[0],
  ], { stdio: "ignore" });

  const killChrome = () => { try { process.kill(chrome.pid, "SIGTERM"); } catch {} };

  try {
    await waitDevtools();

    const b = await puppeteer.connect({ browserURL: `http://127.0.0.1:${PORT}` });
    const page = (await b.pages())[0] || await b.newPage();
    const cdp = await page.target().createCDPSession();

    console.log("Waiting for login (5 minute timeout)...");

    const deadline = Date.now() + 5 * 60 * 1000;
    let out = {};
    let allCookies = [];

    while (Date.now() < deadline) {
      // Visit domains that tend to set required cookies
      for (const u of URLS) {
        try { await page.goto(u, { waitUntil: "domcontentloaded", timeout: 20000 }); } catch {}
      }

      const res = await cdp.send("Storage.getCookies");
      allCookies = res.cookies || [];

      out = {};
      for (const name of [...REQUIRED, ...OPTIONAL]) {
        const best = pickBest(name, allCookies);
        if (best) out[name] = best.value;
      }

      if (REQUIRED.every(n => out[n])) break;
      await sleep(400);
    }

    if (!REQUIRED.every(n => out[n])) {
      const present = Object.keys(out).sort();
      throw new Error(`Missing required cookies. Present: ${present.join(", ") || "(none)"}`);
    }

    const json = JSON.stringify(out);
    console.log("\n✓ Cookies extracted successfully!\n");
    console.log("Send this to @gvoicebot:\n");
    console.log(json);

    try {
      execSync("pbcopy", { input: json });
      console.log("\n(Copied to clipboard)");
    } catch {}

    await b.close();
  } finally {
    killChrome();
    try { fs.rmSync(profile, { recursive: true, force: true }); } catch {}
  }
})().catch(e => {
  console.error(String(e?.message || e));
  process.exit(1);
});
