import { Check, Loader2, Ruler, Save, ZoomIn, ZoomOut } from "lucide-react";
import { useState } from "react";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { cn } from "@/lib/utils";

import type { EditorPageSetup } from "../_hooks/use-editor-page-setup";

interface EditorFooterProps {
  saveStatus: "saved" | "saving" | "unsaved";
  wordCount: number;
  onSave: () => void | Promise<void>;
  pageSetup: EditorPageSetup;
}

export function EditorFooter({ saveStatus, wordCount, onSave, pageSetup }: EditorFooterProps) {
  const [marginOpen, setMarginOpen] = useState(false);
  const zoomLabel = `${Math.round(pageSetup.zoom * 100)}%`;

  return (
    <div className="flex items-center justify-between border-t bg-background px-4 py-2 text-sm">
      <span className="text-muted-foreground">
        字数:
        {wordCount}
      </span>

      <div className="flex items-center gap-2">
        <Button
          type="button"
          variant="ghost"
          size="icon"
          className="h-7 w-7"
          title="缩小"
          onClick={pageSetup.zoomOut}
        >
          <ZoomOut className="h-4 w-4" />
        </Button>

        <Button
          type="button"
          variant="ghost"
          size="sm"
          className="h-7 min-w-[3rem] text-xs"
          title="重置缩放"
          onClick={pageSetup.zoomReset}
        >
          {zoomLabel}
        </Button>

        <Button
          type="button"
          variant="ghost"
          size="icon"
          className="h-7 w-7"
          title="放大"
          onClick={pageSetup.zoomIn}
        >
          <ZoomIn className="h-4 w-4" />
        </Button>
        <Popover open={marginOpen} onOpenChange={setMarginOpen}>
          <PopoverTrigger asChild>
            <Button type="button" variant="ghost" size="sm" className="h-7 text-xs" title="页边距">
              <Ruler className="mr-1 h-3 w-3" />
              边距
            </Button>
          </PopoverTrigger>
          <PopoverContent className="w-64" align="end">
            <div className="grid grid-cols-2 gap-3">
              <MarginInput
                label="上"
                value={pageSetup.margins.top}
                onChange={(v) => pageSetup.setMargins({ top: v })}
              />
              <MarginInput
                label="下"
                value={pageSetup.margins.bottom}
                onChange={(v) => pageSetup.setMargins({ bottom: v })}
              />
              <MarginInput
                label="左"
                value={pageSetup.margins.left}
                onChange={(v) => pageSetup.setMargins({ left: v })}
              />
              <MarginInput
                label="右"
                value={pageSetup.margins.right}
                onChange={(v) => pageSetup.setMargins({ right: v })}
              />
            </div>
            <Button
              type="button"
              variant="outline"
              size="sm"
              className="mt-3 w-full text-xs"
              onClick={pageSetup.resetMargins}
            >
              重置边距
            </Button>
          </PopoverContent>
        </Popover>

        <span
          className={cn(
            "flex items-center gap-1 text-xs",
            saveStatus === "saved" && "text-green-600",
            saveStatus === "saving" && "text-amber-600",
            saveStatus === "unsaved" && "text-orange-600",
          )}
        >
          {saveStatus === "saved" && (
            <>
              <Check className="h-3 w-3" />
              已保存
            </>
          )}
          {saveStatus === "saving" && (
            <>
              <Loader2 className="h-3 w-3 animate-spin" />
              保存中...
            </>
          )}
          {saveStatus === "unsaved" && "未保存"}
        </span>

        <Button type="button" size="sm" variant="outline" onClick={() => void onSave()}>
          <Save className="h-3 w-3" />
          保存
        </Button>
      </div>
    </div>
  );
}

function MarginInput({
  label,
  value,
  onChange,
}: {
  label: string;
  value: number;
  onChange: (value: number) => void;
}) {
  return (
    <div className="space-y-1">
      <label className="text-xs text-muted-foreground">{label} (px)</label>
      <Input
        type="number"
        min={0}
        className="h-7 text-xs"
        value={value}
        onChange={(e) => onChange(Number(e.target.value) || 0)}
      />
    </div>
  );
}
