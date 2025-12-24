// Package main provides the OweHost CLI tool for administrative operations
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/iSundram/OweHost/internal/storage/recovery"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "scan":
		cmdScan(os.Args[2:])
	case "rebuild":
		cmdRebuild(os.Args[2:])
	case "generate":
		cmdGenerate(os.Args[2:])
	case "verify":
		cmdVerify(os.Args[2:])
	case "cleanup":
		cmdCleanup(os.Args[2:])
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`OweHost CLI - Administrative Tool

Usage:
  owehost-cli <command> [options]

Commands:
  scan      Scan filesystem for accounts and resources
  rebuild   Rebuild database from filesystem state
  generate  Regenerate service configurations from filesystem
  verify    Verify consistency between filesystem and database
  cleanup   Clean up stale configurations
  help      Show this help message

Use "owehost-cli <command> -h" for more information about a command.`)
}

// cmdScan scans the filesystem for accounts
func cmdScan(args []string) {
	fs := flag.NewFlagSet("scan", flag.ExitOnError)
	accountID := fs.Int("account", 0, "Scan specific account ID (0 for all)")
	outputJSON := fs.Bool("json", false, "Output as JSON")
	verbose := fs.Bool("verbose", false, "Verbose output")
	fs.Parse(args)

	scanner := recovery.NewScanner()

	var result *recovery.ScanResult
	var err error

	if *accountID > 0 {
		scan, err := scanner.ScanAccount(*accountID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error scanning account %d: %v\n", *accountID, err)
			os.Exit(1)
		}
		result = &recovery.ScanResult{
			Accounts:   []recovery.AccountScan{*scan},
			TotalSites: len(scan.Sites),
			TotalSSL:   len(scan.SSLCerts),
			TotalDBs:   len(scan.Databases),
		}
	} else {
		result, err = scanner.ScanAll()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error scanning filesystem: %v\n", err)
			os.Exit(1)
		}
	}

	if *outputJSON {
		data, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(data))
		return
	}

	// Text output
	fmt.Printf("Scan Results:\n")
	fmt.Printf("  Accounts found: %d\n", len(result.Accounts))
	fmt.Printf("  Total sites:    %d\n", result.TotalSites)
	fmt.Printf("  Total SSL:      %d\n", result.TotalSSL)
	fmt.Printf("  Total DBs:      %d\n", result.TotalDBs)
	fmt.Printf("  Errors:         %d\n", len(result.Errors))

	if *verbose {
		fmt.Println("\nAccounts:")
		for _, acc := range result.Accounts {
			name := "unknown"
			state := "unknown"
			if acc.Identity != nil {
				name = acc.Identity.Name
				state = acc.Identity.State
			}
			fmt.Printf("  - a-%d: %s (state: %s, sites: %d)\n",
				acc.ID, name, state, len(acc.Sites))
		}
	}

	if len(result.Errors) > 0 {
		fmt.Println("\nErrors:")
		for _, err := range result.Errors {
			fmt.Printf("  - [account %d] %s: %s\n", err.AccountID, err.Resource, err.Error)
		}
	}
}

// cmdRebuild rebuilds database from filesystem
func cmdRebuild(args []string) {
	fs := flag.NewFlagSet("rebuild", flag.ExitOnError)
	accountID := fs.Int("account", 0, "Rebuild specific account ID (0 for all)")
	dryRun := fs.Bool("dry-run", false, "Preview changes without applying")
	skipValidation := fs.Bool("skip-validation", false, "Skip validation checks")
	outputJSON := fs.Bool("json", false, "Output as JSON")
	fs.Parse(args)

	rebuilder := recovery.NewRebuilder()

	opts := recovery.RebuildOptions{
		DryRun:         *dryRun,
		SkipValidation: *skipValidation,
	}
	if *accountID > 0 {
		opts.AccountID = accountID
	}

	result, err := rebuilder.Rebuild(context.Background(), opts, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error rebuilding: %v\n", err)
		os.Exit(1)
	}

	if *outputJSON {
		data, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(data))
		return
	}

	fmt.Printf("Rebuild Results:\n")
	fmt.Printf("  Duration:         %v\n", result.Duration)
	fmt.Printf("  Accounts scanned: %d\n", result.AccountsScanned)
	fmt.Printf("  Accounts updated: %d\n", result.AccountsUpdated)
	fmt.Printf("  Sites found:      %d\n", result.SitesFound)
	fmt.Printf("  SSL certs found:  %d\n", result.SSLCertsFound)
	fmt.Printf("  Databases found:  %d\n", result.DatabasesFound)
	fmt.Printf("  Errors:           %d\n", len(result.Errors))
	fmt.Printf("  Warnings:         %d\n", len(result.Warnings))

	if *dryRun {
		fmt.Println("\n[DRY RUN] No changes were made.")
	}

	if len(result.Errors) > 0 {
		fmt.Println("\nErrors:")
		for _, err := range result.Errors {
			fmt.Printf("  - [account %d] %s: %s\n", err.AccountID, err.Resource, err.Error)
		}
	}

	if len(result.Warnings) > 0 {
		fmt.Println("\nWarnings:")
		for _, w := range result.Warnings {
			fmt.Printf("  - %s\n", w)
		}
	}
}

// cmdGenerate regenerates service configurations
func cmdGenerate(args []string) {
	fs := flag.NewFlagSet("generate", flag.ExitOnError)
	accountID := fs.Int("account", 0, "Generate for specific account ID (0 for all)")
	dryRun := fs.Bool("dry-run", false, "Preview changes without applying")
	skipNginx := fs.Bool("skip-nginx", false, "Skip nginx configuration")
	skipPHPFpm := fs.Bool("skip-phpfpm", false, "Skip PHP-FPM pool configuration")
	skipUsers := fs.Bool("skip-users", false, "Skip system user creation")
	outputJSON := fs.Bool("json", false, "Output as JSON")
	fs.Parse(args)

	generator := recovery.NewGenerator()

	opts := recovery.GenerateOptions{
		DryRun:     *dryRun,
		SkipNginx:  *skipNginx,
		SkipPHPFpm: *skipPHPFpm,
		SkipUsers:  *skipUsers,
	}
	if *accountID > 0 {
		opts.AccountID = accountID
	}

	result, err := generator.GenerateAll(opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating configs: %v\n", err)
		os.Exit(1)
	}

	if *outputJSON {
		data, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(data))
		return
	}

	fmt.Printf("Generate Results:\n")
	fmt.Printf("  Nginx configs:  %d\n", result.NginxConfigs)
	fmt.Printf("  PHP-FPM pools:  %d\n", result.PHPFpmPools)
	fmt.Printf("  System users:   %d\n", result.SystemUsers)
	fmt.Printf("  Cgroups:        %d\n", result.Cgroups)
	fmt.Printf("  Errors:         %d\n", len(result.Errors))

	if *dryRun {
		fmt.Println("\n[DRY RUN] No changes were made.")
	}

	if len(result.Errors) > 0 {
		fmt.Println("\nErrors:")
		for _, e := range result.Errors {
			fmt.Printf("  - %s\n", e)
		}
	}
}

// cmdVerify verifies consistency
func cmdVerify(args []string) {
	fs := flag.NewFlagSet("verify", flag.ExitOnError)
	outputJSON := fs.Bool("json", false, "Output as JSON")
	fs.Parse(args)

	scanner := recovery.NewScanner()
	result, err := scanner.ScanAll()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error scanning: %v\n", err)
		os.Exit(1)
	}

	var issues []string
	for _, acc := range result.Accounts {
		accIssues := scanner.ValidateIntegrity(&acc)
		for _, issue := range accIssues {
			issues = append(issues, fmt.Sprintf("account %d: %s", acc.ID, issue))
		}
	}

	if *outputJSON {
		data, _ := json.MarshalIndent(map[string]interface{}{
			"accounts": len(result.Accounts),
			"issues":   issues,
			"valid":    len(issues) == 0,
		}, "", "  ")
		fmt.Println(string(data))
		return
	}

	fmt.Printf("Verification Results:\n")
	fmt.Printf("  Accounts checked: %d\n", len(result.Accounts))
	fmt.Printf("  Issues found:     %d\n", len(issues))

	if len(issues) > 0 {
		fmt.Println("\nIssues:")
		for _, issue := range issues {
			fmt.Printf("  - %s\n", issue)
		}
		os.Exit(1)
	} else {
		fmt.Println("\n✓ All accounts validated successfully.")
	}
}

// cmdCleanup cleans up stale configurations
func cmdCleanup(args []string) {
	fs := flag.NewFlagSet("cleanup", flag.ExitOnError)
	dryRun := fs.Bool("dry-run", false, "Preview changes without applying")
	outputJSON := fs.Bool("json", false, "Output as JSON")
	fs.Parse(args)

	generator := recovery.NewGenerator()

	if *dryRun {
		fmt.Println("[DRY RUN] Would scan for stale configurations...")
		return
	}

	removed, err := generator.CleanupStaleConfigs()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error cleaning up: %v\n", err)
		os.Exit(1)
	}

	if *outputJSON {
		data, _ := json.MarshalIndent(map[string]interface{}{
			"removed": removed,
			"count":   len(removed),
		}, "", "  ")
		fmt.Println(string(data))
		return
	}

	fmt.Printf("Cleanup Results:\n")
	fmt.Printf("  Stale configs removed: %d\n", len(removed))

	if len(removed) > 0 {
		fmt.Println("\nRemoved:")
		for _, r := range removed {
			fmt.Printf("  - %s\n", r)
		}
	}
}
