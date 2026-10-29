"use client";

import * as React from "react";

// G018 M1 (D2, Option B): the advisory address hint of the New Group form. It wraps the server-rendered slug field
// (children) and, when the field loses focus with a value it has not answered yet, asks the same-origin
// GET /groups/address-check once and shows the server-supplied copy for the closed answer in a polite status line that
// the field's aria-describedby already references. A hint is advisory only: it never marks the field invalid, never
// disables or replaces submit and never replaces a server error; the create decides. Every edit clears the hint and
// invalidates any request in flight; an answer is applied only when its generation is still current, the field still
// holds exactly the value asked about and the answer echoes that exact value. One request at a time with a finite
// lifetime; a failure, refusal, timeout or anything outside the closed shape just clears the hint (no retry, no
// polling, no writes). Without script the field is the plain input and the form submits natively. This module imports
// only React: the copy and the bound come from the server page as props.

export type AdvisoryState = "in_use" | "not_seen" | "invalid";
export const ADDRESS_CHECK_PATH = "/groups/address-check";
export const ADDRESS_CHECK_TIMEOUT_MS = 5_000;
// The largest acceptable answer: {"state":"not_seen","slug":"<63 bytes>"} is 93 bytes.
export const ADDRESS_CHECK_MAX_BYTES = 256;
const STATES: ReadonlySet<string> = new Set(["in_use", "not_seen", "invalid"]);

type Fetcher = (input: string, init: RequestInit) => Promise<Response>;
type Options = { maxSlug: number; current: () => string; fetcher?: Fetcher; timeoutMs?: number };

// The response body as text when it is at most max bytes of valid UTF-8; undefined otherwise (reading stops early).
async function readBounded(res: Response, max: number): Promise<string | undefined> {
  const reader = res.body?.getReader();
  if (reader === undefined) return undefined;
  const chunks: Uint8Array[] = [];
  let size = 0;
  for (;;) {
    const { done, value } = await reader.read();
    if (done) break;
    size += value.byteLength;
    if (size > max) {
      void reader.cancel().catch(() => undefined);
      return undefined;
    }
    chunks.push(value);
  }
  const bytes = new Uint8Array(size);
  let at = 0;
  for (const c of chunks) {
    bytes.set(c, at);
    at += c.byteLength;
  }
  try {
    return new TextDecoder("utf-8", { fatal: true }).decode(bytes);
  } catch {
    return undefined;
  }
}

// The closed state of an answer for exactly VALUE, or undefined for anything else.
export async function readAdvisory(res: Response, value: string): Promise<AdvisoryState | undefined> {
  if (res.status !== 200 || !(res.headers.get("content-type") ?? "").toLowerCase().startsWith("application/json")) return undefined;
  const text = await readBounded(res, ADDRESS_CHECK_MAX_BYTES);
  if (text === undefined) return undefined;
  let body: unknown;
  try {
    body = JSON.parse(text);
  } catch {
    return undefined;
  }
  if (typeof body !== "object" || body === null || Array.isArray(body)) return undefined;
  const keys = Object.keys(body).sort();
  if (keys.length !== 2 || keys[0] !== "slug" || keys[1] !== "state") return undefined;
  const { state, slug } = body as { state: unknown; slug: unknown };
  return typeof state === "string" && STATES.has(state) && slug === value ? (state as AdvisoryState) : undefined;
}

// The hint logic without React (unit-tested): show(state) renders a hint, show(undefined) clears it.
export function addressChecker(show: (state: AdvisoryState | undefined) => void, options: Options) {
  const fetcher: Fetcher = options.fetcher ?? ((input, init) => fetch(input, init));
  const timeoutMs = options.timeoutMs ?? ADDRESS_CHECK_TIMEOUT_MS;
  let generation = 0;
  let answered: string | undefined;
  let active: { controller: AbortController; timer: ReturnType<typeof setTimeout> } | undefined;
  const cancel = () => {
    if (active === undefined) return;
    clearTimeout(active.timer);
    active.controller.abort();
    active = undefined;
  };
  const invalidate = () => {
    generation += 1;
    cancel();
    answered = undefined;
  };
  return {
    // Every edit: clear the hint and invalidate the request in flight.
    edited() {
      invalidate();
      show(undefined);
    },
    // Unmount: invalidate without rendering.
    dispose: invalidate,
    // Blur: ask about exactly VALUE (never a trimmed or otherwise different value) unless it is already answered.
    async check(value: string): Promise<void> {
      if (answered !== undefined && value === answered) return;
      invalidate();
      show(undefined);
      if (value === "") return;
      if (value.length > options.maxSlug) {
        // Longer than any valid address: invalid without asking (a valid slug is ASCII, one code unit per byte).
        answered = value;
        show("invalid");
        return;
      }
      const mine = generation;
      const controller = new AbortController();
      const own = { controller, timer: setTimeout(() => controller.abort(), timeoutMs) };
      active = own;
      let state: AdvisoryState | undefined;
      try {
        const res = await fetcher(`${ADDRESS_CHECK_PATH}?slug=${encodeURIComponent(value)}`, {
          method: "GET",
          cache: "no-store",
          credentials: "same-origin",
          redirect: "error",
          headers: { Accept: "application/json" },
          signal: controller.signal,
        });
        state = await readAdvisory(res, value);
      } catch {
        state = undefined;
      } finally {
        if (active === own) {
          clearTimeout(own.timer);
          active = undefined;
        }
      }
      // A stale answer (an edit or a newer check since) or one for a value the field no longer holds is dropped.
      if (mine !== generation || options.current() !== value || controller.signal.aborted) return;
      if (state !== undefined) answered = value;
      show(state);
    },
  };
}

type AddressCheckProps = {
  // The server page's closed copy for each state and the slug bound (this module imports nothing server-side).
  copy: Readonly<Record<AdvisoryState, string>>;
  maxSlug: number;
  // The id of the hint line; the server-rendered slug field lists it in its aria-describedby.
  advisoryId: string;
  children: React.ReactNode;
};

export function AddressCheck({ copy, maxSlug, advisoryId, children }: AddressCheckProps) {
  const ref = React.useRef<HTMLDivElement>(null);
  const [state, setState] = React.useState<AdvisoryState | undefined>(undefined);

  React.useEffect(() => {
    const input = ref.current?.querySelector<HTMLInputElement>('input[name="slug"]');
    if (!input) return;
    const checker = addressChecker(setState, { maxSlug, current: () => input.value });
    const onInput = () => checker.edited();
    const onBlur = () => void checker.check(input.value);
    input.addEventListener("input", onInput);
    input.addEventListener("blur", onBlur);
    return () => {
      input.removeEventListener("input", onInput);
      input.removeEventListener("blur", onBlur);
      checker.dispose();
    };
  }, [maxSlug]);

  return (
    <div ref={ref} className="grid min-w-0 gap-2">
      {children}
      <p id={advisoryId} role="status" data-testid="address-advisory" className="m-0 font-sans text-sm text-foreground/80">
        {state === undefined ? "" : copy[state]}
      </p>
    </div>
  );
}
