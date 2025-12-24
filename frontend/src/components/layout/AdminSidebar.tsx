import { NavLink, useLocation } from 'react-router-dom';
import {
  LayoutDashboard,
  Users,
  Store,
  Zap,
  Globe,
  Database,
  Settings,
  Shield,
  FolderOpen,
  HardDrive,
  Lock,
  Clock,
  Package,
  Server,
  Puzzle,
  FileText,
  Bell,
  Key,
  Wrench,
  ChevronLeft,
  ChevronRight,
} from 'lucide-react';
import { useState } from 'react';

const navItems = [
  { path: '/admin', icon: LayoutDashboard, label: 'Dashboard' },
  { path: '/admin/users', icon: Users, label: 'Users' },
  { path: '/admin/resellers', icon: Store, label: 'Resellers' },
  { path: '/admin/resources', icon: Zap, label: 'Resources' },
  { path: '/admin/domains', icon: Globe, label: 'Domains' },
  { path: '/admin/databases', icon: Database, label: 'Databases' },
  { path: '/admin/dns', icon: Globe, label: 'DNS' },
  { path: '/admin/ssl', icon: Shield, label: 'SSL' },
  { path: '/admin/webserver', icon: Server, label: 'Web Server' },
  { path: '/admin/files', icon: FolderOpen, label: 'Files' },
  { path: '/admin/backups', icon: HardDrive, label: 'Backups' },
  { path: '/admin/security', icon: Lock, label: 'Security' },
  { path: '/admin/cron', icon: Clock, label: 'Cron Jobs' },
  { path: '/admin/apps', icon: Package, label: 'Applications' },
  { path: '/admin/nodes', icon: Server, label: 'Nodes' },
  { path: '/admin/plugins', icon: Puzzle, label: 'Plugins' },
  { path: '/admin/logs', icon: FileText, label: 'Logs' },
  { path: '/admin/notifications', icon: Bell, label: 'Notifications' },
  { path: '/admin/license', icon: Key, label: 'License' },
];

const bottomNavItems = [
  { path: '/admin/settings', icon: Settings, label: 'Settings' },
  { path: '/admin/recovery', icon: Wrench, label: 'Recovery' },
];

export function AdminSidebar() {
  const [collapsed, setCollapsed] = useState(() => {
    // Load collapsed state from localStorage, default to true
    const saved = localStorage.getItem('adminSidebarCollapsed');
    return saved ? JSON.parse(saved) : true;
  });
  const location = useLocation();

  // Save collapsed state to localStorage
  const toggleCollapsed = () => {
    const newState = !collapsed;
    setCollapsed(newState);
    localStorage.setItem('adminSidebarCollapsed', JSON.stringify(newState));
  };

  return (
    <aside
      className={`
        h-full flex flex-col
        bg-[var(--color-surface)]
        border-r border-[var(--color-border-light)]
        transition-all duration-300
        ${collapsed ? 'w-20' : 'w-64'}
      `}
    >
      {/* Logo */}
      <div className="h-16 flex items-center justify-center border-b border-[var(--color-border-light)]">
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 rounded-lg bg-gradient-to-br from-[#7BA4D0] to-[#E7F0FA] flex items-center justify-center shadow-lg">
            <span className="text-white font-bold text-lg">O</span>
          </div>
          {!collapsed && (
            <span className="text-xl font-bold text-[var(--color-text-primary)]">
              Admin Panel
            </span>
          )}
        </div>
      </div>

      {/* Navigation */}
      <nav className="flex-1 py-4 overflow-y-auto">
        <div className="px-3 space-y-1">
          {navItems.map((item) => {
            // Exact match for dashboard, path-based match for others
            const isActive = item.path === '/admin' 
              ? location.pathname === '/admin'
              : location.pathname.startsWith(item.path);
            
            return (
              <NavLink
                key={item.path}
                to={item.path}
                className={`
                  flex items-center gap-3 px-3 py-2.5 rounded-lg
                  transition-colors duration-200 outline-none focus:outline-none
                  ${
                    isActive
                      ? 'bg-[#E7F0FA] !text-[#2E5E99] border-l-4 border-l-[#2E5E99] shadow-sm'
                      : '!text-[var(--color-text-secondary)] hover:bg-[#E7F0FA]/50 hover:!text-[#2E5E99] border-l-4 border-l-transparent'
                  }
                `}
              >
                <item.icon
                  size={20}
                  className={isActive ? 'text-[#2E5E99]' : 'text-[#7BA4D0]'}
                />
                {!collapsed && (
                  <span className="font-medium">{item.label}</span>
                )}
              </NavLink>
            );
          })}
        </div>
      </nav>

      {/* Bottom Navigation */}
      <div className="py-4 border-t border-[var(--color-border-light)]">
        <div className="px-3 space-y-1">
          {bottomNavItems.map((item) => {
            const isActive = location.pathname === item.path;
            return (
              <NavLink
                key={item.path}
                to={item.path}
                className={`
                  flex items-center gap-3 px-3 py-2.5 rounded-lg
                  transition-colors duration-200 outline-none focus:outline-none
                  ${
                    isActive
                      ? 'bg-gradient-to-r from-[#7BA4D0]/50 to-[#E7F0FA]/30 !text-[var(--color-text-primary)]'
                      : '!text-[var(--color-text-secondary)] hover:bg-[var(--color-primary-medium)]/20 hover:!text-[var(--color-text-primary)]'
                  }
                `}
              >
                <item.icon size={20} />
                {!collapsed && (
                  <span className="font-medium">{item.label}</span>
                )}
              </NavLink>
            );
          })}
        </div>
      </div>

      {/* Collapse Button */}
      <div className="p-3 border-t border-[var(--color-border-light)]">
        <button
          onClick={toggleCollapsed}
          className="w-full flex items-center justify-center gap-2 px-3 py-2 rounded-lg
            text-[var(--color-text-secondary)] hover:text-[var(--color-text-primary)]
            hover:bg-[var(--color-primary-medium)]/20
            transition-all duration-200"
        >
          {collapsed ? <ChevronRight size={20} /> : <ChevronLeft size={20} />}
          {!collapsed && <span className="text-sm">Collapse</span>}
        </button>
      </div>
    </aside>
  );
}
