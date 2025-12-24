import { PlaceholderPage } from '../../components/PlaceholderPage';

export function UserWebServerPage() {
  return (
    <PlaceholderPage
      title="Web Server Settings"
      description="Configure web server settings for your account"
      features={[
        'PHP version selection',
        'PHP settings (memory_limit, max_execution_time, etc.)',
        'Error pages (404, 500)',
        'Directory index settings',
        'Hotlink protection',
        'IP blocking',
      ]}
    />
  );
}
