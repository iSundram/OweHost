import { useState, useEffect } from 'react';
import {
  Archive,
  Plus,
  Search,
  Download,
  Trash2,
  RefreshCw,
  AlertCircle,
  CheckCircle,
  Clock,
  HardDrive,
  Calendar,
  RotateCcw,
  Settings,
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
import { backupService, userService } from '../../api/services';
import type { Backup, BackupSchedule, BackupType, BackupStatus, User } from '../../types';

export function AdminBackupsPage() {
  const [backups, setBackups] = useState<Backup[]>([]);
  const [schedules, setSchedules] = useState<BackupSchedule[]>([]);
  const [users, setUsers] = useState<User[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState('');
  const [filterType, setFilterType] = useState<string>('all');
  const [filterStatus, setFilterStatus] = useState<string>('all');
  const [activeTab, setActiveTab] = useState<'backups' | 'schedules'>('backups');

  // Modal states
  const [showCreateBackupModal, setShowCreateBackupModal] = useState(false);
  const [showCreateScheduleModal, setShowCreateScheduleModal] = useState(false);
  const [showRestoreModal, setShowRestoreModal] = useState<Backup | null>(null);
  const [showDeleteConfirm, setShowDeleteConfirm] = useState<{ type: 'backup' | 'schedule'; id: string } | null>(null);

  // Form states
  const [newBackup, setNewBackup] = useState<{
    type: BackupType;
    user_id: string;
    includes: string[];
  }>({
    type: 'full',
    user_id: '',
    includes: ['files', 'databases', 'emails'],
  });
  const [newSchedule, setNewSchedule] = useState({
    type: 'full' as BackupType,
    frequency: 'daily',
    retention_days: 30,
    user_id: '',
  });
  const [restoreOptions, setRestoreOptions] = useState({
    overwrite: false,
    destination: '',
  });
  const [formError, setFormError] = useState<string | null>(null);
  const [isSaving, setIsSaving] = useState(false);

  useEffect(() => {
    loadData();
  }, []);

  const loadData = async () => {
    try {
      setIsLoading(true);
      setError(null);
      const [backupsData, schedulesData, usersData] = await Promise.all([
        backupService.list(),
        backupService.listSchedules(),
        userService.list(),
      ]);
      setBackups(backupsData);
      setSchedules(schedulesData);
      setUsers(usersData);
    } catch (err: any) {
      setError(err.message || 'Failed to load backup data');
    } finally {
      setIsLoading(false);
    }
  };

  const handleCreateBackup = async () => {
    try {
      setIsSaving(true);
      setFormError(null);
      await backupService.create(newBackup);
      await loadData();
      setShowCreateBackupModal(false);
      setNewBackup({ type: 'full', user_id: '', includes: ['files', 'databases', 'emails'] });
    } catch (err: any) {
      setFormError(err.message || 'Failed to create backup');
    } finally {
      setIsSaving(false);
    }
  };

  const handleCreateSchedule = async () => {
    try {
      setIsSaving(true);
      setFormError(null);
      await backupService.createSchedule(newSchedule);
      await loadData();
      setShowCreateScheduleModal(false);
      setNewSchedule({ type: 'full', frequency: 'daily', retention_days: 30, user_id: '' });
    } catch (err: any) {
      setFormError(err.message || 'Failed to create schedule');
    } finally {
      setIsSaving(false);
    }
  };

  const handleRestore = async () => {
    if (!showRestoreModal) return;
    
    try {
      setIsSaving(true);
      setFormError(null);
      await backupService.restore(showRestoreModal.id, restoreOptions);
      await loadData();
      setShowRestoreModal(null);
      setRestoreOptions({ overwrite: false, destination: '' });
    } catch (err: any) {
      setFormError(err.message || 'Failed to restore backup');
    } finally {
      setIsSaving(false);
    }
  };

  const handleDelete = async (type: 'backup' | 'schedule', id: string) => {
    try {
      setIsSaving(true);
      if (type === 'backup') {
        await backupService.delete(id);
      } else {
        await backupService.deleteSchedule(id);
      }
      await loadData();
      setShowDeleteConfirm(null);
    } catch (err: any) {
      setError(err.message || 'Failed to delete');
    } finally {
      setIsSaving(false);
    }
  };

  const handleToggleSchedule = async (schedule: BackupSchedule) => {
    try {
      if (schedule.enabled) {
        await backupService.disableSchedule(schedule.id);
      } else {
        await backupService.enableSchedule(schedule.id);
      }
      await loadData();
    } catch (err: any) {
      setError(err.message || 'Failed to update schedule');
    }
  };

  const handleDownload = async (id: string) => {
    try {
      const result = await backupService.download(id);
      window.open(result.url, '_blank');
    } catch (err: any) {
      setError(err.message || 'Failed to download backup');
    }
  };

  const getStatusBadge = (status: BackupStatus) => {
    const variants: Record<BackupStatus, 'success' | 'error' | 'warning' | 'default'> = {
      completed: 'success',
      failed: 'error',
      in_progress: 'warning',
      pending: 'default',
    };
    const icons: Record<BackupStatus, React.ReactNode> = {
      completed: <CheckCircle size={12} />,
      failed: <AlertCircle size={12} />,
      in_progress: <RefreshCw size={12} className="animate-spin" />,
      pending: <Clock size={12} />,
    };
    return (
      <Badge variant={variants[status]} className="flex items-center gap-1">
        {icons[status]}
        {status.replace('_', ' ')}
      </Badge>
    );
  };

  const getTypeBadge = (type: BackupType) => {
    const variants: Record<BackupType, 'default' | 'success' | 'warning'> = {
      full: 'success',
      incremental: 'warning',
      differential: 'default',
    };
    return <Badge variant={variants[type]}>{type}</Badge>;
  };

  const formatSize = (mb: number) => {
    if (mb >= 1024) {
      return `${(mb / 1024).toFixed(2)} GB`;
    }
    return `${mb.toFixed(2)} MB`;
  };

  const filteredBackups = backups.filter(backup => {
    const matchesSearch = backup.id.toLowerCase().includes(searchQuery.toLowerCase());
    const matchesType = filterType === 'all' || backup.type === filterType;
    const matchesStatus = filterStatus === 'all' || backup.status === filterStatus;
    return matchesSearch && matchesType && matchesStatus;
  });

  const stats = {
    total: backups.length,
    completed: backups.filter(b => b.status === 'completed').length,
    inProgress: backups.filter(b => b.status === 'in_progress').length,
    totalSize: backups.reduce((acc, b) => acc + b.size_mb, 0),
    schedulesActive: schedules.filter(s => s.enabled).length,
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
          <h1 className="text-2xl font-bold text-[var(--color-text-primary)]">Backup Management</h1>
          <p className="text-[var(--color-text-secondary)] mt-1">
            Manage backups and schedules for all users
          </p>
        </div>
        <div className="flex gap-2">
          <Button
            variant="outline"
            leftIcon={<Calendar size={18} />}
            onClick={() => setShowCreateScheduleModal(true)}
          >
            New Schedule
          </Button>
          <Button
            leftIcon={<Archive size={18} />}
            onClick={() => setShowCreateBackupModal(true)}
          >
            Create Backup
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
      <div className="grid grid-cols-1 md:grid-cols-5 gap-4">
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
                <p className="text-sm text-[var(--color-text-secondary)]">In Progress</p>
                <p className="text-2xl font-bold text-[var(--color-warning)]">{stats.inProgress}</p>
              </div>
              <RefreshCw size={24} className="text-[var(--color-warning)]" />
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
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-[var(--color-text-secondary)]">Active Schedules</p>
                <p className="text-2xl font-bold text-[var(--color-text-primary)]">{stats.schedulesActive}</p>
              </div>
              <Calendar size={24} className="text-[#E7F0FA]" />
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Tabs */}
      <div className="flex gap-4 border-b border-[var(--color-border)]">
        <button
          className={`pb-2 px-1 text-sm font-medium transition-colors ${
            activeTab === 'backups'
              ? 'text-[var(--color-text-primary)] border-b-2 border-[#E7F0FA]'
              : 'text-[var(--color-text-muted)] hover:text-[var(--color-text-secondary)]'
          }`}
          onClick={() => setActiveTab('backups')}
        >
          Backups ({backups.length})
        </button>
        <button
          className={`pb-2 px-1 text-sm font-medium transition-colors ${
            activeTab === 'schedules'
              ? 'text-[var(--color-text-primary)] border-b-2 border-[#E7F0FA]'
              : 'text-[var(--color-text-muted)] hover:text-[var(--color-text-secondary)]'
          }`}
          onClick={() => setActiveTab('schedules')}
        >
          Schedules ({schedules.length})
        </button>
      </div>

      {/* Backups List */}
      {activeTab === 'backups' && (
        <Card>
          <CardHeader>
            <div className="flex items-center justify-between flex-wrap gap-4">
              <CardTitle>Backup History</CardTitle>
              <div className="flex gap-4 flex-wrap">
                <Select
                  value={filterType}
                  onChange={(e) => setFilterType(e.target.value)}
                  options={[
                    { value: 'all', label: 'All Types' },
                    { value: 'full', label: 'Full' },
                    { value: 'incremental', label: 'Incremental' },
                    { value: 'differential', label: 'Differential' },
                  ]}
                  className="w-40"
                />
                <Select
                  value={filterStatus}
                  onChange={(e) => setFilterStatus(e.target.value)}
                  options={[
                    { value: 'all', label: 'All Status' },
                    { value: 'completed', label: 'Completed' },
                    { value: 'in_progress', label: 'In Progress' },
                    { value: 'pending', label: 'Pending' },
                    { value: 'failed', label: 'Failed' },
                  ]}
                  className="w-40"
                />
                <div className="relative">
                  <Search className="absolute left-3 top-1/2 -translate-y-1/2 text-[var(--color-text-muted)]" size={18} />
                  <Input
                    placeholder="Search..."
                    value={searchQuery}
                    onChange={(e) => setSearchQuery(e.target.value)}
                    className="pl-10 w-48"
                  />
                </div>
              </div>
            </div>
          </CardHeader>
          <CardContent>
            {filteredBackups.length === 0 ? (
              <div className="flex flex-col items-center justify-center py-12 text-[var(--color-text-muted)]">
                <Archive size={48} className="mb-4 opacity-50" />
                <p>No backups found</p>
                <Button
                  className="mt-4"
                  variant="outline"
                  leftIcon={<Plus size={16} />}
                  onClick={() => setShowCreateBackupModal(true)}
                >
                  Create Backup
                </Button>
              </div>
            ) : (
              <div className="space-y-3">
                {filteredBackups.map(backup => {
                  const user = users.find(u => u.id === backup.user_id);
                  
                  return (
                    <div
                      key={backup.id}
                      className="p-4 rounded-lg bg-[var(--color-primary-dark)]/50 hover:bg-[var(--color-primary-dark)] transition-colors"
                    >
                      <div className="flex items-center justify-between">
                        <div className="flex items-center gap-4">
                          <div className="p-2 rounded-lg bg-[#7BA4D0]/20">
                            <Archive size={20} className="text-[#E7F0FA]" />
                          </div>
                          <div>
                            <div className="flex items-center gap-2">
                              {getTypeBadge(backup.type)}
                              {getStatusBadge(backup.status)}
                            </div>
                            <div className="flex items-center gap-4 mt-1 text-sm text-[var(--color-text-muted)]">
                              {user && <span>User: {user.username}</span>}
                              <span>{formatSize(backup.size_mb)}</span>
                              <span>{new Date(backup.created_at).toLocaleString()}</span>
                            </div>
                            {backup.includes && backup.includes.length > 0 && (
                              <div className="flex gap-1 mt-1">
                                {backup.includes.map(item => (
                                  <span
                                    key={item}
                                    className="text-xs px-2 py-0.5 rounded bg-[var(--color-primary)] text-[var(--color-text-muted)]"
                                  >
                                    {item}
                                  </span>
                                ))}
                              </div>
                            )}
                          </div>
                        </div>
                        
                        <div className="flex items-center gap-2">
                          {backup.status === 'completed' && (
                            <>
                              <Button
                                size="sm"
                                variant="ghost"
                                onClick={() => handleDownload(backup.id)}
                                title="Download"
                              >
                                <Download size={14} />
                              </Button>
                              <Button
                                size="sm"
                                variant="ghost"
                                onClick={() => setShowRestoreModal(backup)}
                                title="Restore"
                              >
                                <RotateCcw size={14} />
                              </Button>
                            </>
                          )}
                          <Button
                            size="sm"
                            variant="ghost"
                            className="text-[var(--color-error)]"
                            onClick={() => setShowDeleteConfirm({ type: 'backup', id: backup.id })}
                          >
                            <Trash2 size={14} />
                          </Button>
                        </div>
                      </div>
                    </div>
                  );
                })}
              </div>
            )}
          </CardContent>
        </Card>
      )}

      {/* Schedules List */}
      {activeTab === 'schedules' && (
        <Card>
          <CardHeader>
            <CardTitle>Backup Schedules</CardTitle>
          </CardHeader>
          <CardContent>
            {schedules.length === 0 ? (
              <div className="flex flex-col items-center justify-center py-12 text-[var(--color-text-muted)]">
                <Calendar size={48} className="mb-4 opacity-50" />
                <p>No schedules configured</p>
                <Button
                  className="mt-4"
                  variant="outline"
                  leftIcon={<Plus size={16} />}
                  onClick={() => setShowCreateScheduleModal(true)}
                >
                  Create Schedule
                </Button>
              </div>
            ) : (
              <div className="space-y-3">
                {schedules.map(schedule => {
                  const user = users.find(u => u.id === schedule.user_id);
                  
                  return (
                    <div
                      key={schedule.id}
                      className="p-4 rounded-lg bg-[var(--color-primary-dark)]/50 hover:bg-[var(--color-primary-dark)] transition-colors"
                    >
                      <div className="flex items-center justify-between">
                        <div className="flex items-center gap-4">
                          <div className="p-2 rounded-lg bg-[#7BA4D0]/20">
                            <Calendar size={20} className="text-[#E7F0FA]" />
                          </div>
                          <div>
                            <div className="flex items-center gap-2">
                              {getTypeBadge(schedule.type)}
                              <Badge variant={schedule.enabled ? 'success' : 'default'}>
                                {schedule.enabled ? 'Active' : 'Disabled'}
                              </Badge>
                            </div>
                            <div className="flex items-center gap-4 mt-1 text-sm text-[var(--color-text-muted)]">
                              {user ? <span>User: {user.username}</span> : <span>All Users</span>}
                              <span className="capitalize">{schedule.frequency}</span>
                              <span>Retention: {schedule.retention_days} days</span>
                            </div>
                            {schedule.next_run_at && (
                              <p className="text-xs text-[var(--color-text-muted)] mt-1">
                                Next run: {new Date(schedule.next_run_at).toLocaleString()}
                              </p>
                            )}
                          </div>
                        </div>
                        
                        <div className="flex items-center gap-2">
                          <Button
                            size="sm"
                            variant={schedule.enabled ? 'outline' : 'ghost'}
                            onClick={() => handleToggleSchedule(schedule)}
                            title={schedule.enabled ? 'Disable' : 'Enable'}
                          >
                            {schedule.enabled ? <RefreshCw size={14} /> : <Settings size={14} />}
                          </Button>
                          <Button
                            size="sm"
                            variant="ghost"
                            className="text-[var(--color-error)]"
                            onClick={() => setShowDeleteConfirm({ type: 'schedule', id: schedule.id })}
                          >
                            <Trash2 size={14} />
                          </Button>
                        </div>
                      </div>
                    </div>
                  );
                })}
              </div>
            )}
          </CardContent>
        </Card>
      )}

      {/* Create Backup Modal */}
      <Modal
        isOpen={showCreateBackupModal}
        onClose={() => {
          setShowCreateBackupModal(false);
          setFormError(null);
        }}
        title="Create Backup"
      >
        <div className="space-y-4">
          {formError && (
            <div className="p-3 rounded bg-[var(--color-error)]/20 text-[var(--color-error)] text-sm">
              {formError}
            </div>
          )}
          
          <div>
            <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-2">
              User
            </label>
            <Select
              value={newBackup.user_id}
              onChange={(e) => setNewBackup({ ...newBackup, user_id: e.target.value })}
              options={[
                { value: '', label: 'All Users (System Backup)' },
                ...users.map(u => ({ value: u.id, label: u.username })),
              ]}
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-2">
              Backup Type
            </label>
            <Select
              value={newBackup.type}
              onChange={(e) => setNewBackup({ ...newBackup, type: e.target.value as BackupType })}
              options={[
                { value: 'full', label: 'Full Backup' },
                { value: 'incremental', label: 'Incremental Backup' },
                { value: 'differential', label: 'Differential Backup' },
              ]}
            />
            <p className="text-xs text-[var(--color-text-muted)] mt-1">
              {newBackup.type === 'full' && 'Complete backup of all selected data'}
              {newBackup.type === 'incremental' && 'Only changes since last backup'}
              {newBackup.type === 'differential' && 'Changes since last full backup'}
            </p>
          </div>

          <div>
            <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-2">
              Include
            </label>
            <div className="space-y-2">
              {['files', 'databases', 'emails', 'dns', 'ssl'].map(item => (
                <label key={item} className="flex items-center gap-2 cursor-pointer">
                  <input
                    type="checkbox"
                    checked={newBackup.includes.includes(item)}
                    onChange={(e) => {
                      if (e.target.checked) {
                        setNewBackup({ ...newBackup, includes: [...newBackup.includes, item] });
                      } else {
                        setNewBackup({ ...newBackup, includes: newBackup.includes.filter(i => i !== item) });
                      }
                    }}
                    className="rounded border-[var(--color-border)]"
                  />
                  <span className="text-sm text-[var(--color-text-primary)] capitalize">{item}</span>
                </label>
              ))}
            </div>
          </div>

          <div className="flex gap-3 justify-end">
            <Button variant="ghost" onClick={() => setShowCreateBackupModal(false)}>
              Cancel
            </Button>
            <Button onClick={handleCreateBackup} isLoading={isSaving}>
              Start Backup
            </Button>
          </div>
        </div>
      </Modal>

      {/* Create Schedule Modal */}
      <Modal
        isOpen={showCreateScheduleModal}
        onClose={() => {
          setShowCreateScheduleModal(false);
          setFormError(null);
        }}
        title="Create Backup Schedule"
      >
        <div className="space-y-4">
          {formError && (
            <div className="p-3 rounded bg-[var(--color-error)]/20 text-[var(--color-error)] text-sm">
              {formError}
            </div>
          )}
          
          <div>
            <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-2">
              User
            </label>
            <Select
              value={newSchedule.user_id}
              onChange={(e) => setNewSchedule({ ...newSchedule, user_id: e.target.value })}
              options={[
                { value: '', label: 'All Users' },
                ...users.map(u => ({ value: u.id, label: u.username })),
              ]}
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-2">
              Backup Type
            </label>
            <Select
              value={newSchedule.type}
              onChange={(e) => setNewSchedule({ ...newSchedule, type: e.target.value as BackupType })}
              options={[
                { value: 'full', label: 'Full Backup' },
                { value: 'incremental', label: 'Incremental Backup' },
                { value: 'differential', label: 'Differential Backup' },
              ]}
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-2">
              Frequency
            </label>
            <Select
              value={newSchedule.frequency}
              onChange={(e) => setNewSchedule({ ...newSchedule, frequency: e.target.value })}
              options={[
                { value: 'daily', label: 'Daily' },
                { value: 'weekly', label: 'Weekly' },
                { value: 'monthly', label: 'Monthly' },
              ]}
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-2">
              Retention (days)
            </label>
            <Input
              type="number"
              value={newSchedule.retention_days}
              onChange={(e) => setNewSchedule({ ...newSchedule, retention_days: parseInt(e.target.value) || 30 })}
              min={1}
              max={365}
            />
            <p className="text-xs text-[var(--color-text-muted)] mt-1">
              Backups older than this will be automatically deleted
            </p>
          </div>

          <div className="flex gap-3 justify-end">
            <Button variant="ghost" onClick={() => setShowCreateScheduleModal(false)}>
              Cancel
            </Button>
            <Button onClick={handleCreateSchedule} isLoading={isSaving}>
              Create Schedule
            </Button>
          </div>
        </div>
      </Modal>

      {/* Restore Modal */}
      <Modal
        isOpen={showRestoreModal !== null}
        onClose={() => {
          setShowRestoreModal(null);
          setFormError(null);
        }}
        title="Restore Backup"
      >
        <div className="space-y-4">
          {formError && (
            <div className="p-3 rounded bg-[var(--color-error)]/20 text-[var(--color-error)] text-sm">
              {formError}
            </div>
          )}
          
          <div className="p-3 rounded bg-[var(--color-warning)]/20 text-[var(--color-warning)] text-sm">
            Warning: Restoring a backup may overwrite existing data. Make sure you have a current backup before proceeding.
          </div>

          <div>
            <label className="flex items-center gap-2 cursor-pointer">
              <input
                type="checkbox"
                checked={restoreOptions.overwrite}
                onChange={(e) => setRestoreOptions({ ...restoreOptions, overwrite: e.target.checked })}
                className="rounded border-[var(--color-border)]"
              />
              <span className="text-sm text-[var(--color-text-primary)]">
                Overwrite existing files
              </span>
            </label>
          </div>

          <div>
            <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-2">
              Destination (Optional)
            </label>
            <Input
              value={restoreOptions.destination}
              onChange={(e) => setRestoreOptions({ ...restoreOptions, destination: e.target.value })}
              placeholder="Leave empty to restore to original location"
            />
          </div>

          <div className="flex gap-3 justify-end">
            <Button variant="ghost" onClick={() => setShowRestoreModal(null)}>
              Cancel
            </Button>
            <Button onClick={handleRestore} isLoading={isSaving}>
              Restore
            </Button>
          </div>
        </div>
      </Modal>

      {/* Delete Confirmation Modal */}
      <Modal
        isOpen={showDeleteConfirm !== null}
        onClose={() => setShowDeleteConfirm(null)}
        title={`Delete ${showDeleteConfirm?.type === 'backup' ? 'Backup' : 'Schedule'}`}
      >
        <div className="space-y-4">
          <p className="text-[var(--color-text-secondary)]">
            Are you sure you want to delete this {showDeleteConfirm?.type}? This action cannot be undone.
          </p>
          <div className="flex gap-3 justify-end">
            <Button variant="ghost" onClick={() => setShowDeleteConfirm(null)}>
              Cancel
            </Button>
            <Button
              variant="danger"
              onClick={() => showDeleteConfirm && handleDelete(showDeleteConfirm.type, showDeleteConfirm.id)}
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
