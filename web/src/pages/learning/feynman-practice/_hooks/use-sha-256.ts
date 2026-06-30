import { useEffect, useState } from "react";

export function useSHA256(text: string) {
  const [hash, setHash] = useState("");

  useEffect(() => {
    let cancelled = false;
    const value = text.trim();
    if (!value) {
      setHash("");
      return;
    }

    crypto.subtle
      .digest("SHA-256", new TextEncoder().encode(value))
      .then((buffer) => {
        if (cancelled) return;
        const bytes = Array.from(new Uint8Array(buffer));
        setHash(bytes.map((byte) => byte.toString(16).padStart(2, "0")).join(""));
      })
      .catch(() => {
        if (!cancelled) setHash("");
      });

    return () => {
      cancelled = true;
    };
  }, [text]);

  return hash;
}
