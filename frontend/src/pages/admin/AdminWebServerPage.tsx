import { PlaceholderPage } from '../../components/PlaceholderPage';

export function AdminWebServerPage() {
  return (
    <PlaceholderPage
      title="Web Server Configuration"
      description="Nginx/Apache configuration and virtual host management"
      features={[
        'Nginx/Apache configuration management',
        'Virtual host management',
        'Global server settings',
        'SSL/TLS configuration',
        'Web server restart/reload controls',
        'Config validation and rollback',
        'Error log viewing',
      ]}
    />
  );
}
