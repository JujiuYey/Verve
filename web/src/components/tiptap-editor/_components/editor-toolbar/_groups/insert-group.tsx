import { CheckSquare, Code2, ImagePlus, Link, Table2 } from "lucide-react";
import { type ChangeEvent, useRef } from "react";

import { useToolbarContext } from "../_context/toolbar-context";
import { ToolbarButton } from "../_primitives/toolbar-button";

export function InsertGroup() {
  const { editor } = useToolbarContext();
  const fileInputRef = useRef<HTMLInputElement>(null);

  const insertLink = () => {
    const href = window.prompt("输入链接地址");
    if (href) {
      editor?.chain().focus().setLink({ href }).run();
    }
  };

  const handleImageChange = (event: ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (!file) {
      return;
    }

    const reader = new FileReader();
    reader.onload = () => {
      const src = typeof reader.result === "string" ? reader.result : "";
      if (src) {
        editor?.chain().focus().setImage({ src }).run();
      }
    };
    reader.readAsDataURL(file);
    event.target.value = "";
  };

  return (
    <>
      <ToolbarButton
        icon={<Link className="h-4 w-4" />}
        title="插入链接"
        disabled={!editor}
        onClick={insertLink}
      />
      <ToolbarButton
        icon={<Table2 className="h-4 w-4" />}
        title="插入表格"
        disabled={!editor}
        onClick={() =>
          editor
            ?.chain()
            .focus()
            .insertTable({
              rows: 3,
              cols: 3,
              withHeaderRow: true,
            })
            .run()
        }
      />
      <ToolbarButton
        icon={<ImagePlus className="h-4 w-4" />}
        title="插入图片"
        disabled={!editor}
        onClick={() => fileInputRef.current?.click()}
      />
      <ToolbarButton
        icon={<Code2 className="h-4 w-4" />}
        title="代码块"
        disabled={!editor}
        onClick={() => editor?.chain().focus().toggleCodeBlock().run()}
      />
      <ToolbarButton
        icon={<CheckSquare className="h-4 w-4" />}
        title="任务列表"
        disabled={!editor}
        onClick={() => editor?.chain().focus().toggleTaskList().run()}
      />
      <input
        ref={fileInputRef}
        type="file"
        accept="image/*"
        className="hidden"
        onChange={handleImageChange}
      />
    </>
  );
}
