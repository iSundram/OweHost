import { useState } from 'react';
import { X, Globe, AlertCircle } from 'lucide-react';
import { Button, Input } from '../ui';
import type { Domain } from '../../types';

interface CreateDomainModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSubmit: (data: { name: string; type: string; document_root?: string }) => Promise<void>;
}

export function CreateDomainModal({ isOpen, onClose, onSubmit }: CreateDomainModalProps) {
  const [formData, setFormData] = useState({
    name: '',
    type: 'primary',
    document_root: '',
  });
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState('');

  if (!isOpen) return null;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setIsLoading(true);

    try {
      const documentRoot = formData.document_root || `/var/www/${formData.name}`;
      await onSubmit({
        name: formData.name,
        type: formData.type,
        document_root: documentRoot,
      });
      setFormData({ name: '', type: 'primary', document_root: '' });
      onClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create domain');
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/50 backdrop-blur-sm">
      <div className="bg-[var(--color-surface)] rounded-2xl shadow-2xl max-w-md w-full border border-[var(--color-border-light)]">
        {/* Header */}
        <div className="flex items-center justify-between p-6 border-b border-[var(--color-border-light)]">
          <div className="flex items-center gap-3">
            <div className="p-2 rounded-lg bg-gradient-to-br from-[#7BA4D0]/30 to-[#E7F0FA]/20">
              <Globe size={20} className="text-[#E7F0FA]" />
            </div>
            <h2 className="text-xl font-semibold text-[var(--color-text-primary)]">Add New Domain</h2>
          </div>
          <button
            onClick={onClose}
            className="p-2 rounded-lg text-[var(--color-text-secondary)] hover:text-[var(--color-text-primary)] hover:bg-[var(--color-primary-dark)] transition-colors"
          >
            <X size={20} />
          </button>
        </div>

        {/* Form */}
        <form onSubmit={handleSubmit} className="p-6 space-y-5">
          {error && (
            <div className="p-3 rounded-lg bg-[var(--color-error)]/10 border border-[var(--color-error)]/20 flex items-start gap-2">
              <AlertCircle size={18} className="text-[var(--color-error)] mt-0.5" />
              <p className="text-sm text-[var(--color-error)]">{error}</p>
            </div>
          )}

          <Input
            label="Domain Name"
            type="text"
            placeholder="example.com"
            value={formData.name}
            onChange={(e) => setFormData({ ...formData, name: e.target.value })}
            required
            autoFocus
          />

          <div>
            <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-2">
              Domain Type
            </label>
            <select
              value={formData.type}
              onChange={(e) => setFormData({ ...formData, type: e.target.value })}
              className="w-full px-4 py-2.5 rounded-lg bg-[var(--color-primary-dark)] border border-[var(--color-border)] text-[var(--color-text-primary)] focus:outline-none focus:ring-2 focus:ring-[#E7F0FA] focus:border-transparent"
            >
              <option value="primary">Primary Domain</option>
              <option value="addon">Addon Domain</option>
              <option value="parked">Parked Domain</option>
              <option value="alias">Alias Domain</option>
            </select>
          </div>

          <Input
            label="Document Root (Optional)"
            type="text"
            placeholder={`/var/www/${formData.name || 'example.com'}`}
            value={formData.document_root}
            onChange={(e) => setFormData({ ...formData, document_root: e.target.value })}
          />

          <div className="bg-[var(--color-info)]/10 border border-[var(--color-info)]/20 rounded-lg p-3">
            <p className="text-sm text-[var(--color-info)]">
              <strong>Note:</strong> Make sure your domain's DNS is pointing to this server's IP address.
            </p>
          </div>

          {/* Actions */}
          <div className="flex gap-3 pt-2">
            <Button
              type="button"
              variant="outline"
              onClick={onClose}
              className="flex-1"
              disabled={isLoading}
            >
              Cancel
            </Button>
            <Button
              type="submit"
              variant="primary"
              className="flex-1"
              isLoading={isLoading}
            >
              Add Domain
            </Button>
          </div>
        </form>
      </div>
    </div>
  );
}
