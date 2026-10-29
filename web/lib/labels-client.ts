import "server-only";

import { runtimeConfig } from "./config";
import { onboardingPage, type OnboardingError, type StatusLink } from "./onboarding-client";
import { deriveContext, errorCategory, outboundHeaders, type RequestContext } from "./request-context";
import { badRequest, field, readCookie, readForm, seeOther } from "./session";

// G020 (Plan r2 §1.5, §4.2, §4.4; D-7, D-9): the server-only client of core's internal labels contract
// (/internal/v1/labels and /internal/v1/groups/{id}/labels; seven method/path pairs on six paths, one function each) and
// the pure form helpers of the label pages and routes. Each call sends exactly the request's diagnostic headers,
// "Authorization: Session <credential>" and, on mutations, Content-Type: application/json with the documented fields
// only. Core alone resolves the Organization from its trusted host registry and decides every read and write (session,
// CSRF token, label.organization.manage, label.group.manage, label.group.read, the Organization-wide context of an
// Organization-owned deletion); nothing here is authority. Answers are validated against strict shapes (the closed key
// sets LABEL_KEYS and GROUP_LABEL_KEYS) with per-endpoint byte bounds; anything outside the contract is "malformed". The
// transport helpers (isObject, hasExactly, readBounded, call, failure, ascending) are copies of membership-client.ts,
// pinned identical by labels-client.test.mjs. Names are data: rendered as escaped text, never logged, never put in a URL.

export const LABELS_TIMEOUT_MS = 10_000;
const MAX_RESPONSE_BYTES = 64 * 1024; // every refusal and any answer without its own bound

export const MAX_LABEL_PAGE = 100; // tenancy.MaxLabelPage
export const MAX_LABEL_NAME = 80; // tenancy.MaxLabelNameLength (code points)
export const MAX_LABELS_PER_SCOPE = 200; // tenancy.MaxLabelDefinitionsPerScope

// Success bounds (the membership-client arithmetic: Go's encoder writes at most 6 bytes per code point, e.g. "<" as an
// escaped sequence). A definition row: a UUID (38), the longest scope/group pair ("group" with a UUID, 45), an
// 80-code-point name (482), the longest visibility (16) and 50 bytes of keys and syntax = 631 ≤ 640. An assignment row:
// a UUID (38), the name (482) and 21 bytes of keys and syntax = 541 ≤ 560. A definition page: 100 rows, 99 commas and a
// 128-byte envelope; the assignments of one Group are at most one per Organization-owned definition (200), unpaged.
// (Plan r2 §1.5 estimated a 400-byte row; the exact sums above supersede that estimate, as G019's did.)
export const LABEL_ROW_MAX_BYTES = 640;
export const GROUP_LABEL_ROW_MAX_BYTES = 560;
const ENVELOPE_BYTES = 128;
export const LABEL_PAGE_MAX_BYTES = ENVELOPE_BYTES + (MAX_LABEL_PAGE - 1) + MAX_LABEL_PAGE * LABEL_ROW_MAX_BYTES;
export const GROUP_LABELS_MAX_BYTES = ENVELOPE_BYTES + (MAX_LABELS_PER_SCOPE - 1) + MAX_LABELS_PER_SCOPE * GROUP_LABEL_ROW_MAX_BYTES;
export const LABEL_MAX_BYTES = ENVELOPE_BYTES + LABEL_ROW_MAX_BYTES;
const SMALL_MAX_BYTES = ENVELOPE_BYTES;

const UUID = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/;
const FIELD_CODE = /^[a-z_]{1,64}$/;
// Control, line-separator and paragraph-separator characters never appear in a stored name (tenancy.ValidateLabelName).
const NAME_FORBIDDEN = /[\p{Cc}\p{Zl}\p{Zp}]/u;
// Core stores a name after Go's strings.TrimSpace, which trims exactly the Unicode White_Space set (U+0009-000D, U+0020,
// U+0085, U+00A0, U+1680, U+2000-200A, U+2028, U+2029, U+202F, U+205F, U+3000), so a stored name never begins or ends
// with one of those. JavaScript's trim() is not that set: it also strips U+FEFF (which core keeps and counts) and does
// not strip U+0085 (a control, refused anywhere), so it is not used here (as mcp-client.ts's name mirror does not).
const NAME_EDGE_SPACE = /^\p{White_Space}|\p{White_Space}$/u;

const LABELS_ERRORS = [
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
  "labels_unavailable",
] as const;
const KNOWN_ERRORS: ReadonlySet<string> = new Set(LABELS_ERRORS);
// Codes that may carry a closed field (core's refuseErr: invalid_request, forbidden and conflict).
const FIELD_ERRORS: ReadonlySet<string> = new Set(["invalid_request", "forbidden", "conflict"]);

export type LabelsError = (typeof LABELS_ERRORS)[number] | "timeout" | "unreachable" | "malformed";
export type LabelsResult<T> = { ok: true; value: T } | { ok: false; error: LabelsError; field?: string };

export type LabelScope = "organization" | "group";
export type LabelVisibility = "owner_scope" | "target_members";
export const VISIBILITIES: readonly LabelVisibility[] = ["owner_scope", "target_members"];
export type Label = { id: string; scope: LabelScope; groupId: string | null; name: string; visibility: LabelVisibility };
export type LabelPage = { labels: Label[]; next: string | null };
export type GroupLabel = { labelId: string; name: string };
export type Deleted = { assignmentsRemoved: number };
export type Mutation = { origin: string; csrfToken: string };

// The closed DTO key sets (Plan r2 §1.5, §4.4; D-9: no existing DTO changes).
export const LABEL_KEYS = ["id", "scope", "group_id", "name", "visibility"] as const;
export const GROUP_LABEL_KEYS = ["label_id", "name"] as const;

// Who reads a definition besides the managers of its owner scope, as the pages state it.
export const VISIBILITY_LABELS: Readonly<Record<LabelScope, Readonly<Partial<Record<LabelVisibility, string>>>>> = {
  organization: { owner_scope: "Organization administrators" },
  group: { owner_scope: "Group administrators", target_members: "Group members" },
};

type Json = Readonly<Record<string, unknown>>;
type Call = { method: "GET" | "POST"; path: string; credential: string; body?: Json; successBytes: number };
type Answer = { kind: "answered"; status: number; body: unknown } | { kind: "failed"; error: "timeout" | "unreachable" | "malformed" };

function isObject(value: unknown): value is Json {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

function hasExactly(o: Json, keys: readonly string[]): boolean {
  const own = Object.keys(o);
  return own.length === keys.length && keys.every((k) => Object.hasOwn(o, k));
}

export function isId(value: unknown): value is string {
  return typeof value === "string" && UUID.test(value);
}

export function isLabelName(value: unknown): value is string {
  if (typeof value !== "string" || value.length === 0 || NAME_EDGE_SPACE.test(value) || NAME_FORBIDDEN.test(value)) return false;
  return Array.from(value).length <= MAX_LABEL_NAME;
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
      signal: AbortSignal.timeout(LABELS_TIMEOUT_MS),
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

// A refusal must be {"error": <closed code>} or, for invalid_request, forbidden and conflict, {"error", "field"}.
function failure(answer: Answer): { ok: false; error: LabelsError; field?: string } {
  if (answer.kind === "failed") return { ok: false, error: answer.error };
  const body = answer.body;
  if (answer.status >= 400 && isObject(body) && typeof body.error === "string" && KNOWN_ERRORS.has(body.error)) {
    const error = body.error as LabelsError;
    if (hasExactly(body, ["error"])) return { ok: false, error };
    if (FIELD_ERRORS.has(error) && hasExactly(body, ["error", "field"]) && typeof body.field === "string" && FIELD_CODE.test(body.field)) {
      return { ok: false, error, field: body.field };
    }
  }
  return { ok: false, error: "malformed" };
}

function ascending(ids: string[], next: unknown): next is string | null {
  for (let i = 1; i < ids.length; i++) if (ids[i] <= ids[i - 1]) return false;
  if (next === null) return true;
  return isId(next) && (ids.length === 0 || next >= ids[ids.length - 1]);
}

// One definition row of the expected owner scope: an Organization-owned row has no Group and owner_scope visibility; a
// Group-owned row names exactly the requested Group.
function label(value: unknown, scope: LabelScope, groupId: string | undefined): Label | undefined {
  if (!isObject(value) || !hasExactly(value, LABEL_KEYS)) return undefined;
  const { id, scope: s, group_id: g, name, visibility: v } = value;
  if (!isId(id) || s !== scope || !isLabelName(name) || typeof v !== "string") return undefined;
  if (scope === "organization" ? g !== null || v !== "owner_scope" : g !== groupId || !(VISIBILITIES as readonly string[]).includes(v)) return undefined;
  return { id, scope, groupId: scope === "organization" ? null : (g as string), name, visibility: v as LabelVisibility };
}

function groupLabel(value: unknown): GroupLabel | undefined {
  if (!isObject(value) || !hasExactly(value, GROUP_LABEL_KEYS)) return undefined;
  const { label_id: id, name } = value;
  return isId(id) && isLabelName(name) ? { labelId: id, name } : undefined;
}

function validScope(scope: LabelScope, groupId: string | undefined): boolean {
  return scope === "organization" ? groupId === undefined : isId(groupId);
}

// GET /internal/v1/labels: one page (up to 100, ascending by id) of the definitions of one owner scope. Organization:
// core decides label.organization.manage; Group: label.group.manage, or label.group.read for the target_members subset.
export async function listLabels(ctx: RequestContext, credential: string, origin: string, scope: LabelScope, groupId?: string, after?: string): Promise<LabelsResult<LabelPage>> {
  if (!validScope(scope, groupId) || (after !== undefined && !isId(after))) return { ok: false, error: "invalid_request" };
  const q = `origin=${encodeURIComponent(origin)}&scope=${scope}${scope === "group" ? `&group_id=${groupId}` : ""}&limit=${MAX_LABEL_PAGE}${after === undefined ? "" : `&after=${after}`}`;
  const answer = await call(ctx, { method: "GET", path: `/internal/v1/labels?${q}`, credential, successBytes: LABEL_PAGE_MAX_BYTES });
  if (answer.kind !== "answered" || answer.status !== 200) return failure(answer);
  const body = answer.body;
  if (!isObject(body) || !hasExactly(body, ["labels", "next"]) || !Array.isArray(body.labels) || body.labels.length > MAX_LABEL_PAGE) {
    return { ok: false, error: "malformed" };
  }
  const labels: Label[] = [];
  for (const item of body.labels as unknown[]) {
    const l = label(item, scope, groupId);
    if (l === undefined) return { ok: false, error: "malformed" };
    labels.push(l);
  }
  if (!ascending(labels.map((l) => l.id), body.next)) return { ok: false, error: "malformed" };
  return { ok: true, value: { labels, next: body.next } };
}

function row(answer: Answer, scope: LabelScope, groupId: string | undefined): LabelsResult<Label> {
  if (answer.kind !== "answered" || answer.status !== 200) return failure(answer);
  const l = label(answer.body, scope, groupId);
  return l === undefined ? { ok: false, error: "malformed" } : { ok: true, value: l };
}

// POST /internal/v1/labels: exactly origin, csrf_token, scope, group_id (Group scope only), name and visibility
// (owner_scope for the Organization scope); the answer is the created row of the requested owner scope.
export async function createLabel(
  ctx: RequestContext,
  credential: string,
  m: Mutation,
  scope: LabelScope,
  groupId: string | undefined,
  name: string,
  visibility: LabelVisibility,
): Promise<LabelsResult<Label>> {
  if (!validScope(scope, groupId)) return { ok: false, error: "invalid_request" };
  const body: Record<string, unknown> = { origin: m.origin, csrf_token: m.csrfToken, scope };
  if (scope === "group") body.group_id = groupId;
  body.name = name;
  body.visibility = visibility;
  const answer = await call(ctx, { method: "POST", path: "/internal/v1/labels", credential, body, successBytes: LABEL_MAX_BYTES });
  const r = row(answer, scope, groupId);
  return r.ok && r.value.visibility !== visibility ? { ok: false, error: "malformed" } : r;
}

// POST /internal/v1/labels/{id}/update: exactly origin, csrf_token, name and visibility; the answer is the row with
// that id in the expected owner scope.
export async function updateLabel(
  ctx: RequestContext,
  credential: string,
  m: Mutation,
  id: string,
  scope: LabelScope,
  groupId: string | undefined,
  name: string,
  visibility: LabelVisibility,
): Promise<LabelsResult<Label>> {
  if (!isId(id) || !validScope(scope, groupId)) return { ok: false, error: "invalid_request" };
  const answer = await call(ctx, {
    method: "POST",
    path: `/internal/v1/labels/${id}/update`,
    credential,
    body: { origin: m.origin, csrf_token: m.csrfToken, name, visibility },
    successBytes: LABEL_MAX_BYTES,
  });
  const r = row(answer, scope, groupId);
  return r.ok && (r.value.id !== id || r.value.visibility !== visibility) ? { ok: false, error: "malformed" } : r;
}

// POST /internal/v1/labels/{id}/delete: exactly origin and csrf_token; the answer reports the deletion and the number of
// assignments removed with it (complete by construction; core refuses a partial deletion).
export async function deleteLabel(ctx: RequestContext, credential: string, m: Mutation, id: string): Promise<LabelsResult<Deleted>> {
  if (!isId(id)) return { ok: false, error: "invalid_request" };
  const answer = await call(ctx, { method: "POST", path: `/internal/v1/labels/${id}/delete`, credential, body: { origin: m.origin, csrf_token: m.csrfToken }, successBytes: SMALL_MAX_BYTES });
  if (answer.kind !== "answered" || answer.status !== 200) return failure(answer);
  const b = answer.body;
  if (!isObject(b) || !hasExactly(b, ["deleted", "assignments_removed"]) || b.deleted !== true || !Number.isSafeInteger(b.assignments_removed) || (b.assignments_removed as number) < 0) {
    return { ok: false, error: "malformed" };
  }
  return { ok: true, value: { assignmentsRemoved: b.assignments_removed as number } };
}

// GET /internal/v1/groups/{id}/labels: the Organization-owned labels assigned to one Group (core decides: Organization
// owner/admin with group.read on the exact Group; anyone else forbidden or not_found). At most one row per definition.
export async function listGroupLabels(ctx: RequestContext, credential: string, origin: string, groupId: string): Promise<LabelsResult<GroupLabel[]>> {
  if (!isId(groupId)) return { ok: false, error: "invalid_request" };
  const answer = await call(ctx, { method: "GET", path: `/internal/v1/groups/${groupId}/labels?origin=${encodeURIComponent(origin)}`, credential, successBytes: GROUP_LABELS_MAX_BYTES });
  if (answer.kind !== "answered" || answer.status !== 200) return failure(answer);
  const body = answer.body;
  if (!isObject(body) || !hasExactly(body, ["labels"]) || !Array.isArray(body.labels) || body.labels.length > MAX_LABELS_PER_SCOPE) return { ok: false, error: "malformed" };
  const out: GroupLabel[] = [];
  const seen = new Set<string>();
  for (const item of body.labels as unknown[]) {
    const l = groupLabel(item);
    if (l === undefined || seen.has(l.labelId)) return { ok: false, error: "malformed" };
    seen.add(l.labelId);
    out.push(l);
  }
  return { ok: true, value: out };
}

async function assignment(ctx: RequestContext, credential: string, m: Mutation, groupId: string, labelId: string, assign: boolean): Promise<LabelsResult<null>> {
  if (!isId(groupId) || !isId(labelId)) return { ok: false, error: "invalid_request" };
  const answer = await call(ctx, {
    method: "POST",
    path: `/internal/v1/groups/${groupId}/labels/${assign ? "assign" : "unassign"}`,
    credential,
    body: { origin: m.origin, csrf_token: m.csrfToken, label_id: labelId },
    successBytes: SMALL_MAX_BYTES,
  });
  if (answer.kind !== "answered" || answer.status !== 200) return failure(answer);
  const b = answer.body;
  return isObject(b) && hasExactly(b, ["assigned"]) && b.assigned === assign ? { ok: true, value: null } : { ok: false, error: "malformed" };
}

// POST /internal/v1/groups/{id}/labels/assign: exactly origin, csrf_token and label_id; the answer is {"assigned": true}.
export function assignLabel(ctx: RequestContext, credential: string, m: Mutation, groupId: string, labelId: string): Promise<LabelsResult<null>> {
  return assignment(ctx, credential, m, groupId, labelId, true);
}

// POST /internal/v1/groups/{id}/labels/unassign: exactly origin, csrf_token and label_id; the answer is {"assigned": false}.
export function unassignLabel(ctx: RequestContext, credential: string, m: Mutation, groupId: string, labelId: string): Promise<LabelsResult<null>> {
  return assignment(ctx, credential, m, groupId, labelId, false);
}

// The frame's fixed refusal pages take the onboarding codes; the labels unavailable code is the same case.
export function labelsPageError(e: LabelsError): OnboardingError {
  return e === "labels_unavailable" ? "onboarding_unavailable" : e;
}

// --- form helpers (pure; unit-tested in labels-forms.test.mjs) --------------------------------------------------------

export const ORGANIZATION_LABELS_PATH = "/organization/labels";
export const groupLabelsPath = (groupId: string): `/${string}` => `/groups/${groupId}/labels`;

export type LabelAction =
  | { action: "create"; name: string; visibility: LabelVisibility }
  | { action: "rename" | "visibility"; labelId: string; name: string; visibility: LabelVisibility }
  | { action: "delete" | "assign" | "unassign"; labelId: string };

// The fields each action may carry besides csrf_token and action (closed: any other or repeated field is no action).
const ACTION_FIELDS: Readonly<Record<LabelScope | "assignment", Readonly<Record<string, readonly string[]>>>> = {
  organization: { create: ["name"], rename: ["label_id", "name"], delete: ["label_id"] },
  group: { create: ["name", "visibility"], rename: ["label_id", "name", "visibility"], visibility: ["label_id", "name", "visibility"], delete: ["label_id"] },
  assignment: { assign: ["label_id"], unassign: ["label_id"] },
};

// The submitted action of a label form, validated for shape only (core decides authority and every rule, including the
// name and visibility rules). An Organization-owned definition's visibility is fixed (owner_scope) and never submitted;
// a Group-owned definition's rename carries its current visibility and its visibility change its current name, because
// core's update sets both. Anything else is no action (the route answers "Request not understood").
export function readLabelAction(kind: LabelScope | "assignment", fields: URLSearchParams): LabelAction | undefined {
  const action = field(fields, "action");
  if (typeof action !== "string" || !Object.hasOwn(ACTION_FIELDS[kind], action)) return undefined;
  const allowed = new Set(["csrf_token", "action", ...ACTION_FIELDS[kind][action]]);
  for (const name of fields.keys()) if (!allowed.has(name)) return undefined;
  const values: Record<string, string> = {};
  for (const name of ACTION_FIELDS[kind][action]) {
    const v = field(fields, name);
    if (typeof v !== "string") return undefined;
    values[name] = v;
  }
  if (values.label_id !== undefined && !isId(values.label_id)) return undefined;
  let visibility: LabelVisibility = "owner_scope";
  if (values.visibility !== undefined) {
    if (!(VISIBILITIES as readonly string[]).includes(values.visibility)) return undefined;
    visibility = values.visibility as LabelVisibility;
  }
  switch (action) {
    case "create":
      return { action, name: values.name, visibility };
    case "rename":
    case "visibility":
      return { action, labelId: values.label_id, name: values.name, visibility };
    case "delete":
    case "assign":
    case "unassign":
      return { action, labelId: values.label_id };
    default:
      return undefined;
  }
}

// The closed ?status= codes of the definition pages and the message each shows; never a submitted value.
// organization_context_required is core's refusal of an Organization-owned deletion under a selected Group (Plan r2 §1.3):
// nothing was deleted.
export const LABEL_STATUS: Readonly<Record<string, string>> = {
  created: "The label was created.",
  renamed: "The label was renamed.",
  visibility_changed: "The label's visibility was changed.",
  deleted: "The label and every assignment of it were deleted.",
  organization_context_required: "Deleting an Organization label requires the Organization-wide context",
};

// The fixed sentences of the Group labels page (Plan r2 §1.5): the notice for a readable Group without a Group role, and
// the managers' statement that no production target exists for Group-owned labels.
export const LABELS_NOT_ADMINISTERED = "Labels of this Group are managed by its owners and admins";
export const LABELS_NO_TARGETS = "Group labels cannot be applied to anything yet: no Group-owned object type exists in this foundation";

// The closed ?status= codes of the Group page's Labels section.
export const GROUP_LABEL_STATUS: Readonly<Record<string, string>> = {
  assigned: "The label was assigned to this Group.",
  unassigned: "The label was removed from this Group.",
};

const STATUS_OF: Readonly<Record<LabelAction["action"], string>> = {
  create: "created",
  rename: "renamed",
  visibility: "visibility_changed",
  delete: "deleted",
  assign: "assigned",
  unassign: "unassigned",
};

// The closed field codes of core's label refusals and the fixed sentence each shows; nothing names a label.
const REFUSAL_FIELDS: Readonly<Record<string, string>> = {
  invalid_name: `Enter a label name of 1 to ${MAX_LABEL_NAME} characters, without leading or trailing spaces or control characters.`,
  invalid_visibility: "Choose one of the offered visibility options.",
  visibility_not_allowed: "Organization labels are visible to Organization administrators only.",
  name_conflict: "A label with this name already exists here. Choose a different name.",
  scope_full: `There are already ${MAX_LABELS_PER_SCOPE} labels here. Delete one before creating another.`,
  already_assigned: "This label is already assigned to this Group.",
  not_assigned: "This label is not assigned to this Group.",
  target_unsupported: "Labels cannot be applied to this.",
};

export function labelRefusalSentence(code: unknown): string | undefined {
  return typeof code === "string" && Object.hasOwn(REFUSAL_FIELDS, code) ? REFUSAL_FIELDS[code] : undefined;
}

// A core outcome that may have committed: no answer in time, no connection after sending, a malformed answer, or
// labels_unavailable.
export function labelOutcomeUnknown(error: LabelsError): boolean {
  return error === "timeout" || error === "unreachable" || error === "malformed" || error === "labels_unavailable";
}

// The no-store page of a refused label action; nothing changed in each case.
export function labelsRefusal(error: LabelsError, code: string | undefined, back: StatusLink): Response {
  const sentence = labelRefusalSentence(code);
  switch (error) {
    case "conflict":
      return onboardingPage(409, "Not changed", [sentence ?? "The current state does not allow this change.", "Nothing was changed. The current state is on the labels page."], [back]);
    case "invalid_request":
      return onboardingPage(400, "Not changed", [sentence ?? "This request could not be read.", "Nothing was changed."], [back]);
    case "unauthenticated":
      return onboardingPage(401, "Not signed in", ["Sign in first, then try again. Nothing was changed."], [{ href: "/sign-in", label: "Sign in" }]);
    case "csrf_failed":
      return onboardingPage(403, "Form expired", ["This form is no longer valid. Reload the page and try again. Nothing was changed."], [back]);
    case "origin_rejected":
      return onboardingPage(403, "Request refused", ["This request did not come from this site's own page. Nothing was changed."], [back]);
    case "not_available":
      return onboardingPage(403, "Not available", ["This is not available for your account in this workspace. Nothing was changed."], [{ href: "/account", label: "Account" }]);
    case "forbidden":
      return onboardingPage(403, "Not permitted", [sentence ?? "Your role does not permit this change.", "Nothing was changed."], [back]);
    case "not_found":
      return onboardingPage(404, "Not found", ["That item does not exist or is not available to you. Nothing was changed."], [back]);
    default:
      return onboardingPage(400, "Request not understood", ["This request could not be read. Nothing was changed."], [back]);
  }
}

// An outcome core may or may not have committed. The page never claims success and never retries.
export function labelsUnknown(back: StatusLink): Response {
  return onboardingPage(503, "Outcome unknown", ["Your request may or may not have completed. Check the current state before trying again."], [back]);
}

type Target = { kind: "organization" } | { kind: "group"; groupId: string };

// The form layer (method, form encoding, exact Origin, 16 KiB body), the closed action and its shape, the session, then
// core decides. Success is a 303 to the page with a closed status; organization_context_required is a 303 to the page
// with that closed status (nothing was deleted); any other refusal is a fixed page ("Not permitted" on forbidden); an
// outcome core may have committed is the unknown-state page, which never claims success and never retries.
async function actionRoute(request: Request, target: Target): Promise<Response> {
  const form = await readForm(request);
  if (!form.ok) return form.response;
  const csrfToken = field(form.fields, "csrf_token");
  const assignmentAction = target.kind === "group" && /^(?:assign|unassign)$/.test(String(field(form.fields, "action")));
  const act = readLabelAction(target.kind === "organization" ? "organization" : assignmentAction ? "assignment" : "group", form.fields);
  if (typeof csrfToken !== "string" || csrfToken === "" || act === undefined) return badRequest();
  const path: `/${string}` =
    target.kind === "organization" ? ORGANIZATION_LABELS_PATH : assignmentAction ? `/groups/${target.groupId}` : groupLabelsPath(target.groupId);
  const back: StatusLink = { href: path, label: assignmentAction ? "Back to the Group" : "Back to the labels page" };
  const credential = readCookie(form.cookie, "session", form.profile);
  if (credential === undefined) return labelsRefusal("unauthenticated", undefined, back);
  const ctx = deriveContext(request.headers);
  const m: Mutation = { origin: form.origin, csrfToken };
  const scope: LabelScope = target.kind;
  const groupId = target.kind === "group" ? target.groupId : undefined;
  let r: LabelsResult<unknown>;
  switch (act.action) {
    case "create":
      r = await createLabel(ctx, credential, m, scope, groupId, act.name, act.visibility);
      break;
    case "rename":
    case "visibility":
      r = await updateLabel(ctx, credential, m, act.labelId, scope, groupId, act.name, act.visibility);
      break;
    case "delete":
      r = await deleteLabel(ctx, credential, m, act.labelId);
      break;
    case "assign":
      r = await assignLabel(ctx, credential, m, groupId as string, act.labelId);
      break;
    case "unassign":
      r = await unassignLabel(ctx, credential, m, groupId as string, act.labelId);
      break;
    default:
      return badRequest();
  }
  if (r.ok) return seeOther(`${path}?status=${STATUS_OF[act.action]}`);
  if (act.action === "delete" && r.error === "invalid_request" && r.field === "organization_context_required") {
    return seeOther(`${path}?status=organization_context_required`);
  }
  return labelOutcomeUnknown(r.error) ? labelsUnknown(back) : labelsRefusal(r.error, r.field, back);
}

// POST /organization/labels/action: Organization-owned definitions (create, rename, delete).
export function labelActionRoute(request: Request): Promise<Response> {
  return actionRoute(request, { kind: "organization" });
}

// POST /groups/<uuid>/labels/action: one Group's own definitions (create, rename, visibility, delete; answered on
// /groups/<uuid>/labels) and the assignment of Organization-owned labels to it (assign, unassign; answered on
// /groups/<uuid>, the Group page's Labels section).
export function groupLabelActionRoute(request: Request, groupId: string): Promise<Response> {
  return actionRoute(request, { kind: "group", groupId });
}
