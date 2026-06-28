"use client";

import { useData } from "@/components/store/DataProvider";

// BitcoinPrice shows the live BTC/USD spot price. The value lives in the global
// DataProvider (polled once, shared across pages), so navigating doesn't refetch.
export function BitcoinPrice() {
  const { btc } = useData();
  if (!btc) return null;
  return (
    <span className="rounded-lg border border-border bg-card px-3 py-1.5 text-sm">
      <span className="text-primary">₿</span> ${btc}
    </span>
  );
}
