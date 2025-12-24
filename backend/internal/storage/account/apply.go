// Package account provides filesystem-based account state management
package account

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"
)

// Applier handles idempotent state application
type Applier struct {
	state *StateManager
}

// NewApplier creates a new applier
func NewApplier() *Applier {
	return &Applier{
		state: NewStateManager(),
	}
}

// NewApplierWithState creates an applier with a custom state manager
func NewApplierWithState(state *StateManager) *Applier {
	return &Applier{state: state}
}

// Apply idempotently applies the desired state to the filesystem
// This is the core function that ensures filesystem state matches desired state
func (a *Applier) Apply(accountID int, config *ApplyConfig) error {
	// Step 1: Validate configuration
	if err := ValidateApplyConfig(config); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Step 2: Ensure directory structure exists
	if err := a.state.CreateAccountStructure(accountID); err != nil {
		return fmt.Errorf("failed to create account structure: %w", err)
	}

	// Step 3: Ensure system user exists (idempotent)
	if config.Identity != nil {
		if err := a.ensureSystemUser(config.Identity); err != nil {
			return fmt.Errorf("failed to ensure system user: %w", err)
		}
	}

	// Step 4: Write state files (atomic writes)
	if config.Identity != nil {
		if err := a.state.WriteIdentity(accountID, config.Identity); err != nil {
			return fmt.Errorf("failed to write identity: %w", err)
		}
	}

	if config.Limits != nil {
		if err := a.state.WriteLimits(accountID, config.Limits); err != nil {
			return fmt.Errorf("failed to write limits: %w", err)
		}
	}

	if config.Status != nil {
		if err := a.state.WriteStatus(accountID, config.Status); err != nil {
			return fmt.Errorf("failed to write status: %w", err)
		}
	}

	if config.Metadata != nil {
		config.Metadata.UpdatedAt = time.Now().Format(time.RFC3339)
		if err := a.state.WriteMetadata(accountID, config.Metadata); err != nil {
			return fmt.Errorf("failed to write metadata: %w", err)
		}
	}

	// Step 5: Apply ownership to account directory
	if config.Identity != nil {
		if err := a.applyOwnership(accountID, config.Identity.UID, config.Identity.GID); err != nil {
			return fmt.Errorf("failed to apply ownership: %w", err)
		}
	}

	// Step 6: Apply resource limits (cgroups, quotas)
	if config.Limits != nil && config.Identity != nil {
		if err := a.applyResourceLimits(accountID, config.Identity.UID, config.Limits); err != nil {
			// Log but don't fail - resource limits are best effort
			fmt.Printf("warning: failed to apply resource limits for account %d: %v\n", accountID, err)
		}
	}

	return nil
}

// ensureSystemUser creates the system user if it doesn't exist (idempotent)
func (a *Applier) ensureSystemUser(identity *AccountIdentity) error {
	// Check if user already exists
	_, err := exec.Command("id", strconv.Itoa(identity.UID)).Output()
	if err == nil {
		return nil // User exists
	}

	// Create group first
	groupArgs := []string{
		"--gid", strconv.Itoa(identity.GID),
		identity.Name,
	}
	if err := exec.Command("groupadd", groupArgs...).Run(); err != nil {
		// Group might already exist, continue
	}

	// Create user
	userArgs := []string{
		"--uid", strconv.Itoa(identity.UID),
		"--gid", strconv.Itoa(identity.GID),
		"--home-dir", filepath.Join(a.state.AccountPath(identity.ID), "home"),
		"--shell", "/bin/bash",
		"--no-create-home", // We create the structure ourselves
		identity.Name,
	}

	if err := exec.Command("useradd", userArgs...).Run(); err != nil {
		// Check if user already exists with different UID
		// This is okay if the user exists
		return nil
	}

	return nil
}

// applyOwnership recursively sets ownership on account directories
func (a *Applier) applyOwnership(accountID, uid, gid int) error {
	basePath := a.state.AccountPath(accountID)

	// Directories that should be owned by the account user
	userOwnedDirs := []string{
		filepath.Join(basePath, "home"),
		filepath.Join(basePath, "web"),
		filepath.Join(basePath, "mail"),
		filepath.Join(basePath, "tmp"),
	}

	for _, dir := range userOwnedDirs {
		if err := chownRecursive(dir, uid, gid); err != nil {
			return fmt.Errorf("failed to chown %s: %w", dir, err)
		}
	}

	// Directories that should be owned by root but readable by user
	rootOwnedDirs := []string{
		filepath.Join(basePath, "logs"),
		filepath.Join(basePath, "backups"),
	}

	for _, dir := range rootOwnedDirs {
		// Set group ownership so user can read
		if err := os.Chown(dir, 0, gid); err != nil {
			return fmt.Errorf("failed to chown %s: %w", dir, err)
		}
		if err := os.Chmod(dir, 0750); err != nil {
			return fmt.Errorf("failed to chmod %s: %w", dir, err)
		}
	}

	return nil
}

// chownRecursive recursively changes ownership of a directory
func chownRecursive(path string, uid, gid int) error {
	return filepath.Walk(path, func(name string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		return os.Chown(name, uid, gid)
	})
}

// applyResourceLimits applies cgroup and quota limits
func (a *Applier) applyResourceLimits(accountID, uid int, limits *ResourceLimits) error {
	// Apply disk quota using setquota (if quotas are enabled)
	if limits.DiskMB > 0 {
		if err := a.setDiskQuota(uid, limits.DiskMB, limits.Inodes); err != nil {
			return err
		}
	}

	// Apply cgroup limits for CPU and memory
	if err := a.applyCgroupLimits(accountID, limits); err != nil {
		return err
	}

	return nil
}

// setDiskQuota sets disk quota for a user
func (a *Applier) setDiskQuota(uid, diskMB, inodes int) error {
	// Convert MB to KB for setquota
	softLimit := diskMB * 1024
	hardLimit := int(float64(diskMB) * 1.1 * 1024) // 10% grace

	inodeSoft := inodes
	inodeHard := int(float64(inodes) * 1.1)

	// setquota -u uid soft hard isoft ihard filesystem
	args := []string{
		"-u", strconv.Itoa(uid),
		strconv.Itoa(softLimit), strconv.Itoa(hardLimit),
		strconv.Itoa(inodeSoft), strconv.Itoa(inodeHard),
		"/srv", // Assuming /srv is the quota-enabled filesystem
	}

	cmd := exec.Command("setquota", args...)
	if err := cmd.Run(); err != nil {
		// Quotas might not be enabled, this is not fatal
		return nil
	}

	return nil
}

// applyCgroupLimits applies cgroup v2 limits
func (a *Applier) applyCgroupLimits(accountID int, limits *ResourceLimits) error {
	cgroupPath := fmt.Sprintf("/sys/fs/cgroup/owehost/account-%d", accountID)

	// Create cgroup directory
	if err := os.MkdirAll(cgroupPath, 0755); err != nil {
		return fmt.Errorf("failed to create cgroup: %w", err)
	}

	// Set CPU limit (cpu.max)
	if limits.CPUPercent > 0 {
		// cpu.max format: "quota period" in microseconds
		// 100% = 100000 / 100000
		quota := limits.CPUPercent * 1000
		cpuMax := fmt.Sprintf("%d 100000", quota)
		if err := os.WriteFile(filepath.Join(cgroupPath, "cpu.max"), []byte(cpuMax), 0644); err != nil {
			// cgroup v2 might not be available
		}
	}

	// Set memory limit (memory.max)
	if limits.RAMMB > 0 {
		memMax := fmt.Sprintf("%d", limits.RAMMB*1024*1024) // Convert to bytes
		if err := os.WriteFile(filepath.Join(cgroupPath, "memory.max"), []byte(memMax), 0644); err != nil {
			// cgroup v2 might not be available
		}
	}

	return nil
}

// Suspend suspends an account
func (a *Applier) Suspend(accountID int, reason, actor string) error {
	now := time.Now().Format(time.RFC3339)
	status := &AccountStatus{
		Suspended:   true,
		Locked:      false,
		Reason:      &reason,
		SuspendedAt: &now,
		SuspendedBy: &actor,
	}

	if err := a.state.WriteStatus(accountID, status); err != nil {
		return err
	}

	// Update identity state
	identity, err := a.state.ReadIdentity(accountID)
	if err != nil {
		return err
	}

	identity.State = StateSuspended
	return a.state.WriteIdentity(accountID, identity)
}

// Unsuspend unsuspends an account
func (a *Applier) Unsuspend(accountID int) error {
	status := &AccountStatus{
		Suspended: false,
		Locked:    false,
	}

	if err := a.state.WriteStatus(accountID, status); err != nil {
		return err
	}

	// Update identity state
	identity, err := a.state.ReadIdentity(accountID)
	if err != nil {
		return err
	}

	identity.State = StateActive
	return a.state.WriteIdentity(accountID, identity)
}

// Terminate terminates an account (marks for deletion)
func (a *Applier) Terminate(accountID int, reason, actor string) error {
	now := time.Now().Format(time.RFC3339)
	status := &AccountStatus{
		Suspended:   true,
		Locked:      true,
		Reason:      &reason,
		SuspendedAt: &now,
		SuspendedBy: &actor,
	}

	if err := a.state.WriteStatus(accountID, status); err != nil {
		return err
	}

	// Update identity state
	identity, err := a.state.ReadIdentity(accountID)
	if err != nil {
		return err
	}

	identity.State = StateTerminated
	return a.state.WriteIdentity(accountID, identity)
}

// Delete completely removes an account from the filesystem
func (a *Applier) Delete(accountID int) error {
	// Read identity to get username for system user deletion
	identity, err := a.state.ReadIdentity(accountID)
	if err == nil {
		// Remove system user
		exec.Command("userdel", identity.Name).Run()
		exec.Command("groupdel", identity.Name).Run()

		// Remove cgroup
		cgroupPath := fmt.Sprintf("/sys/fs/cgroup/owehost/account-%d", accountID)
		os.RemoveAll(cgroupPath)
	}

	// Remove account directory
	return a.state.DeleteAccountStructure(accountID)
}
