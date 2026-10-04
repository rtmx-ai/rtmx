package billing

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateCheckoutSession(t *testing.T) {
	t.Parallel()
	mux := http.NewServeMux()
	mux.HandleFunc("/billing/checkout", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer sess" {
			t.Errorf("auth = %q", r.Header.Get("Authorization"))
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["quantity"].(float64) != 2 {
			t.Errorf("quantity = %v", body["quantity"])
		}
		if body["org_id"] != "org-1" {
			t.Errorf("org_id = %v", body["org_id"])
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"url":        "https://checkout.stripe.com/c/pay/cs_test_x",
			"session_id": "cs_test_x",
			"metadata":   map[string]string{"org_id": "org-1", "quantity": "2"},
		})
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	c := &Client{BaseURL: srv.URL, Token: "sess", HTTPClient: srv.Client()}
	session, err := c.CreateCheckoutSession(context.Background(), CheckoutRequest{
		OrgID:    "org-1",
		Quantity: 2,
	})
	if err != nil {
		t.Fatal(err)
	}
	if session.SessionID != "cs_test_x" {
		t.Fatalf("session = %#v", session)
	}
}

func TestListAndCreateOrg(t *testing.T) {
	t.Parallel()
	mux := http.NewServeMux()
	mux.HandleFunc("/orgs", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			_ = json.NewEncoder(w).Encode(map[string]any{
				"organizations": []map[string]string{
					{"id": "1", "name": "Acme", "slug": "acme", "tier": "free"},
				},
			})
		case http.MethodPost:
			var body map[string]string
			_ = json.NewDecoder(r.Body).Decode(&body)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"id": "2", "name": body["name"], "slug": "new-co", "tier": "free",
			})
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	c := &Client{BaseURL: srv.URL, Token: "t", HTTPClient: srv.Client()}
	orgs, err := c.ListOrgs(context.Background())
	if err != nil || len(orgs) != 1 || orgs[0].Slug != "acme" {
		t.Fatalf("list: %v %#v", err, orgs)
	}
	created, err := c.CreateOrg(context.Background(), "New Co")
	if err != nil || created.Slug != "new-co" {
		t.Fatalf("create: %v %#v", err, created)
	}
}
