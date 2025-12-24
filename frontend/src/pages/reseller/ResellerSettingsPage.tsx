import { PlaceholderPage } from '../../components/PlaceholderPage';

export function ResellerSettingsPage() {
  return (
    <PlaceholderPage
      title="Reseller Settings"
      description="Account information and preferences"
      features={[
        'Account information',
        'Branding customization (if allowed)',
        'Notification preferences',
        'API key management',
        'Password change',
      ]}
    />
  );
}
