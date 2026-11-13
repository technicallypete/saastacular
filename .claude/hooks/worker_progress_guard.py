#!/usr/bin/env python3
"""Event-driven Claude worker stop guard. No polling, model calls or tool execution."""
import fcntl
import json
import os
from pathlib import Path
import sys
import tempfile

LIMIT = 40
def native_refusal(response):
    if isinstance(response, list):
        response = "\n".join(x.get("text", "") for x in response if isinstance(x, dict))
    text = str(response).strip().lower()
    if text.startswith("error:"):
        text = text[6:].lstrip()
    return text.startswith(("stop what you are doing", "the user doesn't want to proceed", "the tool use was rejected", "this agent is isolated in the worktree", "permission denied"))

def transition(state, event):
    state = dict(state)
    kind = event.get("hook_event_name")
    if kind == "PermissionDenied":
        state["held"] = "native refusal: return control; do not retry or reroute"
    elif kind == "PostToolUse" and event.get("tool_name") in ("Write", "Edit", "NotebookEdit"):
        state["written_ids"] = (state.get("written_ids", []) + [event.get("tool_use_id")])[-256:]
    elif kind == "PostCompact":
        if not state.get("artifact_since_compact", False):
            state["held"] = "compaction without observed artifact: return bounded obstacle/design checkpoint"
        state["artifact_since_compact"] = False
    elif kind == "PostToolBatch" and not state.get("held"):
        seen = state.get("seen", [])
        for call in event.get("tool_calls", []):
            ident = call.get("tool_use_id")
            if not ident:
                state["held"] = "missing tool identity: progress tracking unverified"
                break
            if ident in seen:
                continue
            seen.append(ident)
            if native_refusal(call.get("tool_response", "")):
                state["held"] = "native refusal: return control; do not retry or reroute"
                break
            # An artifact signal is not a quality/acceptance decision. Unknown shell
            # writes never silently reset the investigation budget.
            if ident in state.get("written_ids", []):
                state["without_artifact"] = 0
                state["artifact_since_compact"] = True
            else:
                state["without_artifact"] = state.get("without_artifact", 0) + 1
            if state["without_artifact"] >= LIMIT:
                state["held"] = "40 tool calls without observed artifact: return bounded obstacle/design checkpoint"
                break
        state["seen"] = seen[-256:]
    return state

def decision(state, event):
    if event.get("hook_event_name") == "PreToolUse" and event.get("tool_name") == "SubagentHandback":
        return {}
    if state.get("held"):
        return {"continue": False, "stopReason": "Restruct worker guard: " + state["held"] + ". Source/tests remain preserved. No automatic restart."}
    return {}

def worker_event(event):
    return bool(event.get("agent_id")) or "/subagents/" in event.get("transcript_path", "").replace("\\", "/")

def handle(event):
    kind = event.get("hook_event_name")
    if kind not in ("PreToolUse", "PostToolUse", "PostToolBatch", "PostCompact", "PermissionDenied"):
        raise ValueError("native event kind is unverified")
    if not isinstance(event.get("session_id"), str) or not event["session_id"] or not isinstance(event.get("transcript_path"), str):
        raise ValueError("native event identity is unverified")
    if kind == "PostToolBatch":
        calls = event.get("tool_calls")
        if not isinstance(calls, list) or not calls or any(not isinstance(c, dict) or not isinstance(c.get("tool_use_id"), str) or not c["tool_use_id"] or not isinstance(c.get("tool_name"), str) or "tool_response" not in c for c in calls):
            raise ValueError("native tool batch is unverified")
    elif kind != "PostCompact":
        if not isinstance(event.get("tool_name"), str) or not isinstance(event.get("tool_use_id"), str):
            raise ValueError("native tool identity is unverified")
    transcript = Path(event["transcript_path"])
    if not worker_event(event):
        if transcript.stem != event["session_id"]:
            raise ValueError("coordinator identity is unverified")
        return {} # coordinator does not acquire a worker budget
    if transcript.suffix != ".jsonl" or not transcript.is_file() or transcript.is_symlink():
        raise ValueError("worker transcript identity is unverified")
    if transcript.parent.stat().st_uid != os.getuid():
        raise ValueError("worker transcript owner differs")
    statepath = transcript.with_suffix(".restruct-progress.json")
    # Stable lock inode, atomic state replacement: interrupted writes cannot reset a latch.
    lockpath = statepath.with_suffix(".lock")
    try:
        fd = os.open(lockpath, os.O_RDWR | os.O_CREAT | os.O_EXCL | os.O_NOFOLLOW, 0o600)
        new_lock = True
    except FileExistsError:
        fd = os.open(lockpath, os.O_RDWR | os.O_NOFOLLOW)
        new_lock = False
    with os.fdopen(fd, "r+") as lock:
        fcntl.flock(lock, fcntl.LOCK_EX)
        if not new_lock and not statepath.exists():
            raise ValueError("interrupted guard initialization requires recovery")
        if statepath.is_symlink():
            raise ValueError("guard state cannot be a symlink")
        state = json.loads(statepath.read_text()) if statepath.exists() else {}
        if not isinstance(state, dict):
            raise ValueError("guard state is unverified")
        state = transition(state, event)
        pending = None
        try:
            with tempfile.NamedTemporaryFile(mode="w", dir=statepath.parent, prefix=statepath.name+".pending-", delete=False) as f:
                pending = Path(f.name)
                json.dump(state, f); f.flush(); os.fsync(f.fileno())
            os.replace(pending, statepath)
            directory = os.open(statepath.parent, os.O_RDONLY | os.O_DIRECTORY)
            try: os.fsync(directory)
            finally: os.close(directory)
        finally:
            if pending is not None: pending.unlink(missing_ok=True)
    # PostCompact cannot stop Claude: persist a latch for the next PreToolUse.
    if kind == "PostCompact":
        return {}
    return decision(state, event)

def main():
    try:
        result = handle(json.load(sys.stdin))
    except Exception:
        result = {"continue": False, "stopReason": "Restruct worker guard unverified: preserve state and return control; do not retry blindly."}
    print(json.dumps(result))

if __name__ == "__main__":
    main()
