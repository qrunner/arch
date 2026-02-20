import { Routes, Route, Link } from 'react-router-dom';
import DashboardPage from './pages/DashboardPage';
import AssetsPage from './pages/AssetsPage';
import AssetDetailPage from './pages/AssetDetailPage';
import GraphPage from './pages/GraphPage';

export default function App() {
  return (
    <div className="min-h-screen">
      <nav className="bg-white border-b border-gray-200 px-6 py-3">
        <div className="flex items-center justify-between max-w-7xl mx-auto">
          <div className="flex items-center space-x-8">
            <Link to="/" className="text-xl font-bold text-gray-900">
              IT Asset Inventory
            </Link>
            <div className="flex space-x-4">
              <Link to="/" className="text-gray-600 hover:text-gray-900">
                Dashboard
              </Link>
              <Link to="/assets" className="text-gray-600 hover:text-gray-900">
                Assets
              </Link>
              <Link to="/graph" className="text-gray-600 hover:text-gray-900">
                Graph
              </Link>
            </div>
          </div>
        </div>
      </nav>

      <main className="max-w-7xl mx-auto px-6 py-8">
        <Routes>
          <Route path="/" element={<DashboardPage />} />
          <Route path="/assets" element={<AssetsPage />} />
          <Route path="/assets/:id" element={<AssetDetailPage />} />
          <Route path="/graph" element={<GraphPage />} />
          <Route path="/graph/:id" element={<GraphPage />} />
        </Routes>
      </main>
    </div>
  );
}
