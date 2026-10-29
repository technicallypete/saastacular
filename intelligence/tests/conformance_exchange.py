"""Contract conformance exchange for the intelligence contracts (test tooling).

A script, not a pytest module: it is run inside the intelligence check image as
``python tests/conformance_exchange.py <command> ...`` (working directory
``/srv/app``), and ``tests/test_conformance.py`` imports its helpers. It uses
the service's own validator (``streaming.schemas``) with strict RFC 8259
parsing. The schema digest pin, the corpus layout rules and the schema audit
stages live here, in test tooling, not in the runtime module.

Commands (every path is a directory or file inside the container):

``corpus --schemas DIR --corpus DIR --out FILE``
    Normal mode. Checks both schema digests against ``expectations.json``
    first (a mismatch evaluates zero cases), audits both schemas, checks the
    corpus layout, then classifies every case and compares it with its
    declared expectation. Writes the shared result format; prints every
    mismatching case ID; exits 0 only when every case matches.

``ndjson --schemas DIR --corpus DIR --file F --as event|envelope [--require-types a,b]``
    Normal mode. Classifies one NDJSON stream file; prints a JSON summary;
    exits 0 only when the file is valid and contains every required event type.

``schema-negative --schemas DIR --corpus DIR --out FILE``
    Audit mode for SN01-SN07 (declared classes), the positive audit controls
    SA01/SA02, the derivation check of each SN fixture, and the normal-mode
    digest controls SD01/SD02 in a scratch schema directory.

``peer --schemas DIR --corpus DIR --envelopes FILE --out FILE``
    Normal mode. Validates the Go-produced envelopes (an NDJSON file), posts
    each one through ``create_app`` with a stub gateway transport in three
    scenarios (success, gateway error, capability not configured), checks
    every response, and writes the actual raw response bytes to FILE.

Exit status: 0 success; 1 a check failed; 2 usage error.
"""

from __future__ import annotations

import argparse
import asyncio
import copy
import hashlib
import json
import logging
import os
import re
import sys
import tempfile
from collections import Counter
from collections.abc import Mapping, Sequence
from dataclasses import dataclass, field
from datetime import UTC, datetime
from importlib.metadata import version
from pathlib import Path
from typing import Any

import httpx
from jsonschema import Draft202012Validator, SchemaError

from streaming import schemas as contract_schemas
from streaming.schemas import build_validator, profile_violations, strict_loads

VALIDATOR_NAME = f"python-jsonschema {version('jsonschema')}"
SCHEMA_NAMES = ("execution-envelope", "stream-event")
EXPECTATIONS_FILE = "expectations.json"

VALID = "valid"
INVALID_JSON = "invalid-json"
SCHEMA_INVALID = "schema-invalid"
NDJSON_FRAMING = "ndjson-framing"
CASE_CLASSES = (VALID, INVALID_JSON, SCHEMA_INVALID, NDJSON_FRAMING)

AUDIT_OK = "ok"
AUDIT_STAGES = (
    "schema-missing",
    "schema-malformed-json",
    "schema-wrong-draft",
    "schema-meta-invalid",
    "schema-unsupported-keyword",
    "schema-unsupported-format",
    "schema-remote-ref",
    "schema-compile-failed",
)
DIGEST_MISMATCH = "schema-digest-mismatch"

# Fixture directory per (kind, schema) and the file-name rule.
CASE_DIRECTORIES = {
    ("document", "execution-envelope"): ("cases/envelope", ".json"),
    ("document", "stream-event"): ("cases/event", ".json"),
    ("ndjson", "stream-event"): ("cases/ndjson", ".ndjson"),
}
SN_DIRECTORY = ("schema-negative", ".json")
SLUG = re.compile(r"[a-z0-9-]+")


class ConformanceError(Exception):
    """A normal-mode precondition failed; ``cls`` names it."""

    def __init__(self, cls: str, detail: str) -> None:
        self.cls = cls
        super().__init__(f"{cls}: {detail}")


class SchemaDigestMismatch(ConformanceError):
    """Schema bytes differ from the pinned digests; no case was evaluated."""

    def __init__(self, actual: Mapping[str, str | None], mismatched: Sequence[str]) -> None:
        self.actual = dict(actual)
        self.mismatched = list(mismatched)
        super().__init__(DIGEST_MISMATCH, ", ".join(mismatched))


# --------------------------------------------------------------------------
# Hashing and expectations
# --------------------------------------------------------------------------


def sha256_hex(data: bytes) -> str:
    return hashlib.sha256(data).hexdigest()


def corpus_files(corpus: Path) -> list[str]:
    """Every regular file under ``corpus``, relative, sorted bytewise."""
    found: list[str] = []
    for root, dirs, names in os.walk(corpus):
        dirs.sort()
        for name in names:
            path = Path(root) / name
            if path.is_file() and not path.is_symlink():
                found.append(path.relative_to(corpus).as_posix())
    return sorted(found, key=lambda p: p.encode("utf-8"))


def corpus_sha256(corpus: Path) -> str:
    """SHA-256 of ``"<sha256>  <path>\\n"`` lines over the sorted corpus files."""
    lines = "".join(
        f"{sha256_hex((corpus / rel).read_bytes())}  {rel}\n" for rel in corpus_files(corpus)
    )
    return sha256_hex(lines.encode("utf-8"))


def load_expectations(corpus: Path) -> dict[str, Any]:
    try:
        expectations = strict_loads((corpus / EXPECTATIONS_FILE).read_bytes())
    except (OSError, ValueError) as exc:
        raise ConformanceError("expectations-invalid", type(exc).__name__) from None
    if not isinstance(expectations, dict) or expectations.get("version") != 1:
        raise ConformanceError("expectations-invalid", "version must be 1")
    pinned = expectations.get("schemas")
    if not isinstance(pinned, dict) or set(pinned) != set(SCHEMA_NAMES):
        raise ConformanceError("expectations-invalid", "schemas must pin both contracts")
    return expectations


def pinned_schema_bytes(schemas_dir: Path, expectations: Mapping[str, Any]) -> dict[str, bytes]:
    """Read both schemas and check their digests before anything else."""
    data: dict[str, bytes] = {}
    actual: dict[str, str | None] = {}
    mismatched: list[str] = []
    for name in SCHEMA_NAMES:
        spec = expectations["schemas"][name]
        try:
            raw = (schemas_dir / spec["file"]).read_bytes()
        except OSError:
            actual[name] = None
            mismatched.append(name)
            continue
        actual[name] = sha256_hex(raw)
        if actual[name] != spec["sha256"]:
            mismatched.append(name)
        data[name] = raw
    if mismatched:
        raise SchemaDigestMismatch(actual, mismatched)
    return data


# --------------------------------------------------------------------------
# Schema audit (stages 1-8; the first failing stage decides)
# --------------------------------------------------------------------------


def audit_bytes(raw: bytes) -> str:
    """Audit stages 2-8 for schema bytes; returns a class or ``ok``."""
    try:
        schema = strict_loads(raw)
    except ValueError:
        return "schema-malformed-json"
    if not isinstance(schema, dict):
        return "schema-malformed-json"
    if schema.get("$schema") != contract_schemas.DRAFT_2020_12:
        return "schema-wrong-draft"
    try:
        Draft202012Validator.check_schema(schema)
    except SchemaError:
        return "schema-meta-invalid"
    violations = profile_violations(schema)
    for kind, cls in (
        ("keyword", "schema-unsupported-keyword"),
        ("format", "schema-unsupported-format"),
        ("ref", "schema-remote-ref"),
    ):
        if any(v.kind == kind for v in violations):
            return cls
    try:
        validator = build_validator(schema)
        validator.is_valid({})  # resolves references with assertions on
    except Exception:
        return "schema-compile-failed"
    return AUDIT_OK


def audit_file(path: Path) -> str:
    """Audit stages 1-8 for one schema file (audit mode: no digest step)."""
    if not path.is_file():
        return "schema-missing"
    return audit_bytes(path.read_bytes())


def load_validators(schemas_dir: Path, expectations: Mapping[str, Any]) -> dict[str, Any]:
    """Normal mode: digests first, then the audit, then compiled validators."""
    raw = pinned_schema_bytes(schemas_dir, expectations)
    validators: dict[str, Any] = {}
    for name in SCHEMA_NAMES:
        cls = audit_bytes(raw[name])
        if cls != AUDIT_OK:
            raise ConformanceError(cls, name)
        validators[name] = build_validator(strict_loads(raw[name]))
    return validators


# --------------------------------------------------------------------------
# Case classification
# --------------------------------------------------------------------------


def classify_document(raw: bytes, validator: Any) -> tuple[str, Any]:
    try:
        document = strict_loads(raw)
    except ValueError:
        return INVALID_JSON, None
    if not validator.is_valid(document):
        return SCHEMA_INVALID, document
    return VALID, document


@dataclass
class StreamResult:
    cls: str
    line: int | None = None  # 1-based line of the first failure
    documents: list[Any] = field(default_factory=list)
    lines: list[bytes] = field(default_factory=list)

    @property
    def event_types(self) -> list[str]:
        return sorted({d["type"] for d in self.documents if isinstance(d, dict) and "type" in d})


def classify_ndjson(raw: bytes, validator: Any) -> StreamResult:
    """Framing first (empty file, final record without LF, blank line), then
    each line in order: strict parse, then schema."""
    if not raw or not raw.endswith(b"\n"):
        return StreamResult(NDJSON_FRAMING, None if not raw else raw.count(b"\n") + 1)
    lines = raw[:-1].split(b"\n")
    for number, line in enumerate(lines, start=1):
        if line == b"":
            return StreamResult(NDJSON_FRAMING, number)
    result = StreamResult(VALID, lines=lines)
    for number, line in enumerate(lines, start=1):
        cls, document = classify_document(line, validator)
        if cls != VALID:
            return StreamResult(cls, number)
        result.documents.append(document)
    return result


def classify_case(case: Mapping[str, Any], corpus: Path, validators: Mapping[str, Any]) -> str:
    raw = (corpus / case["file"]).read_bytes()
    validator = validators[case["schema"]]
    if case["kind"] == "ndjson":
        return classify_ndjson(raw, validator).cls
    return classify_document(raw, validator)[0]


def case_event_types(
    case: Mapping[str, Any], corpus: Path, validators: Mapping[str, Any]
) -> list[str]:
    """Event types of a valid event case (documents and every NDJSON line)."""
    if case["schema"] != "stream-event":
        return []
    raw = (corpus / case["file"]).read_bytes()
    if case["kind"] == "ndjson":
        return classify_ndjson(raw, validators["stream-event"]).event_types
    cls, document = classify_document(raw, validators["stream-event"])
    return [document["type"]] if cls == VALID else []


# --------------------------------------------------------------------------
# Corpus layout
# --------------------------------------------------------------------------


def _expected_name(identifier: str, rel: str, directory: str, extension: str) -> bool:
    folder, _, name = rel.rpartition("/")
    prefix = identifier.lower() + "-"
    return (
        folder == directory
        and name.startswith(prefix)
        and name.endswith(extension)
        and SLUG.fullmatch(name[len(prefix) : -len(extension)]) is not None
    )


def layout_problems(corpus: Path, expectations: Mapping[str, Any]) -> list[str]:
    """The files must be exactly the declared fixtures plus expectations.json."""
    problems: list[str] = []
    declared = {EXPECTATIONS_FILE}
    ids: list[str] = []
    for case in expectations.get("cases", []):
        ids.append(case["id"])
        directory, extension = CASE_DIRECTORIES.get((case["kind"], case["schema"]), ("?", "?"))
        if not _expected_name(case["id"], case["file"], directory, extension):
            problems.append(f"{case['id']}: file name or directory")
        if case["class"] not in CASE_CLASSES or case["valid"] is not (case["class"] == VALID):
            problems.append(f"{case['id']}: declared class")
        declared.add(case["file"])
    for entry in expectations.get("schema_negative", []):
        ids.append(entry["id"])
        if not _expected_name(entry["id"], entry["file"], *SN_DIRECTORY):
            problems.append(f"{entry['id']}: file name or directory")
        if entry["mutation"]["op"] == "absent":
            if (corpus / entry["file"]).exists():
                problems.append(f"{entry['id']}: must not exist")
        else:
            declared.add(entry["file"])
    duplicates = sorted(i for i, n in Counter(ids).items() if n > 1)
    if duplicates:
        problems.append("duplicate IDs: " + ", ".join(duplicates))
    present = set(corpus_files(corpus))
    for rel in sorted(present - declared):
        problems.append(f"undeclared file: {rel}")
    for rel in sorted(declared - present):
        problems.append(f"missing file: {rel}")
    return problems


# --------------------------------------------------------------------------
# Normal mode: corpus
# --------------------------------------------------------------------------


def run_corpus(schemas_dir: Path, corpus: Path) -> dict[str, Any]:
    """Evaluate every case. Raises before any case on a digest or audit failure."""
    expectations = load_expectations(corpus)
    validators = load_validators(schemas_dir, expectations)
    problems = layout_problems(corpus, expectations)
    if problems:
        raise ConformanceError("corpus-layout", "; ".join(problems))
    results: list[dict[str, Any]] = []
    event_types: set[str] = set()
    mismatches: list[str] = []
    for case in expectations["cases"]:
        cls = classify_case(case, corpus, validators)
        valid = cls == VALID
        results.append({"id": case["id"], "valid": valid, "class": cls})
        if valid:
            event_types.update(case_event_types(case, corpus, validators))
        if valid is not case["valid"] or cls != case["class"]:
            mismatches.append(case["id"])
    by_class = Counter(r["class"] for r in results)
    return {
        "validator": VALIDATOR_NAME,
        "schema_sha256": {name: expectations["schemas"][name]["sha256"] for name in SCHEMA_NAMES},
        "corpus_sha256": corpus_sha256(corpus),
        "results": results,
        "counts": {
            "total": len(results),
            "valid": sum(1 for r in results if r["valid"]),
            "invalid": sum(1 for r in results if not r["valid"]),
            "by_class": dict(sorted(by_class.items())),
            "event_types": sorted(event_types),
        },
        "skipped": 0,
        "mismatches": mismatches,
    }


def _failure_result(error: ConformanceError) -> dict[str, Any]:
    result: dict[str, Any] = {
        "validator": VALIDATOR_NAME,
        "class": error.cls,
        "detail": str(error),
        "evaluated": 0,
        "results": [],
        "skipped": 0,
    }
    if isinstance(error, SchemaDigestMismatch):
        result["schema_sha256"] = error.actual
    return result


def _write_json(path: str, document: Any) -> None:
    with open(path, "w", encoding="utf-8") as fh:
        fh.write(json.dumps(document, indent=2) + "\n")


def command_corpus(args: argparse.Namespace) -> int:
    try:
        result = run_corpus(Path(args.schemas), Path(args.corpus))
    except ConformanceError as error:
        _write_json(args.out, _failure_result(error))
        print(f"corpus: {error}; 0 cases evaluated", file=sys.stderr)
        return 1
    _write_json(args.out, result)
    if result["mismatches"]:
        expected = {c["id"]: c for c in load_expectations(Path(args.corpus))["cases"]}
        actual = {r["id"]: r for r in result["results"]}
        for identifier in result["mismatches"]:
            print(
                f"corpus: mismatch {identifier}: expected {expected[identifier]['class']}, "
                f"got {actual[identifier]['class']}",
                file=sys.stderr,
            )
        return 1
    print(f"corpus: {result['counts']['total']} cases match their expectations", file=sys.stderr)
    return 0


# --------------------------------------------------------------------------
# Normal mode: one NDJSON stream
# --------------------------------------------------------------------------


def command_ndjson(args: argparse.Namespace) -> int:
    schema = "stream-event" if args.as_ == "event" else "execution-envelope"
    required = [t for t in (args.require_types or "").split(",") if t]
    if required and schema != "stream-event":
        print("ndjson: --require-types applies to --as event only", file=sys.stderr)
        return 2
    try:
        expectations = load_expectations(Path(args.corpus))
        validators = load_validators(Path(args.schemas), expectations)
    except ConformanceError as error:
        print(json.dumps({"validator": VALIDATOR_NAME, "class": error.cls, "evaluated": 0}))
        print(f"ndjson: {error}", file=sys.stderr)
        return 1
    try:
        raw = Path(args.file).read_bytes()
    except OSError as exc:
        print(f"ndjson: cannot read the file ({type(exc).__name__})", file=sys.stderr)
        return 1
    result = classify_ndjson(raw, validators[schema])
    types = result.event_types if schema == "stream-event" and result.cls == VALID else []
    missing = sorted(set(required) - set(types)) if result.cls == VALID else sorted(required)
    ok = result.cls == VALID and not missing
    print(
        json.dumps(
            {
                "validator": VALIDATOR_NAME,
                "as": args.as_,
                "valid": result.cls == VALID,
                "class": result.cls,
                "line": result.line,
                "records": len(result.documents),
                "event_types": types,
                "missing_types": missing,
                "ok": ok,
            }
        )
    )
    if not ok:
        reason = result.cls if result.cls != VALID else "missing types " + ",".join(missing)
        where = f" at line {result.line}" if result.line else ""
        print(f"ndjson: rejected: {reason}{where}", file=sys.stderr)
    return 0 if ok else 1


# --------------------------------------------------------------------------
# Audit mode: schema-negative, positive controls, digest controls
# --------------------------------------------------------------------------


def _pointer_tokens(pointer: str) -> list[str]:
    return [t.replace("~1", "/").replace("~0", "~") for t in pointer.split("/")[1:]]


def apply_mutation(document: Any, mutation: Mapping[str, Any]) -> Any:
    """JSON-Patch-style ``add``/``replace`` of one value at a JSON Pointer."""
    result = copy.deepcopy(document)
    tokens = _pointer_tokens(mutation["path"])
    parent = result
    for token in tokens[:-1]:
        parent = parent[int(token)] if isinstance(parent, list) else parent[token]
    last = tokens[-1]
    if mutation["op"] == "replace":
        if isinstance(parent, list):
            parent[int(last)] = mutation["value"]
        else:
            if last not in parent:
                raise KeyError(mutation["path"])
            parent[last] = mutation["value"]
    elif mutation["op"] == "add":
        if isinstance(parent, list):
            parent.insert(len(parent) if last == "-" else int(last), mutation["value"])
        else:
            parent[last] = mutation["value"]
    else:
        raise ValueError(f"unsupported mutation {mutation['op']}")
    return result


def derive_check(entry: Mapping[str, Any], corpus: Path, canonical: Mapping[str, bytes]) -> bool:
    """The SN fixture is exactly the declared single mutation of its base."""
    path = corpus / entry["file"]
    mutation = entry["mutation"]
    base = canonical[entry["base"]]
    if mutation["op"] == "absent":
        return not path.exists()
    if not path.is_file():
        return False
    raw = path.read_bytes()
    if mutation["op"] == "truncate-bytes":
        return raw == base[: mutation["length"]]
    try:
        return strict_loads(raw) == apply_mutation(strict_loads(base), mutation)
    except (ValueError, KeyError, IndexError, TypeError):
        return False


def digest_control(
    control: Mapping[str, Any],
    schemas_dir: Path,
    corpus: Path,
    expectations: Mapping[str, Any],
) -> dict[str, Any]:
    """Run normal mode against a scratch schema directory with altered bytes."""
    derivation = control["derivation"]
    base_file = expectations["schemas"][control["base"]]["file"]
    if derivation == "canonical-plus-lf":
        altered = (schemas_dir / base_file).read_bytes() + b"\n"
    elif derivation.startswith("bytes-of:"):
        source = derivation.removeprefix("bytes-of:")
        entry = next(e for e in expectations["schema_negative"] if e["id"] == source)
        altered = (corpus / entry["file"]).read_bytes()
    else:
        raise ValueError(f"unknown derivation {derivation}")
    evaluated = 0
    with tempfile.TemporaryDirectory(prefix="digest-control-") as scratch:
        scratch_dir = Path(scratch)
        for name in SCHEMA_NAMES:
            spec_file = expectations["schemas"][name]["file"]
            data = altered if name == control["base"] else (schemas_dir / spec_file).read_bytes()
            (scratch_dir / spec_file).write_bytes(data)
        try:
            evaluated = run_corpus(scratch_dir, corpus)["counts"]["total"]
            cls = "ok"
        except ConformanceError as error:
            cls = error.cls
    return {
        "id": control["id"],
        "class": cls,
        "expected_class": control["class"],
        "match": cls == control["class"] and evaluated == 0,
        "evaluated": evaluated,
    }


def run_schema_negative(schemas_dir: Path, corpus: Path) -> dict[str, Any]:
    expectations = load_expectations(corpus)
    canonical = {
        name: (schemas_dir / expectations["schemas"][name]["file"]).read_bytes()
        for name in SCHEMA_NAMES
    }
    results = []
    derive_ok = True
    for entry in expectations["schema_negative"]:
        cls = audit_file(corpus / entry["file"])
        results.append(
            {
                "id": entry["id"],
                "class": cls,
                "expected_class": entry["class"],
                "match": cls == entry["class"],
            }
        )
        derive_ok = derive_check(entry, corpus, canonical) and derive_ok
    positives = []
    for control in expectations["audit_positive"]:
        cls = audit_file(schemas_dir / expectations["schemas"][control["schema"]]["file"])
        positives.append(
            {
                "id": control["id"],
                "schema": control["schema"],
                "class": cls,
                "expected_class": control["class"],
                "match": cls == control["class"],
            }
        )
    digests = [
        digest_control(control, schemas_dir, corpus, expectations)
        for control in expectations["digest_controls"]
    ]
    return {
        "validator": VALIDATOR_NAME,
        "results": results,
        "derive": "pass" if derive_ok else "fail",
        "audit_positive": positives,
        "digest_controls": digests,
    }


def schema_negative_ok(result: Mapping[str, Any]) -> bool:
    return (
        result["derive"] == "pass"
        and all(r["match"] for r in result["results"])
        and all(r["match"] for r in result["audit_positive"])
        and all(r["match"] for r in result["digest_controls"])
    )


def command_schema_negative(args: argparse.Namespace) -> int:
    try:
        result = run_schema_negative(Path(args.schemas), Path(args.corpus))
    except (ConformanceError, OSError, KeyError, StopIteration) as error:
        print(f"schema-negative: {type(error).__name__}: {error}", file=sys.stderr)
        return 1
    _write_json(args.out, result)
    for group in ("results", "audit_positive", "digest_controls"):
        for row in result[group]:
            if not row["match"]:
                print(
                    f"schema-negative: {row['id']}: expected {row['expected_class']}, "
                    f"got {row['class']}",
                    file=sys.stderr,
                )
    if result["derive"] != "pass":
        print("schema-negative: derivation check failed", file=sys.stderr)
    return 0 if schema_negative_ok(result) else 1


# --------------------------------------------------------------------------
# Normal mode: peer exchange
# --------------------------------------------------------------------------

PEER_TOKEN = "peer-exchange-service-value-0001"
PEER_TEXT_CAPABILITIES = ("CHAT", "FAST", "REASONING", "EXTRACTION")
PEER_SCENARIOS = ("success", "gateway-error", "capability-not-configured")
PEER_EXPECTED = {
    "success": [("text_delta", None), ("usage", None), ("done", None)],
    "gateway-error": [("error", "gateway_error"), ("done", None)],
    "capability-not-configured": [("error", "capability_not_configured"), ("done", None)],
}
PEER_CANNED: dict[str, Any] = {
    "id": "gen-synthetic-peer",
    "model": "anthropic/test-model",
    "choices": [{"index": 0, "message": {"role": "assistant", "content": "synthetic answer"}}],
    "usage": {"prompt_tokens": 7, "completion_tokens": 3, "cost": 0.0001},
}


class PeerStubTransport:
    """Deterministic gateway transport (as in conftest): canned body or an error."""

    def __init__(self, error: Exception | None = None) -> None:
        self.error = error
        self.calls = 0

    async def post_json(
        self, path: str, payload: Mapping[str, Any], timeout_s: float
    ) -> Mapping[str, Any]:
        self.calls += 1
        if self.error is not None:
            raise self.error
        return copy.deepcopy(PEER_CANNED)


def _peer_env(scenario: str) -> dict[str, str]:
    env = {"APP_ENV": "test", "INTELLIGENCE_SERVICE_TOKEN": PEER_TOKEN}
    if scenario != "capability-not-configured":
        env.update({f"AI_MODEL_{cap}": "anthropic:test-model" for cap in PEER_TEXT_CAPABILITIES})
    return env


def _peer_post(raw_envelope: bytes, scenario: str) -> tuple[int, str, bytes]:
    """POST the envelope's exact bytes to a fresh app; return the raw response."""
    from main import create_app, get_gateway
    from providers import GatewayError, OpenRouterGateway

    error = None
    if scenario == "gateway-error":
        error = GatewayError("gateway_error", "synthetic gateway failure", retryable=True)
    gateway = OpenRouterGateway(PeerStubTransport(error))
    app = create_app(_peer_env(scenario))
    app.dependency_overrides[get_gateway] = lambda: gateway

    async def post() -> httpx.Response:
        transport = httpx.ASGITransport(app=app)
        async with httpx.AsyncClient(transport=transport, base_url="http://intelligence") as client:
            return await client.post(
                "/internal/v1/execute",
                content=raw_envelope,
                headers={
                    "Authorization": f"Bearer {PEER_TOKEN}",
                    "Content-Type": "application/json",
                },
            )

    response = asyncio.run(post())
    return response.status_code, response.headers.get("content-type", ""), response.content


def _eligible(envelope: Mapping[str, Any]) -> bool:
    """Whether the scenario's exact event sequence is determined by the scenario."""
    if envelope.get("capability", "").upper() not in PEER_TEXT_CAPABILITIES:
        return False
    try:
        deadline = datetime.fromisoformat(envelope["deadline"].replace("Z", "+00:00"))
    except ValueError:
        return False
    return deadline > datetime.now(UTC)


def run_peer(schemas_dir: Path, corpus: Path, envelopes: Path, out: str) -> dict[str, Any]:
    expectations = load_expectations(corpus)
    validators = load_validators(schemas_dir, expectations)
    # The service validates with its own schema copy: it must be the pinned bytes too.
    pinned_schema_bytes(contract_schemas.contracts_dir(), expectations)
    try:
        raw = envelopes.read_bytes()
    except OSError as exc:
        raise ConformanceError("peer-input", type(exc).__name__) from None
    inbound = classify_ndjson(raw, validators["execution-envelope"])
    if inbound.cls != VALID:
        raise ConformanceError("peer-envelopes-rejected", f"{inbound.cls} at line {inbound.line}")

    output = bytearray()
    summary: list[dict[str, Any]] = []
    problems: list[str] = []
    previous = logging.root.manager.disable
    logging.disable(logging.CRITICAL)  # keep the tool's stdout to its own summary
    try:
        for number, (line, envelope) in enumerate(
            zip(inbound.lines, inbound.documents, strict=True), start=1
        ):
            eligible = _eligible(envelope)
            for scenario in PEER_SCENARIOS:
                status, media_type, body = _peer_post(line, scenario)
                where = f"envelope {number} {scenario}"
                if status != 200 or not media_type.startswith("application/x-ndjson"):
                    problems.append(f"{where}: status {status}")
                    continue
                stream = classify_ndjson(body, validators["stream-event"])
                if stream.cls != VALID:
                    problems.append(f"{where}: response {stream.cls} at line {stream.line}")
                    continue
                observed = [(d["type"], d.get("code")) for d in stream.documents]
                if observed[-1][0] != "done":
                    problems.append(f"{where}: does not end with done")
                if any(d["request_id"] != envelope["request_id"] for d in stream.documents):
                    problems.append(f"{where}: request_id differs from the envelope")
                if eligible and observed != PEER_EXPECTED[scenario]:
                    problems.append(f"{where}: unexpected event sequence")
                output.extend(body)
                summary.append(
                    {
                        "envelope": number,
                        "scenario": scenario,
                        "event_types": [t for t, _ in observed],
                    }
                )
    finally:
        logging.disable(previous)
    if problems:
        raise ConformanceError("peer-response-invalid", "; ".join(problems))
    produced = classify_ndjson(bytes(output), validators["stream-event"])
    if produced.cls != VALID:
        raise ConformanceError("peer-output-invalid", produced.cls)
    with open(out, "wb") as fh:
        fh.write(output)
    return {
        "validator": VALIDATOR_NAME,
        "envelopes": len(inbound.documents),
        "responses": summary,
        "event_types": produced.event_types,
        "bytes": len(output),
    }


def command_peer(args: argparse.Namespace) -> int:
    try:
        result = run_peer(Path(args.schemas), Path(args.corpus), Path(args.envelopes), args.out)
    except ConformanceError as error:
        print(f"peer: {error}", file=sys.stderr)
        return 1
    print(json.dumps(result))
    return 0


# --------------------------------------------------------------------------
# Entry point
# --------------------------------------------------------------------------


def parser() -> argparse.ArgumentParser:
    root = argparse.ArgumentParser(
        prog="conformance_exchange.py", description=__doc__.split("\n")[0]
    )
    commands = root.add_subparsers(dest="command", required=True)

    def common(sub: argparse.ArgumentParser) -> None:
        sub.add_argument("--schemas", required=True, help="directory with the two schema files")
        sub.add_argument("--corpus", required=True, help="corpus root (expectations.json)")

    corpus = commands.add_parser("corpus", help="evaluate every corpus case (normal mode)")
    common(corpus)
    corpus.add_argument("--out", required=True)
    corpus.set_defaults(handler=command_corpus)

    ndjson = commands.add_parser("ndjson", help="check one NDJSON stream file (normal mode)")
    common(ndjson)
    ndjson.add_argument("--file", required=True)
    ndjson.add_argument("--as", dest="as_", required=True, choices=("event", "envelope"))
    ndjson.add_argument("--require-types", default="")
    ndjson.set_defaults(handler=command_ndjson)

    negative = commands.add_parser("schema-negative", help="audit SN01-SN07, SA and SD controls")
    common(negative)
    negative.add_argument("--out", required=True)
    negative.set_defaults(handler=command_schema_negative)

    peer = commands.add_parser("peer", help="execute Go envelopes; write raw Python events")
    common(peer)
    peer.add_argument("--envelopes", required=True)
    peer.add_argument("--out", required=True)
    peer.set_defaults(handler=command_peer)
    return root


def main(argv: Sequence[str] | None = None) -> int:
    args = parser().parse_args(argv)
    return args.handler(args)


if __name__ == "__main__":
    sys.exit(main())
