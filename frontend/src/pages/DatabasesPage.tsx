import { useState, useEffect } from 'react';
import {
  Database,
  Plus,
  Search,
  MoreVertical,
  User,
  Trash2,
  Settings,
  Download,
  CheckCircle,
  HardDrive,
} from 'lucide-react';
import { Button, Input, Card, CardContent, CardHeader } from '../components/ui';
import { CreateDatabaseModal } from '../components/databases/CreateDatabaseModal';
import { DeleteDatabaseModal } from '../components/databases/DeleteDatabaseModal';
import { useToast } from '../context/ToastContext';
import type { Database as DatabaseType } from '../types';
import { databaseService } from '../api/services';

interface ExtendedDatabase extends DatabaseType {
  charset?: string;
  collation?: string;
  tables?: number;
}

const TypeBadge = ({ type }: { type: DatabaseType['type'] }) => {
  const styles = {
    mysql: 'bg-[#00758F]/10 text-[#00758F] border-[#00758F]/20',
    postgresql: 'bg-[#336791]/10 text-[#336791] border-[#336791]/20',
    mariadb: 'bg-[#C0765A]/10 text-[#C0765A] border-[#C0765A]/20',
  };

  const labels = {
    mysql: 'MySQL',
    postgresql: 'PostgreSQL',
    mariadb: 'MariaDB',
  };

  return (
    <span
      className={`inline-flex items-center px-2.5 py-1 rounded-full text-xs font-medium border ${styles[type]}`}
    >
      {labels[type]}
    </span>
  );
};

function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}

export function DatabasesPage() {
  const [searchQuery, setSearchQuery] = useState('');
  const [databases, setDatabases] = useState<ExtendedDatabase[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);
  const [isDeleteModalOpen, setIsDeleteModalOpen] = useState(false);
  const [selectedDatabase, setSelectedDatabase] = useState<ExtendedDatabase | null>(null);
  const { showToast } = useToast();

  const fetchDatabases = async () => {
    try {
      const data = await databaseService.list();
      setDatabases(data);
    } catch (error) {
      console.error('Failed to fetch databases:', error);
      showToast('error', 'Failed to load databases');
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchDatabases();
  }, []);

  const handleCreateDatabase = async (data: { name: string; type: string; charset?: string; collation?: string }) => {
    try {
      await databaseService.create(data);
      await fetchDatabases();
      showToast('success', `Database ${data.name} created successfully!`);
    } catch (error) {
      showToast('error', 'Failed to create database');
      throw error;
    }
  };

  const handleDeleteDatabase = async (databaseId: string) => {
    try {
      await databaseService.delete(databaseId);
      await fetchDatabases();
      showToast('success', 'Database deleted successfully!');
    } catch (error) {
      showToast('error', 'Failed to delete database');
      throw error;
    }
  };

  const openDeleteModal = (database: ExtendedDatabase) => {
    setSelectedDatabase(database);
    setIsDeleteModalOpen(true);
  };

  const filteredDatabases = databases.filter((db) =>
    db.name.toLowerCase().includes(searchQuery.toLowerCase())
  );

  const totalSizeBytes = databases.reduce(
    (acc, db) => acc + (db.size_mb ? db.size_mb * 1024 * 1024 : 0),
    0
  );
  const totalTables = databases.reduce((acc, db) => acc + (db.tables || 0), 0);

  return (
    <div className="space-y-6">
      {/* Page Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-[var(--color-text-primary)]">Databases</h1>
          <p className="text-[var(--color-text-secondary)] mt-1">
            Manage your MySQL, PostgreSQL, and MariaDB databases
          </p>
        </div>
        <Button leftIcon={<Plus size={18} />} onClick={() => setIsCreateModalOpen(true)}>
          Create Database
        </Button>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-4 gap-4">
        <Card>
          <CardContent className="flex items-center gap-4 py-4">
            <div className="p-3 rounded-xl bg-gradient-to-br from-[#7BA4D0]/30 to-[#E7F0FA]/20">
              <Database size={24} className="text-[#E7F0FA]" />
            </div>
            <div>
              <p className="text-2xl font-bold text-[var(--color-text-primary)]">{databases.length}</p>
              <p className="text-sm text-[var(--color-text-secondary)]">Total Databases</p>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="flex items-center gap-4 py-4">
            <div className="p-3 rounded-xl bg-[var(--color-info)]/10">
              <HardDrive size={24} className="text-[var(--color-info)]" />
            </div>
            <div>
              <p className="text-2xl font-bold text-[var(--color-text-primary)]">{formatBytes(totalSizeBytes)}</p>
              <p className="text-sm text-[var(--color-text-secondary)]">Total Size</p>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="flex items-center gap-4 py-4">
            <div className="p-3 rounded-xl bg-[var(--color-success)]/10">
              <CheckCircle size={24} className="text-[var(--color-success)]" />
            </div>
            <div>
              <p className="text-2xl font-bold text-[var(--color-text-primary)]">{totalTables}</p>
              <p className="text-sm text-[var(--color-text-secondary)]">Total Tables</p>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="flex items-center gap-4 py-4">
            <div className="p-3 rounded-xl bg-[var(--color-warning)]/10">
              <User size={24} className="text-[var(--color-warning)]" />
            </div>
            <div>
              <p className="text-2xl font-bold text-[var(--color-text-primary)]">{databases.length * 2}</p>
              <p className="text-sm text-[var(--color-text-secondary)]">Database Users</p>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Search and Filter */}
      <Card>
        <CardHeader>
          <div className="flex flex-col sm:flex-row sm:items-center gap-4">
            <div className="flex-1">
              <Input
                placeholder="Search databases..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                leftIcon={<Search size={18} />}
              />
            </div>
            <div className="flex gap-2">
              <Button variant="outline" size="sm">All</Button>
              <Button variant="ghost" size="sm">MySQL</Button>
              <Button variant="ghost" size="sm">PostgreSQL</Button>
              <Button variant="ghost" size="sm">MariaDB</Button>
            </div>
          </div>
        </CardHeader>
        <CardContent className="p-0">
          {/* Databases Table */}
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="border-b border-[var(--color-border-light)]">
                  <th className="text-left text-xs font-medium text-[var(--color-text-muted)] uppercase tracking-wider px-6 py-3">
                    Database
                  </th>
                  <th className="text-left text-xs font-medium text-[var(--color-text-muted)] uppercase tracking-wider px-6 py-3">
                    Type
                  </th>
                  <th className="text-left text-xs font-medium text-[var(--color-text-muted)] uppercase tracking-wider px-6 py-3">
                    Size
                  </th>
                  <th className="text-left text-xs font-medium text-[var(--color-text-muted)] uppercase tracking-wider px-6 py-3">
                    Tables
                  </th>
                  <th className="text-left text-xs font-medium text-[var(--color-text-muted)] uppercase tracking-wider px-6 py-3">
                    Charset
                  </th>
                  <th className="text-right text-xs font-medium text-[var(--color-text-muted)] uppercase tracking-wider px-6 py-3">
                    Actions
                  </th>
                </tr>
              </thead>
              <tbody className="divide-y divide-[var(--color-border-light)]">
                {filteredDatabases.map((db) => (
                  <tr
                    key={db.id}
                    className="hover:bg-[var(--color-primary-dark)]/30 transition-colors"
                  >
                    <td className="px-6 py-4">
                      <div className="flex items-center gap-3">
                        <div className="p-2 rounded-lg bg-[var(--color-primary-dark)]">
                          <Database size={16} className="text-[#E7F0FA]" />
                        </div>
                        <div>
                          <p className="font-medium text-[var(--color-text-primary)]">{db.name}</p>
                          <p className="text-xs text-[var(--color-text-muted)]">
                            Created {new Date(db.created_at).toLocaleDateString()}
                          </p>
                        </div>
                      </div>
                    </td>
                    <td className="px-6 py-4">
                      <TypeBadge type={db.type} />
                    </td>
                    <td className="px-6 py-4">
                      <span className="text-[var(--color-text-primary)]">
                        {formatBytes((db.size_mb || 0) * 1024 * 1024)}
                      </span>
                    </td>
                    <td className="px-6 py-4">
                      <span className="text-[var(--color-text-primary)]">{db.tables || 0}</span>
                    </td>
                    <td className="px-6 py-4">
                      <code className="text-sm text-[var(--color-text-secondary)] bg-[var(--color-primary-dark)] px-2 py-1 rounded">
                        {db.charset || 'N/A'}
                      </code>
                    </td>
                    <td className="px-6 py-4">
                      <div className="flex items-center justify-end gap-2">
                        <button
                          className="p-2 rounded-lg text-[var(--color-text-secondary)] hover:text-[var(--color-text-primary)] hover:bg-[var(--color-primary-medium)]/20 transition-colors"
                          title="Export Database"
                        >
                          <Download size={16} />
                        </button>
                        <button
                          className="p-2 rounded-lg text-[var(--color-text-secondary)] hover:text-[var(--color-text-primary)] hover:bg-[var(--color-primary-medium)]/20 transition-colors"
                          title="Manage Users"
                        >
                          <User size={16} />
                        </button>
                        <button
                          className="p-2 rounded-lg text-[var(--color-text-secondary)] hover:text-[var(--color-text-primary)] hover:bg-[var(--color-primary-medium)]/20 transition-colors"
                          title="Settings"
                        >
                          <Settings size={16} />
                        </button>
                        <button
                          onClick={() => openDeleteModal(db)}
                          className="p-2 rounded-lg text-[var(--color-text-secondary)] hover:text-[var(--color-error)] hover:bg-[var(--color-error)]/10 transition-colors"
                          title="Delete"
                        >
                          <Trash2 size={16} />
                        </button>
                        <button className="p-2 rounded-lg text-[var(--color-text-secondary)] hover:text-[var(--color-text-primary)] hover:bg-[var(--color-primary-medium)]/20 transition-colors">
                          <MoreVertical size={16} />
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          {filteredDatabases.length === 0 && (
            <div className="text-center py-12">
              <Database size={48} className="mx-auto text-[var(--color-text-muted)] mb-4" />
              <p className="text-[var(--color-text-secondary)]">No databases found</p>
              <p className="text-sm text-[var(--color-text-muted)] mt-1">
                Try adjusting your search or create a new database
              </p>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Modals */}
      <CreateDatabaseModal
        isOpen={isCreateModalOpen}
        onClose={() => setIsCreateModalOpen(false)}
        onSubmit={handleCreateDatabase}
      />
      <DeleteDatabaseModal
        isOpen={isDeleteModalOpen}
        database={selectedDatabase}
        onClose={() => {
          setIsDeleteModalOpen(false);
          setSelectedDatabase(null);
        }}
        onConfirm={handleDeleteDatabase}
      />
    </div>
  );
}
