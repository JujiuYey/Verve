import { createFileRoute } from "@tanstack/react-router";

import { AccountPage } from "@/pages/common/account";

export const Route = createFileRoute("/_layout/account")({
  component: AccountPage,
});
