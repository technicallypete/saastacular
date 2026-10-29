import type { Page } from "@playwright/test";

import { dialogSettled } from "./support/contrast";
import { expect, test } from "./support/page-errors";
import { waitForHydration } from "./support/modes";

async function focusTrigger(page: Page) {
  const trigger = page.getByRole("button", { name: "Keyboard help" });
  await trigger.focus();
  await expect(trigger).toBeFocused();
  return trigger;
}

async function expectContained(page: Page) {
  const dialog = page.getByRole("dialog");
  for (let i = 0; i < 10; i++) {
    await page.keyboard.press(i % 3 === 2 ? "Shift+Tab" : "Tab");
    // Base UI traps Tab inside the popup with focus guards beside it (data-base-ui-focus-guard, set by
    // @base-ui/react 1.9.0 utils/FocusGuard): focus may rest on a guard until the next animation frame,
    // which moves it back inside. Nothing else is allowed.
    const transient = await dialog.evaluate((d) => {
      const a = document.activeElement;
      return d.contains(a) || a?.hasAttribute("data-base-ui-focus-guard") === true;
    });
    expect(transient, `focus moved outside the dialog and its guards after ${i + 1} key presses`).toBe(true);
    await page.evaluate(() => new Promise<void>((r) => requestAnimationFrame(() => requestAnimationFrame(() => r()))));
    const inside = await dialog.evaluate((d) => d.contains(document.activeElement));
    expect(inside, `focus left the dialog after ${i + 1} key presses`).toBe(true);
  }
}

test("[S-OVERLAY] keyboard dialog: name, focus, containment, Escape, Close, focus return", async ({ page }) => {
  await page.goto("/");
  await waitForHydration(page);
  const dialog = page.getByRole("dialog");

  for (const openKey of ["Enter", "Space"]) {
    const trigger = await focusTrigger(page);
    await page.keyboard.press(openKey);
    await expect(dialog).toBeVisible();
    await expect(dialog).toHaveAccessibleName("Keyboard help");
    await expect(dialog).toHaveAttribute("aria-modal", "true");
    // Base UI moves initial focus into the popup once it has mounted.
    await expect.poll(() => dialog.evaluate((d) => d.contains(document.activeElement)), { message: "focus moved into the dialog" }).toBe(true);
    await expectContained(page);
    // Everything outside the dialog is inert or hidden from assistive technology while it is open.
    for (const region of ["header", "main#main", "footer"]) {
      await expect
        .poll(() => page.locator(region).evaluate((el) => !!el.closest("[inert], [aria-hidden='true']")), { message: `${region} is inert while the dialog is open` })
        .toBe(true);
    }
    // The Close control meets the 24px target size (WCAG 2.5.8); G067: it is the kit's icon-sm button, 28 x 28px,
    // measured once the popup's zoom has settled. G069: the kit DialogContent's own corner Close, named by its sr-only text.
    await dialogSettled(page);
    const closeBox = await dialog.getByRole("button", { name: "Close" }).boundingBox();
    expect(closeBox!.width, "Close target width").toBeGreaterThanOrEqual(24);
    expect(closeBox!.height, "Close target height").toBeGreaterThanOrEqual(24);
    expect(Math.max(Math.abs(closeBox!.width - 28), Math.abs(closeBox!.height - 28)), `Close is the 28px icon-sm control: ${closeBox!.width} x ${closeBox!.height}`).toBeLessThanOrEqual(0.5);
    if (openKey === "Enter") {
      await page.keyboard.press("Escape");
    } else {
      const close = dialog.getByRole("button", { name: "Close" });
      await close.focus();
      await page.keyboard.press("Enter");
    }
    await expect(dialog).toBeHidden();
    await expect(trigger).toBeFocused();
  }
});

async function popupTiming(page: Page) {
  await page.getByRole("button", { name: "Keyboard help" }).click();
  const popup = page.locator("[data-slot=dialog-content]");
  await expect(popup).toBeVisible();
  await expect(page.locator("[data-slot=dialog-overlay]")).toHaveCount(1);
  const timing = await popup.evaluate((el) => {
    const s = getComputedStyle(el);
    const b = getComputedStyle(document.querySelector("[data-slot=dialog-overlay]")!);
    const trigger = getComputedStyle(document.querySelector('header button[aria-haspopup="dialog"]')!);
    const seconds = (v: string) => Math.max(...v.split(",").map((x) => (x.trim().endsWith("ms") ? parseFloat(x) / 1000 : parseFloat(x))));
    return {
      popupAnimation: seconds(s.animationDuration),
      popupTransition: seconds(s.transitionDuration),
      backdropAnimation: seconds(b.animationDuration),
      buttonTransition: seconds(trigger.transitionDuration),
      name: s.animationName,
      backdropName: b.animationName,
    };
  });
  // Once the open animations have finished, watch the popup and backdrop close after Escape. When Base UI marks a part
  // closed (data-closed) its computed animation name and duration are recorded (the closed state, in both motion
  // modes) together with the CSS animation then attached to the part (getAnimations(), which flushes style); an
  // animationstart on a part not yet recorded as running does the same. Resolves when both parts have both records, or with whatever was recorded after 1.5s (a safety bound:
  // the kit's 100ms exit animations attach in both motion modes).
  type Exit = { name: string; duration: number; started?: string; startedClosed?: boolean };
  type Closing = Partial<Record<"dialog-content" | "dialog-overlay", Exit>>;
  await dialogSettled(page);
  await page.evaluate(() => {
    const seconds = (v: string) => Math.max(...v.split(",").map((x) => (x.trim().endsWith("ms") ? parseFloat(x) / 1000 : parseFloat(x))));
    const parts = ["dialog-content", "dialog-overlay"] as const;
    const slotOf = (el: Element) => parts.find((p) => el.getAttribute("data-slot") === p);
    const closing: Closing = {};
    (window as unknown as { closing: Promise<Closing> }).closing = new Promise((resolve) => {
      const check = () => parts.every((p) => closing[p]?.started !== undefined) && done();
      const observer = new MutationObserver((records) => {
        for (const r of records) {
          const el = r.target as Element;
          const slot = slotOf(el);
          if (!slot || !el.hasAttribute("data-closed") || closing[slot]) continue;
          const s = getComputedStyle(el);
          const exit: Exit = { name: s.animationName, duration: seconds(s.animationDuration) };
          closing[slot] = exit;
          // getAnimations() flushes style, so the exit CSSAnimation exists here. animationstart alone is not a safe proof:
          // when a frame is late (a starved runner) Base UI can unmount the part before the browser dispatches it (rc.22).
          if (el.getAnimations().some((a) => a instanceof CSSAnimation && a.animationName === exit.name)) {
            exit.started = exit.name;
            exit.startedClosed = true; // this path runs only for a part already marked data-closed: attachment at close time
          }
        }
        check();
      });
      const onStart = (e: AnimationEvent) => {
        const el = e.target as Element;
        const slot = slotOf(el);
        if (!slot || closing[slot]?.started !== undefined) return;
        const s = getComputedStyle(el);
        closing[slot] = { ...(closing[slot] ?? { name: s.animationName, duration: seconds(s.animationDuration) }), started: e.animationName, startedClosed: el.hasAttribute("data-closed") };
        check();
      };
      const timer = setTimeout(() => done(), 1500);
      const done = () => {
        clearTimeout(timer);
        observer.disconnect();
        document.removeEventListener("animationstart", onStart, true);
        resolve(closing);
      };
      for (const p of parts) observer.observe(document.querySelector(`[data-slot=${p}]`)!, { attributes: true, attributeFilter: ["data-closed"] });
      document.addEventListener("animationstart", onStart, true);
    });
  });
  await page.keyboard.press("Escape");
  const closing = await page.evaluate(() => (window as unknown as { closing: Promise<Closing> }).closing);
  await expect(popup).toBeHidden();
  return { ...timing, closePopup: closing["dialog-content"], closeBackdrop: closing["dialog-overlay"] };
}

type Timing = Awaited<ReturnType<typeof popupTiming>>;

function expectBriefMotion(t: Timing, mode: string) {
  expect(t.name, `popup open animation (${mode})`).not.toBe("none");
  expect(t.backdropName, `backdrop open animation (${mode})`).not.toBe("none");
  expect(t.popupAnimation, `popup animation duration (${mode})`).toBeGreaterThan(0);
  expect(t.popupAnimation, `popup animation duration (${mode})`).toBeLessThanOrEqual(0.15);
  expect(t.backdropAnimation, `backdrop animation duration (${mode})`).toBeGreaterThan(0);
  expect(t.backdropAnimation, `backdrop animation duration (${mode})`).toBeLessThanOrEqual(0.15);
  expect(t.buttonTransition, `button transitions exist (${mode}, control)`).toBeGreaterThan(0.05);
  for (const [part, exit, open] of [
    ["popup", t.closePopup, t.name],
    ["backdrop", t.closeBackdrop, t.backdropName],
  ] as const) {
    expect(exit, `${part} reached its closed state after Escape (${mode})`).toBeDefined();
    expect(exit!.name, `${part} close animation (${mode})`).not.toBe("none");
    expect(exit!.name, `${part} close animation differs from its open animation (${mode})`).not.toBe(open);
    expect(exit!.duration, `${part} close animation duration (${mode})`).toBeGreaterThan(0);
    expect(exit!.duration, `${part} close animation duration (${mode})`).toBeLessThanOrEqual(0.15);
    expect(exit!.started, `${part} close animation ran (${mode})`).toBe(exit!.name);
    expect(exit!.startedClosed, `${part} was closed (data-closed) when its close animation was attached or started (${mode})`).toBe(true);
  }
}

// The dialog keeps the kit's subtle open/close motion (G067 motion delta, owner preview feedback 2026-10-25,
// superseding D-6's no-motion): the Base UI Backdrop fades and the Popup fades and zooms, each at most 150ms. Closing is
// asserted, not inferred: after Escape both parts take a distinct exit animation in their closed state, which is
// attached to the part (a CSS animation of that name) for at most 150ms before Base UI unmounts the popup. G069 (owner kit-fidelity direction 2026-10-26):
// the app's global reduced-motion rule is withdrawn, and the kit's dialog and button motion has no reduced-motion
// variant (its only such rule stops the ribbon and [data-motion] layers), so under reduced motion the same motion runs
// unchanged.
test("[S-MOTION] the dialog animates briefly on open and close; the kit's dialog and button motion is unchanged under reduced motion", async ({ page }) => {
  await page.goto("/");
  await waitForHydration(page);
  await page.emulateMedia({ reducedMotion: "no-preference" });
  const normal = await popupTiming(page);
  expectBriefMotion(normal, "no preference");
  await page.emulateMedia({ reducedMotion: "reduce" });
  const reduced = await popupTiming(page);
  expectBriefMotion(reduced, "reduced motion");
  expect(reduced.popupAnimation, "popup animation duration is the kit's own under reduced motion").toBe(normal.popupAnimation);
  expect(reduced.popupTransition, "popup transition duration is the kit's own under reduced motion").toBe(normal.popupTransition);
  expect(reduced.backdropAnimation, "backdrop animation duration is the kit's own under reduced motion").toBe(normal.backdropAnimation);
  expect(reduced.buttonTransition, "button transition duration is the kit's own under reduced motion").toBe(normal.buttonTransition);
  expect(reduced.closePopup!.duration, "popup close animation duration is the kit's own under reduced motion").toBe(normal.closePopup!.duration);
  expect(reduced.closeBackdrop!.duration, "backdrop close animation duration is the kit's own under reduced motion").toBe(normal.closeBackdrop!.duration);
});
