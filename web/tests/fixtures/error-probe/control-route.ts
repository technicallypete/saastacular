// Test fixture only: the positive control of the web error-output check. The check copies this
// file to app/api/__probe/control/route.ts inside a disposable copy of web/; it is never part of
// the production tree. It throws PROBE_CANARY WITHOUT the completion wrapper, so Next prints its
// own error output, proving that the check's capture sees that output.
export const dynamic = "force-dynamic";

// A missing or unusable canary answers 422 instead of 500, so the check fails rather than
// passing vacuously.
const CANARY = /^[A-Za-z0-9_.]{8,62}$/;

export function GET(): Response {
  const canary = process.env.PROBE_CANARY ?? "";
  if (!CANARY.test(canary)) {
    return new Response(null, { status: 422, headers: { "Cache-Control": "no-store" } });
  }
  throw new Error(`probe control ${canary}`);
}
