/**
 * Global hotkey registration via the shared runtime adapter.
 */
import {
  registerNativeHotkeys,
  unregisterNativeHotkeys,
} from "$lib/native/runtime";

export interface HotkeyHandlers {
  onQuickAsk?: () => void;
  onToggleSidePanel?: () => void;
}

export async function registerHotkeys(handlers: HotkeyHandlers): Promise<void> {
  try {
    await registerNativeHotkeys(handlers);
  } catch (err) {
    console.warn("[Hotkeys] Failed to register shortcuts:", err);
  }
}

export async function unregisterHotkeys(): Promise<void> {
  try {
    await unregisterNativeHotkeys();
  } catch {
    // Silently fail
  }
}
