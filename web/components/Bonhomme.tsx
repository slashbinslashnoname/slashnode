// Bonhomme rouge ASCII — mascotte de SlashNode.
const ART = `   ___
  / o o\\
  \\  ^  /
   |||||
  / ||| \\
   /   \\`;

export function Bonhomme({ className = "" }: { className?: string }) {
  return (
    <pre
      aria-hidden
      className={`text-primary leading-tight text-sm select-none ${className}`}
    >
      {ART}
    </pre>
  );
}
