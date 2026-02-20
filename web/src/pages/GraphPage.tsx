import { useState } from 'react';
import { useParams } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { getDependencyGraph, getImpactGraph } from '../api/client';

export default function GraphPage() {
  const { id } = useParams<{ id: string }>();
  const [assetId, setAssetId] = useState(id || '');
  const [mode, setMode] = useState<'dependencies' | 'impact'>('dependencies');
  const [depth, setDepth] = useState(3);
  const [queryId, setQueryId] = useState(id || '');

  const graphQuery = useQuery({
    queryKey: ['graph', mode, queryId, depth],
    queryFn: () =>
      mode === 'dependencies'
        ? getDependencyGraph(queryId, depth)
        : getImpactGraph(queryId, depth),
    enabled: !!queryId,
  });

  const handleExplore = () => {
    setQueryId(assetId);
  };

  const graphData = graphQuery.data?.data;

  return (
    <div>
      <h1 className="text-2xl font-bold mb-6">Dependency Graph</h1>

      {/* Controls */}
      <div className="bg-white rounded-lg shadow p-4 mb-6">
        <div className="flex flex-wrap gap-3 items-end">
          <div>
            <label className="block text-sm text-gray-600 mb-1">Asset ID</label>
            <input
              type="text"
              value={assetId}
              onChange={(e) => setAssetId(e.target.value)}
              placeholder="Enter asset UUID..."
              className="border rounded px-3 py-1.5 text-sm w-80 font-mono"
            />
          </div>
          <div>
            <label className="block text-sm text-gray-600 mb-1">Mode</label>
            <select
              value={mode}
              onChange={(e) => setMode(e.target.value as 'dependencies' | 'impact')}
              className="border rounded px-3 py-1.5 text-sm"
            >
              <option value="dependencies">Dependencies (outgoing)</option>
              <option value="impact">Impact / Blast Radius (incoming)</option>
            </select>
          </div>
          <div>
            <label className="block text-sm text-gray-600 mb-1">Depth</label>
            <input
              type="number"
              value={depth}
              onChange={(e) => setDepth(Number(e.target.value))}
              min={1}
              max={10}
              className="border rounded px-3 py-1.5 text-sm w-20"
            />
          </div>
          <button
            onClick={handleExplore}
            disabled={!assetId}
            className="bg-blue-600 text-white px-4 py-1.5 rounded text-sm hover:bg-blue-700 disabled:opacity-50"
          >
            Explore
          </button>
        </div>
      </div>

      {/* Graph Visualization Area */}
      <div className="bg-white rounded-lg shadow p-6">
        {!queryId ? (
          <div className="text-center text-gray-500 py-16">
            <p className="text-lg mb-2">Enter an asset ID to explore its graph</p>
            <p className="text-sm">
              Select an asset from the{' '}
              <a href="/assets" className="text-blue-600 hover:underline">
                Assets page
              </a>{' '}
              and click &quot;View in dependency graph&quot;
            </p>
          </div>
        ) : graphQuery.isLoading ? (
          <div className="text-center text-gray-500 py-16">Loading graph...</div>
        ) : graphData ? (
          <div>
            <div className="mb-4 text-sm text-gray-500">
              {graphData.assets?.length ?? 0} nodes, {graphData.relationships?.length ?? 0} edges
            </div>

            {/* Cytoscape.js will be mounted here in production */}
            <div
              id="cy"
              className="border border-gray-200 rounded bg-gray-50"
              style={{ height: '500px' }}
            >
              {/* Placeholder: Cytoscape.js graph visualization */}
              <div className="flex items-center justify-center h-full text-gray-400">
                <div className="text-center">
                  <p className="mb-2">Graph visualization area</p>
                  <p className="text-xs">
                    Cytoscape.js integration renders the interactive graph here
                  </p>
                </div>
              </div>
            </div>

            {/* Node list fallback */}
            {graphData.assets && graphData.assets.length > 0 && (
              <div className="mt-4">
                <h3 className="text-sm font-semibold text-gray-600 mb-2">Nodes</h3>
                <div className="flex flex-wrap gap-2">
                  {graphData.assets.map((a) => (
                    <a
                      key={a.id}
                      href={`/assets/${a.id}`}
                      className="px-3 py-1 bg-blue-50 text-blue-700 rounded text-xs hover:bg-blue-100"
                    >
                      {a.name || a.id} ({a.asset_type})
                    </a>
                  ))}
                </div>
              </div>
            )}
          </div>
        ) : (
          <div className="text-center text-gray-500 py-16">No graph data available</div>
        )}
      </div>
    </div>
  );
}
