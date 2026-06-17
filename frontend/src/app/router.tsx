import { createBrowserRouter, Navigate } from 'react-router-dom';
import { ProtectedRoute } from './ProtectedRoute';
import { GlobalErrorBoundary } from './ErrorBoundary';
import { LoginPage } from '../features/auth/LoginPage';
import { RegisterPage } from '../features/auth/RegisterPage';
import { ProjectDashboardPage } from '../features/projects/ProjectDashboardPage';
import { ProjectsListPage } from '../features/projects/ProjectsListPage';
import { ApiTokensPage } from '../features/api-tokens/ApiTokensPage';

export const router = createBrowserRouter([
  {
    path: '/login',
    element: <LoginPage />,
    errorElement: <GlobalErrorBoundary />,
  },
  {
    path: '/register',
    element: <RegisterPage />,
    errorElement: <GlobalErrorBoundary />,
  },
  {
    path: '/',
    element: <ProtectedRoute />,
    errorElement: <GlobalErrorBoundary />,
    children: [
      {
        path: '/',
        element: <Navigate to="/projects" replace />,
      },
      {
        path: '/projects',
        element: <ProjectsListPage />,
      },
      {
        path: '/projects/:projectId',
        element: <ProjectDashboardPage />,
      },
      {
        path: '/tokens',
        element: <ApiTokensPage />,
      },
    ],
  },
]);
