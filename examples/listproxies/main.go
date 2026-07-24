// Command listproxies lists the first page of a plan's proxies, then
// iterates the entire proxy list using automatic pagination.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	webshare "github.com/webshare-proxy/webshare-go"
)

// selectPlanID returns the plan to work with: WEBSHARE_PLAN_ID when set,
// otherwise the account's first active plan.
func selectPlanID(ctx context.Context, client *webshare.Client) (int, error) {
	if raw := os.Getenv("WEBSHARE_PLAN_ID"); raw != "" {
		return strconv.Atoi(raw)
	}
	plans, err := client.Plans.List(ctx, webshare.PlanListParams{})
	if err != nil {
		return 0, err
	}
	for _, plan := range plans.Results {
		if plan.Status == webshare.PlanActive {
			return plan.ID, nil
		}
	}
	return 0, fmt.Errorf("no active plan found")
}

func main() {
	// Reads WEBSHARE_API_KEY from the environment.
	client, err := webshare.NewClient()
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()

	planID, err := selectPlanID(ctx, client)
	if err != nil {
		log.Fatal(err)
	}

	// One page at a time.
	page, err := client.Proxies.List(ctx, webshare.ProxyListParams{
		Mode:     webshare.ModeDirect,
		PlanID:   webshare.Int(planID),
		PageSize: webshare.Int(10),
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("plan %d has %d proxies; first page:\n", planID, page.Count)
	for _, proxy := range page.Results {
		// ProxyAddress is nil on residential plans, which connect through
		// the p.webshare.io backbone instead of a direct address.
		address := webshare.BackboneHost
		if proxy.ProxyAddress != nil {
			address = *proxy.ProxyAddress
		}
		fmt.Printf("  %s:%d (%s, valid=%t)\n", address, proxy.Port, proxy.CountryCode, proxy.Valid)
	}

	// Or iterate every proxy across all pages lazily.
	total := 0
	for proxy, err := range client.Proxies.ListAll(ctx, webshare.ProxyListParams{
		Mode:   webshare.ModeDirect,
		PlanID: webshare.Int(planID),
	}) {
		if err != nil {
			log.Fatal(err)
		}
		total++
		_ = proxy
	}
	fmt.Printf("iterated %d proxies in total\n", total)
}
