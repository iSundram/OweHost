import { PlaceholderPage } from '../../components/PlaceholderPage';

export function UserSecurityPage() {
  return (
    <PlaceholderPage
      title="Security Settings"
      description="Configure security settings for your account"
      features={[
        'IP whitelist/blacklist',
        'Password-protected directories',
        'SSL force redirect',
        'Hotlink protection',
        'Security logs',
      ]}
    />
  );
}
