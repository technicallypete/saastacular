#!/usr/bin/env bash
# CI-FEDERATION fixture (G072 Plan r1 §10.3, §10.8, §10.10, §10.13 AC3; r0 §2.4): infra/terraform/bootstrap/
# ci-federation.sh against a fake gcloud that records every call and keeps a small resource and IAM-binding model. No
# cloud, network, Docker or credentials; the GitLab project pins are fake (a disposable tree carries its own
# infra/ci/hosted-context.sh with STG_PROJECT_ID=1001 and the hash of example-group/example-project).
#
#   bash infra/staging/tests/ci-federation.sh --work DIR --report DIR
#
# Cases (base tree):
#   dry-run      setup --dry-run calls no gcloud and prints exactly the reviewed transcript (pool, provider with the
#                three audiences, mapping and attribute condition, sts, the two service accounts, both custom roles,
#                the metadata bucket and its settings, every binding including the exact-key condition); named checks
#                for audiences, mapping, condition, exact-key condition, nothing on the state bucket, no forbidden role
#                or key verb; teardown --dry-run prints exactly the reviewed removal transcript
#   apply        setup against the fake: the final binding set is the bootstrap set plus exactly the reviewed CI set;
#                pool/provider fields, roles, bucket settings and the create-oidc argv are exact; a rerun creates
#                nothing and converges to the same state; the state bucket is never touched
#   verify       read-back passes with read calls only; a ci-stop binding on the state bucket, a widened condition, a
#                user-managed key and an extra project role each fail it; --verify --dry-run renders reads only
#   teardown     removes exactly the created set (snapshot equal to the bootstrap state; sts stays enabled; soft-deleted
#                pool/provider/roles), a second setup undeletes and verifies; with live metadata objects it refuses
#                before any write unless --confirm-meta <env> (a wrong name is refused)
#   ownership    the metadata bucket's owner project number: foreign, missing, malformed or an unreadable (403/503)
#                describe refuses setup and teardown with zero write calls; a matching owner is adopted; a confirmed
#                404 is absent; --verify fails on a foreign owner. A failed live-object listing (403, 404) or policy
#                read refuses teardown with zero write calls; a denied pool describe stops setup before any create; a
#                failed key listing, state-bucket IAM read or secrets listing makes --verify fail (UNREADABLE, no ok);
#                a bucket absent at teardown's first owner check but present before deletion is never deleted.
#                An owner that changes during setup (before the settings update, before the grants, right after
#                this run's create) stops it before any further bucket write; a name taken before the create is
#                refused, not adopted; an unchanged owner is rechecked before each bucket write and succeeds.
#                Under the legal project demo-404-preview, a 503 echoing gs://<bucket> or an account (setup, teardown
#                with live objects, verify, a policy read) is never a not-found; a genuine 404 is still absent
#   inputs       registry/pattern refusals, project-path hash, namespace ID, project-number pin, confirm rules: each
#                exits non-zero with no write call
#   permissions  ci-meta-permissions.tsv is the single source (an added/removed row changes exactly that binding) and
#                is validated (wildcard exact-key grant, write role for ci-read, unknown principal, bad shape, repeat)
#   mutations    each security clause mutated in a copy (an audience removed, every attribute-condition clause
#                dropped, a state-bucket binding, a forbidden role, a keys-create verb, a Secret Manager grant, the
#                exact-key condition dropped, the IAP tunnel grant made project-wide or widened to any instance, a
#                teardown deleting a bootstrap account, the live-object guard removed;
#                with the script's own forbid guard on and, where it applies, off) must fail the assertions
#   controls     the ownership cases against the script with the previous behaviour restored (owner-blind adoption,
#                owner not verified, a failed describe as absent, a failed listing as empty, tolerant verify and
#                value reads, no late owner checks, no owner recheck, the substring not-found parser): every case named for a control must
#                fail
#
# Output: REPORT/cases.tsv (case, result, detail), REPORT/summary.tsv (row, result, passed, failed), transcripts under
# REPORT/transcripts/. Last stdout line: "RESULT PASS passed=N failed=0" or "RESULT FAIL passed=N failed=M".
# Exit: 0 every case PASS; 1 any FAIL; 64 usage; 65 refused (unsafe path, existing report).
set -euo pipefail

REPO="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd -P)"
usage() { echo "usage: ci-federation.sh --work DIR --report DIR" >&2; exit 64; }
die() { echo "ci-federation fixture: $*" >&2; exit 65; }
WORK="" REPORT=""
while (($#)); do
  case "$1" in
    --work) [[ $# -ge 2 ]] || usage; WORK="$2"; shift 2 ;;
    --report) [[ $# -ge 2 ]] || usage; REPORT="$2"; shift 2 ;;
    *) usage ;;
  esac
done
[[ -n "$WORK" && -n "$REPORT" ]] || usage
# A path inside this checkout must be Git-ignored; anything else would change the tree.
safe_dir() {
  mkdir -p "$1"; local p; p="$(cd "$1" && pwd -P)"
  if [[ "$p/" == "$REPO/"* ]] && ! git -C "$REPO" check-ignore -q "${p#"$REPO"/}/"; then die "$1 is inside the checkout and not ignored"; fi
  printf '%s\n' "$p"
}
WORK="$(safe_dir "$WORK")" REPORT="$(safe_dir "$REPORT")"
[[ ! -e "$REPORT/cases.tsv" ]] || die "$REPORT/cases.tsv exists; use a new report directory per run"
RUN_ID="$(date -u +%Y%m%dt%H%M%Sz)-$(head -c 4 /dev/urandom | od -An -tx1 | tr -d ' \n')"
S="$WORK/ci-federation-$RUN_ID"; (umask 077; mkdir -p "$S/bin" "$S/tmp")
trap 'rm -rf "$S"' EXIT
TR="$REPORT/transcripts"; mkdir -p "$TR"
printf 'case\tresult\tdetail\n' >"$REPORT/cases.tsv"
export TMPDIR="$S/tmp"

SRC_SCRIPT="$REPO/infra/terraform/bootstrap/ci-federation.sh"
SRC_PERMS="$REPO/infra/terraform/bootstrap/ci-meta-permissions.tsv"
E=preview-1 P=example-preview-1 N=123456789012 R=us-east1 Z=us-east1-b
GL_PATH=example-group/example-project GL_NS=4242 GL_ID=1001
SA_DEP="$E-deployer@$P.iam.gserviceaccount.com" SA_VM="$E-vm@$P.iam.gserviceaccount.com"
SA_STOP="$E-ci-stop@$P.iam.gserviceaccount.com" SA_READ="$E-ci-read@$P.iam.gserviceaccount.com"
META="$P-$E-ci-meta" STATE="$P-tfstate"
PS="principalSet://iam.googleapis.com/projects/$N/locations/global/workloadIdentityPools/$E-ci/attribute.action"
SEP=$'\x1f'
COMMON=(--project "$P" --project-number "$N" --region "$R" --zone "$Z" --env "$E")
GL=(--gitlab-project-path "$GL_PATH" --gitlab-namespace-id "$GL_NS")
WRITE_RE=' (create|create-oidc|update|update-oidc|delete|undelete|enable|add-iam-policy-binding|remove-iam-policy-binding|rm) '

# ---- the reviewed rendering (oracle) ---------------------------------------------------------------------------------
G="+ gcloud --project $P --quiet"
AUDS="staging-ci:$E:start,staging-ci:$E:stop,staging-ci:$E:read"
MAPPING="google.subject=assertion.sub,attribute.action=assertion.aud == \"staging-ci:$E:start\" ? \"start\" : assertion.aud == \"staging-ci:$E:stop\" ? \"stop\" : assertion.aud == \"staging-ci:$E:read\" ? \"read\" : \"none\""
CLAUSES=(
  "assertion.project_id == \"$GL_ID\""
  "assertion.namespace_id == \"$GL_NS\""
  "assertion.ref_type == \"branch\""
  "assertion.ref == \"staging\""
  "assertion.ref_protected == \"true\""
  "assertion.ci_config_ref_uri == \"gitlab.com/$GL_PATH//.gitlab-ci.yml@refs/heads/staging\""
  "assertion.pipeline_source in [\"web\", \"api\"]"
  "assertion.environment == \"$E\""
  "assertion.environment_protected == \"true\""
  "assertion.runner_environment == \"gitlab-hosted\""
  "assertion.aud in [\"staging-ci:$E:start\", \"staging-ci:$E:stop\", \"staging-ci:$E:read\"]"
)
COND="" ; for c in "${CLAUSES[@]}"; do COND+="${COND:+ && }$c"; done
KEYCOND="resource.name == \"projects/_/buckets/$META/objects/current.tsv\" || resource.name == \"projects/_/buckets/$META/objects/lease\""
IAPCOND="resource.name == \"projects/$N/iap_tunnel/zones/$Z/instances/$E-vm\""
IAPFLAG="--condition=expression=$IAPCOND,title=ci-stop-iap-instance"
STOP_PERMS="compute.instances.get,compute.instances.list,compute.instances.stop,compute.disks.list,compute.addresses.list,compute.zoneOperations.get,compute.projects.get"
# kind SEP resource SEP role SEP member SEP condition: the CI set the fake model must hold after setup.
ORACLE_BINDINGS="$(printf '%s\n' \
  "sa$SEP$SA_DEP${SEP}roles/iam.workloadIdentityUser$SEP$PS/start$SEP" \
  "secret$SEP$E-TERRAFORM_CLOUDFLARE_API_TOKEN${SEP}roles/secretmanager.secretAccessor$SEP$PS/start$SEP" \
  "sa$SEP$SA_STOP${SEP}roles/iam.workloadIdentityUser$SEP$PS/stop$SEP" \
  "sa$SEP$SA_READ${SEP}roles/iam.workloadIdentityUser$SEP$PS/read$SEP" \
  "project$SEP$P${SEP}projects/$P/roles/ciStoppreview_1${SEP}serviceAccount:$SA_STOP$SEP" \
  "project$SEP$P${SEP}roles/iap.tunnelResourceAccessor${SEP}serviceAccount:$SA_STOP$SEP$IAPCOND" \
  "instance$SEP$E-vm${SEP}roles/compute.osAdminLogin${SEP}serviceAccount:$SA_STOP$SEP" \
  "sa$SEP$SA_VM${SEP}roles/iam.serviceAccountUser${SEP}serviceAccount:$SA_STOP$SEP" \
  "repo$SEP$E-images${SEP}roles/artifactregistry.reader${SEP}serviceAccount:$SA_READ$SEP" \
  "bucket$SEP$META${SEP}roles/storage.objectAdmin${SEP}serviceAccount:$SA_DEP$SEP" \
  "bucket$SEP$META${SEP}roles/storage.legacyBucketReader${SEP}serviceAccount:$SA_DEP$SEP" \
  "bucket$SEP$META${SEP}roles/storage.objectViewer${SEP}serviceAccount:$SA_STOP$SEP" \
  "bucket$SEP$META${SEP}roles/storage.objectCreator${SEP}serviceAccount:$SA_STOP$SEP" \
  "bucket$SEP$META${SEP}roles/storage.legacyBucketReader${SEP}serviceAccount:$SA_STOP$SEP" \
  "bucket$SEP$META${SEP}projects/$P/roles/ciMetaMutable${SEP}serviceAccount:$SA_STOP$SEP$KEYCOND" \
  "bucket$SEP$META${SEP}roles/storage.objectViewer${SEP}serviceAccount:$SA_READ$SEP" \
  "bucket$SEP$META${SEP}roles/storage.legacyBucketReader${SEP}serviceAccount:$SA_READ$SEP" | LC_ALL=C sort)"
oracle_setup() {
  local b="storage buckets add-iam-policy-binding gs://$META"
  printf '%s\n' \
    "$G services enable sts.googleapis.com --format=none" \
    "$G iam workload-identity-pools create $E-ci --location=global --display-name=CI $E --description=GitLab CI federation for $E (G072) --format=none" \
    "$G iam workload-identity-pools providers create-oidc gitlab --workload-identity-pool=$E-ci --location=global --issuer-uri=https://gitlab.com --allowed-audiences=$AUDS --attribute-mapping=$MAPPING --attribute-condition=$COND --format=none" \
    "$G iam service-accounts create $E-ci-stop --display-name=CI stop $E --format=none" \
    "$G iam service-accounts create $E-ci-read --display-name=CI read $E --format=none" \
    "$G iam roles create ciStoppreview_1 --title=CI stop $E --permissions=$STOP_PERMS --stage=GA --format=none" \
    "$G iam roles create ciMetaMutable --title=CI metadata mutable objects --permissions=storage.objects.delete --stage=GA --format=none" \
    "$G storage buckets create gs://$META --location=$R --uniform-bucket-level-access --public-access-prevention" \
    "$G storage buckets update gs://$META --versioning --uniform-bucket-level-access --public-access-prevention --lifecycle-file=<tmp>/lifecycle.json --update-labels=managed-by=ci-federation" \
    "$G iam service-accounts add-iam-policy-binding $SA_DEP --member=$PS/start --role=roles/iam.workloadIdentityUser --format=none" \
    "$G secrets add-iam-policy-binding $E-TERRAFORM_CLOUDFLARE_API_TOKEN --member=$PS/start --role=roles/secretmanager.secretAccessor --format=none" \
    "$G iam service-accounts add-iam-policy-binding $SA_STOP --member=$PS/stop --role=roles/iam.workloadIdentityUser --format=none" \
    "$G iam service-accounts add-iam-policy-binding $SA_READ --member=$PS/read --role=roles/iam.workloadIdentityUser --format=none" \
    "$G projects add-iam-policy-binding $P --member=serviceAccount:$SA_STOP --role=projects/$P/roles/ciStoppreview_1 --condition=None --format=none" \
    "$G projects add-iam-policy-binding $P --member=serviceAccount:$SA_STOP --role=roles/iap.tunnelResourceAccessor $IAPFLAG --format=none" \
    "$G compute instances add-iam-policy-binding $E-vm --zone=$Z --member=serviceAccount:$SA_STOP --role=roles/compute.osAdminLogin --format=none" \
    "$G iam service-accounts add-iam-policy-binding $SA_VM --member=serviceAccount:$SA_STOP --role=roles/iam.serviceAccountUser --format=none" \
    "$G artifacts repositories add-iam-policy-binding $E-images --location=$R --member=serviceAccount:$SA_READ --role=roles/artifactregistry.reader --format=none" \
    "$G $b --member=serviceAccount:$SA_DEP --role=roles/storage.objectAdmin --condition=None --format=none" \
    "$G $b --member=serviceAccount:$SA_DEP --role=roles/storage.legacyBucketReader --condition=None --format=none" \
    "$G $b --member=serviceAccount:$SA_STOP --role=roles/storage.objectViewer --condition=None --format=none" \
    "$G $b --member=serviceAccount:$SA_STOP --role=roles/storage.objectCreator --condition=None --format=none" \
    "$G $b --member=serviceAccount:$SA_STOP --role=roles/storage.legacyBucketReader --condition=None --format=none" \
    "$G $b --member=serviceAccount:$SA_STOP --role=projects/$P/roles/ciMetaMutable --condition=expression=$KEYCOND,title=ci-meta-exact-keys --format=none" \
    "$G $b --member=serviceAccount:$SA_READ --role=roles/storage.objectViewer --condition=None --format=none" \
    "$G $b --member=serviceAccount:$SA_READ --role=roles/storage.legacyBucketReader --condition=None --format=none"
}
oracle_teardown() {
  printf '%s\n' \
    "$G iam service-accounts remove-iam-policy-binding $SA_DEP --member=$PS/start --role=roles/iam.workloadIdentityUser --format=none" \
    "$G secrets remove-iam-policy-binding $E-TERRAFORM_CLOUDFLARE_API_TOKEN --member=$PS/start --role=roles/secretmanager.secretAccessor --format=none" \
    "$G iam service-accounts remove-iam-policy-binding $SA_STOP --member=$PS/stop --role=roles/iam.workloadIdentityUser --format=none" \
    "$G iam service-accounts remove-iam-policy-binding $SA_READ --member=$PS/read --role=roles/iam.workloadIdentityUser --format=none" \
    "$G projects remove-iam-policy-binding $P --member=serviceAccount:$SA_STOP --role=projects/$P/roles/ciStoppreview_1 --condition=None --format=none" \
    "$G projects remove-iam-policy-binding $P --member=serviceAccount:$SA_STOP --role=roles/iap.tunnelResourceAccessor $IAPFLAG --format=none" \
    "$G compute instances remove-iam-policy-binding $E-vm --zone=$Z --member=serviceAccount:$SA_STOP --role=roles/compute.osAdminLogin --format=none" \
    "$G iam service-accounts remove-iam-policy-binding $SA_VM --member=serviceAccount:$SA_STOP --role=roles/iam.serviceAccountUser --format=none" \
    "$G artifacts repositories remove-iam-policy-binding $E-images --location=$R --member=serviceAccount:$SA_READ --role=roles/artifactregistry.reader --format=none" \
    "$G iam service-accounts delete $SA_STOP" \
    "$G iam service-accounts delete $SA_READ" \
    "$G iam roles delete ciStoppreview_1 --format=none" \
    "$G iam roles delete ciMetaMutable --format=none" \
    "$G iam workload-identity-pools providers delete gitlab --workload-identity-pool=$E-ci --location=global --format=none" \
    "$G iam workload-identity-pools delete $E-ci --location=global --format=none" \
    "$G storage rm --recursive --all-versions gs://$META"
}

# ---- fake gcloud --------------------------------------------------------------------------------------------------------
cat >"$S/bin/gcloud" <<'FAKE'
#!/usr/bin/env bash
# Fake gcloud: records each call (%q-quoted argv) and keeps resources as key=value files under res/<kind>/<name> and
# IAM bindings as kind US resource US role US member US condition lines. Unknown calls exit 97.
set -u
D="${FAKE_GCLOUD_STATE:?}"
SEP=$'\x1f'
{ printf '%q ' "$@"; printf '\n'; } >>"$D/calls.log"
P="$(sed -n 's/^projectId=//p' "$D"/res/project/*)"
pos=() flags=()
while (($#)); do
  case "$1" in
    --project) [[ "${2:-}" == "$P" ]] || { echo "fake gcloud: wrong project ${2:-}" >&2; exit 98; }; shift 2 ;;
    --quiet) shift ;;
    --*) flags+=("$1"); shift ;;
    *) pos+=("$1"); shift ;;
  esac
done
# faults: "PREFIX<TAB>MESSAGE[<TAB>N]" lines; a call whose positional words start with PREFIX prints MESSAGE and exits
# 1 (only for the first N matching calls when N is given).
if [[ -s "$D/faults" ]]; then
  fl=0
  while IFS=$'\t' read -r fp fm fn; do
    fl=$((fl + 1))
    [[ -n "$fp" && "${pos[*]} " == "$fp "* ]] || continue
    if [[ -n "$fn" ]]; then
      fu="$(cat "$D/faults.used.$fl" 2>/dev/null || echo 0)"; ((fu < fn)) || continue; echo $((fu + 1)) >"$D/faults.used.$fl"
    fi
    printf '%s\n' "$fm" >&2; exit 1
  done <"$D/faults"
fi
# flips: "PREFIX<TAB>BUCKET<TAB>OWNER" lines; a call whose positional words start with PREFIX first sets BUCKET's owner
# (the name deleted and reused by another project meanwhile), then runs as usual.
if [[ -s "$D/flips" ]]; then
  while IFS=$'\t' read -r xp xb xo; do
    [[ -n "$xp" && "${pos[*]} " == "$xp "* ]] || continue
    mkdir -p "$D/owner"; printf '%s\n' "$xo" >"$D/owner/$xb"
  done <"$D/flips"
fi
unsupported() { echo "fake gcloud: unsupported call: ${pos[*]} ${flags[*]}" | tee -a "$D/unsupported.log" >&2; exit 97; }
flag() { local f; for f in "${flags[@]}"; do
  if [[ "$f" == "--$1="* ]]; then printf '%s' "${f#--"$1"=}"; return 0; fi
  if [[ "$f" == "--$1" ]]; then printf 'set'; return 0; fi
done; return 1; }
res() { printf '%s/res/%s/%s' "$D" "$1" "$2"; }
have() { [[ -e "$(res "$1" "$2")" ]]; }
need() { # gcloud's not-found shapes: compute's resource sentence, the API's NOT_FOUND status elsewhere
  have "$1" "$2" && return 0
  if [[ $1 == instance ]]; then printf "ERROR: (gcloud.compute.instances) Could not fetch resource:\n - The resource 'projects/%s/zones/z/instances/%s' was not found\n" "$P" "$2" >&2
  else echo "ERROR: (gcloud.$1) NOT_FOUND: Requested entity was not found ($1 $2)." >&2; fi
  exit 1
}
get() { sed -n "s|^$3=||p" "$(res "$1" "$2")"; }
setkv() { # KIND NAME KEY VALUE: replace the key's line in place (order kept) or append it
  local f; f="$(res "$1" "$2")"
  K="$3" V="$4" awk 'index($0, ENVIRON["K"] "=") == 1 {print ENVIRON["K"] "=" ENVIRON["V"]; done = 1; next} {print}
    END {if (!done) print ENVIRON["K"] "=" ENVIRON["V"]}' "$f" >"$f.n"
  mv "$f.n" "$f"
}
mk() { [[ ! -e "$(res "$1" "$2")" ]] || { echo "fake gcloud: $1 $2 already exists" >&2; exit 1; }; : >"$(res "$1" "$2")"; }
fmt="$(flag format || true)"
show() { # KIND NAME: print per --format
  case "$fmt" in
    ""|none) ;;
    "json(lifecycle_config)") printf '{\n  "lifecycle_config": %s\n}\n' "$(get "$1" "$2" lifecycle)" ;;
    "value("*")") get "$1" "$2" "${fmt:6:${#fmt}-7}" ;;
    *) unsupported ;;
  esac
}
policy() { # add|remove|get KIND RES
  local op=$1 kind=$2 r=$3 member role cond line has_cond=0
  need "$kind" "$r"
  if [[ $op == get ]]; then
    [[ "$fmt" == "value(bindings.role,bindings.members,bindings.condition.expression)" && "$(flag flatten || true)" == "bindings[].members" ]] || unsupported
    awk -F"$SEP" -v k="$kind" -v r="$r" '$1 == k && $2 == r {printf "%s\t%s\t%s\n", $3, $4, $5}' "$D/bindings"
    return 0
  fi
  member="$(flag member)" role="$(flag role)" || unsupported
  if cond="$(flag condition)"; then has_cond=1; else cond=""; fi
  case "$cond" in None|"") cond="" ;; expression=*,title=*) cond="${cond#expression=}"; cond="${cond%,title=*}" ;; *) unsupported ;; esac
  # Non-interactive gcloud refuses an unconditioned change on a policy that holds conditions.
  if ((has_cond == 0)) && awk -F"$SEP" -v k="$kind" -v r="$r" '$1 == k && $2 == r && $5 != "" {f = 1} END {exit !f}' "$D/bindings"; then
    echo "fake gcloud: policy has conditions; --condition is required" >&2; exit 1
  fi
  line="$kind$SEP$r$SEP$role$SEP$member$SEP$cond"
  if [[ $op == add ]]; then grep -qxF -- "$line" "$D/bindings" || printf '%s\n' "$line" >>"$D/bindings"
  else
    grep -qxF -- "$line" "$D/bindings" || { echo "fake gcloud: binding not found" >&2; exit 1; }
    grep -vxF -- "$line" "$D/bindings" >"$D/bindings.n" || true; mv "$D/bindings.n" "$D/bindings"
  fi
}
drop_bindings() { awk -F"$SEP" -v k="$1" -v r="$2" '!($1 == k && $2 == r)' "$D/bindings" >"$D/bindings.n"; mv "$D/bindings.n" "$D/bindings"; }
p0="${pos[0]:-}" p1="${pos[1]:-}" p2="${pos[2]:-}" p3="${pos[3]:-}" p4="${pos[4]:-}"
case "$p0 $p1 $p2" in
  "config get-value account") cat "$D/account" ;;
  "projects describe $P") show project "$P" ;;
  "projects add-iam-policy-binding $P") policy add project "$P" ;;
  "projects remove-iam-policy-binding $P") policy remove project "$P" ;;
  "projects get-iam-policy $P") policy get project "$P" ;;
  "services list ") [[ "$(flag enabled)" == set && "$fmt" == "value(config.name)" ]] || unsupported; cat "$D/apis" ;;
  "services enable "*) printf '%s\n' "$p2" >>"$D/apis" ;;
  "iam workload-identity-pools describe") need pool "$p3"; show pool "$p3" ;;
  "iam workload-identity-pools create") [[ "$(flag location)" == global ]] || unsupported; mk pool "$p3"; setkv pool "$p3" state ACTIVE ;;
  "iam workload-identity-pools undelete") need pool "$p3"; setkv pool "$p3" state ACTIVE ;;
  "iam workload-identity-pools delete") need pool "$p3"; setkv pool "$p3" state DELETED
    for f in "$D/res/provider/$p3--"*; do [[ ! -e "$f" ]] || setkv provider "${f##*/}" state DELETED; done ;;
  "iam workload-identity-pools providers")
    pv="$(flag workload-identity-pool)--$p4" || unsupported
    case "$p3" in
      describe) need provider "$pv"; show provider "$pv" ;;
      create-oidc|update-oidc)
        if [[ $p3 == create-oidc ]]; then [[ "$(get pool "${pv%%--*}" state 2>/dev/null)" == ACTIVE ]] || exit 1; mk provider "$pv"; setkv provider "$pv" state ACTIVE
        else need provider "$pv"; fi
        v="$(flag issuer-uri)" && setkv provider "$pv" oidc.issuerUri "$v"
        v="$(flag allowed-audiences)" && setkv provider "$pv" oidc.allowedAudiences "${v//,/;}"
        v="$(flag attribute-mapping)" && setkv provider "$pv" attributeMapping "${v//,/;}"
        v="$(flag attribute-condition)" && setkv provider "$pv" attributeCondition "$v" ;;
      undelete) need provider "$pv"; setkv provider "$pv" state ACTIVE ;;
      delete) need provider "$pv"; setkv provider "$pv" state DELETED ;;
      *) unsupported ;;
    esac ;;
  "iam service-accounts describe") need sa "$p3"; show sa "$p3" ;;
  "iam service-accounts create") mk sa "$p3@$P.iam.gserviceaccount.com"; setkv sa "$p3@$P.iam.gserviceaccount.com" email "$p3@$P.iam.gserviceaccount.com" ;;
  "iam service-accounts delete") need sa "$p3"; rm -f "$(res sa "$p3")"; drop_bindings sa "$p3" ;;
  "iam service-accounts keys") [[ "$p3" == list ]] || unsupported; a="$(flag iam-account)"; [[ ! -e "$D/keys/$a" ]] || cat "$D/keys/$a" ;;
  "iam service-accounts add-iam-policy-binding") policy add sa "$p3" ;;
  "iam service-accounts remove-iam-policy-binding") policy remove sa "$p3" ;;
  "iam service-accounts get-iam-policy") policy get sa "$p3" ;;
  "iam roles describe") need role "$p3"; show role "$p3" ;;
  "iam roles create") mk role "$p3"; setkv role "$p3" includedPermissions "$(flag permissions | tr , ';')"; setkv role "$p3" deleted False ;;
  "iam roles update") need role "$p3"; [[ "$(get role "$p3" deleted)" == False ]] || exit 1; setkv role "$p3" includedPermissions "$(flag permissions | tr , ';')" ;;
  "iam roles delete") need role "$p3"; setkv role "$p3" deleted True ;;
  "iam roles undelete") need role "$p3"; setkv role "$p3" deleted False ;;
  "secrets describe "*) need secret "$p2"; show secret "$p2" ;;
  "secrets list ") [[ "$fmt" == "value(name.basename())" ]] || unsupported; ls "$D/res/secret" ;;
  "secrets add-iam-policy-binding "*) policy add secret "$p2" ;;
  "secrets remove-iam-policy-binding "*) policy remove secret "$p2" ;;
  "secrets get-iam-policy "*) policy get secret "$p2" ;;
  "artifacts repositories describe") need repo "$p3"; show repo "$p3" ;;
  "artifacts repositories add-iam-policy-binding") policy add repo "$p3" ;;
  "artifacts repositories remove-iam-policy-binding") policy remove repo "$p3" ;;
  "artifacts repositories get-iam-policy") policy get repo "$p3" ;;
  "compute instances describe") need instance "$p3"; show instance "$p3" ;;
  "compute instances add-iam-policy-binding") policy add instance "$p3" ;;
  "compute instances remove-iam-policy-binding") policy remove instance "$p3" ;;
  "compute instances get-iam-policy") policy get instance "$p3" ;;
  "storage buckets describe") b="${p3#gs://}"
    have bucket "$b" || { echo "ERROR: (gcloud.storage.buckets.describe) gs://$b not found: 404." >&2; exit 1; }
    # --raw value(name,projectNumber): the owner is owner/<bucket> when present (a foreign number, "none" for a
    # missing field, any text for a malformed one), else this project's number.
    if [[ "$fmt" == "value(name,projectNumber)" ]]; then
      [[ "$(flag raw || true)" == set ]] || unsupported
      o="$(cat "$D/owner/$b" 2>/dev/null || sed -n 's/^projectNumber=//p' "$D"/res/project/*)"
      if [[ "$o" == none ]]; then printf '%s\n' "$b"; else printf '%s\t%s\n' "$b" "$o"; fi
    else show bucket "$b"; fi ;;
  "storage buckets create") b="${p3#gs://}"
    [[ -n "$(flag location || true)" && "$(flag uniform-bucket-level-access)" == set && "$(flag public-access-prevention)" == set ]] || unsupported
    mk bucket "$b"; setkv bucket "$b" location "$(flag location)"; setkv bucket "$b" uniform_bucket_level_access True
    setkv bucket "$b" public_access_prevention enforced; setkv bucket "$b" versioning_enabled False; : >"$D/objects/$b" ;;
  "storage buckets update") b="${p3#gs://}"; need bucket "$b"
    [[ "$(flag versioning || true)" != set ]] || setkv bucket "$b" versioning_enabled True
    [[ "$(flag uniform-bucket-level-access || true)" != set ]] || setkv bucket "$b" uniform_bucket_level_access True
    [[ "$(flag public-access-prevention || true)" != set ]] || setkv bucket "$b" public_access_prevention enforced
    if v="$(flag lifecycle-file)"; then setkv bucket "$b" lifecycle "$(tr -d '\n' <"$v")"; fi
    if v="$(flag update-labels)"; then setkv bucket "$b" labels "$v"; fi ;;
  "storage buckets add-iam-policy-binding") policy add bucket "${p3#gs://}" ;;
  "storage buckets remove-iam-policy-binding") policy remove bucket "${p3#gs://}" ;;
  "storage buckets get-iam-policy") policy get bucket "${p3#gs://}" ;;
  "storage ls "*) b="${p2#gs://}"; b="${b%%/*}"
    have bucket "$b" || { echo "ERROR: (gcloud.storage.ls) gs://$b not found: 404." >&2; exit 1; }
    # gcloud's answer for a listing that matched nothing in an existing bucket.
    [[ -s "$D/objects/$b" ]] || { echo "ERROR: (gcloud.storage.ls) One or more URLs matched no objects." >&2; exit 1; }
    sed "s|^|gs://$b/|" "$D/objects/$b" ;;
  "storage rm "*) b="${p2#gs://}"; [[ "$b" != */* && "$(flag recursive)" == set && "$(flag all-versions)" == set ]] || unsupported
    need bucket "$b"; rm -f "$(res bucket "$b")" "$D/objects/$b"; drop_bindings bucket "$b" ;;
  *) unsupported ;;
esac
FAKE
chmod 0755 "$S/bin/gcloud"

# ---- helpers -------------------------------------------------------------------------------------------------------------
# The bootstrap.sh state of one environment project (plus another secret and a VM), with a few bootstrap bindings.
init_state() {
  local d=$1 k
  rm -rf "$d"; mkdir -p "$d/keys" "$d/objects"
  for k in project pool provider sa role bucket secret repo instance; do mkdir -p "$d/res/$k"; done
  printf 'projectId=%s\nprojectNumber=%s\n' "$P" "$N" >"$d/res/project/$P"
  echo owner@example.net >"$d/account"
  printf '%s\n' compute.googleapis.com iam.googleapis.com iamcredentials.googleapis.com storage.googleapis.com \
    secretmanager.googleapis.com artifactregistry.googleapis.com >"$d/apis"
  printf 'email=%s\n' "$SA_DEP" >"$d/res/sa/$SA_DEP"; printf 'email=%s\n' "$SA_VM" >"$d/res/sa/$SA_VM"
  : >"$d/res/repo/$E-images"; : >"$d/res/secret/$E-TERRAFORM_CLOUDFLARE_API_TOKEN"; : >"$d/res/secret/$E-CF_API_TOKEN"
  : >"$d/res/instance/$E-vm"
  printf 'location=%s\nuniform_bucket_level_access=True\npublic_access_prevention=enforced\nversioning_enabled=True\n' "$R" >"$d/res/bucket/$STATE"
  printf 'staging/%s/default.tfstate\n' "$E" >"$d/objects/$STATE"
  printf '%s\n' \
    "bucket$SEP$STATE${SEP}roles/storage.objectAdmin${SEP}serviceAccount:$SA_DEP$SEP" \
    "project$SEP$P${SEP}roles/compute.instanceAdmin.v1${SEP}serviceAccount:$SA_DEP$SEP" \
    "sa$SEP$SA_VM${SEP}roles/iam.serviceAccountUser${SEP}serviceAccount:$SA_DEP$SEP" \
    "sa$SEP$SA_DEP${SEP}roles/iam.serviceAccountTokenCreator${SEP}user:owner@example.net$SEP" \
    "repo$SEP$E-images${SEP}roles/artifactregistry.writer${SEP}serviceAccount:$SA_DEP$SEP" \
    "secret$SEP$E-CF_API_TOKEN${SEP}roles/secretmanager.secretAccessor${SEP}serviceAccount:$SA_VM$SEP" >"$d/bindings"
  : >"$d/calls.log"
}
# Resources (soft-deleted ones count as removed), bindings and live objects; APIs separately.
snapshot() {
  local d=$1 f
  for f in "$d"/res/*/*; do
    [[ -e "$f" ]] || continue
    grep -qxE 'state=DELETED|deleted=True' "$f" && continue
    printf 'res %s\n' "${f#"$d"/res/}"
  done | LC_ALL=C sort
  LC_ALL=C sort "$d/bindings" | tr "$SEP" '|' | sed 's/^/binding /'
  for f in "$d"/objects/*; do [[ -e "$f" ]] || continue; sed "s|^|object ${f##*/}/|" "$f"; done | LC_ALL=C sort
}
full_snapshot() {
  local d=$1 f
  snapshot "$d"
  for f in "$d"/res/*/*; do if [[ -e "$f" ]]; then echo "== ${f#"$d"/res/}"; cat "$f"; fi; done
  LC_ALL=C sort "$d/apis"
}
# make_tree DIR: a disposable copy of the script and permissions file with fake GitLab pins and the registry.
REGISTRY_NOTE=""
make_tree() {
  local t=$1
  rm -rf "$t"; mkdir -p "$t/infra/terraform/bootstrap" "$t/infra/ci" "$t/infra/staging"
  cp "$SRC_SCRIPT" "$SRC_PERMS" "$t/infra/terraform/bootstrap/"
  printf 'STG_PROJECT_ID=%s\nSTG_PROJECT_PATH_SHA=%s\n' "$GL_ID" "$(printf '%s' "$GL_PATH" | sha256sum | cut -d' ' -f1)" >"$t/infra/ci/hosted-context.sh"
  if [[ -f "$REPO/infra/staging/ci-environments.txt" ]]; then cp "$REPO/infra/staging/ci-environments.txt" "$t/infra/staging/"; REGISTRY_NOTE="registry from infra/staging/ci-environments.txt"
  else printf 'staging\npreview-1\n' >"$t/infra/staging/ci-environments.txt"; REGISTRY_NOTE="registry absent in this tree: the shared-contract content (staging, preview-1) used"; fi
}
RC=0 OUT="" CALLS=""
fed() { # TREE STATE NAME ARGS...: run the script against the fake; RC, OUT (stdout+stderr file), CALLS (call log copy)
  local tree=$1 st=$2 name=$3; shift 3
  [[ -d "$st" ]] || init_state "$st"
  : >"$st/calls.log"
  OUT="$CUR_TR/$name.out" CALLS="$CUR_TR/$name.calls"
  set +e
  PATH="$S/bin:$PATH" FAKE_GCLOUD_STATE="$st" bash "$tree/infra/terraform/bootstrap/ci-federation.sh" "$@" >"$OUT" 2>&1
  RC=$?
  set -e
  cp "$st/calls.log" "$CALLS"
  # A call the fake does not model (exit 97) would hide behind the script's tolerant reads: collect it.
  if [[ -s "$st/unsupported.log" ]]; then cat "$st/unsupported.log" >>"$CUR_TR/unsupported.log"; : >"$st/unsupported.log"; fi
}
nwrites() { grep -cE -- "$WRITE_RE" "$1" || true; }
ncalls() { grep -c . "$1" || true; }
plus_lines() { grep '^+ ' "$OUT" | sed -E 's#--lifecycle-file=[^ ]*/lifecycle\.json#--lifecycle-file=<tmp>/lifecycle.json#' || true; }
bindings_of() { LC_ALL=C sort "$1/bindings"; }

CUR_LOG="" CUR_TR="" CUR_FAILS=0
pass() { printf 'PASS\t%s\t%s\n' "$1" "$2" >>"$CUR_LOG"; }
fail() { printf 'FAIL\t%s\t%s\n' "$1" "$2" >>"$CUR_LOG"; CUR_FAILS=$((CUR_FAILS + 1)); }
t() { local n=$1 d=$2; shift 2; if "$@"; then pass "$n" "$d"; else fail "$n" "$d"; fi; }

# ---- the assertion suite over one tree -----------------------------------------------------------------------------------
# LEVEL dry: the dry-run renderings only; full: plus apply, idempotence, teardown and the live-object refusal.
suite() { # TREE TAG LEVEL
  local tree=$1 tag=$2 level=$3 st pre post line
  CUR_TR="$TR/$tag"; mkdir -p "$CUR_TR"; CUR_LOG="$CUR_TR/assertions.tsv"; : >"$CUR_LOG"; CUR_FAILS=0
  st="$S/state-$tag-dry"; init_state "$st"
  fed "$tree" "$st" setup-dry "${COMMON[@]}" "${GL[@]}" --dry-run
  plus_lines >"$CUR_TR/setup-dry.plus"; oracle_setup >"$CUR_TR/setup-dry.oracle"
  t dry.exit "setup --dry-run exits 0 (got $RC)" test "$RC" = 0
  t dry.no-gcloud "setup --dry-run calls no gcloud ($(ncalls "$CALLS") calls)" test "$(ncalls "$CALLS")" = 0
  t dry.transcript "setup --dry-run prints exactly the reviewed transcript (diff setup-dry.oracle setup-dry.plus)" cmp -s "$CUR_TR/setup-dry.oracle" "$CUR_TR/setup-dry.plus"
  line="$(grep -F ' providers create-oidc gitlab ' "$CUR_TR/setup-dry.plus" || true)"
  t dry.audiences "create-oidc carries exactly --allowed-audiences=$AUDS" grep -qF -- " --allowed-audiences=$AUDS --attribute-mapping=" <<<"$line"
  t dry.mapping "create-oidc carries the exact attribute mapping (aud suffix -> start|stop|read, else none)" grep -qF -- " --attribute-mapping=$MAPPING --attribute-condition=" <<<"$line"
  t dry.condition "create-oidc carries the exact 11-clause attribute condition" grep -qF -- " --attribute-condition=$COND --format=none" <<<"$line"
  t dry.exact-keys "ciMetaMutable is bound once, only to ci-stop, with exactly the current.tsv/lease condition" \
    test "$(grep -F 'roles/ciMetaMutable' "$CUR_TR/setup-dry.plus" | grep -c 'add-iam-policy-binding' || true)" = 1 -a \
    "$(grep -cF -- "--member=serviceAccount:$SA_STOP --role=projects/$P/roles/ciMetaMutable --condition=expression=$KEYCOND,title=ci-meta-exact-keys --format=none" "$CUR_TR/setup-dry.plus" || true)" = 1
  t dry.iap-instance "iap.tunnelResourceAccessor is bound once, to ci-stop, only with the exact <env>-vm tunnel-resource condition (never project-wide)" \
    test "$(grep -cF -- '--role=roles/iap.tunnelResourceAccessor' "$CUR_TR/setup-dry.plus" || true)" = 1 -a \
    "$(grep -cF -- "projects add-iam-policy-binding $P --member=serviceAccount:$SA_STOP --role=roles/iap.tunnelResourceAccessor $IAPFLAG --format=none" "$CUR_TR/setup-dry.plus" || true)" = 1
  t dry.no-state-bucket "no mutating command names the state bucket gs://$STATE" bash -c '! grep -qF -- "gs://$1" "$2"' _ "$STATE" "$CUR_TR/setup-dry.plus"
  t dry.no-forbidden-role "no basic, IAM-admin, key-admin, token-creator or storage-admin role; Secret Manager only as the start principal's accessor on the Cloudflare secret" \
    no_forbidden "$CUR_TR/setup-dry.plus"
  t dry.no-key-verb "no service-account key verb in the rendering or the script" \
    bash -c '! grep -qE " keys (create|upload)| service-accounts keys " "$1" && ! grep -qE "keys[[:space:]]+(create|upload)" "$2"' _ "$CUR_TR/setup-dry.plus" "$tree/infra/terraform/bootstrap/ci-federation.sh"
  st="$S/state-$tag-tdry"; init_state "$st"
  fed "$tree" "$st" teardown-dry "${COMMON[@]}" --teardown --confirm "$P" --dry-run
  plus_lines >"$CUR_TR/teardown-dry.plus"; oracle_teardown >"$CUR_TR/teardown-dry.oracle"
  t tdry.transcript "teardown --dry-run exits 0 (got $RC), calls no gcloud and prints exactly the reviewed removal transcript" \
    test "$RC" = 0 -a "$(ncalls "$CALLS")" = 0 -a "$(cmp -s "$CUR_TR/teardown-dry.oracle" "$CUR_TR/teardown-dry.plus" && echo same)" = same
  [[ "$level" == full ]] || return 0

  st="$S/state-$tag-apply"; init_state "$st"; pre="$(snapshot "$st")"; bindings_of "$st" >"$CUR_TR/bindings.pre"
  fed "$tree" "$st" apply "${COMMON[@]}" "${GL[@]}"
  t apply.exit "setup against the fake exits 0 (got $RC; apply.out)" test "$RC" = 0
  { cat "$CUR_TR/bindings.pre"; printf '%s\n' "$ORACLE_BINDINGS"; } | LC_ALL=C sort >"$CUR_TR/bindings.want"
  bindings_of "$st" >"$CUR_TR/bindings.got"
  t apply.bindings "final bindings = bootstrap set + exactly the reviewed CI set (diff bindings.want bindings.got)" cmp -s "$CUR_TR/bindings.want" "$CUR_TR/bindings.got"
  t apply.state-bucket "no call names the state bucket; its bindings and objects are unchanged" \
    bash -c '! grep -qF -- "gs://$1" "$2" && [[ "$(grep -F -- "$1" "$3" | sort)" == "$(grep -F -- "$1" "$4" | sort)" ]]' _ "$STATE" "$CALLS" "$CUR_TR/bindings.pre" "$CUR_TR/bindings.got"
  t apply.provider "pool ACTIVE; provider ACTIVE with the exact issuer, audiences, mapping and condition" provider_ok "$st"
  t apply.argv "the create-oidc argv is exact (one argument each)" grep -qxF -- "$(printf '%q ' --project "$P" --quiet iam workload-identity-pools providers create-oidc gitlab \
    "--workload-identity-pool=$E-ci" --location=global --issuer-uri=https://gitlab.com "--allowed-audiences=$AUDS" "--attribute-mapping=$MAPPING" \
    "--attribute-condition=$COND" --format=none)" "$CALLS"
  t apply.identities "created service accounts: exactly ci-stop and ci-read (deployer and vm pre-existing, none other)" \
    test "$(ls "$st/res/sa" | LC_ALL=C sort | tr '\n' ' ')" = "$(printf '%s\n' "$SA_DEP" "$SA_VM" "$SA_STOP" "$SA_READ" | LC_ALL=C sort | tr '\n' ' ')"
  t apply.roles "ciStoppreview_1 = the seven compute permissions; ciMetaMutable = storage.objects.delete" \
    test "$(cat "$st/res/role/ciStoppreview_1" 2>/dev/null)" = "$(printf 'includedPermissions=%s\ndeleted=False' "${STOP_PERMS//,/;}")" -a \
    "$(cat "$st/res/role/ciMetaMutable" 2>/dev/null)" = "$(printf 'includedPermissions=storage.objects.delete\ndeleted=False')"
  t apply.bucket "metadata bucket: region, uniform access, public-access prevention, versioning, 90-day noncurrent expiry, label managed-by=ci-federation" bucket_ok "$st"
  t apply.sts "sts.googleapis.com enabled" grep -qxF sts.googleapis.com "$st/apis"
  post="$(full_snapshot "$st")"
  fed "$tree" "$st" apply-rerun "${COMMON[@]}" "${GL[@]}"
  t apply.idempotent "a rerun exits 0 (got $RC), creates nothing and converges to the same state" \
    test "$RC" = 0 -a "$(grep -cE ' (create|create-oidc|undelete|enable) ' "$CALLS" || true)" = 0 -a "$(full_snapshot "$st")" = "$post"
  fed "$tree" "$st" teardown "${COMMON[@]}" --teardown --confirm "$P"
  t teardown.exact "teardown exits 0 (got $RC) and leaves exactly the bootstrap state (diff teardown.pre teardown.post)" \
    snap_equal "$pre" "$(snapshot "$st")" "$CUR_TR/teardown"
  t teardown.keeps-bootstrap "teardown deletes no bootstrap item (deployer/vm account, repository, secret, VM, state bucket) and keeps sts enabled" \
    bash -c '! grep -E " (delete|rm) " "$1" | grep -qE -- "$2|$3|repositories|secrets|instances|gs://$4" && grep -qxF sts.googleapis.com "$5"' _ "$CALLS" "$SA_DEP" "$SA_VM" "$STATE" "$st/apis"
  st="$S/state-$tag-live"; init_state "$st"
  fed "$tree" "$st" live-setup "${COMMON[@]}" "${GL[@]}"
  printf '%s\n' current.tsv lease runs/1-2.tsv >"$st/objects/$META"
  pre="$(full_snapshot "$st")"
  fed "$tree" "$st" live-teardown "${COMMON[@]}" --teardown --confirm "$P"
  t teardown.live-refused "with live metadata objects, teardown without --confirm-meta exits non-zero (got $RC) with zero write calls and no change" \
    test "$RC" != 0 -a "$(nwrites "$CALLS")" = 0 -a "$(full_snapshot "$st")" = "$pre"
  t teardown.live-message "the refusal names the live objects and --confirm-meta $E" grep -qF -- "holds 3 live metadata object(s); retire the environment first, then pass --confirm-meta $E" "$OUT"
}
no_forbidden() { # FILE: role allowlist and Secret Manager scope on the "+ " lines
  local f=$1 role ok=0
  while IFS= read -r role; do
    case "$role" in
      roles/iam.workloadIdentityUser|roles/secretmanager.secretAccessor|roles/compute.osAdminLogin|roles/iap.tunnelResourceAccessor|\
      roles/iam.serviceAccountUser|roles/artifactregistry.reader|roles/storage.objectAdmin|roles/storage.objectViewer|\
      roles/storage.objectCreator|roles/storage.legacyBucketReader|"projects/$P/roles/ciStoppreview_1"|"projects/$P/roles/ciMetaMutable") ;;
      *) ok=1 ;;
    esac
  done < <(grep -oE -- '--role=[^ ]+' "$f" | sed 's/^--role=//')
  # objectAdmin only for the deployer; Secret Manager only once, for the start principal on the Cloudflare secret.
  [[ "$(grep -F -- '--role=roles/storage.objectAdmin' "$f" | grep -cvF -- "--member=serviceAccount:$SA_DEP " || true)" == 0 ]] || ok=1
  [[ "$(grep -c -- '--role=roles/secretmanager\.' "$f" || true)" == 1 ]] || ok=1
  grep -qF -- "secrets add-iam-policy-binding $E-TERRAFORM_CLOUDFLARE_API_TOKEN --member=$PS/start --role=roles/secretmanager.secretAccessor " "$f" || ok=1
  grep -qE -- '--member=serviceAccount:[^ ]+ --role=roles/iam\.workloadIdentityUser' "$f" && ok=1
  return "$ok"
}
provider_ok() {
  local d=$1 f="$1/res/provider/$E-ci--gitlab"
  [[ "$(cat "$d/res/pool/$E-ci" 2>/dev/null)" == "state=ACTIVE" && -f "$f" ]] || return 1
  [[ "$(LC_ALL=C sort "$f")" == "$(printf '%s\n' "attributeCondition=$COND" "attributeMapping=${MAPPING//,/;}" "oidc.allowedAudiences=${AUDS//,/;}" \
    "oidc.issuerUri=https://gitlab.com" "state=ACTIVE" | LC_ALL=C sort)" ]]
}
bucket_ok() {
  local f="$1/res/bucket/$META"
  [[ -f "$f" && "$(LC_ALL=C sort "$f")" == "$(printf '%s\n' "labels=managed-by=ci-federation" "location=$R" \
    'lifecycle={"rule":[{"action":{"type":"Delete"},"condition":{"isLive":false,"daysSinceNoncurrentTime":90}}]}' \
    "public_access_prevention=enforced" "uniform_bucket_level_access=True" "versioning_enabled=True" | LC_ALL=C sort)" ]]
}
snap_equal() { printf '%s\n' "$1" >"$3.pre"; printf '%s\n' "$2" >"$3.post"; [[ "$RC" == 0 ]] && cmp -s "$3.pre" "$3.post"; }

record() { printf '%s\t%s\t%s\n' "$1" "$2" "${3//$'\t'/ }" >>"$REPORT/cases.tsv"; printf '%s %s %s\n' "$2" "$1" "$3"; }
PASSED=0 FAILED=0
case_pass() { record "$1" PASS "$2"; PASSED=$((PASSED + 1)); }
case_fail() { record "$1" FAIL "$2"; FAILED=$((FAILED + 1)); }
ct() { local n=$1 d=$2; shift 2; if "$@"; then case_pass "$n" "$d"; else case_fail "$n" "$d"; fi; }

# ---- base tree: the whole suite ------------------------------------------------------------------------------------------
BASE="$S/tree-base"; make_tree "$BASE"
suite "$BASE" base full
while IFS=$'\t' read -r r n d; do if [[ $r == PASS ]]; then case_pass "$n" "$d"; else case_fail "$n" "$d"; fi; done <"$TR/base/assertions.tsv"
ct registry "$REGISTRY_NOTE" true
ct pins.readable "the real infra/ci/hosted-context.sh holds exactly one STG_PROJECT_ID=<digits> and one STG_PROJECT_PATH_SHA=<64 hex> line (the script's pin format)" \
  test "$(grep -cE '^STG_PROJECT_ID=[1-9][0-9]*$' "$REPO/infra/ci/hosted-context.sh" || true)" = 1 -a \
  "$(grep -cE '^STG_PROJECT_PATH_SHA=[0-9a-f]{64}$' "$REPO/infra/ci/hosted-context.sh" || true)" = 1

# ---- verify ---------------------------------------------------------------------------------------------------------------
CUR_TR="$TR/verify"; mkdir -p "$CUR_TR"
VS="$S/state-verify"; init_state "$VS"
fed "$BASE" "$VS" setup "${COMMON[@]}" "${GL[@]}"
fed "$BASE" "$VS" verify "${COMMON[@]}" "${GL[@]}" --verify
ct verify.pass "--verify after setup exits 0 (got $RC), every item ok, read calls only ($(nwrites "$CALLS") writes)" \
  test "$RC" = 0 -a "$(nwrites "$CALLS")" = 0 -a "$(grep -c $'\tMISMATCH' "$OUT" || true)" = 0 -a "$(grep -cx $'ci-federation\tverify\tpreview-1\tok' "$OUT" || true)" = 1
tamper_verify() { # NAME DETAIL EXPECT-ITEM MUTATION...: tamper a copy of the verified state; --verify must fail on that item
  local name=$1 detail=$2 item=$3 st="$S/state-verify-$1"; shift 3
  rm -rf "$st"; cp -a "$VS" "$st"; "$@" "$st"
  fed "$BASE" "$st" "verify-$name" "${COMMON[@]}" "${GL[@]}" --verify
  ct "verify.$name" "$detail: --verify exits 1 (got $RC) with MISMATCH on $item and no write call" \
    test "$RC" = 1 -a "$(nwrites "$CALLS")" = 0 -a "$(grep -cF -- $'\tverify\t'"$item"$'\tMISMATCH' "$OUT" || true)" = 1
}
tv_state_bucket() { printf '%s\n' "bucket$SEP$STATE${SEP}roles/storage.objectViewer${SEP}serviceAccount:$SA_STOP$SEP" >>"$1/bindings"; }
tv_condition() { sed -i 's/ && assertion.environment_protected == "true"//' "$1/res/provider/$E-ci--gitlab"; }
tv_key() { echo "projects/$P/serviceAccounts/$SA_STOP/keys/0123abcd" >"$1/keys/$SA_STOP"; }
tv_project_role() { printf '%s\n' "project$SEP$P${SEP}roles/editor${SEP}serviceAccount:$SA_READ$SEP" >>"$1/bindings"; }
tv_secret() { printf '%s\n' "secret$SEP$E-CF_API_TOKEN${SEP}roles/secretmanager.secretAccessor$SEP$PS/stop$SEP" >>"$1/bindings"; }
tamper_verify state-bucket "a ci-stop objectViewer binding on the state bucket" "bindings bucket $STATE" tv_state_bucket
tamper_verify condition "the environment_protected clause removed from the provider" "provider.attributeCondition" tv_condition
tamper_verify key "a user-managed key on ci-stop" "user-managed keys $SA_STOP" tv_key
tamper_verify project-role "roles/editor for ci-read on the project" "bindings project $P" tv_project_role
tv_iap_projectwide() { # the instance-conditioned tunnel grant replaced by an unconditioned (project-wide) one
  awk -F"$SEP" -v OFS="$SEP" '$3 == "roles/iap.tunnelResourceAccessor" {$5 = ""} {print}' "$1/bindings" >"$1/bindings.n"; mv "$1/bindings.n" "$1/bindings"
}
tamper_verify iap-projectwide "iap.tunnelResourceAccessor for ci-stop without the instance condition" "bindings project $P" tv_iap_projectwide
tamper_verify secret "the stop principal granted another secret" "bindings secret $E-CF_API_TOKEN" tv_secret
fed "$BASE" "$S/state-verify-dry" verify-dry "${COMMON[@]}" "${GL[@]}" --verify --dry-run
ct verify.dry-run "--verify --dry-run exits 0 (got $RC), calls no gcloud, prints only read commands (describe/get-iam-policy/list/get-value), including the state-bucket policy and both key listings" \
  test "$RC" = 0 -a "$(ncalls "$CALLS")" = 0 -a "$(grep -c '^+ ' "$OUT" || true)" = 0 -a \
  "$(grep '^? ' "$OUT" | grep -cvE "^\? gcloud --project $P --quiet ([a-z-]+ )+(describe|get-iam-policy|list) " || true)" = 0 -a \
  "$(grep -cF -- "? gcloud --project $P --quiet storage buckets get-iam-policy gs://$STATE " "$OUT" || true)" = 1 -a \
  "$(grep -c '^? .* service-accounts keys list --iam-account=' "$OUT" || true)" = 2

# ---- teardown: re-setup after teardown (undelete), confirm-meta -----------------------------------------------------------
CUR_TR="$TR/teardown"; mkdir -p "$CUR_TR"
US="$S/state-undelete"; init_state "$US"
fed "$BASE" "$US" setup "${COMMON[@]}" "${GL[@]}"
fed "$BASE" "$US" teardown "${COMMON[@]}" --teardown --confirm "$P"
fed "$BASE" "$US" resetup "${COMMON[@]}" "${GL[@]}"
r1=$RC u="$(grep -cE ' undelete ' "$CALLS" || true)"
fed "$BASE" "$US" reverify "${COMMON[@]}" "${GL[@]}" --verify
ct teardown.resetup "setup after teardown exits 0 (got $r1), undeletes the soft-deleted pool, provider and both roles ($u undelete calls), then verifies (exit $RC)" \
  test "$r1" = 0 -a "$u" = 4 -a "$RC" = 0
LS="$S/state-confirm"; init_state "$LS"
fed "$BASE" "$LS" setup "${COMMON[@]}" "${GL[@]}"
printf '%s\n' current.tsv lease >"$LS/objects/$META"
fed "$BASE" "$LS" teardown-wrong-meta "${COMMON[@]}" --teardown --confirm "$P" --confirm-meta staging
ct teardown.confirm-meta-wrong "--confirm-meta naming another environment is refused (exit $RC) with no gcloud call" test "$RC" != 0 -a "$(ncalls "$CALLS")" = 0
fed "$BASE" "$LS" teardown-confirmed "${COMMON[@]}" --teardown --confirm "$P" --confirm-meta "$E"
ct teardown.confirm-meta "--confirm-meta $E removes the bucket with its live objects (exit $RC)" \
  test "$RC" = 0 -a ! -e "$LS/res/bucket/$META" -a ! -e "$LS/objects/$META"

# ---- metadata-bucket owner and failed reads ----------------------------------------------------------------------------------
# Each case is a function of the tree (TREE) so that the controls below can run it against the old behaviour; WHY is
# its observed detail. Faults are realistic gcloud error lines injected for one read.
CUR_TR="$TR/ownership"; mkdir -p "$CUR_TR"
FOREIGN=999999999999
D403="ERROR: (gcloud.storage.buckets.describe) HTTPError 403: owner@example.net does not have storage.buckets.get access to the Google Cloud Storage bucket. Permission 'storage.buckets.get' denied on resource (or it may not exist)."
L403="ERROR: (gcloud.storage.ls) HTTPError 403: owner@example.net does not have storage.objects.list access to the Google Cloud Storage bucket."
L404="ERROR: (gcloud.storage.ls) gs://$META not found: 404."
WHY=""
fault() { printf '%s\t%s\t%s\n' "$2" "$3" "${4:-}" >>"$1/faults"; } # STATE PREFIX MESSAGE [FIRST-N-CALLS]
mkbucket() { # STATE OWNER: an accessible pre-existing gs://META (owner "none": no projectNumber field)
  printf 'location=%s\nuniform_bucket_level_access=True\npublic_access_prevention=enforced\nversioning_enabled=False\n' "$R" >"$1/res/bucket/$META"
  : >"$1/objects/$META"; mkdir -p "$1/owner"; printf '%s\n' "$2" >"$1/owner/$META"
}
st_of() { printf '%s/state-own-%s-%s' "$S" "$1" "${2##*/}"; } # NAME TREE
refused_unchanged() { # STATE PRE MESSAGE: non-zero exit, zero write calls, state unchanged, MESSAGE in the output
  WHY="exit $RC, $(nwrites "$CALLS") writes"
  [[ $RC != 0 && "$(nwrites "$CALLS")" == 0 && "$(full_snapshot "$1")" == "$2" ]] && grep -qF -- "$3" "$OUT"
}
setup_owner() { # TREE NAME OWNER MESSAGE: setup over a pre-existing bucket with that owner is refused before any change
  local st pre; st="$(st_of "setup-$2" "$1")"; init_state "$st"; mkbucket "$st" "$3"; pre="$(full_snapshot "$st")"
  fed "$1" "$st" "setup-$2" "${COMMON[@]}" "${GL[@]}"
  refused_unchanged "$st" "$pre" "$4"
}
c_setup_foreign() { setup_owner "$1" foreign "$FOREIGN" "belongs to project number $FOREIGN, not $P ($N)"; }
c_setup_missing() { setup_owner "$1" missing none "reports no readable owner project number"; }
c_setup_malformed() { setup_owner "$1" malformed 12ab "reports no readable owner project number"; }
c_setup_denied() {
  local st pre; st="$(st_of setup-denied "$1")"; init_state "$st"; mkbucket "$st" "$N"; fault "$st" "storage buckets describe" "$D403"
  pre="$(full_snapshot "$st")"
  fed "$1" "$st" setup-denied "${COMMON[@]}" "${GL[@]}"
  refused_unchanged "$st" "$pre" "gs://$META could not be checked"
}
c_setup_matching() {
  local st; st="$(st_of setup-matching "$1")"; init_state "$st"; mkbucket "$st" "$N"
  fed "$1" "$st" setup-matching "${COMMON[@]}" "${GL[@]}"
  WHY="exit $RC"
  [[ $RC == 0 ]] && grep -qxF -- $'ci-federation\tmeta-bucket\tgs://'"$META"$'\texisting' "$OUT" && bucket_ok "$st" \
    && ! grep -qF -- "storage buckets create" "$CALLS"
}
set_up() { init_state "$1"; fed "$BASE" "$1" "$2" "${COMMON[@]}" "${GL[@]}"; } # STATE NAME: a set-up environment
teardown_refused() { # TREE NAME MESSAGE MUTATION...: teardown (live objects confirmed) refused before any change
  local tree=$1 name=$2 msg=$3 st pre; shift 3
  st="$(st_of "teardown-$name" "$tree")"; set_up "$st" "teardown-$name-setup"
  printf '%s\n' current.tsv lease >"$st/objects/$META"; "$@" "$st"; pre="$(full_snapshot "$st")"
  fed "$tree" "$st" "teardown-$name" "${COMMON[@]}" --teardown --confirm "$P" --confirm-meta "$E"
  refused_unchanged "$st" "$pre" "$msg"
}
m_foreign() { mkdir -p "$1/owner"; echo "$FOREIGN" >"$1/owner/$META"; }
m_describe_503() { fault "$1" "storage buckets describe" "ERROR: (gcloud.storage.buckets.describe) HTTPError 503: backend error"; }
m_list_403() { fault "$1" "storage ls" "$L403"; }
m_list_404() { fault "$1" "storage ls" "$L404"; }
m_sa_policy_503() { fault "$1" "iam service-accounts get-iam-policy" "ERROR: (gcloud.iam.service-accounts.get-iam-policy) UNAVAILABLE: The service is currently unavailable."; }
c_teardown_foreign() { teardown_refused "$1" foreign "belongs to project number $FOREIGN" m_foreign; }
c_teardown_describe_failed() { teardown_refused "$1" describe-failed "gs://$META could not be checked" m_describe_503; }
c_teardown_list_denied() { teardown_refused "$1" list-denied "could not be listed" m_list_403; }
c_teardown_list_404() { teardown_refused "$1" list-404 "could not be listed" m_list_404; }
c_teardown_policy_failed() { teardown_refused "$1" policy-failed "failed and is not a not-found answer" m_sa_policy_503; }
c_teardown_list_denied_unconfirmed() { # without --confirm-meta, over an empty bucket: still refused, not "no live objects"
  local st pre; st="$(st_of teardown-list-unconfirmed "$1")"; set_up "$st" teardown-list-unconfirmed-setup
  m_list_403 "$st"; pre="$(full_snapshot "$st")"
  fed "$1" "$st" teardown-list-unconfirmed "${COMMON[@]}" --teardown --confirm "$P"
  refused_unchanged "$st" "$pre" "could not be listed"
}
c_teardown_absent() { # a confirmed 404 is absent: teardown of a bootstrap-only project changes nothing and succeeds
  local st; st="$(st_of teardown-absent "$1")"; init_state "$st"
  fed "$1" "$st" teardown-absent "${COMMON[@]}" --teardown --confirm "$P"
  WHY="exit $RC, $(nwrites "$CALLS") writes"
  [[ $RC == 0 && "$(nwrites "$CALLS")" == 0 ]] && grep -qxF -- $'ci-federation\tmeta-bucket\tgs://'"$META"$'\tabsent' "$OUT"
}
c_setup_pool_unreadable() { # a failed (non-404) pool describe stops setup; it never creates the pool or anything after it
  local st; st="$(st_of setup-pool "$1")"; init_state "$st"
  fault "$st" "iam workload-identity-pools describe" "ERROR: (gcloud.iam.workload-identity-pools.describe) PERMISSION_DENIED: Permission 'iam.workloadIdentityPools.get' denied on resource (or it may not exist)."
  fed "$1" "$st" setup-pool "${COMMON[@]}" "${GL[@]}"
  WHY="exit $RC, $(grep -cE ' (create|create-oidc|undelete) ' "$CALLS" || true) create calls"
  [[ $RC != 0 && "$(grep -cE ' (create|create-oidc|undelete) ' "$CALLS" || true)" == 0 ]] && grep -qF -- "failed and is not a not-found answer" "$OUT"
}
verify_unreadable() { # TREE NAME PREFIX MESSAGE ITEM: one failed read; --verify exits 1, ITEM UNREADABLE, no ok for it
  local st="$S/state-own-verify-$2-${1##*/}"
  rm -rf "$st"; cp -a "$VS" "$st"; fault "$st" "$3" "$4"
  fed "$1" "$st" "verify-$2" "${COMMON[@]}" "${GL[@]}" --verify
  WHY="exit $RC, $(nwrites "$CALLS") writes"
  [[ $RC == 1 && "$(nwrites "$CALLS")" == 0 ]] && grep -qxF -- $'ci-federation\tverify\t'"$5"$'\tUNREADABLE' "$OUT" \
    && ! grep -qxF -- $'ci-federation\tverify\t'"$5"$'\tok' "$OUT" && ! grep -qx $'ci-federation\tverify\tpreview-1\tok' "$OUT" \
    && grep -qx $'ci-federation\tverify\tpreview-1\tFAILED' "$OUT"
}
c_verify_keys() { verify_unreadable "$1" keys "iam service-accounts keys list" \
  "ERROR: (gcloud.iam.service-accounts.keys.list) PERMISSION_DENIED: Permission 'iam.serviceAccountKeys.list' denied on resource (or it may not exist)." \
  "user-managed keys $SA_STOP"; }
c_verify_state_policy() { verify_unreadable "$1" state-policy "storage buckets get-iam-policy gs://$STATE" \
  "ERROR: (gcloud.storage.buckets.get-iam-policy) HTTPError 403: owner@example.net does not have storage.buckets.getIamPolicy access to the Google Cloud Storage bucket." \
  "bindings bucket $STATE"; }
c_verify_secrets() { verify_unreadable "$1" secrets "secrets list" \
  "ERROR: (gcloud.secrets.list) PERMISSION_DENIED: Permission 'secretmanager.secrets.list' denied for resource 'projects/$P'." "secrets list"; }
c_verify_foreign() { # a foreign owner fails verification on its own item
  local st="$S/state-own-verify-foreign-${1##*/}"
  rm -rf "$st"; cp -a "$VS" "$st"; m_foreign "$st"
  fed "$1" "$st" verify-foreign "${COMMON[@]}" "${GL[@]}" --verify
  WHY="exit $RC"
  [[ $RC == 1 ]] && grep -qxF -- $'ci-federation\tverify\tmeta-bucket owner\tMISMATCH' "$OUT"
}
nc() { # NAME DETAIL FN: the case on the base tree
  if "$3" "$BASE"; then case_pass "$1" "$2 ($WHY)"; else case_fail "$1" "$2 ($WHY)"; fi
}
nc owner.setup-foreign "setup over gs://$META owned by project number $FOREIGN: refused, zero write calls, nothing changed" c_setup_foreign
nc owner.setup-missing "setup over gs://$META without an owner project number: refused, zero write calls" c_setup_missing
nc owner.setup-malformed "setup over gs://$META with a malformed owner (12ab): refused, zero write calls" c_setup_malformed
nc owner.setup-describe-denied "a denied (403) bucket describe is not absence: setup refused, zero write calls" c_setup_denied
nc owner.setup-matching "setup over a pre-existing gs://$META owned by $N: exit 0, existing (not created), settings converge" c_setup_matching
nc owner.teardown-foreign "teardown with gs://$META owned by $FOREIGN (live objects confirmed): refused, zero write calls, nothing deleted" c_teardown_foreign
nc owner.teardown-describe-failed "teardown with a failed (503) bucket describe: refused, zero write calls" c_teardown_describe_failed
nc owner.teardown-absent "a confirmed 404 is absent: teardown of a bootstrap-only project exits 0 with zero write calls, meta-bucket absent" c_teardown_absent
nc teardown.list-denied "a denied (403) live-object listing (even with --confirm-meta): refused, zero write calls, nothing deleted" c_teardown_list_denied
nc teardown.list-404 "a 404 listing of the bucket just described as present: refused, zero write calls" c_teardown_list_404
nc teardown.list-denied-unconfirmed "a denied listing of an empty bucket without --confirm-meta is not 'no live objects': refused, zero write calls" c_teardown_list_denied_unconfirmed
nc teardown.policy-unreadable "a failed (503) policy read is not an absent binding: teardown stops with zero write calls" c_teardown_policy_failed
nc setup.pool-unreadable "a denied pool describe is not absence: setup stops before creating anything" c_setup_pool_unreadable
nc verify.keys-unreadable "a failed user-managed key listing never satisfies the expected-empty check" c_verify_keys
nc verify.state-policy-unreadable "a failed state-bucket IAM read never satisfies its expected-empty binding set" c_verify_state_policy
nc verify.secrets-unreadable "a failed secrets listing fails verification instead of skipping the unexpected-secret checks" c_verify_secrets
nc verify.owner-foreign "gs://$META owned by $FOREIGN: --verify exits 1 with MISMATCH on meta-bucket owner" c_verify_foreign

# A status marker, not a substring: a legal project ID containing -404- (so the bucket, accounts and pool too) must not
# turn a 503 that echoes a name or URL into a not-found; a genuine 404 stays absent.
use_project() { # PROJECT-ID: the project-derived fixture names
  P=$1 META="$1-$E-ci-meta" STATE="$1-tfstate"
  SA_DEP="$E-deployer@$1.iam.gserviceaccount.com" SA_VM="$E-vm@$1.iam.gserviceaccount.com"
  SA_STOP="$E-ci-stop@$1.iam.gserviceaccount.com" SA_READ="$E-ci-read@$1.iam.gserviceaccount.com"
  COMMON=(--project "$1" --project-number "$N" --region "$R" --zone "$Z" --env "$E")
  L404="ERROR: (gcloud.storage.ls) gs://$META not found: 404."
}
D503() { printf 'ERROR: (gcloud.storage.buckets.describe) HTTPError 503: backend error reading gs://%s' "$META"; }
c_teardown_recheck() { # absent at the first owner check, present at the second: never deleted unlisted
  local st; st="$(st_of teardown-recheck "$1")"; set_up "$st" teardown-recheck-setup
  printf '%s\n' current.tsv lease >"$st/objects/$META"
  fault "$st" "storage buckets describe" "ERROR: (gcloud.storage.buckets.describe) gs://$META not found: 404." 1
  fed "$1" "$st" teardown-recheck "${COMMON[@]}" --teardown --confirm "$P"
  WHY="exit $RC, $(grep -cF 'storage rm ' "$CALLS" || true) rm calls"
  [[ $RC != 0 && "$(grep -cF 'storage rm ' "$CALLS" || true)" == 0 && "$(cat "$st/objects/$META" 2>/dev/null)" == $'current.tsv\nlease' ]] \
    && grep -qF -- "was absent when teardown started and is present now" "$OUT"
}
c_p404_id() { # the project-ID pattern accepts the -404- project; its bucket name carries -404-
  local st; st="$(st_of p404-id "$1")"; init_state "$st"
  fed "$1" "$st" p404-id "${COMMON[@]}" "${GL[@]}" --dry-run
  WHY="exit $RC, bucket $META"
  [[ $RC == 0 && $P =~ ^[a-z][a-z0-9-]{4,28}[a-z0-9]$ && $META == *-404-* ]] && grep -qF -- "storage buckets create gs://$META " "$OUT"
}
c_p404_teardown_503() { # first describe only: 503 echoing gs://<bucket with -404->; live objects, no --confirm-meta
  local st pre; st="$(st_of p404-teardown-503 "$1")"; set_up "$st" p404-teardown-503-setup
  printf '%s\n' current.tsv lease >"$st/objects/$META"; fault "$st" "storage buckets describe" "$(D503)" 1; pre="$(full_snapshot "$st")"
  fed "$1" "$st" p404-teardown-503 "${COMMON[@]}" --teardown --confirm "$P"
  refused_unchanged "$st" "$pre" "gs://$META could not be checked" && [[ "$(cat "$st/objects/$META")" == $'current.tsv\nlease' ]]
}
c_p404_setup_503() {
  local st pre; st="$(st_of p404-setup-503 "$1")"; init_state "$st"; mkbucket "$st" "$N"
  fault "$st" "storage buckets describe" "$(D503)" 1; pre="$(full_snapshot "$st")"
  fed "$1" "$st" p404-setup-503 "${COMMON[@]}" "${GL[@]}"
  refused_unchanged "$st" "$pre" "gs://$META could not be checked"
}
c_p404_verify_503() {
  local st; st="$(st_of p404-verify-503 "$1")"; set_up "$st" p404-verify-503-setup
  fault "$st" "storage buckets describe gs://$META" "$(D503)" 1
  fed "$1" "$st" p404-verify-503 "${COMMON[@]}" "${GL[@]}" --verify
  WHY="exit $RC, $(nwrites "$CALLS") writes"
  [[ $RC == 1 && "$(nwrites "$CALLS")" == 0 ]] && grep -qxF -- $'ci-federation\tverify\tmeta-bucket owner\tUNREADABLE' "$OUT" \
    && grep -qx $'ci-federation\tverify\tpreview-1\tFAILED' "$OUT"
}
c_p404_teardown_policy_503() { # probe/value_of: a 503 echoing the account (and its projects/ path) is no absent binding
  local st pre; st="$(st_of p404-policy-503 "$1")"; set_up "$st" p404-policy-503-setup
  fault "$st" "iam service-accounts get-iam-policy" "ERROR: (gcloud.iam.service-accounts.get-iam-policy) HTTPError 503: backend error for $SA_DEP (projects/$P/serviceAccounts/$SA_DEP)"
  pre="$(full_snapshot "$st")"
  fed "$1" "$st" p404-policy-503 "${COMMON[@]}" --teardown --confirm "$P"
  refused_unchanged "$st" "$pre" "failed and is not a not-found answer"
}
c_p404_teardown_absent() { # a genuine 404 for the -404- bucket is still absent
  local st; st="$(st_of p404-absent "$1")"; init_state "$st"
  fed "$1" "$st" p404-absent "${COMMON[@]}" --teardown --confirm "$P"
  WHY="exit $RC, $(nwrites "$CALLS") writes"
  [[ $RC == 0 && "$(nwrites "$CALLS")" == 0 ]] && grep -qxF -- $'ci-federation\tmeta-bucket\tgs://'"$META"$'\tabsent' "$OUT"
}
c_p404_teardown_nomatch_uri() { # "matched no objects" echoing the -404- URL is still confirmed empty
  local st; st="$(st_of p404-nomatch "$1")"; set_up "$st" p404-nomatch-setup
  fault "$st" "storage ls" "ERROR: (gcloud.storage.ls) One or more URLs matched no objects: gs://$META/**" 1
  fed "$1" "$st" p404-nomatch "${COMMON[@]}" --teardown --confirm "$P"
  WHY="exit $RC"
  [[ $RC == 0 && ! -e "$st/res/bucket/$META" ]]
}
nc teardown.owner-recheck "absent at the first owner check (404), present before deletion: refused, no storage rm, live objects kept" c_teardown_recheck

# The owner changes during setup (the bucket deleted and its global name reused): every bucket write is rechecked.
flip() { printf '%s\t%s\t%s\n' "$2" "$META" "$3" >>"$1/flips"; } # STATE PREFIX OWNER
n_upd() { grep -cF -- "storage buckets update gs://$META " "$CALLS" || true; }
n_grant() { grep -cF -- "storage buckets add-iam-policy-binding gs://$META " "$CALLS" || true; }
race_setup() { # TREE NAME PRE-EXISTING(yes|no) FLIP-PREFIX OWNER: setup with the owner flipped at FLIP-PREFIX's call
  local st; st="$(st_of "race-$2" "$1")"; init_state "$st"
  [[ $3 == no ]] || mkbucket "$st" "$N"
  flip "$st" "$4" "$5"
  fed "$1" "$st" "race-$2" "${COMMON[@]}" "${GL[@]}"
  WHY="exit $RC, $(n_upd) bucket updates, $(n_grant) bucket grants"
}
c_race_before_update() { # owned at the first check, foreign from the API enable on: no update, no grant
  race_setup "$1" before-update yes "services enable" "$FOREIGN"
  [[ $RC != 0 && "$(n_upd)" == 0 && "$(n_grant)" == 0 ]] && grep -qF -- "belongs to project number $FOREIGN" "$OUT"
}
c_race_before_grants() { # foreign from the settings update on (the update itself was checked): no grant
  race_setup "$1" before-grants yes "storage buckets update" "$FOREIGN"
  [[ $RC != 0 && "$(n_upd)" == 1 && "$(n_grant)" == 0 ]] && grep -qF -- "belongs to project number $FOREIGN" "$OUT"
}
c_race_after_create() { # absent, created by this run, but reported foreign right after: no update, no grant
  race_setup "$1" after-create no "storage buckets create" "$FOREIGN"
  [[ $RC != 0 && "$(n_upd)" == 0 && "$(n_grant)" == 0 ]] && grep -qF -- "belongs to project number $FOREIGN" "$OUT"
}
c_race_matching() { # the owner rewritten with the same number: setup succeeds, with a fresh check before each bucket write
  local checks
  race_setup "$1" matching yes "services enable" "$N"
  checks="$(grep -cF -- "storage buckets describe gs://$META --raw" "$CALLS" || true)"
  WHY+=", $checks owner checks"
  [[ $RC == 0 && "$(n_upd)" == 1 && "$(n_grant)" == 8 && $checks == 10 ]]
}
c_create_taken() { # absent at the check (404), the name taken (foreign) by the create: refused, not adopted
  local st; st="$(st_of create-taken "$1")"; init_state "$st"; mkbucket "$st" "$FOREIGN"
  fault "$st" "storage buckets describe" "ERROR: (gcloud.storage.buckets.describe) gs://$META not found: 404." 1
  fed "$1" "$st" create-taken "${COMMON[@]}" "${GL[@]}"
  WHY="exit $RC, $(n_upd) bucket updates, $(n_grant) bucket grants"
  [[ $RC != 0 && "$(n_upd)" == 0 && "$(n_grant)" == 0 ]] && grep -qF -- "refused, not adopted" "$OUT"
}
nc race.before-update "owned at the first check, owned by $FOREIGN after the API enable: setup refused, no bucket update, no bucket grant" c_race_before_update
nc race.before-grants "owned through the settings update, owned by $FOREIGN after it: setup refused before any bucket grant" c_race_before_grants
nc race.after-create "created by this run but reported as $FOREIGN's right after: refused, no update, no grant" c_race_after_create
nc race.matching "the owner unchanged ($N) across the run: setup exits 0, one fresh owner check before the update and each of the 8 bucket grants" c_race_matching
nc race.create-taken "absent at the check, the name taken by $FOREIGN before the create: refused (not adopted), no update, no grant" c_create_taken
use_project demo-404-preview
nc p404.project-id "the project-ID pattern accepts demo-404-preview; its bucket gs://$META carries -404- (setup --dry-run renders it)" c_p404_id
nc p404.teardown-503-uri "a first-describe 503 naming gs://$META, live objects, no --confirm-meta: refused before any write, objects kept" c_p404_teardown_503
nc p404.setup-503-uri "setup with a first-describe 503 naming gs://$META: refused, zero write calls" c_p404_setup_503
nc p404.verify-503-uri "--verify with a 503 naming gs://$META on the owner read: meta-bucket owner UNREADABLE, FAILED" c_p404_verify_503
nc p404.teardown-policy-503 "a 503 policy read naming $SA_DEP is not an absent binding: teardown stops with zero write calls" c_p404_teardown_policy_503
nc p404.teardown-404-absent "a genuine 404 for gs://$META is absent: teardown exits 0 with zero write calls" c_p404_teardown_absent
nc p404.teardown-nomatch-uri "'matched no objects' naming gs://$META/** is confirmed empty: teardown removes the bucket" c_p404_teardown_nomatch_uri
use_project example-preview-1

# ---- inputs ----------------------------------------------------------------------------------------------------------------
CUR_TR="$TR/inputs"; mkdir -p "$CUR_TR"
refuse() { # NAME DETAIL ARGS...: exit non-zero, no write call
  local name=$1 detail=$2 st="$S/state-input-$1"; shift 2
  init_state "$st"
  fed "$BASE" "$st" "$name" "$@"
  ct "inputs.$name" "$detail: exit $RC, $(nwrites "$CALLS") write calls" test "$RC" != 0 -a "$(nwrites "$CALLS")" = 0
}
args_with() { # KEY VALUE: COMMON+GL with one value replaced
  local k=$1 v=$2 a=("${COMMON[@]}" "${GL[@]}") i
  for ((i = 0; i < ${#a[@]}; i++)); do [[ "${a[i]}" != "$k" ]] || a[i + 1]=$v; done
  printf '%s\n' "${a[@]}"
}
mapfile -t A < <(args_with --env Preview-1); refuse env-case "--env Preview-1 (pattern)" "${A[@]}"
mapfile -t A < <(args_with --env preview-10); refuse env-unlisted "--env preview-10 (not in the registry)" "${A[@]}"
mapfile -t A < <(args_with --env preview-01); refuse env-padded "--env preview-01 (not in the registry)" "${A[@]}"
mapfile -t A < <(args_with --gitlab-project-path example-group/other-project); refuse project-path "a project path whose hash differs from the pin" "${A[@]}"
mapfile -t A < <(args_with --gitlab-namespace-id 42a); refuse namespace "a non-numeric namespace ID" "${A[@]}"
mapfile -t A < <(args_with --project-number 123456789013); refuse project-number "a project number that differs from the project's" "${A[@]}"
mapfile -t A < <(args_with --zone us-west1-a); refuse zone "a zone outside --region" "${A[@]}"
refuse teardown-confirm "--teardown without --confirm PROJECT" "${COMMON[@]}" --teardown
refuse teardown-verify "--teardown together with --verify" "${COMMON[@]}" --teardown --verify --confirm "$P"
refuse confirm-on-setup "--confirm-meta on setup" "${COMMON[@]}" "${GL[@]}" --confirm-meta "$E"
st="$S/state-input-unknown"; init_state "$st"; fed "$BASE" "$st" unknown-arg "${COMMON[@]}" "${GL[@]}" --images build
ct inputs.unknown-arg "an unknown argument is refused (exit $RC) before any gcloud call" test "$RC" != 0 -a "$(ncalls "$CALLS")" = 0
fed "$BASE" "$st" help --help
ct inputs.help "--help prints the usage (exit $RC = 2)" bash -c '[[ $1 == 2 ]] && grep -q -- "--teardown --confirm ID" "$2"' _ "$RC" "$OUT"

# ---- permissions file --------------------------------------------------------------------------------------------------------
CUR_TR="$TR/permissions"; mkdir -p "$CUR_TR"
perm_tree() { # NAME SED-EXPR: a tree whose permissions file is edited
  local t="$S/tree-perm-$1"; make_tree "$t"; sed -i "$2" "$t/infra/terraform/bootstrap/ci-meta-permissions.tsv"
  cmp -s "$SRC_PERMS" "$t/infra/terraform/bootstrap/ci-meta-permissions.tsv" && { echo "mutation $1 not applied" >&2; return 1; }
  printf '%s\n' "$t"
}
perm_refused() { # NAME DETAIL SED-EXPR
  local t; t="$(perm_tree "$1" "$3")" || { case_fail "permissions.$1" "mutation not applied"; return; }
  fed "$t" "$S/state-perm-$1" "$1" "${COMMON[@]}" "${GL[@]}" --dry-run
  ct "permissions.$1" "$2: refused (exit $RC) with no rendering" test "$RC" != 0 -a "$(grep -c '^+ ' "$OUT" || true)" = 0
}
perm_refused wildcard-mutable "ciMetaMutable with keys * (the exact-key condition dropped)" 's/^ci-stop\tciMetaMutable\t.*/ci-stop\tciMetaMutable\t*/'
perm_refused read-writes "ci-read granted roles/storage.objectCreator" '$a ci-read\troles/storage.objectCreator\t*'
perm_refused stop-admin "ci-stop granted roles/storage.objectAdmin" '$a ci-stop\troles/storage.objectAdmin\t*'
perm_refused storage-admin "a grant outside the contract (roles/storage.admin)" '$a deployer\troles/storage.admin\t*'
perm_refused unknown-principal "an unknown principal" '$a ci-other\troles/storage.objectViewer\t*'
perm_refused shape "a row separated by spaces" 's/^ci-read\troles\/storage.objectViewer\t\*$/ci-read roles\/storage.objectViewer */'
perm_refused repeated "a repeated principal/grant" '$a ci-read\troles/storage.legacyBucketReader\t*'
perm_refused predefined-keys "a predefined role limited to keys" 's/^ci-read\troles\/storage.objectViewer\t\*$/ci-read\troles\/storage.objectViewer\tcurrent.tsv/'
single_source() { # NAME DETAIL SED-EXPR WANT-ADDED WANT-REMOVED
  local t; t="$(perm_tree "$1" "$3")" || { case_fail "permissions.$1" "mutation not applied"; return; }
  fed "$t" "$S/state-perm-$1" "$1" "${COMMON[@]}" "${GL[@]}" --dry-run
  plus_lines >"$CUR_TR/$1.plus"
  ct "permissions.$1" "$2: exit $RC; the rendering differs from the reviewed one by exactly that binding" \
    test "$RC" = 0 -a "$(diff <(oracle_setup) "$CUR_TR/$1.plus" | grep -E '^[<>] ' | sed 's/^> /+/; s/^< /-/' | LC_ALL=C sort | tr '\n' '|')" = "$( { [[ -z "$4" ]] || echo "+$4"; [[ -z "$5" ]] || echo "-$5"; } | LC_ALL=C sort | tr '\n' '|')"
}
single_source added-row "an added row (deployer objectViewer)" '$a deployer\troles/storage.objectViewer\t*' \
  "$G storage buckets add-iam-policy-binding gs://$META --member=serviceAccount:$SA_DEP --role=roles/storage.objectViewer --condition=None --format=none" ""
single_source removed-row "the ci-read objectViewer row removed" '/^ci-read\troles\/storage.objectViewer\t\*$/d' \
  "" "$G storage buckets add-iam-policy-binding gs://$META --member=serviceAccount:$SA_READ --role=roles/storage.objectViewer --condition=None --format=none"

# ---- mutations: each must fail the suite --------------------------------------------------------------------------------------
CUR_TR="$TR/mutations"; mkdir -p "$CUR_TR"
mutant() { # NAME LEVEL GUARD(on|off) FILE(script|perms) SED-EXPR [DETAIL]
  local name=$1 level=$2 guard=$3 which=$4 expr=$5 detail=${6:-} t="$S/tree-mut-$1" f
  make_tree "$t"
  if [[ $which == script ]]; then f="$t/infra/terraform/bootstrap/ci-federation.sh"; else f="$t/infra/terraform/bootstrap/ci-meta-permissions.tsv"; fi
  sed -i -e "$expr" "$f"
  if cmp -s "$f" "${SRC_SCRIPT%/*}/${f##*/}"; then case_fail "mutation.$name" "mutation not applied (sed matched nothing)"; return; fi
  if [[ $guard == off ]]; then sed -i 's/^forbid() {$/forbid() { return 0; }\nforbid_disabled() {/' "$t/infra/terraform/bootstrap/ci-federation.sh"; fi
  suite "$t" "mut-$name" "$level"
  if ((CUR_FAILS > 0)); then
    case_pass "mutation.$name" "${detail:+$detail; }killed: $(awk -F'\t' '$1 == "FAIL" {printf "%s ", $2}' "$CUR_LOG")(guard $guard)"
  else case_fail "mutation.$name" "${detail:+$detail; }survived every assertion (guard $guard)"; fi
}
A_STATE="  run g storage buckets add-iam-policy-binding \"gs://\$state_bucket\" --member=\"serviceAccount:\$stop_sa\" --role=roles/storage.objectViewer --condition=None --format=none"
A_ROLE="  run g projects add-iam-policy-binding \"\$project\" --member=\"serviceAccount:\$read_sa\" --role=roles/editor --condition=None --format=none"
A_KEY="  run g iam service-accounts keys create \"\$tmp_dir/key.json\" --iam-account=\"\$stop_sa\""
A_SECRET="  run g secrets add-iam-policy-binding \"\$env_name-CF_API_TOKEN\" --member=\"serviceAccount:\$stop_sa\" --role=roles/secretmanager.secretAccessor --format=none"
A_TDEL="  run g iam service-accounts delete \"\$deployer_sa\""
mutant remove-audience full on script 's/^ACTIONS=(start stop read)$/ACTIONS=(start stop)/' "the read audience removed"
i=0
for c in "${CLAUSES[@]}"; do
  i=$((i + 1)); key="${c%% *}"; key="${key#assertion.}"
  mutant "condition-$i-$key" dry on script "/^    \"assertion\.$key \(==\|in\) /d" "attribute-condition clause $key dropped"
done
mutant widen-ref-value dry on script 's/"assertion.ref == \\"staging\\""/"assertion.ref != \\"\\""/' "ref == staging widened to ref != empty"
mutant state-bucket full on script "/^  bind_all\$/a\\$A_STATE" "a ci-stop binding on the state bucket"
mutant state-bucket-unguarded full off script "/^  bind_all\$/a\\$A_STATE" "a ci-stop binding on the state bucket"
mutant forbidden-role full on script "/^  bind_all\$/a\\$A_ROLE" "roles/editor for ci-read"
mutant forbidden-role-unguarded full off script "/^  bind_all\$/a\\$A_ROLE" "roles/editor for ci-read"
mutant keys-create full on script "/^  bind_all\$/a\\$A_KEY" "a service-account keys create verb"
mutant keys-create-unguarded full off script "/^  bind_all\$/a\\$A_KEY" "a service-account keys create verb"
mutant secret-grant full on script "/^  bind_all\$/a\\$A_SECRET" "a Secret Manager grant for ci-stop"
mutant secret-grant-unguarded full off script "/^  bind_all\$/a\\$A_SECRET" "a Secret Manager grant for ci-stop"
mutant read-writes-perms full on perms '$a ci-read\troles/storage.objectCreator\t*' "ci-read objectCreator in the permissions file"
mutant drop-exact-key full on script 's/role=$mutable_role cond="$(key_condition "$keys")"/role=$mutable_role cond=""/' "the exact-key condition dropped from the binding"
mutant iap-projectwide full on script 's/"$SEP" "$SEP" "$(iap_condition)"$/"$SEP" "$SEP" ""/' "the IAP tunnel grant made project-wide (instance condition dropped)"
mutant iap-projectwide-unguarded full off script 's/"$SEP" "$SEP" "$(iap_condition)"$/"$SEP" "$SEP" ""/' "the IAP tunnel grant made project-wide (instance condition dropped)"
mutant iap-other-instance dry on script 's|/instances/%s"'"'"' "$project_number" "$zone" "$instance"|/instances/%s"'"'"' "$project_number" "$zone" "*"|' "the IAP condition widened to a wildcard instance name"
mutant wildcard-exact-key full on perms 's/^ci-stop\tciMetaMutable\t.*/ci-stop\tciMetaMutable\t*/' "the exact-key list replaced by *"
mutant prefix-exact-key full on script 's/resource.name == \\"projects/resource.name.startsWith(\\"projects/' "the exact-key equality widened to a prefix"
mutant teardown-deletes-deployer full on script "/^  receipt api sts.googleapis.com \"retained/i\\$A_TDEL" "teardown also deletes the deployer account"
mutant teardown-deletes-deployer-unguarded full off script "/^  receipt api sts.googleapis.com \"retained/i\\$A_TDEL" "teardown also deletes the deployer account"
mutant live-guard-removed full on script 's/if ((n > 0)) && \[\[ "$confirm_meta" != "$env_name" \]\]; then/if false; then/' "the live-object refusal removed"

# Controls: the owner/failed-read cases against the script with the previous behaviour restored; each must fail.
control() { # NAME DETAIL SED-EXPR FN...
  local t="$S/tree-ctl-$1"
  make_tree "$t"; sed -i -e "$3" "$t/infra/terraform/bootstrap/ci-federation.sh"
  control_run "$t" "$@"
}
control_run() { # TREE NAME DETAIL EXPR FN...: the mutated TREE; every FN must fail against it
  local t=$1 name=$2 detail=$3 f fn killed="" survived=""; shift 4
  f="$t/infra/terraform/bootstrap/ci-federation.sh"
  if cmp -s "$f" "$SRC_SCRIPT"; then case_fail "control.$name" "mutation not applied (sed matched nothing)"; return; fi
  CUR_TR="$TR/controls/$name"; mkdir -p "$CUR_TR"
  for fn in "$@"; do if "$fn" "$t"; then survived+="$fn "; else killed+="$fn "; fi; done
  if [[ -z $survived ]]; then case_pass "control.$name" "$detail; killed: $killed"
  else case_fail "control.$name" "$detail; survived: $survived"; fi
}
control owner-blind "an accessible bucket adopted on describe alone (no owner comparison)" \
  's|^  if gcall storage buckets describe "gs://$meta_bucket" --raw --format=.value(name,projectNumber).; then$|  if gcall storage buckets describe "gs://$meta_bucket"; then META_STATE=present; return 0; fi; if false; then|' \
  c_setup_foreign c_setup_missing c_setup_malformed c_teardown_foreign
control owner-unverified "--verify without the owner comparison" '/^  check "meta-bucket owner"/,+1d' c_verify_foreign
control describe-failure-absent "any failed bucket describe taken as absent" \
  's/^  elif not_found "\$G_ERR"; then META_STATE=absent$/  elif true; then META_STATE=absent/' c_setup_denied c_teardown_describe_failed
control list-failure-empty "a failed live-object listing taken as empty (the former || true)" \
  's/^  die "the live objects of .*/  return 0/' c_teardown_list_denied c_teardown_list_404 c_teardown_list_denied_unconfirmed
control verify-reads-tolerated "verify reads turning every error into empty output (the former rd)" \
  's/^  g "\$@" 2>"\$tmp_dir\/rd.err" || rc=\$?$/  g "$@" 2>\/dev\/null || true/' c_verify_keys c_verify_state_policy c_verify_secrets
control value-reads-tolerated "value_of turning every error into empty output (the former value_of)" \
  's/if probe "\$@"; then VALUE=\$G_OUT; fi/VALUE="$(g "$@" 2>\/dev\/null || true)"/' c_teardown_policy_failed c_setup_pool_unreadable
control late-owner-checks-removed "bucket writes trusting the owner state cached at setup's start (no meta_guard, no post-create check)" \
  '/^  meta_guard "\$@"$/d; /is not readable after its creation/d' c_race_before_update c_race_before_grants c_race_after_create
control owner-recheck-removed "teardown deleting a bucket that was absent at its first owner check" \
  '/^  \[\[ \$META_STATE != present || \$first_meta == present \]\] \\$/,+1d' c_teardown_recheck
# The previous substring classifier (any standalone 404, NOT_FOUND or "not found" anywhere in stderr) swapped in.
OLD_NOT_FOUND='not_found() { [[ ! $1 =~ (^|[^0-9A-Za-z])403([^0-9A-Za-z]|$)|PERMISSION_DENIED|[Pp]ermission|[Ff]orbidden|[Dd]enied|does\ not\ have ]] && [[ $1 =~ (^|[^0-9A-Za-z])404([^0-9A-Za-z]|$)|NOT_FOUND|[Nn]ot\ [Ff]ound ]]; }'
t="$S/tree-ctl-substring-parser"; make_tree "$t"
OLD="$OLD_NOT_FOUND" awk 'index($0, "not_found() {") == 1 {print ENVIRON["OLD"]; next} {print}' \
  "$t/infra/terraform/bootstrap/ci-federation.sh" >"$t/ci-federation.sh.n"
mv "$t/ci-federation.sh.n" "$t/infra/terraform/bootstrap/ci-federation.sh"
use_project demo-404-preview
control_run "$t" substring-parser "the previous substring not-found parser, under the -404- project" - \
  c_p404_teardown_503 c_p404_setup_503 c_p404_teardown_policy_503
use_project example-preview-1

unmodelled="$(cat "$TR"/{base,verify,teardown,ownership,inputs,permissions}/unsupported.log 2>/dev/null | grep -c . || true)"
ct fake.coverage "every gcloud call of the unmutated script is modelled by the fake ($unmodelled unsupported calls)" test "$unmodelled" = 0

# ---- summary ------------------------------------------------------------------------------------------------------------------
result=PASS; ((FAILED == 0)) || result=FAIL
printf 'row\tresult\tpassed\tfailed\n%s\t%s\t%s\t%s\n' CI-FEDERATION "$result" "$PASSED" "$FAILED" >"$REPORT/summary.tsv"
echo "RESULT $result passed=$PASSED failed=$FAILED"
[[ $result == PASS ]]
