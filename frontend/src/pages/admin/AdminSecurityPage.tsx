import { PlaceholderPage } from '../../components/PlaceholderPage';

export function AdminSecurityPage() {
  return (
    <PlaceholderPage
      title="Security & Firewall"
      description="Security administration and firewall management"
      features={[
        'Firewall rule management',
        'Rate limiting configuration (per-IP, per-user)',
        'Intrusion detection monitoring',
        'Security audit logs',
        'IP whitelist/blacklist management',
        'DDoS protection settings',
        'Security policy enforcement',
      ]}
    />
  );
}
