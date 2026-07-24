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

	// The download token lives on the proxy config; the config endpoint
	// needs the plan ID from the subscription.
	subscription, err := client.Subscription.Get(ctx)
	if err != nil {
		log.Fatal(err)
	}
	config, err := client.ProxyConfig.Get(ctx, subscription.Plan)
	if err != nil {
		log.Fatal(err)
	}

	params := webshare.ProxyDownloadParams{
		Token:                config.ProxyListDownloadToken,
		AuthenticationMethod: webshare.AuthMethodUsername,
		EndpointMode:         webshare.ModeDirect,
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
