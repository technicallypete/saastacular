#!/usr/bin/env python3
"""Package and adoption checks for the sdlc-management bundle; standard library only.

Usage:
  check_sdlc.py package PATH [--skills-root DIR] [--literals FILE]
  check_sdlc.py adoption REPO --release NAME [--literals FILE]

Prints a JSON summary. Exit 0: pass. 1: findings. 2: usage error. Run with `python3 -B`.

Scopes. `package` is the strict export-privacy check for public skill content: forbidden private paths, hosted
remotes, private keys, non-synthetic UUIDs and `--literals` entries anywhere in the package (P6). `adoption` checks a
target repository without judging its unrelated application code or legitimate project identifiers:
- A3 looks for unresolved `{{...}}` placeholders only in adopted artifacts: the files rendered from the templates
  (and their `.sdlc-proposed` siblings) and the Markdown records under `management/<release>/`.
- A6 applies the strict P6 scan to the installed bundle skills only, and scans the rest of the repository solely for
  source-origin leaks (`--literals` entries) and private-key material. Public URLs, real UUIDs and template syntax
  in the target project are legitimate there.
Private worker checkouts (`.claude/worktrees/`, `.codex/worktrees/`) are not adopted content.
"""
import argparse
import hashlib
import json
import os
import posixpath
import re
import stat
import sys
from pathlib import Path
from urllib.parse import unquote

BUNDLE_NAME = "restruct-sdlc"
BUNDLE_SKILLS = ("sdlc-management", "claude-session-handoff", "repository-supervisor")
MANAGEMENT_SKILL = "sdlc-management"
MAX_SKILL_LINES = 200
ORIGINS = ("new", "verbatim", "adapted")
BUNDLE_ROOT_ENTRIES = {"README.md", "LICENSE", "MANIFEST.tsv", BUNDLE_NAME}
RELEASE_DIRS = ("goals", "specs", "plans", "results", "archive")
INSTANCE_NAMES = {
    "GOAL.md.tmpl": "G000-example.md",
    "SPEC.md.tmpl": "G000-example-spec.md",
    "PLAN.md.tmpl": "G000-example-plan.md",
    "RESULT.md.tmpl": "G000-example-result.md",
    "REVIEW.md.tmpl": "G000-implementation-codex-review.md",
    "PROGRESS-ARCHIVE.md.tmpl": "PROGRESS-0000-00-00.md",
}
GOAL_FIELDS = (
    "Objective", "In Scope", "Out of Scope", "Dependencies", "Completion Definition", "Execution Policy",
    "Iteration Budget", "Context / Scope Budget", "Suggested Subagents / Delegation", "Stop / Escalation Conditions",
)
IGNORE_LINES = ("/management/**/SESSION-HANDOFFS.jsonl", "/.claude/settings.local.json", "/.claude/worktrees/",
                "/.codex/worktrees/", "/.local/")
PRIVATE_ROOTS = (".claude/worktrees/", ".codex/worktrees/")  # private worker checkouts, not adopted content
EXTERNAL_SCHEMES = ("http", "https", "mailto")

# Patterns are written so their own source text does not match them (the checker scans itself).
FORBIDDEN = (
    ("home-path", re.compile(r"/ho[m]e/[A-Za-z0-9._-]+/")),
    ("users-path", re.compile(r"/User[s]/")),
    ("windows-users-path", re.compile(r"[A-Za-z]:\\User[s]\\", re.IGNORECASE)),
    ("run-user-path", re.compile(r"/run/use[r]/[0-9]+")),
    ("private-key", re.compile(r"-{5}BEGIN [A-Z ]*PRIVAT[E] KEY-{5}")),
    ("hosted-remote", re.compile(r"gitla[b]\.com[/:]|githu[b]\.com[/:]|(?<![A-Za-z0-9])gi[t]@", re.IGNORECASE)),
)
UUID_RE = re.compile(r"(?<![0-9A-Fa-f-])[0-9A-Fa-f]{8}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{12}"
                     r"(?![0-9A-Fa-f-])")
SYNTHETIC_UUID_RE = re.compile(r"00000000-0000-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{12}")
PLACEHOLDER_RE = re.compile(r"\{\{(.*?)\}\}")
PARAM_NAME_RE = re.compile(r"[A-Z][A-Z0-9_]*")
LINK_RE = re.compile(r"!?\[[^\]\n]*\]\(\s*<?([^()\s<>]+)>?(?:\s+\"[^\"\n]*\")?\s*\)")
SCHEME_RE = re.compile(r"[A-Za-z][A-Za-z0-9+.-]*:")
WINDOWS_ABS_RE = re.compile(r"[A-Za-z]:[\\/]")
SKILL_NAME_RE = re.compile(r"[a-z0-9-]+")
GOAL_ID_RE = re.compile(r"[A-Z][A-Z0-9]*-?[0-9]+")


class Report:
    def __init__(self, command, path):
        self.command = command
        self.path = str(path)
        self.checks = {}
        self.findings = []

    def ran(self, check):
        self.checks.setdefault(check, "pass")

    def fail(self, check, path, detail):
        self.checks[check] = "fail"
        self.findings.append({"check": check, "path": str(path), "detail": detail})

    def result(self):
        ok = not self.findings
        return {"command": self.command, "path": self.path, "ok": ok, "checks": dict(sorted(self.checks.items())),
                "findings": self.findings}


def read_text(path):
    return Path(path).read_bytes().decode("utf-8", errors="replace")


def sha256_file(path):
    return hashlib.sha256(Path(path).read_bytes()).hexdigest()


def strip_code(text):
    """Remove fenced blocks and inline code spans; links inside code are not links."""
    out, fence = [], None
    for line in text.splitlines():
        marker = line.lstrip()[:3]
        if fence:
            if marker == fence:
                fence = None
            out.append("")
            continue
        if marker in ("```", "~~~"):
            fence = marker
            out.append("")
            continue
        out.append(re.sub(r"`[^`]*`", " ", line))
    return "\n".join(out)


def extract_links(text):
    return [m.group(1) for m in LINK_RE.finditer(strip_code(text))]


def split_target(raw):
    """Return (kind, path): kind is external, anchor, absolute, scheme or relative."""
    target = unquote(raw.split("#", 1)[0])
    if SCHEME_RE.match(raw) and not WINDOWS_ABS_RE.match(raw):
        scheme = raw.split(":", 1)[0].lower()
        return ("external" if scheme in EXTERNAL_SCHEMES else "scheme"), target
    if not target:
        return "anchor", target
    if target.startswith(("/", "\\", "~")) or WINDOWS_ABS_RE.match(target):
        return "absolute", target
    return "relative", target


def inside(path, root):
    real, base = os.path.realpath(path), os.path.realpath(root)
    return real == base or real.startswith(base + os.sep)


def walk_files(root, skip_dirs=()):
    for dirpath, dirnames, filenames in os.walk(root, followlinks=False):
        dirnames[:] = sorted(d for d in dirnames if d not in skip_dirs)
        for name in sorted(filenames):
            yield Path(dirpath) / name


def frontmatter(text):
    if not text.startswith("---\n"):
        return None
    end = text.find("\n---\n", 4)
    if end < 0:
        return None
    fields = {}
    for line in text[4:end].splitlines():
        if line and not line[0].isspace() and ":" in line:
            key, value = line.split(":", 1)
            fields[key.strip()] = value.strip().strip("\"'")
    return fields


# ---------------------------------------------------------------- package checks

def check_p1(report, skill_dir):
    report.ran("P1")
    path = skill_dir / "SKILL.md"
    if not path.is_file():
        report.fail("P1", path, "SKILL.md missing")
        return
    fields = frontmatter(read_text(path))
    if fields is None:
        report.fail("P1", path, "frontmatter not delimited by --- lines")
        return
    name = fields.get("name", "")
    if name != skill_dir.name or not SKILL_NAME_RE.fullmatch(name):
        report.fail("P1", path, f"name {name!r} must equal the directory name and match ^[a-z0-9-]+$")
    if not fields.get("description"):
        report.fail("P1", path, "description is empty")


def check_p2(report, skill_dir):
    report.ran("P2")
    path = skill_dir / "SKILL.md"
    if path.is_file():
        lines = len(read_text(path).splitlines())
        if lines > MAX_SKILL_LINES:
            report.fail("P2", path, f"{lines} lines exceeds {MAX_SKILL_LINES}")


def shipped_docs(root):
    return [p for p in walk_files(root) if p.suffix == ".md" and not p.is_symlink()]


def resolve_shipped(skill_dir, skills_root, rel_src, target):
    """Resolve an L1/L2 link from a shipped document. Returns (path or None, error or None, l1_relpath)."""
    norm = posixpath.normpath(posixpath.join(posixpath.dirname(rel_src), target))
    parts = norm.split("/")
    if parts[0] != "..":
        dest = skill_dir / norm
        if not inside(dest, skill_dir):
            return None, "L1 target escapes the skill directory", None
        return dest, None, norm
    if len(parts) < 3 or parts[1] in ("..", "."):
        return None, "link escapes the declared skills root", None
    if parts[1] not in BUNDLE_SKILLS:
        return None, f"L2 link names unknown skill {parts[1]!r}", None
    if skills_root is None:
        return None, "L2 link without a declared skills root", None
    base = Path(skills_root) / parts[1]
    dest = base / "/".join(parts[2:])
    if not inside(dest, base):
        return None, "L2 target escapes the sibling skill", None
    return dest, None, None


def check_p3_skill(report, skill_dir, skills_root):
    report.ran("P3")
    graph = {}
    for doc in shipped_docs(skill_dir):
        rel = doc.relative_to(skill_dir).as_posix()
        graph[rel] = set()
        for raw in extract_links(read_text(doc)):
            kind, target = split_target(raw)
            if kind in ("external", "anchor"):
                continue
            if kind != "relative":
                report.fail("P3", doc, f"{kind} link {raw!r} is not allowed")
                continue
            dest, error, l1 = resolve_shipped(skill_dir, skills_root, rel, target)
            if error:
                report.fail("P3", doc, f"{raw!r}: {error}")
            elif not dest.exists():
                report.fail("P3", doc, f"{raw!r}: target does not exist")
            elif l1:
                graph[rel].add(l1)
    return graph


def check_p3_readme(report, bundle):
    report.ran("P3")
    readme = bundle / "README.md"
    if not readme.is_file():
        return
    for raw in extract_links(read_text(readme)):
        kind, target = split_target(raw)
        if kind in ("external", "anchor"):
            continue
        if kind != "relative":
            report.fail("P3", readme, f"{kind} link {raw!r} is not allowed")
            continue
        dest = bundle / posixpath.normpath(target)
        if posixpath.normpath(target).startswith("..") or not inside(dest, bundle):
            report.fail("P3", readme, f"{raw!r}: escapes the bundle root")
        elif not dest.exists():
            report.fail("P3", readme, f"{raw!r}: target does not exist")


def load_parameters(skill_dir):
    path = skill_dir / "templates" / "parameters.example.json"
    data = json.loads(read_text(path))
    params = data["parameters"]
    if not isinstance(params, dict):
        raise ValueError("parameters must be an object")
    return {name: spec.get("default") for name, spec in params.items()}


def render_value(value):
    if isinstance(value, bool):
        return "true" if value else "false"
    if isinstance(value, list):
        return ", ".join(str(v) for v in value) if value else "(none)"
    if value is None or value == "":
        return "(none)"
    return str(value)


def render(text, values):
    return PLACEHOLDER_RE.sub(lambda m: render_value(values[m.group(1)]) if m.group(1) in values else m.group(0), text)


def destination(rel_template, release):
    """Template path relative to templates/ -> adopting-repository path (instances get an example name)."""
    parts = rel_template.split("/")
    if parts[0] == "root":
        parts = parts[1:]
    if parts[:2] == ["management", "release"]:
        parts = ["management", release] + parts[2:]
    name = parts[-1]
    parts[-1] = INSTANCE_NAMES.get(name, name[:-5] if name.endswith(".tmpl") else name)
    return "/".join(parts)


def adoption_layout(templates_dir, release):
    created = {".gitignore", "management", f"management/{release}", "docs", "docs/agents"}
    created.update(f"management/{release}/{d}" for d in RELEASE_DIRS)
    for tmpl in walk_files(templates_dir):
        rel = tmpl.relative_to(templates_dir).as_posix()
        if rel.endswith(".tmpl") and tmpl.name not in INSTANCE_NAMES:
            created.add(destination(rel, release))
    return created


def installed_target_exists(target, skill_dir, skills_root):
    parts = target.split("/")
    if parts[:3] != [".agents", "skills", BUNDLE_NAME] or len(parts) < 5 or parts[3] not in BUNDLE_SKILLS:
        return False
    rest = "/".join(parts[4:])
    if parts[3] == MANAGEMENT_SKILL and skill_dir.name == MANAGEMENT_SKILL:
        base = skill_dir
    elif skills_root is not None:
        base = Path(skills_root) / parts[3]
    else:
        return False
    dest = base / rest
    return inside(dest, base) and dest.exists()


def check_p3t(report, skill_dir, skills_root):
    report.ran("P3t")
    templates_dir = skill_dir / "templates"
    try:
        values = load_parameters(skill_dir)
    except (OSError, ValueError, KeyError, json.JSONDecodeError) as exc:
        report.fail("P3t", templates_dir / "parameters.example.json", f"unreadable parameters: {exc}")
        return
    release = render_value(values.get("RELEASE", "v0"))
    layout = adoption_layout(templates_dir, release)
    for tmpl in walk_files(templates_dir):
        if not tmpl.name.endswith(".tmpl"):
            continue
        rel = tmpl.relative_to(templates_dir).as_posix()
        text = read_text(tmpl)
        for m in PLACEHOLDER_RE.finditer(text):
            name = m.group(1)
            if not PARAM_NAME_RE.fullmatch(name):
                report.fail("P3t", tmpl, f"malformed placeholder {m.group(0)!r}")
            elif name not in values:
                report.fail("P3t", tmpl, f"undeclared parameter {name!r}")
        leftover = PLACEHOLDER_RE.sub("", text)
        if "{{" in leftover or "}}" in leftover:
            report.fail("P3t", tmpl, "unbalanced placeholder braces")
        dest_dir = posixpath.dirname(destination(rel, release))
        for raw in extract_links(render(text, values)):
            kind, target = split_target(raw)
            if kind == "anchor":
                continue
            if kind != "relative":
                report.fail("P3t", tmpl, f"{kind} link {raw!r}: template links must be relative")
                continue
            norm = posixpath.normpath(posixpath.join(dest_dir, target))
            if norm == ".." or norm.startswith("../"):
                report.fail("P3t", tmpl, f"{raw!r}: escapes the adopting-repository root")
            elif norm not in layout and not installed_target_exists(norm, skill_dir, skills_root):
                report.fail("P3t", tmpl, f"{raw!r}: target {norm!r} is missing from the documented layout")


def check_p4(report, skill_dir, graph):
    report.ran("P4")
    seen, queue = {"SKILL.md"}, ["SKILL.md"]
    while queue:
        for target in sorted(graph.get(queue.pop(), ())):
            if target not in seen:
                seen.add(target)
                queue.append(target)
    for sub in ("references", "templates"):
        for path in walk_files(skill_dir / sub):
            rel = path.relative_to(skill_dir).as_posix()
            if rel not in seen:
                report.fail("P4", path, "not reachable from SKILL.md")


def check_p5(report, root):
    report.ran("P5")
    for dirpath, dirnames, filenames in os.walk(root, followlinks=False):
        for name in list(dirnames) + list(filenames):
            path = Path(dirpath) / name
            mode = os.lstat(path).st_mode
            if stat.S_ISLNK(mode):
                report.fail("P5", path, "symlink")
            elif name == ".git":
                report.fail("P5", path, "Git metadata")
            elif name == "__pycache__" or name.endswith((".pyc", ".pyo")):
                report.fail("P5", path, "bytecode")
            elif not (stat.S_ISDIR(mode) or stat.S_ISREG(mode)):
                report.fail("P5", path, "special file")
        dirnames[:] = [d for d in dirnames if not os.path.islink(os.path.join(dirpath, d))]


def load_literals(path):
    patterns = []
    if path is None:
        return patterns
    for number, line in enumerate(read_text(path).splitlines(), 1):
        literal = line.strip()
        if literal and not literal.startswith("#"):
            patterns.append((number, re.compile(r"(?<![A-Za-z0-9])" + re.escape(literal) + r"(?![A-Za-z0-9])",
                                                re.IGNORECASE)))
    return patterns


def scan_content(report, check, root, literals, skip_dirs=(".git",), skip_prefixes=(), strict=True):
    """strict: every export-privacy pattern; otherwise only private keys and the source-origin literals."""
    report.ran(check)
    patterns = FORBIDDEN if strict else tuple(p for p in FORBIDDEN if p[0] == "private-key")
    for path in walk_files(root, skip_dirs):
        if path.is_symlink() or not path.is_file():
            continue
        if path.relative_to(root).as_posix().startswith(skip_prefixes):
            continue
        text = read_text(path)
        for label, pattern in patterns:
            if pattern.search(text):
                report.fail(check, path, f"forbidden content: {label}")
        for m in UUID_RE.finditer(text) if strict else ():
            if not SYNTHETIC_UUID_RE.fullmatch(m.group(0)):
                report.fail(check, path, "non-synthetic full UUID")
                break
        for number, pattern in literals:
            if pattern.search(text):
                report.fail(check, path, f"private literal entry #{number}")


def check_p7(report, bundle):
    report.ran("P7")
    entries = {p.name for p in bundle.iterdir()}
    if entries != BUNDLE_ROOT_ENTRIES:
        report.fail("P7", bundle, f"root entries {sorted(entries)} must be exactly {sorted(BUNDLE_ROOT_ENTRIES)}")
    parent = bundle / BUNDLE_NAME
    if parent.is_dir():
        names = {p.name for p in parent.iterdir()}
        expected = {"SKILL.md", "LICENSE", "agents"} | set(BUNDLE_SKILLS)
        if names != expected or not all((parent / n).is_dir() for n in BUNDLE_SKILLS if n in names):
            report.fail("P7", parent, f"bundle entries {sorted(names)} must be exactly {sorted(expected)}")
    manifest = bundle / "MANIFEST.tsv"
    if not manifest.is_file():
        report.fail("P7", manifest, "MANIFEST.tsv missing")
        return
    listed = {}
    for number, line in enumerate(read_text(manifest).splitlines(), 1):
        fields = line.split("\t")
        if len(fields) != 3:
            report.fail("P7", manifest, f"row {number}: expected sha256<TAB>origin<TAB>path")
            continue
        digest, origin, rel = fields
        norm = posixpath.normpath(rel)
        if not re.fullmatch(r"[0-9a-f]{64}", digest):
            report.fail("P7", manifest, f"row {number}: invalid sha256")
        if origin not in ORIGINS:
            report.fail("P7", manifest, f"row {number}: invalid origin {origin!r}")
        if norm != rel or rel.startswith(("/", "..")) or rel == "MANIFEST.tsv":
            report.fail("P7", manifest, f"row {number}: invalid path {rel!r}")
            continue
        if rel in listed:
            report.fail("P7", manifest, f"row {number}: duplicate path {rel!r}")
        listed[rel] = digest
        path = bundle / rel
        if not path.is_file() or path.is_symlink():
            report.fail("P7", manifest, f"row {number}: {rel!r} is not a regular file")
        elif sha256_file(path) != digest:
            report.fail("P7", manifest, f"row {number}: hash mismatch for {rel!r}")
    actual = {p.relative_to(bundle).as_posix() for p in walk_files(bundle) if p.name != "MANIFEST.tsv"
              or p.parent != bundle}
    for rel in sorted(actual - set(listed)):
        report.fail("P7", bundle / rel, "file not listed in MANIFEST.tsv")


def check_skill(report, skill_dir, skills_root, literals, scan=True):
    check_p1(report, skill_dir)
    graph = check_p3_skill(report, skill_dir, skills_root)
    check_p5(report, skill_dir)
    if scan:
        scan_content(report, "P6", skill_dir, literals)
    if skill_dir.name == MANAGEMENT_SKILL:
        check_p2(report, skill_dir)
        check_p3t(report, skill_dir, skills_root)
        check_p4(report, skill_dir, graph)


def is_bundle(path):
    return (path / "MANIFEST.tsv").exists() or ((path / BUNDLE_NAME).is_dir() and not (path / "SKILL.md").exists())


def run_package(path, skills_root, literals_path):
    report = Report("package", path)
    literals = load_literals(literals_path)
    if is_bundle(path):
        root = Path(skills_root) if skills_root else path / BUNDLE_NAME
        check_p5(report, path)
        check_p7(report, path)
        check_p3_readme(report, path)
        scan_content(report, "P6", path, literals)
        parent = path / BUNDLE_NAME
        if parent.is_dir():
            check_p1(report, parent)
            check_p2(report, parent)
        for name in BUNDLE_SKILLS:
            skill_dir = parent / name
            if skill_dir.is_dir():
                check_skill(report, skill_dir, root, [], scan=False)
    else:
        root = Path(skills_root) if skills_root else path.parent
        check_skill(report, path, root, literals)
    return report


# ---------------------------------------------------------------- adoption checks

def adoption_docs(repo):
    docs = [repo / name for name in ("AGENTS.md", "CLAUDE.md") if (repo / name).is_file()]
    for sub in ("management", "docs"):
        if (repo / sub).is_dir():
            docs.extend(p for p in walk_files(repo / sub) if p.suffix == ".md")
    return docs


def adopted_artifacts(repo, release):
    """Files adoption renders (and their proposals) plus the release's Markdown records; never other app code."""
    templates = repo / ".agents" / "skills" / BUNDLE_NAME / MANAGEMENT_SKILL / "templates"
    paths = []
    if templates.is_dir():
        for rel in sorted(adoption_layout(templates, release)):
            for candidate in (repo / rel, repo / (rel + ".sdlc-proposed")):
                if candidate.is_file() and not candidate.is_symlink():
                    paths.append(candidate)
    release_root = repo / "management" / release
    if release_root.is_dir():
        paths.extend(p for p in walk_files(release_root) if p.suffix == ".md" and not p.is_symlink()
                     and p not in paths)
    return paths


def run_adoption(repo, release, literals_path):
    report = Report("adoption", repo)
    literals = load_literals(literals_path)
    rel_root = repo / "management" / release
    report.ran("A1")
    required = ["AGENTS.md", "CLAUDE.md", "management/BUDGET_POLICY.md", "management/SUPERVISOR-HANDOFF.md",
                f"management/{release}/ROADMAP.md", f"management/{release}/PROGRESS.md",
                f"management/{release}/BACKLOG.md"]
    for rel in required:
        if not (repo / rel).is_file():
            report.fail("A1", repo / rel, "required file missing")
    for sub in RELEASE_DIRS:
        if not (rel_root / sub).is_dir():
            report.fail("A1", rel_root / sub, "required directory missing")

    report.ran("A2")
    for doc in adoption_docs(repo):
        for raw in extract_links(read_text(doc)):
            kind, target = split_target(raw)
            if kind in ("external", "anchor"):
                continue
            if kind != "relative":
                report.fail("A2", doc, f"{kind} link {raw!r}")
                continue
            dest = doc.parent / target
            if not inside(dest, repo):
                report.fail("A2", doc, f"{raw!r}: escapes the repository")
            elif not dest.exists():
                report.fail("A2", doc, f"{raw!r}: target does not exist")
    skills_root = repo / ".agents" / "skills" / BUNDLE_NAME
    if not (skills_root / MANAGEMENT_SKILL).is_dir():
        report.fail("A2", skills_root / MANAGEMENT_SKILL, "sdlc-management is not installed")
    for name in BUNDLE_SKILLS:
        if (skills_root / name).is_dir():
            sub = Report("package", skills_root / name)
            check_p3_skill(sub, skills_root / name, skills_root)
            for finding in sub.findings:
                report.fail("A2", finding["path"], "installed skill: " + finding["detail"])

    report.ran("A3")
    for path in adopted_artifacts(repo, release):
        for m in re.finditer(r"\{\{[^{}\n]*\}\}", read_text(path)):
            report.fail("A3", path, f"unresolved placeholder {m.group(0)!r}")
            break

    report.ran("A4")
    goals = sorted((rel_root / "goals").glob("*.md")) if (rel_root / "goals").is_dir() else []
    for goal in goals:
        lines = read_text(goal).splitlines()
        title = next((line for line in lines if line.startswith("# ")), "")
        token = title[2:].split(" ", 1)[0] if title else ""
        if not GOAL_ID_RE.fullmatch(token):
            report.fail("A4", goal, "field 1 (Goal ID and title) missing")
        headings = [line[3:].strip() for line in lines if line.startswith("## ")]
        position = 0
        for field in GOAL_FIELDS:
            try:
                position = headings.index(field, position) + 1
            except ValueError:
                report.fail("A4", goal, f"field {field!r} missing or out of order")

    report.ran("A5")
    ignore = repo / ".gitignore"
    present = set(read_text(ignore).splitlines()) if ignore.is_file() else set()
    for line in IGNORE_LINES:
        if line not in present:
            report.fail("A5", ignore, f"missing ignore line {line!r}")

    installed = tuple(f".agents/skills/{BUNDLE_NAME}/{name}/" for name in BUNDLE_SKILLS)
    scan_content(report, "A6", repo, literals, skip_prefixes=PRIVATE_ROOTS + installed, strict=False)
    for name in BUNDLE_SKILLS:
        if (skills_root / name).is_dir() and not (skills_root / name).is_symlink():
            scan_content(report, "A6", skills_root / name, literals)
    return report


def main(argv=None):
    parser = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    sub = parser.add_subparsers(dest="command", required=True)
    package = sub.add_parser("package")
    package.add_argument("path")
    package.add_argument("--skills-root")
    package.add_argument("--literals")
    adoption = sub.add_parser("adoption")
    adoption.add_argument("repo")
    adoption.add_argument("--release", required=True)
    adoption.add_argument("--literals")
    args = parser.parse_args(argv)
    literals = getattr(args, "literals", None)
    if literals is not None and not Path(literals).is_file():
        parser.exit(2, f"literals file not found: {literals}\n")
    if args.command == "package":
        path = Path(args.path)
        if not path.is_dir() or not ((path / "SKILL.md").is_file() or is_bundle(path)):
            parser.exit(2, f"not a skill directory or bundle root: {path}\n")
        if args.skills_root is not None and not Path(args.skills_root).is_dir():
            parser.exit(2, f"skills root not found: {args.skills_root}\n")
        report = run_package(path, args.skills_root, literals)
    else:
        repo = Path(args.repo)
        if not repo.is_dir() or not re.fullmatch(r"[A-Za-z0-9._-]+", args.release) or args.release in (".", ".."):
            parser.exit(2, "adoption needs an existing repository directory and a plain release name\n")
        report = run_adoption(repo, args.release, literals)
    result = report.result()
    print(json.dumps(result, indent=2))
    return 0 if result["ok"] else 1


if __name__ == "__main__":
    sys.exit(main())
