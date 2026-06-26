// Client-side background preference (mirrors the dark/light theme: stored in
// localStorage, applied in the browser, no server round-trip). The custom image
// is downscaled before storing so it fits comfortably in localStorage.

export type BgKind = "none" | "starmind" | "apollo" | "custom";

export const BG_KEY = "slashnode-bg";
export const BG_CUSTOM_KEY = "slashnode-bg-custom";
export const BG_EVENT = "slashnode-bg-change";

// Bundled predefined photo (Apollo 11, AS11-40-5927).
export const APOLLO_SRC = "/backgrounds/as11-40-5927.jpg";

export function getBg(): BgKind {
  if (typeof window === "undefined") return "none";
  const v = localStorage.getItem(BG_KEY) as BgKind | null;
  return v ?? "none";
}

export function setBg(kind: BgKind) {
  localStorage.setItem(BG_KEY, kind);
  window.dispatchEvent(new CustomEvent(BG_EVENT));
}

export function getCustom(): string | null {
  if (typeof window === "undefined") return null;
  return localStorage.getItem(BG_CUSTOM_KEY);
}

export function setCustom(dataUrl: string) {
  localStorage.setItem(BG_CUSTOM_KEY, dataUrl);
  window.dispatchEvent(new CustomEvent(BG_EVENT));
}

// readAndDownscale loads a user-selected image file and returns a downscaled
// JPEG data URL (max 1920px on the long edge) to keep localStorage small.
export function readAndDownscale(file: File, maxEdge = 1920): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onerror = () => reject(new Error("read failed"));
    reader.onload = () => {
      const img = new Image();
      img.onerror = () => reject(new Error("decode failed"));
      img.onload = () => {
        const scale = Math.min(1, maxEdge / Math.max(img.width, img.height));
        const w = Math.round(img.width * scale);
        const h = Math.round(img.height * scale);
        const canvas = document.createElement("canvas");
        canvas.width = w;
        canvas.height = h;
        const ctx = canvas.getContext("2d");
        if (!ctx) return reject(new Error("no canvas"));
        ctx.drawImage(img, 0, 0, w, h);
        resolve(canvas.toDataURL("image/jpeg", 0.82));
      };
      img.src = reader.result as string;
    };
    reader.readAsDataURL(file);
  });
}
