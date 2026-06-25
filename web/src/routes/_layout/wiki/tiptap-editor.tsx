import { createFileRoute } from "@tanstack/react-router";

import { CanvasTiptapPage } from "@/pages/wiki/tiptap-editor";

interface SearchSchema {
  docId?: string;
}

export const Route = createFileRoute("/_layout/wiki/tiptap-editor")({
  component: CanvasTiptapPage,
  validateSearch: (search: Record<string, unknown>): SearchSchema => {
    return {
      docId: search.docId as string | undefined,
    };
  },
});
