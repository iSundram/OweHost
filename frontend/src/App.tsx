import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { AuthProvider } from './context/AuthContext';
import { ToastProvider as OldToastProvider } from './context/ToastContext';
import { ToastProvider, LoadingBar } from './components/ui';
import { useAuth } from './hooks/useAuth';
import { AdminLayout, ResellerLayout, UserLayout } from './components/layout';
import {
  EmailLoginPage,
  PasswordLoginPage,
  DashboardPage,
  AdminDashboardPage,
  ResellerDashboardPage,
} from './pages';
import { AdminUsersPage } from './pages/admin/AdminUsersPage';
import { AdminResellersPage } from './pages/admin/AdminResellersPage';
import { AdminResourcesPage } from './pages/admin/AdminResourcesPage';
import { AdminPackagesPage } from './pages/admin/AdminPackagesPage';
import { AdminFeaturesPage } from './pages/admin/AdminFeaturesPage';
import { AdminDomainsPage } from './pages/admin/AdminDomainsPage';
import { AdminDatabasesPage } from './pages/admin/AdminDatabasesPage';
import { AdminDNSPage } from './pages/admin/AdminDNSPage';
import { AdminSSLPage } from './pages/admin/AdminSSLPage';
import { AdminWebServerPage } from './pages/admin/AdminWebServerPage';
import { AdminFilesPage } from './pages/admin/AdminFilesPage';
import { AdminBackupsPage } from './pages/admin/AdminBackupsPage';
import { AdminSecurityPage } from './pages/admin/AdminSecurityPage';
import { AdminCronPage } from './pages/admin/AdminCronPage';
import { AdminAppsPage } from './pages/admin/AdminAppsPage';
import { AdminNodesPage } from './pages/admin/AdminNodesPage';
import { AdminPluginsPage } from './pages/admin/AdminPluginsPage';
import { AdminLogsPage } from './pages/admin/AdminLogsPage';
import { AdminNotificationsPage } from './pages/admin/AdminNotificationsPage';
import { AdminLicensePage } from './pages/admin/AdminLicensePage';
import { AdminSettingsPage } from './pages/admin/AdminSettingsPage';
import { AdminRecoveryPage } from './pages/admin/AdminRecoveryPage';
import { ResellerCustomersPage } from './pages/reseller/ResellerCustomersPage';
import { ResellerPackagesPage } from './pages/reseller/ResellerPackagesPage';
import { ResellerResourcesPage } from './pages/reseller/ResellerResourcesPage';
import { ResellerDomainsPage } from './pages/reseller/ResellerDomainsPage';
import { ResellerDatabasesPage } from './pages/reseller/ResellerDatabasesPage';
import { ResellerDNSPage } from './pages/reseller/ResellerDNSPage';
import { ResellerSSLPage } from './pages/reseller/ResellerSSLPage';
import { ResellerFilesPage } from './pages/reseller/ResellerFilesPage';
import { ResellerBackupsPage } from './pages/reseller/ResellerBackupsPage';
import { ResellerCronPage } from './pages/reseller/ResellerCronPage';
import { ResellerAppsPage } from './pages/reseller/ResellerAppsPage';
import { ResellerSubResellersPage } from './pages/reseller/ResellerSubResellersPage';
import { ResellerReportsPage } from './pages/reseller/ResellerReportsPage';
import { ResellerSettingsPage } from './pages/reseller/ResellerSettingsPage';
import { UserDomainsPage } from './pages/user/UserDomainsPage';
import { UserDNSPage } from './pages/user/UserDNSPage';
import { UserDatabasesPage } from './pages/user/UserDatabasesPage';
import { UserFilesPage } from './pages/user/UserFilesPage';
import { UserSSLPage } from './pages/user/UserSSLPage';
import { UserEmailPage } from './pages/user/UserEmailPage';
import { UserFTPPage } from './pages/user/UserFTPPage';
import { UserWebServerPage } from './pages/user/UserWebServerPage';
import { UserCronPage } from './pages/user/UserCronPage';
import { UserAppsPage } from './pages/user/UserAppsPage';
import { UserBackupsPage } from './pages/user/UserBackupsPage';
import { UserStatsPage } from './pages/user/UserStatsPage';
import { UserSecurityPage } from './pages/user/UserSecurityPage';
import { UserSettingsPage } from './pages/user/UserSettingsPage';
import { UserSupportPage } from './pages/user/UserSupportPage';
import type { UserRole } from './types';

// Protected Route wrapper
function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const { isAuthenticated, isLoading } = useAuth();

  if (isLoading) {
    return (
      <div className="min-h-screen bg-[var(--color-background)] flex items-center justify-center">
        {/* Only progress bar, no overlay */}
        <LoadingBar isLoading={true} message="Authenticating..." />
        <p className="text-[var(--color-text-secondary)]">Loading...</p>
      </div>
    );
  }

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }

  return <>{children}</>;
}

// Role-based Protected Route
function RoleProtectedRoute({ 
  children, 
  allowedRoles 
}: { 
  children: React.ReactNode;
  allowedRoles: UserRole[];
}) {
  const { user, isLoading } = useAuth();

  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-[var(--color-background)]">
        <div className="flex flex-col items-center gap-4">
          <div className="w-12 h-12 rounded-xl bg-gradient-to-br from-[#7BA4D0] to-[#E7F0FA] flex items-center justify-center animate-pulse">
            <span className="text-xl font-bold text-white">O</span>
          </div>
          <p className="text-[var(--color-text-secondary)]">Loading...</p>
        </div>
      </div>
    );
  }

  if (!user || !allowedRoles.includes(user.role)) {
    // Redirect to appropriate dashboard based on role
    const dashboardRoute = user?.role === 'admin' ? '/admin' : 
                          user?.role === 'reseller' ? '/reseller' : '/user';
    return <Navigate to={dashboardRoute} replace />;
  }

  return <>{children}</>;
}

// Get dashboard route based on user role
function getDashboardRoute(role?: UserRole): string {
  switch (role) {
    case 'admin':
      return '/admin';
    case 'reseller':
      return '/reseller';
    case 'user':
    default:
      return '/user';
  }
}


function AppRoutes() {
  const { user } = useAuth();

  return (
    <Routes>
      <Route path="/login" element={<EmailLoginPage />} />
      <Route path="/login/password" element={<PasswordLoginPage />} />
      
      {/* Admin Routes */}
      <Route
        path="/admin/*"
        element={
          <ProtectedRoute>
            <RoleProtectedRoute allowedRoles={['admin']}>
              <AdminLayout />
            </RoleProtectedRoute>
          </ProtectedRoute>
        }
      >
        <Route index element={<AdminDashboardPage />} />
        <Route path="users" element={<AdminUsersPage />} />
        <Route path="resellers" element={<AdminResellersPage />} />
        <Route path="packages" element={<AdminPackagesPage />} />
        <Route path="features" element={<AdminFeaturesPage />} />
        <Route path="resources" element={<AdminResourcesPage />} />
        <Route path="domains" element={<AdminDomainsPage />} />
        <Route path="databases" element={<AdminDatabasesPage />} />
        <Route path="dns" element={<AdminDNSPage />} />
        <Route path="ssl" element={<AdminSSLPage />} />
        <Route path="webserver" element={<AdminWebServerPage />} />
        <Route path="files" element={<AdminFilesPage />} />
        <Route path="backups" element={<AdminBackupsPage />} />
        <Route path="security" element={<AdminSecurityPage />} />
        <Route path="cron" element={<AdminCronPage />} />
        <Route path="apps" element={<AdminAppsPage />} />
        <Route path="nodes" element={<AdminNodesPage />} />
        <Route path="plugins" element={<AdminPluginsPage />} />
        <Route path="logs" element={<AdminLogsPage />} />
        <Route path="notifications" element={<AdminNotificationsPage />} />
        <Route path="license" element={<AdminLicensePage />} />
        <Route path="settings" element={<AdminSettingsPage />} />
        <Route path="recovery" element={<AdminRecoveryPage />} />
      </Route>

      {/* Reseller Routes */}
      <Route
        path="/reseller/*"
        element={
          <ProtectedRoute>
            <RoleProtectedRoute allowedRoles={['admin', 'reseller']}>
              <ResellerLayout />
            </RoleProtectedRoute>
          </ProtectedRoute>
        }
      >
        <Route index element={<ResellerDashboardPage />} />
        <Route path="customers" element={<ResellerCustomersPage />} />
        <Route path="packages" element={<ResellerPackagesPage />} />
        <Route path="resources" element={<ResellerResourcesPage />} />
        <Route path="domains" element={<ResellerDomainsPage />} />
        <Route path="databases" element={<ResellerDatabasesPage />} />
        <Route path="dns" element={<ResellerDNSPage />} />
        <Route path="ssl" element={<ResellerSSLPage />} />
        <Route path="files" element={<ResellerFilesPage />} />
        <Route path="backups" element={<ResellerBackupsPage />} />
        <Route path="cron" element={<ResellerCronPage />} />
        <Route path="apps" element={<ResellerAppsPage />} />
        <Route path="sub-resellers" element={<ResellerSubResellersPage />} />
        <Route path="reports" element={<ResellerReportsPage />} />
        <Route path="settings" element={<ResellerSettingsPage />} />
      </Route>

      {/* User Routes */}
      <Route
        path="/user/*"
        element={
          <ProtectedRoute>
            <RoleProtectedRoute allowedRoles={['admin', 'reseller', 'user']}>
              <UserLayout />
            </RoleProtectedRoute>
          </ProtectedRoute>
        }
      >
        <Route index element={<DashboardPage />} />
        <Route path="domains" element={<UserDomainsPage />} />
        <Route path="dns" element={<UserDNSPage />} />
        <Route path="databases" element={<UserDatabasesPage />} />
        <Route path="files" element={<UserFilesPage />} />
        <Route path="ssl" element={<UserSSLPage />} />
        <Route path="email" element={<UserEmailPage />} />
        <Route path="ftp" element={<UserFTPPage />} />
        <Route path="webserver" element={<UserWebServerPage />} />
        <Route path="cron" element={<UserCronPage />} />
        <Route path="apps" element={<UserAppsPage />} />
        <Route path="backups" element={<UserBackupsPage />} />
        <Route path="stats" element={<UserStatsPage />} />
        <Route path="security" element={<UserSecurityPage />} />
        <Route path="settings" element={<UserSettingsPage />} />
        <Route path="support" element={<UserSupportPage />} />
      </Route>

      {/* Legacy routes - redirect to role-based dashboard */}
      <Route
        path="/"
        element={
          <ProtectedRoute>
            <Navigate to={getDashboardRoute(user?.role)} replace />
          </ProtectedRoute>
        }
      />
      
      <Route path="*" element={<Navigate to={getDashboardRoute(user?.role)} replace />} />
    </Routes>
  );
}

function App() {
  return (
    <BrowserRouter>
      <AuthProvider>
        <ToastProvider>
          <OldToastProvider>
            <AppRoutes />
          </OldToastProvider>
        </ToastProvider>
      </AuthProvider>
    </BrowserRouter>
  );
}

export default App;
