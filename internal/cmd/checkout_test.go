package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rtmx-ai/rtmx/internal/auth"
	"github.com/rtmx-ai/rtmx/internal/output"
	"github.com/rtmx-ai/rtmx/pkg/rtmx"
	"github.com/spf13/cobra"
)

func createCheckoutTestCmd() *cobra.Command {
	root := &cobra.Command{Use: "rtmx", SilenceUsage: true, SilenceErrors: true}
	c := &cobra.Command{Use: "checkout", RunE: runCheckout}
	c.Flags().StringVar(&checkoutOrg, "org", "", "")
	c.Flags().StringVar(&checkoutCreateOrg, "create-org", "", "")
	c.Flags().IntVar(&checkoutSeats, "seats", 1, "")
	c.Flags().StringVar(&checkoutSyncURL, "sync-http-url", "", "")
	c.Flags().StringVar(&checkoutProvider, "provider", "github", "")
	c.Flags().BoolVar(&checkoutNoBrowser, "no-browser", false, "")
	root.AddCommand(c)
	return root
}

func TestCheckoutOpensStripeURL(t *testing.T) {
	rtmx.Req(t, "REQ-GO-084")
	output.DisableColor()
	defer output.EnableColor()

	mux := http.NewServeMux()
	mux.HandleFunc("/orgs", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"organizations": []map[string]string{
				{"id": "org-1", "name": "Acme", "slug": "acme", "tier": "free"},
			},
		})
	})
	mux.HandleFunc("/billing/checkout", func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if int(body["quantity"].(float64)) != 1 {
			t.Errorf("quantity = %v", body["quantity"])
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"url":        "https://checkout.stripe.com/c/pay/cs_test_cli",
			"session_id": "cs_test_cli",
			"metadata":   map[string]string{"org_id": "org-1"},
		})
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	oldLogin := ensureManagedLogin
	ensureManagedLogin = func(context.Context) (*auth.TokenSet, error) {
		return &auth.TokenSet{AccessToken: "sess", Mode: "managed"}, nil
	}
	defer func() { ensureManagedLogin = oldLogin }()

	opened := ""
	oldBrowser := checkoutBrowser
	checkoutBrowser = func(u string) error {
		opened = u
		return nil
	}
	defer func() { checkoutBrowser = oldBrowser }()

	cmd := createCheckoutTestCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"checkout", "--sync-http-url", srv.URL, "--seats", "1"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("checkout: %v\n%s", err, buf.String())
	}
	if opened != "https://checkout.stripe.com/c/pay/cs_test_cli" {
		t.Fatalf("opened = %q", opened)
	}
	out := buf.String()
	if !strings.Contains(out, "webhook") && !strings.Contains(out, "Entitlement") {
		t.Fatalf("expected entitlement messaging:\n%s", out)
	}
}

func TestCheckoutTriggersLoginWhenUnauthenticated(t *testing.T) {
	rtmx.Req(t, "REQ-GO-084")
	output.DisableColor()
	defer output.EnableColor()

	mux := http.NewServeMux()
	mux.HandleFunc("/orgs", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"organizations": []map[string]string{
				{"id": "o", "name": "Solo", "slug": "solo"},
			},
		})
	})
	mux.HandleFunc("/billing/checkout", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"url": "https://checkout.stripe.com/c/pay/cs_test_2", "session_id": "cs_test_2",
		})
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	loginCalled := false
	oldLogin := ensureManagedLogin
	ensureManagedLogin = func(context.Context) (*auth.TokenSet, error) {
		loginCalled = true
		return &auth.TokenSet{AccessToken: "after-login"}, nil
	}
	defer func() { ensureManagedLogin = oldLogin }()

	oldBrowser := checkoutBrowser
	checkoutBrowser = func(string) error { return nil }
	defer func() { checkoutBrowser = oldBrowser }()

	checkoutNoBrowser = false
	cmd := createCheckoutTestCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"checkout", "--sync-http-url", srv.URL})

	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !loginCalled {
		t.Fatal("expected ensureManagedLogin")
	}
}

func TestCheckoutCreateOrgFlag(t *testing.T) {
	rtmx.Req(t, "REQ-GO-084")
	output.DisableColor()
	defer output.EnableColor()

	mux := http.NewServeMux()
	mux.HandleFunc("/orgs", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s", r.Method)
		}
		_ = json.NewEncoder(w).Encode(map[string]string{
			"id": "new", "name": "Fresh", "slug": "fresh",
		})
	})
	mux.HandleFunc("/billing/checkout", func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["org_id"] != "new" {
			t.Fatalf("org_id = %v", body["org_id"])
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"url": "https://checkout.stripe.com/c/pay/cs_test_3", "session_id": "cs_test_3",
		})
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	oldLogin := ensureManagedLogin
	ensureManagedLogin = func(context.Context) (*auth.TokenSet, error) {
		return &auth.TokenSet{AccessToken: "t"}, nil
	}
	defer func() { ensureManagedLogin = oldLogin }()
	oldBrowser := checkoutBrowser
	checkoutBrowser = func(string) error { return nil }
	defer func() { checkoutBrowser = oldBrowser }()

	cmd := createCheckoutTestCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"checkout", "--sync-http-url", srv.URL, "--create-org", "Fresh", "--seats", "3"})

	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "Fresh") {
		t.Fatalf("output: %s", buf.String())
	}
}
