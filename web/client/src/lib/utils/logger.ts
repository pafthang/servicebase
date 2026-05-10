// Lightweight logger that writes to the Tauri log plugin (file + stdout + webview)
// when running inside a Tauri app, and falls back to console otherwise.
//
// Usage:
//   import { logger } from "$lib/utils/logger";
//   logger.info("Connected to backend");
//   logger.error("Failed to load sessions", err);

import { writeNativeLog } from "$lib/native/runtime";

type LogFn = (...args: unknown[]) => void;

interface Logger {
  trace: LogFn;
  debug: LogFn;
  info: LogFn;
  warn: LogFn;
  error: LogFn;
}

let _loading: Promise<void> | null = null;

async function loadPlugin() {
  try {
    await writeNativeLog("debug", "[logger] native log bridge ready");
  } catch {
    // Native logger not available
  }
}

function log(level: "trace" | "debug" | "info" | "warn" | "error", args: unknown[]) {
  const message = args
    .map((a) => (a instanceof Error ? `${a.message}\n${a.stack}` : String(a)))
    .join(" ");

  // Always write to console
  // eslint-disable-next-line no-console
  console[level === "trace" ? "debug" : level](message);

  // Also write to Tauri log plugin (async, fire-and-forget)
  if (!_loading) {
    _loading = loadPlugin();
  }
  writeNativeLog(level, message).catch(() => {});
}

export const logger: Logger = {
  trace: (...args) => log("trace", args),
  debug: (...args) => log("debug", args),
  info: (...args) => log("info", args),
  warn: (...args) => log("warn", args),
  error: (...args) => log("error", args),
};
