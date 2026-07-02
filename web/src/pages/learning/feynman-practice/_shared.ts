import type { LearningObjective } from "@/api/learning";

export type WorkbenchPhase = "reading" | "answering";

export type MarkdownCatalogItem = {
  line: number;
  level: number;
  text: string;
};

export const masteryLabels: Record<string, string> = {
  none: "未验证",
  seen: "看过",
  heard: "听过",
  explained: "能解释",
  written: "能写出",
  verified: "已验证",
};

export const verdictLabels: Record<string, string> = {
  pass: "通过",
  partial: "部分掌握",
  fail: "未通过",
};

export function extractMarkdownCatalog(markdown: string): MarkdownCatalogItem[] {
  return markdown
    .split("\n")
    .map((line, index) => {
      const match = /^(#{1,4})\s+(.+)$/.exec(line.trim());
      if (!match) return null;
      return {
        line: index,
        level: match[1].length,
        text: match[2].replace(/[#*_`]/g, "").trim(),
      };
    })
    .filter((item): item is MarkdownCatalogItem => !!item && item.text.length > 0);
}

export function scrollToMarkdownHeading(text: string) {
  const headings = Array.from(
    document.querySelectorAll<HTMLElement>("article h1, article h2, article h3, article h4"),
  );
  const heading = headings.find((item) => item.textContent?.trim() === text);
  heading?.scrollIntoView({ block: "start", behavior: "smooth" });
}

export function buildPrompt(objective: LearningObjective) {
  return [
    `请用自己的话解释：${objective.title}`,
    objective.detail ? `本轮目标：${objective.detail}` : "",
    "请只围绕当前学习小节判定，不要要求学习者一次讲完整篇资料。",
  ]
    .filter(Boolean)
    .join("\n");
}
