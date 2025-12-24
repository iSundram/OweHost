import { useState, useEffect } from 'react';
import {
  Globe,
  Plus,
  Search,
  Shield,
  Lock,
  Unlock,
  RefreshCw,
  Trash2,
  Edit2,
  ChevronDown,
  ChevronRight,
  AlertCircle,
  CheckCircle,
  Copy,
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

const RECORD_TYPES: DNSRecordType[] = ['A', 'AAAA', 'CNAME', 'MX', 'TXT', 'SRV', 'NS', 'CAA', 'PTR'];

const DEFAULT_TTL_OPTIONS = [
  { value: 300, label: '5 minutes' },
  { value: 600, label: '10 minutes' },
  { value: 1800, label: '30 minutes' },
  { value: 3600, label: '1 hour' },
  { value: 14400, label: '4 hours' },
  { value: 86400, label: '1 day' },
];

export function AdminDNSPage() {
  const [zones, setZones] = useState<DNSZone[]>([]);
  const [domains, setDomains] = useState<Domain[]>([]);
  const [selectedZone, setSelectedZone] = useState<DNSZone | null>(null);
  const [records, setRecords] = useState<DNSRecord[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isLoadingRecords, setIsLoadingRecords] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState('');
  
  // Modal states
  const [showCreateZoneModal, setShowCreateZoneModal] = useState(false);
  const [showCreateRecordModal, setShowCreateRecordModal] = useState(false);
  const [showEditRecordModal, setShowEditRecordModal] = useState(false);
  const [showDeleteConfirm, setShowDeleteConfirm] = useState<{ type: 'zone' | 'record'; id: string } | null>(null);
  
  // Form states
  const [newZone, setNewZone] = useState({ domain_id: '', name: '' });
  const [newRecord, setNewRecord] = useState<{
    name: string;
    type: DNSRecordType;
    content: string;
    ttl: number;
    priority?: number;
  }>({ name: '', type: 'A', content: '', ttl: 3600 });
  const [editingRecord, setEditingRecord] = useState<DNSRecord | null>(null);
  const [formError, setFormError] = useState<string | null>(null);
  const [isSaving, setIsSaving] = useState(false);

  // Expanded zones in the list
  const [expandedZones, setExpandedZones] = useState<Set<string>>(new Set());

  useEffect(() => {
    loadData();
  }, []);

  const loadData = async () => {
    try {
      setIsLoading(true);
      setError(null);
      const [zonesData, domainsData] = await Promise.all([
        dnsService.listZones(),
        domainService.list(),
      ]);
      setZones(zonesData);
      setDomains(domainsData);
    } catch (err: any) {
      setError(err.message || 'Failed to load DNS data');
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

  const toggleZoneExpand = (zoneId: string) => {
    setExpandedZones(prev => {
      const next = new Set(prev);
      if (next.has(zoneId)) {
        next.delete(zoneId);
      } else {
        next.add(zoneId);
      }
      return next;
    });
  };

  const handleCreateZone = async () => {
    if (!newZone.domain_id || !newZone.name) {
      setFormError('Please fill in all fields');
      return;
    }
    
    try {
      setIsSaving(true);
      setFormError(null);
      await dnsService.createZone(newZone);
      await loadData();
      setShowCreateZoneModal(false);
      setNewZone({ domain_id: '', name: '' });
    } catch (err: any) {
      setFormError(err.message || 'Failed to create zone');
    } finally {
      setIsSaving(false);
    }
  };

  const handleDeleteZone = async (id: string) => {
    try {
      setIsSaving(true);
      await dnsService.deleteZone(id);
      await loadData();
      if (selectedZone?.id === id) {
        setSelectedZone(null);
        setRecords([]);
      }
      setShowDeleteConfirm(null);
    } catch (err: any) {
      setError(err.message || 'Failed to delete zone');
    } finally {
      setIsSaving(false);
    }
  };

  const handleLockZone = async (id: string, lock: boolean) => {
    try {
      if (lock) {
        await dnsService.lockZone(id);
      } else {
        await dnsService.unlockZone(id);
      }
      await loadData();
    } catch (err: any) {
      setError(err.message || 'Failed to update zone lock status');
    }
  };

  const handleEnableDNSSEC = async (zoneId: string) => {
    try {
      await dnsService.enableDNSSEC(zoneId);
      await loadData();
    } catch (err: any) {
      setError(err.message || 'Failed to enable DNSSEC');
    }
  };

  const handleSyncZone = async (zoneId: string) => {
    try {
      await dnsService.syncZone(zoneId);
      await loadData();
    } catch (err: any) {
      setError(err.message || 'Failed to sync zone');
    }
  };

  const handleCreateRecord = async () => {
    if (!selectedZone || !newRecord.name || !newRecord.content) {
      setFormError('Please fill in all required fields');
      return;
    }
    
    try {
      setIsSaving(true);
      setFormError(null);
      await dnsService.createRecord(selectedZone.id, newRecord);
      await loadRecords(selectedZone.id);
      setShowCreateRecordModal(false);
      setNewRecord({ name: '', type: 'A', content: '', ttl: 3600 });
    } catch (err: any) {
      setFormError(err.message || 'Failed to create record');
    } finally {
      setIsSaving(false);
    }
  };

  const handleUpdateRecord = async () => {
    if (!editingRecord) return;
    
    try {
      setIsSaving(true);
      setFormError(null);
      await dnsService.updateRecord(editingRecord.id, {
        name: editingRecord.name,
        type: editingRecord.type,
        content: editingRecord.content,
        ttl: editingRecord.ttl,
        priority: editingRecord.priority,
      });
      if (selectedZone) {
        await loadRecords(selectedZone.id);
      }
      setShowEditRecordModal(false);
      setEditingRecord(null);
    } catch (err: any) {
      setFormError(err.message || 'Failed to update record');
    } finally {
      setIsSaving(false);
    }
  };

  const handleDeleteRecord = async (id: string) => {
    try {
      setIsSaving(true);
      await dnsService.deleteRecord(id);
      if (selectedZone) {
        await loadRecords(selectedZone.id);
      }
      setShowDeleteConfirm(null);
    } catch (err: any) {
      setError(err.message || 'Failed to delete record');
    } finally {
      setIsSaving(false);
    }
  };

  const filteredZones = zones.filter(zone =>
    zone.name.toLowerCase().includes(searchQuery.toLowerCase())
  );

  const getRecordTypeColor = (type: DNSRecordType) => {
    const colors: Record<DNSRecordType, string> = {
      A: 'bg-blue-500/20 text-blue-400',
      AAAA: 'bg-purple-500/20 text-purple-400',
      CNAME: 'bg-green-500/20 text-green-400',
      MX: 'bg-orange-500/20 text-orange-400',
      TXT: 'bg-yellow-500/20 text-yellow-400',
      SRV: 'bg-pink-500/20 text-pink-400',
      NS: 'bg-cyan-500/20 text-cyan-400',
      CAA: 'bg-red-500/20 text-red-400',
      PTR: 'bg-indigo-500/20 text-indigo-400',
    };
    return colors[type] || 'bg-gray-500/20 text-gray-400';
  };

  if (isLoading) {
    return (
      <div className="space-y-6">
        <div className="h-8 bg-[var(--color-primary)]/50 rounded animate-pulse w-48" />
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          <div className="h-96 bg-[var(--color-primary)]/50 rounded animate-pulse" />
          <div className="lg:col-span-2 h-96 bg-[var(--color-primary)]/50 rounded animate-pulse" />
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-[var(--color-text-primary)]">DNS Management</h1>
          <p className="text-[var(--color-text-secondary)] mt-1">
            Manage DNS zones and records for all domains
          </p>
        </div>
        <Button
          leftIcon={<Plus size={18} />}
          onClick={() => setShowCreateZoneModal(true)}
        >
          Create Zone
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

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Zones List */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Globe size={20} />
              DNS Zones ({zones.length})
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            {/* Search */}
            <div className="relative">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 text-[var(--color-text-muted)]" size={18} />
              <Input
                placeholder="Search zones..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="pl-10"
              />
            </div>

            {/* Zone List */}
            <div className="space-y-2 max-h-[500px] overflow-y-auto">
              {filteredZones.length === 0 ? (
                <p className="text-center text-[var(--color-text-muted)] py-8">
                  No zones found
                </p>
              ) : (
                filteredZones.map(zone => (
                  <div
                    key={zone.id}
                    className={`p-3 rounded-lg cursor-pointer transition-colors ${
                      selectedZone?.id === zone.id
                        ? 'bg-[var(--color-accent)]/20 border border-[var(--color-accent)]'
                        : 'bg-[var(--color-primary-dark)]/50 hover:bg-[var(--color-primary-dark)]'
                    }`}
                    onClick={() => handleZoneClick(zone)}
                  >
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-2">
                        <button
                          onClick={(e) => {
                            e.stopPropagation();
                            toggleZoneExpand(zone.id);
                          }}
                          className="text-[var(--color-text-muted)]"
                        >
                          {expandedZones.has(zone.id) ? (
                            <ChevronDown size={16} />
                          ) : (
                            <ChevronRight size={16} />
                          )}
                        </button>
                        <span className="font-medium text-[var(--color-text-primary)]">
                          {zone.name}
                        </span>
                      </div>
                      <div className="flex items-center gap-1">
                        {zone.dnssec_enabled && (
                          <span title="DNSSEC enabled"><Shield size={14} className="text-[var(--color-success)]" /></span>
                        )}
                        {zone.locked ? (
                          <span title="Locked"><Lock size={14} className="text-[var(--color-warning)]" /></span>
                        ) : (
                          <span title="Unlocked"><Unlock size={14} className="text-[var(--color-text-muted)]" /></span>
                        )}
                      </div>
                    </div>
                    
                    {expandedZones.has(zone.id) && (
                      <div className="mt-2 pt-2 border-t border-[var(--color-border)] space-y-1">
                        <div className="flex gap-1">
                          <Button
                            size="sm"
                            variant="ghost"
                            onClick={(e) => {
                              e.stopPropagation();
                              handleLockZone(zone.id, !zone.locked);
                            }}
                          >
                            {zone.locked ? <Unlock size={14} /> : <Lock size={14} />}
                          </Button>
                          {!zone.dnssec_enabled && (
                            <Button
                              size="sm"
                              variant="ghost"
                              onClick={(e) => {
                                e.stopPropagation();
                                handleEnableDNSSEC(zone.id);
                              }}
                              title="Enable DNSSEC"
                            >
                              <Shield size={14} />
                            </Button>
                          )}
                          <Button
                            size="sm"
                            variant="ghost"
                            onClick={(e) => {
                              e.stopPropagation();
                              handleSyncZone(zone.id);
                            }}
                            title="Sync with provider"
                          >
                            <RefreshCw size={14} />
                          </Button>
                          <Button
                            size="sm"
                            variant="ghost"
                            className="text-[var(--color-error)]"
                            onClick={(e) => {
                              e.stopPropagation();
                              setShowDeleteConfirm({ type: 'zone', id: zone.id });
                            }}
                          >
                            <Trash2 size={14} />
                          </Button>
                        </div>
                      </div>
                    )}
                  </div>
                ))
              )}
            </div>
          </CardContent>
        </Card>

        {/* Records Panel */}
        <div className="lg:col-span-2">
          <Card>
            <CardHeader>
              <div className="flex items-center justify-between">
                <CardTitle>
                  {selectedZone ? (
                    <span className="flex items-center gap-2">
                      DNS Records - {selectedZone.name}
                      {selectedZone.locked && (
                        <Badge variant="warning" size="sm">Locked</Badge>
                      )}
                    </span>
                  ) : (
                    'Select a zone to view records'
                  )}
                </CardTitle>
                {selectedZone && !selectedZone.locked && (
                  <Button
                    size="sm"
                    leftIcon={<Plus size={16} />}
                    onClick={() => setShowCreateRecordModal(true)}
                  >
                    Add Record
                  </Button>
                )}
              </div>
            </CardHeader>
            <CardContent>
              {!selectedZone ? (
                <div className="flex flex-col items-center justify-center py-12 text-[var(--color-text-muted)]">
                  <Globe size={48} className="mb-4 opacity-50" />
                  <p>Select a DNS zone from the list to manage its records</p>
                </div>
              ) : isLoadingRecords ? (
                <div className="space-y-2">
                  {[1, 2, 3].map(i => (
                    <div key={i} className="h-12 bg-[var(--color-primary-dark)]/50 rounded animate-pulse" />
                  ))}
                </div>
              ) : records.length === 0 ? (
                <div className="flex flex-col items-center justify-center py-12 text-[var(--color-text-muted)]">
                  <p>No records found</p>
                  {!selectedZone.locked && (
                    <Button
                      className="mt-4"
                      variant="outline"
                      leftIcon={<Plus size={16} />}
                      onClick={() => setShowCreateRecordModal(true)}
                    >
                      Add First Record
                    </Button>
                  )}
                </div>
              ) : (
                <div className="space-y-2">
                  {/* Records Table Header */}
                  <div className="grid grid-cols-12 gap-4 px-3 py-2 text-sm font-medium text-[var(--color-text-secondary)]">
                    <div className="col-span-1">Type</div>
                    <div className="col-span-3">Name</div>
                    <div className="col-span-4">Content</div>
                    <div className="col-span-2">TTL</div>
                    <div className="col-span-2">Actions</div>
                  </div>
                  
                  {/* Records */}
                  {records.map(record => (
                    <div
                      key={record.id}
                      className="grid grid-cols-12 gap-4 px-3 py-3 rounded-lg bg-[var(--color-primary-dark)]/50 items-center"
                    >
                      <div className="col-span-1">
                        <span className={`px-2 py-1 rounded text-xs font-medium ${getRecordTypeColor(record.type)}`}>
                          {record.type}
                        </span>
                      </div>
                      <div className="col-span-3 text-[var(--color-text-primary)] font-mono text-sm truncate">
                        {record.name}
                      </div>
                      <div className="col-span-4 text-[var(--color-text-secondary)] font-mono text-sm truncate flex items-center gap-1">
                        {record.priority !== undefined && (
                          <span className="text-[var(--color-text-muted)]">{record.priority}</span>
                        )}
                        {record.content}
                        <button
                          onClick={() => navigator.clipboard.writeText(record.content)}
                          className="ml-1 text-[var(--color-text-muted)] hover:text-[var(--color-text-primary)]"
                          title="Copy"
                        >
                          <Copy size={12} />
                        </button>
                      </div>
                      <div className="col-span-2 text-[var(--color-text-muted)] text-sm">
                        {record.ttl}s
                      </div>
                      <div className="col-span-2 flex gap-1">
                        {!selectedZone.locked && (
                          <>
                            <Button
                              size="sm"
                              variant="ghost"
                              onClick={() => {
                                setEditingRecord(record);
                                setShowEditRecordModal(true);
                              }}
                            >
                              <Edit2 size={14} />
                            </Button>
                            <Button
                              size="sm"
                              variant="ghost"
                              className="text-[var(--color-error)]"
                              onClick={() => setShowDeleteConfirm({ type: 'record', id: record.id })}
                            >
                              <Trash2 size={14} />
                            </Button>
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

      {/* Create Zone Modal */}
      <Modal
        isOpen={showCreateZoneModal}
        onClose={() => {
          setShowCreateZoneModal(false);
          setFormError(null);
        }}
        title="Create DNS Zone"
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
              value={newZone.domain_id}
              onChange={(e) => {
                const domain = domains.find(d => d.id === e.target.value);
                setNewZone({
                  domain_id: e.target.value,
                  name: domain?.name || '',
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
              Zone Name
            </label>
            <Input
              value={newZone.name}
              onChange={(e) => setNewZone({ ...newZone, name: e.target.value })}
              placeholder="example.com"
            />
          </div>
          
          <div className="flex gap-3 justify-end">
            <Button variant="ghost" onClick={() => setShowCreateZoneModal(false)}>
              Cancel
            </Button>
            <Button onClick={handleCreateZone} isLoading={isSaving}>
              Create Zone
            </Button>
          </div>
        </div>
      </Modal>

      {/* Create Record Modal */}
      <Modal
        isOpen={showCreateRecordModal}
        onClose={() => {
          setShowCreateRecordModal(false);
          setFormError(null);
        }}
        title="Add DNS Record"
      >
        <div className="space-y-4">
          {formError && (
            <div className="p-3 rounded bg-[var(--color-error)]/20 text-[var(--color-error)] text-sm">
              {formError}
            </div>
          )}
          
          <div>
            <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-2">
              Record Type
            </label>
            <Select
              value={newRecord.type}
              onChange={(e) => setNewRecord({ ...newRecord, type: e.target.value as DNSRecordType })}
              options={RECORD_TYPES.map(t => ({ value: t, label: t }))}
            />
          </div>
          
          <div>
            <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-2">
              Name
            </label>
            <Input
              value={newRecord.name}
              onChange={(e) => setNewRecord({ ...newRecord, name: e.target.value })}
              placeholder="@ or subdomain"
            />
            <p className="text-xs text-[var(--color-text-muted)] mt-1">
              Use @ for root domain or enter subdomain name
            </p>
          </div>
          
          <div>
            <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-2">
              Content
            </label>
            <Input
              value={newRecord.content}
              onChange={(e) => setNewRecord({ ...newRecord, content: e.target.value })}
              placeholder={
                newRecord.type === 'A' ? '192.168.1.1' :
                newRecord.type === 'AAAA' ? '2001:db8::1' :
                newRecord.type === 'CNAME' ? 'target.example.com' :
                newRecord.type === 'MX' ? 'mail.example.com' :
                newRecord.type === 'TXT' ? 'v=spf1 include:_spf.google.com ~all' :
                'Value'
              }
            />
          </div>
          
          {(newRecord.type === 'MX' || newRecord.type === 'SRV') && (
            <div>
              <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-2">
                Priority
              </label>
              <Input
                type="number"
                value={newRecord.priority || ''}
                onChange={(e) => setNewRecord({ ...newRecord, priority: parseInt(e.target.value) || undefined })}
                placeholder="10"
              />
            </div>
          )}
          
          <div>
            <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-2">
              TTL
            </label>
            <Select
              value={newRecord.ttl.toString()}
              onChange={(e) => setNewRecord({ ...newRecord, ttl: parseInt(e.target.value) })}
              options={DEFAULT_TTL_OPTIONS.map(o => ({ value: o.value.toString(), label: o.label }))}
            />
          </div>
          
          <div className="flex gap-3 justify-end">
            <Button variant="ghost" onClick={() => setShowCreateRecordModal(false)}>
              Cancel
            </Button>
            <Button onClick={handleCreateRecord} isLoading={isSaving}>
              Add Record
            </Button>
          </div>
        </div>
      </Modal>

      {/* Edit Record Modal */}
      <Modal
        isOpen={showEditRecordModal}
        onClose={() => {
          setShowEditRecordModal(false);
          setEditingRecord(null);
          setFormError(null);
        }}
        title="Edit DNS Record"
      >
        {editingRecord && (
          <div className="space-y-4">
            {formError && (
              <div className="p-3 rounded bg-[var(--color-error)]/20 text-[var(--color-error)] text-sm">
                {formError}
              </div>
            )}
            
            <div>
              <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-2">
                Record Type
              </label>
              <Select
                value={editingRecord.type}
                onChange={(e) => setEditingRecord({ ...editingRecord, type: e.target.value as DNSRecordType })}
                options={RECORD_TYPES.map(t => ({ value: t, label: t }))}
              />
            </div>
            
            <div>
              <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-2">
                Name
              </label>
              <Input
                value={editingRecord.name}
                onChange={(e) => setEditingRecord({ ...editingRecord, name: e.target.value })}
              />
            </div>
            
            <div>
              <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-2">
                Content
              </label>
              <Input
                value={editingRecord.content}
                onChange={(e) => setEditingRecord({ ...editingRecord, content: e.target.value })}
              />
            </div>
            
            {(editingRecord.type === 'MX' || editingRecord.type === 'SRV') && (
              <div>
                <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-2">
                  Priority
                </label>
                <Input
                  type="number"
                  value={editingRecord.priority || ''}
                  onChange={(e) => setEditingRecord({ ...editingRecord, priority: parseInt(e.target.value) || undefined })}
                />
              </div>
            )}
            
            <div>
              <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-2">
                TTL
              </label>
              <Select
                value={editingRecord.ttl.toString()}
                onChange={(e) => setEditingRecord({ ...editingRecord, ttl: parseInt(e.target.value) })}
                options={DEFAULT_TTL_OPTIONS.map(o => ({ value: o.value.toString(), label: o.label }))}
              />
            </div>
            
            <div className="flex gap-3 justify-end">
              <Button variant="ghost" onClick={() => setShowEditRecordModal(false)}>
                Cancel
              </Button>
              <Button onClick={handleUpdateRecord} isLoading={isSaving}>
                Save Changes
              </Button>
            </div>
          </div>
        )}
      </Modal>

      {/* Delete Confirmation Modal */}
      <Modal
        isOpen={showDeleteConfirm !== null}
        onClose={() => setShowDeleteConfirm(null)}
        title={`Delete ${showDeleteConfirm?.type === 'zone' ? 'Zone' : 'Record'}`}
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
              onClick={() => {
                if (showDeleteConfirm?.type === 'zone') {
                  handleDeleteZone(showDeleteConfirm.id);
                } else if (showDeleteConfirm?.type === 'record') {
                  handleDeleteRecord(showDeleteConfirm.id);
                }
              }}
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
