# RTMX Language Benchmark Candidates

Each language has a proposed open source exemplar project and a marker style.
Projects are selected for: moderate build time, good test coverage, representative
test framework usage, and permissive license.

## Tier 1 -- Core Languages (highest priority)

### Go
- **Exemplar**: [cli/cli](https://github.com/cli/cli) (GitHub CLI, ~800 tests, Go stdlib testing)
- **Build time**: ~2 min
- **Marker style**: `rtmx.Req(t, "REQ-ID")`
- **Verify command**: `go test -json ./...`

### Python
- **Exemplar**: [psf/requests](https://github.com/psf/requests) (HTTP library, pytest, ~500 tests)
- **Build time**: ~1 min
- **Marker style**: `@pytest.mark.req("REQ-ID")`
- **Verify command**: `pytest --rtmx-output results.json`

### Rust
- **Exemplar**: [rtmx-ai/aegis-cli](https://github.com/rtmx-ai/aegis-cli) (our own, ~437 tests)
- **Build time**: ~3 min
- **Marker style**: `// @req REQ-ID`
- **Verify command**: `cargo test --workspace`

### JavaScript/TypeScript
- **Exemplar**: [sindresorhus/got](https://github.com/sindresorhus/got) (HTTP client, AVA/Jest, ~400 tests)
- **Build time**: ~1 min
- **Marker style**: `// rtmx:req REQ-ID`
- **Verify command**: `npm test`

### Java
- **Exemplar**: [google/gson](https://github.com/google/gson) (JSON library, JUnit 5, ~1000 tests)
- **Build time**: ~3 min
- **Marker style**: `@Req("REQ-ID")` annotation
- **Verify command**: `mvn test`

### C#/.NET
- **Exemplar**: [jbogard/MediatR](https://github.com/jbogard/MediatR) (mediator pattern, xUnit, ~200 tests)
- **Build time**: ~2 min
- **Marker style**: `[Req("REQ-ID")]` attribute
- **Verify command**: `dotnet test`

## Tier 2 -- Systems Languages

### C/C++
- **Exemplar**: [nlohmann/json](https://github.com/nlohmann/json) (JSON for C++, Catch2, ~40K assertions)
- **Build time**: ~5 min (large test suite)
- **Marker style**: `// rtmx:req REQ-ID` + `RTMX_REQ("REQ-ID")`
- **Verify command**: `cmake --build . && ctest`
- **Note**: May want a smaller project like [fmtlib/fmt](https://github.com/fmtlib/fmt)

### Swift
- **Exemplar**: [apple/swift-argument-parser](https://github.com/apple/swift-argument-parser) (CLI framework, XCTest, ~300 tests)
- **Build time**: ~2 min
- **Marker style**: `// rtmx:req REQ-ID`
- **Verify command**: `swift test`

### Dart
- **Exemplar**: [dart-lang/path](https://github.com/dart-lang/path) (path manipulation, dart test, ~200 tests)
- **Build time**: ~1 min
- **Marker style**: `// rtmx:req REQ-ID`
- **Verify command**: `dart test`

### Kotlin
- **Exemplar**: [square/okio](https://github.com/square/okio) (I/O library, JUnit/kotest, ~500 tests)
- **Build time**: ~3 min
- **Marker style**: `@Req("REQ-ID")`
- **Verify command**: `gradle test`

## Tier 3 -- Domain-Specific Languages

### MATLAB
- **Exemplar**: [fieldtrip/fieldtrip](https://github.com/fieldtrip/fieldtrip) (neuroimaging, matlab.unittest)
- **Build time**: N/A (requires MATLAB license)
- **Marker style**: `% rtmx:req REQ-ID`
- **Alternative**: Create a minimal exemplar with GNU Octave compatibility
- **Note**: License-dependent; may need exemplar-only approach

### Verilog/SystemVerilog
- **Exemplar**: [lowRISC/ibex](https://github.com/lowRISC/ibex) (RISC-V core, UVM, ~100 tests)
- **Build time**: ~5 min (requires Verilator or commercial simulator)
- **Marker style**: `// rtmx:req REQ-ID`
- **Alternative**: [steveicarus/iverilog](https://github.com/steveicarus/iverilog) test suite
- **Note**: May need Verilator in CI; exemplar may be more practical

### R
- **Exemplar**: [tidyverse/dplyr](https://github.com/tidyverse/dplyr) (data manipulation, testthat, ~2000 tests)
- **Build time**: ~3 min
- **Marker style**: `# rtmx:req REQ-ID`
- **Verify command**: `Rscript -e "testthat::test_dir('tests')"`

### Julia
- **Exemplar**: [JuliaLang/JSON.jl](https://github.com/JuliaLang/JSON.jl) (JSON parser, Test stdlib, ~100 tests)
- **Build time**: ~2 min (first-run JIT compilation)
- **Marker style**: `# rtmx:req REQ-ID`
- **Verify command**: `julia -e "using Pkg; Pkg.test()"`

## Tier 4 -- Infrastructure Languages

### Terraform
- **Exemplar**: [hashicorp/terraform-provider-aws](https://github.com/hashicorp/terraform-provider-aws) (too large)
- **Alternative**: [cloudposse/terraform-aws-vpc](https://github.com/cloudposse/terraform-aws-vpc) (~20 tests)
- **Build time**: ~2 min (mocked)
- **Marker style**: `# rtmx:req REQ-ID` in .tftest.hcl
- **Verify command**: `terraform test`

### PHP
- **Exemplar**: [symfony/console](https://github.com/symfony/console) (CLI framework, PHPUnit, ~500 tests)
- **Build time**: ~1 min
- **Marker style**: `// rtmx:req REQ-ID`
- **Verify command**: `phpunit`

### Elixir
- **Exemplar**: [elixir-lang/plug](https://github.com/elixir-lang/plug) (HTTP middleware, ExUnit, ~300 tests)
- **Build time**: ~2 min
- **Marker style**: `# rtmx:req REQ-ID`
- **Verify command**: `mix test`

## Tier 5 -- Legacy Languages

### COBOL
- **Exemplar**: [openmainframeproject/cobol-programming-course](https://github.com/openmainframeproject/cobol-programming-course)
- **Build time**: ~1 min (GnuCOBOL)
- **Marker style**: `* rtmx:req REQ-ID` (fixed-format)
- **Verify command**: `cobc -x test.cob && ./test`
- **Note**: Limited test frameworks; exemplar-only may be best

### Fortran
- **Exemplar**: [fortran-lang/stdlib](https://github.com/fortran-lang/stdlib) (standard library, pFUnit, ~200 tests)
- **Build time**: ~3 min
- **Marker style**: `! rtmx:req REQ-ID`
- **Verify command**: `fpm test`

### Ada/SPARK
- **Exemplar**: [AdaCore/Ada_Drivers_Library](https://github.com/AdaCore/Ada_Drivers_Library)
- **Build time**: ~2 min (GNAT)
- **Marker style**: `-- rtmx:req REQ-ID`
- **Verify command**: `gprbuild && ./test_runner`
- **Alternative**: [annexi-strayline/AURA](https://github.com/annexi-strayline/AURA)

### Perl
- **Exemplar**: [mojolicious/mojo](https://github.com/mojolicious/mojo) (web framework, Test::More, ~1000 tests)
- **Build time**: ~1 min
- **Marker style**: `# rtmx:req REQ-ID`
- **Verify command**: `prove -l t/`

## Tier 6 -- Functional and Niche

### Haskell
- **Exemplar**: [aeson](https://github.com/haskell/aeson) (JSON library, HSpec, ~200 tests)
- **Build time**: ~5 min (GHC compilation)
- **Marker style**: `-- rtmx:req REQ-ID`
- **Verify command**: `cabal test`

### Lua
- **Exemplar**: [lunarmodules/luacheck](https://github.com/lunarmodules/luacheck) (linter, busted, ~100 tests)
- **Build time**: ~30 sec
- **Marker style**: `-- rtmx:req REQ-ID`
- **Verify command**: `busted`

### Scala
- **Exemplar**: [circe/circe](https://github.com/circe/circe) (JSON library, ScalaTest, ~500 tests)
- **Build time**: ~5 min (SBT)
- **Marker style**: `// rtmx:req REQ-ID`
- **Verify command**: `sbt test`

### Assembly
- **Exemplar**: Minimal exemplar only (no standard test framework)
- **Build time**: ~10 sec
- **Marker style**: `; rtmx:req REQ-ID`
- **Verify command**: `nasm && ./test`

### Ruby
- **Exemplar**: [rack/rack](https://github.com/rack/rack) (HTTP interface, RSpec, ~500 tests)
- **Build time**: ~1 min
- **Marker style**: `# rtmx:req REQ-ID`
- **Verify command**: `bundle exec rspec`

## Benchmark Execution Model

```
rtmx-ai/rtmx-benchmarks/
  benchmarks/
    go/
      cli-cli/          # pinned commit of cli/cli
      exemplar/          # minimal hello-rtmx-go
    rust/
      aegis-cli/        # pinned commit
      exemplar/
    python/
      requests/         # pinned commit
      exemplar/
    ...
  scripts/
    run-benchmark.sh    # orchestrator
    report.sh           # generate comparison report
  .github/workflows/
    nightly.yml         # scheduled run
  results/
    latest.json         # most recent results
    history/            # timestamped results
```

Each benchmark:
1. Clone exemplar at pinned commit
2. Add rtmx markers to tests (via patch file or fork)
3. Run `rtmx from-tests` -- verify marker count matches expected
4. Run `rtmx verify --command` -- verify test output parsing
5. Compare results to previous run -- flag regressions
