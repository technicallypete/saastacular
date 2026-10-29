import "server-only";

import { randomBytes } from "node:crypto";

// Request-local diagnostic context and the bounded completion record for web's
// health routes. Context travels only as an explicit RequestContext argument:
// there is no module-level request state and no AsyncLocalStorage. The module
// has no relative or alias imports, so the unit tests load it directly.

export type IdSource = "inbound" | "generated";
export type TraceFlags = "00" | "01";
export type CoreProbe = "ok" | "not_ready" | "timeout" | "unreachable";
export type WebRoute = "/api/health" | "/api/ready";

// Closed error_type categories. A category is chosen with instanceof checks
// against built-in constructors and written from this code, never read from
// the thrown value.
export type ErrorCategory =
  | "AbortError"
  | "TimeoutError"
  | "DOMException"
  | "TypeError"
  | "RangeError"
  | "SyntaxError"
  | "ReferenceError"
  | "EvalError"
  | "URIError"
  | "AggregateError"
  | "Error"
  | "NonError";

export type RequestContext = {
  readonly requestId: string;
  readonly requestIdSource: IdSource;
  readonly traceId: string;
  readonly traceSource: IdSource;
  readonly spanId: string;
  // Present only when the inbound traceparent was valid.
  readonly parentSpanId?: string;
  readonly traceFlags: TraceFlags;
};

export type OutboundHeaders = {
  readonly "X-Request-Id": string;
  readonly traceparent: string;
};

// The readiness handler reports its core probe with its response.
export type ProbedResponse = {
  readonly response: Response;
  readonly coreProbe: CoreProbe;
};

export type RouteHandler = (request: Request) => Promise<Response>;

type Awaitable<T> = T | Promise<T>;
type Outcome = "completed" | "cancelled" | "failed";
type RecordValue = string | number;

const REQUEST_ID = /^[A-Za-z0-9._-]{8,64}$/;
const TRACEPARENT = /^00-([0-9a-f]{32})-([0-9a-f]{16})-([0-9a-f]{2})$/;
// Header values are byte strings, so length is the byte count.
const TRACEPARENT_MAX_BYTES = 512;
const ALL_ZERO = /^0+$/;
const METHODS: ReadonlySet<string> = new Set(["GET", "HEAD", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"]);
const CORE_PROBES: ReadonlySet<string> = new Set(["ok", "not_ready", "timeout", "unreachable"]);
const SENSITIVE_KEY = /(password|passwd|secret|token|api_?key|authorization|cookie|dsn|credential)/i;

const BUILT_IN_CATEGORIES: ReadonlyArray<readonly [abstract new (...args: never[]) => unknown, ErrorCategory]> = [
  [TypeError, "TypeError"],
  [RangeError, "RangeError"],
  [SyntaxError, "SyntaxError"],
  [ReferenceError, "ReferenceError"],
  [EvalError, "EvalError"],
  [URIError, "URIError"],
  [AggregateError, "AggregateError"],
];

// The built-in accessor reads the name a DOMException was constructed with; own
// properties and subclass overrides are never consulted.
const DOM_EXCEPTION_NAME = Object.getOwnPropertyDescriptor(DOMException.prototype, "name")?.get;

function randomHex(bytes: number): string {
  return randomBytes(bytes).toString("hex");
}

function freshId(bytes: number, avoid?: string): string {
  let value: string;
  do {
    value = randomHex(bytes);
  } while (ALL_ZERO.test(value) || value === avoid);
  return value;
}

type InboundTrace = { traceId: string; parentSpanId: string; sampled: boolean };

function parseTraceparent(value: string | null): InboundTrace | undefined {
  if (value === null || value.length > TRACEPARENT_MAX_BYTES) return undefined;
  const match = TRACEPARENT.exec(value);
  if (match === null) return undefined;
  const [, traceId, parentSpanId, flags] = match;
  if (ALL_ZERO.test(traceId) || ALL_ZERO.test(parentSpanId)) return undefined;
  return { traceId, parentSpanId, sampled: (Number.parseInt(flags, 16) & 0x01) === 0x01 };
}

// Accepts exactly one valid X-Request-Id and one valid traceparent. Duplicated
// fields arrive comma-joined and fail the patterns. A rejected value is never
// kept: a fresh ID or trace replaces it.
export function deriveContext(headers: Headers): RequestContext {
  const inboundId = headers.get("x-request-id");
  const acceptedId = inboundId !== null && REQUEST_ID.test(inboundId) ? inboundId : undefined;
  const inbound = parseTraceparent(headers.get("traceparent"));
  const requestId = acceptedId ?? randomHex(16);
  const requestIdSource: IdSource = acceptedId === undefined ? "generated" : "inbound";
  if (inbound === undefined) {
    return {
      requestId,
      requestIdSource,
      traceId: freshId(16),
      traceSource: "generated",
      spanId: freshId(8),
      traceFlags: "00",
    };
  }
  return {
    requestId,
    requestIdSource,
    traceId: inbound.traceId,
    traceSource: "inbound",
    spanId: freshId(8, inbound.parentSpanId),
    parentSpanId: inbound.parentSpanId,
    traceFlags: inbound.sampled ? "01" : "00",
  };
}

// The only headers application code sets on an outbound call.
export function outboundHeaders(ctx: RequestContext): OutboundHeaders {
  return {
    "X-Request-Id": ctx.requestId,
    traceparent: `00-${ctx.traceId}-${ctx.spanId}-${ctx.traceFlags}`,
  };
}

function domExceptionCategory(value: DOMException): ErrorCategory {
  try {
    const name: unknown = DOM_EXCEPTION_NAME === undefined ? undefined : Reflect.apply(DOM_EXCEPTION_NAME, value, []);
    if (name === "AbortError") return "AbortError";
    if (name === "TimeoutError") return "TimeoutError";
  } catch {
    // Not a constructed DOMException; it keeps the generic category.
  }
  return "DOMException";
}

// Maps a thrown value to its closed category. The value's name, message, cause,
// stack and own properties are never read.
export function errorCategory(thrown: unknown): ErrorCategory {
  try {
    if (thrown instanceof DOMException) return domExceptionCategory(thrown);
    for (const [builtIn, category] of BUILT_IN_CATEGORIES) {
      if (thrown instanceof builtIn) return category;
    }
    return thrown instanceof Error ? "Error" : "NonError";
  } catch {
    // A value whose prototype cannot be inspected is not a verifiable Error.
    return "NonError";
  }
}

function asResponse(result: Response | ProbedResponse): Response {
  if (result instanceof Response) return result;
  throw new TypeError("route handler did not return a Response");
}

function asProbedResponse(result: Response | ProbedResponse): ProbedResponse {
  if (!(result instanceof Response) && result.response instanceof Response && CORE_PROBES.has(result.coreProbe)) {
    return result;
  }
  throw new TypeError("readiness handler did not return a response and core probe");
}

function writeRecord(entries: ReadonlyArray<readonly [string, RecordValue]>): void {
  const record = Object.fromEntries(entries.filter(([key]) => !SENSITIVE_KEY.test(key)));
  try {
    process.stdout.write(`${JSON.stringify(record)}\n`);
  } catch {
    // Logging never changes the response.
  }
}

export function withCompletionLog(
  route: "/api/health",
  handler: (request: Request, ctx: RequestContext) => Awaitable<Response>,
): RouteHandler;
export function withCompletionLog(
  route: "/api/ready",
  handler: (request: Request, ctx: RequestContext) => Awaitable<ProbedResponse>,
): RouteHandler;
// Runs the handler with a fresh context, adds X-Request-Id to its response and
// writes exactly one completion record. A thrown value is never rethrown: it
// becomes an empty 500 with Cache-Control: no-store, so Next prints nothing.
export function withCompletionLog(
  route: WebRoute,
  handler: (request: Request, ctx: RequestContext) => Awaitable<Response | ProbedResponse>,
): RouteHandler {
  return async (request: Request): Promise<Response> => {
    const started = performance.now();
    const ctx = deriveContext(request.headers);
    let response: Response;
    let status = 0;
    let outcome: Outcome = "completed";
    let errorType: ErrorCategory | undefined;
    let coreProbe: CoreProbe | undefined;
    try {
      const result = await handler(request, ctx);
      if (route === "/api/ready") {
        ({ response, coreProbe } = asProbedResponse(result));
      } else {
        response = asResponse(result);
      }
      response.headers.set("X-Request-Id", ctx.requestId);
      status = response.status;
    } catch (thrown) {
      outcome = "failed";
      errorType = errorCategory(thrown);
      coreProbe = undefined;
      response = new Response(null, {
        status: 500,
        headers: { "Cache-Control": "no-store", "X-Request-Id": ctx.requestId },
      });
      status = 500;
    } finally {
      // The client left before any header was sent.
      if (request.signal.aborted) {
        outcome = "cancelled";
        errorType = undefined;
        status = 0;
      }
      const entries: Array<readonly [string, RecordValue]> = [
        ["time", new Date().toISOString()],
        ["level", outcome === "failed" ? "error" : status >= 500 ? "warn" : "info"],
        ["service", "web"],
        ["event", "http.request.completed"],
        ["method", METHODS.has(request.method) ? request.method : "OTHER"],
        ["route", route],
        ["status", status],
        ["duration_ms", Math.max(0, Math.floor(performance.now() - started))],
        ["request_id", ctx.requestId],
        ["request_id_source", ctx.requestIdSource],
        ["trace_id", ctx.traceId],
        ["trace_source", ctx.traceSource],
        ["span_id", ctx.spanId],
      ];
      if (ctx.parentSpanId !== undefined) entries.push(["parent_span_id", ctx.parentSpanId]);
      entries.push(["trace_flags", ctx.traceFlags], ["outcome", outcome]);
      if (outcome === "failed" && errorType !== undefined) entries.push(["error_type", errorType]);
      if (route === "/api/ready" && coreProbe !== undefined) entries.push(["core_probe", coreProbe]);
      writeRecord(entries);
    }
    return response;
  };
}
