// Package main is the entry point for OweHost backend server
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/iSundram/OweHost/internal/accountsvc"
	"github.com/iSundram/OweHost/internal/api/middleware"
	v1 "github.com/iSundram/OweHost/internal/api/v1"
	"github.com/iSundram/OweHost/internal/appinstaller"
	"github.com/iSundram/OweHost/internal/audit"
	"github.com/iSundram/OweHost/internal/auth"
	"github.com/iSundram/OweHost/internal/authorization"
	"github.com/iSundram/OweHost/internal/backup"
	"github.com/iSundram/OweHost/internal/cluster"
	"github.com/iSundram/OweHost/internal/cron"
	"github.com/iSundram/OweHost/internal/database"
	"github.com/iSundram/OweHost/internal/dns"
	"github.com/iSundram/OweHost/internal/domain"
	"github.com/iSundram/OweHost/internal/feature"
	"github.com/iSundram/OweHost/internal/filesystem"
	"github.com/iSundram/OweHost/internal/firewall"
	"github.com/iSundram/OweHost/internal/ftp"
	"github.com/iSundram/OweHost/internal/git"
	"github.com/iSundram/OweHost/internal/installation"
	"github.com/iSundram/OweHost/internal/licensing"
	"github.com/iSundram/OweHost/internal/logging"
	"github.com/iSundram/OweHost/internal/metrics"
	"github.com/iSundram/OweHost/internal/notification"
	"github.com/iSundram/OweHost/internal/oscontrol"
	"github.com/iSundram/OweHost/internal/packages"
	"github.com/iSundram/OweHost/internal/plugin"
	"github.com/iSundram/OweHost/internal/provisioning"
	"github.com/iSundram/OweHost/internal/recovery"
	"github.com/iSundram/OweHost/internal/reseller"
	"github.com/iSundram/OweHost/internal/resource"
	"github.com/iSundram/OweHost/internal/runtime"
	"github.com/iSundram/OweHost/internal/ssh"
	"github.com/iSundram/OweHost/internal/ssl"
	"github.com/iSundram/OweHost/internal/stats"
	"github.com/iSundram/OweHost/internal/twofactor"
	"github.com/iSundram/OweHost/internal/user"
	"github.com/iSundram/OweHost/internal/webserver"
	"github.com/iSundram/OweHost/internal/websocket"
	"github.com/iSundram/OweHost/pkg/config"
	pkgdb "github.com/iSundram/OweHost/pkg/database"
)

// Server represents the OweHost API server
type Server struct {
	config *config.Config
	api    *http.Server
	user   *http.Server
	admin  *http.Server
	db     *pkgdb.DB

	// Services
	authService          *auth.Service
	authorizationService *authorization.Service
	userService          *user.Service
	accountService       *accountsvc.Service
	packageService       *packages.Service
	featureService       *feature.Service
	resellerService      *reseller.Service
	resourceService      *resource.Service
	domainService        *domain.Service
	dnsService           *dns.Service
	webserverService     *webserver.Service
	runtimeService       *runtime.Service
	databaseService      *database.Service
	filesystemService    *filesystem.Service
	backupService        *backup.Service
	sslService           *ssl.Service
	firewallService      *firewall.Service
	cronService          *cron.Service
	appinstallerService  *appinstaller.Service
	pluginService        *plugin.Service
	loggingService       *logging.Service
	notificationService  *notification.Service
	clusterService       *cluster.Service
	oscontrolService     *oscontrol.Service
	licensingService     *licensing.Service
	recoveryService      *recovery.Service
	installationService  *installation.Service
	provisioningService  *provisioning.Service

	// New enhanced services
	ftpService       *ftp.Service
	sshService       *ssh.Service
	gitService       *git.Service
	statsService     *stats.Service
	twoFactorService *twofactor.Service
	auditService     *audit.Service
	metricsService   *metrics.Metrics
	wsHub            *websocket.Hub
	rateLimiter      *middleware.UserRateLimiter
}

// NewServer creates a new server instance
func NewServer(cfg *config.Config) *Server {
	s := &Server{
		config: cfg,
	}

	// Initialize all services
	s.initServices()

	return s
}

// initServices initializes all services
func (s *Server) initServices() {
	// Initialize Database
	var err error
	s.db, err = pkgdb.New(pkgdb.Config{
		Driver:          s.config.Database.Driver,
		Host:            s.config.Database.Host,
		Port:            s.config.Database.Port,
		Database:        s.config.Database.Name,
		Username:        s.config.Database.User,
		Password:        s.config.Database.Password,
		SSLMode:         s.config.Database.SSLMode,
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 5 * time.Minute,
	})
	if err != nil {
		fmt.Printf("Warning: Failed to connect to database: %v\n", err)
	}

	// Repositories
	var userRepo *user.Repository
	if s.db != nil {
		userRepo = user.NewRepository(s.db)
	}

	s.loggingService = logging.NewService()
	s.authService = auth.NewService(s.config)
	s.authorizationService = authorization.NewService()
	s.userService = user.NewService(s.config, userRepo)
	s.accountService = accountsvc.NewService()
	s.packageService = packages.NewService()
	s.featureService = feature.NewService()
	s.resellerService = reseller.NewService()
	s.resourceService = resource.NewService()
	s.domainService = domain.NewService()
	s.dnsService = dns.NewService()
	s.webserverService = webserver.NewService()
	s.runtimeService = runtime.NewService()
	s.databaseService = database.NewService()
	s.filesystemService = filesystem.NewService()
	s.backupService = backup.NewService()
	s.sslService = ssl.NewService()
	s.firewallService = firewall.NewService()
	s.cronService = cron.NewService()
	s.appinstallerService = appinstaller.NewService()
	s.pluginService = plugin.NewService()
	s.notificationService = notification.NewService()
	s.clusterService = cluster.NewService()
	s.oscontrolService = oscontrol.NewService()
	s.licensingService = licensing.NewService()
	s.recoveryService = recovery.NewService()
	s.installationService = installation.NewService(s.config, s.userService)
	s.provisioningService = provisioning.NewService(
		s.userService,
		s.domainService,
		s.databaseService,
		s.filesystemService,
		s.dnsService,
		s.sslService,
		s.resourceService,
		s.oscontrolService,
		s.webserverService,
		s.backupService,
	)

	// Initialize new enhanced services
	s.ftpService = ftp.NewService()
	s.sshService = ssh.NewService()
	s.gitService = git.NewService()
	s.statsService = stats.NewService()
	s.twoFactorService = twofactor.NewService()
	s.auditService = audit.NewService()
	s.metricsService = metrics.NewMetrics()
	s.wsHub = websocket.NewHub()
	s.rateLimiter = middleware.NewUserRateLimiter()
}

// setupAPIRoutes sets up API routes (Port 8080)
func (s *Server) setupAPIRoutes() http.Handler {
	mux := http.NewServeMux()

	// Create handlers
	authHandler := v1.NewAuthHandler(s.authService, s.userService)
	userHandler := v1.NewUserHandler(s.userService)
	resellerHandler := v1.NewResellerHandler(s.resellerService, s.userService)
	domainHandler := v1.NewDomainHandler(s.domainService, s.userService)
	databaseHandler := v1.NewDatabaseHandler(s.databaseService, s.userService)
	healthHandler := v1.NewHealthHandler(s.loggingService)
	installationHandler := v1.NewInstallationHandler(s.installationService)
	adminHandler := v1.NewAdminHandler(s.userService, s.resellerService, s.domainService, s.databaseService, s.oscontrolService)
	dnsHandler := v1.NewDNSHandler(s.dnsService, s.domainService)
	accountHandler := v1.NewAccountHandler(s.accountService, s.userService, s.domainService, s.dnsService)
	packageHandler := v1.NewPackageHandler(s.packageService)
	featureHandler := v1.NewFeatureHandler(s.featureService)

	// New handlers
	sslHandler := v1.NewSSLHandler(s.sslService, s.userService)
	backupHandler := v1.NewBackupHandler(s.backupService, s.userService)
	cronHandler := v1.NewCronHandler(s.cronService, s.userService)
	filesystemHandler := v1.NewFileSystemHandler(s.filesystemService, s.userService)
	appinstallerHandler := v1.NewAppInstallerHandler(s.appinstallerService, s.userService)
	webserverHandler := v1.NewWebServerHandler(s.webserverService, s.userService)
	notificationHandler := v1.NewNotificationHandler(s.notificationService, s.userService)
	firewallHandler := v1.NewFirewallHandler(s.firewallService, s.userService)
	resourceHandler := v1.NewResourceHandler(s.resourceService, s.userService)
	provisioningHandler := v1.NewProvisioningHandler(s.provisioningService, s.userService)

	// Enhanced service handlers
	ftpHandler := v1.NewFTPHandler(s.ftpService, s.userService)
	sshHandler := v1.NewSSHHandler(s.sshService, s.userService)
	twoFactorHandler := v1.NewTwoFactorHandler(s.twoFactorService, s.userService)
	auditHandler := v1.NewAuditHandler(s.auditService, s.userService)

	// Missing handlers that need routes registered
	statsHandler := v1.NewStatsHandler(s.statsService)
	gitHandler := v1.NewGitHandler(s.gitService)
	clusterHandler := v1.NewClusterHandler(s.clusterService)
	runtimeHandler := v1.NewRuntimeHandler(s.runtimeService)
	loggingHandler := v1.NewLoggingHandler(s.loggingService)
	pluginHandler := v1.NewPluginHandler(s.pluginService)

	// Helper wrappers
	authWrap := func(handler http.HandlerFunc) http.Handler {
		return middleware.AuthMiddleware(s.authService)(http.HandlerFunc(handler))
	}
	adminWrap := func(handler http.HandlerFunc) http.Handler {
		return middleware.AuthMiddleware(s.authService)(
			middleware.AdminOnlyMiddleware(s.userService)(http.HandlerFunc(handler)),
		)
	}
	adminOrResellerWrap := func(handler http.HandlerFunc) http.Handler {
		return middleware.AuthMiddleware(s.authService)(
			middleware.AdminOrResellerMiddleware(s.userService)(http.HandlerFunc(handler)),
		)
	}

	// Installation endpoints (no auth required)
	mux.HandleFunc("/api/v1/installation/check", installationHandler.Check)
	mux.HandleFunc("/api/v1/installation/install", installationHandler.Install)

	// Health endpoints (no auth required)
	mux.HandleFunc("/health", healthHandler.Health)
	mux.HandleFunc("/ready", healthHandler.Ready)
	mux.Handle("/metrics", s.metricsService.Handler())

	// Auth endpoints (no auth required)
	mux.HandleFunc("/api/v1/auth/login", authHandler.Login)
	mux.HandleFunc("/api/v1/auth/refresh", authHandler.Refresh)
	mux.HandleFunc("/api/v1/auth/logout", authHandler.Logout)

	// User endpoints (protected)
	mux.Handle("/api/v1/users", authWrap(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			userHandler.List(w, r)
		case http.MethodPost:
			userHandler.Create(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	mux.Handle("/api/v1/users/me", authWrap(userHandler.Me))
	mux.Handle("/api/v1/users/", authWrap(func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) >= 5 {
			action := parts[len(parts)-1]

			if action == "suspend" && r.Method == http.MethodPost {
				userHandler.Suspend(w, r)
				return
			}
			if action == "terminate" && r.Method == http.MethodPost {
				userHandler.Terminate(w, r)
				return
			}
		}

		switch r.Method {
		case http.MethodGet:
			userHandler.Get(w, r)
		case http.MethodPut, http.MethodPatch:
			userHandler.Update(w, r)
		case http.MethodDelete:
			userHandler.Delete(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	// Domain endpoints (protected)
	mux.Handle("/api/v1/domains", authWrap(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			domainHandler.List(w, r)
		case http.MethodPost:
			domainHandler.Create(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	mux.Handle("/api/v1/domains/", authWrap(func(w http.ResponseWriter, r *http.Request) {
		// Handle validation and subdomain operations based on trailing segment
		parts := strings.Split(strings.TrimSuffix(r.URL.Path, "/"), "/")
		// e.g. /api/v1/domains/{id}/validate or /api/v1/domains/{id}/subdomains
		if len(parts) >= 6 && parts[len(parts)-1] == "validate" && r.Method == http.MethodPost {
			domainHandler.Validate(w, r)
			return
		}
		if len(parts) >= 6 && parts[len(parts)-1] == "subdomains" {
			if r.Method == http.MethodPost {
				domainHandler.CreateSubdomain(w, r)
				return
			}
			if r.Method == http.MethodGet {
				domainHandler.ListSubdomains(w, r)
				return
			}
		}

		// Default domain operations
		switch r.Method {
		case http.MethodGet:
			domainHandler.Get(w, r)
		case http.MethodDelete:
			domainHandler.Delete(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	// Subdomain deletion by ID
	mux.Handle("/api/v1/subdomains/", authWrap(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			domainHandler.DeleteSubdomain(w, r)
			return
		}
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}))

	// Database endpoints (protected)
	mux.Handle("/api/v1/databases", authWrap(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			databaseHandler.List(w, r)
		case http.MethodPost:
			databaseHandler.Create(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	mux.Handle("/api/v1/databases/", authWrap(func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(strings.TrimSuffix(r.URL.Path, "/"), "/")

		// Nested resources: users/backups
		if len(parts) >= 6 && parts[len(parts)-1] == "users" {
			if r.Method == http.MethodPost {
				databaseHandler.CreateUser(w, r)
				return
			}
			if r.Method == http.MethodGet {
				databaseHandler.ListUsers(w, r)
				return
			}
		}
		if len(parts) >= 6 && parts[len(parts)-1] == "backups" {
			if r.Method == http.MethodPost {
				databaseHandler.CreateBackup(w, r)
				return
			}
			if r.Method == http.MethodGet {
				databaseHandler.ListBackups(w, r)
				return
			}
		}

		// Default database operations
		switch r.Method {
		case http.MethodGet:
			databaseHandler.Get(w, r)
		case http.MethodDelete:
			databaseHandler.Delete(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	// Database user deletion
	mux.Handle("/api/v1/database-users/", authWrap(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			databaseHandler.DeleteUser(w, r)
			return
		}
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}))
	// Database backup restore
	mux.Handle("/api/v1/database-backups/", authWrap(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			databaseHandler.RestoreBackup(w, r)
			return
		}
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}))

	// Account functions (admin/reseller)
	mux.Handle("/api/v1/accounts", adminOrResellerWrap(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			accountHandler.List(w, r)
		case http.MethodPost:
			accountHandler.Create(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	mux.Handle("/api/v1/accounts/", adminOrResellerWrap(func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(strings.TrimSuffix(r.URL.Path, "/"), "/")
		if len(parts) >= 6 && (parts[len(parts)-1] == "suspend" || parts[len(parts)-1] == "unsuspend" || parts[len(parts)-1] == "terminate") {
			accountHandler.UpdateStatus(w, r)
			return
		}
		http.Error(w, "Not found", http.StatusNotFound)
	}))

	// Package/Plan endpoints (protected)
	mux.Handle("/api/v1/packages", authWrap(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			packageHandler.List(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	mux.Handle("/api/v1/packages/", authWrap(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			packageHandler.Get(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	// Feature manager endpoints (admin only)
	mux.Handle("/api/v1/features", adminWrap(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			featureHandler.List(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	mux.Handle("/api/v1/features/", adminWrap(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			featureHandler.Get(w, r)
		case http.MethodPut, http.MethodPatch:
			featureHandler.Update(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	// DNS functions
	mux.Handle("/api/v1/dns/zones", authWrap(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			dnsHandler.ListZones(w, r)
		case http.MethodPost:
			dnsHandler.CreateZone(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	mux.Handle("/api/v1/dns/zones/", authWrap(func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(strings.TrimSuffix(r.URL.Path, "/"), "/")
		// /api/v1/dns/zones/{id}
		if len(parts) == 6 {
			switch r.Method {
			case http.MethodGet:
				dnsHandler.GetZone(w, r)
			case http.MethodDelete:
				dnsHandler.DeleteZone(w, r)
			default:
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
			return
		}

		// nested: /api/v1/dns/zones/{id}/records
		if len(parts) >= 7 && parts[len(parts)-1] == "records" {
			if r.Method == http.MethodGet {
				dnsHandler.ListRecords(w, r)
				return
			}
			if r.Method == http.MethodPost {
				dnsHandler.CreateRecord(w, r)
				return
			}
		}
		// DNSSEC enable /api/v1/dns/zones/{id}/dnssec
		if len(parts) >= 7 && parts[len(parts)-1] == "dnssec" && r.Method == http.MethodPost {
			dnsHandler.EnableDNSSEC(w, r)
			return
		}
		// Sync /api/v1/dns/zones/{id}/sync
		if len(parts) >= 7 && parts[len(parts)-1] == "sync" && r.Method == http.MethodPost {
			dnsHandler.SyncZone(w, r)
			return
		}

		http.Error(w, "Not found", http.StatusNotFound)
	}))
	// DNS records direct /api/v1/dns/records/{id}
	mux.Handle("/api/v1/dns/records/", authWrap(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodDelete:
			dnsHandler.DeleteRecord(w, r)
		case http.MethodPut, http.MethodPatch:
			dnsHandler.UpdateRecord(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	// Reseller endpoints (protected - admin and reseller access)
	mux.Handle("/api/v1/resellers", adminOrResellerWrap(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			resellerHandler.List(w, r)
		case http.MethodPost:
			resellerHandler.Create(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	// Admin endpoints (protected - admin only)
	mux.Handle("/api/v1/admin/stats", adminWrap(adminHandler.GetSystemStats))
	mux.Handle("/api/v1/admin/services", adminWrap(adminHandler.GetServiceStatus))
	mux.Handle("/api/v1/resellers/me", authWrap(resellerHandler.GetByUserID))
	mux.Handle("/api/v1/resellers/", adminOrResellerWrap(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			resellerHandler.Get(w, r)
		case http.MethodPut, http.MethodPatch:
			resellerHandler.Update(w, r)
		case http.MethodDelete:
			resellerHandler.Delete(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	// SSL Certificate endpoints
	mux.Handle("/api/v1/ssl/certificates", authWrap(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			sslHandler.ListCertificates(w, r)
		case http.MethodPost:
			sslHandler.CreateCertificate(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	mux.Handle("/api/v1/ssl/certificates/", authWrap(func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(strings.TrimSuffix(r.URL.Path, "/"), "/")
		if len(parts) >= 7 && parts[len(parts)-1] == "renew" && r.Method == http.MethodPost {
			sslHandler.RenewCertificate(w, r)
			return
		}
		switch r.Method {
		case http.MethodGet:
			sslHandler.GetCertificate(w, r)
		case http.MethodDelete:
			sslHandler.DeleteCertificate(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	mux.Handle("/api/v1/ssl/letsencrypt", authWrap(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			sslHandler.RequestLetsEncrypt(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	mux.Handle("/api/v1/ssl/install", authWrap(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			sslHandler.InstallCertificate(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	mux.Handle("/api/v1/ssl/csr", authWrap(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			sslHandler.GenerateCSR(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	mux.Handle("/api/v1/ssl/verify-domain", authWrap(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			sslHandler.VerifyDomain(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	// Backup endpoints
	mux.Handle("/api/v1/backups", authWrap(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			backupHandler.ListBackups(w, r)
		case http.MethodPost:
			backupHandler.CreateBackup(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	mux.Handle("/api/v1/backups/", authWrap(func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(strings.TrimSuffix(r.URL.Path, "/"), "/")
		if len(parts) >= 6 {
			action := parts[len(parts)-1]
			if action == "restore" && r.Method == http.MethodPost {
				backupHandler.RestoreBackup(w, r)
				return
			}
			if action == "download" && r.Method == http.MethodGet {
				backupHandler.DownloadBackup(w, r)
				return
			}
		}
		switch r.Method {
		case http.MethodGet:
			backupHandler.GetBackup(w, r)
		case http.MethodDelete:
			backupHandler.DeleteBackup(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	mux.Handle("/api/v1/backups/schedule", authWrap(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			backupHandler.GetBackupSchedule(w, r)
		case http.MethodPut:
			backupHandler.UpdateBackupSchedule(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	// Cron Job endpoints
	mux.Handle("/api/v1/cron/jobs", authWrap(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			cronHandler.ListCronJobs(w, r)
		case http.MethodPost:
			cronHandler.CreateCronJob(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	mux.Handle("/api/v1/cron/jobs/", authWrap(func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(strings.TrimSuffix(r.URL.Path, "/"), "/")
		if len(parts) >= 7 {
			action := parts[len(parts)-1]
			if action == "enable" && r.Method == http.MethodPost {
				cronHandler.EnableCronJob(w, r)
				return
			}
			if action == "disable" && r.Method == http.MethodPost {
				cronHandler.DisableCronJob(w, r)
				return
			}
			if action == "executions" && r.Method == http.MethodGet {
				cronHandler.GetCronJobExecutions(w, r)
				return
			}
		}
		switch r.Method {
		case http.MethodGet:
			cronHandler.GetCronJob(w, r)
		case http.MethodPut, http.MethodPatch:
			cronHandler.UpdateCronJob(w, r)
		case http.MethodDelete:
			cronHandler.DeleteCronJob(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	mux.Handle("/api/v1/cron/validate", authWrap(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			cronHandler.ValidateCronExpression(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	// File System endpoints
	mux.Handle("/api/v1/files", authWrap(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			filesystemHandler.ListFiles(w, r)
		case http.MethodPost:
			filesystemHandler.CreateFile(w, r)
		case http.MethodPut:
			filesystemHandler.UpdateFile(w, r)
		case http.MethodDelete:
			filesystemHandler.DeleteFile(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	mux.Handle("/api/v1/files/download", authWrap(filesystemHandler.DownloadFile))
	mux.Handle("/api/v1/files/upload", authWrap(filesystemHandler.UploadFile))
	mux.Handle("/api/v1/files/copy", authWrap(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			filesystemHandler.CopyFile(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	mux.Handle("/api/v1/files/move", authWrap(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			filesystemHandler.MoveFile(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	mux.Handle("/api/v1/files/permissions", authWrap(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			filesystemHandler.GetPermissions(w, r)
		case http.MethodPut:
			filesystemHandler.UpdatePermissions(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	mux.Handle("/api/v1/files/compress", authWrap(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			filesystemHandler.CompressFiles(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	mux.Handle("/api/v1/files/extract", authWrap(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			filesystemHandler.ExtractArchive(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	mux.Handle("/api/v1/files/search", authWrap(filesystemHandler.SearchFiles))
	mux.Handle("/api/v1/files/disk-usage", authWrap(filesystemHandler.GetDiskUsage))

	// App Installer endpoints
	mux.Handle("/api/v1/apps", authWrap(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			appinstallerHandler.ListAvailableApps(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	mux.Handle("/api/v1/apps/", authWrap(func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(strings.TrimSuffix(r.URL.Path, "/"), "/")
		if len(parts) >= 6 && parts[len(parts)-1] == "install" && r.Method == http.MethodPost {
			appinstallerHandler.InstallApp(w, r)
			return
		}
		if r.Method == http.MethodGet {
			appinstallerHandler.GetApp(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	mux.Handle("/api/v1/apps/installed", authWrap(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			appinstallerHandler.ListInstalledApps(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	mux.Handle("/api/v1/apps/installed/", authWrap(func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(strings.TrimSuffix(r.URL.Path, "/"), "/")
		if len(parts) >= 7 && parts[len(parts)-1] == "status" && r.Method == http.MethodGet {
			appinstallerHandler.GetInstallationStatus(w, r)
			return
		}
		switch r.Method {
		case http.MethodGet:
			appinstallerHandler.GetInstalledApp(w, r)
		case http.MethodPut:
			appinstallerHandler.UpdateInstalledApp(w, r)
		case http.MethodDelete:
			appinstallerHandler.UninstallApp(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	// Web Server endpoints
	mux.Handle("/api/v1/webserver/config", authWrap(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			webserverHandler.GetConfig(w, r)
		case http.MethodPut:
			webserverHandler.UpdateConfig(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	mux.Handle("/api/v1/webserver/php", authWrap(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			webserverHandler.ListPHPVersions(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	mux.Handle("/api/v1/webserver/php/switch", authWrap(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			webserverHandler.SwitchPHPVersion(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	mux.Handle("/api/v1/webserver/modules", authWrap(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			webserverHandler.ListModules(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	mux.Handle("/api/v1/webserver/modules/enable", authWrap(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			webserverHandler.EnableModule(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	mux.Handle("/api/v1/webserver/modules/disable", authWrap(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			webserverHandler.DisableModule(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	mux.Handle("/api/v1/webserver/restart", authWrap(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			webserverHandler.RestartServer(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	mux.Handle("/api/v1/webserver/error-logs", authWrap(webserverHandler.GetErrorLogs))
	mux.Handle("/api/v1/webserver/access-logs", authWrap(webserverHandler.GetAccessLogs))
	mux.Handle("/api/v1/webserver/status", authWrap(webserverHandler.GetServerStatus))
	mux.Handle("/api/v1/webserver/test-config", authWrap(webserverHandler.TestConfiguration))

	// Notification endpoints
	mux.Handle("/api/v1/notifications", authWrap(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			notificationHandler.ListNotifications(w, r)
		case http.MethodPost:
			notificationHandler.CreateNotification(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	mux.Handle("/api/v1/notifications/", authWrap(func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(strings.TrimSuffix(r.URL.Path, "/"), "/")
		if len(parts) >= 6 && parts[len(parts)-1] == "read" && r.Method == http.MethodPut {
			notificationHandler.MarkAsRead(w, r)
			return
		}
		switch r.Method {
		case http.MethodGet:
			notificationHandler.GetNotification(w, r)
		case http.MethodDelete:
			notificationHandler.DeleteNotification(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	mux.Handle("/api/v1/notifications/read-all", authWrap(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut {
			notificationHandler.MarkAllAsRead(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	mux.Handle("/api/v1/notifications/settings", authWrap(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			notificationHandler.GetSettings(w, r)
		case http.MethodPut:
			notificationHandler.UpdateSettings(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	mux.Handle("/api/v1/notifications/unread-count", authWrap(notificationHandler.GetUnreadCount))

	// Firewall endpoints (admin only)
	mux.Handle("/api/v1/firewall/rules", adminWrap(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			firewallHandler.ListRules(w, r)
		case http.MethodPost:
			firewallHandler.CreateRule(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	mux.Handle("/api/v1/firewall/rules/", adminWrap(func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(strings.TrimSuffix(r.URL.Path, "/"), "/")
		if len(parts) >= 7 {
			action := parts[len(parts)-1]
			if action == "enable" && r.Method == http.MethodPost {
				firewallHandler.EnableRule(w, r)
				return
			}
			if action == "disable" && r.Method == http.MethodPost {
				firewallHandler.DisableRule(w, r)
				return
			}
		}
		switch r.Method {
		case http.MethodGet:
			firewallHandler.GetRule(w, r)
		case http.MethodPut:
			firewallHandler.UpdateRule(w, r)
		case http.MethodDelete:
			firewallHandler.DeleteRule(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	mux.Handle("/api/v1/firewall/status", adminWrap(firewallHandler.GetStatus))
	mux.Handle("/api/v1/firewall/enable", adminWrap(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			firewallHandler.EnableFirewall(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	mux.Handle("/api/v1/firewall/disable", adminWrap(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			firewallHandler.DisableFirewall(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	mux.Handle("/api/v1/firewall/block-ip", adminWrap(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			firewallHandler.BlockIP(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	mux.Handle("/api/v1/firewall/unblock-ip", adminWrap(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			firewallHandler.UnblockIP(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	mux.Handle("/api/v1/firewall/blocked-ips", adminWrap(firewallHandler.GetBlockedIPs))

	// Resource endpoints
	mux.Handle("/api/v1/resources/usage", authWrap(resourceHandler.GetUsage))
	mux.Handle("/api/v1/resources/limits", authWrap(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			resourceHandler.GetLimits(w, r)
		case http.MethodPut:
			resourceHandler.UpdateLimits(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	mux.Handle("/api/v1/resources/history", authWrap(resourceHandler.GetHistory))
	mux.Handle("/api/v1/resources/system", adminWrap(resourceHandler.GetSystemResources))

	// Provisioning endpoints (admin/reseller only)
	mux.Handle("/api/v1/provisioning/accounts", adminOrResellerWrap(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			provisioningHandler.ProvisionAccount(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	mux.Handle("/api/v1/provisioning/status", authWrap(provisioningHandler.GetProvisioningStatus))
	mux.Handle("/api/v1/provisioning/deprovision", adminOrResellerWrap(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			provisioningHandler.DeprovisionAccount(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	// FTP endpoints
	mux.Handle("/api/v1/ftp/accounts", authWrap(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			ftpHandler.ListAccounts(w, r)
		case http.MethodPost:
			ftpHandler.CreateAccount(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	mux.Handle("/api/v1/ftp/accounts/", authWrap(func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(strings.TrimSuffix(r.URL.Path, "/"), "/")
		if len(parts) >= 7 {
			action := parts[len(parts)-1]
			if action == "suspend" && r.Method == http.MethodPost {
				ftpHandler.SuspendAccount(w, r)
				return
			}
			if action == "unsuspend" && r.Method == http.MethodPost {
				ftpHandler.UnsuspendAccount(w, r)
				return
			}
		}
		switch r.Method {
		case http.MethodGet:
			ftpHandler.GetAccount(w, r)
		case http.MethodPut, http.MethodPatch:
			ftpHandler.UpdateAccount(w, r)
		case http.MethodDelete:
			ftpHandler.DeleteAccount(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	mux.Handle("/api/v1/ftp/config", adminWrap(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			ftpHandler.GetConfig(w, r)
		case http.MethodPut:
			ftpHandler.UpdateConfig(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	mux.Handle("/api/v1/ftp/sessions", adminWrap(ftpHandler.GetSessions))

	// SSH endpoints
	mux.Handle("/api/v1/ssh/keys", authWrap(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			sshHandler.ListKeys(w, r)
		case http.MethodPost:
			sshHandler.AddKey(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	mux.Handle("/api/v1/ssh/keys/generate", authWrap(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			sshHandler.GenerateKey(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	mux.Handle("/api/v1/ssh/keys/", authWrap(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			sshHandler.GetKey(w, r)
		case http.MethodDelete:
			sshHandler.DeleteKey(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	mux.Handle("/api/v1/ssh/access", authWrap(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			sshHandler.GetAccess(w, r)
		case http.MethodPut:
			sshHandler.UpdateAccess(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	mux.Handle("/api/v1/ssh/access/enable", authWrap(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			sshHandler.EnableAccess(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	mux.Handle("/api/v1/ssh/access/disable", authWrap(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			sshHandler.DisableAccess(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	mux.Handle("/api/v1/ssh/sessions", adminWrap(sshHandler.GetSessions))

	// Two-Factor Authentication endpoints
	mux.Handle("/api/v1/2fa/status", authWrap(twoFactorHandler.GetStatus))
	mux.Handle("/api/v1/2fa/setup", authWrap(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			twoFactorHandler.SetupTOTP(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	mux.Handle("/api/v1/2fa/verify-setup", authWrap(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			twoFactorHandler.VerifySetup(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	mux.Handle("/api/v1/2fa/disable", authWrap(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			twoFactorHandler.Disable(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	mux.Handle("/api/v1/2fa/backup-codes", authWrap(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			twoFactorHandler.RegenerateBackupCodes(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	mux.Handle("/api/v1/2fa/login-attempts", authWrap(twoFactorHandler.GetLoginAttempts))

	// Audit Log endpoints (admin only)
	mux.Handle("/api/v1/audit/logs", adminWrap(auditHandler.ListLogs))
	mux.Handle("/api/v1/audit/stats", adminWrap(auditHandler.GetStats))
	mux.Handle("/api/v1/audit/activity", authWrap(auditHandler.GetUserActivity))
	mux.Handle("/api/v1/audit/resources", adminWrap(auditHandler.GetResourceLogs))
	mux.Handle("/api/v1/audit/security-events", adminWrap(auditHandler.GetSecurityEvents))
	mux.Handle("/api/v1/audit/security-events/", adminWrap(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			auditHandler.ResolveSecurityEvent(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	// Stats endpoints
	mux.Handle("/api/v1/stats/bandwidth/", authWrap(statsHandler.GetBandwidthStats))
	mux.Handle("/api/v1/stats/visitors/", authWrap(statsHandler.GetVisitorStats))
	mux.Handle("/api/v1/stats/errors/", authWrap(statsHandler.GetErrorLogs))
	mux.Handle("/api/v1/stats/access/", authWrap(statsHandler.GetAccessLogs))
	mux.Handle("/api/v1/stats/domain/", authWrap(statsHandler.GetDomainSummary))
	mux.Handle("/api/v1/stats/resources", authWrap(statsHandler.GetResourceStats))
	mux.Handle("/api/v1/stats/user/summary", authWrap(statsHandler.GetUserSummary))
	mux.Handle("/api/v1/stats/user/", adminWrap(statsHandler.GetUserResourceStats))

	// Git endpoints
	mux.Handle("/api/v1/git/repositories", authWrap(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			gitHandler.List(w, r)
		case http.MethodPost:
			gitHandler.CreateRepository(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	mux.Handle("/api/v1/git/repositories/", authWrap(func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(strings.TrimSuffix(r.URL.Path, "/"), "/")
		if len(parts) >= 6 {
			action := parts[len(parts)-1]
			switch action {
			case "clone":
				if r.Method == http.MethodPost {
					gitHandler.Clone(w, r)
					return
				}
			case "pull":
				if r.Method == http.MethodPost {
					gitHandler.Pull(w, r)
					return
				}
			case "deploy":
				if r.Method == http.MethodPost {
					gitHandler.Deploy(w, r)
					return
				}
			case "branches":
				if r.Method == http.MethodGet {
					gitHandler.GetBranches(w, r)
					return
				}
			case "commits":
				if r.Method == http.MethodGet {
					gitHandler.GetCommits(w, r)
					return
				}
			case "deploy-keys":
				if r.Method == http.MethodGet {
					gitHandler.ListDeployKeys(w, r)
					return
				}
				if r.Method == http.MethodPost {
					gitHandler.AddDeployKey(w, r)
					return
				}
			case "webhooks":
				if r.Method == http.MethodGet {
					gitHandler.ListWebhooks(w, r)
					return
				}
				if r.Method == http.MethodPost {
					gitHandler.AddWebhook(w, r)
					return
				}
			}
		}
		switch r.Method {
		case http.MethodGet:
			gitHandler.Get(w, r)
		case http.MethodPut, http.MethodPatch:
			gitHandler.Update(w, r)
		case http.MethodDelete:
			gitHandler.Delete(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	mux.Handle("/api/v1/git/deploy-keys/", authWrap(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			gitHandler.RemoveDeployKey(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	mux.Handle("/api/v1/git/webhooks/", authWrap(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			gitHandler.RemoveWebhook(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	// Cluster endpoints (admin only)
	mux.Handle("/api/v1/cluster/nodes", adminWrap(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			clusterHandler.ListNodes(w, r)
		case http.MethodPost:
			clusterHandler.RegisterNode(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	mux.Handle("/api/v1/cluster/nodes/online", adminWrap(clusterHandler.GetOnlineNodes))
	mux.Handle("/api/v1/cluster/nodes/", adminWrap(func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(strings.TrimSuffix(r.URL.Path, "/"), "/")
		if len(parts) >= 7 {
			action := parts[len(parts)-1]
			if action == "status" && (r.Method == http.MethodPut || r.Method == http.MethodPatch) {
				clusterHandler.UpdateNodeStatus(w, r)
				return
			}
			if action == "heartbeat" && r.Method == http.MethodPost {
				clusterHandler.ProcessHeartbeat(w, r)
				return
			}
		}
		switch r.Method {
		case http.MethodGet:
			clusterHandler.GetNode(w, r)
		case http.MethodDelete:
			clusterHandler.RemoveNode(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	mux.Handle("/api/v1/cluster/capabilities", adminWrap(clusterHandler.DiscoverCapabilities))
	mux.Handle("/api/v1/cluster/placement", adminWrap(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			clusterHandler.PlaceResource(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	mux.Handle("/api/v1/cluster/dead-nodes", adminWrap(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			clusterHandler.CheckDeadNodes(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	mux.Handle("/api/v1/cluster/providers", adminWrap(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			clusterHandler.ListCloudProviders(w, r)
		case http.MethodPost:
			clusterHandler.RegisterCloudProvider(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	mux.Handle("/api/v1/cluster/providers/", adminWrap(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			clusterHandler.GetCloudProvider(w, r)
		case http.MethodDelete:
			clusterHandler.DeleteCloudProvider(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	mux.Handle("/api/v1/cluster/vms", adminWrap(clusterHandler.ListVMInstances))
	mux.Handle("/api/v1/cluster/vm-lifecycle", adminWrap(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			clusterHandler.VMLifecycle(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	// Runtime endpoints
	mux.Handle("/api/v1/runtimes/", authWrap(func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(strings.TrimSuffix(r.URL.Path, "/"), "/")
		if len(parts) >= 6 && parts[len(parts)-1] == "versions" {
			runtimeHandler.ListVersions(w, r)
			return
		}
		http.Error(w, "Not found", http.StatusNotFound)
	}))
	mux.Handle("/api/v1/runtime/php/pools", authWrap(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			runtimeHandler.ListPHPPools(w, r)
		case http.MethodPost:
			runtimeHandler.CreatePHPPool(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	mux.Handle("/api/v1/runtime/php/pools/", authWrap(func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(strings.TrimSuffix(r.URL.Path, "/"), "/")
		if len(parts) >= 8 && parts[len(parts)-1] == "extensions" && r.Method == http.MethodPost {
			runtimeHandler.EnablePHPExtension(w, r)
			return
		}
		if len(parts) >= 9 && parts[len(parts)-2] == "extensions" && r.Method == http.MethodDelete {
			runtimeHandler.DisablePHPExtension(w, r)
			return
		}
		switch r.Method {
		case http.MethodGet:
			runtimeHandler.GetPHPPool(w, r)
		case http.MethodPut, http.MethodPatch:
			runtimeHandler.UpdatePHPPool(w, r)
		case http.MethodDelete:
			runtimeHandler.DeletePHPPool(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	mux.Handle("/api/v1/runtime/nodejs/apps", authWrap(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			runtimeHandler.CreateNodeJSApp(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	mux.Handle("/api/v1/runtime/nodejs/apps/", authWrap(func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(strings.TrimSuffix(r.URL.Path, "/"), "/")
		if len(parts) >= 8 {
			action := parts[len(parts)-1]
			if action == "start" && r.Method == http.MethodPost {
				runtimeHandler.StartNodeJSApp(w, r)
				return
			}
			if action == "stop" && r.Method == http.MethodPost {
				runtimeHandler.StopNodeJSApp(w, r)
				return
			}
			if action == "restart" && r.Method == http.MethodPost {
				runtimeHandler.RestartNodeJSApp(w, r)
				return
			}
		}
		switch r.Method {
		case http.MethodGet:
			runtimeHandler.GetNodeJSApp(w, r)
		case http.MethodDelete:
			runtimeHandler.DeleteNodeJSApp(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	mux.Handle("/api/v1/runtime/python/apps", authWrap(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			runtimeHandler.CreatePythonApp(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	mux.Handle("/api/v1/runtime/python/apps/", authWrap(func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(strings.TrimSuffix(r.URL.Path, "/"), "/")
		if len(parts) >= 8 && parts[len(parts)-1] == "virtualenv" && r.Method == http.MethodPost {
			runtimeHandler.ProvisionVirtualenv(w, r)
			return
		}
		switch r.Method {
		case http.MethodGet:
			runtimeHandler.GetPythonApp(w, r)
		case http.MethodDelete:
			runtimeHandler.DeletePythonApp(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	// Logging endpoints (admin only)
	mux.Handle("/api/v1/logging/logs", adminWrap(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet, http.MethodPost:
			if r.Method == http.MethodPost && r.URL.Query().Get("action") != "query" {
				loggingHandler.CreateLog(w, r)
			} else {
				loggingHandler.QueryLogs(w, r)
			}
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	mux.Handle("/api/v1/logging/audits", adminWrap(loggingHandler.QueryAudits))
	mux.Handle("/api/v1/logging/metrics", adminWrap(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			loggingHandler.ListMetrics(w, r)
		case http.MethodPost:
			loggingHandler.RecordMetric(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	mux.Handle("/api/v1/logging/counters", adminWrap(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			loggingHandler.IncrementCounter(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	mux.Handle("/api/v1/logging/gauges", adminWrap(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			loggingHandler.SetGauge(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	mux.Handle("/api/v1/logging/prometheus", http.HandlerFunc(loggingHandler.ExportPrometheus))

	// Plugin endpoints (admin only)
	mux.Handle("/api/v1/plugins", adminWrap(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			pluginHandler.List(w, r)
		case http.MethodPost:
			pluginHandler.Install(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	mux.Handle("/api/v1/plugins/scopes", adminWrap(pluginHandler.ListScopes))
	mux.Handle("/api/v1/plugins/hooks/", adminWrap(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			pluginHandler.UnregisterHook(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	mux.Handle("/api/v1/plugins/", adminWrap(func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(strings.TrimSuffix(r.URL.Path, "/"), "/")
		if len(parts) >= 6 {
			action := parts[len(parts)-1]
			switch action {
			case "activate":
				if r.Method == http.MethodPost {
					pluginHandler.Activate(w, r)
					return
				}
			case "deactivate":
				if r.Method == http.MethodPost {
					pluginHandler.Deactivate(w, r)
					return
				}
			case "configure":
				if r.Method == http.MethodPut || r.Method == http.MethodPatch {
					pluginHandler.Configure(w, r)
					return
				}
			case "scopes":
				if r.Method == http.MethodPost {
					pluginHandler.GrantScope(w, r)
					return
				}
			case "hooks":
				if r.Method == http.MethodGet {
					pluginHandler.ListHooks(w, r)
					return
				}
				if r.Method == http.MethodPost {
					pluginHandler.RegisterHook(w, r)
					return
				}
			}
			// Handle scope revocation: /api/v1/plugins/{id}/scopes/{scope}
			if len(parts) >= 7 && parts[len(parts)-2] == "scopes" && r.Method == http.MethodDelete {
				pluginHandler.RevokeScope(w, r)
				return
			}
		}
		switch r.Method {
		case http.MethodGet:
			pluginHandler.Get(w, r)
		case http.MethodDelete:
			pluginHandler.Uninstall(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	// WebSocket endpoint
	mux.Handle("/api/v1/ws", s.wsHub.Handler(func(r *http.Request) string {
		// Extract user ID from token
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			return ""
		}
		if strings.HasPrefix(authHeader, "Bearer ") {
			token := strings.TrimPrefix(authHeader, "Bearer ")
			claims, err := s.authService.ValidateToken(token)
			if err != nil {
				return ""
			}
			return claims.UserID
		}
		return ""
	}))

	// Apply middleware
	var handler http.Handler = mux
	handler = middleware.RequestIDMiddleware(handler)
	handler = middleware.LoggingMiddleware(s.loggingService)(handler)
	handler = metrics.MetricsMiddleware(s.metricsService)(handler)
	handler = middleware.UserRateLimitMiddleware(s.rateLimiter)(handler)
	handler = middleware.CORSMiddleware(handler)
	handler = middleware.ContentTypeMiddleware(handler)
	handler = middleware.RecoveryMiddleware(s.loggingService)(handler)

	return handler
}

// Start starts all servers
func (s *Server) Start() error {
	apiHandler := s.setupAPIRoutes()

	s.api = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", s.config.Server.Host, s.config.Server.Port),
		Handler:      apiHandler,
		ReadTimeout:  s.config.Server.ReadTimeout,
		WriteTimeout: s.config.Server.WriteTimeout,
	}

	s.loggingService.Info("server", fmt.Sprintf("Starting OweHost API server on %s:%d", s.config.Server.Host, s.config.Server.Port))
	s.loggingService.Info("server", "User Panel frontend should run on port 2083")
	s.loggingService.Info("server", "Admin Panel frontend should run on port 2087")

	return s.api.ListenAndServe()
}

// Shutdown gracefully shuts down all servers
func (s *Server) Shutdown(ctx context.Context) error {
	s.loggingService.Info("server", "Shutting down server...")

	var err error
	if s.api != nil {
		err = s.api.Shutdown(ctx)
	}

	if s.db != nil {
		s.db.Close()
	}

	return err
}

// setupUserPanelRoutes sets up User Panel routes (Port 2083 - cPanel-like)
func (s *Server) setupUserPanelRoutes() http.Handler {
	mux := http.NewServeMux()

	// Serve static files for user panel
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head><title>OweHost User Panel</title></head>
<body>
<h1>OweHost User Panel</h1>
<p>User control panel - Port 2083</p>
<p>cPanel-like interface for individual users</p>
</body>
</html>`)
	})

	var handler http.Handler = mux
	handler = middleware.CORSMiddleware(handler)
	handler = middleware.LoggingMiddleware(s.loggingService)(handler)

	return handler
}

// setupAdminPanelRoutes sets up Admin Panel routes (Port 2087 - WHM-like)
func (s *Server) setupAdminPanelRoutes() http.Handler {
	mux := http.NewServeMux()

	// Create handlers
	authHandler := v1.NewAuthHandler(s.authService, s.userService)
	userHandler := v1.NewUserHandler(s.userService)
	domainHandler := v1.NewDomainHandler(s.domainService, s.userService)
	databaseHandler := v1.NewDatabaseHandler(s.databaseService, s.userService)
	healthHandler := v1.NewHealthHandler(s.loggingService)
	installationHandler := v1.NewInstallationHandler(s.installationService)

	// Installation endpoints
	mux.HandleFunc("/api/v1/installation/check", installationHandler.Check)
	mux.HandleFunc("/api/v1/installation/install", installationHandler.Install)

	// Health endpoints
	mux.HandleFunc("/health", healthHandler.Health)

	// Auth endpoints
	mux.HandleFunc("/api/v1/auth/login", authHandler.Login)
	mux.HandleFunc("/api/v1/auth/refresh", authHandler.Refresh)
	mux.HandleFunc("/api/v1/auth/logout", authHandler.Logout)

	// Admin endpoints
	mux.HandleFunc("/api/v1/users", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			userHandler.List(w, r)
		case http.MethodPost:
			userHandler.Create(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/v1/users/me", userHandler.Me)
	mux.HandleFunc("/api/v1/domains", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			domainHandler.List(w, r)
		case http.MethodPost:
			domainHandler.Create(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/v1/databases", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			databaseHandler.List(w, r)
		case http.MethodPost:
			databaseHandler.Create(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	var handler http.Handler = mux
	handler = middleware.RequestIDMiddleware(handler)
	handler = middleware.LoggingMiddleware(s.loggingService)(handler)
	handler = middleware.CORSMiddleware(handler)
	handler = middleware.ContentTypeMiddleware(handler)
	handler = middleware.RecoveryMiddleware(s.loggingService)(handler)

	return handler
}

func main() {
	// Load configuration
	cfg := config.Load()

	// Create and start server
	server := NewServer(cfg)

	// Handle graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Printf("Server shutdown error: %v", err)
		}
	}()

	// Start server
	log.Printf("Starting OweHost API server on %s:%d", cfg.Server.Host, cfg.Server.Port)
	if err := server.Start(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}
}
