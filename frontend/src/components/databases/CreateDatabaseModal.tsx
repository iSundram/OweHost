import { useState } from 'react';
import { X, Database as DatabaseIcon, AlertCircle } from 'lucide-react';
import { Button, Input } from '../ui';

interface CreateDatabaseModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSubmit: (data: { name: string; type: string; charset?: string; collation?: string }) => Promise<void>;
}

export function CreateDatabaseModal({ isOpen, onClose, onSubmit }: CreateDatabaseModalProps) {
  const [formData, setFormData] = useState({
    name: '',
    type: 'mysql',
    charset: 'utf8mb4',
    collation: 'utf8mb4_general_ci',
  });
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState('');

  if (!isOpen) return null;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setIsLoading(true);

    try {
      await onSubmit(formData);
      setFormData({ name: '', type: 'mysql', charset: 'utf8mb4', collation: 'utf8mb4_general_ci' });
      onClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create database');
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
              <DatabaseIcon size={20} className="text-[#E7F0FA]" />
            </div>
            <h2 className="text-xl font-semibold text-[var(--color-text-primary)]">Create New Database</h2>
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
            label="Database Name"
            type="text"
            placeholder="my_database"
            value={formData.name}
            onChange={(e) => setFormData({ ...formData, name: e.target.value })}
            required
            autoFocus
          />

          <div>
            <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-2">
              Database Type
            </label>
            <select
              value={formData.type}
              onChange={(e) => setFormData({ ...formData, type: e.target.value })}
              className="w-full px-4 py-2.5 rounded-lg bg-[var(--color-primary-dark)] border border-[var(--color-border)] text-[var(--color-text-primary)] focus:outline-none focus:ring-2 focus:ring-[#E7F0FA] focus:border-transparent"
            >
              <option value="mysql">MySQL</option>
              <option value="mariadb">MariaDB</option>
              <option value="postgresql">PostgreSQL</option>
            </select>
          </div>

          <div>
            <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-2">
              Character Set
            </label>
            <select
              value={formData.charset}
              onChange={(e) => setFormData({ ...formData, charset: e.target.value })}
              className="w-full px-4 py-2.5 rounded-lg bg-[var(--color-primary-dark)] border border-[var(--color-border)] text-[var(--color-text-primary)] focus:outline-none focus:ring-2 focus:ring-[#E7F0FA] focus:border-transparent"
            >
              <option value="utf8mb4">utf8mb4 (Recommended)</option>
              <option value="utf8">utf8</option>
              <option value="latin1">latin1</option>
            </select>
          </div>

          <div>
            <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-2">
              Collation
            </label>
            <select
              value={formData.collation}
              onChange={(e) => setFormData({ ...formData, collation: e.target.value })}
              className="w-full px-4 py-2.5 rounded-lg bg-[var(--color-primary-dark)] border border-[var(--color-border)] text-[var(--color-text-primary)] focus:outline-none focus:ring-2 focus:ring-[#E7F0FA] focus:border-transparent"
            >
              <option value="utf8mb4_general_ci">utf8mb4_general_ci</option>
              <option value="utf8mb4_unicode_ci">utf8mb4_unicode_ci</option>
              <option value="utf8mb4_bin">utf8mb4_bin</option>
            </select>
          </div>

          <div className="bg-[var(--color-info)]/10 border border-[var(--color-info)]/20 rounded-lg p-3">
            <p className="text-sm text-[var(--color-info)]">
              <strong>Note:</strong> utf8mb4 with utf8mb4_general_ci is recommended for modern applications with emoji support.
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
              Create Database
            </Button>
          </div>
        </form>
      </div>
    </div>
  );
}
