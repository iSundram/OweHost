import { PlaceholderPage } from '../../components/PlaceholderPage';

export function AdminPluginsPage() {
  return (
    <PlaceholderPage
      title="Plugin Management"
      description="Install and manage plugins and extensions"
      features={[
        'Install/Remove plugins',
        'Plugin signature verification',
        'Plugin configuration',
        'Sandboxed execution controls',
        'Plugin API access management',
      ]}
    />
  );
}
