package broker

// WindowsKeyACL describes the intended ACL on the Windows verifier private key
// (REQ-MONO-069bc). Application of the ACL is platform-specific; this type is
// the portable contract the installer must enforce.
type WindowsKeyACL struct {
	GrantRead []string
	DenyRead  []string
}

// DefaultWindowsKeyACL grants NT SERVICE\rtmx and denies a typical developer
// interactive account identity used in ACL tests.
func DefaultWindowsKeyACL(developerAccount string) WindowsKeyACL {
	return WindowsKeyACL{
		GrantRead: []string{`NT SERVICE\rtmx`},
		DenyRead:  []string{developerAccount},
	}
}

// GrantsVerifierDeniesDeveloper reports whether the ACL grants the service
// account and denies the developer account.
func (a WindowsKeyACL) GrantsVerifierDeniesDeveloper(developerAccount string) bool {
	grantOK := false
	for _, g := range a.GrantRead {
		if g == `NT SERVICE\rtmx` {
			grantOK = true
			break
		}
	}
	denyOK := false
	for _, d := range a.DenyRead {
		if d == developerAccount {
			denyOK = true
			break
		}
	}
	return grantOK && denyOK
}
