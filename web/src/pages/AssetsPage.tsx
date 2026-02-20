import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Link } from 'react-router-dom';
import { listAssets, type AssetFilter } from '../api/client';

export default function AssetsPage() {
  const [filter, setFilter] = useState<AssetFilter>({ limit: 50, offset: 0 });
  const [search, setSearch] = useState('');

  const query = useQuery({
    queryKey: ['assets', filter],
    queryFn: () => listAssets(filter),
  });

  const handleSearch = () => {
    setFilter((f) => ({ ...f, search, offset: 0 }));
  };

  const assets = query.data?.data ?? [];
  const total = query.data?.total ?? 0;

  return (
    <div>
      <h1 className="text-2xl font-bold mb-6">Assets</h1>

      {/* Filters */}
      <div className="bg-white rounded-lg shadow p-4 mb-6">
        <div className="flex flex-wrap gap-3 items-end">
          <div>
            <label className="block text-sm text-gray-600 mb-1">Search</label>
            <input
              type="text"
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              onKeyDown={(e) => e.key === 'Enter' && handleSearch()}
              placeholder="Name or FQDN..."
              className="border rounded px-3 py-1.5 text-sm w-64"
            />
          </div>
          <div>
            <label className="block text-sm text-gray-600 mb-1">Source</label>
            <select
              value={filter.source || ''}
              onChange={(e) => setFilter((f) => ({ ...f, source: e.target.value || undefined, offset: 0 }))}
              className="border rounded px-3 py-1.5 text-sm"
            >
              <option value="">All Sources</option>
              <option value="nmap">Nmap</option>
              <option value="vmware">VMware</option>
              <option value="zabbix">Zabbix</option>
              <option value="ansible">Ansible</option>
              <option value="k8s">Kubernetes</option>
              <option value="netscaler">NetScaler</option>
            </select>
          </div>
          <div>
            <label className="block text-sm text-gray-600 mb-1">Status</label>
            <select
              value={filter.status || ''}
              onChange={(e) => setFilter((f) => ({ ...f, status: e.target.value || undefined, offset: 0 }))}
              className="border rounded px-3 py-1.5 text-sm"
            >
              <option value="">All</option>
              <option value="active">Active</option>
              <option value="stale">Stale</option>
              <option value="removed">Removed</option>
            </select>
          </div>
          <button
            onClick={handleSearch}
            className="bg-blue-600 text-white px-4 py-1.5 rounded text-sm hover:bg-blue-700"
          >
            Search
          </button>
        </div>
      </div>

      {/* Asset Table */}
      <div className="bg-white rounded-lg shadow overflow-hidden">
        <table className="w-full text-sm">
          <thead className="bg-gray-50 text-left">
            <tr>
              <th className="px-4 py-3 font-medium">Name</th>
              <th className="px-4 py-3 font-medium">Type</th>
              <th className="px-4 py-3 font-medium">Source</th>
              <th className="px-4 py-3 font-medium">IPs</th>
              <th className="px-4 py-3 font-medium">Status</th>
              <th className="px-4 py-3 font-medium">Last Seen</th>
            </tr>
          </thead>
          <tbody>
            {query.isLoading ? (
              <tr>
                <td colSpan={6} className="px-4 py-8 text-center text-gray-500">
                  Loading...
                </td>
              </tr>
            ) : assets.length === 0 ? (
              <tr>
                <td colSpan={6} className="px-4 py-8 text-center text-gray-500">
                  No assets found
                </td>
              </tr>
            ) : (
              assets.map((asset) => (
                <tr key={asset.id} className="border-t border-gray-100 hover:bg-gray-50">
                  <td className="px-4 py-3">
                    <Link to={`/assets/${asset.id}`} className="text-blue-600 hover:underline">
                      {asset.name || asset.external_id}
                    </Link>
                    {asset.fqdn && (
                      <span className="block text-xs text-gray-400">{asset.fqdn}</span>
                    )}
                  </td>
                  <td className="px-4 py-3">
                    <span className="bg-gray-100 px-2 py-0.5 rounded text-xs">{asset.asset_type}</span>
                  </td>
                  <td className="px-4 py-3 capitalize">{asset.source}</td>
                  <td className="px-4 py-3 font-mono text-xs">
                    {asset.ip_addresses?.slice(0, 2).join(', ')}
                    {(asset.ip_addresses?.length ?? 0) > 2 && '...'}
                  </td>
                  <td className="px-4 py-3">
                    <StatusBadge status={asset.status} />
                  </td>
                  <td className="px-4 py-3 text-gray-500">
                    {new Date(asset.last_seen).toLocaleString()}
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>

        {/* Pagination */}
        {total > (filter.limit ?? 50) && (
          <div className="flex items-center justify-between px-4 py-3 border-t border-gray-100">
            <span className="text-sm text-gray-500">
              Showing {(filter.offset ?? 0) + 1}-{Math.min((filter.offset ?? 0) + (filter.limit ?? 50), total)} of {total}
            </span>
            <div className="flex gap-2">
              <button
                onClick={() => setFilter((f) => ({ ...f, offset: Math.max(0, (f.offset ?? 0) - (f.limit ?? 50)) }))}
                disabled={(filter.offset ?? 0) === 0}
                className="px-3 py-1 border rounded text-sm disabled:opacity-50"
              >
                Previous
              </button>
              <button
                onClick={() => setFilter((f) => ({ ...f, offset: (f.offset ?? 0) + (f.limit ?? 50) }))}
                disabled={(filter.offset ?? 0) + (filter.limit ?? 50) >= total}
                className="px-3 py-1 border rounded text-sm disabled:opacity-50"
              >
                Next
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}

function StatusBadge({ status }: { status: string }) {
  const colorMap: Record<string, string> = {
    active: 'bg-green-100 text-green-800',
    stale: 'bg-yellow-100 text-yellow-800',
    removed: 'bg-red-100 text-red-800',
  };

  return (
    <span className={`px-2 py-0.5 rounded text-xs font-medium ${colorMap[status] || 'bg-gray-100'}`}>
      {status}
    </span>
  );
}
