import { PlaceholderPage } from '../../components/PlaceholderPage';

export function UserEmailPage() {
  return (
    <PlaceholderPage
      title="Email Management"
      description="Manage email accounts and settings (Future feature)"
      features={[
        'Create/Delete email accounts',
        'Email forwarders',
        'Email aliases',
        'Autoresponders',
        'Webmail access',
        'Spam filters',
      ]}
    />
  );
}
