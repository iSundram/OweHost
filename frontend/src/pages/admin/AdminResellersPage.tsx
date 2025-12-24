import { useState, useEffect } from 'react';
import type { ReactNode } from 'react';
import {
  Store,
  Plus,
  Search,
  Edit,
  Trash2,
  Users,
  Zap,
  HardDrive,
  Globe,
  Database,
  AlertTriangle,
  CheckCircle,
} from 'lucide-react';
import { Button, Input, Card, CardContent, CardHeader, Modal } from '../../components/ui';
import { resellerService, userService } from '../../api/services';
import type { Reseller, ResellerCreateRequest } from '../../api/services';
import type { User } from '../../types';

export function AdminResellersPage() {
  const [searchQuery, setSearchQuery] = useState('');
  const [resellers, setResellers] = useState<Reseller[]>([]);
  const [users, setUsers] = useState<User[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [selectedReseller, setSelectedReseller] = useState<Reseller | null>(null);
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);
  const [isEditModalOpen, setIsEditModalOpen] = useState(false);
  const [isDeleteModalOpen, setIsDeleteModalOpen] = useState(false);

  const loadData = () => {
    setIsLoading(true);
    Promise.all([resellerService.list(), userService.list()])
      .then(([resellersData, usersData]) => {
        setResellers(resellersData);
        setUsers(usersData.filter((u) => u.role === 'user' || u.role === 'reseller'));
      })
      .catch(console.error)
      .finally(() => setIsLoading(false));
  };

  useEffect(() => {
    loadData();
  }, []);

  const filteredResellers = resellers.filter((reseller) => {
    const user = users.find((u) => u.id === reseller.user_id);
    return (
      reseller.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      user?.username.toLowerCase().includes(searchQuery.toLowerCase()) ||
      user?.email.toLowerCase().includes(searchQuery.toLowerCase())
    );
  });

  const handleCreate = async (data: ResellerCreateRequest) => {
    try {
      await resellerService.create(data);
      setIsCreateModalOpen(false);
      loadData();
    } catch (error) {
      console.error('Failed to create reseller:', error);
      alert('Failed to create reseller');
    }
  };

  const handleUpdate = async (id: string, data: Partial<ResellerCreateRequest>) => {
    try {
      await resellerService.update(id, data);
      setIsEditModalOpen(false);
      setSelectedReseller(null);
      loadData();
    } catch (error) {
      console.error('Failed to update reseller:', error);
      alert('Failed to update reseller');
    }
  };

  const handleDelete = async (id: string) => {
    try {
      await resellerService.delete(id);
      setIsDeleteModalOpen(false);
      setSelectedReseller(null);
      loadData();
    } catch (error) {
      console.error('Failed to delete reseller:', error);
      alert('Failed to delete reseller');
    }
  };

  return (
    <div className="space-y-6">
      {/* Page Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-[var(--color-text-primary)]">Reseller Management</h1>
          <p className="text-[var(--color-text-secondary)] mt-1">
            Manage reseller accounts and resource allocation
          </p>
        </div>
        <Button leftIcon={<Plus size={18} />} onClick={() => setIsCreateModalOpen(true)}>
          Create Reseller
        </Button>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-4 gap-4">
        <Card>
          <CardContent className="flex items-center gap-4 py-4">
            <div className="p-3 rounded-xl bg-gradient-to-br from-[#7BA4D0]/30 to-[#E7F0FA]/20">
              <Store size={24} className="text-[#E7F0FA]" />
            </div>
            <div>
              <p className="text-2xl font-bold text-[var(--color-text-primary)]">{resellers.length}</p>
              <p className="text-sm text-[var(--color-text-secondary)]">Total Resellers</p>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="flex items-center gap-4 py-4">
            <div className="p-3 rounded-xl bg-[var(--color-success)]/10">
              <Users size={24} className="text-[var(--color-success)]" />
            </div>
            <div>
              <p className="text-2xl font-bold text-[var(--color-text-primary)]">
                {resellers.reduce((sum, r) => sum + r.resource_pool.max_users, 0)}
              </p>
              <p className="text-sm text-[var(--color-text-secondary)]">Max Customers</p>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="flex items-center gap-4 py-4">
            <div className="p-3 rounded-xl bg-[#7BA4D0]/20">
              <HardDrive size={24} className="text-[#E7F0FA]" />
            </div>
            <div>
              <p className="text-2xl font-bold text-[var(--color-text-primary)]">
                {Math.round(resellers.reduce((sum, r) => sum + r.resource_pool.max_disk_mb, 0) / 1024)}
              </p>
              <p className="text-sm text-[var(--color-text-secondary)]">Total Disk (GB)</p>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="flex items-center gap-4 py-4">
            <div className="p-3 rounded-xl bg-[var(--color-warning)]/10">
              <Zap size={24} className="text-[var(--color-warning)]" />
            </div>
            <div>
              <p className="text-2xl font-bold text-[var(--color-text-primary)]">
                {resellers.reduce((sum, r) => sum + r.resource_pool.max_cpu_quota, 0)}
              </p>
              <p className="text-sm text-[var(--color-text-secondary)]">Total CPU Cores</p>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Resellers Table */}
      <Card>
        <CardHeader>
          <div className="flex flex-col sm:flex-row sm:items-center gap-4">
            <div className="flex-1">
              <Input
                placeholder="Search resellers by name or user..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                leftIcon={<Search size={18} />}
              />
            </div>
          </div>
        </CardHeader>
        <CardContent className="p-0">
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="border-b border-[var(--color-border-light)]">
                  <th className="text-left text-xs font-medium text-[var(--color-text-muted)] uppercase tracking-wider px-6 py-3">
                    Reseller
                  </th>
                  <th className="text-left text-xs font-medium text-[var(--color-text-muted)] uppercase tracking-wider px-6 py-3">
                    User
                  </th>
                  <th className="text-left text-xs font-medium text-[var(--color-text-muted)] uppercase tracking-wider px-6 py-3">
                    Resource Pool
                  </th>
                  <th className="text-left text-xs font-medium text-[var(--color-text-muted)] uppercase tracking-wider px-6 py-3">
                    Created
                  </th>
                  <th className="text-right text-xs font-medium text-[var(--color-text-muted)] uppercase tracking-wider px-6 py-3">
                    Actions
                  </th>
                </tr>
              </thead>
              <tbody className="divide-y divide-[var(--color-border-light)]">
                {isLoading ? (
                  <tr>
                    <td colSpan={5} className="px-6 py-12 text-center">
                      <div className="flex flex-col items-center gap-2">
                        <div className="w-8 h-8 border-2 border-[#E7F0FA] border-t-transparent rounded-full animate-spin" />
                        <p className="text-[var(--color-text-secondary)]">Loading resellers...</p>
                      </div>
                    </td>
                  </tr>
                ) : filteredResellers.length === 0 ? (
                  <tr>
                    <td colSpan={5} className="px-6 py-12 text-center">
                      <Store size={48} className="mx-auto text-[var(--color-text-muted)] mb-4" />
                      <p className="text-[var(--color-text-secondary)]">No resellers found</p>
                      <p className="text-sm text-[var(--color-text-muted)] mt-1">
                        Create a new reseller to get started
                      </p>
                    </td>
                  </tr>
                ) : (
                  filteredResellers.map((reseller) => {
                    const user = users.find((u) => u.id === reseller.user_id);
                    return (
                      <tr
                        key={reseller.id}
                        className="hover:bg-[var(--color-primary-dark)]/30 transition-colors"
                      >
                        <td className="px-6 py-4">
                          <div className="flex items-center gap-3">
                            <div className="w-10 h-10 rounded-full bg-gradient-to-br from-[#7BA4D0] to-[#E7F0FA] flex items-center justify-center">
                              <Store size={20} className="text-white" />
                            </div>
                            <div>
                              <p className="font-medium text-[var(--color-text-primary)]">
                                {reseller.name}
                              </p>
                              <p className="text-xs text-[var(--color-text-muted)]">ID: {reseller.id}</p>
                            </div>
                          </div>
                        </td>
                        <td className="px-6 py-4">
                          {user ? (
                            <div>
                              <p className="text-[var(--color-text-primary)]">{user.username}</p>
                              <p className="text-xs text-[var(--color-text-muted)]">{user.email}</p>
                            </div>
                          ) : (
                            <span className="text-[var(--color-text-muted)]">Unknown</span>
                          )}
                        </td>
                        <td className="px-6 py-4">
                          <div className="space-y-1 text-sm">
                            <div className="flex items-center gap-2">
                              <Users size={14} className="text-[var(--color-text-muted)]" />
                              <span className="text-[var(--color-text-secondary)]">
                                {reseller.resource_pool.max_users} customers
                              </span>
                            </div>
                            <div className="flex items-center gap-2">
                              <HardDrive size={14} className="text-[var(--color-text-muted)]" />
                              <span className="text-[var(--color-text-secondary)]">
                                {Math.round(reseller.resource_pool.max_disk_mb / 1024)} GB disk
                              </span>
                            </div>
                            <div className="flex items-center gap-2">
                              <Zap size={14} className="text-[var(--color-text-muted)]" />
                              <span className="text-[var(--color-text-secondary)]">
                                {reseller.resource_pool.max_cpu_quota} CPU cores
                              </span>
                            </div>
                          </div>
                        </td>
                        <td className="px-6 py-4">
                          <span className="text-[var(--color-text-secondary)]">
                            {new Date(reseller.created_at).toLocaleDateString()}
                          </span>
                        </td>
                        <td className="px-6 py-4">
                          <div className="flex items-center justify-end gap-2">
                            <button
                              onClick={() => {
                                setSelectedReseller(reseller);
                                setIsEditModalOpen(true);
                              }}
                              className="p-2 rounded-lg text-[var(--color-text-secondary)] hover:text-[var(--color-text-primary)] hover:bg-[var(--color-primary-medium)]/20 transition-colors"
                              title="Edit"
                            >
                              <Edit size={16} />
                            </button>
                            <button
                              onClick={() => {
                                setSelectedReseller(reseller);
                                setIsDeleteModalOpen(true);
                              }}
                              className="p-2 rounded-lg text-[var(--color-text-secondary)] hover:text-[var(--color-error)] hover:bg-[var(--color-error)]/10 transition-colors"
                              title="Delete"
                            >
                              <Trash2 size={16} />
                            </button>
                          </div>
                        </td>
                      </tr>
                    );
                  })
                )}
              </tbody>
            </table>
          </div>
        </CardContent>
      </Card>

      {/* Create Reseller Modal */}
      <CreateResellerModal
        isOpen={isCreateModalOpen}
        onClose={() => setIsCreateModalOpen(false)}
        onSubmit={handleCreate}
        users={users}
      />

      {/* Edit Reseller Modal */}
      {selectedReseller && (
        <>
          <EditResellerModal
            isOpen={isEditModalOpen}
            onClose={() => {
              setIsEditModalOpen(false);
              setSelectedReseller(null);
            }}
            reseller={selectedReseller}
            onSubmit={handleUpdate}
          />
          <DeleteResellerModal
            isOpen={isDeleteModalOpen}
            onClose={() => {
              setIsDeleteModalOpen(false);
              setSelectedReseller(null);
            }}
            reseller={selectedReseller}
            onConfirm={handleDelete}
          />
        </>
      )}
    </div>
  );
}

// Create Reseller Modal
function CreateResellerModal({
  isOpen,
  onClose,
  onSubmit,
  users,
}: {
  isOpen: boolean;
  onClose: () => void;
  onSubmit: (data: ResellerCreateRequest) => void;
  users: User[];
}) {
  const [formData, setFormData] = useState<ResellerCreateRequest>({
    user_id: '',
    name: '',
    resource_pool: {
      max_users: 10,
      max_domains: 50,
      max_disk_mb: 10240,
      max_bandwidth_mb: 102400,
      max_databases: 20,
      max_cpu_quota: 2,
      max_memory_mb: 2048,
    },
  });
  const [isSubmitting, setIsSubmitting] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsSubmitting(true);
    try {
      await onSubmit(formData);
      setFormData({
        user_id: '',
        name: '',
        resource_pool: {
          max_users: 10,
          max_domains: 50,
          max_disk_mb: 10240,
          max_bandwidth_mb: 102400,
          max_databases: 20,
          max_cpu_quota: 2,
          max_memory_mb: 2048,
        },
      });
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <Modal isOpen={isOpen} onClose={onClose} title="Create Reseller" size="lg">
      <form onSubmit={handleSubmit} className="space-y-4">
        <div>
          <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-2">
            Select User
          </label>
          <select
            value={formData.user_id}
            onChange={(e) => {
              const user = users.find((u) => u.id === e.target.value);
              setFormData({
                ...formData,
                user_id: e.target.value,
                name: user?.username || '',
              });
            }}
            className="w-full px-4 py-2 rounded-lg bg-[var(--color-primary-dark)] border border-[var(--color-border)] text-[var(--color-text-primary)] focus:outline-none focus:ring-2 focus:ring-[#E7F0FA]"
            required
          >
            <option value="">Select a user...</option>
            {users
              .filter((u) => u.role === 'user')
              .map((user) => (
                <option key={user.id} value={user.id}>
                  {user.username} ({user.email})
                </option>
              ))}
          </select>
        </div>
        <Input
          label="Reseller Name"
          value={formData.name}
          onChange={(e) => setFormData({ ...formData, name: e.target.value })}
          required
          placeholder="Enter reseller name"
        />
        <div className="grid grid-cols-2 gap-4">
          <Input
            label="Max Customers"
            type="number"
            value={formData.resource_pool.max_users}
            onChange={(e) =>
              setFormData({
                ...formData,
                resource_pool: {
                  ...formData.resource_pool,
                  max_users: parseInt(e.target.value) || 0,
                },
              })
            }
            required
            min={1}
          />
          <Input
            label="Max Domains"
            type="number"
            value={formData.resource_pool.max_domains}
            onChange={(e) =>
              setFormData({
                ...formData,
                resource_pool: {
                  ...formData.resource_pool,
                  max_domains: parseInt(e.target.value) || 0,
                },
              })
            }
            required
            min={1}
          />
          <Input
            label="Max Disk (MB)"
            type="number"
            value={formData.resource_pool.max_disk_mb}
            onChange={(e) =>
              setFormData({
                ...formData,
                resource_pool: {
                  ...formData.resource_pool,
                  max_disk_mb: parseInt(e.target.value) || 0,
                },
              })
            }
            required
            min={1024}
          />
          <Input
            label="Max Bandwidth (MB)"
            type="number"
            value={formData.resource_pool.max_bandwidth_mb}
            onChange={(e) =>
              setFormData({
                ...formData,
                resource_pool: {
                  ...formData.resource_pool,
                  max_bandwidth_mb: parseInt(e.target.value) || 0,
                },
              })
            }
            required
            min={1024}
          />
          <Input
            label="Max Databases"
            type="number"
            value={formData.resource_pool.max_databases}
            onChange={(e) =>
              setFormData({
                ...formData,
                resource_pool: {
                  ...formData.resource_pool,
                  max_databases: parseInt(e.target.value) || 0,
                },
              })
            }
            required
            min={1}
          />
          <Input
            label="Max CPU Cores"
            type="number"
            value={formData.resource_pool.max_cpu_quota}
            onChange={(e) =>
              setFormData({
                ...formData,
                resource_pool: {
                  ...formData.resource_pool,
                  max_cpu_quota: parseInt(e.target.value) || 0,
                },
              })
            }
            required
            min={1}
          />
          <Input
            label="Max Memory (MB)"
            type="number"
            value={formData.resource_pool.max_memory_mb}
            onChange={(e) =>
              setFormData({
                ...formData,
                resource_pool: {
                  ...formData.resource_pool,
                  max_memory_mb: parseInt(e.target.value) || 0,
                },
              })
            }
            required
            min={512}
          />
        </div>
        <div className="flex gap-3 pt-4">
          <Button type="button" variant="ghost" onClick={onClose} className="flex-1">
            Cancel
          </Button>
          <Button type="submit" variant="primary" className="flex-1" isLoading={isSubmitting}>
            Create Reseller
          </Button>
        </div>
      </form>
    </Modal>
  );
}

// Edit Reseller Modal
function EditResellerModal({
  isOpen,
  onClose,
  reseller,
  onSubmit,
}: {
  isOpen: boolean;
  onClose: () => void;
  reseller: Reseller;
  onSubmit: (id: string, data: Partial<ResellerCreateRequest>) => void;
}) {
  const [formData, setFormData] = useState({
    name: reseller.name,
    resource_pool: reseller.resource_pool,
  });
  const [isSubmitting, setIsSubmitting] = useState(false);

  useEffect(() => {
    setFormData({
      name: reseller.name,
      resource_pool: reseller.resource_pool,
    });
  }, [reseller]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsSubmitting(true);
    try {
      await onSubmit(reseller.id, formData);
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <Modal isOpen={isOpen} onClose={onClose} title="Edit Reseller" size="lg">
      <form onSubmit={handleSubmit} className="space-y-4">
        <Input
          label="Reseller Name"
          value={formData.name}
          onChange={(e) => setFormData({ ...formData, name: e.target.value })}
          required
        />
        <div className="grid grid-cols-2 gap-4">
          <Input
            label="Max Customers"
            type="number"
            value={formData.resource_pool.max_users}
            onChange={(e) =>
              setFormData({
                ...formData,
                resource_pool: {
                  ...formData.resource_pool,
                  max_users: parseInt(e.target.value) || 0,
                },
              })
            }
            required
            min={1}
          />
          <Input
            label="Max Domains"
            type="number"
            value={formData.resource_pool.max_domains}
            onChange={(e) =>
              setFormData({
                ...formData,
                resource_pool: {
                  ...formData.resource_pool,
                  max_domains: parseInt(e.target.value) || 0,
                },
              })
            }
            required
            min={1}
          />
          <Input
            label="Max Disk (MB)"
            type="number"
            value={formData.resource_pool.max_disk_mb}
            onChange={(e) =>
              setFormData({
                ...formData,
                resource_pool: {
                  ...formData.resource_pool,
                  max_disk_mb: parseInt(e.target.value) || 0,
                },
              })
            }
            required
            min={1024}
          />
          <Input
            label="Max Bandwidth (MB)"
            type="number"
            value={formData.resource_pool.max_bandwidth_mb}
            onChange={(e) =>
              setFormData({
                ...formData,
                resource_pool: {
                  ...formData.resource_pool,
                  max_bandwidth_mb: parseInt(e.target.value) || 0,
                },
              })
            }
            required
            min={1024}
          />
          <Input
            label="Max Databases"
            type="number"
            value={formData.resource_pool.max_databases}
            onChange={(e) =>
              setFormData({
                ...formData,
                resource_pool: {
                  ...formData.resource_pool,
                  max_databases: parseInt(e.target.value) || 0,
                },
              })
            }
            required
            min={1}
          />
          <Input
            label="Max CPU Cores"
            type="number"
            value={formData.resource_pool.max_cpu_quota}
            onChange={(e) =>
              setFormData({
                ...formData,
                resource_pool: {
                  ...formData.resource_pool,
                  max_cpu_quota: parseInt(e.target.value) || 0,
                },
              })
            }
            required
            min={1}
          />
          <Input
            label="Max Memory (MB)"
            type="number"
            value={formData.resource_pool.max_memory_mb}
            onChange={(e) =>
              setFormData({
                ...formData,
                resource_pool: {
                  ...formData.resource_pool,
                  max_memory_mb: parseInt(e.target.value) || 0,
                },
              })
            }
            required
            min={512}
          />
        </div>
        <div className="flex gap-3 pt-4">
          <Button type="button" variant="ghost" onClick={onClose} className="flex-1">
            Cancel
          </Button>
          <Button type="submit" variant="primary" className="flex-1" isLoading={isSubmitting}>
            Save Changes
          </Button>
        </div>
      </form>
    </Modal>
  );
}

// Delete Reseller Modal
function DeleteResellerModal({
  isOpen,
  onClose,
  reseller,
  onConfirm,
}: {
  isOpen: boolean;
  onClose: () => void;
  reseller: Reseller;
  onConfirm: (id: string) => void;
}) {
  const [isDeleting, setIsDeleting] = useState(false);

  const handleConfirm = async () => {
    setIsDeleting(true);
    try {
      await onConfirm(reseller.id);
    } finally {
      setIsDeleting(false);
    }
  };

  return (
    <Modal isOpen={isOpen} onClose={onClose} title="Delete Reseller" size="md">
      <div className="space-y-4">
        <div className="p-4 rounded-lg bg-[var(--color-error)]/10 border border-[var(--color-error)]/20">
          <p className="text-[var(--color-text-primary)]">
            Are you sure you want to delete reseller <strong>{reseller.name}</strong>?
          </p>
          <p className="text-sm text-[var(--color-text-secondary)] mt-2">
            This will convert the user back to a regular user account. All reseller data will be removed.
          </p>
        </div>
        <div className="flex gap-3 pt-4">
          <Button variant="ghost" onClick={onClose} className="flex-1">
            Cancel
          </Button>
          <Button
            variant="primary"
            onClick={handleConfirm}
            className="flex-1 bg-[var(--color-error)] hover:bg-[var(--color-error)]/90"
            isLoading={isDeleting}
          >
            Delete Reseller
          </Button>
        </div>
      </div>
    </Modal>
  );
}
