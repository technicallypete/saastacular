import "server-only";

import { runtimeConfig } from "./config";
import { errorCategory, outboundHeaders, type CoreProbe, type RequestContext } from "./request-context";

// Server-only bridge to core. The browser never calls core or intelligence
// directly; this thin BFF forwards to core over the internal edge network.
// Authority, policy and domain rules live in core, never here. The only
// headers sent are the request's diagnostic context (X-Request-Id and
// traceparent); the result is a probe state by name, never error detail.
export async function coreReady(ctx: RequestContext, timeoutMs = 2000): Promise<CoreProbe> {
  const { coreInternalUrl } = runtimeConfig();
  try {
    const response = await fetch(new URL("/readyz", coreInternalUrl), {
      cache: "no-store",
      headers: outboundHeaders(ctx),
      signal: AbortSignal.timeout(timeoutMs),
    });
    return response.ok ? "ok" : "not_ready";
  } catch (error) {
    return errorCategory(error) === "TimeoutError" ? "timeout" : "unreachable";
  }
}
