# Baseline fixtures (historical copies; do not edit)

`staging.sh` and `lib.sh` here are byte-exact historical copies of `infra/staging/staging.sh` and
`infra/staging/lib.sh` from the revision that introduced the staging lifecycle tooling, before the CI lease and
fence interfaces were added. They are test data, not code that runs in a deployment:

- `infra/staging/check.sh` (row STAGING-GREP) compares the current `forbid()` in `infra/staging/lib.sh` with the one
  in `lib.sh` here.
- `infra/staging/tests/env-lease.sh` (row ENV-LEASE) runs both copies as the legacy baseline of its parity cases.

Both checks pin each file's SHA-256 and fail closed when a copy is missing, is not a regular file or differs:

| File | SHA-256 |
|---|---|
| `staging.sh` | `7c319382c48e8b3f089bee07ec33746c364567bac61131cf9edb775f47393ee7` |
| `lib.sh` | `9d11fe404de15c362137b7457caa779debf53196a40bbd5b968082f6983cb06b` |

STAGING-GREP and the shell lint lists skip `infra/staging/tests/fixtures/`, so these copies are not linted or
scanned as current tooling. Replacing them changes what the checks treat as the baseline; update both pins together
and only on purpose.
