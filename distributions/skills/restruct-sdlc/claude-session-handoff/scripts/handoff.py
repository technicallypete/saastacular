#!/usr/bin/env python3
"""Compact Linux Claude Code protocol-1 handoffs; stdlib only."""
import argparse
from collections import Counter
import contextlib
import errno
import fcntl
import hashlib
import json
import os
from pathlib import Path
import re
import shlex
import socket
import stat
import struct
import subprocess
import sys
import time
import uuid
from datetime import datetime, timezone

# Protocol 1: 2.1.296 reverified by authenticated G051 handoff replay and
# acknowledgment on 2026-10-10. Versions are diagnostic baselines, not transport gates.
BASELINE_VERSIONS = {"2.1.290", "2.1.291", "2.1.296"}
# The read-only creation gate accepts only the version whose creation interface was
# read locally. A change is a bounded escalation; never widen this automatically.
CREATION_VERIFIED_VERSIONS = {"2.1.296"}
PIN_SCHEMA = "claude-session-handoff/pin/v1"
MANIFEST_SCHEMA = "claude-session-handoff/compaction-manifest/v1"
DEFAULT_RECORD = "management/SESSION-HANDOFFS.jsonl"
# Committed helper locations, in order: the source-maintainer bundle, then an adopting repository's installed
# bundle. A symlinked .agents/skills/restruct-sdlc alias is never a tree path in HEAD, so the choice is deterministic.
HELPER_CANDIDATES = ("distributions/skills/restruct-sdlc/claude-session-handoff/scripts/handoff.py",
                     ".agents/skills/restruct-sdlc/claude-session-handoff/scripts/handoff.py")
# Retired standalone helper location. Scanned only (never selected as the writer) so a helper or observer still
# running from it during the migration/cutover window keeps the in-flight-helper refusal; absent paths are ignored.
LEGACY_HELPER_RELATIVE = ".agents/skills/claude-session-handoff/scripts/handoff.py"
COMMANDS = ("discover", "check", "send", "receipt", "pin", "status", "unpin", "compact")
PROC = Path("/proc")
CLI_TIMEOUT = 15
TASK_RE = re.compile(r"G\d{3}")
HEX64_RE = re.compile(r"[0-9a-f]{64}")
DELIVERY_OUTCOMES = {"receipt-confirmed", "delivery-unconfirmed"}
# Bytes this process loaded, hashed once at start; later edits to the file do not change it.
SELF_SHA256 = hashlib.sha256(Path(__file__).resolve().read_bytes()).hexdigest()


class Blocked(Exception):
    def __init__(self, reason, details=None):
        super().__init__(reason)
        self.details = details or {}


def require(condition, reason):
    if not condition:
        raise Blocked(reason)


def digest(data):
    return hashlib.sha256(data).hexdigest()


def canonical_uuid(value):
    """Full lowercase UUID text only; names and short IDs are never identities."""
    require(isinstance(value, str) and str(uuid.UUID(value)) == value, "invalid-uuid")
    return value


def utc_now():
    return datetime.now(timezone.utc)


def git(root, *arguments, text=True):
    return subprocess.run(["git", "--no-optional-locks", "-C", str(root), *arguments],
                          capture_output=True, text=text)


def repo_root(directory):
    result = subprocess.run(
        ["git", "--no-optional-locks", "-C", str(directory), "rev-parse", "--show-toplevel"],
        capture_output=True, text=True,
    )
    require(result.returncode == 0, "repository-unavailable")
    return Path(result.stdout.strip()).resolve()


def repo_identity(directory):
    result = subprocess.run(
        ["git", "--no-optional-locks", "-C", str(directory), "rev-parse",
         "--path-format=absolute", "--git-common-dir"], capture_output=True, text=True,
    )
    require(result.returncode == 0, "repository-unavailable")
    return Path(result.stdout.strip()).resolve()


def main_checkout(root):
    result = git(root, "worktree", "list", "--porcelain")
    require(result.returncode == 0, "repository-unavailable")
    first = result.stdout.splitlines()[0] if result.stdout else ""
    require(first.startswith("worktree "), "repository-unavailable")
    return Path(first[len("worktree "):]).resolve()


def worktree_paths(root):
    result = git(root, "worktree", "list", "--porcelain")
    require(result.returncode == 0, "repository-unavailable")
    return [Path(line[len("worktree "):]) for line in result.stdout.splitlines() if line.startswith("worktree ")]


NOT_A_REPOSITORY = "fatal: not a git repository (or any of the parent directories): .git"
GIT_DISCOVERY_VARIABLES = ("GIT_DIR", "GIT_WORK_TREE", "GIT_COMMON_DIR", "GIT_CEILING_DIRECTORIES",
                           "GIT_DISCOVERY_ACROSS_FILESYSTEM")


def same_repo(cwd, identity):
    """True or False when a recorded cwd is established to be inside or outside this repository; None otherwise.

    False needs a positive determination: either another repository, or an existing, listable directory for which
    Git, run in the C locale without discovery-affecting environment variables, reports exactly its canonical
    "not a git repository" diagnosis (exit 128). Missing or inaccessible paths, unsafe or dubious ownership, an
    invalid Git context and any other Git outcome stay unresolved (None).
    """
    if not isinstance(cwd, str) or not cwd.startswith("/"):
        return None
    try:
        if not stat.S_ISDIR(os.stat(cwd).st_mode):
            return None
        with os.scandir(cwd):
            pass
    except OSError:
        return None
    environment = {key: value for key, value in os.environ.items() if key not in GIT_DISCOVERY_VARIABLES}
    environment.update(LC_ALL="C", LANGUAGE="C")
    try:
        result = subprocess.run(["git", "--no-optional-locks", "-C", cwd, "rev-parse", "--path-format=absolute",
                                 "--git-common-dir"], capture_output=True, text=True, env=environment)
    except OSError:
        return None
    if result.returncode == 0 and result.stdout.strip():
        return Path(result.stdout.strip()).resolve() == identity
    if result.returncode == 128 and result.stderr.strip() == NOT_A_REPOSITORY:
        return False
    return None


def management_path(root, relative):
    path = Path(relative)
    require(not path.is_absolute() and ".." not in path.parts, "invalid-management-path")
    require(path.parts and path.parts[0] == "management", "outside-management")
    resolved = (root / path).resolve()
    require(resolved.is_relative_to(root / "management"), "escaped-management-path")
    require((root / "management").resolve() == root / "management", "redirected-management-root")
    return resolved


def real_management_file(root, relative, reason):
    """(path, bytes) of a regular file under management/ reached through no symlink, read through an
    O_NOFOLLOW descriptor that is the same inode lstat checked, so every name, class and tracked check
    applies to the bytes actually read. (None, None) when the path does not exist."""
    try:
        path = management_path(root, relative)
    except (RuntimeError, OSError):
        raise Blocked(reason)
    require(path == Path(root).resolve() / Path(relative), reason)
    try:
        info = os.lstat(path)
    except FileNotFoundError:
        return None, None
    require(stat.S_ISREG(info.st_mode), reason)
    try:
        fd = os.open(path, os.O_RDONLY | os.O_NOFOLLOW)
    except OSError:
        raise Blocked(reason)
    with os.fdopen(fd, "rb") as stream:
        opened = os.fstat(stream.fileno())
        require(stat.S_ISREG(opened.st_mode)
                and (opened.st_dev, opened.st_ino) == (info.st_dev, info.st_ino), reason)
        return path, stream.read()


def is_canonical_uuid(value):
    try:
        canonical_uuid(value)
        return True
    except (Blocked, ValueError, TypeError, AttributeError):
        return False


def record_path(root, relative):
    """Canonical delivery-log path; symlinked components are refused."""
    require(Path(relative).suffix == ".jsonl", "record-must-be-jsonl")
    path = management_path(root, relative)
    require(path == Path(root).resolve() / Path(relative), "record-unsafe")
    return path


def live_registration(registry, session, root):
    matches = []
    for path in registry.glob("*.json"):
        try:
            row = json.loads(path.read_text())
            if row.get("sessionId") != session:
                continue
            pid = row["pid"]
            require(type(pid) is int and pid > 0, "invalid-pid")
            os.kill(pid, 0)
            start = Path(f"/proc/{pid}/stat").read_text().rsplit(")", 1)[1].split()[19]
            if str(row.get("procStart")) != start:
                continue
            if repo_identity(Path(row["cwd"])) == repo_identity(root):
                matches.append(row)
        except (OSError, ValueError, KeyError, Blocked):
            continue
    require(len(matches) == 1, "missing-or-ambiguous-live-session")
    row = matches[0]
    require(type(row.get("peerProtocol")) is int and row["peerProtocol"] == 1, "changed-peer-protocol")
    return row


def tail_records(path, max_bytes=131072):
    with path.open("rb") as stream:
        size = path.stat().st_size
        start = max(0, size - max_bytes)
        stream.seek(start)
        data = stream.read(max_bytes)
    lines = data.decode("utf-8", errors="replace").splitlines()
    if start:
        lines = lines[1:]
    for line in lines:
        try:
            row = json.loads(line)
            if isinstance(row, dict):
                yield row
        except ValueError:
            continue


def texts(row):
    content = row.get("message", {}).get("content", [])
    if isinstance(content, str):
        return [content]
    return [item.get("text", "") for item in content
            if isinstance(item, dict) and item.get("type") == "text"]


def transcripts(projects, session, child=None):
    parents = list(projects.glob(f"*/{session}.jsonl"))
    if child:
        require(re.fullmatch(r"[a-zA-Z0-9_-]+", child), "invalid-child-id")
        return [p.with_suffix("") / "subagents" / f"agent-{child}.jsonl"
                for p in parents if (p.with_suffix("") / "subagents" / f"agent-{child}.jsonl").exists()]
    return parents


def receipt_status(projects, session, handoff_id):
    """Receipt needs a user replay; an enqueue row alone is diagnostic only."""
    marker = f"[handoff-id: {handoff_id}]"
    found = queued = False
    for path in transcripts(projects, session):
        for row in tail_records(path, max_bytes=2097152):
            if row.get("type") == "user" and any(marker in text for text in texts(row)):
                found = True
            elif (row.get("type") == "queue-operation" and row.get("operation") == "enqueue"
                  and isinstance(row.get("content"), str) and marker in row["content"]):
                queued = True
    return {"status": "receipt-confirmed" if found else "delivery-unconfirmed",
            "handoff_id": handoff_id, "execution_confirmed": False,
            "queued_observed": queued and not found}


def activity(projects, session):
    paths = transcripts(projects, session)
    return max((p.stat().st_mtime for p in paths), default=0)


def newer_sessions(registry, projects, root, selected):
    """Surface potentially stale targets; never choose or message a replacement."""
    candidates = set()
    for path in registry.glob("*.json"):
        try:
            value = json.loads(path.read_text()).get("sessionId")
            if value and value != selected:
                candidates.add(value)
        except (OSError, ValueError):
            continue
    result = []
    selected_time = activity(projects, selected)
    for session in candidates:
        try:
            row = live_registration(registry, session, root)
            stamp = activity(projects, session)
            if stamp > selected_time:
                result.append({"session": session, "pid": row["pid"],
                               "checkout": row["cwd"], "transcript_updated": stamp})
        except (OSError, ValueError, KeyError, Blocked):
            continue
    return sorted(result, key=lambda r: r["transcript_updated"], reverse=True)[:5]


def safe_excerpt(text, limit):
    text = re.sub(r"\x1b\][^\x07]*(?:\x07|\x1b\\)|\x1b\[[0-?]*[ -/]*[@-~]", "", text)
    text = re.sub(r"[\x00-\x08\x0b-\x1f\x7f]", "", text)
    text = re.sub(r"\b(?:sk_[A-Za-z0-9_-]+|glpat-[A-Za-z0-9_-]+|eyJ[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+)\b", "[REDACTED]", text)
    return text[:limit]


def safe_name(value):
    return safe_excerpt(value, 120) if isinstance(value, str) else None


def append_record(root, relative, record):
    """Append one JSON line under the log's inode lock, always starting on a line boundary."""
    path = record_path(root, relative)
    path.parent.mkdir(parents=True, exist_ok=True)
    require(management_path(root, relative) == path, "record-path-drift")
    try:
        fd = os.open(path, os.O_RDWR | os.O_APPEND | os.O_CREAT | os.O_NOFOLLOW, 0o600)
    except OSError as error:
        require(error.errno != errno.ELOOP, "record-unsafe")
        raise
    try:
        fcntl.flock(fd, fcntl.LOCK_EX)
        info = os.fstat(fd)
        require(stat.S_ISREG(info.st_mode) and info.st_uid == os.getuid(), "record-unsafe")
        data = (json.dumps(record, separators=(",", ":")) + "\n").encode()
        if info.st_size and os.pread(fd, 1, info.st_size - 1) != b"\n":
            data = b"\n" + data
        os.write(fd, data)
    finally:
        os.close(fd)


def compatibility_warnings(registration):
    return [] if registration.get("version") in BASELINE_VERSIONS else ["unbaselined-claude-version"]


def connect_peer(registry, registration):
    """Prepare an authenticated peer connection; no bytes are written here."""
    pid = registration["pid"]
    address = registration["messagingSocketPath"]
    require(isinstance(address, str) and Path(address).is_absolute() and os.path.realpath(address) == address,
            "noncanonical-socket")
    require(stat.S_ISSOCK(os.stat(address).st_mode), "not-a-unix-socket")
    key = registry / f"{pid}.{digest(address.encode())}.key"
    info = key.stat()
    require(info.st_uid == os.getuid() and info.st_mode & 0o077 == 0, "unsafe-peer-key")
    token = json.loads(key.read_text())["peerToken"]
    require(isinstance(token, str) and bool(token), "unsupported-peer-key-shape")
    client = socket.socket(socket.AF_UNIX, socket.SOCK_STREAM)
    try:
        client.settimeout(5)
        client.connect(address)
        peer_pid, peer_uid, _ = struct.unpack("3i", client.getsockopt(socket.SOL_SOCKET, socket.SO_PEERCRED, 12))
        require(peer_pid == pid and peer_uid == os.getuid(), "socket-peer-mismatch")
    except BaseException:
        client.close()
        raise
    return client, token


def send(registry, registration, message, handoff_id):
    # Refresh the same UUID from the same repository before every connection, not a cached socket/PID.
    root = repo_root(Path(registration["cwd"]))
    session = canonical_uuid(registration["sessionId"])
    for attempt in range(2):
        current = live_registration(registry, session, root)
        try:
            client, token = connect_peer(registry, current)
        except OSError as error:
            # Only these pre-write failures prove nothing could have been delivered. Never retry sendall.
            if attempt == 0 and error.errno in (errno.ENOENT, errno.ECONNREFUSED):
                continue
            raise
        frame = {"type": "user", "session_id": session,
                 "uuid": handoff_id, "from": "codex-supervisor", "priority": "next",
                 "message": {"role": "user", "content": message}}
        with client:
            client.sendall((json.dumps({"type": "auth", "token": token}) + "\n"
                            + json.dumps(frame) + "\n").encode())
        return current


# --- private machine-local state (pin, maintenance lock, archives) ----------------------------

def state_root():
    return Path.home() / ".local" / "state" / "claude-session-handoff"


def private_dir(path, create):
    """A user-owned 0700 directory that is not a symlink; None when absent and not created."""
    try:
        info = os.lstat(path)
    except FileNotFoundError:
        if not create:
            return None
        os.mkdir(path, 0o700)
        info = os.lstat(path)
    require(stat.S_ISDIR(info.st_mode), "state-unsafe")
    require(info.st_uid == os.getuid() and info.st_mode & 0o077 == 0, "state-unsafe")
    return path


def repo_key(root):
    return digest(str(repo_identity(root)).encode())


def repo_state(root, create=True):
    base = state_root()
    if create:
        base.parent.mkdir(parents=True, exist_ok=True)
    if private_dir(base, create) is None:
        return None
    return private_dir(base / repo_key(root), create)


def state_subdir(root, name):
    return private_dir(repo_state(root) / name, True)


def read_private(path):
    try:
        fd = os.open(path, os.O_RDONLY | os.O_NOFOLLOW)
    except FileNotFoundError:
        return None
    except OSError as error:
        require(error.errno != errno.ELOOP, "state-unsafe")
        raise
    with os.fdopen(fd, "rb") as stream:
        info = os.fstat(stream.fileno())
        require(stat.S_ISREG(info.st_mode) and info.st_uid == os.getuid()
                and info.st_mode & 0o077 == 0, "state-unsafe")
        return stream.read()


def fsync_dir(path):
    fd = os.open(path, os.O_RDONLY | os.O_DIRECTORY)
    try:
        os.fsync(fd)
    finally:
        os.close(fd)


def write_new_private(path, data):
    """Create a new 0600 file (never overwrite or follow a link) and fsync it."""
    fd = os.open(path, os.O_WRONLY | os.O_CREAT | os.O_EXCL | os.O_NOFOLLOW, 0o600)
    try:
        view = memoryview(data)
        while view:
            view = view[os.write(fd, view):]
        os.fsync(fd)
    finally:
        os.close(fd)


def write_private_atomic(directory, name, data):
    temporary = directory / f".{name}.{uuid.uuid4().hex}.tmp"
    try:
        write_new_private(temporary, data)
        os.replace(temporary, directory / name)
    finally:
        with contextlib.suppress(FileNotFoundError):
            os.unlink(temporary)
    fsync_dir(directory)


def read_pin(root):
    """(state, pin, reason) where state is unpinned, pinned or unsafe."""
    try:
        directory = repo_state(root, create=False)
        data = None if directory is None else read_private(directory / "pin.json")
    except Blocked as error:
        return "unsafe", None, str(error)
    if data is None:
        return "unpinned", None, None
    try:
        pin = json.loads(data)
        require(isinstance(pin, dict) and pin.get("schema") == PIN_SCHEMA, "pin-schema")
        require(pin.get("repository") == str(repo_identity(root)), "pin-repository")
        canonical_uuid(pin.get("session"))
    except Blocked as error:
        return "unsafe", None, str(error)
    except (ValueError, TypeError, AttributeError):
        return "unsafe", None, "pin-malformed"
    return "pinned", pin, None


@contextlib.contextmanager
def maintenance_lock(root, log, exclusive, wait_seconds):
    """Stable lock beside the pin, never replaced by compaction: shared for send, exclusive for apply."""
    path = state_subdir(root, "locks") / f"{digest(str(log).encode())}.lock"
    try:
        fd = os.open(path, os.O_RDWR | os.O_CREAT | os.O_NOFOLLOW, 0o600)
    except OSError as error:
        require(error.errno != errno.ELOOP, "state-unsafe")
        raise
    try:
        info = os.fstat(fd)
        require(stat.S_ISREG(info.st_mode) and info.st_uid == os.getuid()
                and info.st_mode & 0o077 == 0, "state-unsafe")
        mode = fcntl.LOCK_EX if exclusive else fcntl.LOCK_SH
        deadline = time.monotonic() + wait_seconds
        while True:
            try:
                fcntl.flock(fd, mode | fcntl.LOCK_NB)
                break
            except BlockingIOError:
                require(time.monotonic() < deadline, "maintenance-lock-busy")
                time.sleep(0.05)
        yield
    finally:
        os.close(fd)


# --- duplicate-safe discovery -------------------------------------------------------------------

def run_claude(arguments, timeout=CLI_TIMEOUT):
    """The only CLI entry point: read-only listing and version queries. It never starts a session."""
    listing = (len(arguments) == 5 and arguments[:4] == ["agents", "--json", "--all", "--cwd"]
               and isinstance(arguments[4], str) and arguments[4].startswith("/"))
    require(arguments == ["--version"] or listing, "unsupported-cli-query")
    return subprocess.run(["claude", *arguments], capture_output=True, text=True, timeout=timeout)


def cli_rows(main):
    try:
        result = run_claude(["agents", "--json", "--all", "--cwd", str(main)])
    except (OSError, subprocess.TimeoutExpired):
        return None
    if result.returncode != 0:
        return None
    try:
        rows = json.loads(result.stdout)
    except ValueError:
        return None
    if not isinstance(rows, list) or not all(isinstance(row, dict) for row in rows):
        return None
    return rows


def cli_version():
    try:
        result = run_claude(["--version"])
    except (OSError, subprocess.TimeoutExpired):
        return None
    match = re.match(r"\s*(\d+\.\d+\.\d+)\b", result.stdout or "") if result.returncode == 0 else None
    return match.group(1) if match else None


def process_start(pid):
    try:
        return (PROC / str(pid) / "stat").read_text().rsplit(")", 1)[1].split()[19]
    except (OSError, IndexError):
        return None


def classify_pid(pid, proc_start):
    if type(pid) is not int or pid <= 0:
        return "unverifiable"
    try:
        os.kill(pid, 0)
    except ProcessLookupError:
        return "dead"
    except PermissionError:
        # Another user's process: a different start time proves the registration's PID was reused.
        start = process_start(pid)
        return "start-mismatch" if start is not None and start != str(proc_start) else "unverifiable"
    start = process_start(pid)
    if start is None:
        return "unverifiable"
    return "verified-live" if str(proc_start) == start else "start-mismatch"


def pid_visible(pid):
    return (PROC / str(pid)).is_dir()


def cli_live(row):
    return type(row.get("pid")) is int and row["pid"] > 0


def discovery(root, registry):
    """Registry plus CLI listing; any blind spot is unknown, never 'no coordinator'."""
    identity = repo_identity(root)
    reasons = set()
    registry_rows = []
    if not registry.is_dir():
        reasons.add("registry-unavailable")
    else:
        for path in sorted(registry.glob("*.json")):
            try:
                row = json.loads(path.read_text())
                require(isinstance(row, dict), "unreadable")
            except (OSError, ValueError, Blocked):
                # An unreadable registration could be a live same-repository session: unknown, not absent.
                registry_rows.append({"file": path.name, "classification": "unreadable"})
                reasons.add("registry-row-unreadable")
                continue
            classification = classify_pid(row.get("pid"), row.get("procStart"))
            in_repo = same_repo(row.get("cwd"), identity)
            if classification == "verified-live" and in_repo is None:
                reasons.add("live-cwd-unresolved")
            if classification == "unverifiable" and in_repo is not False:
                reasons.add("registry-row-unverifiable")
            registry_rows.append({"file": path.name, "session": row.get("sessionId"), "pid": row.get("pid"),
                                  "classification": classification, "cwd": row.get("cwd"),
                                  "version": row.get("version"), "peer_protocol": row.get("peerProtocol"),
                                  "name": safe_name(row.get("name")), "same_repository": in_repo})
    listing = cli_rows(main_checkout(root))
    cli = None
    if listing is None:
        reasons.add("cli-listing-unavailable")
    else:
        cli = []
        for row in listing:
            live = cli_live(row)
            in_repo = same_repo(row.get("cwd"), identity)
            visible = pid_visible(row["pid"]) if live else None
            if live and not visible:
                reasons.add("pid-visibility-unproven")
            if live and in_repo is None:
                reasons.add("live-cwd-unresolved")
            cli.append({"id": row.get("id"), "session": row.get("sessionId"), "name": safe_name(row.get("name")),
                        "kind": row.get("kind"), "state": row.get("state"), "pid": row.get("pid"),
                        "cwd": row.get("cwd"), "same_repository": in_repo, "pid_visible": visible})
    live_sessions = []
    seen = set()
    for row in registry_rows:
        if row.get("classification") == "verified-live" and row.get("same_repository") is True:
            seen.add(row["session"])
            live_sessions.append({k: row[k] for k in ("session", "pid", "cwd", "name", "version", "peer_protocol")}
                                 | {"source": "registry"})
    for row in cli or []:
        if cli_live(row) and row["same_repository"] is True and row["session"] not in seen:
            seen.add(row["session"])
            live_sessions.append({"session": row["session"], "pid": row["pid"], "cwd": row["cwd"],
                                  "name": row["name"], "source": "cli"})
    names = {}
    for row in [r for r in registry_rows if r.get("same_repository") is True] + [r for r in cli or [] if r["same_repository"] is True]:
        if row.get("name") and row.get("session"):
            names.setdefault(row["name"], set()).add(row["session"])
    registrations = Counter(r["session"] for r in registry_rows
                            if r.get("same_repository") is True and r.get("session"))
    return {
        "state": "unknown" if reasons else "complete",
        "reasons": sorted(reasons),
        "registry": [r for r in registry_rows if r.get("same_repository") is not False],
        "cli": None if cli is None else [r for r in cli if r["same_repository"] is not False],
        "live_sessions": live_sessions,
        "duplicate_names": {name: sorted(s) for name, s in sorted(names.items()) if len(s) > 1},
        "duplicate_registrations": {s: n for s, n in sorted(registrations.items()) if n > 1},
        "stopped_or_done": [{"id": r["id"], "session": r["session"], "kind": r["kind"], "state": r["state"]}
                            for r in cli or [] if r["state"] in ("stopped", "done")],
        "stale_registrations": [r["file"] for r in registry_rows
                                if r.get("classification") in ("dead", "start-mismatch")
                                and r.get("same_repository") is not False],
    }


def live_rows(registry, session, root):
    identity = repo_identity(root)
    rows = []
    for path in registry.glob("*.json"):
        try:
            row = json.loads(path.read_text())
        except (OSError, ValueError):
            continue
        if (isinstance(row, dict) and row.get("sessionId") == session
                and classify_pid(row.get("pid"), row.get("procStart")) == "verified-live"
                and same_repo(row.get("cwd"), identity) is True):
            rows.append(row)
    return rows


def pinned_route(registry, session, root):
    rows = live_rows(registry, session, root)
    if len(rows) > 1:
        return "pinned-ambiguous", None
    if not rows:
        return "pinned-not-live", None
    row = rows[0]
    if type(row.get("peerProtocol")) is not int or row["peerProtocol"] != 1:
        return "pinned-unverified", row
    return "pinned-live-verified", row


def public_pin(pin):
    return {k: pin.get(k) for k in ("session", "checkout_at_pin", "name_at_pin", "cli_version_at_pin",
                                    "selection_ref", "pinned_at", "replaces")}


def resolve_target(session, root, registry):
    """Explicit --session routes as before; otherwise only a verified live pin routes."""
    if session:
        uuid.UUID(session)
        return session, "explicit"
    state, pin, _ = read_pin(root)
    require(state != "unsafe", "default-route-pin-unsafe")
    require(state == "pinned", "default-route-unpinned")
    route, _ = pinned_route(registry, pin["session"], root)
    require(route == "pinned-live-verified", f"default-route-{route}")
    return pin["session"], "preferred"


def mark_peers(peers, root):
    state, pin, _ = read_pin(root)
    if state != "pinned":
        return peers
    return [{**peer, "shares_pinned_checkout": peer.get("checkout") == pin.get("checkout_at_pin")}
            for peer in peers]


def handoff_sessions(root, relative, handoff_id):
    path = record_path(root, relative)
    try:
        fd = os.open(path, os.O_RDONLY | os.O_NOFOLLOW)
    except FileNotFoundError:
        return set()
    with os.fdopen(fd, "rb") as stream:
        data = stream.read()
    sessions = set()
    for line in data.splitlines():
        try:
            row = json.loads(line)
        except ValueError:
            continue
        if isinstance(row, dict) and row.get("handoff_id") == handoff_id and isinstance(row.get("session"), str):
            sessions.add(row["session"])
    return sessions


# --- status, pin and the read-only creation gate --------------------------------------------------

def status_report(root, registry, projects):
    found = discovery(root, registry)
    state, pin, reason = read_pin(root)
    if state == "unsafe":
        return {"status": "pin-unsafe", "reason": reason, "discovery": found}
    if state == "unpinned":
        status = "unpinned" if found["state"] == "complete" else "discovery-unknown"
        return {"status": status, "candidates": found["live_sessions"], "discovery": found}
    route, row = pinned_route(registry, pin["session"], root)
    if route == "pinned-not-live" and found["state"] != "complete":
        route = "discovery-unknown"
    report = {"status": route, "pin": public_pin(pin),
              "candidates": [c for c in found["live_sessions"] if c["session"] != pin["session"]],
              "pinned_cli_rows": [{"id": r["id"], "kind": r["kind"], "state": r["state"]}
                                  for r in found["cli"] or [] if r["session"] == pin["session"]],
              "discovery": found}
    if row:
        report["compatibility_warnings"] = compatibility_warnings(row)
        report["live"] = {"pid": row.get("pid"), "checkout": row.get("cwd"), "name": safe_name(row.get("name")),
                          "version": row.get("version"), "peer_protocol": row.get("peerProtocol")}
        report["changes"] = {
            "checkout": None if row.get("cwd") == pin.get("checkout_at_pin")
            else {"at_pin": pin.get("checkout_at_pin"), "now": row.get("cwd")},
            "name": None if safe_name(row.get("name")) == safe_name(pin.get("name_at_pin"))
            else {"at_pin": safe_name(pin.get("name_at_pin")), "now": safe_name(row.get("name"))},
        }
    return report


def tracked_and_clean(root, relative):
    """The exact path (literal pathspec, no wildcard expansion) is tracked and has no local change."""
    if git(root, "--literal-pathspecs", "ls-files", "--error-unmatch", "--", relative).returncode != 0:
        return False
    result = git(root, "--literal-pathspecs", "status", "--porcelain", "--", relative)
    return result.returncode == 0 and result.stdout == ""


def verify_authority(root, args):
    """Mechanical bindings only; whether the text authorizes creation is Codex's recorded judgement."""
    failures = []
    if not (args.authority_ref and args.authority_sha256 and args.authority_text):
        return ["authority-arguments-missing"], ""
    try:
        path, data = real_management_file(root, args.authority_ref, "authority-unsafe")
    except Blocked as error:
        return [str(error)], ""
    if path is None:
        return ["authority-absent"], ""
    if not tracked_and_clean(root, Path(args.authority_ref).as_posix()):
        failures.append("authority-untracked-or-modified")
    if digest(data) != args.authority_sha256:
        failures.append("authority-hash-mismatch")
    if args.authority_text not in data.decode("utf-8", errors="replace"):
        failures.append("authority-quote-missing")
    # Candidates must be named in the verified quoted decision itself, not elsewhere in the file.
    return failures, args.authority_text if not failures else ""


def creation_check(args, root, registry, projects):
    refusals = []

    def refuse(code, detail):
        refusals.append({"code": code, "detail": detail})

    found = discovery(root, registry)
    if found["state"] != "complete":
        refuse("discovery-unknown", found["reasons"])
    coordinators = {canonical_uuid(value) for value in args.coordinator_uuid}
    state, pin, reason = read_pin(root)
    if state == "unsafe":
        refuse("discovery-unknown", [f"pin-unsafe:{reason}"])
    elif state == "pinned":
        coordinators |= {pin["session"]} | ({pin["replaces"]} if pin.get("replaces") else set())
    observed = [row["session"] for row in found["live_sessions"]]
    observed += [r["session"] for r in found["cli"] or [] if cli_live(r) and r["same_repository"] is True]
    invalid = sorted({repr(value) for value in observed if not is_canonical_uuid(value)})
    live = {value for value in observed if is_canonical_uuid(value)}
    live_coordinators = sorted(coordinators & live)
    if live_coordinators:
        refuse("reusable-live-coordinator", live_coordinators)
    stopped = sorted(
        session for session in coordinators - live
        if any(r["session"] == session and r["kind"] == "background" and r["state"] in ("stopped", "done")
               for r in found["cli"] or [])
        or transcripts(projects, session))
    if stopped:
        refuse("reusable-stopped-coordinator", stopped)
    failures, quote = verify_authority(root, args)
    if failures:
        refuse("missing-authority", failures)
    unnamed = sorted(session for session in live - coordinators if session not in quote)
    if unnamed or invalid:
        refuse("ambiguous-candidates", unnamed + [f"invalid-session-id:{value}" for value in invalid])
    conflicts = sorted(session for session in coordinators if quote and session in quote)
    if conflicts:
        refuse("coordinator-candidate-conflict", conflicts)
    version = cli_version()
    if version not in CREATION_VERIFIED_VERSIONS:
        refuse("unverified-cli-version", version or "unavailable")
    command = None
    if not refusals:
        name = args.creation_name or ""
        prompt = Path(args.creation_prompt_file).read_text() if args.creation_prompt_file else ""
        # The installed CLI help documents no "--" terminator, so none is guessed. A prompt that starts with
        # "-" or is a single token could be parsed as an option or a subcommand, so it is refused instead.
        if not (re.fullmatch(r"[^\x00-\x1f\x7f-]{1}[^\x00-\x1f\x7f]{0,79}", name)
                and prompt.strip() and not prompt.startswith("-") and re.search(r"\s", prompt.strip())
                and len(prompt.encode()) <= 32768):
            refuse("creation-command-arguments-invalid", "--creation-name and --creation-prompt-file")
        else:
            command = ["claude", "--bg", "--name", name, "--effort", "xhigh", prompt]
    if refusals:
        return {"status": "blocked", "reason": "creation-refused", "refusals": refusals, "spawned": 0,
                "next": "Reuse, restart, select or escalate; nothing overrides a refusal."}
    main = main_checkout(root)
    return {"status": "creation-permitted", "command": command, "cwd": str(main),
            "shell": f"cd {shlex.quote(str(main))} && {shlex.join(command)}", "spawned": 0,
            "next": "The operator runs this once from cwd, re-runs status, expects exactly one new live "
                    "session, verifies it, pins it explicitly and confirms a receipt."}


def cmd_pin(args, root, registry, projects):
    session = canonical_uuid(args.session)
    selection = management_path(root, args.selection_ref)
    require(selection.is_file(), "missing-selection-ref")
    registration = live_registration(registry, session, root)
    newer = newer_sessions(registry, projects, root, session)
    require(args.confirm_target or not newer, "newer-live-session-activity-check-task-owner-and-confirm-target")
    state, pin, _ = read_pin(root)
    require(state != "unsafe", "pin-unsafe")
    if state == "pinned" and pin["session"] == session:
        return {"status": "pinned", "unchanged": True, "session": session}
    require(state == "unpinned" or args.replace, "existing-pin-requires-replace")
    record = {"schema": PIN_SCHEMA, "repository": str(repo_identity(root)), "session": session,
              "checkout_at_pin": registration.get("cwd"), "name_at_pin": registration.get("name"),
              "cli_version_at_pin": registration.get("version"),
              "selection_ref": {"path": args.selection_ref, "sha256": digest(selection.read_bytes())},
              "pinned_at": utc_now().isoformat(), "replaces": pin["session"] if state == "pinned" else None}
    write_private_atomic(repo_state(root), "pin.json", (json.dumps(record, indent=1) + "\n").encode())
    return {"status": "pinned", "session": session, "replaces": record["replaces"],
            "selection_ref": record["selection_ref"],
            "confirmed_newer_peers": [p.get("session") for p in newer] if newer else []}


def cmd_unpin(args, root):
    session = canonical_uuid(args.session)
    state, pin, _ = read_pin(root)
    require(state != "unsafe", "pin-unsafe")
    require(state == "pinned", "not-pinned")
    require(pin["session"] == session, "pin-session-mismatch")
    directory = repo_state(root, create=False)
    os.unlink(directory / "pin.json")
    fsync_dir(directory)
    return {"status": "unpinned", "removed": session}


# --- task-closeout compaction ----------------------------------------------------------------------

def is_main_checkout(root):
    return Path(root).resolve() == main_checkout(root)


def verify_closure(root, args):
    """Mechanical checks of a supervisor-verified terminal record; they bind, not interpret, closure."""
    relative = Path(args.closure_ref).as_posix()
    # No symlink component, so a committed link cannot launder a Goal or an untracked file past the checks.
    path, data = real_management_file(root, relative, "closure-unsafe")
    require(not {"goals", "specs", "plans"} & set(Path(relative).parts[1:-1]), "closure-not-terminal-record")
    if args.disposition == "completed":
        require(re.fullmatch(rf"management/[^/]+/results/{args.task}-implementation-codex-review[^/]*\.md",
                             relative), "closure-not-implementation-review")
    require(path is not None, "closure-missing")
    require(tracked_and_clean(root, relative), "closure-untracked-or-modified")
    require(digest(data) == args.closure_sha256, "closure-hash-mismatch")
    text = args.closure_text
    require(text and text in data.decode("utf-8", errors="replace"), "closure-quote-missing")
    require(args.task in text, "closure-quote-lacks-task")
    token = {"completed": "ACCEPT" in text, "canceled": re.search("cancel", text, re.I) is not None,
             "superseded": re.search("supersed", text, re.I) is not None}[args.disposition]
    require(token, "closure-disposition-mismatch")
    return {"path": relative, "sha256": args.closure_sha256, "text": text, "disposition": args.disposition,
            "task": args.task, "verified_by": args.verified_by}


def read_log(path):
    try:
        fd = os.open(path, os.O_RDONLY | os.O_NOFOLLOW)
    except FileNotFoundError:
        raise Blocked("missing-delivery-log")
    except OSError as error:
        require(error.errno != errno.ELOOP, "record-unsafe")
        raise
    with os.fdopen(fd, "rb") as stream:
        info = os.fstat(stream.fileno())
        require(stat.S_ISREG(info.st_mode) and info.st_uid == os.getuid(), "record-unsafe")
        return stream.read(), (info.st_dev, info.st_ino)


def classify_log(data, task, listed):
    """Group by handoff ID; only whole groups attributed to the verified-terminal task move."""
    lines = data.splitlines(keepends=True)
    groups = {}
    malformed = []
    for index, line in enumerate(lines):
        record = None
        if line.endswith(b"\n"):
            try:
                value = json.loads(line)
                if isinstance(value, dict):
                    canonical_uuid(value.get("handoff_id"))
                    record = value
            except (ValueError, Blocked):
                record = None
        if record is None:
            malformed.append(index)
        else:
            groups.setdefault(record["handoff_id"], []).append((index, record))
    selected, selected_ids, incomplete = set(), [], []
    retained = Counter()
    for handoff_id, items in groups.items():
        records = [record for _, record in items]
        labels = {record["task"] for record in records if "task" in record}
        if any(not isinstance(r.get("session"), str) or not isinstance(r.get("status"), str) for r in records):
            reason = "malformed-group"
        elif len({record["session"] for record in records}) != 1:
            reason = "multi-session"
        elif labels - {task}:
            reason = "conflicting-label" if handoff_id in listed or task in labels else "other-task"
        elif labels == {task} or handoff_id in listed:
            reason = None
        else:
            reason = "unattributed"
        if reason:
            retained[reason] += len(items)
            continue
        selected.update(index for index, _ in items)
        selected_ids.append(handoff_id)
        if not any(record["status"] in DELIVERY_OUTCOMES for record in records):
            incomplete.append(handoff_id)
    if malformed:
        retained["malformed-or-truncated"] += len(malformed)
    return {"archive": b"".join(line for i, line in enumerate(lines) if i in selected),
            "retained": b"".join(line for i, line in enumerate(lines) if i not in selected),
            "selected_ids": selected_ids, "incomplete_delivery": incomplete,
            "archived_records": len(selected), "retained_records": len(lines) - len(selected),
            "retained_by_reason": dict(sorted(retained.items()))}


def selection_digest(closure, selection):
    payload = {"closure": closure, "selected_ids": sorted(selection["selected_ids"]),
               "archive_sha256": digest(selection["archive"])}
    return digest(json.dumps(payload, sort_keys=True, separators=(",", ":")).encode())


def helper_dirs(root):
    """The skill scripts directories a process could run: every worktree's copy (current and retired layouts), this
    copy, the user copies."""
    directories = {Path(__file__).resolve().parent}
    relatives = HELPER_CANDIDATES + (LEGACY_HELPER_RELATIVE,)
    helpers = [path / relative for path in worktree_paths(root) for relative in relatives]
    helpers.append(Path.home() / ".agents/skills/restruct-sdlc/claude-session-handoff/scripts/handoff.py")
    helpers.append(Path.home() / LEGACY_HELPER_RELATIVE)
    for helper in helpers:
        with contextlib.suppress(OSError, RuntimeError):
            if helper.parent.is_dir():
                directories.add(helper.parent.resolve())
    return directories


def is_python(arguments):
    return bool(arguments) and re.fullmatch(r"python[0-9.]*", os.path.basename(arguments[0])) is not None


def helper_argument(arguments, cwd, directories):
    """Index of the argument naming a helper: any handoff.py, or for a Python process any path that
    resolves (relative to its cwd, through symlinks) into a skill scripts directory."""
    for index, argument in enumerate(arguments):
        if os.path.basename(argument) == "handoff.py":
            return index
    if not is_python(arguments):
        return None
    for index, argument in enumerate(arguments[1:], 1):
        if argument.startswith("-") or (not argument.startswith("/") and cwd is None):
            continue
        try:
            resolved = Path(os.path.realpath(argument if argument.startswith("/") else cwd / argument))
        except (OSError, ValueError):
            continue
        if resolved.parent in directories:
            return index
    return None


def probe(entry, action):
    """(value, None); (None, "exited") when the /proc entry vanished; (None, "unreadable") otherwise."""
    try:
        return action(), None
    except OSError as error:
        if isinstance(error, (FileNotFoundError, ProcessLookupError)) and not entry.exists():
            return None, "exited"
        return None, "unreadable"


def process_scan(log_id, directories):
    """Scan same-user processes (by status real Uid), excluding this one, failing closed inside the writer model.

    Only the Python helper's `send` appends to the delivery log; Claude sessions use native messaging. So a
    process that runs a helper, holds the live log open, or cannot be inspected where a helper could hide refuses.
    A non-Python process that names no helper but whose descriptors cannot be listed (for example the session's
    `(sd-pam)`) is evidence only; a non-helper descriptor holder is left to the documented quiescence procedure.
    Returns (invocations, holders, uninspectable, nonhelper); evidence carries PID and comm, never full argv.
    """
    invocations, holders, uninspectable, nonhelper = [], [], [], []
    for entry in PROC.iterdir():
        if not entry.name.isdigit() or int(entry.name) == os.getpid():
            continue
        pid = int(entry.name)
        status, problem = probe(entry, lambda: (entry / "status").read_text())
        if problem == "exited":
            continue
        fields = dict(line.split(":", 1) for line in (status or "").splitlines() if ":" in line)
        try:
            uid = int(fields["Uid"].split()[0])
        except (KeyError, ValueError, IndexError):
            uninspectable.append({"pid": pid, "reason": "status-unreadable" if problem else "status-malformed"})
            continue
        if uid != os.getuid() or fields.get("State", "").strip().startswith("Z"):
            continue
        comm = fields.get("Name", "").strip()[:32]
        raw, problem = probe(entry, lambda: (entry / "cmdline").read_bytes())
        if problem == "exited":
            continue
        arguments = [a.decode("utf-8", "replace") for a in (raw or b"").split(b"\0") if a]
        if problem or not arguments:
            uninspectable.append({"pid": pid, "comm": comm,
                                  "reason": "cmdline-unreadable" if problem else "cmdline-empty"})
            continue
        link, problem = probe(entry, lambda: os.readlink(entry / "cwd"))
        if problem == "exited":
            continue
        cwd = Path(link) if link else None
        python = is_python(arguments)
        if python and cwd is None and any(not a.startswith(("-", "/")) for a in arguments[1:]):
            uninspectable.append({"pid": pid, "comm": comm, "reason": "cwd-unreadable"})
        index = helper_argument(arguments, cwd, directories)
        if index is not None:
            command = next((a for a in arguments[index + 1:] if a in COMMANDS), None)
            script = Path(arguments[index])
            if not script.is_absolute() and cwd is not None:
                script = cwd / script
            try:
                path_hash = digest(script.read_bytes()) if script.is_absolute() else None
            except OSError:
                path_hash = None
            # A running interpreter may have loaded older code before its source pathname was
            # replaced, so the current path hash never classifies it as upgraded.
            invocations.append({"pid": pid, "command": command, "source_path_sha256_untrusted": path_hash})
        writer_capable = python or index is not None
        descriptors, problem = probe(entry, lambda: list((entry / "fd").iterdir()))
        if problem == "exited":
            continue
        if problem:
            if writer_capable:
                uninspectable.append({"pid": pid, "comm": comm, "reason": "fd-unlistable"})
            else:
                nonhelper.append({"pid": pid, "comm": comm})
            continue
        for descriptor in descriptors:
            try:
                info = os.stat(descriptor)
            except FileNotFoundError:
                continue
            except OSError:
                if writer_capable:
                    uninspectable.append({"pid": pid, "comm": comm, "reason": "fd-unstatable"})
                    break
                continue
            if (info.st_dev, info.st_ino) == log_id:
                holders.append({"pid": pid, "runs_helper": index is not None})
                break
    return invocations, holders, uninspectable, nonhelper


def helper_relative(root):
    """The first HELPER_CANDIDATES path that HEAD tracks as a file (blob); None when neither is tracked."""
    for candidate in HELPER_CANDIDATES:
        result = git(root, "cat-file", "-t", f"HEAD:{candidate}")
        if result.returncode == 0 and result.stdout.strip() == "blob":
            return candidate
    return None


def committed_helper_sha(root, relative):
    if relative is None:
        return None
    result = git(root, "show", f"HEAD:{relative}", text=False)
    return digest(result.stdout) if result.returncode == 0 else None


def prerequisite_evidence(root, log_id, writer_sha256):
    failures = []
    main = is_main_checkout(root)
    if not main:
        failures.append("repo-not-main-checkout")
    relative = helper_relative(root)
    committed = committed_helper_sha(root, relative)
    try:
        working = digest((Path(root) / relative).read_bytes()) if relative is not None else None
    except OSError:
        working = None
    if not (SELF_SHA256 == committed == working == writer_sha256):
        failures.append("invocation-path-not-upgraded")
    invocations, holders, uninspectable, nonhelper = process_scan(log_id, helper_dirs(root))
    if invocations:
        failures.append("other-helper-invocation-in-flight")
    if holders:
        failures.append("log-held-by-other-process")
    if uninspectable:
        failures.append("uninspectable-process")
    listing = cli_rows(main_checkout(root))
    pids = [row["pid"] for row in listing or [] if cli_live(row)]
    invisible = [pid for pid in pids if not pid_visible(pid)]
    if listing is None or not pids or invisible:
        failures.append("pid-visibility-unproven")
    differing = []
    for path in worktree_paths(root) if relative is not None else ():
        try:
            value = digest((path / relative).read_bytes())
        except OSError:
            continue
        if value != writer_sha256:
            differing.append({"worktree": str(path), "helper_sha256": value})
    return {"satisfied": not failures, "failures": failures, "running_sha256": SELF_SHA256,
            "committed_sha256": committed, "working_sha256": working, "writer_sha256": writer_sha256,
            "main_checkout": main, "helper_invocations": invocations, "log_holders": holders,
            "uninspectable_processes": uninspectable, "uninspectable_nonhelper": nonhelper,
            "cli_listing": listing is not None, "visible_cli_pids": len(pids) - len(invisible),
            "invisible_cli_pids": len(invisible),
            "other_worktree_helpers_differing": differing}


def write_archive(root, closure, selection, source, evidence, selection_sha):
    directory = state_subdir(root, "archive")
    stamp = utc_now().strftime("%Y%m%dT%H%M%S%fZ")
    name = f"{stamp}-{closure['task']}.jsonl"
    try:
        write_new_private(directory / name, selection["archive"])
        fsync_dir(directory)
        require(read_private(directory / name) == selection["archive"], "archive-verification-failed")
    except OSError:
        raise Blocked("archive-write-failed")
    manifest = {
        "schema": MANIFEST_SCHEMA, "created_at": utc_now().isoformat(), "task": closure["task"],
        "disposition": closure["disposition"], "verified_by": closure["verified_by"],
        "closure": {k: closure[k] for k in ("path", "sha256", "text")},
        "source": {"path": source["path"], "sha256": digest(source["data"]), "bytes": len(source["data"])},
        "archive": {"name": name, "sha256": digest(selection["archive"])},
        "retained_sha256": digest(selection["retained"]),
        "counts": {"archived_records": selection["archived_records"],
                   "archived_groups": len(selection["selected_ids"]),
                   "retained_records": selection["retained_records"],
                   "retained_by_reason": selection["retained_by_reason"]},
        "handoff_ids": selection["selected_ids"], "incomplete_delivery": selection["incomplete_delivery"],
        "writer_sha256": evidence["writer_sha256"], "selection_sha256": selection_sha,
        "prerequisite": {k: evidence[k] for k in ("satisfied", "failures", "main_checkout", "visible_cli_pids",
                                                  "uninspectable_nonhelper")},
        "pruned": False, "prune_status": "pending",
    }
    manifest_name = f"{stamp}-{closure['task']}.manifest.json"
    data = (json.dumps(manifest, indent=1, sort_keys=True) + "\n").encode()
    try:
        write_new_private(directory / manifest_name, data)
        fsync_dir(directory)
    except OSError:
        raise Blocked("archive-write-failed")
    return name, manifest_name, manifest


def finalize_manifest(root, manifest_name, manifest, **updates):
    """Atomically record the actual outcome (pruned or why not) beside the verified archive."""
    data = (json.dumps({**manifest, **updates}, indent=1, sort_keys=True) + "\n").encode()
    write_private_atomic(state_subdir(root, "archive"), manifest_name, data)
    return digest(data)


def replace_file(source, target):
    os.replace(source, target)


def replace_log(log, retained):
    """Atomic same-directory replacement; the temporary name stays covered by the log ignore rule."""
    staging = log.parent / f".compaction-{uuid.uuid4().hex}"
    os.mkdir(staging, 0o700)
    temporary = staging / log.name
    try:
        try:
            write_new_private(temporary, retained)
            replace_file(temporary, log)
        except OSError:
            raise Blocked("replace-failed")
        try:
            fsync_dir(log.parent)
        except OSError:
            raise Blocked("pruned-directory-fsync-failed", {"pruned": True})
    finally:
        # Cleanup never masks the outcome; a leftover staging name stays covered by the log ignore rule.
        with contextlib.suppress(OSError):
            os.unlink(temporary)
        with contextlib.suppress(OSError):
            os.rmdir(staging)
    try:
        current = read_log(log)[0]
    except (Blocked, OSError):
        current = None
    if current != retained:
        raise Blocked("pruned-verification-failed", {"pruned": True})


def cmd_compact(args, root):
    require(TASK_RE.fullmatch(args.task), "invalid-task")
    require(HEX64_RE.fullmatch(args.closure_sha256) and HEX64_RE.fullmatch(args.writer_sha256), "invalid-sha256")
    listed = {canonical_uuid(value) for value in args.handoff_id}
    require(is_main_checkout(root), "compact-requires-main-checkout")
    closure = verify_closure(root, args)
    log = record_path(root, args.record)

    def summary(selection, data):
        return {"task": args.task, "disposition": args.disposition, "verified_by": args.verified_by,
                "closure": {k: closure[k] for k in ("path", "sha256")},
                "selection_sha256": selection_digest(closure, selection),
                "archived_records": selection["archived_records"],
                "archived_groups": len(selection["selected_ids"]), "handoff_ids": selection["selected_ids"],
                "incomplete_delivery": selection["incomplete_delivery"],
                "retained_records": selection["retained_records"],
                "retained_by_reason": selection["retained_by_reason"], "source_sha256": digest(data)}

    data, log_id = read_log(log)
    selection = classify_log(data, args.task, listed)
    if not args.apply:
        return {"status": "compaction-dry-run", **summary(selection, data),
                "prerequisite": prerequisite_evidence(root, log_id, args.writer_sha256), "writes": 0}
    require(args.selection_sha256 == selection_digest(closure, selection), "stale-selection")
    with maintenance_lock(root, log, True, args.lock_wait_seconds):
        data, log_id = read_log(log)
        selection = classify_log(data, args.task, listed)
        selection_sha = selection_digest(closure, selection)
        require(selection_sha == args.selection_sha256, "stale-selection")
        evidence = prerequisite_evidence(root, log_id, args.writer_sha256)
        if not evidence["satisfied"]:
            raise Blocked("legacy-writer-prerequisite-unmet", {"prerequisite": evidence})
        if not selection["selected_ids"]:
            return {"status": "compaction-applied", **summary(selection, data), "archive": None,
                    "prerequisite": {k: evidence[k] for k in ("satisfied", "failures", "uninspectable_nonhelper")}}
        name, manifest_name, manifest = write_archive(root, closure, selection,
                                                      {"path": args.record, "data": data}, evidence, selection_sha)

        def refuse(reason, pruned=False, **details):
            try:
                finalize_manifest(root, manifest_name, manifest, pruned=pruned, prune_status=reason, **details)
            except (OSError, Blocked):
                pass  # The verified archive and the reported reason remain; the manifest says "pending".
            raise Blocked(reason, {"archive": name, "pruned": pruned, **details})

        final = prerequisite_evidence(root, log_id, args.writer_sha256)
        if not final["satisfied"]:
            refuse("legacy-writer-prerequisite-unmet", final_prerequisite={
                k: final[k] for k in ("satisfied", "failures", "helper_invocations", "log_holders",
                                      "uninspectable_processes", "uninspectable_nonhelper")})
        if read_log(log)[0] != data:
            refuse("log-changed-during-compaction")
        try:
            replace_log(log, selection["retained"])
        except Blocked as error:
            refuse(str(error), pruned=error.details.get("pruned", False))
        try:
            manifest_sha = finalize_manifest(root, manifest_name, manifest, pruned=True, prune_status="pruned",
                                             live_sha256=digest(selection["retained"]))
        except (OSError, Blocked):
            raise Blocked("pruned-manifest-update-failed", {"archive": name, "pruned": True})
    return {"status": "compaction-applied", **summary(selection, data), "archive": name,
            "manifest_sha256": manifest_sha, "live_sha256": digest(selection["retained"]),
            "prerequisite": {k: final[k] for k in ("satisfied", "failures", "uninspectable_nonhelper")}}


# --- command line ------------------------------------------------------------------------------------

def parser():
    root_parser = argparse.ArgumentParser(description=__doc__)
    root_parser.add_argument("--repo", default=".", help="Repository or service directory")
    sub = root_parser.add_subparsers(dest="command", required=True)
    sub.add_parser("discover", help="Registered sessions plus combined, classified discovery")
    check = sub.add_parser("check", help="Verified live session and bounded assistant text only")
    check.add_argument("--session", help="Explicit target; omit to use only a verified live pin")
    check.add_argument("--child")
    check.add_argument("--match", default="")
    check.add_argument("--limit", type=int, default=3, choices=range(1, 6))
    sender = sub.add_parser("send", help="Send an already-authorized handoff, once")
    sender.add_argument("--session", help="Explicit target; omit to use only a verified live pin")
    sender.add_argument("--message-file", required=True)
    sender.add_argument("--document", action="append", default=[], help="Root-relative management reference")
    sender.add_argument("--record", default=DEFAULT_RECORD)
    sender.add_argument("--task", help="Optional task label GNNN recorded with each delivery record")
    sender.add_argument("--wait-seconds", type=int, default=20, choices=range(0, 61))
    sender.add_argument("--lock-wait-seconds", type=int, default=10, choices=range(0, 61))
    sender.add_argument("--confirm-target", action="store_true",
                        help="Use only after verifying this task belongs to the selected session despite newer peers")
    receipt = sub.add_parser("receipt", help="Read-only verification of transcript replay, not execution")
    receipt.add_argument("--session", help="Needed only when the log record is absent or ambiguous")
    receipt.add_argument("--handoff-id", required=True)
    receipt.add_argument("--record", default=DEFAULT_RECORD)
    pin = sub.add_parser("pin", help="Record the verified preferred coordinator (never sends or spawns)")
    pin.add_argument("--session", required=True)
    pin.add_argument("--selection-ref", required=True, help="Management record of the owner's selection")
    pin.add_argument("--replace", action="store_true")
    pin.add_argument("--confirm-target", action="store_true")
    status = sub.add_parser("status", help="Pin state and duplicate-safe discovery (read-only)")
    status.add_argument("--creation-check", action="store_true", help="Read-only creation gate")
    status.add_argument("--coordinator-uuid", action="append", default=[])
    status.add_argument("--authority-ref")
    status.add_argument("--authority-sha256")
    status.add_argument("--authority-text")
    status.add_argument("--creation-name")
    status.add_argument("--creation-prompt-file")
    unpin = sub.add_parser("unpin", help="Remove the pin only when the UUID matches")
    unpin.add_argument("--session", required=True)
    compact = sub.add_parser("compact", help="Archive-before-prune task closeout (dry-run by default)")
    compact.add_argument("--task", required=True)
    compact.add_argument("--disposition", required=True, choices=("completed", "canceled", "superseded"))
    compact.add_argument("--closure-ref", required=True)
    compact.add_argument("--closure-sha256", required=True)
    compact.add_argument("--closure-text", required=True)
    compact.add_argument("--verified-by", required=True, choices=("codex", "owner"))
    compact.add_argument("--writer-sha256", required=True)
    compact.add_argument("--handoff-id", action="append", default=[])
    compact.add_argument("--record", default=DEFAULT_RECORD)
    compact.add_argument("--apply", action="store_true")
    compact.add_argument("--selection-sha256")
    compact.add_argument("--lock-wait-seconds", type=int, default=10, choices=range(0, 61))
    return root_parser


def main():
    args = parser().parse_args()
    if args.command == "compact" and args.apply and not args.selection_sha256:
        raise Blocked("apply-requires-selection-sha256")
    root = repo_root(Path(args.repo))
    registry = Path.home() / ".claude" / "sessions"
    projects = Path.home() / ".claude" / "projects"
    if args.command == "discover":
        identity = repo_identity(root)
        rows = []
        for path in registry.glob("*.json"):
            try:
                row = json.loads(path.read_text())
                if same_repo(row.get("cwd"), identity) is True:
                    rows.append({k: row.get(k) for k in ("sessionId", "pid", "cwd", "status", "version")})
            except (OSError, ValueError, AttributeError):
                continue
        return {"status": "registered-not-verified", "sessions": rows, "discovery": discovery(root, registry)}
    if args.command == "status":
        return creation_check(args, root, registry, projects) if args.creation_check \
            else status_report(root, registry, projects)
    if args.command == "pin":
        return cmd_pin(args, root, registry, projects)
    if args.command == "unpin":
        return cmd_unpin(args, root)
    if args.command == "compact":
        return cmd_compact(args, root)
    if args.command == "receipt":
        uuid.UUID(args.handoff_id)
        logged = handoff_sessions(root, args.record, args.handoff_id)
        if args.session:
            uuid.UUID(args.session)
            require(not logged or args.session in logged, "receipt-session-mismatch")
            session = args.session
        else:
            require(len(logged) == 1, "receipt-session-unresolved-pass-session")
            session = next(iter(logged))
        return receipt_status(projects, session, args.handoff_id)
    session, route = resolve_target(args.session, root, registry)
    registration = live_registration(registry, session, root)
    if args.command == "check":
        excerpts = []
        for path in transcripts(projects, session, args.child):
            for row in tail_records(path):
                if row.get("type") == "assistant":
                    for text in texts(row):
                        if args.match.lower() in text.lower():
                            excerpts.append({"timestamp": row.get("timestamp", ""),
                                             "text": safe_excerpt(text, 800)})
        excerpts.sort(key=lambda row: row["timestamp"])
        return {"status": registration.get("status"), "session": session, "route": route,
                "pid": registration["pid"], "version": registration["version"],
                "peer_protocol": registration["peerProtocol"],
                "compatibility_warnings": compatibility_warnings(registration),
                "checkout": registration["cwd"], "reported_not_validated": excerpts[-args.limit:],
                "newer_live_sessions_check_target": mark_peers(newer_sessions(registry, projects, root, session), root)}
    newer = newer_sessions(registry, projects, root, session)
    require(args.confirm_target or not newer, "newer-live-session-activity-check-task-owner-and-confirm-target")
    require(args.task is None or TASK_RE.fullmatch(args.task), "invalid-task-label")
    message = Path(args.message_file).read_text()
    require(message.strip() and len(message.encode()) <= 32768, "empty-or-oversized-handoff")
    references = []
    for name in args.document:
        path = management_path(root, name)
        require(path.is_file(), "missing-management-document")
        references.append({"path": name, "sha256": digest(path.read_bytes())})
    handoff_id = str(uuid.uuid4())
    if references:
        message += "\nRepository management references (read these; unchanged approvals remain binding):\n"
        message += "\n".join(f"{root / ref['path']} SHA-256 {ref['sha256']}" for ref in references)
    message += f"\n[handoff-id: {handoff_id}]"
    record = {"time": utc_now().isoformat(), "handoff_id": handoff_id,
              "session": session, "checkout": registration["cwd"],
              "message_sha256": digest(message.encode()), "documents": references,
              "route": route, "status": "prepared-not-sent"}
    if args.task:
        record["task"] = args.task
    if args.confirm_target and newer:
        record["confirmed_newer_peers"] = [peer.get("session") for peer in newer]
    # Shared for the whole invocation, so a handoff's records are never split by compaction.
    with maintenance_lock(root, record_path(root, args.record), False, args.lock_wait_seconds):
        append_record(root, args.record, record)  # Fail before delivery if tracking cannot be written.
        delivered_to = send(registry, registration, message, handoff_id)
        record = {**record, "status": "sent-receipt-pending",
                  "compatibility_warnings": compatibility_warnings(delivered_to or registration)}
        append_record(root, args.record, record)
        deadline = time.monotonic() + args.wait_seconds
        result = receipt_status(projects, session, handoff_id)
        while result["status"] != "receipt-confirmed" and time.monotonic() < deadline:
            time.sleep(min(0.5, max(0, deadline - time.monotonic())))
            result = receipt_status(projects, session, handoff_id)
        append_record(root, args.record, {**record, "status": result["status"]})
    return {**result, "record": args.record, "route": route,
            "compatibility_warnings": record.get("compatibility_warnings", []),
            "newer_live_sessions_check_target": mark_peers(newer_sessions(registry, projects, root, session), root),
            "next": "Verify model acknowledgment and task dispatch." if result["status"] == "receipt-confirmed"
            else "Do not claim delivery or retry automatically. Re-discover the intended session; inspect approval/refusal/queue state."}


if __name__ == "__main__":
    try:
        result = main()
        print(json.dumps(result, separators=(",", ":")))
        if result.get("status") == "blocked":
            sys.exit(1)
        if result.get("status") == "delivery-unconfirmed":
            sys.exit(2)
    except Blocked as error:
        print(json.dumps({"status": "blocked", "reason": str(error), **error.details}, separators=(",", ":")))
        sys.exit(1)
    except (OSError, ValueError, KeyError, TypeError, subprocess.SubprocessError):
        print(json.dumps({"status": "blocked", "reason": "local-input-process-or-transport-error"}))
        sys.exit(1)
