> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G017 Implementation Plan — Organization branding and logged-out shell

Revision **r2** (provisional; second bounded correction round), 2026-10-16 UTC. Author: Claude (model Fable 5.1;
planning effort "high (requested)" — the runtime setting is not verifiable from inside the worker, actual value
**unknown**). Worktree: generated name `agent-<id>` (fresh planning-only successor to the r0/r1 worker
`agent-<id>`; recorded per AGENTS "Task-prefixed Claude worktrees"). Branch `g017-plan-r2`, created at
the r1 commit `61d6ff66` and containing `origin/staging` `21469fc0`. Status: **proposed for Codex
review; not an execution binding; no implementation, build, test, probe, container or push.** Tier: architecture/
security/data with a bounded ordinary UI delta. Budget **B5, checkpoint after attempt 3**; planning spends no attempt.

**r1 revision note.** r0 is archived byte-exact by Codex as
r0 (not included) (SHA-256
`5d9a80e69c342075270c809954d4bd7f241c084cd17b747366bffd14aa644121`) with
allow-list r0 (not included) (`fe3352dc6554d77b1df76d5b342b09708a596ec2c7609af335c0e4c95772b406`);
both are kept. This revision answers the eight corrections of the
r0 review (not included) (SHA-256
`c88b8822f5f0f0fdbfa42ebccb4be40b144970a8fd06f54d8ca193598459f7f8`) and its decisions; unchanged design is
restated only where a correction touches it.

| Review item | r1 change (section) |
|---|---|
| 1 Complete image safety | §1.4.1 V: bounded **full decode** + strict trailing-content policy + PNG chunk allow-list; BRD-9 extended |
| 2 Upload recovery / uncertain commit | §1.4.2: the uploads ledger is **removed**; discoverability comes from the object rows themselves; `commit_outcome_unknown` reconciles before any compensation; stages on the existing fault seam; BRD-10 cases |
| 3 Deletion and retention | §4: **no foreign keys** from branding to `stored_objects`; guard-trigger checks at write; projection and management views join on live state; purge/tombstones untouched; current-logo deletion → fallback; lock order pinned |
| 4 Auth-reader RLS path | §4 items 7–8: exact FORCE-RLS SELECT policies for `<db>_auth_reader` (branding `USING (true)`; `stored_objects` only rows that are a current logo); CT-BRD-1 and BRD-11 pins; public HTTP carries no key/id |
| 5 Query layer | §3/§5.1: SQL lives in `core/internal/tenancy/branding.go` (store methods, `sql*` constants) under the existing `TestSQLBoundary`/`TestObjectSQLBoundary` pattern; `access` holds no SQL; no sqlc |
| 6 Effective contrast and states | §1.3: pairs re-pinned on the **actual** roles (`ink`, `text-2`, `text-3` inherit unless `foreground` is overridden), hover/press brightness transforms, control/focus checks, partial-override negatives |
| 7 Focused validation | §6.3: only new tests, changed pins and narrow affected regressions; browser through the existing `identity:browser` and `platform:browser` commands with `--project` selection; **no** root/ops umbrella; no speculative fallback |
| 8 Manual preview / early draft MR | §6.4: the existing identity overlay on the operator workstation (`127.0.0.1:543`, run CA, platform hosts) as the preview; readiness handoff and cleanup; draft MR opened early (G018 trigger) with manual feedback before final qualification/merge |
| Status, pins, counts, C-1..C-3 | §0.1, §0.4, §11.2, §12 corrected; C-1 and C-3 resolved by inspection |
| D-1 | **open owner decision** (§10); D-2/D-9/D-10/C-2 accepted as directed; D-4/D-7/D-8 Spec-authorized narrow work subject to corrections 1–4; D-5 bounded copy (extraction patterns removed); D-6 kept with pins; cache contract restated truthfully (§5.5); PG-008 as a source estimate (§5.6) |

**r2 revision note.** r1 is archived byte-exact as
r1 (not included) (SHA-256
`5b72dbf2c561aadb0c40b6dc2d09b96f62bcef88e817eaf62374453d77204204`) with
allow-list r1 (not included) (`dc6373012c18f0b6003c615b79ef095a27a86a10dee3f829545678523b4ad8a1`);
the allow-list is **unchanged** in r2 (same bytes, same hash). This revision answers only the three remaining
corrections of the r1 review (not included) (SHA-256
`bde78eb4cad312ffd7a65dd1d2568aa4f3051b2d7fda6918a947ba15d683982a`) and applies its binding directions. Every r1
correction in the table above stays closed as written; unrelated sections are untouched.

| r1 review item | r2 change (section) |
|---|---|
| 1 Uncertain publication outcome | §1.4.2 R/C: retain the object and report `publication_outcome_unknown` unless non-commit is definitive; in-flight and advanced-version cases covered; actual object states reported; BRD-10 extended on the existing seam only |
| 2 Complete-image/trailing validation | §1.4.1: bounded stdlib decode **and re-encode** to a normalized output with final byte/type limits; buffering-independent PNG/JPEG structural walks keep the appended-content negatives; memory/work bounds corrected (16-bit PNG = 8 B/px, 32 MiB); BRD-9 kept and extended |
| 3 Actual contrast states | §1.3: whole-button transforms (foreground **and** fill); the invented fill-boundary comparisons (r1 N1, N4) withdrawn; focus outline against the enclosing page/band/dialog surface; dialog Close icon `muted-foreground`/`popover` at 3:1; thresholds unchanged; defaults checked by source arithmetic before execution |
| C-4 cache | §5.5: the disclosed 24 h public-logo cache is retained; the revalidation alternative is withdrawn |
| C-5 helper-role RLS | §4 item 6: the explicit auth-reader policies are the chosen design; the definer-only fallback is withdrawn (it cannot bypass FORCE RLS); public-reader test wording corrected (BRD-11) |
| Manual preview | §6.4: the emulator-backed identity overlay (L2) with its seeded users; no WorkOS L4 settings or credentials; no tunnel, no deployment, no mobile assumption |
| D-1 | still **PENDING** with the owner (§0.1, §10); no answer is assumed |
| PG-008 | source estimate only; no benchmark gate (§5.6) |

> **Provisional base.** The design still rests on the unmerged G016 implementation (MR !34). Per the review, G016's
> source/delta ACCEPT is recorded at staging `26ccab6d`; the post-MR fast/Duo correction is G016 **attempt 2**, with
> its final dependency/qualification holds intact. §0.2 (r0) pins the G016 interfaces at `0a163c7f`; **R0 (§0.3)
> re-reads them against the integrated staging content after G016's final ACCEPT**. Provisional design review may
> proceed now; execution reassessment and release wait.

- Inputs: Goal and Spec (digests §2); Architecture §§37A, 50, 51, 61, 63; ADR-001/003/004/005/006/007; WEB-SHELL,
  AUTHORIZATION, OBJECT-STORAGE, DOMAINS, AUDIT, DATABASE-ISOLATION, RECOVERY, IDENTITY, SECURITY, PERFORMANCE
  (PG-008); the Vestive skill (`web/.claude/skills/vestive-design/SKILL.md`, precedence rules 1–7 applied).
- Planning rules honoured: read-only inspection; no install, generation, build, test, Docker, provider call or
  settings change; edits limited to this Plan and the allow-list proposal.

## 0. Gates, provisional G016 interfaces and R0

### 0.1 Gates

| Gate | Status | Owner |
|---|---|---|
| G016: source/delta ACCEPT at `26ccab6d`; post-MR fast/Duo correction = attempt 2; final ACCEPT, hosted qualification, P-1..P-4 and integration | **open** | Codex / owner |
| **D-1** privileged mapping (owner+admin vs owner-only) | **asked of the owner; PENDING**. The recommendation is not approval. Planning may continue; implementation cannot | owner |
| R0 reconciliation (§0.3) after G016 integration | after integration | Claude; Codex reviews |
| Codex review of this r2; Plan ACCEPT; allow-list import; initial Result; Git-first binding; B5 ledger; resource readiness; recorded execution release | after the above | Codex |
| G018 provisional planning trigger: this Goal's first actual (draft) implementation MR (§6.4) | — | — |
| G021 barrier closed | — | — |

### 0.2 Provisional G016 interfaces (P-G016-1..12)

Unchanged from r0 (file:line at `0a163c7f`); the rows relied on below are P-G016-1 (resolver), -3 (session scope
`begin/scope/failure` in `domains/api/http.go:262-376`), -4 (`s.guard` helper), -5 (twelve capabilities), -6 (audit
35), -7 (00012 last; auth-reader definer pattern), -8 (web gate), -9 (session answer and `app.is_organization_member`),
-10 (overlay seeds and projects), -11 (harness rows 119), -12 (storage not wired in `cmd/server`).

### 0.3 R0 after G016 integration

As r0: re-read every P-G016 row against integrated staging, re-pin §11.2, rebind §2, record drift; material drift
returns as a Plan revision.

### 0.4 Resolutions by inspection (C-1..C-3 of r0)

| Item | Finding (read-only, at `b9b6e5c0`) | Plan consequence |
|---|---|---|
| C-1 non-member fixture | `identitysetup` (`infra/validation/tools/cmd/identitysetup/main.go` header) generates CA/leaf keys, the emulator seed, `browser/users.json` and secrets; it has **no membership operation**. Memberships are local rows seeded by `infra/identity/mcp-seed.sql`; the Playwright runner has no database access | The revoked/non-member indicator state is proven at the HTTP/core level (`TestBranding` BRD-13: session answer `membership: "none"` after the binding row is removed in the test transaction; `http_test` renders the closed answer) plus one web unit case of the indicator component on a stubbed `SessionView`. The browser E-BRD-4 proves the **member** state; no new tool, seeded credential or fixture capability is added |
| C-2 tenant entry content | directed by the review | identity + sign-in + appearance/help only (§1.5) |
| C-3 log/route enums | `infra/validation/tools/internal/logrecord/completion.schema.json:40-42,91` enumerates every `/internal/v1/*` route; `events.json:135-168` pins the `domain.request` event with its closed reasons | **Required**: add the four branding routes to both route enums, one `branding.request` event (operation, closed outcome/reason, host tag) to `events.json`, and the `logrecord_test.go` pin — exactly as G016 amendments 04/05 did. Patterns are unconditional in the allow-list |

## 1. Approach

### 1.1 Trust boundaries (as r0, with the corrected containments)

| Boundary | Containment (r1) |
|---|---|
| Public projection | closed definer read after G016 `Lookup`: `display_label`, `light`, `dark`, `version`, and the logo's SHA-256 prefix and content type; never `organizations.name`, memberships, people, policy, provider IDs, keys, object UUIDs, domain proof |
| Logo bytes | PNG/JPEG only, **complete decode and re-encode** within pinned bounds (stored bytes are the server's own encoder output, never the caller's), PNG chunk allow-list and JPEG segment walk over the whole buffer (buffering-independent appended-content rejection), server-chosen MIME, input and output ≤ 512 KiB, 16–2048 px; no caller-supplied object id; served `nosniff`/`inline`, bytes verified against the stored SHA-256 while streaming; zero publication on any failure |
| Tenant-host shell | label rendered as text; per-request resolution from the request's own origin after the gate; HTML `no-store`; server-rendered scope style (no wrong-tenant paint) |
| Cache | only the versioned logo response is cacheable (§5.5, truthful bounded contract); nothing personalized is cacheable |
| Privileged mutation | thirteenth capability (D-1 pending), FORCE RLS, `expected_version` with `FOR NO KEY UPDATE`, `access.denied` on denial, zero effect on invalid input |
| Durable state | additive 00013 without foreign keys into `stored_objects`; tombstones, purge and G013 retention untouched; down refused while rows/events exist |

### 1.2 Riskiest slice first — as r0 (§8).

### 1.3 Configuration model and effective contrast (R1, R2; correction 6)

Row shape, label rule (≤ 80 code points, NFC, no control characters), 15 overridable tokens (no `overlay`), grammar
`^#[0-9a-f]{6}$` after lowercase normalization, and the no-brand fallback to `PRODUCT_NAME` (never
`organizations.name`) are as r0 (accepted D-2/D-9/D-10).

**Effective palette per appearance** (computed in core on every write, Light and Dark separately):
`eff[t] = override[t] ?? platform[t]` for the 15 tokens; roles `ink`, `text-2`, `text-3` = `override.foreground` when
that appearance overrides `foreground`, otherwise the **platform role values of that appearance** (WEB-SHELL "Internal
text roles": Light `#0d0d0d`/`#36322e`/`#4d4843`, Dark `#ffffff`/`#c9c5c0`/`#a8a39d`) — exactly what
`tokenOverrideStyle()` emits. Transformed states follow the unchanged button classes (`components/ui/button.tsx`):
filled variants `hover:brightness-[1.12]` and `active:brightness-90`; ghost `hover:bg-border hover:text-ink`; outline
`hover:text-ink`. `brightness(f)` is a CSS filter on the **whole rendered button**, so the oracle transforms the
foreground **and** the fill (and any painted outline) before comparing them (r2): `brightness(f)` is modelled as each
sRGB channel × f, clamped to [0, 255] (the filter's component transfer); S-STATES already measures the painted result,
which the oracle's unit test is cross-checked against on the platform defaults.

Pinned pairs (WCAG 2.x contrast; text 4.5:1, non-text 3:1), each evaluated for Light and Dark effective values:

| # | Foreground on surface | Threshold | Rendered use |
|---|---|---|---|
| T1 | `ink` on `background` | 4.5 | H1/H2 headings (first paint), hover ink of outline/ghost over the page |
| T2 | `text-2` on `background`; `text-3` on `background`; `foreground` on `background` | 4.5 | lead; body; nav links, outline label, indicator text |
| T3 | `ink`, `text-2`, `text-3`, `foreground` on `muted` | 4.5 | headings, lead, body and links inside the band |
| T4 | `muted-foreground` on `background`; on `muted` | 4.5 | index, footer meta (text; the dialog Close icon is N4) |
| T5 | `accent` on `background`; on `muted`; on `popover` | 4.5 | kicker, H3, links, ghost/link buttons, in-dialog links |
| T6 | `popover-foreground` on `popover`; `ink` on `popover`; `text-3` on `popover` | 4.5 | dialog body, dialog title, dialog description |
| T7 | `ink` on `border` | 4.5 | ghost hover (`hover:bg-border` + `hover:text-ink`) and the muted button label |
| T8 | `primary-foreground×f` on `primary×f` for f ∈ {1, 1.12, 0.90} | 4.5 | filled action rest/hover/press (whole-button transform); selected appearance |
| T9 | `secondary-foreground×f` on `secondary×f`, same f | 4.5 | secondary action states |
| T10 | `background×f` on `destructive×f`, same f | 4.5 | destructive label states (K10) |
| N1 | — withdrawn in r2 (was: filled fills against `background`/`popover`) | — | a filled text button is identified by its label (T8–T10); no fill-boundary requirement applies, and the Dark default cannot meet one (below) |
| N2 | `ring` against `background`, `muted`, `popover` | 3.0 | the 2px-offset focus outline against the enclosing page, band or dialog surface it is drawn on; the offset separates it from the control's own fill, which is therefore not its adjacent surface |
| N3 | `input` against `background`, `popover` | 3.0 | essential control boundaries: the outline button's frame (its only boundary) and the input field edge (X8) |
| N4 | `muted-foreground` against `popover` | 3.0 | the existing dialog Close icon at rest (non-text glyph on the popover; `components/ui/dialog.tsx`); its hover is `ink` on `popover` (T6). Replaces the r1 `border`/`background` pair, which was a decorative-fill comparison |

Failure rejects the whole write (`contrast_failed`; closed field = pair id + appearance; never values). The platform
defaults must pass every pair (unit precondition). **Checked before execution** (r2) by source arithmetic over the
`web/app/globals.css` Light/Dark values with WCAG relative luminance and the ×f model above — the lowest remaining
ratios are T8 `primary×1.12` 4.68 (both appearances), T9 Dark `secondary×0.90` 4.77, N2 Light `ring`/`muted` 4.85,
N3 Dark `input`/`popover` 5.43, N4 6.70 Light / 5.43 Dark; every pinned pair passes. The withdrawn r1 comparisons were
impossible for the defaults, which is why they are removed now rather than rediscovered and redefined later: r1 N4
`border`/`background` is 1.39 Light / 1.28 Dark (the muted button fill and ghost hover fill are decorative surfaces
whose label is already held to 4.5 by T7), and r1 N1 is 2.93 for Dark `primary×0.90` against `background` (2.80
against `popover`). Thresholds are unchanged: actual text 4.5, essential boundaries and focus 3.0. If a default fails
at implementation time the pair is corrected in the Result and Plan revision, not the threshold. **Rejection edges**
pinned as unit cases: Light `background` only (roles stay platform Light → T1/T2 fail for a dark background); Dark
`foreground` only (Dark roles follow it, Light untouched); a `primary`/`primary-foreground` pair whose whole-button
×1.12 hover loses the label (T8 f = 1.12 fails while f = 1 passes); `ring` equal to `muted` (N2); `accent` passing on
`background` but failing on `muted` (T5); `muted-foreground` within 3:1 of `popover` (N4). Browser: the branded fixture scope exercises
S-STATES/S-CONTRAST/S-TOKENS as rendered; no permutation matrix.

### 1.4 Mutations, stale edits and the logo lifecycle (R1, R3; corrections 1–3)

Five guarded operations, capability `organization.manage_branding` (D-1 pending), Organization target, inside the
facade's guarded TENANT-ORG transaction; lock order **user → session → Organization (FOR SHARE) → branding row
(`FOR NO KEY UPDATE`; created at version 1 when absent, by Update/SetLogo only) → referenced stored object
(`FOR SHARE`, only when the reference changes)**. The G009 delete path locks only the object row (D1), so the two
orders cannot deadlock; whichever commits first decides (§1.4.2).

| Operation | Input | Effect | Conflict rule |
|---|---|---|---|
| `GetBranding(tc)` | — | view: label, overrides, version, current logo `{object_id, content_type, byte_size, sha12, state}` (state read live: `ready`, `deleting`, `deleted` or `absent` after purge), and **unreferenced logo uploads** (§1.4.2 discovery) | — |
| `UpdateBranding(tc, in)` | `expected_version` (0 = no row), label?, `light`, `dark` | validate (§1.3), write, `version + 1` | `conflict`, zero effect, when the locked version ≠ expected |
| `ResetBranding(tc, v)` | | label null, overrides `{}`, reference null, `version + 1`; deletes no object | as above |
| `SetBrandingLogo(tc, v, body, size)` | bytes | §1.4.1/§1.4.2 | as above, inside A |
| `RemoveBrandingLogo(tc, v)` | | reference null, `version + 1`; deletes no object | as above |

Unreferenced logo uploads are cleaned up only through the existing guarded `DeleteObject` (D-3 as disposed: a thin
authenticated route, same-Organization, current role, restricted to objects the view lists as branding uploads; no
generic object surface). Truthful outcomes (`deleted`/`deletion_pending`) and G013 purge afterwards.

#### 1.4.1 Image validation and normalization (r0 correction 1; r1-review correction 2)

Input is read once into a bounded buffer of at most 524 288 + 1 bytes; a longer body is `invalid_input` before any
store call. Then, in order, all in core with the standard library only (`image`, `image/png`, `image/jpeg`,
`image/draw`; no other decoder registered, no third-party package, no imaging framework):

1. **Sniff**: `http.DetectContentType(first 512 bytes)` ∈ {`image/png`, `image/jpeg`}; the request's claimed type is
   ignored and never stored.
2. **Header bounds**: `image.DecodeConfig` must report the same format and 16 ≤ width, height ≤ 2048 (≤ 4 194 304
   pixels; one decode and one encode per request, under the request deadline).
3. **Structural walk over the whole buffer** (buffering-independent: it never consults the decoder's reader
   position, so decoder read-ahead is irrelevant):
   - PNG, as r1: signature, chunk length ≤ remaining bytes, CRC verified, chunk types restricted to `IHDR PLTE IDAT
     IEND tRNS gAMA cHRM sRGB iCCP pHYs tEXt zTXt iTXt bKGD sBIT hIST tIME`; `acTL`/`fcTL`/`fdAT` (animation) and
     any other type reject; exactly one `IHDR` first and one `IEND`, which must end the buffer.
   - JPEG, the marker walk of the same shape (r2): `SOI`; then segments skipped by their 16-bit length (`APPn`, `COM`,
     `DQT`, `DHT`, `SOFn`, `DRI`, …); after each `SOS` the entropy-coded data runs until a marker byte other than
     `FF00` stuffing or `RSTn`, repeated per scan (progressive); `EOI` must end the buffer. Any byte after `EOI`, a
     segment length past the buffer, or a missing `EOI` rejects. An `APP1` (EXIF) thumbnail is skipped by its
     segment length, so its embedded `EOI` cannot end the walk early.
4. **Complete decode**: `image.Decode` over the buffer must succeed (truncated or corrupt data rejects; progressive
   and EXIF-bearing JPEGs decode normally; the Go PNG decoder itself rejects bad chunk CRCs).
5. **Normalize (re-encode)** (r2; permitted by the Spec and the r0 review): the decoded image is drawn onto an 8-bit
   `image.NRGBA` canvas of the decoded size and re-encoded in the same **family** as detected — PNG →
   `png.Encoder{CompressionLevel: DefaultCompression}` (alpha kept; 16-bit, palette, gray and interlaced sources become
   8-bit RGBA); JPEG → `jpeg.Encode` at quality 90 (progressive, CMYK and YCbCr sources become baseline YCbCr).
   Ancillary data (EXIF, ICC profiles, text chunks, comments) is dropped by construction; the management copy says so.
   The encoder writes through a counting writer that fails beyond 524 288 bytes: a normalized output larger than that
   is `invalid_input` (closed reason `normalized_too_large`) and nothing is stored.

**Stored bytes are the server's own encoder output**, with the server-detected MIME; the caller's bytes never reach
the store, so appended content, polyglots and unknown ancillary data cannot be published even if a walk were wrong.
`v` (§1.5) and the stored SHA-256 are computed over the normalized bytes. **Bounds, pinned (correcting r1's
2048×2048×4 claim):** input ≤ 512 KiB; pixels ≤ 2048²; decoded image ≤ **8 bytes/pixel** — an accepted 16-bit PNG
decodes to `NRGBA64`/`RGBA64`, i.e. up to **32 MiB**, not 16 MiB; 8-bit PNG and CMYK JPEG ≤ 16 MiB; YCbCr ≤ 12 MiB;
the canvas 16 MiB; output ≤ 512 KiB; work = one structural walk over ≤ 512 KiB, one decode and one encode of ≤ 4.2 M
pixels. Peak is therefore about 49 MiB per request, freed at request end; no timing is claimed (PG-008 stays a source
estimate). Concurrency is bounded by the existing capability-guarded route (authenticated owners/admins only); no new
semaphore or worker. No third-party dependency, no SVG/WebP, no resizing or colour management (dimensions are
preserved; only the encoding is normalized).

BRD-9 (§6.2) keeps the header-only/truncated, appended-content (after `IEND` and after `EOI`, including a JPEG with
an `APP1` thumbnail followed by appended bytes), animated-PNG, unknown-chunk, bad-CRC, SVG/HTML/GIF/WebP and
oversize/undersize dimension negatives, and adds the normalization controls: 16-bit, palette and interlaced PNG and
progressive/CMYK JPEG positives (stored bytes decode to the same dimensions, are the normalized family, and differ
from the input), and a `normalized_too_large` rejection. OBJECT-STORAGE.md records the normalization.

#### 1.4.2 Upload → association, uncertainty and recovery (corrections 2 and 3)

`SetBrandingLogo` stages (seam: the existing `objectHooks.stage/crash` test seam, `s.at("branding.set_logo", stage)`;
no new framework):

| Stage | Action | Failure outcome |
|---|---|---|
| **V** validate | §1.4.1 in memory | `input_invalid`; nothing stored |
| **P** upload | `PutObject(tc, {Group: zero, ContentType: detected, Size, Body})` — unchanged G009 path with its own transactions, `object.*` audit, compensation and recovery semantics | any G009 outcome; a failed/unknown upload is handled by G009 (never published, visible to `recovery.sh status`) |
| **A** associate | guarded write transaction: lock/create the row, check `expected_version`, lock the new object `FOR SHARE` and re-read it under RLS (`ready`, same Organization, type/size as validated), set the reference, `version + 1` | a **definitive** abort — `conflict`, `not_found`, a guard/constraint failure, or a store error returned before `COMMIT` was sent, so the transaction is known rolled back → **C**; `commit_outcome_unknown` — the reply to `COMMIT` was lost, or any error whose rollback is not known → **R** |
| **R** report (r2; no reconciliation verdict) | one guarded read (`GetBranding`-style, snapshot) of the branding row and the object's live state, taken after the unknown outcome. A snapshot observes only what has already committed: it can **confirm** a commit but can never **prove** non-commit — A may still be in flight, or a later writer may have advanced the row. So: (i) the row references the new object → A committed → return the view (no compensation); (ii) otherwise — the row is still at the pre-A version, or its version has **advanced** without referencing the object (A may have committed and been replaced, or never committed), or the read itself fails — return `unavailable` with the closed field `publication_outcome_unknown`, the object id and the object's **actual** live state as read (`ready`, `deleting`, `deleted`, or `unread` when the read failed). The object is **retained** untouched; C never follows R. `GetBranding` lists it live afterwards — as the current logo if A lands later, otherwise as an unreferenced `ready` upload that the owner may delete through the guarded route or supersede by a new `SetBrandingLogo`. No background or deferred action exists; the next owner read shows the true state | — |
| **C** compensate | `DeleteObject(tc, id)` under the same context, only after a **definitive** abort known from A itself; the result is reported in the error's closed field exactly as G009 returned it (`compensated` for `deleted`, `compensation_pending` for `deletion_pending`) | compensation failure: the error carries the object's actual state as G009 returned it — `ready` and listed as unreferenced when the delete did not take effect, `deletion_pending` only when G009 said so; nothing published; no retry loop |

Invariants: no double publication (A is bound to `expected_version`; a second A conflicts); a possibly-committed
association never triggers deletion (C only after a definitive abort known from A itself; R never yields a non-commit
verdict — mirroring G009 I-FIN/R); **no possibly published bytes are deleted to make an error look cleaned up**;
cancellation between P and A leaves a listed unreferenced upload and no publication; request bodies are read once;
no transaction spans a provider call.

**Discovery without a ledger.** The r0 uploads ledger is removed. `GetBranding` lists as "unreferenced logo uploads"
the Organization's `stored_objects` rows with `group_id IS NULL`, `content_type IN ('image/png','image/jpeg')`,
`state IN ('ready','deleting')` and `id <> current reference` (one indexed query on the existing Organization index
of `stored_objects`; name confirmed at R0; bounded by LIMIT 100 with a count). Since G017 is the first production
writer of the store, every such object was created by `SetBrandingLogo`; OBJECT-STORAGE.md records this assumption
and that a later uploader must add a purpose discriminator before reusing Organization-scoped image objects.

**Deletion and retention (correction 3).** No foreign key references `stored_objects` from branding: the guard trigger
(§4) validates a *changed* reference at write time; the projection and the views join `stored_objects` on
`(id, organization_id)` and expose the logo only while `state = 'ready'`. Consequences, all finite: a guarded
`DeleteObject` of the current logo (owner/admin) proceeds unchanged (D1 marks `deleting`; the shell falls back to the
text mark from the next request; `GetBranding` shows the reference with state `deleting`/`deleted`; `RemoveBrandingLogo`
clears it); G013 purge of the tombstone proceeds unchanged (the reference then reads `absent`); reset/removal/replace
leave objects private and deletable; concurrent delete vs associate is serialized on the object row (A's `FOR SHARE`
against D1's `FOR NO KEY UPDATE`): if the delete wins, A sees a non-ready object and takes **C**. No cascade, no
runtime DELETE grant beyond G009's, no cleanup service. Required additional paths: none in `infra/recovery`
(`retention-objects.sql` is unchanged).

### 1.5 Public projection, shell and indicators (R4, R5; D-6)

As r0 with these pins: the projection answers `{revision, display_label?, light, dark, logo?: {v, content_type}}` —
`v` is the first 12 hex characters of the object's SHA-256; no key, id or size leaves core's public routes. The tenant
`/` entry renders identity (logo mark or label/product name), one neutral "Sign in" action, the appearance control,
keyboard help and the configured attribution; nothing else (C-2 as directed). The authenticated indicator reads the
**existing** session route extended with `membership: "member" | "none"`, computed inside the same core call by one
`app.is_organization_member(user, org)` statement on the request's own origin binding (A2: a credential issued for
another host is `unauthenticated` here as today — the indicator can never make a foreign-host credential usable). Pins:
`core/internal/identity/{identity,http,service}.go` (+ `http_test.go`, `service_test.go`), `web/lib/identity-client.ts`
(`SessionView.membership`), `web/tests/unit/identity-client.test.mjs`, `completion.schema.json` unchanged for this
route. Cost: +1 statement inside the existing session call; no extra round trip.

### 1.6 Appearance and accessibility — as r0 (no new font, geometry, motion, library; brand mark in
`components/layout/`).

### 1.7 Options — as r0; the ledger option is withdrawn (§1.4.2).

## 2. Source binding (provisional; rebind at R0)

| Input | Revision | SHA-256 |
|---|---|---|
| Base: `origin/staging` `b9b6e5c0` merged into `g017-plan` (G016 `0a163c7f` content retained) | — | — |
| Goal `management/v0/goals/G017-organization-branding-logged-out-shell.md` | `b9b6e5c0` | `aefe78c82c0cc4c3a6d4a771e13f26b9287a28db6b6291ab7d7c355b45f74058` |
| Spec `management/v0/specs/G017-organization-branding-logged-out-shell-spec.md` | `b9b6e5c0` | `c90d4cdef488030312451ee259a02bc0ed662472686802a08d9600734ba6c3f9` |
| r0 review `management/v0/results/G017-plan-r0-codex-review.md` | `b9b6e5c0` | `c88b8822f5f0f0fdbfa42ebccb4be40b144970a8fd06f54d8ca193598459f7f8` |
| Allow-list proposal r1 `management/v0/plans/G017-scope-allowlist.txt` | this commit | §13 |

## 3. Files (touch-list; corrections 2, 3, 5 and C-3 applied)

**Create**

| Path | Content |
|---|---|
| `core/migrations/00013_organization_branding.sql` | §4 |
| `core/internal/branding/{branding,colors,image}.go` + `_test.go` | pure validation: label, grammar, effective palette and the T/N oracle, PNG chunk walk, decode/trailing policy; stdlib only |
| `core/internal/branding/api/http.go` + `http_test.go` | routes §5.3; **bounded copy** of the G016 session-scope behaviour (`begin`/`scope`/`failure`, D-5), with a pin test asserting byte-equivalence of the copied refusal table to `domains/api` |
| `core/internal/tenancy/branding.go` + `branding_test.go` | **the SQL** (`sqlBranding*` constants and `Store` methods: get/lock/create/update, logo-state join, unreferenced-upload listing, the definer-read wrappers); joins the `TestSQLBoundary`/`TestObjectSQLBoundary` pattern (correction 5) |
| `core/internal/tenancy/access/branding.go` + `branding_integration_test.go` | guarded ops, `PublicBranding`, `PublicBrandingLogo`; no SQL literals (`access/boundary_test.go` rule) |
| `core/internal/migrate/branding_integration_test.go` | `TestBrandingCatalog` |
| `web/lib/branding.ts`, `web/tests/unit/branding-client.test.mjs`, `web/app/branding/logo/route.ts`, `web/app/organization/branding/{page.tsx, update/route.ts, reset/route.ts, logo/route.ts, logo/remove/route.ts, objects/delete/route.ts}`, `web/components/layout/{brand-mark,context-indicator}.tsx`, `web/components/branding/*.tsx`, `web/tests/unit/context-indicator.test.mjs`, `web/tests/fixtures/branding/*` (test images) | as r0 |
| `management/v0/results/G017-organization-branding-logged-out-shell-result.md`, `G017-scope-allowlist.txt` | Result; imported allow-list |

**Modify (production)**: `core/cmd/server/main.go` (store from `config.Load`, `access.NewServiceWithObjects`, branding
handler), `core/internal/tenancy/access/{policy,errors,log,access}.go`, `core/internal/tenancy/store.go` (only if the
store interface needs the new methods declared there), `core/internal/audit/audit.go`, `core/internal/identity/
{identity,http,service}.go`, `web/components/layout/app-shell.tsx`, `web/app/{page,sign-in/page,account/page,
connections/mcp/page}.tsx`, `web/lib/{identity-client,tokens}.ts`, `infra/identity/mcp-seed.sql`. **No** `domains/api`
or shared-helper change (D-5); `.env.example` only if R0 finds a variable is needed (none expected).

**Modify (tests, pins, harness, docs)**: `core/internal/tenancy/{boundary,parity}_test.go` (SQL/table pins),
`core/internal/tenancy/access/{boundary,matrix,policy,rls_expected,attack_expected}_test.go` (importers, capability,
table sets), `core/internal/audit/audit_test.go` (36), `core/internal/identity/{http,service,store,rls_pool_integration}
_test.go`, version-13 pins (`core/internal/migrate/{runner,integration,mcp_integration,rls_integration,domains_integration}
_test.go`, `core/migrations/migrations_test.go`, `core/internal/tenancy/integration_test.go`, `core/internal/tenancy/
access/recovery_integration_test.go`), `web/tests/fixtures/token-scopes/{fixtures.ts,page.tsx}`, `web/tests/browser/
{identity,tokens}.spec.ts`, `web/tests/browser/support/modes.ts`, `web/tests/unit/identity-client.test.mjs`,
`infra/validation/run.sh`, `required-checks.tsv`, `infra/ci/validate-controls.sh` (two rows, count pins),
`infra/validation/tools/internal/logrecord/{events.json,completion.schema.json,logrecord_test.go}` (C-3, required),
docs `docs/foundation/{WEB-SHELL,OBJECT-STORAGE,AUTHORIZATION,AUDIT,DATABASE-ISOLATION,MIGRATIONS,DOMAINS,RECOVERY}.md`,
`docs/{SECURITY,PERFORMANCE}.md` (PG-008 source estimate), `docs/ARCHITECTURE.md` only on Codex request.

## 4. Schema and migration 00013 (provisional number; corrections 3 and 4)

1. `CREATE FUNCTION app.branding_overrides_valid(jsonb) RETURNS boolean IMMUTABLE` — ≤ 15 keys from the pinned list,
   values `~ '^#[0-9a-f]{6}$'`.
2. `CREATE TABLE app.organization_branding (organization_id uuid PRIMARY KEY REFERENCES app.organizations(id),
   display_label text CHECK (…1–80, no [[:cntrl:]]…), logo_object_id uuid, light jsonb NOT NULL DEFAULT '{}' CHECK
   (app.branding_overrides_valid(light)), dark jsonb … , version bigint NOT NULL DEFAULT 1 CHECK (version >= 1),
   updated_at timestamptz NOT NULL DEFAULT now(), updated_by uuid NOT NULL REFERENCES app.users(id))`. **No foreign
   key to `stored_objects`** and no uploads ledger (§1.4.2); no change to `stored_objects` (the r0 unique constraint
   is withdrawn).
3. Guard `app.organization_branding_guard` BEFORE UPDATE: `organization_id` immutable; `version = OLD.version + 1`;
   when `NEW.logo_object_id IS DISTINCT FROM OLD.logo_object_id AND NEW.logo_object_id IS NOT NULL`, the referenced
   `stored_objects` row must exist with the same `organization_id`, `state = 'ready'`, `content_type IN ('image/png',
   'image/jpeg')` and `byte_size <= 524288` (the trigger runs as the table owner's definer-style function with
   `search_path` pinned, as the G016 guards do). The same check runs BEFORE INSERT when a reference is present.
4. Audit capture `audit.capture_organization_branding()` AFTER INSERT OR UPDATE → one `organization.branding_changed`
   (`reason` ∈ {`updated`, `reset`, `logo_set`, `logo_removed`}; no values); `audit.events_attribute()` version-13
   branch (down restores the version-12 body).
5. Runtime grants: `SELECT`, `INSERT (organization_id, display_label, logo_object_id, light, dark, version,
   updated_by)`, `UPDATE (display_label, logo_object_id, light, dark, version, updated_at, updated_by)`; no DELETE.
   `ENABLE` + `FORCE ROW LEVEL SECURITY`; runtime policies `organization_branding_{select,insert,update}` with the same
   Organization-context predicate as `organization_domains_*` (00011).
6. **Auth-reader path (correction 4)** — the helper role is not BYPASSRLS, so grants alone read nothing under FORCE
   RLS; the exact policies:
   - `GRANT SELECT (organization_id, display_label, logo_object_id, light, dark, version) ON app.organization_branding
     TO <db>_auth_reader`; `CREATE POLICY organization_branding_auth_reader ON app.organization_branding FOR SELECT TO
     <db>_auth_reader USING (true)` (the row content is intentionally public; the HTTP layer answers only for the
     Organization the request's host routes to).
   - `GRANT SELECT (id, organization_id, state, content_type, byte_size, sha256, storage_key) ON app.stored_objects TO
     <db>_auth_reader`; `CREATE POLICY stored_objects_auth_reader ON app.stored_objects FOR SELECT TO <db>_auth_reader
     USING (EXISTS (SELECT 1 FROM app.organization_branding b WHERE b.logo_object_id = stored_objects.id AND
     b.organization_id = stored_objects.organization_id))` — the reader sees **only rows that are some Organization's
     current logo**; ordinary objects stay invisible to it; runtime/tenant policies and denials are unchanged.
   - `CREATE FUNCTION app.public_branding(p_org uuid) RETURNS TABLE (revision bigint, display_label text, light jsonb,
     dark jsonb, logo_sha256 bytea, logo_content_type text, logo_size bigint, logo_key text) SECURITY DEFINER` owned by
     the auth reader (created under `SET LOCAL ROLE <db>_auth_reader` with the transient `GRANT CREATE ON SCHEMA app`,
     as 00011), joining the logo only where `state = 'ready'`; `REVOKE ALL FROM PUBLIC`; `GRANT EXECUTE` to runtime.
     `logo_key`/`sha256` reach core only; the public HTTP output excludes key, id and size (§1.5).
   - Catalog proof CT-BRD-1 pins both policies and grants; runtime proof BRD-11 executes, as `<db>_auth_reader` (via
     `SET ROLE` in the integration test as the G008/G016 catalog tests do), a `SELECT` by id of a non-logo object (0
     rows) and of the current logo (1 row). **Corrected wording (r2):** `public_branding(B)` legitimately returns B's
     intentionally public projection, so foreign-logo rejection is not a property of the function and is not claimed
     of it. It is tested where it holds: at the trusted-host resolution — a request on a host that G016 `Lookup`
     resolves to Organization A asking for B's `v` → `not_found` — and on the private paths: B's object through A's
     runtime context (`GetObject`, the unreferenced-upload listing) keeps its unchanged RLS/tenant denial.
   - **C-5 disposed (r1 review):** these explicit helper-role policies are the chosen design. A SECURITY DEFINER
     function alone cannot bypass FORCE RLS for its non-BYPASSRLS owner, so the r1 "probe-function fallback" (§10) is
     withdrawn. The policy dependency graph is acyclic (`stored_objects_auth_reader` → `organization_branding`, whose
     helper policy is `USING (true)`); ordinary objects stay invisible; runtime tenant denials and the
     selected-logo-only read are preserved.
7. Down: refuse while any branding row or `organization.branding_changed` event exists; drop policies, grants,
   function, triggers, table, validator; restore `events_attribute()` version 12.

Locks: one `CREATE TABLE`, policy/grant DDL, `CREATE OR REPLACE FUNCTION audit.events_attribute` (brief ACCESS
EXCLUSIVE on the function as in 00011/00012); no data migration, no change to existing tables' rows or constraints.

## 5. Contracts and interfaces

### 5.1 Go surfaces

Operations as r0 (`GetBranding`, `UpdateBranding`, `ResetBranding`, `SetBrandingLogo`, `RemoveBrandingLogo`,
`PublicBranding`, `PublicBrandingLogo`) with the r1 semantics of §1.4. Query layer (correction 5): all SQL in
`core/internal/tenancy/branding.go` as `sql*` constants naming only `app.organization_branding`, one join constant on
`app.stored_objects` for the live logo state, one listing constant on `app.stored_objects`, and the two definer-function
calls; `access/branding.go` calls those store methods only. `tenancy/boundary_test.go` gains `TestBrandingSQLBoundary`
(file `branding.go`, constant prefix, relation set) beside `TestObjectSQLBoundary`; `access/boundary_test.go`'s
SQL-literal rule is unchanged and must pass. No sqlc.

Errors: reuse `unauthenticated`, `context_denied`, `forbidden`, `not_found`, `invalid_input`, `conflict`,
`unavailable`; new `contrast_failed`; closed fields `publication_outcome_unknown` (with object id and actual live
state), `compensated`, `compensation_pending`; `invalid_input` reason `normalized_too_large` (r2; `upload_unreferenced`
is withdrawn).
Operations: `branding.get/update/reset/set_logo/remove_logo`; log event `branding.request` (closed outcome/reason, host
as the G016 12-hex tag) and `branding.public` for the anonymous reads.

### 5.2 Capability — `organization.manage_branding` (thirteenth; **D-1 pending**: owner+admin recommended, owner-only
the alternative; the matrix row is written only after the owner decides).

### 5.3 HTTP — as r0 (public `GET /internal/v1/branding/public?origin=` and `GET /internal/v1/branding/logo?origin=&v=`;
session `GET /internal/v1/branding`, `POST /internal/v1/branding`, `/reset`, `/logo` (raw body, `Content-Length` ≤
524288 else 413, `X-Expected-Version`), `/logo/remove`, `/objects/delete {object_id, expected_version}`; existing
`GET /internal/v1/auth/session` + `membership`). All routes join the `completion.schema.json` enums (C-3). No MCP,
public API or `contracts/` change.

### 5.4 Web — as r0 (`lib/branding.ts`, `AppShell.brand`, pages, management forms following `connections/mcp`).

### 5.5 Cache contract (R4, AC8; truthful statement)

| Response | Cache-Control | Contract |
|---|---|---|
| HTML pages, JSON routes, refusals | `no-store` | never cached; indicators never shared |
| `GET /branding/logo?v=<sha12>` with `v` current and the host routing | `public, max-age=86400, immutable`, `ETag`, `nosniff`, `inline` | **bounded cached-byte contract**: a browser or intermediary that holds a fresh copy serves it without contacting the origin for up to 24 h; the origin is **not** consulted on a fresh hit, so removal, replacement, suspension or domain removal does **not** stop already-cached bytes within that window. What the origin guarantees is separate and tested separately: every request that reaches it is re-resolved and re-associated (new or stale `v`, removed/suspended host, deleted logo → 404 `no-store`), and every newly rendered page references only the current `v` (so no new page can point at removed bytes). No retroactive erasure, no in-flight-race guarantee |
| stale/missing `v`, non-routing host | `no-store` 404 | never cacheable |

**C-4 disposed (r1 review):** the explicitly disclosed 24 h cache lifetime for intentionally public logo bytes is
retained as the contract above; origin-request checks and fresh cache hits remain separate guarantees, tested
separately (`http_test`, E-BRD-5/6). The r1 `no-cache` revalidation alternative is withdrawn; no revalidation feature
is added.

### 5.6 PG-008 (source estimate; no benchmark, no gate)

Per tenant page: +1 web→core round trip = `Lookup` (1 unique-index probe) + `public_branding` (1 PK probe + 1 PK join
on `stored_objects`). Authenticated pages: +1 `is_organization_member` probe inside the existing session call. Logo
origin fetch: 2 statements + 1 provider GET ≤ 512 KiB; cache hits: 0. Management `GetBranding`: 1 row lock/read + 1
indexed listing (`stored_objects` Organization index, LIMIT 100). Writes: one short guarded transaction (locks §1.4);
`SetBrandingLogo` adds the unchanged G009 Put footprint (PG-003) plus A (and R/C on failure). Migration: one new
table, no index build on existing tables. `TestBranding` BRD-12 records the statement counts with the existing
`e.hook.counts()` mechanism; unexpected scans or lock amplification escalate. Image normalization (§1.4.1) is
per-request CPU/memory inside core, not a database cost; its bounds are part of this source estimate and carry no
benchmark gate (r1 review: PG-008 remains a source estimate only).

## 6. Validation and test mapping

### 6.1 Existing coverage inspected — as r0, plus `TestSQLBoundary`/`TestObjectSQLBoundary`, the `identity:browser`
command (IDENTITY.md; projects `domain-platform`/`domain-custom` selectable) and the `platform:browser` runner (README;
accepts `--project` flags, as `run.sh` invokes it).

### 6.2 Scenario groups → tests (reuse / extend / new)

| Case | Kind | Test | Proves |
|---|---|---|---|
| CT-BRD-1/NEG/UPDOWN | **new** `TestBrandingCatalog` (`./internal/migrate/`) | exact version-13 delta on the unchanged version-12 catalog (table, validator, guard, capture, runtime policies, **both auth-reader policies**, definer function, grants), rolled-back missing-protection negatives, down refusal | AC1, AC5 |
| BRD-1..4 | **new** `TestBranding` (`./internal/tenancy/access/`) | owner/admin update/reset; member/Group-owner/foreign denial with `access.denied`; invalid/oversized/unknown-key/bad-grammar inputs commit nothing | AC1 |
| BRD-5..7 | **new** | stale version on update/reset/logo → `conflict` (state unchanged); two concurrent updates; reset preserves objects; audit rows without payload | AC2 |
| BRD-8 | **new** | PNG and JPEG accepted; stored bytes are the normalized output (same dimensions, same family, server MIME, differ from the input); `v` is over the stored bytes; object ready; reference set; view lists it as current | AC5 |
| BRD-9 | **new** | rejections: header-only/truncated PNG and JPEG, appended bytes after `IEND`/`EOI` (incl. a JPEG with an `APP1` thumbnail then appended bytes), animated PNG (`acTL`), unknown/private chunk, bad CRC, SVG, HTML, GIF, WebP, 15 px and 2049 px dimensions, 524 289 bytes, `normalized_too_large`; positives: 16-bit/palette/interlaced PNG and progressive/CMYK JPEG normalized; zero effects, no pending row left (or G009-compensated) | AC5 (r0 correction 1; r1 correction 2) |
| BRD-10 | **new** (seam `branding.set_logo` stages; the existing `objectHooks.stage/crash` seam only) | crash after P (listed unreferenced, not published); definitive conflict at A (compensated, reported); `commit_outcome_unknown` at A with committed association (R returns the view, **no** deletion); unknown at A while A is still **in flight** (the A-stage callback runs R's read from the test before A's `COMMIT` is sent, then lets A commit: `publication_outcome_unknown`, object retained, and the next `GetBranding` shows it as the current logo — no deletion of published bytes); unknown at A with a genuine rollback (`publication_outcome_unknown`, object retained and listed `ready`; the owner's guarded delete then succeeds); unknown at A with an **advanced version** that does not reference the object (a concurrent update: `publication_outcome_unknown`, retained); R's read failing (`publication_outcome_unknown`, state `unread`); compensation failure (actual G009 state reported, listed); cancellation between P and A | AC2, AC5 (r0 correction 2; r1 correction 1) |
| BRD-11 | **new** | `PublicBrandingLogo`: current ready logo only; wrong `v`, unreferenced/deleting/deleted object → not_found; a host resolving to A asking for B's `v` → not_found (trusted-host resolution); as `<db>_auth_reader`: non-logo object invisible, current logo visible, `public_branding(B)` returns B's public projection by design; B's object under A's runtime context denied as today | AC5, AC6 (r0 correction 4; r1 C-5) |
| BRD-12 | **new** | statement counts (PG-008) | AC9 |
| BRD-13 | **new** | session answer `membership` flips to `none` when the Organization binding is removed in-test; foreign-origin credential stays `unauthenticated` | AC7 (C-1) |
| Current-logo delete / purge | **new** (BRD-14) | guarded `DeleteObject` of the current logo → projection falls back, view shows `deleting`/`deleted`; after a simulated purge (row removed by the owner role in-test) → `absent`; `RemoveBrandingLogo` clears it | AC5 (correction 3) |
| U-BRD-* | **new** unit (`./internal/branding/`) | grammar/injection negatives; effective palette with inherited roles; T1–T10/N2–N4 with whole-button ×1.12/×0.90 transforms; the §1.3 rejection edges; platform defaults pass; PNG chunk walk, JPEG segment walk and normalization (family, dimensions, size cap, ancillary data dropped) on small synthetic files | AC3 (correction 6) |
| Capability/pins | **extend** `TestAccessPolicy`, `TestToolEligibility` (empty-Organization pin), `TestProductionImporters`, `audit_test.go` (36), `rls_expected`/`attack_expected` table sets, `TestBrandingSQLBoundary`, `TestMigrationParity` | the thirteenth capability and new table join accepted invariants | AC1 |
| `http_test.go` (branding/api) | **new** | origin/session/CSRF/foreign-binding refusals, 413, exact public field set, logo headers, `v` mismatch, copied refusal-table equivalence with `domains/api` (D-5) | AC6, AC8 |
| Web unit | **new** `branding-client.test.mjs`, `context-indicator.test.mjs`; **extend** `identity-client.test.mjs` | closed reading, grammar re-check, no credential on public calls; indicator states incl. `none` | AC6, AC7 |
| Browser fixture | **extend** `token-scopes` with a `branded` scope → existing S-TOKENS/S-PORTAL/S-PORTAL-INIT/S-STATES/S-CONTRAST/S-TEXT-SPACING | rendered pairs/states and sibling isolation | AC3, AC4 |
| `appearance` project | **reuse** unchanged | branding never sets the mode | AC4 |
| E-BRD-1..6 (`domain-platform`, `domain-custom`) | **new** in `identity.spec.ts` | E-BRD-1 branded A at `portal.acme.test` (escaped label, logo bytes/headers, Light/Dark); E-BRD-2 unbranded C fallback; E-BRD-3 removed/unknown hosts stay 404 (reuse E-DOM-4/5 assertions); E-BRD-4 member indicator after sign-in; E-BRD-5 management: upload PNG (accepted) and SVG (rejected, nothing changes), replacement changes `v`, old `v` → 404 at the origin; E-BRD-6 interleave anonymous A / anonymous B / member A / anonymous A: exact per-host bodies and cache headers, old-`v` origin refusal | AC6, AC7 (member), AC8 |

Expected counts (pin at R0): `TestBrandingCatalog` 3; `TestBranding` 14; `branding` unit 5; E-BRD 6.

### 6.3 Batches (focused executable selectors; correction 7)

Before every batch: fetch `origin/staging`, assess the delta, rebase/reconcile, `bash infra/ci/docker-desktop-mounts.sh`,
the standard `run.sh --base <sha> --allow-file management/v0/results/G017-scope-allowlist.txt` form; no edits during a
running batch; finish, fix, rerun, advance.

| Batch | Selectors | Why each |
|---|---|---|
| b1 unit/static | check image `docker build --target check -t g017-core-check:b1 -f infra/docker/core.Dockerfile core`; `docker run --rm --network none --entrypoint sh g017-core-check:b1 -c 'gofmt -l . && go vet ./internal/branding/... ./internal/tenancy/ ./internal/tenancy/access/ ./internal/identity/ ./internal/audit/ ./cmd/server/ && go vet -tags integration ./internal/tenancy/access/ ./internal/migrate/ && go test -count=1 ./internal/branding/... && go test -count=1 -run "^(TestAccessPolicy\|TestToolEligibility\|TestProductionImporters\|TestSQLBoundary\|TestObjectSQLBoundary\|TestBrandingSQLBoundary\|TestPackageBoundary\|TestMigrationParity)$" ./internal/tenancy/ ./internal/tenancy/access/ && go test -count=1 ./internal/audit/ -run "^TestCatalog"'` (exact audit test name at R0); web-check image: `node --conditions=react-server --test tests/unit/branding-client.test.mjs tests/unit/context-indicator.test.mjs tests/unit/identity-client.test.mjs`; `npm run lint && npm run typecheck` | changed packages/files only |
| b1 harness | `run.sh lint lint-base lint-web lint-core` | B-WEB-TOKENS/COMPONENTS/DESIGN/BOUNDARY over changed web sources, B-SECRET-CANARY (lint with lint-base), B-MIGRATE-SET, B-SCOPE |
| b2 focused integration | focus database `g017-focus`: `go test -tags integration -count=1 -timeout 30m -v -run '^(TestBrandingCatalog)$' ./internal/migrate/` plus **only** the test functions in the version-13 pin files whose pins change (names recorded at R0, e.g. the runner/migrations-set pins); `… -run '^(TestBranding)$' ./internal/tenancy/access/`; `… -run '^(TestRLSIdentityPool)$' ./internal/identity/` and the identity `http`/`service` unit tests for the session answer. `TestAuditCatalog`, `TestRLSCatalog`, `TestIsolationAttacks`, `TestDomains*`, `TestObjects*`, `TestMCP*` are **not** run by default: their inputs are unchanged except where a pin file changes, in which case exactly that test runs (recorded with the changed pin as the reason). Attributable existing evidence and the hosted full gate cover the unchanged families | new behaviour, changed guards/pins, narrow affected regressions |
| b3 browser fixture/appearance | README `platform:browser` command with `--project=tokens-light --project=tokens-dark --project=appearance` appended (the runner passes its arguments to `playwright test`, as `run.sh` and `identity:browser` do) | branded scope rendering; appearance unchanged |
| b4 browser tenant hosts | IDENTITY.md `identity:up`/seed blocks, then the `identity:browser` command **narrowed** to `--project=domain-platform --project=domain-custom`; `identity:stop` afterwards | E-BRD-1..6 at real HTTPS hostnames through the existing overlay — the entrypoint G016 already used; **no `run.sh ops`, no root umbrella** |
| S-review | read-only (§7) | findings → rerun only affected selectors |
| manual | §6.4 | before final qualification |
| final (local) | no new suite run; reconcile the delta since each attributable pass; rerun only changed selectors; Result | attributable coverage |

Harness delta: `branding_checks()` beside `domains_checks()` with **B-BRANDING-CATALOG** and **B-BRANDING-IT**
(`required-checks.tsv` 119 → 121; `validate-controls.sh` pins). Path selection: migrations/tenancy/web shared paths
select **full** for the MR; the post-merge staging full gate supplies hosted rows after an authorized conditional
merge. No technical limitation is claimed: every needed guarantee has a supported narrow selector above.

### 6.4 Manual review, readiness and the early draft MR (correction 8)

- **Preview environment (existing, disposable, local, emulator-backed; r2):** the identity overlay in its **L2
  emulator** form (IDENTITY.md "The emulator (L2)"), not the L4 WorkOS procedure: the `identity:start` block with
  `IDENTITY_PROXY_PUBLISH=127.0.0.1:543` in the environment (the overlay's documented publish variable; its default is
  a random loopback port), then the `identity:migrate` block, then the compose `cp identity-setup:/identity/ca.pem
  <file>` line to export the run CA, trusted only in a dedicated browser profile. Sign-in uses the emulator's seeded
  users (`browser/users.json` in the run volume, written by `identitysetup`: password sign-in, no TOTP/MFA). **Not
  required and not used:** `IDENTITY_MANUAL_ENV_FILE`, WorkOS staging client/issuer/API-key settings,
  `TestManualQualification`/`workos_manual`, the MQ steps and Radar. The reviewed preview inputs are exactly: the
  candidate SHA, `.env` from README, the emulator seed digest, the run CA path and the seeded-user file. Reachable URLs
  (resolve natively in Chromium/Firefox without hosts-file edits): `https://org-a.localhost:543/` (branded by the
  seeded row in `mcp-seed.sql`), `https://org-c.localhost:543/` (no-brand fallback), `https://org-a.localhost:543/
  organization/branding` (management, signed in as a seeded owner). The custom host `portal.acme.test` is covered by
  E-BRD-1 automation; it is not required for manual review (it would need a hosts-file entry, which this Plan does
  not assume). No mobile device, public tunnel or hosted environment is assumed or authorized; narrow viewports are
  reviewed with the desktop browser's device toolbar at 320 px.
- **Readiness handoff:** Claude starts the overlay at the exact candidate SHA, records the SHA, seed digest and CA
  path, points the operator at `browser/users.json` for the seeded credentials (values never in chat/logs), and asks
  for explicit readiness confirmation before any timer. Feedback items (Light/Dark/System, 320/1280, dialog in the
  tenant scope, sign-in as member, logo upload and SVG rejection) are recorded in the Result bound to that SHA.
- **Resolution:** findings → fixes → rerun affected E-BRD/S-*/unit selectors → repeat M1 only for changed surfaces.
  **Cleanup:** `identity:stop` (`down -v`) and the owned-resource receipt.
- **Sequencing:** the **draft implementation MR is opened as soon as slice 2 compiles and its focused checks pass**
  (its creation is G018's provisional-planning trigger); manual feedback (M1) is completed before final local
  qualification, Codex pre-integration acceptance and any merge. Screenshots in the Result support visual feedback
  only; interactive outcomes (upload, sign-in) are attested by the operator and the automated E-BRD cases.

## 7. Delegation — as r0 (one implementer; two optional read-only reviewers R-A after slice 2, R-B after slice 3;
task-prefixed worktree rule carried in delegated instructions).

## 8. Sequence

0. R0 after G016 integration; rebind; Codex Plan ACCEPT; D-1 owner answer; allow-list import; release.
1. Slice 1: `branding` package + units; `tenancy/branding.go` SQL + boundary pins; migration 00013; capability (per
   D-1); `access/branding.go` Get/Update/Reset/RemoveLogo; audit T30; `TestBrandingCatalog`; BRD-1..7, 14; version-13
   pins. b1, b2.
2. Slice 2: image validation; `SetBrandingLogo` with R/C; `PublicBranding`/`PublicBrandingLogo`; `branding/api`
   (bounded scope copy); server storage wiring; logrecord enums; BRD-8..12; `http_test`. b1, b2. **Open the draft
   MR.** R-A review.
3. Slice 3: `lib/branding.ts`, shell/brand mark/entry page, session `membership` + indicator (BRD-13), fixtures,
   `modes.ts`, web units. b1, b3. R-B review.
4. Slice 4: management page/routes incl. the unreferenced-upload delete route; seed; E-BRD-1..6; harness rows. b4.
5. Slice 5: docs; M1 manual feedback; reconcile; final targeted reruns; Result; safekeeping push; MR to ready after
   Codex pre-integration acceptance.

## 9. Acceptance-criteria mapping

| AC | Evidence |
|---|---|
| AC1 | CT-BRD-1; BRD-1..4; capability/eligibility/importer/SQL-boundary pins; AUTHORIZATION row (after D-1) |
| AC2 | BRD-5..7, BRD-10 (unknown outcome retained and reported; compensation only on a definitive abort), audit without payload |
| AC3 | U-BRD contrast/grammar with inherited roles and transformed states; BRD-4; S-CONTRAST/S-STATES over the branded scope |
| AC4 | `appearance` unchanged; S-TOKENS/S-PORTAL/S-PORTAL-INIT/S-TEXT-SPACING over the branded scope; M1 |
| AC5 | BRD-8..11, 14; `http_test` headers; E-BRD-1/5 |
| AC6 | E-BRD-1..3; `http_test` public field set; canary private records (Organization name, member names, challenge, storage key) absent from body/headers/logo/logs (`support/sanitize.ts`, identity log checks) |
| AC7 | E-BRD-4 (member); BRD-13 + `context-indicator` unit (non-member/revoked); identity/MCP suites untouched |
| AC8 | E-BRD-6; `http_test` cache headers; §5.5 contract documented in WEB-SHELL/OBJECT-STORAGE |
| AC9 | B-WEB-DESIGN/B-WEB-TOKENS; BRD-12 → PG-008 source estimate updated with counts |
| AC10 | Result with commands/outcomes/provenance, M1 record, MR/staging disposition, cleanup receipt |

## 10. Decisions, risks and escalations

| ID | Status after the r0 and r1 reviews |
|---|---|
| **D-1** | **OPEN — owner decision pending** (unchanged in r2; no answer assumed): `organization.manage_branding` for Organization owner **and** admin (recommended) or owner-only. Not approved by the recommendation; implementation cannot start before the answer |
| D-2, D-9, D-10, C-2 | disposed by the review as Spec-conforming; applied as written |
| D-3 | disposed: narrow authenticated delete route reusing guarded `DeleteObject`, restricted to listed branding uploads; no recovery promised beyond what the listing supplies (§1.4.2) |
| D-4, D-7, D-8 | Spec-authorized narrow logo projection, auth-reader path and existing-store wiring; **safety design accepted only when corrections 1–4 close**; no new dependency, bucket policy or live infrastructure |
| D-5 | bounded exact copy of the G016 scope behaviour in `branding/api` with an equivalence pin test; extraction patterns removed from the allow-list |
| D-6 | kept: internal membership fact on the existing session read, exact origin/Organization binding, no authority expansion; pins and cost in §1.5 |
| Cache (C-4) | **disposed by the r1 review**: the disclosed 24 h bounded cached-byte contract is retained (§5.5); the revalidation alternative is withdrawn |
| C-5 | **disposed by the r1 review**: explicit helper-role RLS policies (§4 item 6); the definer-only probe fallback is withdrawn |
| PG-008 | source estimate (§5.6); no benchmark, no gate (reaffirmed by the r1 review) |
| r1-review corrections 1–3 | applied in r2 (§1.4.2, §1.4.1, §1.3); thresholds, seam, schema and scope unchanged |

Escalation risks: G016 contract drift at R0; D-1 answer changing the matrix; any need for SVG/WebP, resizing or
colour management, remote fetch, a cache service or Organization-name exposure (out of scope; normalization
re-encoding is in scope per the Spec and the r0/r1 reviews). The r1 probe-function fallback for the auth-reader path
is withdrawn (C-5 disposed).

## 11. Budget, forecast, pins and scope

### 11.1 Budget — B5, checkpoint after attempt 3; forecast 2 attempts.

### 11.2 Count pins (provisional; re-pin at R0)

Migrations 12 → **13**; audit actions 35 → **36**; capabilities 12 → **13** (after D-1); `required-checks.tsv` 119 →
**121**; `access.denied` operations +5; `completion.schema.json` route enums +4 (two occurrences each); `events.json`
+1 event; `web/public` inventory unchanged (8); `components/ui` set unchanged; `SEMANTIC_COLOR_TOKENS` 16 unchanged,
new exported 15-name subset; Playwright projects unchanged; `appearance-init` hash unchanged; `stored_objects` schema
unchanged.

### 11.3 Scope allow-list — `management/v0/plans/G017-scope-allowlist.txt` (r1, **unchanged in r2** — same bytes, SHA-256 `dc6373012c18f0b6003c615b79ef095a27a86a10dee3f829545678523b4ad8a1`: 41 executable patterns, 52 lines with comments, versus 43 in r0; the D-5
extraction patterns and the ledger are removed; `tenancy/branding.go`, boundary/parity pins, identity test files and
the unconditional logrecord enums are added). Codex imports it as `management/v0/results/G017-scope-allowlist.txt`.

## 12. Open questions for Codex

- None new in r2. C-4 and C-5 were disposed by the r1 review and are applied as directed (§5.5, §4 item 6). The only
  open item is the owner's D-1 answer (§0.1, §10), which this Plan does not assume.

## 13. Revision history

| Rev | Date (UTC) | Base | Change |
|---|---|---|---|
| r0 | 2026-10-16 | `0a163c7f` | initial provisional Plan; archived by Codex as `…-plan-r0.md` (`5d9a80e6…4121`) with allow-list r0 (`fe3352dc…b406`) |
| r1 | 2026-10-16 | `b9b6e5c0` merged into `g017-plan` | one bounded correction round per the r0 review (`c88b8822…f7f8`): image safety, upload recovery/uncertain commit, FK/retention, auth-reader RLS, query layer, effective contrast/states, focused validation, manual preview/draft MR; status/pins/counts; C-1..C-3 resolved; D-1 recorded as pending. Allow-list r1 SHA-256: `dc6373012c18f0b6003c615b79ef095a27a86a10dee3f829545678523b4ad8a1` |
| r2 | 2026-10-16 | `61d6ff66` (r1) with `21469fc0` merged into `g017-plan-r2` | second bounded correction round per the r1 review (`bde78eb4…982a`): unknown publication outcome retained and reported (`publication_outcome_unknown`, in-flight/advanced-version cases, actual states); image normalization by stdlib re-encode with buffering-independent PNG/JPEG walks and corrected memory bounds (16-bit PNG 32 MiB); whole-button contrast transforms, r1 N1/N4 fill-boundary pairs withdrawn, focus outline against the enclosing surface, dialog Close icon pair added; C-4 (24 h cache) and C-5 (explicit helper-role policies) applied as disposed; BRD-11 public-reader wording corrected; emulator-backed manual preview; D-1 still pending; PG-008 source estimate only. r1 archived byte-exact as `…-plan-r1.md` (`5b72dbf2…4204`); allow-list unchanged (`dc637301…4ad8a1`) |
