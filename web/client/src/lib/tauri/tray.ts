/**
 * Frontend tray state management.
 * Listens for tray events emitted by the Rust backend.
 */

import { listen } from "$lib/native/runtime";

export interface TrayEventHandlers {
  onNavigate?: (path: string) => void;
}

const unlisten: (() => void)[] = [];

export async function setupTrayListeners(handlers: TrayEventHandlers): Promise<void> {
  try {
    if (handlers.onNavigate) {
      const cb = handlers.onNavigate;
      const u = await listen<string>("tray-navigate", cb);
      unlisten.push(u);
    }
  } catch (err) {
    console.warn("[Tray] Failed to setup listeners:", err);
  }
}

export function cleanupTrayListeners(): void {
  for (const u of unlisten) u();
  unlisten.length = 0;
}
