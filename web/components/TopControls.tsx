import { ThemeToggle } from "@/components/ThemeToggle";
import { ThemePicker } from "@/components/ThemePicker";
import { SignOutButton } from "@/components/SignOutButton";

// TopControls is the fixed top-right control cluster: the dark/light toggle, the
// background picker (to its right) and, when the UI is password protected, a
// sign-out button. Password protection is driven by the daemon (passed in as an
// env var when it launches the front).
export function TopControls() {
  const protectedMode = process.env.SLASHNODE_PASSWORD_PROTECTED === "true";
  return (
    <div className="fixed top-4 right-4 z-50 flex items-center gap-2">
      <ThemeToggle />
      <ThemePicker />
      {protectedMode && <SignOutButton />}
    </div>
  );
}
