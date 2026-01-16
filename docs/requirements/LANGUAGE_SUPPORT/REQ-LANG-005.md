# REQ-LANG-005: NUnit/xUnit Extension for C#

## Status: MISSING
## Priority: MEDIUM
## Phase: 14

## Description
System shall provide .NET test framework extensions supporting NUnit, xUnit, and MSTest that enable requirement traceability through custom attributes, published as NuGet packages for the .NET ecosystem.

## Acceptance Criteria
- [ ] NuGet package `RTMX.NUnit` published to nuget.org
- [ ] NuGet package `RTMX.xUnit` published to nuget.org
- [ ] NuGet package `RTMX.MSTest` published to nuget.org
- [ ] `[Req("REQ-XXX-NNN")]` attribute marks test methods with requirement IDs
- [ ] `[RtmxScope(Scope.Unit)]` attribute specifies test scope
- [ ] `[RtmxTechnique(Technique.Nominal)]` attribute specifies test technique
- [ ] `[RtmxEnv(Environment.Simulation)]` attribute specifies test environment
- [ ] NUnit: Custom `ITestAction` implementation captures requirement associations
- [ ] xUnit: Custom `ITestOutputHelper` extension for requirement metadata
- [ ] MSTest: Custom `TestMethodAttribute` extension for requirement tracking
- [ ] `dotnet rtmx` CLI tool extracts markers and generates RTM JSON
- [ ] `rtmx from-dotnet <results.json>` imports .NET test results into RTM database
- [ ] Support for `[Theory]` (xUnit), `[TestCase]` (NUnit), `[DataRow]` (MSTest) parameterized tests

## Technical Notes
- Shared `RTMX.Core` package contains common attributes and models
- Attributes use `[AttributeUsage(AttributeTargets.Method)]` with `AllowMultiple = true`
- NUnit extension: `ITestAction.BeforeTest()` and `AfterTest()` hooks
- xUnit extension: Custom `IXunitTestCaseDiscoverer` for enriched test case metadata
- MSTest extension: `TestMethodAttribute` subclass with result capture
- .NET CLI tool: `dotnet tool install -g rtmx-dotnet`
- Consider source generators for compile-time marker validation (.NET 6+)
- JSON output follows RTMX marker specification schema (REQ-LANG-007)
- Support .NET Standard 2.0 for broad framework compatibility

## Test Cases
1. `tests/test_lang_dotnet.py::test_req_attribute_parsing` - Parse [Req] attributes
2. `tests/test_lang_dotnet.py::test_scope_technique_env_attributes` - Parse extended attributes
3. `tests/test_lang_dotnet.py::test_nunit_test_action_capture` - NUnit ITestAction hooks
4. `tests/test_lang_dotnet.py::test_xunit_discoverer_enrichment` - xUnit metadata enrichment
5. `tests/test_lang_dotnet.py::test_mstest_attribute_extension` - MSTest integration
6. `tests/test_lang_dotnet.py::test_parameterized_test_support` - Theory/TestCase/DataRow
7. `tests/test_lang_dotnet.py::test_dotnet_cli_extraction` - CLI tool output
8. `tests/test_lang_dotnet.py::test_from_dotnet_import` - CLI import command
9. `tests/test_lang_dotnet.py::test_netstandard_compatibility` - Framework compatibility

## Dependencies
- REQ-LANG-007: Language-agnostic marker annotation spec

## Blocks
- None

## Effort
3.5 weeks
