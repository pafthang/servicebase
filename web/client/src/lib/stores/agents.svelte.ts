import { client } from '$lib/api';

export interface AgentModel {
    id: string;
    name: string;
    role: string;
    description?: string;
    backend?: string;
    model?: string;
    system_prompt?: string;
    tools?: string[];
    enabled: boolean;
    created?: string;
    updated?: string;
}

export interface SessionModel {
    id: string;
    agent_id: string;
    user_id: string;
    status: string;
    messages?: any[];
    created?: string;
    updated?: string;
}

export interface ChatMessage {
    role: string;
    content: string;
    timestamp?: Date;
}

class AgentsStore {
    agents = $state<AgentModel[]>([]);
    sessions = $state<SessionModel[]>([]);
    activeSession = $state<SessionModel | null>(null);
    messages = $state<ChatMessage[]>([]);
    isLoading = $state(false);
    isStreaming = $state(false);
    error = $state<string | null>(null);
    streamController = $state<AbortController | null>(null);

    async loadAgents() {
        this.isLoading = true;
        this.error = null;
        try {
            const result = await client.listAgents();
            this.agents = result.agents;
            return result.agents;
        } catch (err: any) {
            this.error = err.message || 'Failed to load agents';
            console.error('Failed to load agents:', err);
            return [];
        } finally {
            this.isLoading = false;
        }
    }

    async createAgent(data: {
        name: string;
        role: string;
        description?: string;
        backend?: string;
        model?: string;
        system_prompt?: string;
        tools?: string[];
        enabled?: boolean;
    }) {
        this.isLoading = true;
        this.error = null;
        try {
            const agent = await client.createAgent(data);
            this.agents.push(agent);
            return agent;
        } catch (err: any) {
            this.error = err.message || 'Failed to create agent';
            throw err;
        } finally {
            this.isLoading = false;
        }
    }

    async updateAgent(id: string, data: Partial<AgentModel>) {
        this.error = null;
        try {
            const agent = await client.updateAgent(id, data);
            const index = this.agents.findIndex(a => a.id === id);
            if (index !== -1) {
                this.agents[index] = agent;
            }
            return agent;
        } catch (err: any) {
            this.error = err.message || 'Failed to update agent';
            throw err;
        }
    }

    async deleteAgent(id: string) {
        this.error = null;
        try {
            await client.deleteAgent(id);
            this.agents = this.agents.filter(a => a.id !== id);
            this.sessions = this.sessions.filter(s => s.agent_id !== id);
        } catch (err: any) {
            this.error = err.message || 'Failed to delete agent';
            throw err;
        }
    }

    async loadSessions(agentId: string) {
        this.isLoading = true;
        this.error = null;
        try {
            const result = await client.listAgentSessions(agentId);
            this.sessions = result.sessions;
            return result.sessions;
        } catch (err: any) {
            this.error = err.message || 'Failed to load sessions';
            console.error('Failed to load sessions:', err);
            return [];
        } finally {
            this.isLoading = false;
        }
    }

    async createSession(agentId: string, userId?: string) {
        this.error = null;
        try {
            const session = await client.createAgentSession(agentId, userId);
            this.sessions.push(session);
            this.activeSession = session;
            this.messages = [];
            return session;
        } catch (err: any) {
            this.error = err.message || 'Failed to create session';
            throw err;
        }
    }

    async setActiveSession(session: SessionModel | null) {
        this.activeSession = session;
        if (session) {
            await this.loadSessionHistory(session.id);
        } else {
            this.messages = [];
        }
    }

    async loadSessionHistory(sessionId: string) {
        this.isLoading = true;
        try {
            const session = await client.getAgentSession(sessionId);
            if (session.messages && Array.isArray(session.messages)) {
                this.messages = session.messages.map((m: any) => ({
                    role: m.role,
                    content: m.content,
                    timestamp: new Date(),
                }));
            }
            return session;
        } catch (err: any) {
            console.error('Failed to load session history:', err);
            return null;
        } finally {
            this.isLoading = false;
        }
    }

    async sendMessage(content: string, callbacks?: {
        onChunk?: (chunk: string) => void;
        onComplete?: () => void;
        onError?: (error: string) => void;
    }) {
        if (!this.activeSession) {
            throw new Error('No active session');
        }

        this.isStreaming = true;
        this.error = null;

        const userMessage: ChatMessage = {
            role: 'user',
            content,
            timestamp: new Date(),
        };
        this.messages.push(userMessage);

        const assistantMessage: ChatMessage = {
            role: 'assistant',
            content: '',
            timestamp: new Date(),
        };
        this.messages.push(assistantMessage);

        try {
            await client.streamAgent(
                this.activeSession.id,
                content,
                {
                    onChunk: (chunk: any) => {
                        const delta = chunk.choices?.[0]?.delta?.content || '';
                        if (delta) {
                            assistantMessage.content += delta;
                            callbacks?.onChunk?.(delta);
                        }
                    },
                    onToolStart: (data: any) => {
                        console.log('Tool started:', data);
                    },
                    onToolResult: (data: any) => {
                        console.log('Tool result:', data);
                    },
                    onError: (error: any) => {
                        this.error = error.message || 'Stream error';
                        callbacks?.onError?.(this.error);
                    },
                    onEnd: () => {
                        this.isStreaming = false;
                        callbacks?.onComplete?.();
                    },
                }
            );
        } catch (err: any) {
            this.isStreaming = false;
            this.error = err.message || 'Failed to send message';
            callbacks?.onError?.(this.error);
            throw err;
        }
    }

    stopStreaming() {
        if (this.streamController) {
            this.streamController.abort();
            this.streamController = null;
        }
        this.isStreaming = false;
    }

    async abortSession() {
        if (!this.activeSession) return;
        
        try {
            await client.abortAgentSession(this.activeSession.id);
            this.isStreaming = false;
        } catch (err: any) {
            console.error('Failed to abort session:', err);
        }
    }

    clearMessages() {
        this.messages = [];
    }

    clearError() {
        this.error = null;
    }
}

export const agentsStore = new AgentsStore();
