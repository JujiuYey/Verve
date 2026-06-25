import { CustomError } from "@/components/sag-ui";

export function NotFoundPage() {
  return (
    <div className="flex items-center justify-center h-screen">
      <CustomError
        code={404}
        subtitle="页面未找到"
        error="您访问的页面可能已被删除、更名或暂时不可用。"
      />
    </div>
  );
}
