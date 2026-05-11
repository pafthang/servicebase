import { client } from '$lib/api';

export interface ProviderModel {
    id: string;
    name: string;
    base_url: string;
    api_key?: string;
    model: string;
    enabled: boolean;
    timeout?: number;
    max_retries?: number;
    created?: string;
    updated?: string;
}

export interface ModelInfo {
    id: string;
    object: string;
    created: number;
    owned_by: string;
}

class ProvidersStore {
    providers = $state<ProviderModel[]>([]);
    models = $state<ModelInfo[]>([]);
    isLoading = $state(false);
    isTesting = $state<string | null>(null);
    error = $state<string | null>(null);

    async loadProviders(enabled?: boolean) {
        this.isLoading = true;
        this.error = null;
        try {
            const result = await client.listProviders(enabled);
            this.providers = result.providers;
            return result.providers;
        } catch (err: any) {
            this.error = err.message || 'Failed to load providers';
            console.error('Failed to load providers:', err);
            return [];
        } finally {
            this.isLoading = false;
        }
    }

    async createProvider(data: {
        name: string;
        base_url: string;
        api_key?: string;
        model: string;
        enabled?: boolean;
        timeout?: number;
        max_retries?: number;
    }) {
        this.isLoading = true;
        this.error = null;
        try {
            const provider = await client.createProvider(data);
            this.providers.push(provider);
            return provider;
        } catch (err: any) {
            this.error = err.message || 'Failed to create provider';
            throw err;
        } finally {
            this.isLoading = false;
        }
    }

    async updateProvider(id: string, data: Partial<ProviderModel>) {
        this.error = null;
        try {
            const provider = await client.updateProvider(id, data);
            const index = this.providers.findIndex(p => p.id === id);
            if (index !== -1) {
                this.providers[index] = provider;
            }
            return provider;
        } catch (err: any) {
            this.error = err.message || 'Failed to update provider';
            throw err;
        }
    }

    async deleteProvider(id: string) {
        this.error = null;
        try {
            await client.deleteProvider(id);
            this.providers = this.providers.filter(p => p.id !== id);
        } catch (err: any) {
            this.error = err.message || 'Failed to delete provider';
            throw err;
        }
    }

    async toggleProvider(id: string, enabled: boolean) {
        return this.updateProvider(id, { enabled });
    }

    async loadModels(providerId: string) {
        this.isLoading = true;
        this.error = null;
        try {
            const result = await client.listProviderModels(providerId);
            this.models = result.models;
            return result.models;
        } catch (err: any) {
            this.error = err.message || 'Failed to load models';
            console.error('Failed to load models:', err);
            return [];
        } finally {
            this.isLoading = false;
        }
    }

    async testProvider(providerId: string, message: string = 'Hello!'): Promise<boolean> {
        this.isTesting = providerId;
        this.error = null;
        try {
            const response = await client.providerChat(providerId, {
                messages: [
                    { role: 'system', content: 'You are a helpful assistant.' },
                    { role: 'user', content: message },
                ],
                max_tokens: 50,
            });
            
            return !!response.choices?.[0]?.message?.content;
        } catch (err: any) {
            this.error = err.message || 'Provider test failed';
            return false;
        } finally {
            this.isTesting = null;
        }
    }

    async testProviderStream(providerId: string, callbacks: {
        onChunk?: (chunk: string) => void;
        onComplete?: () => void;
        onError?: (error: string) => void;
    }) {
        this.isTesting = providerId;
        this.error = null;
        
        try {
            await client.providerChatStream(
                providerId,
                {
                    messages: [
                        { role: 'system', content: 'You are a helpful assistant.' },
                        { role: 'user', content: 'Say hello!' },
                    ],
                    max_tokens: 50,
                },
                {
                    onChunk: (chunk: any) => {
                        const delta = chunk.choices?.[0]?.delta?.content || '';
                        if (delta) {
                            callbacks.onChunk?.(delta);
                        }
                    },
                    onError: (error: any) => {
                        this.error = error.message || 'Stream test failed';
                        callbacks.onError?.(this.error);
                    },
                    onEnd: () => {
                        this.isTesting = null;
                        callbacks.onComplete?.();
                    },
                }
            );
        } catch (err: any) {
            this.isTesting = null;
            this.error = err.message || 'Stream test failed';
            callbacks.onError?.(this.error);
        }
    }

    getEnabledProviders(): ProviderModel[] {
        return this.providers.filter(p => p.enabled);
    }

    getProviderById(id: string): ProviderModel | undefined {
        return this.providers.find(p => p.id === id);
    }

    clearError() {
        this.error = null;
    }
}

export const providersStore = new ProvidersStore();
