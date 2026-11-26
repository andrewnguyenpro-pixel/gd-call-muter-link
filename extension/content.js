// Runs inside discord.com and clicks the Deafen button when told to.
function findDeafenButton() {
  for (const b of document.querySelectorAll('button[aria-label], [role="switch"][aria-label]')) {
    const label = (b.getAttribute("aria-label") || "").toLowerCase();
    if (label === "deafen" || label === "undeafen") return b;
  }
  return null;
}

function isDeafened(btn) {
  const c = btn.getAttribute("aria-checked");
  if (c === "true") return true;
  if (c === "false") return false;
  return (btn.getAttribute("aria-label") || "").toLowerCase() === "undeafen";
}

chrome.runtime.onMessage.addListener(msg => {
  if (!msg || msg.type !== "action") return;
  const btn = findDeafenButton();
  if (!btn) return console.log("[GD Auto-Deafen] Deafen button not found");
  const deaf = isDeafened(btn);
  if (msg.action === "toggle" ||
      (msg.action === "deafen" && !deaf) || (msg.action === "undeafen" && deaf)) btn.click();
});

function ping() { try { chrome.runtime.sendMessage({ type: "ping" }).catch(() => {}); } catch (e) {} }
ping();
setInterval(ping, 20000);
