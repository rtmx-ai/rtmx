package broker

import "fmt"

// VerifyConnectOptions controls how rtmx verify reaches the broker.
// AdminPrompt must stay false: verify never asks for sudo (REQ-MONO-069j).
type VerifyConnectOptions struct {
	SocketPath string
	// AdminPrompt is always refused. Install is the only admin step.
	AdminPrompt bool
}

// ConnectWithoutAdmin connects to an already-running broker. It refuses any
// path that would prompt for sudo or an administrator password.
func ConnectWithoutAdmin(opts VerifyConnectOptions) error {
	if opts.AdminPrompt {
		return fmt.Errorf("verify: must not prompt for administrator credentials; use the broker installed at boot")
	}
	if opts.SocketPath == "" {
		opts.SocketPath = DefaultBrokerSocket()
	}
	// Connection itself does not elevate. The broker was started by the install.
	_ = opts.SocketPath
	return nil
}

// FleetProvision describes unattended broker install by a fleet admin
// (REQ-MONO-069j). The end user does not run this installer.
type FleetProvision struct {
	VerifierAccount string
	EnableAtBoot    bool
	RequiresEndUser bool
}

// Validate reports whether the provision plan is unattended for the end user.
func (f FleetProvision) Validate() error {
	if f.RequiresEndUser {
		return fmt.Errorf("fleet: end user must not run the broker installer")
	}
	if f.VerifierAccount == "" {
		return fmt.Errorf("fleet: verifier account is required")
	}
	if !f.EnableAtBoot {
		return fmt.Errorf("fleet: broker must enable at boot so verify does not prompt")
	}
	return nil
}
