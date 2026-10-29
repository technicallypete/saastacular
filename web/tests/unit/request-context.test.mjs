// Unit tests for lib/request-context.ts: context validation and generation, the outbound header
// allow-list, the completion record, failure and cancellation handling, error categories and
// isolation under concurrency. Run by `npm run test:unit` in the web check image with networking
// disabled: Node strips the module's types, and the react-server condition resolves `server-only`
// to its empty entry.
import assert from "node:assert/strict";
import crypto from "node:crypto";
import { syncBuiltinESMExports } from "node:module";
import { describe, it, mock } from "node:test";

import { deriveContext, errorCategory, outboundHeaders, withCompletionLog } from "../../lib/request-context.ts";

const URL_HEALTH = "http://127.0.0.1:3000/api/health";
const URL_READY = "http://127.0.0.1:3000/api/ready";
const RECORD_PREFIX = '{"time":';
const HEX32 = /^[0-9a-f]{32}$/;
const HEX16 = /^[0-9a-f]{16}$/;
const ZERO32 = "0".repeat(32);
const ZERO16 = "0".repeat(16);
const TIME = /^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{3}Z$/;
const ERROR_TYPE = /^[A-Za-z][A-Za-z0-9_.]{0,63}$/;
const SENSITIVE_KEY = /(password|passwd|secret|token|api_?key|authorization|cookie|dsn|credential)/i;

const BASE_KEYS = [
  "time",
  "level",
  "service",
  "event",
  "method",
  "route",
  "status",
  "duration_ms",
  "request_id",
  "request_id_source",
  "trace_id",
  "trace_source",
  "span_id",
];

const hex = (bytes) => crypto.randomBytes(bytes).toString("hex");
const canary = () => `Cn${hex(16)}`;

function headersOf(entries) {
  const headers = new Headers();
  for (const [name, value] of entries) headers.append(name, value);
  return headers;
}

function request(url, entries = [], init = {}) {
  return new Request(url, { ...init, headers: headersOf(entries) });
}

// Captures every string written to stdout and stderr while `run` executes. Completion records are
// kept from the real stdout; any other write passes through unchanged (the test runner reports on
// stdout). The originals are restored afterwards, so nothing leaks into later tests.
async function captured(run) {
  const writes = [];
  const raw = [];
  const stdout = process.stdout.write;
  const stderr = process.stderr.write;
  const out = mock.method(process.stdout, "write", function (chunk, ...rest) {
    if (typeof chunk === "string") {
      writes.push(chunk);
      if (chunk.startsWith(RECORD_PREFIX)) {
        raw.push(chunk);
        return true;
      }
    }
    return stdout.call(this, chunk, ...rest);
  });
  const err = mock.method(process.stderr, "write", function (chunk, ...rest) {
    if (typeof chunk === "string") writes.push(chunk);
    return stderr.call(this, chunk, ...rest);
  });
  try {
    const result = await run();
    return { result, raw, writes, records: raw.map((line) => JSON.parse(line)) };
  } finally {
    out.mock.restore();
    err.mock.restore();
  }
}

function assertOneLine(raw) {
  for (const line of raw) {
    assert.ok(line.endsWith("\n"), "record ends with a newline");
    assert.equal(line.indexOf("\n"), line.length - 1, "record is a single line");
  }
}

function assertAbsent(value, haystacks) {
  for (const text of haystacks) assert.ok(!text.includes(value), "value must not appear in the output");
}

function responseHeaderText(response) {
  return [...response.headers].map(([name, value]) => `${name}: ${value}`).join("\n");
}

const alive = () => Response.json({ status: "alive" }, { headers: { "Cache-Control": "no-store" } });

describe("deriveContext: X-Request-Id", () => {
  const valid = [
    ["8 characters", "abcd1234"],
    ["typical", "req-12345678"],
    ["dots and underscores", "a.b_c-d.e_f-0123"],
    ["64 characters", "A".repeat(64)],
  ];
  for (const [label, value] of valid) {
    it(`accepts ${label}`, () => {
      const ctx = deriveContext(headersOf([["X-Request-Id", value]]));
      assert.equal(ctx.requestId, value);
      assert.equal(ctx.requestIdSource, "inbound");
    });
  }

  const invalid = [
    ["absent", []],
    ["empty", [["X-Request-Id", ""]]],
    ["7 characters", [["X-Request-Id", "abc1234"]]],
    ["65 characters", [["X-Request-Id", "A".repeat(65)]]],
    ["8 KiB", [["X-Request-Id", "A".repeat(8192)]]],
    ["internal space", [["X-Request-Id", "req 12345678"]]],
    ["slash", [["X-Request-Id", "req/12345678"]]],
    ["colon", [["X-Request-Id", "req:12345678"]]],
    ["non-ASCII", [["X-Request-Id", "réq-12345678"]]],
    ["comma-joined", [["X-Request-Id", "abcdefgh1,abcdefgh2"]]],
    ["comma-space-joined", [["X-Request-Id", "abcdefgh1, abcdefgh2"]]],
    ["duplicated field", [["X-Request-Id", "abcdefgh1"], ["X-Request-Id", "abcdefgh2"]]],
  ];
  for (const [label, entries] of invalid) {
    it(`rejects ${label} and generates 32 hex`, () => {
      const ctx = deriveContext(headersOf(entries));
      assert.match(ctx.requestId, HEX32);
      assert.equal(ctx.requestIdSource, "generated");
      for (const [, value] of entries) if (value !== "") assert.ok(!ctx.requestId.includes(value));
    });
  }
});

describe("deriveContext: traceparent", () => {
  const trace = "4bf92f3577b34da6a3ce929d0e0e4736";
  const parent = "00f067aa0ba902b7";

  const accepted = [
    ["sampled", "01", "01"],
    ["not sampled", "00", "00"],
    ["unknown bits with sampled", "03", "01"],
    ["unknown bits without sampled", "02", "00"],
    ["all bits", "ff", "01"],
  ];
  for (const [label, flags, expected] of accepted) {
    it(`accepts a valid value (${label})`, () => {
      const ctx = deriveContext(headersOf([["traceparent", `00-${trace}-${parent}-${flags}`]]));
      assert.equal(ctx.traceId, trace);
      assert.equal(ctx.traceSource, "inbound");
      assert.equal(ctx.parentSpanId, parent);
      assert.equal(ctx.traceFlags, expected);
      assert.match(ctx.spanId, HEX16);
      assert.notEqual(ctx.spanId, ZERO16);
      assert.notEqual(ctx.spanId, parent);
    });
  }

  const rejected = [
    ["absent", []],
    ["all-zero trace", [["traceparent", `00-${ZERO32}-${parent}-01`]]],
    ["all-zero parent", [["traceparent", `00-${trace}-${ZERO16}-01`]]],
    ["uppercase", [["traceparent", `00-${trace.toUpperCase()}-${parent}-01`]]],
    ["version 01", [["traceparent", `01-${trace}-${parent}-01`]]],
    ["version ff", [["traceparent", `ff-${trace}-${parent}-01`]]],
    ["extra field", [["traceparent", `00-${trace}-${parent}-01-extra`]]],
    ["short trace", [["traceparent", `00-${trace.slice(1)}-${parent}-01`]]],
    ["long parent", [["traceparent", `00-${trace}-${parent}0-01`]]],
    ["non-hex", [["traceparent", `00-${trace.slice(0, 31)}g-${parent}-01`]]],
    ["internal whitespace", [["traceparent", `00-${trace} -${parent}-01`]]],
    ["comma-joined", [["traceparent", `00-${trace}-${parent}-01,00-${trace}-${parent}-01`]]],
    ["duplicated field", [["traceparent", `00-${trace}-${parent}-01`], ["traceparent", `00-${trace}-${parent}-01`]]],
    ["513 bytes", [["traceparent", `00-${trace}-${parent}-01`.padEnd(513, "0")]]],
    ["8 KiB", [["traceparent", `00-${trace}-${parent}-01`.padEnd(8192, "0")]]],
  ];
  for (const [label, entries] of rejected) {
    it(`replaces ${label} with a new local trace`, () => {
      const ctx = deriveContext(headersOf(entries));
      assert.equal(ctx.traceSource, "generated");
      assert.match(ctx.traceId, HEX32);
      assert.notEqual(ctx.traceId, ZERO32);
      assert.notEqual(ctx.traceId, trace, "an invalid inbound trace id is never kept");
      assert.equal(ctx.parentSpanId, undefined);
      assert.equal("parentSpanId" in ctx, false);
      assert.equal(ctx.traceFlags, "00");
      assert.match(ctx.spanId, HEX16);
      assert.notEqual(ctx.spanId, parent);
    });
  }

  it("ignores tracestate and baggage", () => {
    const ctx = deriveContext(headersOf([["tracestate", "k=v"], ["baggage", "k=v"]]));
    assert.equal(ctx.traceSource, "generated");
    assert.deepEqual(Object.keys(outboundHeaders(ctx)), ["X-Request-Id", "traceparent"]);
  });
});

describe("deriveContext: generated identifiers", () => {
  it("are well-formed, nonzero and distinct", () => {
    const spans = new Set();
    const traces = new Set();
    for (let i = 0; i < 1000; i += 1) {
      const ctx = deriveContext(new Headers());
      assert.match(ctx.requestId, HEX32);
      assert.match(ctx.traceId, HEX32);
      assert.match(ctx.spanId, HEX16);
      assert.notEqual(ctx.traceId, ZERO32);
      assert.notEqual(ctx.spanId, ZERO16);
      spans.add(ctx.spanId);
      traces.add(ctx.traceId);
    }
    assert.equal(spans.size, 1000);
    assert.equal(traces.size, 1000);
  });

  it("regenerates an all-zero trace or span id", () => {
    const real = crypto.randomBytes;
    let calls = 0;
    // The first four draws are zero; without regeneration a context needs only three draws.
    const draw = mock.method(crypto, "randomBytes", (size) => (calls++ < 4 ? Buffer.alloc(size) : real(size)));
    syncBuiltinESMExports();
    try {
      const ctx = deriveContext(new Headers());
      assert.ok(calls > 4, `randomBytes drawn ${calls} times`);
      assert.match(ctx.traceId, HEX32);
      assert.match(ctx.spanId, HEX16);
      assert.notEqual(ctx.traceId, ZERO32);
      assert.notEqual(ctx.spanId, ZERO16);
    } finally {
      draw.mock.restore();
      syncBuiltinESMExports();
    }
  });
});

describe("outboundHeaders", () => {
  it("returns exactly X-Request-Id and traceparent with the caller's span", () => {
    const trace = hex(16);
    const parent = hex(8);
    const ctx = deriveContext(
      headersOf([
        ["X-Request-Id", "req-outbound-1"],
        ["traceparent", `00-${trace}-${parent}-01`],
        ["Cookie", "session=abc"],
        ["Authorization", "Bearer abc"],
        ["baggage", "k=v"],
        ["tracestate", "k=v"],
        ["X-Api-Key", "abc"],
      ]),
    );
    const headers = outboundHeaders(ctx);
    assert.deepEqual(Object.keys(headers), ["X-Request-Id", "traceparent"]);
    assert.equal(headers["X-Request-Id"], "req-outbound-1");
    assert.equal(headers.traceparent, `00-${trace}-${ctx.spanId}-01`);
    assert.notEqual(ctx.spanId, parent);
  });

  it("starts a new trace with flags 00", () => {
    const ctx = deriveContext(new Headers());
    assert.equal(outboundHeaders(ctx).traceparent, `00-${ctx.traceId}-${ctx.spanId}-00`);
  });
});

describe("withCompletionLog: completion record", () => {
  it("writes one record with exact keys and types for a valid inbound context", async () => {
    const trace = hex(16);
    const parent = hex(8);
    const handler = withCompletionLog("/api/health", alive);
    const { result, raw, records } = await captured(() =>
      handler(request(URL_HEALTH, [["X-Request-Id", "req-valid-0001"], ["traceparent", `00-${trace}-${parent}-01`]])),
    );
    assertOneLine(raw);
    assert.equal(records.length, 1);
    const [record] = records;
    assert.deepEqual(Object.keys(record), [...BASE_KEYS, "parent_span_id", "trace_flags", "outcome"]);
    assert.match(record.time, TIME);
    assert.equal(record.level, "info");
    assert.equal(record.service, "web");
    assert.equal(record.event, "http.request.completed");
    assert.equal(record.method, "GET");
    assert.equal(record.route, "/api/health");
    assert.equal(record.status, 200);
    assert.ok(Number.isInteger(record.duration_ms) && record.duration_ms >= 0);
    assert.equal(record.request_id, "req-valid-0001");
    assert.equal(record.request_id_source, "inbound");
    assert.equal(record.trace_id, trace);
    assert.equal(record.trace_source, "inbound");
    assert.match(record.span_id, HEX16);
    assert.notEqual(record.span_id, ZERO16);
    assert.equal(record.parent_span_id, parent);
    assert.equal(record.trace_flags, "01");
    assert.equal(record.outcome, "completed");
    for (const key of Object.keys(record)) assert.doesNotMatch(key, SENSITIVE_KEY);

    assert.equal(result.status, 200);
    assert.equal(result.headers.get("X-Request-Id"), "req-valid-0001");
    assert.equal(result.headers.get("Cache-Control"), "no-store");
    assert.equal(await result.text(), '{"status":"alive"}');
  });

  it("omits parent_span_id for a generated context and reports generated sources", async () => {
    const handler = withCompletionLog("/api/health", alive);
    const { result, records } = await captured(() => handler(request(URL_HEALTH)));
    assert.equal(records.length, 1);
    const [record] = records;
    assert.deepEqual(Object.keys(record), [...BASE_KEYS, "trace_flags", "outcome"]);
    assert.equal(record.request_id_source, "generated");
    assert.equal(record.trace_source, "generated");
    assert.match(record.request_id, HEX32);
    assert.match(record.trace_id, HEX32);
    assert.equal(record.trace_flags, "00");
    assert.equal(result.headers.get("X-Request-Id"), record.request_id);
  });

  it("records core_probe on the readiness route only, with the handled status and level", async () => {
    const handler = withCompletionLog("/api/ready", async () => ({
      coreProbe: "not_ready",
      response: Response.json(
        { status: "unready", checks: { core: "unavailable" } },
        { status: 503, headers: { "Cache-Control": "no-store" } },
      ),
    }));
    const { result, records } = await captured(() => handler(request(URL_READY)));
    assert.equal(records.length, 1);
    const [record] = records;
    assert.deepEqual(Object.keys(record), [...BASE_KEYS, "trace_flags", "outcome", "core_probe"]);
    assert.equal(record.route, "/api/ready");
    assert.equal(record.status, 503);
    assert.equal(record.level, "warn");
    assert.equal(record.outcome, "completed");
    assert.equal(record.core_probe, "not_ready");
    assert.equal(result.status, 503);
    assert.equal(result.headers.get("X-Request-Id"), record.request_id);
    assert.equal(await result.text(), '{"status":"unready","checks":{"core":"unavailable"}}');

    for (const coreProbe of ["ok", "timeout", "unreachable"]) {
      const ready = withCompletionLog("/api/ready", () => ({ coreProbe, response: alive() }));
      const run = await captured(() => ready(request(URL_READY)));
      assert.equal(run.records[0].core_probe, coreProbe);
    }
  });

  it("treats a malformed readiness result as a failure", async () => {
    const handler = withCompletionLog("/api/ready", async () => ({ coreProbe: "maybe", response: alive() }));
    const { result, records } = await captured(() => handler(request(URL_READY)));
    assert.equal(result.status, 500);
    assert.equal(records.length, 1);
    assert.equal(records[0].outcome, "failed");
    assert.equal(records[0].error_type, "TypeError");
    assert.equal("core_probe" in records[0], false);
  });

  it("logs a returned 500 as warn and normalizes the method", async () => {
    const failing = withCompletionLog("/api/health", () => new Response(null, { status: 500 }));
    const head = withCompletionLog("/api/health", alive);
    const other = withCompletionLog("/api/health", alive);
    const { records } = await captured(async () => {
      await failing(request(URL_HEALTH));
      await head(request(URL_HEALTH, [], { method: "HEAD" }));
      await other(request(URL_HEALTH, [], { method: "PROPFIND" }));
    });
    assert.equal(records.length, 3);
    assert.equal(records[0].level, "warn");
    assert.equal(records[0].outcome, "completed");
    assert.equal(records[1].method, "HEAD");
    assert.equal(records[2].method, "OTHER");
  });
});

describe("withCompletionLog: failures", () => {
  it("turns a thrown error into an empty no-store 500 with exactly one failed record", async () => {
    const secret = canary();
    const handler = withCompletionLog("/api/health", () => {
      const error = new Error(`message ${secret}`, { cause: new Error(`cause ${secret}`) });
      throw Object.assign(error, { detail: secret });
    });
    let run;
    await assert.doesNotReject(async () => {
      run = await captured(() => handler(request(URL_HEALTH, [["X-Request-Id", "req-failure-01"]])));
    });
    const { result, raw, records, writes } = run;
    assertOneLine(raw);
    assert.equal(records.length, 1);
    const [record] = records;
    assert.equal(record.outcome, "failed");
    assert.equal(record.error_type, "Error");
    assert.equal(record.level, "error");
    assert.equal(record.status, 500);
    assert.deepEqual(Object.keys(record), [...BASE_KEYS, "trace_flags", "outcome", "error_type"]);
    assert.equal(result.status, 500);
    assert.equal(result.body, null);
    assert.equal(await result.text(), "");
    assert.equal(result.headers.get("Cache-Control"), "no-store");
    assert.equal(result.headers.get("X-Request-Id"), "req-failure-01");
    assert.deepEqual([...result.headers.keys()].sort(), ["cache-control", "x-request-id"]);
    assertAbsent(secret, [...writes, responseHeaderText(result)]);
  });

  it("does not reject when an async handler rejects", async () => {
    const handler = withCompletionLog("/api/ready", async () => {
      throw new RangeError("out of range");
    });
    const { result, records } = await captured(() => handler(request(URL_READY)));
    assert.equal(result.status, 500);
    assert.equal(records.length, 1);
    assert.equal(records[0].error_type, "RangeError");
    assert.equal("core_probe" in records[0], false);
  });

  class CustomTypeError extends TypeError {}
  class CustomRangeError extends RangeError {}
  class CustomDOMException extends DOMException {}

  // [label, factory(canary), expected category]
  const cases = [
    ["Error", (c) => new Error(c), "Error"],
    ["Error with a valid canary-bearing name", (c) => Object.assign(new Error(c), { name: `Cx${c}` }), "Error"],
    ["Error with an invalid name", (c) => Object.assign(new Error(c), { name: `bad name ${c}` }), "Error"],
    ["Error with a name getter that throws", (c) => Object.defineProperty(new Error(c), "name", { get() { throw new Error(c); } }), "Error"],
    ["Error with a message getter that throws", (c) => Object.defineProperty(new Error(), "message", { get() { throw new Error(c); } }), "Error"],
    ["TypeError subclass with a custom name", (c) => Object.assign(new CustomTypeError(c), { name: `Cx${c}` }), "TypeError"],
    ["TypeError", (c) => new TypeError(c), "TypeError"],
    ["RangeError subclass", (c) => new CustomRangeError(c), "RangeError"],
    ["SyntaxError", (c) => new SyntaxError(c), "SyntaxError"],
    ["ReferenceError", (c) => new ReferenceError(c), "ReferenceError"],
    ["EvalError", (c) => new EvalError(c), "EvalError"],
    ["URIError", (c) => new URIError(c), "URIError"],
    ["AggregateError", (c) => new AggregateError([new Error(c)], c), "AggregateError"],
    ["DOMException with an arbitrary name", (c) => new DOMException(c, `Cx${c}`), "DOMException"],
    ["DOMException subclass with an arbitrary name", (c) => new CustomDOMException(c, `Cx${c}`), "DOMException"],
    ["AbortError", (c) => new DOMException(c, "AbortError"), "AbortError"],
    ["TimeoutError", (c) => new DOMException(c, "TimeoutError"), "TimeoutError"],
    ["abort reason", () => AbortSignal.abort().reason, "AbortError"],
    ["DOMException with a spoofed own name", (c) => Object.defineProperty(new DOMException(c, "SyntaxError"), "name", { value: "AbortError" }), "DOMException"],
    ["object inheriting from DOMException", () => Object.create(DOMException.prototype), "DOMException"],
    ["string", (c) => c, "NonError"],
    ["error-like object", (c) => ({ name: `Cx${c}`, message: c, stack: c }), "NonError"],
    ["null", () => null, "NonError"],
    ["undefined", () => undefined, "NonError"],
    ["number", () => 42, "NonError"],
    ["proxy whose prototype lookup throws", (c) => new Proxy({}, { getPrototypeOf() { throw new Error(c); } }), "NonError"],
    ["revoked proxy", () => { const { proxy, revoke } = Proxy.revocable({}, {}); revoke(); return proxy; }, "NonError"],
  ];

  for (const [label, make, expected] of cases) {
    it(`categorizes ${label} as ${expected} without reading the value`, async () => {
      const secret = canary();
      assert.equal(errorCategory(make(secret)), expected);
      const handler = withCompletionLog("/api/health", () => {
        throw make(secret);
      });
      const { result, records, writes } = await captured(() => handler(request(URL_HEALTH)));
      assert.equal(records.length, 1);
      assert.equal(records[0].outcome, "failed");
      assert.equal(records[0].error_type, expected);
      assert.match(records[0].error_type, ERROR_TYPE);
      assert.equal(result.status, 500);
      assert.equal(await result.text(), "");
      assertAbsent(secret, [...writes, responseHeaderText(result)]);
    });
  }

  it("categorizes AbortSignal.timeout's reason as TimeoutError", async () => {
    const signal = AbortSignal.timeout(1);
    await new Promise((resolve) => signal.addEventListener("abort", resolve, { once: true }));
    assert.equal(errorCategory(signal.reason), "TimeoutError");
  });
});

describe("withCompletionLog: cancellation", () => {
  it("records cancelled once when the client aborts during the handler", async () => {
    const controller = new AbortController();
    const handler = withCompletionLog("/api/health", async () => {
      controller.abort();
      return alive();
    });
    const { result, records } = await captured(() => handler(request(URL_HEALTH, [], { signal: controller.signal })));
    assert.equal(records.length, 1);
    const [record] = records;
    assert.equal(record.outcome, "cancelled");
    assert.equal(record.status, 0);
    assert.equal(record.level, "info");
    assert.equal("error_type" in record, false);
    assert.equal(result.headers.get("X-Request-Id"), record.request_id);
  });

  it("records cancelled once for an already aborted request", async () => {
    const handler = withCompletionLog("/api/ready", () => ({ coreProbe: "ok", response: alive() }));
    const { records } = await captured(() => handler(request(URL_READY, [], { signal: AbortSignal.abort() })));
    assert.equal(records.length, 1);
    assert.equal(records[0].outcome, "cancelled");
    assert.equal(records[0].core_probe, "ok");
  });

  it("records cancelled, not failed, and still answers 500 when the handler throws after an abort", async () => {
    const secret = canary();
    const controller = new AbortController();
    const handler = withCompletionLog("/api/health", () => {
      controller.abort();
      throw new Error(secret);
    });
    let run;
    await assert.doesNotReject(async () => {
      run = await captured(() => handler(request(URL_HEALTH, [], { signal: controller.signal })));
    });
    const { result, records, writes } = run;
    assert.equal(records.length, 1);
    assert.equal(records[0].outcome, "cancelled");
    assert.equal("error_type" in records[0], false);
    assert.equal(result.status, 500);
    assert.equal(result.headers.get("Cache-Control"), "no-store");
    assertAbsent(secret, writes);
  });
});

describe("withCompletionLog: invalid values are never echoed", () => {
  it("replaces canary-bearing request ids and traceparents everywhere", async () => {
    const secret = hex(16);
    const parent = hex(8);
    const inputs = [
      [["X-Request-Id", `bad ${secret}`]],
      [["X-Request-Id", `${secret}${secret}${secret}`]],
      [["X-Request-Id", `${secret},${secret}`]],
      [["X-Request-Id", secret], ["X-Request-Id", secret]],
      [["traceparent", `01-${secret}-${parent}-01`]],
      [["traceparent", `00-${secret}-${parent}-01-x`]],
      [["traceparent", `00-${secret.toUpperCase()}-${parent}-01`]],
      [["traceparent", `00-${secret}-${ZERO16}-01`]],
      [["traceparent", `00-${secret}-${parent}-01`], ["traceparent", `00-${secret}-${parent}-01`]],
    ];
    for (const entries of inputs) {
      let outbound;
      const handler = withCompletionLog("/api/health", (_request, ctx) => {
        outbound = JSON.stringify(outboundHeaders(ctx));
        return alive();
      });
      const { result, records, writes } = await captured(() => handler(request(URL_HEALTH, entries)));
      assert.equal(records.length, 1);
      assert.equal(records[0].status, 200, "an invalid value never causes a 500");
      assertAbsent(secret, [...writes, responseHeaderText(result), outbound]);
      assertAbsent(secret.toUpperCase(), [...writes, responseHeaderText(result), outbound]);
    }
  });
});

describe("withCompletionLog: isolation", () => {
  it("keeps 200 concurrent invocations with distinct contexts apart", async () => {
    const inputs = Array.from({ length: 200 }, (_, i) => ({
      id: `req-${String(i).padStart(4, "0")}-${hex(4)}`,
      trace: hex(16),
      parent: hex(8),
      flags: i % 2 === 0 ? "01" : "00",
      delay: i % 7,
      fail: i % 10 === 0,
    }));
    const handler = withCompletionLog("/api/health", async (req, ctx) => {
      const params = new URL(req.url).searchParams;
      await new Promise((resolve) => setTimeout(resolve, Number(params.get("delay"))));
      if (params.get("fail") === "1") throw new Error("concurrent failure");
      return Response.json({ ctx, outbound: outboundHeaders(ctx) });
    });
    const { result: responses, records } = await captured(() =>
      Promise.all(
        inputs.map((input) =>
          handler(
            request(`${URL_HEALTH}?delay=${input.delay}&fail=${input.fail ? 1 : 0}`, [
              ["X-Request-Id", input.id],
              ["traceparent", `00-${input.trace}-${input.parent}-${input.flags}`],
            ]),
          ),
        ),
      ),
    );
    assert.equal(records.length, 200);
    const byId = new Map(records.map((record) => [record.request_id, record]));
    assert.equal(byId.size, 200);
    assert.equal(new Set(records.map((record) => record.span_id)).size, 200);
    for (const [i, input] of inputs.entries()) {
      const response = responses[i];
      const record = byId.get(input.id);
      assert.ok(record, `record for ${input.id}`);
      assert.equal(response.headers.get("X-Request-Id"), input.id);
      assert.equal(record.trace_id, input.trace);
      assert.equal(record.parent_span_id, input.parent);
      assert.equal(record.trace_flags, input.flags);
      if (input.fail) {
        assert.equal(response.status, 500);
        assert.equal(record.outcome, "failed");
        continue;
      }
      const body = await response.json();
      assert.equal(record.outcome, "completed");
      assert.equal(body.ctx.requestId, input.id);
      assert.equal(body.ctx.traceId, input.trace);
      assert.equal(body.ctx.parentSpanId, input.parent);
      assert.equal(body.ctx.spanId, record.span_id);
      assert.deepEqual(body.outbound, {
        "X-Request-Id": input.id,
        traceparent: `00-${input.trace}-${record.span_id}-${input.flags}`,
      });
    }
  });
});
