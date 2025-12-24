// User types
export type UserRole = 'admin' | 'reseller' | 'user';
export type UserStatus = 'active' | 'suspended' | 'terminated';

export interface User {
  id: string;
  tenant_id: string;
  username: string;
  email: string;
  role: UserRole;
  status: UserStatus;
  uid: number;
  gid: number;
  home_directory: string;
  namespace: string;
  created_at: string;
  updated_at: string;
}

// Domain types
export interface Domain {
  id: string;
  user_id: string;
  name: string;
  type: 'primary' | 'addon' | 'parked' | 'alias';
  status: 'active' | 'pending' | 'suspended';
  document_root: string;
  validated: boolean;
  created_at: string;
}

// Database types
export interface Database {
  id: string;
  user_id: string;
  name: string;
  type: 'mysql' | 'postgresql' | 'mariadb' | 'mongodb' | 'redis' | 'sqlite';
  charset?: string;
  collation?: string;
  size_mb: number;
  created_at: string;
  updated_at: string;
}

// Installation types
export interface DatabaseEngine {
  type: string;
  name: string;
  description: string;
  default_port: number;
  is_installed: boolean;
}

export interface InstallationCheckResponse {
  is_installed: boolean;
  requires_setup: boolean;
  supported_engines: DatabaseEngine[];
}

export interface InstallationRequest {
  database_engine: string;
  database_host: string;
  database_port: number;
  database_name: string;
  database_user: string;
  database_password: string;
  admin_email: string;
}

export interface Installation {
  id: string;
  status: 'pending' | 'in_progress' | 'completed' | 'failed';
  database_engine: string;
  database_host: string;
  database_port: number;
  database_name: string;
  admin_username: string;
  admin_email: string;
  installation_step: number;
  total_steps: number;
  error_message?: string;
  created_at: string;
  completed_at?: string;
}

// Auth types
export interface LoginRequest {
  username: string;
  password: string;
}

export interface LoginResponse {
  access_token: string;
  refresh_token: string;
  expires_in: number;
  token_type: string;
}

// API Response types
export interface ApiResponse<T> {
  success: boolean;
  data?: T;
  error?: {
    code: string;
    message: string;
  };
}

// Stats types for dashboard
export interface DashboardStats {
  totalDomains: number;
  activeDomains: number;
  totalDatabases: number;
  totalUsers: number;
  diskUsage: {
    used: number;
    total: number;
  };
  bandwidthUsage: {
    used: number;
    total: number;
  };
}

// DNS types
export interface DNSZone {
  id: string;
  domain_id: string;
  name: string;
  locked: boolean;
  dnssec_enabled: boolean;
  created_at: string;
  updated_at: string;
}

export type DNSRecordType = 'A' | 'AAAA' | 'CNAME' | 'MX' | 'TXT' | 'SRV' | 'NS' | 'CAA' | 'PTR';

export interface DNSRecord {
  id: string;
  zone_id: string;
  name: string;
  type: DNSRecordType;
  content: string;
  ttl: number;
  priority?: number;
  created_at: string;
  updated_at: string;
}

export interface DNSRecordCreateRequest {
  name: string;
  type: DNSRecordType;
  content: string;
  ttl?: number;
  priority?: number;
}

// SSL Certificate types
export type CertificateType = 'letsencrypt' | 'custom' | 'self-signed';
export type CertificateStatus = 'active' | 'expired' | 'pending' | 'revoked';

export interface Certificate {
  id: string;
  domain_id: string;
  type: CertificateType;
  status: CertificateStatus;
  common_name: string;
  sans?: string[];
  issuer: string;
  serial_number: string;
  issued_at: string;
  expires_at: string;
  auto_renew: boolean;
  created_at: string;
  updated_at: string;
}

export interface CSRCreateRequest {
  domain_id: string;
  common_name: string;
  organization: string;
  country: string;
  state: string;
  city: string;
  sans?: string[];
}

export interface CSR {
  id: string;
  domain_id: string;
  common_name: string;
  csr_data: string;
  created_at: string;
}

// File System types
export type FileType = 'file' | 'directory' | 'symlink';

export interface FileEntry {
  name: string;
  path: string;
  type: FileType;
  size: number;
  permissions: string;
  owner: string;
  group: string;
  modified_at: string;
}

export interface FileContent {
  path: string;
  content: string;
  encoding: string;
}

// Backup types
export type BackupType = 'full' | 'incremental' | 'differential';
export type BackupStatus = 'pending' | 'in_progress' | 'completed' | 'failed';

export interface Backup {
  id: string;
  user_id: string;
  type: BackupType;
  status: BackupStatus;
  size_mb: number;
  includes: string[];
  storage_location: string;
  created_at: string;
  completed_at?: string;
  expires_at?: string;
}

export interface BackupSchedule {
  id: string;
  user_id: string;
  type: BackupType;
  frequency: 'daily' | 'weekly' | 'monthly';
  retention_days: number;
  enabled: boolean;
  next_run_at: string;
  created_at: string;
}

// Cron Job types
export type CronJobStatus = 'active' | 'paused' | 'disabled';

export interface CronJob {
  id: string;
  user_id: string;
  name: string;
  command: string;
  schedule: string;
  status: CronJobStatus;
  last_run_at?: string;
  next_run_at?: string;
  last_exit_code?: number;
  created_at: string;
  updated_at: string;
}

export interface CronJobLog {
  id: string;
  job_id: string;
  started_at: string;
  completed_at: string;
  exit_code: number;
  stdout: string;
  stderr: string;
}
