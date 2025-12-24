import { NavLink, useLocation } from 'react-router-dom';
import {
  LayoutDashboard,
  Globe,
  Database,
  FolderOpen,
  Shield,
  Mail,
  Plug,
  Settings as SettingsIcon,
  Clock,
  Package,
  HardDrive,
  BarChart,
  Lock,
  HelpCircle,
  ChevronLeft,
  ChevronRight,
} from 'lucide-react';
import { useState } from 'react';

const navItems = [
  { path: '/user', icon: LayoutDashboard, label: 'Dashboard' },
  { path: '/user/domains', icon: Globe, label: 'Domains' },
  { path: '/user/dns', icon: Globe, label: 'DNS' },
  { path: '/user/databases', icon: Database, label: 'Databases' },
  { path: '/user/files', icon: FolderOpen, label: 'File Manager' },
  { path: '/user/ssl', icon: Shield, label: 'SSL' },
  { path: '/user/email', icon: Mail, label: 'Email' },
  { path: '/user/ftp', icon: Plug, label: 'FTP' },
  { path: '/user/webserver', icon: SettingsIcon, label: 'Web Server' },
  { path: '/user/cron', icon: Clock, label: 'Cron Jobs' },
  { path: '/user/apps', icon: Package, label: 'Applications' },
  { path: '/user/backups', icon: HardDrive, label: 'Backups' },
  { path: '/user/stats', icon: BarChart, label: 'Statistics' },
  { path: '/user/security', icon: Lock, label: 'Security' },
];

const bottomNavItems = [
  { path: '/user/settings', icon: SettingsIcon, label: 'Settings' },
  { path: '/user/support', icon: HelpCircle, label: 'Support' },
];

export function UserSidebar() {
  const [collapsed, setCollapsed] = useState(() => {
    // Load collapsed state from localStorage, default to true
    const saved = localStorage.getItem('userSidebarCollapsed');
    return saved ? JSON.parse(saved) : true;
  });
  const location = useLocation();

  // Save collapsed state to localStorage
  const toggleCollapsed = () => {
    const newState = !collapsed;
    setCollapsed(newState);
    localStorage.setItem('userSidebarCollapsed', JSON.stringify(newState));
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
              OweHost
            </span>
          )}
        </div>
      </div>

      {/* Navigation */}
      <nav className="flex-1 py-4 overflow-y-auto">
        <div className="px-3 space-y-1">
          {navItems.map((item) => {
            // Exact match for dashboard, path-based match for others
            const isActive = item.path === '/user' 
              ? location.pathname === '/user'
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
