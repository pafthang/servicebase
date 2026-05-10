/**
 * Native notification helpers via the shared runtime adapter.
 */
import {
  requestNotificationPermissionNative,
  sendNotificationNative,
} from "$lib/native/runtime";

export async function requestNotificationPermission(): Promise<boolean> {
  try {
    return await requestNotificationPermissionNative();
  } catch {
    return false;
  }
}

export async function sendNotification(title: string, body: string): Promise<void> {
  try {
    await sendNotificationNative(title, body);
  } catch {
    // Silently fail in non-Tauri environments
  }
}

/**
 * Notify when agent completes a task while window is not focused.
 */
export async function notifyAgentComplete(summary: string): Promise<void> {
  if (typeof document !== "undefined" && document.hasFocus()) return;
  await sendNotification("PocketPaw", summary);
}

/**
 * Notify when Guardian AI blocks something.
 */
export async function notifyGuardianBlock(detail: string): Promise<void> {
  await sendNotification("PocketPaw - Guardian AI", detail);
}
