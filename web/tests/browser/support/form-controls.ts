import { expect, type Locator, type Page, type TestInfo } from "@playwright/test";

import { cssToRgba, distance, hex, over, rgb, type RGBA } from "./color";
import { expectFocusRing, expectNonText, haloOf, settled } from "./contrast";
import { expectAtKit } from "./modes";
import { pixelAt, type RGB } from "./png";

// Shared checks for the kit form controls of the labels, enrollment, membership, onboarding and branding forms (G069
// Plan §6.2, retyping the G068 helpers): the kit RadioGroup options (components/controls/radio-choice over the bare kit
// RadioGroupItem) and the focus-state motion the kit fields share (Input and Textarea: transition-[color,box-shadow]
// duration-200; tokens.spec.ts S-FIELD-MOTION). Since the owner's kit-fidelity direction (2026-10-26: "exact kit,
// overrides included") the kit renders unaltered: no app boundary class, no app reduced-motion rule; a contrast pair the
// kit paints below its former threshold is asserted at the kit's ratio as OWNER-WAIVED (support/modes.ts expectAtKit).
// Expected colours are read from the page's own tokens at the control (so an Organization's palette applies), never
// from fixed values.

const seconds = (v: string) => (v.trim().endsWith("ms") ? parseFloat(v) / 1000 : parseFloat(v));
const frames = (page: Page) => page.evaluate(() => new Promise<void>((r) => requestAnimationFrame(() => requestAnimationFrame(() => r()))));
const blur = (page: Page) => page.evaluate(() => (document.activeElement as HTMLElement | null)?.blur());
const WHITE: RGB = [255, 255, 255];

async function near(page: Page, actual: string, expected: string, what: string) {
  const a = rgb(await cssToRgba(page, actual));
  const e = rgb(await cssToRgba(page, expected));
  expect(distance(a, e), `${what}: ${hex(a)}, expected ${hex(e)}`).toBeLessThanOrEqual(2);
}

// One token's computed value at the control (a probe next to it, so the control's own palette applies).
async function tokenAt(control: Locator, name: string): Promise<string> {
  return control.evaluate((e, n) => {
    const probe = document.createElement("span");
    e.parentElement!.append(probe);
    probe.style.color = `var(--${n})`;
    const value = getComputedStyle(probe).color;
    probe.remove();
    return value;
  }, name);
}

// A token at an opacity modifier (Tailwind's `<utility>-<token>/<percent>`: color-mix in oklab with transparent),
// resolved by the browser through a detached probe and composited over a ground.
async function mixOver(page: Page, token: string, percent: number, ground: RGB): Promise<RGB> {
  const computed = await page.evaluate((e) => {
    const probe = document.createElement("div");
    probe.style.backgroundColor = e;
    document.body.appendChild(probe);
    const c = getComputedStyle(probe).backgroundColor;
    probe.remove();
    return c;
  }, `color-mix(in oklab, ${token} ${percent}%, transparent)`);
  return over(await cssToRgba(page, computed), ground);
}

// The kit's settled keyboard focus state on a field or radio (G069 S22): no outline, the border in the ring token and
// a 3px ring of the ring token at 30 % (focus-visible:border-ring focus-visible:ring-ring/30 focus-visible:ring-3). The
// ring is compared composited over white: a translucent colour read back through the canvas is premultiplied at 8 bits.
async function expectKitFocusState(page: Page, control: Locator, what: string): Promise<void> {
  const ring = await tokenAt(control, "ring");
  const s = await control.evaluate((e) => {
    const c = getComputedStyle(e);
    return { outline: c.outlineStyle, border: c.borderTopColor, shadow: c.boxShadow };
  });
  expect(s.outline, `${what}: settled focus draws no outline`).toBe("none");
  await near(page, s.border, ring, `${what}: settled focus border (the ring token)`);
  const halo = haloOf(s.shadow);
  expect(halo, `${what}: settled focus ring (${s.shadow})`).not.toBeNull();
  expect(halo!.spread, `${what}: ring width`).toBe(3);
  const haloColor = await cssToRgba(page, halo!.color);
  expect(Math.abs(haloColor[3] - 0.3 * 255), `${what}: ring alpha ${haloColor[3]}`).toBeLessThanOrEqual(2);
  const want: RGBA = [...rgb(await cssToRgba(page, ring)), Math.round(0.3 * 255)] as RGBA;
  expect(distance(over(haloColor, WHITE), over(want, WHITE)), `${what}: ring colour (the ring token at 30 %)`).toBeLessThanOrEqual(2);
}

// Keyboard focus on a control: focused programmatically, then Shift+Tab and Tab, so :focus-visible applies as for a
// keyboard user; waits for the focus transition.
async function keyboardFocus(page: Page, control: Locator, what: string): Promise<void> {
  await control.focus();
  await page.keyboard.press("Shift+Tab");
  await page.keyboard.press("Tab");
  await expect(control, `${what}: keyboard focus`).toBeFocused();
  await settled(control);
  await frames(page);
}

// A kit field (Input, Textarea; G069 retype of the G068 outline motion): its focus state (colour and box-shadow)
// transitions over the kit's 200ms under both motion preferences (the kit's components carry no reduced-motion variant
// and its own reduce rule, in the kit's brand stylesheet, stops only the brand ribbon motion; the app's global reduced-motion rule is
// withdrawn by the owner's kit-fidelity direction, 2026-10-26), nothing is drawn at rest (no ring, no outline), and the
// settled keyboard focus is the kit's ring state. Ends blurred with reduced-motion emulation off.
export async function expectFocusMotion(page: Page, control: Locator, what: string): Promise<void> {
  for (const motion of ["no-preference", "reduce"] as const) {
    await blur(page);
    await page.emulateMedia({ reducedMotion: motion });
    await frames(page);
    const s = await control.evaluate((e) => {
      const c = getComputedStyle(e);
      return { property: c.transitionProperty, duration: c.transitionDuration, outline: c.outlineStyle };
    });
    expect(s.property.split(",").map((p) => p.trim()), `${what} (${motion}): transitioned properties`).toEqual(expect.arrayContaining(["color", "box-shadow"]));
    const durations = s.duration.split(",").map(seconds);
    expect(durations.every((d) => Math.abs(d - 0.2) < 1e-6), `${what} (${motion}): the kit's own 200ms transition, never suppressed (${s.duration})`).toBe(true);
    expect(s.outline, `${what} (${motion}): no outline at rest`).toBe("none");
    // The blur above starts the ring's 200ms transition back to nothing (in both motion preferences): wait it out.
    await expect
      .poll(() => control.evaluate((e) => getComputedStyle(e).boxShadow).then((v) => haloOf(v)), { message: `${what} (${motion}): no focus ring at rest`, timeout: 2000 })
      .toBeNull();
    await keyboardFocus(page, control, `${what} (${motion})`);
    await expectKitFocusState(page, control, `${what} (${motion})`);
  }
  await blur(page);
  await page.emulateMedia({ reducedMotion: null });
}

type RadioFacts = {
  role: string | null;
  ariaChecked: string | null;
  checked: boolean;
  disabled: boolean;
  dark: boolean;
  width: number;
  height: number;
  radius: number;
  border: string;
  borderColor: string;
  background: string;
  opacity: string;
  dot: { width: number; height: number } | null;
  label: string;
  labelColor: string;
  rootFont: number;
  font: number;
  line: number;
  offset: number;
  lines: number;
  transition: string;
  tokens: Record<"primary" | "primary-foreground" | "input" | "muted-foreground" | "foreground", string>;
};

// The kit radio is Base UI's role="radio" element with aria-checked and data-checked/data-disabled; its label is the
// kit Label that RadioChoice renders beside it.
async function radioFacts(radio: Locator): Promise<RadioFacts> {
  return radio.evaluate((el) => {
    const label = el.parentElement!.querySelector<HTMLElement>(":scope > label");
    if (!label) throw new Error("radio option: no label beside the radio");
    const c = getComputedStyle(el);
    const t = getComputedStyle(label);
    const r = el.getBoundingClientRect();
    const tr = label.getBoundingClientRect();
    const dot = el.querySelector<HTMLElement>("[data-slot=radio-group-indicator] span");
    const probe = document.createElement("span");
    el.parentElement!.append(probe);
    const tokens = {} as Record<string, string>;
    for (const name of ["primary", "primary-foreground", "input", "muted-foreground", "foreground"]) {
      probe.style.color = `var(--${name})`;
      tokens[name] = getComputedStyle(probe).color;
    }
    probe.remove();
    const line = parseFloat(t.lineHeight);
    const db = dot?.getBoundingClientRect();
    return {
      role: el.getAttribute("role"),
      ariaChecked: el.getAttribute("aria-checked"),
      checked: el.getAttribute("aria-checked") === "true",
      disabled: el.hasAttribute("data-disabled"),
      dark: document.documentElement.classList.contains("dark"),
      width: r.width,
      height: r.height,
      radius: parseFloat(c.borderTopLeftRadius),
      border: c.borderTopWidth,
      borderColor: c.borderTopColor,
      background: c.backgroundColor,
      opacity: c.opacity,
      dot: db && db.width > 0 ? { width: db.width, height: db.height } : null,
      label: (label.textContent ?? "").replace(/\s+/g, " ").trim(),
      labelColor: t.color,
      rootFont: parseFloat(getComputedStyle(document.documentElement).fontSize),
      font: parseFloat(t.fontSize),
      line,
      // The circle's centre against the first line box's centre of its label.
      offset: r.top + r.height / 2 - (tr.top + line / 2),
      lines: Math.round(tr.height / line),
      transition: c.transitionDuration,
      tokens: tokens as RadioFacts["tokens"],
    };
  });
}

async function expectAligned(radios: Locator, what: string): Promise<number> {
  let wrapped = 0;
  for (let i = 0; i < (await radios.count()); i++) {
    const f = await radioFacts(radios.nth(i));
    expect(Math.abs(f.offset), `${what} option ${i + 1}: circle centred on the first text line (offset ${f.offset.toFixed(2)}px)`).toBeLessThanOrEqual(1);
    if (f.lines > 1) wrapped++;
  }
  return wrapped;
}

// AC1/AC3 (G068, retyped to the kit radio, G069 Plan §6.2; G069 owner kit-fidelity direction, 2026-10-26: the bare kit
// item): every option of the group is a kit radio (role="radio" with aria-checked) named by its label, a 16px circle
// with the kit's 1px transparent border (border-transparent; the app's solid input boundary is withdrawn) on the kit's
// input fill at 90 % at rest; checked is the kit's primary disc with its primary-foreground dot (8px, 10px in Dark),
// its border still transparent; a disabled option is undimmed with foreground text (the kit's disabled:opacity-50 and
// the Label's peer-disabled:opacity-50 match :disabled, which the Base UI <span> root never does; data-disabled carries
// the state); the circle is centred on the first line of 14px text at 1.375 line height. The visible edge of an enabled
// option against its surface (the resting fill, or the checked disc) keeps 3:1 where the kit meets it and is otherwise
// asserted at the kit's ratio as OWNER-WAIVED. The form submits exactly one value for the group's `name` (the kit's
// documented form bridge) when an option is checked. Keyboard focus shows the kit's whole focus state (>= 3:1 against
// the page) in both motion preferences; the kit radio has no transition at all. Attaches a capture of the group; with
// `narrow`, also at 320px wide, where at least one option wraps and stays aligned.
export async function expectRadioOptions(page: Page, testInfo: TestInfo, fieldset: Locator, what: string, name: string, narrow = false): Promise<void> {
  const radios = fieldset.getByRole("radio");
  const n = await radios.count();
  expect(n, `${what}: radio options`).toBeGreaterThan(0);
  await blur(page);
  await fieldset.scrollIntoViewIfNeeded();
  // A group inside a kit Table (the label list's visibility choice) sits on a row that paints a hover fill
  // (hover:bg-muted/50 with transition-colors, web/src/components/ui/table.tsx TableRow): a pointer left over the row
  // by an earlier step, or a row scrolled under it, would change the sampled surface between the samples below. Park
  // the pointer outside the content and let the row's colour transition finish before sampling.
  await page.mouse.move(0, 0);
  await frames(page);
  await fieldset.evaluate((e) => Promise.allSettled((e.closest("tr")?.getAnimations() ?? []).map((a) => a.finished)).then(() => undefined));
  let target: Locator | undefined;
  let checkedCount = 0;
  for (let i = 0; i < n; i++) {
    const radio = radios.nth(i);
    const f = await radioFacts(radio);
    const option = `${what} option ${i + 1}`;
    expect(f.role, `${option}: the kit radio`).toBe("radio");
    expect(["true", "false"], `${option}: aria-checked ${f.ariaChecked}`).toContain(f.ariaChecked);
    await expect(radio, `${option}: named by its label`).toHaveAccessibleName(f.label);
    expect([Math.round(f.width * 2) / 2, Math.round(f.height * 2) / 2], `${option}: 16px circle`).toEqual([f.rootFont, f.rootFont]);
    expect(f.radius, `${option}: round`).toBeGreaterThanOrEqual(f.width / 2);
    expect(f.border, `${option}: 1px boundary`).toBe("1px");
    expect(f.font, `${option}: 14px label text`).toBeCloseTo(f.rootFont * 0.875, 1);
    expect(f.line / f.font, `${option}: label line height`).toBeCloseTo(1.375, 2);
    expect(Math.abs(f.offset), `${option}: circle centred on the first text line (offset ${f.offset.toFixed(2)}px)`).toBeLessThanOrEqual(1);
    expect(Math.max(...f.transition.split(",").map(seconds)), `${option}: the kit radio has no transition (${f.transition})`).toBe(0);
    expect((await cssToRgba(page, f.borderColor))[3], `${option}: ${f.checked ? "checked" : "resting"} border is the kit's border-transparent (${f.borderColor})`).toBe(0);
    const b = await radio.boundingBox();
    const cx = b!.x + b!.width / 2;
    const cy = b!.y + b!.height / 2;
    // The surface the option sits on (the pixel just left of the circle).
    const surface: RGB = await pixelAt(page, b!.x - 6, cy);
    // The kit's visible edge at the option's state, as the kit paints it from the page's own tokens (the surface sampled
    // beside the circle): the primary disc when checked, the input fill at 90 % at rest.
    const kitEdge = f.checked ? rgb(await cssToRgba(page, f.tokens.primary)) : await mixOver(page, f.tokens.input, 90, surface);
    if (f.checked) {
      checkedCount++;
      await near(page, f.background, f.tokens.primary, `${option}: checked disc (the primary token)`);
      expect(f.dot, `${option}: checked dot`).not.toBeNull();
      const size = f.dark ? 10 : 8;
      expect([Math.round(f.dot!.width), Math.round(f.dot!.height)], `${option}: dot of ${size}px`).toEqual([size, size]);
      if (!f.disabled) {
        // As painted: the disc on the centre row 5px right of the centre (clear of the border's anti-aliasing and of the
        // dot), the dot at the centre (tokens.spec.ts S-FIELD-STATES samples the same points).
        // The pixel is the first whole one at or beyond 5px: a circle at a fractional x would otherwise put the sample across the
        // dot's edge (Dark's 10px dot ends exactly there). Past it the pixel carries up to a pixel's share of the border's
        // anti-aliasing (6 to 7 measured here against 138 for the dot), so the disc is matched within 10, not 2.
        const disc = await pixelAt(page, Math.ceil(cx + 5), cy);
        const dot = await pixelAt(page, cx, cy);
        expect(distance(disc, rgb(await cssToRgba(page, f.tokens.primary))), `${option}: painted disc ${hex(disc)}`).toBeLessThanOrEqual(10);
        expect(distance(dot, rgb(await cssToRgba(page, f.tokens["primary-foreground"]))), `${option}: painted dot ${hex(dot)}`).toBeLessThanOrEqual(2);
        expectNonText(`${option} checked dot against its disc`, dot, disc, 3);
      }
    } else {
      expect(f.dot, `${option}: no dot at rest`).toBeNull();
      const want = await mixOver(page, f.tokens.input, 90, surface);
      const fill = over(await cssToRgba(page, f.background), surface);
      expect(distance(fill, want), `${option}: resting fill ${hex(fill)} (input at 90 % over the surface, ${hex(want)})`).toBeLessThanOrEqual(2);
      if (!f.disabled) {
        const painted = await pixelAt(page, cx, cy);
        expect(distance(painted, want), `${option}: painted resting fill ${hex(painted)}`).toBeLessThanOrEqual(2);
      }
    }
    // The kit leaves a disabled option undimmed (its :disabled variants never match the Base UI <span> root) and its
    // label in the inherited foreground, like an enabled one.
    expect(f.opacity, `${option}: ${f.disabled ? "disabled (the kit's disabled:opacity-50 does not match the span root)" : "enabled"} circle`).toBe("1");
    await near(page, f.labelColor, f.tokens.foreground, `${option}: ${f.disabled ? "disabled " : ""}text (inherited foreground)`);
    if (!f.disabled) {
      // The visible edge, checked or not, against the surface it sits on: the centre-row pixel 5px right of the centre for
      // a checked disc (clear of the dot), the centre for a resting fill.
      const painted = await pixelAt(page, f.checked ? Math.ceil(cx + 5) : cx, cy);
      expectAtKit(`${option} ${f.checked ? "checked disc" : "resting fill"} against its surface`, painted, surface, kitEdge, surface, 3);
      if (f.checked) target = radio;
      else target ??= radio;
    }
  }
  expect(checkedCount, `${what}: at most one checked option`).toBeLessThanOrEqual(1);
  // The kit's form bridge: the enclosing form submits the group's name once when an option is checked, never otherwise.
  const submitted = await fieldset.evaluate((el, key) => {
    const form = el.closest("form");
    return form === null ? -1 : new FormData(form).getAll(key).length;
  }, name);
  expect(submitted, `${what}: submitted values for "${name}"`).toBe(checkedCount);
  await testInfo.attach(`radios-${what.replace(/[^a-z0-9]+/gi, "-").toLowerCase()}-${testInfo.project.name}.png`, {
    body: await fieldset.screenshot({ animations: "disabled" }),
    contentType: "image/png",
  });
  // Keyboard focus: Tab into the group lands on the checked option (the first enabled one when none is checked).
  expect(target, `${what}: an enabled option`).toBeDefined();
  for (const motion of ["no-preference", "reduce"] as const) {
    await blur(page);
    await page.emulateMedia({ reducedMotion: motion });
    await keyboardFocus(page, target!, `${what} (${motion})`);
    await expectKitFocusState(page, target!, `${what} radio (${motion})`);
    await expectFocusRing(page, `${what} (${motion})`, target!);
  }
  await blur(page);
  await page.emulateMedia({ reducedMotion: null });
  if (narrow) {
    const size = page.viewportSize() ?? { width: 1280, height: 800 };
    await page.setViewportSize({ width: 320, height: 640 });
    await fieldset.scrollIntoViewIfNeeded();
    expect(await expectAligned(radios, `${what} at 320px`), `${what} at 320px: a wrapped option`).toBeGreaterThan(0);
    await testInfo.attach(`radios-${what.replace(/[^a-z0-9]+/gi, "-").toLowerCase()}-320-${testInfo.project.name}.png`, {
      body: await fieldset.screenshot({ animations: "disabled" }),
      contentType: "image/png",
    });
    await page.setViewportSize(size);
  }
}
