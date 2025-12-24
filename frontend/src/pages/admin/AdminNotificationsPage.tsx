import { PlaceholderPage } from '../../components/PlaceholderPage';

export function AdminNotificationsPage() {
  return (
    <PlaceholderPage
      title="Notification Management"
      description="Configure event notifications and webhooks"
      features={[
        'Event bus monitoring',
        'Webhook configuration',
        'Email notification settings',
        'Alert rules and thresholds',
        'Notification history',
      ]}
    />
  );
}
