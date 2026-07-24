// Command proxyconnection builds proxy connection URLs with the ProxyURL
// helper and uses one to configure an *http.Client.
package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"

	webshare "github.com/webshare-proxy/webshare-go"
)

func main() {
	// Direct mode: connect straight to a proxy from the proxy list API.
	direct, err := webshare.ProxyURL(webshare.ProxyURLParams{
		Mode:     webshare.ModeDirect,
		Username: "myuser",
		Password: "mypassword",
		Address:  "203.0.113.10",
		Port:     8168,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("direct:", direct)

	// Backbone mode with a sticky session pinned to US exits.
	sticky, err := webshare.ProxyURL(webshare.ProxyURLParams{
		Mode:         webshare.ModeBackbone,
		Username:     "myuser",
		Password:     "mypassword",
		CountryCodes: []string{"US"},
		SessionID:    "1234",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("backbone sticky:", sticky)

	// Backbone mode rotating to a new IP on every request.
	rotate, err := webshare.ProxyURL(webshare.ProxyURLParams{
		Mode:         webshare.ModeBackbone,
		Username:     "myuser",
		Password:     "mypassword",
		CountryCodes: []string{"DE"},
		City:         "munich",
		Rotate:       true,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("backbone rotate:", rotate)

	// Use a proxy URL with the standard library HTTP client.
	proxyURL, err := url.Parse(sticky)
	if err != nil {
		log.Fatal(err)
	}
	httpClient := &http.Client{
		Transport: &http.Transport{Proxy: http.ProxyURL(proxyURL)},
	}
	_ = httpClient // httpClient now routes requests through the proxy
}
