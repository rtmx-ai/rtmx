# REQ-DIST-006: Homebrew-core Formula Submission

## Metadata
- **Category**: DIST
- **Subcategory**: Homebrew
- **Priority**: HIGH
- **Phase**: 25
- **Status**: MISSING
- **Effort**: 1 week
- **Dependencies**: REQ-REL-007 (v1.0.0 tagged), REQ-GO-044 (custom tap works)
- **Blocks**: REQ-LAUNCH-002

## Requirement

RTMX shall be submitted to Homebrew/homebrew-core so that `brew install rtmx`
works without a custom tap prefix. The formula must pass `brew audit --strict
--new rtmx` and be accepted by Homebrew maintainers.

## Rationale

`brew install rtmx-ai/tap/rtmx` works but requires users to know the tap name.
`brew install rtmx` is the expected install experience. Homebrew-core acceptance
validates the project as a legitimate, stable CLI tool.

## Homebrew-core Requirements

1. **Stable version**: v1.0.0+ tagged release (not pre-release)
2. **Notable project**: GitHub stars, downloads, or community evidence
3. **No vendored dependencies**: Must build from source (Go modules)
4. **Test stanza**: Formula includes a working test block
5. **Audit clean**: `brew audit --strict --new rtmx` passes

## Formula Design

```ruby
class Rtmx < Formula
  desc "Requirements traceability management CLI"
  homepage "https://rtmx.ai"
  url "https://github.com/rtmx-ai/rtmx/archive/refs/tags/v1.0.0.tar.gz"
  sha256 "PLACEHOLDER"
  license "Apache-2.0"

  depends_on "go" => :build

  def install
    ldflags = %W[
      -s -w
      -X github.com/rtmx-ai/rtmx/internal/cmd.Version=#{version}
      -X github.com/rtmx-ai/rtmx/internal/cmd.Commit=#{tap.user}
      -X github.com/rtmx-ai/rtmx/internal/cmd.Date=#{time.iso8601}
    ]
    system "go", "build", *std_go_args(ldflags:), "./cmd/rtmx"
  end

  test do
    system "#{bin}/rtmx", "version"
    system "#{bin}/rtmx", "init"
    assert_predicate testpath/".rtmx", :directory?
  end
end
```

## Acceptance Criteria

1. Formula PR submitted to Homebrew/homebrew-core
2. `brew audit --strict --new rtmx` passes locally
3. Formula builds from source (not downloading prebuilt binary)
4. Test stanza verifies `rtmx version` and `rtmx init`
5. Formula accepted and merged by Homebrew maintainers
6. `brew install rtmx` works on a clean machine

## Verification Test

Test validates formula file exists and passes structural checks. Actual
homebrew-core submission is verified by inspection.

## Files to Create

- `Formula/rtmx.rb` -- Homebrew formula (for local testing before PR)

## Notes

- Custom tap (rtmx-ai/homebrew-tap) continues to work as a fallback
- Homebrew-core formula is maintained by the Homebrew community after merge
- Updates to the formula happen via `brew bump-formula-pr` or automation
