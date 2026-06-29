import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { CreatePlatformDialog } from "./create-platform-dialog";

const mutateAsync = vi.fn();

vi.mock("@/api/ai/model-config", () => ({
  useCreateAIPlatform: () => ({
    mutateAsync,
    isPending: false,
  }),
}));

describe("CreatePlatformDialog", () => {
  it("requires api key before enabling create and submits api_key in payload", async () => {
    const onOpenChange = vi.fn();
    const onCreated = vi.fn();

    mutateAsync.mockResolvedValueOnce({ id: "platform-1" });

    render(
      <CreatePlatformDialog open onOpenChange={onOpenChange} onCreated={onCreated} />,
    );

    const createButton = screen.getByRole("button", { name: "创建" });
    expect(createButton).toBeDisabled();
    expect(screen.getByText("API Key 为必填项")).toBeInTheDocument();

    fireEvent.change(screen.getByPlaceholderText("例如: My LLM Platform"), {
      target: { value: "uniapi" },
    });
    fireEvent.change(screen.getByPlaceholderText("https://api.example.com/v1"), {
      target: { value: "https://api.uniapi.io" },
    });

    expect(createButton).toBeDisabled();

    fireEvent.change(screen.getByPlaceholderText("请输入 API Key"), {
      target: { value: "sk-test-123" },
    });

    expect(createButton).not.toBeDisabled();

    fireEvent.click(createButton);

    await waitFor(() => {
      expect(mutateAsync).toHaveBeenCalledWith({
        name: "uniapi",
        provider_type: "openai_compatible",
        default_base_url: "https://api.uniapi.io",
        base_url: "https://api.uniapi.io",
        api_key: "sk-test-123",
        model_list_path: "/models",
        auth_scheme: "bearer",
      });
    });

    expect(onOpenChange).toHaveBeenCalledWith(false);
    expect(onCreated).toHaveBeenCalledWith("platform-1");
  });
});
