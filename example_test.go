package webshare_test

import (
	"context"
	"fmt"
	"log"

	webshare "github.com/webshare-proxy/webshare-go"
)

func ExampleNewClient() {
	client, err := webshare.NewClient(webshare.WithAPIKey("your-api-key"))
	if err != nil {
		log.Fatal(err)
	}
	profile, err := client.Profile.Get(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(profile.Email)
}

func ExampleProxiesService_ListAll() {
	client, err := webshare.NewClient()
	if err != nil {
		log.Fatal(err)
	}
	for proxy, err := range client.Proxies.ListAll(context.Background(), webshare.ProxyListParams{Mode: webshare.ModeDirect}) {
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s (%s)\n", proxy.ID, proxy.CountryCode)
	}
}

func ExampleProxyURL() {
	proxyURL, err := webshare.ProxyURL(webshare.ProxyURLParams{
		Mode:         webshare.ModeBackbone,
		Username:     "myuser",
		Password:     "mypassword",
		CountryCodes: []string{"US"},
		SessionID:    "1234",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(proxyURL)
	// Output: http://myuser-us-1234:mypassword@p.webshare.io:80
}
