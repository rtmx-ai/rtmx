"""NIST 800-53 Compliance Tests for RTMX.

This package contains tests that map to NIST 800-53 security controls,
providing evidence for compliance audits and certifications.

Control Families Covered:

AC (Access Control):
- AC-2: Account Management
- AC-3: Access Enforcement
- AC-4: Information Flow Enforcement
- AC-6: Least Privilege
- AC-17: Remote Access
- AC-21: Information Sharing

AU (Audit and Accountability):
- AU-2: Audit Events
- AU-3: Content of Audit Records
- AU-6: Audit Review, Analysis, and Reporting
- AU-9: Protection of Audit Information
- AU-12: Audit Generation

IA (Identification and Authentication):
- IA-2: Identification and Authentication (Organizational Users)
- IA-4: Identifier Management
- IA-5: Authenticator Management
- IA-8: Identification and Authentication (Non-Organizational Users)
- IA-9: Service Identification and Authentication

SC (System and Communications Protection):
- SC-7: Boundary Protection
- SC-8: Transmission Confidentiality and Integrity
- SC-12: Cryptographic Key Establishment and Management
- SC-13: Cryptographic Protection
- SC-23: Session Authenticity
- SC-28: Protection of Information at Rest

Each test function documents:
- Control ID (e.g., AC-2)
- Control Name
- How RTMX implements the control
- Test verification approach

Usage:
    pytest tests/compliance/ -v                    # Run all compliance tests
    pytest tests/compliance/test_ac_*.py -v       # Run AC family only
    pytest tests/compliance/ --compliance-report  # Generate compliance report
"""
