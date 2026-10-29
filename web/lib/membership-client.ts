import "server-only";

import { runtimeConfig } from "./config";
import { onboardingPage, type StatusLink } from "./onboarding-client";
import { deriveContext, errorCategory, outboundHeaders, type RequestContext } from "./request-context";
import { badRequest, field, readCookie, readForm, seeOther } from "./session";

// G019 (Plan r1 §1.6, §5.3–§5.4; r2 addendum §7): the server-only client of core's internal membership contract
// (/internal/v1/members, /internal/v1/ownership/transfer, /internal/v1/invitations, /internal/v1/account/
// retirement-eligibility; sixteen method/path pairs, one function each; G060 adds the Organization's Group-authority
// read and owner-only write on /internal/v1/organization/group-authority: eighteen) and the pure form helpers of the
// membership pages and routes. Each call sends exactly the request's diagnostic headers, "Authorization: Session <credential>" and,
// on mutations, Content-Type: application/json with the documented fields only. Core alone resolves the Organization
// from its trusted host registry and decides every read and write (session, CSRF token, membership, capability, owner
// rules); the can_* values are display hints from core's snapshot and authorize nothing. Answers are validated against
// strict shapes with per-endpoint byte bounds; anything outside the contract is "malformed". The invitation token is
// returned to the caller once and is never logged, put in a URL, a Location header or a page title by this module.

export const MEMBERSHIP_TIMEOUT_MS = 10_000;
const MAX_RESPONSE_BYTES = 64 * 1024; // every refusal and any answer without its own bound

export const MAX_PAGE = 100; // tenancy.MaxMembersPage
export const PAGE_SIZE = 50;
export const MAX_DISPLAY_NAME = 200; // app.users display_name CHECK (code points)
export const MAX_ADDRESS_BYTES = 254; // tenancy maxAddressBytes
// The closed invitation lifetimes offered by the forms (tenancy: minimum 1 h, default 168 h, maximum 720 h).
export const LIFETIME_HOURS = [24, 168, 720] as const;
export const DEFAULT_LIFETIME_HOURS = 168;

// Success bounds (the onboarding-client arithmetic: Go writes at most 6 bytes per code point or byte when escaping). A
// member row: a UUID (38), a 200-code-point name (1,202), the longest role (14) and 169 bytes of keys and syntax = 1,423
// ≤ 1,500. An invitation row: four UUIDs (152), the issuer name (1,202), a 254-byte address with every byte escapable
// (1,526), the timestamp, scope, role and status (60) and about 190 bytes of keys and syntax = 3,130 ≤ 3,200. A candidate
// row: 1,268 ≤ 1,300. An eligibility scope: two UUIDs and the scope type, 136 ≤ 160. A list: 100 rows, 99 commas and a
// 128-byte envelope. (Plan r1 §5.4 estimated 1,400 and 2 KiB; the exact sums above supersede those estimates.)
export const MEMBER_ROW_MAX_BYTES = 1_500;
export const INVITATION_ROW_MAX_BYTES = 3_200;
export const CANDIDATE_ROW_MAX_BYTES = 1_300;
export const ELIGIBILITY_SCOPE_MAX_BYTES = 160;
const ENVELOPE_BYTES = 128;
const list = (row: number) => ENVELOPE_BYTES + (MAX_PAGE - 1) + MAX_PAGE * row;
export const MEMBER_LIST_MAX_BYTES = list(MEMBER_ROW_MAX_BYTES);
export const INVITATION_LIST_MAX_BYTES = list(INVITATION_ROW_MAX_BYTES);
export const CANDIDATE_LIST_MAX_BYTES = list(CANDIDATE_ROW_MAX_BYTES);
export const ELIGIBILITY_MAX_BYTES = list(ELIGIBILITY_SCOPE_MAX_BYTES);
export const MEMBER_MAX_BYTES = ENVELOPE_BYTES + MEMBER_ROW_MAX_BYTES;
export const INVITATION_MAX_BYTES = ENVELOPE_BYTES + INVITATION_ROW_MAX_BYTES;
export const ISSUED_MAX_BYTES = ENVELOPE_BYTES + INVITATION_ROW_MAX_BYTES + 64;
const SMALL_MAX_BYTES = ENVELOPE_BYTES;

const UUID = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/;
const TOKEN = /^[A-Za-z0-9_-]{43}$/; // 32 random bytes, base64url without padding
const INSTANT = /^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$/;
const FIELD_CODE = /^[a-z_]{1,64}$/;

const MEMBERSHIP_ERRORS = [
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
  "membership_unavailable",
] as const;
const KNOWN_ERRORS: ReadonlySet<string> = new Set(MEMBERSHIP_ERRORS);
// Codes that may carry a closed field (core's refuseErr: invalid_request, forbidden and conflict).
const FIELD_ERRORS: ReadonlySet<string> = new Set(["invalid_request", "forbidden", "conflict"]);

export type MembershipError = (typeof MEMBERSHIP_ERRORS)[number] | "timeout" | "unreachable" | "malformed";
export type MembershipResult<T> = { ok: true; value: T } | { ok: false; error: MembershipError; field?: string };

export type ScopeKind = "organization" | "group";
export type MemberScope = { kind: "organization" } | { kind: "group"; groupId: string };
export const ROLES = {
  organization: { owner: "org_owner", admin: "org_admin", member: "org_member" },
  group: { owner: "group_owner", admin: "group_admin", member: "group_member" },
} as const;
export type Role = (typeof ROLES)[ScopeKind][keyof (typeof ROLES)[ScopeKind]];
export type OrdinaryRole = "org_admin" | "org_member" | "group_admin" | "group_member";
// The client-facing role labels: one closed map, the only place a role name becomes text (no label configuration
// exists; Plan r1 §1.7 adds none). Pages never render the raw role code.
export const ROLE_LABELS: Readonly<Record<Role, string>> = {
  org_owner: "Owner",
  org_admin: "Admin",
  org_member: "Member",
  group_owner: "Owner",
  group_admin: "Admin",
  group_member: "Member",
};

export type Member = {
  userId: string;
  displayName: string | null;
  role: Role;
  self: boolean;
  disabled: boolean;
  canChangeRole: boolean;
  canRemove: boolean;
  canPromote: boolean;
  canDemote: boolean;
  canTransfer: boolean;
};
export type MemberPage = { members: Member[]; next: string | null };
export type SelfMember = { userId: string; displayName: string | null; role: Role; canLeave: boolean; lastOwner: boolean; canManage: boolean };
export type InvitationStatus = "pending" | "expired" | "revoked" | "consumed";
export type Invitation = {
  id: string;
  scope: ScopeKind;
  groupId: string | null;
  // An invitation grants only the same scope's admin or member role (Spec; Plan r1 §1.5): never an owner role.
  role: OrdinaryRole;
  issuerDisplayName: string | null;
  intendedUserId: string | null;
  // Present only in the issuing scope's administrator view; never in the recipient view.
  intendedAddress: string | null;
  expiresAt: string;
  status: InvitationStatus;
  consumedBy: string | null;
};
export type InvitationPage = { invitations: Invitation[]; next: string | null };
export type Candidate = { userId: string; displayName: string | null };
export type CandidatePage = { candidates: Candidate[]; next: string | null };
export type Issued = { invitation: Invitation; token: string };
export type AcceptOutcome = "joined" | "upgraded" | "already_member" | "consumed";
export type Accepted = { outcome: AcceptOutcome; groupId: string | null };
// The Account notice reads only the count and the truncation flag; scope ids are validated and dropped.
export type Eligibility = { count: number; truncated: boolean };
export type Mutation = { origin: string; csrfToken: string };
export type Recipient = { userId: string } | { address: string };

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

function optionalId(value: unknown): value is string | null {
  return value === null || isId(value);
}

function displayName(value: unknown): value is string | null {
  return value === null || (typeof value === "string" && value.length > 0 && Array.from(value).length <= MAX_DISPLAY_NAME);
}

export function roleOf(kind: ScopeKind, value: unknown): Role | undefined {
  return (Object.values(ROLES[kind]) as Role[]).find((r) => r === value);
}

export function isOrdinaryRole(kind: ScopeKind, value: unknown): value is OrdinaryRole {
  return value === ROLES[kind].admin || value === ROLES[kind].member;
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
      signal: AbortSignal.timeout(MEMBERSHIP_TIMEOUT_MS),
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
function failure(answer: Answer): { ok: false; error: MembershipError; field?: string } {
  if (answer.kind === "failed") return { ok: false, error: answer.error };
  const body = answer.body;
  if (answer.status >= 400 && isObject(body) && typeof body.error === "string" && KNOWN_ERRORS.has(body.error)) {
    const error = body.error as MembershipError;
    if (hasExactly(body, ["error"])) return { ok: false, error };
    if (FIELD_ERRORS.has(error) && hasExactly(body, ["error", "field"]) && typeof body.field === "string" && FIELD_CODE.test(body.field)) {
      return { ok: false, error, field: body.field };
    }
  }
  return { ok: false, error: "malformed" };
}

const MEMBER_KEYS = ["user_id", "display_name", "role", "self", "disabled", "can_change_role", "can_remove", "can_promote", "can_demote", "can_transfer"];

function member(kind: ScopeKind, value: unknown): Member | undefined {
  if (!isObject(value) || !hasExactly(value, MEMBER_KEYS)) return undefined;
  const role = roleOf(kind, value.role);
  const flags = MEMBER_KEYS.slice(3).map((k) => value[k]);
  if (!isId(value.user_id) || !displayName(value.display_name) || role === undefined || !flags.every((f) => typeof f === "boolean")) return undefined;
  const [self, disabled, canChangeRole, canRemove, canPromote, canDemote, canTransfer] = flags as boolean[];
  return { userId: value.user_id, displayName: value.display_name, role, self, disabled, canChangeRole, canRemove, canPromote, canDemote, canTransfer };
}

function selfMember(kind: ScopeKind, value: unknown): SelfMember | undefined {
  if (!isObject(value) || !hasExactly(value, ["user_id", "display_name", "role", "self", "can_leave", "last_owner", "can_manage"])) return undefined;
  const role = roleOf(kind, value.role);
  const { user_id: id, display_name: name, self, can_leave: canLeave, last_owner: lastOwner, can_manage: canManage } = value;
  if (!isId(id) || !displayName(name) || role === undefined || self !== true || typeof canLeave !== "boolean" || typeof lastOwner !== "boolean" ||
    typeof canManage !== "boolean") {
    return undefined;
  }
  // A last owner can never leave (Spec R1): core's hints must agree.
  if (lastOwner && (canLeave || role !== ROLES[kind].owner)) return undefined;
  return { userId: id, displayName: name, role, canLeave, lastOwner, canManage };
}

const INVITATION_KEYS = ["id", "scope", "group_id", "role", "issuer_display_name", "intended_user_id", "expires_at", "status", "consumed_by"];
const STATUSES: ReadonlySet<string> = new Set(["pending", "expired", "revoked", "consumed"]);

function invitation(value: unknown): Invitation | undefined {
  if (!isObject(value)) return undefined;
  const withAddress = Object.hasOwn(value, "intended_address");
  if (!hasExactly(value, withAddress ? [...INVITATION_KEYS, "intended_address"] : INVITATION_KEYS)) return undefined;
  const { id, scope, group_id: groupId, issuer_display_name: issuer, intended_user_id: intended, expires_at: expiresAt, status, consumed_by: consumedBy } = value;
  if (scope !== "organization" && scope !== "group") return undefined;
  // Strict: only the same scope's admin or member role; an owner role (or the other scope's role) is malformed.
  const role = value.role;
  if (!isId(id) || !isOrdinaryRole(scope, role) || !optionalId(groupId) || (scope === "group") !== (groupId !== null)) return undefined;
  if (!displayName(issuer) || !optionalId(intended) || !optionalId(consumedBy)) return undefined;
  if (typeof expiresAt !== "string" || !INSTANT.test(expiresAt) || Number.isNaN(Date.parse(expiresAt))) return undefined;
  if (typeof status !== "string" || !STATUSES.has(status) || (status === "consumed") !== (consumedBy !== null)) return undefined;
  let address: string | null = null;
  if (withAddress) {
    const a = value.intended_address;
    if (typeof a !== "string" || a.length === 0 || Buffer.byteLength(a) > MAX_ADDRESS_BYTES) return undefined;
    address = a;
  }
  // Exactly one recipient form is stored; the recipient view omits the address, so an absent address is allowed. G058
  // (scope amendment05): consuming an address-addressed Organization invitation records the consumer as intended_user_id
  // (00015 invitations_consumed_is_intended), so both forms appear only on a consumed Organization row whose intended
  // user is its consumer.
  if (intended !== null && address !== null && !(scope === "organization" && status === "consumed" && intended === consumedBy)) return undefined;
  return {
    id,
    scope,
    groupId,
    role,
    issuerDisplayName: issuer,
    intendedUserId: intended,
    intendedAddress: address,
    expiresAt,
    status: status as InvitationStatus,
    consumedBy,
  };
}

function candidate(value: unknown): Candidate | undefined {
  if (!isObject(value) || !hasExactly(value, ["user_id", "display_name"]) || !isId(value.user_id) || !displayName(value.display_name)) return undefined;
  return { userId: value.user_id, displayName: value.display_name };
}

// A page of ids: strictly ascending (core's keyset order), at most MAX_PAGE rows, next null or a UUID not below the last.
function ascending(ids: string[], next: unknown): next is string | null {
  for (let i = 1; i < ids.length; i++) if (ids[i] <= ids[i - 1]) return false;
  if (next === null) return true;
  return isId(next) && (ids.length === 0 || next >= ids[ids.length - 1]);
}

type Page = { after?: string; limit?: number };

function pageQuery(origin: string, page: Page): string | undefined {
  if (page.after !== undefined && !isId(page.after)) return undefined;
  const limit = page.limit ?? PAGE_SIZE;
  if (!Number.isInteger(limit) || limit < 1 || limit > MAX_PAGE) return undefined;
  return `origin=${encodeURIComponent(origin)}&limit=${limit}${page.after === undefined ? "" : `&after=${page.after}`}`;
}

function scopeQuery(scope: MemberScope): string | undefined {
  if (scope.kind === "organization") return "&scope=organization";
  return isId(scope.groupId) ? `&scope=group&group_id=${scope.groupId}` : undefined;
}

function scopeBody(m: Mutation, scope: MemberScope): Record<string, unknown> {
  const body: Record<string, unknown> = { origin: m.origin, csrf_token: m.csrfToken, scope: scope.kind };
  if (scope.kind === "group") body.group_id = scope.groupId;
  return body;
}

function validScope(scope: MemberScope): boolean {
  return scope.kind === "organization" || isId(scope.groupId);
}

function rows<T>(answer: Answer, key: string, read: (v: unknown) => T | undefined, id: (t: T) => string): MembershipResult<{ rows: T[]; next: string | null }> {
  if (answer.kind !== "answered" || answer.status !== 200) return failure(answer);
  const body = answer.body;
  if (!isObject(body) || !hasExactly(body, [key, "next"]) || !Array.isArray(body[key])) return { ok: false, error: "malformed" };
  const items = body[key] as unknown[];
  if (items.length > MAX_PAGE) return { ok: false, error: "malformed" };
  const out: T[] = [];
  for (const item of items) {
    const v = read(item);
    if (v === undefined) return { ok: false, error: "malformed" };
    out.push(v);
  }
  if (!ascending(out.map(id), body.next)) return { ok: false, error: "malformed" };
  return { ok: true, value: { rows: out, next: body.next } };
}

function wrapped<T>(answer: Answer, status: number, key: string, read: (value: unknown) => T | undefined): MembershipResult<T> {
  if (answer.kind !== "answered" || answer.status !== status) return failure(answer);
  const body = answer.body;
  const value = isObject(body) && hasExactly(body, [key]) ? read(body[key]) : undefined;
  return value === undefined ? { ok: false, error: "malformed" } : { ok: true, value };
}

function empty(answer: Answer): MembershipResult<null> {
  if (answer.kind !== "answered" || answer.status !== 200) return failure(answer);
  return isObject(answer.body) && hasExactly(answer.body, []) ? { ok: true, value: null } : { ok: false, error: "malformed" };
}

// GET /internal/v1/members: one page of the scope's members (administrators; role=owner_candidates pages the
// eligible promotion and transfer recipients).
export async function listMembers(
  ctx: RequestContext,
  credential: string,
  origin: string,
  scope: MemberScope,
  page: Page & { ownerCandidates?: boolean } = {},
): Promise<MembershipResult<MemberPage>> {
  const q = pageQuery(origin, page);
  const s = scopeQuery(scope);
  if (q === undefined || s === undefined) return { ok: false, error: "invalid_request" };
  const role = page.ownerCandidates === true ? "&role=owner_candidates" : "";
  const answer = await call(ctx, { method: "GET", path: `/internal/v1/members?${q}${s}${role}`, credential, successBytes: MEMBER_LIST_MAX_BYTES });
  const r = rows(answer, "members", (v) => member(scope.kind, v), (m) => m.userId);
  return r.ok ? { ok: true, value: { members: r.value.rows, next: r.value.next } } : r;
}

// The owner's Transfer picker (Plan r1 §1.6/C5: a paged eligible-admin picker): one bounded page of core's
// owner_candidates list per request, the next page reached by a link carrying core's cursor (owners_after); never a
// fetch-all loop. The cursor is only a position: core re-decides the chosen recipient when the transfer is posted.
export const OWNER_CANDIDATE_PAGE = MAX_PAGE;
export const OWNER_CANDIDATE_CURSOR = "owners_after";

export function ownerCandidatesNext(path: `/${string}`, next: string | null): `/${string}` | undefined {
  return isId(next) ? `${path}?${OWNER_CANDIDATE_CURSOR}=${next}#transfer-title` : undefined;
}

// GET /internal/v1/members/self: the caller's own binding in the scope with the leave and manage hints (r2 §7).
export async function selfMembership(ctx: RequestContext, credential: string, origin: string, scope: MemberScope): Promise<MembershipResult<SelfMember>> {
  const s = scopeQuery(scope);
  if (s === undefined) return { ok: false, error: "not_found" };
  const answer = await call(ctx, { method: "GET", path: `/internal/v1/members/self?origin=${encodeURIComponent(origin)}${s}`, credential, successBytes: MEMBER_MAX_BYTES });
  return wrapped(answer, 200, "member", (v) => selfMember(scope.kind, v));
}

// GET /internal/v1/invitations: the scope's invitations (administrator view, addresses included).
export async function listInvitations(ctx: RequestContext, credential: string, origin: string, scope: MemberScope, page: Page = {}): Promise<MembershipResult<InvitationPage>> {
  const q = pageQuery(origin, page);
  const s = scopeQuery(scope);
  if (q === undefined || s === undefined) return { ok: false, error: "invalid_request" };
  const answer = await call(ctx, { method: "GET", path: `/internal/v1/invitations?${q}${s}`, credential, successBytes: INVITATION_LIST_MAX_BYTES });
  const r = rows(answer, "invitations", invitation, (i) => i.id);
  if (!r.ok) return r;
  const groupId = scope.kind === "group" ? scope.groupId : null;
  if (r.value.rows.some((i) => i.scope !== scope.kind || i.groupId !== groupId)) return { ok: false, error: "malformed" };
  return { ok: true, value: { invitations: r.value.rows, next: r.value.next } };
}

// GET /internal/v1/invitations/mine: the caller's own pending Group invitations (recipient view: no address).
export async function myInvitations(ctx: RequestContext, credential: string, origin: string, page: Page = {}): Promise<MembershipResult<InvitationPage>> {
  const q = pageQuery(origin, page);
  if (q === undefined) return { ok: false, error: "invalid_request" };
  const answer = await call(ctx, { method: "GET", path: `/internal/v1/invitations/mine?${q}`, credential, successBytes: INVITATION_LIST_MAX_BYTES });
  const r = rows(answer, "invitations", invitation, (i) => i.id);
  if (!r.ok) return r;
  if (r.value.rows.some((i) => i.scope !== "group" || i.intendedAddress !== null)) return { ok: false, error: "malformed" };
  return { ok: true, value: { invitations: r.value.rows, next: r.value.next } };
}

// GET /internal/v1/invitations/candidates: current Organization members not in the Group (Group managers; D-13).
export async function invitationCandidates(ctx: RequestContext, credential: string, origin: string, groupId: string, page: Page = {}): Promise<MembershipResult<CandidatePage>> {
  const q = pageQuery(origin, page);
  if (q === undefined || !isId(groupId)) return { ok: false, error: "invalid_request" };
  const answer = await call(ctx, { method: "GET", path: `/internal/v1/invitations/candidates?${q}&group_id=${groupId}`, credential, successBytes: CANDIDATE_LIST_MAX_BYTES });
  const r = rows(answer, "candidates", candidate, (c) => c.userId);
  return r.ok ? { ok: true, value: { candidates: r.value.rows, next: r.value.next } } : r;
}

// GET /internal/v1/invitations/{id}: one invitation as core lets the caller read it (recipient or administrator).
export async function getInvitation(ctx: RequestContext, credential: string, origin: string, id: string): Promise<MembershipResult<Invitation>> {
  if (!isId(id)) return { ok: false, error: "not_found" };
  const answer = await call(ctx, { method: "GET", path: `/internal/v1/invitations/${id}?origin=${encodeURIComponent(origin)}`, credential, successBytes: INVITATION_MAX_BYTES });
  const r = wrapped(answer, 200, "invitation", invitation);
  return r.ok && r.value.id !== id ? { ok: false, error: "malformed" } : r;
}

// POST /internal/v1/invitations: issue an invitation (201 with the one-time token; Organization: by address, Group: by
// the candidate's user id). expiresInHours absent is core's default lifetime.
export async function issueInvitation(
  ctx: RequestContext,
  credential: string,
  m: Mutation,
  scope: MemberScope,
  input: { role: OrdinaryRole; recipient: Recipient; expiresInHours?: number },
): Promise<MembershipResult<Issued>> {
  if (!validScope(scope)) return { ok: false, error: "invalid_request" };
  const body = scopeBody(m, scope);
  body.role = input.role;
  if ("userId" in input.recipient) body.intended_user_id = input.recipient.userId;
  else body.intended_address = input.recipient.address;
  if (input.expiresInHours !== undefined) body.expires_in_hours = input.expiresInHours;
  const answer = await call(ctx, { method: "POST", path: "/internal/v1/invitations", credential, body, successBytes: ISSUED_MAX_BYTES });
  if (answer.kind !== "answered" || answer.status !== 201) return failure(answer);
  const b = answer.body;
  if (!isObject(b) || !hasExactly(b, ["invitation", "token"]) || typeof b.token !== "string" || !TOKEN.test(b.token)) return { ok: false, error: "malformed" };
  const inv = invitation(b.invitation);
  const groupId = scope.kind === "group" ? scope.groupId : null;
  if (inv === undefined || inv.scope !== scope.kind || inv.groupId !== groupId || inv.status !== "pending" || inv.role !== input.role) {
    return { ok: false, error: "malformed" };
  }
  return { ok: true, value: { invitation: inv, token: b.token } };
}

// POST /internal/v1/invitations/{id}/revoke.
export async function revokeInvitation(ctx: RequestContext, credential: string, m: Mutation, id: string): Promise<MembershipResult<null>> {
  if (!isId(id)) return { ok: false, error: "not_found" };
  const answer = await call(ctx, { method: "POST", path: `/internal/v1/invitations/${id}/revoke`, credential, body: { origin: m.origin, csrf_token: m.csrfToken }, successBytes: SMALL_MAX_BYTES });
  return empty(answer);
}

const OUTCOMES: ReadonlySet<string> = new Set(["joined", "upgraded", "already_member", "consumed"]);

// POST /internal/v1/invitations/{id}/accept: the token travels in the JSON body only. Unknown, expired, revoked,
// foreign, wrong-token and wrong-recipient are one closed not_found.
export async function acceptInvitation(ctx: RequestContext, credential: string, m: Mutation, id: string, token: string): Promise<MembershipResult<Accepted>> {
  if (!isId(id)) return { ok: false, error: "not_found" };
  if (!TOKEN.test(token)) return { ok: false, error: "invalid_request", field: "invalid_token" };
  const answer = await call(ctx, { method: "POST", path: `/internal/v1/invitations/${id}/accept`, credential, body: { origin: m.origin, csrf_token: m.csrfToken, token }, successBytes: SMALL_MAX_BYTES });
  if (answer.kind !== "answered" || answer.status !== 200) return failure(answer);
  const b = answer.body;
  if (!isObject(b) || !hasExactly(b, ["outcome", "group_id"]) || typeof b.outcome !== "string" || !OUTCOMES.has(b.outcome) || !optionalId(b.group_id)) {
    return { ok: false, error: "malformed" };
  }
  return { ok: true, value: { outcome: b.outcome as AcceptOutcome, groupId: b.group_id } };
}

async function memberMutation(
  ctx: RequestContext,
  credential: string,
  m: Mutation,
  scope: MemberScope,
  path: string,
  extra: Record<string, unknown>,
  answers: "member" | "empty",
): Promise<MembershipResult<Member | null>> {
  if (!validScope(scope)) return { ok: false, error: "invalid_request" };
  const answer = await call(ctx, { method: "POST", path, credential, body: { ...scopeBody(m, scope), ...extra }, successBytes: answers === "member" ? MEMBER_MAX_BYTES : SMALL_MAX_BYTES });
  return answers === "member" ? wrapped(answer, 200, "member", (v) => member(scope.kind, v)) : empty(answer);
}

// POST /internal/v1/members/{user_id}/role: an admin or member subject to the other ordinary role.
type Changed = Promise<MembershipResult<Member | null>>;

export async function changeRole(ctx: RequestContext, credential: string, m: Mutation, scope: MemberScope, userId: string, role: OrdinaryRole): Changed {
  if (!isId(userId)) return { ok: false, error: "not_found" };
  return memberMutation(ctx, credential, m, scope, `/internal/v1/members/${userId}/role`, { role }, "member");
}

// POST /internal/v1/members/{user_id}/remove.
export async function removeMember(ctx: RequestContext, credential: string, m: Mutation, scope: MemberScope, userId: string): Changed {
  if (!isId(userId)) return { ok: false, error: "not_found" };
  return memberMutation(ctx, credential, m, scope, `/internal/v1/members/${userId}/remove`, {}, "empty");
}

// POST /internal/v1/members/{user_id}/promote: an eligible admin to owner (owners only).
export async function promote(ctx: RequestContext, credential: string, m: Mutation, scope: MemberScope, userId: string): Changed {
  if (!isId(userId)) return { ok: false, error: "not_found" };
  return memberMutation(ctx, credential, m, scope, `/internal/v1/members/${userId}/promote`, {}, "member");
}

// POST /internal/v1/members/{user_id}/demote: another owner to an ordinary role (never the last owner).
export async function demote(ctx: RequestContext, credential: string, m: Mutation, scope: MemberScope, userId: string, role: OrdinaryRole): Changed {
  if (!isId(userId)) return { ok: false, error: "not_found" };
  return memberMutation(ctx, credential, m, scope, `/internal/v1/members/${userId}/demote`, { role }, "member");
}

// POST /internal/v1/ownership/transfer: ownership to the recipient; the actor keeps formerRole or leaves ("none").
export async function transfer(ctx: RequestContext, credential: string, m: Mutation, scope: MemberScope, recipientUserId: string, formerRole: OrdinaryRole | "none"): Changed {
  if (!isId(recipientUserId)) return { ok: false, error: "invalid_request", field: "invalid_id" };
  return memberMutation(ctx, credential, m, scope, "/internal/v1/ownership/transfer", { recipient_user_id: recipientUserId, former_role: formerRole }, "empty");
}

// POST /internal/v1/members/leave: the caller's own departure (never the last owner).
export async function leave(ctx: RequestContext, credential: string, m: Mutation, scope: MemberScope): Changed {
  return memberMutation(ctx, credential, m, scope, "/internal/v1/members/leave", {}, "empty");
}

// GET /internal/v1/account/retirement-eligibility: the caller's own sole-owned scopes (IDENTITY shape; no tenant). Only
// the count and the truncation flag leave this function; no scope id or name is rendered (C6).
export async function retirementEligibility(ctx: RequestContext, credential: string, origin: string): Promise<MembershipResult<Eligibility>> {
  const answer = await call(ctx, { method: "GET", path: `/internal/v1/account/retirement-eligibility?origin=${encodeURIComponent(origin)}`, credential, successBytes: ELIGIBILITY_MAX_BYTES });
  if (answer.kind !== "answered" || answer.status !== 200) return failure(answer);
  const b = answer.body;
  if (!isObject(b) || !hasExactly(b, ["count", "scopes", "truncated"]) || !Array.isArray(b.scopes) || b.scopes.length > MAX_PAGE) return { ok: false, error: "malformed" };
  const { count, truncated } = b;
  if (typeof count !== "number" || !Number.isSafeInteger(count) || count < b.scopes.length || typeof truncated !== "boolean") return { ok: false, error: "malformed" };
  if (!truncated && count !== b.scopes.length) return { ok: false, error: "malformed" };
  for (const s of b.scopes as unknown[]) {
    if (!isObject(s) || !hasExactly(s, ["organization_id", "scope_type", "scope_id"]) || !isId(s.organization_id) || !isId(s.scope_id) ||
      (s.scope_type !== "organization" && s.scope_type !== "group")) {
      return { ok: false, error: "malformed" };
    }
  }
  return { ok: true, value: { count, truncated } };
}

// G060 (Plan r0 §1.4, §4.2, D-3, D-11): the Organization's Group-management override. overrideEnabled is the current
// setting (default off); canChange is core's owner hint (organization.manage_policy), never authority: the write is
// re-decided by core and a non-owner is refused (forbidden). Every current Organization member may read it.
export type GroupAuthority = { overrideEnabled: boolean; canChange: boolean };

function groupAuthorityOf(answer: Answer): MembershipResult<GroupAuthority> {
  if (answer.kind !== "answered" || answer.status !== 200) return failure(answer);
  const b = answer.body;
  if (!isObject(b) || !hasExactly(b, ["override_enabled", "can_change"]) || typeof b.override_enabled !== "boolean" || typeof b.can_change !== "boolean") {
    return { ok: false, error: "malformed" };
  }
  return { ok: true, value: { overrideEnabled: b.override_enabled, canChange: b.can_change } };
}

// GET /internal/v1/organization/group-authority.
export async function groupAuthority(ctx: RequestContext, credential: string, origin: string): Promise<MembershipResult<GroupAuthority>> {
  const answer = await call(ctx, { method: "GET", path: `/internal/v1/organization/group-authority?origin=${encodeURIComponent(origin)}`, credential, successBytes: SMALL_MAX_BYTES });
  return groupAuthorityOf(answer);
}

// POST /internal/v1/organization/group-authority: an absolute assignment (exactly origin, csrf_token and enabled). The
// answer must report the requested value.
export async function setGroupAuthority(ctx: RequestContext, credential: string, m: Mutation, enabled: boolean): Promise<MembershipResult<GroupAuthority>> {
  const answer = await call(ctx, {
    method: "POST",
    path: "/internal/v1/organization/group-authority",
    credential,
    body: { origin: m.origin, csrf_token: m.csrfToken, enabled },
    successBytes: SMALL_MAX_BYTES,
  });
  const r = groupAuthorityOf(answer);
  return r.ok && r.value.overrideEnabled !== enabled ? { ok: false, error: "malformed" } : r;
}

// --- form helpers (pure; unit-tested in membership-forms.test.mjs) ----------------------------------------------------

type Get = (name: string) => string | undefined | symbol;

export const MEMBER_ACTIONS = ["invite", "revoke", "change_role", "remove", "promote", "demote", "transfer", "leave"] as const;
export type MemberAction =
  | { action: "invite"; role: OrdinaryRole; recipient: Recipient; expiresInHours: number }
  | { action: "revoke"; invitationId: string }
  | { action: "change_role" | "demote"; userId: string; role: OrdinaryRole }
  | { action: "remove" | "promote"; userId: string }
  | { action: "transfer"; recipientUserId: string; formerRole: OrdinaryRole | "none" }
  | { action: "leave" };

function one(get: Get, name: string): string | undefined {
  const v = get(name);
  return typeof v === "string" ? v : undefined;
}

// The submitted action of a members form, validated for shape only (core decides authority and every rule). A repeated,
// missing or malformed field is no action (the route answers "Request not understood"). An Organization invitation is
// by address (passed as typed: core normalizes and validates it); a Group invitation is by a candidate's user id.
export function readMemberAction(kind: ScopeKind, get: Get): MemberAction | undefined {
  const action = one(get, "action");
  const userId = one(get, "user_id");
  const role = one(get, "role");
  switch (action) {
    case "invite": {
      const hours = Number(one(get, "expires_in_hours") ?? "");
      if (!isOrdinaryRole(kind, role) || !LIFETIME_HOURS.some((h) => h === hours)) return undefined;
      if (kind === "group") {
        const id = one(get, "intended_user_id");
        return isId(id) ? { action, role, recipient: { userId: id }, expiresInHours: hours } : undefined;
      }
      const address = one(get, "intended_address");
      if (address === undefined || address === "" || Buffer.byteLength(address) > 4 * MAX_ADDRESS_BYTES) return undefined;
      return { action, role, recipient: { address }, expiresInHours: hours };
    }
    case "revoke": {
      const id = one(get, "invitation_id");
      return isId(id) ? { action, invitationId: id } : undefined;
    }
    case "change_role":
    case "demote":
      return isId(userId) && isOrdinaryRole(kind, role) ? { action, userId, role } : undefined;
    case "remove":
    case "promote":
      return isId(userId) ? { action, userId } : undefined;
    case "transfer": {
      const former = one(get, "former_role");
      if (!isId(userId)) return undefined;
      return former === "none" || isOrdinaryRole(kind, former) ? { action, recipientUserId: userId, formerRole: former } : undefined;
    }
    case "leave":
      return { action };
    default:
      return undefined;
  }
}

// The closed outcome codes of the action route (303 ?status=) and the message each page shows; never a submitted value.
export const ACTION_STATUS: Readonly<Record<string, string>> = {
  revoked: "The invitation was revoked.",
  role_changed: "The member's role was changed.",
  removed: "The member was removed.",
  promoted: "The member is now an owner.",
  demoted: "The owner's role was changed.",
  transferred: "Ownership was transferred.",
  left: "You have left.",
};

export const STATUS_OF: Readonly<Record<Exclude<MemberAction["action"], "invite">, string>> = {
  revoke: "revoked",
  change_role: "role_changed",
  remove: "removed",
  promote: "promoted",
  demote: "demoted",
  transfer: "transferred",
  leave: "left",
};

// The closed field codes of core's refusals and the fixed sentence each shows. Conflicts state the current rule, never
// success; nothing names another member, an address or a token.
const REFUSAL_FIELDS: Readonly<Record<string, string>> = {
  last_owner: "This would leave no owner. Promote or transfer ownership to another member first.",
  recipient_ineligible: "That person cannot receive this role or invitation here (for example, they are not a current member or their account is disabled).",
  invitation_limit: "This workspace has the maximum number of pending invitations. Revoke one before inviting again.",
  context: "This change needs the workspace itself as its context. Open the members page of the workspace and try again.",
  dependent_bindings: "This member still has Group roles in this workspace. Remove those first.",
  invitation_not_pending: "That invitation is no longer pending.",
  issuer_authority_lost: "The person who sent this invitation can no longer grant it. Ask for a new invitation.",
  owner_subject: "Owners cannot be changed or removed this way. Use the ownership actions instead.",
  self_target: "You cannot change your own role or invite yourself this way.",
  self_escalation: "You cannot raise your own role in a Group this way.",
  not_owner: "That member is not an owner.",
  invalid_address: "Enter a single valid email address.",
  invalid_role: "Choose one of the offered roles.",
  invalid_lifetime: "Choose one of the offered expiry periods.",
  recipient: "That recipient does not fit this kind of invitation.",
  invalid_token: "Enter the invitation code exactly as you received it.",
};

export function refusalSentence(field: unknown): string | undefined {
  return typeof field === "string" && Object.hasOwn(REFUSAL_FIELDS, field) ? REFUSAL_FIELDS[field] : undefined;
}

// A core outcome that may have committed: no answer in time, no connection after sending, a malformed answer, or
// membership_unavailable.
export function outcomeUnknown(error: MembershipError): boolean {
  return error === "timeout" || error === "unreachable" || error === "malformed" || error === "membership_unavailable";
}

// The no-store page of a refused action; nothing changed in each case (the page names the closed field's sentence).
export function membershipRefusal(error: MembershipError, field: string | undefined, back: StatusLink): Response {
  const sentence = refusalSentence(field);
  switch (error) {
    case "conflict":
      return onboardingPage(409, "Not changed", [sentence ?? "The current state does not allow this change.", "Nothing was changed. The current state is on the members page."], [back]);
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
export function membershipUnknown(back: StatusLink): Response {
  return onboardingPage(503, "Outcome unknown", ["Your request may or may not have completed. Check the current state before trying again."], [back]);
}

// The direct 200 no-store result of an invitation (C4: a redirect cannot carry the one-time token). The link carries the
// token in its fragment, which a browser never sends to a server; the token is shown once, on this page only, and never
// in its title, a URL this server receives, a Location header or a log. G058 (scope amendment04): the link follows the
// invitation's own scope, never the recipient or an account: an Organization invitation opens the enrollment entry
// /join/invitations/<id> (the recipient signs in, or creates an account, there and core admits them); a Group invitation
// keeps the existing member page /invitations/<id>.
export function invitationIssuedPage(origin: string, issued: Issued, back: StatusLink): Response {
  const entry = issued.invitation.scope === "organization" ? "/join/invitations/" : "/invitations/";
  const link = `${origin}${entry}${issued.invitation.id}#${issued.token}`;
  const paragraphs = [
    "The invitation was created. Copy the link below now and send it to the person you invited: it is shown only once.",
    link,
    `Expires: ${issued.invitation.expiresAt.replace("T", " ").replace("Z", " UTC")}.`,
    "If this page was lost before you copied the link, revoke the pending invitation on the members page and create a new one.",
  ];
  if (issued.invitation.scope === "organization") {
    paragraphs.push("The person opens the link, signs in (or creates an account with the invited address) and joins this workspace.");
  }
  return onboardingPage(200, "Invitation created", paragraphs, [back]);
}

// --- the Group-authority setting (G060; POST /organization/group-authority/action) -----------------------------------

export const GROUP_AUTHORITY_PATH = "/organization/group-authority";

// The setting form's one choice, validated for shape only: exactly one "enabled" field, "on" or "off" (core decides who
// may change it). Anything else is no choice.
export function readGroupAuthorityForm(get: Get): boolean | undefined {
  const v = one(get, "enabled");
  return v === "on" ? true : v === "off" ? false : undefined;
}

// The closed ?status= codes of the setting page; the message names the state core reported, never a submitted value.
export const GROUP_AUTHORITY_STATUS: Readonly<Record<string, string>> = {
  enabled: "Group administration by Organization staff is now on.",
  disabled: "Group administration by Organization staff is now off. Organization owners and admins keep only the Group roles they actually hold.",
};

// POST /organization/group-authority/action: the form layer (method, form encoding, exact Origin, 16 KiB body), the
// closed choice, the session, then core decides (owner only, CSRF, current state). Success is a 303 to the page with the
// state core reported; a refusal is a fixed page ("Not permitted" for every non-owner); an outcome core may have
// committed is the unknown-state page, which never claims success and never retries.
export async function groupAuthorityActionRoute(request: Request): Promise<Response> {
  const form = await readForm(request);
  if (!form.ok) return form.response;
  const get = (name: string) => field(form.fields, name);
  const csrfToken = get("csrf_token");
  const enabled = readGroupAuthorityForm(get);
  if (typeof csrfToken !== "string" || csrfToken === "" || enabled === undefined) return badRequest();
  const back: StatusLink = { href: GROUP_AUTHORITY_PATH, label: "Back to Group administration" };
  const credential = readCookie(form.cookie, "session", form.profile);
  if (credential === undefined) return membershipRefusal("unauthenticated", undefined, back);
  const r = await setGroupAuthority(deriveContext(request.headers), credential, { origin: form.origin, csrfToken }, enabled);
  if (r.ok) return seeOther(`${GROUP_AUTHORITY_PATH}?status=${r.value.overrideEnabled ? "enabled" : "disabled"}`);
  return outcomeUnknown(r.error) ? membershipUnknown(back) : membershipRefusal(r.error, r.field, back);
}

// --- the shared action route (POST /organization/members/action and /groups/<uuid>/members/action) -------------------

// The existing form-route order (method, form encoding, exact Origin, the 16 KiB body budget), then the closed action
// field and its shape, then core decides. Every action except invite answers 303 to the members page with a closed
// status; invite answers the direct no-store result page; a refusal is a fixed page; an outcome core may have committed
// is the unknown-state page, which never retries.
export async function memberActionRoute(request: Request, scope: MemberScope, membersPath: `/${string}`): Promise<Response> {
  const form = await readForm(request);
  if (!form.ok) return form.response;
  const get = (name: string) => field(form.fields, name);
  const csrfToken = get("csrf_token");
  const act = readMemberAction(scope.kind, get);
  if (typeof csrfToken !== "string" || csrfToken === "" || act === undefined) return badRequest();
  const back: StatusLink = { href: membersPath, label: "Back to the members page" };
  const credential = readCookie(form.cookie, "session", form.profile);
  if (credential === undefined) return membershipRefusal("unauthenticated", undefined, back);
  const ctx = deriveContext(request.headers);
  const m: Mutation = { origin: form.origin, csrfToken };
  if (act.action === "invite") {
    const r = await issueInvitation(ctx, credential, m, scope, act);
    if (r.ok) return invitationIssuedPage(form.origin, r.value, back);
    return outcomeUnknown(r.error) ? membershipUnknown(back) : membershipRefusal(r.error, r.field, back);
  }
  let r: MembershipResult<Member | null>;
  switch (act.action) {
    case "revoke":
      r = await revokeInvitation(ctx, credential, m, act.invitationId);
      break;
    case "change_role":
      r = await changeRole(ctx, credential, m, scope, act.userId, act.role);
      break;
    case "demote":
      r = await demote(ctx, credential, m, scope, act.userId, act.role);
      break;
    case "remove":
      r = await removeMember(ctx, credential, m, scope, act.userId);
      break;
    case "promote":
      r = await promote(ctx, credential, m, scope, act.userId);
      break;
    case "transfer":
      r = await transfer(ctx, credential, m, scope, act.recipientUserId, act.formerRole);
      break;
    case "leave":
      r = await leave(ctx, credential, m, scope);
      break;
    default:
      return badRequest();
  }
  if (r.ok) return seeOther(`${membersPath}?status=${STATUS_OF[act.action]}`);
  return outcomeUnknown(r.error) ? membershipUnknown(back) : membershipRefusal(r.error, r.field, back);
}
