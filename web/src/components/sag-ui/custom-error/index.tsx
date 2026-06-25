import { useNavigate } from "@tanstack/react-router";

import { Button } from "@/components/ui/button";

interface CustomErrorProps {
  code: number;
  subtitle: string;
  error: string;
  children?: React.ReactNode;
}

export function CustomError({ code, subtitle, error, children }: CustomErrorProps) {
  const navigate = useNavigate();

  return (
    <div className="max-w-2xl mx-auto text-center">
      <h1 className="font-bold text-8xl">{code}</h1>
      <h2 className="mt-4 text-2xl font-bold">{subtitle}</h2>
      <p className="text-stone-400">{error}</p>

      <footer className="mt-8">
        {children ?? (
          <div className="flex justify-center gap-2">
            <Button variant="outline" onClick={() => window.history.back()}>
              返回
            </Button>
            <Button onClick={() => navigate({ to: "/" })}>返回首页</Button>
          </div>
        )}
      </footer>
    </div>
  );
}
