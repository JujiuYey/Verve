# Wiki Document Card Readability Design

**Date:** 2026-04-05

## Goal

Improve document card readability in the wiki folders page by prioritizing filename visibility over card density.

## Current Context

- [`DocumentCard`](/Users/jujiuyey/Projects/sag-wiki/web/src/pages/wiki/folders/_components/document-card.tsx) currently uses a compact horizontal layout with a single-line truncated title.
- The menu trigger is vertically centered on the right, which pulls visual attention away from the filename.
- [`ItemGrid`](/Users/jujiuyey/Projects/sag-wiki/web/src/pages/wiki/folders/_components/item-grid.tsx) renders document cards with the same dense grid treatment used for folders, even though documents have longer labels and more metadata.
- In practice, long filenames are cut off too aggressively, so users cannot identify documents quickly.

## Proposed Design

### 1. Give documents a wider grid than folders

Keep the folders section unchanged, but render the documents section with a less dense grid so each document card gets more horizontal space.

- Prefer fewer columns for documents at medium and large breakpoints.
- Do not change the section structure or document count header.

### 2. Rebuild the document card around filename readability

Restructure [`DocumentCard`](/Users/jujiuyey/Projects/sag-wiki/web/src/pages/wiki/folders/_components/document-card.tsx) into a top-aligned card layout.

- Keep the file-type icon on the left.
- Move the menu trigger to the top-right corner instead of vertically centering it.
- Increase card minimum height slightly so content has stable vertical spacing.
- Allow the filename to wrap to two lines with clamping instead of a single-line truncate.

### 3. Separate primary and secondary information

Use a clearer vertical information hierarchy inside the card.

- Primary: filename.
- Secondary: file size and optional chunk count.
- Status badge: placed in the lower metadata area so it reads as document state, not as the title.

This keeps the user’s eye on the filename first and still preserves operational metadata.

## Testing Strategy

- Add or update a component test for [`DocumentCard`](/Users/jujiuyey/Projects/sag-wiki/web/src/pages/wiki/folders/_components/document-card.tsx) to verify long filenames are rendered without single-line truncation.
- Add or update a component test for [`ItemGrid`](/Users/jujiuyey/Projects/sag-wiki/web/src/pages/wiki/folders/_components/item-grid.tsx) to verify the documents section uses the intended grid classes.
- Run the focused test file(s) plus project type-checking.

## Non-Goals

- No redesign of folder cards.
- No switch from card view to list view.
- No change to document actions, download flow, or delete behavior.
- No backend or API changes.

## Risks And Mitigations

- Risk: wider document cards reduce the number of items visible at once.
  Mitigation: scope the wider layout to the document section only, leaving the folder section density unchanged.

- Risk: two-line titles can make card heights feel uneven.
  Mitigation: use a stable minimum height and consistent spacing so all cards still align cleanly in the grid.

- Risk: moving the menu trigger could affect usability.
  Mitigation: keep the trigger visible, preserve its current interaction behavior, and only change its placement.
