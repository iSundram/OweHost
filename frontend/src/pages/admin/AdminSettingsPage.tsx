import { PlaceholderPage } from '../../components/PlaceholderPage';

export function AdminSettingsPage() {
  return (
    <PlaceholderPage
      title="System Settings"
      description="Configure system-wide settings and preferences"
      features={[
        'General settings (branding, timezone)',
        'API configuration (rate limits, versioning)',
        'Authentication settings (JWT expiry, session limits)',
        'Default resource limits',
        'System maintenance mode',
        'Feature toggles',
        'OS-level controls',
      ]}
    />
  );
}
