# TreeSelect Clear Option Design

**Date:** 2026-04-04

## Goal

Allow form fields that use `TreeSelect` to explicitly clear their current selection and submit an empty value again.

## Design

- Add optional clear-entry support to `TreeSelect`.
- When enabled, the popover renders a top-level action above the tree items.
- Clicking the clear action emits an empty string value and `null` node, then closes the popover.
- Keep this behavior opt-in so existing `TreeSelect` usages remain unchanged.
- Enable the clear action in the department form with the label `无上级部门`.
- Map the cleared value back to `undefined` in the form field so submission behavior stays consistent with the current API contract.

## Testing

- Add a component test proving the clear entry is rendered when enabled.
- Add a component test proving selecting the clear entry calls `onValueChange("", null)` and closes the popover.
- Run the focused `TreeSelect` tests plus TypeScript type-checking.
