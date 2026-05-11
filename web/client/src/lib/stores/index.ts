// Re-export all existing stores
export * from './chat.svelte';
export * from './sessions.svelte';
export * from './settings.svelte';
export * from './activity.svelte';
export * from './explorer.svelte';
export * from './kits.svelte';
export * from './metrics.svelte';
export * from './mission-control.svelte';
export * from './platform.svelte';
export * from './projects.svelte';
export * from './skills.svelte';
export * from './ui.svelte';
export * from './connection.svelte';

// New stores
export * from './agents.svelte';
export * from './providers.svelte';

import { connectionStore } from './connection.svelte';

export async function initializeStores(
  token: string,
  baseUrl?: string,
  wsToken?: string,
) {
  await connectionStore.initialize(token, baseUrl, wsToken);
}
