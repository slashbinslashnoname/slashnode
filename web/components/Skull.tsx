// Red ASCII skull — the SlashNode mascot.
const ART = `     .-------.
    /         \\
   |  ()   ()  |
   |     ^     |
   |  '-----'  |
    \\  |||||  /
     '-------'`;

export function Skull({ className = "" }: { className?: string }) {
  return (
    <pre
      aria-hidden
      className={`text-primary leading-tight text-sm select-none ${className}`}
    >
      {ART}
    </pre>
  );
}
