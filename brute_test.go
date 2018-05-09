package brute

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

const googleDNS = "8.8.8.8:53"

func TestBrute(t *testing.T) {
	tt := []struct {
		name     string
		wordlist []string
		cfg      Bruter
		exp      []string
	}{
		{
			name:     "valid 1 worker 1 domain",
			wordlist: []string{"maps"},
			cfg: Bruter{
				Domain:  "google.com",
				Retries: 1,
				Record:  "A",
				Servers: []string{googleDNS},
				Workers: 1,
			},
			exp: []string{"maps"},
		},
		{
			name:     "2 valid domains 1 invalid",
			wordlist: []string{"maps", "docs", "gopherpleasedonotexist"},
			cfg: Bruter{
				Domain:  "google.com",
				Retries: 0,
				Record:  "A",
				Servers: []string{googleDNS},
				Workers: 1,
			},
			exp: []string{"maps", "docs"},
		},
		{
			name:     "wildcard",
			wordlist: []string{"maps"},
			cfg: Bruter{
				Domain:  "myshopify.com",
				Retries: 0,
				Record:  "A",
				Servers: []string{googleDNS},
				Workers: 1,
			},
			exp: nil,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			var b strings.Builder
			for _, v := range tc.wordlist {
				if _, err := b.WriteString(fmt.Sprintf("%s\n", v)); err != nil {
					t.Fatal(err)
				}
			}
			r := strings.NewReader(b.String())

			out := make(chan string)
			done := make(chan struct{})
			var res []string
			go func() {
				for v := range out {
					res = append(res, v)
				}
				done <- struct{}{}
			}()

			err := tc.cfg.Brute(context.Background(), r, out)
			if err != nil && tc.exp != nil {
				t.Fatal(err)
			}
			<-done

			if !equal(t, tc.exp, res) {
				t.Fatalf("expected %v got %v", tc.exp, res)
			}
		})
	}
}

func TestIsWildcard(t *testing.T) {
	tt := []struct {
		name   string
		domain string
		srv    string
		exp    bool
	}{
		{
			name:   "non wildcard domain",
			domain: "bugcrowd.com",
			srv:    "8.8.8.8:53",
			exp:    false,
		},
		{
			name:   "wildcard domain",
			domain: "myshopify.com",
			srv:    "8.8.8.8:53",
			exp:    true,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			out := IsWildcard(context.Background(), tc.domain, tc.srv)
			if out != tc.exp {
				t.Fatalf("expected %v got %v", tc.exp, out)
			}
		})
	}
}

func equal(t *testing.T, a, b []string) bool {
	t.Helper()
	m1 := make(map[string]struct{}, len(a))
	for _, v := range a {
		m1[v] = struct{}{}
	}

	m2 := make(map[string]struct{}, len(b))
	for _, v := range b {
		m2[v] = struct{}{}
	}

	return reflect.DeepEqual(m1, m2)
}
