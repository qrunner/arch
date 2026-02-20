import { useQuery } from '@tanstack/react-query';
import { getDashboardStats, getRecentChanges } from '../api/client';

export default function DashboardPage() {
  const statsQuery = useQuery({
    queryKey: ['dashboard-stats'],
    queryFn: () => getDashboardStats(),
  });

  const changesQuery = useQuery({
    queryKey: ['recent-changes'],
    queryFn: () => getRecentChanges(10),
  });

  const stats = statsQuery.data?.data;

  return (
    <div>
      <h1 className="text-2xl font-bold mb-6">Dashboard</h1>

      {/* Summary Cards */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-8">
        <StatCard
          title="Total Assets"
          value={stats?.total_assets ?? 0}
          color="blue"
        />
        <StatCard
          title="Active"
          value={stats?.by_status?.active ?? 0}
          color="green"
        />
        <StatCard
          title="Stale"
          value={stats?.by_status?.stale ?? 0}
          color="yellow"
        />
        <StatCard
          title="Recent Changes (24h)"
          value={stats?.recent_changes ?? 0}
          color="purple"
        />
      </div>

      {/* By Source */}
      {stats?.by_source && Object.keys(stats.by_source).length > 0 && (
        <div className="bg-white rounded-lg shadow p-6 mb-6">
          <h2 className="text-lg font-semibold mb-4">Assets by Source</h2>
          <div className="grid grid-cols-2 md:grid-cols-3 gap-3">
            {Object.entries(stats.by_source).map(([source, count]) => (
              <div key={source} className="flex justify-between items-center px-3 py-2 bg-gray-50 rounded">
                <span className="text-gray-700 capitalize">{source}</span>
                <span className="font-mono font-semibold">{count}</span>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Recent Changes */}
      <div className="bg-white rounded-lg shadow p-6">
        <h2 className="text-lg font-semibold mb-4">Recent Changes</h2>
        {changesQuery.isLoading ? (
          <p className="text-gray-500">Loading...</p>
        ) : changesQuery.data?.data?.length ? (
          <div className="space-y-2">
            {changesQuery.data.data.map((event) => (
              <div key={event.id} className="flex items-center gap-3 py-2 border-b border-gray-100">
                <ActionBadge action={event.action} />
                <span className="text-sm text-gray-600">{event.source}</span>
                <span className="text-sm text-gray-400 ml-auto">
                  {new Date(event.timestamp).toLocaleString()}
                </span>
              </div>
            ))}
          </div>
        ) : (
          <p className="text-gray-500">No recent changes</p>
        )}
      </div>
    </div>
  );
}

function StatCard({ title, value, color }: { title: string; value: number; color: string }) {
  const colorMap: Record<string, string> = {
    blue: 'bg-blue-50 text-blue-700 border-blue-200',
    green: 'bg-green-50 text-green-700 border-green-200',
    yellow: 'bg-yellow-50 text-yellow-700 border-yellow-200',
    purple: 'bg-purple-50 text-purple-700 border-purple-200',
  };

  return (
    <div className={`rounded-lg border p-4 ${colorMap[color] || colorMap.blue}`}>
      <p className="text-sm opacity-75">{title}</p>
      <p className="text-3xl font-bold">{value}</p>
    </div>
  );
}

function ActionBadge({ action }: { action: string }) {
  const colorMap: Record<string, string> = {
    'asset.created': 'bg-green-100 text-green-800',
    'asset.updated': 'bg-blue-100 text-blue-800',
    'asset.removed': 'bg-red-100 text-red-800',
    'relationship.changed': 'bg-purple-100 text-purple-800',
  };

  return (
    <span className={`px-2 py-0.5 rounded text-xs font-medium ${colorMap[action] || 'bg-gray-100 text-gray-800'}`}>
      {action}
    </span>
  );
}
