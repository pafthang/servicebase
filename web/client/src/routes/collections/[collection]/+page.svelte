<script lang="ts">
  import AppShell from "$lib/components/AppShell.svelte";
  import { onMount } from "svelte";
  import { page } from "$app/state";
  import { getCollection, listCollectionRecords } from "$lib/api/collections";
  import { getCollectionFields, type ServicebaseCollection } from "$lib/types/collections";

  let collection = $state<ServicebaseCollection | null>(null);
  let records = $state<Record<string, unknown>[]>([]);
  let loading = $state(true);
  let error = $state<string | null>(null);

  let collectionName = $derived(page.params.collection);
  let fields = $derived(collection ? getCollectionFields(collection) : []);

  onMount(async () => {
    try {
      collection = await getCollection(collectionName);
      const result = await listCollectionRecords(collectionName);
      records = result.items;
    } catch (err) {
      error = err instanceof Error ? err.message : "Failed to load collection";
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
          <div class="text-xs uppercase tracking-wide text-muted-foreground">
            Collection
          </div>
          <h1 class="mt-1 text-2xl font-semibold tracking-tight">
            {collectionName}
          </h1>
        </div>

        {#if collection}
          <div class="rounded-lg border border-border bg-card px-3 py-2 text-sm text-muted-foreground">
            {collection.type}
          </div>
        {/if}
      </div>
    </div>

    <div class="flex-1 overflow-auto p-6">
      {#if loading}
        <div class="text-sm text-muted-foreground">Loading collection…</div>
      {:else if error}
        <div class="rounded-xl border border-destructive/30 bg-destructive/10 p-4 text-sm text-destructive">
          {error}
        </div>
      {:else if collection}
        <div class="grid gap-6 xl:grid-cols-[320px,1fr]">
          <section class="rounded-2xl border border-border bg-card p-4">
            <div class="mb-4">
              <h2 class="text-sm font-medium">Schema</h2>
              <p class="mt-1 text-xs text-muted-foreground">
                Fields available in this collection.
              </p>
            </div>

            <div class="space-y-2">
              {#each fields as field}
                <div class="rounded-xl border border-border bg-background px-3 py-2">
                  <div class="flex items-center justify-between gap-2">
                    <div class="font-medium">{field.name}</div>
                    <div class="text-xs text-muted-foreground">
                      {field.type}
                    </div>
                  </div>
                </div>
              {/each}
            </div>
          </section>

          <section class="rounded-2xl border border-border bg-card p-4 overflow-hidden">
            <div class="mb-4 flex items-center justify-between gap-4">
              <div>
                <h2 class="text-sm font-medium">Records</h2>
                <p class="mt-1 text-xs text-muted-foreground">
                  Live collection records browser.
                </p>
              </div>

              <div class="text-xs text-muted-foreground">
                {records.length} records
              </div>
            </div>

            <div class="overflow-auto rounded-xl border border-border">
              <table class="min-w-full divide-y divide-border text-sm">
                <thead class="bg-muted/40">
                  <tr>
                    {#each fields.slice(0, 6) as field}
                      <th class="px-3 py-2 text-left font-medium text-muted-foreground">
                        {field.name}
                      </th>
                    {/each}
                  </tr>
                </thead>

                <tbody class="divide-y divide-border">
                  {#each records as record}
                    <tr class="hover:bg-accent/20">
                      {#each fields.slice(0, 6) as field}
                        <td class="max-w-[220px] truncate px-3 py-2">
                          {String(record[field.name] ?? "")}
                        </td>
                      {/each}
                    </tr>
                  {/each}
                </tbody>
              </table>
            </div>
          </section>
        </div>
      {/if}
    </div>
  </div>
</AppShell>
