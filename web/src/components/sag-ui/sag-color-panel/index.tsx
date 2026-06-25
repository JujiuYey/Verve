import "./index.css";
import { useEffect, useState } from "react";
import { HexColorPicker } from "react-colorful";

import { Input } from "@/components/ui/input";
import { cn } from "@/lib/utils";
import {
  hexToRgbDraft,
  isLightColor,
  normalizeHexColor,
  type RgbDraft,
  rgbDraftToHex,
  sanitizeHexInput,
} from "@/utils/color";

interface SagColorPanelProps {
  color: string;
  onChange: (color: string) => void;
}

export function SagColorPanel({ color, onChange }: SagColorPanelProps) {
  const normalizedColor = normalizeHexColor(color) ?? "#000000";
  const [hexValue, setHexValue] = useState(normalizedColor);
  const [rgbValue, setRgbValue] = useState<RgbDraft>(
    () => hexToRgbDraft(normalizedColor) ?? { r: "0", g: "0", b: "0" },
  );

  /* eslint-disable react-hooks/set-state-in-effect */
  useEffect(() => {
    setHexValue(normalizedColor);
    setRgbValue(hexToRgbDraft(normalizedColor) ?? { r: "0", g: "0", b: "0" });
  }, [normalizedColor]);
  /* eslint-enable react-hooks/set-state-in-effect */

  const syncInputs = (nextColor: string) => {
    const normalizedNextColor = normalizeHexColor(nextColor);
    if (!normalizedNextColor) {
      return;
    }

    setHexValue(normalizedNextColor);
    setRgbValue(hexToRgbDraft(normalizedNextColor) ?? { r: "0", g: "0", b: "0" });
  };

  const handleColorChange = (nextColor: string) => {
    const normalizedNextColor = normalizeHexColor(nextColor);
    if (!normalizedNextColor) {
      return;
    }

    syncInputs(normalizedNextColor);
    onChange(normalizedNextColor);
  };

  return (
    <div>
      <div className="space-y-3 rounded-md">
        <div className="flex items-center gap-3">
          <div
            className={cn(
              "h-10 w-10 rounded-full border shadow-inner",
              isLightColor(normalizedColor) ? "border-zinc-300" : "border-transparent",
            )}
            style={{ backgroundColor: normalizedColor }}
          />
          <div className="min-w-0">
            <div className="text-xs text-muted-foreground">当前颜色</div>
            <div className="font-mono text-sm font-medium uppercase">{normalizedColor}</div>
          </div>
        </div>

        <div className="space-y-2">
          <div className="space-y-1">
            <div className="text-xs text-muted-foreground">色盘</div>
            <HexColorPicker
              color={normalizedColor}
              onChange={handleColorChange}
              className={cn(
                "sag-color-panel-picker",
                "!h-[180px] !w-full",
                "[&_.react-colorful__saturation]:rounded-sm",
                "[&_.react-colorful__hue]:mt-3 [&_.react-colorful__hue]:h-2 [&_.react-colorful__hue]:rounded-full",
              )}
            />
          </div>

          <div className="space-y-1">
            <div className="text-xs text-muted-foreground">HEX</div>
            <Input
              value={hexValue}
              placeholder="#RRGGBB"
              className="h-8 font-mono uppercase"
              onChange={(event) => {
                const nextHexValue = sanitizeHexInput(event.target.value);
                setHexValue(nextHexValue);

                const normalizedNextColor = normalizeHexColor(nextHexValue);
                if (!normalizedNextColor) {
                  return;
                }

                syncInputs(normalizedNextColor);
                onChange(normalizedNextColor);
              }}
              onBlur={() => {
                if (!normalizeHexColor(hexValue)) {
                  syncInputs(normalizedColor);
                }
              }}
            />
          </div>

          <div className="space-y-1">
            <div className="text-xs text-muted-foreground">RGB</div>
            <div className="grid grid-cols-3 gap-2">
              {(["r", "g", "b"] as const).map((channel) => (
                <div key={channel} className="space-y-1">
                  <div className="text-[11px] font-medium uppercase text-muted-foreground">
                    {channel}
                  </div>
                  <Input
                    type="text"
                    inputMode="numeric"
                    value={rgbValue[channel]}
                    className="h-8 px-2 text-center font-mono"
                    onChange={(event) => {
                      const nextChannelValue = event.target.value.replace(/\D/g, "").slice(0, 3);
                      const nextRgbValue = { ...rgbValue, [channel]: nextChannelValue };
                      setRgbValue(nextRgbValue);

                      const nextHexColor = rgbDraftToHex(nextRgbValue);
                      if (!nextHexColor) {
                        return;
                      }

                      syncInputs(nextHexColor);
                      onChange(nextHexColor);
                    }}
                    onBlur={() => {
                      if (!rgbDraftToHex(rgbValue)) {
                        syncInputs(normalizedColor);
                      }
                    }}
                  />
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
