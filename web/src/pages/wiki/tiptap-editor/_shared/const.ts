import type { Editor } from "@tiptap/react";

export interface RangeStyle {
  bold: boolean;
  italic: boolean;
  underline: boolean;
  strikethrough: boolean;
  code: boolean;
  highlight: string | null;
  textColor: string | null;
  fontFamily: string | null;
  fontSize: string | null;
  textAlign: "left" | "center" | "right" | "justify" | null;
  heading: number | null;
  bulletList: boolean;
  orderedList: boolean;
  blockquote: boolean;
}

export interface ToolbarContextValue {
  editor: Editor | null;
  rangeStyle: RangeStyle;
  docTitle: string;
}

export const DEFAULT_RANGE_STYLE: RangeStyle = {
  bold: false,
  italic: false,
  underline: false,
  strikethrough: false,
  code: false,
  highlight: null,
  textColor: null,
  fontFamily: null,
  fontSize: null,
  textAlign: null,
  heading: null,
  bulletList: false,
  orderedList: false,
  blockquote: false,
};

export const FONT_FAMILIES = [
  { label: "微软雅黑", value: "Microsoft YaHei" },
  { label: "华文宋体", value: "华文宋体" },
  { label: "华文黑体", value: "华文黑体" },
  { label: "华文仿宋", value: "华文仿宋" },
  { label: "华文楷体", value: "华文楷体" },
  { label: "华文琥珀", value: "华文琥珀" },
  { label: "华文隶书", value: "华文隶书" },
  { label: "华文新魏", value: "华文新魏" },
  { label: "华文行楷", value: "华文行楷" },
  { label: "华文中宋", value: "华文中宋" },
  { label: "华文彩云", value: "华文彩云" },
  { label: "Arial", value: "Arial" },
  { label: "Segoe UI", value: "Segoe UI" },
  { label: "Ink Free", value: "Ink Free" },
  { label: "Fantasy", value: "Fantasy" },
];

export const FONT_SIZES = [
  { label: "初号", value: "56" },
  { label: "小初", value: "48" },
  { label: "一号", value: "34" },
  { label: "小一", value: "32" },
  { label: "二号", value: "29" },
  { label: "小二", value: "24" },
  { label: "三号", value: "21" },
  { label: "小三", value: "20" },
  { label: "四号", value: "18" },
  { label: "小四", value: "16" },
  { label: "五号", value: "14" },
  { label: "小五", value: "12" },
  { label: "六号", value: "10" },
  { label: "小六", value: "8" },
  { label: "七号", value: "7" },
  { label: "八号", value: "6" },
];

export const TITLE_LEVELS = [
  { label: "正文", value: "16" },
  { label: "标题1", value: "26" },
  { label: "标题2", value: "24" },
  { label: "标题3", value: "22" },
  { label: "标题4", value: "20" },
  { label: "标题5", value: "18" },
];

export const PAPER_SIZES = [
  { label: "A4", value: "794*1123" },
  { label: "A2", value: "1593*2251" },
  { label: "A3", value: "1125*1593" },
  { label: "A5", value: "565*796" },
  { label: "5号信封", value: "412*488" },
  { label: "6号信封", value: "450*866" },
  { label: "7号信封", value: "609*862" },
  { label: "9号信封", value: "862*1221" },
  { label: "法律用纸", value: "813*1266" },
  { label: "信纸", value: "813*1054" },
];

export const ROW_MARGINS = [
  { label: "1", value: "1" },
  { label: "1.25", value: "1.25" },
  { label: "1.5", value: "1.5" },
  { label: "1.75", value: "1.75" },
  { label: "2", value: "2" },
  { label: "2.5", value: "2.5" },
  { label: "3", value: "3" },
];

export const TEXT_COLORS = ["#000000", "#e63946", "#457b9d", "#2a9d8f", "#f4a261"];
export const HIGHLIGHT_COLORS = ["#fff3cd", "#fde68a", "#bfdbfe", "#c7f9cc", "#fecdd3"];

const THEME_COLOR_COLUMNS = [
  {
    label: "白色",
    values: ["#FFFFFF", "#F2F2F2", "#D9D9D9", "#BFBFBF", "#A6A6A6", "#7F7F7F"],
  },
  {
    label: "黑色",
    values: ["#000000", "#7F7F7F", "#595959", "#3F3F3F", "#262626", "#0D0D0D"],
  },
  {
    label: "浅灰",
    values: ["#D9D9D9", "#ECECEC", "#D0CECE", "#B1AFAF", "#918D8D", "#666666"],
  },
  {
    label: "深灰蓝",
    values: ["#44546A", "#D6DCE4", "#ADB9CA", "#8497B0", "#5B6F8A", "#333F50"],
  },
  {
    label: "蓝色",
    values: ["#5B73C4", "#D9E2F3", "#B4C6E7", "#8EA9DB", "#6D84CF", "#3E5AA8"],
  },
  {
    label: "橙色",
    values: ["#F28C28", "#FCE4D6", "#F8CBAD", "#F4B183", "#F09B54", "#C55A11"],
  },
  {
    label: "黄色",
    values: ["#E6C229", "#FFF2CC", "#FFE699", "#FFD966", "#F1C232", "#BF9000"],
  },
  {
    label: "绿色",
    values: ["#8CC63E", "#E2F0D9", "#C6E0B4", "#A9D18E", "#70AD47", "#548235"],
  },
  {
    label: "青色",
    values: ["#55C0B8", "#D9EEEA", "#B7E1DD", "#95D5D0", "#6EC8C1", "#31817A"],
  },
  {
    label: "红色",
    values: ["#E5667A", "#F4D7DC", "#EBB1BC", "#E48E9D", "#D96D7F", "#C0504D"],
  },
] as const;

// 主题色 - 参照截图整理成 6 x 10 的 Office/WPS 风格色板
export const THEME_COLORS = THEME_COLOR_COLUMNS[0].values.flatMap((_, rowIndex) =>
  THEME_COLOR_COLUMNS.map((column) => ({
    label: `${column.label}-${rowIndex + 1}`,
    value: column.values[rowIndex],
  })),
);

// 标准色 - 底部一排高饱和度快捷色
export const STANDARD_COLORS = [
  { label: "深红", value: "#C00000" },
  { label: "红色", value: "#FF0000" },
  { label: "橙黄", value: "#FFC000" },
  { label: "亮黄", value: "#FFFF00" },
  { label: "浅绿", value: "#92D050" },
  { label: "绿色", value: "#00B050" },
  { label: "天蓝", value: "#00B0F0" },
  { label: "蓝色", value: "#5B9BD5" },
  { label: "深蓝", value: "#002060" },
  { label: "紫色", value: "#7030A0" },
];

export const ZOOM_MIN = 0.5;
export const ZOOM_MAX = 2.0;
export const ZOOM_STEP = 0.1;
export const ZOOM_DEFAULT = 1.0;

export const DEFAULT_MARGINS = {
  top: 16,
  bottom: 16,
  left: 32,
  right: 32,
};

export interface PaperSize {
  width: number;
  height: number;
}

export function parsePaperSize(value: string): PaperSize {
  const [w, h] = value.split("*").map(Number);
  return { width: w, height: h };
}

export const DEFAULT_PAPER_SIZE: PaperSize = parsePaperSize(PAPER_SIZES[0].value);
