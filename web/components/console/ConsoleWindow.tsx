"use client";

import { useEffect, useRef, useState } from "react";
import { Terminal } from "@xterm/xterm";
import { FitAddon } from "@xterm/addon-fit";
import { SearchAddon } from "@xterm/addon-search";
import "@xterm/xterm/css/xterm.css";
import { useFloating } from "@/components/windows/useFloating";

// ConsoleWindow is a floating, draggable, resizable terminal connected to a
// container via the slashnoded console WebSocket. Multiple can be open at once.
export function ConsoleWindow({
  container,
  index,
  onClose,
}: {
  container: string;
  index: number;
  onClose: () => void;
}) {
  const termHost = useRef<HTMLDivElement>(null);
  const fitRef = useRef<FitAddon | null>(null);
  const searchRef = useRef<SearchAddon | null>(null);
  const wsRef = useRef<WebSocket | null>(null);
  const [query, setQuery] = useState("");
  const { pos, size, z, bringFront, startDrag, startResize } = useFloating(index, {
    w: 640,
    h: 380,
  });

  // Terminal + WebSocket lifecycle.
  useEffect(() => {
    if (!termHost.current) return;
    const term = new Terminal({
      fontSize: 13,
      fontFamily: "ui-monospace, SFMono-Regular, Menlo, monospace",
      theme: { background: "#0e0e10", foreground: "#f5f5f5" },
    });
    const fit = new FitAddon();
    fitRef.current = fit;
    term.loadAddon(fit);
    const search = new SearchAddon();
    searchRef.current = search;
    term.loadAddon(search);
    term.open(termHost.current);
    fit.fit();

    const proto = location.protocol === "https:" ? "wss:" : "ws:";
    const ws = new WebSocket(
      `${proto}//${location.host}/__console?c=${encodeURIComponent(container)}`,
    );
    ws.binaryType = "arraybuffer";
    wsRef.current = ws;
    const enc = new TextEncoder();

    const sendResize = () => {
      if (ws.readyState === WebSocket.OPEN) {
        ws.send(JSON.stringify({ resize: { cols: term.cols, rows: term.rows } }));
      }
    };
    ws.onopen = () => {
      term.writeln(`\x1b[2mconnected to ${container}\x1b[0m`);
      sendResize();
    };
    ws.onmessage = (e) => term.write(new Uint8Array(e.data as ArrayBuffer));
    ws.onclose = () => term.writeln("\r\n\x1b[2m[disconnected]\x1b[0m");
    ws.onerror = () =>
      term.writeln("\r\n\x1b[31mconnection failed\x1b[0m");

    const disp = term.onData((d) => {
      if (ws.readyState === WebSocket.OPEN) ws.send(enc.encode(d));
    });

    // Refit + tell the PTY whenever the window resizes.
    const ro = new ResizeObserver(() => {
      try {
        fit.fit();
        sendResize();
      } catch {}
    });
    ro.observe(termHost.current);

    return () => {
      ro.disconnect();
      disp.dispose();
      ws.close();
      term.dispose();
    };
  }, [container]);

  return (
    <div
      className="fixed flex flex-col overflow-hidden rounded-xl border border-border bg-[#0e0e10] shadow-2xl"
      style={{ left: pos.x, top: pos.y, width: size.w, height: size.h, zIndex: z }}
      onPointerDownCapture={bringFront}
    >
      <div
        onPointerDown={startDrag}
        className="flex cursor-move items-center justify-between border-b border-border px-3 py-1.5 select-none"
      >
        <span className="font-mono text-xs text-fg">
          <span className="text-primary">$</span> {container}
        </span>
        <div
          className="flex items-center gap-1"
          onPointerDown={(e) => e.stopPropagation()}
        >
          <input
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === "Enter") {
                e.preventDefault();
                if (query) {
                  if (e.shiftKey) searchRef.current?.findPrevious(query);
                  else searchRef.current?.findNext(query);
                }
              }
            }}
            placeholder="search…"
            className="w-28 rounded border border-border bg-bg px-2 py-0.5 font-mono text-xs outline-none focus:border-primary"
          />
          <button
            onClick={() => query && searchRef.current?.findPrevious(query)}
            className="cursor-pointer rounded px-1 text-muted hover:text-primary"
            title="previous match (Shift+Enter)"
          >
            ↑
          </button>
          <button
            onClick={() => query && searchRef.current?.findNext(query)}
            className="cursor-pointer rounded px-1 text-muted hover:text-primary"
            title="next match (Enter)"
          >
            ↓
          </button>
          <button
            onClick={onClose}
            className="rounded px-2 text-muted hover:text-primary"
          >
            ✕
          </button>
        </div>
      </div>
      <div ref={termHost} className="flex-1 overflow-hidden p-1.5" />
      <div
        onPointerDown={startResize}
        className="absolute bottom-0 right-0 h-4 w-4 cursor-nwse-resize text-muted"
        style={{ lineHeight: "1rem" }}
        title="resize"
      >
        ◢
      </div>
    </div>
  );
}
