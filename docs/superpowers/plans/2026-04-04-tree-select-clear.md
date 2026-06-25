# TreeSelect Clear Option Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add an opt-in clear action to `TreeSelect` and wire it into the department parent selector so users can reset the field to empty.

**Architecture:** Extend `TreeSelect` with a small, optional clear entry rendered inside the existing popover. Keep the component API backward compatible, then adapt the department form to translate the cleared value back to `undefined`.

**Tech Stack:** React 19, Vitest, Testing Library, TypeScript, existing project UI primitives

---

### Task 1: Add Failing Coverage For Clear Behavior

**Files:**
- Modify: `web/src/components/sag-ui/tree-select/index.test.tsx`
- Test: `web/src/components/sag-ui/tree-select/index.test.tsx`

- [ ] **Step 1: Write the failing test**

```tsx
it("clears the selected value when the clear action is chosen", () => {
  const onValueChange = vi.fn();

  render(
    <TreeSelect
      items={items}
      value="engineering"
      onValueChange={onValueChange}
      placeholder="请选择部门"
      allowClear
      clearLabel="无上级部门"
    />,
  );

  fireEvent.click(screen.getByRole("combobox"));
  fireEvent.click(screen.getByText("无上级部门"));

  expect(onValueChange).toHaveBeenCalledWith("", null);
  expect(screen.getByRole("combobox")).toHaveTextContent("请选择部门");
});
```

- [ ] **Step 2: Run test to verify it fails**

Run: `pnpm vitest run src/components/sag-ui/tree-select/index.test.tsx`
Expected: FAIL because `TreeSelect` does not yet expose a clear action.

- [ ] **Step 3: Write minimal implementation**

```tsx
export interface TreeSelectProps<T extends TreeSelectNode<T>> {
  allowClear?: boolean;
  clearLabel?: string;
}

const handleClear = () => {
  onValueChange("", null);
  setOpen(false);
};
```

- [ ] **Step 4: Run test to verify it passes**

Run: `pnpm vitest run src/components/sag-ui/tree-select/index.test.tsx`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add web/src/components/sag-ui/tree-select/index.tsx web/src/components/sag-ui/tree-select/index.test.tsx
git commit -m "feat: add clear action to tree select"
```

### Task 2: Wire Clear Behavior Into Department Form

**Files:**
- Modify: `web/src/pages/system/department/_components/department-form-modal.tsx`
- Test: `web/src/components/sag-ui/tree-select/index.test.tsx`

- [ ] **Step 1: Write the failing integration expectation**

```tsx
<TreeSelect
  allowClear
  clearLabel="无上级部门"
  onValueChange={(value) => field.handleChange(value || undefined)}
/>
```

- [ ] **Step 2: Run type-check to verify the current call site is not yet updated**

Run: `pnpm lint:type`
Expected: PASS before the change and PASS after the change. This step exists to confirm the integration stays type-safe.

- [ ] **Step 3: Write minimal implementation**

```tsx
<TreeSelect
  items={departmentTreeItems}
  value={field.state.value}
  onValueChange={value => field.handleChange(value || undefined)}
  placeholder="请选择上级部门（可选）"
  className="w-full"
  allowClear
  clearLabel="无上级部门"
/>
```

- [ ] **Step 4: Run verification**

Run: `pnpm vitest run src/components/sag-ui/tree-select/index.test.tsx && pnpm lint:type`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add web/src/pages/system/department/_components/department-form-modal.tsx
git commit -m "feat: support clearing department parent selection"
```
