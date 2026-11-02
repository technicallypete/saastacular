import "server-only";

import { readMutationFields, type MutationFields } from "./branding";
import { runtimeConfig } from "./config";
import { isProfileId, PROFILE_TIMEOUT_MS, readRefusal, type Mutation, type ProfileFailure, type ProfileResult } from "./profile-client";
import { errorCategory, outboundHeaders, type RequestContext } from "./request-context";
import { badRequest, field, forbidden, seeOther, statusPage } from "./session";

// G075 (Plan r2 §3.4, §3.5, §3.7, C8, C10): the server-only client of core's two internal Group member profile routes
// (core/internal/profile/api), the selected-Group context selector and the pure helpers of the label form route. Each
// call sends exactly the request's diagnostic headers and "Authorization: Session <credential>"; the label mutation adds
// Content-Type: application/json with the documented fields only. Core alone resolves the Organization, validates the
// selector's selectability (Resolve) and decides every card and every write; nothing here is authority. Answers are
// validated against strict shapes (exact key sets) with byte bounds; anything outside the contract is "malformed", never
// a truncated or guessed list. Labels, Group names, subject ids and selector values are never logged by this module, and
// no request value is ever reflected: redirects carry closed codes and the validated selector only.

// Field maxima in Unicode code points (tenancy.RoleLabelMax; the Group name bound of tenancy names.go).
export const MAX_ROLE_LABEL = 100;
export const MAX_GROUP_NAME = 200;
// The projection's card bound (C8): at most 200 cards are ever emitted; more is the narrow_required outcome.
export const MAX_CARDS = 200;
const MAX_INT32 = 2_147_483_647;

// Success bounds (the onboarding-client arithmetic). Go's encoder writes at most 6 bytes per code point (\u escapes of <,
// >, &, controls, U+2028/9; UTF-8 is at most 4). A card: a UUID (36), a 200-code-point Group name (1,200), a
// 100-code-point label (600), "can_edit" and an int32 version with 85 bytes of keys and syntax = 1,921 ≤ 2,000. The list
// envelope {"user_id","cards":[],"narrow_required"} with the trailing newline is 86 bytes ≤ 128; a list is at most 200
// cards and 199 commas. A refusal is a closed code with at most one closed field or reason.
const CARD_MAX_BYTES = 2_000;
const ENVELOPE_BYTES = 128;
export const CARDS_MAX_BYTES = ENVELOPE_BYTES + (MAX_CARDS - 1) + MAX_CARDS * CARD_MAX_BYTES;
export const CARD_ANSWER_MAX_BYTES = ENVELOPE_BYTES + CARD_MAX_BYTES;
const REFUSAL_MAX_BYTES = 8 * 1024;

const UUID = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/;

export const CARD_KEYS: readonly string[] = ["group_id", "group_name", "role_label", "can_edit", "version"];
export const CARDS_KEYS: readonly string[] = ["user_id", "cards", "narrow_required"];

// One Group member card: the Group, its name, the member's role label there (null: none), core's edit hint (display only;
// every write is re-decided) and the label's optimistic version (0: no label row yet).
export type GroupCard = { groupId: string; groupName: string; roleLabel: string | null; canEdit: boolean; version: number };
// The subject's cards the viewer may see now. narrowRequired: more cards than the bound in Organization-only context, so
// no card is listed (no total, no cursor, no partial list).
export type GroupCards = { userId: string; cards: GroupCard[]; narrowRequired: boolean };

// ---------------------------------------------------------------------------------------------------------------------
// The selected-Group context selector (C10). It travels as ?context=<uuid> on the page and the JSON route, as the hidden
// context_group form field, and to core as context_group (query on GET, body field on POST). Absent means
// Organization-only. Present, it must be exactly one canonical lowercase UUID: an empty, malformed or repeated value is
// invalid and is refused, never dropped to Organization-only (which would broaden the request). Core's Resolve alone
// decides whether the Group is selectable; the selector only narrows.

export const CONTEXT_PARAM = "context";
export const CONTEXT_FIELD = "context_group";

export type ContextSelector = { ok: true; group: string | undefined } | { ok: false };

export function isGroupId(value: unknown): value is string {
  return typeof value === "string" && UUID.test(value);
}

// Every value of the selector as the request carried it (URLSearchParams.getAll): none is Organization-only, exactly one
// canonical UUID is that Group, anything else (an empty value, a malformed value, two or more values) is invalid.
export function readContextValues(values: readonly string[]): ContextSelector {
  if (values.length === 0) return { ok: true, group: undefined };
  if (values.length === 1 && isGroupId(values[0])) return { ok: true, group: values[0] };
  return { ok: false };
}

// The selector as a page's searchParams carry it (a repeated parameter arrives as an array).
export function readContextParam(value: string | string[] | undefined): ContextSelector {
  if (value === undefined) return { ok: true, group: undefined };
  return readContextValues(typeof value === "string" ? [value] : value);
}

// ---------------------------------------------------------------------------------------------------------------------
// Strict answer shapes.

type Json = Readonly<Record<string, unknown>>;
type Answer = { kind: "answered"; status: number; body: unknown } | { kind: "failed"; error: "timeout" | "unreachable" | "malformed" };

function isObject(value: unknown): value is Json {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

function hasExactly(o: Json, keys: readonly string[]): boolean {
  const own = Object.keys(o);
  return own.length === keys.length && keys.every((k) => Object.hasOwn(o, k));
}

function text(value: unknown, max: number): value is string {
  return typeof value === "string" && value.length > 0 && Array.from(value).length <= max;
}

function version(value: unknown): value is number {
  return typeof value === "number" && Number.isSafeInteger(value) && value >= 0 && value <= MAX_INT32;
}

// The closed reading of one card {group_id, group_name, role_label, can_edit, version} (exported for the unit tests). A
// label is null or 1–100 code points (core trims and refuses controls; re-checked loosely); a label row has been written
// at least once, so a label never comes with version 0.
export function readGroupCard(value: unknown): GroupCard | undefined {
  if (!isObject(value) || !hasExactly(value, CARD_KEYS)) return undefined;
  const { group_id: id, group_name: name, role_label: label, can_edit: canEdit, version: v } = value;
  if (typeof id !== "string" || !UUID.test(id) || !text(name, MAX_GROUP_NAME) || typeof canEdit !== "boolean" || !version(v)) return undefined;
  const roleLabel: string | null | undefined = label === null ? null : text(label, MAX_ROLE_LABEL) ? label : undefined;
  if (roleLabel === undefined || (roleLabel !== null && v < 1)) return undefined;
  return { groupId: id, groupName: name, roleLabel, canEdit, version: v };
}

// The closed reading of the card list {user_id, cards, narrow_required} for the requested subject and selector
// (exported for the unit tests): at most 200 cards with distinct Group ids; narrow_required only with no card; in a
// selected-Group context at most that one Group's card and never narrow_required.
export function readGroupCards(value: unknown, userId: string, contextGroup?: string): GroupCards | undefined {
  if (!isObject(value) || !hasExactly(value, CARDS_KEYS)) return undefined;
  const { user_id: subject, cards, narrow_required: narrow } = value;
  if (subject !== userId || !Array.isArray(cards) || cards.length > MAX_CARDS || typeof narrow !== "boolean") return undefined;
  const out: GroupCard[] = [];
  const seen = new Set<string>();
  for (const raw of cards) {
    const card = readGroupCard(raw);
    if (card === undefined || seen.has(card.groupId)) return undefined;
    seen.add(card.groupId);
    out.push(card);
  }
  if (narrow && out.length > 0) return undefined;
  if (contextGroup !== undefined && (narrow || out.length > 1 || out.some((c) => c.groupId !== contextGroup))) return undefined;
  return { userId, cards: out, narrowRequired: narrow };
}

// ---------------------------------------------------------------------------------------------------------------------
// Calls.

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

type Call = { method: "GET" | "POST"; path: string; credential: string; json?: Json; successBytes: number };

async function call(ctx: RequestContext, c: Call): Promise<Answer> {
  const headers: Record<string, string> = { ...outboundHeaders(ctx) };
  headers.Authorization = `Session ${c.credential}`;
  let body: string | undefined;
  if (c.json !== undefined) {
    headers["Content-Type"] = "application/json";
    body = JSON.stringify(c.json);
  }
  let response: Response;
  try {
    response = await fetch(new URL(c.path, runtimeConfig().coreInternalUrl), {
      method: c.method,
      headers,
      body,
      cache: "no-store",
      redirect: "error",
      signal: AbortSignal.timeout(PROFILE_TIMEOUT_MS),
    });
  } catch (error) {
    return { kind: "failed", error: errorCategory(error) === "TimeoutError" ? "timeout" : "unreachable" };
  }
  try {
    const bytes = await readBounded(response, response.status === 200 ? c.successBytes : REFUSAL_MAX_BYTES);
    if (bytes === undefined) return { kind: "failed", error: "malformed" };
    if (bytes.byteLength === 0) return { kind: "answered", status: response.status, body: undefined };
    return { kind: "answered", status: response.status, body: JSON.parse(new TextDecoder().decode(bytes)) as unknown };
  } catch (error) {
    if (errorCategory(error) === "TimeoutError") return { kind: "failed", error: "timeout" };
    return { kind: "answered", status: response.status, body: undefined };
  }
}

function failure(answer: Answer): ProfileFailure {
  if (answer.kind === "failed") return { ok: false, error: answer.error };
  return readRefusal(answer.status, answer.body);
}

// GET /internal/v1/profiles/{user_id}/groups?origin=&context_group=: the subject's cards the viewer may see now, in
// Organization-only context (contextGroup absent) or narrowed to one selected Group. An unknown, foreign, unrelated or
// revoked subject and an unselectable selector are one closed not_found (core's identical refusal). A malformed selector
// is refused here as invalid_request before any call: it never becomes Organization-only.
export async function listGroupCards(
  ctx: RequestContext,
  credential: string,
  origin: string,
  userId: string,
  contextGroup?: string,
): Promise<ProfileResult<GroupCards>> {
  if (!isProfileId(userId)) return { ok: false, error: "not_found" };
  if (contextGroup !== undefined && !isGroupId(contextGroup)) return { ok: false, error: "invalid_request", field: CONTEXT_FIELD };
  const q = new URLSearchParams({ origin });
  if (contextGroup !== undefined) q.set(CONTEXT_FIELD, contextGroup);
  const answer = await call(ctx, { method: "GET", path: `/internal/v1/profiles/${userId}/groups?${q.toString()}`, credential, successBytes: CARDS_MAX_BYTES });
  if (answer.kind !== "answered" || answer.status !== 200) return failure(answer);
  const value = readGroupCards(answer.body, userId, contextGroup);
  return value === undefined ? { ok: false, error: "malformed" } : { ok: true, value };
}

// The label mutation: the G074 mutation fields plus the validated selector (absent: Organization-only).
export type LabelMutation = Mutation & { contextGroup: string | undefined };

// POST /internal/v1/profiles/{user_id}/groups/{group_id}/label: set the subject's role label in that Group, or clear it
// with "" (core trims; whitespace clears). The value is sent as submitted; core validates it (invalid_request or
// invalid_input, field role_label) and decides the writer (the member itself, or the exact Group's own owner or admin).
// The edited Group (the path) and the selector (the body's context_group) are separate values, never substituted.
export async function setGroupLabel(
  ctx: RequestContext,
  credential: string,
  m: LabelMutation,
  userId: string,
  groupId: string,
  label: string,
): Promise<ProfileResult<GroupCard>> {
  if (!isProfileId(userId) || !isGroupId(groupId)) return { ok: false, error: "not_found" };
  if (m.contextGroup !== undefined && !isGroupId(m.contextGroup)) return { ok: false, error: "invalid_request", field: CONTEXT_FIELD };
  const json: Record<string, unknown> = { origin: m.origin, csrf_token: m.csrfToken, expected_version: m.expectedVersion, role_label: label };
  if (m.contextGroup !== undefined) json[CONTEXT_FIELD] = m.contextGroup;
  const answer = await call(ctx, {
    method: "POST",
    path: `/internal/v1/profiles/${userId}/groups/${groupId}/label`,
    credential,
    json,
    successBytes: CARD_ANSWER_MAX_BYTES,
  });
  if (answer.kind !== "answered" || answer.status !== 200) return failure(answer);
  const card = readGroupCard(answer.body);
  if (card === undefined || card.groupId !== groupId) return { ok: false, error: "malformed" };
  return { ok: true, value: card };
}

// ---------------------------------------------------------------------------------------------------------------------
// The label form route (POST /profiles/<id>/groups/<group>/label): pure helpers. POST-redirect-GET to the profile page
// with one closed group_status and the same validated selector; the page maps the code to fixed text.

// The closed outcome codes of the label form (the page's group_status parameter).
export const GROUP_STATUSES = [
  "label_saved",
  "label_cleared",
  "stale_version",
  "invalid_label",
  "invalid_context",
  "not_permitted",
  "unavailable",
] as const;
export type GroupStatus = (typeof GROUP_STATUSES)[number];

// The composite profile page of a subject, optionally in a selected-Group context and with one closed status.
export function profilePagePath(userId: string, contextGroup?: string, status?: GroupStatus): string {
  const q = new URLSearchParams();
  if (status !== undefined) q.set("group_status", status);
  if (contextGroup !== undefined) q.set(CONTEXT_PARAM, contextGroup);
  const query = q.toString();
  return `/profiles/${userId}${query === "" ? "" : `?${query}`}`;
}

// The label form's action for one card.
export function labelFormAction(userId: string, groupId: string): `/${string}` {
  return `/profiles/${userId}/groups/${groupId}/label`;
}

// The page a label form returns to, with one closed status and the request's own validated selector, at the Groups
// section.
export function groupBack(userId: string, contextGroup: string | undefined, status: GroupStatus): Response {
  return seeOther(`${profilePagePath(userId, contextGroup, status)}#groups`);
}

export type LabelForm = { m: MutationFields; label: string };

// The label form's fields: csrf_token and expected_version (readMutationFields) and exactly one role_label field ("" clears
// the label). The label is passed as submitted: core trims and validates it. A missing or repeated field, or a version
// outside the grammar, is undefined (the route answers 400). The context_group field is read separately
// (readContextValues over every value), so a repeated or malformed selector is refused rather than dropped.
export function readLabelForm(fields: URLSearchParams): LabelForm | undefined {
  const m = readMutationFields((n) => field(fields, n));
  const label = field(fields, "role_label");
  if (m === undefined || typeof label !== "string") return undefined;
  return { m, label };
}

// The closed status of a refused or unconfirmed label write that returns to the page, or undefined when the refusal is
// one of the form layer's fixed pages (exported for the unit tests).
export function groupStatusOf(failure: ProfileFailure): GroupStatus | undefined {
  switch (failure.error) {
    case "version_conflict":
      return "stale_version";
    case "invalid_request":
    case "invalid_input":
      if (failure.field === "role_label") return "invalid_label";
      return failure.field === CONTEXT_FIELD ? "invalid_context" : undefined;
    case "forbidden":
      return "not_permitted";
    case "unavailable":
    case "timeout":
    case "unreachable":
    case "malformed":
      return "unavailable";
    default:
      return undefined;
  }
}

const NOTHING_CHANGED = "Nothing was changed.";

// A label form's answer (POST-redirect-GET): success and the recoverable refusals return to the page with one closed
// status and the same selector; the session, CSRF and origin refusals are the form layer's fixed pages; an invisible
// subject, an unknown, foreign or unreadable Group, a Group outside the selected context, a subject without a current
// membership there and an unselectable selector are one fixed 404.
export function groupLabelAnswer(result: ProfileResult<GroupCard>, userId: string, contextGroup: string | undefined): Response {
  if (result.ok) return groupBack(userId, contextGroup, result.value.roleLabel === null ? "label_cleared" : "label_saved");
  const status = groupStatusOf(result);
  // A selector core itself refused as invalid is not carried back (as the route's own invalid_context).
  if (status !== undefined) return groupBack(userId, status === "invalid_context" ? undefined : contextGroup, status);
  switch (result.error) {
    case "csrf_failed":
      return statusPage(403, "Form expired", `This form is no longer valid. Reload the profile page and try again. ${NOTHING_CHANGED}`);
    case "unauthenticated":
      return statusPage(401, "Not signed in", `Sign in first, then try again. ${NOTHING_CHANGED}`);
    case "origin_rejected":
      return forbidden();
    case "not_available":
      return statusPage(403, "Not available", `This role cannot be changed from this workspace right now. ${NOTHING_CHANGED}`);
    case "not_found":
    case "unknown_host":
      return statusPage(404, "Not found", "There is nothing at this address.");
    default:
      return badRequest();
  }
}
