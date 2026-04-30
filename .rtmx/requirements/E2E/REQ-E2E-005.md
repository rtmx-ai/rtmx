# REQ-E2E-005: Init/Setup Coherence and Dogfood Testing

## Metadata
- **Category**: E2E
- **Subcategory**: Onboarding
- **Priority**: P0
- **Phase**: 20
- **Status**: MISSING
- **Dependencies**: REQ-GO-016, REQ-GO-026
- **Blocks**: REQ-GO-047

## Requirement

The `rtmx setup` command shall detect and respect an existing `.rtmx/` directory
structure created by `rtmx init`. When `.rtmx/` exists, setup shall use modern
paths (`.rtmx/database.csv`, `.rtmx/requirements/`) instead of hardcoded legacy
paths (`docs/rtm_database.csv`, `docs/requirements/`). E2E tests shall verify
coherence across all init/setup sequencing and dogfood against the rtmx repo's
own `.rtmx/` infrastructure.

## Rationale

Running `rtmx init` followed by `rtmx setup` creates two parallel database
structures with no shared state. Users encounter divergent `.rtmx/database.csv`
and `docs/rtm_database.csv` files, leading to confusion about which is
authoritative. The `init` command defaults to the modern `.rtmx/` layout, but
`setup` hardcodes the legacy `docs/` layout (setup.go lines ~192, 229-231),
ignoring the modern structure entirely.

The `migrate` command already knows how to detect and convert between the two
layouts, confirming `.rtmx/` is the intended modern default. The fix wires
this detection into `setup`.

## Design

### Bug Fix (setup.go)

1. `detectProject()` shall check for `.rtmx/database.csv` and `.rtmx/config.yaml`
   in addition to the existing `docs/rtm_database.csv` and `rtmx.yaml` checks.
2. When `.rtmx/` structure is detected, Phase 2 (Configuration) shall skip
   creating root `rtmx.yaml` if `.rtmx/config.yaml` already exists.
3. When `.rtmx/` structure is detected, Phase 3 (RTM Database) shall use
   `.rtmx/database.csv` and `.rtmx/requirements/` instead of `docs/` paths.
4. When neither structure exists, `setup` shall default to the modern `.rtmx/`
   layout (matching `init` default behavior).

### E2E Test Scenarios

```go
// test/onboarding_e2e_test.go

func TestInitThenSetup(t *testing.T)
// 1. Run rtmx init in temp dir
// 2. Run rtmx setup in same dir
// 3. Assert: NO docs/ directory created
// 4. Assert: .rtmx/database.csv is the only database
// 5. Assert: .rtmx/requirements/ is the only req tree
// 6. Assert: rtmx status succeeds and reads correct database

func TestSetupAlone(t *testing.T)
// 1. Run rtmx setup in empty temp dir (no prior init)
// 2. Assert: .rtmx/ structure created (modern default)
// 3. Assert: NO docs/ directory created
// 4. Assert: rtmx status succeeds

func TestSetupOnExistingProject(t *testing.T)
// 1. Run rtmx setup in temp dir with pre-existing .rtmx/
//    containing populated database and requirements
// 2. Assert: existing database content preserved
// 3. Assert: existing requirements preserved
// 4. Assert: no docs/ directory created
// 5. Assert: rtmx status reports correct counts

func TestSetupLegacyMode(t *testing.T)
// 1. Run rtmx init --legacy in temp dir
// 2. Run rtmx setup in same dir
// 3. Assert: setup detects docs/ layout and uses it
// 4. Assert: no .rtmx/ directory created
// 5. Assert: rtmx status succeeds

func TestDogfoodSelf(t *testing.T)
// 1. Run rtmx setup --dry-run in the rtmx repo itself
// 2. Assert: detects existing .rtmx/ structure
// 3. Assert: would NOT create docs/ directory
// 4. Assert: reports .rtmx/database.csv as RTM database
// 5. Assert: all existing requirements remain intact
```

### Benchmark Scenario (optional follow-up)

Add an init/setup benchmark to `benchmarks/` that:
1. Clones a fresh exemplar project
2. Runs `rtmx init` -> `rtmx bootstrap --from-tests` -> `rtmx verify`
3. Validates the full onboarding pipeline end-to-end

## Acceptance Criteria

1. `rtmx init` then `rtmx setup` produces exactly one database and one requirements tree
2. `rtmx setup` alone (no prior init) creates the modern `.rtmx/` structure
3. `rtmx setup` on a project with existing `.rtmx/` preserves all content
4. `rtmx setup` on a project with existing `docs/` (legacy) respects that layout
5. `rtmx setup --dry-run` against the rtmx repo reports `.rtmx/` paths, not `docs/`
6. All tests pass with race detector enabled
7. No regression in existing setup unit tests

## Files to Create/Modify

- `internal/cmd/setup.go` - Fix path resolution to detect and use `.rtmx/`
- `internal/cmd/setup_test.go` - Update unit tests for new detection logic
- `test/onboarding_e2e_test.go` - New E2E test file
- `.rtmx/requirements/E2E/REQ-E2E-005.md` - This file
- `.rtmx/database.csv` - Add requirement entry
