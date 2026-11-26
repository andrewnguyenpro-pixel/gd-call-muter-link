// Keeps one request open to GD Call Muter v10 (127.0.0.1:47322).
// The app answers the moment something changes:
//   discord: deafen / undeafen -> clicks the Deafen button in discord.com tabs
//   insta:   mute / unmute     -> mutes or unmutes every instagram.com tab
let seq = null, dseq = null, iseq = null;
let looping = false;

const DISCORD = ["https://discord.com/*", "https://*.discord.com/*"];
const INSTA = ["https://www.instagram.com/*", "https://instagram.com/*"];

async function discord(action) {
  for (const t of await chrome.tabs.query({ url: DISCORD }))
    chrome.tabs.sendMessage(t.id, { type: "action", action }).catch(() => {});
}

async function insta(action) {
  for (const t of await chrome.tabs.query({ url: INSTA }))
    chrome.tabs.update(t.id, { muted: action === "mute" }).catch(() => {});
}

async function loop() {
  if (looping) return;
  looping = true;
  while (true) {
    try {
      const url = "http://127.0.0.1:47322/wait?seq=" + (seq === null ? -1 : seq);
      const s = await (await fetch(url, { cache: "no-store" })).json();
      if (dseq !== null && s.dseq !== dseq && s.discord !== "none") discord(s.discord);
      if (iseq !== null && s.iseq !== iseq && s.insta !== "none") insta(s.insta);
      seq = s.seq; dseq = s.dseq; iseq = s.iseq;
    } catch (e) {
      await new Promise(r => setTimeout(r, 1000)); // app not open yet; retry
    }
  }
}

// Open discord.com / instagram.com tabs ping us every 20 s to keep this worker awake.
chrome.runtime.onMessage.addListener((msg, sender, reply) => {
  if (msg && msg.type === "ping") { loop(); reply({ ok: true }); }
});
chrome.runtime.onStartup.addListener(loop);
chrome.runtime.onInstalled.addListener(loop);
loop();
