<script lang="ts">
    import { agentsStore } from '$lib/stores/agents.svelte';
    import { onMount } from 'svelte';
    import AgentList from '$lib/components/agents/AgentList.svelte';
    import Chat from '$lib/components/agents/Chat.svelte';

    onMount(async () => {
        await agentsStore.loadAgents();
    });
</script>

<svelte:window title="Agents - ServiceBase" />

<div class="agents-page">
    <div class="agents-layout">
        <aside class="agents-sidebar">
            <AgentList />
        </aside>
        <main class="agents-main">
            {#if agentsStore.activeSession}
                <div class="chat-header">
                    <h2>Chat</h2>
                    <span class="session-info">Session: {agentsStore.activeSession.id.slice(0, 8)}</span>
                </div>
                <Chat />
            {:else}
                <div class="no-session">
                    <div class="no-session-content">
                        <h1>Select an agent to start chatting</h1>
                        <p>Choose an agent from the sidebar or create a new one to begin.</p>
                    </div>
                </div>
            {/if}
        </main>
    </div>
</div>

<style>
    .agents-page {
        height: 100vh;
        width: 100%;
        overflow: hidden;
    }

    .agents-layout {
        display: flex;
        height: 100%;
    }

    .agents-sidebar {
        width: 320px;
        border-right: 1px solid rgb(229 231 235);
        overflow-y: auto;
        background: rgb(249 250 251);
    }

    .agents-main {
        flex: 1;
        display: flex;
        flex-direction: column;
        overflow: hidden;
    }

    .chat-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        padding: 1rem 1.5rem;
        border-bottom: 1px solid rgb(229 231 235);
        background: white;
    }

    .chat-header h2 {
        font-size: 1.25rem;
        font-weight: 600;
    }

    .session-info {
        font-size: 0.75rem;
        color: rgb(107 114 128);
        font-family: monospace;
        background: rgb(243 244 246);
        padding: 0.25rem 0.5rem;
        border-radius: 0.25rem;
    }

    .no-session {
        flex: 1;
        display: flex;
        align-items: center;
        justify-content: center;
        background: rgb(249 250 251);
    }

    .no-session-content {
        text-align: center;
        max-width: 400px;
        padding: 2rem;
    }

    .no-session-content h1 {
        font-size: 1.5rem;
        font-weight: 600;
        margin-bottom: 0.5rem;
        color: rgb(17 24 39);
    }

    .no-session-content p {
        color: rgb(107 114 128);
        line-height: 1.6;
    }
</style>
