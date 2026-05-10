export type NativeRuntime = "wails" | "browser";

type AnyFn = (...args: any[]) => any;

interface WailsRuntimeLike {
  EventsEmit?: AnyFn;
  EventsOn?: AnyFn;
}

function getWindow(): Window & { go?: any; runtime?: WailsRuntimeLike } {
  return window as Window & { go?: any; runtime?: WailsRuntimeLike };
}

export function getNativeRuntime(): NativeRuntime {
  if (typeof window === "undefined") return "browser";

  const w = getWindow();
  if (w.go || w.runtime) return "wails";

  return "browser";
}

export function isNativeApp(): boolean {
  return getNativeRuntime() === "wails";
}

export async function invoke<T = unknown>(
  command: string,
  args?: Record<string, any>,
): Promise<T> {
  if (getNativeRuntime() === "wails") {
    const w = getWindow();
    const app = w.go?.main?.App;
    if (!app || typeof app.Invoke !== "function") {
      throw new Error("Wails bridge is not initialized");
    }

    return app.Invoke(command, args ?? {});
  }

  throw new Error("Native invoke is unavailable in browser runtime");
}

export async function emit(event: string, payload?: any): Promise<void> {
  if (getNativeRuntime() !== "wails") return;

  const w = getWindow();
  if (typeof w.runtime?.EventsEmit === "function") {
    w.runtime.EventsEmit(event, payload);
  }
}

export async function listen<T = unknown>(
  event: string,
  handler: (payload: T) => void,
): Promise<() => void> {
  if (getNativeRuntime() !== "wails") return () => {};

  const w = getWindow();
  if (typeof w.runtime?.EventsOn === "function") {
    const off = w.runtime.EventsOn(event, handler);
    if (typeof off === "function") return off;
  }

  return () => {};
}

export async function convertFileSrc(path: string): Promise<string> {
  if (getNativeRuntime() === "wails") {
    return `wails://wails/${encodeURI(path)}`;
  }

  return path;
}

export async function detectPlatform(): Promise<string> {
  if (getNativeRuntime() === "wails") {
    try {
      const value = await invoke<string>("platform");
      return value || "unknown";
    } catch {
      return "unknown";
    }
  }

  return "unknown";
}

export async function openUrl(url: string): Promise<void> {
  if (typeof window !== "undefined") {
    window.open(url, "_blank", "noopener,noreferrer");
  }
}

export async function openPath(path: string): Promise<void> {
  if (typeof window !== "undefined") {
    window.open(path, "_blank", "noopener,noreferrer");
  }
}

export async function closeCurrentWindow(): Promise<void> {
  if (getNativeRuntime() === "wails") {
    await invoke("window_close");
  }
}

export async function windowStartDragging(): Promise<void> {
  if (getNativeRuntime() === "wails") {
    await invoke("window_start_dragging");
  }
}

export async function windowHide(): Promise<void> {
  if (getNativeRuntime() === "wails") {
    await invoke("window_hide");
  }
}

export async function windowMinimize(): Promise<void> {
  if (getNativeRuntime() === "wails") {
    await invoke("window_minimize");
  }
}

export async function windowToggleMaximize(): Promise<void> {
  if (getNativeRuntime() === "wails") {
    await invoke("window_toggle_maximize");
  }
}

export async function windowIsMaximized(): Promise<boolean> {
  if (getNativeRuntime() === "wails") {
    return invoke<boolean>("window_is_maximized");
  }
  return false;
}

export interface NativeUpdateInfo {
  version: string;
  notes: string;
  install: (onProgress?: (progress: number) => void) => Promise<void>;
}

export async function checkForNativeUpdate(): Promise<NativeUpdateInfo | null> {
  if (getNativeRuntime() !== "wails") return null;

  const update = await invoke<{ version: string; notes?: string } | null>("updater_check");
  if (!update) return null;

  return {
    version: update.version,
    notes: update.notes ?? "",
    install: async (onProgress?: (progress: number) => void) => {
      await invoke("updater_download_and_install");
      if (onProgress) onProgress(100);
    },
  };
}

export async function relaunchNativeApp(): Promise<void> {
  if (getNativeRuntime() === "wails") {
    await invoke("app_relaunch");
  }
}

export async function pickDirectory(title = "Choose folder"): Promise<string | null> {
  if (getNativeRuntime() === "wails") {
    return invoke<string | null>("dialog_pick_directory", { title });
  }
  return null;
}

export async function isAutoStartEnabledNative(): Promise<boolean> {
  if (getNativeRuntime() === "wails") {
    return invoke<boolean>("autostart_is_enabled");
  }
  return false;
}

export async function setAutoStartEnabledNative(enabled: boolean): Promise<void> {
  if (getNativeRuntime() === "wails") {
    await invoke("autostart_set_enabled", { enabled });
  }
}

export async function requestNotificationPermissionNative(): Promise<boolean> {
  if (getNativeRuntime() === "wails") {
    return invoke<boolean>("notifications_request_permission");
  }
  return false;
}

export async function sendNotificationNative(title: string, body: string): Promise<void> {
  if (getNativeRuntime() === "wails") {
    await invoke("notifications_send", { title, body });
  }
}

export async function registerNativeHotkeys(handlers: {
  onQuickAsk?: () => void;
  onToggleSidePanel?: () => void;
}): Promise<void> {
  // Wails implementation will be added with native hooks.
  // Keep as a no-op to preserve behavior in browser and avoid regressions.
  void handlers;
}

export async function unregisterNativeHotkeys(): Promise<void> {
  // no-op for now
}

export async function writeNativeLog(
  level: "trace" | "debug" | "info" | "warn" | "error",
  message: string,
): Promise<void> {
  if (getNativeRuntime() === "wails") {
    await invoke("log_write", { level, message });
  }
}
