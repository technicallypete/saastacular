# Contract conformance corpus

Fixtures for checking that independent validators agree on the intelligence wire contracts:

- the execution envelope, `intelligence/execution-envelope.schema.json`, sent from core to intelligence;
- the stream event, `intelligence/stream-event.schema.json`, sent from intelligence to core.

This directory holds test fixtures only. The schemas are unchanged and remain the single source of the contract. The corpus is kept outside `contracts/intelligence/`, so it never enters the intelligence runtime image.

## Provenance

- All data is synthetic, hand-specified for G003, and generated deterministically.
- Identifiers are RFC-style UUIDs or plain synthetic strings. No value is a credential.
- The field names `api_key`, `password`, `access_token` and `authorization` in EI18–EI21 exist to prove that the envelope rejects them. Their values are the placeholder `example`.
- Each `schema-negative/` fixture is the canonical schema with exactly one declared mutation (see below).

## Layout

```text
intelligence/expectations.json      declared outcome of every case, pinned schema digests, audit profile
intelligence/cases/envelope/*.json  EV01–EV14 (valid), EI01–EI50 (invalid): one envelope document each
intelligence/cases/event/*.json     XV01–XV15 (valid), XI01–XI34 (invalid): one stream event each
intelligence/cases/ndjson/*.ndjson  NV01–NV02 (valid), NI01–NI05 (invalid): stream files
intelligence/schema-negative/*.json SN02–SN07: mutated schemas for the audit (SN01 is declared absent)
```

- **File names.** Each file is named after its lowercase case ID, a hyphen, a slug matching `[a-z0-9-]+`, and its extension.
- **Exact file set.** The files under `intelligence/` are exactly `expectations.json` plus the files declared there: 126 fixtures (120 cases and SN02–SN07).

## Case classes

Every validator computes these classes in the same way, and `expectations.json` declares the expected one for each case.

| Class | Meaning |
|---|---|
| `valid` | The document parses as strict RFC 8259 JSON and validates against its schema. |
| `invalid-json` | Strict parsing fails: truncation, a trailing comma, a `NaN` or `Infinity` literal, or a non-JSON line. |
| `schema-invalid` | The document parses, but fails the schema. A top-level array is included here. |
| `ndjson-framing` | Stream files only. The file is empty, a final record has no LF, or there is a blank (zero-length) line. |

- **Format assertion.** `format` is asserted:
  - `date-time` is RFC 3339 with calendar and leap-second rules;
  - `uuid` is checked as a UUID.
- **Pattern semantics.** `pattern` uses ECMA-262 semantics: `$` does not match before a final newline.
- **Format-only cases.** EI39–EI45 and EI50 satisfy the `deadline` pattern and fail only on `format: date-time`. They prove that format assertion is in effect.

## Stream-check rules (NDJSON)

These are test-profile rules for checking producer output. They make no wire-contract change. The first failing rule decides the class.

1. **Framing**, checked before any line is parsed. The file has at least one record, every record ends with LF, and no line is blank.
2. **Each line, in order.** A line must first parse as strict JSON (otherwise `invalid-json`). It must then validate against the stream-event schema (otherwise `schema-invalid`).

## Schema digests and audit

- **Digest pins.** `expectations.json` pins the SHA-256 of both schema files. In normal mode (`corpus`, `ndjson`, `peer`), every validator checks the digests before anything else. A mismatch fails with `schema-digest-mismatch`, and no case is evaluated. No option disables this check.
- **Audit mode.** The audit classifies a schema file by the first stage that fails, in this order:
  1. `schema-missing`
  2. `schema-malformed-json`
  3. `schema-wrong-draft`
  4. `schema-meta-invalid`
  5. `schema-unsupported-keyword`
  6. `schema-unsupported-format`
  7. `schema-remote-ref`
  8. `schema-compile-failed`

  A schema that passes every stage is `ok`.

| ID | Base | Mutation | Class |
|---|---|---|---|
| SN01 | envelope | the file is absent | `schema-missing` |
| SN02 | envelope | the first 64 bytes of the canonical file | `schema-malformed-json` |
| SN03 | envelope | replace `/properties/authorization_ref/type` with `5` | `schema-meta-invalid` |
| SN04 | envelope | add `/dependentSchemas` | `schema-unsupported-keyword` |
| SN05 | envelope | add `/properties/authorization_ref/format` `"ipv4"` | `schema-unsupported-format` |
| SN06 | stream-event | replace `/oneOf/0/$ref` with a remote URL | `schema-remote-ref` |
| SN07 | envelope | replace `/$schema` with draft-07 | `schema-wrong-draft` |

- **Digest controls.** SD01 (the canonical bytes plus one LF) and SD02 (SN03's bytes in place of the canonical file) are derived in a scratch directory at test time. Each must fail in normal mode with `schema-digest-mismatch`, with zero cases evaluated.
- **Positive controls.** SA01 and SA02 require both canonical schemas to audit as `ok`.

## How to run

- **Go.** The Go validator (`contractcheck`) runs from the README `platform:contracts` block in the repository root.
- **Python.** The Python validator runs inside the intelligence check image. That image holds the schemas at `/srv/contracts/intelligence` and this corpus at `/srv/conformance/intelligence`. Its `pytest` run includes `tests/test_conformance.py`.

To run the Python commands directly:

```sh
docker build -q -f infra/docker/intelligence.Dockerfile --target check --build-context contracts=contracts -t intelligence-check intelligence
docker run --rm --network none intelligence-check python tests/conformance_exchange.py corpus --schemas /srv/contracts/intelligence --corpus /srv/conformance/intelligence --out /dev/null
docker run --rm --network none intelligence-check python tests/conformance_exchange.py schema-negative --schemas /srv/contracts/intelligence --corpus /srv/conformance/intelligence --out /dev/null
```
