<script lang="ts">
    import { agentsStore, type ChatMessage } from '$lib/stores/agents.svelte';
    import { Send } from '@lucide/svelte';

    let message = $state('');
    let messagesEnd: HTMLDivElement;

    function scrollToBottom() {
        messagesEnd?.scrollIntoView({ behavior: 'smooth' });
    }

    $effect(() => {
        scrollToBottom();
    });

    async function sendMessage() {
        if (!message.trim() || agentsStore.isStreaming) return;
        
        const content = message;
        message = '';
        
        await agentsStore.sendMessage(content, {
            onChunk: () => {
                scrollToBottom();
            },
        });
    }

    function handleKeydown(e: KeyboardEvent) {
        if (e.key === 'Enter' && !e.shiftKey) {
            e.preventDefault();
            sendMessage();
        }
    }

    function stopGenerating() {
        agentsStore.stopStreaming();
    }
</script>

<div class="chat-container">
    {#if agentsStore.error}
        <div class="error-banner">
            {agentsStore.error}
            <button onclick={() => agentsStore.clearError()}>Dismiss</button>
        </div>
    {/if}

    <div class="messages">
        {#if agentsStore.messages.length === 0}
            <div class="empty-state">
                <h2>Start a conversation</h2>
                <p>Send a message to begin chatting with the agent.</p>
            </div>
        {:else}
            {#each agentsStore.messages as msg (msg.timestamp)}
                <div class="message {msg.role}">
                    <div class="message-avatar">
                        {#if msg.role === 'user'}
                            <span>👤</span>
                        {:else}
                            <span>🤖</span>
                        {/if}
                    </div>
                    <div class="message-content">
                        {msg.content}
                    </div>
                </div>
            {/each}
        {/if}
        <div bind:this={messagesEnd}></div>
    </div>

    <div class="input-area">
        {#if agentsStore.isStreaming}
            <button class="stop-btn" onclick={stopGenerating}>
                <span>⏹</span> Stop generating
            </button>
        {:else}
            <form on:submit|preventDefault={sendMessage}>
                <textarea
                    bind:value={message}
                    on:keydown={handleKeydown}
                    placeholder="Type a message..."
                    rows="1"
                    disabled={!agentsStore.activeSession}
                ></textarea>
                <button 
                    type="submit" 
                    disabled={!message.trim() || !agentsStore.activeSession || agentsStore.isLoading}
                >
                    <Send size={20} />
                </button>
            </form>
        {/if}
    </div>
</div>

<style>
    .chat-container {
        display: flex;
        flex-direction: column;
        height: 100%;
        max-height: 600px;
    }

    .error-banner {
        background: rgb(254 226 226);
        color: rgb(185 28 28);
        padding: 0.75rem 1rem;
        display: flex;
        justify-content: space-between;
        align-items: center;
    }

    .error-banner button {
        background: transparent;
        border: none;
        color: inherit;
        cursor: pointer;
        font-weight: 600;
    }

    .messages {
        flex: 1;
        overflow-y: auto;
        padding: 1rem;
        display: flex;
        flex-direction: column;
        gap: 1rem;
    }

    .empty-state {
        text-align: center;
        padding: 3rem;
        color: rgb(107 114 128);
    }

    .empty-state h2 {
        font-size: 1.5rem;
        margin-bottom: 0.5rem;
    }

    .message {
        display: flex;
        gap: 0.75rem;
        max-width: 80%;
    }

    .message.user {
        align-self: flex-end;
        flex-direction: row-reverse;
    }

    .message-avatar {
        font-size: 1.5rem;
        flex-shrink: 0;
    }

    .message-content {
        background: rgb(243 244 246);
        padding: 0.75rem 1rem;
        border-radius: 0.75rem;
        line-height: 1.5;
    }

    .message.user .message-content {
        background: rgb(59 130 246);
        color: white;
    }

    .input-area {
        border-top: 1px solid rgb(229 231 235);
        padding: 1rem;
    }

    form {
        display: flex;
        gap: 0.5rem;
        align-items: flex-end;
    }

    textarea {
        flex: 1;
        resize: none;
        padding: 0.75rem;
        border: 1px solid rgb(209 213 219);
        border-radius: 0.5rem;
        font-family: inherit;
        font-size: 0.875rem;
        line-height: 1.5;
    }

    textarea:focus {
        outline: none;
        border-color: rgb(59 130 246);
        box-shadow: 0 0 0 3px rgb(59 130 246 / 0.1);
    }

    button {
        background: rgb(59 130 246);
        color: white;
        border: none;
        padding: 0.75rem;
        border-radius: 0.5rem;
        cursor: pointer;
        display: flex;
        align-items: center;
        justify-content: center;
    }

    button:disabled {
        opacity: 0.5;
        cursor: not-allowed;
    }

    button:hover:not(:disabled) {
        background: rgb(37 99 235);
    }

    .stop-btn {
        width: 100%;
        background: rgb(239 68 68);
        color: white;
        border: none;
        padding: 0.75rem 1.5rem;
        border-radius: 0.5rem;
        cursor: pointer;
        display: flex;
        align-items: center;
        justify-content: center;
        gap: 0.5rem;
        font-weight: 500;
    }

    .stop-btn:hover {
        background: rgb(220 38 38);
    }
</style>
