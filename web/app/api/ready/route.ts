import { coreReady } from "@/lib/core-client";
import { withCompletionLog } from "@/lib/request-context";

// Readiness: web can serve only when core is ready. Reports state by name,
// never error detail.
export const dynamic = "force-dynamic";

export const GET = withCompletionLog("/api/ready", async (_request, ctx) => {
  const coreProbe = await coreReady(ctx);
  const ready = coreProbe === "ok";
  return {
    coreProbe,
    response: Response.json(
      { status: ready ? "ready" : "unready", checks: { core: ready ? "ok" : "unavailable" } },
      { status: ready ? 200 : 503, headers: { "Cache-Control": "no-store" } },
    ),
  };
});
