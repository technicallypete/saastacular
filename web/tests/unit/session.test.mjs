// Unit tests for lib/session.ts: the cookie profile, cookie names and fixed attributes, exactly-one cookie
// reading, own-origin derivation, the browser-form checks in their order and the minimal no-store responses.
// Run by `npm run test:unit` in the web check image with networking disabled: Node strips the module's types,
// and the react-server condition resolves `server-only` to its empty entry.
import assert from "node:assert/strict";
import { afterEach, beforeEach, describe, it } from "node:test";

import {
  AMBIGUOUS,
  CHALLENGE_COOKIE_MAX_AGE,
  LOGOUT_COOKIE_MAX_AGE,
  SIGNUP_COOKIE_MAX_AGE,
  readSignupFormCookie,
  setSignupFormCookie,
  badRequest,
  clearCookie,
  cookieName,
  cookieProfile,
  field,
  forbidden,
  hasCookie,
  jsonResponse,
  methodNotAllowed,
  ownOrigin,
  readCookie,
  readForm,
  seeOther,
  setCookie,
  statusPage,
} from "../../lib/session.ts";

const SECRET = "A".repeat(21) + "-_" + "z".repeat(20); // 43 base64url characters
const OTHER = "B".repeat(43);
const FORM = "application/x-www-form-urlencoded";

let savedEnv;
beforeEach(() => {
  savedEnv = { APP_ENV: process.env.APP_ENV, IDENTITY_COOKIE_PROFILE: process.env.IDENTITY_COOKIE_PROFILE };
});
afterEach(() => {
  for (const [key, value] of Object.entries(savedEnv)) {
    if (value === undefined) delete process.env[key];
    else process.env[key] = value;
  }
});

function useProfile(appEnv, profile) {
  process.env.APP_ENV = appEnv;
  if (profile === undefined) delete process.env.IDENTITY_COOKIE_PROFILE;
  else process.env.IDENTITY_COOKIE_PROFILE = profile;
}

function post(headers, body = "") {
  return new Request("http://web.internal/auth/logout", { method: "POST", headers, body });
}

describe("cookieProfile", () => {
  it("defaults to https in production and loopback-http elsewhere", () => {
    assert.equal(cookieProfile({ APP_ENV: "production" }), "https");
    assert.equal(cookieProfile({ APP_ENV: "development" }), "loopback-http");
    assert.equal(cookieProfile({ APP_ENV: "test", IDENTITY_COOKIE_PROFILE: "  " }), "loopback-http");
  });
  it("accepts the two profiles and refuses loopback-http in production", () => {
    assert.equal(cookieProfile({ APP_ENV: "production", IDENTITY_COOKIE_PROFILE: "https" }), "https");
    assert.equal(cookieProfile({ APP_ENV: "test", IDENTITY_COOKIE_PROFILE: "https" }), "https");
    assert.equal(cookieProfile({ APP_ENV: "development", IDENTITY_COOKIE_PROFILE: "loopback-http" }), "loopback-http");
    assert.throws(() => cookieProfile({ APP_ENV: "production", IDENTITY_COOKIE_PROFILE: "loopback-http" }), /IDENTITY_COOKIE_PROFILE is invalid/);
  });
  it("refuses any other value without echoing it", () => {
    for (const value of ["http", "HTTPS", "none", "https,loopback-http", "Cn-canary-value"]) {
      assert.throws(
        () => cookieProfile({ APP_ENV: "test", IDENTITY_COOKIE_PROFILE: value }),
        (error) => error instanceof Error && !error.message.includes("canary") && /is invalid$/.test(error.message),
      );
    }
  });
});

describe("cookie names and attributes", () => {
  it("uses __Host- names for https and plain names for loopback-http", () => {
    assert.equal(cookieName("session", "https"), "__Host-sid");
    assert.equal(cookieName("login", "https"), "__Host-login");
    assert.equal(cookieName("session", "loopback-http"), "sid");
    assert.equal(cookieName("login", "loopback-http"), "login");
    assert.equal(cookieName("logout", "https"), "__Host-logout");
    assert.equal(cookieName("logout", "loopback-http"), "logout");
  });
  it("issues the logout reference cookie host-only, HttpOnly, SameSite=Lax, Secure (https) for at most ten minutes", () => {
    assert.equal(LOGOUT_COOKIE_MAX_AGE, 600);
    assert.equal(setCookie("logout", SECRET, 600, "https"), `__Host-logout=${SECRET}; Path=/; Max-Age=600; HttpOnly; SameSite=Lax; Secure`);
    assert.equal(setCookie("logout", SECRET, 600, "loopback-http"), `logout=${SECRET}; Path=/; Max-Age=600; HttpOnly; SameSite=Lax`);
    assert.equal(clearCookie("logout", "https"), "__Host-logout=; Path=/; Max-Age=0; HttpOnly; SameSite=Lax; Secure");
    for (const age of [601, 86400, 0]) assert.throws(() => setCookie("logout", SECRET, age, "https"), RangeError, String(age));
    assert.throws(() => setCookie("logout", "not-a-reference", 600, "https"), TypeError);
    assert.doesNotMatch(setCookie("logout", SECRET, 600, "https"), /domain|expires/i);
  });
  it("sets HttpOnly, Path=/, SameSite=Lax, no Domain, and Secure only for https", () => {
    assert.equal(setCookie("session", SECRET, 28800, "https"), `__Host-sid=${SECRET}; Path=/; Max-Age=28800; HttpOnly; SameSite=Lax; Secure`);
    assert.equal(setCookie("login", SECRET, 600, "https"), `__Host-login=${SECRET}; Path=/; Max-Age=600; HttpOnly; SameSite=Lax; Secure`);
    assert.equal(setCookie("session", SECRET, 60, "loopback-http"), `sid=${SECRET}; Path=/; Max-Age=60; HttpOnly; SameSite=Lax`);
    for (const value of [setCookie("session", SECRET, 1, "https"), clearCookie("login", "https")]) {
      assert.doesNotMatch(value, /domain|expires/i);
    }
  });
  it("clears with an empty value and Max-Age=0 under the same attributes", () => {
    assert.equal(clearCookie("session", "https"), "__Host-sid=; Path=/; Max-Age=0; HttpOnly; SameSite=Lax; Secure");
    assert.equal(clearCookie("login", "loopback-http"), "login=; Path=/; Max-Age=0; HttpOnly; SameSite=Lax");
  });
  it("refuses values that are not a canonical credential, including header injection", () => {
    for (const value of ["", "A".repeat(42), "A".repeat(44), `${"A".repeat(42)}=`, `${"A".repeat(42)};`, `${"A".repeat(41)}\r\n`, `${"A".repeat(42)}+`]) {
      assert.throws(() => setCookie("session", value, 60, "https"), TypeError, JSON.stringify(value));
    }
  });
  it("refuses lifetimes outside one second to 24 hours", () => {
    for (const age of [0, -1, 1.5, 86401, Number.NaN, Number.POSITIVE_INFINITY]) {
      assert.throws(() => setCookie("session", SECRET, age, "https"), RangeError, String(age));
    }
    assert.match(setCookie("session", SECRET, 86400, "https"), /Max-Age=86400;/);
  });
});

describe("readCookie", () => {
  it("returns exactly one well-formed value", () => {
    assert.equal(readCookie(`__Host-sid=${SECRET}`, "session", "https"), SECRET);
    assert.equal(readCookie(`a=1; __Host-login=${OTHER}; __Host-sid=${SECRET}`, "session", "https"), SECRET);
    assert.equal(readCookie(`a=1;__Host-login=${OTHER};b=2`, "login", "https"), OTHER);
    assert.equal(readCookie(`  sid = ${SECRET} `, "session", "loopback-http"), SECRET);
  });
  it("yields nothing when absent, repeated or malformed", () => {
    assert.equal(readCookie(null, "session", "https"), undefined);
    assert.equal(readCookie("", "session", "https"), undefined);
    assert.equal(readCookie(`__Host-sid=${SECRET}; __Host-sid=${OTHER}`, "session", "https"), undefined);
    assert.equal(readCookie(`__Host-sid=${SECRET}; __Host-sid=${SECRET}`, "session", "https"), undefined);
    assert.equal(readCookie(`__Host-sid=${"A".repeat(42)}`, "session", "https"), undefined);
    assert.equal(readCookie(`__Host-sid=${SECRET}x`, "session", "https"), undefined);
    assert.equal(readCookie(`__Host-sid="${SECRET}"`, "session", "https"), undefined);
    assert.equal(readCookie("__Host-sid", "session", "https"), undefined);
  });
  it("reads the logout reference like any other: exactly one well-formed value", () => {
    assert.equal(readCookie(`__Host-sid=${OTHER}; __Host-logout=${SECRET}`, "logout", "https"), SECRET);
    assert.equal(readCookie(`__Host-logout=${SECRET}; __Host-logout=${OTHER}`, "logout", "https"), undefined);
    assert.equal(readCookie(`__Host-logout=forged`, "logout", "https"), undefined);
    assert.equal(readCookie(`logout=${SECRET}`, "logout", "https"), undefined);
  });
  it("hasCookie reports presence by exact name only, whatever the value or count", () => {
    assert.equal(hasCookie(null, "logout", "https"), false);
    assert.equal(hasCookie("", "logout", "https"), false);
    assert.equal(hasCookie(`__Host-logout=${SECRET}`, "logout", "https"), true);
    assert.equal(hasCookie("__Host-logout=forged", "logout", "https"), true);
    assert.equal(hasCookie(`a=1; __Host-logout=x; __Host-logout=y`, "logout", "https"), true);
    assert.equal(hasCookie(`__Host-logoutx=${SECRET}; x__Host-logout=${SECRET}`, "logout", "https"), false);
    assert.equal(hasCookie(`logout=${SECRET}`, "logout", "https"), false);
    assert.equal(hasCookie(`logout=${SECRET}`, "logout", "loopback-http"), true);
    assert.equal(hasCookie("__Host-logout", "logout", "https"), false);
  });
  it("keeps the profiles' names apart and matches names exactly", () => {
    assert.equal(readCookie(`sid=${SECRET}`, "session", "https"), undefined);
    assert.equal(readCookie(`__Host-sid=${SECRET}`, "session", "loopback-http"), undefined);
    assert.equal(readCookie(`x__Host-sid=${SECRET}`, "session", "https"), undefined);
    assert.equal(readCookie(`__host-sid=${SECRET}`, "session", "https"), undefined);
    assert.equal(readCookie(`__Host-login=${SECRET}`, "session", "https"), undefined);
  });
});

describe("G062 signup cookies", () => {
  it("names the root signup credential, the tenant challenge and the form record per profile", () => {
    assert.equal(cookieName("signup", "https"), "__Host-signup");
    assert.equal(cookieName("challenge", "https"), "__Host-sct");
    assert.equal(cookieName("signup_form", "https"), "__Host-signup-form");
    assert.equal(cookieName("signup", "loopback-http"), "signup");
    assert.equal(cookieName("challenge", "loopback-http"), "sct");
    assert.equal(cookieName("signup_form", "loopback-http"), "signup-form");
  });
  it("bounds the signup credential to 1800 s and the challenge to 120 s, host-only, HttpOnly, SameSite=Lax, Secure", () => {
    assert.equal(SIGNUP_COOKIE_MAX_AGE, 1800);
    assert.equal(CHALLENGE_COOKIE_MAX_AGE, 120);
    assert.equal(setCookie("signup", SECRET, 1800, "https"), `__Host-signup=${SECRET}; Path=/; Max-Age=1800; HttpOnly; SameSite=Lax; Secure`);
    assert.equal(setCookie("challenge", SECRET, 120, "https"), `__Host-sct=${SECRET}; Path=/; Max-Age=120; HttpOnly; SameSite=Lax; Secure`);
    assert.equal(setCookie("challenge", SECRET, 120, "loopback-http"), `sct=${SECRET}; Path=/; Max-Age=120; HttpOnly; SameSite=Lax`);
    assert.throws(() => setCookie("signup", SECRET, 1801, "https"), RangeError);
    assert.throws(() => setCookie("challenge", SECRET, 121, "https"), RangeError);
    assert.equal(clearCookie("challenge", "https"), "__Host-sct=; Path=/; Max-Age=0; HttpOnly; SameSite=Lax; Secure");
    assert.equal(clearCookie("signup_form", "https"), "__Host-signup-form=; Path=/; Max-Age=0; HttpOnly; SameSite=Lax; Secure");
    assert.throws(() => setCookie("signup_form", SECRET, 60, "https"), TypeError);
    assert.equal(readCookie(`__Host-signup=${SECRET}; __Host-signup-form=v1.QQ.a`, "signup", "https"), SECRET);
  });
  it("records exactly the submitted name and slug, host-only, HttpOnly, SameSite=Lax, Secure, at most 30 minutes", () => {
    const name = "  Acme Südwerk 工場 \u{1F3ED} ";
    const cookie = setSignupFormCookie({ name, slug: "acme-sudwerk" }, "https");
    const encoded = Buffer.from(name, "utf8").toString("base64url");
    assert.equal(cookie, `__Host-signup-form=v1.${encoded}.acme-sudwerk; Path=/; Max-Age=1800; HttpOnly; SameSite=Lax; Secure`);
    assert.doesNotMatch(cookie, /domain|expires/i);
    assert.equal(setSignupFormCookie({ name: "A", slug: "" }, "loopback-http"), "signup-form=v1.QQ.; Path=/; Max-Age=1800; HttpOnly; SameSite=Lax");
    // The record round-trips byte-identical, so a retry repeats the identical payload.
    const header = `a=1; ${cookie.split(";")[0]}; b=2`;
    assert.deepEqual(readSignupFormCookie(header, "https"), { name, slug: "acme-sudwerk" });
    assert.deepEqual(readSignupFormCookie("signup-form=v1.QQ.", "loopback-http"), { name: "A", slug: "" });
  });
  it("refuses to record values outside the bounds", () => {
    for (const form of [
      { name: "", slug: "a" },
      { name: "\u0000x", slug: "a" },
      { name: "x".repeat(801), slug: "a" },
      { name: "\u{1F3ED}".repeat(201), slug: "a" },
      { name: "\uD800", slug: "a" },
      { name: "Acme", slug: "Acme" },
      { name: "Acme", slug: "-acme" },
      { name: "Acme", slug: "acme-" },
      { name: "Acme", slug: "a.b" },
      { name: "Acme", slug: "a".repeat(64) },
      { name: "Acme", slug: "a;b" },
    ]) {
      assert.equal(setSignupFormCookie(form, "https"), undefined, JSON.stringify(form).slice(0, 60));
    }
    assert.ok(setSignupFormCookie({ name: "x".repeat(800), slug: "a".repeat(63) }, "https"));
  });
  it("reads only exactly one canonical, bounded record; anything else is absent", () => {
    const ok = `v1.${Buffer.from("Acme", "utf8").toString("base64url")}.acme`;
    assert.deepEqual(readSignupFormCookie(`__Host-signup-form=${ok}`, "https"), { name: "Acme", slug: "acme" });
    for (const header of [
      null,
      "",
      `__Host-signup-form=${ok}; __Host-signup-form=${ok}`,
      `signup-form=${ok}`,
      `__Host-signup-form=v2.QWNtZQ.acme`,
      `__Host-signup-form=v1.QWNtZQ`,
      `__Host-signup-form=v1.QWNtZQ==.acme`,
      `__Host-signup-form=v1.QWNtZR.acme`, // not canonical base64url
      `__Host-signup-form=v1.QWNtZQ.Acme`,
      `__Host-signup-form=v1.QWNtZQ.acme-`,
      `__Host-signup-form=v1.QWNtZQ.${"a".repeat(64)}`,
      `__Host-signup-form=v1.${Buffer.from([0xff, 0xfe]).toString("base64url")}.acme`, // not UTF-8
      `__Host-signup-form=v1.${Buffer.from("\u0000", "utf8").toString("base64url")}.acme`,
      `__Host-signup-form=v1.${"A".repeat(1068)}.acme`,
      `__Host-signup-form=v1.${Buffer.from("x".repeat(801), "utf8").toString("base64url")}.acme`,
    ]) {
      assert.equal(readSignupFormCookie(header, "https"), undefined, String(header).slice(0, 60));
    }
  });
});

describe("ownOrigin", () => {
  const h = (host) => new Headers(host === undefined ? {} : { host });
  it("is https://<Host> for https, lower-cased, with any port", () => {
    assert.equal(ownOrigin(h("org-a.localhost:543"), "https"), "https://org-a.localhost:543");
    assert.equal(ownOrigin(h("Org-A.LocalHost"), "https"), "https://org-a.localhost");
    assert.equal(ownOrigin(h("app.example.com"), "https"), "https://app.example.com");
  });
  it("serves loopback hosts only under loopback-http", () => {
    assert.equal(ownOrigin(h("127.0.0.1:3000"), "loopback-http"), "http://127.0.0.1:3000");
    assert.equal(ownOrigin(h("localhost:3100"), "loopback-http"), "http://localhost:3100");
    for (const host of ["org-a.localhost:543", "localhost.example.com", "10.0.0.1:3000", "app.example.com"]) {
      assert.equal(ownOrigin(h(host), "loopback-http"), undefined, host);
    }
  });
  it("refuses a missing or malformed Host", () => {
    for (const host of [undefined, "", "evil.example/x", "a b", "user@host", "[::1]:3000", "host:", "host:123456", "-bad.example", "bad-.example", "a..b", `${"a".repeat(64)}.example`]) {
      assert.equal(ownOrigin(h(host), "https"), undefined, String(host));
    }
  });
  it("never reads forwarding headers", () => {
    const headers = new Headers({ host: "org-a.localhost:543", "x-forwarded-host": "evil.example", "x-forwarded-proto": "http", forwarded: "host=evil.example" });
    assert.equal(ownOrigin(headers, "https"), "https://org-a.localhost:543");
  });
});

describe("responses", () => {
  it("seeOther is a no-store 303 with every cookie kept", () => {
    const response = seeOther("/account?status=signed_in", "a=1; Path=/", "b=2; Path=/");
    assert.equal(response.status, 303);
    assert.equal(response.headers.get("location"), "/account?status=signed_in");
    assert.equal(response.headers.get("cache-control"), "no-store");
    assert.equal(response.headers.get("x-content-type-options"), "nosniff");
    assert.equal(response.headers.get("referrer-policy"), "no-referrer"); // G016 §1.7.2
    assert.deepEqual(response.headers.getSetCookie(), ["a=1; Path=/", "b=2; Path=/"]);
  });
  it("statusPage is minimal, escaped, no-store HTML", async () => {
    const response = statusPage(503, "Sign-out <failed>", "Fixed & safe \"text\"", "c=; Max-Age=0");
    assert.equal(response.status, 503);
    assert.equal(response.headers.get("content-type"), "text/html; charset=utf-8");
    assert.equal(response.headers.get("cache-control"), "no-store");
    assert.equal(response.headers.get("x-content-type-options"), "nosniff");
    assert.equal(response.headers.get("referrer-policy"), "no-referrer"); // G016 §1.7.2
    assert.deepEqual(response.headers.getSetCookie(), ["c=; Max-Age=0"]);
    const body = await response.text();
    assert.match(body, /^<!doctype html><html lang="en">/);
    assert.match(body, /<title>Sign-out &#60;failed&#62;<\/title>/);
    assert.match(body, /<h1>Sign-out &#60;failed&#62;<\/h1><p>Fixed &#38; safe &#34;text&#34;<\/p>/);
    assert.doesNotMatch(body, /<script|<link|<img|style=/);
  });
  it("405, 403 and 400 pages carry their statuses", () => {
    const response = methodNotAllowed("POST");
    assert.equal(response.status, 405);
    assert.equal(response.headers.get("allow"), "POST");
    assert.equal(methodNotAllowed("GET").headers.get("allow"), "GET");
    assert.equal(forbidden().status, 403);
    assert.equal(badRequest().status, 400);
  });
  it("jsonResponse is no-store JSON", async () => {
    const response = jsonResponse(401, { error: "unauthenticated" });
    assert.equal(response.status, 401);
    assert.equal(response.headers.get("cache-control"), "no-store");
    assert.match(response.headers.get("content-type"), /^application\/json/);
    assert.deepEqual(await response.json(), { error: "unauthenticated" });
  });
});

describe("readForm", () => {
  const good = { host: "org-a.localhost:543", origin: "https://org-a.localhost:543", "content-type": FORM };
  beforeEach(() => useProfile("test", "https"));

  it("accepts a same-origin form post and parses its fields", async () => {
    const form = await readForm(post({ ...good, cookie: `__Host-sid=${SECRET}` }, `csrf_token=${OTHER}&x=1`));
    assert.equal(form.ok, true);
    assert.equal(form.profile, "https");
    assert.equal(form.origin, "https://org-a.localhost:543");
    assert.equal(form.cookie, `__Host-sid=${SECRET}`);
    assert.equal(form.fields.get("csrf_token"), OTHER);
    const charset = await readForm(post({ ...good, "content-type": `${FORM}; charset=UTF-8` }, "a=1"));
    assert.equal(charset.ok, true);
  });
  it("refuses any other body type with 415 before looking at the origin", async () => {
    for (const type of [null, "application/json", "text/plain", "multipart/form-data; boundary=x", `${FORM}; charset=latin1`, `${FORM}, ${FORM}`, `${FORM}x`]) {
      const headers = { host: good.host, origin: "https://attacker.localhost" };
      if (type !== null) headers["content-type"] = type;
      const form = await readForm(post(headers, "{}"));
      assert.equal(form.ok, false, String(type));
      assert.equal(form.response.status, 415, String(type));
    }
  });
  it("refuses a missing, foreign, null or downgraded Origin with 403", async () => {
    for (const origin of [undefined, "null", "https://attacker.localhost", "https://org-b.localhost:543", "http://org-a.localhost:543", "https://org-a.localhost", "https://org-a.localhost:543/"]) {
      const headers = { ...good };
      if (origin === undefined) delete headers.origin;
      else headers.origin = origin;
      const form = await readForm(post(headers, "a=1"));
      assert.equal(form.ok, false, String(origin));
      assert.equal(form.response.status, 403, String(origin));
    }
  });
  it("refuses a request whose Host cannot be an origin", async () => {
    const form = await readForm(post({ ...good, host: "evil.example/x", origin: "https://evil.example/x" }, "a=1"));
    assert.equal(form.ok, false);
    assert.equal(form.response.status, 403);
  });
  it("refuses a body over 16 KiB with 400", async () => {
    const form = await readForm(post(good, `a=${"x".repeat(16 * 1024)}`));
    assert.equal(form.ok, false);
    assert.equal(form.response.status, 400);
    const edge = await readForm(post(good, `a=${"x".repeat(16 * 1024 - 2)}`));
    assert.equal(edge.ok, true);
  });
  it("keeps the 16 KiB default without options and honours an explicit maxBytes (G018)", async () => {
    for (const options of [undefined, {}, { maxBytes: undefined }]) {
      const over = await readForm(post(good, `a=${"x".repeat(16 * 1024)}`), options);
      assert.equal(over.ok, false, JSON.stringify(options));
      assert.equal(over.response.status, 400);
    }
    const max = 96 * 1024;
    const within = await readForm(post(good, `a=${"x".repeat(max - 2)}`), { maxBytes: max });
    assert.equal(within.ok, true);
    assert.equal(within.fields.get("a").length, max - 2);
    const over = await readForm(post(good, `a=${"x".repeat(max - 1)}`), { maxBytes: max });
    assert.equal(over.ok, false);
    assert.equal(over.response.status, 400);
    const small = await readForm(post(good, "a=12345"), { maxBytes: 6 });
    assert.equal(small.ok, false);
    assert.equal(small.response.status, 400);
  });
  it("uses the loopback-http profile's http origin", async () => {
    useProfile("development", undefined);
    const ok = await readForm(post({ host: "127.0.0.1:3000", origin: "http://127.0.0.1:3000", "content-type": FORM }, ""));
    assert.equal(ok.ok, true);
    assert.equal(ok.origin, "http://127.0.0.1:3000");
    const https = await readForm(post({ host: "127.0.0.1:3000", origin: "https://127.0.0.1:3000", "content-type": FORM }, ""));
    assert.equal(https.response.status, 403);
  });
  it("answers 503 for an invalid cookie profile", async () => {
    useProfile("production", "loopback-http");
    const form = await readForm(post(good, "a=1"));
    assert.equal(form.ok, false);
    assert.equal(form.response.status, 503);
  });
});

describe("field", () => {
  it("is one value, absent, or ambiguous when repeated", () => {
    const fields = new URLSearchParams("a=1&b=2&b=3&c=");
    assert.equal(field(fields, "a"), "1");
    assert.equal(field(fields, "c"), "");
    assert.equal(field(fields, "missing"), undefined);
    assert.equal(field(fields, "b"), AMBIGUOUS);
  });
});
