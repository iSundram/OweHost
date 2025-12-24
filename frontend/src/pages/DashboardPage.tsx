import { useState, useEffect } from 'react';
import {
  Globe,
  Database,
  Users,
  HardDrive,
  TrendingUp,
  Activity,
  Shield,
  Clock,
  ArrowUpRight,
  ArrowDownRight,
} from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle } from '../components/ui/Card';
import { domainService, databaseService, userService } from '../api/services';

interface StatCardProps {
  title: string;
  value: string | number;
  icon: React.ReactNode;
  trend?: {
    value: number;
    isPositive: boolean;
  };
  description?: string;
}

function StatCard({ title, value, icon, trend, description }: StatCardProps) {
  return (
    <Card hover>
      <CardContent className="flex items-start justify-between">
        <div>
          <p className="text-sm font-medium text-[var(--color-text-secondary)]">{title}</p>
          <p className="text-3xl font-bold text-[var(--color-text-primary)] mt-2">{value}</p>
          {trend && (
            <div className="flex items-center gap-1 mt-2">
              {trend.isPositive ? (
                <ArrowUpRight size={16} className="text-[var(--color-success)]" />
              ) : (
                <ArrowDownRight size={16} className="text-[var(--color-error)]" />
              )}
              <span
                className={`text-sm font-medium ${
                  trend.isPositive ? 'text-[var(--color-success)]' : 'text-[var(--color-error)]'
                }`}
              >
                {trend.value}%
              </span>
              <span className="text-sm text-[var(--color-text-muted)]">from last month</span>
            </div>
          )}
          {description && (
            <p className="text-sm text-[var(--color-text-muted)] mt-2">{description}</p>
          )}
        </div>
        <div className="p-3 rounded-xl bg-gradient-to-br from-[#7BA4D0]/30 to-[#E7F0FA]/20">
          {icon}
        </div>
      </CardContent>
    </Card>
  );
}

interface ProgressBarProps {
  value: number;
  max: number;
  label: string;
  color?: 'primary' | 'success' | 'warning' | 'error';
}

function ProgressBar({ value, max, label, color = 'primary' }: ProgressBarProps) {
  const percentage = Math.round((value / max) * 100);
  const colors = {
    primary: 'bg-gradient-to-r from-[#7BA4D0] to-[#E7F0FA]',
    success: 'bg-[var(--color-success)]',
    warning: 'bg-[var(--color-warning)]',
    error: 'bg-[var(--color-error)]',
  };

  return (
    <div className="space-y-2">
      <div className="flex justify-between text-sm">
        <span className="text-[var(--color-text-secondary)]">{label}</span>
        <span className="text-[var(--color-text-primary)] font-medium">{percentage}%</span>
      </div>
      <div className="h-2 bg-[var(--color-primary-dark)] rounded-full overflow-hidden">
        <div
          className={`h-full ${colors[color]} rounded-full transition-all duration-500`}
          style={{ width: `${percentage}%` }}
        />
      </div>
      <p className="text-xs text-[var(--color-text-muted)]">
        {value.toFixed(1)} GB of {max} GB used
      </p>
    </div>
  );
}

interface RecentActivityItem {
  id: string;
  action: string;
  target: string;
  time: string;
  icon: React.ReactNode;
}

const recentActivities: RecentActivityItem[] = [
  {
    id: '1',
    action: 'Domain added',
    target: 'example.com',
    time: '2 minutes ago',
    icon: <Globe size={16} className="text-[#E7F0FA]" />,
  },
  {
    id: '2',
    action: 'Database created',
    target: 'wordpress_db',
    time: '15 minutes ago',
    icon: <Database size={16} className="text-[#E7F0FA]" />,
  },
  {
    id: '3',
    action: 'SSL installed',
    target: 'mysite.com',
    time: '1 hour ago',
    icon: <Shield size={16} className="text-[var(--color-success)]" />,
  },
  {
    id: '4',
    action: 'Cron job executed',
    target: 'backup.sh',
    time: '2 hours ago',
    icon: <Clock size={16} className="text-[var(--color-warning)]" />,
  },
];

export function DashboardPage() {
  const [domains, setDomains] = useState<any[]>([]);
  const [databases, setDatabases] = useState<any[]>([]);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    const loadData = async () => {
      try {
        const [domainsRes, databasesRes] = await Promise.all([
          domainService.list().catch(() => []),
          databaseService.list().catch(() => []),
        ]);
        setDomains(domainsRes || []);
        setDatabases(databasesRes || []);
      } catch (error) {
        console.error('Failed to load dashboard data:', error);
      } finally {
        setIsLoading(false);
      }
    };
    loadData();
  }, []);

  const stats = {
    totalDomains: domains.length,
    activeDomains: domains.filter(d => d.status === 'active').length,
    totalDatabases: databases.length,
  };

  return (
    <div className="space-y-6">
      {/* Page Header */}
      <div>
        <h1 className="text-2xl font-bold text-[var(--color-text-primary)]">Dashboard</h1>
        <p className="text-[var(--color-text-secondary)] mt-1">
          Welcome back! Here's an overview of your hosting account.
        </p>
      </div>

      {/* Stats Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        <StatCard
          title="Total Domains"
          value={stats.totalDomains}
          icon={<Globe size={24} className="text-[#E7F0FA]" />}
          description={`${stats.activeDomains} active`}
        />
        <StatCard
          title="Databases"
          value={stats.totalDatabases}
          icon={<Database size={24} className="text-[#E7F0FA]" />}
          description={`${databases.filter(d => d.type === 'mysql').length} MySQL, ${databases.filter(d => d.type === 'postgresql').length} PostgreSQL`}
        />
        <StatCard
          title="Active Services"
          value={8}
          icon={<Activity size={24} className="text-[var(--color-success)]" />}
          description="All services running"
        />
      </div>

      {/* Main Content Grid */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Resource Usage */}
        <Card className="lg:col-span-2">
          <CardHeader>
            <CardTitle>Resource Usage</CardTitle>
          </CardHeader>
          <CardContent className="space-y-6">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              <ProgressBar value={45.2} max={100} label="Disk Space" color="primary" />
              <ProgressBar value={2.8} max={10} label="Bandwidth" color="success" />
              <ProgressBar value={512} max={1024} label="Memory" color="warning" />
              <ProgressBar value={8.5} max={10} label="CPU Usage" color="error" />
            </div>
          </CardContent>
        </Card>

        {/* Quick Stats */}
        <Card>
          <CardHeader>
            <CardTitle>Quick Stats</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex items-center justify-between p-3 rounded-lg bg-[var(--color-primary-dark)]/50">
              <div className="flex items-center gap-3">
                <HardDrive size={20} className="text-[#E7F0FA]" />
                <span className="text-sm text-[var(--color-text-secondary)]">Inodes</span>
              </div>
              <span className="font-semibold text-[var(--color-text-primary)]">45,231</span>
            </div>
            <div className="flex items-center justify-between p-3 rounded-lg bg-[var(--color-primary-dark)]/50">
              <div className="flex items-center gap-3">
                <Globe size={20} className="text-[#E7F0FA]" />
                <span className="text-sm text-[var(--color-text-secondary)]">Subdomains</span>
              </div>
              <span className="font-semibold text-[var(--color-text-primary)]">8</span>
            </div>
            <div className="flex items-center justify-between p-3 rounded-lg bg-[var(--color-primary-dark)]/50">
              <div className="flex items-center gap-3">
                <Shield size={20} className="text-[var(--color-success)]" />
                <span className="text-sm text-[var(--color-text-secondary)]">SSL Certs</span>
              </div>
              <span className="font-semibold text-[var(--color-text-primary)]">12</span>
            </div>
            <div className="flex items-center justify-between p-3 rounded-lg bg-[var(--color-primary-dark)]/50">
              <div className="flex items-center gap-3">
                <TrendingUp size={20} className="text-[var(--color-success)]" />
                <span className="text-sm text-[var(--color-text-secondary)]">Uptime</span>
              </div>
              <span className="font-semibold text-[var(--color-success)]">99.9%</span>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Recent Activity */}
      <Card>
        <CardHeader>
          <CardTitle>Recent Activity</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            {recentActivities.map((activity) => (
              <div
                key={activity.id}
                className="flex items-center gap-4 p-3 rounded-lg hover:bg-[var(--color-primary-dark)]/30 transition-colors"
              >
                <div className="p-2 rounded-lg bg-[var(--color-primary-dark)]">
                  {activity.icon}
                </div>
                <div className="flex-1">
                  <p className="text-sm text-[var(--color-text-primary)]">
                    {activity.action}{' '}
                    <span className="font-medium text-[#E7F0FA]">{activity.target}</span>
                  </p>
                  <p className="text-xs text-[var(--color-text-muted)]">{activity.time}</p>
                </div>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
