export type CollectionType = "base" | "auth" | "view" | "users" | string;

export interface CollectionSchemaField {
  id?: string;
  name: string;
  type: string;
  required?: boolean;
  hidden?: boolean;
  presentable?: boolean;
  unique?: boolean;
  options?: Record<string, unknown> | null;
  [key: string]: unknown;
}

export interface ServicebaseCollection {
  id: string;
  name: string;
  type: CollectionType;
  system?: boolean;
  schema?: CollectionSchemaField[] | { fields?: CollectionSchemaField[] };
  indexes?: string[];
  listRule?: string | null;
  viewRule?: string | null;
  createRule?: string | null;
  updateRule?: string | null;
  deleteRule?: string | null;
  options?: Record<string, unknown> | null;
  created?: string;
  updated?: string;
  recordsCount?: number;
  [key: string]: unknown;
}

export interface CollectionRecordList {
  page: number;
  perPage: number;
  totalItems: number;
  totalPages: number;
  items: Record<string, unknown>[];
}

export function getCollectionFields(collection: ServicebaseCollection): CollectionSchemaField[] {
  if (Array.isArray(collection.schema)) return collection.schema;
  if (collection.schema && Array.isArray(collection.schema.fields)) return collection.schema.fields;
  return [];
}
