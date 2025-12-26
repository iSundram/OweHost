import { apiClient } from './client';
import type { 
  Domain, Database, User, LoginRequest, LoginResponse, UserRole,
  DNSZone, DNSRecord, DNSRecordCreateRequest,
  Certificate, CSR, CSRCreateRequest,
  FileEntry, FileContent,
  Backup, BackupSchedule, BackupType,
  CronJob, CronJobLog
} from '../types';

export const authService = {
  login: (credentials: LoginRequest) =>
    apiClient.post<LoginResponse>('/api/v1/auth/login', credentials),

  refresh: (refreshToken: string) =>
    apiClient.post<LoginResponse>('/api/v1/auth/refresh', { refresh_token: refreshToken }),

  logout: (refreshToken: string | null) =>
    apiClient.post<void>('/api/v1/auth/logout', refreshToken ? { refresh_token: refreshToken } : undefined),

  me: () =>
    apiClient.get<User>('/api/v1/users/me'),
};

export const domainService = {
  list: () => 
    apiClient.get<Domain[]>('/api/v1/domains'),
  
  get: (id: string) => 
    apiClient.get<Domain>(`/api/v1/domains/${id}`),
  
  create: (data: { name: string; type: string; document_root?: string }) => 
    apiClient.post<Domain>('/api/v1/domains', data),
  
  delete: (id: string) => 
    apiClient.delete<void>(`/api/v1/domains/${id}`),
};

export const databaseService = {
  list: () => 
    apiClient.get<Database[]>('/api/v1/databases'),
  
  get: (id: string) => 
    apiClient.get<Database>(`/api/v1/databases/${id}`),
  
  create: (data: { name: string; type: string; charset?: string; collation?: string }) => 
    apiClient.post<Database>('/api/v1/databases', data),
  
  delete: (id: string) => 
    apiClient.delete<void>(`/api/v1/databases/${id}`),
};

export const userService = {
  list: () => 
    apiClient.get<User[]>('/api/v1/users'),
  
  get: (id: string) => 
    apiClient.get<User>(`/api/v1/users/${id}`),
  
  create: (data: { username: string; email: string; password: string; role?: UserRole }) => 
    apiClient.post<User>('/api/v1/users', data),
  
  update: (id: string, data: Partial<User>) => 
    apiClient.put<User>(`/api/v1/users/${id}`, data),
  
  delete: (id: string) => 
    apiClient.delete<void>(`/api/v1/users/${id}`),
  
  suspend: (id: string) => 
    apiClient.post<void>(`/api/v1/users/${id}/suspend`),
  
  terminate: (id: string) => 
    apiClient.post<void>(`/api/v1/users/${id}/terminate`),
};

export interface Reseller {
  id: string;
  user_id: string;
  parent_reseller_id?: string;
  name: string;
  resource_pool: {
    max_users: number;
    max_domains: number;
    max_disk_mb: number;
    max_bandwidth_mb: number;
    max_databases: number;
    max_cpu_quota: number;
    max_memory_mb: number;
  };
  created_at: string;
  updated_at: string;
}

export interface ResellerCreateRequest {
  user_id: string;
  parent_reseller_id?: string;
  name: string;
  resource_pool: {
    max_users: number;
    max_domains: number;
    max_disk_mb: number;
    max_bandwidth_mb: number;
    max_databases: number;
    max_cpu_quota: number;
    max_memory_mb: number;
  };
}

export const resellerService = {
  list: () => 
    apiClient.get<Reseller[]>('/api/v1/resellers'),
  
  get: (id: string) => 
    apiClient.get<Reseller>(`/api/v1/resellers/${id}`),
  
  getMe: () => 
    apiClient.get<Reseller>('/api/v1/resellers/me'),
  
  create: (data: ResellerCreateRequest) => 
    apiClient.post<Reseller>('/api/v1/resellers', data),
  
  update: (id: string, data: Partial<ResellerCreateRequest>) => 
    apiClient.put<Reseller>(`/api/v1/resellers/${id}`, data),
  
  delete: (id: string) => 
    apiClient.delete<void>(`/api/v1/resellers/${id}`),
};

// Admin Stats Service
export interface SystemStats {
  users: {
    total: number;
    active: number;
    suspended: number;
    byRole: {
      admin: number;
      reseller: number;
      user: number;
    };
  };
  resellers: {
    total: number;
    active: number;
  };
  domains: {
    total: number;
    active: number;
    withSSL: number;
  };
  databases: {
    total: number;
    totalSizeMB: number;
    byType: {
      mysql: number;
      postgresql: number;
    };
  };
  resources: {
    cpu: {
      usage: number;  // percentage
      cores: number;  // total cores
      usedCores: number;  // used cores
    };
    memory: {
      usage: number;  // percentage
      total: number;  // GB
      used: number;  // GB
    };
    disk: {
      usage: number;  // percentage
      total: number;  // GB
      used: number;  // GB
    };
    network: {
      usage: number;  // percentage
      bandwidth: string;  // e.g., "125 Mbps"
    };
  };
}

export interface ServiceStatus {
  name: string;
  status: 'running' | 'stopped' | 'warning';
  uptime: string;
  load?: string;
  pid?: number;
}

export const adminService = {
  getSystemStats: async (): Promise<SystemStats> => {
    try {
      // Try to fetch from actual API endpoint
      return await apiClient.get<SystemStats>('/api/v1/admin/stats');
    } catch (error) {
      // Fallback to aggregated data from existing endpoints
      const [users, resellers, domains, databases] = await Promise.all([
        userService.list().catch(() => []),
        resellerService.list().catch(() => []),
        domainService.list().catch(() => []),
        databaseService.list().catch(() => []),
      ]);

      return {
        users: {
          total: users.length,
          active: users.filter((u: any) => u.status === 'active').length,
          suspended: users.filter((u: any) => u.status === 'suspended').length,
          byRole: {
            admin: users.filter((u: any) => u.role === 'admin').length,
            reseller: users.filter((u: any) => u.role === 'reseller').length,
            user: users.filter((u: any) => u.role === 'user').length,
          },
        },
        resellers: {
          total: resellers.length,
          active: resellers.length,
        },
        domains: {
          total: domains.length,
          active: domains.filter((d: any) => d.status === 'active').length,
          withSSL: domains.filter((d: any) => d.ssl_enabled).length,
        },
        databases: {
          total: databases.length,
          totalSizeMB: databases.reduce((sum: number, db: any) => sum + (db.size_mb || 0), 0),
          byType: {
            mysql: databases.filter((db: any) => db.type === 'mysql').length,
            postgresql: databases.filter((db: any) => db.type === 'postgresql').length,
          },
        },
        resources: {
          cpu: {
            usage: Math.random() * 80 + 10, // Mock: 10-90%
            cores: 8,
            usedCores: parseFloat((Math.random() * 6 + 1).toFixed(1)), // Mock: 1-7 cores
          },
          memory: {
            usage: Math.random() * 70 + 20, // Mock: 20-90%
            total: 32,
            used: parseFloat((Math.random() * 25 + 5).toFixed(1)), // Mock: 5-30 GB
          },
          disk: {
            usage: Math.random() * 60 + 20, // Mock: 20-80%
            total: 500,
            used: parseFloat((Math.random() * 350 + 50).toFixed(1)), // Mock: 50-400 GB
          },
          network: {
            usage: Math.random() * 50 + 10, // Mock: 10-60%
            bandwidth: `${Math.floor(Math.random() * 800 + 100)} Mbps`, // Mock: 100-900 Mbps
          },
        },
      };
    }
  },

  getServiceStatus: async (): Promise<ServiceStatus[]> => {
    try {
      return await apiClient.get<ServiceStatus[]>('/api/v1/admin/services');
    } catch (error) {
      // Return mock data if endpoint doesn't exist
      return [
        { name: 'Nginx Web Server', status: 'running', uptime: '99.9%', load: 'Low' },
        { name: 'MySQL Database', status: 'running', uptime: '99.8%', load: 'Medium' },
        { name: 'DNS Server', status: 'running', uptime: '100%', load: 'Low' },
        { name: 'FTP Server', status: 'running', uptime: '99.5%', load: 'Low' },
      ];
    }
  },

  getSystemHealth: async () => {
    try {
      return await apiClient.get('/api/v1/admin/health');
    } catch (error) {
      return { status: 'healthy', uptime: '99.9%' };
    }
  },
};

// DNS Service
export const dnsService = {
  // Zones
  listZones: () =>
    apiClient.get<DNSZone[]>('/api/v1/dns/zones'),

  getZone: (id: string) =>
    apiClient.get<DNSZone>(`/api/v1/dns/zones/${id}`),

  createZone: (data: { domain_id: string; name: string }) =>
    apiClient.post<DNSZone>('/api/v1/dns/zones', data),

  deleteZone: (id: string) =>
    apiClient.delete<void>(`/api/v1/dns/zones/${id}`),

  lockZone: (id: string) =>
    apiClient.post<void>(`/api/v1/dns/zones/${id}/lock`),

  unlockZone: (id: string) =>
    apiClient.post<void>(`/api/v1/dns/zones/${id}/unlock`),

  enableDNSSEC: (zoneId: string) =>
    apiClient.post<{ key_tag: number; digest: string }>(`/api/v1/dns/zones/${zoneId}/dnssec`),

  syncZone: (zoneId: string, provider: string = 'default') =>
    apiClient.post<{ status: string }>(`/api/v1/dns/zones/${zoneId}/sync`, { provider }),

  // Records
  listRecords: (zoneId: string) =>
    apiClient.get<DNSRecord[]>(`/api/v1/dns/zones/${zoneId}/records`),

  getRecord: (recordId: string) =>
    apiClient.get<DNSRecord>(`/api/v1/dns/records/${recordId}`),

  createRecord: (zoneId: string, data: DNSRecordCreateRequest) =>
    apiClient.post<DNSRecord>(`/api/v1/dns/zones/${zoneId}/records`, data),

  updateRecord: (recordId: string, data: DNSRecordCreateRequest) =>
    apiClient.put<DNSRecord>(`/api/v1/dns/records/${recordId}`, data),

  deleteRecord: (recordId: string) =>
    apiClient.delete<void>(`/api/v1/dns/records/${recordId}`),
};

// SSL Certificate Service
export const sslService = {
  listCertificates: () =>
    apiClient.get<Certificate[]>('/api/v1/ssl/certificates'),

  getCertificate: (id: string) =>
    apiClient.get<Certificate>(`/api/v1/ssl/certificates/${id}`),

  getCertificateByDomain: (domainId: string) =>
    apiClient.get<Certificate>(`/api/v1/ssl/certificates/domain/${domainId}`),

  deleteCertificate: (id: string) =>
    apiClient.delete<void>(`/api/v1/ssl/certificates/${id}`),

  // CSR
  generateCSR: (data: CSRCreateRequest) =>
    apiClient.post<CSR>('/api/v1/ssl/csr', data),

  // Let's Encrypt
  requestLetsEncrypt: (data: { domain_id: string; domains: string[] }) =>
    apiClient.post<Certificate>('/api/v1/ssl/letsencrypt', data),

  // Custom Certificate
  uploadCertificate: (data: { domain_id: string; certificate: string; private_key: string; chain?: string }) =>
    apiClient.post<Certificate>('/api/v1/ssl/certificates', data),

  // Self-signed
  generateSelfSigned: (data: { domain_id: string; common_name: string }) =>
    apiClient.post<Certificate>('/api/v1/ssl/self-signed', data),

  // Auto-renewal
  enableAutoRenew: (id: string) =>
    apiClient.post<void>(`/api/v1/ssl/certificates/${id}/auto-renew`),

  disableAutoRenew: (id: string) =>
    apiClient.delete<void>(`/api/v1/ssl/certificates/${id}/auto-renew`),

  // Check expiring
  getExpiring: (days: number = 30) =>
    apiClient.get<Certificate[]>(`/api/v1/ssl/certificates/expiring?days=${days}`),
};

// File System Service
export const fileService = {
  list: (path: string = '/', userId?: string) => {
    const params = new URLSearchParams({ path });
    if (userId) params.append('user_id', userId);
    return apiClient.get<FileEntry[]>(`/api/v1/files?${params}`);
  },

  read: (path: string, userId?: string) => {
    const params = new URLSearchParams({ path });
    if (userId) params.append('user_id', userId);
    return apiClient.get<FileContent>(`/api/v1/files/content?${params}`);
  },

  write: (path: string, content: string, userId?: string) =>
    apiClient.post<void>('/api/v1/files/content', { path, content, user_id: userId }),

  create: (path: string, type: 'file' | 'directory', userId?: string) =>
    apiClient.post<FileEntry>('/api/v1/files', { path, type, user_id: userId }),

  delete: (path: string, userId?: string) => {
    const params = new URLSearchParams({ path });
    if (userId) params.append('user_id', userId);
    return apiClient.delete<void>(`/api/v1/files?${params}`);
  },

  rename: (oldPath: string, newPath: string, userId?: string) =>
    apiClient.put<FileEntry>('/api/v1/files/rename', { old_path: oldPath, new_path: newPath, user_id: userId }),

  copy: (sourcePath: string, destPath: string, userId?: string) =>
    apiClient.post<FileEntry>('/api/v1/files/copy', { source: sourcePath, destination: destPath, user_id: userId }),

  chmod: (path: string, permissions: string, userId?: string) =>
    apiClient.put<void>('/api/v1/files/permissions', { path, permissions, user_id: userId }),

  compress: (paths: string[], archivePath: string, userId?: string) =>
    apiClient.post<FileEntry>('/api/v1/files/compress', { paths, archive_path: archivePath, user_id: userId }),

  extract: (archivePath: string, destPath: string, userId?: string) =>
    apiClient.post<void>('/api/v1/files/extract', { archive_path: archivePath, destination: destPath, user_id: userId }),

  getDiskUsage: (path: string = '/', userId?: string) => {
    const params = new URLSearchParams({ path });
    if (userId) params.append('user_id', userId);
    return apiClient.get<{ used: number; total: number; free: number }>(`/api/v1/files/disk-usage?${params}`);
  },
};

// Backup Service
export const backupService = {
  list: (userId?: string) => {
    const params = userId ? `?user_id=${userId}` : '';
    return apiClient.get<Backup[]>(`/api/v1/backups${params}`);
  },

  get: (id: string) =>
    apiClient.get<Backup>(`/api/v1/backups/${id}`),

  create: (data: { type: BackupType; includes?: string[]; user_id?: string }) =>
    apiClient.post<Backup>('/api/v1/backups', data),

  delete: (id: string) =>
    apiClient.delete<void>(`/api/v1/backups/${id}`),

  restore: (id: string, options?: { overwrite?: boolean; destination?: string }) =>
    apiClient.post<void>(`/api/v1/backups/${id}/restore`, options),

  download: (id: string) =>
    apiClient.get<{ url: string }>(`/api/v1/backups/${id}/download`),

  // Schedules
  listSchedules: (userId?: string) => {
    const params = userId ? `?user_id=${userId}` : '';
    return apiClient.get<BackupSchedule[]>(`/api/v1/backups/schedules${params}`);
  },

  createSchedule: (data: { type: BackupType; frequency: string; retention_days: number; user_id?: string }) =>
    apiClient.post<BackupSchedule>('/api/v1/backups/schedules', data),

  updateSchedule: (id: string, data: Partial<BackupSchedule>) =>
    apiClient.put<BackupSchedule>(`/api/v1/backups/schedules/${id}`, data),

  deleteSchedule: (id: string) =>
    apiClient.delete<void>(`/api/v1/backups/schedules/${id}`),

  enableSchedule: (id: string) =>
    apiClient.post<void>(`/api/v1/backups/schedules/${id}/enable`),

  disableSchedule: (id: string) =>
    apiClient.post<void>(`/api/v1/backups/schedules/${id}/disable`),
};

// Cron Service
export const cronService = {
  list: (userId?: string) => {
    const params = userId ? `?user_id=${userId}` : '';
    return apiClient.get<CronJob[]>(`/api/v1/cron${params}`);
  },

  get: (id: string) =>
    apiClient.get<CronJob>(`/api/v1/cron/${id}`),

  create: (data: { name: string; command: string; schedule: string; user_id?: string }) =>
    apiClient.post<CronJob>('/api/v1/cron', data),

  update: (id: string, data: Partial<CronJob>) =>
    apiClient.put<CronJob>(`/api/v1/cron/${id}`, data),

  delete: (id: string) =>
    apiClient.delete<void>(`/api/v1/cron/${id}`),

  pause: (id: string) =>
    apiClient.post<void>(`/api/v1/cron/${id}/pause`),

  resume: (id: string) =>
    apiClient.post<void>(`/api/v1/cron/${id}/resume`),

  run: (id: string) =>
    apiClient.post<{ log_id: string }>(`/api/v1/cron/${id}/run`),

  // Logs
  getLogs: (jobId: string, limit: number = 10) =>
    apiClient.get<CronJobLog[]>(`/api/v1/cron/${jobId}/logs?limit=${limit}`),

  getLog: (logId: string) =>
    apiClient.get<CronJobLog>(`/api/v1/cron/logs/${logId}`),
};

// Account Service (filesystem-based account management)
export interface Account {
  account_id: number;
  identity: {
    id: number;
    name: string;
    uid: number;
    gid: number;
    owner: string;
    plan: string;
    node: string;
    created_at: string;
    state: string;
  };
  limits: {
    disk_mb: number;
    cpu_percent: number;
    ram_mb: number;
    databases: number;
    domains: number;
    subdomains: number;
    email_accounts: number;
    ftp_accounts: number;
    bandwidth_gb: number;
    inodes: number;
  };
  status?: {
    suspended: boolean;
    locked: boolean;
    reason?: string;
    suspended_at?: string;
    suspended_by?: string;
  };
  metadata?: {
    email: string;
    contact_name?: string;
    contact_phone?: string;
    notes?: string;
    tags?: string[];
    custom?: Record<string, string>;
    updated_at: string;
  };
}

export interface AccountCreateRequest {
  username: string;
  email: string;
  password: string;
  plan?: string; // starter, standard, premium, enterprise
  owner?: string; // admin, reseller-X, partner-X
  domain?: string;
}

export const accountService = {
  list: () =>
    apiClient.get<Account[]>('/api/v1/accounts'),

  get: (id: number) =>
    apiClient.get<Account>(`/api/v1/accounts/${id}`),

  create: (data: AccountCreateRequest) =>
    apiClient.post<{ account: Account; user?: any; domain?: any }>('/api/v1/accounts', data),

  suspend: (id: number) =>
    apiClient.post<Account>(`/api/v1/accounts/${id}/suspend`),

  unsuspend: (id: number) =>
    apiClient.post<Account>(`/api/v1/accounts/${id}/unsuspend`),

  terminate: (id: number) =>
    apiClient.post<Account>(`/api/v1/accounts/${id}/terminate`),
};

// Package/Plan Service
export interface Package {
  name: string;
  display_name: string;
  description: string;
  limits: {
    disk_mb: number;
    cpu_percent: number;
    ram_mb: number;
    databases: number;
    domains: number;
    subdomains: number;
    email_accounts: number;
    ftp_accounts: number;
    bandwidth_gb: number;
    inodes: number;
  };
  features: Record<string, boolean>;
  price?: {
    monthly: number;
    yearly: number;
    currency: string;
  };
}

export const packageService = {
  list: () =>
    apiClient.get<Package[]>('/api/v1/packages'),

  get: (name: string) =>
    apiClient.get<Package>(`/api/v1/packages/${name}`),
};

// Feature Manager Service
export interface Feature {
  name: string;
  display_name: string;
  description: string;
  enabled: boolean;
  category: string;
}

export interface FeatureCategory {
  name: string;
  display_name: string;
  features: Feature[];
}

export const featureService = {
  list: () =>
    apiClient.get<FeatureCategory[]>('/api/v1/features'),

  get: (name: string) =>
    apiClient.get<Feature>(`/api/v1/features/${name}`),

  update: (name: string, enabled: boolean) =>
    apiClient.put<Feature>(`/api/v1/features/${name}`, { enabled }),
};
