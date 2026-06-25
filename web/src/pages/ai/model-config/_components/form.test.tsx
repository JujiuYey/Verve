import { fireEvent, render, screen } from "@testing-library/react";
import { useState } from "react";
import { beforeAll, describe, expect, it, vi } from "vitest";

import type { ModelConfig } from "@/api/ai/model-config";

const mutateAsync = vi.fn();

vi.mock("@/api/ai/model-config", () => ({
  useCreateModelConfig: () => ({
    isPending: false,
    mutateAsync,
  }),
  useUpdateModelConfig: () => ({
    isPending: false,
    mutateAsync,
  }),
}));

import { ModelConfigForm } from "./form";

beforeAll(() => {
  class ResizeObserverMock {
    observe() {}
    unobserve() {}
    disconnect() {}
  }

  vi.stubGlobal("ResizeObserver", ResizeObserverMock);
});

const baseConfig: ModelConfig = {
  id: "config-1",
  vendor: "custom",
  name: "Custom Model",
  api_key: "sk-test",
  base_url: "https://example.com/v1",
  model_type: "chat",
  model: "custom-model-v1",
  temperature: 0.7,
  top_p: 0.9,
  max_tokens: undefined,
  top_k: undefined,
  is_active: true,
  is_default: false,
  created_at: "2026-03-30T00:00:00Z",
  updated_at: "2026-03-30T00:00:00Z",
};

function ReselectHarness() {
  const [config, setConfig] = useState<ModelConfig>(baseConfig);

  return (
    <>
      <button
        type="button"
        onClick={() => {
          setConfig({ ...baseConfig });
        }}
      >
        reselect
      </button>

      <ModelConfigForm key={config.id} config={config} onRefresh={vi.fn()} />
    </>
  );
}

describe("ModelConfigForm", () => {
  it("preserves the edited model when the same config is selected again", () => {
    render(<ReselectHarness />);

    const modelInput = screen.getByPlaceholderText("请输入自定义模型名称");
    fireEvent.change(modelInput, { target: { value: "custom-model-v2" } });

    fireEvent.click(screen.getByRole("button", { name: "reselect" }));

    expect(screen.getByPlaceholderText("请输入自定义模型名称")).toHaveValue("custom-model-v2");
  });
});
