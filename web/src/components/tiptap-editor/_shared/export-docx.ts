import type { Editor } from "@tiptap/react";
import { Packer, ShadingType } from "docx";
import { saveAs } from "file-saver";
import { defaultMarks, defaultNodes, DocxSerializer } from "prosemirror-docx";

/* eslint-disable @typescript-eslint/no-explicit-any */

// prosemirror-docx uses snake_case node names, TipTap uses camelCase.
const tiptapNodes: Record<string, any> = {
  text: defaultNodes.text,
  paragraph: defaultNodes.paragraph,
  heading: defaultNodes.heading,
  blockquote: defaultNodes.blockquote,
  codeBlock: defaultNodes.code_block,
  horizontalRule: defaultNodes.horizontal_rule,
  hardBreak: defaultNodes.hard_break,
  orderedList: defaultNodes.ordered_list,
  bulletList: defaultNodes.bullet_list,
  listItem: defaultNodes.list_item,
  image: defaultNodes.image,
  table: defaultNodes.table,
  tableRow(state: any, node: any) {
    state.renderContent(node);
  },
  tableHeader(state: any, node: any) {
    state.renderContent(node);
  },
  tableCell(state: any, node: any) {
    state.renderContent(node);
  },
  taskList(state: any, node: any) {
    state.renderList(node, "bullets");
  },
  taskItem(state: any, node: any) {
    const checked = node.attrs.checked as boolean;
    state.text(checked ? "☑ " : "☐ ");
    state.renderListItem(node);
  },
};

const tiptapMarks: Record<string, any> = {
  ...defaultMarks,
  highlight(_state: any, _node: any, mark: any) {
    const color = (mark.attrs.color as string | null)?.replace("#", "") ?? "FFFF00";
    return {
      shading: { type: ShadingType.SOLID, color, fill: color },
    };
  },
  textStyle(_state: any, _node: any, mark: any) {
    const opts: Record<string, unknown> = {};
    if (mark.attrs.color) {
      opts.color = (mark.attrs.color as string).replace("#", "");
    }
    if (mark.attrs.fontFamily) {
      opts.font = { name: mark.attrs.fontFamily as string };
    }
    if (mark.attrs.fontSize) {
      const px = Number.parseInt(mark.attrs.fontSize as string, 10);
      if (!Number.isNaN(px)) {
        opts.size = px * 2;
      }
    }
    return opts;
  },
};

/* eslint-enable @typescript-eslint/no-explicit-any */

const serializer = new DocxSerializer(tiptapNodes, tiptapMarks);

export async function exportToDocx(editor: Editor, title: string): Promise<void> {
  const doc = serializer.serialize(editor.state.doc, {
    getImageBuffer: () => new Uint8Array(0),
  });
  console.log("🚀 ~ exportToDocx ~ doc:", doc);
  const blob = await Packer.toBlob(doc);
  saveAs(blob, `${title}.docx`);
}
