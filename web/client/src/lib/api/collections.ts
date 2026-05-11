import { API_BASE } from "./config";
import type { CollectionRecordList, ServicebaseCollection } from "$lib/types/collections";

function authHeaders(): Record<string, string> {
  const headers: Record<string, string> = { "Content-Type": "application/json" };

  if (typeof localStorage === "undefined") return headers;

  const raw = localStorage.getItem("servicebase_auth");
  if (!raw) return headers;

  try {
    const auth = JSON.parse(raw) as { access_token?: string; token?: string };
    const token = auth.access_token ?? auth.token;
    if (token) headers.Authorization = `Bearer ${token}`;
  } catch {
    // Ignore malformed local auth state. The request will fail with 401.
  }

  return headers;
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    ...init,
    headers: {
      ...authHeaders(),
      ...(init?.headers ?? {}),
    },
  });

  if (!res.ok) {
    const detail = await res.text().catch(() => "");
    throw new Error(`${res.status} ${res.statusText}${detail ? `: ${detail}` : ""}`);
  }

  if (res.status === 204) return undefined as T;
  return (await res.json()) as T;
}

export async function listCollections(): Promise<ServicebaseCollection[]> {
  const data = await request<ServicebaseCollection[] | { collections?: ServicebaseCollection[]; items?: ServicebaseCollection[] }>(
    "/collections",
  );

  if (Array.isArray(data)) return data;
  return data.collections ?? data.items ?? [];
}

export async function getCollection(collection: string): Promise<ServicebaseCollection> {
  return request<ServicebaseCollection>(`/collections/${encodeURIComponent(collection)}`);
}

export async function listCollectionRecords(
  collection: string,
  page = 1,
  perPage = 50,
): Promise<CollectionRecordList> {
  const params = new URLSearchParams({
    page: String(page),
    perPage: String(perPage),
  });

  const data = await request<CollectionRecordList | Record<string, unknown>[]>(
    `/collections/${encodeURIComponent(collection)}/records?${params}`,
  );

  if (Array.isArray(data)) {
    return {
      page,
      perPage,
      totalItems: data.length,
      totalPages: 1,
      items: data,
    };
  }

  return data;
}
