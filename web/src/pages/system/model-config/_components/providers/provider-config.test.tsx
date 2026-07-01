import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { ProviderConfig } from "./provider-config";

const mutateAsync = vi.fn();

vi.mock("@/api", () => ({
  useUpdateAIPlatformConfig: () => ({
    mutateAsync,
    isPending: false,
  }),
}));

describe("ProviderConfig", () => {
  it("offers a clear-key action and sends clear_api_key when confirmed", async () => {
    const platform = {
      id: "p-1",
      name: "DeepSeek",
      provider_type: "openai_compatible",
      default_base_url: "https://api.deepseek.com/v1",
      base_url: "https://api.deepseek.com/v1",
      api_key_hint: "sk****1234",
      model_list_path: "/models",
      auth_scheme: "bearer",
      enabled: true,
      sort_order: 0,
    };

    mutateAsync.mockReset();
    mutateAsync.mockResolvedValueOnce(undefined);

    render(<ProviderConfig platform={platform as never} />);

    fireEvent.click(screen.getByRole("button", { name: "清空密钥" }));
    fireEvent.click(screen.getByRole("button", { name: "确认清空" }));

    await waitFor(() => {
      expect(mutateAsync).toHaveBeenCalledWith({
        platformId: "p-1",
        data: {
          base_url: "https://api.deepseek.com/v1",
          clear_api_key: true,
        },
      });
    });
  });
});
