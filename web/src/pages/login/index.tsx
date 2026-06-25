import { BookOpen, Boxes, Files, Sparkles } from "lucide-react";
import { useEffect, useState } from "react";

import logo from "@/assets/logo.svg";

import { LoginForm } from "./login-form";

const slides = [
  {
    title: "沉淀结构化知识资产",
    subtitle: "围绕团队资料、规范与流程，建立可管理、可复用、可追踪的知识底座。",
    icon: BookOpen,
  },
  {
    title: "构建高质量知识库",
    subtitle: "将目录、文档与片段统一归档，让知识检索、编辑与协作保持同一套语义结构。",
    icon: Files,
  },
  {
    title: "连接 RAG 与模型能力",
    subtitle: "把检索、向量库与模型配置串成闭环，降低问答系统上线与维护成本。",
    icon: Boxes,
  },
  {
    title: "让系统助手持续工作",
    subtitle: "在队列监控、系统助手和运营配置之间建立稳定协同，形成长期演进的知识工作流。",
    icon: Sparkles,
  },
];

export function LoginPage() {
  const [current, setCurrent] = useState(0);
  const [fading, setFading] = useState(false);

  useEffect(() => {
    const timer = setInterval(() => {
      setFading(true);
      window.setTimeout(() => {
        setCurrent((prev) => (prev + 1) % slides.length);
        setFading(false);
      }, 350);
    }, 4000);

    return () => clearInterval(timer);
  }, []);

  const slide = slides[current];
  const Icon = slide.icon;

  return (
    <div className="flex min-h-svh w-full">
      <div className="hidden bg-gradient-to-br from-blue-600 via-indigo-600 to-violet-700 lg:flex lg:w-1/2 lg:flex-col lg:items-center lg:justify-center lg:p-12">
        <div
          className="flex w-full max-w-md flex-col items-center text-center"
          style={{
            opacity: fading ? 0 : 1,
            transform: fading ? "translateY(6px)" : "translateY(0)",
            transition: "opacity 0.35s ease, transform 0.35s ease",
          }}
        >
          <div className="relative mb-8 flex h-80 w-full items-center justify-center overflow-hidden rounded-[2rem] border border-white/12 bg-white/10 shadow-[0_24px_80px_rgb(0_0_0/0.22)] backdrop-blur">
            <div className="absolute inset-0 bg-[radial-gradient(circle_at_top,rgba(255,255,255,0.35),transparent_58%)]" />
            <div className="absolute -left-10 top-10 h-32 w-32 rounded-full bg-cyan-300/24 blur-3xl" />
            <div className="absolute -right-8 bottom-8 h-36 w-36 rounded-full bg-blue-300/18 blur-3xl" />
            <div className="relative grid w-[78%] gap-4">
              <div className="flex items-center justify-between rounded-2xl border border-white/18 bg-slate-950/18 px-4 py-3 text-white/88">
                <div className="flex items-center gap-3">
                  <div className="rounded-xl bg-white/14 p-2">
                    <Icon className="size-5" />
                  </div>
                  <div className="text-left">
                    <div className="text-sm font-medium">知识工作流</div>
                    <div className="text-xs text-white/60">检索、整理、问答、沉淀</div>
                  </div>
                </div>
                <div className="rounded-full bg-emerald-300/18 px-2.5 py-1 text-xs text-emerald-100">
                  在线
                </div>
              </div>
              <div className="rounded-2xl border border-white/18 bg-white/14 p-4 text-left text-white shadow-lg">
                <div className="mb-3 flex items-center justify-between">
                  <span className="text-sm font-medium">知识面板</span>
                  <span className="text-xs text-white/60">SAG Wiki</span>
                </div>
                <div className="space-y-3">
                  <div className="h-2 rounded-full bg-white/18">
                    <div className="h-2 w-[68%] rounded-full bg-cyan-300/80" />
                  </div>
                  <div className="grid grid-cols-2 gap-3">
                    <div className="rounded-xl bg-slate-950/18 p-3">
                      <div className="mb-2 text-xs text-white/60">知识库</div>
                      <div className="text-lg font-semibold">248</div>
                    </div>
                    <div className="rounded-xl bg-slate-950/18 p-3">
                      <div className="mb-2 text-xs text-white/60">会话</div>
                      <div className="text-lg font-semibold">1.4k</div>
                    </div>
                  </div>
                  <div className="rounded-xl bg-slate-950/18 p-3">
                    <div className="mb-2 text-xs text-white/60">当前焦点</div>
                    <div className="text-sm leading-6 text-white/84">{slide.title}</div>
                  </div>
                </div>
              </div>
            </div>
          </div>
          <h2 className="mb-2 text-xl font-semibold text-white">{slide.title}</h2>
          <p className="text-sm leading-relaxed text-white/70">{slide.subtitle}</p>
        </div>

        <div className="mt-10 flex gap-2">
          {slides.map((_, i) => (
            <button
              key={i}
              type="button"
              onClick={() => {
                setFading(true);
                window.setTimeout(() => {
                  setCurrent(i);
                  setFading(false);
                }, 350);
              }}
              className={`rounded-full transition-all duration-300 ${
                i === current ? "h-2 w-6 bg-white" : "h-2 w-2 bg-white/30 hover:bg-white/50"
              }`}
              aria-label={`切换到第 ${i + 1} 张介绍卡片`}
            />
          ))}
        </div>
      </div>

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
