import { PlaceholderPage } from '../../components/PlaceholderPage';

export function UserFTPPage() {
  return (
    <PlaceholderPage
      title="FTP Accounts"
      description="Manage FTP accounts for file access"
      features={[
        'Create FTP accounts',
        'Delete FTP accounts',
        'Set FTP directory',
        'Change FTP passwords',
        'FTP connection details',
      ]}
    />
  );
}
