import type { GuidePracticePoint, GuideResult, LearningObjective } from "@/api/learning";

export type WorkbenchPhase = "reading" | "answering";

export type MarkdownCatalogItem = {
  line: number;
  level: number;
  text: string;
};

export type GuideContent = {
  summary: string;
  focusItems: string[];
  practicePoints: GuidePracticePoint[];
  readingSteps: string[];
  pitfalls: string[];
  selfCheckQuestions: string[];
  evidenceItems: string[];
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

export function guideResultToContent(result: GuideResult): GuideContent {
  return {
    summary: result.summary,
    focusItems: toArray(result.mastery_goals),
    practicePoints: toArray(result.practice_points),
    readingSteps: toArray(result.reading_steps),
    pitfalls: toArray(result.pitfalls),
    selfCheckQuestions: toArray(result.self_check_questions),
    evidenceItems: toArray(result.evidence),
  };
}

export function toArray<T>(value: T[] | null | undefined): T[] {
  return Array.isArray(value) ? value : [];
}

export function scrollToMarkdownHeading(text: string) {
  const headings = Array.from(
    document.querySelectorAll<HTMLElement>("article h1, article h2, article h3, article h4"),
  );
  const heading = headings.find((item) => item.textContent?.trim() === text);
  heading?.scrollIntoView({ block: "start", behavior: "smooth" });
}

export function buildPrompt(objective: LearningObjective, practicePoint: GuidePracticePoint | null) {
  if (!practicePoint) {
    return `请用自己的话解释：${objective.title}`;
  }

  return [
    `请用自己的话解释：${objective.title}`,
    `本轮只判断这个复述小点：${practicePoint.title}`,
    practicePoint.goal ? `本轮目标：${practicePoint.goal}` : "",
    "请不要按整篇资料要求判定，只看这个小点是否讲清楚。",
  ]
    .filter(Boolean)
    .join("\n");
}
