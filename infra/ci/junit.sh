#!/usr/bin/env bash
# Supplemental JUnit report for infra/ci/validate.sh (G050; docs/foundation/OPERATIONS.md). Sourced, never run.
# junit_report builds ci-evidence/junit.xml for GitLab's artifacts:reports:junit from fields validate.sh has
# already validated: the sanitized status columns its completeness step writes (ID<TAB>STATUS), the recorded
# phase statuses and the job and profile tokens. It never reads free-text detail. It writes only OUT (through
# OUT.tmp); it emits
# no DOCTYPE, entity, timestamp, path, URL or text node; the only time attribute is the bounded measured
# validation elapsed seconds on the one timing case (never a sum of overlapping suite or phase durations);
# identical input gives identical bytes.
# The report is supplementary: the job exit status and the final job-log line stay authoritative.

# junit_attr VALUE: VALUE escaped for a double-quoted XML attribute (& < > " '). The replacements are quoted,
# so Bash 5.2's patsub_replacement never expands "&".
junit_attr() {
  local s="$1"
  s=${s//'&'/'&amp;'}
  s=${s//'<'/'&lt;'}
  s=${s//'>'/'&gt;'}
  s=${s//'"'/'&quot;'}
  s=${s//"'"/'&apos;'}
  printf '%s' "$s"
}

# junit_case CLASSNAME NAME [TYPE MESSAGE]: one testcase element, with a failure child when TYPE is given.
junit_case() {
  local tm=""
  if [[ "${JUNIT_CASE_SECONDS:-}" =~ ^[0-9]{1,5}$ ]]; then tm=" time=\"$JUNIT_CASE_SECONDS\""; fi
  if (( $# == 2 )); then
    printf '<testcase classname="%s" name="%s"%s/>' "$(junit_attr "$1")" "$(junit_attr "$2")" "$tm"
  else
    printf '<testcase classname="%s" name="%s"%s><failure type="%s" message="%s"/></testcase>' \
      "$(junit_attr "$1")" "$(junit_attr "$2")" "$tm" "$(junit_attr "$3")" "$(junit_attr "$4")"
  fi
}

# junit_report OUT JOB PROFILE COLUMNS NAMES [NAME=STATUS...]
#   OUT          the report to create; it must not exist (written to OUT.tmp, then renamed)
#   JOB          fast | gate | staging | full | local (validate.sh JOB_KIND)
#   PROFILE      fast | affected | integration | full | staging | none (the executed profile; none when nothing
#                executed) | remainder (G051: a gate's partition of an integration or full pipeline)
#   COLUMNS      the sanitized status-columns file, or "" for a phase-only suite
#   NAMES        the space-separated phase names (validate.sh PHASES)
#   NAME=STATUS  the recorded phases, in PHASES order
# Optional JUNIT_DURATIONS_FILE: validated suite|phase|elapsed<TAB>name<TAB>measured-seconds rows (elapsed: at
# most one row, named validation). GitLab's test-report totals sum case times, and suite and phase scopes overlap
# (background suites run concurrently; the harness phase contains them), so suite and phase rows are diagnostic
# only and never a case or a case time (2026-10-19 timing repair; the per-suite breakdown stays in timings.tsv).
# The elapsed row gives the one timing case: class validate.JOB.timing, name timing:validation-elapsed, passing,
# with that measured wall-clock value as its time; no real check or phase case carries a time. Absent check
# durations are never inferred.
# Cases: one per exported check in COLUMNS order (PASS passes; FAIL and UNMET fail with a fixed message; INFO
# is omitted), then the timing case (when an elapsed row is given), then one per phase whose numeric status is
# not 0 (skip, n/a, not-run and not-required are omitted). On success it prints "TESTS FAILURES" and returns 0.
# Anything else (a malformed or unknown status column, a duplicate non-INFO ID, a missing COLUMNS file, an
# unknown phase or status, an invalid label or duration row, an existing OUT) is a converter error: nothing is
# written and the status is 19.
junit_report() {
  local out="${1:-}" job="${2:-}" prof="${3:-}" cols="${4:-}" names=" ${5:-} " line id st kv name val tests=0 fails=0 cases="" doc
  local re=$'^([A-Z][A-Z0-9-]*)\t(PASS|FAIL|UNMET|INFO)$'
  local -A seen=() pseen=() duration=()
  local timing_kind timing_name timing_seconds elapsed=""
  if (( $# < 5 )); then return 19; fi
  shift 5
  if [[ ! "$job" =~ ^(fast|gate|staging|full|local)$ ]]; then return 19; fi
  if [[ ! "$prof" =~ ^(fast|affected|integration|full|staging|none|remainder)$ ]]; then return 19; fi
  if [[ -z "$out" || -e "$out" || -L "$out" || -e "$out.tmp" || -L "$out.tmp" ]]; then return 19; fi
  if [[ -n "${JUNIT_DURATIONS_FILE:-}" ]]; then
    [[ -f "$JUNIT_DURATIONS_FILE" && ! -L "$JUNIT_DURATIONS_FILE" ]] || return 19
    while IFS=$'\t' read -r timing_kind timing_name timing_seconds; do
      [[ "$timing_kind" =~ ^(suite|phase|elapsed)$ && "$timing_name" =~ ^[a-z][a-z0-9_-]{0,63}$ &&
         "$timing_seconds" =~ ^[0-9]{1,5}$ ]] || return 19
      [[ "$timing_kind" != elapsed || "$timing_name" == validation ]] || return 19
      [[ -z "${duration[$timing_kind:$timing_name]+x}" ]] || return 19
      duration[$timing_kind:$timing_name]="$timing_seconds"
      if [[ "$timing_kind" == elapsed ]]; then elapsed="$timing_seconds"; fi
    done <"$JUNIT_DURATIONS_FILE"
  fi
  if [[ -n "$cols" ]]; then
    if [[ -L "$cols" || ! -f "$cols" ]]; then return 19; fi
    while IFS= read -r line || [[ -n "$line" ]]; do
      if [[ ! "$line" =~ $re ]]; then return 19; fi
      id="${BASH_REMATCH[1]}" st="${BASH_REMATCH[2]}"
      if [[ "$st" == INFO ]]; then continue; fi
      if [[ -n "${seen[$id]+x}" ]]; then return 19; fi
      seen[$id]=1
      tests=$((tests + 1))
      case "$st" in
        PASS) cases+="$(junit_case "validate.$job.check" "$id")" ;;
        FAIL) cases+="$(junit_case "validate.$job.check" "$id" FAIL "check failed")"; fails=$((fails + 1)) ;;
        UNMET) cases+="$(junit_case "validate.$job.check" "$id" UNMET "required prerequisite unmet")"; fails=$((fails + 1)) ;;
      esac
    done <"$cols" || return 19
  fi
  # The one timing case: the measured validation elapsed seconds, never a sum of suite or phase durations.
  if [[ -n "$elapsed" ]]; then
    cases+="$(JUNIT_CASE_SECONDS="$elapsed" junit_case "validate.$job.timing" "timing:validation-elapsed")"
    tests=$((tests + 1))
  fi
  for kv in "$@"; do
    if [[ ! "$kv" =~ ^([a-z][a-z_]{0,31})=(.*)$ ]]; then return 19; fi
    name="${BASH_REMATCH[1]}" val="${BASH_REMATCH[2]}"
    if [[ "$names" != *" $name "* || -n "${pseen[$name]+x}" ]]; then return 19; fi
    pseen[$name]=1
    if [[ "$val" =~ ^[0-9]{1,3}$ ]]; then
      if (( 10#$val != 0 )); then
        cases+="$(junit_case "validate.$job.phase" "phase:$name" phase-error "phase $name exited $((10#$val))")"
        tests=$((tests + 1)) fails=$((fails + 1))
      fi
    elif [[ ! "$val" =~ ^(skip|n/a|not-run|not-required)$ ]]; then
      return 19
    fi
  done
  doc="$(printf '<?xml version="1.0" encoding="UTF-8"?>\n<testsuites tests="%s" failures="%s" errors="0" skipped="0"><testsuite name="%s" tests="%s" failures="%s" errors="0" skipped="0">%s</testsuite></testsuites>' \
    "$tests" "$fails" "$(junit_attr "$job:$prof")" "$tests" "$fails" "$cases")"
  if printf '%s\n' "$doc" >"$out.tmp" && mv -T -- "$out.tmp" "$out" && printf '%s %s\n' "$tests" "$fails"; then :; else
    rm -f -- "$out.tmp" "$out" || true
    return 19
  fi
}
