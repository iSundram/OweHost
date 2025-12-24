import { useState, useEffect } from 'react';
import {
  Shield,
  Plus,
  AlertCircle,
  CheckCircle,
  Clock,
  Lock,
  Calendar,
  RefreshCw,
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

export function ResellerSSLPage() {
  const [certificates, setCertificates] = useState<Certificate[]>([]);
  const [domains, setDomains] = useState<Domain[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Modal states
  const [showLetsEncryptModal, setShowLetsEncryptModal] = useState(false);
  const [letsEncryptForm, setLetsEncryptForm] = useState({ domain_id: '', domains: [''] });
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
      setFormError('Please select a domain');
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
    return Math.ceil((expiry.getTime() - now.getTime()) / (1000 * 60 * 60 * 24));
  };

  const getExpiryColor = (days: number) => {
    if (days <= 0) return 'text-[var(--color-error)]';
    if (days <= 7) return 'text-[var(--color-error)]';
    if (days <= 30) return 'text-[var(--color-warning)]';
    return 'text-[var(--color-success)]';
  };

  if (isLoading) {
    return (
      <div className="space-y-6">
        <div className="h-8 bg-[var(--color-primary)]/50 rounded animate-pulse w-48" />
        <div className="h-96 bg-[var(--color-primary)]/50 rounded animate-pulse" />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-[var(--color-text-primary)]">SSL Certificates (Customer)</h1>
          <p className="text-[var(--color-text-secondary)] mt-1">
            Manage SSL certificates for customer domains with SSL/TLS certificates
          </p>
        </div>
        <Button
          leftIcon={<Shield size={18} />}
          onClick={() => setShowLetsEncryptModal(true)}
        >
          Get Free SSL
        </Button>
      </div>

      {/* Error Alert */}
      {error && (
        <div className="flex items-center gap-2 p-4 rounded-lg bg-[var(--color-error)]/20 text-[var(--color-error)]">
          <AlertCircle size={20} />
          <span>{error}</span>
          <button onClick={() => setError(null)} className="ml-auto">×</button>
        </div>
      )}

      {/* Certificates List */}
      <Card>
        <CardHeader>
          <CardTitle>Your Certificates</CardTitle>
        </CardHeader>
        <CardContent>
          {certificates.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-12 text-[var(--color-text-muted)]">
              <Shield size={48} className="mb-4 opacity-50" />
              <p className="mb-2">No SSL certificates installed</p>
              <p className="text-sm mb-4">Get a free Let's Encrypt certificate for your domains</p>
              <Button
                variant="outline"
                leftIcon={<Plus size={16} />}
                onClick={() => setShowLetsEncryptModal(true)}
              >
                Get Free SSL
              </Button>
            </div>
          ) : (
            <div className="space-y-3">
              {certificates.map(cert => {
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
                          <p className="text-sm text-[var(--color-text-muted)] mt-1">
                            Issuer: {cert.issuer}
                          </p>
                        </div>
                      </div>
                      
                      <div className="flex items-center gap-6">
                        <div className="text-right">
                          <div className={`flex items-center gap-1 ${getExpiryColor(daysUntilExpiry)}`}>
                            <Calendar size={14} />
                            <span className="text-sm font-medium">
                              {daysUntilExpiry > 0 ? `${daysUntilExpiry} days left` : 'Expired'}
                            </span>
                          </div>
                          <p className="text-xs text-[var(--color-text-muted)] mt-0.5">
                            Expires: {new Date(cert.expires_at).toLocaleDateString()}
                          </p>
                        </div>
                        
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
        title="Get Free SSL Certificate"
      >
        <div className="space-y-4">
          <div className="p-3 rounded bg-[var(--color-success)]/20 text-[var(--color-success)] text-sm">
            Let's Encrypt provides free SSL certificates that auto-renew every 90 days.
          </div>

          {formError && (
            <div className="p-3 rounded bg-[var(--color-error)]/20 text-[var(--color-error)] text-sm">
              {formError}
            </div>
          )}
          
          <div>
            <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-2">
              Select Domain
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

          <div className="flex gap-3 justify-end">
            <Button variant="ghost" onClick={() => setShowLetsEncryptModal(false)}>
              Cancel
            </Button>
            <Button onClick={handleRequestLetsEncrypt} isLoading={isSaving}>
              Get Certificate
            </Button>
          </div>
        </div>
      </Modal>
    </div>
  );
}
