package main

import (
"context"
"fmt"
"log"

"github.com/iSundram/OweHost/internal/accountsvc"
"github.com/iSundram/OweHost/internal/storage/web"
)

func main() {
// Create service
svc := accountsvc.NewService()

// Create account for sundram
resp, err := svc.Create(context.Background(), &accountsvc.CreateRequest{
Username: "sundram",
Email:    "sundram@example.com",
Plan:     "standard",
Owner:    "admin",
}, "admin", "admin", "127.0.0.1")

if err != nil {
log.Fatalf("Failed to create account: %v", err)
}

fmt.Printf("✅ Account created!\n")
fmt.Printf("   Account ID: %d\n", resp.AccountID)
fmt.Printf("   Username: %s\n", resp.Identity.Name)
fmt.Printf("   UID/GID: %d/%d\n", resp.Identity.UID, resp.Identity.GID)

// Add domain admini.tech
site := &web.SiteDescriptor{
Domain:       "admini.tech",
Runtime:      "php-8.2",
SSL:          true,
SSLRedirect:  true,
DocumentRoot: "public",
}

err = svc.AddDomain(context.Background(), resp.AccountID, site, "admin", "admin")
if err != nil {
log.Fatalf("Failed to add domain: %v", err)
}

fmt.Printf("✅ Domain added: admini.tech\n")

// Verify - list domains
domains, err := svc.ListDomains(context.Background(), resp.AccountID)
if err != nil {
log.Fatalf("Failed to list domains: %v", err)
}
fmt.Printf("   Domains: %d\n", len(domains))
for _, d := range domains {
fmt.Printf("   - %s (runtime: %s)\n", d.Domain, d.Runtime)
}
}
