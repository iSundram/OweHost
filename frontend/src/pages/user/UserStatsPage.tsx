import { PlaceholderPage } from '../../components/PlaceholderPage';

export function UserStatsPage() {
  return (
    <PlaceholderPage
      title="Statistics & Analytics"
      description="View usage statistics and analytics"
      features={[
        'Bandwidth usage graph',
        'Disk usage breakdown',
        'Visitor statistics (if available)',
        'Database size trends',
        'Resource usage history',
      ]}
    />
  );
}
