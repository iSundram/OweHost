package main

import (
"context"
"fmt"
"log"

"github.com/iSundram/OweHost/internal/accountsvc"
"github.com/iSundram/OweHost/internal/storage/web"
)

func main() {
svc := accountsvc.NewService()

// Add domain admini.tech to account 10001
site := &web.SiteDescriptor{
Domain:       "admini.tech",
Runtime:      "php-8.2",
SSL:          true,
SSLRedirect:  true,
DocumentRoot: "public",
}

err := svc.AddDomain(context.Background(), 10001, site, "admin", "admin")
if err != nil {
log.Fatalf("Failed to add domain: %v", err)
}

fmt.Printf("✅ Domain added: admini.tech\n")

// Verify - list domains
domains, err := svc.ListDomains(context.Background(), 10001)
if err != nil {
log.Fatalf("Failed to list domains: %v", err)
}
fmt.Printf("   Domains: %d\n", len(domains))
for _, d := range domains {
fmt.Printf("   - %s (runtime: %s, ssl: %v)\n", d.Domain, d.Runtime, d.SSL)
}
}
