// Command plans lists the account's proxy plans, selects one, and uses its
// ID across plan-scoped calls: the proxy list, the proxy configuration and
// the proxy list download URL.
package main

import (
	"context"
	"fmt"
	"log"

	webshare "github.com/webshare-proxy/webshare-go"
)

func main() {
	client, err := webshare.NewClient()
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()

	// 1. List the plans and pick one. The plan ID is also visible in the
	// dashboard URL when viewing a plan.
	plans, err := client.Plans.List(ctx, webshare.PlanListParams{})
	if err != nil {
		log.Fatal(err)
	}
	var plan *webshare.Plan
	for i := range plans.Results {
		fmt.Printf("plan %d: %s %s/%s, %d proxies\n",
			plans.Results[i].ID, plans.Results[i].Status,
			plans.Results[i].ProxyType, plans.Results[i].ProxySubtype,
			plans.Results[i].ProxyCount)
		if plan == nil && plans.Results[i].Status == webshare.PlanActive {
			plan = &plans.Results[i]
		}
	}
	if plan == nil {
		log.Fatal("no active plan found")
	}
	fmt.Printf("using plan %d\n\n", plan.ID)

	// 2. List that plan's proxies.
	page, err := client.Proxies.List(ctx, webshare.ProxyListParams{
		Mode:     webshare.ModeDirect,
		PlanID:   webshare.Int(plan.ID),
		PageSize: webshare.Int(5),
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("plan has %d proxies\n", page.Count)

	// 3. Read the plan's proxy configuration.
	config, err := client.ProxyConfig.Get(ctx, plan.ID)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("request timeout: %ds\n", config.RequestTimeout)

	// 4. Build the plan-scoped proxy list download URL.
	downloadURL, err := client.Proxies.DownloadURL(webshare.ProxyDownloadParams{
		Token:                config.ProxyListDownloadToken,
		AuthenticationMethod: webshare.AuthMethodUsername,
		EndpointMode:         webshare.ModeDirect,
		PlanID:               webshare.Int(plan.ID),
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("download URL:", downloadURL)
}
