import "server-only";

import { runtimeConfig } from "./config";
import { errorCategory, outboundHeaders, type RequestContext } from "./request-context";

// G018 (Plan r5 §1.4–§1.6, §5.3–§5.4): the server-only client of core's internal onboarding contract
// (/internal/v1/groups, /internal/v1/business-profiles) and the pure form helpers of the onboarding pages and routes.
// Each call sends exactly the request's diagnostic headers, "Authorization: Session <credential>" and, on mutations,
// Content-Type: application/json with the documented fields only. Core alone resolves the Organization from its trusted
// host registry and decides every read and write (session, CSRF token, membership, capability); nothing here is
// authority. Answers are validated against strict shapes with per-endpoint byte bounds and reduced to closed outcomes;
// anything outside the contract is "malformed", never a truncated list or a guessed state.

export const ONBOARDING_TIMEOUT_MS = 10_000;
const MAX_RESPONSE_BYTES = 64 * 1024; // every refusal and any answer without its own bound

// Field maxima in Unicode code points (core/internal/tenancy names.go, slug.go, business_context.go).
export const MAX_NAME = 200;
export const MAX_SLUG = 63;
export const PROFILE_MAXIMA = { display_name: 255, legal_name: 255, description: 4000, industry: 255, website: 2048 } as const;
export type ProfileField = keyof typeof PROFILE_MAXIMA;
export const PROFILE_FIELDS: readonly ProfileField[] = ["display_name", "legal_name", "description", "industry", "website"];
export const MAX_PAGE = 100; // tenancy.MaxOnboardingPage
export const PAGE_SIZE = 50; // the web page size (Plan r5 §1.5)

// Success bounds (Plan r5 §1.5; the MCP_LIST_MAX_RESPONSE_BYTES arithmetic). Go's encoder writes at most 6 bytes per code
// point (\u escapes of <, >, &, controls, U+2028/9; UTF-8 is at most 4). A Group row: two UUIDs (72), a 200-code-point
// name (1,200), a 63-byte slug and 71 bytes of keys and syntax = 1,406 ≤ 1,500. A profile summary row: a UUID, two
// 255-code-point texts (2 × 1,530) and 45 bytes of syntax = 3,141 ≤ 3,200. A list: 100 rows, 99 commas and a 128-byte
// envelope with the next cursor and the trailing newline. A single profile: three 255-code-point texts (3 × 1,530), the
// 4,000-code-point description (24,000), the 2,048-byte ASCII website with every byte escapable (12,288) and 200 bytes
// of keys, syntax and the can_update envelope = 41,078 < 42 KiB.
// G060 (Plan r0 §4.2, D-9; r1 C2): a Group row gains self_role, can_manage and management_source — at most
// `,"self_role":"group_member"` (27), `,"can_manage":false` (19) and `,"management_source":"organization_override"` (44),
// 90 bytes — so 1,496 ≤ 1,600 (100 bytes of margin kept); a profile summary row gains `,"can_select":false` (19) and
// `,"can_update":false` (19): 3,179 ≤ 3,200.
export const GROUP_ROW_MAX_BYTES = 1_600;
export const PROFILE_ROW_MAX_BYTES = 3_200;
const ENVELOPE_BYTES = 128;
export const GROUP_LIST_MAX_BYTES = ENVELOPE_BYTES + (MAX_PAGE - 1) + MAX_PAGE * GROUP_ROW_MAX_BYTES;
export const PROFILE_LIST_MAX_BYTES = ENVELOPE_BYTES + (MAX_PAGE - 1) + MAX_PAGE * PROFILE_ROW_MAX_BYTES;
export const GROUP_MAX_BYTES = ENVELOPE_BYTES + GROUP_ROW_MAX_BYTES;
export const PROFILE_MAX_BYTES = 42 * 1024;
export const SLUG_MAX_BYTES = ENVELOPE_BYTES;

// The profile form budget (Plan r5 §1.6, review item 3). application/x-www-form-urlencoded spends at most 12 bytes per
// code point (an astral code point is four UTF-8 bytes, each percent-encoded as three; a BMP one at most nine; a
// textarea newline is submitted as %0D%0A, six). The five profile fields total 6,813 code points (≤ 81,756 bytes); field
// names, separators, mode, the identifier and the CSRF token stay within 1 KiB: ≤ 82,780 bytes < 96 KiB. Every other
// form keeps readForm's 16 KiB default.
export const FORM_BYTES_PER_CODE_POINT = 12;
export const FORM_OVERHEAD_BYTES = 1024;
export const PROFILE_FORM_BYTES = 96 * 1024;

const UUID = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/;
const SLUG = /^[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?$/;
const FIELD_CODE = /^[a-z_]{1,64}$/;

const ONBOARDING_ERRORS = [
  "invalid_request",
  "ambiguous_parameter",
  "origin_rejected",
  "unauthenticated",
  "csrf_failed",
  "not_available",
  "forbidden",
  "not_found",
  "method_not_allowed",
  "conflict",
  "onboarding_unavailable",
] as const;
const KNOWN_ERRORS: ReadonlySet<string> = new Set(ONBOARDING_ERRORS);

// A closed core code, or the call's own failure: no answer in time, no connection, or an answer outside the contract.
export type OnboardingError = (typeof ONBOARDING_ERRORS)[number] | "timeout" | "unreachable" | "malformed";
export type OnboardingResult<T> = { ok: true; value: T } | { ok: false; error: OnboardingError; field?: string };

// G060 (Plan r0 §1.6, §4.2, D-9; r1 C7): core's display hints on every Group answer. selfRole is the caller's own
// explicit Group role (null without a binding: an Organization role or the management override is never a self
// membership); canManage is core's decision of group.manage_members on the same facts and managementSource says which
// path allowed it ("group_role": the caller's own Group owner or admin role; "organization_override": the owner-set
// Organization setting; null: no management). The hints only choose what a page offers; core re-decides every request.
export type GroupRole = "group_owner" | "group_admin" | "group_member";
export type ManagementSource = "group_role" | "organization_override";
export type GroupView = {
  id: string;
  name: string;
  slug: string;
  businessProfileId: string | null;
  complete: boolean;
  selfRole: GroupRole | null;
  canManage: boolean;
  managementSource: ManagementSource | null;
};
export type GroupPage = { groups: GroupView[]; next: string | null };
// G060 (r1 C2): the listing is a metadata read; canSelect (select_for_onboarding) and canUpdate (profile.update) are
// independent display hints, so a readable row may be neither selectable nor editable.
export type ProfileSummary = { id: string; displayName: string; industry: string | null; canSelect: boolean; canUpdate: boolean };
export type ProfilePage = { profiles: ProfileSummary[]; next: string | null };
export type ProfileView = {
  id: string;
  displayName: string;
  legalName: string | null;
  description: string | null;
  industry: string | null;
  website: string | null;
};
// The editable context as submitted (optional fields: null when empty); core validates and trims.
export type ProfileInput = Record<ProfileField, string | null> & { display_name: string };
export type Mutation = { origin: string; csrfToken: string };

type Json = Readonly<Record<string, unknown>>;
type Call = { method: "GET" | "POST" | "PUT"; path: string; credential: string; body?: Json; successBytes: number };
type Answer = { kind: "answered"; status: number; body: unknown } | { kind: "failed"; error: "timeout" | "unreachable" | "malformed" };

function isObject(value: unknown): value is Json {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

function hasExactly(o: Json, keys: readonly string[]): boolean {
  const own = Object.keys(o);
  return own.length === keys.length && keys.every((k) => Object.hasOwn(o, k));
}

export function codePoints(s: string): number {
  return Array.from(s).length;
}

function text(value: unknown, max: number): value is string {
  return typeof value === "string" && value.length > 0 && codePoints(value) <= max;
}

function optionalText(value: unknown, max: number): value is string | null {
  return value === null || text(value, max);
}

async function readBounded(response: Response, maxBytes: number): Promise<string | undefined> {
  if (response.body === null) return "";
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
  return new TextDecoder().decode(Buffer.concat(chunks));
}

async function call(ctx: RequestContext, c: Call): Promise<Answer> {
  const headers: Record<string, string> = { ...outboundHeaders(ctx) };
  headers.Authorization = `Session ${c.credential}`;
  if (c.body !== undefined) headers["Content-Type"] = "application/json";
  let response: Response;
  try {
    response = await fetch(new URL(c.path, runtimeConfig().coreInternalUrl), {
      method: c.method,
      headers,
      body: c.body === undefined ? undefined : JSON.stringify(c.body),
      cache: "no-store",
      redirect: "error",
      signal: AbortSignal.timeout(ONBOARDING_TIMEOUT_MS),
    });
  } catch (error) {
    return { kind: "failed", error: errorCategory(error) === "TimeoutError" ? "timeout" : "unreachable" };
  }
  try {
    const ok = response.status >= 200 && response.status < 300;
    const body = await readBounded(response, ok ? c.successBytes : MAX_RESPONSE_BYTES);
    if (body === undefined) return { kind: "failed", error: "malformed" };
    return { kind: "answered", status: response.status, body: JSON.parse(body) as unknown };
  } catch (error) {
    if (errorCategory(error) === "TimeoutError") return { kind: "failed", error: "timeout" };
    return { kind: "failed", error: "malformed" };
  }
}

// A refusal must be {"error": <closed code>} or, for invalid_request and conflict, {"error", "field": <closed code>}.
function failure(answer: Answer): { ok: false; error: OnboardingError; field?: string } {
  if (answer.kind === "failed") return { ok: false, error: answer.error };
  const body = answer.body;
  if (answer.status >= 400 && isObject(body) && typeof body.error === "string" && KNOWN_ERRORS.has(body.error)) {
    const error = body.error as OnboardingError;
    if (hasExactly(body, ["error"])) return { ok: false, error };
    if ((error === "invalid_request" || error === "conflict") && hasExactly(body, ["error", "field"]) &&
      typeof body.field === "string" && FIELD_CODE.test(body.field)) {
      return { ok: false, error, field: body.field };
    }
  }
  return { ok: false, error: "malformed" };
}

const GROUP_KEYS = ["id", "name", "slug", "business_profile_id", "complete", "self_role", "can_manage", "management_source"];
const GROUP_ROLES: ReadonlySet<string> = new Set(["group_owner", "group_admin", "group_member"]);

function group(value: unknown): GroupView | undefined {
  if (!isObject(value) || !hasExactly(value, GROUP_KEYS)) return undefined;
  const { id, name, slug, business_profile_id: profile, complete, self_role: role, can_manage: canManage, management_source: source } = value;
  if (typeof id !== "string" || !UUID.test(id) || !text(name, MAX_NAME) || typeof slug !== "string" || !SLUG.test(slug)) return undefined;
  if (!(profile === null || (typeof profile === "string" && UUID.test(profile)))) return undefined;
  // Completion is derived from the association (D-2): complete exactly when a profile is attached.
  if (typeof complete !== "boolean" || complete !== (profile !== null)) return undefined;
  if (!(role === null || (typeof role === "string" && GROUP_ROLES.has(role))) || typeof canManage !== "boolean") return undefined;
  if (!(source === null || source === "group_role" || source === "organization_override")) return undefined;
  // G060: core decides the explicit Group path first, so a source exists exactly when management is allowed, the Group
  // path means the caller's own owner or admin role, and the override never stands in for one.
  const manager = role === "group_owner" || role === "group_admin";
  if (canManage !== (source !== null) || (source === "group_role") !== (canManage && manager) || (source === "organization_override" && manager)) {
    return undefined;
  }
  return { id, name, slug, businessProfileId: profile, complete, selfRole: role as GroupRole | null, canManage, managementSource: source };
}

function summary(value: unknown): ProfileSummary | undefined {
  if (!isObject(value) || !hasExactly(value, ["id", "display_name", "industry", "can_select", "can_update"])) return undefined;
  const { id, display_name: name, industry, can_select: canSelect, can_update: canUpdate } = value;
  if (typeof id !== "string" || !UUID.test(id) || !text(name, PROFILE_MAXIMA.display_name) || !optionalText(industry, PROFILE_MAXIMA.industry)) {
    return undefined;
  }
  if (typeof canSelect !== "boolean" || typeof canUpdate !== "boolean") return undefined;
  return { id, displayName: name, industry, canSelect, canUpdate };
}

function profile(value: unknown): ProfileView | undefined {
  if (!isObject(value) || !hasExactly(value, ["id", "display_name", "legal_name", "description", "industry", "website"])) return undefined;
  const { id, display_name: name, legal_name: legal, description, industry, website } = value;
  if (typeof id !== "string" || !UUID.test(id) || !text(name, PROFILE_MAXIMA.display_name) || !optionalText(legal, PROFILE_MAXIMA.legal_name) ||
    !optionalText(description, PROFILE_MAXIMA.description) || !optionalText(industry, PROFILE_MAXIMA.industry) ||
    !optionalText(website, PROFILE_MAXIMA.website)) {
    return undefined;
  }
  return { id, displayName: name, legalName: legal, description, industry, website };
}

// A page of ids: strictly ascending (core's ORDER BY id), at most MAX_PAGE rows, next null or a UUID not below the last row.
function ascending(ids: string[], next: unknown): next is string | null {
  for (let i = 1; i < ids.length; i++) if (ids[i] <= ids[i - 1]) return false;
  if (next === null) return true;
  return typeof next === "string" && UUID.test(next) && (ids.length === 0 || next >= ids[ids.length - 1]);
}

function pageQuery(origin: string, page: { after?: string; limit?: number; slug?: string }): string | undefined {
  if (page.after !== undefined && !UUID.test(page.after)) return undefined;
  if (page.slug !== undefined && !SLUG.test(page.slug)) return undefined;
  const limit = page.limit ?? PAGE_SIZE;
  if (!Number.isInteger(limit) || limit < 1 || limit > MAX_PAGE) return undefined;
  let q = `origin=${encodeURIComponent(origin)}&limit=${limit}`;
  if (page.after !== undefined) q += `&after=${page.after}`;
  if (page.slug !== undefined) q += `&slug=${page.slug}`;
  return q;
}

function wrapped<T>(answer: Answer, status: number, key: string, read: (value: unknown) => T | undefined): OnboardingResult<T> {
  if (answer.kind !== "answered" || answer.status !== status) return failure(answer);
  const body = answer.body;
  const value = isObject(body) && hasExactly(body, [key]) ? read(body[key]) : undefined;
  return value === undefined ? { ok: false, error: "malformed" } : { ok: true, value };
}

// GET /internal/v1/groups?origin=&limit=[&after=][&slug=]: one page of the Organization's Groups the caller may read,
// each with its derived onboarding state. A malformed cursor or slug is refused as invalid_request without calling core.
export async function listGroups(
  ctx: RequestContext,
  credential: string,
  origin: string,
  page: { after?: string; limit?: number; slug?: string } = {},
): Promise<OnboardingResult<GroupPage>> {
  const q = pageQuery(origin, page);
  if (q === undefined) return { ok: false, error: "invalid_request" };
  const answer = await call(ctx, { method: "GET", path: `/internal/v1/groups?${q}`, credential, successBytes: GROUP_LIST_MAX_BYTES });
  if (answer.kind !== "answered" || answer.status !== 200) return failure(answer);
  const body = answer.body;
  if (!isObject(body) || !hasExactly(body, ["groups", "next"]) || !Array.isArray(body.groups) || body.groups.length > MAX_PAGE) {
    return { ok: false, error: "malformed" };
  }
  const groups: GroupView[] = [];
  for (const item of body.groups) {
    const g = group(item);
    if (g === undefined) return { ok: false, error: "malformed" };
    groups.push(g);
  }
  if (!ascending(groups.map((g) => g.id), body.next)) return { ok: false, error: "malformed" };
  if (page.slug !== undefined && (groups.length > 1 || groups.some((g) => g.slug !== page.slug))) return { ok: false, error: "malformed" };
  return { ok: true, value: { groups, next: body.next } };
}

// POST /internal/v1/groups/slug-suggestion: core's suggestion for a name (an idempotent read; CSRF-checked by core).
export async function suggestSlug(ctx: RequestContext, credential: string, m: Mutation, name: string): Promise<OnboardingResult<string>> {
  const answer = await call(ctx, {
    method: "POST",
    path: "/internal/v1/groups/slug-suggestion",
    credential,
    body: { origin: m.origin, csrf_token: m.csrfToken, name },
    successBytes: SLUG_MAX_BYTES,
  });
  return wrapped(answer, 200, "slug", (v) => (typeof v === "string" && SLUG.test(v) ? v : undefined));
}

// POST /internal/v1/groups: create a Group owned by the caller (201). A taken slug is conflict (field slug).
export async function createGroup(
  ctx: RequestContext,
  credential: string,
  m: Mutation,
  input: { name: string; slug: string | null },
): Promise<OnboardingResult<GroupView>> {
  const body: Record<string, string> = { origin: m.origin, csrf_token: m.csrfToken, name: input.name };
  if (input.slug !== null) body.slug = input.slug;
  const answer = await call(ctx, { method: "POST", path: "/internal/v1/groups", credential, body, successBytes: GROUP_MAX_BYTES });
  return wrapped(answer, 201, "group", group);
}

// GET /internal/v1/groups/{id}?origin=: the Group's onboarding view.
export async function getGroup(ctx: RequestContext, credential: string, origin: string, id: string): Promise<OnboardingResult<GroupView>> {
  if (!UUID.test(id)) return { ok: false, error: "not_found" };
  const answer = await call(ctx, {
    method: "GET",
    path: `/internal/v1/groups/${id}?origin=${encodeURIComponent(origin)}`,
    credential,
    successBytes: GROUP_MAX_BYTES,
  });
  const r = wrapped(answer, 200, "group", group);
  return r.ok && r.value.id !== id ? { ok: false, error: "malformed" } : r;
}

export type Completion = { mode: "new"; profile: ProfileInput } | { mode: "existing"; businessProfileId: string };

// POST /internal/v1/groups/{id}/business-profile: complete onboarding with a new or an existing profile. A Group that
// already has a profile is conflict (field association_exists); the caller re-reads and shows the current state.
export async function completeOnboarding(
  ctx: RequestContext,
  credential: string,
  m: Mutation,
  id: string,
  choice: Completion,
): Promise<OnboardingResult<GroupView>> {
  if (!UUID.test(id)) return { ok: false, error: "not_found" };
  const body: Record<string, unknown> = { origin: m.origin, csrf_token: m.csrfToken, mode: choice.mode };
  if (choice.mode === "new") body.profile = choice.profile;
  else body.business_profile_id = choice.businessProfileId;
  const answer = await call(ctx, { method: "POST", path: `/internal/v1/groups/${id}/business-profile`, credential, body, successBytes: GROUP_MAX_BYTES });
  const r = wrapped(answer, 200, "group", group);
  return r.ok && (r.value.id !== id || !r.value.complete) ? { ok: false, error: "malformed" } : r;
}

// GET /internal/v1/business-profiles?origin=&limit=[&after=]: one page of the Business Profiles the caller may read
// (G060 r1 C2: a metadata listing with the can_select and can_update hints per row). A page may hold fewer rows than the
// limit while next is present (next is the last scanned id).
export async function listProfiles(
  ctx: RequestContext,
  credential: string,
  origin: string,
  page: { after?: string; limit?: number } = {},
): Promise<OnboardingResult<ProfilePage>> {
  const q = pageQuery(origin, page);
  if (q === undefined) return { ok: false, error: "invalid_request" };
  const answer = await call(ctx, { method: "GET", path: `/internal/v1/business-profiles?${q}`, credential, successBytes: PROFILE_LIST_MAX_BYTES });
  if (answer.kind !== "answered" || answer.status !== 200) return failure(answer);
  const body = answer.body;
  if (!isObject(body) || !hasExactly(body, ["business_profiles", "next"]) || !Array.isArray(body.business_profiles) ||
    body.business_profiles.length > MAX_PAGE) {
    return { ok: false, error: "malformed" };
  }
  const profiles: ProfileSummary[] = [];
  for (const item of body.business_profiles) {
    const p = summary(item);
    if (p === undefined) return { ok: false, error: "malformed" };
    profiles.push(p);
  }
  if (!ascending(profiles.map((p) => p.id), body.next)) return { ok: false, error: "malformed" };
  return { ok: true, value: { profiles, next: body.next } };
}

// G060 (correction02 §P): the onboarding picker's rows — the validated page filtered to core's can_select hint. The
// cursor is never altered: next (and the First/Next links built from it) stays the unfiltered page's, so a readable page
// with no selectable row still leads to the following page.
export function selectableProfiles(page: ProfilePage): ProfilePage {
  return { profiles: page.profiles.filter((p) => p.canSelect), next: page.next };
}

// GET /internal/v1/business-profiles/{id}?origin=: the shared profile and the can_update display hint (authorizes
// nothing: every update is re-decided by core).
export async function getProfile(
  ctx: RequestContext,
  credential: string,
  origin: string,
  id: string,
): Promise<OnboardingResult<{ profile: ProfileView; canUpdate: boolean }>> {
  if (!UUID.test(id)) return { ok: false, error: "not_found" };
  const answer = await call(ctx, {
    method: "GET",
    path: `/internal/v1/business-profiles/${id}?origin=${encodeURIComponent(origin)}`,
    credential,
    successBytes: PROFILE_MAX_BYTES,
  });
  if (answer.kind !== "answered" || answer.status !== 200) return failure(answer);
  const body = answer.body;
  if (!isObject(body) || !hasExactly(body, ["business_profile", "can_update"]) || typeof body.can_update !== "boolean") {
    return { ok: false, error: "malformed" };
  }
  const p = profile(body.business_profile);
  if (p === undefined || p.id !== id) return { ok: false, error: "malformed" };
  return { ok: true, value: { profile: p, canUpdate: body.can_update } };
}

// PUT /internal/v1/business-profiles/{id}: replace the editable context (last commit wins; disclosed on the page).
export async function updateProfile(
  ctx: RequestContext,
  credential: string,
  m: Mutation,
  id: string,
  input: ProfileInput,
): Promise<OnboardingResult<ProfileView>> {
  if (!UUID.test(id)) return { ok: false, error: "not_found" };
  const answer = await call(ctx, {
    method: "PUT",
    path: `/internal/v1/business-profiles/${id}`,
    credential,
    body: { origin: m.origin, csrf_token: m.csrfToken, profile: input },
    successBytes: PROFILE_MAX_BYTES,
  });
  const r = wrapped(answer, 200, "business_profile", profile);
  return r.ok && r.value.id !== id ? { ok: false, error: "malformed" } : r;
}

export function isId(value: unknown): value is string {
  return typeof value === "string" && UUID.test(value);
}

export function isSlug(value: unknown): value is string {
  return typeof value === "string" && SLUG.test(value);
}

// --- form helpers (pure; unit-tested in onboarding-forms.test.mjs) ----------------------------------------------------

type Get = (name: string) => string | undefined | symbol;

// The five profile fields of a submitted form: each at most once (a repeated field is no input), the description's
// browser newline normalization (CRLF) undone so the stored value is what was typed; an empty optional field is null.
export function readProfileInput(get: Get): ProfileInput | undefined {
  const out: Partial<Record<ProfileField, string | null>> = {};
  for (const name of PROFILE_FIELDS) {
    const raw = get(name);
    if (typeof raw === "symbol") return undefined;
    let value = raw ?? "";
    if (name === "description") value = value.replace(/\r\n/g, "\n");
    out[name] = name === "display_name" || value !== "" ? value : null;
  }
  return out as ProfileInput;
}

// Labels, hints and the message of each core field code, keyed by the form field it belongs to. A page renders the
// message next to its field (id fieldErrorId(field), referenced by the field's aria-describedby) and focuses that field.
export const FIELD_LABELS: Readonly<Record<ProfileField | "name" | "slug", string>> = {
  name: "Group name",
  slug: "Group address (slug)",
  display_name: "Display name",
  legal_name: "Legal name",
  description: "Description",
  industry: "Industry",
  website: "Website",
};

const FIELD_ERRORS: Readonly<Record<string, { field: ProfileField | "name" | "slug"; message: string }>> = {
  invalid_name: { field: "name", message: "Enter a Group name of 1 to 200 characters without control characters." },
  invalid_slug: { field: "slug", message: "Use 1 to 63 lowercase letters, digits or hyphens, starting and ending with a letter or digit." },
  slug_reserved: { field: "slug", message: "That address is reserved. Choose another one." },
  // G018 M1 (D3): web-local, never sent by core: the create route's code for core's definitive conflict on the address.
  slug_taken: { field: "slug", message: "That address is already in use in this workspace. Choose another one." },
  invalid_display_name: { field: "display_name", message: "Enter a display name of 1 to 255 characters without control characters." },
  invalid_legal_name: { field: "legal_name", message: "A legal name may have up to 255 characters without control characters." },
  invalid_description: { field: "description", message: "A description may have up to 4,000 characters; tabs and line breaks are allowed." },
  invalid_industry: { field: "industry", message: "An industry may have up to 255 characters without control characters." },
  invalid_website: { field: "website", message: "Enter a full web address starting with https:// or http://, such as https://example.com." },
};

export function fieldError(code: unknown): { field: ProfileField | "name" | "slug"; message: string } | undefined {
  return typeof code === "string" && Object.hasOwn(FIELD_ERRORS, code) ? FIELD_ERRORS[code] : undefined;
}

export function fieldErrorId(field: string): string {
  return `${field}-error`;
}

// G060 (r1 C6): the notice of a readable incomplete Group whose onboarding the viewer does not administer (no Resume
// link, no forms; core's group.complete_onboarding refuses a forged completion).
export const ONBOARDING_NOT_ADMINISTERED = "Onboarding is administered by this Group's owners and admins.";

// G018 M1 (D3): the inline conflict of a create whose address is taken. Core's conflict is definitive (nothing was
// created by that request); the holder sentence appears only for a Group a fresh authorized read returns.
export const CREATE_CONFLICT = {
  summary: "Nothing was created by this request.",
  holder: "A Group you can see already uses it:",
  interrupted: "If an earlier attempt was interrupted, this may be that Group.",
} as const;

// G018 M1 (D2): the advisory address hints of the create form, by the closed state of GET /groups/address-check. A
// filtered absence is "not seen", never availability: the create itself decides uniqueness.
export const ADDRESS_ADVISORY: Readonly<Record<"in_use" | "not_seen" | "invalid", string>> = {
  in_use: "A Group you can see already uses this address. Choose another one.",
  not_seen: "No Group you can see uses this address. The final check happens when you create the Group.",
  invalid: FIELD_ERRORS.invalid_slug.message,
};

// --- status pages of the onboarding routes ----------------------------------------------------------------------------

const NO_STORE: Readonly<Record<string, string>> = {
  "Cache-Control": "no-store",
  "X-Content-Type-Options": "nosniff",
  "Referrer-Policy": "no-referrer",
};

function escapeHtml(s: string): string {
  return s.replace(/[&<>"']/g, (c) => `&#${c.charCodeAt(0)};`);
}

export type StatusLink = { href: `/${string}`; label: string };

// A minimal no-store HTML page for a route outcome, with the given links (escaped; hrefs are site paths built from
// validated identifiers). Values shown are core's validated answer or the validated request; nothing else is reflected.
export function onboardingPage(status: number, title: string, paragraphs: readonly string[], links: readonly StatusLink[]): Response {
  const body =
    `<!doctype html><html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1">` +
    `<title>${escapeHtml(title)}</title></head><body><main style="overflow-wrap:anywhere"><h1>${escapeHtml(title)}</h1>` +
    paragraphs.map((p) => `<p>${escapeHtml(p)}</p>`).join("") +
    `<p>${links.map((l) => `<a href="${escapeHtml(l.href)}">${escapeHtml(l.label)}</a>`).join(" · ")}</p></main></body></html>`;
  return new Response(body, { status, headers: { ...NO_STORE, "Content-Type": "text/html; charset=utf-8" } });
}

const GROUPS: StatusLink = { href: "/groups", label: "Your Groups" };

// An outcome core may or may not have committed (no answer, a malformed answer or onboarding_unavailable). The page
// never claims success and never retries; the person checks the current state.
export function unknownOutcomePage(what: "group" | "completion" | "update", id?: string): Response {
  if (what === "group") {
    return onboardingPage(503, "Outcome unknown", ["Your request may or may not have completed. Check your Groups before trying again."], [GROUPS]);
  }
  if (what === "completion" && id !== undefined) {
    return onboardingPage(503, "Outcome unknown", ["Your request may or may not have completed. Open the Group to see its current state."], [
      { href: `/groups/${id}`, label: "Open the Group" },
    ]);
  }
  return onboardingPage(
    503,
    "Outcome unknown",
    ["Your change may or may not have been saved. Open the Business Profile to see its current values."],
    id === undefined ? [{ href: "/business-profiles", label: "Your Business Profiles" }] : [{ href: `/business-profiles/${id}`, label: "Open the Business Profile" }],
  );
}

// The fixed refusal pages shared by the three onboarding form routes (nothing changed in each case).
export function refusalPage(error: OnboardingError, back: StatusLink): Response {
  switch (error) {
    case "unauthenticated":
      return onboardingPage(401, "Not signed in", ["Sign in first, then try again. Nothing was changed."], [{ href: "/sign-in", label: "Sign in" }]);
    case "csrf_failed":
      return onboardingPage(403, "Form expired", ["This form is no longer valid. Reload the page and try again. Nothing was changed."], [back]);
    case "origin_rejected":
      return onboardingPage(403, "Request refused", ["This request did not come from this site's own page. Nothing was changed."], [back]);
    case "not_available":
      return onboardingPage(403, "Not available", ["This is not available for your account in this workspace. Nothing was changed."], [
        { href: "/account", label: "Account" },
      ]);
    case "forbidden":
      return onboardingPage(403, "Not permitted", ["Your role does not permit this change. Nothing was changed."], [back]);
    case "not_found":
      return onboardingPage(404, "Not found", ["That item does not exist or is not available to you. Nothing was changed."], [GROUPS]);
    default:
      return onboardingPage(400, "Request not understood", ["This request could not be read. Nothing was changed."], [back]);
  }
}

// A core outcome that may have committed: no answer in time, no connection after sending, a malformed answer, or
// onboarding_unavailable (Plan r5 §5.3: possibly durable).
export function outcomeUnknown(error: OnboardingError): boolean {
  return error === "timeout" || error === "unreachable" || error === "malformed" || error === "onboarding_unavailable";
}
