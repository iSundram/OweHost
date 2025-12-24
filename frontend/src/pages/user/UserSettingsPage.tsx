import { PlaceholderPage } from '../../components/PlaceholderPage';

export function UserSettingsPage() {
  return (
    <PlaceholderPage
      title="Account Settings"
      description="Manage your account information and preferences"
      features={[
        'Profile information',
        'Change password',
        'Email notifications preferences',
        'Two-factor authentication (if implemented)',
        'API keys',
        'Session management (view active sessions, logout)',
      ]}
    />
  );
}
