# Homebrew formula for rtmx
# This file is used for local testing before submitting to homebrew-core.
# Submission: brew tap-new rtmx-ai/homebrew-core && brew create ...
# Then: brew audit --strict --new rtmx && brew test rtmx

class Rtmx < Formula
  desc "Requirements traceability management CLI -- track what you built, tested, and what's next"
  homepage "https://rtmx.ai"
  url "https://github.com/rtmx-ai/rtmx/archive/refs/tags/v1.0.0.tar.gz"
  sha256 "PLACEHOLDER_SHA256"
  license "Apache-2.0"
  head "https://github.com/rtmx-ai/rtmx.git", branch: "main"

  depends_on "go" => :build

  def install
    ldflags = %W[
      -s -w
      -X github.com/rtmx-ai/rtmx/internal/cmd.Version=#{version}
      -X github.com/rtmx-ai/rtmx/internal/cmd.Commit=HEAD
      -X github.com/rtmx-ai/rtmx/internal/cmd.Date=#{time.iso8601}
    ]
    system "go", "build", *std_go_args(ldflags:), "./cmd/rtmx"
  end

  test do
    # Verify binary runs and reports correct version
    assert_match version.to_s, shell_output("#{bin}/rtmx version")

    # Verify init creates project structure
    system "#{bin}/rtmx", "init"
    assert_predicate testpath/".rtmx", :directory?
    assert_predicate testpath/"rtmx.yaml", :file?
  end
end
