import { useState, useEffect } from 'react';
import {
  Globe,
  Plus,
  Search,
  Trash2,
  Edit2,
  AlertCircle,
  Copy,
  Lock,
  Unlock,
  RefreshCw,
  ChevronDown,
  ChevronRight,
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
import { dnsService, domainService } from '../../api/services';
import type { DNSZone, DNSRecord, DNSRecordType, Domain } from '../../types';

const RECORD_TYPES: DNSRecordType[] = ['A', 'AAAA', 'CNAME', 'MX', 'TXT', 'SRV'];

const DEFAULT_TTL_OPTIONS = [
  { value: 300, label: '5 minutes' },
  { value: 3600, label: '1 hour' },
  { value: 86400, label: '1 day' },
];

export function ResellerDNSPage() {
  const [zones, setZones] = useState<DNSZone[]>([]);
  const [domains, setDomains] = useState<Domain[]>([]);
  const [selectedZone, setSelectedZone] = useState<DNSZone | null>(null);
  const [records, setRecords] = useState<DNSRecord[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isLoadingRecords, setIsLoadingRecords] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState('');
  const [expandedZones, setExpandedZones] = useState<Set<string>>(new Set());
  
  const [showCreateRecordModal, setShowCreateRecordModal] = useState(false);
  const [showEditRecordModal, setShowEditRecordModal] = useState(false);
  const [showDeleteConfirm, setShowDeleteConfirm] = useState<string | null>(null);
  
  const [newRecord, setNewRecord] = useState<{ name: string; type: DNSRecordType; content: string; ttl: number; priority?: number }>({ 
    name: '', type: 'A', content: '', ttl: 3600 
  });
  const [editingRecord, setEditingRecord] = useState<DNSRecord | null>(null);
  const [formError, setFormError] = useState<string | null>(null);
  const [isSaving, setIsSaving] = useState(false);

  useEffect(() => { loadData(); }, []);

  const loadData = async () => {
    try {
      setIsLoading(true);
      const [zonesData, domainsData] = await Promise.all([dnsService.listZones(), domainService.list()]);
      setZones(zonesData);
      setDomains(domainsData);
    } catch (err: any) {
      setError(err.message || 'Failed to load data');
    } finally {
      setIsLoading(false);
    }
  };

  const loadRecords = async (zoneId: string) => {
    try {
      setIsLoadingRecords(true);
      const recordsData = await dnsService.listRecords(zoneId);
      setRecords(recordsData);
    } catch (err: any) {
      console.error('Failed to load records:', err);
    } finally {
      setIsLoadingRecords(false);
    }
  };

  const handleZoneClick = async (zone: DNSZone) => {
    setSelectedZone(zone);
    await loadRecords(zone.id);
  };

  const handleCreateRecord = async () => {
    if (!selectedZone || !newRecord.name || !newRecord.content) {
      setFormError('Please fill in all fields');
      return;
    }
    try {
      setIsSaving(true);
      await dnsService.createRecord(selectedZone.id, newRecord);
      await loadRecords(selectedZone.id);
      setShowCreateRecordModal(false);
      setNewRecord({ name: '', type: 'A', content: '', ttl: 3600 });
    } catch (err: any) {
      setFormError(err.message || 'Failed to create');
    } finally {
      setIsSaving(false);
    }
  };

  const handleUpdateRecord = async () => {
    if (!editingRecord) return;
    try {
      setIsSaving(true);
      await dnsService.updateRecord(editingRecord.id, {
        name: editingRecord.name, type: editingRecord.type, content: editingRecord.content,
        ttl: editingRecord.ttl, priority: editingRecord.priority
      });
      if (selectedZone) await loadRecords(selectedZone.id);
      setShowEditRecordModal(false);
      setEditingRecord(null);
    } catch (err: any) {
      setFormError(err.message || 'Failed to update');
    } finally {
      setIsSaving(false);
    }
  };

  const handleDeleteRecord = async (id: string) => {
    try {
      setIsSaving(true);
      await dnsService.deleteRecord(id);
      if (selectedZone) await loadRecords(selectedZone.id);
      setShowDeleteConfirm(null);
    } catch (err: any) {
      setError(err.message || 'Failed to delete');
    } finally {
      setIsSaving(false);
    }
  };

  const getRecordTypeColor = (type: DNSRecordType) => {
    const colors: Record<string, string> = {
      A: 'bg-blue-500/20 text-blue-400', AAAA: 'bg-purple-500/20 text-purple-400',
      CNAME: 'bg-green-500/20 text-green-400', MX: 'bg-orange-500/20 text-orange-400',
      TXT: 'bg-yellow-500/20 text-yellow-400', SRV: 'bg-pink-500/20 text-pink-400',
    };
    return colors[type] || 'bg-gray-500/20 text-gray-400';
  };

  const filteredZones = zones.filter(z => z.name.toLowerCase().includes(searchQuery.toLowerCase()));

  if (isLoading) return <div className="h-96 bg-[var(--color-primary)]/50 rounded animate-pulse" />;

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-[var(--color-text-primary)]">DNS Management</h1>
          <p className="text-[var(--color-text-secondary)] mt-1">Manage DNS for customer domains</p>
        </div>
      </div>

      {error && (
        <div className="flex items-center gap-2 p-4 rounded-lg bg-[var(--color-error)]/20 text-[var(--color-error)]">
          <AlertCircle size={20} /><span>{error}</span>
          <button onClick={() => setError(null)} className="ml-auto">×</button>
        </div>
      )}

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <Card>
          <CardHeader><CardTitle className="flex items-center gap-2"><Globe size={20} />DNS Zones ({zones.length})</CardTitle></CardHeader>
          <CardContent className="space-y-4">
            <div className="relative">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 text-[var(--color-text-muted)]" size={18} />
              <Input placeholder="Search zones..." value={searchQuery} onChange={(e) => setSearchQuery(e.target.value)} className="pl-10" />
            </div>
            <div className="space-y-2 max-h-[500px] overflow-y-auto">
              {filteredZones.map(zone => (
                <div key={zone.id}
                  className={`p-3 rounded-lg cursor-pointer transition-colors ${selectedZone?.id === zone.id ? 'bg-[var(--color-accent)]/20 border border-[var(--color-accent)]' : 'bg-[var(--color-primary-dark)]/50 hover:bg-[var(--color-primary-dark)]'}`}
                  onClick={() => handleZoneClick(zone)}>
                  <div className="flex items-center justify-between">
                    <span className="font-medium text-[var(--color-text-primary)]">{zone.name}</span>
                    {zone.locked && <Lock size={14} className="text-[var(--color-warning)]" />}
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>

        <div className="lg:col-span-2">
          <Card>
            <CardHeader>
              <div className="flex items-center justify-between">
                <CardTitle>{selectedZone ? `Records - ${selectedZone.name}` : 'Select a zone'}</CardTitle>
                {selectedZone && !selectedZone.locked && (
                  <Button size="sm" leftIcon={<Plus size={16} />} onClick={() => setShowCreateRecordModal(true)}>Add Record</Button>
                )}
              </div>
            </CardHeader>
            <CardContent>
              {!selectedZone ? (
                <div className="flex flex-col items-center justify-center py-12 text-[var(--color-text-muted)]">
                  <Globe size={48} className="mb-4 opacity-50" />
                  <p>Select a zone to manage records</p>
                </div>
              ) : isLoadingRecords ? (
                <div className="space-y-2">{[1,2,3].map(i => <div key={i} className="h-12 bg-[var(--color-primary-dark)]/50 rounded animate-pulse" />)}</div>
              ) : records.length === 0 ? (
                <div className="text-center py-12 text-[var(--color-text-muted)]">No records</div>
              ) : (
                <div className="space-y-2">
                  {records.map(record => (
                    <div key={record.id} className="grid grid-cols-12 gap-4 px-3 py-3 rounded-lg bg-[var(--color-primary-dark)]/50 items-center">
                      <div className="col-span-1"><span className={`px-2 py-1 rounded text-xs font-medium ${getRecordTypeColor(record.type)}`}>{record.type}</span></div>
                      <div className="col-span-3 font-mono text-sm truncate">{record.name}</div>
                      <div className="col-span-4 font-mono text-sm truncate text-[var(--color-text-secondary)]">{record.content}</div>
                      <div className="col-span-2 text-sm text-[var(--color-text-muted)]">{record.ttl}s</div>
                      <div className="col-span-2 flex gap-1">
                        {!selectedZone?.locked && (
                          <>
                            <Button size="sm" variant="ghost" onClick={() => { setEditingRecord(record); setShowEditRecordModal(true); }}><Edit2 size={14} /></Button>
                            <Button size="sm" variant="ghost" className="text-[var(--color-error)]" onClick={() => setShowDeleteConfirm(record.id)}><Trash2 size={14} /></Button>
                          </>
                        )}
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </CardContent>
          </Card>
        </div>
      </div>

      {/* Create Record Modal */}
      <Modal isOpen={showCreateRecordModal} onClose={() => setShowCreateRecordModal(false)} title="Add DNS Record">
        <div className="space-y-4">
          {formError && <div className="p-3 rounded bg-[var(--color-error)]/20 text-[var(--color-error)] text-sm">{formError}</div>}
          <Select value={newRecord.type} onChange={(e) => setNewRecord({...newRecord, type: e.target.value as DNSRecordType})} options={RECORD_TYPES.map(t => ({value: t, label: t}))} label="Type" />
          <Input value={newRecord.name} onChange={(e) => setNewRecord({...newRecord, name: e.target.value})} placeholder="@ or subdomain" label="Name" />
          <Input value={newRecord.content} onChange={(e) => setNewRecord({...newRecord, content: e.target.value})} placeholder="Value" label="Content" />
          {(newRecord.type === 'MX' || newRecord.type === 'SRV') && (
            <Input type="number" value={newRecord.priority || ''} onChange={(e) => setNewRecord({...newRecord, priority: parseInt(e.target.value)})} placeholder="10" label="Priority" />
          )}
          <Select value={newRecord.ttl.toString()} onChange={(e) => setNewRecord({...newRecord, ttl: parseInt(e.target.value)})} options={DEFAULT_TTL_OPTIONS.map(o => ({value: o.value.toString(), label: o.label}))} label="TTL" />
          <div className="flex gap-3 justify-end">
            <Button variant="ghost" onClick={() => setShowCreateRecordModal(false)}>Cancel</Button>
            <Button onClick={handleCreateRecord} isLoading={isSaving}>Add</Button>
          </div>
        </div>
      </Modal>

      {/* Edit Record Modal */}
      <Modal isOpen={showEditRecordModal} onClose={() => setShowEditRecordModal(false)} title="Edit Record">
        {editingRecord && (
          <div className="space-y-4">
            {formError && <div className="p-3 rounded bg-[var(--color-error)]/20 text-[var(--color-error)] text-sm">{formError}</div>}
            <Select value={editingRecord.type} onChange={(e) => setEditingRecord({...editingRecord, type: e.target.value as DNSRecordType})} options={RECORD_TYPES.map(t => ({value: t, label: t}))} label="Type" />
            <Input value={editingRecord.name} onChange={(e) => setEditingRecord({...editingRecord, name: e.target.value})} label="Name" />
            <Input value={editingRecord.content} onChange={(e) => setEditingRecord({...editingRecord, content: e.target.value})} label="Content" />
            <Select value={editingRecord.ttl.toString()} onChange={(e) => setEditingRecord({...editingRecord, ttl: parseInt(e.target.value)})} options={DEFAULT_TTL_OPTIONS.map(o => ({value: o.value.toString(), label: o.label}))} label="TTL" />
            <div className="flex gap-3 justify-end">
              <Button variant="ghost" onClick={() => setShowEditRecordModal(false)}>Cancel</Button>
              <Button onClick={handleUpdateRecord} isLoading={isSaving}>Save</Button>
            </div>
          </div>
        )}
      </Modal>

      {/* Delete Confirm */}
      <Modal isOpen={showDeleteConfirm !== null} onClose={() => setShowDeleteConfirm(null)} title="Delete Record">
        <div className="space-y-4">
          <p className="text-[var(--color-text-secondary)]">Delete this record?</p>
          <div className="flex gap-3 justify-end">
            <Button variant="ghost" onClick={() => setShowDeleteConfirm(null)}>Cancel</Button>
            <Button variant="danger" onClick={() => showDeleteConfirm && handleDeleteRecord(showDeleteConfirm)} isLoading={isSaving}>Delete</Button>
          </div>
        </div>
      </Modal>
    </div>
  );
}
