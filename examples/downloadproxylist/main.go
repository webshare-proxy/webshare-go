// Command downloadproxylist fetches the proxy list download token from the
// proxy config API and downloads the proxy list as plain text.
package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	webshare "github.com/webshare-proxy/webshare-go"
)

func main() {
	client, err := webshare.NewClient()
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()

	// Pick the plan to download the list for: proxy operations are
	// plan-scoped, so list the plans and select one.
	plans, err := client.Plans.List(ctx, webshare.PlanListParams{})
	if err != nil {
		log.Fatal(err)
	}
	planID := 0
	for _, plan := range plans.Results {
		if plan.Status == webshare.PlanActive {
			planID = plan.ID
			break
		}
	}
	if planID == 0 {
		log.Fatal("no active plan found")
	}

	// The download token lives on that plan's proxy config.
	config, err := client.ProxyConfig.Get(ctx, planID)
	if err != nil {
		log.Fatal(err)
	}

	params := webshare.ProxyDownloadParams{
		Token:                config.ProxyListDownloadToken,
		AuthenticationMethod: webshare.AuthMethodUsername,
		EndpointMode:         webshare.ModeDirect,
		PlanID:               webshare.Int(planID),
	}

	// The URL alone can be shared with tools that fetch the list directly.
	downloadURL, err := client.Proxies.DownloadURL(params)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("download URL:", downloadURL)

	// Or download the list in place: one address:port:username:password
	// line per proxy.
	list, err := client.Proxies.Download(ctx, params)
	if err != nil {
		log.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(list), "\n")
	fmt.Printf("downloaded %d proxies\n", len(lines))
	for i, line := range lines {
		if i == 3 {
			fmt.Println("  ...")
			break
		}
		fmt.Println(" ", line)
	}
}
