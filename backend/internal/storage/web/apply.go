// Package web provides filesystem-based web/site state management
package web

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/iSundram/OweHost/internal/storage/account"
)

// Applier handles idempotent site configuration application
type Applier struct {
	state          *StateManager
	nginxTemplate  *template.Template
	phpFpmTemplate *template.Template
}

// NewApplier creates a new web applier
func NewApplier() *Applier {
	a := &Applier{
		state: NewStateManager(),
	}
	a.initTemplates()
	return a
}

// initTemplates initializes configuration templates
func (a *Applier) initTemplates() {
	a.nginxTemplate = template.Must(template.New("nginx").Parse(nginxConfigTemplate))
	a.phpFpmTemplate = template.Must(template.New("phpfpm").Parse(phpFpmPoolTemplate))
}

// ApplySite applies a site configuration (creates directories, generates configs)
func (a *Applier) ApplySite(accountID int, site *SiteDescriptor) error {
	// Step 1: Validate
	if err := ValidateSite(site); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Step 2: Write site state
	if err := a.state.WriteSite(accountID, site); err != nil {
		return fmt.Errorf("failed to write site: %w", err)
	}

	// Step 3: Create default index if needed
	docRoot := site.DocumentRoot
	if docRoot == "" {
		docRoot = "public"
	}
	if err := a.state.CreateDefaultIndex(accountID, site.Domain, docRoot); err != nil {
		fmt.Printf("warning: failed to create default index: %v\n", err)
	}

	// Step 4: Generate and apply web server config
	if err := a.GenerateNginxConfig(accountID, site); err != nil {
		return fmt.Errorf("failed to generate nginx config: %w", err)
	}

	// Step 5: Generate PHP-FPM pool if needed
	if strings.HasPrefix(site.Runtime, "php-") {
		if err := a.GeneratePHPFpmPool(accountID, site); err != nil {
			return fmt.Errorf("failed to generate PHP-FPM pool: %w", err)
		}
	}

	// Step 6: Set up SSL if enabled
	if site.SSL {
		if err := a.ensureSSL(accountID, site.Domain); err != nil {
			fmt.Printf("warning: failed to set up SSL: %v\n", err)
		}
	}

	return nil
}

// DeleteSite removes a site and its configuration
func (a *Applier) DeleteSite(accountID int, domain string) error {
	// Remove nginx config
	nginxPath := fmt.Sprintf("/etc/nginx/sites-enabled/a-%d-%s.conf", accountID, domain)
	os.Remove(nginxPath)
	os.Remove(strings.Replace(nginxPath, "sites-enabled", "sites-available", 1))

	// Remove PHP-FPM pool if exists
	phpFpmPath := fmt.Sprintf("/etc/php/8.2/fpm/pool.d/a-%d-%s.conf", accountID, domain)
	os.Remove(phpFpmPath)

	// Remove site directory
	return a.state.DeleteSite(accountID, domain)
}

// GenerateNginxConfig generates and writes nginx configuration
func (a *Applier) GenerateNginxConfig(accountID int, site *SiteDescriptor) error {
	config, err := a.renderNginxConfig(accountID, site)
	if err != nil {
		return err
	}

	// Write to sites-available
	availablePath := fmt.Sprintf("/etc/nginx/sites-available/a-%d-%s.conf", accountID, site.Domain)
	if err := os.MkdirAll(filepath.Dir(availablePath), 0755); err != nil {
		return err
	}
	if err := os.WriteFile(availablePath, []byte(config), 0644); err != nil {
		return err
	}

	// Symlink to sites-enabled
	enabledPath := fmt.Sprintf("/etc/nginx/sites-enabled/a-%d-%s.conf", accountID, site.Domain)
	os.Remove(enabledPath)
	return os.Symlink(availablePath, enabledPath)
}

// renderNginxConfig renders the nginx configuration
func (a *Applier) renderNginxConfig(accountID int, site *SiteDescriptor) (string, error) {
	data := struct {
		*SiteDescriptor
		AccountID    int
		AccountPath  string
		SitePath     string
		DocumentPath string
		PHPSocket    string
		ServerNames  string
	}{
		SiteDescriptor: site,
		AccountID:      accountID,
		AccountPath:    fmt.Sprintf("%s/%s%d", account.BaseAccountPath, account.AccountPrefix, accountID),
		SitePath:       a.state.SitePath(accountID, site.Domain),
	}

	docRoot := site.DocumentRoot
	if docRoot == "" {
		docRoot = "public"
	}
	data.DocumentPath = filepath.Join(data.SitePath, docRoot)

	if site.PHPSettings != nil {
		data.PHPSocket = fmt.Sprintf("/run/php/php%s-fpm-a%d.sock", site.PHPSettings.Version, accountID)
	} else if strings.HasPrefix(site.Runtime, "php-") {
		version := strings.TrimPrefix(site.Runtime, "php-")
		data.PHPSocket = fmt.Sprintf("/run/php/php%s-fpm-a%d.sock", version, accountID)
	}

	serverNames := []string{site.Domain, "www." + site.Domain}
	serverNames = append(serverNames, site.Aliases...)
	data.ServerNames = strings.Join(serverNames, " ")

	var buf strings.Builder
	if err := a.nginxTemplate.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// GeneratePHPFpmPool generates PHP-FPM pool configuration
func (a *Applier) GeneratePHPFpmPool(accountID int, site *SiteDescriptor) error {
	version := "8.2"
	if site.PHPSettings != nil && site.PHPSettings.Version != "" {
		version = site.PHPSettings.Version
	} else if strings.HasPrefix(site.Runtime, "php-") {
		version = strings.TrimPrefix(site.Runtime, "php-")
	}

	data := struct {
		AccountID   int
		Domain      string
		Version     string
		PHPSettings *PHPSettings
		AccountPath string
	}{
		AccountID:   accountID,
		Domain:      site.Domain,
		Version:     version,
		PHPSettings: site.PHPSettings,
		AccountPath: fmt.Sprintf("%s/%s%d", account.BaseAccountPath, account.AccountPrefix, accountID),
	}

	if data.PHPSettings == nil {
		data.PHPSettings = DefaultPHPSettings(version)
	}

	var buf strings.Builder
	if err := a.phpFpmTemplate.Execute(&buf, data); err != nil {
		return err
	}

	poolPath := fmt.Sprintf("/etc/php/%s/fpm/pool.d/a-%d-%s.conf", version, accountID, site.Domain)
	if err := os.MkdirAll(filepath.Dir(poolPath), 0755); err != nil {
		return err
	}

	return os.WriteFile(poolPath, []byte(buf.String()), 0644)
}

// ensureSSL ensures SSL is set up for the domain
func (a *Applier) ensureSSL(accountID int, domain string) error {
	meta, _ := a.state.ReadSSLMeta(accountID, domain)
	if meta != nil {
		return nil
	}
	return a.createSelfSignedCert(accountID, domain)
}

// createSelfSignedCert creates a self-signed certificate for development
func (a *Applier) createSelfSignedCert(accountID int, domain string) error {
	sslPath := a.state.SSLPath(accountID, domain)
	if err := os.MkdirAll(sslPath, 0700); err != nil {
		return err
	}

	certPath := filepath.Join(sslPath, "cert.pem")
	keyPath := filepath.Join(sslPath, "key.pem")

	cmd := exec.Command("openssl", "req",
		"-x509", "-nodes",
		"-days", "365",
		"-newkey", "rsa:2048",
		"-keyout", keyPath,
		"-out", certPath,
		"-subj", fmt.Sprintf("/CN=%s", domain),
	)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to generate self-signed cert: %w", err)
	}

	meta := &SSLMeta{
		Domain:    domain,
		Type:      "self-signed",
		AutoRenew: false,
	}

	return a.state.atomicWrite(filepath.Join(sslPath, "meta.json"), meta)
}

// ReloadNginx reloads nginx configuration
func (a *Applier) ReloadNginx() error {
	if err := exec.Command("nginx", "-t").Run(); err != nil {
		return fmt.Errorf("nginx config test failed: %w", err)
	}
	return exec.Command("nginx", "-s", "reload").Run()
}

// ReloadPHPFpm reloads PHP-FPM
func (a *Applier) ReloadPHPFpm(version string) error {
	serviceName := fmt.Sprintf("php%s-fpm", version)
	return exec.Command("systemctl", "reload", serviceName).Run()
}

// Nginx configuration template
const nginxConfigTemplate = `# Generated by OweHost for {{ .Domain }}
# Account: {{ .AccountID }}

server {
    listen 80;
    listen [::]:80;
    server_name {{ .ServerNames }};

    root {{ .DocumentPath }};
    index index.php index.html index.htm;

    access_log {{ .SitePath }}/logs/access.log;
    error_log {{ .SitePath }}/logs/error.log;

    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;

    {{ if .SSL }}{{ if .SSLRedirect }}
    return 301 https://$server_name$request_uri;
    {{ end }}{{ end }}

    location / {
        try_files $uri $uri/ /index.php?$query_string;
    }

    {{ if .PHPSocket }}
    location ~ \.php$ {
        fastcgi_pass unix:{{ .PHPSocket }};
        fastcgi_index index.php;
        fastcgi_param SCRIPT_FILENAME $realpath_root$fastcgi_script_name;
        include fastcgi_params;
    }
    {{ end }}

    location ~ /\.(ht|git|svn) {
        deny all;
    }

    location ~ /\.(env|json|lock|md)$ {
        deny all;
    }

    location ~* \.(jpg|jpeg|png|gif|ico|css|js|woff2?)$ {
        expires 30d;
        add_header Cache-Control "public, immutable";
    }
}

{{ if .SSL }}
server {
    listen 443 ssl http2;
    listen [::]:443 ssl http2;
    server_name {{ .ServerNames }};

    root {{ .DocumentPath }};
    index index.php index.html index.htm;

    ssl_certificate {{ .AccountPath }}/ssl/{{ .Domain }}/cert.pem;
    ssl_certificate_key {{ .AccountPath }}/ssl/{{ .Domain }}/key.pem;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256;
    ssl_prefer_server_ciphers off;

    access_log {{ .SitePath }}/logs/access.log;
    error_log {{ .SitePath }}/logs/error.log;

    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;

    location / {
        try_files $uri $uri/ /index.php?$query_string;
    }

    {{ if .PHPSocket }}
    location ~ \.php$ {
        fastcgi_pass unix:{{ .PHPSocket }};
        fastcgi_index index.php;
        fastcgi_param SCRIPT_FILENAME $realpath_root$fastcgi_script_name;
        include fastcgi_params;
    }
    {{ end }}

    location ~ /\.(ht|git|svn) {
        deny all;
    }

    location ~* \.(jpg|jpeg|png|gif|ico|css|js|woff2?)$ {
        expires 30d;
        add_header Cache-Control "public, immutable";
    }
}
{{ end }}
`

// PHP-FPM pool template
const phpFpmPoolTemplate = `; Generated by OweHost for {{ .Domain }}
; Account: {{ .AccountID }}

[a{{ .AccountID }}-{{ .Domain }}]
user = a{{ .AccountID }}
group = a{{ .AccountID }}

listen = /run/php/php{{ .Version }}-fpm-a{{ .AccountID }}.sock
listen.owner = www-data
listen.group = www-data
listen.mode = 0660

pm = dynamic
pm.max_children = 10
pm.start_servers = 2
pm.min_spare_servers = 1
pm.max_spare_servers = 4
pm.max_requests = 500

chdir = {{ .AccountPath }}/web/{{ .Domain }}

{{ with .PHPSettings }}
php_admin_value[max_execution_time] = {{ .MaxExecutionTime }}
php_admin_value[memory_limit] = {{ .MemoryLimit }}
php_admin_value[post_max_size] = {{ .PostMaxSize }}
php_admin_value[upload_max_filesize] = {{ .UploadMaxFilesize }}
php_admin_value[max_input_vars] = {{ .MaxInputVars }}
php_admin_value[display_errors] = {{ if .DisplayErrors }}On{{ else }}Off{{ end }}
{{ end }}

php_admin_value[open_basedir] = {{ .AccountPath }}:/tmp:/usr/share/php
php_admin_value[disable_functions] = exec,passthru,shell_exec,system,proc_open,popen
php_admin_value[error_log] = {{ .AccountPath }}/logs/php-error.log
`
