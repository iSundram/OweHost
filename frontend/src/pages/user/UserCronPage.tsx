import { useState, useEffect } from 'react';
import {
  Clock,
  Plus,
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
import { cronService } from '../../api/services';
import type { CronJob, CronJobLog, CronJobStatus } from '../../types';

const CRON_PRESETS = [
  { label: 'Every minute', value: '* * * * *' },
  { label: 'Every 5 minutes', value: '*/5 * * * *' },
  { label: 'Every 15 minutes', value: '*/15 * * * *' },
  { label: 'Every hour', value: '0 * * * *' },
  { label: 'Every day at midnight', value: '0 0 * * *' },
  { label: 'Every week (Sunday)', value: '0 0 * * 0' },
  { label: 'Custom', value: 'custom' },
];

export function UserCronPage() {
  const [jobs, setJobs] = useState<CronJob[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Modal states
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [showEditModal, setShowEditModal] = useState(false);
  const [showLogsModal, setShowLogsModal] = useState(false);
  const [showDeleteConfirm, setShowDeleteConfirm] = useState<string | null>(null);
  
  // Form states
  const [editingJob, setEditingJob] = useState<CronJob | null>(null);
  const [newJob, setNewJob] = useState({ name: '', command: '', schedule: '0 * * * *' });
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
      const jobsData = await cronService.list();
      setJobs(jobsData);
    } catch (err: any) {
      setError(err.message || 'Failed to load cron jobs');
    } finally {
      setIsLoading(false);
    }
  };

  const loadLogs = async (jobId: string) => {
    try {
      setIsLoadingLogs(true);
      const logsData = await cronService.getLogs(jobId, 10);
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
      setFormError('Please fill in all fields');
      return;
    }

    try {
      setIsSaving(true);
      setFormError(null);
      await cronService.create({ ...newJob, schedule });
      await loadData();
      setShowCreateModal(false);
      setNewJob({ name: '', command: '', schedule: '0 * * * *' });
      setSelectedPreset('0 * * * *');
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
      await cronService.update(editingJob.id, { name: editingJob.name, command: editingJob.command, schedule });
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
      setError(err.message || 'Failed to update job');
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
    setSelectedPreset(preset ? job.schedule : 'custom');
    setCustomSchedule(preset ? '' : job.schedule);
    setShowEditModal(true);
  };

  const openLogsModal = async (job: CronJob) => {
    setSelectedJobLogs(job);
    setShowLogsModal(true);
    await loadLogs(job.id);
  };

  const getStatusBadge = (status: CronJobStatus) => {
    const variants: Record<CronJobStatus, 'success' | 'warning' | 'default'> = {
      active: 'success', paused: 'warning', disabled: 'default',
    };
    return <Badge variant={variants[status]}>{status}</Badge>;
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
          <h1 className="text-2xl font-bold text-[var(--color-text-primary)]">Cron Jobs</h1>
          <p className="text-[var(--color-text-secondary)] mt-1">Schedule and manage automated tasks</p>
        </div>
        <Button leftIcon={<Plus size={18} />} onClick={() => setShowCreateModal(true)}>
          Create Job
        </Button>
      </div>

      {error && (
        <div className="flex items-center gap-2 p-4 rounded-lg bg-[var(--color-error)]/20 text-[var(--color-error)]">
          <AlertCircle size={20} />
          <span>{error}</span>
          <button onClick={() => setError(null)} className="ml-auto">×</button>
        </div>
      )}

      <Card>
        <CardHeader><CardTitle>Your Scheduled Jobs</CardTitle></CardHeader>
        <CardContent>
          {jobs.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-12 text-[var(--color-text-muted)]">
              <Clock size={48} className="mb-4 opacity-50" />
              <p>No cron jobs configured</p>
              <Button className="mt-4" variant="outline" leftIcon={<Plus size={16} />} onClick={() => setShowCreateModal(true)}>
                Create First Job
              </Button>
            </div>
          ) : (
            <div className="space-y-3">
              {jobs.map(job => (
                <div key={job.id} className="p-4 rounded-lg bg-[var(--color-primary-dark)]/50">
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-4 flex-1 min-w-0">
                      <Terminal size={20} className="text-[#E7F0FA]" />
                      <div className="flex-1 min-w-0">
                        <div className="flex items-center gap-2">
                          <h3 className="font-medium text-[var(--color-text-primary)]">{job.name}</h3>
                          {getStatusBadge(job.status)}
                          {job.last_exit_code !== undefined && job.last_exit_code !== 0 && (
                            <Badge variant="error" size="sm">Failed</Badge>
                          )}
                        </div>
                        <code className="text-xs text-[var(--color-text-muted)] mt-1 block">{job.schedule} - {job.command}</code>
                        {job.next_run_at && (
                          <p className="text-xs text-[var(--color-text-muted)]">Next: {new Date(job.next_run_at).toLocaleString()}</p>
                        )}
                      </div>
                    </div>
                    <div className="flex items-center gap-2">
                      <Button size="sm" variant="ghost" onClick={() => handleRunNow(job.id)} title="Run now"><Play size={14} /></Button>
                      <Button size="sm" variant="ghost" onClick={() => handlePauseResume(job)} title={job.status === 'active' ? 'Pause' : 'Resume'}>
                        {job.status === 'active' ? <Pause size={14} /> : <RefreshCw size={14} />}
                      </Button>
                      <Button size="sm" variant="ghost" onClick={() => openLogsModal(job)} title="Logs"><History size={14} /></Button>
                      <Button size="sm" variant="ghost" onClick={() => openEditModal(job)}><Edit2 size={14} /></Button>
                      <Button size="sm" variant="ghost" className="text-[var(--color-error)]" onClick={() => setShowDeleteConfirm(job.id)}><Trash2 size={14} /></Button>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      {/* Create Modal */}
      <Modal isOpen={showCreateModal} onClose={() => setShowCreateModal(false)} title="Create Cron Job">
        <div className="space-y-4">
          {formError && <div className="p-3 rounded bg-[var(--color-error)]/20 text-[var(--color-error)] text-sm">{formError}</div>}
          <div>
            <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-2">Job Name</label>
            <Input value={newJob.name} onChange={(e) => setNewJob({ ...newJob, name: e.target.value })} placeholder="Daily backup" />
          </div>
          <div>
            <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-2">Schedule</label>
            <Select value={selectedPreset} onChange={(e) => setSelectedPreset(e.target.value)} options={CRON_PRESETS} />
            {selectedPreset === 'custom' && (
              <Input className="mt-2" value={customSchedule} onChange={(e) => setCustomSchedule(e.target.value)} placeholder="* * * * *" />
            )}
          </div>
          <div>
            <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-2">Command</label>
            <textarea value={newJob.command} onChange={(e) => setNewJob({ ...newJob, command: e.target.value })} placeholder="/usr/bin/php ~/script.php"
              className="w-full h-20 px-4 py-2.5 rounded-lg bg-[var(--color-surface)] border border-[var(--color-border)] text-[var(--color-text-primary)] font-mono text-sm" />
          </div>
          <div className="flex gap-3 justify-end">
            <Button variant="ghost" onClick={() => setShowCreateModal(false)}>Cancel</Button>
            <Button onClick={handleCreate} isLoading={isSaving}>Create Job</Button>
          </div>
        </div>
      </Modal>

      {/* Edit Modal */}
      <Modal isOpen={showEditModal} onClose={() => setShowEditModal(false)} title="Edit Cron Job">
        {editingJob && (
          <div className="space-y-4">
            {formError && <div className="p-3 rounded bg-[var(--color-error)]/20 text-[var(--color-error)] text-sm">{formError}</div>}
            <div>
              <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-2">Job Name</label>
              <Input value={editingJob.name} onChange={(e) => setEditingJob({ ...editingJob, name: e.target.value })} />
            </div>
            <div>
              <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-2">Schedule</label>
              <Select value={selectedPreset} onChange={(e) => setSelectedPreset(e.target.value)} options={CRON_PRESETS} />
              {selectedPreset === 'custom' && (
                <Input className="mt-2" value={customSchedule} onChange={(e) => setCustomSchedule(e.target.value)} />
              )}
            </div>
            <div>
              <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-2">Command</label>
              <textarea value={editingJob.command} onChange={(e) => setEditingJob({ ...editingJob, command: e.target.value })}
                className="w-full h-20 px-4 py-2.5 rounded-lg bg-[var(--color-surface)] border border-[var(--color-border)] text-[var(--color-text-primary)] font-mono text-sm" />
            </div>
            <div className="flex gap-3 justify-end">
              <Button variant="ghost" onClick={() => setShowEditModal(false)}>Cancel</Button>
              <Button onClick={handleUpdate} isLoading={isSaving}>Save</Button>
            </div>
          </div>
        )}
      </Modal>

      {/* Logs Modal */}
      <Modal isOpen={showLogsModal} onClose={() => { setShowLogsModal(false); setSelectedJobLogs(null); }} title={`Logs - ${selectedJobLogs?.name}`}>
        <div className="space-y-4">
          {isLoadingLogs ? (
            <div className="h-32 bg-[var(--color-primary-dark)]/50 rounded animate-pulse" />
          ) : logs.length === 0 ? (
            <p className="text-center text-[var(--color-text-muted)] py-8">No logs found</p>
          ) : (
            <div className="space-y-2 max-h-80 overflow-y-auto">
              {logs.map(log => (
                <div key={log.id} className="p-3 rounded-lg bg-[var(--color-primary-dark)]/50">
                  <div className="flex items-center gap-2 mb-1">
                    {log.exit_code === 0 ? <CheckCircle size={14} className="text-[var(--color-success)]" /> : <XCircle size={14} className="text-[var(--color-error)]" />}
                    <span className="text-sm text-[var(--color-text-primary)]">{new Date(log.started_at).toLocaleString()}</span>
                    <Badge variant={log.exit_code === 0 ? 'success' : 'error'} size="sm">Exit: {log.exit_code}</Badge>
                  </div>
                  {log.stdout && <pre className="text-xs bg-[var(--color-primary-dark)] p-2 rounded mt-2">{log.stdout.slice(0, 300)}</pre>}
                  {log.stderr && <pre className="text-xs bg-[var(--color-error)]/10 text-[var(--color-error)] p-2 rounded mt-2">{log.stderr.slice(0, 300)}</pre>}
                </div>
              ))}
            </div>
          )}
          <div className="flex justify-end"><Button variant="ghost" onClick={() => setShowLogsModal(false)}>Close</Button></div>
        </div>
      </Modal>

      {/* Delete Confirm */}
      <Modal isOpen={showDeleteConfirm !== null} onClose={() => setShowDeleteConfirm(null)} title="Delete Cron Job">
        <div className="space-y-4">
          <p className="text-[var(--color-text-secondary)]">Are you sure you want to delete this job?</p>
          <div className="flex gap-3 justify-end">
            <Button variant="ghost" onClick={() => setShowDeleteConfirm(null)}>Cancel</Button>
            <Button variant="danger" onClick={() => showDeleteConfirm && handleDelete(showDeleteConfirm)} isLoading={isSaving}>Delete</Button>
          </div>
        </div>
      </Modal>
    </div>
  );
}
