type ServerEvent = { type: string; payload: any };
type Listener = (event: ServerEvent) => void;

const RECONNECT_DELAY_MS = 2000;

let socket: WebSocket | null = null;
let listeners: Listener[] = [];
let openListeners: Array<() => void> = [];
let currentToken: string | null = null;
let currentBaseUrl = "";
let reconnectTimer: ReturnType<typeof setTimeout> | null = null;

function openSocket() {
  if (!currentToken) return;

  socket = new WebSocket(`${currentBaseUrl}/ws?token=${currentToken}`);

  socket.onopen = () => {
    openListeners.forEach((listener) => listener());
  };

  socket.onmessage = (raw) => {
    try {
      const event: ServerEvent = JSON.parse(raw.data);
      listeners.forEach((listener) => listener(event));
    } catch {
      // ignore malformed frames
    }
  };

  // A closed socket is only ever a symptom (network blip, backgrounded tab,
  // backend restart) — without this, the instructor's location heartbeat
  // silently stops going out and they fall out of matching with no visible
  // sign anything is wrong. Reconnect as long as we still have a token.
  socket.onclose = () => {
    socket = null;
    if (currentToken) {
      reconnectTimer = setTimeout(openSocket, RECONNECT_DELAY_MS);
    }
  };

  socket.onerror = () => {
    socket?.close();
  };
}

export function connect(token: string, wsBaseUrl: string) {
  currentToken = token;
  currentBaseUrl = wsBaseUrl;
  if (socket) return;
  openSocket();
}

export function disconnect() {
  currentToken = null;
  if (reconnectTimer) {
    clearTimeout(reconnectTimer);
    reconnectTimer = null;
  }
  socket?.close();
  socket = null;
}

export function send(type: string, payload: unknown) {
  socket?.send(JSON.stringify({ type, payload }));
}

export function subscribe(listener: Listener) {
  listeners.push(listener);
  return () => {
    listeners = listeners.filter((l) => l !== listener);
  };
}

export function subscribeOpen(listener: () => void) {
  openListeners.push(listener);
  return () => {
    openListeners = openListeners.filter((l) => l !== listener);
  };
}
