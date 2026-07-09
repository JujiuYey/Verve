import { createFileRoute } from "@tanstack/react-router";

import { FeynmanExercisePage } from "@/pages/learning/feynman";

interface SearchSchema {
  agentInstanceId?: string;
  rootFolderId?: string;
  rootFolderName?: string;
}

export const Route = createFileRoute("/_layout/learn/feynman")({
  component: FeynmanExercisePage,
  validateSearch: (search: Record<string, unknown>): SearchSchema => ({
    agentInstanceId: search.agentInstanceId as string | undefined,
    rootFolderId: search.rootFolderId as string | undefined,
    rootFolderName: search.rootFolderName as string | undefined,
  }),
});
