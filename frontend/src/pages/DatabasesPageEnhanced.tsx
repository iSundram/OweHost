import { useState } from 'react';
import { 
  Database, 
  Trash2, 
  Plus,
  Search,
  RefreshCw 
} from 'lucide-react';
import { 
  Card, 
  CardContent, 
  CardHeader, 
  CardTitle,
  Button,
  Input,
  Table,
  Badge,
  EmptyState,
  useToast,
  ConfirmDialog
} from '../components/ui';
import type { Column } from '../components/ui/Table';
import { CreateDatabaseModal } from '../components/databases/CreateDatabaseModal';

interface DatabaseItem {
  id: string;
  name: string;
  type: 'mysql' | 'postgresql';
  size: string;
  status: 'active' | 'inactive';
  createdAt: string;
}

export function DatabasesPage() {
  const { showToast } = useToast();
  const [databases, setDatabases] = useState<DatabaseItem[]>([
    {
      id: '1',
      name: 'wordpress_db',
      type: 'mysql',
      size: '125 MB',
      status: 'active',
      createdAt: '2024-01-15',
    },
    {
      id: '2',
      name: 'app_database',
      type: 'postgresql',
      size: '89 MB',
      status: 'active',
      createdAt: '2024-01-20',
    },
  ]);
  const [searchQuery, setSearchQuery] = useState('');
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);
  const [deleteDialog, setDeleteDialog] = useState<{ isOpen: boolean; database?: DatabaseItem }>({
    isOpen: false,
  });
  const [isDeleting, setIsDeleting] = useState(false);

  const filteredDatabases = databases.filter(db =>
    db.name.toLowerCase().includes(searchQuery.toLowerCase())
  );

  const handleDelete = async () => {
    if (!deleteDialog.database) return;
    
    setIsDeleting(true);
    // Simulate API call
    setTimeout(() => {
      setDatabases(prev => prev.filter(db => db.id !== deleteDialog.database?.id));
      showToast({
        type: 'success',
        title: 'Database deleted',
        message: `${deleteDialog.database.name} has been deleted successfully`,
      });
      setDeleteDialog({ isOpen: false });
      setIsDeleting(false);
    }, 1000);
  };

  const handleRefresh = () => {
    showToast({
      type: 'info',
      title: 'Refreshing',
      message: 'Database list updated',
    });
  };

  const columns: Column<DatabaseItem>[] = [
    {
      key: 'name',
      header: 'Database Name',
      sortable: true,
      render: (item) => (
        <div className="flex items-center gap-2">
          <Database size={16} className="text-[var(--color-text-muted)]" />
          <span className="font-medium text-[var(--color-text-primary)]">{item.name}</span>
        </div>
      ),
    },
    {
      key: 'type',
      header: 'Type',
      sortable: true,
      render: (item) => (
        <Badge variant={item.type === 'mysql' ? 'primary' : 'info'}>
          {item.type.toUpperCase()}
        </Badge>
      ),
    },
    {
      key: 'size',
      header: 'Size',
      sortable: true,
    },
    {
      key: 'status',
      header: 'Status',
      sortable: true,
      render: (item) => (
        <Badge variant={item.status === 'active' ? 'success' : 'default'} dot>
          {item.status}
        </Badge>
      ),
    },
    {
      key: 'createdAt',
      header: 'Created',
      sortable: true,
    },
    {
      key: 'actions',
      header: 'Actions',
      render: (item) => (
        <Button
          variant="ghost"
          size="sm"
          onClick={(e) => {
            e.stopPropagation();
            setDeleteDialog({ isOpen: true, database: item });
          }}
          leftIcon={<Trash2 size={16} />}
        >
          Delete
        </Button>
      ),
    },
  ];

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-[var(--color-text-primary)]">Databases</h1>
          <p className="text-[var(--color-text-secondary)] mt-1">
            Manage your MySQL and PostgreSQL databases
          </p>
        </div>
        <div className="flex items-center gap-3">
          <Button
            variant="ghost"
            size="md"
            leftIcon={<RefreshCw size={18} />}
            onClick={handleRefresh}
          >
            Refresh
          </Button>
          <Button
            variant="primary"
            size="md"
            leftIcon={<Plus size={18} />}
            onClick={() => setIsCreateModalOpen(true)}
          >
            Create Database
          </Button>
        </div>
      </div>

      {/* Search and Filters */}
      <Card>
        <CardContent>
          <div className="flex items-center gap-4">
            <div className="flex-1">
              <Input
                placeholder="Search databases..."
                leftIcon={<Search size={18} />}
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
              />
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Databases Table */}
      <Card>
        <CardHeader>
          <CardTitle>All Databases ({filteredDatabases.length})</CardTitle>
        </CardHeader>
        <CardContent>
          {filteredDatabases.length === 0 ? (
            <EmptyState
              icon={<Database size={32} className="text-[var(--color-text-muted)]" />}
              title="No databases found"
              description={
                searchQuery
                  ? 'No databases match your search criteria'
                  : 'Get started by creating your first database'
              }
              action={
                !searchQuery
                  ? {
                      label: 'Create Database',
                      onClick: () => setIsCreateModalOpen(true),
                    }
                  : undefined
              }
            />
          ) : (
            <Table
              data={filteredDatabases}
              columns={columns}
              keyExtractor={(item) => item.id}
            />
          )}
        </CardContent>
      </Card>

      {/* Create Modal */}
      <CreateDatabaseModal
        isOpen={isCreateModalOpen}
        onClose={() => setIsCreateModalOpen(false)}
      />

      {/* Delete Confirmation */}
      <ConfirmDialog
        isOpen={deleteDialog.isOpen}
        onClose={() => setDeleteDialog({ isOpen: false })}
        onConfirm={handleDelete}
        title="Delete Database"
        message={`Are you sure you want to delete "${deleteDialog.database?.name}"? This action cannot be undone and all data will be permanently lost.`}
        confirmText="Delete Database"
        variant="danger"
        loading={isDeleting}
      />
    </div>
  );
}
