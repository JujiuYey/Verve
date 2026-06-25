import data from "@/app/dashboard/data.json";
import type { Department } from "@/api/system/department";
import { ChartAreaInteractive } from "@/pages/dashboard/chart-area-interactive";
import { SectionCards } from "@/pages/dashboard/section-cards";
import { DataTable } from "@/pages/system/department/_components/data-table";

const dashboardRows: Department[] = data.slice(0, 12).map((item) => ({
  id: String(item.id),
  name: item.header,
  description: `${item.type} · ${item.status} · 审核人：${item.reviewer}`,
  created_at: new Date().toISOString(),
  updated_at: new Date().toISOString(),
}));

export function DashboardPage() {
  return (
    <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
      <SectionCards />
      <div className="px-4 lg:px-6">
        <ChartAreaInteractive />
      </div>
      <DataTable data={dashboardRows} />
    </div>
  );
}
