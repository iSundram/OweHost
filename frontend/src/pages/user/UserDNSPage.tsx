import { useState, useEffect } from 'react';
import {
  Globe,
  Plus,
  Search,
  Trash2,
  Edit2,
  AlertCircle,
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

const RECORD_TYPES: DNSRecordType[] = ['A', 'AAAA', 'CNAME', 'MX', 'TXT', 'SRV'];

const DEFAULT_TTL_OPTIONS = [
  { value: 300, label: '5 minutes' },
  { value: 600, label: '10 minutes' },
  { value: 1800, label: '30 minutes' },
  { value: 3600, label: '1 hour' },
  { value: 14400, label: '4 hours' },
  { value: 86400, label: '1 day' },
];

export function UserDNSPage() {
  const [zones, setZones] = useState<DNSZone[]>([]);
  const [domains, setDomains] = useState<Domain[]>([]);
  const [selectedZone, setSelectedZone] = useState<DNSZone | null>(null);
  const [records, setRecords] = useState<DNSRecord[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isLoadingRecords, setIsLoadingRecords] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState('');
  
  // Modal states
  const [showCreateRecordModal, setShowCreateRecordModal] = useState(false);
  const [showEditRecordModal, setShowEditRecordModal] = useState(false);
  const [showDeleteConfirm, setShowDeleteConfirm] = useState<string | null>(null);
  
  // Form states
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
      
      // Auto-select first zone
      if (zonesData.length > 0 && !selectedZone) {
        setSelectedZone(zonesData[0]);
        await loadRecords(zonesData[0].id);
      }
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

  const handleZoneChange = async (zoneId: string) => {
    const zone = zones.find(z => z.id === zoneId);
    if (zone) {
      setSelectedZone(zone);
      await loadRecords(zone.id);
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

  const getRecordTypeColor = (type: DNSRecordType) => {
    const colors: Record<string, string> = {
      A: 'bg-blue-500/20 text-blue-400',
      AAAA: 'bg-purple-500/20 text-purple-400',
      CNAME: 'bg-green-500/20 text-green-400',
      MX: 'bg-orange-500/20 text-orange-400',
      TXT: 'bg-yellow-500/20 text-yellow-400',
      SRV: 'bg-pink-500/20 text-pink-400',
    };
    return colors[type] || 'bg-gray-500/20 text-gray-400';
  };

  const filteredRecords = records.filter(record =>
    record.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
    record.content.toLowerCase().includes(searchQuery.toLowerCase())
  );

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
          <h1 className="text-2xl font-bold text-[var(--color-text-primary)]">DNS Records</h1>
          <p className="text-[var(--color-text-secondary)] mt-1">
            Manage DNS records for your domains
          </p>
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

      {zones.length === 0 ? (
        <Card>
          <CardContent className="py-12 text-center">
            <Globe size={48} className="mx-auto mb-4 text-[var(--color-text-muted)] opacity-50" />
            <h3 className="text-lg font-medium text-[var(--color-text-primary)] mb-2">
              No DNS Zones
            </h3>
            <p className="text-[var(--color-text-muted)]">
              DNS zones are created automatically when you add domains.
            </p>
          </CardContent>
        </Card>
      ) : (
        <Card>
          <CardHeader>
            <div className="flex items-center justify-between flex-wrap gap-4">
              <div className="flex items-center gap-4">
                <CardTitle>DNS Records</CardTitle>
                <Select
                  value={selectedZone?.id || ''}
                  onChange={(e) => handleZoneChange(e.target.value)}
                  options={zones.map(z => ({ value: z.id, label: z.name }))}
                  className="w-48"
                />
              </div>
              <div className="flex gap-4">
                <div className="relative">
                  <Search className="absolute left-3 top-1/2 -translate-y-1/2 text-[var(--color-text-muted)]" size={18} />
                  <Input
                    placeholder="Search records..."
                    value={searchQuery}
                    onChange={(e) => setSearchQuery(e.target.value)}
                    className="pl-10 w-64"
                  />
                </div>
                {selectedZone && !selectedZone.locked && (
                  <Button
                    leftIcon={<Plus size={18} />}
                    onClick={() => setShowCreateRecordModal(true)}
                  >
                    Add Record
                  </Button>
                )}
              </div>
            </div>
          </CardHeader>
          <CardContent>
            {selectedZone?.locked && (
              <div className="mb-4 p-3 rounded bg-[var(--color-warning)]/20 text-[var(--color-warning)] text-sm">
                This zone is locked. Contact your administrator to make changes.
              </div>
            )}

            {isLoadingRecords ? (
              <div className="space-y-2">
                {[1, 2, 3].map(i => (
                  <div key={i} className="h-12 bg-[var(--color-primary-dark)]/50 rounded animate-pulse" />
                ))}
              </div>
            ) : filteredRecords.length === 0 ? (
              <div className="flex flex-col items-center justify-center py-12 text-[var(--color-text-muted)]">
                <Globe size={48} className="mb-4 opacity-50" />
                <p>No records found</p>
                {!selectedZone?.locked && (
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
                {filteredRecords.map(record => (
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
                      {!selectedZone?.locked && (
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
                            onClick={() => setShowDeleteConfirm(record.id)}
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
      )}

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
        title="Delete Record"
      >
        <div className="space-y-4">
          <p className="text-[var(--color-text-secondary)]">
            Are you sure you want to delete this DNS record? This may affect your domain's functionality.
          </p>
          <div className="flex gap-3 justify-end">
            <Button variant="ghost" onClick={() => setShowDeleteConfirm(null)}>
              Cancel
            </Button>
            <Button
              variant="danger"
              onClick={() => showDeleteConfirm && handleDeleteRecord(showDeleteConfirm)}
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
