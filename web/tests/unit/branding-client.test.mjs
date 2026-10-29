// Unit tests for lib/branding.ts (G017 Plan r2 §6.2 "Web unit"): the closed reading of core's public projection and
// management answers (grammar and name re-check, label bounds, exact field sets), the exact outbound public requests
// (diagnostic headers only, no credential, no body) and the bounded logo proxy reading. fetch is replaced by a recorder,
// so nothing leaves the process. The lib modules import their siblings without extensions, as the Next bundler resolves
// them; a resolve hook scoped to the lib directory maps them to the .ts sources.
import assert from "node:assert/strict";
import { registerHooks } from "node:module";
import { afterEach, beforeEach, describe, it, mock } from "node:test";

registerHooks({
  resolve(specifier, context, nextResolve) {
    if (/^\.\/[a-z-]+$/.test(specifier) && /\/lib\/[a-z-]+\.ts$/.test(context.parentURL ?? "")) {
      return nextResolve(`${specifier}.ts`, context);
    }
    return nextResolve(specifier, context);
  },
});

const branding = await import("../../lib/branding.ts");
const { deriveContext } = await import("../../lib/request-context.ts");
const tenantCssModule = await import("../../lib/tenant-css.ts");

const CORE = "http://core.internal:8180";
const ORIGIN = "https://org-a.localhost:543";
const SECRET = "S".repeat(43);
const CSRF = "C".repeat(43);
const OBJ = "0b8f6c1e-2a4d-4c7b-9e1f-3a5b7c9d1e2f";
const CTX = deriveContext(new Headers({ "x-request-id": "req-branding-1", traceparent: `00-${"a".repeat(32)}-${"b".repeat(16)}-01` }));
const ENV = { PRODUCT_NAME: "Test Product", PLATFORM_BASE_DOMAIN: "example.test", CUSTOM_DOMAIN_CNAME_TARGET: "edge.example.test", CORE_INTERNAL_URL: CORE };

let saved;
let calls;
let reply;

beforeEach(() => {
  saved = Object.fromEntries(Object.keys(ENV).map((key) => [key, process.env[key]]));
  Object.assign(process.env, ENV);
  calls = [];
  reply = () => Response.json({ error: "branding_unavailable" }, { status: 503 });
  mock.method(globalThis, "fetch", async (url, init) => {
    calls.push({ url: String(url), init });
    return reply();
  });
});

afterEach(() => {
  mock.restoreAll();
  for (const [key, value] of Object.entries(saved)) {
    if (value === undefined) delete process.env[key];
    else process.env[key] = value;
  }
});

const GOOD = { revision: 3, display_label: "Acme <Tools>", light: { primary: "#1d4ed8" }, dark: {}, logo: { v: "0123456789ab", content_type: "image/png" } };

describe("public projection", () => {
  it("reads the exact projection, keeps the label as text and maps the logo to its version", () => {
    assert.deepEqual(branding.readPublicBranding(200, GOOD), {
      ok: true,
      value: { revision: 3, label: "Acme <Tools>", logo: { version: "0123456789ab" }, tokens: { light: { primary: "#1d4ed8" }, dark: {} } },
    });
    assert.deepEqual(branding.readPublicBranding(200, { revision: 0, light: {}, dark: {} }), {
      ok: true,
      value: { revision: 0, label: null, logo: null, tokens: { light: {}, dark: {} } },
    });
  });
  it("treats anything outside the grammar, the 15 names or the field set as malformed (never rendered)", () => {
    for (const change of [
      { light: { primary: "#1D4ED8" } },
      { light: { primary: "url(x)" } },
      { light: { primary: "#1d4ed8;}" } },
      { light: { primary: "#1d4ed8ff" } },
      { light: { overlay: "#000000" } },
      { light: { ink: "#000000" } },
      { dark: { primary: 7 } },
      { dark: [] },
      { revision: -1 },
      { revision: "3" },
      { display_label: "" },
      { display_label: " padded" },
      { display_label: "line\nbreak" },
      { display_label: "x".repeat(81) },
      { logo: { v: "0123456789AB", content_type: "image/png" } },
      { logo: { v: "0123456789ab", content_type: "image/svg+xml" } },
      { logo: { v: "0123456789ab", content_type: "image/png", key: "k" } },
      { organization_name: "Acme Ltd" },
    ]) {
      assert.deepEqual(branding.readPublicBranding(200, { ...GOOD, ...change }), { ok: false, error: "malformed" }, JSON.stringify(change));
    }
    assert.equal(branding.readPublicBranding(200, { ...GOOD, display_label: "é".repeat(80) }).ok, true);
  });
  it("maps unknown_host, an outage and other answers to closed errors", () => {
    assert.deepEqual(branding.readPublicBranding(404, { error: "unknown_host" }), { ok: false, error: "unknown_host" });
    assert.deepEqual(branding.readPublicBranding(503, { error: "branding_unavailable" }), { ok: false, error: "unavailable" });
    assert.deepEqual(branding.readPublicBranding(404, { error: "not_found" }), { ok: false, error: "malformed" });
    assert.deepEqual(branding.readPublicBranding(400, { error: "ambiguous_parameter" }), { ok: false, error: "malformed" });
  });
  it("sends a GET with the diagnostic headers only: no credential, cookie or body", async () => {
    reply = () => Response.json(GOOD);
    const result = await branding.resolveBranding(CTX, ORIGIN);
    assert.equal(result.ok, true);
    assert.equal(calls.length, 1);
    const { url, init } = calls[0];
    assert.equal(url, `${CORE}/internal/v1/branding/public?origin=${encodeURIComponent(ORIGIN)}`);
    assert.equal(init.method, "GET");
    assert.equal(init.body, undefined);
    assert.equal(init.cache, "no-store");
    assert.equal(init.redirect, "error");
    assert.deepEqual(Object.keys(init.headers).sort(), ["X-Request-Id", "traceparent"]);
  });
  it("is unavailable, never a clean unknown host, when core does not answer", async () => {
    reply = () => {
      throw new TypeError("fetch failed");
    };
    assert.deepEqual(await branding.resolveBranding(CTX, ORIGIN), { ok: false, error: "unavailable" });
  });
});

describe("logo proxy reading", () => {
  const PNG = new Uint8Array([137, 80, 78, 71, 1, 2, 3]);
  const logo = (headers, body = PNG, status = 200) => () => new Response(body, { status, headers });
  const OK = { "content-type": "image/png", "content-length": String(PNG.byteLength), etag: '"0123456789ab"' };
  it("accepts only the current version's bytes with the reviewed headers", async () => {
    reply = logo(OK);
    const r = await branding.fetchLogo(CTX, ORIGIN, "0123456789ab");
    assert.equal(r.ok, true);
    assert.deepEqual([...r.bytes], [...PNG]);
    assert.equal(calls[0].url, `${CORE}/internal/v1/branding/logo?origin=${encodeURIComponent(ORIGIN)}&v=0123456789ab`);
    assert.deepEqual(Object.keys(calls[0].init.headers).sort(), ["X-Request-Id", "traceparent"]);
  });
  it("refuses a bad version before any call and maps refusals and contract breaks", async () => {
    assert.deepEqual(await branding.fetchLogo(CTX, ORIGIN, "0123456789AB"), { ok: false, error: "not_found" });
    assert.equal(calls.length, 0);
    reply = () => Response.json({ error: "not_found" }, { status: 404 });
    assert.deepEqual(await branding.fetchLogo(CTX, ORIGIN, "0123456789ab"), { ok: false, error: "not_found" });
    for (const headers of [
      { ...OK, "content-type": "image/svg+xml" },
      { ...OK, etag: '"ffffffffffff"' },
      { ...OK, "content-length": "524289" },
      { ...OK, "content-length": "3" },
    ]) {
      reply = logo(headers);
      assert.deepEqual(await branding.fetchLogo(CTX, ORIGIN, "0123456789ab"), { ok: false, error: "unavailable" }, JSON.stringify(headers));
    }
  });
});

describe("management answers", () => {
  const VIEW = {
    version: 4,
    display_label: null,
    light: { primary: "#1d4ed8" },
    dark: {},
    logo: { object_id: OBJ, content_type: "image/png", byte_size: 812, state: "ready", v: "0123456789ab" },
    uploads: [{ object_id: OBJ, content_type: "image/jpeg", byte_size: 90, state: "ready" }],
    uploads_total: 1,
  };
  it("reads the view, an absent (purged) logo and the closed refusals with their field", () => {
    const r = branding.readView(200, { branding: VIEW });
    assert.equal(r.ok && r.value.logo.version, "0123456789ab");
    assert.equal(r.ok && r.value.uploads[0].version, null);
    const absent = branding.readView(200, { branding: { ...VIEW, logo: { object_id: OBJ, content_type: "", byte_size: 0, state: "absent" } } });
    assert.equal(absent.ok && absent.value.logo.state, "absent");
    assert.deepEqual(branding.readView(422, { error: "contrast_failed", field: "T7:light" }), { ok: false, error: "contrast_failed", field: "T7:light" });
    assert.deepEqual(branding.readView(409, { error: "version_conflict" }), { ok: false, error: "version_conflict" });
    assert.deepEqual(branding.readView(500, undefined), { ok: false, error: "unavailable" });
    assert.deepEqual(branding.readView(418, { error: "teapot" }), { ok: false, error: "malformed" });
    for (const change of [{ light: { primary: "red" } }, { logo: { ...VIEW.logo, content_type: "" } }, { uploads: [{ object_id: "x" }] }, { version: -1 }]) {
      assert.deepEqual(branding.readView(200, { branding: { ...VIEW, ...change } }), { ok: false, error: "malformed" }, JSON.stringify(change));
    }
  });
  it("sends the session credential and the documented JSON fields only", async () => {
    reply = () => Response.json({ branding: VIEW });
    await branding.updateBranding(CTX, { origin: ORIGIN, csrfToken: CSRF, expectedVersion: 4 }, { label: "Acme", light: {}, dark: {} }, SECRET);
    const { url, init } = calls[0];
    assert.equal(url, `${CORE}/internal/v1/branding`);
    assert.equal(init.method, "POST");
    assert.equal(init.headers.Authorization, `Session ${SECRET}`);
    assert.equal(init.headers["Content-Type"], "application/json");
    assert.deepEqual(JSON.parse(init.body), { origin: ORIGIN, csrf_token: CSRF, expected_version: 4, display_label: "Acme", light: {}, dark: {} });
    await branding.setBrandingLogo(CTX, { origin: ORIGIN, csrfToken: CSRF, expectedVersion: 4 }, new Uint8Array([1, 2]), SECRET);
    assert.equal(calls[1].url, `${CORE}/internal/v1/branding/logo?origin=${encodeURIComponent(ORIGIN)}`);
    assert.equal(calls[1].init.headers["X-CSRF-Token"], CSRF);
    assert.equal(calls[1].init.headers["X-Expected-Version"], "4");
  });
});

describe("management forms", () => {
  const form = (pairs) => new URLSearchParams(pairs);
  const get = (fields) => (name) => {
    const v = fields.getAll(name);
    return v.length > 1 ? undefined : v[0];
  };

  it("reads the label and the colour fields: blank inherits, #RRGGBB is lowercased, label trimmed and NFC", () => {
    const r = branding.readBrandingInput(form([["display_label", "  Café "], ["light.primary", "#0F6B5C"], ["light.ring", ""], ["dark.accent", " #f0c674 "]]));
    assert.deepEqual(r, { ok: true, value: { label: "Café", light: { primary: "#0f6b5c" }, dark: { accent: "#f0c674" } } });
    assert.deepEqual(branding.readBrandingInput(form([["display_label", "   "]])), { ok: true, value: { label: null, light: {}, dark: {} } });
  });

  it("names the closed field of an invalid value; a repeated field is a bad request", () => {
    assert.deepEqual(branding.readBrandingInput(form([["light.primary", "red"]])), { ok: false, field: "light" });
    assert.deepEqual(branding.readBrandingInput(form([["dark.border", "#12345"]])), { ok: false, field: "dark" });
    assert.deepEqual(branding.readBrandingInput(form([["display_label", "a\u0007b"]])), { ok: false, field: "display_label" });
    assert.deepEqual(branding.readBrandingInput(form([["display_label", "x".repeat(81)]])), { ok: false, field: "display_label" });
    assert.deepEqual(branding.readBrandingInput(form([["light.primary", "#000000"], ["light.primary", "#ffffff"]])), { ok: false });
    // overlay and unknown names are never read.
    assert.deepEqual(branding.readBrandingInput(form([["light.overlay", "#000000"], ["light.unknown", "x"]])), { ok: true, value: { label: null, light: {}, dark: {} } });
  });

  it("reads the CSRF token and expected version", () => {
    assert.deepEqual(branding.readMutationFields(get(form([["csrf_token", CSRF], ["expected_version", "7"]]))), { csrfToken: CSRF, expectedVersion: 7 });
    assert.deepEqual(branding.readMutationFields(get(form([["expected_version", "0"]]))), { csrfToken: "", expectedVersion: 0 });
    for (const v of ["", "-1", "01", "1.5", "9".repeat(16)]) {
      assert.equal(branding.readMutationFields(get(form([["csrf_token", CSRF], ["expected_version", v]]))), undefined, v);
    }
  });

  it("accepts exactly one multipart/form-data Content-Type with a boundary", () => {
    assert.equal(branding.isMultipartForm("multipart/form-data; boundary=----WebKitFormBoundaryAbC123"), true);
    for (const v of [null, "multipart/form-data", "multipart/form-data; boundary=a, text/plain", "application/x-www-form-urlencoded", "multipart/mixed; boundary=x"]) {
      assert.equal(branding.isMultipartForm(v), false, String(v));
    }
  });

  async function multipart(parts) {
    const fd = new FormData();
    for (const [name, value] of parts) fd.append(name, value);
    const r = new Request("http://form.test/", { method: "POST", body: fd });
    return { type: r.headers.get("content-type"), body: new Uint8Array(await r.arrayBuffer()) };
  }

  it("reduces the logo form to the raw bytes, at most 512 KiB, with exactly the three fields", async () => {
    const png = new Uint8Array([0x89, 0x50, 0x4e, 0x47, 1, 2, 3]);
    let m = await multipart([["csrf_token", CSRF], ["expected_version", "3"], ["logo", new Blob([png], { type: "image/svg+xml" })]]);
    const ok = await branding.parseLogoUpload(m.type, m.body);
    assert.equal(ok.ok, true);
    assert.deepEqual(ok.fields, { csrfToken: CSRF, expectedVersion: 3 });
    assert.deepEqual([...ok.bytes], [...png]);
    m = await multipart([["csrf_token", CSRF], ["expected_version", "3"], ["logo", new Blob([new Uint8Array(branding.MAX_LOGO_BYTES + 1)])]]);
    assert.deepEqual(await branding.parseLogoUpload(m.type, m.body), { ok: false, error: "too_large" });
    for (const parts of [
      [["csrf_token", CSRF], ["expected_version", "3"]],
      [["csrf_token", CSRF], ["expected_version", "3"], ["logo", new Blob([])]],
      [["csrf_token", CSRF], ["expected_version", "3"], ["logo", "text"]],
      [["csrf_token", CSRF], ["expected_version", "3"], ["logo", new Blob([png])], ["logo", new Blob([png])]],
      [["csrf_token", CSRF], ["expected_version", "3"], ["logo", new Blob([png])], ["other", "x"]],
      [["csrf_token", CSRF], ["logo", new Blob([png])]],
    ]) {
      m = await multipart(parts);
      assert.deepEqual(await branding.parseLogoUpload(m.type, m.body), { ok: false, error: "bad_request" });
    }
    assert.deepEqual(await branding.parseLogoUpload("multipart/form-data; boundary=x", new Uint8Array([1, 2, 3])), { ok: false, error: "bad_request" });
  });

  it("answers each management outcome with the closed status, field or fixed page", () => {
    const at = (res) => [res.status, res.headers.get("location"), res.headers.get("cache-control")];
    assert.deepEqual(at(branding.brandingAnswer({ ok: true, value: {} }, "logo_set", "x")), [303, "/organization/branding?status=logo_set", "no-store"]);
    assert.deepEqual(at(branding.brandingAnswer({ ok: false, error: "version_conflict" }, "updated", "x")), [303, "/organization/branding?status=version_conflict", "no-store"]);
    assert.deepEqual(at(branding.brandingAnswer({ ok: false, error: "contrast_failed", field: "T10:dark" }, "updated", "x"))[1], "/organization/branding?status=contrast_failed&field=T10%3Adark");
    assert.equal(at(branding.brandingAnswer({ ok: false, error: "contrast_failed", field: "T11:dark" }, "updated", "x"))[1], "/organization/branding?status=contrast_failed");
    // G069 (Plan r1 S11): core's new T12 pair (text on the field fill) round-trips from its refusal to the page's field.
    const t12 = branding.readView(422, { error: "contrast_failed", field: "T12:light" });
    assert.deepEqual(t12, { ok: false, error: "contrast_failed", field: "T12:light" });
    assert.equal(at(branding.brandingAnswer(t12, "updated", "x"))[1], "/organization/branding?status=contrast_failed&field=T12%3Alight");
    assert.equal(at(branding.brandingAnswer({ ok: false, error: "contrast_failed", field: "T13:light" }, "updated", "x"))[1], "/organization/branding?status=contrast_failed");
    assert.equal(at(branding.brandingAnswer({ ok: false, error: "invalid_request", field: "light" }, "updated", "x"))[1], "/organization/branding?status=invalid_input&field=light");
    assert.equal(branding.brandingAnswer({ ok: false, error: "invalid_request", field: "other" }, "updated", "x").status, 400);
    assert.equal(at(branding.brandingAnswer({ ok: false, error: "publication_outcome_unknown" }, "logo_set", "x"))[1], "/organization/branding?status=publication_outcome_unknown");
    assert.equal(at(branding.brandingAnswer({ ok: false, error: "not_found" }, "upload_deleted", "x"))[1], "/organization/branding?status=upload_not_found");
    for (const [error, status] of [["forbidden", 404], ["not_available", 404], ["not_found", 404], ["unknown_host", 404], ["csrf_failed", 403], ["origin_rejected", 403], ["unauthenticated", 401], ["unavailable", 503], ["branding_unavailable", 503], ["malformed", 503]]) {
      const res = branding.brandingAnswer({ ok: false, error }, "updated", "The change");
      assert.equal(res.status, status, error);
      assert.equal(res.headers.get("cache-control"), "no-store", error);
    }
  });
});

// G069 (Plan r1 §2.5, S7–S9): lib/tenant-css.ts, the saved-branding fan-out onto the kit's variables. The expected
// table is typed here from the Plan, independently of the module's own table.
describe("tenant stylesheet fan-out", () => {
  const { tenantCss, previewVariables } = tenantCssModule;
  const LIGHT = "html:root:not(.dark)";
  const DARK = "html.dark:root";
  const FAN_OUT = [
    ["background", ["--background", "--sidebar"]],
    ["foreground", ["--foreground", "--accent-foreground", "--sidebar-foreground"]],
    ["primary", ["--primary", "--vx-blue", "--sidebar-primary"]],
    ["primary-foreground", ["--primary-foreground", "--sidebar-primary-foreground"]],
    ["accent", ["--vx-gold"]],
    ["muted", ["--muted", "--accent", "--sidebar-accent"]],
    ["muted-foreground", ["--muted-foreground"]],
    ["secondary", ["--secondary"]],
    ["secondary-foreground", ["--secondary-foreground"]],
    ["popover", ["--popover", "--card"]],
    ["popover-foreground", ["--popover-foreground", "--card-foreground"]],
    ["border", ["--border", "--sidebar-border"]],
    ["input", ["--input"]],
    ["destructive", ["--destructive"]],
    ["ring", ["--ring", "--sidebar-ring"]],
  ];
  const decls = (vars, colour) => vars.map((v) => `${v}:${colour}`).join(";");

  it("sets exactly the fan-out variables of each saved name, per appearance", () => {
    assert.equal(FAN_OUT.length, branding.BRANDING_TOKENS.length);
    for (const [name, vars] of FAN_OUT) {
      assert.ok(branding.BRANDING_TOKENS.includes(name), name);
      assert.equal(tenantCss({ light: { [name]: "#0a5a4d" } }), `${LIGHT}{${decls(vars, "#0a5a4d")}}`, `light ${name}`);
      assert.equal(tenantCss({ dark: { [name]: "#ff9c8f" } }), `${DARK}{${decls(vars, "#ff9c8f")}}`, `dark ${name}`);
      assert.deepEqual(previewVariables({ [name]: "#123abc" }), Object.fromEntries(vars.map((v) => [v, "#123abc"])), `preview ${name}`);
    }
  });

  it("never sets the kit's neutral accent from the saved accent, and orders a full palette by the table", () => {
    assert.equal(tenantCss({ light: { accent: "#7a4a00" } }), `${LIGHT}{--vx-gold:#7a4a00}`);
    const shuffled = Object.fromEntries([...FAN_OUT].reverse().map(([name], i) => [name, `#0000${String(i).padStart(2, "0")}`]));
    const expected = FAN_OUT.map(([name, vars]) => decls(vars, shuffled[name])).join(";");
    const css = tenantCss({ light: shuffled, dark: shuffled });
    assert.equal(css, `${LIGHT}{${expected}}${DARK}{${expected}}`);
    assert.match(css, /^[a-z0-9#:;{}().-]+$/);
  });

  it("emits only the appearances that carry values", () => {
    assert.equal(tenantCss({ light: { primary: "#525252" } }), `${LIGHT}{--primary:#525252;--vx-blue:#525252;--sidebar-primary:#525252}`);
    assert.equal(tenantCss({ dark: { destructive: "#e8a6a6" }, light: {} }), `${DARK}{--destructive:#e8a6a6}`);
  });

  it("is empty without branding (loopback or unknown host, platform defaults)", () => {
    assert.equal(tenantCss(undefined), "");
    assert.equal(tenantCss({}), "");
    assert.equal(tenantCss({ light: {}, dark: {} }), "");
    assert.deepEqual(previewVariables(undefined), {});
  });

  it("re-checks every value and name and refuses the whole palette on any failure", () => {
    const refused = [
      { primary: "#0A5A4D" },
      { primary: "#abc" },
      { primary: "#0a5a4d0" },
      { primary: "red" },
      { primary: "#12345g" },
      { primary: "#0a5a4d;}body{color:red" },
      { primary: " #0a5a4d" },
      { primary: 0x0a5a4d },
      { primary: null },
      { overlay: "#000000" },
      { "--primary": "#0a5a4d" },
      { constructor: "#0a5a4d" },
      JSON.parse('{"__proto__":"#0a5a4d"}'),
    ];
    for (const bad of refused) {
      const label = JSON.stringify(bad);
      assert.equal(tenantCss({ light: bad }), "", `light ${label}`);
      assert.equal(tenantCss({ light: { background: "#ffffff" }, dark: bad }), "", `dark ${label} refuses the light values too`);
      assert.deepEqual(previewVariables(bad), {}, `preview ${label}`);
    }
    assert.equal(tenantCss({ light: "#ffffff" }), "");
  });
});
