import { useState, useEffect } from 'react';
import {
  Clock,
  Plus,
  Search,
  Play,
  Pause,
  Trash2,
  Edit2,
  Terminal,
  AlertCircle,
  CheckCircle,
  XCircle,
  RefreshCw,
  History,
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
import { cronService, userService } from '../../api/services';
import type { CronJob, CronJobLog, CronJobStatus, User } from '../../types';

const CRON_PRESETS = [
  { label: 'Every minute', value: '* * * * *' },
  { label: 'Every 5 minutes', value: '*/5 * * * *' },
  { label: 'Every 15 minutes', value: '*/15 * * * *' },
  { label: 'Every hour', value: '0 * * * *' },
  { label: 'Every 6 hours', value: '0 */6 * * *' },
  { label: 'Every day at midnight', value: '0 0 * * *' },
  { label: 'Every day at noon', value: '0 12 * * *' },
  { label: 'Every week (Sunday)', value: '0 0 * * 0' },
  { label: 'Every month (1st)', value: '0 0 1 * *' },
  { label: 'Custom', value: 'custom' },
];

export function AdminCronPage() {
  const [jobs, setJobs] = useState<CronJob[]>([]);
  const [users, setUsers] = useState<User[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState('');
  const [filterStatus, setFilterStatus] = useState<string>('all');
  const [filterUser, setFilterUser] = useState<string>('all');

  // Modal states
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [showEditModal, setShowEditModal] = useState(false);
  const [showLogsModal, setShowLogsModal] = useState(false);
  const [showDeleteConfirm, setShowDeleteConfirm] = useState<string | null>(null);
  
  // Form states
  const [editingJob, setEditingJob] = useState<CronJob | null>(null);
  const [newJob, setNewJob] = useState({
    name: '',
    command: '',
    schedule: '0 * * * *',
    user_id: '',
  });
  const [selectedPreset, setSelectedPreset] = useState('0 * * * *');
  const [customSchedule, setCustomSchedule] = useState('');
  const [formError, setFormError] = useState<string | null>(null);
  const [isSaving, setIsSaving] = useState(false);

  // Logs
  const [selectedJobLogs, setSelectedJobLogs] = useState<CronJob | null>(null);
  const [logs, setLogs] = useState<CronJobLog[]>([]);
  const [isLoadingLogs, setIsLoadingLogs] = useState(false);

  useEffect(() => {
    loadData();
  }, []);

  const loadData = async () => {
    try {
      setIsLoading(true);
      setError(null);
      const [jobsData, usersData] = await Promise.all([
        cronService.list(),
        userService.list(),
      ]);
      setJobs(jobsData);
      setUsers(usersData);
    } catch (err: any) {
      setError(err.message || 'Failed to load cron jobs');
    } finally {
      setIsLoading(false);
    }
  };

  const loadLogs = async (jobId: string) => {
    try {
      setIsLoadingLogs(true);
      const logsData = await cronService.getLogs(jobId, 20);
      setLogs(logsData);
    } catch (err: any) {
      console.error('Failed to load logs:', err);
    } finally {
      setIsLoadingLogs(false);
    }
  };

  const handleCreate = async () => {
    const schedule = selectedPreset === 'custom' ? customSchedule : selectedPreset;
    if (!newJob.name || !newJob.command || !schedule) {
      setFormError('Please fill in all required fields');
      return;
    }

    try {
      setIsSaving(true);
      setFormError(null);
      await cronService.create({
        ...newJob,
        schedule,
      });
      await loadData();
      setShowCreateModal(false);
      setNewJob({ name: '', command: '', schedule: '0 * * * *', user_id: '' });
      setSelectedPreset('0 * * * *');
      setCustomSchedule('');
    } catch (err: any) {
      setFormError(err.message || 'Failed to create job');
    } finally {
      setIsSaving(false);
    }
  };

  const handleUpdate = async () => {
    if (!editingJob) return;
    
    const schedule = selectedPreset === 'custom' ? customSchedule : selectedPreset;
    
    try {
      setIsSaving(true);
      setFormError(null);
      await cronService.update(editingJob.id, {
        name: editingJob.name,
        command: editingJob.command,
        schedule,
      });
      await loadData();
      setShowEditModal(false);
      setEditingJob(null);
    } catch (err: any) {
      setFormError(err.message || 'Failed to update job');
    } finally {
      setIsSaving(false);
    }
  };

  const handleDelete = async (id: string) => {
    try {
      setIsSaving(true);
      await cronService.delete(id);
      await loadData();
      setShowDeleteConfirm(null);
    } catch (err: any) {
      setError(err.message || 'Failed to delete job');
    } finally {
      setIsSaving(false);
    }
  };

  const handlePauseResume = async (job: CronJob) => {
    try {
      if (job.status === 'active') {
        await cronService.pause(job.id);
      } else {
        await cronService.resume(job.id);
      }
      await loadData();
    } catch (err: any) {
      setError(err.message || 'Failed to update job status');
    }
  };

  const handleRunNow = async (id: string) => {
    try {
      await cronService.run(id);
      await loadData();
    } catch (err: any) {
      setError(err.message || 'Failed to run job');
    }
  };

  const openEditModal = (job: CronJob) => {
    setEditingJob(job);
    const preset = CRON_PRESETS.find(p => p.value === job.schedule);
    if (preset) {
      setSelectedPreset(job.schedule);
      setCustomSchedule('');
    } else {
      setSelectedPreset('custom');
      setCustomSchedule(job.schedule);
    }
    setShowEditModal(true);
  };

  const openLogsModal = async (job: CronJob) => {
    setSelectedJobLogs(job);
    setShowLogsModal(true);
    await loadLogs(job.id);
  };

  const getStatusBadge = (status: CronJobStatus) => {
    const variants: Record<CronJobStatus, 'success' | 'warning' | 'default'> = {
      active: 'success',
      paused: 'warning',
      disabled: 'default',
    };
    return <Badge variant={variants[status]}>{status}</Badge>;
  };

  const getExitCodeBadge = (code: number) => {
    if (code === 0) {
      return <Badge variant="success" size="sm">Success</Badge>;
    }
    return <Badge variant="error" size="sm">Exit: {code}</Badge>;
  };

  const filteredJobs = jobs.filter(job => {
    const matchesSearch = 
      job.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      job.command.toLowerCase().includes(searchQuery.toLowerCase());
    const matchesStatus = filterStatus === 'all' || job.status === filterStatus;
    const matchesUser = filterUser === 'all' || job.user_id === filterUser;
    return matchesSearch && matchesStatus && matchesUser;
  });

  const stats = {
    total: jobs.length,
    active: jobs.filter(j => j.status === 'active').length,
    paused: jobs.filter(j => j.status === 'paused').length,
    failed: jobs.filter(j => j.last_exit_code && j.last_exit_code !== 0).length,
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
          <h1 className="text-2xl font-bold text-[var(--color-text-primary)]">Cron Job Management</h1>
          <p className="text-[var(--color-text-secondary)] mt-1">
            Manage scheduled tasks across all users
          </p>
        </div>
        <Button
          leftIcon={<Plus size={18} />}
          onClick={() => setShowCreateModal(true)}
        >
          Create Job
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

      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-[var(--color-text-secondary)]">Total Jobs</p>
                <p className="text-2xl font-bold text-[var(--color-text-primary)]">{stats.total}</p>
              </div>
              <Clock size={24} className="text-[#E7F0FA]" />
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
                <p className="text-sm text-[var(--color-text-secondary)]">Paused</p>
                <p className="text-2xl font-bold text-[var(--color-warning)]">{stats.paused}</p>
              </div>
              <Pause size={24} className="text-[var(--color-warning)]" />
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-[var(--color-text-secondary)]">Failed (Last Run)</p>
                <p className="text-2xl font-bold text-[var(--color-error)]">{stats.failed}</p>
              </div>
              <XCircle size={24} className="text-[var(--color-error)]" />
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Jobs List */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between flex-wrap gap-4">
            <CardTitle>Scheduled Jobs</CardTitle>
            <div className="flex gap-4 flex-wrap">
              <Select
                value={filterUser}
                onChange={(e) => setFilterUser(e.target.value)}
                options={[
                  { value: 'all', label: 'All Users' },
                  ...users.map(u => ({ value: u.id, label: u.username })),
                ]}
                className="w-40"
              />
              <Select
                value={filterStatus}
                onChange={(e) => setFilterStatus(e.target.value)}
                options={[
                  { value: 'all', label: 'All Status' },
                  { value: 'active', label: 'Active' },
                  { value: 'paused', label: 'Paused' },
                  { value: 'disabled', label: 'Disabled' },
                ]}
                className="w-40"
              />
              <div className="relative">
                <Search className="absolute left-3 top-1/2 -translate-y-1/2 text-[var(--color-text-muted)]" size={18} />
                <Input
                  placeholder="Search jobs..."
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  className="pl-10 w-64"
                />
              </div>
            </div>
          </div>
        </CardHeader>
        <CardContent>
          {filteredJobs.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-12 text-[var(--color-text-muted)]">
              <Clock size={48} className="mb-4 opacity-50" />
              <p>No cron jobs found</p>
              <Button
                className="mt-4"
                variant="outline"
                leftIcon={<Plus size={16} />}
                onClick={() => setShowCreateModal(true)}
              >
                Create First Job
              </Button>
            </div>
          ) : (
            <div className="space-y-3">
              {filteredJobs.map(job => {
                const user = users.find(u => u.id === job.user_id);
                
                return (
                  <div
                    key={job.id}
                    className="p-4 rounded-lg bg-[var(--color-primary-dark)]/50 hover:bg-[var(--color-primary-dark)] transition-colors"
                  >
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-4 flex-1 min-w-0">
                        <div className="p-2 rounded-lg bg-[#7BA4D0]/20">
                          <Terminal size={20} className="text-[#E7F0FA]" />
                        </div>
                        <div className="flex-1 min-w-0">
                          <div className="flex items-center gap-2 flex-wrap">
                            <h3 className="font-medium text-[var(--color-text-primary)]">
                              {job.name}
                            </h3>
                            {getStatusBadge(job.status)}
                            {job.last_exit_code !== undefined && job.last_exit_code !== null && (
                              getExitCodeBadge(job.last_exit_code)
                            )}
                          </div>
                          <div className="flex items-center gap-4 mt-1 text-sm text-[var(--color-text-muted)] flex-wrap">
                            <code className="px-2 py-0.5 rounded bg-[var(--color-primary-dark)] font-mono text-xs">
                              {job.schedule}
                            </code>
                            {user && <span>User: {user.username}</span>}
                            {job.next_run_at && (
                              <span>Next: {new Date(job.next_run_at).toLocaleString()}</span>
                            )}
                          </div>
                          <div className="mt-1">
                            <code className="text-xs text-[var(--color-text-secondary)] font-mono truncate block">
                              {job.command}
                            </code>
                          </div>
                        </div>
                      </div>
                      
                      <div className="flex items-center gap-2 ml-4">
                        <Button
                          size="sm"
                          variant="ghost"
                          onClick={() => handleRunNow(job.id)}
                          title="Run now"
                        >
                          <Play size={14} />
                        </Button>
                        <Button
                          size="sm"
                          variant="ghost"
                          onClick={() => handlePauseResume(job)}
                          title={job.status === 'active' ? 'Pause' : 'Resume'}
                        >
                          {job.status === 'active' ? <Pause size={14} /> : <RefreshCw size={14} />}
                        </Button>
                        <Button
                          size="sm"
                          variant="ghost"
                          onClick={() => openLogsModal(job)}
                          title="View logs"
                        >
                          <History size={14} />
                        </Button>
                        <Button
                          size="sm"
                          variant="ghost"
                          onClick={() => openEditModal(job)}
                        >
                          <Edit2 size={14} />
                        </Button>
                        <Button
                          size="sm"
                          variant="ghost"
                          className="text-[var(--color-error)]"
                          onClick={() => setShowDeleteConfirm(job.id)}
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

      {/* Create Job Modal */}
      <Modal
        isOpen={showCreateModal}
        onClose={() => {
          setShowCreateModal(false);
          setFormError(null);
        }}
        title="Create Cron Job"
      >
        <div className="space-y-4">
          {formError && (
            <div className="p-3 rounded bg-[var(--color-error)]/20 text-[var(--color-error)] text-sm">
              {formError}
            </div>
          )}
          
          <div>
            <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-2">
              Job Name
            </label>
            <Input
              value={newJob.name}
              onChange={(e) => setNewJob({ ...newJob, name: e.target.value })}
              placeholder="Daily backup"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-2">
              User (Optional)
            </label>
            <Select
              value={newJob.user_id}
              onChange={(e) => setNewJob({ ...newJob, user_id: e.target.value })}
              options={[
                { value: '', label: 'System (root)' },
                ...users.map(u => ({ value: u.id, label: u.username })),
              ]}
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-2">
              Schedule
            </label>
            <Select
              value={selectedPreset}
              onChange={(e) => setSelectedPreset(e.target.value)}
              options={CRON_PRESETS}
            />
            {selectedPreset === 'custom' && (
              <Input
                className="mt-2"
                value={customSchedule}
                onChange={(e) => setCustomSchedule(e.target.value)}
                placeholder="* * * * * (min hour day month weekday)"
              />
            )}
            <p className="text-xs text-[var(--color-text-muted)] mt-1">
              Format: minute hour day month weekday
            </p>
          </div>

          <div>
            <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-2">
              Command
            </label>
            <textarea
              value={newJob.command}
              onChange={(e) => setNewJob({ ...newJob, command: e.target.value })}
              placeholder="/usr/bin/php /home/user/script.php"
              className="w-full h-24 px-4 py-2.5 rounded-lg bg-[var(--color-surface)] border border-[var(--color-border)] text-[var(--color-text-primary)] font-mono text-sm"
            />
          </div>

          <div className="flex gap-3 justify-end">
            <Button variant="ghost" onClick={() => setShowCreateModal(false)}>
              Cancel
            </Button>
            <Button onClick={handleCreate} isLoading={isSaving}>
              Create Job
            </Button>
          </div>
        </div>
      </Modal>

      {/* Edit Job Modal */}
      <Modal
        isOpen={showEditModal}
        onClose={() => {
          setShowEditModal(false);
          setEditingJob(null);
          setFormError(null);
        }}
        title="Edit Cron Job"
      >
        {editingJob && (
          <div className="space-y-4">
            {formError && (
              <div className="p-3 rounded bg-[var(--color-error)]/20 text-[var(--color-error)] text-sm">
                {formError}
              </div>
            )}
            
            <div>
              <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-2">
                Job Name
              </label>
              <Input
                value={editingJob.name}
                onChange={(e) => setEditingJob({ ...editingJob, name: e.target.value })}
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-2">
                Schedule
              </label>
              <Select
                value={selectedPreset}
                onChange={(e) => setSelectedPreset(e.target.value)}
                options={CRON_PRESETS}
              />
              {selectedPreset === 'custom' && (
                <Input
                  className="mt-2"
                  value={customSchedule}
                  onChange={(e) => setCustomSchedule(e.target.value)}
                  placeholder="* * * * *"
                />
              )}
            </div>

            <div>
              <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-2">
                Command
              </label>
              <textarea
                value={editingJob.command}
                onChange={(e) => setEditingJob({ ...editingJob, command: e.target.value })}
                className="w-full h-24 px-4 py-2.5 rounded-lg bg-[var(--color-surface)] border border-[var(--color-border)] text-[var(--color-text-primary)] font-mono text-sm"
              />
            </div>

            <div className="flex gap-3 justify-end">
              <Button variant="ghost" onClick={() => setShowEditModal(false)}>
                Cancel
              </Button>
              <Button onClick={handleUpdate} isLoading={isSaving}>
                Save Changes
              </Button>
            </div>
          </div>
        )}
      </Modal>

      {/* Logs Modal */}
      <Modal
        isOpen={showLogsModal}
        onClose={() => {
          setShowLogsModal(false);
          setSelectedJobLogs(null);
          setLogs([]);
        }}
        title={`Execution Logs - ${selectedJobLogs?.name}`}
      >
        <div className="space-y-4">
          {isLoadingLogs ? (
            <div className="space-y-2">
              {[1, 2, 3].map(i => (
                <div key={i} className="h-16 bg-[var(--color-primary-dark)]/50 rounded animate-pulse" />
              ))}
            </div>
          ) : logs.length === 0 ? (
            <p className="text-center text-[var(--color-text-muted)] py-8">
              No execution logs found
            </p>
          ) : (
            <div className="space-y-2 max-h-96 overflow-y-auto">
              {logs.map(log => (
                <div
                  key={log.id}
                  className="p-3 rounded-lg bg-[var(--color-primary-dark)]/50"
                >
                  <div className="flex items-center justify-between mb-2">
                    <div className="flex items-center gap-2">
                      {log.exit_code === 0 ? (
                        <CheckCircle size={16} className="text-[var(--color-success)]" />
                      ) : (
                        <XCircle size={16} className="text-[var(--color-error)]" />
                      )}
                      <span className="text-sm text-[var(--color-text-primary)]">
                        {new Date(log.started_at).toLocaleString()}
                      </span>
                    </div>
                    {getExitCodeBadge(log.exit_code)}
                  </div>
                  {log.stdout && (
                    <div className="mt-2">
                      <p className="text-xs text-[var(--color-text-muted)] mb-1">Output:</p>
                      <pre className="text-xs bg-[var(--color-primary-dark)] p-2 rounded overflow-x-auto text-[var(--color-text-secondary)]">
                        {log.stdout.slice(0, 500)}{log.stdout.length > 500 ? '...' : ''}
                      </pre>
                    </div>
                  )}
                  {log.stderr && (
                    <div className="mt-2">
                      <p className="text-xs text-[var(--color-error)] mb-1">Error:</p>
                      <pre className="text-xs bg-[var(--color-error)]/10 p-2 rounded overflow-x-auto text-[var(--color-error)]">
                        {log.stderr.slice(0, 500)}{log.stderr.length > 500 ? '...' : ''}
                      </pre>
                    </div>
                  )}
                </div>
              ))}
            </div>
          )}
          
          <div className="flex justify-end">
            <Button variant="ghost" onClick={() => setShowLogsModal(false)}>
              Close
            </Button>
          </div>
        </div>
      </Modal>

      {/* Delete Confirmation Modal */}
      <Modal
        isOpen={showDeleteConfirm !== null}
        onClose={() => setShowDeleteConfirm(null)}
        title="Delete Cron Job"
      >
        <div className="space-y-4">
          <p className="text-[var(--color-text-secondary)]">
            Are you sure you want to delete this cron job? This action cannot be undone.
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
