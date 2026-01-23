# REQ-VERIFY-004: NIST 800-53 Compliance Mapping

## Status: NOT STARTED
## Priority: MEDIUM
## Phase: 13
## Effort: 3.0 weeks

## Description

RTMX shall provide a comprehensive mapping to NIST 800-53 security controls, with automated compliance testing that generates audit evidence. This enables customers to include RTMX in FedRAMP-authorized environments and provides CISOs with compliance documentation.

## Acceptance Criteria

- [ ] Compliance mapping document covers all applicable AC controls
- [ ] Automated tests tagged with NIST control IDs
- [ ] Compliance test runner generates evidence reports
- [ ] AC-2 (Account Management): Zitadel user lifecycle tests
- [ ] AC-3 (Access Enforcement): Grant-based access tests
- [ ] AC-4 (Information Flow): Cross-repo visibility tests
- [ ] AC-6 (Least Privilege): Role minimization tests
- [ ] AC-17 (Remote Access): Zero-trust network tests
- [ ] Reports in JSON and PDF formats for auditors
- [ ] CI generates compliance report on release

## Test Cases

- `tests/compliance/test_ac_2_account_management.py` - AC-2 controls
- `tests/compliance/test_ac_3_access_enforcement.py` - AC-3 controls
- `tests/compliance/test_ac_4_information_flow.py` - AC-4 controls
- `tests/compliance/test_ac_6_least_privilege.py` - AC-6 controls
- `tests/compliance/test_ac_17_remote_access.py` - AC-17 controls

## Technical Notes

### NIST 800-53 Control Mapping

| Control | Title | RTMX Implementation | Test |
|---------|-------|---------------------|------|
| AC-2 | Account Management | Zitadel user provisioning, deprovisioning | test_ac_2_* |
| AC-2(1) | Automated System Account Management | Zitadel API automation | test_ac_2_automated |
| AC-3 | Access Enforcement | Grant-based authorization | test_ac_3_* |
| AC-3(7) | Role-Based Access Control | Role hierarchy in grants | test_ac_3_rbac |
| AC-4 | Information Flow Enforcement | Shadow requirements, category constraints | test_ac_4_* |
| AC-6 | Least Privilege | Default deny, minimal roles | test_ac_6_* |
| AC-6(1) | Authorize Access to Security Functions | Admin role separation | test_ac_6_admin |
| AC-17 | Remote Access | OpenZiti zero-trust network | test_ac_17_* |
| AC-17(1) | Automated Monitoring | Connection logging | test_ac_17_logging |

### Compliance Test Structure

```python
# tests/compliance/test_ac_2_account_management.py
import pytest

@pytest.mark.compliance
@pytest.mark.nist("AC-2")
class TestAC2AccountManagement:
    """NIST 800-53 AC-2: Account Management tests."""

    @pytest.mark.nist("AC-2(a)")
    def test_account_types_defined(self, zitadel_client):
        """AC-2(a): Organization defines account types.

        Evidence: Zitadel project has defined roles for RTMX users.
        """
        roles = zitadel_client.list_project_roles("rtmx")

        required_roles = {
            "dependency_viewer",
            "status_observer",
            "requirement_editor",
            "admin"
        }

        assert required_roles <= set(roles)

    @pytest.mark.nist("AC-2(c)")
    def test_account_creation_requires_authorization(self, zitadel_client):
        """AC-2(c): Account creation requires manager authorization.

        Evidence: User provisioning requires admin role.
        """
        # Attempt user creation without admin role
        with pytest.raises(PermissionError):
            zitadel_client.create_user(
                "test@example.com",
                requester_roles=["viewer"]  # Not admin
            )

    @pytest.mark.nist("AC-2(i)")
    def test_account_termination(self, zitadel_client, test_user):
        """AC-2(i): Accounts disabled when no longer required.

        Evidence: Deprovisioned users immediately lose access.
        """
        # User has access
        assert zitadel_client.can_access(test_user, "rtmx")

        # Deactivate user
        zitadel_client.deactivate_user(test_user)

        # Access immediately revoked
        assert not zitadel_client.can_access(test_user, "rtmx")
```

### Compliance Report Generator

```python
# rtmx/compliance/reporter.py
import json
from datetime import datetime
from pathlib import Path

class ComplianceReporter:
    """Generate NIST 800-53 compliance evidence reports."""

    def __init__(self, results: list[TestResult]):
        self.results = results
        self.timestamp = datetime.utcnow()

    def generate_json_report(self) -> dict:
        """Generate JSON compliance report."""
        controls = {}

        for result in self.results:
            control_id = result.nist_control
            if control_id not in controls:
                controls[control_id] = {
                    "control_id": control_id,
                    "title": CONTROL_TITLES[control_id],
                    "tests": [],
                    "status": "PASS"
                }

            controls[control_id]["tests"].append({
                "name": result.test_name,
                "docstring": result.docstring,
                "status": "PASS" if result.passed else "FAIL",
                "duration_ms": result.duration_ms
            })

            if not result.passed:
                controls[control_id]["status"] = "FAIL"

        return {
            "report_type": "NIST 800-53 Compliance",
            "generated_at": self.timestamp.isoformat(),
            "system": "RTMX",
            "controls": list(controls.values()),
            "summary": {
                "total_controls": len(controls),
                "passing": sum(1 for c in controls.values() if c["status"] == "PASS"),
                "failing": sum(1 for c in controls.values() if c["status"] == "FAIL")
            }
        }

    def generate_pdf_report(self, output_path: Path) -> None:
        """Generate PDF compliance report for auditors."""
        # Uses reportlab or weasyprint
        ...
```

### CI Integration

```yaml
# .github/workflows/compliance.yml
compliance-report:
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v4
    - name: Run compliance tests
      run: |
        pytest tests/compliance/ -v \
          --compliance-report=nist-800-53.json \
          --junit-xml=compliance-junit.xml
    - name: Generate PDF report
      run: |
        rtmx compliance report \
          --input nist-800-53.json \
          --output compliance-report.pdf
    - name: Upload reports
      uses: actions/upload-artifact@v4
      with:
        name: compliance-reports
        path: |
          nist-800-53.json
          compliance-report.pdf
```

## Files to Create/Modify

- `tests/compliance/__init__.py` - Compliance test package
- `tests/compliance/conftest.py` - Compliance fixtures and markers
- `tests/compliance/test_ac_2_account_management.py` - AC-2 tests
- `tests/compliance/test_ac_3_access_enforcement.py` - AC-3 tests
- `tests/compliance/test_ac_4_information_flow.py` - AC-4 tests
- `tests/compliance/test_ac_6_least_privilege.py` - AC-6 tests
- `tests/compliance/test_ac_17_remote_access.py` - AC-17 tests
- `src/rtmx/compliance/reporter.py` - Report generator
- `src/rtmx/cli/compliance.py` - Compliance CLI commands
- `docs/security/compliance-mapping.md` - Control mapping document
- `.github/workflows/compliance.yml` - CI workflow

## Dependencies

- REQ-ZT-001: Zitadel OIDC (for AC-2 tests)
- REQ-ZT-002: OpenZiti (for AC-17 tests)
- REQ-ZT-003: JWT validation (for AC-3 tests)
- REQ-VERIFY-001: Property tests (evidence source)
- REQ-VERIFY-002: TLA+ spec (evidence source)
- REQ-VERIFY-003: Chaos tests (evidence source)

## Blocks

None (terminal requirement in verification chain)
