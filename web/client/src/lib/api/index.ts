export * from "./types";
export { BACKEND_URL, API_PREFIX, API_BASE } from "./config";
export { PocketPawClient } from "./client";
export { PocketPawWebSocket, type ConnectionState } from "./websocket";

import { PocketPawClient } from "./client";

export const client = new PocketPawClient();