import { PlaceholderPage } from '../../components/PlaceholderPage';

export function AdminAppsPage() {
  return (
    <PlaceholderPage
      title="Application Installer Management"
      description="Manage available applications and installation templates"
      features={[
        'Manage available applications (WordPress, etc.)',
        'Version management',
        'Custom app definitions',
        'Installation templates',
        'Update management',
      ]}
    />
  );
}
