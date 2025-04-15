// Make typst look like LaTeX 

#let my-template(body) = {
  // Global document settings:
  set page(margin: 1.25in)
  set par(leading: 0.55em, justify: true)
  set text(12pt, font: "New Computer Modern")
  // show raw: set text(font: "JetBrains Mono")
  // show par: set block(spacing: 0.55em)
  // show par: set par(spacing: 0.55em)
  show heading: set block(above: 1.4em, below: 1em)
  show heading.where(level: 1): set text(size: 14pt)
  show heading.where(level: 2): set text(size: 13pt)
  show heading.where(level: 3): set text(size: 12pt)

  let argmax = math.op("argmax", limits: true)
  set math.vec(delim: "[")
  set math.mat(delim: "[")
  set heading(numbering: "1.1.")
  
  // Now insert the body content:
  body
}

#let fixedHeightMat = (mat, entryHeight) => {
  // return if matrix is empty
  if mat.rows.len() == 0 or mat.rows.at(0).len() == 0 {return mat} 
  let firstEntry = mat.rows.at(0).at(0)
  // return if entries are already wrapped in blocks (using the first entry) to avoid infinite recursion
  if firstEntry.func() == block {return mat} 
  // wrap every entry in block with fixed height
  let entries = mat.rows.map(row => {
    row.map(entry => {
      block(entry, height: entryHeight)
    })
  })
  set align(horizon)
  let delim = if mat.has("delim") {mat.delim} else {"["}
  math.mat(..entries, delim: delim)
}

// Equation numbering
#let numbered_eq(content) = math.equation(
    block: true,
    numbering: "(1)",
    content,
)

#let styled-link = (
  url,
  label,
  color: blue,
  uline: true,
  italic: false,
) => {
  // Apply color and optional italics to the text
  let styled-text = text(
    label,
    fill: blue,
    style: if italic { "italic" } else { "normal" },
  )

  // Optionally add underlining
  if uline {
    styled-text = underline(styled-text)
  }

  // Create the hyperlink
  // link(url)[styled-text]
  link(url, styled-text)
}


// First page stuff
#let title(content) = align(center, text(1.95em)[#content])
#let subtitle(content) = align(center, text(1.4em)[#content])
#let author(content) = align(center, par(text(size: 0.9em,[#content]), justify: false))
#let long_date(d, m, y) = datetime(year: y, month: m, day: d, ).display("[weekday], [day] [month repr:long] [year]")
#let short_date(d, m, y) = datetime(year: y, month: m, day: d).display("[day]-[month]-[year]")

// Math stuff
#let partial(f, x) = $(diff #f) / (diff #x)$
#let deriv(f, x) = $(dif #f) / (dif #x)$
#let summation(i, a, b) = $limits(sum)_(#i=#a)^(#b)$
#let product(i, a, b) = $limits(\u{220F})_(#i=#a)^(#b)$
#let integral(x, a, b, content) = $limits(\u{222B})_(#a)^(#b) #content dif #x$
#let limit(x, a) = $limits(lim)_(#x -> #a) $

#let hello(a, b) = grid(
  columns: (1fr, 1fr, 1fr),
  "",align(center)[#a], align(right)[#b]
)

#let colred(x) = text(fill: red, $#x$) 
