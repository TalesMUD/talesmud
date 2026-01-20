const defaultOrigin =
  typeof window !== "undefined" ? window.location.origin : "http://localhost:8010";

const defaultWsOrigin =
  typeof window !== "undefined"
    ? `${window.location.protocol === "https:" ? "wss" : "ws"}://${window.location.host}`
    : "ws://localhost:8010";

// Allow overrides via env (in case you host API elsewhere), otherwise default to same-origin.
const backend = import.meta?.env?.VITE_API_BASE_URL || `${defaultOrigin}/api`;
const wsbackend = import.meta?.env?.VITE_WS_BASE_URL || `${defaultWsOrigin}/ws`;

export { backend, wsbackend };
