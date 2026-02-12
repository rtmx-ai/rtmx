// RTMX Whitepaper Template
// REQ-DOC-005: Professional whitepaper template with RTMX branding
//
// Usage:
//   #import "whitepaper.typ": *
//   #show: whitepaper.with(
//     title: "Your Title",
//     subtitle: "Your Subtitle",
//     author: "Author Name",
//     date: datetime.today(),
//     version: "1.0",
//   )

#import "rtmx-theme.typ": *

// Whitepaper template function
#let whitepaper(
  title: "RTMX Whitepaper",
  subtitle: none,
  author: "RTMX Engineering",
  email: "dev@rtmx.ai",
  date: datetime.today(),
  version: "1.0",
  abstract: none,
  keywords: (),
  paper: "a4",  // "a4" or "us-letter"
  doc,
) = {
  // Page setup
  let page-size = if paper == "us-letter" { "us-letter" } else { "a4" }

  set document(
    title: title,
    author: author,
    keywords: keywords,
    date: date,
  )

  set page(
    paper: page-size,
    fill: background,
    margin: (
      top: 3cm,
      bottom: 2.5cm,
      left: 2.5cm,
      right: 2.5cm,
    ),
    header: context {
      if counter(page).get().first() > 1 {
        grid(
          columns: (1fr, auto),
          align: (left, right),
          text(size: small-size, fill: secondary)[#title],
          text(size: small-size, fill: secondary)[v#version],
        )
        line(length: 100%, stroke: 0.5pt + border)
      }
    },
    footer: context {
      line(length: 100%, stroke: 0.5pt + border)
      v(0.3em)
      grid(
        columns: (1fr, auto, 1fr),
        align: (left, center, right),
        text(size: small-size, fill: secondary)[RTMX],
        text(size: small-size, fill: secondary)[
          #counter(page).display("1 / 1", both: true)
        ],
        text(size: small-size, fill: secondary)[#email],
      )
    },
  )

  // Apply theme
  show: apply-theme

  // Title page
  v(3cm)

  // RTMX Logo placeholder (text version)
  align(center)[
    #box(
      fill: slate-700,
      inset: 16pt,
      radius: 8pt,
      [
        #text(size: 32pt, fill: primary, weight: "bold", tracking: 2pt)[RTMX]
      ]
    )
  ]

  v(2cm)

  // Title
  align(center)[
    #text(size: title-size, fill: foreground, weight: "bold")[#title]
  ]

  if subtitle != none {
    v(0.5em)
    align(center)[
      #text(size: h2-size, fill: secondary)[#subtitle]
    ]
  }

  v(2cm)

  // Metadata
  align(center)[
    #text(fill: secondary)[
      #author\
      #link("mailto:" + email)[#email]\
      \
      Version #version\
      #date.display("[month repr:long] [day], [year]")
    ]
  ]

  // Keywords
  if keywords.len() > 0 {
    v(1cm)
    align(center)[
      #for keyword in keywords {
        badge(keyword)
        h(0.5em)
      }
    ]
  }

  pagebreak()

  // Abstract
  if abstract != none {
    heading(level: 1, outlined: false)[Abstract]
    text(style: "italic")[#abstract]
    v(section-spacing)
  }

  // Table of contents
  heading(level: 1, outlined: false)[Contents]
  outline(
    title: none,
    indent: 1.5em,
    fill: repeat[#text(fill: border)[.]],
  )

  pagebreak()

  // Main content
  doc
}

// Re-export theme utilities
#let info(body) = callout(body, type: "info")
#let success(body) = callout(body, type: "success")
#let warning(body) = callout(body, type: "warning")
#let error(body) = callout(body, type: "error")

// Figure styling
#let rtmx-figure(body, caption: none) = {
  figure(
    body,
    caption: if caption != none { text(fill: secondary, size: small-size)[#caption] },
  )
}

// Quote styling
#let rtmx-quote(body, attribution: none) = {
  block(
    inset: (left: 1.5em),
    stroke: (left: 3pt + primary),
    [
      #text(style: "italic")[#body]
      #if attribution != none {
        v(0.5em)
        text(fill: secondary, size: small-size)[— #attribution]
      }
    ]
  )
}

// Definition list
#let definition(term, body) = {
  [*#text(fill: primary)[#term]* — #body]
}

// Requirements reference
#let req(id) = {
  link(label(id))[#badge(id, color: primary)]
}

// Status indicators
#let status-complete = badge("COMPLETE", color: success)
#let status-partial = badge("PARTIAL", color: warning)
#let status-missing = badge("MISSING", color: error)

// Code with caption
#let code-block(code, language: none, caption: none) = {
  block(
    fill: code-bg,
    inset: 0pt,
    radius: 4pt,
    width: 100%,
    clip: true,
    [
      #if caption != none {
        block(
          fill: slate-600,
          inset: (x: 10pt, y: 6pt),
          width: 100%,
          text(size: small-size, fill: secondary)[#caption]
        )
      }
      #block(
        inset: 10pt,
        width: 100%,
        raw(code, lang: language, block: true)
      )
    ]
  )
}
