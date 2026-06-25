"use client";

import { useEffect, useState } from "react";

// BitcoinPrice shows the live BTC/USD spot price (Coinbase public API, fetched
// from the browser).
export function BitcoinPrice() {
  const [price, setPrice] = useState<string | null>(null);

  useEffect(() => {
    let alive = true;
    const load = async () => {
      try {
        const r = await fetch("https://api.coinbase.com/v2/prices/BTC-USD/spot");
        const j = await r.json();
        const amount = Number(j?.data?.amount);
        if (alive && amount) {
          setPrice(amount.toLocaleString("en-US", { maximumFractionDigits: 0 }));
        }
      } catch {
        /* offline — hide */
      }
    };
    load();
    const t = setInterval(load, 60_000);
    return () => {
      alive = false;
      clearInterval(t);
    };
  }, []);

  if (!price) return null;
  return (
    <span className="rounded-lg border border-border bg-card px-3 py-1.5 text-sm">
      <span className="text-primary">₿</span> ${price}
    </span>
  );
}
