import { PlaceholderPage } from '../../components/PlaceholderPage';

export function ResellerResourcesPage() {
  return (
    <PlaceholderPage
      title="Resource Allocation"
      description="Manage resource pool and customer allocations"
      features={[
        'View allocated resource pool',
        'Distribute resources to customers',
        'Set per-customer limits (CPU, RAM, Disk, Bandwidth)',
        'Create custom packages',
        'Resource usage analytics',
        'Over-allocation warnings',
      ]}
    />
  );
}
