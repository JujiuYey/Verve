import { createFileRoute } from "@tanstack/react-router";

import { AppSettingPage } from "@/pages/common/app-setting";

export const Route = createFileRoute("/_layout/app-setting")({
  component: AppSettingPage,
});
