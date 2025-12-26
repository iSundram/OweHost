import { useState, useEffect } from 'react';
import { Package, Plus, Edit, Trash2, CheckCircle, XCircle } from 'lucide-react';
import { Button, Card, CardContent, CardHeader, Modal, Input } from '../../components/ui';
import { packageService, type Package as PackageType } from '../../api/services';

export function AdminPackagesPage() {
  const [packages, setPackages] = useState<PackageType[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [selectedPackage, setSelectedPackage] = useState<PackageType | null>(null);
  const [isViewModalOpen, setIsViewModalOpen] = useState(false);

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
          <h1 className="text-2xl font-bold text-[var(--color-text-primary)]">Packages & Plans</h1>
          <p className="text-[var(--color-text-secondary)] mt-1">
            Manage hosting packages and resource limits
          </p>
        </div>
      </div>

      {/* Packages Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        {packages.map((pkg) => (
          <Card
            key={pkg.name}
            className="cursor-pointer hover:border-[#7BA4D0] transition-colors"
            onClick={() => {
              setSelectedPackage(pkg);
              setIsViewModalOpen(true);
            }}
          >
            <CardHeader className="pb-3">
              <div className="flex items-center justify-between">
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

      {/* View Package Modal */}
      {selectedPackage && (
        <Modal
          isOpen={isViewModalOpen}
          onClose={() => {
            setIsViewModalOpen(false);
            setSelectedPackage(null);
          }}
          title={`${selectedPackage.display_name} Package Details`}
          size="lg"
        >
          <div className="space-y-6">
            <div>
              <h3 className="text-sm font-medium text-[var(--color-text-secondary)] mb-2">
                Description
              </h3>
              <p className="text-[var(--color-text-primary)]">{selectedPackage.description}</p>
            </div>

            <div>
              <h3 className="text-sm font-medium text-[var(--color-text-secondary)] mb-3">
                Resource Limits
              </h3>
              <div className="grid grid-cols-2 gap-4">
                <div className="p-3 rounded-lg bg-[var(--color-primary-dark)]">
                  <div className="text-xs text-[var(--color-text-secondary)]">Disk Space</div>
                  <div className="text-lg font-semibold text-[var(--color-text-primary)]">
                    {formatBytes(selectedPackage.limits.disk_mb)}
                  </div>
                </div>
                <div className="p-3 rounded-lg bg-[var(--color-primary-dark)]">
                  <div className="text-xs text-[var(--color-text-secondary)]">CPU</div>
                  <div className="text-lg font-semibold text-[var(--color-text-primary)]">
                    {selectedPackage.limits.cpu_percent === -1
                      ? 'Unlimited'
                      : `${selectedPackage.limits.cpu_percent}%`}
                  </div>
                </div>
                <div className="p-3 rounded-lg bg-[var(--color-primary-dark)]">
                  <div className="text-xs text-[var(--color-text-secondary)]">RAM</div>
                  <div className="text-lg font-semibold text-[var(--color-text-primary)]">
                    {formatBytes(selectedPackage.limits.ram_mb)}
                  </div>
                </div>
                <div className="p-3 rounded-lg bg-[var(--color-primary-dark)]">
                  <div className="text-xs text-[var(--color-text-secondary)]">Bandwidth</div>
                  <div className="text-lg font-semibold text-[var(--color-text-primary)]">
                    {formatNumber(selectedPackage.limits.bandwidth_gb)} GB
                  </div>
                </div>
                <div className="p-3 rounded-lg bg-[var(--color-primary-dark)]">
                  <div className="text-xs text-[var(--color-text-secondary)]">Domains</div>
                  <div className="text-lg font-semibold text-[var(--color-text-primary)]">
                    {formatNumber(selectedPackage.limits.domains)}
                  </div>
                </div>
                <div className="p-3 rounded-lg bg-[var(--color-primary-dark)]">
                  <div className="text-xs text-[var(--color-text-secondary)]">Subdomains</div>
                  <div className="text-lg font-semibold text-[var(--color-text-primary)]">
                    {formatNumber(selectedPackage.limits.subdomains)}
                  </div>
                </div>
                <div className="p-3 rounded-lg bg-[var(--color-primary-dark)]">
                  <div className="text-xs text-[var(--color-text-secondary)]">Databases</div>
                  <div className="text-lg font-semibold text-[var(--color-text-primary)]">
                    {formatNumber(selectedPackage.limits.databases)}
                  </div>
                </div>
                <div className="p-3 rounded-lg bg-[var(--color-primary-dark)]">
                  <div className="text-xs text-[var(--color-text-secondary)]">Email Accounts</div>
                  <div className="text-lg font-semibold text-[var(--color-text-primary)]">
                    {formatNumber(selectedPackage.limits.email_accounts)}
                  </div>
                </div>
                <div className="p-3 rounded-lg bg-[var(--color-primary-dark)]">
                  <div className="text-xs text-[var(--color-text-secondary)]">FTP Accounts</div>
                  <div className="text-lg font-semibold text-[var(--color-text-primary)]">
                    {formatNumber(selectedPackage.limits.ftp_accounts)}
                  </div>
                </div>
                <div className="p-3 rounded-lg bg-[var(--color-primary-dark)]">
                  <div className="text-xs text-[var(--color-text-secondary)]">Inodes</div>
                  <div className="text-lg font-semibold text-[var(--color-text-primary)]">
                    {formatNumber(selectedPackage.limits.inodes)}
                  </div>
                </div>
              </div>
            </div>

            <div>
              <h3 className="text-sm font-medium text-[var(--color-text-secondary)] mb-3">
                Features
              </h3>
              <div className="grid grid-cols-2 gap-2">
                {Object.entries(selectedPackage.features).map(([feature, enabled]) => (
                  <div
                    key={feature}
                    className="flex items-center gap-2 p-2 rounded-lg bg-[var(--color-primary-dark)]"
                  >
                    {enabled ? (
                      <CheckCircle size={16} className="text-[var(--color-success)]" />
                    ) : (
                      <XCircle size={16} className="text-[var(--color-text-muted)]" />
                    )}
                    <span className="text-sm text-[var(--color-text-primary)] capitalize">
                      {feature.replace(/_/g, ' ')}
                    </span>
                  </div>
                ))}
              </div>
            </div>
          </div>
        </Modal>
      )}
    </div>
  );
}
