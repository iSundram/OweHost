import { useState, useEffect } from 'react';
import {
  Folder,
  File,
  ChevronRight,
  ChevronDown,
  Upload,
  Download,
  Trash2,
  Edit2,
  Plus,
  Search,
  AlertCircle,
  FolderPlus,
  FilePlus,
  Copy,
  Scissors,
  Archive,
  RefreshCw,
  Home,
  HardDrive,
} from 'lucide-react';
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  Button,
  Badge,
  Input,
  Modal,
  Select,
} from '../../components/ui';
import { fileService, userService } from '../../api/services';
import type { FileEntry, User } from '../../types';

export function AdminFilesPage() {
  const [files, setFiles] = useState<FileEntry[]>([]);
  const [users, setUsers] = useState<User[]>([]);
  const [currentPath, setCurrentPath] = useState('/');
  const [selectedUser, setSelectedUser] = useState<string>('');
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedFiles, setSelectedFiles] = useState<Set<string>>(new Set());

  // Modal states
  const [showCreateModal, setShowCreateModal] = useState<'file' | 'directory' | null>(null);
  const [showRenameModal, setShowRenameModal] = useState<FileEntry | null>(null);
  const [showEditModal, setShowEditModal] = useState<FileEntry | null>(null);
  const [showDeleteConfirm, setShowDeleteConfirm] = useState<string | null>(null);
  const [showCompressModal, setShowCompressModal] = useState(false);
  const [showExtractModal, setShowExtractModal] = useState<FileEntry | null>(null);
  const [showChmodModal, setShowChmodModal] = useState<FileEntry | null>(null);

  // Form states
  const [newName, setNewName] = useState('');
  const [editContent, setEditContent] = useState('');
  const [archiveName, setArchiveName] = useState('');
  const [extractPath, setExtractPath] = useState('');
  const [permissions, setPermissions] = useState('644');
  const [formError, setFormError] = useState<string | null>(null);
  const [isSaving, setIsSaving] = useState(false);

  // Clipboard for copy/cut
  const [clipboard, setClipboard] = useState<{ files: string[]; action: 'copy' | 'cut' } | null>(null);

  // Disk usage
  const [diskUsage, setDiskUsage] = useState<{ used: number; total: number; free: number } | null>(null);

  useEffect(() => {
    loadUsers();
  }, []);

  useEffect(() => {
    loadFiles();
    loadDiskUsage();
  }, [currentPath, selectedUser]);

  const loadUsers = async () => {
    try {
      const usersData = await userService.list();
      setUsers(usersData);
    } catch (err: any) {
      console.error('Failed to load users:', err);
    }
  };

  const loadFiles = async () => {
    try {
      setIsLoading(true);
      setError(null);
      const filesData = await fileService.list(currentPath, selectedUser || undefined);
      setFiles(filesData);
      setSelectedFiles(new Set());
    } catch (err: any) {
      setError(err.message || 'Failed to load files');
    } finally {
      setIsLoading(false);
    }
  };

  const loadDiskUsage = async () => {
    try {
      const usage = await fileService.getDiskUsage(currentPath, selectedUser || undefined);
      setDiskUsage(usage);
    } catch (err: any) {
      console.error('Failed to load disk usage:', err);
    }
  };

  const navigateTo = (path: string) => {
    setCurrentPath(path);
  };

  const navigateUp = () => {
    const parts = currentPath.split('/').filter(p => p);
    parts.pop();
    setCurrentPath('/' + parts.join('/'));
  };

  const handleFileClick = (file: FileEntry) => {
    if (file.type === 'directory') {
      navigateTo(file.path);
    }
  };

  const toggleFileSelection = (path: string) => {
    setSelectedFiles(prev => {
      const next = new Set(prev);
      if (next.has(path)) {
        next.delete(path);
      } else {
        next.add(path);
      }
      return next;
    });
  };

  const handleCreate = async () => {
    if (!newName) {
      setFormError('Please enter a name');
      return;
    }

    try {
      setIsSaving(true);
      setFormError(null);
      const fullPath = currentPath === '/' ? `/${newName}` : `${currentPath}/${newName}`;
      await fileService.create(fullPath, showCreateModal!, selectedUser || undefined);
      await loadFiles();
      setShowCreateModal(null);
      setNewName('');
    } catch (err: any) {
      setFormError(err.message || 'Failed to create');
    } finally {
      setIsSaving(false);
    }
  };

  const handleRename = async () => {
    if (!showRenameModal || !newName) return;

    try {
      setIsSaving(true);
      setFormError(null);
      const parts = showRenameModal.path.split('/');
      parts.pop();
      const newPath = [...parts, newName].join('/');
      await fileService.rename(showRenameModal.path, newPath, selectedUser || undefined);
      await loadFiles();
      setShowRenameModal(null);
      setNewName('');
    } catch (err: any) {
      setFormError(err.message || 'Failed to rename');
    } finally {
      setIsSaving(false);
    }
  };

  const handleEdit = async () => {
    if (!showEditModal) return;

    try {
      setIsSaving(true);
      setFormError(null);
      await fileService.write(showEditModal.path, editContent, selectedUser || undefined);
      await loadFiles();
      setShowEditModal(null);
      setEditContent('');
    } catch (err: any) {
      setFormError(err.message || 'Failed to save file');
    } finally {
      setIsSaving(false);
    }
  };

  const handleDelete = async () => {
    if (!showDeleteConfirm) return;

    try {
      setIsSaving(true);
      await fileService.delete(showDeleteConfirm, selectedUser || undefined);
      await loadFiles();
      setShowDeleteConfirm(null);
    } catch (err: any) {
      setError(err.message || 'Failed to delete');
    } finally {
      setIsSaving(false);
    }
  };

  const handleCopy = () => {
    setClipboard({ files: Array.from(selectedFiles), action: 'copy' });
  };

  const handleCut = () => {
    setClipboard({ files: Array.from(selectedFiles), action: 'cut' });
  };

  const handlePaste = async () => {
    if (!clipboard || clipboard.files.length === 0) return;

    try {
      setIsSaving(true);
      for (const sourcePath of clipboard.files) {
        const fileName = sourcePath.split('/').pop();
        const destPath = currentPath === '/' ? `/${fileName}` : `${currentPath}/${fileName}`;
        
        if (clipboard.action === 'copy') {
          await fileService.copy(sourcePath, destPath, selectedUser || undefined);
        } else {
          await fileService.rename(sourcePath, destPath, selectedUser || undefined);
        }
      }
      await loadFiles();
      if (clipboard.action === 'cut') {
        setClipboard(null);
      }
    } catch (err: any) {
      setError(err.message || 'Failed to paste');
    } finally {
      setIsSaving(false);
    }
  };

  const handleCompress = async () => {
    if (selectedFiles.size === 0 || !archiveName) return;

    try {
      setIsSaving(true);
      setFormError(null);
      const archivePath = currentPath === '/' ? `/${archiveName}` : `${currentPath}/${archiveName}`;
      await fileService.compress(Array.from(selectedFiles), archivePath, selectedUser || undefined);
      await loadFiles();
      setShowCompressModal(false);
      setArchiveName('');
      setSelectedFiles(new Set());
    } catch (err: any) {
      setFormError(err.message || 'Failed to compress');
    } finally {
      setIsSaving(false);
    }
  };

  const handleExtract = async () => {
    if (!showExtractModal) return;

    try {
      setIsSaving(true);
      setFormError(null);
      const dest = extractPath || currentPath;
      await fileService.extract(showExtractModal.path, dest, selectedUser || undefined);
      await loadFiles();
      setShowExtractModal(null);
      setExtractPath('');
    } catch (err: any) {
      setFormError(err.message || 'Failed to extract');
    } finally {
      setIsSaving(false);
    }
  };

  const handleChmod = async () => {
    if (!showChmodModal) return;

    try {
      setIsSaving(true);
      setFormError(null);
      await fileService.chmod(showChmodModal.path, permissions, selectedUser || undefined);
      await loadFiles();
      setShowChmodModal(null);
      setPermissions('644');
    } catch (err: any) {
      setFormError(err.message || 'Failed to change permissions');
    } finally {
      setIsSaving(false);
    }
  };

  const openEditFile = async (file: FileEntry) => {
    try {
      setIsLoading(true);
      const content = await fileService.read(file.path, selectedUser || undefined);
      setEditContent(content.content);
      setShowEditModal(file);
    } catch (err: any) {
      setError(err.message || 'Failed to read file');
    } finally {
      setIsLoading(false);
    }
  };

  const formatSize = (bytes: number) => {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  const getFileIcon = (file: FileEntry) => {
    if (file.type === 'directory') {
      return <Folder size={18} className="text-[#E7F0FA]" />;
    }
    return <File size={18} className="text-[var(--color-text-muted)]" />;
  };

  const isArchive = (name: string) => {
    return /\.(zip|tar|tar\.gz|tgz|tar\.bz2|rar|7z)$/i.test(name);
  };

  const filteredFiles = files.filter(file =>
    file.name.toLowerCase().includes(searchQuery.toLowerCase())
  );

  const breadcrumbs = currentPath.split('/').filter(p => p);

  if (isLoading && files.length === 0) {
    return (
      <div className="space-y-6">
        <div className="h-8 bg-[var(--color-primary)]/50 rounded animate-pulse w-48" />
        <div className="h-96 bg-[var(--color-primary)]/50 rounded animate-pulse" />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-[var(--color-text-primary)]">File Manager</h1>
          <p className="text-[var(--color-text-secondary)] mt-1">
            Browse and manage files across all user accounts
          </p>
        </div>
        <div className="flex gap-2">
          <Select
            value={selectedUser}
            onChange={(e) => {
              setSelectedUser(e.target.value);
              setCurrentPath('/');
            }}
            options={[
              { value: '', label: 'System Files' },
              ...users.map(u => ({ value: u.id, label: `${u.username} (${u.home_directory})` })),
            ]}
            className="w-64"
          />
        </div>
      </div>

      {/* Error Alert */}
      {error && (
        <div className="flex items-center gap-2 p-4 rounded-lg bg-[var(--color-error)]/20 text-[var(--color-error)]">
          <AlertCircle size={20} />
          <span>{error}</span>
          <button onClick={() => setError(null)} className="ml-auto">×</button>
        </div>
      )}

      {/* Disk Usage */}
      {diskUsage && (
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-4">
                <HardDrive size={24} className="text-[#E7F0FA]" />
                <div>
                  <p className="text-sm text-[var(--color-text-secondary)]">Disk Usage</p>
                  <p className="text-lg font-semibold text-[var(--color-text-primary)]">
                    {formatSize(diskUsage.used)} / {formatSize(diskUsage.total)}
                  </p>
                </div>
              </div>
              <div className="w-48">
                <div className="h-2 bg-[var(--color-primary-dark)] rounded-full overflow-hidden">
                  <div
                    className="h-full bg-gradient-to-r from-[#7BA4D0] to-[#E7F0FA] rounded-full"
                    style={{ width: `${(diskUsage.used / diskUsage.total) * 100}%` }}
                  />
                </div>
                <p className="text-xs text-[var(--color-text-muted)] mt-1 text-right">
                  {formatSize(diskUsage.free)} free
                </p>
              </div>
            </div>
          </CardContent>
        </Card>
      )}

      {/* File Browser */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between flex-wrap gap-4">
            {/* Breadcrumb */}
            <div className="flex items-center gap-1 text-sm">
              <button
                onClick={() => setCurrentPath('/')}
                className="flex items-center gap-1 text-[var(--color-text-muted)] hover:text-[var(--color-text-primary)]"
              >
                <Home size={16} />
              </button>
              {breadcrumbs.map((crumb, idx) => (
                <span key={idx} className="flex items-center gap-1">
                  <ChevronRight size={14} className="text-[var(--color-text-muted)]" />
                  <button
                    onClick={() => navigateTo('/' + breadcrumbs.slice(0, idx + 1).join('/'))}
                    className="text-[var(--color-text-muted)] hover:text-[var(--color-text-primary)]"
                  >
                    {crumb}
                  </button>
                </span>
              ))}
            </div>

            {/* Actions */}
            <div className="flex gap-2 flex-wrap">
              <Button
                size="sm"
                variant="ghost"
                onClick={loadFiles}
                title="Refresh"
              >
                <RefreshCw size={16} />
              </Button>
              <Button
                size="sm"
                variant="outline"
                leftIcon={<FolderPlus size={16} />}
                onClick={() => setShowCreateModal('directory')}
              >
                New Folder
              </Button>
              <Button
                size="sm"
                variant="outline"
                leftIcon={<FilePlus size={16} />}
                onClick={() => setShowCreateModal('file')}
              >
                New File
              </Button>
              {selectedFiles.size > 0 && (
                <>
                  <Button size="sm" variant="ghost" onClick={handleCopy} title="Copy">
                    <Copy size={16} />
                  </Button>
                  <Button size="sm" variant="ghost" onClick={handleCut} title="Cut">
                    <Scissors size={16} />
                  </Button>
                  <Button
                    size="sm"
                    variant="ghost"
                    onClick={() => setShowCompressModal(true)}
                    title="Compress"
                  >
                    <Archive size={16} />
                  </Button>
                </>
              )}
              {clipboard && (
                <Button size="sm" variant="outline" onClick={handlePaste}>
                  Paste ({clipboard.files.length})
                </Button>
              )}
            </div>
          </div>
        </CardHeader>
        <CardContent>
          {/* Search */}
          <div className="mb-4">
            <div className="relative">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 text-[var(--color-text-muted)]" size={18} />
              <Input
                placeholder="Search files..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="pl-10"
              />
            </div>
          </div>

          {/* File List */}
          {filteredFiles.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-12 text-[var(--color-text-muted)]">
              <Folder size={48} className="mb-4 opacity-50" />
              <p>No files found</p>
            </div>
          ) : (
            <div className="space-y-1">
              {/* Go up */}
              {currentPath !== '/' && (
                <div
                  className="flex items-center gap-3 p-3 rounded-lg hover:bg-[var(--color-primary-dark)] cursor-pointer"
                  onClick={navigateUp}
                >
                  <Folder size={18} className="text-[var(--color-text-muted)]" />
                  <span className="text-[var(--color-text-secondary)]">..</span>
                </div>
              )}

              {/* Files */}
              {filteredFiles.map(file => (
                <div
                  key={file.path}
                  className={`flex items-center gap-3 p-3 rounded-lg hover:bg-[var(--color-primary-dark)] cursor-pointer ${
                    selectedFiles.has(file.path) ? 'bg-[var(--color-accent)]/20' : ''
                  }`}
                >
                  <input
                    type="checkbox"
                    checked={selectedFiles.has(file.path)}
                    onChange={() => toggleFileSelection(file.path)}
                    onClick={(e) => e.stopPropagation()}
                    className="rounded border-[var(--color-border)]"
                  />
                  <div
                    className="flex items-center gap-3 flex-1 min-w-0"
                    onClick={() => handleFileClick(file)}
                  >
                    {getFileIcon(file)}
                    <span className="text-[var(--color-text-primary)] truncate flex-1">
                      {file.name}
                    </span>
                  </div>
                  <div className="flex items-center gap-4 text-sm text-[var(--color-text-muted)]">
                    <span className="font-mono text-xs">{file.permissions}</span>
                    <span className="w-20 text-right">{formatSize(file.size)}</span>
                    <span className="w-32 text-right">
                      {new Date(file.modified_at).toLocaleDateString()}
                    </span>
                    <div className="flex gap-1">
                      {file.type === 'file' && (
                        <Button
                          size="sm"
                          variant="ghost"
                          onClick={(e) => {
                            e.stopPropagation();
                            openEditFile(file);
                          }}
                          title="Edit"
                        >
                          <Edit2 size={14} />
                        </Button>
                      )}
                      {isArchive(file.name) && (
                        <Button
                          size="sm"
                          variant="ghost"
                          onClick={(e) => {
                            e.stopPropagation();
                            setShowExtractModal(file);
                          }}
                          title="Extract"
                        >
                          <Archive size={14} />
                        </Button>
                      )}
                      <Button
                        size="sm"
                        variant="ghost"
                        onClick={(e) => {
                          e.stopPropagation();
                          setNewName(file.name);
                          setShowRenameModal(file);
                        }}
                        title="Rename"
                      >
                        <Edit2 size={14} />
                      </Button>
                      <Button
                        size="sm"
                        variant="ghost"
                        onClick={(e) => {
                          e.stopPropagation();
                          setPermissions(file.permissions);
                          setShowChmodModal(file);
                        }}
                        title="Permissions"
                      >
                        🔒
                      </Button>
                      <Button
                        size="sm"
                        variant="ghost"
                        className="text-[var(--color-error)]"
                        onClick={(e) => {
                          e.stopPropagation();
                          setShowDeleteConfirm(file.path);
                        }}
                      >
                        <Trash2 size={14} />
                      </Button>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      {/* Create Modal */}
      <Modal
        isOpen={showCreateModal !== null}
        onClose={() => {
          setShowCreateModal(null);
          setNewName('');
          setFormError(null);
        }}
        title={`Create ${showCreateModal === 'directory' ? 'Folder' : 'File'}`}
      >
        <div className="space-y-4">
          {formError && (
            <div className="p-3 rounded bg-[var(--color-error)]/20 text-[var(--color-error)] text-sm">
              {formError}
            </div>
          )}
          
          <div>
            <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-2">
              Name
            </label>
            <Input
              value={newName}
              onChange={(e) => setNewName(e.target.value)}
              placeholder={showCreateModal === 'directory' ? 'folder-name' : 'filename.txt'}
              autoFocus
            />
          </div>

          <div className="flex gap-3 justify-end">
            <Button variant="ghost" onClick={() => setShowCreateModal(null)}>
              Cancel
            </Button>
            <Button onClick={handleCreate} isLoading={isSaving}>
              Create
            </Button>
          </div>
        </div>
      </Modal>

      {/* Rename Modal */}
      <Modal
        isOpen={showRenameModal !== null}
        onClose={() => {
          setShowRenameModal(null);
          setNewName('');
          setFormError(null);
        }}
        title="Rename"
      >
        <div className="space-y-4">
          {formError && (
            <div className="p-3 rounded bg-[var(--color-error)]/20 text-[var(--color-error)] text-sm">
              {formError}
            </div>
          )}
          
          <div>
            <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-2">
              New Name
            </label>
            <Input
              value={newName}
              onChange={(e) => setNewName(e.target.value)}
              autoFocus
            />
          </div>

          <div className="flex gap-3 justify-end">
            <Button variant="ghost" onClick={() => setShowRenameModal(null)}>
              Cancel
            </Button>
            <Button onClick={handleRename} isLoading={isSaving}>
              Rename
            </Button>
          </div>
        </div>
      </Modal>

      {/* Edit File Modal */}
      <Modal
        isOpen={showEditModal !== null}
        onClose={() => {
          setShowEditModal(null);
          setEditContent('');
          setFormError(null);
        }}
        title={`Edit: ${showEditModal?.name}`}
      >
        <div className="space-y-4">
          {formError && (
            <div className="p-3 rounded bg-[var(--color-error)]/20 text-[var(--color-error)] text-sm">
              {formError}
            </div>
          )}
          
          <textarea
            value={editContent}
            onChange={(e) => setEditContent(e.target.value)}
            className="w-full h-96 px-4 py-2.5 rounded-lg bg-[var(--color-surface)] border border-[var(--color-border)] text-[var(--color-text-primary)] font-mono text-sm resize-none"
          />

          <div className="flex gap-3 justify-end">
            <Button variant="ghost" onClick={() => setShowEditModal(null)}>
              Cancel
            </Button>
            <Button onClick={handleEdit} isLoading={isSaving}>
              Save
            </Button>
          </div>
        </div>
      </Modal>

      {/* Compress Modal */}
      <Modal
        isOpen={showCompressModal}
        onClose={() => {
          setShowCompressModal(false);
          setArchiveName('');
          setFormError(null);
        }}
        title="Compress Files"
      >
        <div className="space-y-4">
          {formError && (
            <div className="p-3 rounded bg-[var(--color-error)]/20 text-[var(--color-error)] text-sm">
              {formError}
            </div>
          )}
          
          <p className="text-sm text-[var(--color-text-secondary)]">
            Compressing {selectedFiles.size} file(s)
          </p>

          <div>
            <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-2">
              Archive Name
            </label>
            <Input
              value={archiveName}
              onChange={(e) => setArchiveName(e.target.value)}
              placeholder="archive.zip"
            />
          </div>

          <div className="flex gap-3 justify-end">
            <Button variant="ghost" onClick={() => setShowCompressModal(false)}>
              Cancel
            </Button>
            <Button onClick={handleCompress} isLoading={isSaving}>
              Compress
            </Button>
          </div>
        </div>
      </Modal>

      {/* Extract Modal */}
      <Modal
        isOpen={showExtractModal !== null}
        onClose={() => {
          setShowExtractModal(null);
          setExtractPath('');
          setFormError(null);
        }}
        title="Extract Archive"
      >
        <div className="space-y-4">
          {formError && (
            <div className="p-3 rounded bg-[var(--color-error)]/20 text-[var(--color-error)] text-sm">
              {formError}
            </div>
          )}
          
          <p className="text-sm text-[var(--color-text-secondary)]">
            Extracting: {showExtractModal?.name}
          </p>

          <div>
            <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-2">
              Destination (leave empty for current directory)
            </label>
            <Input
              value={extractPath}
              onChange={(e) => setExtractPath(e.target.value)}
              placeholder={currentPath}
            />
          </div>

          <div className="flex gap-3 justify-end">
            <Button variant="ghost" onClick={() => setShowExtractModal(null)}>
              Cancel
            </Button>
            <Button onClick={handleExtract} isLoading={isSaving}>
              Extract
            </Button>
          </div>
        </div>
      </Modal>

      {/* Chmod Modal */}
      <Modal
        isOpen={showChmodModal !== null}
        onClose={() => {
          setShowChmodModal(null);
          setPermissions('644');
          setFormError(null);
        }}
        title="Change Permissions"
      >
        <div className="space-y-4">
          {formError && (
            <div className="p-3 rounded bg-[var(--color-error)]/20 text-[var(--color-error)] text-sm">
              {formError}
            </div>
          )}
          
          <p className="text-sm text-[var(--color-text-secondary)]">
            File: {showChmodModal?.name}
          </p>

          <div>
            <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-2">
              Permissions (octal)
            </label>
            <Input
              value={permissions}
              onChange={(e) => setPermissions(e.target.value)}
              placeholder="644"
              maxLength={4}
            />
            <p className="text-xs text-[var(--color-text-muted)] mt-1">
              Common: 644 (files), 755 (directories/scripts), 600 (private)
            </p>
          </div>

          <div className="flex gap-3 justify-end">
            <Button variant="ghost" onClick={() => setShowChmodModal(null)}>
              Cancel
            </Button>
            <Button onClick={handleChmod} isLoading={isSaving}>
              Apply
            </Button>
          </div>
        </div>
      </Modal>

      {/* Delete Confirmation Modal */}
      <Modal
        isOpen={showDeleteConfirm !== null}
        onClose={() => setShowDeleteConfirm(null)}
        title="Delete"
      >
        <div className="space-y-4">
          <p className="text-[var(--color-text-secondary)]">
            Are you sure you want to delete this? This action cannot be undone.
          </p>
          <code className="block p-2 rounded bg-[var(--color-primary-dark)] text-sm text-[var(--color-text-muted)]">
            {showDeleteConfirm}
          </code>
          <div className="flex gap-3 justify-end">
            <Button variant="ghost" onClick={() => setShowDeleteConfirm(null)}>
              Cancel
            </Button>
            <Button variant="danger" onClick={handleDelete} isLoading={isSaving}>
              Delete
            </Button>
          </div>
        </div>
      </Modal>
    </div>
  );
}
