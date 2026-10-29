#!/usr/bin/env python3
"""Focused observer checks with fake coordinator-status, transcript and Codex CLI boundaries; no live messages.

Run: python3 -B -m unittest discover -s scripts -p 'test_watch.py' -v
"""
import fcntl
import json
import os
from pathlib import Path
import shutil
import subprocess
import sys
import tempfile
import time
import unittest
from unittest.mock import patch
import uuid

import handoff
import watch

COORD = "00000000-0000-4000-8000-0000000000c1"
OTHER = "00000000-0000-4000-8000-0000000000c2"
SUPER = "00000000-0000-4000-8000-0000000000d1"


class FakeCodex:
    def __init__(self):
        self.calls, self.queue_rc, self.cli, self.daemon, self.raise_timeout = [], 0, "0.160.0", "0.162.1", False
        self.help_text = "Usage: codex queue [OPTIONS] --thread <THREAD> --message <TEXT>"

    def __call__(self, arguments, timeout):
        self.calls.append(list(arguments))
        if arguments == ["queue", "--help"]:
            return subprocess.CompletedProcess(arguments, 0, self.help_text, "")
        if arguments[:1] == ["--version"]:
            return subprocess.CompletedProcess(arguments, 0, f"codex-cli {self.cli}\n", "")
        if arguments[:3] == ["app-server", "daemon", "version"]:
            return subprocess.CompletedProcess(arguments, 0, json.dumps({"appServerVersion": self.daemon}), "")
        if self.raise_timeout:
            raise subprocess.TimeoutExpired(arguments, timeout)
        return subprocess.CompletedProcess(arguments, self.queue_rc, "", "")

    def queues(self):
        return [c for c in self.calls if c[:2] == ["queue", "--thread"]]


class Base(unittest.TestCase):
    def setUp(self):
        self.tmp = Path(tempfile.mkdtemp(prefix="watch-test-"))
        self.addCleanup(shutil.rmtree, self.tmp, True)
        env = patch.dict(os.environ, {"HOME": str(self.tmp / "home"), "GIT_CONFIG_GLOBAL": os.devnull,
                                      "GIT_CONFIG_NOSYSTEM": "1"})
        env.start()
        self.addCleanup(env.stop)
        (self.tmp / "home").mkdir()
        self.repo = self.tmp / "repo"
        subprocess.run(["git", "init", "-q", "-b", "main", str(self.repo)], check=True)
        (self.repo / "management").mkdir()
        (self.repo / "management/PROGRESS.md").write_text("# Progress\n")
        self.root = handoff.repo_root(self.repo)
        self.projects = self.tmp / "home" / ".claude" / "projects" / "fixture"
        self.projects.mkdir(parents=True)
        self.transcript = self.projects / f"{COORD}.jsonl"
        self.transcript.write_text(self.row("assistant", "Earlier: ready for review (history, not replayed).") + "\n")
        self.status = {"status": "pinned-live-verified", "pin": {"session": COORD}, "registration_status": "busy"}
        self.codex = FakeCodex()
        p1 = patch.object(watch, "coordinator_status", lambda root: json.loads(json.dumps(self.status)))
        p2 = patch.object(watch, "run_codex", self.codex)
        for p in (p1, p2):
            p.start()
            self.addCleanup(p.stop)

    @staticmethod
    def row(kind, text, tool=None, sidechain=False):
        content = [{"type": "text", "text": text}]
        if tool:
            content.append({"type": "tool_use", "name": tool, "input": {}})
        return json.dumps({"type": kind, "uuid": str(uuid.uuid4()), "timestamp": "2026-10-12T00:00:00Z",
                           "isSidechain": sidechain, "message": {"role": kind, "content": content}})

    def append(self, *rows):
        with self.transcript.open("a") as stream:
            for row in rows:
                stream.write(row + "\n")

    def bind(self, *extra):
        return watch.main(["--repo", str(self.repo), "bind", "--supervisor", SUPER, *extra])

    def poll(self):
        return watch.main(["--repo", str(self.repo), "poll-once"])

    def state(self):
        return json.loads(handoff.read_private(watch.observer_dir(self.root) / "state.json"))


class Polling(Base):
    def test_idle_polling_makes_zero_queue_or_cli_calls(self):
        self.bind()
        for _ in range(3):
            self.assertEqual(self.poll()["queue_calls"], 0)
        self.assertEqual(self.codex.calls, [])

    def test_milestone_notifies_once_and_history_is_not_replayed(self):
        self.bind()
        self.poll()
        self.append(self.row("assistant", "Result submitted for Codex review."))
        result = self.poll()
        self.assertEqual((result["status"], result["queue_calls"]), ("polled", 1))
        call = self.codex.queues()[0]
        self.assertEqual(call[:3], ["queue", "--thread", SUPER])
        self.assertIn("[observer-notice:", call[4])
        self.assertIn("review-ready", call[4])
        self.assertIn("no approval or task authority", call[4])
        self.assertEqual(self.poll()["queue_calls"], 0)

    def test_restart_keeps_dedup_and_offsets(self):
        self.bind()
        self.poll()
        self.append(self.row("assistant", "Work is complete; handback sent."))
        self.poll()
        self.assertEqual(len(self.codex.queues()), 1)
        offsets = self.state()["offsets"]
        self.assertEqual(self.poll()["queue_calls"], 0)  # a fresh process reads the same private state
        self.assertEqual(self.state()["offsets"], offsets)
        self.append(self.row("assistant", "I am blocked on a missing input."))
        self.assertEqual(self.poll()["queue_calls"], 1)

    def test_partial_line_waits_until_complete(self):
        self.bind()
        self.poll()
        line = self.row("assistant", "Plan is review-ready.")
        with self.transcript.open("a") as stream:
            stream.write(line[:25])
        self.assertEqual(self.poll()["queue_calls"], 0)
        with self.transcript.open("a") as stream:
            stream.write(line[25:] + "\n")
        self.assertEqual(self.poll()["queue_calls"], 1)

    def test_rotated_and_truncated_transcripts(self):
        self.bind()
        self.poll()
        first = self.row("assistant", "Escalating a scope question.")
        self.append(first)
        self.poll()
        rotated = self.transcript.with_suffix(".new")
        rotated.write_text(first + "\n" + self.row("assistant", "Completed the attempt; handback ready.") + "\n")
        os.replace(rotated, self.transcript)
        self.assertEqual(self.poll()["queue_calls"], 1)  # only the new row; the old key is deduplicated
        self.transcript.write_text(self.row("assistant", "Result is review-ready again after fixes.") + "\n")
        self.assertEqual(self.poll()["queue_calls"], 1)

    def test_question_tool_sidechain_and_non_milestone_rows(self):
        self.bind()
        self.poll()
        self.append(self.row("assistant", "Choosing an option.", tool="AskUserQuestion"),
                    self.row("assistant", "Subagent says blocked.", sidechain=True),
                    self.row("assistant", "Running tests now."))
        self.assertEqual(self.poll()["queue_calls"], 1)
        self.assertIn("(question", self.codex.queues()[0][4])

    def test_waiting_requires_native_inspection_and_notifies_on_transition(self):
        self.bind()
        self.poll()
        self.status["registration_status"] = "waiting"
        self.assertEqual(self.poll()["queue_calls"], 1)
        self.assertIn("native prompt was not inspected", self.codex.queues()[0][4])
        self.assertEqual(self.poll()["queue_calls"], 0)
        self.status["registration_status"] = "busy"
        self.poll()
        self.status["registration_status"] = "waiting"
        self.assertEqual(self.poll()["queue_calls"], 1)


class NotificationSources(Base):
    def test_working_marker_suppresses_history_and_recovery_prose(self):
        self.bind()
        self.poll()
        self.append(self.row("assistant", "Earlier the gate failed. Rerun the failed batch, then continue.\n"
                             "COORDINATOR_STATUS: working | G008 | Implementer resumed."))
        self.assertEqual(self.poll()["queue_calls"], 0)
        self.assertEqual(self.codex.calls, [])

    def test_mixed_task_working_summary_cannot_hide_review_handoff(self):
        self.bind()
        self.poll()
        self.append(self.row("assistant", "G013 gate is running. G014 Plan r7 needs re-review.\n"
                             "COORDINATOR_STATUS: working | G013 | Staging gate running; "
                             "G014 r7 ready for Codex re-review."))
        self.assertEqual(self.poll()["queue_calls"], 1)
        notice = self.codex.queues()[0][4]
        self.assertIn("milestone (review-ready", notice)
        self.assertIn("G014 r7 ready for Codex re-review", notice)
        self.assertEqual(self.poll()["queue_calls"], 0)

    def test_working_handoff_is_classified_before_notice_truncation(self):
        self.bind()
        self.poll()
        tail = "G014 ready for Codex re-review."
        summary = "x" * (240 - len(tail) - 1) + " " + tail
        self.append(self.row("assistant", f"COORDINATOR_STATUS: working | {'G' * 64} | {summary}"))
        self.assertEqual(self.poll()["queue_calls"], 1)
        self.assertIn("milestone (review-ready", self.codex.queues()[0][4])

    def test_oversized_attention_summary_is_delivered_once_with_bounded_excerpt(self):
        self.bind()
        self.poll()
        missed = ("Scope gap 2: audit_integration_test.go AT-INV requires an observed branding_changed producer "
                  "(outside 43-pattern list); request exact path for one guarded UpdateBranding call; "
                  "WIP aaaa1111 (Result bbbb2222); fresh successor after disposition")
        self.assertEqual(len(missed), 241)
        for category in ("decision", "question", "review-ready", "blocked", "failed", "completed"):
            for summary in ("x" * 240, missed, "x" * 8192):
                with self.subTest(category=category, length=len(summary)):
                    row = self.row("assistant", f"COORDINATOR_STATUS: {category} | G017 | {summary}")
                    found = watch.milestone(json.loads(row))
                    self.assertIsNotNone(found)
                    self.assertEqual(found[0], category)
                    self.assertLessEqual(len(found[1]), 240)
                    self.append(row)
                    self.assertEqual(self.poll()["queue_calls"], 1)
                    self.assertIn(f"milestone ({category}", self.codex.queues()[-1][4])
                    self.assertEqual(self.poll()["queue_calls"], 0)

    def test_owner_decision_notifies_while_work_continues(self):
        self.bind()
        self.poll()
        marker = ("COORDINATOR_STATUS: decision | G066 | Owner requires N<=TOTAL; "
                  "ref owner-message-1; source acceptance must follow the fixed head.")
        row = self.row("assistant", "The existing writer is still working.\n" + marker)
        self.append(row)
        self.assertEqual(self.poll()["queue_calls"], 1)
        notice = self.codex.queues()[-1][4]
        self.assertIn("milestone (decision", notice)
        self.assertIn("owner-message-1", notice)
        self.assertIn("no approval or task authority", notice)
        self.assertEqual(self.poll()["queue_calls"], 0)

    def test_decision_marker_keeps_source_and_echo_boundaries(self):
        marker = "COORDINATOR_STATUS: decision | G066 | Owner changed scope; ref owner-message-1."
        for row in (self.row("user", marker), self.row("assistant", marker, sidechain=True),
                    self.row("assistant", "[handoff-id: echo]\n" + marker),
                    self.row("assistant", "> " + marker),
                    self.row("assistant", "```text\n" + marker + "\n```")):
            with self.subTest(row=row):
                self.assertIsNone(watch.milestone(json.loads(row)))

    def test_oversized_working_summary_classifies_tail_before_truncation(self):
        self.bind()
        self.poll()
        for tail, category in (("G018 ready for Codex re-review.", "review-ready"),
                               ("G018 owner approval pending.", "question")):
            summary = "x" * 8192 + " " + tail
            self.append(self.row("assistant", f"COORDINATOR_STATUS: working | G017 | {summary}"))
            self.assertEqual(self.poll()["queue_calls"], 1)
            self.assertIn(f"milestone ({category}", self.codex.queues()[-1][4])
            self.assertEqual(self.poll()["queue_calls"], 0)
        self.append(self.row("assistant", "Earlier the gate failed.\n"
                             "COORDINATOR_STATUS: working | G017 | " + "x" * 8192))
        self.assertEqual(self.poll()["queue_calls"], 0)

    def test_working_summary_reports_pending_decision_or_review(self):
        self.bind()
        self.poll()
        for index, (summary, category) in enumerate((
                ("G014 owner approval pending.", "question"),
                ("G014 awaiting supervisor decision.", "question"),
                ("G014 awaiting Codex review.", "review-ready"))):
            with self.subTest(summary=summary):
                self.append(self.row("assistant", f"COORDINATOR_STATUS: working | G013 | {summary}"))
                self.assertEqual(self.poll()["queue_calls"], 1)
                self.assertIn(f"milestone ({category}", self.codex.queues()[index][4])

    def test_working_handoff_fallback_keeps_negations_and_source_exclusions(self):
        self.bind()
        self.poll()
        marker = "COORDINATOR_STATUS: working | G013 | G014 ready for Codex re-review."
        for row in (self.row("assistant", "COORDINATOR_STATUS: working | G013 | "
                                           "No approval pending; not ready for Codex review."),
                    self.row("user", marker),
                    self.row("assistant", marker, sidechain=True),
                    self.row("assistant", "[observer-notice: echo]\n" + marker),
                    self.row("assistant", "[handoff-id: echo]\n" + marker)):
            self.append(row)
            self.assertEqual(self.poll()["queue_calls"], 0)
        self.assertEqual(self.codex.calls, [])

    def test_ready_marker_selects_the_handoff_instead_of_failed_batch_prose(self):
        self.bind()
        self.poll()
        self.append(self.row("assistant", "Rerun the failed batch after review.\n"
                             "COORDINATOR_STATUS: review-ready | G008 | Result ready for review."))
        self.assertEqual(self.poll()["queue_calls"], 1)
        self.assertIn("milestone (review-ready", self.codex.queues()[0][4])
        self.assertIn("G008: Result ready for review.", self.codex.queues()[0][4])
        self.assertEqual(self.poll()["queue_calls"], 0)

    def test_explicit_attention_classes_notify_once(self):
        self.bind()
        self.poll()
        for category in ("review-ready", "blocked", "question", "failed", "completed"):
            with self.subTest(category=category):
                self.append(self.row("assistant", f"COORDINATOR_STATUS: {category} | OS001 | Check records."))
                self.assertEqual(self.poll()["queue_calls"], 1)
                self.assertIn(f"milestone ({category}", self.codex.queues()[-1][4])
                self.assertIn("no approval or task authority", self.codex.queues()[-1][4])
                self.assertEqual(self.poll()["queue_calls"], 0)

    def test_invalid_marker_does_not_suppress_a_real_prose_failure(self):
        self.bind()
        self.poll()
        for marker in ("COORDINATOR_STATUS: unknown | G008 | Check records.",
                       "COORDINATOR_STATUS: working | ../G008 | Check records.",
                       "COORDINATOR_STATUS: working | G008 | ",
                       "COORDINATOR_STATUS: working | G008 | invalid | separator"):
            self.append(self.row("assistant", "The staging gate failed.\n" + marker))
            self.assertEqual(self.poll()["queue_calls"], 1)
            self.assertIn("milestone (failed", self.codex.queues()[-1][4])

    def test_incoming_sidechain_and_echo_markers_never_notify(self):
        self.bind()
        self.poll()
        marker = "COORDINATOR_STATUS: review-ready | G008 | Result ready."
        self.append(self.row("user", marker), self.row("assistant", marker, sidechain=True),
                    self.row("assistant", "[observer-notice: example]\n" + marker),
                    self.row("assistant", "[handoff-id: example]\n" + marker))
        self.assertEqual(self.poll()["queue_calls"], 0)
        self.assertEqual(self.codex.calls, [])

    def test_nonterminal_quoted_and_fenced_markers_never_notify(self):
        self.bind()
        self.poll()
        marker = "COORDINATOR_STATUS: review-ready | G008 | Result ready."
        for text in (marker + "\nThis was an example.", "> " + marker,
                     "```text\n" + marker + "\n```", "```text\n" + marker,
                     "~~~text\n" + marker):
            self.append(self.row("assistant", text))
            self.assertEqual(self.poll()["queue_calls"], 0)

    def test_final_marker_after_closed_fence_is_recognized(self):
        self.bind()
        self.poll()
        self.append(self.row("assistant", "```text\nexample\n```\n"
                             "COORDINATOR_STATUS: question | G008 | Choose the next action."))
        self.assertEqual(self.poll()["queue_calls"], 1)
        self.assertIn("milestone (question", self.codex.queues()[0][4])

    def test_native_question_is_not_hidden_by_working_marker(self):
        self.bind()
        self.poll()
        self.append(self.row("assistant", "COORDINATOR_STATUS: working | G008 | Preparing.",
                             tool="AskUserQuestion"))
        self.assertEqual(self.poll()["queue_calls"], 1)
        self.assertIn("asked a native question", self.codex.queues()[0][4])

    def test_staging_failure_report_notifies_once(self):
        self.bind()
        self.poll()
        self.append(self.row("assistant", "The staging gate failed, so final acceptance must wait.\n"
                             "140 tests passed and 1 failed.\n**Decision needed (supervisor or owner):** retry or investigate."))
        self.assertEqual(self.poll()["queue_calls"], 1)
        self.assertIn("milestone (failed", self.codex.queues()[0][4])
        self.assertEqual(self.poll()["queue_calls"], 0)

    def test_absent_expected_or_resolved_failures_do_not_notify(self):
        self.bind()
        self.poll()
        for text in ("No failures were found.", "0 unexpected failures.", "4 expected failures before the fix.",
                     "The failure is resolved.", "The failed gate now passed.", "The **failure** was already fixed."):
            with self.subTest(text=text):
                self.append(self.row("assistant", text))
                self.assertEqual(self.poll()["queue_calls"], 0)
        self.assertEqual(self.codex.queues(), [])

    def test_resolved_failure_does_not_hide_a_new_failure(self):
        self.bind()
        self.poll()
        self.append(self.row("assistant", "The failure is resolved; a different job failed."))
        self.assertEqual(self.poll()["queue_calls"], 1)
        self.assertIn("milestone (failed", self.codex.queues()[0][4])

    def test_conditional_failure_rules_do_not_notify(self):
        self.bind()
        self.poll()
        for text in ("Any failure rolls back steps 3–5 together.",
                     "Any failure stops it and comes back to me.",
                     "On failure we stop.", "Any failure would stop execution."):
            with self.subTest(text=text):
                self.append(self.row("assistant", text))
                self.assertEqual(self.poll()["queue_calls"], 0)
        self.assertEqual(self.codex.queues(), [])

    def test_conditional_rule_does_not_hide_review_ready_headline(self):
        self.bind()
        self.poll()
        self.append(self.row("assistant", "**G008 Plan r3 is ready for Codex review.**\n"
                             "Any failure rolls back steps 3–5 together."))
        self.assertEqual(self.poll()["queue_calls"], 1)
        self.assertIn("milestone (review-ready", self.codex.queues()[0][4])
        self.assertIn("G008 Plan r3", self.codex.queues()[0][4])
        self.assertEqual(self.poll()["queue_calls"], 0)

    def test_conditional_rule_does_not_hide_real_failure(self):
        self.bind()
        self.poll()
        self.append(self.row("assistant", "Any failure stops it; the staging gate failed."))
        self.assertEqual(self.poll()["queue_calls"], 1)
        self.assertIn("milestone (failed", self.codex.queues()[0][4])

    def test_mechanical_binding_ready_notifies_once(self):
        self.bind()
        self.poll()
        for text in ("G008 execution binding proposed and ready for Codex's mechanical verification.",
                     "The binding is ready for Codex’s mechanical confirmation.",
                     "The binding is ready for supervisor mechanical confirmation."):
            with self.subTest(text=text):
                self.append(self.row("assistant", text))
                self.assertEqual(self.poll()["queue_calls"], 1)
                self.assertIn("milestone (review-ready", self.codex.queues()[-1][4])
                self.assertEqual(self.poll()["queue_calls"], 0)

    def test_decision_and_approval_requests_notify_but_negated_requests_do_not(self):
        self.bind()
        self.poll()
        for text in ("No decision needed.", "No approval is required."):
            self.append(self.row("assistant", text))
            self.assertEqual(self.poll()["queue_calls"], 0)
        for text in ("**Decision needed (supervisor or owner):** choose the recovery.", "Approval is required."):
            self.append(self.row("assistant", text))
            self.assertEqual(self.poll()["queue_calls"], 1)
            self.assertIn("milestone (question", self.codex.queues()[-1][4])

    def test_resolved_and_negated_blockers_do_not_notify(self):
        self.bind()
        self.poll()
        for text in ("The permission blocker is resolved.", "No blocker is known.",
                     "The action is no longer blocked.", "The blocker was already cleared."):
            with self.subTest(text=text):
                self.append(self.row("assistant", text))
                self.assertEqual(self.poll()["queue_calls"], 0)
        self.assertEqual(self.codex.queues(), [])

    def test_resolved_blocker_does_not_hide_another_active_blocker(self):
        self.bind()
        self.poll()
        self.append(self.row("assistant", "The permission blocker is resolved; currently blocked on credentials."))
        self.assertEqual(self.poll()["queue_calls"], 1)
        self.assertIn("milestone (blocked", self.codex.queues()[0][4])

    def test_codex_repository_edits_cause_zero_queue_calls(self):
        self.bind()
        self.poll()
        (self.repo / "management/PROGRESS.md").write_text("# Progress\n\nG002 review-ready; blocked; handback.\n")
        (self.repo / "management/review.md").write_text("Outcome: ACCEPT; ready for review\n")
        self.assertEqual(self.poll()["queue_calls"], 0)
        self.assertEqual(self.codex.calls, [])

    def test_incoming_and_echoed_handoffs_cause_zero_queue_calls(self):
        self.bind()
        self.poll()
        incoming = "Codex correction: Result is review-ready? blocked? [handoff-id: 00000000-0000-4000-8000-0000000000e1]"
        self.append(self.row("user", incoming),
                    json.dumps({"type": "queue-operation", "operation": "enqueue", "content": incoming}),
                    self.row("assistant", "Received: " + incoming))
        self.assertEqual(self.poll()["queue_calls"], 0)

    def test_notice_echo_does_not_retrigger_then_genuine_milestone_notifies_once(self):
        self.bind()
        self.poll()
        self.append(self.row("assistant", "Implementation submitted for Codex review."))
        self.poll()
        notice = self.codex.queues()[0][4]
        self.append(self.row("user", "Supervisor replied to " + notice),
                    self.row("assistant", "Acknowledged " + notice + " (review-ready, handback)."))
        self.assertEqual(self.poll()["queue_calls"], 0)
        self.append(self.row("assistant", "Correction complete: work is complete and handed back."))
        self.assertEqual(self.poll()["queue_calls"], 1)
        self.assertEqual(len(self.codex.queues()), 2)

    def test_task_closeout_keeps_monitoring_and_dedup(self):
        self.bind()
        self.poll()
        self.append(self.row("assistant", "G001 handback sent."))
        self.poll()
        keys = len(self.state()["keys"])
        worktree = self.repo / ".claude" / "worktrees" / "g001-agent-x-impl"
        worktree.mkdir(parents=True)
        shutil.rmtree(worktree)                                     # task worktree cleanup
        (self.repo / "management/SESSION-HANDOFFS.jsonl").write_text("")  # delivery log compaction
        self.assertEqual(len(self.state()["keys"]), keys)
        self.assertEqual(self.poll()["queue_calls"], 0)
        self.append(self.row("assistant", "G002 plan is review-ready."))
        self.assertEqual(self.poll()["queue_calls"], 1)


class Identity(Base):
    def test_unknown_identity_suspends_until_explicit_rebind(self):
        self.bind()
        self.poll()
        self.status = {"status": "discovery-unknown"}
        self.append(self.row("assistant", "Ready for review."))
        result = self.poll()
        self.assertEqual((result["status"], result["reason"]), ("suspended", "coordinator-discovery-unknown"))
        self.status = {"status": "pinned-live-verified", "pin": {"session": COORD}, "registration_status": "busy"}
        self.assertEqual(self.poll()["status"], "suspended")
        self.assertEqual(self.codex.calls, [])
        self.bind("--replace")
        self.assertEqual(self.poll()["status"], "polled")

    def test_changed_coordinator_never_retargets(self):
        self.bind()
        self.poll()
        self.status["pin"]["session"] = OTHER
        result = self.poll()
        self.assertEqual(result["reason"], "coordinator-changed-rebind-required")
        self.assertEqual(self.codex.calls, [])
        with self.assertRaises(handoff.Blocked):
            self.bind()  # rebinding needs an explicit --replace

    def test_retired_session_blocks_delivery(self):
        self.bind()
        self.poll()
        self.status = {"status": "pinned-not-live", "pin": {"session": COORD}}
        self.append(self.row("assistant", "Completed the task."))
        self.assertEqual(self.poll()["status"], "suspended")
        self.status = {"status": "pinned-live-verified", "pin": {"session": COORD}, "registration_status": "busy"}
        self.assertEqual(self.poll()["status"], "suspended")
        self.assertEqual(self.codex.queues(), [])

    def test_bind_requires_verified_pin_and_canonical_uuid(self):
        self.status = {"status": "unpinned"}
        with self.assertRaises(handoff.Blocked):
            self.bind()
        self.status = {"status": "pinned-live-verified", "pin": {"session": COORD}}
        with self.assertRaises(handoff.Blocked):
            watch.main(["--repo", str(self.repo), "bind", "--supervisor", "supervisor-name"])

    def test_status_shows_fingerprint_not_uuid(self):
        self.bind()
        text = json.dumps(watch.main(["--repo", str(self.repo), "status"]))
        self.assertNotIn(SUPER, text)
        self.assertIn(watch.fingerprint(SUPER), text)


class StateAndQueue(Base):
    def test_unsafe_state_is_refused(self):
        self.bind()
        directory = watch.observer_dir(self.root)
        os.chmod(directory / "binding.json", 0o644)
        with self.assertRaises(handoff.Blocked):
            self.poll()
        os.chmod(directory / "binding.json", 0o600)
        target = self.tmp / "elsewhere.json"
        target.write_text("{}")
        os.chmod(target, 0o600)
        (directory / "state.json").unlink(missing_ok=True)
        os.symlink(target, directory / "state.json")
        with self.assertRaises(handoff.Blocked):
            self.poll()
        os.unlink(directory / "state.json")
        os.chmod(directory, 0o755)
        with self.assertRaises(handoff.Blocked):
            self.poll()
        self.assertEqual(self.codex.calls, [])

    def test_queue_failure_is_uncertain_and_never_retried(self):
        self.bind()
        self.poll()
        self.codex.queue_rc = 1
        self.append(self.row("assistant", "Ready for review."))
        result = self.poll()
        self.assertEqual((result["status"], result["reason"]), ("suspended", "queue-uncertain"))
        self.assertEqual(self.state()["deliveries"][-1]["outcome"], "queue-uncertain-exit-1")
        self.codex.queue_rc = 0
        self.append(self.row("assistant", "Blocked now."))
        self.assertEqual(self.poll()["queue_calls"], 0)
        self.assertEqual(len(self.codex.queues()), 1)

    def test_queue_timeout_is_uncertain(self):
        self.bind()
        self.poll()
        self.codex.raise_timeout = True
        self.append(self.row("assistant", "Handback sent."))
        self.assertEqual(self.poll()["reason"], "queue-uncertain")
        self.assertEqual(self.state()["deliveries"][-1]["outcome"], "queue-uncertain-error")

    def test_verified_daemon_upgrade_queues_once(self):
        self.bind()
        self.poll()
        self.codex.daemon = "0.163.0"
        self.append(self.row("assistant", "COORDINATOR_STATUS: review-ready | G013 | Closeout ready for Codex review"))
        self.assertEqual(self.poll()["queue_calls"], 1)
        self.assertEqual(self.poll()["queue_calls"], 0)
        self.assertEqual(len(self.codex.queues()), 1)
        self.assertIsNone(self.state()["suspended"])

    def test_unavailable_codex_capabilities_send_nothing(self):
        self.bind()
        self.poll()
        self.codex.daemon = "unknown"
        self.append(self.row("assistant", "Ready for review."))
        self.assertEqual(self.poll()["reason"], "incompatible-codex-capabilities")
        self.assertEqual(self.codex.queues(), [])

    def test_unbaselined_versions_with_supported_capabilities_queue_once(self):
        self.bind()
        self.poll()
        self.codex.cli, self.codex.daemon = "0.999.7", "9.9.9"
        self.append(self.row("assistant", "Ready for review."))
        self.assertEqual(self.poll()["queue_calls"], 1)
        self.assertEqual(self.state()["deliveries"][-1]["compatibility_warnings"],
                         ["unbaselined-codex-cli", "unbaselined-codex-daemon"])
        self.assertEqual(self.poll()["queue_calls"], 0)
        self.assertEqual(len(self.codex.queues()), 1)

    def test_missing_queue_options_never_attempt_delivery(self):
        for help_text in ("Usage: codex queue --thread <THREAD>",
                          "Usage: other --thread <THREAD> --message <TEXT>"):
            with self.subTest(help_text=help_text):
                self.setUp()
                self.bind()
                self.poll()
                self.codex.help_text = help_text
                self.append(self.row("assistant", "Ready for review."))
                self.assertEqual(self.poll()["reason"], "incompatible-codex-capabilities")
                self.assertEqual(self.codex.queues(), [])
                self.assertEqual(self.state()["keys"], [])

    def test_help_probe_timeout_is_known_unsent(self):
        self.bind()
        self.poll()
        original = self.codex
        def timeout_help(arguments, timeout):
            if arguments == ["queue", "--help"]:
                raise subprocess.TimeoutExpired(arguments, timeout)
            return original(arguments, timeout)
        self.append(self.row("assistant", "Ready for review."))
        with patch.object(watch, "run_codex", timeout_help):
            self.assertEqual(self.poll()["reason"], "incompatible-codex-capabilities")
        self.assertEqual(self.codex.queues(), [])

    def test_smoke_delivers_exactly_once(self):
        self.bind()
        result = watch.main(["--repo", str(self.repo), "smoke", "--label", "AC13 smoke"])
        self.assertEqual(result["status"], "queued-not-acknowledged")
        self.assertIn("OBSERVER SMOKE TEST", self.codex.queues()[0][4])
        with self.assertRaises(handoff.Blocked):
            watch.main(["--repo", str(self.repo), "smoke", "--label", "again"])
        self.assertEqual(len(self.codex.queues()), 1)


class KnownUnsentRecovery(Base):
    kinds = ("milestone", "advisory", "waiting", "smoke")

    def prepare(self, kind):
        self.bind("--size-threshold", "4096")
        self.poll()
        if kind == "milestone":
            row = self.row("assistant", "Plan is ready for review.")
            self.append(row)
            return watch.row_key(json.loads(row)), "review-ready"
        if kind == "advisory":
            with self.transcript.open("a") as stream:
                stream.write(" " * 5000 + "\n")
            return f"size:{COORD}:4096", "size-advisory"
        if kind == "waiting":
            self.status["registration_status"] = "waiting"
            return "waiting:legacy", "waiting"
        return "smoke:recovery smoke", "smoke"

    def fire(self, kind):
        if kind == "smoke":
            return watch.main(["--repo", str(self.repo), "smoke", "--label", "recovery smoke"])
        return self.poll()

    def legacy_blocked(self, kind, key, category):
        state = self.state()
        state["keys"].append(key)
        state["deliveries"].append({"key": key, "kind": category,
                                    "outcome": "not-sent-unverified-codex-version"})
        state["suspended"] = {"reason": "incompatible-codex-capabilities", "time": "2026-10-12T00:00:00Z"}
        if kind == "advisory":
            state["size_advised"].append(key)
            state["transcript"]["advised"] = True
        elif kind == "waiting":
            state["waiting"] = True
        elif kind == "smoke":
            state["smoke"] = {"label": "recovery smoke", "outcome": "pending-intent"}
        watch.save(watch.observer_dir(self.root), "state.json", state)

    def assert_consumed_once(self, kind):
        if kind == "smoke":
            with self.assertRaises(handoff.Blocked):
                self.fire(kind)
        else:
            self.assertEqual(self.poll()["queue_calls"], 0)
        self.assertEqual(len(self.codex.queues()), 1)

    def test_capability_blocked_markers_recover_only_after_explicit_rebind(self):
        for kind in self.kinds:
            for version in ("cli", "daemon"):
                with self.subTest(kind=kind, version=version):
                    self.setUp()
                    self.prepare(kind)
                    original = getattr(self.codex, version)
                    setattr(self.codex, version, "unknown")
                    self.assertEqual(self.fire(kind)["reason"], "incompatible-codex-capabilities")
                    state = self.state()
                    self.assertEqual(state["keys"], [])
                    self.assertEqual(state["size_advised"], [])
                    self.assertFalse(state["waiting"])
                    self.assertIsNone(state["smoke"])
                    self.assertEqual(self.codex.queues(), [])
                    setattr(self.codex, version, original)
                    self.assertEqual(self.poll()["status"], "suspended")
                    self.bind("--replace", "--size-threshold", "4096")
                    self.assertEqual(self.codex.queues(), [])  # rebinding is never delivery
                    self.assertEqual(self.fire(kind)["queue_calls"], 1)
                    self.assert_consumed_once(kind)

    def test_waiting_survives_another_event_being_version_blocked(self):
        self.prepare("milestone")
        self.status["registration_status"] = "waiting"
        self.codex.daemon = "unknown"
        self.assertEqual(self.poll()["reason"], "incompatible-codex-capabilities")
        self.assertFalse(self.state()["waiting"])
        self.codex.daemon = "0.162.1"
        self.bind("--replace", "--size-threshold", "4096")
        self.assertEqual(self.poll()["queue_calls"], 2)
        self.assertIn("(waiting", self.codex.queues()[1][4])
        self.assertEqual(self.poll()["queue_calls"], 0)

    def test_legacy_known_unsent_markers_recover_on_rebind(self):
        for kind in self.kinds:
            with self.subTest(kind=kind):
                self.setUp()
                key, category = self.prepare(kind)
                self.legacy_blocked(kind, key, category)
                self.bind("--replace", "--size-threshold", "4096")
                self.assertNotIn(key, self.state()["keys"])
                self.assertEqual(self.codex.queues(), [])
                self.assertEqual(self.fire(kind)["queue_calls"], 1)
                self.assert_consumed_once(kind)

    def test_legacy_unsent_record_never_releases_an_attempted_key(self):
        for kind in self.kinds:
            for outcome in ("queued-not-acknowledged", "uncertain-interrupted", "pending-intent"):
                with self.subTest(kind=kind, outcome=outcome):
                    self.setUp()
                    key, category = self.prepare(kind)
                    self.legacy_blocked(kind, key, category)
                    state = self.state()
                    state["deliveries"].append({"key": key, "kind": category, "outcome": outcome,
                                                "notice_id": "00000000-0000-4000-8000-0000000000e1"})
                    watch.save(watch.observer_dir(self.root), "state.json", state)
                    self.bind("--replace", "--size-threshold", "4096")
                    self.assertIn(key, self.state()["keys"])
                    if kind == "smoke":
                        with self.assertRaises(handoff.Blocked):
                            self.fire(kind)
                    else:
                        self.assertEqual(self.poll()["queue_calls"], 0)
                    self.assertEqual(self.codex.queues(), [])

    def test_version_change_mid_poll_does_not_repeat_the_first_notice(self):
        self.prepare("milestone")
        self.append(self.row("assistant", "Now blocked on input."))
        def change_version(arguments, timeout):
            result = self.codex(arguments, timeout)
            if arguments[:2] == ["queue", "--thread"]:
                self.codex.daemon = "unknown"
            return result
        with patch.object(watch, "run_codex", change_version):
            result = self.poll()
        self.assertEqual((result["reason"], result["queue_calls"]), ("incompatible-codex-capabilities", 1))
        self.codex.daemon = "0.162.1"
        self.bind("--replace", "--size-threshold", "4096")
        self.assertEqual(self.poll()["queue_calls"], 1)
        self.assertIn("(blocked", self.codex.queues()[1][4])
        self.assertEqual(len(self.codex.queues()), 2)
        self.assertEqual(self.poll()["queue_calls"], 0)


class SizeAdvisory(Base):
    def grow(self, size):
        with self.transcript.open("a") as stream:
            stream.write(" " * max(0, size - self.transcript.stat().st_size - 1) + "\n")

    def advisories(self):
        return [c for c in self.codex.queues() if "transcript size advisory" in c[4]]

    def test_default_threshold_is_100_mib_and_status_reports_it(self):
        self.bind()
        self.assertEqual(watch.DEFAULT_SIZE_THRESHOLD, 104857600)
        self.poll()
        status = watch.main(["--repo", str(self.repo), "status"])
        self.assertEqual(status["size_threshold"], 104857600)
        self.assertEqual(status["transcript"]["bytes"], self.transcript.stat().st_size)
        self.assertFalse(status["transcript"]["advised"])
        self.assertEqual(self.codex.calls, [])

    def test_below_at_and_above_threshold(self):
        self.bind("--size-threshold", "4096")
        self.poll()
        self.grow(4095)
        self.assertEqual(self.poll()["queue_calls"], 0)
        self.grow(4096)
        self.assertEqual(self.poll()["queue_calls"], 1)
        message = self.advisories()[0][4]
        self.assertIn("4096 bytes", message)
        self.assertIn("4096-byte threshold", message)
        self.assertIn("safe session handoff or archive", message)
        self.grow(9000)
        self.assertEqual(self.poll()["queue_calls"], 0)  # already advised for this coordinator and threshold
        self.assertTrue(watch.main(["--repo", str(self.repo), "status"])["transcript"]["advised"])

    def test_first_poll_already_large_qualifies(self):
        self.grow(5000)
        self.bind("--size-threshold", "4096")
        self.assertEqual(self.poll()["queue_calls"], 1)
        self.assertEqual(len(self.advisories()), 1)

    def test_restart_rotation_and_decrease_do_not_repeat(self):
        self.bind("--size-threshold", "4096")
        self.grow(5000)
        self.poll()
        self.assertEqual(self.poll()["queue_calls"], 0)       # restart: same private state
        self.transcript.write_text(self.row("assistant", "Running tests.") + "\n")  # decrease / truncation
        self.assertEqual(self.poll()["queue_calls"], 0)
        rotated = self.transcript.with_suffix(".new")
        rotated.write_text(" " * 6000 + "\n")
        os.replace(rotated, self.transcript)                   # rotation, large again
        self.assertEqual(self.poll()["queue_calls"], 0)
        self.assertEqual(len(self.advisories()), 1)

    def test_changed_threshold_and_opt_out(self):
        self.bind("--size-threshold", "4096")
        self.grow(5000)
        self.poll()
        self.bind("--replace", "--size-threshold", "2048")     # a new configured threshold is a new advisory key
        self.assertEqual(self.poll()["queue_calls"], 1)
        self.bind("--replace", "--no-size-advisory")
        self.grow(100000)
        self.assertEqual(self.poll()["queue_calls"], 0)
        self.assertIsNone(watch.main(["--repo", str(self.repo), "status"])["size_threshold"])
        with self.assertRaises(handoff.Blocked):
            self.bind("--replace", "--size-threshold", "10")

    def test_replacement_coordinator_is_assessed_separately(self):
        self.bind("--size-threshold", "4096")
        self.grow(5000)
        self.poll()
        other = self.projects / f"{OTHER}.jsonl"
        other.write_text(" " * 6000 + "\n")
        self.status["pin"]["session"] = OTHER
        self.assertEqual(self.poll()["status"], "suspended")   # never retargets implicitly
        self.bind("--replace", "--size-threshold", "4096")
        self.assertEqual(self.poll()["queue_calls"], 1)
        self.assertEqual(len(self.advisories()), 2)

    def test_uncertain_advisory_delivery_is_not_retried(self):
        self.bind("--size-threshold", "4096")
        self.poll()
        self.codex.queue_rc = 1
        self.grow(5000)
        self.assertEqual(self.poll()["reason"], "queue-uncertain")
        self.codex.queue_rc = 0
        self.bind("--replace", "--size-threshold", "4096")
        self.assertEqual(self.poll()["queue_calls"], 0)
        self.assertEqual(self.state()["deliveries"][-1]["outcome"], "queue-uncertain-exit-1")

    def test_explicit_zero_threshold_is_refused_without_binding_change(self):
        with self.assertRaises(handoff.Blocked) as ctx:
            self.bind("--size-threshold", "0")                     # no prior binding: none is created
        self.assertEqual(str(ctx.exception), "size-threshold-too-small")
        self.assertIsNone(handoff.read_private(watch.observer_dir(self.root) / "binding.json"))
        self.bind("--size-threshold", "4096")
        before = handoff.read_private(watch.observer_dir(self.root) / "binding.json")
        with self.assertRaises(handoff.Blocked) as ctx:
            self.bind("--replace", "--size-threshold", "0")        # explicit 0 never becomes the 100 MiB default
        self.assertEqual(str(ctx.exception), "size-threshold-too-small")
        self.assertEqual(handoff.read_private(watch.observer_dir(self.root) / "binding.json"), before)
        self.assertEqual(self.codex.calls, [])

    def test_omitted_valid_and_opt_out_thresholds(self):
        self.bind()
        self.assertEqual(watch.main(["--repo", str(self.repo), "status"])["size_threshold"], 104857600)
        self.bind("--replace", "--size-threshold", "2048")
        self.assertEqual(watch.main(["--repo", str(self.repo), "status"])["size_threshold"], 2048)
        self.bind("--replace", "--no-size-advisory")
        self.assertIsNone(watch.main(["--repo", str(self.repo), "status"])["size_threshold"])
        self.assertEqual(self.codex.calls, [])

    def test_unverified_identity_blocks_advisory(self):
        self.grow(5000)
        self.bind("--size-threshold", "4096")
        self.status = {"status": "pinned-unverified", "pin": {"session": COORD}}
        self.assertEqual(self.poll()["status"], "suspended")
        self.assertEqual(self.codex.calls, [])


class DurableIntent(Base):
    """C4: the intent and dedup key are persisted before `codex queue`; interruption never causes a blind retry."""

    class Died(BaseException):
        pass

    def die_after_enqueue(self):
        real = self.codex.__call__

        def call(arguments, timeout):
            result = real(arguments, timeout)
            if arguments[:2] == ["queue", "--thread"]:
                raise self.Died()  # the process dies right after the external enqueue
            return result
        return patch.object(watch, "run_codex", call)

    def fail_final_save(self):
        real = watch.save

        def save(directory, name, value):
            if name == "state.json" and self.codex.queues() and any(
                    d.get("outcome") != "pending-intent" for d in value.get("deliveries", [])[-1:]):
                raise OSError("disk full after enqueue")
            return real(directory, name, value)
        return patch.object(watch, "save", save)

    def fail_initial_save(self):
        real = watch.save

        def save(directory, name, value):
            if name == "state.json" and any(d.get("outcome") == "pending-intent"
                                            for d in value.get("deliveries", [])):
                raise OSError("cannot persist the intent")
            return real(directory, name, value)
        return patch.object(watch, "save", save)

    def trigger(self, kind):
        if kind == "milestone":
            self.append(self.row("assistant", "Result submitted for Codex review."))
            return self.poll
        if kind == "advisory":
            with self.transcript.open("a") as stream:
                stream.write(" " * 5000 + "\n")
            return self.poll
        return lambda: watch.main(["--repo", str(self.repo), "smoke", "--label", "AC13 smoke"])

    def prepare(self, kind):
        self.bind("--size-threshold", "4096") if kind == "advisory" else self.bind()
        if kind != "smoke":
            self.poll()

    def assert_recovered_without_retry(self, kind):
        if kind == "smoke":
            state = self.state()
            self.assertEqual(state["deliveries"][-1]["outcome"], "pending-intent")
            with self.assertRaises(handoff.Blocked):
                watch.main(["--repo", str(self.repo), "smoke", "--label", "AC13 smoke"])
            self.assertEqual(self.state()["smoke"]["outcome"], "uncertain-interrupted")
        else:
            result = self.poll()
            self.assertEqual((result["status"], result["reason"], result["queue_calls"]),
                             ("suspended", "interrupted-delivery-uncertain", 0))
        self.assertEqual(self.state()["deliveries"][-1]["outcome"], "uncertain-interrupted")
        self.bind("--replace", "--size-threshold", "4096")
        if kind != "smoke":
            self.assertEqual(self.poll()["queue_calls"], 0)  # never retried, even after explicit rebinding
        self.assertEqual(len(self.codex.queues()), 1)

    def test_interruption_after_enqueue_never_retries(self):
        for kind in ("milestone", "advisory", "smoke"):
            with self.subTest(kind=kind):
                self.setUp()
                self.prepare(kind)
                run = self.trigger(kind)
                with self.die_after_enqueue(), self.assertRaises(self.Died):
                    run()
                self.assertEqual(len(self.codex.queues()), 1)
                self.assert_recovered_without_retry(kind)

    def test_failed_final_persistence_never_retries(self):
        for kind in ("milestone", "advisory", "smoke"):
            with self.subTest(kind=kind):
                self.setUp()
                self.prepare(kind)
                run = self.trigger(kind)
                with self.fail_final_save(), self.assertRaises(OSError):
                    run()
                self.assertEqual(len(self.codex.queues()), 1)
                self.assert_recovered_without_retry(kind)

    def test_failed_initial_persistence_never_queues(self):
        for kind in ("milestone", "advisory", "smoke"):
            with self.subTest(kind=kind):
                self.setUp()
                self.prepare(kind)
                run = self.trigger(kind)
                with self.fail_initial_save(), self.assertRaises(OSError):
                    run()
                self.assertEqual(self.codex.queues(), [])

    def test_later_events_in_the_same_poll_are_not_lost(self):
        self.bind()
        self.poll()
        self.append(self.row("assistant", "Plan is review-ready."), self.row("assistant", "Now blocked on input."))
        with self.die_after_enqueue(), self.assertRaises(self.Died):
            self.poll()
        self.bind("--replace")
        self.assertEqual(self.poll()["queue_calls"], 1)  # the unattempted second milestone, offsets not advanced
        self.assertIn("(blocked", self.codex.queues()[-1][4])


class UnicodeSeparatorFraming(Base):
    """C6: rows are framed by LF only; U+2028, U+2029 and U+0085 inside JSON strings are content."""

    def raw_row(self, text):
        row = json.loads(self.row("assistant", text))
        return json.dumps(row, ensure_ascii=False)  # keeps the separator as a raw character in the transcript

    def test_raw_unicode_separators_do_not_split_rows(self):
        for separator in ("\u2028", "\u2029", "\u0085"):
            with self.subTest(separator=hex(ord(separator))):
                self.setUp()
                self.bind()
                self.poll()
                first = self.raw_row(f"Plan is review-ready.{separator}Details follow{separator}here.")
                self.assertIn(separator.encode("utf-8"), first.encode("utf-8"))
                self.append(first, self.raw_row(f"Now blocked{separator}on a decision."))
                self.assertEqual(self.poll()["queue_calls"], 2)
                messages = [c[4] for c in self.codex.queues()]
                self.assertTrue(any("(review-ready" in m for m in messages) and any("(blocked" in m for m in messages))
                self.assertEqual(self.poll()["queue_calls"], 0)      # each delivered exactly once
                offsets = self.state()["offsets"][str(self.transcript)]
                self.assertEqual(offsets["offset"], self.transcript.stat().st_size)
                self.assertNotIn("oversized", offsets)
                self.assertEqual(watch.main(["--repo", str(self.repo), "status"])["parse_limit"]["oversized_rows"], 0)


class OversizedRows(Base):
    """C5: bounded progress through rows longer than the read limit, with explicit parse-limit reporting."""

    def setUp(self):
        super().setUp()
        for name, value in (("MAX_READ", 1024), ("MAX_SCAN", 4096), ("PREFIX_BYTES", 256)):
            p = patch.object(watch, name, value)
            p.start()
            self.addCleanup(p.stop)
        self.bind()
        self.poll()

    def big(self, kind, size, text="padding"):
        row = json.loads(self.row(kind, text))
        row["pad"] = "x" * size
        return json.dumps(row)

    def offset(self):
        return self.state()["offsets"][str(self.transcript)]

    def test_complete_oversized_nonmilestone_row_then_milestone(self):
        self.append(self.big("user", 3000, "Incoming text says review-ready"),
                    self.row("assistant", "Result submitted for Codex review."))
        result = self.poll()
        self.assertEqual(result["queue_calls"], 1)
        self.assertIn("review-ready", self.codex.queues()[0][4])
        self.assertEqual(self.offset()["offset"], self.transcript.stat().st_size)
        self.assertNotIn("oversized", self.offset())
        limit = watch.main(["--repo", str(self.repo), "status"])["parse_limit"]
        self.assertEqual((limit["oversized_rows"], limit["last"]["coordinator_output"], limit["row_bytes"]),
                         (1, False, 1024))

    def test_oversized_coordinator_row_is_reported_not_silently_dropped(self):
        self.append(self.big("assistant", 3000, "Handback: work is complete."))
        self.assertEqual(self.poll()["queue_calls"], 1)
        notice = self.codex.queues()[0][4]
        self.assertIn("parse-limit", notice)
        self.assertIn("exceeded the 1024-byte parsing limit", notice)
        self.assertEqual(self.poll()["queue_calls"], 0)

    def test_growing_partial_oversized_row_progresses_and_completes(self):
        row = self.big("assistant", 9000, "Plan is review-ready.")
        start = self.transcript.stat().st_size
        with self.transcript.open("a") as stream:
            stream.write(row[:2500])
        self.assertEqual(self.poll()["queue_calls"], 0)
        first = self.offset()
        self.assertEqual((first["offset"], first["oversized"]["start"], first["oversized"]["scanned"]),
                         (start, start, start + 2500))
        with self.transcript.open("a") as stream:
            stream.write(row[2500:8000])
        self.assertEqual(self.poll()["queue_calls"], 0)
        second = self.offset()
        self.assertEqual(second["oversized"]["start"], start)
        self.assertGreater(second["oversized"]["scanned"], first["oversized"]["scanned"])  # resumes, no reread
        self.assertLessEqual(second["oversized"]["scanned"] - first["oversized"]["scanned"],
                             watch.MAX_SCAN + watch.MAX_READ)                   # bounded work per poll
        with self.transcript.open("a") as stream:
            stream.write(row[8000:] + "\n")
        self.append(self.row("assistant", "Now blocked on a decision."))
        calls = 0
        for _ in range(4):
            calls += self.poll()["queue_calls"]
        self.assertEqual(calls, 2)  # one parse-limit report and the following milestone; no fragment records
        kinds = [c[4] for c in self.codex.queues()]
        self.assertTrue(any("parse-limit" in k for k in kinds) and any("(blocked" in k for k in kinds))
        self.assertEqual(self.offset()["offset"], self.transcript.stat().st_size)
        self.assertEqual(watch.main(["--repo", str(self.repo), "status"])["parse_limit"]["oversized_rows"], 1)

    def test_same_inode_truncation_below_pending_oversized_scan_recovers(self):
        row_start = self.transcript.stat().st_size
        with self.transcript.open("a") as stream:
            stream.write(self.big("user", 9000)[:5000])          # partial oversized row
        self.poll()
        scanned = self.offset()["oversized"]["scanned"]
        self.assertEqual(self.offset()["oversized"]["start"], row_start)
        self.assertGreater(scanned, row_start + 2000)                # the scan already passed the truncation point
        inode = self.transcript.stat().st_ino
        with self.transcript.open("r+") as stream:               # in-place truncation, same inode
            stream.truncate(row_start + 2000)
            stream.seek(row_start + 2000)
            stream.write("\n")                                   # ends the truncated fragment
        self.append(self.row("assistant", "Result submitted for Codex review."))
        self.assertEqual(self.transcript.stat().st_ino, inode)
        self.assertGreater(self.transcript.stat().st_size, row_start)
        self.assertLess(self.transcript.stat().st_size, scanned)
        calls = sum(self.poll()["queue_calls"] for _ in range(4))
        self.assertEqual(calls, 1)                                # delivered exactly once, eventually
        self.assertIn("review-ready", self.codex.queues()[0][4])
        self.assertEqual(self.offset()["offset"], self.transcript.stat().st_size)
        self.assertNotIn("oversized", self.offset())

    def test_rotation_clears_a_pending_oversized_row(self):
        with self.transcript.open("a") as stream:
            stream.write(self.big("user", 3000)[:2000])
        self.poll()
        self.assertIn("oversized", self.offset())
        rotated = self.transcript.with_suffix(".new")
        rotated.write_text(self.row("assistant", "Ready for review after rotation.") + "\n")
        os.replace(rotated, self.transcript)
        self.assertEqual(self.poll()["queue_calls"], 1)
        self.assertNotIn("oversized", self.offset())


class Lifecycle(Base):
    def fake_child(self, root, interval):
        return [sys.executable, "-c", "import time; time.sleep(60)", "observer-fixture", str(interval)]

    def start_fixture(self, *extra):
        with patch.object(watch, "child_argv", self.fake_child):
            started = watch.main(["--repo", str(self.repo), "start", *extra])
        if started["status"] == "started":
            def cleanup():
                watch.main(["--repo", str(self.repo), "stop"])
                os.waitpid(started["pid"], 0)
            self.addCleanup(cleanup)
        return started

    def test_fresh_start_skips_saved_backlog_and_keeps_dedup(self):
        self.bind()
        self.poll()
        self.append(self.row("assistant", "Work is complete; handback sent."))
        self.poll()
        keys = self.state()["keys"]
        self.append(self.row("assistant", "Old Result is review-ready."))
        self.start_fixture()
        self.assertEqual(self.poll()["queue_calls"], 0)
        self.assertEqual(self.state()["keys"], keys)
        self.append(self.row("assistant", "New Result is review-ready."))
        self.assertEqual(self.poll()["queue_calls"], 1)
        offsets = self.state()["offsets"]
        self.append(self.row("assistant", "Another Result is review-ready."))
        self.assertEqual(self.start_fixture()["status"], "already-running")
        self.assertEqual(self.state()["offsets"], offsets)
        self.assertEqual(self.poll()["queue_calls"], 1)

    def test_resume_catches_up_once_from_saved_cursor(self):
        self.bind()
        self.poll()
        self.append(self.row("assistant", "Missed Result is review-ready."))
        offsets = self.state()["offsets"]
        self.start_fixture("--resume")
        self.assertEqual(self.state()["offsets"], offsets)
        self.assertEqual(self.poll()["queue_calls"], 1)
        self.assertEqual(self.poll()["queue_calls"], 0)

    def test_fresh_start_discards_existing_partial_row_without_parse_warning(self):
        self.bind()
        self.poll()
        row = self.row("assistant", "Old partial Result is review-ready.")
        with self.transcript.open("a") as stream:
            stream.write(row[:50])
        self.start_fixture()
        with self.transcript.open("a") as stream:
            stream.write(row[50:] + "\n")
        self.append(self.row("assistant", "Current Result is review-ready."))
        self.assertEqual(self.poll()["queue_calls"], 1)
        self.assertEqual(len(self.codex.queues()), 1)
        self.assertNotIn("oversized", self.state()["offsets"][str(self.transcript)])
        self.assertEqual(self.state().get("parse_limits", []), [])

    def test_fresh_start_does_not_clear_uncertain_intent(self):
        self.bind()
        self.poll()
        directory = watch.observer_dir(self.root)
        state = self.state()
        offsets = state["offsets"]
        state["deliveries"].append({"key": "interrupted", "outcome": "pending-intent"})
        watch.save(directory, "state.json", state)
        result = self.start_fixture()
        self.assertEqual(result["status"], "suspended")
        self.assertEqual(self.state()["offsets"], offsets)
        self.assertEqual(self.state()["deliveries"][-1]["outcome"], "uncertain-interrupted")
        self.assertEqual(self.codex.queues(), [])

    def test_start_duplicate_stop_and_cleanup(self):
        self.bind()
        with patch.object(watch, "child_argv", self.fake_child):
            started = watch.main(["--repo", str(self.repo), "start", "--interval", "15"])
            self.assertEqual(started["status"], "started")
            self.assertEqual(watch.main(["--repo", str(self.repo), "start"])["status"], "already-running")
            self.assertEqual(watch.main(["--repo", str(self.repo), "status"])["status"], "running")
            stopped = watch.main(["--repo", str(self.repo), "stop"])
        self.assertEqual(stopped["status"], "stopped")
        # In the test the fixture process is our child: reap it to prove it exited (the CLI's would reparent).
        pid, status = os.waitpid(started["pid"], 0)
        self.assertEqual(pid, started["pid"])
        with self.assertRaises(ProcessLookupError):
            os.kill(started["pid"], 0)
        self.assertEqual(watch.main(["--repo", str(self.repo), "status"])["status"], "stopped")

    def test_default_interval_is_thirty_and_override_is_honoured(self):
        self.assertEqual(watch.DEFAULT_INTERVAL, 30)
        self.assertEqual(watch.parser().parse_args(["--repo", "x", "start"]).interval, 30)
        self.assertEqual(watch.parser().parse_args(["--repo", "x", "run"]).interval, 30)
        self.bind()
        directory = watch.observer_dir(self.root)
        with patch.object(watch, "child_argv", self.fake_child):
            for argv, expected in ((["start"], 30), (["start", "--interval", "45"], 45)):
                started = watch.main(["--repo", str(self.repo), *argv])
                record = watch.load(directory, "process.json")
                self.assertEqual((started["interval"], record["interval"], record["argv"][-1]),
                                 (expected, expected, str(expected)))
                self.assertEqual(watch.main(["--repo", str(self.repo), "stop"])["status"], "stopped")
                os.waitpid(started["pid"], 0)

    def test_start_stays_inactive_without_a_verified_pair(self):
        self.bind()
        self.status = {"status": "discovery-unknown"}
        with patch.object(watch, "child_argv", self.fake_child):
            result = watch.main(["--repo", str(self.repo), "start"])
        self.assertEqual((result["status"], result["reason"]), ("inactive", "coordinator-discovery-unknown"))
        self.assertIsNone(watch.load(watch.observer_dir(self.root), "process.json"))

    def test_stop_refuses_reused_pid(self):
        self.bind()
        other = subprocess.Popen([sys.executable, "-c", "import time; time.sleep(30)"])
        self.addCleanup(other.wait)
        self.addCleanup(other.kill)
        directory = watch.observer_dir(self.root)
        watch.save(directory, "process.json", {"pid": other.pid, "start": handoff.process_start(other.pid),
                                               "argv": watch.child_argv(self.root, 20)})
        result = watch.main(["--repo", str(self.repo), "stop"])
        self.assertEqual(result["status"], "stale-record-not-signalled")
        self.assertIsNone(other.poll())

    def test_start_requires_binding_and_interval_bounds(self):
        with self.assertRaises(handoff.Blocked):
            watch.main(["--repo", str(self.repo), "start"])
        self.bind()
        with self.assertRaises(handoff.Blocked):
            watch.main(["--repo", str(self.repo), "start", "--interval", "1"])

    def test_run_loop_is_single_instance_and_stops_on_signal(self):
        self.bind()
        directory = watch.observer_dir(self.root)
        holder = os.open(directory / "run.lock", os.O_RDWR | os.O_CREAT, 0o600)
        fcntl.flock(holder, fcntl.LOCK_EX)
        self.assertEqual(watch.cmd_run(self.root, 5)["status"], "already-running")
        os.close(holder)
        polls = []

        def poll(root):
            polls.append(1)
            if len(polls) == 2:
                watch.STOP = True
            return {"status": "polled", "queue_calls": 0}
        with patch.object(watch, "poll_once", poll), patch.object(watch, "signal"):
            watch.STOP = False
            result = watch.cmd_run(self.root, 5)
        watch.STOP = False
        self.assertEqual((result["status"], len(polls)), ("stopped", 2))


if __name__ == "__main__":
    unittest.main()
