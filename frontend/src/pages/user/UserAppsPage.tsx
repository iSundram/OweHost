import { PlaceholderPage } from '../../components/PlaceholderPage';

export function UserAppsPage() {
  return (
    <PlaceholderPage
      title="One-Click Apps"
      description="Install and manage applications"
      features={[
        'Browse available applications (WordPress, Joomla, etc.)',
        'Install application',
        'Manage installed apps',
        'Update applications',
        'Uninstall applications',
      ]}
    />
  );
}
