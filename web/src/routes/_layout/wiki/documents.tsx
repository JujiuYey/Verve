import { createFileRoute } from "@tanstack/react-router";

import { DocumentsPage } from "@/pages/wiki/documents";

export const Route = createFileRoute("/_layout/wiki/documents")({
  component: DocumentsPage,
});
