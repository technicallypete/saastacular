import { withCompletionLog } from "@/lib/request-context";

// Test fixture only: the web error-output check copies this file to
// app/api/__probe/wrapped/route.ts inside a disposable copy of web/; it is never part of the
// production tree. Every mode runs through the completion wrapper, and every thrown value carries
// PROBE_CANARY so the check can prove the canary never reaches the output. The record uses the
// liveness route label: the completion contract admits only the two production route labels.
//   ?mode=ok              200
//   ?mode=throw           Error with the canary in its message, cause and an own property
//   ?mode=nonerror        a thrown non-Error value carrying the canary
//   ?mode=badname         Error whose name is invalid and carries the canary
//   ?mode=badname-valid   Error whose name is a valid identifier carrying the canary (Cx<canary>)
//   ?mode=slow            200 after 3 s (for a client abort)
export const dynamic = "force-dynamic";

// A missing or unusable canary answers 422, so the check fails rather than passing vacuously.
// "Cx" plus 8 to 62 characters is always a valid error_type-shaped name.
const CANARY = /^[A-Za-z0-9_.]{8,62}$/;
const NO_STORE = { "Cache-Control": "no-store" };

function canaryError(canary: string): Error {
  const error = new Error(`probe failure ${canary}`, { cause: new Error(`probe cause ${canary}`) });
  return Object.assign(error, { probeDetail: canary });
}

export const GET = withCompletionLog("/api/health", async (request) => {
  const canary = process.env.PROBE_CANARY ?? "";
  if (!CANARY.test(canary)) return new Response(null, { status: 422, headers: NO_STORE });
  switch (new URL(request.url).searchParams.get("mode")) {
    case "ok":
      return Response.json({ status: "ok" }, { headers: NO_STORE });
    case "throw":
      throw canaryError(canary);
    case "nonerror":
      throw { name: `Probe${canary}`, message: `probe failure ${canary}`, probeDetail: canary };
    case "badname": {
      const error = canaryError(canary);
      error.name = `bad name ${canary}`;
      throw error;
    }
    case "badname-valid": {
      const error = canaryError(canary);
      error.name = `Cx${canary}`;
      throw error;
    }
    case "slow":
      await new Promise((resolve) => setTimeout(resolve, 3000));
      return Response.json({ status: "ok" }, { headers: NO_STORE });
    default:
      return new Response(null, { status: 400, headers: NO_STORE });
  }
});
