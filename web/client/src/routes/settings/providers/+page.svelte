<script lang="ts">
    import { providersStore } from '$lib/stores/providers.svelte';
    import { onMount } from 'svelte';
    import ProviderList from '$lib/components/providers/ProviderList.svelte';

    onMount(async () => {
        await providersStore.loadProviders();
    });
</script>

<svelte:window title="Providers - Settings" />

<div class="providers-settings-page">
    <div class="settings-header">
        <div class="breadcrumb">
            <a href="/settings">Settings</a>
            <span>/</span>
            <span>Providers</span>
        </div>
        <h1>LLM Providers</h1>
        <p class="description">
            Configure your AI model providers. Add OpenAI-compatible endpoints to use with agents.
        </p>
    </div>

    <div class="settings-content">
        <ProviderList />
    </div>
</div>

<style>
    .providers-settings-page {
        max-width: 1200px;
        margin: 0 auto;
        padding: 2rem;
    }

    .settings-header {
        margin-bottom: 2rem;
    }

    .breadcrumb {
        display: flex;
        gap: 0.5rem;
        align-items: center;
        font-size: 0.875rem;
        color: rgb(107 114 128);
        margin-bottom: 1rem;
    }

    .breadcrumb a {
        color: rgb(59 130 246);
        text-decoration: none;
    }

    .breadcrumb a:hover {
        text-decoration: underline;
    }

    h1 {
        font-size: 1.875rem;
        font-weight: 700;
        margin-bottom: 0.5rem;
        color: rgb(17 24 39);
    }

    .description {
        color: rgb(107 114 128);
        font-size: 1rem;
        line-height: 1.6;
        max-width: 600px;
    }

    .settings-content {
        background: white;
        border: 1px solid rgb(229 231 235);
        border-radius: 0.75rem;
        overflow: hidden;
    }
</style>
