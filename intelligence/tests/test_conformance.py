"""I-CONFORMANCE: the shared contract corpus, schema audit and digest controls.

Every case in ``contracts/conformance/intelligence/expectations.json`` is
classified by the service's own validator (strict JSON parsing) and compared
with its independently declared expectation. The schema-negative fixtures,
the positive audit controls and the digest controls run here as well. No
case is ever skipped: a missing corpus is a failure, not a skip.
"""

from __future__ import annotations

import json
import shutil
from pathlib import Path
from typing import Any

import pytest

import conformance_exchange as cx
from conftest import make_envelope
from streaming.schemas import (
    ENVELOPE_SCHEMA,
    IMAGE_CONTRACTS_DIR,
    PROFILE_FORMATS,
    PROFILE_KEYWORDS,
    STREAM_EVENT_SCHEMA,
    contracts_dir,
    format_checker,
)

IMAGE_CORPUS_DIR = Path("/srv/conformance/intelligence")
# tests/test_conformance.py -> parents[2] is the repository root outside the image.
PROJECT_CORPUS_DIR = Path(__file__).resolve().parents[2] / "contracts/conformance/intelligence"
CORPUS = IMAGE_CORPUS_DIR if IMAGE_CORPUS_DIR.is_dir() else PROJECT_CORPUS_DIR
SCHEMAS = contracts_dir()

# Loaded at collection: a missing or unreadable corpus fails collection.
EXPECTATIONS: dict[str, Any] = cx.load_expectations(CORPUS)
CASES = EXPECTATIONS["cases"]
SCHEMA_NEGATIVE = EXPECTATIONS["schema_negative"]
DIGEST_CONTROLS = EXPECTATIONS["digest_controls"]
AUDIT_POSITIVE = EXPECTATIONS["audit_positive"]
ALL_EVENT_TYPES = [
    "citation",
    "done",
    "error",
    "structured_output",
    "text_delta",
    "tool_call",
    "tool_result",
    "usage",
]


@pytest.fixture(scope="module")
def validators() -> dict[str, Any]:
    # Normal mode: the digests are checked before any validator exists.
    return cx.load_validators(SCHEMAS, EXPECTATIONS)


def test_i_conformance_runs_from_the_image_locations() -> None:
    assert SCHEMAS == IMAGE_CONTRACTS_DIR
    assert CORPUS == IMAGE_CORPUS_DIR


def test_i_conformance_catalogue_is_complete() -> None:
    ids = [case["id"] for case in CASES]
    expected = (
        [f"EV{n:02d}" for n in range(1, 15)]
        + [f"EI{n:02d}" for n in range(1, 51)]
        + [f"XV{n:02d}" for n in range(1, 16)]
        + [f"XI{n:02d}" for n in range(1, 35)]
        + ["NV01", "NV02"]
        + [f"NI{n:02d}" for n in range(1, 6)]
    )
    assert ids == expected
    assert [e["id"] for e in SCHEMA_NEGATIVE] == [f"SN{n:02d}" for n in range(1, 8)]
    assert [c["id"] for c in DIGEST_CONTROLS] == ["SD01", "SD02"]
    assert [c["id"] for c in AUDIT_POSITIVE] == ["SA01", "SA02"]
    assert set(EXPECTATIONS["profile"]["keywords"]) == PROFILE_KEYWORDS
    assert set(EXPECTATIONS["profile"]["formats"]) == PROFILE_FORMATS


def test_i_conformance_schema_digests_are_pinned() -> None:
    pinned = {name: EXPECTATIONS["schemas"][name]["sha256"] for name in cx.SCHEMA_NAMES}
    assert pinned == {
        "execution-envelope": "bf3cec365e291da2cf9b201b9d0194632818721671a5844589954b978accba4f",
        "stream-event": "e1025031134253ac8d19a514f6d7a4d9563281b5006ebe9be40a84e4d63fd102",
    }
    raw = cx.pinned_schema_bytes(SCHEMAS, EXPECTATIONS)
    assert {name: cx.sha256_hex(data) for name, data in raw.items()} == pinned
    assert EXPECTATIONS["schemas"]["execution-envelope"]["file"] == ENVELOPE_SCHEMA
    assert EXPECTATIONS["schemas"]["stream-event"]["file"] == STREAM_EVENT_SCHEMA


def test_i_conformance_corpus_layout_is_exactly_the_declared_files() -> None:
    assert cx.layout_problems(CORPUS, EXPECTATIONS) == []
    files = cx.corpus_files(CORPUS)
    assert len(files) == 127  # 126 fixtures and expectations.json
    assert not (CORPUS / "schema-negative/sn01-missing.json").exists()


@pytest.mark.parametrize("case", CASES, ids=[c["id"] for c in CASES])
def test_i_conformance_case_matches_expectation(case: dict[str, Any], validators) -> None:
    cls = cx.classify_case(case, CORPUS, validators)
    assert (cls == cx.VALID, cls) == (case["valid"], case["class"])


def test_i_conformance_full_run_skips_nothing_and_covers_every_event_type() -> None:
    result = cx.run_corpus(SCHEMAS, CORPUS)
    assert result["mismatches"] == []
    assert result["skipped"] == 0
    counts = result["counts"]
    assert counts["total"] == len(CASES) == 120
    assert (counts["valid"], counts["invalid"]) == (31, 89)
    assert counts["event_types"] == ALL_EVENT_TYPES
    assert result["corpus_sha256"] == cx.corpus_sha256(CORPUS)


def test_i_conformance_format_only_cases_pass_the_pattern(validators) -> None:
    """EI39-EI45 and EI50 fail only on ``format: date-time``."""
    for identifier in ("EI39", "EI40", "EI41", "EI42", "EI43", "EI44", "EI45", "EI50"):
        case = next(c for c in CASES if c["id"] == identifier)
        document = cx.strict_loads((CORPUS / case["file"]).read_bytes())
        errors = list(validators["execution-envelope"].iter_errors(document))
        assert [(list(e.absolute_path), e.validator) for e in errors] == [(["deadline"], "format")]


def test_i_conformance_strict_parsing_rejects_nan() -> None:
    with pytest.raises(ValueError):
        cx.strict_loads(b'{"a": NaN}')
    with pytest.raises(ValueError):
        cx.strict_loads(b"[Infinity]")
    assert cx.strict_loads(b'{"a": 1.0}') == {"a": 1.0}


def test_i_conformance_profile_checkers_are_present() -> None:
    assert {"date-time", "uuid"} <= set(format_checker().checkers)


@pytest.mark.parametrize("entry", SCHEMA_NEGATIVE, ids=[e["id"] for e in SCHEMA_NEGATIVE])
def test_i_conformance_schema_negative_audit_class(entry: dict[str, Any]) -> None:
    assert cx.audit_file(CORPUS / entry["file"]) == entry["class"]


@pytest.mark.parametrize("entry", SCHEMA_NEGATIVE, ids=[e["id"] for e in SCHEMA_NEGATIVE])
def test_i_conformance_schema_negative_is_a_single_declared_mutation(entry: dict[str, Any]) -> None:
    canonical = {
        name: (SCHEMAS / EXPECTATIONS["schemas"][name]["file"]).read_bytes()
        for name in cx.SCHEMA_NAMES
    }
    assert cx.derive_check(entry, CORPUS, canonical)


@pytest.mark.parametrize("control", AUDIT_POSITIVE, ids=[c["id"] for c in AUDIT_POSITIVE])
def test_i_conformance_canonical_schemas_audit_ok(control: dict[str, Any]) -> None:
    path = SCHEMAS / EXPECTATIONS["schemas"][control["schema"]]["file"]
    assert cx.audit_file(path) == control["class"] == "ok"


@pytest.mark.parametrize("control", DIGEST_CONTROLS, ids=[c["id"] for c in DIGEST_CONTROLS])
def test_i_conformance_digest_control_evaluates_no_case(
    control: dict[str, Any], tmp_path: Path, monkeypatch
) -> None:
    for name in cx.SCHEMA_NAMES:
        file = EXPECTATIONS["schemas"][name]["file"]
        shutil.copyfile(SCHEMAS / file, tmp_path / file)
    target = tmp_path / EXPECTATIONS["schemas"][control["base"]]["file"]
    if control["derivation"] == "canonical-plus-lf":
        target.write_bytes(target.read_bytes() + b"\n")
    else:
        source = control["derivation"].removeprefix("bytes-of:")
        entry = next(e for e in SCHEMA_NEGATIVE if e["id"] == source)
        target.write_bytes((CORPUS / entry["file"]).read_bytes())

    evaluated: list[str] = []
    real = cx.classify_case

    def counting(case, corpus, validators):
        evaluated.append(case["id"])
        return real(case, corpus, validators)

    monkeypatch.setattr(cx, "classify_case", counting)
    with pytest.raises(cx.SchemaDigestMismatch) as raised:
        cx.run_corpus(tmp_path, CORPUS)
    assert raised.value.cls == control["class"] == "schema-digest-mismatch"
    assert raised.value.mismatched == [control["base"]]
    assert evaluated == []


def test_i_conformance_schema_negative_command_agrees(tmp_path: Path) -> None:
    out = tmp_path / "schema-negative-python.json"
    args = ["--schemas", str(SCHEMAS), "--corpus", str(CORPUS), "--out", str(out)]
    code = cx.main(["schema-negative", *args])
    result = cx.strict_loads(out.read_bytes())
    assert code == 0
    assert result["derive"] == "pass"
    assert [(r["id"], r["class"]) for r in result["results"]] == [
        (e["id"], e["class"]) for e in SCHEMA_NEGATIVE
    ]
    assert [(r["id"], r["class"]) for r in result["audit_positive"]] == [
        ("SA01", "ok"),
        ("SA02", "ok"),
    ]
    assert [(r["id"], r["class"], r["evaluated"]) for r in result["digest_controls"]] == [
        ("SD01", "schema-digest-mismatch", 0),
        ("SD02", "schema-digest-mismatch", 0),
    ]


def _flipped_corpus(tmp_path: Path, identifier: str) -> Path:
    copy = tmp_path / "corpus"
    shutil.copytree(CORPUS, copy)
    expectations = cx.strict_loads((copy / cx.EXPECTATIONS_FILE).read_bytes())
    for case in expectations["cases"]:
        if case["id"] == identifier:
            case["valid"], case["class"] = True, "valid"
    (copy / cx.EXPECTATIONS_FILE).write_text(json.dumps(expectations, indent=2) + "\n")
    return copy


def test_i_conformance_flipped_ei39_fails_naming_it(tmp_path: Path, capsys) -> None:
    corpus = _flipped_corpus(tmp_path, "EI39")
    out = tmp_path / "results-python.json"
    code = cx.main(
        ["corpus", "--schemas", str(SCHEMAS), "--corpus", str(corpus), "--out", str(out)]
    )
    assert code != 0
    assert cx.strict_loads(out.read_bytes())["mismatches"] == ["EI39"]
    assert "EI39" in capsys.readouterr().err


def test_i_conformance_corpus_command_writes_shared_result(tmp_path: Path) -> None:
    out = tmp_path / "results-python.json"
    code = cx.main(
        ["corpus", "--schemas", str(SCHEMAS), "--corpus", str(CORPUS), "--out", str(out)]
    )
    result = cx.strict_loads(out.read_bytes())
    assert code == 0
    assert result["validator"].startswith("python-jsonschema ")
    assert set(result) >= {"schema_sha256", "corpus_sha256", "results", "counts", "skipped"}
    assert result["schema_sha256"] == {
        name: EXPECTATIONS["schemas"][name]["sha256"] for name in cx.SCHEMA_NAMES
    }
    assert [r["id"] for r in result["results"]] == [c["id"] for c in CASES]


@pytest.mark.parametrize(
    ("data", "cls", "line"),
    [
        (b"", cx.NDJSON_FRAMING, None),
        (b'{"type":"done","request_id":"req-00000001"}', cx.NDJSON_FRAMING, 1),
        (b'{"type":"done","request_id":"req-00000001"}\n\n', cx.NDJSON_FRAMING, 2),
        (b"\n", cx.NDJSON_FRAMING, 1),
        (b'garbage\n{"type":"done","request_id":"req-00000001"}\n', cx.INVALID_JSON, 1),
        (b'{"type":"done","request_id":"r"}\n', cx.SCHEMA_INVALID, 1),
        (b'{"type":"done","request_id":"req-00000001"}\n', cx.VALID, None),
    ],
    ids=["empty", "no-final-lf", "trailing-blank", "only-lf", "garbage", "schema", "valid"],
)
def test_i_conformance_ndjson_rules(data: bytes, cls: str, line: int | None, validators) -> None:
    result = cx.classify_ndjson(data, validators["stream-event"])
    assert (result.cls, result.line) == (cls, line)


def test_i_conformance_ndjson_command_requires_types(tmp_path: Path, capsys) -> None:
    stream = tmp_path / "events.ndjson"
    stream.write_bytes((CORPUS / "cases/ndjson").joinpath(_file("NV01")).read_bytes())
    base = ["ndjson", "--schemas", str(SCHEMAS), "--corpus", str(CORPUS), "--file", str(stream)]
    assert cx.main([*base, "--as", "event", "--require-types", "text_delta,usage,done"]) == 0
    assert cx.main([*base, "--as", "event", "--require-types", "error"]) == 1
    assert cx.main([*base, "--as", "envelope"]) == 1  # events are not envelopes
    capsys.readouterr()


def _file(identifier: str) -> str:
    return next(c["file"] for c in CASES if c["id"] == identifier).rsplit("/", 1)[1]


def _ndjson(documents: list[dict[str, Any]]) -> bytes:
    return b"".join(json.dumps(d, separators=(",", ":")).encode() + b"\n" for d in documents)


def test_i_conformance_peer_executes_envelopes_and_writes_raw_events(
    tmp_path: Path, validators, capsys
) -> None:
    envelopes = tmp_path / "go-envelopes.ndjson"
    envelopes.write_bytes(
        _ndjson([make_envelope(), make_envelope(request_id="req-peer-0002", capability="fast")])
    )
    out = tmp_path / "python-events.ndjson"
    base = ["--schemas", str(SCHEMAS), "--corpus", str(CORPUS)]
    assert cx.main(["peer", *base, "--envelopes", str(envelopes), "--out", str(out)]) == 0
    summary = json.loads(capsys.readouterr().out)
    assert [(r["envelope"], r["scenario"]) for r in summary["responses"]] == [
        (n, s) for n in (1, 2) for s in cx.PEER_SCENARIOS
    ]
    produced = cx.classify_ndjson(out.read_bytes(), validators["stream-event"])
    assert produced.cls == cx.VALID
    assert produced.event_types == ["done", "error", "text_delta", "usage"]
    assert [d["type"] for d in produced.documents].count("done") == 6
    required = ["--require-types", "text_delta,usage,error,done"]
    assert cx.main(["ndjson", *base, "--file", str(out), "--as", "event", *required]) == 0
    capsys.readouterr()


@pytest.mark.parametrize(
    "corruption", ["truncated-last-line", "garbage-line", "invalid-envelope", "blank-line"]
)
def test_i_conformance_peer_rejects_corrupt_envelopes(
    corruption: str, tmp_path: Path, capsys
) -> None:
    good = _ndjson([make_envelope(), make_envelope(request_id="req-peer-0002")])
    corrupt = {
        "truncated-last-line": good[:-20],
        "garbage-line": good.replace(b"\n", b"\nnot json\n", 1),
        "invalid-envelope": good + _ndjson([make_envelope(capability="vision")]),
        "blank-line": good.replace(b"\n", b"\n\n", 1),
    }[corruption]
    envelopes = tmp_path / "go-envelopes.ndjson"
    envelopes.write_bytes(corrupt)
    out = tmp_path / "python-events.ndjson"
    base = ["--schemas", str(SCHEMAS), "--corpus", str(CORPUS)]
    assert cx.main(["peer", *base, "--envelopes", str(envelopes), "--out", str(out)]) == 1
    assert not out.exists()
    assert "peer-envelopes-rejected" in capsys.readouterr().err
