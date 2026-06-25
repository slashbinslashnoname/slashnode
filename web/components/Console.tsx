"use client";

import { useEffect, useRef } from "react";
import { Terminal } from "@xterm/xterm";
import { FitAddon } from "@xterm/addon-fit";
import "@xterm/xterm/css/xterm.css";

// Console opens an interactive shell into a container over a WebSocket
// (slashnoded `docker exec -it`, proxied by Caddy at /__console).
export function Console({
  container,
  onClose,
}: {
  container: string;
  onClose: () => void;
}) {
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!ref.current) return;
    const term = new Terminal({
      convertEol: true,
      fontSize: 13,
      fontFamily: "ui-monospace, SFMono-Regular, Menlo, monospace",
      theme: { background: "#0e0e10", foreground: "#f5f5f5" },
    });
    const fit = new FitAddon();
    term.loadAddon(fit);
    term.open(ref.current);
    fit.fit();

    const proto = location.protocol === "https:" ? "wss:" : "ws:";
    const ws = new WebSocket(
      `${proto}//${location.host}/__console?c=${encodeURIComponent(container)}`,
    );
    ws.binaryType = "arraybuffer";

    ws.onopen = () => term.writeln(`\x1b[2mconnected to ${container}\x1b[0m`);
    ws.onmessage = (e) => term.write(new Uint8Array(e.data as ArrayBuffer));
    ws.onclose = () => term.writeln("\r\n\x1b[2m[disconnected]\x1b[0m");
    ws.onerror = () =>
      term.writeln(
        "\r\n\x1b[31mconnection failed — open via the Caddy URL (https://…) for the console\x1b[0m",
      );

    const disp = term.onData((d) => {
      if (ws.readyState === WebSocket.OPEN) ws.send(d);
    });
    const onResize = () => fit.fit();
    window.addEventListener("resize", onResize);

    return () => {
      window.removeEventListener("resize", onResize);
      disp.dispose();
      ws.close();
      term.dispose();
    };
  }, [container]);

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4"
      onClick={onClose}
    >
      <div
        className="flex h-[70vh] w-full max-w-3xl flex-col overflow-hidden rounded-xl border border-border bg-[#0e0e10]"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="flex items-center justify-between border-b border-border px-4 py-2">
          <span className="font-mono text-sm text-fg">
            <span className="text-primary">$</span> {container}
          </span>
          <button
            onClick={onClose}
            className="rounded px-2 text-muted hover:text-primary"
          >
            ✕
          </button>
        </div>
        <div ref={ref} className="flex-1 overflow-hidden p-2" />
      </div>
    </div>
  );
}
