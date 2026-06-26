import Link from "next/link";
import { ThemeToggle } from "@/components/ThemeToggle";
import { ThemePicker } from "@/components/ThemePicker";
import { NodeLinks } from "@/components/NodeLinks";
import { SignOutButton } from "@/components/SignOutButton";

// TopControls is the fixed top-right control cluster: the dark/light toggle, the
// background picker (to its right) and, when the UI is password protected, a
// sign-out button. Password protection is driven by the daemon (passed in as an
// env var when it launches the front).
export function TopControls() {
  return (
    <div className="fixed top-4 right-4 z-50 flex items-center gap-2">
      <NodeLinks />
      <Link
        href="/settings"
        aria-label="Settings"
        title="Settings"
        className="cursor-pointer rounded-lg border border-border bg-card px-3 py-1.5 text-sm hover:border-primary transition-colors"
      >
        ⚙
      </Link>
      <ThemeToggle />
      <ThemePicker />
      <SignOutButton />
    </div>
  );
}
