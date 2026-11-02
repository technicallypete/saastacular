import "server-only";

import { readMutationFields, type MutationFields } from "./branding";
import { isHold, MAX_AVATAR_BYTES, type ProfileFailure, type ProfileResult } from "./profile-client";
import { AMBIGUOUS, badRequest, field, forbidden, seeOther, statusPage } from "./session";

// G074 (Plan r3.1 §3.7, §3.10): the pure form helpers of the three /account profile form routes (native POST forms;
// POST-redirect-GET to /account with one closed profile_status). The multipart avatar form is reduced to the raw image
// bytes with the G017 logo form's shape (lib/branding.ts readBoundedBytes / parseLogoUpload), generalized to the avatar
// field. Query values are never reflected: every redirect carries a closed code only, and the page maps it to fixed
// text. Core alone validates names and images and decides every write.

export const ACCOUNT_PAGE = "/account";
// The multipart avatar form: the 512 KiB image plus at most 16 KiB of form framing; anything larger is never parsed.
export const MAX_AVATAR_FORM_BYTES = MAX_AVATAR_BYTES + 16 * 1024;

// The closed outcome codes of the profile forms (the page's profile_status parameter, disjoint from its status and
// mcp_status parameters).
export const PROFILE_STATUSES = [
  "profile_saved",
  "avatar_set",
  "avatar_removed",
  "stale_version",
  "invalid_name",
  "invalid_image",
  "payload_too_large",
  "storage_unavailable",
  "avatar_hold",
] as const;
export type ProfileStatus = (typeof PROFILE_STATUSES)[number];
export type ProfileSuccess = "profile_saved" | "avatar_set" | "avatar_removed";

// Core's closed invalid-input fields of an avatar upload (the branding.NormalizeLogo fields mapped logo_* → avatar_*).
export const AVATAR_INPUT_FIELDS: readonly string[] = ["avatar_type", "avatar_dimensions", "avatar_malformed", "normalized_too_large"];

// The page a profile form returns to, with one closed status, at the Profile section.
export function profileBack(status: ProfileStatus): Response {
  return seeOther(`${ACCOUNT_PAGE}?profile_status=${status}#profile`);
}

// The name form's fields: csrf_token and expected_version (readMutationFields) and exactly one name field ("" clears
// the explicit name). The name is passed as submitted: core trims and validates it. A missing or repeated field, or a
// version outside the grammar, is undefined (the route answers 400).
export function readNameForm(fields: URLSearchParams): { m: MutationFields; name: string } | undefined {
  const m = readMutationFields((n) => field(fields, n));
  const value = field(fields, "name");
  if (m === undefined || typeof value !== "string") return undefined;
  return { m, name: value };
}

export type AvatarUpload = { ok: true; fields: MutationFields; bytes: Uint8Array } | { ok: false; error: "bad_request" | "too_large" };

// The avatar form's multipart body: exactly the fields csrf_token, expected_version and one non-empty avatar file of at
// most 512 KiB, each at most once. The file's declared name and type are ignored: core validates the bytes and decides.
export async function parseAvatarUpload(contentType: string, body: Uint8Array): Promise<AvatarUpload> {
  let form: FormData;
  try {
    form = await new Response(Buffer.from(body), { headers: { "Content-Type": contentType } }).formData();
  } catch {
    return { ok: false, error: "bad_request" };
  }
  const names = [...form.keys()];
  if (names.some((n) => n !== "csrf_token" && n !== "expected_version" && n !== "avatar")) return { ok: false, error: "bad_request" };
  const text = (n: string): string | undefined | typeof AMBIGUOUS => {
    const values = form.getAll(n);
    if (values.length > 1) return AMBIGUOUS;
    return typeof values[0] === "string" ? values[0] : values[0] === undefined ? undefined : AMBIGUOUS;
  };
  const fields = readMutationFields(text);
  const files = form.getAll("avatar");
  const file = files.length === 1 && typeof files[0] !== "string" ? files[0] : undefined;
  if (fields === undefined || file === undefined || file.size === 0) return { ok: false, error: "bad_request" };
  if (file.size > MAX_AVATAR_BYTES) return { ok: false, error: "too_large" };
  return { ok: true, fields, bytes: new Uint8Array(await file.arrayBuffer()) };
}

// The closed status of a refused or unconfirmed profile write that returns to the page, or undefined when the refusal
// is one of the form layer's fixed pages (exported for the unit tests). The three avatar HOLD reasons are one status;
// every other unavailable answer, and no answer, is storage_unavailable (the outcome is not claimed either way).
export function profileStatusOf(failure: ProfileFailure): ProfileStatus | undefined {
  switch (failure.error) {
    case "version_conflict":
      return "stale_version";
    case "invalid_request":
      return failure.field === "name" ? "invalid_name" : undefined;
    case "invalid_input":
      if (failure.field === "avatar_too_large") return "payload_too_large";
      return failure.field !== undefined && AVATAR_INPUT_FIELDS.includes(failure.field) ? "invalid_image" : undefined;
    case "payload_too_large":
      return "payload_too_large";
    case "unavailable":
      return isHold(failure) ? "avatar_hold" : "storage_unavailable";
    case "timeout":
    case "unreachable":
    case "malformed":
      return "storage_unavailable";
    default:
      return undefined;
  }
}

// A profile form's answer (POST-redirect-GET): success and the recoverable refusals return to the page with one closed
// status; the session, CSRF and origin refusals are the form layer's fixed pages; a viewer whose context core refuses
// gets a fixed not-available page.
export function profileAnswer(result: ProfileResult<unknown>, success: ProfileSuccess): Response {
  if (result.ok) return profileBack(success);
  const status = profileStatusOf(result);
  if (status !== undefined) return profileBack(status);
  switch (result.error) {
    case "csrf_failed":
      return statusPage(403, "Form expired", "This form is no longer valid. Reload your account page and try again. Nothing was changed.");
    case "unauthenticated":
      return statusPage(401, "Not signed in", "Sign in first, then try again. Nothing was changed.");
    case "origin_rejected":
      return forbidden();
    case "forbidden":
    case "not_available":
      return statusPage(403, "Not available", "Your profile cannot be changed from this workspace right now. Nothing was changed.");
    case "not_found":
    case "unknown_host":
      return statusPage(404, "Not found", "There is nothing at this address.");
    default:
      return badRequest();
  }
}
