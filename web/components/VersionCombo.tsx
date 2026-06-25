"use client";

import { useEffect, useRef, useState } from "react";

// VersionCombo is an autocomplete combobox for picking an image tag out of a
// long list. Unlike a native <datalist>, a pre-filled value never filters the
// options away: the full list shows until the operator actually types to narrow
// it. Free text is allowed (any tag can be entered), so it stays usable even
// when the registry returns nothing.
export function VersionCombo({
  value,
  options,
  onChange,
  disabled,
  placeholder = "tag",
}: {
  value: string;
  options: string[];
  onChange: (tag: string) => void;
  disabled?: boolean;
  placeholder?: string;
}) {
  const [open, setOpen] = useState(false);
  const [query, setQuery] = useState("");
  const [touched, setTouched] = useState(false);
  const [active, setActive] = useState(0);
  const wrap = useRef<HTMLDivElement>(null);

  // Close on outside click.
  useEffect(() => {
    function onDoc(e: MouseEvent) {
      if (wrap.current && !wrap.current.contains(e.target as Node)) {
        setOpen(false);
        setTouched(false);
      }
    }
    document.addEventListener("mousedown", onDoc);
    return () => document.removeEventListener("mousedown", onDoc);
  }, []);

  // Only filter once the operator types; a prefilled value shows the full list.
  const filtered =
    touched && query.trim()
      ? options.filter((o) => o.toLowerCase().includes(query.toLowerCase()))
      : options;

  function choose(tag: string) {
    onChange(tag);
    setOpen(false);
    setTouched(false);
    setQuery("");
  }

  function onKey(e: React.KeyboardEvent) {
    if (e.key === "ArrowDown") {
      e.preventDefault();
      setOpen(true);
      setActive((a) => Math.min(a + 1, filtered.length - 1));
    } else if (e.key === "ArrowUp") {
      e.preventDefault();
      setActive((a) => Math.max(a - 1, 0));
    } else if (e.key === "Enter") {
      if (open && filtered[active]) {
        e.preventDefault();
        choose(filtered[active]);
      }
    } else if (e.key === "Escape") {
      setOpen(false);
      setTouched(false);
    }
  }

  return (
    <div ref={wrap} className="relative w-40">
      <input
        value={touched ? query : value}
        autoComplete="off"
        spellCheck={false}
        disabled={disabled}
        placeholder={placeholder}
        onFocus={() => {
          setOpen(true);
          setActive(0);
        }}
        onChange={(e) => {
          setTouched(true);
          setOpen(true);
          setActive(0);
          setQuery(e.target.value);
          onChange(e.target.value);
        }}
        onKeyDown={onKey}
        className="w-full rounded-md border border-border bg-bg px-2 py-1 text-xs outline-none focus:border-primary"
      />
      {open && filtered.length > 0 && (
        <ul className="absolute z-20 mt-1 max-h-56 w-full overflow-auto rounded-md border border-border bg-card py-1 text-xs shadow-lg">
          {filtered.map((o, i) => (
            <li key={o}>
              <button
                type="button"
                onMouseDown={(e) => {
                  e.preventDefault();
                  choose(o);
                }}
                onMouseEnter={() => setActive(i)}
                className={`block w-full px-2 py-1 text-left ${
                  i === active ? "bg-primary text-white" : "hover:bg-bg"
                } ${o === value ? "font-semibold" : ""}`}
              >
                {o}
              </button>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}
