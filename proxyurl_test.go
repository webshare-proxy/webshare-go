package webshare

import (
	"strings"
	"testing"
)

func TestProxyURL(t *testing.T) {
	tests := []struct {
		name    string
		params  ProxyURLParams
		want    string
		wantErr string
	}{
		{
			name: "direct with credentials",
			params: ProxyURLParams{
				Mode:     ModeDirect,
				Username: "user",
				Password: "pass",
				Address:  "1.2.3.4",
				Port:     8168,
			},
			want: "http://user:pass@1.2.3.4:8168",
		},
		{
			name: "direct ip authorization has no credentials",
			params: ProxyURLParams{
				Mode:    ModeDirect,
				Address: "1.2.3.4",
				Port:    8168,
			},
			want: "http://1.2.3.4:8168",
		},
		{
			name: "backbone defaults host and port",
			params: ProxyURLParams{
				Mode:     ModeBackbone,
				Username: "myuser",
				Password: "password",
			},
			want: "http://myuser:password@p.webshare.io:80",
		},
		{
			name: "backbone with session",
			params: ProxyURLParams{
				Mode:         ModeBackbone,
				Username:     "myuser",
				Password:     "password",
				CountryCodes: []string{"US"},
				SessionID:    "1234",
			},
			want: "http://myuser-us-1234:password@p.webshare.io:80",
		},
		{
			name: "backbone with city and rotate",
			params: ProxyURLParams{
				Mode:         ModeBackbone,
				Username:     "myuser",
				Password:     "password",
				CountryCodes: []string{"us"},
				City:         "los_angeles",
				Rotate:       true,
			},
			want: "http://myuser-us-city_los_angeles-rotate:password@p.webshare.io:80",
		},
		{
			name: "backbone parameter ordering country city session",
			params: ProxyURLParams{
				Mode:         ModeBackbone,
				Username:     "myuser",
				Password:     "password",
				CountryCodes: []string{"de"},
				City:         "munich",
				SessionID:    "1234",
			},
			want: "http://myuser-de-city_munich-1234:password@p.webshare.io:80",
		},
		{
			name: "backbone multiple countries",
			params: ProxyURLParams{
				Mode:         ModeBackbone,
				Username:     "myuser",
				Password:     "password",
				CountryCodes: []string{"US", "FR", "DE"},
			},
			want: "http://myuser-us-fr-de:password@p.webshare.io:80",
		},
		{
			name: "custom scheme and port",
			params: ProxyURLParams{
				Mode:     ModeBackbone,
				Scheme:   "socks5",
				Username: "myuser",
				Password: "password",
				Port:     1080,
			},
			want: "socks5://myuser:password@p.webshare.io:1080",
		},
		{
			name: "password is escaped",
			params: ProxyURLParams{
				Mode:     ModeBackbone,
				Username: "myuser",
				Password: "p@ss/word",
			},
			want: "http://myuser:p%40ss%2Fword@p.webshare.io:80",
		},
		{
			name:    "missing mode",
			params:  ProxyURLParams{Username: "u", Password: "p"},
			wantErr: "mode is required",
		},
		{
			name:    "invalid mode",
			params:  ProxyURLParams{Mode: "tunnel", Username: "u", Password: "p"},
			wantErr: "invalid mode",
		},
		{
			name:    "direct requires address",
			params:  ProxyURLParams{Mode: ModeDirect, Username: "u", Password: "p", Port: 80},
			wantErr: "address is required",
		},
		{
			name:    "direct requires port",
			params:  ProxyURLParams{Mode: ModeDirect, Username: "u", Password: "p", Address: "1.2.3.4"},
			wantErr: "port is required",
		},
		{
			name: "direct rejects backbone username params",
			params: ProxyURLParams{
				Mode: ModeDirect, Username: "u", Password: "p",
				Address: "1.2.3.4", Port: 80, CountryCodes: []string{"us"},
			},
			wantErr: "only supported in backbone mode",
		},
		{
			name: "session and rotate are mutually exclusive",
			params: ProxyURLParams{
				Mode: ModeBackbone, Username: "u", Password: "p",
				SessionID: "1", Rotate: true,
			},
			wantErr: "mutually exclusive",
		},
		{
			name: "session must be numeric",
			params: ProxyURLParams{
				Mode: ModeBackbone, Username: "u", Password: "p",
				SessionID: "abc",
			},
			wantErr: "must be numeric",
		},
		{
			name: "city must be letters and underscores",
			params: ProxyURLParams{
				Mode: ModeBackbone, Username: "u", Password: "p",
				City: "san jose",
			},
			wantErr: "letters and underscores",
		},
		{
			name: "invalid country code",
			params: ProxyURLParams{
				Mode: ModeBackbone, Username: "u", Password: "p",
				CountryCodes: []string{"usa"},
			},
			wantErr: "invalid country code",
		},
		{
			name:    "username without password",
			params:  ProxyURLParams{Mode: ModeBackbone, Username: "u"},
			wantErr: "password is required",
		},
		{
			name:    "password without username",
			params:  ProxyURLParams{Mode: ModeBackbone, Password: "p"},
			wantErr: "username is required",
		},
		{
			name: "ip auth cannot use username params",
			params: ProxyURLParams{
				Mode:         ModeBackbone,
				CountryCodes: []string{"us"},
			},
			wantErr: "require username/password",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ProxyURL(tt.params)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("err = %v, want error containing %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("ProxyURL: %v", err)
			}
			if got != tt.want {
				t.Errorf("ProxyURL = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestProxyDownloadURL(t *testing.T) {
	client, err := NewClient(WithAPIKey("k"))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	t.Run("all defaults", func(t *testing.T) {
		got, err := client.Proxies.DownloadURL(ProxyDownloadParams{
			Token:                "tok123",
			AuthenticationMethod: AuthMethodUsername,
			EndpointMode:         ModeDirect,
		})
		if err != nil {
			t.Fatalf("DownloadURL: %v", err)
		}
		want := "https://proxy.webshare.io/api/v2/proxy/list/download/tok123/-/any/username/direct/-/"
		if got != want {
			t.Errorf("DownloadURL = %q, want %q", got, want)
		}
	})

	t.Run("countries search and plan", func(t *testing.T) {
		got, err := client.Proxies.DownloadURL(ProxyDownloadParams{
			Token:                "tok123",
			CountryCodes:         []string{"US", "FR"},
			AuthenticationMethod: AuthMethodSourceIP,
			EndpointMode:         ModeBackbone,
			Search:               "new york",
			PlanID:               Int(7),
		})
		if err != nil {
			t.Fatalf("DownloadURL: %v", err)
		}
		want := "https://proxy.webshare.io/api/v2/proxy/list/download/tok123/US-FR/any/sourceip/backbone/new%20york/?plan_id=7"
		if got != want {
			t.Errorf("DownloadURL = %q, want %q", got, want)
		}
	})

	t.Run("missing token", func(t *testing.T) {
		_, err := client.Proxies.DownloadURL(ProxyDownloadParams{
			AuthenticationMethod: AuthMethodUsername,
			EndpointMode:         ModeDirect,
		})
		if err == nil {
			t.Fatal("expected error for missing token")
		}
	})

	t.Run("missing authentication method", func(t *testing.T) {
		_, err := client.Proxies.DownloadURL(ProxyDownloadParams{
			Token:        "tok123",
			EndpointMode: ModeDirect,
		})
		if err == nil {
			t.Fatal("expected error for missing authentication method")
		}
	})
}
