import logo from "@/assets/logo.svg";

import { LoginForm } from "./login-form";

export function LoginPage() {
  return (
    <div className="flex h-screen w-full justify-center items-center">
      <div className="bg-background flex w-full items-center justify-center p-8 lg:w-1/2">
        <div className="w-full max-w-sm">
          <div className="mb-6 flex items-center gap-3">
            <img src={logo} alt="SAG Wiki" className="size-10 shrink-0 rounded-lg object-contain" />
            <div>
              <h1 className="text-foreground text-xl font-bold">SAG Wiki</h1>
              <p className="text-muted-foreground text-xs">智能知识运营中枢</p>
            </div>
          </div>

          <LoginForm />
        </div>
      </div>
    </div>
  );
}
