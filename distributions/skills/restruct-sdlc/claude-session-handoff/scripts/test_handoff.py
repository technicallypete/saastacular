#!/usr/bin/env python3
"""Focused receipt/routing regression checks; no live messages or credentials."""
import json
from pathlib import Path
import tempfile
import unittest
from unittest.mock import patch
import handoff

SESSION = "00000000-0000-4000-8000-0000000000a1"
OTHER = "00000000-0000-4000-8000-0000000000a2"
HANDOFF = "00000000-0000-4000-8000-0000000000a3"

class ReceiptRouting(unittest.TestCase):
    def test_receipt_requires_selected_session_user_replay(self):
        with tempfile.TemporaryDirectory() as d:
            projects = Path(d); p = projects / "project"; p.mkdir()
            marker = f"[handoff-id: {HANDOFF}]"
            def record(kind):
                return json.dumps({"type": kind, "message": {"content": marker}}) + "\n"
            (p / f"{OTHER}.jsonl").write_text(record("user"))
            (p / f"{SESSION}.jsonl").write_text(record("assistant"))
            self.assertEqual(handoff.receipt_status(projects, SESSION, HANDOFF)["status"], "delivery-unconfirmed")
            (p / f"{SESSION}.jsonl").write_text(record("user"))
            result = handoff.receipt_status(projects, SESSION, HANDOFF)
            self.assertEqual(result["status"], "receipt-confirmed")
            self.assertFalse(result["execution_confirmed"])

    def test_missing_receipt_is_explicitly_unconfirmed(self):
        with tempfile.TemporaryDirectory() as d:
            self.assertEqual(handoff.receipt_status(Path(d), SESSION, HANDOFF)["status"], "delivery-unconfirmed")

    def test_newer_peer_blocks_before_tracking_or_delivery(self):
        with patch("sys.argv", ["handoff", "send", "--session", SESSION, "--message-file", "/missing"]), \
             patch.object(handoff, "repo_root", return_value=Path("/repo")), \
             patch.object(handoff, "live_registration", return_value={}), \
             patch.object(handoff, "newer_sessions", return_value=[{"session": OTHER}]), \
             patch.object(handoff, "append_record") as track, patch.object(handoff, "send") as send:
            with self.assertRaises(handoff.Blocked):
                handoff.main()
            track.assert_not_called()
            send.assert_not_called()

    def test_no_receipt_does_not_retry_or_report_delivery(self):
        with tempfile.TemporaryDirectory() as d:
            root = Path(d); message = root / "message"; message.write_text("Authorized test fixture")
            argv = ["handoff", "send", "--session", SESSION, "--message-file", str(message), "--wait-seconds", "0"]
            with patch("sys.argv", argv), patch.object(handoff, "repo_root", return_value=root), \
                 patch.object(handoff, "live_registration", return_value={"cwd": str(root)}), \
                 patch.object(handoff, "newer_sessions", return_value=[]), patch.object(handoff, "send") as send, \
                 patch.object(handoff, "receipt_status", return_value={"status": "delivery-unconfirmed", "execution_confirmed": False}):
                result = handoff.main()
                self.assertEqual(result["status"], "delivery-unconfirmed")
                send.assert_called_once()
                records = [json.loads(line) for line in (root / "management/SESSION-HANDOFFS.jsonl").read_text().splitlines()]
                self.assertEqual(records[-1]["status"], "delivery-unconfirmed")

# --- G052 Plan r3 fixtures (T1-T15): temporary HOME, state, Git repositories with worktrees, fake registry,
# transcripts and /proc, and stubbed transport/CLI. No live messages, spawns or credentials.
import contextlib
import hashlib
import io
import os
import shutil
import stat
import subprocess
import sys
import threading
import time
import uuid

HELPER_BYTES = Path(handoff.__file__).read_bytes()
SELF_SHA = hashlib.sha256(HELPER_BYTES).hexdigest()
REAL_RUN_CLAUDE = handoff.run_claude
REVIEW = "management/v0/results/G006-implementation-codex-review.md"
QUOTE = "**ACCEPT** — G006 implementation accepted after the staging pipeline succeeded"
LOG = "management/SESSION-HANDOFFS.jsonl"
AUTHORITY = "management/v0/results/G999-coordinator-creation-authority.md"
DECISION = "Owner authorizes creating one new coordinator for this repository"
VISIBLE_PID = 910000
_STATE = {}


def _fixture_identity(directory):
    """The four legacy tests use a plain temporary root without Git; only such roots map to themselves."""
    try:
        return _STATE["identity"](directory)
    except handoff.Blocked:
        path = Path(directory)
        if (path.is_dir() and not (path / ".git").exists()
                and path.resolve().is_relative_to(Path(tempfile.gettempdir()).resolve())):
            return path.resolve()
        raise


def setUpModule():
    home = Path(tempfile.mkdtemp(prefix="handoff-home-"))
    keys = ("HOME", "GIT_CONFIG_GLOBAL", "GIT_CONFIG_NOSYSTEM")
    _STATE.update(home=home, env={key: os.environ.get(key) for key in keys}, identity=handoff.repo_identity)
    os.environ.update(HOME=str(home), GIT_CONFIG_GLOBAL=os.devnull, GIT_CONFIG_NOSYSTEM="1")
    _STATE["patch"] = patch.object(handoff, "repo_identity", _fixture_identity)
    _STATE["patch"].start()


def tearDownModule():
    _STATE["patch"].stop()
    for key, value in _STATE["env"].items():
        if value is None:
            os.environ.pop(key, None)
        else:
            os.environ[key] = value
    shutil.rmtree(_STATE["home"], ignore_errors=True)


def run_git(*arguments):
    subprocess.run(["git", *map(str, arguments)], check=True, capture_output=True)


def run_main(*argv):
    with patch("sys.argv", ["handoff", *map(str, argv)]):
        return handoff.main()


def proc_start(pid):
    return Path(f"/proc/{pid}/stat").read_text().rsplit(")", 1)[1].split()[19]


def dead_pid():
    child = subprocess.Popen([sys.executable, "-c", ""])
    child.wait()
    return child.pid


def record(handoff_id, session, status, task=None):
    value = {"time": "2026-10-10T00:00:00+00:00", "handoff_id": handoff_id, "session": session,
             "checkout": "/fixture", "message_sha256": "0" * 64, "documents": [], "status": status}
    if task:
        value["task"] = task
    return (json.dumps(value, separators=(",", ":")) + "\n").encode()


def ids_in(data):
    found = set()
    for line in data.splitlines():
        with contextlib.suppress(ValueError, AttributeError):
            found.add(json.loads(line).get("handoff_id"))
    return found - {None}


class CliStub:
    """Stands in for run_claude: read-only listing/version answers, and a record of every request."""

    def __init__(self):
        self.rows, self.fail, self.version, self.calls = [], False, "2.1.296 (Claude Code)", []

    def __call__(self, arguments, timeout=None):
        self.calls.append(list(arguments))
        if arguments == ["--version"]:
            return subprocess.CompletedProcess(arguments, 0, self.version + "\n", "")
        if self.fail:
            return subprocess.CompletedProcess(arguments, 1, "", "unavailable")
        return subprocess.CompletedProcess(arguments, 0, json.dumps(self.rows), "")

    def spawns(self):
        return [call for call in self.calls if "--bg" in call or "--resume" in call]


class Base(unittest.TestCase):
    def setUp(self):
        self.home = _STATE["home"]
        for name in (".claude", ".local"):
            shutil.rmtree(self.home / name, ignore_errors=True)
        self.registry = self.home / ".claude" / "sessions"
        self.registry.mkdir(parents=True)
        self.projects = self.home / ".claude" / "projects" / "fixture"
        self.projects.mkdir(parents=True)
        self.tmp = Path(tempfile.mkdtemp(prefix="case-", dir=self.home))
        self.addCleanup(shutil.rmtree, self.tmp, True)
        self.cli = CliStub()
        self.patch(handoff, "run_claude", self.cli)
        self.deliveries = []
        self.message = self.tmp / "message.txt"
        self.message.write_text("Authorized fixture handoff")

    def patch(self, target, name, value):
        patcher = patch.object(target, name, value)
        patcher.start()
        self.addCleanup(patcher.stop)

    def tearDown(self):
        self.assertEqual(self.cli.spawns(), [], "the helper must never start a Claude session")

    # repositories ----------------------------------------------------------------------------------
    def make_repo(self, name="repo", helper=None, extra=None):
        root = self.tmp / name
        run_git("init", "-q", "-b", "main", root)
        files = {".gitignore": b"/management/**/SESSION-HANDOFFS.jsonl\n",
                 handoff.HELPER_CANDIDATES[0]: HELPER_BYTES if helper is None else helper,
                 "management/v0/PROGRESS.md": b"# Progress\n\nOwner selects the existing coordinator.\n",
                 REVIEW: f"# G006 review\n\nOutcome: {QUOTE}.\n".encode(), **(extra or {})}
        for relative, data in files.items():
            path = root / relative
            path.parent.mkdir(parents=True, exist_ok=True)
            path.write_bytes(data)
        self.commit(root)
        return root.resolve()

    def commit(self, root):
        run_git("-C", root, "add", "-A")
        run_git("-C", root, "-c", "user.name=fixture", "-c", "user.email=fixture@example.invalid",
                "-c", "commit.gpgsign=false", "commit", "-q", "-m", "fixture")

    def worktree(self, root, name):
        path = self.tmp / name
        run_git("-C", root, "worktree", "add", "-q", "-b", name, path)
        return path.resolve()

    def status(self, root):
        return subprocess.run(["git", "-C", str(root), "status", "--porcelain", "--ignored"],
                              capture_output=True, text=True, check=True).stdout.splitlines()

    # sessions --------------------------------------------------------------------------------------
    def register(self, session, cwd, pid=None, start=None, protocol=1, version="2.1.296", name="worker"):
        pid = pid or os.getpid()
        row = {"pid": pid, "sessionId": session, "cwd": str(cwd), "procStart": start or proc_start(pid),
               "peerProtocol": protocol, "version": version, "name": name, "status": "busy",
               "messagingSocketPath": "/nonexistent/fixture.sock"}
        path = self.registry / f"{pid}-{uuid.uuid4().hex[:8]}.json"
        path.write_text(json.dumps(row))
        return path

    def unregister(self, session):
        for path in self.registry.glob("*.json"):
            if json.loads(path.read_text()).get("sessionId") == session:
                path.unlink()

    def transcript(self, session, mtime, rows=()):
        path = self.projects / f"{session}.jsonl"
        with path.open("a") as stream:
            for row in rows:
                stream.write(json.dumps(row) + "\n")
        os.utime(path, (mtime, mtime))

    def cli_row(self, session, cwd, pid=None, state="working", kind="background", name="worker"):
        row = {"id": session[:8], "sessionId": session, "name": name, "kind": kind, "state": state,
               "cwd": str(cwd), "startedAt": 0}
        if pid:
            row["pid"] = pid
        return row

    def deliver(self, registry, registration, message, handoff_id):
        self.deliveries.append(registration["sessionId"])
        with (self.projects / f"{registration['sessionId']}.jsonl").open("a") as stream:
            stream.write(json.dumps({"type": "user", "message": {"content": message}}) + "\n")

    def pin(self, root, session, *extra):
        return run_main("--repo", root, "pin", "--session", session, "--selection-ref", "management/v0/PROGRESS.md", *extra)

    def pin_file(self, root):
        return handoff.state_root() / handoff.repo_key(root) / "pin.json"

    # compaction ------------------------------------------------------------------------------------
    def write_log(self, root, data):
        path = root / LOG
        path.write_bytes(data)
        path.chmod(0o600)
        return path

    def compact_args(self, root, *extra, task="G006", disposition="completed", ref=REVIEW, sha=None,
                     text=QUOTE, writer=None):
        sha = sha or hashlib.sha256((root / ref).read_bytes()).hexdigest()
        return ["--repo", root, "compact", "--task", task, "--disposition", disposition, "--closure-ref", ref,
                "--closure-sha256", sha, "--closure-text", text, "--verified-by", "codex",
                "--writer-sha256", writer or SELF_SHA, *extra]

    def dry(self, root, *extra, **options):
        return run_main(*self.compact_args(root, *extra, **options))

    def apply(self, root, selection, *extra, **options):
        return run_main(*self.compact_args(root, "--apply", "--selection-sha256", selection, *extra, **options))

    def fake_proc(self, processes=(), cli_pids=(VISIBLE_PID,)):
        """Fake /proc: (pid, argv, fds[, cwd]) with a status file (real Uid = ours) and a cwd link (default: the
        case directory; cwd=None means no readable link). Tests may then remove or chmod entries."""
        proc = self.tmp / f"proc-{uuid.uuid4().hex[:6]}"
        uid = os.getuid()
        for pid, argv, fds, *cwd in [(VISIBLE_PID, ["claude"], [])] + list(processes):
            (proc / str(pid) / "fd").mkdir(parents=True)
            (proc / str(pid) / "cmdline").write_bytes(b"\0".join(a.encode() for a in argv) + b"\0")
            name = os.path.basename(argv[0])[:15] if argv else ""
            (proc / str(pid) / "status").write_text(f"Name:\t{name}\nState:\tS (sleeping)\n"
                                                    f"Uid:\t{uid}\t{uid}\t{uid}\t{uid}\n")
            for number, target in enumerate(fds, 3):
                os.symlink(target, proc / str(pid) / "fd" / str(number))
            target = cwd[0] if cwd else str(self.tmp)
            if target is not None:
                os.symlink(target, proc / str(pid) / "cwd")
        self.patch(handoff, "PROC", proc)
        self.cli.rows = [self.cli_row(str(uuid.uuid4()), "/fixture", pid=pid) for pid in cli_pids]
        return proc

    def closure_log(self, session):
        """A G006 triple (labelled) and an unattributed group: the smallest compaction fixture."""
        a, b = str(uuid.uuid4()), str(uuid.uuid4())
        data = (record(a, session, "prepared-not-sent", "G006") + record(a, session, "sent-receipt-pending", "G006")
                + record(b, session, "prepared-not-sent") + record(a, session, "receipt-confirmed", "G006"))
        return data, a, b


class T01RepositoryKey(Base):
    def test_t1_worktrees_share_a_key_and_other_repositories_do_not(self):
        root = self.make_repo()
        linked = self.worktree(root, "linked")
        other = self.make_repo("other")
        self.assertEqual(handoff.repo_key(root), handoff.repo_key(linked))
        self.assertNotEqual(handoff.repo_key(root), handoff.repo_key(other))
        self.assertRegex(handoff.repo_key(root), r"^[0-9a-f]{64}$")


class T02StateSafety(Base):
    def setUp(self):
        super().setUp()
        self.root = self.make_repo()
        self.session = str(uuid.uuid4())
        self.register(self.session, self.root)
        self.pin(self.root, self.session)

    def test_t2_modes_and_atomic_replace_leave_no_temporary_file(self):
        repo_dir = self.pin_file(self.root).parent
        self.assertEqual(stat.S_IMODE(os.stat(handoff.state_root()).st_mode), 0o700)
        self.assertEqual(stat.S_IMODE(os.stat(repo_dir).st_mode), 0o700)
        self.assertEqual(stat.S_IMODE(os.stat(self.pin_file(self.root)).st_mode), 0o600)
        second = str(uuid.uuid4())
        self.register(second, self.root)
        self.transcript(second, time.time())
        self.pin(self.root, second, "--replace")
        self.assertEqual(sorted(p.name for p in repo_dir.iterdir()), ["pin.json"])
        self.assertEqual(json.loads(self.pin_file(self.root).read_text())["replaces"], self.session)

    def test_t2_unsafe_state_is_rejected(self):
        pin = self.pin_file(self.root)
        original = pin.read_bytes()
        repo_dir = pin.parent

        def assert_unsafe():
            self.assertEqual(handoff.read_pin(self.root)[0], "unsafe")
            self.assertEqual(run_main("--repo", self.root, "status")["status"], "pin-unsafe")
            with self.assertRaises(handoff.Blocked):
                self.pin(self.root, self.session, "--replace")

        pin.chmod(0o644)
        assert_unsafe()
        pin.chmod(0o600)
        repo_dir.chmod(0o755)
        assert_unsafe()
        repo_dir.chmod(0o700)
        with patch.object(handoff.os, "getuid", return_value=os.getuid() + 1):
            self.assertEqual(handoff.read_pin(self.root)[0], "unsafe")
        for change in ({"schema": "other"}, {"repository": "/elsewhere"}, {"session": "not-a-uuid"}):
            pin.write_text(json.dumps({**json.loads(original), **change}))
            assert_unsafe()
        pin.write_bytes(original)
        target = repo_dir / "elsewhere.json"
        target.write_bytes(original)
        target.chmod(0o600)
        pin.unlink()
        os.symlink(target, pin)
        assert_unsafe()
        pin.unlink()
        target.rename(pin)
        self.assertEqual(handoff.read_pin(self.root)[0], "pinned")
        moved = self.tmp / "real-state"
        shutil.move(str(handoff.state_root()), str(moved))
        os.symlink(moved, handoff.state_root())
        assert_unsafe()

    def test_t2_symlinked_delivery_log_is_refused(self):
        (self.root / "management/other.jsonl").write_text("")
        os.symlink(self.root / "management/other.jsonl", self.root / LOG)
        with self.assertRaises(handoff.Blocked) as caught:
            handoff.append_record(self.root, LOG, {"handoff_id": "x"})
        self.assertEqual(str(caught.exception), "record-unsafe")
        self.assertEqual((self.root / "management/other.jsonl").read_text(), "")


class T03RepositoryCleanliness(Base):
    def test_t3_pin_send_and_compact_leave_only_the_ignored_log(self):
        root = self.make_repo()
        self.assertEqual(self.status(root), [])
        session = str(uuid.uuid4())
        self.register(session, root)
        self.pin(root, session)
        with patch.object(handoff, "send", side_effect=self.deliver):
            result = run_main("--repo", root, "send", "--message-file", self.message, "--task", "G006")
        self.assertEqual(result["status"], "receipt-confirmed")
        self.fake_proc()
        selection = self.dry(root)["selection_sha256"]
        self.assertEqual(self.apply(root, selection)["archived_records"], 3)
        self.assertEqual(self.status(root), ["!! " + LOG])


class T04Pin(Base):
    def setUp(self):
        super().setUp()
        self.root = self.make_repo()
        self.session = str(uuid.uuid4())
        patcher = patch.object(handoff, "send")
        self.transport = patcher.start()
        self.addCleanup(patcher.stop)

    def rejects(self, *extra, reason=None):
        before = self.pin_file(self.root).read_bytes() if self.pin_file(self.root).exists() else None
        with self.assertRaises(handoff.Blocked) as caught:
            self.pin(self.root, self.session, *extra)
        if reason:
            self.assertEqual(str(caught.exception), reason)
        after = self.pin_file(self.root).read_bytes() if self.pin_file(self.root).exists() else None
        self.assertEqual(before, after)
        self.transport.assert_not_called()

    def test_t4_accepts_a_verified_live_session(self):
        self.register(self.session, self.root, name="Coordinator")
        result = self.pin(self.root, self.session)
        self.assertEqual(result["status"], "pinned")
        pin = json.loads(self.pin_file(self.root).read_text())
        self.assertEqual(pin["session"], self.session)
        self.assertEqual(pin["selection_ref"]["path"], "management/v0/PROGRESS.md")
        self.assertNotIn("pid", pin)
        self.assertFalse({"token", "peerToken", "messagingSocketPath"} & set(pin))
        self.transport.assert_not_called()

    def test_t4_rejections(self):
        other_repo = self.make_repo("other")
        self.register(self.session, other_repo)
        self.rejects(reason="missing-or-ambiguous-live-session")
        self.unregister(self.session)
        self.register(self.session, self.root, pid=dead_pid(), start="1")
        self.rejects(reason="missing-or-ambiguous-live-session")
        self.unregister(self.session)
        self.register(self.session, self.root, start="1")
        self.rejects(reason="missing-or-ambiguous-live-session")
        self.unregister(self.session)
        self.register(self.session, self.root, protocol=2)
        self.rejects(reason="changed-peer-protocol")
        self.unregister(self.session)
        self.register(self.session, self.root)
        self.register(self.session, self.root)
        self.rejects(reason="missing-or-ambiguous-live-session")
        self.unregister(self.session)
        self.register(self.session, self.root)
        peer = str(uuid.uuid4())
        self.register(peer, self.root)
        self.transcript(self.session, 1000)
        self.transcript(peer, 2000)
        self.rejects(reason="newer-live-session-activity-check-task-owner-and-confirm-target")
        self.assertEqual(self.pin(self.root, self.session, "--confirm-target")["confirmed_newer_peers"], [peer])
        session, self.session = self.session, peer
        self.rejects("--confirm-target", reason="existing-pin-requires-replace")
        self.assertEqual(self.pin(self.root, peer, "--replace")["replaces"], session)


class T05Routing(Base):
    def setUp(self):
        super().setUp()
        self.root = self.make_repo()
        self.pinned, self.other = str(uuid.uuid4()), str(uuid.uuid4())
        self.register(self.pinned, self.root)
        self.register(self.other, self.root)
        self.transcript(self.pinned, 2000)
        self.transcript(self.other, 1000)
        self.pin(self.root, self.pinned)
        patcher = patch.object(handoff, "send", side_effect=self.deliver)
        patcher.start()
        self.addCleanup(patcher.stop)

    def send(self, *extra):
        return run_main("--repo", self.root, "send", "--message-file", self.message, "--wait-seconds", "1", *extra)

    def explicit_still_works(self):
        result = self.send("--session", self.other, "--confirm-target")
        self.assertEqual((result["status"], result["route"], self.deliveries[-1]),
                         ("receipt-confirmed", "explicit", self.other))

    def test_t5_default_reaches_the_pin_and_explicit_keeps_the_pin(self):
        result = self.send()
        self.assertEqual((result["status"], result["route"], self.deliveries), ("receipt-confirmed", "preferred", [self.pinned]))
        check = run_main("--repo", self.root, "check")
        self.assertEqual((check["session"], check["route"]), (self.pinned, "preferred"))
        before = self.pin_file(self.root).read_bytes()
        with self.assertRaises(handoff.Blocked):
            self.send("--session", self.other)
        self.explicit_still_works()
        self.assertEqual(self.pin_file(self.root).read_bytes(), before)
        records = [json.loads(line) for line in (self.root / LOG).read_text().splitlines()]
        self.assertEqual(records[-1]["route"], "explicit")
        self.assertEqual(records[-1]["confirmed_newer_peers"], [self.pinned])
        self.assertEqual({r["route"] for r in records[:3]}, {"preferred"})

    def test_t5_default_route_blocks_unless_live_verified(self):
        pin = self.pin_file(self.root)
        cases = []
        pin.chmod(0o644)
        cases.append("default-route-pin-unsafe")
        with self.assertRaises(handoff.Blocked) as caught:
            self.send()
        self.assertEqual(str(caught.exception), cases[-1])
        self.explicit_still_works()
        pin.chmod(0o600)
        extra = self.register(self.pinned, self.root)
        with self.assertRaises(handoff.Blocked) as caught:
            self.send()
        self.assertEqual(str(caught.exception), "default-route-pinned-ambiguous")
        self.explicit_still_works()
        extra.unlink()
        self.unregister(self.pinned)
        unverified = self.register(self.pinned, self.root, protocol=2)
        self.assertEqual(run_main("--repo", self.root, "status")["status"], "pinned-unverified")
        with self.assertRaises(handoff.Blocked) as caught:
            self.send()
        self.assertEqual(str(caught.exception), "default-route-pinned-unverified")
        self.explicit_still_works()
        unverified.unlink()
        with self.assertRaises(handoff.Blocked) as caught:
            self.send()
        self.assertEqual(str(caught.exception), "default-route-pinned-not-live")
        self.explicit_still_works()
        run_main("--repo", self.root, "unpin", "--session", self.pinned)
        with self.assertRaises(handoff.Blocked) as caught:
            self.send()
        self.assertEqual(str(caught.exception), "default-route-unpinned")
        self.explicit_still_works()
        self.assertNotIn(self.pinned, self.deliveries)


class T06Receipt(Base):
    def setUp(self):
        super().setUp()
        self.root = self.make_repo()
        self.session, self.other = str(uuid.uuid4()), str(uuid.uuid4())
        self.register(self.session, self.root)
        self.transcript(self.session, time.time())

    def test_t6_receipt_uses_the_logged_session(self):
        with patch.object(handoff, "send", side_effect=self.deliver):
            sent = run_main("--repo", self.root, "send", "--session", self.session, "--message-file", self.message)
        result = run_main("--repo", self.root, "receipt", "--handoff-id", sent["handoff_id"])
        self.assertEqual(result["status"], "receipt-confirmed")
        with self.assertRaises(handoff.Blocked) as caught:
            run_main("--repo", self.root, "receipt", "--handoff-id", sent["handoff_id"], "--session", self.other)
        self.assertEqual(str(caught.exception), "receipt-session-mismatch")

    def test_t6_missing_or_ambiguous_record_needs_session(self):
        missing = str(uuid.uuid4())
        with self.assertRaises(handoff.Blocked) as caught:
            run_main("--repo", self.root, "receipt", "--handoff-id", missing)
        self.assertEqual(str(caught.exception), "receipt-session-unresolved-pass-session")
        self.assertEqual(run_main("--repo", self.root, "receipt", "--handoff-id", missing,
                                  "--session", self.session)["status"], "delivery-unconfirmed")
        ambiguous = str(uuid.uuid4())
        self.write_log(self.root, record(ambiguous, self.session, "prepared-not-sent")
                       + record(ambiguous, self.other, "prepared-not-sent"))
        with self.assertRaises(handoff.Blocked):
            run_main("--repo", self.root, "receipt", "--handoff-id", ambiguous)
        self.assertEqual(run_main("--repo", self.root, "receipt", "--handoff-id", ambiguous,
                                  "--session", self.session)["status"], "delivery-unconfirmed")

    def test_t6_enqueue_alone_is_only_queued_observed(self):
        queued = str(uuid.uuid4())
        self.transcript(self.session, time.time(), [{"type": "queue-operation", "operation": "enqueue",
                                                      "content": f"Fixture\n[handoff-id: {queued}]"}])
        result = handoff.receipt_status(self.projects.parent, self.session, queued)
        self.assertEqual((result["status"], result["queued_observed"]), ("delivery-unconfirmed", True))


class T07Discovery(Base):
    def setUp(self):
        super().setUp()
        self.root = self.make_repo()

    def test_t7_complete_listing_outcomes(self):
        empty = run_main("--repo", self.root, "status")
        self.assertEqual((empty["status"], empty["discovery"]["state"], empty["candidates"]), ("unpinned", "complete", []))
        one = str(uuid.uuid4())
        self.register(one, self.root, name="Coordinator")
        self.cli.rows = [self.cli_row(one, self.root, pid=os.getpid(), name="Coordinator")]
        result = run_main("--repo", self.root, "status")
        self.assertEqual([c["session"] for c in result["candidates"]], [one])
        two = str(uuid.uuid4())
        self.register(two, self.root, name="Coordinator")
        self.register(two, self.root, pid=dead_pid(), start="1", name="Coordinator")
        result = run_main("--repo", self.root, "status")
        self.assertEqual(sorted(c["session"] for c in result["candidates"]), sorted([one, two]))
        self.assertEqual(result["discovery"]["duplicate_names"], {"Coordinator": sorted([one, two])})
        self.assertEqual(result["discovery"]["duplicate_registrations"], {two: 2})
        self.assertEqual(len(result["discovery"]["stale_registrations"]), 1)

    def test_t7_unknown_is_never_none(self):
        self.cli.fail = True
        result = run_main("--repo", self.root, "status")
        self.assertEqual((result["status"], result["discovery"]["reasons"]), ("discovery-unknown", ["cli-listing-unavailable"]))
        self.cli.fail = False
        self.cli.rows = [self.cli_row(str(uuid.uuid4()), self.root, pid=dead_pid())]
        self.assertEqual(run_main("--repo", self.root, "status")["discovery"]["reasons"], ["pid-visibility-unproven"])
        self.cli.rows = []
        gone = self.tmp / "removed-worktree"
        self.register(str(uuid.uuid4()), gone)
        self.assertEqual(run_main("--repo", self.root, "status")["discovery"]["reasons"], ["live-cwd-unresolved"])

    def test_t7_registry_access_failures_are_unknown(self):
        def reasons():
            return run_main("--repo", self.root, "status")["discovery"]["reasons"]
        broken = self.registry / "broken.json"
        broken.write_text("{not json")
        self.assertEqual(reasons(), ["registry-row-unreadable"])
        broken.unlink()
        row = {"pid": "not-a-pid", "sessionId": str(uuid.uuid4()), "cwd": str(self.root), "procStart": "1"}
        (self.registry / "bad-pid.json").write_text(json.dumps(row))
        self.assertEqual(reasons(), ["registry-row-unverifiable"])
        other = self.make_repo("other")
        (self.registry / "bad-pid.json").write_text(json.dumps({**row, "cwd": str(other)}))
        self.assertEqual(reasons(), [])
        (self.registry / "bad-pid.json").unlink()
        child = subprocess.Popen([sys.executable, "-c", "import time; time.sleep(60)"])
        self.addCleanup(child.wait)
        self.addCleanup(child.kill)
        hidden = child.pid  # a live process with a readable /proc start time in every context (no parent PID 0)
        self.register(str(uuid.uuid4()), self.root, pid=hidden)
        real_kill = os.kill

        def kill(pid, signal):
            if pid == hidden:
                raise PermissionError(1, "Operation not permitted")
            return real_kill(pid, signal)
        with patch.object(handoff.os, "kill", side_effect=kill):
            self.assertEqual(reasons(), ["registry-row-unverifiable"])
            self.unregister_all()
            self.register(str(uuid.uuid4()), self.root, pid=hidden, start="1")
            result = run_main("--repo", self.root, "status")
            self.assertEqual(result["discovery"]["reasons"], [], "a reused PID is a stale registration, not unknown")
            self.assertEqual(len(result["discovery"]["stale_registrations"]), 1)

    def unregister_all(self):
        for path in self.registry.glob("*.json"):
            path.unlink()

    def test_t7_c1_readable_non_git_cwd_is_outside_the_repository(self):
        identity = handoff.repo_identity(self.root)
        plain = self.tmp / "plain-directory"
        plain.mkdir()
        foreign = self.make_repo("foreign")
        self.assertIs(handoff.same_repo(str(self.root), identity), True)
        self.assertIs(handoff.same_repo(str(plain), identity), False)
        self.assertIs(handoff.same_repo(str(foreign), identity), False)
        outsider, stranger = str(uuid.uuid4()), str(uuid.uuid4())
        self.register(outsider, plain)
        self.register(stranger, foreign)
        self.cli.rows = [self.cli_row(outsider, plain, pid=os.getpid()), self.cli_row(stranger, foreign, pid=os.getpid())]
        result = run_main("--repo", self.root, "status")
        self.assertEqual((result["status"], result["discovery"]["state"], result["discovery"]["reasons"]),
                         ("unpinned", "complete", []))
        self.assertEqual(result["candidates"], [])
        gate = run_main("--repo", self.root, "status", "--creation-check", "--coordinator-uuid", str(uuid.uuid4()))
        self.assertNotIn("discovery-unknown", {r["code"] for r in gate["refusals"]})
        self.assertNotIn("ambiguous-candidates", {r["code"] for r in gate["refusals"]})

    def test_t7_c1_unresolved_cwds_stay_unknown_and_prevent_creation_or_send(self):
        identity = handoff.repo_identity(self.root)
        missing = self.tmp / "missing"
        not_directory = self.tmp / "a-file"
        not_directory.write_text("x")
        invalid = self.tmp / "invalid-gitfile"
        invalid.mkdir()
        (invalid / ".git").write_text("gitdir: /nonexistent/g052-fixture\n")
        unreadable = self.tmp / "unreadable"
        unreadable.mkdir()
        for path in (missing, not_directory, invalid):
            self.assertIsNone(handoff.same_repo(str(path), identity), path)
        real_scandir, real_run = os.scandir, subprocess.run

        def scandir(path):
            if os.fspath(path) == str(unreadable):
                raise PermissionError(13, "Permission denied (injected)", str(unreadable))
            return real_scandir(path)
        with patch.object(handoff.os, "scandir", side_effect=scandir):
            self.assertIsNone(handoff.same_repo(str(unreadable), identity))
        plain = self.tmp / "plain-but-git-fails"
        plain.mkdir()
        outcomes = [
            (128, "fatal: detected dubious ownership in repository at '/srv/repo'\n"
                  "To add an exception for this directory, call:\n\n\tgit config --global --add safe.directory /srv/repo\n"),
            (128, "fatal: not a git repository (or any parent up to mount point /mnt)\n"
                  "Stopping at filesystem boundary (GIT_DISCOVERY_ACROSS_FILESYSTEM not set).\n"),
            (129, "error: unknown option `path-format=absolute'\n"),
            (1, ""),
        ]
        for code, stderr in outcomes:
            def run(arguments, *args, **kwargs):
                if "--git-common-dir" in arguments and str(plain) in arguments:
                    return subprocess.CompletedProcess(arguments, code, "", stderr)
                return real_run(arguments, *args, **kwargs)
            with patch.object(handoff.subprocess, "run", side_effect=run):
                self.assertIsNone(handoff.same_repo(str(plain), identity), (code, stderr))
        patcher = patch.object(handoff, "send")
        transport = patcher.start()
        self.addCleanup(patcher.stop)
        session = str(uuid.uuid4())
        self.register(session, self.root)
        self.pin(self.root, session)
        self.unregister(session)
        self.register(session, unreadable)
        with patch.object(handoff.os, "scandir", side_effect=scandir):
            status = run_main("--repo", self.root, "status")
            self.assertEqual((status["status"], status["discovery"]["reasons"]), ("discovery-unknown", ["live-cwd-unresolved"]))
            gate = run_main("--repo", self.root, "status", "--creation-check", "--coordinator-uuid", str(uuid.uuid4()))
            self.assertIn("discovery-unknown", {r["code"] for r in gate["refusals"]})
            self.assertEqual((gate["status"], gate["spawned"], "command" in gate), ("blocked", 0, False))
            with self.assertRaises(handoff.Blocked) as caught:
                run_main("--repo", self.root, "send", "--message-file", self.message)
            self.assertEqual(str(caught.exception), "default-route-pinned-not-live")
        transport.assert_not_called()

    def test_t7_same_uuid_new_pid_and_changed_uuid(self):
        session = str(uuid.uuid4())
        first = self.register(session, self.root)
        self.pin(self.root, session)
        pin_bytes = self.pin_file(self.root).read_bytes()
        child = subprocess.Popen([sys.executable, "-c", "import time; time.sleep(30)"])
        self.addCleanup(child.wait)
        self.addCleanup(child.kill)
        first.unlink()
        self.register(session, self.root, pid=child.pid)
        self.assertEqual(run_main("--repo", self.root, "status")["status"], "pinned-live-verified")
        self.unregister(session)
        fork = str(uuid.uuid4())
        self.register(fork, self.root)
        result = run_main("--repo", self.root, "status")
        self.assertEqual(result["status"], "pinned-not-live")
        self.assertEqual([c["session"] for c in result["candidates"]], [fork])
        self.cli.fail = True
        self.assertEqual(run_main("--repo", self.root, "status")["status"], "discovery-unknown")
        self.assertEqual(self.pin_file(self.root).read_bytes(), pin_bytes)


class T08CreationGate(Base):
    FORBIDDEN = ("--permission", "--model", "--settings", "--dangerously", "--allowed", "--disallowed",
                 "--resume", "--continue", "--mcp", "--add-dir", "--fork")

    def setUp(self):
        super().setUp()
        self.coordinator = str(uuid.uuid4())
        self.prompt = self.tmp / "prompt.txt"
        self.prompt.write_text("Coordinate the repository tasks.")
        self.calls = []
        self.real_run = subprocess.run
        real = self.real_run

        def spy(arguments, *args, **kwargs):
            self.calls.append(list(arguments))
            return real(arguments, *args, **kwargs)
        self.patch(handoff.subprocess, "run", spy)

    def authority_repo(self, names=(), commit=True, outside_quote=()):
        """The verified quote names the non-reusable live candidates; outside_quote names others elsewhere."""
        quote = f"{DECISION}; non-reusable live candidates: {' '.join(names) or 'none'}"
        text = f"# Creation authority\n\n{quote}.\n\nOther notes: {' '.join(outside_quote)}\n"
        root = self.make_repo(f"repo-{uuid.uuid4().hex[:6]}", extra={AUTHORITY: text.encode()} if commit else None)
        if not commit:
            (root / AUTHORITY).write_text(text)
        self.quote = quote
        return root

    def gate(self, root, sha=None, text=None, ref=AUTHORITY, coordinators=None):
        sha = sha or (hashlib.sha256((root / ref).read_bytes()).hexdigest() if (root / ref).exists() else "0" * 64)
        argv = ["--repo", root, "status", "--creation-check", "--authority-ref", ref, "--authority-sha256", sha,
                "--authority-text", text or self.quote, "--creation-name", "Coordinator",
                "--creation-prompt-file", self.prompt]
        for value in (coordinators if coordinators is not None else [self.coordinator]):
            argv += ["--coordinator-uuid", value]
        return run_main(*argv)

    def refused(self, result, *codes):
        self.assertEqual((result["status"], result["reason"], result["spawned"]), ("blocked", "creation-refused", 0))
        self.assertNotIn("command", result)
        self.assertTrue(set(codes) <= {r["code"] for r in result["refusals"]}, result["refusals"])
        return {r["code"]: r["detail"] for r in result["refusals"]}

    def operator_runs(self, result):
        """The documented procedure: the operator runs the printed command once from the printed cwd. A stub
        `claude` first on PATH records exactly what it receives and starts nothing."""
        bin_dir, log = self.tmp / "operator-bin", self.tmp / "claude-stub.jsonl"
        bin_dir.mkdir()
        stub = bin_dir / "claude"
        stub.write_text(f"#!{sys.executable}\nimport json, os, sys\n"
                        "open(os.environ['CLAUDE_STUB_LOG'], 'a').write("
                        "json.dumps({'argv': sys.argv[1:], 'cwd': os.getcwd()}) + '\\n')\n")
        stub.chmod(0o700)
        env = {**os.environ, "PATH": f"{bin_dir}{os.pathsep}{os.environ.get('PATH', '')}", "CLAUDE_STUB_LOG": str(log)}
        self.real_run(result["command"], cwd=result["cwd"], env=env, check=True)
        return [json.loads(line) for line in log.read_text().splitlines()]

    def test_t8_only_bound_zero_reusable_case_returns_a_command(self):
        candidate = str(uuid.uuid4())
        root = self.authority_repo([candidate])
        self.register(candidate, root)
        self.cli.rows = [self.cli_row(candidate, root, pid=os.getpid())]
        before = self.status(root)
        result = self.gate(root)
        self.assertEqual(result["status"], "creation-permitted")
        self.assertEqual(result["cwd"], str(root))
        self.assertFalse([c for c in self.calls if c and c[0] == "claude"], "the helper itself never runs claude")
        invocations = self.operator_runs(result)
        self.assertEqual(invocations, [{"argv": ["--bg", "--name", "Coordinator", "--effort", "xhigh",
                                                 "Coordinate the repository tasks."], "cwd": str(root)}])
        self.assertFalse([a for a in invocations[0]["argv"] if a.startswith(self.FORBIDDEN)])
        self.assertEqual(self.status(root), before)
        self.assertFalse(handoff.state_root().exists())

    def test_t8_refusals_and_zero_spawns(self):
        root = self.authority_repo()
        self.register(self.coordinator, root)
        self.refused(self.gate(root), "reusable-live-coordinator")
        self.unregister(self.coordinator)
        self.cli.rows = [self.cli_row(self.coordinator, root, state="stopped")]
        self.refused(self.gate(root), "reusable-stopped-coordinator")
        self.cli.rows = []
        self.transcript(self.coordinator, time.time())
        self.refused(self.gate(root), "reusable-stopped-coordinator")
        (self.projects / f"{self.coordinator}.jsonl").unlink()
        stranger = str(uuid.uuid4())
        self.register(stranger, root)
        self.refused(self.gate(root), "ambiguous-candidates")
        elsewhere = self.authority_repo(outside_quote=[stranger])
        self.register(stranger, elsewhere)
        self.assertEqual(self.refused(self.gate(elsewhere), "ambiguous-candidates")["ambiguous-candidates"], [stranger])
        self.unregister(stranger)
        self.cli.fail = True
        self.refused(self.gate(root), "discovery-unknown")
        self.cli.fail = False
        self.cli.rows = [self.cli_row(str(uuid.uuid4()), root, pid=dead_pid())]
        self.refused(self.gate(root), "discovery-unknown")
        self.cli.rows = []
        self.refused(self.gate(root, sha="1" * 64), "missing-authority")
        self.refused(self.gate(root, text="a different decision"), "missing-authority")
        self.refused(self.gate(root, ref="management/v0/results/absent.md"), "missing-authority")
        self.refused(run_main("--repo", root, "status", "--creation-check", "--coordinator-uuid", self.coordinator),
                     "missing-authority")
        root_quote = f"{DECISION}; non-reusable live candidates: none"
        untracked = self.authority_repo(commit=False)
        self.assertEqual(self.refused(self.gate(untracked, text=root_quote), "missing-authority")["missing-authority"],
                         ["authority-untracked-or-modified"])
        named = self.authority_repo([self.coordinator])
        self.refused(self.gate(named), "coordinator-candidate-conflict")
        with (root / AUTHORITY).open("a") as stream:
            stream.write("Later local edit.\n")
        self.assertEqual(self.refused(self.gate(root, text=root_quote), "missing-authority")["missing-authority"],
                         ["authority-untracked-or-modified"])
        clean = self.authority_repo()
        self.cli.version = "2.1.297 (Claude Code)"
        self.refused(self.gate(clean), "unverified-cli-version")
        self.cli.fail = True
        self.register(self.coordinator, clean)
        self.refused(self.gate(clean), "unverified-cli-version", "discovery-unknown", "reusable-live-coordinator")
        self.assertFalse([c for c in self.calls if c and c[0] == "claude"])

    def test_t8_symlinked_authority_invalid_ids_and_unsafe_prompts_refuse(self):
        root = self.authority_repo()
        real = root / "management/v0/results/real-authority.md"
        shutil.move(str(root / AUTHORITY), str(real))
        os.symlink(real.name, root / AUTHORITY)
        self.commit(root)
        details = self.refused(self.gate(root), "missing-authority")
        self.assertEqual(details["missing-authority"], ["authority-unsafe"])
        clean = self.authority_repo()
        row = {"pid": os.getpid(), "sessionId": "", "cwd": str(clean), "procStart": proc_start(os.getpid()),
               "peerProtocol": 1, "version": "2.1.296"}
        (self.registry / "empty-session.json").write_text(json.dumps(row))
        details = self.refused(self.gate(clean), "ambiguous-candidates")
        self.assertEqual(details["ambiguous-candidates"], ["invalid-session-id:''"])
        (self.registry / "empty-session.json").unlink()
        self.cli.rows = [self.cli_row("not-a-uuid-session", clean, pid=os.getpid())]
        self.refused(self.gate(clean), "ambiguous-candidates")
        self.cli.rows = []
        for prompt in ("update", "-p hello", " "):
            self.prompt.write_text(prompt)
            self.refused(self.gate(clean), "creation-command-arguments-invalid")
        self.assertFalse([c for c in self.calls if c and c[0] == "claude"])

    def test_t8_no_create_command_and_cli_entry_refuses_spawns(self):
        with contextlib.redirect_stderr(io.StringIO()), self.assertRaises(SystemExit):
            run_main("create")
        with patch.object(handoff.subprocess, "run") as never:
            for arguments in (["--bg", "--name", "x", "prompt"], ["--resume", str(uuid.uuid4())], ["agents"],
                              ["agents", "--json"], ["agents", "--json", "--all", "--cwd", "relative"],
                              ["agents", "--json", "--all", "--cwd", "/repo", "--extra"], ["--version", "--bg"]):
                with self.assertRaises(handoff.Blocked):
                    REAL_RUN_CLAUDE(arguments)
            never.assert_not_called()


class T09CompactionSelection(Base):
    def setUp(self):
        super().setUp()
        self.root = self.make_repo()
        s, t = str(uuid.uuid4()), str(uuid.uuid4())
        self.ids = {name: str(uuid.uuid4()) for name in "ABCDEFGH"}
        i = self.ids
        self.lines = [
            ("A", record(i["A"], s, "prepared-not-sent", "G006")), ("E", record(i["E"], s, "prepared-not-sent", "G007")),
            ("B", record(i["B"], s, "prepared-not-sent")), ("A", record(i["A"], s, "sent-receipt-pending", "G006")),
            ("M", b"not json\n"), ("B", record(i["B"], s, "sent-receipt-pending")),
            ("C", record(i["C"], s, "prepared-not-sent", "G006")), ("F", record(i["F"], s, "prepared-not-sent", "G007")),
            ("C", record(i["C"], s, "sent-receipt-pending", "G006")), ("G", record(i["G"], s, "prepared-not-sent")),
            ("G", record(i["G"], t, "sent-receipt-pending")), ("C", record(i["C"], s, "delivery-unconfirmed", "G006")),
            ("D", record(i["D"], s, "prepared-not-sent", "G006")), ("H", record(i["H"], s, "prepared-not-sent")),
            ("A", record(i["A"], s, "receipt-confirmed", "G006")), ("T", record(i["H"], s, "sent")[:-9]),
        ]
        self.data = b"".join(line for _, line in self.lines)
        self.log = self.write_log(self.root, self.data)
        self.listed = [arg for name in "BFG" for arg in ("--handoff-id", i[name])]

    def expected(self, archived):
        return b"".join(line for name, line in self.lines if (name in "ABCD") == archived)

    def test_t9_dry_run_selects_whole_attributed_groups_and_writes_nothing(self):
        result = self.dry(self.root, *self.listed)
        self.assertEqual((result["status"], result["writes"]), ("compaction-dry-run", 0))
        self.assertEqual(sorted(result["handoff_ids"]), sorted(self.ids[n] for n in "ABCD"))
        self.assertEqual(sorted(result["incomplete_delivery"]), sorted([self.ids["B"], self.ids["D"]]))
        self.assertEqual(result["archived_records"], 9)
        self.assertEqual(result["retained_by_reason"], {"conflicting-label": 1, "malformed-or-truncated": 2,
                                                        "multi-session": 2, "other-task": 1, "unattributed": 1})
        self.assertEqual(self.log.read_bytes(), self.data)
        self.assertFalse((handoff.state_root()).exists() and any(handoff.state_root().rglob("archive")))

    def test_t9_apply_moves_exact_bytes_and_refuses_a_stale_selection(self):
        selection = self.dry(self.root, *self.listed)["selection_sha256"]
        self.fake_proc()
        with self.assertRaises(handoff.Blocked) as caught:
            self.apply(self.root, "0" * 64, *self.listed)
        self.assertEqual(str(caught.exception), "stale-selection")
        self.assertEqual(self.log.read_bytes(), self.data)
        result = self.apply(self.root, selection, *self.listed)
        self.assertEqual(result["status"], "compaction-applied")
        archive_dir = handoff.state_root() / handoff.repo_key(self.root) / "archive"
        archive = archive_dir / result["archive"]
        self.assertEqual(archive.read_bytes(), self.expected(True))
        self.assertEqual(self.log.read_bytes(), self.expected(False))
        manifest_path = archive.with_name(archive.name.replace(".jsonl", ".manifest.json"))
        manifest = json.loads(manifest_path.read_bytes())
        self.assertEqual(hashlib.sha256(manifest_path.read_bytes()).hexdigest(), result["manifest_sha256"])
        self.assertEqual(manifest["closure"]["text"], QUOTE)
        self.assertEqual(manifest["source"]["sha256"], hashlib.sha256(self.data).hexdigest())
        self.assertEqual(manifest["counts"]["archived_records"], 9)
        self.assertEqual(manifest["counts"]["archived_groups"], 4)
        self.assertEqual((manifest["pruned"], manifest["prune_status"]), (True, "pruned"))
        self.assertEqual(manifest["live_sha256"], hashlib.sha256(self.expected(False)).hexdigest())
        self.assertEqual(manifest["archive"]["sha256"], hashlib.sha256(self.expected(True)).hexdigest())
        for path in (archive, manifest_path):
            self.assertEqual(stat.S_IMODE(path.stat().st_mode), 0o600)
        self.assertEqual(stat.S_IMODE(archive_dir.stat().st_mode), 0o700)


class T10ClosureRefusals(Base):
    def setUp(self):
        super().setUp()
        self.task_id = str(uuid.uuid4())
        statement = f"G052 Claude planning delegation requested; handoff {self.task_id}; ACCEPT pending."
        extra = {f"management/v0/{kind}/G052-fixture.md": statement.encode() for kind in ("goals", "specs", "plans")}
        extra["management/v0/PROGRESS.md"] = f"# Progress\n\n{statement}\n".encode()
        extra["management/v0/results/G006-other-record.md"] = f"{QUOTE}.\n".encode()
        self.root = self.make_repo(extra=extra)
        self.statement = statement
        self.data = record(self.task_id, str(uuid.uuid4()), "prepared-not-sent", "G052")
        self.log = self.write_log(self.root, self.data)

    def refused(self, reason, **options):
        with self.assertRaises(handoff.Blocked) as caught:
            self.dry(self.root, "--handoff-id", self.task_id, **options)
        self.assertEqual(str(caught.exception), reason)
        self.assertEqual(self.log.read_bytes(), self.data)

    def test_t10_active_references_and_mismatches_are_refused(self):
        for kind in ("goals", "specs", "plans"):
            for disposition in ("completed", "canceled"):
                self.refused("closure-not-terminal-record", task="G052", disposition=disposition,
                             ref=f"management/v0/{kind}/G052-fixture.md", text=self.statement)
        self.refused("closure-not-implementation-review", task="G052", ref="management/v0/PROGRESS.md", text=self.statement)
        self.refused("closure-disposition-mismatch", task="G052", disposition="canceled",
                     ref="management/v0/PROGRESS.md", text=self.statement)
        self.refused("closure-disposition-mismatch", disposition="canceled")
        self.refused("closure-not-implementation-review", ref="management/v0/results/G006-other-record.md")

    def test_t10_ambiguous_closure_is_refused(self):
        self.refused("closure-quote-missing", text="**ACCEPT** — G006 something else")
        self.refused("closure-quote-lacks-task", text="**ACCEPT**")
        self.refused("closure-hash-mismatch", sha="1" * 64)
        untracked = "management/v0/results/G006-implementation-codex-review-r2.md"
        (self.root / untracked).write_text(f"{QUOTE}.\n")
        self.refused("closure-untracked-or-modified", ref=untracked)
        with (self.root / REVIEW).open("a") as stream:
            stream.write("Reopened locally.\n")
        self.refused("closure-untracked-or-modified")

    def test_t10_symlinked_closure_reference_is_refused(self):
        to_goal = "management/v0/results/G052-implementation-codex-review.md"
        os.symlink("../goals/G052-fixture.md", self.root / to_goal)
        (self.root / "management/v0/results/untracked.md").write_text(f"{QUOTE}.\n")
        to_untracked = "management/v0/results/G006-implementation-codex-review-alias.md"
        os.symlink("untracked.md", self.root / to_untracked)
        run_git("-C", self.root, "add", to_goal, to_untracked)
        run_git("-C", self.root, "-c", "user.name=fixture", "-c", "user.email=fixture@example.invalid",
                "-c", "commit.gpgsign=false", "commit", "-q", "-m", "links")
        self.refused("closure-unsafe", task="G052", ref=to_goal, text=self.statement)
        self.refused("closure-unsafe", ref=to_untracked)

    def test_t10_literal_pathspecs_and_same_inode_read(self):
        self.assertTrue(handoff.tracked_and_clean(self.root, REVIEW))
        self.assertFalse(handoff.tracked_and_clean(self.root, "management/v0/results/*.md"))
        self.assertFalse(handoff.tracked_and_clean(self.root, "management/v0/results/G006-implementation-codex-review.m?"))
        real_lstat, other = os.lstat, self.root / "management/v0/PROGRESS.md"

        def swapped(path, *args, **kwargs):
            return real_lstat(other) if Path(path) == self.root / REVIEW else real_lstat(path, *args, **kwargs)
        with patch.object(handoff.os, "lstat", side_effect=swapped):
            self.refused("closure-unsafe")


class T11Concurrency(Base):
    def setUp(self):
        super().setUp()
        self.root = self.make_repo()
        self.session = str(uuid.uuid4())
        self.register(self.session, self.root)
        self.transcript(self.session, time.time())
        self.data, self.closed, self.kept = self.closure_log(self.session)
        self.log = self.write_log(self.root, self.data)
        self.transport = self.deliver
        patcher = patch.object(handoff, "send", side_effect=lambda *a: self.transport(*a))
        patcher.start()
        self.addCleanup(patcher.stop)
        self.selection = self.dry(self.root)["selection_sha256"]
        self.fake_proc()
        self.outcome = {}

    def sender(self):
        self.outcome["result"] = run_main("--repo", self.root, "send", "--session", self.session, "--message-file",
                                          self.message, "--wait-seconds", "2", "--lock-wait-seconds", "20")

    def archives(self):
        return list((handoff.state_root() / handoff.repo_key(self.root) / "archive").glob("*.jsonl"))

    def test_t11_in_flight_send_delays_compact_apply_which_refuses(self):
        entered, release = threading.Event(), threading.Event()

        def slow_transport(*args):
            entered.set()
            release.wait(20)
            self.deliver(*args)
        self.transport = slow_transport
        thread = threading.Thread(target=self.sender)
        thread.start()
        self.assertTrue(entered.wait(10))
        started = time.monotonic()
        with self.assertRaises(handoff.Blocked) as caught:
            self.apply(self.root, self.selection, "--lock-wait-seconds", "1")
        self.assertGreaterEqual(time.monotonic() - started, 0.9)
        self.assertEqual(str(caught.exception), "maintenance-lock-busy")
        release.set()
        thread.join(30)
        self.assertEqual(self.outcome["result"]["status"], "receipt-confirmed")
        self.assertTrue(self.log.read_bytes().startswith(self.data))
        self.assertEqual(len(self.log.read_bytes().splitlines()), 4 + 3)
        self.assertEqual(self.archives(), [])

    def test_t11_send_started_during_compact_apply_appends_to_the_new_file(self):
        real_write_archive = handoff.write_archive
        thread = threading.Thread(target=self.sender)

        def archive_while_send_waits(*args, **kwargs):
            thread.start()
            time.sleep(0.5)
            self.assertEqual(self.log.read_bytes(), self.data, "the send must wait for the exclusive lock")
            return real_write_archive(*args, **kwargs)
        with patch.object(handoff, "write_archive", side_effect=archive_while_send_waits):
            result = self.apply(self.root, self.selection)
        thread.join(30)
        self.assertEqual(result["archived_records"], 3)
        self.assertEqual(self.outcome["result"]["status"], "receipt-confirmed")
        lines = self.log.read_bytes().splitlines(keepends=True)
        self.assertEqual(lines[0], record(self.kept, self.session, "prepared-not-sent"))
        self.assertEqual([json.loads(line)["status"] for line in lines[1:]],
                         ["prepared-not-sent", "sent-receipt-pending", "receipt-confirmed"])
        archived = set().union(*(ids_in(path.read_bytes()) for path in self.archives()))
        self.assertEqual(archived, {self.closed})
        self.assertEqual(archived | ids_in(self.log.read_bytes()),
                         {self.closed, self.kept, self.outcome["result"]["handoff_id"]})

    def test_t11_repeated_closeout_archives_zero(self):
        self.assertEqual(self.apply(self.root, self.selection)["archived_records"], 3)
        again = self.dry(self.root)
        self.assertEqual(again["archived_records"], 0)
        repeat = self.apply(self.root, again["selection_sha256"])
        self.assertEqual((repeat["archived_records"], repeat["archive"]), (0, None))
        self.assertEqual(len(self.archives()), 1)

    def test_t11_send_blocks_without_record_or_delivery_when_the_lock_is_busy(self):
        delivered = []
        self.transport = lambda *args: delivered.append(args)
        with handoff.maintenance_lock(self.root, self.log, True, 0):
            with self.assertRaises(handoff.Blocked) as caught:
                run_main("--repo", self.root, "send", "--session", self.session, "--message-file", self.message,
                         "--lock-wait-seconds", "1")
        self.assertEqual(str(caught.exception), "maintenance-lock-busy")
        self.assertEqual(self.log.read_bytes(), self.data)
        self.assertEqual(delivered, [])


class T12ExistingBehaviour(Base):
    def test_t12_no_writes_under_claude_state_and_no_retries(self):
        root = self.make_repo()
        session = str(uuid.uuid4())
        self.register(session, root)
        self.transcript(session, time.time())

        def tree():
            return sorted((str(p), p.stat().st_size, p.stat().st_mtime_ns) for p in (self.home / ".claude").rglob("*"))
        before = tree()
        self.pin(root, session)
        run_main("--repo", root, "status")
        with patch.object(handoff, "send") as transport:
            result = run_main("--repo", root, "send", "--message-file", self.message, "--wait-seconds", "0")
        self.assertEqual(result["status"], "delivery-unconfirmed")
        transport.assert_called_once()
        self.dry(root)
        self.assertEqual(tree(), before)


class T13LineBoundary(Base):
    def test_t13_append_after_a_missing_final_newline_starts_a_new_line(self):
        root = self.make_repo()
        log = self.write_log(root, b'{"partial":')
        handoff.append_record(root, LOG, {"handoff_id": "x", "status": "prepared-not-sent"})
        self.assertEqual(log.read_bytes(), b'{"partial":\n{"handoff_id":"x","status":"prepared-not-sent"}\n')
        handoff.append_record(root, LOG, {"handoff_id": "y"})
        self.assertTrue(log.read_bytes().endswith(b'"prepared-not-sent"}\n{"handoff_id":"y"}\n'))


class T14LegacyWriter(Base):
    def setUp(self):
        super().setUp()
        self.root = self.make_repo()
        self.data, self.closed, self.kept = self.closure_log(str(uuid.uuid4()))
        self.log = self.write_log(self.root, self.data)
        self.selection = self.dry(self.root)["selection_sha256"]

    def refused(self, *failures, root=None, **options):
        root = root or self.root
        dry = self.dry(root, **options)["prerequisite"]
        self.assertFalse(dry["satisfied"])
        self.assertTrue(set(failures) <= set(dry["failures"]), dry["failures"])
        with self.assertRaises(handoff.Blocked) as caught:
            self.apply(root, self.selection, **options)
        self.assertEqual(str(caught.exception), "legacy-writer-prerequisite-unmet")
        self.assertTrue(set(failures) <= set(caught.exception.details["prerequisite"]["failures"]))
        self.assertEqual((root / LOG).read_bytes(), self.data)
        self.assertFalse(list((handoff.state_root() / handoff.repo_key(root)).glob("archive/*")))
        return caught.exception.details["prerequisite"]

    def test_t14_satisfied_prerequisite_is_reported_by_dry_run(self):
        self.fake_proc()
        evidence = self.dry(self.root)["prerequisite"]
        self.assertEqual((evidence["satisfied"], evidence["failures"]), (True, []))
        self.assertEqual(evidence["running_sha256"], SELF_SHA)

    def test_t14_legacy_processes_refuse(self):
        legacy = ["python3", str(self.tmp / "legacy/handoff.py"), "send", "--session", "x"]
        self.fake_proc([(910001, legacy, [str(self.log)])])
        self.refused("other-helper-invocation-in-flight", "log-held-by-other-process")
        self.fake_proc([(910002, legacy, [])])
        self.refused("other-helper-invocation-in-flight")
        upgraded = self.tmp / "replaced/handoff.py"
        upgraded.parent.mkdir()
        upgraded.write_bytes(HELPER_BYTES)
        self.fake_proc([(910003, ["python3", str(upgraded), "send"], [])])
        evidence = self.refused("other-helper-invocation-in-flight")
        self.assertEqual(evidence["helper_invocations"][0]["source_path_sha256_untrusted"], SELF_SHA)

    def test_t14_renamed_relative_and_module_helper_invocations_refuse(self):
        scripts = self.root / Path(handoff.HELPER_CANDIDATES[0]).parent
        alias = self.tmp / "coordinator-tool"
        os.symlink(scripts / "handoff.py", alias)
        cases = [
            (910011, ["python3", str(alias), "send"], []),
            (910012, ["python3.12", str(scripts / "maintenance.py")], []),
            (910013, ["python3", "handoff_copy.py", "send"], [], str(scripts)),
            (910014, ["python3", "-m", "handoff", "send"], [], str(scripts)),
        ]
        for case in cases:
            self.fake_proc([case])
            evidence = self.refused("other-helper-invocation-in-flight")
            self.assertEqual([i["pid"] for i in evidence["helper_invocations"]], [case[0]])
        self.fake_proc([(910015, ["less", str(scripts / "notes.txt")], [])])
        self.assertEqual(self.dry(self.root)["prerequisite"]["helper_invocations"], [])

    @contextlib.contextmanager
    def denied(self, target, boundary):
        """Inject PermissionError at one exact /proc boundary for `target` only. Unlike chmod(0), this is
        deterministic as a normal user and as root (root can read mode-0 files)."""
        target = Path(target)

        def matches(path):
            try:
                return Path(os.fsdecode(os.fspath(path))) == target
            except TypeError:
                return False

        owner = Path if boundary in ("read_text", "read_bytes", "iterdir") else handoff.os
        real = getattr(owner, boundary)

        def fake(path, *args, **kwargs):
            if matches(path):
                raise PermissionError(13, "Permission denied (injected)", str(target))
            return real(path, *args, **kwargs)
        with patch.object(owner, boundary, fake):
            yield

    def test_t14_uninspectable_processes_refuse(self):
        python = ["python3", "/opt/tool.py"]
        cases = [  # (pid, argv, (action, entry name), expected reason[, cwd])
            (910021, python, ("unlink", "status"), "status-unreadable"),
            (910028, python, ("read_text", "status"), "status-unreadable"),
            (910022, python, ("write", "status"), "status-malformed"),
            (910023, python, ("read_bytes", "cmdline"), "cmdline-unreadable"),
            (910024, python, ("empty", "cmdline"), "cmdline-empty"),
            (910025, ["python3", "tool.py"], (None, None), "cwd-unreadable", None),
            (910029, ["python3", "tool.py"], ("readlink", "cwd"), "cwd-unreadable"),
            (910026, python, ("iterdir", "fd"), "fd-unlistable"),
            (910027, ["sh", "/opt/handoff.py"], ("iterdir", "fd"), "fd-unlistable"),
            (910030, python, ("stat", "fd/3"), "fd-unstatable"),
        ]
        for pid, argv, (action, name), reason, *cwd in cases:
            proc = self.fake_proc([(pid, argv, [str(self.tmp)] if action == "stat" else [], *cwd)])
            entry = proc / str(pid)
            if action == "unlink":
                (entry / name).unlink()
            elif action == "write":
                (entry / name).write_text("Name:\tpython3\n")
            elif action == "empty":
                (entry / name).write_bytes(b"")
            injected = (self.denied(entry / name, action)
                        if action in ("read_text", "read_bytes", "iterdir", "readlink", "stat")
                        else contextlib.nullcontext())
            with injected:
                evidence = self.refused("uninspectable-process")
            self.assertIn(reason, [u["reason"] for u in evidence["uninspectable_processes"] if u["pid"] == pid],
                          (pid, action, evidence["uninspectable_processes"]))

    def test_t14_nonhelper_with_unlistable_descriptors_is_evidence_only(self):
        proc = self.fake_proc([(910031, ["(sd-pam)"], [])])
        expected = [{"pid": 910031, "comm": "(sd-pam)"}]
        with self.denied(proc / "910031" / "fd", "iterdir"):
            evidence = self.dry(self.root)["prerequisite"]
            self.assertEqual((evidence["satisfied"], evidence["uninspectable_nonhelper"]), (True, expected))
            self.assertEqual(evidence["uninspectable_processes"], [], "a non-helper is evidence only, never a refusal")
            result = self.apply(self.root, self.selection)
        self.assertEqual((result["status"], result["prerequisite"]["uninspectable_nonhelper"]),
                         ("compaction-applied", expected))
        archive = handoff.state_root() / handoff.repo_key(self.root) / "archive" / result["archive"]
        manifest = json.loads(archive.with_name(archive.name.replace(".jsonl", ".manifest.json")).read_bytes())
        self.assertEqual(manifest["prerequisite"]["uninspectable_nonhelper"], expected)

    def test_t14_zombie_and_vanished_entries_are_skipped(self):
        uid = os.getuid()
        proc = self.fake_proc([(910032, ["python3"], [])])
        (proc / "910032" / "status").write_text(f"Name:\tpython3\nState:\tZ (zombie)\nUid:\t{uid}\t{uid}\t{uid}\t{uid}\n")
        (proc / "910032" / "cmdline").write_bytes(b"")
        self.assertEqual(self.dry(self.root)["prerequisite"]["failures"], [])

        class Vanishing:
            """Lists one more entry than exists: a process that exits between listing and inspection."""
            def iterdir(self):
                return list(proc.iterdir()) + [proc / "910033"]

            def __truediv__(self, other):
                return proc / other
        self.patch(handoff, "PROC", Vanishing())
        evidence = self.dry(self.root)["prerequisite"]
        self.assertEqual((evidence["failures"], evidence["uninspectable_processes"]), ([], []))

    def test_t14_non_dumpable_same_user_python_is_scanned_by_status_uid(self):
        proc = self.fake_proc([(910034, ["python3", "/opt/tool.py"], [])])
        acting = os.getuid() + 1
        (proc / "910034" / "status").write_text(f"Name:\tpython3\nState:\tS (sleeping)\n"
                                                f"Uid:\t{acting}\t{acting}\t{acting}\t{acting}\n")
        # As uid `acting`, this entry's directory is owned by someone else, as for a non-dumpable process,
        # while its status real Uid is ours: it must be scanned and refuse.
        with self.denied(proc / "910034" / "fd", "iterdir"):
            with patch.object(handoff.os, "getuid", return_value=acting):
                _, _, uninspectable, _ = handoff.process_scan((0, 0), set())
            self.assertEqual(uninspectable, [{"pid": 910034, "comm": "python3", "reason": "fd-unlistable"}])
            _, _, uninspectable, _ = handoff.process_scan((0, 0), set())
            self.assertEqual(uninspectable, [], "another user's process is outside the same-user scan")

    def test_t14_incomplete_pid_visibility_refuses(self):
        self.fake_proc(cli_pids=(VISIBLE_PID, 910099))
        self.refused("pid-visibility-unproven")
        self.fake_proc(cli_pids=())
        self.refused("pid-visibility-unproven")
        self.fake_proc()
        self.cli.fail = True
        self.refused("pid-visibility-unproven")

    def test_t14_invocation_path_must_be_upgraded(self):
        self.fake_proc()
        self.refused("invocation-path-not-upgraded", writer=hashlib.sha256(b"other").hexdigest())
        legacy_root = self.make_repo("legacy", helper=b"# legacy helper\n")
        self.write_log(legacy_root, self.data)
        self.selection = self.dry(legacy_root)["selection_sha256"]
        self.refused("invocation-path-not-upgraded", root=legacy_root)

    def commit_alias(self, root):
        """The source-maintainer discovery alias: a committed relative symlink to the canonical bundle."""
        alias = root / ".agents" / "skills" / "restruct-sdlc"
        alias.parent.mkdir(parents=True)
        os.symlink("../../distributions/skills/restruct-sdlc", alias)
        self.commit(root)
        return alias

    def relocate_helper(self, name, destination):
        root = self.make_repo(name)
        (root / destination).parent.mkdir(parents=True, exist_ok=True)
        run_git("-C", root, "mv", handoff.HELPER_CANDIDATES[0], destination)
        self.commit(root)
        self.write_log(root, self.data)
        return root

    def test_helper_relative_prefers_tracked_canonical_path(self):
        """AC6: the source-maintainer path wins (a committed alias symlink is never a tree path), the adopter path is
        second, and a repository tracking neither (the retired standalone path) refuses."""
        self.fake_proc()
        self.commit_alias(self.root)
        self.assertEqual(handoff.helper_relative(self.root), handoff.HELPER_CANDIDATES[0])
        evidence = self.dry(self.root)["prerequisite"]
        self.assertEqual((evidence["satisfied"], evidence["committed_sha256"], evidence["working_sha256"]),
                         (True, SELF_SHA, SELF_SHA))
        adopter = self.relocate_helper("adopter", handoff.HELPER_CANDIDATES[1])
        self.assertEqual(handoff.helper_relative(adopter), handoff.HELPER_CANDIDATES[1])
        evidence = self.dry(adopter)["prerequisite"]
        self.assertEqual((evidence["satisfied"], evidence["committed_sha256"], evidence["working_sha256"]),
                         (True, SELF_SHA, SELF_SHA))
        retired = self.relocate_helper("retired", ".agents/skills/claude-session-handoff/scripts/handoff.py")
        self.assertIsNone(handoff.helper_relative(retired))
        self.selection = self.dry(retired)["selection_sha256"]
        evidence = self.refused("invocation-path-not-upgraded", root=retired)
        self.assertEqual((evidence["committed_sha256"], evidence["working_sha256"]), (None, None))

    def test_alias_invocation_resolves_same_repository(self):
        """AC6: a helper invoked through the alias path is recognized by helper_dirs/process_scan and resolves to the
        same Git root; another repository's directories are never selected."""
        alias = self.commit_alias(self.root)
        other = self.make_repo("other")
        alias_scripts = alias / "claude-session-handoff" / "scripts"
        canonical = (self.root / handoff.HELPER_CANDIDATES[0]).parent.resolve()
        self.assertEqual(alias_scripts.resolve(), canonical)
        self.assertEqual(handoff.repo_root(alias_scripts), self.root)
        directories = handoff.helper_dirs(self.root)
        self.assertIn(canonical, directories)
        self.assertNotIn((other / handoff.HELPER_CANDIDATES[0]).parent.resolve(), directories)
        for pid, argv in ((910021, ["python3", str(alias_scripts / "handoff.py"), "send"]),
                          (910022, ["python3", str(alias_scripts / "maintenance.py")])):
            self.fake_proc([(pid, argv, [])])
            evidence = self.refused("other-helper-invocation-in-flight")
            self.assertEqual([i["pid"] for i in evidence["helper_invocations"]], [pid])

    def test_retired_standalone_helper_path_is_still_scanned(self):
        """Migration window: a watcher or helper still running from the retired standalone scripts directory (in a
        worktree or the user copy) is detected by process_scan and compact refuses as an in-flight helper."""
        retired = self.root / handoff.LEGACY_HELPER_RELATIVE
        retired.parent.mkdir(parents=True)
        (retired.parent / "watch.py").write_bytes(b"# retired watcher fixture\n")
        home_retired = (Path.home() / handoff.LEGACY_HELPER_RELATIVE).parent
        home_retired.mkdir(parents=True)
        directories = handoff.helper_dirs(self.root)
        self.assertIn(retired.parent.resolve(), directories)
        self.assertIn(home_retired.resolve(), directories)
        self.assertEqual(handoff.helper_relative(self.root), handoff.HELPER_CANDIDATES[0], "writer path unchanged")
        cases = [(910031, ["python3", str(retired.parent / "watch.py"), "--repo", str(self.root), "run"], []),
                 (910032, ["python3", "watch.py", "status"], [], str(retired.parent)),
                 (910033, ["python3.12", str(home_retired / "maintenance.py")], [])]
        for case in cases:
            self.fake_proc([case])
            evidence = self.refused("other-helper-invocation-in-flight")
            self.assertEqual([i["pid"] for i in evidence["helper_invocations"]], [case[0]])

    def test_t14_linked_worktree_is_refused(self):
        linked = self.worktree(self.root, "linked")
        (linked / LOG).write_bytes(self.data)
        for runner in (lambda: self.dry(linked), lambda: self.apply(linked, self.selection)):
            with self.assertRaises(handoff.Blocked) as caught:
                runner()
            self.assertEqual(str(caught.exception), "compact-requires-main-checkout")
        self.assertEqual((linked / LOG).read_bytes(), self.data)
        self.assertEqual(self.log.read_bytes(), self.data)

    def test_t14_real_process_holding_the_log_refuses(self):
        self.cli.rows = [self.cli_row(str(uuid.uuid4()), self.root, pid=os.getpid())]
        child = subprocess.Popen([sys.executable, "-c", "import sys, time; f = open(sys.argv[1], 'rb'); "
                                  "print('ready', flush=True); time.sleep(60)", str(self.log)],
                                 stdout=subprocess.PIPE, text=True)
        self.addCleanup(child.stdout.close)
        self.addCleanup(child.wait)
        self.addCleanup(child.kill)
        self.assertEqual(child.stdout.readline().strip(), "ready")
        evidence = self.refused("log-held-by-other-process")
        self.assertIn(child.pid, [h["pid"] for h in evidence["log_holders"]])


class T15Failures(Base):
    def setUp(self):
        super().setUp()
        self.root = self.make_repo()
        self.data, self.closed, self.kept = self.closure_log(str(uuid.uuid4()))
        self.log = self.write_log(self.root, self.data)
        self.selection = self.dry(self.root)["selection_sha256"]
        self.fake_proc()
        self.archive_dir = handoff.state_root() / handoff.repo_key(self.root) / "archive"

    def preserved(self):
        archived = set()
        for path in self.archive_dir.glob("*.jsonl") if self.archive_dir.exists() else []:
            archived |= ids_in(path.read_bytes())
        self.assertEqual(archived | ids_in(self.log.read_bytes()), {self.closed, self.kept})

    def rerun_is_safe(self):
        result = self.apply(self.root, self.selection)
        self.assertEqual(result["status"], "compaction-applied")
        self.assertEqual(ids_in(self.log.read_bytes()), {self.kept})
        self.preserved()

    def fails(self, exception, reason=None):
        with self.assertRaises(exception) as caught:
            self.apply(self.root, self.selection)
        if reason:
            self.assertEqual(str(caught.exception), reason)
        self.assertEqual(self.log.read_bytes(), self.data)
        self.assertFalse(list(self.log.parent.glob(".compaction-*")))
        self.preserved()

    def test_t15_archive_write_error(self):
        with patch.object(handoff, "write_new_private", side_effect=OSError("disk full")):
            self.fails(handoff.Blocked, "archive-write-failed")
        self.rerun_is_safe()

    def test_t15_archive_hash_mismatch(self):
        real = handoff.read_private
        with patch.object(handoff, "read_private", side_effect=lambda p: b"corrupt" if p.suffix == ".jsonl" else real(p)):
            self.fails(handoff.Blocked, "archive-verification-failed")
        self.rerun_is_safe()

    def test_t15_failed_replace(self):
        with patch.object(handoff, "replace_file", side_effect=OSError("replace")):
            self.fails(handoff.Blocked, "replace-failed")
        self.assertEqual(len(list(self.archive_dir.glob("*.jsonl"))), 1)
        self.rerun_is_safe()

    def test_t15_crash_between_archive_and_replace(self):
        with patch.object(handoff, "replace_log", side_effect=RuntimeError("simulated crash")):
            self.fails(RuntimeError)
        self.assertEqual((self.manifest()["pruned"], self.manifest()["prune_status"]), (False, "pending"))
        self.rerun_is_safe()

    def manifest(self):
        paths = sorted(self.archive_dir.glob("*.manifest.json"))
        return json.loads(paths[0].read_bytes())

    def test_t15_final_scan_refusal_after_archive(self):
        real, calls = handoff.prerequisite_evidence, []

        def evidence(*args):
            result = real(*args)
            calls.append(result)
            if len(calls) == 2:
                result = {**result, "satisfied": False, "failures": ["other-helper-invocation-in-flight"]}
            return result
        with patch.object(handoff, "prerequisite_evidence", side_effect=evidence):
            self.fails(handoff.Blocked, "legacy-writer-prerequisite-unmet")
        manifest = self.manifest()
        self.assertEqual((manifest["pruned"], manifest["prune_status"]), (False, "legacy-writer-prerequisite-unmet"))
        self.assertEqual(manifest["final_prerequisite"]["failures"], ["other-helper-invocation-in-flight"])
        self.rerun_is_safe()

    def test_t15_log_change_after_archive_refuses(self):
        real = handoff.write_archive
        late = str(uuid.uuid4())

        def archive_then_legacy_append(*args, **kwargs):
            outcome = real(*args, **kwargs)
            with self.log.open("ab") as stream:
                stream.write(record(late, str(uuid.uuid4()), "prepared-not-sent"))
            return outcome
        with patch.object(handoff, "write_archive", side_effect=archive_then_legacy_append):
            with self.assertRaises(handoff.Blocked) as caught:
                self.apply(self.root, self.selection)
        self.assertEqual(str(caught.exception), "log-changed-during-compaction")
        self.assertTrue(self.log.read_bytes().startswith(self.data))
        self.assertEqual(ids_in(self.log.read_bytes()), {self.closed, self.kept, late})
        self.assertEqual((self.manifest()["pruned"], self.manifest()["prune_status"]),
                         (False, "log-changed-during-compaction"))

    def test_t15_directory_fsync_failure_after_replace_reports_the_prune(self):
        real = handoff.fsync_dir

        def fsync(path):
            if Path(path) == self.log.parent:
                raise OSError("fsync")
            return real(path)
        with patch.object(handoff, "fsync_dir", side_effect=fsync):
            with self.assertRaises(handoff.Blocked) as caught:
                self.apply(self.root, self.selection)
        self.assertEqual(str(caught.exception), "pruned-directory-fsync-failed")
        self.assertTrue(caught.exception.details["pruned"])
        self.assertEqual(ids_in(self.log.read_bytes()), {self.kept})
        self.assertFalse(list(self.log.parent.glob(".compaction-*")))
        self.assertEqual((self.manifest()["pruned"], self.manifest()["prune_status"]),
                         (True, "pruned-directory-fsync-failed"))
        self.preserved()

    def test_t15_post_replace_paths_keep_pruned_true(self):
        with patch.object(handoff.os, "rmdir", side_effect=OSError("busy")):
            result = self.apply(self.root, self.selection)
        self.assertEqual(result["status"], "compaction-applied")
        self.assertEqual((self.manifest()["pruned"], self.manifest()["prune_status"]), (True, "pruned"))
        for leftover in self.log.parent.glob(".compaction-*"):
            leftover.rmdir()
        self.preserved()

    def test_t15_failed_verification_read_after_replace_reports_the_prune(self):
        real, calls = handoff.read_log, []

        def read(path):
            calls.append(path)
            if len(calls) == 4:
                raise handoff.Blocked("record-unsafe")
            return real(path)
        with patch.object(handoff, "read_log", side_effect=read):
            with self.assertRaises(handoff.Blocked) as caught:
                self.apply(self.root, self.selection)
        self.assertEqual((str(caught.exception), caught.exception.details["pruned"]), ("pruned-verification-failed", True))
        self.assertEqual((self.manifest()["pruned"], self.manifest()["prune_status"]), (True, "pruned-verification-failed"))
        self.assertEqual(ids_in(self.log.read_bytes()), {self.kept})
        self.preserved()

    def test_t15_manifest_update_failure_after_replace_reports_the_prune(self):
        with patch.object(handoff, "finalize_manifest", side_effect=handoff.Blocked("state-unsafe")):
            with self.assertRaises(handoff.Blocked) as caught:
                self.apply(self.root, self.selection)
        self.assertEqual((str(caught.exception), caught.exception.details["pruned"]), ("pruned-manifest-update-failed", True))
        self.assertEqual(ids_in(self.log.read_bytes()), {self.kept})
        self.preserved()

class TransportCompatibility(Base):
    def setUp(self):
        super().setUp()
        self.root = self.make_repo()
        self.session = str(uuid.uuid4())

    def test_new_version_keeps_protocol_gate_and_warns(self):
        self.register(self.session, self.root, version="9.9.9")
        self.pin(self.root, self.session)
        result = run_main("--repo", self.root, "status")
        self.assertEqual(result["status"], "pinned-live-verified")
        self.assertEqual(result["compatibility_warnings"], ["unbaselined-claude-version"])
        self.unregister(self.session)
        for protocol in (2, True):
            self.register(self.session, self.root, version="9.9.9", protocol=protocol)
            with self.assertRaises(handoff.Blocked):
                handoff.live_registration(self.registry, self.session, self.root)
            self.unregister(self.session)

    def test_stale_prewrite_connect_refreshes_same_session_once(self):
        from unittest.mock import MagicMock
        row = {"cwd": str(self.root), "sessionId": self.session}
        refreshed = {**row, "pid": 2, "version": "9.9.9"}
        client = MagicMock()
        with patch.object(handoff, "live_registration", side_effect=[row, refreshed]) as lookup, \
             patch.object(handoff, "connect_peer", side_effect=[FileNotFoundError(2, "stale"), (client, "fixture")]) as connect:
            self.assertEqual(handoff.send(self.registry, row, "authorized fixture", HANDOFF), refreshed)
        self.assertEqual(lookup.call_count, 2)
        self.assertTrue(all(call.args[1] == self.session for call in lookup.call_args_list))
        self.assertEqual(connect.call_count, 2)
        client.sendall.assert_called_once()
        frames = [json.loads(line) for line in client.sendall.call_args.args[0].splitlines()]
        self.assertEqual((frames[1]["session_id"], frames[1]["uuid"]), (self.session, HANDOFF))

    def test_connect_retry_is_bounded_and_never_retargets(self):
        row = {"cwd": str(self.root), "sessionId": self.session}
        with patch.object(handoff, "live_registration", return_value=row) as lookup, \
             patch.object(handoff, "connect_peer", side_effect=ConnectionRefusedError(111, "stale")) as connect:
            with self.assertRaises(OSError):
                handoff.send(self.registry, row, "fixture", HANDOFF)
        self.assertEqual((lookup.call_count, connect.call_count), (2, 2))

    def test_peer_mismatch_and_uncertain_write_are_never_retried(self):
        from unittest.mock import MagicMock
        row = {"cwd": str(self.root), "sessionId": self.session}
        with patch.object(handoff, "live_registration", return_value=row), \
             patch.object(handoff, "connect_peer", side_effect=handoff.Blocked("socket-peer-mismatch")) as connect:
            with self.assertRaises(handoff.Blocked):
                handoff.send(self.registry, row, "fixture", HANDOFF)
            connect.assert_called_once()
        client = MagicMock()
        client.sendall.side_effect = BrokenPipeError(32, "uncertain partial send")
        with patch.object(handoff, "live_registration", return_value=row) as lookup, \
             patch.object(handoff, "connect_peer", return_value=(client, "fixture")) as connect:
            with self.assertRaises(OSError):
                handoff.send(self.registry, row, "fixture", HANDOFF)
        self.assertEqual((lookup.call_count, connect.call_count, client.sendall.call_count), (1, 1, 1))

    def test_real_peer_credentials_checked_before_any_bytes(self):
        import socket
        address = str(self.tmp / "peer.sock")
        with socket.socket(socket.AF_UNIX, socket.SOCK_STREAM) as listener:
            listener.bind(address)
            listener.listen(1)
            listener.settimeout(2)
            pid = os.getpid() + 1
            key = self.registry / f"{pid}.{handoff.digest(address.encode())}.key"
            key.write_text(json.dumps({"peerToken": "fixture"}))
            key.chmod(0o600)
            with self.assertRaises(handoff.Blocked) as caught:
                handoff.connect_peer(self.registry, {"pid": pid, "messagingSocketPath": address})
            self.assertEqual(str(caught.exception), "socket-peer-mismatch")
            peer, _ = listener.accept()
            with peer:
                peer.settimeout(2)
                self.assertEqual(peer.recv(4096), b"")

if __name__ == "__main__":
    unittest.main()
