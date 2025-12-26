import { useState, useEffect } from 'react';
import { Package } from 'lucide-react';
import { Card, CardContent, CardHeader } from '../../components/ui';
import { packageService, type Package as PackageType } from '../../api/services';

export function ResellerPackagesPage() {
  const [packages, setPackages] = useState<PackageType[]>([]);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    loadPackages();
  }, []);

  const loadPackages = async () => {
    setIsLoading(true);
    try {
      const data = await packageService.list();
      setPackages(data);
    } catch (error) {
      console.error('Failed to load packages:', error);
    } finally {
      setIsLoading(false);
    }
  };

  const formatBytes = (bytes: number) => {
    if (bytes === -1) return 'Unlimited';
    if (bytes < 1024) return `${bytes} MB`;
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} GB`;
    return `${(bytes / (1024 * 1024)).toFixed(1)} TB`;
  };

  const formatNumber = (num: number) => {
    if (num === -1) return 'Unlimited';
    return num.toLocaleString();
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="text-[var(--color-text-secondary)]">Loading packages...</div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Page Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-[var(--color-text-primary)]">Available Packages</h1>
          <p className="text-[var(--color-text-secondary)] mt-1">
            View available hosting packages for your customers
          </p>
        </div>
      </div>

      {/* Packages Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        {packages.map((pkg) => (
          <Card key={pkg.name}>
            <CardHeader className="pb-3">
              <div className="flex items-center gap-2">
                <div className="p-2 rounded-lg bg-gradient-to-br from-[#7BA4D0]/30 to-[#E7F0FA]/20">
                  <Package size={20} className="text-[#E7F0FA]" />
                </div>
                <div>
                  <h3 className="font-semibold text-[var(--color-text-primary)]">
                    {pkg.display_name}
                  </h3>
                  <p className="text-xs text-[var(--color-text-secondary)]">{pkg.name}</p>
                </div>
              </div>
            </CardHeader>
            <CardContent>
              <div className="space-y-2 text-sm">
                <div className="flex justify-between">
                  <span className="text-[var(--color-text-secondary)]">Disk:</span>
                  <span className="font-medium text-[var(--color-text-primary)]">
                    {formatBytes(pkg.limits.disk_mb)}
                  </span>
                </div>
                <div className="flex justify-between">
                  <span className="text-[var(--color-text-secondary)]">CPU:</span>
                  <span className="font-medium text-[var(--color-text-primary)]">
                    {pkg.limits.cpu_percent === -1 ? 'Unlimited' : `${pkg.limits.cpu_percent}%`}
                  </span>
                </div>
                <div className="flex justify-between">
                  <span className="text-[var(--color-text-secondary)]">RAM:</span>
                  <span className="font-medium text-[var(--color-text-primary)]">
                    {formatBytes(pkg.limits.ram_mb)}
                  </span>
                </div>
                <div className="flex justify-between">
                  <span className="text-[var(--color-text-secondary)]">Domains:</span>
                  <span className="font-medium text-[var(--color-text-primary)]">
                    {formatNumber(pkg.limits.domains)}
                  </span>
                </div>
                <div className="flex justify-between">
                  <span className="text-[var(--color-text-secondary)]">Databases:</span>
                  <span className="font-medium text-[var(--color-text-primary)]">
                    {formatNumber(pkg.limits.databases)}
                  </span>
                </div>
              </div>
              <div className="mt-4 pt-4 border-t border-[var(--color-border)]">
                <p className="text-xs text-[var(--color-text-secondary)] line-clamp-2">
                  {pkg.description}
                </p>
              </div>
            </CardContent>
          </Card>
        ))}
      </div>
    </div>
  );
}
