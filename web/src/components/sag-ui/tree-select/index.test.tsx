import { fireEvent, render, screen } from "@testing-library/react";
import { useState } from "react";
import { beforeAll, describe, expect, it, vi } from "vitest";

import { TreeSelect, type TreeSelectItem, type TreeSelectNode } from ".";

interface TestNode extends TreeSelectNode<TestNode> {
  name: string;
}

beforeAll(() => {
  class ResizeObserverMock {
    observe() {}
    unobserve() {}
    disconnect() {}
  }

  vi.stubGlobal("ResizeObserver", ResizeObserverMock);
});

const items: TreeSelectItem<TestNode>[] = [
  {
    value: "engineering",
    label: "研发部",
    node: { id: "engineering", name: "研发部" },
    children: [
      {
        value: "frontend",
        label: "前端组",
        node: { id: "frontend", name: "前端组" },
      },
    ],
  },
];

describe("TreeSelect", () => {
  it("shows tree items after opening the select", () => {
    render(<TreeSelect items={items} onValueChange={vi.fn()} placeholder="请选择部门" />);

    fireEvent.click(screen.getByRole("combobox"));

    expect(screen.getByText("研发部")).toBeVisible();
  });

  it("closes the list after selecting an item", () => {
    render(<TreeSelect items={items} onValueChange={vi.fn()} placeholder="请选择部门" />);

    const trigger = screen.getByRole("combobox");

    fireEvent.click(trigger);

    fireEvent.click(screen.getByText("研发部"));

    expect(trigger).toHaveAttribute("aria-expanded", "false");
  });

  it("clears the selected value when the clear action is chosen", () => {
    const onValueChange = vi.fn();

    function Harness() {
      const [value, setValue] = useState<string | undefined>("engineering");

      return (
        <TreeSelect
          items={items}
          value={value}
          onValueChange={(nextValue, node) => {
            onValueChange(nextValue, node);
            setValue(nextValue || undefined);
          }}
          placeholder="请选择部门"
          allowClear
          clearLabel="无上级部门"
        />
      );
    }

    render(<Harness />);

    const trigger = screen.getByRole("combobox");

    fireEvent.click(trigger);

    expect(screen.getByText("无上级部门")).toBeVisible();

    fireEvent.click(screen.getByText("无上级部门"));

    expect(onValueChange).toHaveBeenCalledWith("", null);
    expect(trigger).toHaveTextContent("请选择部门");
    expect(trigger).toHaveAttribute("aria-expanded", "false");
  });
});
