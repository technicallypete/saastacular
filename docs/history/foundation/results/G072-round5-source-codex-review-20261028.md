> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G072 source review round 5 — 2026-10-28 UTC

**Outcome: ACCEPT for the submitted source at `a17d1537`** (executable tree
`d9909865`), review range `318f0b40..a17d1537`. This accepts the source against
the approved Plan and resolves the round-five P2. It does not claim that pending
hosted checks or Duo review have passed, and is not final Goal ACCEPT.

## Finding and correction

On `318f0b40`, setup checked the globally named metadata bucket once and relied
on cached ownership after API, federation and role operations. If the bucket
were deleted and its global name reused during that interval, later settings or
IAM writes could affect another project's bucket.

At `a17d1537`, `meta_guard` runs immediately before each non-dry-run mutation
that names the metadata bucket. It rechecks the bucket name and pinned project
number before settings updates, each bucket IAM binding (including retries),
and recursive deletion. A changed, absent or unreadable owner stops the action.
If a bucket absent at preflight cannot be created because its name was taken,
setup stops without adopting it; a successful create is checked before further
writes.

I inspected `infra/terraform/bootstrap/ci-federation.sh:239–257, 511–541,
634–687` and the corresponding fixture cases. Tests cover owner changes before
the update and grants, a post-create ownership mismatch, a matching owner,
and a name taken at creation. The `late-owner-checks-removed` control kills the
race regressions. An independent read-only review reached the same ACCEPT.
Neither reviewer ran tests. Claude's Result reports CI-FEDERATION 132/0,
S2 7/7 and S1 PASS on executable tree `d9909865`; these are attributed reports,
not runs by Codex.

## Remaining integration gates

The coordinator reports hosted pipeline for head `a17d1537` is running. Verify it
finishes successfully, and obtain Duo's re-review of this executable delta with
all discussions resolved. Then follow the normal MR integration path, including
input reconciliation and affected post-merge checks. G057's staging identity
hold and all separately gated live GCP/GitLab/IAM/DNS actions remain in force.
No test or cloud action was performed for this review.

## Source snapshot

The coordinator Result imported byte-for-byte from `g072-impl` at `a17d153763c1d4d71dabd800ae74a2ef19e5379e`;
SHA-256 `59a69e2f883d77f2d92495619c4b469f621a6614a6b13137895b90b8adbea3e9`. Its append-only history records all earlier failures and
corrections. See [Result](G072-manual-staging-ci-result.md).
