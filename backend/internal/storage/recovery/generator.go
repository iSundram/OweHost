// Package recovery provides filesystem scanning and database rebuild functionality
package recovery

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/iSundram/OweHost/internal/storage/account"
	"github.com/iSundram/OweHost/internal/storage/web"
)

// Generator regenerates service configurations from filesystem state
type Generator struct {
	scanner        *Scanner
	webApplier     *web.Applier
	nginxTemplate  *template.Template
	phpFpmTemplate *template.Template
}

// NewGenerator creates a new configuration generator
func NewGenerator() *Generator {
	return &Generator{
		scanner:    NewScanner(),
		webApplier: web.NewApplier(),
	}
}

// GenerateResult represents the result of configuration generation
type GenerateResult struct {
	NginxConfigs   int      `json:"nginx_configs"`
	PHPFpmPools    int      `json:"phpfpm_pools"`
	SystemUsers    int      `json:"system_users"`
	Cgroups        int      `json:"cgroups"`
	Errors         []string `json:"errors,omitempty"`
	ConfigsPaths   []string `json:"config_paths,omitempty"`
}

// GenerateOptions configures generation
type GenerateOptions struct {
	DryRun      bool // Preview without writing
	AccountID   *int // Generate for specific account only
	SkipNginx   bool
	SkipPHPFpm  bool
	SkipUsers   bool
	SkipCgroups bool
	Verbose     bool
}

// GenerateAll regenerates all service configurations
func (g *Generator) GenerateAll(opts GenerateOptions) (*GenerateResult, error) {
	result := &GenerateResult{
		Errors:       []string{},
		ConfigsPaths: []string{},
	}

	// Scan filesystem
	var accounts []AccountScan

	if opts.AccountID != nil {
		scan, err := g.scanner.ScanAccount(*opts.AccountID)
		if err != nil {
			return nil, err
		}
		accounts = []AccountScan{*scan}
	} else {
		scanResult, err := g.scanner.ScanAll()
		if err != nil {
			return nil, err
		}
		accounts = scanResult.Accounts
	}

	// Process each account
	for _, scan := range accounts {
		if err := g.generateForAccount(&scan, opts, result); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("account %d: %v", scan.ID, err))
		}
	}

	// Reload services if not dry run
	if !opts.DryRun {
		if !opts.SkipNginx && result.NginxConfigs > 0 {
			if err := g.reloadNginx(); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("nginx reload: %v", err))
			}
		}

		if !opts.SkipPHPFpm && result.PHPFpmPools > 0 {
			if err := g.reloadPHPFpm(); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("php-fpm reload: %v", err))
			}
		}
	}

	return result, nil
}

// generateForAccount generates configs for a single account
func (g *Generator) generateForAccount(scan *AccountScan, opts GenerateOptions, result *GenerateResult) error {
	// Generate system user
	if !opts.SkipUsers && scan.Identity != nil {
		if err := g.ensureSystemUser(scan.Identity, opts.DryRun); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("user %s: %v", scan.Identity.Name, err))
		} else {
			result.SystemUsers++
		}
	}

	// Generate cgroup
	if !opts.SkipCgroups && scan.Limits != nil {
		if err := g.ensureCgroup(scan.ID, scan.Limits, opts.DryRun); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("cgroup %d: %v", scan.ID, err))
		} else {
			result.Cgroups++
		}
	}

	// Generate nginx configs for each site
	for _, site := range scan.Sites {
		if !opts.SkipNginx {
			if opts.DryRun {
				result.NginxConfigs++
			} else {
				if err := g.webApplier.GenerateNginxConfig(scan.ID, &site); err != nil {
					result.Errors = append(result.Errors, fmt.Sprintf("nginx %s: %v", site.Domain, err))
				} else {
					result.NginxConfigs++
					result.ConfigsPaths = append(result.ConfigsPaths,
						fmt.Sprintf("/etc/nginx/sites-available/a-%d-%s.conf", scan.ID, site.Domain))
				}
			}
		}

		// Generate PHP-FPM pool if PHP runtime
		if !opts.SkipPHPFpm && strings.HasPrefix(site.Runtime, "php-") {
			if opts.DryRun {
				result.PHPFpmPools++
			} else {
				if err := g.webApplier.GeneratePHPFpmPool(scan.ID, &site); err != nil {
					result.Errors = append(result.Errors, fmt.Sprintf("phpfpm %s: %v", site.Domain, err))
				} else {
					result.PHPFpmPools++
				}
			}
		}
	}

	return nil
}

// ensureSystemUser ensures the system user exists
func (g *Generator) ensureSystemUser(identity *account.AccountIdentity, dryRun bool) error {
	if dryRun {
		return nil
	}

	// Check if user exists
	if _, err := exec.Command("id", identity.Name).Output(); err == nil {
		return nil // User exists
	}

	// Create group
	exec.Command("groupadd",
		"--gid", fmt.Sprintf("%d", identity.GID),
		identity.Name,
	).Run()

	// Create user
	homePath := filepath.Join(account.BaseAccountPath, fmt.Sprintf("%s%d", account.AccountPrefix, identity.ID), "home")
	return exec.Command("useradd",
		"--uid", fmt.Sprintf("%d", identity.UID),
		"--gid", fmt.Sprintf("%d", identity.GID),
		"--home-dir", homePath,
		"--shell", "/bin/bash",
		"--no-create-home",
		identity.Name,
	).Run()
}

// ensureCgroup ensures cgroup limits are set
func (g *Generator) ensureCgroup(accountID int, limits *account.ResourceLimits, dryRun bool) error {
	if dryRun {
		return nil
	}

	cgroupPath := fmt.Sprintf("/sys/fs/cgroup/owehost/account-%d", accountID)

	// Create cgroup directory
	if err := os.MkdirAll(cgroupPath, 0755); err != nil {
		return err
	}

	// Set CPU limit
	if limits.CPUPercent > 0 {
		quota := limits.CPUPercent * 1000 // Convert to microseconds
		cpuMax := fmt.Sprintf("%d 100000", quota)
		os.WriteFile(filepath.Join(cgroupPath, "cpu.max"), []byte(cpuMax), 0644)
	}

	// Set memory limit
	if limits.RAMMB > 0 {
		memMax := fmt.Sprintf("%d", limits.RAMMB*1024*1024)
		os.WriteFile(filepath.Join(cgroupPath, "memory.max"), []byte(memMax), 0644)
	}

	return nil
}

// reloadNginx reloads nginx configuration
func (g *Generator) reloadNginx() error {
	// Test config first
	if err := exec.Command("nginx", "-t").Run(); err != nil {
		return fmt.Errorf("config test failed: %w", err)
	}
	return exec.Command("nginx", "-s", "reload").Run()
}

// reloadPHPFpm reloads all PHP-FPM services
func (g *Generator) reloadPHPFpm() error {
	// Find and reload all PHP-FPM versions
	versions := []string{"7.4", "8.0", "8.1", "8.2", "8.3"}
	for _, version := range versions {
		service := fmt.Sprintf("php%s-fpm", version)
		// Check if service exists before reloading
		if exec.Command("systemctl", "is-active", service).Run() == nil {
			exec.Command("systemctl", "reload", service).Run()
		}
	}
	return nil
}

// CleanupStaleConfigs removes nginx configs for non-existent sites
func (g *Generator) CleanupStaleConfigs() ([]string, error) {
	var removed []string

	// Scan filesystem for current sites
	scanResult, err := g.scanner.ScanAll()
	if err != nil {
		return nil, err
	}

	// Build set of valid config names
	validConfigs := make(map[string]bool)
	for _, account := range scanResult.Accounts {
		for _, site := range account.Sites {
			configName := fmt.Sprintf("a-%d-%s.conf", account.ID, site.Domain)
			validConfigs[configName] = true
		}
	}

	// Scan nginx sites-available
	entries, err := os.ReadDir("/etc/nginx/sites-available")
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !strings.HasPrefix(entry.Name(), "a-") {
			continue // Not an OweHost config
		}

		if !validConfigs[entry.Name()] {
			// Config is stale, remove it
			availPath := filepath.Join("/etc/nginx/sites-available", entry.Name())
			enabledPath := filepath.Join("/etc/nginx/sites-enabled", entry.Name())

			os.Remove(enabledPath)
			os.Remove(availPath)
			removed = append(removed, entry.Name())
		}
	}

	return removed, nil
}

// GenerateMainNginxConfig generates the main nginx include config
func (g *Generator) GenerateMainNginxConfig() error {
	content := `# OweHost managed configuration
# Include all account site configurations
include /etc/nginx/sites-enabled/a-*.conf;
`
	return os.WriteFile("/etc/nginx/conf.d/owehost.conf", []byte(content), 0644)
}
