# Wiki Upload Dialog TreeSelect Design

**Date:** 2026-04-04

## Goal

Replace the folder picker in the wiki upload dialog with `TreeSelect`, while allowing `FoldersPage` to pass its already-loaded folder tree into the dialog and keeping `DocumentsPage` working through an internal fallback fetch.

## Current Context

- [`UploadDialog`](/Users/jujiuyey/Projects/sag-wiki/web/src/pages/wiki/documents/_components/upload-dialog.tsx) currently loads a flat folder list with `folderApi.list()` when the dialog opens and renders a plain `Select`.
- [`FoldersPage`](/Users/jujiuyey/Projects/sag-wiki/web/src/pages/wiki/folders/index.tsx) already has a tree navigation use case, but the tree data is fetched inside [`FolderTree`](/Users/jujiuyey/Projects/sag-wiki/web/src/pages/wiki/folders/_components/folder-tree.tsx), so the parent cannot reuse that data for the upload dialog.
- [`DocumentsPage`](/Users/jujiuyey/Projects/sag-wiki/web/src/pages/wiki/documents/index.tsx) does not currently own folder tree data.

## Proposed Design

### 1. UploadDialog accepts optional tree data

Add `folderTree?: FolderTreeNode[]` to `UploadDialogProps`.

- If `folderTree` is provided, `UploadDialog` uses it directly.
- If `folderTree` is not provided, `UploadDialog` fetches tree data with `folderApi.tree()` when opened.
- The dialog keeps `selectedFolderId`, `defaultFolderId`, file selection, upload submission, and reset behavior unchanged.

### 2. UploadDialog uses TreeSelect instead of Select

Inside `UploadDialog`:

- Convert `FolderTreeNode[]` to `TreeSelectItem<FolderTreeNode>[]` with a small local helper.
- Replace the flat `Select` field with `TreeSelect`.
- Keep the field required.
- Do not enable `allowClear`, because uploading still requires an explicit destination folder.
- Continue showing the current placeholder when nothing is selected.
- If the tree is empty, reuse the existing empty-state behavior through `TreeSelect`'s `emptyMessage`.

### 3. FoldersPage lifts tree data and reuses it

In `FoldersPage`:

- Add parent-owned `folderTreeData` state.
- Load it with `folderApi.tree()` alongside the existing folder list workflow.
- Pass `folderTreeData` into both `FolderTree` and `UploadDialog`.
- Refresh the tree data after folder create, update, or delete operations so the left tree and upload dialog stay in sync.

This avoids an extra tree request each time the upload dialog opens from the folders page.

### 4. FolderTree supports controlled data input

Extend `FolderTree` with an optional prop such as `data?: FolderTreeNode[]`.

- When `data` is provided, render it directly.
- When `data` is absent, keep the current internal `folderApi.tree()` loading behavior.

This keeps existing call sites compatible while letting `FoldersPage` reuse a single tree source.

## Data Flow

### Folders page path

1. `FoldersPage` loads `folderTreeData`.
2. `FoldersPage` passes `folderTreeData` to `FolderTree`.
3. `FoldersPage` passes the same `folderTreeData` to `UploadDialog`.
4. `UploadDialog` converts the tree data to `TreeSelectItem[]` and updates `selectedFolderId` when the user chooses a folder.

### Documents page path

1. `DocumentsPage` opens `UploadDialog` without `folderTree`.
2. `UploadDialog` fetches tree data with `folderApi.tree()` when opened.
3. `UploadDialog` converts the fetched tree to `TreeSelectItem[]` and proceeds normally.

## Testing Strategy

Add focused component tests for `UploadDialog`:

1. When `folderTree` is passed, opening the picker shows tree nodes and selecting one updates the chosen folder.
2. When `folderTree` is not passed, opening the dialog triggers `folderApi.tree()` and renders the returned tree nodes.

Verification commands:

- `pnpm vitest run src/pages/wiki/documents/_components/upload-dialog.test.tsx`
- `pnpm lint:type`

## Non-Goals

- No changes to `folderApi` shapes or server APIs.
- No page-level refactor for `DocumentsPage`.
- No new shared hook for tree loading in this change.
- No change to upload validation, upload request flow, or post-upload refresh logic.

## Risks And Mitigations

- Risk: `FoldersPage` tree data becomes stale after folder mutations.
  Mitigation: refresh the lifted tree data after create, update, and delete flows.

- Risk: `UploadDialog` fallback fetch duplicates logic.
  Mitigation: keep the fallback minimal and only use it when `folderTree` is absent.

- Risk: Tree conversion logic gets duplicated later.
  Mitigation: keep the helper small and local for now; only extract it if another call site appears.
