import "server-only";

import { runtimeConfig } from "./config";
import { errorCategory, outboundHeaders, type RequestContext } from "./request-context";

// G074 (Plan r3.1 §3.5, §3.7): the server-only client of core's internal account-profile contract
// (core/internal/profile/api; six method/path pairs, one function each). Each call sends exactly the request's diagnostic
// headers and "Authorization: Session <credential>"; the JSON mutations add Content-Type: application/json with the
// documented fields only; the avatar upload sends the raw image bytes (core validates and re-encodes them) as
// application/octet-stream with the CSRF token, expected version and origin as X-CSRF-Token, X-Expected-Version and
// X-Origin (the origin also as the G017 logo route's ?origin= query, so either reading of the convention finds exactly
// one value). Core alone resolves
// the Organization from its trusted host registry and decides every read and write (session, CSRF token, liveness,
// self-only writes, the relationship predicate on every read). Answers are validated against strict shapes (exact key
// sets) with byte bounds; anything outside the contract is "malformed". Names, subject ids and image bytes are never
// logged by this module; core's error detail never leaves it beyond its closed code, closed field and closed reason.

export const PROFILE_TIMEOUT_MS = 10_000;
const UPLOAD_TIMEOUT_MS = 30_000;
const MAX_JSON_BYTES = 8 * 1024; // §3.7: ≤ 8 KiB bodies
export const MAX_AVATAR_BYTES = 524_288; // §3.8: ≤ 512 KiB, in and re-encoded out
export const MAX_PROFILE_NAME = 200; // code points (tenancy.ProfileNameMax)
const MAX_INT32 = 2_147_483_647;

const UUID = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/;
const CODE = /^[a-z_]{1,64}$/;
const AVATAR_TYPES: ReadonlySet<string> = new Set(["image/png", "image/jpeg"]);

const PROFILE_ERRORS = [
  "invalid_request",
  "invalid_input",
  "ambiguous_parameter",
  "unsupported_media_type",
  "origin_rejected",
  "unauthenticated",
  "csrf_failed",
  "not_available",
  "forbidden",
  "not_found",
  "unknown_host",
  "version_conflict",
  "payload_too_large",
  "unavailable",
] as const;
const KNOWN_ERRORS: ReadonlySet<string> = new Set(PROFILE_ERRORS);

// The three repairable HOLD reasons of a refused avatar upload (§3.5: 503 unavailable with Retry-After). They are shown as
// one closed status (avatar_hold); the reason itself and the Retry-After value are never reflected to the browser.
export const HOLD_REASONS: readonly string[] = ["avatar_cleanup_pending", "avatar_upload_in_progress", "commit_outcome_unknown"];

export type ProfileError = (typeof PROFILE_ERRORS)[number] | "timeout" | "unreachable" | "malformed";
export type ProfileFailure = { ok: false; error: ProfileError; field?: string; reason?: string };
export type ProfileResult<T> = { ok: true; value: T } | ProfileFailure;

export type AvatarRef = { version: number; contentType: "image/png" | "image/jpeg" };
// The caller's own profile: name is the explicit override (null: none, the sign-in provider's name applies),
// effectiveName the presentation name (null only when neither exists), version the optimistic row version (0: no row).
export type OwnProfile = { userId: string; name: string | null; effectiveName: string | null; avatar: AvatarRef | null; version: number };
// Another user's profile as a related viewer sees it now: the effective name and the current avatar reference.
export type VisibleProfile = { userId: string; name: string | null; avatar: AvatarRef | null };
export type Mutation = { origin: string; csrfToken: string; expectedVersion: number };

type Json = Readonly<Record<string, unknown>>;
type Answer = { kind: "answered"; status: number; body: unknown } | { kind: "failed"; error: "timeout" | "unreachable" | "malformed" };

function isObject(value: unknown): value is Json {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

function hasExactly(o: Json, keys: readonly string[]): boolean {
  const own = Object.keys(o);
  return own.length === keys.length && keys.every((k) => Object.hasOwn(o, k));
}

export function isProfileId(value: unknown): value is string {
  return typeof value === "string" && UUID.test(value);
}

// An avatar version as a URL carries it: a canonical positive int32 (no sign, no leading zero).
export function isAvatarVersion(value: unknown): value is string {
  return typeof value === "string" && /^[1-9][0-9]{0,9}$/.test(value) && Number(value) <= MAX_INT32;
}

function count(value: unknown, min: number): value is number {
  return typeof value === "number" && Number.isSafeInteger(value) && value >= min && value <= MAX_INT32;
}

// A name as core stores or projects it: null, or 1–200 code points (core trims and refuses controls; re-checked loosely).
function name(value: unknown): value is string | null {
  return value === null || (typeof value === "string" && value.length > 0 && Array.from(value).length <= MAX_PROFILE_NAME);
}

// A presentation name: as name(), and "" (no explicit name and no sign-in provider name, as a Go string carries it) is
// read as null. undefined: outside the contract.
function presentation(value: unknown): string | null | undefined {
  if (value === "") return null;
  return name(value) ? value : undefined;
}

function avatar(value: unknown): AvatarRef | null | undefined {
  if (value === null) return null;
  if (!isObject(value) || !hasExactly(value, ["version", "content_type"]) || !count(value.version, 1) ||
    typeof value.content_type !== "string" || !AVATAR_TYPES.has(value.content_type)) {
    return undefined;
  }
  return { version: value.version, contentType: value.content_type as AvatarRef["contentType"] };
}

// The closed reading of the own view {user_id, name, effective_name, avatar, version} (exported for the unit tests).
export function readOwnProfile(value: unknown): OwnProfile | undefined {
  if (!isObject(value) || !hasExactly(value, ["user_id", "name", "effective_name", "avatar", "version"])) return undefined;
  const a = avatar(value.avatar);
  const { user_id: id, name: explicit, version } = value;
  const effective = presentation(value.effective_name);
  if (!isProfileId(id) || !name(explicit) || effective === undefined || a === undefined || !count(version, 0)) return undefined;
  // An explicit name is the presentation name; a row with an avatar has been written at least once.
  if ((explicit !== null && effective !== explicit) || (a !== null && version < 1)) return undefined;
  return { userId: id, name: explicit, effectiveName: effective, avatar: a, version };
}

// The closed reading of another user's view {user_id, name, avatar} (exported for the unit tests).
export function readVisibleProfile(value: unknown): VisibleProfile | undefined {
  if (!isObject(value) || !hasExactly(value, ["user_id", "name", "avatar"])) return undefined;
  const a = avatar(value.avatar);
  const effective = presentation(value.name);
  if (!isProfileId(value.user_id) || effective === undefined || a === undefined) return undefined;
  return { userId: value.user_id, name: effective, avatar: a };
}

async function readBounded(response: Response, maxBytes: number): Promise<Uint8Array | undefined> {
  if (response.body === null) return new Uint8Array(0);
  const reader = response.body.getReader();
  const chunks: Uint8Array[] = [];
  let size = 0;
  for (;;) {
    const { done, value } = await reader.read();
    if (done) break;
    size += value.byteLength;
    if (size > maxBytes) {
      await reader.cancel();
      return undefined;
    }
    chunks.push(value);
  }
  return new Uint8Array(Buffer.concat(chunks));
}

type Call = {
  method: "GET" | "POST";
  path: string;
  credential: string;
  json?: Json;
  raw?: { bytes: Uint8Array; m: Mutation };
};

async function call(ctx: RequestContext, c: Call): Promise<Answer> {
  const headers: Record<string, string> = { ...outboundHeaders(ctx) };
  headers.Authorization = `Session ${c.credential}`;
  let body: BodyInit | undefined;
  if (c.json !== undefined) {
    headers["Content-Type"] = "application/json";
    body = JSON.stringify(c.json);
  } else if (c.raw !== undefined) {
    headers["Content-Type"] = "application/octet-stream";
    headers["X-CSRF-Token"] = c.raw.m.csrfToken;
    headers["X-Expected-Version"] = String(c.raw.m.expectedVersion);
    headers["X-Origin"] = c.raw.m.origin;
    body = Buffer.from(c.raw.bytes);
  }
  let response: Response;
  try {
    response = await fetch(new URL(c.path, runtimeConfig().coreInternalUrl), {
      method: c.method,
      headers,
      body,
      cache: "no-store",
      redirect: "error",
      signal: AbortSignal.timeout(c.raw === undefined ? PROFILE_TIMEOUT_MS : UPLOAD_TIMEOUT_MS),
    });
  } catch (error) {
    return { kind: "failed", error: errorCategory(error) === "TimeoutError" ? "timeout" : "unreachable" };
  }
  try {
    const bytes = await readBounded(response, MAX_JSON_BYTES);
    if (bytes === undefined) return { kind: "failed", error: "malformed" };
    if (bytes.byteLength === 0) return { kind: "answered", status: response.status, body: undefined };
    return { kind: "answered", status: response.status, body: JSON.parse(new TextDecoder().decode(bytes)) as unknown };
  } catch (error) {
    if (errorCategory(error) === "TimeoutError") return { kind: "failed", error: "timeout" };
    return { kind: "answered", status: response.status, body: undefined };
  }
}

// A refusal must be {"error": <closed code>}, optionally with one closed "field" (an invalid input) or one closed
// "reason" (an unavailable answer, e.g. a HOLD). No answer, or a 5xx without a closed code, is unavailable; any other
// answer outside the contract is malformed.
export function readRefusal(status: number, body: unknown): ProfileFailure {
  if (status >= 400 && isObject(body) && typeof body.error === "string" && KNOWN_ERRORS.has(body.error)) {
    const error = body.error as ProfileError;
    if (hasExactly(body, ["error"])) return { ok: false, error };
    if (hasExactly(body, ["error", "field"]) && typeof body.field === "string" && CODE.test(body.field)) return { ok: false, error, field: body.field };
    if (hasExactly(body, ["error", "reason"]) && typeof body.reason === "string" && CODE.test(body.reason)) return { ok: false, error, reason: body.reason };
  }
  return { ok: false, error: status >= 500 ? "unavailable" : "malformed" };
}

// Whether a failure is one of the three repairable avatar HOLDs (503 unavailable with a closed HOLD reason).
export function isHold(failure: ProfileFailure): boolean {
  return failure.error === "unavailable" && failure.reason !== undefined && HOLD_REASONS.includes(failure.reason);
}

function failure(answer: Answer): ProfileFailure {
  if (answer.kind === "failed") return { ok: false, error: answer.error };
  return readRefusal(answer.status, answer.body);
}

function ownView(answer: Answer): ProfileResult<OwnProfile> {
  if (answer.kind !== "answered" || answer.status !== 200) return failure(answer);
  const value = readOwnProfile(answer.body);
  return value === undefined ? { ok: false, error: "malformed" } : { ok: true, value };
}

// GET /internal/v1/profiles/me?origin=: the caller's own profile (version 0 and the provider name before any edit).
export async function getOwnProfile(ctx: RequestContext, credential: string, origin: string): Promise<ProfileResult<OwnProfile>> {
  const answer = await call(ctx, { method: "GET", path: `/internal/v1/profiles/me?origin=${encodeURIComponent(origin)}`, credential });
  return ownView(answer);
}

// POST /internal/v1/profiles/me: set the explicit name, or clear it with "" (the provider name applies again). The
// value is sent as submitted; core trims and validates it (invalid_request, field "name").
export async function updateProfileName(ctx: RequestContext, credential: string, m: Mutation, value: string): Promise<ProfileResult<OwnProfile>> {
  const answer = await call(ctx, {
    method: "POST",
    path: "/internal/v1/profiles/me",
    credential,
    json: { origin: m.origin, csrf_token: m.csrfToken, expected_version: m.expectedVersion, name: value },
  });
  return ownView(answer);
}

// POST /internal/v1/profiles/me/avatar: the raw image (at most 512 KiB; core validates, normalizes and stores it).
export async function setAvatar(ctx: RequestContext, credential: string, m: Mutation, bytes: Uint8Array): Promise<ProfileResult<OwnProfile>> {
  if (bytes.byteLength > MAX_AVATAR_BYTES) return { ok: false, error: "payload_too_large" };
  const answer = await call(ctx, {
    method: "POST",
    path: `/internal/v1/profiles/me/avatar?origin=${encodeURIComponent(m.origin)}`,
    credential,
    raw: { bytes, m },
  });
  return ownView(answer);
}

// POST /internal/v1/profiles/me/avatar/remove: withdraw the current avatar (no avatar: a no-op).
export async function removeAvatar(ctx: RequestContext, credential: string, m: Mutation): Promise<ProfileResult<OwnProfile>> {
  const answer = await call(ctx, {
    method: "POST",
    path: "/internal/v1/profiles/me/avatar/remove",
    credential,
    json: { origin: m.origin, csrf_token: m.csrfToken, expected_version: m.expectedVersion },
  });
  return ownView(answer);
}

// GET /internal/v1/profiles/{user_id}?origin=: another user's profile when core's relationship predicate admits it now;
// unknown, foreign, unrelated and revoked subjects are one closed not_found.
export async function getProfile(ctx: RequestContext, credential: string, origin: string, userId: string): Promise<ProfileResult<VisibleProfile>> {
  if (!isProfileId(userId)) return { ok: false, error: "not_found" };
  const answer = await call(ctx, { method: "GET", path: `/internal/v1/profiles/${userId}?origin=${encodeURIComponent(origin)}`, credential });
  if (answer.kind !== "answered" || answer.status !== 200) return failure(answer);
  const value = readVisibleProfile(answer.body);
  if (value === undefined || value.userId !== userId) return { ok: false, error: "malformed" };
  return { ok: true, value };
}

export type AvatarResult =
  | { ok: true; contentType: "image/png" | "image/jpeg"; bytes: Uint8Array }
  | { ok: false; error: "not_found" | "unavailable" };

// GET /internal/v1/profiles/{user_id}/avatar?origin=&v=: the subject's current avatar bytes, only for its current version
// and only while core's predicate admits the viewer (re-decided on every read). Core's refusals (including an ended
// session) are not_found; anything outside the contract (type, length, ETag, size) is unavailable.
export async function fetchAvatar(ctx: RequestContext, credential: string, origin: string, userId: string, version: string): Promise<AvatarResult> {
  if (!isProfileId(userId) || !isAvatarVersion(version)) return { ok: false, error: "not_found" };
  const path = `/internal/v1/profiles/${userId}/avatar?origin=${encodeURIComponent(origin)}&v=${version}`;
  let response: Response;
  try {
    response = await fetch(new URL(path, runtimeConfig().coreInternalUrl), {
      method: "GET",
      headers: { ...outboundHeaders(ctx), Authorization: `Session ${credential}` },
      cache: "no-store",
      redirect: "error",
      signal: AbortSignal.timeout(PROFILE_TIMEOUT_MS),
    });
  } catch {
    return { ok: false, error: "unavailable" };
  }
  if (response.status >= 400 && response.status < 500) {
    await response.body?.cancel();
    return { ok: false, error: "not_found" };
  }
  const type = response.headers.get("content-type") ?? "";
  const length = Number(response.headers.get("content-length"));
  if (response.status !== 200 || !AVATAR_TYPES.has(type) || response.headers.get("etag") !== `"v${version}"` ||
    !Number.isSafeInteger(length) || length < 1 || length > MAX_AVATAR_BYTES) {
    await response.body?.cancel();
    return { ok: false, error: "unavailable" };
  }
  try {
    const bytes = await readBounded(response, MAX_AVATAR_BYTES);
    if (bytes === undefined || bytes.byteLength !== length) return { ok: false, error: "unavailable" };
    return { ok: true, contentType: type as "image/png" | "image/jpeg", bytes };
  } catch {
    return { ok: false, error: "unavailable" };
  }
}
