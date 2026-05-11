<script lang="ts">
    import { agentsStore, type AgentModel } from '$lib/stores/agents.svelte';
    import { Plus, Settings, Trash2, MessageSquare } from '@lucide/svelte';

    let showCreateForm = $state(false);
    let newAgent = $state({
        name: '',
        role: '',
        description: '',
        backend: 'openai',
        model: 'gpt-4o',
        system_prompt: '',
        enabled: true,
    });

    async function handleCreate() {
        try {
            await agentsStore.createAgent(newAgent);
            showCreateForm = false;
            newAgent = {
                name: '',
                role: '',
                description: '',
                backend: 'openai',
                model: 'gpt-4o',
                system_prompt: '',
                enabled: true,
            };
        } catch (err) {
            // Error handled by store
        }
    }

    async function handleDelete(agent: AgentModel) {
        if (confirm(`Delete agent "${agent.name}"?`)) {
            await agentsStore.deleteAgent(agent.id);
        }
    }

    async function handleCreateSession(agent: AgentModel) {
        await agentsStore.createSession(agent.id, 'current-user');
    }
</script>

<div class="agent-list">
    <div class="header">
        <h2>Agents</h2>
        <button onclick={() => showCreateForm = !showCreateForm}>
            <Plus size={18} />
            New Agent
        </button>
    </div>

    {#if showCreateForm}
        <div class="create-form">
            <div class="form-group">
                <label>Name</label>
                <input type="text" bind:value={newAgent.name} placeholder="Assistant" />
            </div>
            <div class="form-group">
                <label>Role</label>
                <input type="text" bind:value={newAgent.role} placeholder="Helpful assistant" />
            </div>
            <div class="form-group">
                <label>Backend</label>
                <select bind:value={newAgent.backend}>
                    <option value="openai">OpenAI</option>
                    <option value="groq">Groq</option>
                    <option value="local">Local</option>
                </select>
            </div>
            <div class="form-group">
                <label>Model</label>
                <input type="text" bind:value={newAgent.model} placeholder="gpt-4o" />
            </div>
            <div class="form-group">
                <label>System Prompt</label>
                <textarea bind:value={newAgent.system_prompt} rows="3" placeholder="You are a helpful assistant..." />
            </div>
            <div class="form-group">
                <label>
                    <input type="checkbox" bind:checked={newAgent.enabled} />
                    Enabled
                </label>
            </div>
            <div class="form-actions">
                <button type="button" onclick={() => showCreateForm = false}>Cancel</button>
                <button type="submit" onclick={handleCreate}>Create</button>
            </div>
        </div>
    {/if}

    {#if agentsStore.isLoading && agentsStore.agents.length === 0}
        <div class="loading">Loading agents...</div>
    {:else if agentsStore.agents.length === 0}
        <div class="empty-state">
            <p>No agents yet. Create your first agent to get started.</p>
        </div>
    {:else}
        <div class="agents-grid">
            {#each agentsStore.agents as agent (agent.id)}
                <div class="agent-card {agent.enabled ? '' : 'disabled'}">
                    <div class="agent-header">
                        <h3>{agent.name}</h3>
                        <span class="status">{agent.enabled ? 'Active' : 'Inactive'}</span>
                    </div>
                    <p class="role">{agent.role}</p>
                    {#if agent.description}
                        <p class="description">{agent.description}</p>
                    {/if}
                    <div class="agent-meta">
                        <span class="backend">{agent.backend}</span>
                        <span class="model">{agent.model}</span>
                    </div>
                    <div class="agent-actions">
                        <button onclick={() => handleCreateSession(agent)} title="New chat">
                            <MessageSquare size={16} />
                            Chat
                        </button>
                        <button onclick={() => {}} title="Settings">
                            <Settings size={16} />
                        </button>
                        <button onclick={() => handleDelete(agent)} title="Delete" class="danger">
                            <Trash2 size={16} />
                        </button>
                    </div>
                </div>
            {/each}
        </div>
    {/if}
</div>

<style>
    .agent-list {
        padding: 1rem;
    }

    .header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-bottom: 1.5rem;
    }

    .header h2 {
        font-size: 1.5rem;
        font-weight: 600;
    }

    .header button {
        display: flex;
        align-items: center;
        gap: 0.5rem;
        background: rgb(59 130 246);
        color: white;
        border: none;
        padding: 0.5rem 1rem;
        border-radius: 0.5rem;
        cursor: pointer;
        font-weight: 500;
    }

    .header button:hover {
        background: rgb(37 99 235);
    }

    .create-form {
        background: rgb(249 250 251);
        padding: 1.5rem;
        border-radius: 0.75rem;
        margin-bottom: 1.5rem;
    }

    .form-group {
        margin-bottom: 1rem;
    }

    .form-group label {
        display: block;
        font-weight: 500;
        margin-bottom: 0.5rem;
        font-size: 0.875rem;
    }

    .form-group input[type="text"],
    .form-group select,
    .form-group textarea {
        width: 100%;
        padding: 0.5rem 0.75rem;
        border: 1px solid rgb(209 213 219);
        border-radius: 0.5rem;
        font-family: inherit;
        font-size: 0.875rem;
    }

    .form-group input[type="checkbox"] {
        margin-right: 0.5rem;
    }

    .form-actions {
        display: flex;
        gap: 0.5rem;
        justify-content: flex-end;
    }

    .form-actions button {
        padding: 0.5rem 1rem;
        border-radius: 0.5rem;
        border: none;
        cursor: pointer;
        font-weight: 500;
    }

    .form-actions button[type="button"] {
        background: rgb(229 231 235);
        color: rgb(107 114 128);
    }

    .form-actions button[type="submit"] {
        background: rgb(59 130 246);
        color: white;
    }

    .loading, .empty-state {
        text-align: center;
        padding: 3rem;
        color: rgb(107 114 128);
    }

    .agents-grid {
        display: grid;
        grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
        gap: 1rem;
    }

    .agent-card {
        background: white;
        border: 1px solid rgb(229 231 235);
        border-radius: 0.75rem;
        padding: 1rem;
    }

    .agent-card.disabled {
        opacity: 0.6;
    }

    .agent-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-bottom: 0.5rem;
    }

    .agent-header h3 {
        font-size: 1.125rem;
        font-weight: 600;
    }

    .status {
        font-size: 0.75rem;
        padding: 0.25rem 0.5rem;
        background: rgb(34 197 94);
        color: white;
        border-radius: 9999px;
    }

    .role {
        color: rgb(107 114 128);
        font-size: 0.875rem;
        margin-bottom: 0.5rem;
    }

    .description {
        font-size: 0.875rem;
        margin-bottom: 0.75rem;
    }

    .agent-meta {
        display: flex;
        gap: 0.5rem;
        margin-bottom: 1rem;
    }

    .agent-meta span {
        font-size: 0.75rem;
        padding: 0.25rem 0.5rem;
        background: rgb(243 244 246);
        border-radius: 0.25rem;
        color: rgb(107 114 128);
    }

    .agent-actions {
        display: flex;
        gap: 0.5rem;
    }

    .agent-actions button {
        display: flex;
        align-items: center;
        gap: 0.25rem;
        padding: 0.375rem 0.75rem;
        border: 1px solid rgb(229 231 235);
        background: white;
        border-radius: 0.375rem;
        font-size: 0.75rem;
        cursor: pointer;
    }

    .agent-actions button:hover {
        background: rgb(249 250 251);
    }

    .agent-actions button.danger {
        color: rgb(220 38 38);
        border-color: rgb(254 226 226);
    }

    .agent-actions button.danger:hover {
        background: rgb(254 226 226);
    }
</style>
