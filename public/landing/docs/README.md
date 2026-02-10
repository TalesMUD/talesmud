# TalesMUD Documentation Website

This directory contains the HTML documentation for TalesMUD, converted from Markdown sources.

## Structure

- `index.html` - Documentation hub page listing all guides
- `docs-style.css` - Shared stylesheet matching Veilspan landing page aesthetic
- `01-getting-started.html` through `13-glossary.html` - Individual documentation pages

## Features

- **Consistent Styling**: Matches the main Veilspan landing page with:
  - Dark fantasy aesthetic (deep blacks, green/amber accents)
  - Cinzel fonts for headings
  - Cormorant Garamond for body text
  - Fira Code for code/monospace elements
  - Scanline overlay effects

- **Navigation**:
  - Top navigation bar linking back to landing page
  - Sidebar navigation (sticky on desktop)
  - Previous/Next links at bottom of each page
  - Active page highlighting in sidebar

- **Responsive Design**:
  - Desktop: Two-column layout with sticky sidebar
  - Tablet: Single column with sidebar above content
  - Mobile: Optimized spacing and touch targets

- **Content Formatting**:
  - Tables with hover effects
  - Syntax-highlighted code blocks
  - Organized lists
  - Info boxes (tips, warnings)
  - Proper semantic HTML

## Source Files

Markdown source files are located in `/mud_client_docs/` directory at the repository root.

## Regenerating

To regenerate the HTML files from markdown sources, run the Python conversion script that processes all markdown files and applies the template.

## Links

- Main landing page: `/public/landing/index.html`
- Game client: `/public/landing/play/`
- Documentation: `/public/landing/docs/`
