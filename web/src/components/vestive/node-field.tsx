import * as React from "react"

type Chain = { dur: number; off: number; pts: [number, number][] };

const W = 1360, H = 620;
const RAMP = ["#0075ff", "#0cacf1", "#23c3e7", "#48bdc2", "#5abab0", "#6cb79e", "#91b179", "#b6ab54"];
export const DEFAULT_CHAINS: Chain[] = [
  { dur: 16, off: 0, pts: [[70, 520], [250, 404], [452, 556], [664, 318], [846, 470], [1010, 250], [1180, 392], [1300, 180]] },
  { dur: 13, off: 3.4, pts: [[120, 96], [330, 168], [548, 72], [760, 150], [966, 60], [1210, 140]] },
  { dur: 19, off: 6.1, pts: [[150, 300], [392, 236], [600, 430], [838, 262], [1064, 352], [1290, 296]] },
  { dur: 11, off: 9.2, pts: [[300, 596], [556, 470], [820, 592], [1120, 548]] },
];

export interface NodeFieldProps {
  chains?: Chain[];
  opacity?: number;
}

/** NodeField — the hero background: nodes fire, tracers run to the next node and dissolve. Place inside a relative, overflow-hidden section. */
export function NodeField({ chains = DEFAULT_CHAINS, opacity = 0.62 }: NodeFieldProps) {
  const pc = (v: number, t: number) => `${((v / t) * 100).toFixed(3)}%`;
  const parts: React.ReactNode[] = [];
  chains.forEach((ch, ci) => {
    const step = +(ch.dur * 0.1).toFixed(2);
    ch.pts.forEach(([x0, y0], i) => {
      const c = RAMP[(i + ci * 2) % RAMP.length];
      const t = `${(ch.off + i * step).toFixed(2)}s`;
      if (i < ch.pts.length - 1) {
        const [x1, y1] = ch.pts[i + 1];
        const len = Math.hypot(x1 - x0, y1 - y0);
        const ang = (Math.atan2(y1 - y0, x1 - x0) * 180) / Math.PI;
        parts.push(
          <div key={`e${ci}-${i}`} style={{ position: "absolute", left: pc(x0, W), top: pc(y0, H), width: len, height: 1, transform: `rotate(${ang}deg)`, transformOrigin: "0 50%" }}>
            <span data-motion="edge" style={{ position: "absolute", left: 0, top: 0, height: 1, width: 0, background: `linear-gradient(90deg,rgba(255,255,255,.04),${c})`, animation: `vx-edge-trace ${ch.dur}s linear ${t} infinite` }}>
              <span style={{ position: "absolute", right: -2, top: -2, width: 5, height: 5, background: c }} />
            </span>
          </div>
        );
      }
      const pos: React.CSSProperties = { position: "absolute", left: pc(x0, W), top: pc(y0, H), width: 7, height: 7, margin: "-3px 0 0 -3px" };
      parts.push(<span key={`n${ci}-${i}`} data-motion="node" style={{ ...pos, background: c, animation: `vx-node-fire ${ch.dur}s ease-out ${t} infinite` }} />);
      parts.push(<span key={`r${ci}-${i}`} data-motion="ring" style={{ ...pos, border: `1px solid ${c}`, animation: `vx-node-ring ${ch.dur}s ease-out ${t} infinite` }} />);
    });
  });
  return (
    <div aria-hidden="true" data-slot="node-field" style={{ position: "absolute", left: "50%", top: 0, width: W, height: H, marginLeft: -W / 2, pointerEvents: "none", opacity }}>
      {parts}
    </div>
  );
}
