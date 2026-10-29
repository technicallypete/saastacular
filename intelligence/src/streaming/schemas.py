"""Loads the cross-service JSON Schemas, the single source of the wire contract.

In the image the schemas live under ``/srv/contracts/intelligence`` (copied
from the project's ``contracts/`` directory through a named build context).
Outside the image the project-relative ``contracts/intelligence`` is used.

The validator follows Draft 2020-12 with two corrections over the library
defaults, so that it agrees with other conforming implementations:

* ``pattern`` uses ECMA-262 semantics. Python's ``$`` also matches before a
  final newline; an unescaped ``$`` outside a character class is therefore
  translated to ``\\Z``. Constructs whose meaning differs between ECMA-262 and
  Python ``re`` are rejected when the validator is built.
* ``format: date-time`` is asserted by a project-owned RFC 3339 checker with
  calendar validation (the library registers its checker only when an
  optional dependency is installed). ``uuid`` keeps the library checker.

Schemas are also audited against the supported keyword profile when the
validator is built, so a schema that relies on anything else fails fast.
"""

from __future__ import annotations

import json
import re
from collections.abc import Iterator
from dataclasses import dataclass
from functools import cache
from pathlib import Path
from typing import Any

from jsonschema import Draft202012Validator, FormatChecker, ValidationError, validators

IMAGE_CONTRACTS_DIR = Path("/srv/contracts/intelligence")
# src/streaming/schemas.py -> parents[3] is the project root.
PROJECT_CONTRACTS_DIR = Path(__file__).resolve().parents[3] / "contracts" / "intelligence"

ENVELOPE_SCHEMA = "execution-envelope.schema.json"
STREAM_EVENT_SCHEMA = "stream-event.schema.json"

DRAFT_2020_12 = "https://json-schema.org/draft/2020-12/schema"

# The audited keyword profile shared (by specification) with every other
# validator of these contracts. Anything else in a loaded schema is rejected.
PROFILE_KEYWORDS = frozenset(
    {
        "$schema",
        "$id",
        "$defs",
        "$ref",
        "title",
        "description",
        "type",
        "properties",
        "required",
        "additionalProperties",
        "items",
        "minItems",
        "maxItems",
        "minLength",
        "maxLength",
        "pattern",
        "enum",
        "const",
        "oneOf",
        "minimum",
        "format",
    }
)
PROFILE_FORMATS = frozenset({"date-time", "uuid"})
LOCAL_REF_PREFIX = "#/$defs/"


# --------------------------------------------------------------------------
# Strict JSON (RFC 8259): NaN and Infinity literals are rejected.
# --------------------------------------------------------------------------


def _reject_constant(name: str) -> Any:
    raise ValueError(f"non-standard JSON literal {name}")


def strict_loads(data: bytes | str) -> Any:
    """Parse one RFC 8259 JSON text; raise ``ValueError`` on anything else.

    Bytes must be UTF-8 (no byte-order mark). ``NaN``, ``Infinity`` and
    ``-Infinity`` are rejected, unlike ``json.loads``.
    """
    text = data.decode("utf-8") if isinstance(data, bytes) else data
    return json.loads(text, parse_constant=_reject_constant)


# --------------------------------------------------------------------------
# Keyword profile audit
# --------------------------------------------------------------------------


@dataclass(frozen=True, slots=True)
class ProfileViolation:
    """One departure from the supported profile: ``keyword``, ``format`` or ``ref``."""

    kind: str
    pointer: str
    detail: str


class SchemaProfileError(ValueError):
    """A schema uses something outside the supported profile."""

    def __init__(self, violations: list[ProfileViolation]) -> None:
        self.violations = violations
        super().__init__(
            "schema outside the supported profile: "
            + "; ".join(f"{v.kind} {v.detail} at {v.pointer or '/'}" for v in violations)
        )


def _escape_pointer(token: str) -> str:
    return token.replace("~", "~0").replace("/", "~1")


def _resolve_local(root: Any, ref: str) -> bool:
    """True when ``ref`` (``#/...``) resolves inside ``root``."""
    node = root
    for raw in ref[2:].split("/"):
        token = raw.replace("~1", "/").replace("~0", "~")
        if isinstance(node, dict) and token in node:
            node = node[token]
        else:
            return False
    return True


def _walk_schema(node: Any, pointer: str) -> Iterator[tuple[str, dict[str, Any]]]:
    """Yield every subschema object with its JSON Pointer."""
    if not isinstance(node, dict):
        return  # boolean schemas carry no keywords
    yield pointer, node
    for keyword in ("properties", "$defs"):
        children = node.get(keyword)
        if isinstance(children, dict):
            for name, child in children.items():
                yield from _walk_schema(child, f"{pointer}/{keyword}/{_escape_pointer(name)}")
    for keyword in ("items", "additionalProperties"):
        if isinstance(node.get(keyword), dict):
            yield from _walk_schema(node[keyword], f"{pointer}/{keyword}")
    if isinstance(node.get("oneOf"), list):
        for index, child in enumerate(node["oneOf"]):
            yield from _walk_schema(child, f"{pointer}/oneOf/{index}")


def profile_violations(schema: Any) -> list[ProfileViolation]:
    """Every keyword, format and ``$ref`` outside the profile, in document order."""
    violations: list[ProfileViolation] = []
    for pointer, node in _walk_schema(schema, ""):
        for keyword in node:
            if keyword not in PROFILE_KEYWORDS:
                violations.append(ProfileViolation("keyword", pointer, keyword))
        if "format" in node and node["format"] not in PROFILE_FORMATS:
            violations.append(ProfileViolation("format", pointer, str(node["format"])))
        if "$ref" in node:
            ref = node["$ref"]
            if not (
                isinstance(ref, str)
                and ref.startswith(LOCAL_REF_PREFIX)
                and _resolve_local(schema, ref)
            ):
                violations.append(ProfileViolation("ref", pointer, "non-local or unresolved"))
    return violations


# --------------------------------------------------------------------------
# ECMA-262 pattern semantics
# --------------------------------------------------------------------------


class UnsupportedPattern(ValueError):
    """A pattern uses a construct whose ECMA-262 and Python meanings differ."""


# Escapes whose meaning is identical in ECMA-262 (non-unicode mode) and Python.
_SAFE_LETTER_ESCAPES = frozenset("tnrfv")
# ECMA-262 '.' excludes every line terminator; Python's excludes only "\n".
_ECMA_DOT = "[^\\n\\r\\u2028\\u2029]"


def _hex_run(pattern: str, start: int, count: int) -> bool:
    run = pattern[start : start + count]
    return len(run) == count and all(c in "0123456789abcdefABCDEF" for c in run)


@cache
def ecma_to_python(pattern: str) -> str:
    """Translate an ECMA-262 pattern to an equivalent Python ``re`` pattern.

    Raises ``UnsupportedPattern`` for ``\\d``, ``\\w``, ``\\s``, ``\\b``,
    ``\\B``, ``\\p``, ``(?`` (lookarounds, flags, named groups), backreferences,
    other letter or digit escapes, and empty or negated-empty classes.
    """
    out: list[str] = []
    in_class = False
    i = 0
    n = len(pattern)
    while i < n:
        char = pattern[i]
        if char == "\\":
            if i + 1 >= n:
                raise UnsupportedPattern("trailing backslash")
            nxt = pattern[i + 1]
            if nxt.isascii() and nxt.isalnum():
                if nxt in _SAFE_LETTER_ESCAPES:
                    out.append(pattern[i : i + 2])
                    i += 2
                    continue
                if nxt == "x" and _hex_run(pattern, i + 2, 2):
                    out.append(pattern[i : i + 4])
                    i += 4
                    continue
                if nxt == "u" and _hex_run(pattern, i + 2, 4):
                    out.append(pattern[i : i + 6])
                    i += 6
                    continue
                raise UnsupportedPattern(f"escape \\{nxt} is not supported")
            out.append(pattern[i : i + 2])  # identity escape of a symbol
            i += 2
            continue
        if in_class:
            if char == "]":
                in_class = False
            elif char == "[":
                raise UnsupportedPattern("'[' inside a character class")
            out.append(char)
            i += 1
            continue
        if char == "[":
            body = 2 if pattern.startswith("[^", i) else 1
            if pattern.startswith("]", i + body):
                raise UnsupportedPattern("empty or negated-empty character class")
            in_class = True
            out.append(pattern[i : i + body])
            i += body
            continue
        if char == "(" and pattern.startswith("(?", i):
            raise UnsupportedPattern("'(?' groups and lookarounds are not supported")
        if char == "$":
            out.append("\\Z")
        elif char == ".":
            out.append(_ECMA_DOT)
        else:
            out.append(char)
        i += 1
    if in_class:
        raise UnsupportedPattern("unterminated character class")
    return "".join(out)


@cache
def _compiled(pattern: str) -> re.Pattern[str]:
    return re.compile(ecma_to_python(pattern))


def _ecma_pattern(validator: Any, pattern: str, instance: Any, schema: Any) -> Iterator[Any]:
    if validator.is_type(instance, "string") and not _compiled(pattern).search(instance):
        # The value is never echoed into the error message.
        yield ValidationError("string does not match the schema pattern")


ContractValidator = validators.extend(Draft202012Validator, {"pattern": _ecma_pattern})


def _patterns(schema: Any) -> Iterator[str]:
    for _, node in _walk_schema(schema, ""):
        if isinstance(node.get("pattern"), str):
            yield node["pattern"]


# --------------------------------------------------------------------------
# RFC 3339 date-time
# --------------------------------------------------------------------------

_DATE_TIME = re.compile(
    r"([0-9]{4})-([0-9]{2})-([0-9]{2})[Tt]"
    r"([0-9]{2}):([0-9]{2}):([0-9]{2})(?:\.[0-9]+)?"
    r"(?:([Zz])|([+-])([0-9]{2}):([0-9]{2}))"
)
_MONTH_DAYS = (31, 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31)


def _is_leap_year(year: int) -> bool:
    return year % 4 == 0 and (year % 100 != 0 or year % 400 == 0)


def is_rfc3339_date_time(value: object) -> bool:
    """RFC 3339 §5.6 ``date-time`` with calendar and leap-second rules.

    Non-strings are not this format's concern (JSON Schema semantics).
    """
    if not isinstance(value, str):
        return True
    match = _DATE_TIME.fullmatch(value)
    if match is None:
        return False
    year, month, day, hour, minute, second = (int(match.group(k)) for k in range(1, 7))
    if not 1 <= month <= 12:
        return False
    days = 29 if month == 2 and _is_leap_year(year) else _MONTH_DAYS[month - 1]
    if not 1 <= day <= days:
        return False
    if hour > 23 or minute > 59 or second > 60:
        return False
    offset_minutes = 0
    if match.group(7) is None:
        offset_hour, offset_minute = int(match.group(9)), int(match.group(10))
        if offset_hour > 23 or offset_minute > 59:
            return False
        offset_minutes = offset_hour * 60 + offset_minute
        if match.group(8) == "-":
            offset_minutes = -offset_minutes
    if second == 60:
        # A leap second is only possible at 23:59 UTC.
        utc_minutes = (hour * 60 + minute - offset_minutes) % (24 * 60)
        if utc_minutes != 23 * 60 + 59:
            return False
    return True


def _check_date_time(value: object) -> bool:
    if not is_rfc3339_date_time(value):
        raise ValueError("not an RFC 3339 date-time")
    return True


def format_checker() -> FormatChecker:
    """Exactly the profile formats: project ``date-time`` and library ``uuid``."""
    checker = FormatChecker(formats=("uuid",))
    checker.checks("date-time", raises=ValueError)(_check_date_time)
    return checker


# --------------------------------------------------------------------------
# Validators
# --------------------------------------------------------------------------


def build_validator(schema: Any) -> Any:
    """A meta-schema-checked, profile-audited validator with format assertions.

    Raises ``jsonschema.SchemaError`` (meta-invalid), ``SchemaProfileError``
    (outside the profile) or ``UnsupportedPattern``.
    """
    Draft202012Validator.check_schema(schema)
    violations = profile_violations(schema)
    if violations:
        raise SchemaProfileError(violations)
    for pattern in _patterns(schema):
        _compiled(pattern)
    checker = format_checker()
    missing = PROFILE_FORMATS - checker.checkers.keys()
    if missing:
        raise SchemaProfileError(
            [ProfileViolation("format", "", f"no checker for {name}") for name in sorted(missing)]
        )
    return ContractValidator(schema, format_checker=checker)


def contracts_dir() -> Path:
    for candidate in (IMAGE_CONTRACTS_DIR, PROJECT_CONTRACTS_DIR):
        if (candidate / ENVELOPE_SCHEMA).is_file():
            return candidate
    raise FileNotFoundError("intelligence contract schemas not found")


def load_schema(name: str) -> dict[str, Any]:
    with (contracts_dir() / name).open(encoding="utf-8") as fh:
        return json.load(fh)


@cache
def validator(name: str) -> Any:
    """The contract validator for one schema file (built once per process)."""
    return build_validator(load_schema(name))


def envelope_validator() -> Any:
    return validator(ENVELOPE_SCHEMA)


def event_validator() -> Any:
    return validator(STREAM_EVENT_SCHEMA)
