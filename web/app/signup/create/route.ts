import { deriveContext } from "@/lib/request-context";
import {
  createOrganization,
  formExpired,
  outcomeUnknown,
  refusedRoot,
  rootOfRootRequest,
  suggestSlug,
  type Created,
  type SignupError,
  type SignupResult,
} from "@/lib/signup-client";
import {
  AMBIGUOUS,
  badRequest,
  clearCookie,
  field,
  forbidden,
  methodNotAllowed,
  readCookie,
  readForm,
  readSignupFormCookie,
  seeOther,
  setSignupFormCookie,
  type CookieProfile,
  type SignupForm,
} from "@/lib/session";

// G062 (Plan r5 §1.3, §1.4, §4.4; Codex Round 14 option A): POST /signup/create on the platform root. The existing
// form-route order, the root check, the signup credential's CSRF token (core checks it) and the signup credential from
// its host-only cookie; core decides every write. Three closed shapes:
// - {csrf_token, name}: core's advisory suggestion for the name; the form restoration cookie records the name and the
//   suggested address, and the details page shows both for editing;
// - {csrf_token, name, slug}: constrained provisioning; the restoration cookie records the exact submitted name and slug
//   (also on a field refusal or an unknown outcome, so the form is restored and an identical retry answers
//   already_created). 201 keeps the signup cookie (Plan r5 r4-2) and 303s into the continuation core named;
// - {csrf_token, retry: "1"}: the failure-path "Try again", which repeats the recorded payload unchanged (a tampered or
//   changed record is refused by core as intent_used and creates nothing; the record is never rewritten here).
// A dead credential clears both cookies. Nothing submitted is ever put in a URL or a log.
export const dynamic = "force-dynamic";

export async function POST(request: Request): Promise<Response> {
  const form = await readForm(request);
  if (!form.ok) return form.response;
  const csrfToken = field(form.fields, "csrf_token");
  const name = field(form.fields, "name");
  const slug = field(form.fields, "slug");
  const retry = field(form.fields, "retry");
  if (typeof csrfToken !== "string" || csrfToken === "" || name === AMBIGUOUS || slug === AMBIGUOUS || retry === AMBIGUOUS) return badRequest();
  if (retry !== undefined ? retry !== "1" || name !== undefined || slug !== undefined : name === undefined) return badRequest();
  const ctx = deriveContext(request.headers);
  const root = await rootOfRootRequest(ctx, form.origin);
  if (!root.ok) return refusedRoot(root.error);
  const credential = readCookie(form.cookie, "signup", form.profile);
  if (credential === undefined) return expired(form.profile);
  const m = { origin: form.origin, csrfToken };
  if (retry !== undefined) {
    const recorded = readSignupFormCookie(form.cookie, form.profile);
    if (recorded === undefined || recorded.slug === "") return seeOther("/signup/details?status=retry_unavailable");
    return created(await createOrganization(ctx, credential, m, recorded.name, recorded.slug), form.profile);
  }
  if (slug === undefined) {
    const s = await suggestSlug(ctx, credential, m, name as string);
    if (s.ok) return withForm(seeOther("/signup/details"), { name: name as string, slug: s.value.suggested ?? "" }, form.profile);
    return refused(s, form.profile);
  }
  const submitted: SignupForm = { name: name as string, slug };
  const result = await createOrganization(ctx, credential, m, submitted.name, submitted.slug);
  // The exact submitted values are recorded on success, a field refusal, a slug conflict, a rate limit and an unknown
  // outcome (so the form is restored and an identical retry is possible). Not recorded: intent_used (the row was
  // provisioned with another payload; the existing record is kept unchanged), a dead credential (both cookies are
  // cleared), and a refused form or origin.
  const keep = result.ok || !(NOT_RECORDED.has(result.error) || (result.error === "conflict" && result.field === "intent_used"));
  const res = created(result, form.profile);
  return keep ? withForm(res, submitted, form.profile) : res;
}

const NOT_RECORDED: ReadonlySet<string> = new Set(["unauthenticated", "csrf_failed", "origin_rejected"]);

// The answer of a create or retry: 303 into core's continuation, or the details page with a closed status.
function created(result: SignupResult<Created>, profile: CookieProfile): Response {
  if (result.ok) return seeOther(result.value.continueUrl);
  if (result.error === "conflict") return seeOther(`/signup/details?status=${result.field === "slug_taken" ? "slug_taken" : "intent_used"}`);
  return refused(result, profile);
}

function refused(result: { ok: false; error: SignupError; field?: string }, profile: CookieProfile): Response {
  switch (result.error) {
    case "unauthenticated":
      return expired(profile);
    case "csrf_failed":
      return formExpired();
    case "origin_rejected":
      return forbidden();
    case "user_disabled":
      return seeOther("/signup?status=user_disabled");
    case "rate_limited":
      return seeOther("/signup/details?status=rate_limited");
    case "invalid_request":
      if (result.field === "name") return seeOther("/signup/details?status=invalid_name");
      if (result.field === "slug") return seeOther("/signup/details?status=invalid_slug");
      return badRequest();
    default:
      return seeOther(`/signup/details?status=${outcomeUnknown(result.error) ? "unknown" : "unavailable"}`);
  }
}

// A dead or missing signup credential: both root signup cookies are cleared and the browser starts again.
function expired(profile: CookieProfile): Response {
  return seeOther("/signup?status=expired", clearCookie("signup", profile), clearCookie("signup_form", profile));
}

// Appends the form restoration record when the values fit its bounds (a name or slug that cannot be recorded is simply
// not restored).
function withForm(res: Response, form: SignupForm, profile: CookieProfile): Response {
  const cookie = setSignupFormCookie(form, profile) ?? setSignupFormCookie({ name: form.name, slug: "" }, profile);
  if (cookie !== undefined) res.headers.append("Set-Cookie", cookie);
  return res;
}

export function GET(): Response {
  return methodNotAllowed("POST");
}

export function PUT(): Response {
  return methodNotAllowed("POST");
}

export function PATCH(): Response {
  return methodNotAllowed("POST");
}

export function DELETE(): Response {
  return methodNotAllowed("POST");
}

export function OPTIONS(): Response {
  return methodNotAllowed("POST");
}
