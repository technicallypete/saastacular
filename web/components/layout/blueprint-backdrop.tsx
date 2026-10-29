// Decorative blueprint imagery (G048): an empty, aria-hidden layer painted by globals.css (E. Decorative imagery)
// behind the in-flow content of an isolated, relatively positioned container. Never focusable, no text.
export function BlueprintBackdrop({ className }: { className?: string }) {
  return <div aria-hidden="true" data-blueprint="" className={className} />;
}
