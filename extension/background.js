const api = typeof browser !== "undefined" ? browser : chrome;
const API_URL = "http://127.0.0.1:8384/tab";

let lastKey = null;
let timer = null;

function scheduleSend() {
  clearTimeout(timer);
  timer = setTimeout(sendActiveTab, 150);
}

async function sendActiveTab() {
  try {
    const tabs = await api.tabs.query({ active: true, currentWindow: true });
    const tab = tabs[0];
    if (!tab || !tab.url) return;

    const payload = {
      tabId: tab.id,
      title: tab.title || "",
      url: tab.url,
      ts: Date.now()
    };

    const key = payload.tabId + payload.url;
    if (key === lastKey) return;
    lastKey = key;

    await fetch(API_URL, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload)
    });

    console.log("Sent to tracker:", payload);
  } catch (e) {
    console.debug("send error:", e);
  }
}

api.tabs.onActivated.addListener(scheduleSend);
api.tabs.onUpdated.addListener((_, changeInfo) => {
  if (changeInfo.url) scheduleSend();
});
api.windows.onFocusChanged.addListener(scheduleSend);

// Send initial tab on load
scheduleSend();
