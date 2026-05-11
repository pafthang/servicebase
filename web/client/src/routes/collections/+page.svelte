<script lang="ts">
  import AppShell from "$lib/components/AppShell.svelte";
  import { onMount } from "svelte";
  import { goto } from "$app/navigation";
  import { Database, Shield, Eye, Search } from "@lucide/svelte";
  import { listCollections } from "$lib/api/collections";
  import type { ServicebaseCollection } from "$lib/types/collections";

  let collections = $state<ServicebaseCollection[]>([]);
  let loading = $state(true);
  let error = $state<string | null>(null);
  let query = $state("");

  const filteredCollections = $derived(
    collections.filter((collection) =>
      collection.name.toLowerCase().includes(query.toLowerCase()),
    ),
  );

  function collectionIcon(type: string) {
    switch (type) {
      case "auth":
      case "users":
        return Shield;
      case "view":
        return Eye;
      default:
        return Database;
    }
  }

  function openCollection(name: string) {
    goto(`/collections/${name}`);
  }

  onMount(async () => {
    try {
      collections = await listCollections();
    } catch (err) {
      error = err instanceof Error ? err.message : "Failed to load collections";
    } finally {
      loading = false;
    }
  });
</script>

<AppShell>
  <div class="flex h-full flex-col overflow-hidden bg-background">
    <div class="border-b border-border px-6 py-4">
      <div class="flex items-center justify-between gap-4">
        <div>
          <h1 class="text-2xl font-semibold tracking-tight">Collections</h1>
          <p class="mt-1 text-sm text-muted-foreground">
            Browse collections, records, schema and rules.
          </p>
        </div>
      </div>

      <div class="relative mt-4 max-w-sm">
        <Search class="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
        <input
          bind:value={query}
          placeholder="Filter collections"
          class="h-10 w-full rounded-lg border border-border bg-background pl-10 pr-4 text-sm outline-none transition-colors focus:border-primary"
        />
      </div>
    </div>

    <div class="flex-1 overflow-auto p-6">
      {#if loading}
        <div class="text-sm text-muted-foreground">Loading collections…</div>
      {:else if error}
        <div class="rounded-xl border border-destructive/30 bg-destructive/10 p-4 text-sm text-destructive">
          {error}
        </div>
      {:else}
        <div class="grid gap-3 sm:grid-cols-2 xl:grid-cols-3">
          {#each filteredCollections as collection (collection.id)}
            {@const Icon = collectionIcon(collection.type)}
            <button
              onclick={() => openCollection(collection.name)}
              class="group rounded-2xl border border-border bg-card p-4 text-left transition-all hover:border-primary/40 hover:bg-accent/30"
            >
              <div class="flex items-start justify-between gap-3">
                <div class="flex items-center gap-3">
                  <div class="rounded-xl border border-border bg-background p-2 text-muted-foreground group-hover:text-foreground">
                    <Icon class="h-5 w-5" />
                  </div>
                  <div>
                    <div class="font-medium">{collection.name}</div>
                    <div class="mt-1 text-xs text-muted-foreground">
                      {collection.type}
                    </div>
                  </div>
                </div>

                {#if collection.system}
                  <span class="rounded-md border border-border bg-muted px-2 py-1 text-[10px] uppercase tracking-wide text-muted-foreground">
                    system
                  </span>
                {/if}
              </div>
            </button>
          {/each}
        </div>
      {/if}
    </div>
  </div>
</AppShell>
