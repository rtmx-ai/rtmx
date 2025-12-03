# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 0.1.x   | :white_check_mark: |

## Reporting a Vulnerability

We take security vulnerabilities seriously. If you discover a security issue, please report it responsibly.

### How to Report

**Do NOT open a public GitHub issue for security vulnerabilities.**

Instead, please report security vulnerabilities by emailing:

**security@iotactical.com**

Include the following information:
- Type of vulnerability (e.g., injection, authentication bypass, etc.)
- Full path to the affected source file(s)
- Steps to reproduce the issue
- Proof-of-concept or exploit code (if possible)
- Impact assessment

### What to Expect

1. **Acknowledgment**: We will acknowledge receipt within 48 hours
2. **Assessment**: We will assess the vulnerability within 7 days
3. **Resolution**: We aim to resolve critical issues within 30 days
4. **Disclosure**: We will coordinate disclosure timing with you

### Scope

This security policy covers:
- The rtmx Python package
- CLI commands and their handling of user input
- Configuration file parsing
- Integration with external services (GitHub, Jira)
- MCP server implementation

### Out of Scope

- Issues in dependencies (report to the respective project)
- Theoretical attacks without proof of concept
- Social engineering attacks

## Security Measures

### Current Protections

- **Input validation**: All CLI inputs are validated
- **No eval/exec**: We never execute arbitrary code from user input
- **Token handling**: API tokens are read from environment variables only
- **Dependency scanning**: Automated pip-audit checks in CI
- **SBOM generation**: CycloneDX SBOM attached to releases

### Best Practices for Users

1. **Keep updated**: Always use the latest version
2. **Secure tokens**: Store API tokens securely in environment variables
3. **Review permissions**: Only grant necessary access to GitHub/Jira tokens
4. **Audit dependencies**: Use `pip-audit` to check for vulnerabilities

## Security Updates

Security updates are released as patch versions (e.g., 0.1.1, 0.1.2).

Subscribe to GitHub releases to be notified of security updates.

## Acknowledgments

We thank security researchers who help keep RTMX secure. Contributors who report valid security issues will be acknowledged in release notes (unless they prefer anonymity).
