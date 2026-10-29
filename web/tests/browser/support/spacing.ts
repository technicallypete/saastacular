import { expect, type Page } from "@playwright/test";

import { polygonPoints } from "./angle";

// Bounded text-spacing trial (WCAG 1.4.12 values) for touched shell and control text. This is a
// focused regression check, not a full 1.4.12 certification.
export const SPACING_CSS =
  "* { line-height: 1.5 !important; letter-spacing: 0.12em !important; word-spacing: 0.16em !important; } p { margin-bottom: 2em !important; }";

// text: a block of text. control: a focusable control whose label stays inside its box (G067: every
// everyday Button and ButtonLink, whose fill is the element's own rounded background). angled: a
// control whose label also stays inside its angled fill (G067: the cta variant only; an angled item
// without a clipped ::before fill is reported). icon: a control without visible text, so its box is
// checked instead of its (screen-reader-only) label.
export type ItemKind = "text" | "control" | "angled" | "icon";

// An expected element: the selector must match exactly texts.length elements, in document order,
// each with exactly that collapsed textContent.
export type SpacingItem = { name: string; selector: string; texts: string[]; kind?: ItemKind; disabled?: boolean };

// The open help dialog's expected elements (exact texts from the shell help dialog source).
export const POPUP = '[data-slot="dialog-content"]';
export const MODAL_ITEMS: SpacingItem[] = [
  { name: "dialog title", selector: `${POPUP} [data-slot="dialog-title"]`, texts: ["Keyboard help"] },
  { name: "dialog description", selector: `${POPUP} [data-slot="dialog-description"]`, texts: ["Everything here works with a keyboard alone."] },
  {
    name: "dialog tips",
    selector: `${POPUP} li span:last-child`,
    texts: [
      "Tab and Shift+Tab move between links and controls.",
      "The first Tab on a page reveals Skip to main content.",
      "Arrow keys choose between Light, Dark and System appearance.",
      "Escape or the Close button closes this dialog.",
    ],
  },
  // G069: the kit DialogContent's own corner Close (showCloseButton; owner kit-fidelity direction, 2026-10-26), a button
  // directly in the popup with its sr-only name "Close" as text, no aria-label attribute.
  { name: "Close", selector: `${POPUP} > [data-slot="dialog-close"]`, texts: ["Close"], kind: "icon" },
];

// The bundled faces (G048) that must be loaded before text spacing is measured (the bounded phase-B
// readiness gate, support/fonts.ts, runs first; this is the in-page precondition). G069: a page loads the faces its text
// uses (the shell pages never request Geist Mono), so the precondition is that no bundled face is still loading or
// failed and that every face of each family a measured item renders in is loaded; a face no text uses stays unloaded.
const BUNDLED_FAMILIES = ["Oswald", "Barlow Condensed", "Geist", "Geist Mono"];
// One FontFace per pinned binary: Oswald, Barlow Condensed 400 and 600, Geist, Geist Mono (G067).
const BUNDLED_FACES = 5;

export type Finding = { item: string; index: number; check: string; detail: string };
type Rect = { x: number; y: number; w: number; h: number };
type Measured = { item: string; index: number; kind: ItemKind; rects: Rect[]; pad?: Rect; poly?: string };
export type SpacingState = { state: string; findings: Finding[]; items: Measured[] };

// Measures every expected element in the current state. With `popup`, the state is an open modal:
// every hit must also lie inside that popup, and the popup and its Close control are checked. With
// `focused`, the state is that element focused: it is (re)focused first and must match :focus.
export async function inspectSpacing(page: Page, state: string, items: SpacingItem[], popup?: string, focused?: string): Promise<SpacingState> {
  if (!focused) await page.mouse.move(0, 0);
  const out = await page.evaluate(
    async ({ items, popup, focused, families, faceCount }) => {
      const findings: { item: string; index: number; check: string; detail: string }[] = [];
      const familyOf = (value: string) => value.split(",")[0].trim().replace(/["']/g, "");
      const faces = [...document.fonts].filter((ff) => families.includes(familyOf(ff.family)));
      const used = new Set(items.flatMap((it) => [...document.querySelectorAll(it.selector)].map((e) => familyOf(getComputedStyle(e).fontFamily))));
      const unready = faces.filter((ff) => ff.status === "loading" || ff.status === "error" || (used.has(familyOf(ff.family)) && ff.status !== "loaded"));
      if (faces.length !== faceCount || unready.length > 0) {
        findings.push({ item: "state", index: 0, check: "fonts", detail: `bundled faces not loaded: ${faces.map((ff) => `${ff.family} ${ff.weight} ${ff.status}`).join("; ")}` });
      }
      const f = focused ? document.querySelector<HTMLElement>(focused) : null;
      f?.focus({ preventScroll: true });
      // Measure the settled state: the kit's own transitions run in full under reduced motion too (G069: the app's
      // global reduced-motion rule is withdrawn; for example the skip link's translate on focus and the Buttons'
      // 150ms transition-all), and they only advance on the next frames.
      await Promise.race([Promise.allSettled(document.getAnimations().map((a) => a.finished)), new Promise((r) => setTimeout(r, 2000))]);
      await new Promise<void>((r) => requestAnimationFrame(() => requestAnimationFrame(() => r())));
      if (focused && (!f || !f.matches(":focus"))) {
        findings.push({ item: "state", index: 0, check: "focus", detail: `${focused} is not :focus (active ${document.activeElement?.tagName ?? "none"}, hasFocus ${document.hasFocus()})` });
      }
      const measured: { item: string; index: number; kind: "text" | "control" | "angled" | "icon"; rects: { x: number; y: number; w: number; h: number }[]; pad?: { x: number; y: number; w: number; h: number }; poly?: string }[] = [];
      const tol = 0.5;
      const collapse = (s: string | null) => (s ?? "").replace(/\s+/g, " ").trim();
      const plain = (r: DOMRect) => ({ x: r.left, y: r.top, w: r.width, h: r.height });
      const describe = (el: Element) => `${el.tagName.toLowerCase()}${el.id ? `#${el.id}` : ""} ${(el.getAttribute("class") ?? "").slice(0, 80)}`.trim();
      const popupEl = popup ? document.querySelector<HTMLElement>(popup) : null;
      if (popup && !popupEl) findings.push({ item: "modal", index: 0, check: "missing", detail: popup });

      // Painted text: rects of non-empty text nodes, excluding screen-reader-only text.
      const textRects = (el: Element) => {
        const rects: { r: DOMRect; from: Element }[] = [];
        const walker = document.createTreeWalker(el, NodeFilter.SHOW_TEXT);
        for (let n = walker.nextNode(); n; n = walker.nextNode()) {
          const parent = n.parentElement;
          if (!parent || !n.textContent?.trim() || parent.closest(".sr-only")) continue;
          const range = document.createRange();
          range.selectNodeContents(n);
          for (const r of range.getClientRects()) if (r.width > 0 && r.height > 0) rects.push({ r, from: parent });
        }
        return rects;
      };
      const rectsOf = (el: HTMLElement, kind: string) => (kind === "icon" ? [{ r: el.getBoundingClientRect(), from: el.parentElement ?? el }] : textRects(el));

      const all: { key: string; el: HTMLElement; kind: string }[] = [];
      for (const item of items) {
        const kind = item.kind ?? "text";
        const els = [...document.querySelectorAll<HTMLElement>(item.selector)];
        if (els.length !== item.texts.length) findings.push({ item: item.name, index: -1, check: "missing", detail: `found ${els.length}, expected ${item.texts.length}` });
        els.forEach((el, i) => {
          const f = (check: string, detail: string) => findings.push({ item: item.name, index: i, check, detail });
          // Drawn text only: Base UI's Select icon carries a "▼" text node inside its svg (never painted, aria-hidden).
          const clone = el.cloneNode(true) as HTMLElement;
          clone.querySelectorAll("svg").forEach((s) => s.remove());
          const own = collapse(clone.textContent);
          if (i < item.texts.length && own !== item.texts[i]) f("text", `"${own}" instead of "${item.texts[i]}"`);
          const b = el.getBoundingClientRect();
          if (!el.checkVisibility({ checkOpacity: true, checkVisibilityCSS: true }) || el.getClientRects().length === 0 || b.width <= 0 || b.height <= 0) {
            f("hidden", "not visible");
            return;
          }
          el.scrollIntoView({ block: "center", inline: "nearest" });
          const rects = rectsOf(el, kind);
          if (rects.length === 0) {
            f("hidden", "no painted text");
            return;
          }
          all.push({ key: `${item.name}#${i}`, el, kind });
          // Not covered: the hit at the centre of each text rect is the item (pointer-events are
          // forced on for the probe, since disabled controls do not take pointer events).
          const saved = el.style.getPropertyValue("pointer-events");
          const savedPriority = el.style.getPropertyPriority("pointer-events");
          el.style.setProperty("pointer-events", "auto", "important");
          for (const { r } of rects) {
            const x = r.left + r.width / 2;
            const y = r.top + r.height / 2;
            const hit = document.elementFromPoint(x, y);
            if (!hit || !(hit === el || el.contains(hit))) f("covered", `at (${x.toFixed(1)}, ${y.toFixed(1)}) by ${hit ? describe(hit) : "nothing"}`);
            else if (popupEl && !popupEl.contains(hit)) f("covered", `at (${x.toFixed(1)}, ${y.toFixed(1)}) outside the modal`);
          }
          el.style.setProperty("pointer-events", saved, savedPriority);
          // Not clipped by an overflow: hidden/clip ancestor; reachable in every scrolling ancestor
          // and in the viewport after scrolling it into view.
          for (const { r, from } of rects) {
            for (let a: Element | null = from; a && a !== document.documentElement; a = a.parentElement) {
              const s = getComputedStyle(a);
              const ab = a.getBoundingClientRect();
              const left = ab.left + a.clientLeft;
              const top = ab.top + a.clientTop;
              const right = left + a.clientWidth;
              const bottom = top + a.clientHeight;
              const outX = r.left < left - tol || r.right > right + tol;
              const outY = r.top < top - tol || r.bottom > bottom + tol;
              for (const [axis, overflow, out] of [["x", s.overflowX, outX], ["y", s.overflowY, outY]] as const) {
                if (!out) continue;
                if (overflow === "hidden" || overflow === "clip") f("clipped", `${axis} by ${describe(a)}`);
                else if (overflow === "auto" || overflow === "scroll") f("unreachable", `${axis} in scrolling ${describe(a)}`);
              }
            }
            if (r.left < -tol || r.top < -tol || r.right > window.innerWidth + tol || r.bottom > window.innerHeight + tol) {
              f("unreachable", `outside the viewport after scrolling: ${JSON.stringify(plain(r))}`);
            }
          }
          // Labels stay inside their control's box.
          if (kind === "control" || kind === "angled") {
            for (const { r } of rects) {
              const cb = el.getBoundingClientRect();
              if (r.left < cb.left - tol || r.top < cb.top - tol || r.right > cb.right + tol || r.bottom > cb.bottom + tol) f("label-outside", `label ${JSON.stringify(plain(r))} outside ${JSON.stringify(plain(cb))}`);
            }
          }
          const m: (typeof measured)[number] = { item: item.name, index: i, kind, rects: rects.map(({ r }) => plain(r)) };
          if (kind === "angled") {
            // ::before (absolute, inset 0) fills the padding box, inside the border.
            const cb = el.getBoundingClientRect();
            m.pad = { x: cb.left + el.clientLeft, y: cb.top + el.clientTop, w: el.clientWidth, h: el.clientHeight };
            m.poly = getComputedStyle(el, "::before").clipPath;
          }
          measured.push(m);
        });
      }

      // No overlap: no two items' text rects intersect, measured in one pass without scrolling.
      const now = all.map(({ key, el, kind }) => ({ key, rects: rectsOf(el, kind).map(({ r }) => r) }));
      for (let i = 0; i < now.length; i++)
        for (let j = i + 1; j < now.length; j++)
          for (const a of now[i].rects)
            for (const b of now[j].rects)
              if (a.left < b.right - tol && b.left < a.right - tol && a.top < b.bottom - tol && b.top < a.bottom - tol) {
                const [name, index] = now[i].key.split("#");
                findings.push({ item: name, index: Number(index), check: "overlap", detail: `with ${now[j].key}` });
              }

      // Controls stay usable: enabled controls take focus; disabled ones are :disabled and do not.
      for (const item of items) {
        const kind = item.kind ?? "text";
        if (kind === "text") continue;
        document.querySelectorAll<HTMLElement>(item.selector).forEach((el, i) => {
          const target = el.matches("a, button, input, select") ? el : el.querySelector<HTMLElement>("a, button, input, select");
          const prev = document.activeElement as HTMLElement | null;
          if (!target) {
            findings.push({ item: item.name, index: i, check: "unusable", detail: "no focusable control" });
            return;
          }
          target.focus({ preventScroll: true });
          const focused = document.activeElement === target;
          if (item.disabled) {
            if (!target.matches(":disabled")) findings.push({ item: item.name, index: i, check: "unusable", detail: "not :disabled" });
            if (focused) findings.push({ item: item.name, index: i, check: "unusable", detail: "a disabled control took focus" });
          } else if (!focused) findings.push({ item: item.name, index: i, check: "unusable", detail: "does not take focus" });
          prev?.focus({ preventScroll: true });
        });
      }

      const doc = document.documentElement;
      if (doc.scrollWidth > doc.clientWidth + tol) findings.push({ item: "page", index: 0, check: "page-overflow", detail: `scrollWidth ${doc.scrollWidth} > ${doc.clientWidth}` });

      if (popupEl) {
        popupEl.scrollTop = 0;
        const inView = (r: DOMRect) => r.left >= -tol && r.top >= -tol && r.right <= window.innerWidth + tol && r.bottom <= window.innerHeight + tol;
        const pr = popupEl.getBoundingClientRect();
        if (!inView(pr)) findings.push({ item: "modal", index: 0, check: "modal", detail: `popup outside the viewport: ${JSON.stringify(plain(pr))}` });
        const close = popupEl.querySelector<HTMLElement>(":scope > button");
        if (!close) findings.push({ item: "modal", index: 0, check: "modal", detail: "no Close control" });
        else {
          const cr = close.getBoundingClientRect();
          if (!inView(cr)) findings.push({ item: "modal", index: 0, check: "modal", detail: `Close outside the viewport: ${JSON.stringify(plain(cr))}` });
          for (const sel of ["[data-slot=dialog-title]", "[data-slot=dialog-description]"]) {
            const t = popupEl.querySelector(sel);
            for (const { r } of t ? textRects(t) : [])
              if (cr.left < r.right - tol && r.left < cr.right - tol && cr.top < r.bottom - tol && r.top < cr.bottom - tol) findings.push({ item: "modal", index: 0, check: "modal", detail: `Close intersects ${sel}` });
          }
        }
      }
      return { findings, measured };
    },
    { items, popup, focused, families: BUNDLED_FAMILIES, faceCount: BUNDLED_FACES },
  );
  // Angled labels stay inside the polygon of their clipped fill. An item declared angled whose ::before
  // carries no polygon (an everyday button since G067) is a finding, never a skipped check.
  for (const m of out.measured) {
    if (m.kind !== "angled") continue;
    if (!m.pad || !m.poly || !m.poly.startsWith("polygon(")) {
      out.findings.push({ item: m.item, index: m.index, check: "angle", detail: `no angled ::before fill (clip-path ${m.poly ?? "unread"})` });
      continue;
    }
    const pad = m.pad;
    const pts = polygonPoints(m.poly, pad.w, pad.h).map(([x, y]) => [pad.x + x, pad.y + y]);
    const inside = (x: number, y: number) => {
      let r = false;
      for (let i = 0, j = pts.length - 1; i < pts.length; j = i++) {
        const [xi, yi] = pts[i];
        const [xj, yj] = pts[j];
        if (yi > y !== yj > y && x < ((xj - xi) * (y - yi)) / (yj - yi) + xi) r = !r;
      }
      return r;
    };
    for (const t of m.rects) {
      for (const [x, y] of [[t.x, t.y], [t.x + t.w, t.y], [t.x, t.y + t.h], [t.x + t.w, t.y + t.h]]) {
        if (!inside(x, y)) out.findings.push({ item: m.item, index: m.index, check: "angle", detail: `label corner (${x.toFixed(1)}, ${y.toFixed(1)}) under the angled cut` });
      }
    }
  }
  return { state, findings: out.findings, items: out.measured };
}

export function expectNoFindings(result: SpacingState) {
  expect(result.findings, `text spacing: ${result.state}`).toEqual([]);
}

// Focused negative cases: scratch elements that each break exactly one property must be reported by
// the matching check, and text below the fold of a scrolling box must pass. The scratch elements are
// removed afterwards. With `modal`, an opaque element placed over the open dialog's Close control
// must be reported as covering it.
export async function expectNegativeCases(page: Page, modal?: string): Promise<SpacingState[]> {
  const states: SpacingState[] = [];
  if (modal) {
    // The cover is placed inside the popup (absolutely, over Close), so the modal never marks it inert.
    await page.evaluate((sel) => {
      const popup = document.querySelector<HTMLElement>(sel)!;
      const close = popup.querySelector(":scope > button")!.getBoundingClientRect();
      const p = popup.getBoundingClientRect();
      const cover = document.createElement("div");
      cover.id = "spacing-negative";
      cover.style.cssText = `position:absolute;left:${close.left - p.left - popup.clientLeft}px;top:${close.top - p.top - popup.clientTop + popup.scrollTop}px;width:${close.width}px;height:${close.height}px;background:#000;z-index:2147483647`;
      popup.appendChild(cover);
    }, modal);
    const r = await inspectSpacing(page, "negative: covered Close", [{ name: "neg-close", selector: `${modal} > [data-slot="dialog-close"]`, texts: ["Close"], kind: "icon" }], modal);
    expect(r.findings.filter((f) => f.item === "neg-close").map((f) => f.check), "an opaque element over Close is reported").toContain("covered");
    states.push(r);
  } else {
    await page.evaluate(() => {
      const box = document.createElement("div");
      box.id = "spacing-negative";
      box.style.cssText = `position:absolute;left:8px;top:${window.scrollY + 8}px;width:280px;z-index:2147483646;background:#fff;color:#000`;
      box.innerHTML = [
        '<p data-neg="vis" style="visibility:hidden">Hidden text</p>',
        '<p data-neg="none" style="display:none">Undisplayed text</p>',
        '<div style="position:relative"><p data-neg="covered">Covered text</p><div style="position:absolute;inset:0;background:#000"></div></div>',
        '<div style="position:relative;height:3em"><p data-neg="ov1" style="position:absolute;top:0;left:0;margin:0">Overlap one</p><p data-neg="ov2" style="position:absolute;top:0;left:10px;margin:0">Overlap two</p></div>',
        '<div style="width:60px;overflow:hidden"><p data-neg="clip" style="white-space:nowrap">This text is far too long to fit</p></div>',
        '<div style="height:60px;overflow:auto"><p data-neg="unreach" style="position:relative;left:-600px;white-space:nowrap">Unreachable text</p></div>',
        // An angled fill clipped to the kit's angle token (G069: --vx-angle, kit styles/theme.css), with no padding.
        '<style>#spacing-negative [data-neg="angle"]{position:relative;isolation:isolate}#spacing-negative [data-neg="angle"]::before{content:"";position:absolute;inset:0;z-index:-1;clip-path:var(--vx-angle)}</style>',
        '<button data-neg="angle" style="padding:0;height:32px">Cut label</button>',
        '<div style="height:60px;overflow:auto"><div style="height:240px"></div><p data-neg="below">Below the fold</p></div>',
      ].join("");
      document.body.appendChild(box);
    });
    const neg = (name: string, text: string, kind?: ItemKind): SpacingItem => ({ name, selector: `#spacing-negative [data-neg="${name}"]`, texts: [text], kind });
    const cases: [SpacingItem, string][] = [
      [neg("missing", "Missing text"), "missing"],
      [neg("vis", "Hidden text"), "hidden"],
      [neg("none", "Undisplayed text"), "hidden"],
      [neg("covered", "Covered text"), "covered"],
      [neg("clip", "This text is far too long to fit"), "clipped"],
      [neg("unreach", "Unreachable text"), "unreachable"],
      [neg("angle", "Cut label", "angled"), "angle"],
    ];
    for (const [item, check] of cases) {
      const r = await inspectSpacing(page, `negative: ${item.name}`, [item]);
      expect(r.findings.map((f) => f.check), `negative case ${item.name} is reported as ${check}`).toContain(check);
      states.push(r);
    }
    const overlap = await inspectSpacing(page, "negative: overlap", [neg("ov1", "Overlap one"), neg("ov2", "Overlap two")]);
    expect(overlap.findings.map((f) => f.check), "overlapping blocks are reported").toContain("overlap");
    states.push(overlap);
    const below = await inspectSpacing(page, "positive: below the fold of a scrolling box", [neg("below", "Below the fold")]);
    expect(below.findings, "text below the fold of a scrolling box is reachable").toEqual([]);
    states.push(below);
  }
  await page.evaluate(() => document.getElementById("spacing-negative")?.remove());
  await page.evaluate(() => window.scrollTo(0, 0));
  return states;
}

export type Reflow = {
  scrollWidth: number;
  clientWidth: number;
  outside: string[];
  overlaps: string[];
  h1Overflow: boolean;
  h1Lines: number;
  columns: number;
};

// The responsive overlap detector shared by S-RESPONSIVE, S-LONG-IDENTITY and the font checks. `scope` limits
// the overlap and inside-the-viewport checks to one region (the header while a dialog covers the page).
export async function reflowLayout(page: Page, scope = "body"): Promise<Reflow> {
  return page.evaluate((scopeSelector) => {
    const doc = document.documentElement;
    // The appearance control counts once, as its wrapper (the kit Select's trigger inside it is the same box).
    const controls = [...document.querySelector(scopeSelector)!.querySelectorAll<HTMLElement>("a, button, [data-appearance-control]")]
      .filter((el) => !el.classList.contains("sr-only") && el.getClientRects().length > 0)
      .filter((el) => el.matches("[data-appearance-control]") || el.closest("[data-appearance-control]") === null)
      .map((el) => ({ name: (el.textContent ?? "").trim().slice(0, 30), r: el.getBoundingClientRect() }));
    const overlaps: string[] = [];
    for (let i = 0; i < controls.length; i++)
      for (let j = i + 1; j < controls.length; j++) {
        const a = controls[i].r, b = controls[j].r;
        if (a.left < b.right - 0.5 && b.left < a.right - 0.5 && a.top < b.bottom - 0.5 && b.top < a.bottom - 0.5) overlaps.push(`${controls[i].name} / ${controls[j].name}`);
      }
    const h1 = document.querySelector("h1")!;
    const h1Style = getComputedStyle(h1);
    const lineHeight = parseFloat(h1Style.lineHeight) || parseFloat(h1Style.fontSize) * 1.2;
    const grid = document.querySelector("[data-slot=ruled-grid] > div") as HTMLElement | null;
    // G069 S14: the grid's auto-fit repeat collapses its empty tracks to 0px, and the computed track list still names
    // them; the columns a reader sees are the tracks with a width.
    const tracks = grid ? getComputedStyle(grid).gridTemplateColumns.split(" ").filter(Boolean) : [];
    return {
      scrollWidth: doc.scrollWidth,
      clientWidth: doc.clientWidth,
      outside: controls.filter((c) => c.r.left < -0.5 || c.r.right > window.innerWidth + 0.5).map((c) => c.name),
      overlaps,
      h1Overflow: h1.scrollWidth > h1.clientWidth + 0.5,
      h1Lines: Math.round(h1.getBoundingClientRect().height / lineHeight),
      columns: tracks.filter((t) => parseFloat(t) > 0).length,
    };
  }, scope);
}

export function expectReflow(layout: Reflow) {
  expect(layout.scrollWidth, "no horizontal page scroll").toBeLessThanOrEqual(layout.clientWidth);
  expect(layout.outside, "controls outside the viewport").toEqual([]);
  expect(layout.overlaps, "overlapping controls").toEqual([]);
  expect(layout.h1Overflow, "the product name wraps").toBe(false);
}
