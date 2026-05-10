/**
 * Auto-start helpers via the shared native runtime adapter.
 */
import {
  isAutoStartEnabledNative,
  setAutoStartEnabledNative,
} from "$lib/native/runtime";

export async function isAutoStartEnabled(): Promise<boolean> {
  try {
    return await isAutoStartEnabledNative();
  } catch {
    return false;
  }
}

export async function enableAutoStart(): Promise<void> {
  try {
    await setAutoStartEnabledNative(true);
  } catch (err) {
    console.warn("[Autostart] Failed to enable:", err);
  }
}

export async function disableAutoStart(): Promise<void> {
  try {
    await setAutoStartEnabledNative(false);
  } catch (err) {
    console.warn("[Autostart] Failed to disable:", err);
  }
}

export async function toggleAutoStart(enabled: boolean): Promise<void> {
  if (enabled) {
    await enableAutoStart();
  } else {
    await disableAutoStart();
  }
}
