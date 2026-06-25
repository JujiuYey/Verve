import { FileDown, Printer, Search } from "lucide-react";
import { useMemo, useState } from "react";

import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";

import { exportToDocx } from "../../../_shared/export-docx";
import { useToolbarContext } from "../_context/toolbar-context";
import { ToolbarButton } from "../_primitives/toolbar-button";

function getSearchMatches(text: string, query: string) {
  if (!query) {
    return [];
  }

  const matches: { from: number; to: number }[] = [];
  let index = text.indexOf(query);

  while (index >= 0) {
    matches.push({
      from: index + 1,
      to: index + query.length + 1,
    });
    index = text.indexOf(query, index + query.length);
  }

  return matches;
}

export function SearchPrintGroup() {
  const { editor, docTitle } = useToolbarContext();
  const [open, setOpen] = useState(false);
  const [query, setQuery] = useState("");
  const [replaceValue, setReplaceValue] = useState("");
  const [activeIndex, setActiveIndex] = useState(-1);

  const matches = useMemo(
    () => getSearchMatches(editor?.state?.doc?.textContent || "", query),
    [editor, query],
  );

  const jumpToMatch = (index: number) => {
    const match = matches[index];
    if (!match) {
      return;
    }

    editor?.chain().focus().setTextSelection({ from: match.from, to: match.to }).run();
  };

  const handleNext = () => {
    if (matches.length === 0) {
      return;
    }

    const nextIndex = activeIndex >= matches.length - 1 ? 0 : activeIndex + 1;
    setActiveIndex(nextIndex);
    jumpToMatch(nextIndex);
  };

  const handlePrevious = () => {
    if (matches.length === 0) {
      return;
    }

    const previousIndex = activeIndex <= 0 ? matches.length - 1 : activeIndex - 1;
    setActiveIndex(previousIndex);
    jumpToMatch(previousIndex);
  };

  const handleReplaceCurrent = () => {
    if (!replaceValue || activeIndex < 0) {
      return;
    }

    const match = matches[activeIndex];
    if (!match) {
      return;
    }

    editor
      ?.chain()
      .focus()
      .setTextSelection({ from: match.from, to: match.to })
      .insertContent(replaceValue)
      .run();
  };

  return (
    <>
      <ToolbarButton
        icon={<Search className="h-4 w-4" />}
        title="搜索"
        disabled={!editor}
        onClick={() => setOpen(true)}
      />
      <ToolbarButton
        icon={<Printer className="h-4 w-4" />}
        title="打印"
        onClick={() => window.print()}
      />
      <ToolbarButton
        icon={<FileDown className="h-4 w-4" />}
        title="导出 Word"
        disabled={!editor}
        onClick={() => editor && exportToDocx(editor, docTitle)}
      />

      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>搜索</DialogTitle>
            <DialogDescription>在当前文档中搜索文本并跳转到下一个匹配项。</DialogDescription>
          </DialogHeader>

          <div className="space-y-3">
            <Input
              aria-label="搜索内容"
              value={query}
              onChange={(event) => {
                setQuery(event.target.value);
                setActiveIndex(-1);
              }}
              placeholder="输入搜索内容..."
            />
            <Input
              aria-label="替换内容"
              value={replaceValue}
              onChange={(event) => setReplaceValue(event.target.value)}
              placeholder="输入替换内容..."
            />
            <div className="flex items-center gap-2">
              <Button type="button" variant="outline" onClick={handlePrevious}>
                上一个
              </Button>
              <Button type="button" onClick={handleNext}>
                下一个
              </Button>
              <Button type="button" variant="outline" onClick={handleReplaceCurrent}>
                替换当前
              </Button>
            </div>
          </div>
        </DialogContent>
      </Dialog>
    </>
  );
}
