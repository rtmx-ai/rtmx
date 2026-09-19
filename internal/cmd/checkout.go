package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/rtmx-ai/rtmx/internal/auth"
	"github.com/rtmx-ai/rtmx/internal/billing"
	"github.com/rtmx-ai/rtmx/internal/output"
	"github.com/spf13/cobra"
)

var (
	checkoutOrg       string
	checkoutCreateOrg string
	checkoutSeats     int
	checkoutSyncURL   string
	checkoutProvider  string
	checkoutNoBrowser bool
)

// checkoutBrowser opens the Stripe Checkout URL (injectable in tests).
var checkoutBrowser = openSystemBrowser

// ensureManagedLogin runs managed login when no stored credential exists.
var ensureManagedLogin = func(ctx context.Context) (*auth.TokenSet, error) {
	tokens, err := auth.LoadStoredTokens(resolveTokenPath())
	if err == nil && tokens != nil && tokens.AccessToken != "" && !tokens.IsExpired() {
		return tokens, nil
	}
	result, err := managedLoginFunc(ctx, auth.ManagedLoginConfig{
		SyncURL:   resolveManagedSyncURL(),
		Provider:  checkoutProvider,
		TokenPath: resolveTokenPath(),
		Browser:   openSystemBrowser,
	})
	if err != nil {
		return nil, err
	}
	return result.Tokens, nil
}

var checkoutCmd = &cobra.Command{
	Use:   "checkout",
	Short: "Start managed Team Checkout in the browser",
	Long: `Create a Stripe Checkout Session for managed Sync Team seats.

Zero-config defaults:
  sync host: https://sync.rtmx.ai (or --sync-http-url / $RTMX_SYNC_HTTP_URL)
  seats:     1
  offering:  managed_sync / team

If you are not signed in, runs managed login first (rtmx login).
Organization selection:
  --org SLUG_OR_ID     use an existing org
  --create-org NAME    create a new org, then checkout
  (no flag)            use the sole org, or error if zero/many

Entitlement applies after Stripe completes and the webhook lands —
this command only opens Checkout.`,
	RunE: runCheckout,
}

func init() {
	checkoutCmd.Flags().StringVar(&checkoutOrg, "org", "", "organization id or slug")
	checkoutCmd.Flags().StringVar(&checkoutCreateOrg, "create-org", "", "create an organization with this name, then checkout")
	checkoutCmd.Flags().IntVar(&checkoutSeats, "seats", 1, "Team seat quantity (default 1)")
	checkoutCmd.Flags().StringVar(&checkoutSyncURL, "sync-http-url", "", "managed sync HTTP base URL")
	checkoutCmd.Flags().StringVar(&checkoutProvider, "provider", "github", "OAuth provider if login is required")
	checkoutCmd.Flags().BoolVar(&checkoutNoBrowser, "no-browser", false, "print the Checkout URL without opening a browser")

	rootCmd.AddCommand(checkoutCmd)
}

func runCheckout(cmd *cobra.Command, args []string) error {
	if noColor {
		output.DisableColor()
	}
	if checkoutSeats < 1 {
		return fmt.Errorf("--seats must be >= 1")
	}

	// Prefer checkout-specific sync URL flag when set.
	if checkoutSyncURL != "" {
		authSyncURL = checkoutSyncURL
	}
	syncURL := resolveManagedSyncURL()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	cmd.Println("Preparing managed Team Checkout...")
	tokens, err := ensureManagedLogin(ctx)
	if err != nil {
		return fmt.Errorf("authentication required: %w\nRun 'rtmx login' and retry", err)
	}

	client := &billing.Client{
		BaseURL: syncURL,
		Token:   tokens.AccessToken,
	}

	orgID, orgLabel, err := resolveCheckoutOrg(ctx, client)
	if err != nil {
		return err
	}
	cmd.Printf("Organization: %s\n", orgLabel)
	cmd.Printf("Seats: %d\n", checkoutSeats)

	session, err := client.CreateCheckoutSession(ctx, billing.CheckoutRequest{
		Offering: "managed_sync",
		Tier:     "team",
		Quantity: checkoutSeats,
		OrgID:    orgID,
	})
	if err != nil {
		return fmt.Errorf("could not create Checkout Session: %w", err)
	}

	cmd.Printf("%s Checkout Session %s created.\n", output.Color("OK", output.Green), session.SessionID)
	if checkoutNoBrowser {
		cmd.Printf("Open this URL to pay:\n  %s\n", session.URL)
	} else {
		cmd.Println("Opening Stripe Checkout in your browser...")
		if err := checkoutBrowser(session.URL); err != nil {
			cmd.Printf("Could not open browser (%v). Open manually:\n  %s\n", err, session.URL)
		}
	}
	cmd.Println()
	cmd.Println("Complete payment in Stripe. Entitlement applies after the")
	cmd.Println("webhook lands — this command does not grant write access by itself.")
	cmd.Println("Then: rtmx sync --sync-url wss://sync.rtmx.ai/sync/<org>/<room> --pull")
	return nil
}

func resolveCheckoutOrg(ctx context.Context, client *billing.Client) (id, label string, err error) {
	if checkoutCreateOrg != "" {
		org, err := client.CreateOrg(ctx, checkoutCreateOrg)
		if err != nil {
			return "", "", fmt.Errorf("create organization: %w", err)
		}
		return org.ID, fmt.Sprintf("%s (%s)", org.Name, org.Slug), nil
	}

	orgs, err := client.ListOrgs(ctx)
	if err != nil {
		return "", "", fmt.Errorf("list organizations: %w", err)
	}

	if checkoutOrg != "" {
		ref := strings.TrimSpace(checkoutOrg)
		for _, o := range orgs {
			if o.ID == ref || o.Slug == ref || strings.EqualFold(o.Name, ref) {
				return o.ID, fmt.Sprintf("%s (%s)", o.Name, o.Slug), nil
			}
		}
		// Allow passing a fresh org id not yet listed (unlikely) by using the ref directly.
		return ref, ref, nil
	}

	switch len(orgs) {
	case 0:
		return "", "", fmt.Errorf("no organization found; pass --create-org NAME or create one on rtmx.ai/checkout")
	case 1:
		o := orgs[0]
		return o.ID, fmt.Sprintf("%s (%s)", o.Name, o.Slug), nil
	default:
		return "", "", fmt.Errorf("multiple organizations; pass --org SLUG_OR_ID (have %d)", len(orgs))
	}
}
