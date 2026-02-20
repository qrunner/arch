import { useParams, Link } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { getAsset, getAssetHistory, getAssetRelationships } from '../api/client';

export default function AssetDetailPage() {
  const { id } = useParams<{ id: string }>();

  const assetQuery = useQuery({
    queryKey: ['asset', id],
    queryFn: () => getAsset(id!),
    enabled: !!id,
  });

  const historyQuery = useQuery({
    queryKey: ['asset-history', id],
    queryFn: () => getAssetHistory(id!),
    enabled: !!id,
  });

  const relsQuery = useQuery({
    queryKey: ['asset-relationships', id],
    queryFn: () => getAssetRelationships(id!),
    enabled: !!id,
  });

  const asset = assetQuery.data?.data;

  if (assetQuery.isLoading) {
    return <p className="text-gray-500">Loading...</p>;
  }

  if (!asset) {
    return <p className="text-red-500">Asset not found</p>;
  }

  return (
    <div>
      <div className="flex items-center gap-3 mb-6">
        <Link to="/assets" className="text-blue-600 hover:underline text-sm">
          &larr; Assets
        </Link>
        <h1 className="text-2xl font-bold">{asset.name}</h1>
        <span className={`px-2 py-0.5 rounded text-xs font-medium ${
          asset.status === 'active' ? 'bg-green-100 text-green-800' :
          asset.status === 'stale' ? 'bg-yellow-100 text-yellow-800' :
          'bg-red-100 text-red-800'
        }`}>
          {asset.status}
        </span>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Asset Info */}
        <div className="bg-white rounded-lg shadow p-6">
          <h2 className="text-lg font-semibold mb-4">Details</h2>
          <dl className="grid grid-cols-2 gap-y-3 text-sm">
            <dt className="text-gray-500">ID</dt>
            <dd className="font-mono text-xs">{asset.id}</dd>
            <dt className="text-gray-500">External ID</dt>
            <dd>{asset.external_id}</dd>
            <dt className="text-gray-500">Source</dt>
            <dd className="capitalize">{asset.source}</dd>
            <dt className="text-gray-500">Type</dt>
            <dd>{asset.asset_type}</dd>
            <dt className="text-gray-500">FQDN</dt>
            <dd>{asset.fqdn || '-'}</dd>
            <dt className="text-gray-500">IP Addresses</dt>
            <dd className="font-mono text-xs">{asset.ip_addresses?.join(', ') || '-'}</dd>
            <dt className="text-gray-500">First Seen</dt>
            <dd>{new Date(asset.first_seen).toLocaleString()}</dd>
            <dt className="text-gray-500">Last Seen</dt>
            <dd>{new Date(asset.last_seen).toLocaleString()}</dd>
          </dl>

          <Link
            to={`/graph/${asset.id}`}
            className="inline-block mt-4 text-sm text-blue-600 hover:underline"
          >
            View in dependency graph &rarr;
          </Link>
        </div>

        {/* Attributes */}
        <div className="bg-white rounded-lg shadow p-6">
          <h2 className="text-lg font-semibold mb-4">Attributes</h2>
          <pre className="bg-gray-50 rounded p-4 text-xs overflow-auto max-h-80">
            {JSON.stringify(asset.attributes, null, 2)}
          </pre>
        </div>

        {/* Relationships */}
        <div className="bg-white rounded-lg shadow p-6">
          <h2 className="text-lg font-semibold mb-4">Relationships</h2>
          {relsQuery.isLoading ? (
            <p className="text-gray-500">Loading...</p>
          ) : relsQuery.data?.data?.length ? (
            <ul className="space-y-2">
              {relsQuery.data.data.map((rel) => (
                <li key={rel.id} className="flex items-center gap-2 text-sm">
                  <span className="bg-purple-100 text-purple-800 px-2 py-0.5 rounded text-xs">
                    {rel.type}
                  </span>
                  <span className="text-gray-500">&rarr;</span>
                  <Link
                    to={`/assets/${rel.to_id === id ? rel.from_id : rel.to_id}`}
                    className="text-blue-600 hover:underline font-mono text-xs"
                  >
                    {rel.to_id === id ? rel.from_id : rel.to_id}
                  </Link>
                </li>
              ))}
            </ul>
          ) : (
            <p className="text-gray-500 text-sm">No relationships found</p>
          )}
        </div>

        {/* Change History */}
        <div className="bg-white rounded-lg shadow p-6">
          <h2 className="text-lg font-semibold mb-4">Change History</h2>
          {historyQuery.isLoading ? (
            <p className="text-gray-500">Loading...</p>
          ) : historyQuery.data?.data?.length ? (
            <div className="space-y-3">
              {historyQuery.data.data.map((event) => (
                <div key={event.id} className="border-l-2 border-gray-200 pl-3">
                  <div className="flex items-center gap-2">
                    <span className={`px-2 py-0.5 rounded text-xs font-medium ${
                      event.action === 'asset.created' ? 'bg-green-100 text-green-800' :
                      event.action === 'asset.updated' ? 'bg-blue-100 text-blue-800' :
                      'bg-red-100 text-red-800'
                    }`}>
                      {event.action}
                    </span>
                    <span className="text-xs text-gray-400">
                      {new Date(event.timestamp).toLocaleString()}
                    </span>
                  </div>
                  {event.diff && (
                    <pre className="text-xs text-gray-600 mt-1">
                      {JSON.stringify(event.diff, null, 2)}
                    </pre>
                  )}
                </div>
              ))}
            </div>
          ) : (
            <p className="text-gray-500 text-sm">No history found</p>
          )}
        </div>
      </div>
    </div>
  );
}
