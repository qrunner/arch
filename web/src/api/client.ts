const API_BASE = '/api/v1';

export interface Asset {
  id: string;
  external_id: string;
  source: string;
  asset_type: string;
  name: string;
  fqdn?: string;
  ip_addresses: string[];
  attributes: Record<string, unknown>;
  first_seen: string;
  last_seen: string;
  status: 'active' | 'stale' | 'removed';
  created_at: string;
  updated_at: string;
}

export interface ChangeEvent {
  id: string;
  asset_id: string;
  action: string;
  source: string;
  diff: Record<string, unknown> | null;
  timestamp: string;
}

export interface Relationship {
  id: string;
  from_id: string;
  to_id: string;
  type: string;
  source: string;
  properties: Record<string, unknown>;
}

export interface DashboardStats {
  total_assets: number;
  by_source: Record<string, number>;
  by_type: Record<string, number>;
  by_status: Record<string, number>;
  recent_changes: number;
}

export interface ApiResponse<T> {
  data: T;
  error?: string;
  total?: number;
}

export interface AssetFilter {
  source?: string;
  asset_type?: string;
  status?: string;
  search?: string;
  limit?: number;
  offset?: number;
}

async function fetchApi<T>(path: string, init?: RequestInit): Promise<ApiResponse<T>> {
  const res = await fetch(`${API_BASE}${path}`, {
    headers: { 'Content-Type': 'application/json' },
    ...init,
  });
  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new Error(body.error || `API error: ${res.status}`);
  }
  return res.json();
}

export async function listAssets(filter: AssetFilter = {}): Promise<ApiResponse<Asset[]>> {
  const params = new URLSearchParams();
  if (filter.source) params.set('source', filter.source);
  if (filter.asset_type) params.set('asset_type', filter.asset_type);
  if (filter.status) params.set('status', filter.status);
  if (filter.search) params.set('search', filter.search);
  if (filter.limit) params.set('limit', String(filter.limit));
  if (filter.offset) params.set('offset', String(filter.offset));
  return fetchApi<Asset[]>(`/assets?${params}`);
}

export async function getAsset(id: string): Promise<ApiResponse<Asset>> {
  return fetchApi<Asset>(`/assets/${id}`);
}

export async function getAssetHistory(
  id: string,
  limit = 50,
  offset = 0,
): Promise<ApiResponse<ChangeEvent[]>> {
  return fetchApi<ChangeEvent[]>(`/assets/${id}/history?limit=${limit}&offset=${offset}`);
}

export async function getAssetRelationships(id: string): Promise<ApiResponse<Relationship[]>> {
  return fetchApi<Relationship[]>(`/assets/${id}/relationships`);
}

export async function getDependencyGraph(
  id: string,
  depth = 3,
): Promise<ApiResponse<{ assets: Asset[]; relationships: Relationship[] }>> {
  return fetchApi(`/graph/dependencies/${id}?depth=${depth}`);
}

export async function getImpactGraph(
  id: string,
  depth = 3,
): Promise<ApiResponse<{ assets: Asset[]; relationships: Relationship[] }>> {
  return fetchApi(`/graph/impact/${id}?depth=${depth}`);
}

export async function getDashboardStats(): Promise<ApiResponse<DashboardStats>> {
  return fetchApi<DashboardStats>('/dashboard/stats');
}

export async function getRecentChanges(
  limit = 50,
  offset = 0,
): Promise<ApiResponse<ChangeEvent[]>> {
  return fetchApi<ChangeEvent[]>(`/changes?limit=${limit}&offset=${offset}`);
}
