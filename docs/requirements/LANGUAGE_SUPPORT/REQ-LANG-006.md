# REQ-LANG-006: RSpec Integration for Ruby

## Status: MISSING
## Priority: LOW
## Phase: 14

## Description
System shall provide RSpec integration for Ruby projects that enables requirement traceability through DSL methods and comment markers, with a custom formatter for RTM-compatible output.

## Acceptance Criteria
- [ ] Ruby gem `rtmx-rspec` published to rubygems.org
- [ ] DSL method `rtmx_req "REQ-XXX-NNN"` within `describe`/`context`/`it` blocks
- [ ] Comment marker format: `# @rtmx REQ-XXX-NNN` parsed before examples
- [ ] Extended DSL: `rtmx_req "REQ-XXX-NNN", scope: :unit, technique: :nominal, env: :simulation`
- [ ] Extended comment: `# @rtmx REQ-XXX-NNN scope=unit technique=nominal`
- [ ] Custom RSpec formatter `RTMX::Formatter` outputs RTM-compatible JSON
- [ ] Requirement inheritance: nested groups inherit parent requirements
- [ ] Support for `shared_examples` with requirement propagation
- [ ] Tree-sitter Ruby parser extracts markers from spec file AST
- [ ] `rtmx from-rspec <results.json>` imports RSpec results into RTM database
- [ ] RSpec metadata integration via `:rtmx` tag for flexible marker attachment

## Technical Notes
- Gem structure: `lib/rtmx/rspec.rb` with `RTMX::RSpec` module
- DSL implementation via `RSpec.configure` block extending example groups
- Comment marker extraction requires parsing source files separately from execution
- Formatter inherits from `RSpec::Core::Formatters::BaseFormatter`
- Hook into `example_started`, `example_passed`, `example_failed` events
- Requirement inheritance: merge parent metadata with child overrides
- `shared_examples` support: track requirement association at include site
- Consider Minitest support in future version
- Output format: JSON array with example metadata and requirement associations

## Test Cases
1. `tests/test_lang_rspec.py::test_dsl_method_extraction` - Parse rtmx_req DSL
2. `tests/test_lang_rspec.py::test_comment_marker_extraction` - Parse # @rtmx comments
3. `tests/test_lang_rspec.py::test_extended_dsl_attributes` - Parse scope/technique/env
4. `tests/test_lang_rspec.py::test_formatter_json_output` - Validate formatter output
5. `tests/test_lang_rspec.py::test_requirement_inheritance` - Nested group inheritance
6. `tests/test_lang_rspec.py::test_shared_examples_propagation` - shared_examples support
7. `tests/test_lang_rspec.py::test_metadata_tag_integration` - :rtmx tag support
8. `tests/test_lang_rspec.py::test_from_rspec_import` - CLI import command

## Dependencies
- REQ-LANG-007: Language-agnostic marker annotation spec

## Blocks
- None

## Effort
2.5 weeks
