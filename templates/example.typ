// Example RTMX Whitepaper
// Demonstrates the whitepaper template and theme
//
// Compile with: typst compile example.typ example.pdf

#import "whitepaper.typ": *

#show: whitepaper.with(
  title: "RTMX Template Demo",
  subtitle: "A Professional Whitepaper Template",
  author: "RTMX Engineering",
  email: "dev@rtmx.ai",
  version: "1.0",
  keywords: ("requirements", "traceability", "documentation"),
  abstract: [
    This document demonstrates the RTMX whitepaper template, showcasing
    the dark theme, typography, code blocks, tables, and other formatting
    elements. The template is designed for technical documentation with
    a professional, modern appearance matching the rtmx.ai website.
  ],
)

= Introduction

RTMX (Requirements Traceability Matrix) is a toolkit for managing software
requirements with full traceability from specification to implementation.
This document demonstrates the whitepaper template styling.

== Purpose

The whitepaper template provides:

- Consistent branding with rtmx.ai
- Professional typography for technical documents
- Code blocks with syntax highlighting
- Custom callout boxes for important information
- Table styling for data presentation

= Theme Elements

== Typography

The template uses *Inter* for body text and headings, with *JetBrains Mono*
for code. Text colors are carefully chosen for readability on the dark
background.

=== Heading Levels

Headings use a hierarchy of sizes and colors:
- Level 1: Sky blue (#0ea5e9), 24pt
- Level 2: White (#f1f5f9), 20pt
- Level 3: Slate (#94a3b8), 16pt

== Code Blocks

Inline code uses a `monospace font` with a subtle background.

#code-block(
  ```python
  from rtmx import RTMDatabase

  # Load the requirements database
  db = RTMDatabase.load("docs/rtm_database.csv")

  # Check completion status
  print(f"Complete: {db.complete_count}/{db.total_count}")
  ```,
  language: "python",
  caption: "Loading and checking RTM database"
)

== Callouts

#info[
  *Information*: Use callouts to highlight important information that
  readers should pay attention to.
]

#success[
  *Success*: Indicate successful outcomes or best practices with
  green callouts.
]

#warning[
  *Warning*: Warn readers about potential issues or deprecated
  features with amber callouts.
]

#error[
  *Error*: Alert readers to critical issues or breaking changes
  with red callouts.
]

== Tables

#rtmx-table(
  columns: (auto, 1fr, auto),
  [*Requirement*], [*Description*], [*Status*],
  [REQ-CORE-001], [System shall load RTM from CSV], status-complete,
  [REQ-CLI-001], [CLI shall show status summary], status-complete,
  [REQ-SYNC-001], [CRDT-based offline sync], status-partial,
  [REQ-ZT-001], [Zitadel OIDC integration], status-missing,
)

== Quotes

#rtmx-quote(
  [Requirements traceability is not just about compliance—it's about
  understanding the *why* behind every line of code.],
  attribution: "RTMX Engineering"
)

== Definitions

#definition("RTM")[Requirements Traceability Matrix - a document that maps
requirements to their implementation and verification artifacts.]

#definition("CRDT")[Conflict-free Replicated Data Type - a data structure
that enables distributed systems to converge without coordination.]

= Conclusion

The RTMX whitepaper template provides all the elements needed for
professional technical documentation. Use it for:

- Product whitepapers
- Technical specifications
- Architecture documents
- Research papers

For more information, visit #link("https://rtmx.ai")[rtmx.ai].
