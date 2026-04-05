import { BrowserRouter, Routes, Route } from 'react-router-dom';
import { Layout } from './components/layout/Layout';
import { Dashboard } from './pages/Dashboard';
import { ToolBrowser } from './pages/ToolBrowser';
import { ToolExecutor } from './pages/ToolExecutor';
import { QuickActions } from './pages/QuickActions';
import { WorkflowBuilder } from './pages/WorkflowBuilder';
import { Settings } from './pages/Settings';
import { InventoryDashboard } from './pages/InventoryDashboard';
import { InventoryList } from './pages/InventoryList';
import { InventoryDetail } from './pages/InventoryDetail';

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Layout />}>
          <Route index element={<Dashboard />} />
          <Route path="tools" element={<ToolBrowser />} />
          <Route path="tools/:toolName" element={<ToolExecutor />} />
          <Route path="actions" element={<QuickActions />} />
          <Route path="workflows" element={<WorkflowBuilder />} />
          <Route path="inventory" element={<InventoryDashboard />} />
          <Route path="inventory/list" element={<InventoryList />} />
          <Route path="inventory/:sku" element={<InventoryDetail />} />
          <Route path="settings" element={<Settings />} />
        </Route>
      </Routes>
    </BrowserRouter>
  );
}
