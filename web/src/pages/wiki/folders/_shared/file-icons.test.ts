import { readFileSync } from "node:fs";
import { join } from "node:path";

import { describe, expect, it } from "vitest";

import audioIcon from "@/assets/icon/file_icon_audio.svg";
import docxIcon from "@/assets/icon/file_icon_docx.svg";
import folderIcon from "@/assets/icon/file_icon_folder.svg";
import imageIcon from "@/assets/icon/file_icon_image.svg";
import mdIcon from "@/assets/icon/file_icon_md.svg";
import otherIcon from "@/assets/icon/file_icon_other.svg";
import pdfIcon from "@/assets/icon/file_icon_pdf.svg";
import pptxIcon from "@/assets/icon/file_icon_pptx.svg";
import scriptIcon from "@/assets/icon/file_icon_script.svg";
import txtIcon from "@/assets/icon/file_icon_txt.svg";
import videoIcon from "@/assets/icon/file_icon_video.svg";
import xlsxIcon from "@/assets/icon/file_icon_xlsx.svg";
import zipIcon from "@/assets/icon/file_icon_zip.svg";

import { folderIconAsset, getDocumentIconAsset } from "./file-icons";

describe("file icon assets", () => {
  it("renders markdown as a clean centered M monogram without a light badge", () => {
    const svg = readFileSync(join(process.cwd(), "src/assets/icon/file_icon_md.svg"), "utf8");

    expect(svg).not.toContain("<rect");
    expect(svg).toContain('fill="#8B5CF6"');
    expect(svg).toContain('fill="#6D28D9"');
    expect(svg).toContain('<g transform="translate(63.85 -29.525) scale(1.05)">');
    expect(svg).toContain(
      '<path d="M256 742V439h67l96 128 96-128h67v303h-61V532l-82 109h-41L317 532v210z" fill="#FFFFFF"/>',
    );
    expect(svg).not.toContain('fill-rule="evenodd"');
  });

  it("maps common wiki document types to the matching svg asset", () => {
    expect(getDocumentIconAsset({ contentType: "application/pdf", filename: "spec.pdf" }).src).toBe(
      pdfIcon,
    );
    expect(getDocumentIconAsset({ contentType: "image/png", filename: "cover.png" }).src).toBe(
      imageIcon,
    );
    expect(getDocumentIconAsset({ contentType: "text/markdown", filename: "guide.md" }).src).toBe(
      mdIcon,
    );
    expect(getDocumentIconAsset({ contentType: "text/plain", filename: "readme.txt" }).src).toBe(
      txtIcon,
    );
    expect(getDocumentIconAsset({ contentType: "audio/mpeg", filename: "voice.mp3" }).src).toBe(
      audioIcon,
    );
    expect(getDocumentIconAsset({ contentType: "video/mp4", filename: "demo.mp4" }).src).toBe(
      videoIcon,
    );
    expect(
      getDocumentIconAsset({
        contentType: "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
        filename: "proposal.docx",
      }).src,
    ).toBe(docxIcon);
    expect(
      getDocumentIconAsset({
        contentType: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
        filename: "budget.xlsx",
      }).src,
    ).toBe(xlsxIcon);
    expect(
      getDocumentIconAsset({
        contentType: "application/vnd.openxmlformats-officedocument.presentationml.presentation",
        filename: "review.pptx",
      }).src,
    ).toBe(pptxIcon);
    expect(
      getDocumentIconAsset({ contentType: "application/zip", filename: "archive.zip" }).src,
    ).toBe(zipIcon);
  });

  it("falls back by filename extension and then to the generic asset", () => {
    expect(
      getDocumentIconAsset({ contentType: "application/octet-stream", filename: "notes.markdown" })
        .src,
    ).toBe(mdIcon);
    expect(
      getDocumentIconAsset({ contentType: "application/octet-stream", filename: "script.py" }).src,
    ).toBe(scriptIcon);
    expect(
      getDocumentIconAsset({ contentType: "application/octet-stream", filename: "unknown.bin" })
        .src,
    ).toBe(otherIcon);
  });

  it("exposes the folder asset for folder surfaces", () => {
    expect(folderIconAsset.src).toBe(folderIcon);
    expect(folderIconAsset.alt).toBe("文件夹图标");
  });
});
