export interface RgbDraft {
  r: string;
  g: string;
  b: string;
}

function rgbChannelsToHex(channels: number[]) {
  return `#${channels
    .map((channel) => channel.toString(16).padStart(2, "0").toUpperCase())
    .join("")}`;
}

export function normalizeHexColor(value: string) {
  const normalizedValue = value.trim();
  const hex = normalizedValue.replace(/^#/, "");
  if (!/^[0-9A-Fa-f]{6}$/.test(hex)) {
    const rgbMatch = normalizedValue.match(/^rgba?\((.*)\)$/i);
    if (!rgbMatch) {
      return null;
    }

    const channels = rgbMatch[1]
      .split(/[\s,/]+/)
      .filter(Boolean)
      .slice(0, 3);

    if (channels.length !== 3 || !channels.every((channel) => /^\d{1,3}$/.test(channel))) {
      return null;
    }

    const numericChannels = channels.map(Number);
    if (!numericChannels.every((channel) => channel >= 0 && channel <= 255)) {
      return null;
    }

    return rgbChannelsToHex(numericChannels);
  }
  return `#${hex.toUpperCase()}`;
}

export function sanitizeHexInput(value: string) {
  const hex = value
    .replace(/[^0-9A-Fa-f]/g, "")
    .slice(0, 6)
    .toUpperCase();
  return hex ? `#${hex}` : "";
}

function isValidRgbChannel(value: string) {
  if (!/^\d{1,3}$/.test(value)) {
    return false;
  }

  const numericValue = Number(value);
  return numericValue >= 0 && numericValue <= 255;
}

export function hexToRgbDraft(value: string): RgbDraft | null {
  const normalizedColor = normalizeHexColor(value);
  if (!normalizedColor) {
    return null;
  }

  return {
    r: String(Number.parseInt(normalizedColor.slice(1, 3), 16)),
    g: String(Number.parseInt(normalizedColor.slice(3, 5), 16)),
    b: String(Number.parseInt(normalizedColor.slice(5, 7), 16)),
  };
}

export function rgbDraftToHex(value: RgbDraft) {
  const channels = [value.r, value.g, value.b];
  if (!channels.every(isValidRgbChannel)) {
    return null;
  }

  return rgbChannelsToHex(channels.map((channel) => Number(channel)));
}

export function isLightColor(value: string) {
  const rgb = hexToRgbDraft(value);
  if (!rgb) {
    return false;
  }

  return [rgb.r, rgb.g, rgb.b].every((channel) => Number(channel) >= 235);
}
