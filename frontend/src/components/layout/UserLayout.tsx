import { Outlet } from 'react-router-dom';
import { UserSidebar } from './UserSidebar';
import { Header } from './Header';

export function UserLayout() {
  return (
    <div className="h-screen flex overflow-hidden">
      <UserSidebar />
      <div className="flex-1 flex flex-col overflow-hidden">
        <Header />
        <main className="flex-1 overflow-auto bg-[var(--color-background)] p-6">
          <Outlet />
        </main>
      </div>
    </div>
  );
}
