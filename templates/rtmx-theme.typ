// RTMX Theme Configuration for Typst
// REQ-DOC-005: Typst whitepaper template matching rtmx.ai dark theme
//
// Color palette derived from rtmx.ai website (Tailwind CSS colors)

// Primary colors
#let slate-800 = rgb("#1e293b")  // Background
#let slate-700 = rgb("#334155")  // Secondary background
#let slate-600 = rgb("#475569")  // Borders
#let slate-400 = rgb("#94a3b8")  // Muted text
#let slate-300 = rgb("#cbd5e1")  // Secondary text
#let slate-100 = rgb("#f1f5f9")  // Primary text

// Accent colors
#let sky-500 = rgb("#0ea5e9")    // Primary accent (links, headings)
#let sky-400 = rgb("#38bdf8")    // Hover state
#let green-500 = rgb("#22c55e")  // Success
#let amber-500 = rgb("#f59e0b")  // Warning
#let red-500 = rgb("#ef4444")    // Error

// Semantic aliases
#let background = slate-800
#let foreground = slate-100
#let primary = sky-500
#let secondary = slate-400
#let success = green-500
#let warning = amber-500
#let error = red-500
#let border = slate-600
#let code-bg = slate-700

// Typography
#let heading-font = "Inter"
#let body-font = "Inter"
#let code-font = "JetBrains Mono"

// Font sizes
#let title-size = 28pt
#let h1-size = 24pt
#let h2-size = 20pt
#let h3-size = 16pt
#let body-size = 11pt
#let small-size = 9pt
#let code-size = 10pt

// Spacing
#let page-margin = 2.5cm
#let section-spacing = 1.5em
#let paragraph-spacing = 0.8em

// Apply theme to document
#let apply-theme(doc) = {
  set page(
    fill: background,
    margin: page-margin,
  )

  set text(
    font: body-font,
    size: body-size,
    fill: foreground,
  )

  set par(
    leading: 0.8em,
    justify: true,
  )

  // Headings
  show heading.where(level: 1): set text(
    font: heading-font,
    size: h1-size,
    fill: primary,
    weight: "bold",
  )

  show heading.where(level: 2): set text(
    font: heading-font,
    size: h2-size,
    fill: foreground,
    weight: "semibold",
  )

  show heading.where(level: 3): set text(
    font: heading-font,
    size: h3-size,
    fill: secondary,
    weight: "medium",
  )

  // Links
  show link: set text(fill: primary)
  show link: underline

  // Code blocks
  show raw.where(block: true): block.with(
    fill: code-bg,
    inset: 10pt,
    radius: 4pt,
    width: 100%,
  )

  show raw: set text(
    font: code-font,
    size: code-size,
  )

  // Inline code
  show raw.where(block: false): box.with(
    fill: code-bg,
    inset: (x: 4pt, y: 2pt),
    radius: 2pt,
  )

  doc
}

// Table styling
#let rtmx-table(columns: auto, ..args) = {
  table(
    columns: columns,
    stroke: (x, y) => if y == 0 { (bottom: 1pt + border) } else { none },
    fill: (x, y) => if y == 0 { slate-700 } else if calc.odd(y) { slate-700.lighten(5%) } else { none },
    inset: 8pt,
    ..args
  )
}

// Callout boxes
#let callout(body, type: "info") = {
  let (icon, color) = if type == "info" {
    ("ℹ️", primary)
  } else if type == "success" {
    ("✓", success)
  } else if type == "warning" {
    ("⚠", warning)
  } else if type == "error" {
    ("✗", error)
  } else {
    ("•", secondary)
  }

  block(
    fill: color.lighten(85%),
    stroke: (left: 3pt + color),
    inset: 12pt,
    radius: (right: 4pt),
    width: 100%,
    [#text(fill: color, weight: "bold")[#icon] #body]
  )
}

// Highlight box
#let highlight-box(body, title: none) = {
  block(
    fill: slate-700,
    stroke: 1pt + border,
    inset: 12pt,
    radius: 4pt,
    width: 100%,
    [
      #if title != none {
        [#text(fill: primary, weight: "bold")[#title]\ ]
      }
      #body
    ]
  )
}

// Badge/tag component
#let badge(label, color: primary) = {
  box(
    fill: color.lighten(80%),
    stroke: 1pt + color,
    inset: (x: 6pt, y: 2pt),
    radius: 2pt,
    text(fill: color, size: small-size, weight: "medium")[#label]
  )
}
