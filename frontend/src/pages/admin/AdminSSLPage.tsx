import { useState, useEffect } from 'react';
import {
  Shield,
  Plus,
  Search,
  RefreshCw,
  Trash2,
  Download,
  AlertCircle,
  CheckCircle,
  Clock,
  Lock,
  Upload,
  Key,
  Calendar,
} from 'lucide-react';
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  Button,
  Badge,
  Input,
  Modal,
  Select,
} from '../../components/ui';
import { sslService, domainService } from '../../api/services';
import type { Certificate, CertificateStatus, Domain } from '../../types';

export function AdminSSLPage() {
  const [certificates, setCertificates] = useState<Certificate[]>([]);
  const [domains, setDomains] = useState<Domain[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState('');
  const [filterStatus, setFilterStatus] = useState<string>('all');

  // Modal states
  const [showLetsEncryptModal, setShowLetsEncryptModal] = useState(false);
  const [showUploadModal, setShowUploadModal] = useState(false);
  const [showSelfSignedModal, setShowSelfSignedModal] = useState(false);
  const [showDeleteConfirm, setShowDeleteConfirm] = useState<string | null>(null);

  // Form states
  const [letsEncryptForm, setLetsEncryptForm] = useState({ domain_id: '', domains: [''] });
  const [uploadForm, setUploadForm] = useState({ domain_id: '', certificate: '', private_key: '', chain: '' });
  const [selfSignedForm, setSelfSignedForm] = useState({ domain_id: '', common_name: '' });
  const [formError, setFormError] = useState<string | null>(null);
  const [isSaving, setIsSaving] = useState(false);

  useEffect(() => {
    loadData();
  }, []);

  const loadData = async () => {
    try {
      setIsLoading(true);
      setError(null);
      const [certsData, domainsData] = await Promise.all([
        sslService.listCertificates(),
        domainService.list(),
      ]);
      setCertificates(certsData);
      setDomains(domainsData);
    } catch (err: any) {
      setError(err.message || 'Failed to load SSL data');
    } finally {
      setIsLoading(false);
    }
  };

  const handleRequestLetsEncrypt = async () => {
    if (!letsEncryptForm.domain_id || letsEncryptForm.domains.filter(d => d).length === 0) {
      setFormError('Please select a domain and enter at least one domain name');
      return;
    }

    try {
      setIsSaving(true);
      setFormError(null);
      await sslService.requestLetsEncrypt({
        domain_id: letsEncryptForm.domain_id,
        domains: letsEncryptForm.domains.filter(d => d),
      });
      await loadData();
      setShowLetsEncryptModal(false);
      setLetsEncryptForm({ domain_id: '', domains: [''] });
    } catch (err: any) {
      setFormError(err.message || 'Failed to request certificate');
    } finally {
      setIsSaving(false);
    }
  };

  const handleUploadCertificate = async () => {
    if (!uploadForm.domain_id || !uploadForm.certificate || !uploadForm.private_key) {
      setFormError('Please fill in all required fields');
      return;
    }

    try {
      setIsSaving(true);
      setFormError(null);
      await sslService.uploadCertificate(uploadForm);
      await loadData();
      setShowUploadModal(false);
      setUploadForm({ domain_id: '', certificate: '', private_key: '', chain: '' });
    } catch (err: any) {
      setFormError(err.message || 'Failed to upload certificate');
    } finally {
      setIsSaving(false);
    }
  };

  const handleGenerateSelfSigned = async () => {
    if (!selfSignedForm.domain_id || !selfSignedForm.common_name) {
      setFormError('Please fill in all fields');
      return;
    }

    try {
      setIsSaving(true);
      setFormError(null);
      await sslService.generateSelfSigned(selfSignedForm);
      await loadData();
      setShowSelfSignedModal(false);
      setSelfSignedForm({ domain_id: '', common_name: '' });
    } catch (err: any) {
      setFormError(err.message || 'Failed to generate certificate');
    } finally {
      setIsSaving(false);
    }
  };

  const handleDelete = async (id: string) => {
    try {
      setIsSaving(true);
      await sslService.deleteCertificate(id);
      await loadData();
      setShowDeleteConfirm(null);
    } catch (err: any) {
      setError(err.message || 'Failed to delete certificate');
    } finally {
      setIsSaving(false);
    }
  };

  const handleToggleAutoRenew = async (id: string, enable: boolean) => {
    try {
      if (enable) {
        await sslService.enableAutoRenew(id);
      } else {
        await sslService.disableAutoRenew(id);
      }
      await loadData();
    } catch (err: any) {
      setError(err.message || 'Failed to update auto-renewal');
    }
  };

  const getStatusBadge = (status: CertificateStatus) => {
    const variants: Record<CertificateStatus, 'success' | 'error' | 'warning' | 'default'> = {
      active: 'success',
      expired: 'error',
      pending: 'warning',
      revoked: 'error',
    };
    return <Badge variant={variants[status]}>{status}</Badge>;
  };

  const getDaysUntilExpiry = (expiresAt: string) => {
    const expiry = new Date(expiresAt);
    const now = new Date();
    const days = Math.ceil((expiry.getTime() - now.getTime()) / (1000 * 60 * 60 * 24));
    return days;
  };

  const getExpiryColor = (days: number) => {
    if (days <= 0) return 'text-[var(--color-error)]';
    if (days <= 7) return 'text-[var(--color-error)]';
    if (days <= 30) return 'text-[var(--color-warning)]';
    return 'text-[var(--color-success)]';
  };

  const filteredCertificates = certificates.filter(cert => {
    const matchesSearch = cert.common_name.toLowerCase().includes(searchQuery.toLowerCase());
    const matchesStatus = filterStatus === 'all' || cert.status === filterStatus;
    return matchesSearch && matchesStatus;
  });

  const stats = {
    total: certificates.length,
    active: certificates.filter(c => c.status === 'active').length,
    expiring: certificates.filter(c => {
      const days = getDaysUntilExpiry(c.expires_at);
      return days > 0 && days <= 30;
    }).length,
    expired: certificates.filter(c => c.status === 'expired').length,
  };

  if (isLoading) {
    return (
      <div className="space-y-6">
        <div className="h-8 bg-[var(--color-primary)]/50 rounded animate-pulse w-48" />
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
          {[1, 2, 3, 4].map(i => (
            <div key={i} className="h-24 bg-[var(--color-primary)]/50 rounded animate-pulse" />
          ))}
        </div>
        <div className="h-96 bg-[var(--color-primary)]/50 rounded animate-pulse" />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-[var(--color-text-primary)]">SSL Certificate Management</h1>
          <p className="text-[var(--color-text-secondary)] mt-1">
            Manage SSL/TLS certificates for all domains
          </p>
        </div>
        <div className="flex gap-2">
          <Button
            variant="outline"
            leftIcon={<Upload size={18} />}
            onClick={() => setShowUploadModal(true)}
          >
            Upload
          </Button>
          <Button
            variant="outline"
            leftIcon={<Key size={18} />}
            onClick={() => setShowSelfSignedModal(true)}
          >
            Self-Signed
          </Button>
          <Button
            leftIcon={<Shield size={18} />}
            onClick={() => setShowLetsEncryptModal(true)}
          >
            Let's Encrypt
          </Button>
        </div>
      </div>

      {/* Error Alert */}
      {error && (
        <div className="flex items-center gap-2 p-4 rounded-lg bg-[var(--color-error)]/20 text-[var(--color-error)]">
          <AlertCircle size={20} />
          <span>{error}</span>
          <button onClick={() => setError(null)} className="ml-auto">×</button>
        </div>
      )}

      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-[var(--color-text-secondary)]">Total Certificates</p>
                <p className="text-2xl font-bold text-[var(--color-text-primary)]">{stats.total}</p>
              </div>
              <Shield size={24} className="text-[#E7F0FA]" />
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-[var(--color-text-secondary)]">Active</p>
                <p className="text-2xl font-bold text-[var(--color-success)]">{stats.active}</p>
              </div>
              <CheckCircle size={24} className="text-[var(--color-success)]" />
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-[var(--color-text-secondary)]">Expiring Soon</p>
                <p className="text-2xl font-bold text-[var(--color-warning)]">{stats.expiring}</p>
              </div>
              <Clock size={24} className="text-[var(--color-warning)]" />
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-[var(--color-text-secondary)]">Expired</p>
                <p className="text-2xl font-bold text-[var(--color-error)]">{stats.expired}</p>
              </div>
              <AlertCircle size={24} className="text-[var(--color-error)]" />
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Certificates List */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <CardTitle>Certificates</CardTitle>
            <div className="flex gap-4">
              <Select
                value={filterStatus}
                onChange={(e) => setFilterStatus(e.target.value)}
                options={[
                  { value: 'all', label: 'All Status' },
                  { value: 'active', label: 'Active' },
                  { value: 'pending', label: 'Pending' },
                  { value: 'expired', label: 'Expired' },
                  { value: 'revoked', label: 'Revoked' },
                ]}
                className="w-40"
              />
              <div className="relative">
                <Search className="absolute left-3 top-1/2 -translate-y-1/2 text-[var(--color-text-muted)]" size={18} />
                <Input
                  placeholder="Search certificates..."
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  className="pl-10 w-64"
                />
              </div>
            </div>
          </div>
        </CardHeader>
        <CardContent>
          {filteredCertificates.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-12 text-[var(--color-text-muted)]">
              <Shield size={48} className="mb-4 opacity-50" />
              <p>No certificates found</p>
              <Button
                className="mt-4"
                variant="outline"
                leftIcon={<Plus size={16} />}
                onClick={() => setShowLetsEncryptModal(true)}
              >
                Request Certificate
              </Button>
            </div>
          ) : (
            <div className="space-y-3">
              {filteredCertificates.map(cert => {
                const daysUntilExpiry = getDaysUntilExpiry(cert.expires_at);
                
                return (
                  <div
                    key={cert.id}
                    className="p-4 rounded-lg bg-[var(--color-primary-dark)]/50 hover:bg-[var(--color-primary-dark)] transition-colors"
                  >
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-4">
                        <div className="p-2 rounded-lg bg-[#7BA4D0]/20">
                          <Lock size={20} className="text-[#E7F0FA]" />
                        </div>
                        <div>
                          <div className="flex items-center gap-2">
                            <h3 className="font-medium text-[var(--color-text-primary)]">
                              {cert.common_name}
                            </h3>
                            {getStatusBadge(cert.status)}
                            <Badge variant="default" size="sm">{cert.type}</Badge>
                          </div>
                          <div className="flex items-center gap-4 mt-1 text-sm text-[var(--color-text-muted)]">
                            <span>Issuer: {cert.issuer}</span>
                            {cert.sans && cert.sans.length > 1 && (
                              <span>+{cert.sans.length - 1} SANs</span>
                            )}
                          </div>
                        </div>
                      </div>
                      
                      <div className="flex items-center gap-6">
                        <div className="text-right">
                          <div className={`flex items-center gap-1 ${getExpiryColor(daysUntilExpiry)}`}>
                            <Calendar size={14} />
                            <span className="text-sm font-medium">
                              {daysUntilExpiry > 0 
                                ? `${daysUntilExpiry} days left`
                                : 'Expired'
                              }
                            </span>
                          </div>
                          <p className="text-xs text-[var(--color-text-muted)] mt-0.5">
                            Expires: {new Date(cert.expires_at).toLocaleDateString()}
                          </p>
                        </div>
                        
                        <div className="flex items-center gap-2">
                          {cert.type === 'letsencrypt' && (
                            <Button
                              size="sm"
                              variant={cert.auto_renew ? 'outline' : 'ghost'}
                              onClick={() => handleToggleAutoRenew(cert.id, !cert.auto_renew)}
                              title={cert.auto_renew ? 'Disable auto-renewal' : 'Enable auto-renewal'}
                            >
                              <RefreshCw size={14} className={cert.auto_renew ? 'text-[var(--color-success)]' : ''} />
                            </Button>
                          )}
                          <Button
                            size="sm"
                            variant="ghost"
                            className="text-[var(--color-error)]"
                            onClick={() => setShowDeleteConfirm(cert.id)}
                          >
                            <Trash2 size={14} />
                          </Button>
                        </div>
                      </div>
                    </div>
                  </div>
                );
              })}
            </div>
          )}
        </CardContent>
      </Card>

      {/* Let's Encrypt Modal */}
      <Modal
        isOpen={showLetsEncryptModal}
        onClose={() => {
          setShowLetsEncryptModal(false);
          setFormError(null);
        }}
        title="Request Let's Encrypt Certificate"
      >
        <div className="space-y-4">
          {formError && (
            <div className="p-3 rounded bg-[var(--color-error)]/20 text-[var(--color-error)] text-sm">
              {formError}
            </div>
          )}
          
          <div>
            <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-2">
              Domain
            </label>
            <Select
              value={letsEncryptForm.domain_id}
              onChange={(e) => {
                const domain = domains.find(d => d.id === e.target.value);
                setLetsEncryptForm({
                  domain_id: e.target.value,
                  domains: domain ? [domain.name] : [''],
                });
              }}
              options={[
                { value: '', label: 'Select a domain' },
                ...domains.map(d => ({ value: d.id, label: d.name })),
              ]}
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-2">
              Domain Names (for certificate)
            </label>
            {letsEncryptForm.domains.map((domain, idx) => (
              <div key={idx} className="flex gap-2 mb-2">
                <Input
                  value={domain}
                  onChange={(e) => {
                    const newDomains = [...letsEncryptForm.domains];
                    newDomains[idx] = e.target.value;
                    setLetsEncryptForm({ ...letsEncryptForm, domains: newDomains });
                  }}
                  placeholder="example.com or *.example.com"
                />
                {idx > 0 && (
                  <Button
                    variant="ghost"
                    onClick={() => {
                      const newDomains = letsEncryptForm.domains.filter((_, i) => i !== idx);
                      setLetsEncryptForm({ ...letsEncryptForm, domains: newDomains });
                    }}
                  >
                    <Trash2 size={16} />
                  </Button>
                )}
              </div>
            ))}
            <Button
              variant="ghost"
              size="sm"
              onClick={() => setLetsEncryptForm({
                ...letsEncryptForm,
                domains: [...letsEncryptForm.domains, ''],
              })}
            >
              <Plus size={14} className="mr-1" /> Add Domain
            </Button>
          </div>

          <div className="flex gap-3 justify-end">
            <Button variant="ghost" onClick={() => setShowLetsEncryptModal(false)}>
              Cancel
            </Button>
            <Button onClick={handleRequestLetsEncrypt} isLoading={isSaving}>
              Request Certificate
            </Button>
          </div>
        </div>
      </Modal>

      {/* Upload Certificate Modal */}
      <Modal
        isOpen={showUploadModal}
        onClose={() => {
          setShowUploadModal(false);
          setFormError(null);
        }}
        title="Upload Custom Certificate"
      >
        <div className="space-y-4">
          {formError && (
            <div className="p-3 rounded bg-[var(--color-error)]/20 text-[var(--color-error)] text-sm">
              {formError}
            </div>
          )}
          
          <div>
            <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-2">
              Domain
            </label>
            <Select
              value={uploadForm.domain_id}
              onChange={(e) => setUploadForm({ ...uploadForm, domain_id: e.target.value })}
              options={[
                { value: '', label: 'Select a domain' },
                ...domains.map(d => ({ value: d.id, label: d.name })),
              ]}
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-2">
              Certificate (PEM)
            </label>
            <textarea
              value={uploadForm.certificate}
              onChange={(e) => setUploadForm({ ...uploadForm, certificate: e.target.value })}
              placeholder="-----BEGIN CERTIFICATE-----&#10;...&#10;-----END CERTIFICATE-----"
              className="w-full h-32 px-4 py-2.5 rounded-lg bg-[var(--color-surface)] border border-[var(--color-border)] text-[var(--color-text-primary)] font-mono text-sm"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-2">
              Private Key (PEM)
            </label>
            <textarea
              value={uploadForm.private_key}
              onChange={(e) => setUploadForm({ ...uploadForm, private_key: e.target.value })}
              placeholder="-----BEGIN PRIVATE KEY-----&#10;...&#10;-----END PRIVATE KEY-----"
              className="w-full h-32 px-4 py-2.5 rounded-lg bg-[var(--color-surface)] border border-[var(--color-border)] text-[var(--color-text-primary)] font-mono text-sm"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-2">
              Certificate Chain (Optional)
            </label>
            <textarea
              value={uploadForm.chain}
              onChange={(e) => setUploadForm({ ...uploadForm, chain: e.target.value })}
              placeholder="-----BEGIN CERTIFICATE-----&#10;...&#10;-----END CERTIFICATE-----"
              className="w-full h-24 px-4 py-2.5 rounded-lg bg-[var(--color-surface)] border border-[var(--color-border)] text-[var(--color-text-primary)] font-mono text-sm"
            />
          </div>

          <div className="flex gap-3 justify-end">
            <Button variant="ghost" onClick={() => setShowUploadModal(false)}>
              Cancel
            </Button>
            <Button onClick={handleUploadCertificate} isLoading={isSaving}>
              Upload Certificate
            </Button>
          </div>
        </div>
      </Modal>

      {/* Self-Signed Modal */}
      <Modal
        isOpen={showSelfSignedModal}
        onClose={() => {
          setShowSelfSignedModal(false);
          setFormError(null);
        }}
        title="Generate Self-Signed Certificate"
      >
        <div className="space-y-4">
          {formError && (
            <div className="p-3 rounded bg-[var(--color-error)]/20 text-[var(--color-error)] text-sm">
              {formError}
            </div>
          )}
          
          <div className="p-3 rounded bg-[var(--color-warning)]/20 text-[var(--color-warning)] text-sm">
            Self-signed certificates are not trusted by browsers and should only be used for testing.
          </div>
          
          <div>
            <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-2">
              Domain
            </label>
            <Select
              value={selfSignedForm.domain_id}
              onChange={(e) => {
                const domain = domains.find(d => d.id === e.target.value);
                setSelfSignedForm({
                  domain_id: e.target.value,
                  common_name: domain?.name || '',
                });
              }}
              options={[
                { value: '', label: 'Select a domain' },
                ...domains.map(d => ({ value: d.id, label: d.name })),
              ]}
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-2">
              Common Name
            </label>
            <Input
              value={selfSignedForm.common_name}
              onChange={(e) => setSelfSignedForm({ ...selfSignedForm, common_name: e.target.value })}
              placeholder="example.com"
            />
          </div>

          <div className="flex gap-3 justify-end">
            <Button variant="ghost" onClick={() => setShowSelfSignedModal(false)}>
              Cancel
            </Button>
            <Button onClick={handleGenerateSelfSigned} isLoading={isSaving}>
              Generate Certificate
            </Button>
          </div>
        </div>
      </Modal>

      {/* Delete Confirmation Modal */}
      <Modal
        isOpen={showDeleteConfirm !== null}
        onClose={() => setShowDeleteConfirm(null)}
        title="Delete Certificate"
      >
        <div className="space-y-4">
          <p className="text-[var(--color-text-secondary)]">
            Are you sure you want to delete this certificate? This action cannot be undone and may cause HTTPS to stop working for the domain.
          </p>
          <div className="flex gap-3 justify-end">
            <Button variant="ghost" onClick={() => setShowDeleteConfirm(null)}>
              Cancel
            </Button>
            <Button
              variant="danger"
              onClick={() => showDeleteConfirm && handleDelete(showDeleteConfirm)}
              isLoading={isSaving}
            >
              Delete
            </Button>
          </div>
        </div>
      </Modal>
    </div>
  );
}
