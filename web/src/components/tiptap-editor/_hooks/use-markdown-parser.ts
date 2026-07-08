export function useMarkdownParser() {
  return {
    normalize: (markdown: string) => markdown.trim(),
  };
}
