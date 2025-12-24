import { useState, useEffect } from 'react';
import {
  Users,
  Zap,
  HardDrive,
  Globe,
  Database,
  TrendingUp,
  AlertCircle,
  Activity,
} from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle } from '../../components/ui/Card';

interface StatCardProps {
  title: string;
  value: string | number;
  icon: React.ReactNode;
  subtitle?: string;
  color?: 'primary' | 'success' | 'warning' | 'error';
}

function StatCard({ title, value, icon, subtitle, color = 'primary' }: StatCardProps) {
  const colorClasses = {
    primary: 'from-[#7BA4D0]/30 to-[#E7F0FA]/20',
    success: 'from-green-500/30 to-emerald-500/20',
    warning: 'from-yellow-500/30 to-orange-500/20',
    error: 'from-red-500/30 to-rose-500/20',
  };

  return (
    <Card hover>
      <CardContent className="flex items-start justify-between">
        <div>
          <p className="text-sm font-medium text-[var(--color-text-secondary)]">{title}</p>
          <p className="text-3xl font-bold text-[var(--color-text-primary)] mt-2">{value}</p>
          {subtitle && (
            <p className="text-sm text-[var(--color-text-muted)] mt-1">{subtitle}</p>
          )}
        </div>
        <div className={`p-3 rounded-xl bg-gradient-to-br ${colorClasses[color]}`}>
          {icon}
        </div>
      </CardContent>
    </Card>
  );
}

export function ResellerDashboardPage() {
  const [stats, setStats] = useState({
    totalCustomers: 0,
    activeCustomers: 0,
    totalDomains: 0,
    totalDatabases: 0,
    resourcePool: {
      cpu: { allocated: 0, used: 0 },
      ram: { allocated: 0, used: 0 },
      disk: { allocated: 0, used: 0 },
      bandwidth: { allocated: 0, used: 0 },
    },
  });

  useEffect(() => {
    // Load reseller stats
    // TODO: Replace with actual API calls
    setStats({
      totalCustomers: 45,
      activeCustomers: 42,
      totalDomains: 128,
      totalDatabases: 89,
      resourcePool: {
        cpu: { allocated: 100, used: 65 },
        ram: { allocated: 512, used: 342 },
        disk: { allocated: 1000, used: 456 },
        bandwidth: { allocated: 5000, used: 1234 },
      },
    });
  }, []);

  const getPercentage = (used: number, allocated: number) => {
    return allocated > 0 ? Math.round((used / allocated) * 100) : 0;
  };

  return (
    <div className="space-y-6">
      {/* Page Header */}
      <div>
        <h1 className="text-2xl font-bold text-[var(--color-text-primary)]">Reseller Dashboard</h1>
        <p className="text-[var(--color-text-secondary)] mt-1">
          Manage your customers and resource allocation
        </p>
      </div>

      {/* Stats Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        <StatCard
          title="Total Customers"
          value={stats.totalCustomers}
          icon={<Users size={24} className="text-[#E7F0FA]" />}
          subtitle={`${stats.activeCustomers} active`}
        />
        <StatCard
          title="Customer Domains"
          value={stats.totalDomains}
          icon={<Globe size={24} className="text-[#E7F0FA]" />}
        />
        <StatCard
          title="Customer Databases"
          value={stats.totalDatabases}
          icon={<Database size={24} className="text-[#E7F0FA]" />}
        />
        <StatCard
          title="Resource Usage"
          value={`${getPercentage(
            stats.resourcePool.disk.used,
            stats.resourcePool.disk.allocated
          )}%`}
          icon={<Activity size={24} className="text-[#E7F0FA]" />}
          subtitle="Pool utilization"
        />
      </div>

      {/* Resource Pool Overview */}
      <Card>
        <CardHeader>
          <CardTitle>Resource Pool Allocation</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div>
            <div className="flex justify-between text-sm mb-2">
              <span className="text-[var(--color-text-secondary)]">CPU Cores</span>
              <span className="text-[var(--color-text-primary)] font-medium">
                {stats.resourcePool.cpu.used} / {stats.resourcePool.cpu.allocated}
              </span>
            </div>
            <div className="h-2 bg-[var(--color-primary-dark)] rounded-full overflow-hidden">
              <div
                className="h-full bg-gradient-to-r from-[#7BA4D0] to-[#E7F0FA] rounded-full transition-all duration-500"
                style={{
                  width: `${getPercentage(stats.resourcePool.cpu.used, stats.resourcePool.cpu.allocated)}%`,
                }}
              />
            </div>
          </div>
          <div>
            <div className="flex justify-between text-sm mb-2">
              <span className="text-[var(--color-text-secondary)]">RAM (GB)</span>
              <span className="text-[var(--color-text-primary)] font-medium">
                {stats.resourcePool.ram.used} / {stats.resourcePool.ram.allocated} GB
              </span>
            </div>
            <div className="h-2 bg-[var(--color-primary-dark)] rounded-full overflow-hidden">
              <div
                className="h-full bg-gradient-to-r from-[#7BA4D0] to-[#E7F0FA] rounded-full transition-all duration-500"
                style={{
                  width: `${getPercentage(stats.resourcePool.ram.used, stats.resourcePool.ram.allocated)}%`,
                }}
              />
            </div>
          </div>
          <div>
            <div className="flex justify-between text-sm mb-2">
              <span className="text-[var(--color-text-secondary)]">Disk Space (GB)</span>
              <span className="text-[var(--color-text-primary)] font-medium">
                {stats.resourcePool.disk.used} / {stats.resourcePool.disk.allocated} GB
              </span>
            </div>
            <div className="h-2 bg-[var(--color-primary-dark)] rounded-full overflow-hidden">
              <div
                className="h-full bg-gradient-to-r from-[#7BA4D0] to-[#E7F0FA] rounded-full transition-all duration-500"
                style={{
                  width: `${getPercentage(stats.resourcePool.disk.used, stats.resourcePool.disk.allocated)}%`,
                }}
              />
            </div>
          </div>
          <div>
            <div className="flex justify-between text-sm mb-2">
              <span className="text-[var(--color-text-secondary)]">Bandwidth (GB)</span>
              <span className="text-[var(--color-text-primary)] font-medium">
                {stats.resourcePool.bandwidth.used} / {stats.resourcePool.bandwidth.allocated} GB
              </span>
            </div>
            <div className="h-2 bg-[var(--color-primary-dark)] rounded-full overflow-hidden">
              <div
                className="h-full bg-gradient-to-r from-[#7BA4D0] to-[#E7F0FA] rounded-full transition-all duration-500"
                style={{
                  width: `${getPercentage(
                    stats.resourcePool.bandwidth.used,
                    stats.resourcePool.bandwidth.allocated
                  )}%`,
                }}
              />
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Quick Actions */}
      <Card>
        <CardHeader>
          <CardTitle>Quick Actions</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            <button className="p-4 rounded-lg bg-[var(--color-primary-dark)]/50 hover:bg-[var(--color-primary-dark)] transition-colors text-left">
              <Users size={24} className="text-[#E7F0FA] mb-2" />
              <p className="text-sm font-medium text-[var(--color-text-primary)]">Add Customer</p>
            </button>
            <button className="p-4 rounded-lg bg-[var(--color-primary-dark)]/50 hover:bg-[var(--color-primary-dark)] transition-colors text-left">
              <Zap size={24} className="text-[#E7F0FA] mb-2" />
              <p className="text-sm font-medium text-[var(--color-text-primary)]">Allocate Resources</p>
            </button>
            <button className="p-4 rounded-lg bg-[var(--color-primary-dark)]/50 hover:bg-[var(--color-primary-dark)] transition-colors text-left">
              <Globe size={24} className="text-[#E7F0FA] mb-2" />
              <p className="text-sm font-medium text-[var(--color-text-primary)]">Manage Domains</p>
            </button>
            <button className="p-4 rounded-lg bg-[var(--color-primary-dark)]/50 hover:bg-[var(--color-primary-dark)] transition-colors text-left">
              <TrendingUp size={24} className="text-[#E7F0FA] mb-2" />
              <p className="text-sm font-medium text-[var(--color-text-primary)]">View Reports</p>
            </button>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
