const api = typeof browser !== "undefined" ? browser : chrome;
const HOST_NAME = "com.example.tabhost";

let port = null;
let lastKey = null;
let timer = null;

function connect() {
  if (port) return;

  port = api.runtime.connectNative(HOST_NAME);

  port.onDisconnect.addListener(() => {
    console.log("Native host disconnected");
    port = null;
    setTimeout(connect, 1000);
  });
}

function scheduleSend() {
  clearTimeout(timer);
  timer = setTimeout(sendActiveTab, 150);
}

async function sendActiveTab() {
  try {
    connect();

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

    port.postMessage(payload);
    console.log("Sent to native app:", payload);
  } catch (e) {
    console.debug("send error:", e);
  }
}

api.tabs.onActivated.addListener(scheduleSend);
api.tabs.onUpdated.addListener((_, changeInfo) => {
  if (changeInfo.url) scheduleSend();
});
api.windows.onFocusChanged.addListener(scheduleSend);

connect();