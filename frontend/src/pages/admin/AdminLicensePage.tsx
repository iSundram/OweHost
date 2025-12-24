import { PlaceholderPage } from '../../components/PlaceholderPage';

export function AdminLicensePage() {
  return (
    <PlaceholderPage
      title="License Management"
      description="View license status and manage feature flags"
      features={[
        'View license status',
        'Feature flag configuration',
        'Plan-based gating controls',
        'License validation monitoring',
        'Usage reporting',
      ]}
    />
  );
}
