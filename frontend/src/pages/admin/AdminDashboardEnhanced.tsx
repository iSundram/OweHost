import { useState, useEffect } from 'react';
import {
  Users,
  Store,
  Server,
  Activity,
  Shield,
  AlertCircle,
  CheckCircle,
  TrendingUp,
  Database,
  Globe,
  Settings,
  Bell,
} from 'lucide-react';
import { 
  Card, 
  CardContent, 
  CardHeader, 
  CardTitle,
  Button,
  Badge,
  StatusBadge,
  DashboardSkeleton
} from '../../components/ui';

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
        <div className="flex-1">
          <p className="text-sm font-medium text-[var(--color-text-secondary)]">{title}</p>
          <p className="text-3xl font-bold text-[var(--color-text-primary)] mt-2">{value}</p>
          {trend && (
            <div className="flex items-center gap-1 mt-2">
              <TrendingUp
                size={16}
                className={trend.isPositive ? 'text-[var(--color-success)]' : 'text-[var(--color-error)] rotate-180'}
              />
              <span
                className={`text-sm font-medium ${
                  trend.isPositive ? 'text-[var(--color-success)]' : 'text-[var(--color-error)]'
                }`}
              >
                {Math.abs(trend.value)}%
              </span>
              <span className="text-sm text-[var(--color-text-muted)]">vs last month</span>
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

interface ServiceStatus {
  name: string;
  status: 'running' | 'stopped' | 'warning';
  uptime: string;
  load?: string;
}

export function AdminDashboardEnhanced() {
  const [isLoading, setIsLoading] = useState(true);
  const [systemStats, setSystemStats] = useState({
    totalUsers: 125,
    totalResellers: 8,
    totalDomains: 342,
    totalDatabases: 189,
    activeUsers: 118,
    cpuUsage: 45.2,
    memoryUsage: 68.5,
    diskUsage: 52.3,
    networkUsage: 12.8,
  });

  const [services, setServices] = useState<ServiceStatus[]>([
    { name: 'Nginx Web Server', status: 'running', uptime: '99.9%', load: 'Low' },
    { name: 'MySQL Database', status: 'running', uptime: '99.8%', load: 'Medium' },
    { name: 'DNS Server', status: 'running', uptime: '100%', load: 'Low' },
    { name: 'Mail Server', status: 'running', uptime: '99.5%', load: 'Low' },
    { name: 'FTP Server', status: 'running', uptime: '99.7%', load: 'Low' },
  ]);

  const [alerts] = useState([
    { id: 1, type: 'warning', message: 'User "john_doe" approaching disk quota (90% used)', time: '5m ago' },
    { id: 2, type: 'info', message: '3 SSL certificates expiring in 30 days', time: '1h ago' },
    { id: 3, type: 'success', message: 'System backup completed successfully', time: '2h ago' },
  ]);

  useEffect(() => {
    // Simulate loading
    setTimeout(() => setIsLoading(false), 500);
  }, []);

  if (isLoading) {
    return <DashboardSkeleton />;
  }

  return (
    <div className="space-y-6">
      {/* Page Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-[var(--color-text-primary)]">System Dashboard</h1>
          <p className="text-[var(--color-text-secondary)] mt-1">
            Overview of system health, resources, and activity
          </p>
        </div>
        <div className="flex items-center gap-3">
          <Button variant="ghost" leftIcon={<Settings size={18} />}>
            Settings
          </Button>
          <Button variant="ghost" leftIcon={<Bell size={18} />}>
            Alerts ({alerts.filter(a => a.type !== 'success').length})
          </Button>
        </div>
      </div>

      {/* System Stats */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        <StatCard
          title="Total Users"
          value={systemStats.totalUsers}
          icon={<Users size={24} className="text-[#E7F0FA]" />}
          trend={{ value: 12, isPositive: true }}
          onClick={() => window.location.href = '/admin/users'}
        />
        <StatCard
          title="Resellers"
          value={systemStats.totalResellers}
          icon={<Store size={24} className="text-[#E7F0FA]" />}
          trend={{ value: 5, isPositive: true }}
          onClick={() => window.location.href = '/admin/resellers'}
        />
        <StatCard
          title="Total Domains"
          value={systemStats.totalDomains}
          icon={<Globe size={24} className="text-[#E7F0FA]" />}
          trend={{ value: 8, isPositive: true }}
          onClick={() => window.location.href = '/admin/domains'}
        />
        <StatCard
          title="Databases"
          value={systemStats.totalDatabases}
          icon={<Database size={24} className="text-[#E7F0FA]" />}
          trend={{ value: 3, isPositive: true }}
          onClick={() => window.location.href = '/admin/databases'}
        />
      </div>

      {/* Main Content */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* System Resources */}
        <Card className="lg:col-span-2">
          <CardHeader>
            <CardTitle>System Resources</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            {[
              { label: 'CPU Usage', value: systemStats.cpuUsage, color: 'primary' as const },
              { label: 'Memory Usage', value: systemStats.memoryUsage, color: 'warning' as const },
              { label: 'Disk Usage', value: systemStats.diskUsage, color: 'success' as const },
              { label: 'Network Usage', value: systemStats.networkUsage, color: 'info' as const },
            ].map((resource) => (
              <div key={resource.label}>
                <div className="flex justify-between text-sm mb-2">
                  <span className="text-[var(--color-text-secondary)]">{resource.label}</span>
                  <span className="text-[var(--color-text-primary)] font-medium">
                    {resource.value}%
                  </span>
                </div>
                <div className="h-2 bg-[var(--color-primary-dark)] rounded-full overflow-hidden">
                  <div
                    className={`h-full bg-gradient-to-r from-[#7BA4D0] to-[#E7F0FA] rounded-full transition-all duration-500`}
                    style={{ width: `${resource.value}%` }}
                  />
                </div>
              </div>
            ))}
          </CardContent>
        </Card>

        {/* System Health */}
        <Card>
          <CardHeader>
            <CardTitle>System Health</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex items-center justify-between p-3 rounded-lg bg-[var(--color-primary-dark)]/50">
              <span className="text-sm text-[var(--color-text-secondary)]">Active Users</span>
              <Badge variant="success">{systemStats.activeUsers}</Badge>
            </div>
            <div className="flex items-center justify-between p-3 rounded-lg bg-[var(--color-primary-dark)]/50">
              <span className="text-sm text-[var(--color-text-secondary)]">Avg Load</span>
              <Badge variant="primary">0.45</Badge>
            </div>
            <div className="flex items-center justify-between p-3 rounded-lg bg-[var(--color-primary-dark)]/50">
              <span className="text-sm text-[var(--color-text-secondary)]">Uptime</span>
              <Badge variant="success">99.9%</Badge>
            </div>
            <div className="flex items-center justify-between p-3 rounded-lg bg-[var(--color-primary-dark)]/50">
              <span className="text-sm text-[var(--color-text-secondary)]">Services</span>
              <Badge variant="success">{services.filter(s => s.status === 'running').length}/{services.length}</Badge>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Services & Alerts */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Service Status */}
        <Card>
          <CardHeader>
            <div className="flex items-center justify-between">
              <CardTitle>Service Status</CardTitle>
              <Button variant="ghost" size="sm">Manage</Button>
            </div>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              {services.map((service) => (
                <div
                  key={service.name}
                  className="flex items-center justify-between p-3 rounded-lg bg-[var(--color-primary-dark)]/50 hover:bg-[var(--color-primary-dark)]/70 transition-colors"
                >
                  <div className="flex items-center gap-3">
                    {service.status === 'running' ? (
                      <CheckCircle size={20} className="text-[var(--color-success)]" />
                    ) : (
                      <AlertCircle size={20} className="text-[var(--color-error)]" />
                    )}
                    <div>
                      <span className="text-sm text-[var(--color-text-primary)] font-medium">
                        {service.name}
                      </span>
                      <p className="text-xs text-[var(--color-text-muted)]">
                        {service.uptime} uptime • {service.load} load
                      </p>
                    </div>
                  </div>
                  <StatusBadge status={service.status} />
                </div>
              ))}
            </div>
          </CardContent>
        </Card>

        {/* Recent Alerts */}
        <Card>
          <CardHeader>
            <div className="flex items-center justify-between">
              <CardTitle>Recent Alerts</CardTitle>
              <Button variant="ghost" size="sm">View All</Button>
            </div>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              {alerts.map((alert) => (
                <div
                  key={alert.id}
                  className="flex items-start gap-3 p-3 rounded-lg bg-[var(--color-primary-dark)]/50"
                >
                  {alert.type === 'warning' && (
                    <AlertCircle size={18} className="text-yellow-400 flex-shrink-0 mt-0.5" />
                  )}
                  {alert.type === 'info' && (
                    <Shield size={18} className="text-blue-400 flex-shrink-0 mt-0.5" />
                  )}
                  {alert.type === 'success' && (
                    <CheckCircle size={18} className="text-green-400 flex-shrink-0 mt-0.5" />
                  )}
                  <div className="flex-1 min-w-0">
                    <p className="text-sm text-[var(--color-text-primary)]">{alert.message}</p>
                    <p className="text-xs text-[var(--color-text-muted)] mt-1">{alert.time}</p>
                  </div>
                </div>
              ))}
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
            {[
              { icon: <Users size={24} />, label: 'Create User', path: '/admin/users' },
              { icon: <Store size={24} />, label: 'Add Reseller', path: '/admin/resellers' },
              { icon: <Server size={24} />, label: 'View Logs', path: '/admin/logs' },
              { icon: <Shield size={24} />, label: 'Security', path: '/admin/security' },
            ].map((action, index) => (
              <button
                key={index}
                onClick={() => window.location.href = action.path}
                className="p-4 rounded-lg bg-[var(--color-primary-dark)]/50 hover:bg-[var(--color-primary-dark)] transition-colors text-left group"
              >
                <div className="text-[#E7F0FA] mb-2 transform group-hover:scale-110 transition-transform">
                  {action.icon}
                </div>
                <p className="text-sm font-medium text-[var(--color-text-primary)]">
                  {action.label}
                </p>
              </button>
            ))}
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
