import { useState } from 'react';
import { 
  Globe, 
  Trash2, 
  Plus,
  Search,
  RefreshCw,
  ExternalLink,
  Shield
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
  ConfirmDialog,
  StatusBadge
} from '../components/ui';
import type { Column } from '../components/ui/Table';
import { CreateDomainModal } from '../components/domains/CreateDomainModal';

interface DomainItem {
  id: string;
  name: string;
  status: 'active' | 'pending' | 'suspended';
  ssl: boolean;
  ipAddress: string;
  createdAt: string;
}

export function DomainsPageEnhanced() {
  const { showToast } = useToast();
  const [domains, setDomains] = useState<DomainItem[]>([
    {
      id: '1',
      name: 'example.com',
      status: 'active',
      ssl: true,
      ipAddress: '192.168.1.1',
      createdAt: '2024-01-15',
    },
    {
      id: '2',
      name: 'mysite.net',
      status: 'active',
      ssl: false,
      ipAddress: '192.168.1.2',
      createdAt: '2024-01-20',
    },
  ]);
  const [searchQuery, setSearchQuery] = useState('');
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);
  const [deleteDialog, setDeleteDialog] = useState<{ isOpen: boolean; domain?: DomainItem }>({
    isOpen: false,
  });
  const [isDeleting, setIsDeleting] = useState(false);

  const filteredDomains = domains.filter(domain =>
    domain.name.toLowerCase().includes(searchQuery.toLowerCase())
  );

  const handleDelete = async () => {
    if (!deleteDialog.domain) return;
    
    setIsDeleting(true);
    setTimeout(() => {
      setDomains(prev => prev.filter(d => d.id !== deleteDialog.domain?.id));
      showToast({
        type: 'success',
        title: 'Domain deleted',
        message: `${deleteDialog.domain.name} has been removed`,
      });
      setDeleteDialog({ isOpen: false });
      setIsDeleting(false);
    }, 1000);
  };

  const handleRefresh = () => {
    showToast({
      type: 'info',
      title: 'Refreshing',
      message: 'Domain list updated',
    });
  };

  const columns: Column<DomainItem>[] = [
    {
      key: 'name',
      header: 'Domain Name',
      sortable: true,
      render: (item) => (
        <div className="flex items-center gap-2">
          <Globe size={16} className="text-[var(--color-text-muted)]" />
          <span className="font-medium text-[var(--color-text-primary)]">{item.name}</span>
          <button className="text-[var(--color-text-muted)] hover:text-[var(--color-text-primary)]">
            <ExternalLink size={14} />
          </button>
        </div>
      ),
    },
    {
      key: 'status',
      header: 'Status',
      sortable: true,
      render: (item) => <StatusBadge status={item.status} />,
    },
    {
      key: 'ssl',
      header: 'SSL',
      render: (item) => (
        <div className="flex items-center gap-2">
          <Shield
            size={16}
            className={item.ssl ? 'text-green-400' : 'text-gray-400'}
          />
          <span className="text-sm text-[var(--color-text-secondary)]">
            {item.ssl ? 'Enabled' : 'Disabled'}
          </span>
        </div>
      ),
    },
    {
      key: 'ipAddress',
      header: 'IP Address',
      sortable: true,
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
            setDeleteDialog({ isOpen: true, domain: item });
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
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-[var(--color-text-primary)]">Domains</h1>
          <p className="text-[var(--color-text-secondary)] mt-1">
            Manage your domain names and DNS settings
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
            Add Domain
          </Button>
        </div>
      </div>

      <Card>
        <CardContent>
          <Input
            placeholder="Search domains..."
            leftIcon={<Search size={18} />}
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
          />
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>All Domains ({filteredDomains.length})</CardTitle>
        </CardHeader>
        <CardContent>
          {filteredDomains.length === 0 ? (
            <EmptyState
              icon={<Globe size={32} className="text-[var(--color-text-muted)]" />}
              title="No domains found"
              description={
                searchQuery
                  ? 'No domains match your search'
                  : 'Add your first domain to get started'
              }
              action={
                !searchQuery
                  ? {
                      label: 'Add Domain',
                      onClick: () => setIsCreateModalOpen(true),
                    }
                  : undefined
              }
            />
          ) : (
            <Table
              data={filteredDomains}
              columns={columns}
              keyExtractor={(item) => item.id}
            />
          )}
        </CardContent>
      </Card>

      <CreateDomainModal
        isOpen={isCreateModalOpen}
        onClose={() => setIsCreateModalOpen(false)}
      />

      <ConfirmDialog
        isOpen={deleteDialog.isOpen}
        onClose={() => setDeleteDialog({ isOpen: false })}
        onConfirm={handleDelete}
        title="Delete Domain"
        message={`Are you sure you want to delete "${deleteDialog.domain?.name}"? This will remove all DNS records and configurations.`}
        confirmText="Delete Domain"
        variant="danger"
        loading={isDeleting}
      />
    </div>
  );
}
