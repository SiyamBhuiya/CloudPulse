export const API_URL = import.meta.env.VITE_API_URL ?? "http://localhost:8080";
export const WS_URL = import.meta.env.VITE_WS_URL ?? "ws://localhost:8080/ws";

export const TOKEN_KEY = "cloudpulse_token";
export const USER_KEY = "cloudpulse_user";

export const HISTORY_SIZE = 60;     // seconds shown on the live graph
export const NETWORK_MAX_MBPS = 60; // network line is drawn as % of this