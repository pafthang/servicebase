// Thin wrapper around Tauri invoke for persisting OAuth tokens to disk
import { invoke } from "$lib/native/runtime";

export interface OAuthTokens {
  access_token: string;
  refresh_token: string | null;
  expires_at: number;
  scopes: string[];
}

export async function readTokens(): Promise<OAuthTokens | null> {
  try {
    return await invoke<OAuthTokens>("read_oauth_tokens");
  } catch {
    return null;
  }
}

export async function saveTokens(tokens: OAuthTokens): Promise<void> {
  await invoke("save_oauth_tokens", { tokens });
}

export async function clearTokens(): Promise<void> {
  await invoke("clear_oauth_tokens");
}
