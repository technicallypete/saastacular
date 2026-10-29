import { withCompletionLog } from "@/lib/request-context";

// Liveness: the web process is serving. No dependencies.
export const dynamic = "force-dynamic";

export const GET = withCompletionLog("/api/health", () =>
  Response.json({ status: "alive" }, { headers: { "Cache-Control": "no-store" } }),
);
