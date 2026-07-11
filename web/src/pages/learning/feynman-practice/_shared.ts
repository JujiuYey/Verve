export type MarkdownCatalogItem = {
  line: number;
  level: number;
  text: string;
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
