import { useState, useEffect } from 'react';
import {
  Users,
  Store,
  Server,
  Activity,
  Shield,
  HardDrive,
  Zap,
  AlertCircle,
  CheckCircle,
  Clock,
  TrendingUp,
  Database,
  Globe,
  RefreshCw,
} from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle } from '../../components/ui/Card';
import { Button, DashboardSkeleton, Badge } from '../../components/ui';
import { adminService, type SystemStats, type ServiceStatus } from '../../api/services';
import { useNavigate } from 'react-router-dom';

interface StatCardProps {
  title: string;
  value: string | number;
  icon: React.ReactNode;
  trend?: {
    value: number;
    isPositive: boolean;
  };
  color?: 'primary' | 'success' | 'warning' | 'error';
  onClick?: () => void;
}

function StatCard({ title, value, icon, trend, color = 'primary', onClick }: StatCardProps) {
  const colorClasses = {
    primary: 'from-[#7BA4D0]/30 to-[#E7F0FA]/20',
    success: 'from-green-500/30 to-emerald-500/20',
    warning: 'from-yellow-500/30 to-orange-500/20',
    error: 'from-red-500/30 to-rose-500/20',
  };

  return (
    <Card hover className={onClick ? 'cursor-pointer' : ''} onClick={onClick}>
      <CardContent className="flex items-start justify-between">
        <div>
          <p className="text-sm font-medium text-[var(--color-text-secondary)]">{title}</p>
          <p className="text-3xl font-bold text-[var(--color-text-primary)] mt-2">{value}</p>
          {trend && (
            <div className="flex items-center gap-1 mt-2">
              <TrendingUp
                size={16}
                className={trend.isPositive ? 'text-[var(--color-success)]' : 'text-[var(--color-error)]'}
              />
              <span
                className={`text-sm font-medium ${
                  trend.isPositive ? 'text-[var(--color-success)]' : 'text-[var(--color-error)]'
                }`}
              >
                {trend.value}%
              </span>
            </div>
          )}
        </div>
        <div className={`p-3 rounded-xl bg-gradient-to-br ${colorClasses[color]}`}>
          {icon}
        </div>
      </CardContent>
    </Card>
  );
}

export function AdminDashboardPage() {
  const navigate = useNavigate();
  const [isLoading, setIsLoading] = useState(true);
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [systemStats, setSystemStats] = useState<SystemStats | null>(null);
  const [services, setServices] = useState<ServiceStatus[]>([]);

  const loadDashboardData = async (showRefreshing = false) => {
    try {
      if (showRefreshing) setIsRefreshing(true);
      else setIsLoading(true);
      
      setError(null);

      const [stats, serviceStatus] = await Promise.all([
        adminService.getSystemStats(),
        adminService.getServiceStatus(),
      ]);

      setSystemStats(stats);
      setServices(serviceStatus);
    } catch (err: any) {
      console.error('Failed to load dashboard data:', err);
      setError(err.message || 'Failed to load dashboard data');
    } finally {
      setIsLoading(false);
      setIsRefreshing(false);
    }
  };

  useEffect(() => {
    loadDashboardData();

    // Auto-refresh every 5 seconds for real-time updates
    const interval = setInterval(() => {
      loadDashboardData(true);
    }, 5000);

    return () => clearInterval(interval);
  }, []);

  if (isLoading && !systemStats) {
    return <DashboardSkeleton />;
  }

  if (error && !systemStats) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <Card className="max-w-md">
          <CardContent className="p-8 text-center">
            <AlertCircle size={48} className="mx-auto text-[var(--color-error)] mb-4" />
            <h2 className="text-xl font-bold text-[var(--color-text-primary)] mb-2">
              Failed to Load Dashboard
            </h2>
            <p className="text-[var(--color-text-secondary)] mb-4">{error}</p>
            <Button onClick={() => loadDashboardData()}>Try Again</Button>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Page Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-[var(--color-text-primary)]">System Dashboard</h1>
          <p className="text-[var(--color-text-secondary)] mt-1">
            Overview of system health, resources, and activity
            <span className="ml-2 text-xs opacity-75">(Auto-refreshing every 5s)</span>
          </p>
        </div>
        <Button
          variant="ghost"
          onClick={() => loadDashboardData(true)}
          disabled={isRefreshing}
          leftIcon={<RefreshCw size={18} className={isRefreshing ? 'animate-spin' : ''} />}
        >
          {isRefreshing ? 'Refreshing...' : 'Refresh'}
        </Button>
      </div>

      {/* System Health Stats */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        <StatCard
          title="Total Users"
          value={systemStats?.users.total || 0}
          icon={<Users size={24} className="text-[#E7F0FA]" />}
          onClick={() => navigate('/admin/users')}
        />
        <StatCard
          title="Resellers"
          value={systemStats?.resellers.total || 0}
          icon={<Store size={24} className="text-[#E7F0FA]" />}
          onClick={() => navigate('/admin/resellers')}
        />
        <StatCard
          title="Total Domains"
          value={systemStats?.domains.total || 0}
          icon={<Globe size={24} className="text-[#E7F0FA]" />}
          onClick={() => navigate('/admin/domains')}
        />
        <StatCard
          title="Databases"
          value={systemStats?.databases.total || 0}
          icon={<Database size={24} className="text-[#E7F0FA]" />}
          onClick={() => navigate('/admin/databases')}
        />
      </div>

      {/* System Resources */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <Card>
          <CardHeader>
            <CardTitle>System Resources</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div>
              <div className="flex justify-between text-sm mb-2">
                <span className="text-[var(--color-text-secondary)]">CPU Usage</span>
                <span className="text-[var(--color-text-primary)] font-medium">
                  {systemStats?.resources.cpu.usedCores.toFixed(1)} / {systemStats?.resources.cpu.cores} cores ({systemStats?.resources.cpu.usage.toFixed(1)}%)
                </span>
              </div>
              <div className="h-2 bg-[var(--color-primary-dark)] rounded-full overflow-hidden">
                <div
                  className="h-full bg-gradient-to-r from-[#7BA4D0] to-[#E7F0FA] rounded-full transition-all duration-500"
                  style={{ width: `${systemStats?.resources.cpu.usage}%` }}
                />
              </div>
            </div>
            <div>
              <div className="flex justify-between text-sm mb-2">
                <span className="text-[var(--color-text-secondary)]">Memory Usage</span>
                <span className="text-[var(--color-text-primary)] font-medium">
                  {systemStats?.resources.memory.used.toFixed(1)} / {systemStats?.resources.memory.total} GB ({systemStats?.resources.memory.usage.toFixed(1)}%)
                </span>
              </div>
              <div className="h-2 bg-[var(--color-primary-dark)] rounded-full overflow-hidden">
                <div
                  className="h-full bg-gradient-to-r from-[#7BA4D0] to-[#E7F0FA] rounded-full transition-all duration-500"
                  style={{ width: `${systemStats?.resources.memory.usage}%` }}
                />
              </div>
            </div>
            <div>
              <div className="flex justify-between text-sm mb-2">
                <span className="text-[var(--color-text-secondary)]">Disk Usage</span>
                <span className="text-[var(--color-text-primary)] font-medium">
                  {systemStats?.resources.disk.used.toFixed(1)} / {systemStats?.resources.disk.total} GB ({systemStats?.resources.disk.usage.toFixed(1)}%)
                </span>
              </div>
              <div className="h-2 bg-[var(--color-primary-dark)] rounded-full overflow-hidden">
                <div
                  className="h-full bg-gradient-to-r from-[#7BA4D0] to-[#E7F0FA] rounded-full transition-all duration-500"
                  style={{ width: `${systemStats?.resources.disk.usage}%` }}
                />
              </div>
            </div>
            <div>
              <div className="flex justify-between text-sm mb-2">
                <span className="text-[var(--color-text-secondary)]">Network Usage</span>
                <span className="text-[var(--color-text-primary)] font-medium">
                  {systemStats?.resources.network.bandwidth} ({systemStats?.resources.network.usage.toFixed(1)}%)
                </span>
              </div>
              <div className="h-2 bg-[var(--color-primary-dark)] rounded-full overflow-hidden">
                <div
                  className="h-full bg-gradient-to-r from-[#7BA4D0] to-[#E7F0FA] rounded-full transition-all duration-500"
                  style={{ width: `${systemStats?.resources.network.usage}%` }}
                />
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Service Status */}
        <Card>
          <CardHeader>
            <CardTitle>Service Status</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              {services.map((service) => (
                <div
                  key={service.name}
                  className="flex items-center justify-between p-3 rounded-lg bg-[var(--color-primary-dark)]/50"
                >
                  <div className="flex items-center gap-3">
                    {service.status === 'running' ? (
                      <CheckCircle size={20} className="text-[var(--color-success)]" />
                    ) : (
                      <AlertCircle size={20} className="text-[var(--color-error)]" />
                    )}
                    <span className="text-sm text-[var(--color-text-primary)]">{service.name}</span>
                  </div>
                  <div className="flex items-center gap-2">
                    <span className="text-xs text-[var(--color-text-muted)]">{service.uptime}</span>
                    <Badge
                      variant={service.status === 'running' ? 'success' : 'error'}
                      size="sm"
                    >
                      {service.status}
                    </Badge>
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Quick Stats */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-[var(--color-text-secondary)]">Active Users</p>
                <p className="text-2xl font-bold text-[var(--color-text-primary)] mt-1">
                  {systemStats?.users.active || 0}
                </p>
              </div>
              <Activity size={24} className="text-[var(--color-success)]" />
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="p-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-[var(--color-text-secondary)]">SSL Secured</p>
                <p className="text-2xl font-bold text-[var(--color-text-primary)] mt-1">
                  {systemStats?.domains.withSSL || 0}
                </p>
              </div>
              <Shield size={24} className="text-[var(--color-success)]" />
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="p-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-[var(--color-text-secondary)]">DB Storage</p>
                <p className="text-2xl font-bold text-[var(--color-text-primary)] mt-1">
                  {((systemStats?.databases.totalSizeMB || 0) / 1024).toFixed(1)} GB
                </p>
              </div>
              <HardDrive size={24} className="text-[#E7F0FA]" />
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="p-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-[var(--color-text-secondary)]">Active Domains</p>
                <p className="text-2xl font-bold text-[var(--color-text-primary)] mt-1">
                  {systemStats?.domains.active || 0}
                </p>
              </div>
              <Globe size={24} className="text-[#E7F0FA]" />
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Quick Actions */}
      <Card>
        <CardHeader>
          <CardTitle>Quick Actions</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            <button
              onClick={() => navigate('/admin/users')}
              className="p-4 rounded-lg bg-[var(--color-primary-dark)]/50 hover:bg-[var(--color-primary-dark)] transition-colors text-left"
            >
              <Users size={24} className="text-[#E7F0FA] mb-2" />
              <p className="text-sm font-medium text-[var(--color-text-primary)]">Manage Users</p>
            </button>
            <button
              onClick={() => navigate('/admin/resellers')}
              className="p-4 rounded-lg bg-[var(--color-primary-dark)]/50 hover:bg-[var(--color-primary-dark)] transition-colors text-left"
            >
              <Store size={24} className="text-[#E7F0FA] mb-2" />
              <p className="text-sm font-medium text-[var(--color-text-primary)]">Resellers</p>
            </button>
            <button
              onClick={() => navigate('/admin/logs')}
              className="p-4 rounded-lg bg-[var(--color-primary-dark)]/50 hover:bg-[var(--color-primary-dark)] transition-colors text-left"
            >
              <Server size={24} className="text-[#E7F0FA] mb-2" />
              <p className="text-sm font-medium text-[var(--color-text-primary)]">View Logs</p>
            </button>
            <button
              onClick={() => navigate('/admin/security')}
              className="p-4 rounded-lg bg-[var(--color-primary-dark)]/50 hover:bg-[var(--color-primary-dark)] transition-colors text-left"
            >
              <Shield size={24} className="text-[#E7F0FA] mb-2" />
              <p className="text-sm font-medium text-[var(--color-text-primary)]">Security</p>
            </button>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
