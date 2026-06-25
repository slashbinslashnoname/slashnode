import { Skull } from "@/components/Skull";
import { ThemeToggle } from "@/components/ThemeToggle";
import { LoginForm } from "@/components/LoginForm";

export const dynamic = "force-dynamic";

export default async function Login({
  searchParams,
}: {
  searchParams: Promise<{ next?: string }>;
}) {
  const { next } = await searchParams;

  return (
    <main className="flex min-h-screen flex-col items-center justify-center gap-6 px-4">
      <ThemeToggle />
      <Skull />
      <div className="text-center">
        <h1 className="text-2xl font-bold tracking-widest">
          <span className="text-primary">/</span>SlashNode
        </h1>
        <p className="text-muted">enter your password to continue</p>
      </div>
      <LoginForm next={next ?? "/"} />
    </main>
  );
}
