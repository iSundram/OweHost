import { useState, useEffect } from 'react';
import {
  Archive,
  Plus,
  Download,
  Trash2,
  AlertCircle,
  CheckCircle,
  Clock,
  HardDrive,
  RotateCcw,
  RefreshCw,
} from 'lucide-react';
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  Button,
  Badge,
  Modal,
  Select,
  Input,
} from '../../components/ui';
import { backupService } from '../../api/services';
import type { Backup, BackupType, BackupStatus } from '../../types';

export function UserBackupsPage() {
  const [backups, setBackups] = useState<Backup[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Modal states
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [showRestoreModal, setShowRestoreModal] = useState<Backup | null>(null);
  const [showDeleteConfirm, setShowDeleteConfirm] = useState<string | null>(null);

  // Form states
  const [newBackup, setNewBackup] = useState<{ type: BackupType; includes: string[] }>({
    type: 'full',
    includes: ['files', 'databases'],
  });
  const [restoreOptions, setRestoreOptions] = useState({ overwrite: false, destination: '' });
  const [formError, setFormError] = useState<string | null>(null);
  const [isSaving, setIsSaving] = useState(false);

  useEffect(() => {
    loadData();
  }, []);

  const loadData = async () => {
    try {
      setIsLoading(true);
      setError(null);
      const backupsData = await backupService.list();
      setBackups(backupsData);
    } catch (err: any) {
      setError(err.message || 'Failed to load backups');
    } finally {
      setIsLoading(false);
    }
  };

  const handleCreate = async () => {
    try {
      setIsSaving(true);
      setFormError(null);
      await backupService.create(newBackup);
      await loadData();
      setShowCreateModal(false);
      setNewBackup({ type: 'full', includes: ['files', 'databases'] });
    } catch (err: any) {
      setFormError(err.message || 'Failed to create backup');
    } finally {
      setIsSaving(false);
    }
  };

  const handleRestore = async () => {
    if (!showRestoreModal) return;
    try {
      setIsSaving(true);
      await backupService.restore(showRestoreModal.id, restoreOptions);
      await loadData();
      setShowRestoreModal(null);
      setRestoreOptions({ overwrite: false, destination: '' });
    } catch (err: any) {
      setFormError(err.message || 'Failed to restore');
    } finally {
      setIsSaving(false);
    }
  };

  const handleDelete = async (id: string) => {
    try {
      setIsSaving(true);
      await backupService.delete(id);
      await loadData();
      setShowDeleteConfirm(null);
    } catch (err: any) {
      setError(err.message || 'Failed to delete');
    } finally {
      setIsSaving(false);
    }
  };

  const handleDownload = async (id: string) => {
    try {
      const result = await backupService.download(id);
      window.open(result.url, '_blank');
    } catch (err: any) {
      setError(err.message || 'Failed to download');
    }
  };

  const getStatusBadge = (status: BackupStatus) => {
    const variants: Record<BackupStatus, 'success' | 'error' | 'warning' | 'default'> = {
      completed: 'success', failed: 'error', in_progress: 'warning', pending: 'default',
    };
    const icons: Record<BackupStatus, React.ReactNode> = {
      completed: <CheckCircle size={12} />, failed: <AlertCircle size={12} />,
      in_progress: <RefreshCw size={12} className="animate-spin" />, pending: <Clock size={12} />,
    };
    return <Badge variant={variants[status]} className="flex items-center gap-1">{icons[status]} {status.replace('_', ' ')}</Badge>;
  };

  const formatSize = (mb: number) => mb >= 1024 ? `${(mb / 1024).toFixed(2)} GB` : `${mb.toFixed(2)} MB`;

  const stats = {
    total: backups.length,
    completed: backups.filter(b => b.status === 'completed').length,
    totalSize: backups.reduce((acc, b) => acc + b.size_mb, 0),
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
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-[var(--color-text-primary)]">Backups</h1>
          <p className="text-[var(--color-text-secondary)] mt-1">Manage your account backups</p>
        </div>
        <Button leftIcon={<Archive size={18} />} onClick={() => setShowCreateModal(true)}>
          Create Backup
        </Button>
      </div>

      {error && (
        <div className="flex items-center gap-2 p-4 rounded-lg bg-[var(--color-error)]/20 text-[var(--color-error)]">
          <AlertCircle size={20} /><span>{error}</span>
          <button onClick={() => setError(null)} className="ml-auto">×</button>
        </div>
      )}

      {/* Stats */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-[var(--color-text-secondary)]">Total Backups</p>
                <p className="text-2xl font-bold text-[var(--color-text-primary)]">{stats.total}</p>
              </div>
              <Archive size={24} className="text-[#E7F0FA]" />
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-[var(--color-text-secondary)]">Completed</p>
                <p className="text-2xl font-bold text-[var(--color-success)]">{stats.completed}</p>
              </div>
              <CheckCircle size={24} className="text-[var(--color-success)]" />
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-[var(--color-text-secondary)]">Total Size</p>
                <p className="text-2xl font-bold text-[var(--color-text-primary)]">{formatSize(stats.totalSize)}</p>
              </div>
              <HardDrive size={24} className="text-[#E7F0FA]" />
            </div>
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader><CardTitle>Backup History</CardTitle></CardHeader>
        <CardContent>
          {backups.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-12 text-[var(--color-text-muted)]">
              <Archive size={48} className="mb-4 opacity-50" />
              <p>No backups yet</p>
              <Button className="mt-4" variant="outline" leftIcon={<Plus size={16} />} onClick={() => setShowCreateModal(true)}>
                Create First Backup
              </Button>
            </div>
          ) : (
            <div className="space-y-3">
              {backups.map(backup => (
                <div key={backup.id} className="p-4 rounded-lg bg-[var(--color-primary-dark)]/50">
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-4">
                      <Archive size={20} className="text-[#E7F0FA]" />
                      <div>
                        <div className="flex items-center gap-2">
                          <Badge variant={backup.type === 'full' ? 'success' : 'warning'}>{backup.type}</Badge>
                          {getStatusBadge(backup.status)}
                        </div>
                        <div className="flex items-center gap-4 mt-1 text-sm text-[var(--color-text-muted)]">
                          <span>{formatSize(backup.size_mb)}</span>
                          <span>{new Date(backup.created_at).toLocaleString()}</span>
                        </div>
                        {backup.includes && backup.includes.length > 0 && (
                          <div className="flex gap-1 mt-1">
                            {backup.includes.map(item => (
                              <span key={item} className="text-xs px-2 py-0.5 rounded bg-[var(--color-primary)] text-[var(--color-text-muted)]">{item}</span>
                            ))}
                          </div>
                        )}
                      </div>
                    </div>
                    <div className="flex items-center gap-2">
                      {backup.status === 'completed' && (
                        <>
                          <Button size="sm" variant="ghost" onClick={() => handleDownload(backup.id)} title="Download"><Download size={14} /></Button>
                          <Button size="sm" variant="ghost" onClick={() => setShowRestoreModal(backup)} title="Restore"><RotateCcw size={14} /></Button>
                        </>
                      )}
                      <Button size="sm" variant="ghost" className="text-[var(--color-error)]" onClick={() => setShowDeleteConfirm(backup.id)}><Trash2 size={14} /></Button>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      {/* Create Modal */}
      <Modal isOpen={showCreateModal} onClose={() => setShowCreateModal(false)} title="Create Backup">
        <div className="space-y-4">
          {formError && <div className="p-3 rounded bg-[var(--color-error)]/20 text-[var(--color-error)] text-sm">{formError}</div>}
          <div>
            <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-2">Backup Type</label>
            <Select value={newBackup.type} onChange={(e) => setNewBackup({ ...newBackup, type: e.target.value as BackupType })}
              options={[{ value: 'full', label: 'Full Backup' }, { value: 'incremental', label: 'Incremental' }]} />
          </div>
          <div>
            <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-2">Include</label>
            <div className="space-y-2">
              {['files', 'databases', 'emails'].map(item => (
                <label key={item} className="flex items-center gap-2 cursor-pointer">
                  <input type="checkbox" checked={newBackup.includes.includes(item)}
                    onChange={(e) => setNewBackup({
                      ...newBackup,
                      includes: e.target.checked ? [...newBackup.includes, item] : newBackup.includes.filter(i => i !== item)
                    })} className="rounded" />
                  <span className="text-sm capitalize">{item}</span>
                </label>
              ))}
            </div>
          </div>
          <div className="flex gap-3 justify-end">
            <Button variant="ghost" onClick={() => setShowCreateModal(false)}>Cancel</Button>
            <Button onClick={handleCreate} isLoading={isSaving}>Start Backup</Button>
          </div>
        </div>
      </Modal>

      {/* Restore Modal */}
      <Modal isOpen={showRestoreModal !== null} onClose={() => setShowRestoreModal(null)} title="Restore Backup">
        <div className="space-y-4">
          <div className="p-3 rounded bg-[var(--color-warning)]/20 text-[var(--color-warning)] text-sm">
            Warning: Restoring may overwrite existing data.
          </div>
          <label className="flex items-center gap-2 cursor-pointer">
            <input type="checkbox" checked={restoreOptions.overwrite} onChange={(e) => setRestoreOptions({ ...restoreOptions, overwrite: e.target.checked })} />
            <span className="text-sm">Overwrite existing files</span>
          </label>
          <div className="flex gap-3 justify-end">
            <Button variant="ghost" onClick={() => setShowRestoreModal(null)}>Cancel</Button>
            <Button onClick={handleRestore} isLoading={isSaving}>Restore</Button>
          </div>
        </div>
      </Modal>

      {/* Delete Confirm */}
      <Modal isOpen={showDeleteConfirm !== null} onClose={() => setShowDeleteConfirm(null)} title="Delete Backup">
        <div className="space-y-4">
          <p className="text-[var(--color-text-secondary)]">Are you sure you want to delete this backup?</p>
          <div className="flex gap-3 justify-end">
            <Button variant="ghost" onClick={() => setShowDeleteConfirm(null)}>Cancel</Button>
            <Button variant="danger" onClick={() => showDeleteConfirm && handleDelete(showDeleteConfirm)} isLoading={isSaving}>Delete</Button>
          </div>
        </div>
      </Modal>
    </div>
  );
}
