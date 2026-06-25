import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "SlashNode",
  description: "your node, your rules",
};

// Anti-flash script: applies the theme before the first paint.
const themeScript = `
(function () {
  try {
    var t = localStorage.getItem("slashnode-theme") || "system";
    var dark = t === "dark" || (t === "system" && matchMedia("(prefers-color-scheme: dark)").matches);
    document.documentElement.classList.toggle("dark", dark);
    document.documentElement.dataset.themeMode = t;
  } catch (e) {}
})();
`;

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en" suppressHydrationWarning>
      <head>
        <script dangerouslySetInnerHTML={{ __html: themeScript }} />
      </head>
      <body className="min-h-screen antialiased">{children}</body>
    </html>
  );
}
