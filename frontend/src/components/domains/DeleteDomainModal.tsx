import { useState } from 'react';
import { X, AlertTriangle } from 'lucide-react';
import { Button } from '../ui';
import type { Domain } from '../../types';

interface DeleteDomainModalProps {
  isOpen: boolean;
  domain: Domain | null;
  onClose: () => void;
  onConfirm: (domainId: string) => Promise<void>;
}

export function DeleteDomainModal({ isOpen, domain, onClose, onConfirm }: DeleteDomainModalProps) {
  const [isLoading, setIsLoading] = useState(false);
  const [confirmText, setConfirmText] = useState('');

  if (!isOpen || !domain) return null;

  const handleConfirm = async () => {
    if (confirmText !== domain.name) return;
    
    setIsLoading(true);
    try {
      await onConfirm(domain.id);
      setConfirmText('');
      onClose();
    } catch (err) {
      console.error('Failed to delete domain:', err);
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
            <div className="p-2 rounded-lg bg-[var(--color-error)]/10">
              <AlertTriangle size={20} className="text-[var(--color-error)]" />
            </div>
            <h2 className="text-xl font-semibold text-[var(--color-text-primary)]">Delete Domain</h2>
          </div>
          <button
            onClick={onClose}
            className="p-2 rounded-lg text-[var(--color-text-secondary)] hover:text-[var(--color-text-primary)] hover:bg-[var(--color-primary-dark)] transition-colors"
          >
            <X size={20} />
          </button>
        </div>

        {/* Content */}
        <div className="p-6 space-y-5">
          <div className="bg-[var(--color-error)]/10 border border-[var(--color-error)]/20 rounded-lg p-4">
            <p className="text-[var(--color-error)] font-medium mb-2">
              ⚠️ This action cannot be undone
            </p>
            <p className="text-sm text-[var(--color-text-secondary)]">
              Deleting <strong className="text-[var(--color-text-primary)]">{domain.name}</strong> will:
            </p>
            <ul className="mt-2 space-y-1 text-sm text-[var(--color-text-secondary)] list-disc list-inside">
              <li>Remove all DNS records</li>
              <li>Delete web server configuration</li>
              <li>Remove SSL certificates</li>
              <li>The files will remain but website will be inaccessible</li>
            </ul>
          </div>

          <div>
            <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-2">
              Type <strong className="text-[var(--color-text-primary)]">{domain.name}</strong> to confirm
            </label>
            <input
              type="text"
              value={confirmText}
              onChange={(e) => setConfirmText(e.target.value)}
              placeholder={domain.name}
              className="w-full px-4 py-2.5 rounded-lg bg-[var(--color-primary-dark)] border border-[var(--color-border)] text-[var(--color-text-primary)] placeholder-[var(--color-text-muted)] focus:outline-none focus:ring-2 focus:ring-[var(--color-error)] focus:border-transparent"
              autoFocus
            />
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
              type="button"
              onClick={handleConfirm}
              className="flex-1 bg-[var(--color-error)] hover:bg-[var(--color-error)]/90"
              isLoading={isLoading}
              disabled={confirmText !== domain.name}
            >
              Delete Domain
            </Button>
          </div>
        </div>
      </div>
    </div>
  );
}
