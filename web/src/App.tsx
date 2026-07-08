import { createBrowserRouter, RouterProvider } from 'react-router-dom';
import { Layout } from './components/Layout';
import { ProtectedRoute } from './components/ProtectedRoute';
import { LoginPage } from './components/LoginPage';
import { DashboardPage } from './pages/DashboardPage';
import { ProductsPage } from './pages/ProductsPage';
import { ProductDetailPage } from './pages/ProductDetailPage';
import { KanbanPage } from './pages/KanbanPage';
import { FindingsPage } from './pages/FindingsPage';
import { AdminPanelPage } from './pages/AdminPanelPage';
import { ReportsPage } from './pages/ReportsPage';
import { RulesPage } from './pages/RulesPage';
import { TopologyPage } from './pages/TopologyPage';
import { ScannersPage } from './pages/ScannersPage';
import { TerminalPage } from './pages/TerminalPage';
import { AIChatPage } from './pages/AIChatPage';
import { CommandCenterPage } from './pages/CommandCenterPage';
import { FAQPage } from './pages/FAQPage';
import { RouteError } from './ui/ErrorBoundary';

const router = createBrowserRouter([
  {
    path: '/login',
    element: <LoginPage />,
  },
  {
    path: '/',
    element: <Layout />,
    errorElement: <RouteError />,
    children: [
      {
        index: true,
        element: (
          <ProtectedRoute>
            <DashboardPage />
          </ProtectedRoute>
        ),
      },
      {
        path: 'cc',
        element: (
          <ProtectedRoute>
            <CommandCenterPage />
          </ProtectedRoute>
        ),
      },
      {
        path: 'products',
        element: (
          <ProtectedRoute>
            <ProductsPage />
          </ProtectedRoute>
        ),
      },
      {
        path: 'products/:id',
        element: (
          <ProtectedRoute>
            <ProductDetailPage />
          </ProtectedRoute>
        ),
      },
      {
        path: 'kanban',
        element: (
          <ProtectedRoute>
            <KanbanPage />
          </ProtectedRoute>
        ),
      },
      {
        path: 'findings',
        element: (
          <ProtectedRoute>
            <FindingsPage />
          </ProtectedRoute>
        ),
      },
      {
        path: 'chat',
        element: (
          <ProtectedRoute>
            <AIChatPage />
          </ProtectedRoute>
        ),
      },
      {
        path: 'rules',
        element: (
          <ProtectedRoute>
            <RulesPage />
          </ProtectedRoute>
        ),
      },
      {
        path: 'topology',
        element: (
          <ProtectedRoute>
            <TopologyPage />
          </ProtectedRoute>
        ),
      },
      {
        path: 'scanners',
        element: (
          <ProtectedRoute>
            <ScannersPage />
          </ProtectedRoute>
        ),
      },
      {
        path: 'terminal',
        element: (
          <ProtectedRoute>
            <TerminalPage />
          </ProtectedRoute>
        ),
      },
      {
        path: 'faq',
        element: (
          <ProtectedRoute>
            <FAQPage />
          </ProtectedRoute>
        ),
      },
      {
        path: 'reports',
        element: (
          <ProtectedRoute allowedRoles={['security_lead', 'admin', 'superadmin']}>
            <ReportsPage />
          </ProtectedRoute>
        ),
      },
      {
        path: 'admin',
        element: (
          <ProtectedRoute allowedRoles={['admin', 'superadmin']}>
            <AdminPanelPage />
          </ProtectedRoute>
        ),
      },
    ],
  },
]);

function App() {
  return <RouterProvider router={router} />;
}

export default App;
