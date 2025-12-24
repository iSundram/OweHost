import { useState, useEffect } from 'react';
import {
  Globe,
  Database,
  HardDrive,
  TrendingUp,
  Activity,
  Shield,
  Clock,
  ArrowUpRight,
  Plus,
  Zap,
} from 'lucide-react';
import { 
  Card, 
  CardContent, 
  CardHeader, 
  CardTitle,
  Button,
  Badge,
  DashboardSkeleton
} from '../../components/ui';
import { domainService, databaseService } from '../../api/services';

interface StatCardProps {
  title: string;
  value: string | number;
  icon: React.ReactNode;
  trend?: {
    value: number;
    isPositive: boolean;
  };
  description?: string;
  onClick?: () => void;
}

function StatCard({ title, value, icon, trend, description, onClick }: StatCardProps) {
  return (
    <Card hover className={onClick ? 'cursor-pointer' : ''} onClick={onClick}>
      <CardContent className="flex items-start justify-between">
        <div className="flex-1">
          <p className="text-sm font-medium text-[var(--color-text-secondary)]">{title}</p>
          <p className="text-3xl font-bold text-[var(--color-text-primary)] mt-2">{value}</p>
          {trend && (
            <div className="flex items-center gap-1 mt-2">
              {trend.isPositive ? (
                <ArrowUpRight size={16} className="text-[var(--color-success)]" />
              ) : (
                <ArrowUpRight size={16} className="text-[var(--color-error)] rotate-90" />
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
  label: string;
  value: number;
  max: number;
  color?: 'primary' | 'success' | 'warning' | 'error';
}

function ProgressBar({ label, value, max, color = 'primary' }: ProgressBarProps) {
  const percentage = Math.min(Math.round((value / max) * 100), 100);
  
  const colors = {
    primary: 'bg-gradient-to-r from-[#7BA4D0] to-[#E7F0FA]',
    success: 'bg-[var(--color-success)]',
    warning: 'bg-[var(--color-warning)]',
    error: 'bg-[var(--color-error)]',
  };

  const getColor = () => {
    if (percentage >= 90) return 'error';
    if (percentage >= 75) return 'warning';
    return color;
  };

  const selectedColor = getColor();

  return (
    <div className="space-y-2">
      <div className="flex justify-between text-sm">
        <span className="text-[var(--color-text-secondary)]">{label}</span>
        <span className="text-[var(--color-text-primary)] font-medium">{percentage}%</span>
      </div>
      <div className="h-2 bg-[var(--color-primary-dark)] rounded-full overflow-hidden">
        <div
          className={`h-full ${colors[selectedColor]} rounded-full transition-all duration-500`}
          style={{ width: `${percentage}%` }}
        />
      </div>
      <p className="text-xs text-[var(--color-text-muted)]">
        {value.toFixed(1)} {typeof max === 'number' ? `of ${max}` : ''} GB used
      </p>
    </div>
  );
}

interface QuickAction {
  icon: React.ReactNode;
  label: string;
  onClick: () => void;
}

export function UserDashboardEnhanced() {
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
    sslCertificates: domains.filter(d => d.ssl).length,
  };

  const quickActions: QuickAction[] = [
    {
      icon: <Globe size={24} className="text-[#E7F0FA]" />,
      label: 'Add Domain',
      onClick: () => console.log('Add Domain'),
    },
    {
      icon: <Database size={24} className="text-[#E7F0FA]" />,
      label: 'Create Database',
      onClick: () => console.log('Create Database'),
    },
    {
      icon: <Shield size={24} className="text-[#E7F0FA]" />,
      label: 'Install SSL',
      onClick: () => console.log('Install SSL'),
    },
    {
      icon: <Zap size={24} className="text-[#E7F0FA]" />,
      label: 'View Stats',
      onClick: () => console.log('View Stats'),
    },
  ];

  if (isLoading) {
    return <DashboardSkeleton />;
  }

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
          onClick={() => window.location.href = '/user/domains'}
        />
        <StatCard
          title="Databases"
          value={stats.totalDatabases}
          icon={<Database size={24} className="text-[#E7F0FA]" />}
          onClick={() => window.location.href = '/user/databases'}
        />
        <StatCard
          title="SSL Certificates"
          value={stats.sslCertificates}
          icon={<Shield size={24} className="text-[var(--color-success)]" />}
          description="Auto-renewed"
        />
        <StatCard
          title="Uptime"
          value="99.9%"
          icon={<Activity size={24} className="text-[var(--color-success)]" />}
          trend={{ value: 0.1, isPositive: true }}
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
              <ProgressBar value={8.5} max={10} label="Inodes" color="error" />
            </div>
            <div className="pt-4 border-t border-[var(--color-border-light)]">
              <Button variant="outline" size="sm" className="w-full">
                Upgrade Resources
              </Button>
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
                <span className="text-sm text-[var(--color-text-secondary)]">Backups</span>
              </div>
              <Badge variant="success" size="sm">12</Badge>
            </div>
            <div className="flex items-center justify-between p-3 rounded-lg bg-[var(--color-primary-dark)]/50">
              <div className="flex items-center gap-3">
                <Clock size={20} className="text-[#E7F0FA]" />
                <span className="text-sm text-[var(--color-text-secondary)]">Cron Jobs</span>
              </div>
              <Badge variant="primary" size="sm">5</Badge>
            </div>
            <div className="flex items-center justify-between p-3 rounded-lg bg-[var(--color-primary-dark)]/50">
              <div className="flex items-center gap-3">
                <Globe size={20} className="text-[#E7F0FA]" />
                <span className="text-sm text-[var(--color-text-secondary)]">Subdomains</span>
              </div>
              <Badge variant="info" size="sm">8</Badge>
            </div>
            <div className="flex items-center justify-between p-3 rounded-lg bg-[var(--color-primary-dark)]/50">
              <div className="flex items-center gap-3">
                <TrendingUp size={20} className="text-[var(--color-success)]" />
                <span className="text-sm text-[var(--color-text-secondary)]">Visitors</span>
              </div>
              <span className="font-semibold text-[var(--color-success)]">+24%</span>
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
            {quickActions.map((action, index) => (
              <button
                key={index}
                onClick={action.onClick}
                className="p-4 rounded-lg bg-[var(--color-primary-dark)]/50 hover:bg-[var(--color-primary-dark)] transition-colors text-left group"
              >
                <div className="mb-3 transform group-hover:scale-110 transition-transform">
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

      {/* Recent Activity */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <CardTitle>Recent Activity</CardTitle>
            <Button variant="ghost" size="sm">View All</Button>
          </div>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            {[
              { action: 'Domain added', target: 'example.com', time: '2 minutes ago', icon: <Globe size={16} /> },
              { action: 'Database created', target: 'wordpress_db', time: '15 minutes ago', icon: <Database size={16} /> },
              { action: 'SSL installed', target: 'mysite.com', time: '1 hour ago', icon: <Shield size={16} /> },
              { action: 'Backup completed', target: 'Full backup', time: '2 hours ago', icon: <HardDrive size={16} /> },
            ].map((activity, index) => (
              <div
                key={index}
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
