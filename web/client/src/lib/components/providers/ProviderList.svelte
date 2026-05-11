<script lang="ts">
    import { providersStore, type ProviderModel } from '$lib/stores/providers.svelte';
    import { Plus, Settings, Trash2, Zap, Check, X } from '@lucide/svelte';

    let showCreateForm = $state(false);
    let testResult = $state<{ [key: string]: boolean }>({});
    let newProvider = $state({
        name: '',
        base_url: '',
        api_key: '',
        model: 'gpt-4o',
        enabled: true,
        timeout: 60,
        max_retries: 3,
    });

    async function handleCreate() {
        try {
            await providersStore.createProvider(newProvider);
            showCreateForm = false;
            newProvider = {
                name: '',
                base_url: '',
                api_key: '',
                model: 'gpt-4o',
                enabled: true,
                timeout: 60,
                max_retries: 3,
            };
        } catch (err) {}
    }

    async function handleDelete(provider: ProviderModel) {
        if (confirm(`Delete provider "${provider.name}"?`)) {
            await providersStore.deleteProvider(provider.id);
        }
    }

    async function handleTest(provider: ProviderModel) {
        const success = await providersStore.testProvider(provider.id);
        testResult[provider.id] = success;
        setTimeout(() => { delete testResult[provider.id]; }, 3000);
    }

    async function handleToggle(provider: ProviderModel) {
        await providersStore.toggleProvider(provider.id, !provider.enabled);
    }
</script>

<div class="provider-list">
    <div class="header">
        <h2>LLM Providers</h2>
        <button onclick={() => showCreateForm = !showCreateForm}>
            <Plus size={18} /> New Provider
        </button>
    </div>

    {#if showCreateForm}
        <div class="create-form">
            <div class="form-group">
                <label>Name</label>
                <input type="text" bind:value={newProvider.name} placeholder="OpenAI" />
            </div>
            <div class="form-group">
                <label>Base URL</label>
                <input type="text" bind:value={newProvider.base_url} placeholder="https://api.openai.com/v1" />
            </div>
            <div class="form-group">
                <label>API Key</label>
                <input type="password" bind:value={newProvider.api_key} placeholder="sk-..." />
            </div>
            <div class="form-group">
                <label>Default Model</label>
                <input type="text" bind:value={newProvider.model} placeholder="gpt-4o" />
            </div>
            <div class="form-actions">
                <button type="button" onclick={() => showCreateForm = false}>Cancel</button>
                <button type="submit" onclick={handleCreate}>Create</button>
            </div>
        </div>
    {/if}

    {#if providersStore.providers.length === 0}
        <div class="empty-state"><p>No providers configured.</p></div>
    {:else}
        <div class="providers-grid">
            {#each providersStore.providers as provider (provider.id)}
                <div class="provider-card {provider.enabled ? '' : 'disabled'}">
                    <div class="provider-header">
                        <h3>{provider.name}</h3>
                        <button onclick={() => handleToggle(provider)} class="toggle-btn">
                            {#if provider.enabled}<Check size={16} />{:else}<X size={16} />{/if}
                        </button>
                    </div>
                    <p class="base-url">{provider.base_url}</p>
                    <div class="provider-meta">
                        <span>{provider.model}</span>
                        <span>{provider.timeout}s timeout</span>
                    </div>
                    {#if testResult[provider.id] !== undefined}
                        <div class="test-result {testResult[provider.id] ? 'success' : 'error'}">
                            {#if testResult[provider.id]}<Check size={14} /> Passed{:else}<X size={14} /> Failed{/if}
                        </div>
                    {/if}
                    <div class="provider-actions">
                        <button onclick={() => handleTest(provider)}>Test</button>
                        <button onclick={() => {}}>Settings</button>
                        <button onclick={() => handleDelete(provider)} class="danger">Delete</button>
                    </div>
                </div>
            {/each}
        </div>
    {/if}
</div>

<style>
    .provider-list { padding: 1rem; }
    .header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 1.5rem; }
    .header h2 { font-size: 1.5rem; font-weight: 600; }
    .header button { display: flex; align-items: center; gap: 0.5rem; background: rgb(59 130 246); color: white; border: none; padding: 0.5rem 1rem; border-radius: 0.5rem; cursor: pointer; }
    .create-form { background: rgb(249 250 251); padding: 1.5rem; border-radius: 0.75rem; margin-bottom: 1.5rem; max-width: 500px; }
    .form-group { margin-bottom: 1rem; }
    .form-group label { display: block; font-weight: 500; margin-bottom: 0.5rem; }
    .form-group input { width: 100%; padding: 0.5rem 0.75rem; border: 1px solid rgb(209 213 219); border-radius: 0.5rem; }
    .form-actions { display: flex; gap: 0.5rem; justify-content: flex-end; }
    .form-actions button { padding: 0.5rem 1rem; border-radius: 0.5rem; border: none; cursor: pointer; }
    .form-actions button[type="button"] { background: rgb(229 231 235); }
    .form-actions button[type="submit"] { background: rgb(59 130 246); color: white; }
    .empty-state { text-align: center; padding: 3rem; color: rgb(107 114 128); }
    .providers-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(300px, 1fr)); gap: 1rem; }
    .provider-card { background: white; border: 1px solid rgb(229 231 235); border-radius: 0.75rem; padding: 1rem; }
    .provider-card.disabled { opacity: 0.6; }
    .provider-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 0.5rem; }
    .base-url { color: rgb(107 114 128); font-size: 0.875rem; font-family: monospace; }
    .provider-meta { display: flex; gap: 0.75rem; margin: 0.75rem 0; }
    .provider-meta span { font-size: 0.75rem; padding: 0.25rem 0.5rem; background: rgb(243 244 246); border-radius: 0.25rem; }
    .test-result { display: flex; align-items: center; gap: 0.25rem; padding: 0.5rem; border-radius: 0.375rem; margin-bottom: 1rem; font-size: 0.875rem; }
    .test-result.success { background: rgb(220 252 231); color: rgb(22 163 74); }
    .test-result.error { background: rgb(254 226 226); color: rgb(220 38 38); }
    .provider-actions { display: flex; gap: 0.5rem; }
    .provider-actions button { padding: 0.375rem 0.75rem; border: 1px solid rgb(229 231 235); background: white; border-radius: 0.375rem; font-size: 0.75rem; cursor: pointer; }
    .provider-actions button.danger { color: rgb(220 38 38); border-color: rgb(254 226 226); }
</style>
