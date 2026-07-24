// Command errorhandling demonstrates how to inspect API and transport errors
// returned by the SDK.
package main

import (
	"context"
	"errors"
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

	// Force a validation error: the proxy list requires a mode parameter.
	_, err = client.Proxies.List(ctx, webshare.ProxyListParams{})
	if err == nil {
		fmt.Println("unexpectedly succeeded")
		return
	}

	var apiErr *webshare.Error
	var reqErr *webshare.RequestError
	switch {
	case errors.As(err, &apiErr):
		fmt.Printf("API error: status=%d code=%q request_id=%q\n", apiErr.StatusCode, apiErr.Code, apiErr.RequestID)
		fmt.Printf("detail: %s\n", apiErr.Detail)
		for field, messages := range apiErr.FieldErrors {
			fmt.Printf("field %q: %v\n", field, messages)
		}
		switch apiErr.Code {
		case "account_suspended":
			fmt.Println("check client.Verification.GetSuspension for details")
		case "account_deleted":
			fmt.Println("the account was deleted; all API calls return this code")
		}
	case errors.As(err, &reqErr):
		fmt.Printf("transport error: %v\n", reqErr.Unwrap())
	default:
		fmt.Printf("error: %v\n", err)
	}
}
