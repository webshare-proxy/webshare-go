// Command listproxies lists the first page of proxies, then iterates the
// entire proxy list using automatic pagination.
package main

import (
	"context"
	"fmt"
	"log"

	webshare "github.com/webshare-proxy/webshare-go"
)

func main() {
	// Reads WEBSHARE_API_KEY from the environment.
	client, err := webshare.NewClient()
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()

	// One page at a time.
	page, err := client.Proxies.List(ctx, webshare.ProxyListParams{
		Mode:     webshare.ModeDirect,
		PageSize: webshare.Int(10),
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("account has %d proxies; first page:\n", page.Count)
	for _, proxy := range page.Results {
		fmt.Printf("  %s:%d (%s, valid=%t)\n", *proxy.ProxyAddress, proxy.Port, proxy.CountryCode, proxy.Valid)
	}

	// Or iterate every proxy across all pages lazily.
	total := 0
	for proxy, err := range client.Proxies.ListAll(ctx, webshare.ProxyListParams{Mode: webshare.ModeDirect}) {
		if err != nil {
			log.Fatal(err)
		}
		total++
		_ = proxy
	}
	fmt.Printf("iterated %d proxies in total\n", total)
}
