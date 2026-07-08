import audioIcon from "@/assets/icon/file_icon_audio.svg";
import docxIcon from "@/assets/icon/file_icon_docx.svg";
import exeIcon from "@/assets/icon/file_icon_exe.svg";
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

interface IconAsset {
  alt: string;
  src: string;
}

interface DocumentIconInput {
  contentType: string;
  filename: string;
}

const imageExtensions = new Set(["png", "jpg", "jpeg", "gif", "webp", "svg", "bmp"]);
const audioExtensions = new Set(["mp3", "wav", "m4a", "aac", "flac", "ogg"]);
const videoExtensions = new Set(["mp4", "mov", "avi", "mkv", "webm"]);
const markdownExtensions = new Set(["md", "markdown"]);
const scriptExtensions = new Set([
  "js",
  "jsx",
  "ts",
  "tsx",
  "py",
  "sh",
  "sql",
  "json",
  "yaml",
  "yml",
  "toml",
]);
const textExtensions = new Set(["txt"]);
const archiveExtensions = new Set(["zip", "rar", "7z", "tar", "gz"]);

function getFileExtension(filename: string) {
  const segments = filename.toLowerCase().split(".");
  return segments.length > 1 ? (segments.at(-1) ?? "") : "";
}

function matchesExecutable(contentType: string, extension: string) {
  return contentType.includes("exe") || contentType.includes("executable") || extension === "exe";
}

export const folderIconAsset: IconAsset = {
  alt: "文件夹图标",
  src: folderIcon,
};

export function getDocumentIconAsset({ contentType, filename }: DocumentIconInput): IconAsset {
  const normalizedType = contentType.toLowerCase();
  const extension = getFileExtension(filename);

  if (normalizedType.includes("pdf") || extension === "pdf") {
    return { alt: "PDF 文件图标", src: pdfIcon };
  }

  if (normalizedType.includes("image") || imageExtensions.has(extension)) {
    return { alt: "图片文件图标", src: imageIcon };
  }

  if (normalizedType.includes("audio") || audioExtensions.has(extension)) {
    return { alt: "音频文件图标", src: audioIcon };
  }

  if (normalizedType.includes("video") || videoExtensions.has(extension)) {
    return { alt: "视频文件图标", src: videoIcon };
  }

  if (normalizedType.includes("wordprocessingml") || extension === "docx") {
    return { alt: "Word 文件图标", src: docxIcon };
  }

  if (normalizedType.includes("spreadsheetml") || extension === "xlsx") {
    return { alt: "Excel 文件图标", src: xlsxIcon };
  }

  if (normalizedType.includes("presentationml") || extension === "pptx") {
    return { alt: "PPT 文件图标", src: pptxIcon };
  }

  if (normalizedType.includes("zip") || archiveExtensions.has(extension)) {
    return { alt: "压缩文件图标", src: zipIcon };
  }

  if (matchesExecutable(normalizedType, extension)) {
    return { alt: "可执行文件图标", src: exeIcon };
  }

  if (normalizedType.includes("markdown") || markdownExtensions.has(extension)) {
    return { alt: "Markdown 文件图标", src: mdIcon };
  }

  if (scriptExtensions.has(extension)) {
    return { alt: "脚本文件图标", src: scriptIcon };
  }

  if (normalizedType.includes("text/plain") || textExtensions.has(extension)) {
    return { alt: "文本文件图标", src: txtIcon };
  }

  return { alt: "文件图标", src: otherIcon };
}
