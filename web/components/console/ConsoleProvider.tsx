"use client";

import {
  createContext,
  useCallback,
  useContext,
  useRef,
  useState,
} from "react";
import { ConsoleWindow } from "./ConsoleWindow";

type Win = { id: number; container: string };

const Ctx = createContext<{ open: (container: string) => void }>({
  open: () => {},
});

export function useConsole() {
  return useContext(Ctx);
}

// ConsoleProvider lets any component open container consoles; each is a separate
// movable/resizable window and several can be open at once.
export function ConsoleProvider({ children }: { children: React.ReactNode }) {
  const [wins, setWins] = useState<Win[]>([]);
  const idRef = useRef(0);

  const open = useCallback((container: string) => {
    setWins((w) => [...w, { id: ++idRef.current, container }]);
  }, []);
  const close = (id: number) => setWins((w) => w.filter((x) => x.id !== id));

  return (
    <Ctx.Provider value={{ open }}>
      {children}
      {wins.map((w, i) => (
        <ConsoleWindow
          key={w.id}
          container={w.container}
          index={i}
          onClose={() => close(w.id)}
        />
      ))}
    </Ctx.Provider>
  );
}
