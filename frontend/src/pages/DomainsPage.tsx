import React, { useState, useEffect } from 'react';
import {
  Globe,
  Plus,
  Search,
  MoreVertical,
  Shield,
  ExternalLink,
  Trash2,
  Settings,
  CheckCircle,
  Clock,
  AlertCircle,
} from 'lucide-react';
import { Button, Input, Card, CardContent, CardHeader } from '../components/ui';
import { CreateDomainModal } from '../components/domains/CreateDomainModal';
import { DeleteDomainModal } from '../components/domains/DeleteDomainModal';
import { useToast } from '../context/ToastContext';
import type { Domain } from '../types';
import { domainService } from '../api/services';

const StatusBadge = ({ status }: { status: Domain['status'] }) => {
  const styles: Record<string, string> = {
    active: 'bg-[var(--color-success)]/10 text-[var(--color-success)] border-[var(--color-success)]/20',
    pending: 'bg-[var(--color-warning)]/10 text-[var(--color-warning)] border-[var(--color-warning)]/20',
    suspended: 'bg-[var(--color-error)]/10 text-[var(--color-error)] border-[var(--color-error)]/20',
  };

  const icons: Record<string, React.ReactNode> = {
    active: <CheckCircle size={14} />,
    pending: <Clock size={14} />,
    suspended: <AlertCircle size={14} />,
  };

  const safeStatus = status || 'pending';

  return (
    <span
      className={`inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-medium border ${styles[safeStatus] || styles.pending}`}
    >
      {icons[safeStatus] || icons.pending}
      {safeStatus.charAt(0).toUpperCase() + safeStatus.slice(1)}
    </span>
  );
};

const TypeBadge = ({ type }: { type: Domain['type'] }) => {
  const styles: Record<string, string> = {
    primary: 'bg-[#7BA4D0]/20 text-[#E7F0FA] border-[#7BA4D0]/30',
    addon: 'bg-[var(--color-info)]/10 text-[var(--color-info)] border-[var(--color-info)]/20',
    parked: 'bg-[var(--color-text-muted)]/10 text-[var(--color-text-muted)] border-[var(--color-text-muted)]/20',
    alias: 'bg-[var(--color-text-muted)]/10 text-[var(--color-text-muted)] border-[var(--color-text-muted)]/20',
  };

  const safeType = type || 'primary';

  return (
    <span
      className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium border ${styles[safeType] || styles.primary}`}
    >
      {safeType.charAt(0).toUpperCase() + safeType.slice(1)}
    </span>
  );
};

export function DomainsPage() {
  const [searchQuery, setSearchQuery] = useState('');
  const [domains, setDomains] = useState<Domain[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);
  const [isDeleteModalOpen, setIsDeleteModalOpen] = useState(false);
  const [selectedDomain, setSelectedDomain] = useState<Domain | null>(null);
  const { showToast } = useToast();

  const fetchDomains = async () => {
    try {
      const data = await domainService.list();
      setDomains(Array.isArray(data) ? data : []);
    } catch (error) {
      console.error('Failed to fetch domains:', error);
      showToast('error', 'Failed to load domains');
      setDomains([]);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchDomains();
  }, []);

  const handleCreateDomain = async (data: { name: string; type: string; document_root?: string }) => {
    try {
      await domainService.create(data);
      await fetchDomains();
      showToast('success', `Domain ${data.name} created successfully!`);
    } catch (error) {
      showToast('error', 'Failed to create domain');
      throw error;
    }
  };

  const handleDeleteDomain = async (domainId: string) => {
    try {
      await domainService.delete(domainId);
      await fetchDomains();
      showToast('success', 'Domain deleted successfully!');
    } catch (error) {
      showToast('error', 'Failed to delete domain');
      throw error;
    }
  };

  const openDeleteModal = (domain: Domain) => {
    setSelectedDomain(domain);
    setIsDeleteModalOpen(true);
  };

  const filteredDomains = domains.filter((domain) =>
    domain.name.toLowerCase().includes(searchQuery.toLowerCase())
  );

  return (
    <div className="space-y-6">
      {/* Page Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-[var(--color-text-primary)]">Domains</h1>
          <p className="text-[var(--color-text-secondary)] mt-1">
            Manage your domains and subdomains
          </p>
        </div>
        <Button leftIcon={<Plus size={18} />} onClick={() => setIsCreateModalOpen(true)}>
          Add Domain
        </Button>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <Card>
          <CardContent className="flex items-center gap-4 py-4">
            <div className="p-3 rounded-xl bg-gradient-to-br from-[#7BA4D0]/30 to-[#E7F0FA]/20">
              <Globe size={24} className="text-[#E7F0FA]" />
            </div>
            <div>
              <p className="text-2xl font-bold text-[var(--color-text-primary)]">{domains.length}</p>
              <p className="text-sm text-[var(--color-text-secondary)]">Total Domains</p>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="flex items-center gap-4 py-4">
            <div className="p-3 rounded-xl bg-[var(--color-success)]/10">
              <CheckCircle size={24} className="text-[var(--color-success)]" />
            </div>
            <div>
              <p className="text-2xl font-bold text-[var(--color-text-primary)]">
                {domains.filter((d) => d.status === 'active').length}
              </p>
              <p className="text-sm text-[var(--color-text-secondary)]">Active</p>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="flex items-center gap-4 py-4">
            <div className="p-3 rounded-xl bg-[var(--color-success)]/10">
              <Shield size={24} className="text-[var(--color-success)]" />
            </div>
            <div>
              <p className="text-2xl font-bold text-[var(--color-text-primary)]">
                {domains.filter((d) => d.validated).length}
              </p>
              <p className="text-sm text-[var(--color-text-secondary)]">SSL Secured</p>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Search and Filter */}
      <Card>
        <CardHeader>
          <div className="flex flex-col sm:flex-row sm:items-center gap-4">
            <div className="flex-1">
              <Input
                placeholder="Search domains..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                leftIcon={<Search size={18} />}
              />
            </div>
            <div className="flex gap-2">
              <Button variant="outline" size="sm">All</Button>
              <Button variant="ghost" size="sm">Active</Button>
              <Button variant="ghost" size="sm">Pending</Button>
            </div>
          </div>
        </CardHeader>
        <CardContent className="p-0">
          {/* Domains Table */}
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="border-b border-[var(--color-border-light)]">
                  <th className="text-left text-xs font-medium text-[var(--color-text-muted)] uppercase tracking-wider px-6 py-3">
                    Domain
                  </th>
                  <th className="text-left text-xs font-medium text-[var(--color-text-muted)] uppercase tracking-wider px-6 py-3">
                    Type
                  </th>
                  <th className="text-left text-xs font-medium text-[var(--color-text-muted)] uppercase tracking-wider px-6 py-3">
                    Status
                  </th>
                  <th className="text-left text-xs font-medium text-[var(--color-text-muted)] uppercase tracking-wider px-6 py-3">
                    SSL
                  </th>
                  <th className="text-left text-xs font-medium text-[var(--color-text-muted)] uppercase tracking-wider px-6 py-3">
                    Document Root
                  </th>
                  <th className="text-right text-xs font-medium text-[var(--color-text-muted)] uppercase tracking-wider px-6 py-3">
                    Actions
                  </th>
                </tr>
              </thead>
              <tbody className="divide-y divide-[var(--color-border-light)]">
                {filteredDomains.map((domain) => (
                  <tr
                    key={domain.id}
                    className="hover:bg-[var(--color-primary-dark)]/30 transition-colors"
                  >
                    <td className="px-6 py-4">
                      <div className="flex items-center gap-3">
                        <div className="p-2 rounded-lg bg-[var(--color-primary-dark)]">
                          <Globe size={16} className="text-[#E7F0FA]" />
                        </div>
                        <div>
                          <p className="font-medium text-[var(--color-text-primary)]">
                            {domain.name || 'Unknown'}
                          </p>
                          <p className="text-xs text-[var(--color-text-muted)]">
                            Added {domain.created_at ? new Date(domain.created_at).toLocaleDateString() : 'N/A'}
                          </p>
                        </div>
                      </div>
                    </td>
                    <td className="px-6 py-4">
                      <TypeBadge type={domain.type} />
                    </td>
                    <td className="px-6 py-4">
                      <StatusBadge status={domain.status} />
                    </td>
                    <td className="px-6 py-4">
                      {domain.validated ? (
                        <span className="inline-flex items-center gap-1.5 text-[var(--color-success)]">
                          <Shield size={16} />
                          <span className="text-sm">Secured</span>
                        </span>
                      ) : (
                        <span className="inline-flex items-center gap-1.5 text-[var(--color-warning)]">
                          <AlertCircle size={16} />
                          <span className="text-sm">Not Secured</span>
                        </span>
                      )}
                    </td>
                    <td className="px-6 py-4">
                      <code className="text-sm text-[var(--color-text-secondary)] bg-[var(--color-primary-dark)] px-2 py-1 rounded">
                        {domain.document_root || '/var/www'}
                      </code>
                    </td>
                    <td className="px-6 py-4">
                      <div className="flex items-center justify-end gap-2">
                        <button className="p-2 rounded-lg text-[var(--color-text-secondary)] hover:text-[var(--color-text-primary)] hover:bg-[var(--color-primary-medium)]/20 transition-colors">
                          <ExternalLink size={16} />
                        </button>
                        <button className="p-2 rounded-lg text-[var(--color-text-secondary)] hover:text-[var(--color-text-primary)] hover:bg-[var(--color-primary-medium)]/20 transition-colors">
                          <Settings size={16} />
                        </button>
                        <button
                          onClick={() => openDeleteModal(domain)}
                          className="p-2 rounded-lg text-[var(--color-text-secondary)] hover:text-[var(--color-error)] hover:bg-[var(--color-error)]/10 transition-colors"
                        >
                          <Trash2 size={16} />
                        </button>
                        <button className="p-2 rounded-lg text-[var(--color-text-secondary)] hover:text-[var(--color-text-primary)] hover:bg-[var(--color-primary-medium)]/20 transition-colors">
                          <MoreVertical size={16} />
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          {filteredDomains.length === 0 && (
            <div className="text-center py-12">
              <Globe size={48} className="mx-auto text-[var(--color-text-muted)] mb-4" />
              <p className="text-[var(--color-text-secondary)]">No domains found</p>
              <p className="text-sm text-[var(--color-text-muted)] mt-1">
                Try adjusting your search or add a new domain
              </p>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Modals */}
      <CreateDomainModal
        isOpen={isCreateModalOpen}
        onClose={() => setIsCreateModalOpen(false)}
        onSubmit={handleCreateDomain}
      />
      <DeleteDomainModal
        isOpen={isDeleteModalOpen}
        domain={selectedDomain}
        onClose={() => {
          setIsDeleteModalOpen(false);
          setSelectedDomain(null);
        }}
        onConfirm={handleDeleteDomain}
      />
    </div>
  );
}
