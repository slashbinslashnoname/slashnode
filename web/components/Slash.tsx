// Red ASCII slash — the SlashNode logo.
const ART = `      //
     //
    //
   //
  //
 //`;

export function Slash({ className = "" }: { className?: string }) {
  return (
    <pre
      aria-hidden
      className={`text-primary leading-tight text-sm select-none ${className}`}
    >
      {ART}
    </pre>
  );
}
