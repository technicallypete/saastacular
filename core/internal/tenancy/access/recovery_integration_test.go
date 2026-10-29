//go:build integration

// G013 B-RECOVERY-DRILL and B-MAINTENANCE-DRILL Go phases (Plan r3 §2.2, §10.2), driven by infra/validation/run.sh around
// infra/recovery/recovery.sh:
//   - TestRecoverySeed (RD-S seed): the pinned two-Organization fixture through the real G004 sign-in (fake WorkOS) and
//     the real G009 facade (private MinIO), split by a recorded database timestamp T_mid; prints RECOVERY-SEED; the old
//     authentication material is written only to RECOVERY_OLD_DIR (a run-owned named volume), never to logs.
//   - TestRecoveryPost (RD-P): the six post-checkpoint restriction kinds on the source; prints RECOVERY-POST.
//   - TestRecoveryVerify (RD-G V1-V9): on the restored, invalidated and replayed target, as the actual runtime role through
//     the guarded facade.
//   - TestRecoveryRepair (MD-R): pending and deleting repair through the authorized guarded delete (RECOVERY_REPAIR_STEP).
//
// The environment keeps the identity namespace row (unlike setupIT) and fixes the fake WorkOS issuer, so restored
// principals stay valid across processes (E11; test-only). Identifiers (UUIDs, fixture subjects) pass between phases in
// RECOVERY_IDS; credentials never do.
//
// Required environment: the B-OBJ-IT settings plus RECOVERY_OLD_DIR; Verify and Repair also RECOVERY_IDS, and Verify
// RECOVERY_CHECKPOINT_END (the checkpoint's window end, database clock, from the manifest).
package access

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"core/internal/identity"
	"core/internal/identity/identitytest"
	"core/internal/migrate"
	"core/internal/objectstore"
	"core/internal/observability"
	"core/internal/tenancy"
	"core/migrations"
)

// recIssuer is the drill's fixed fake WorkOS issuer (test-only; production keeps one WorkOS namespace).
const recIssuer = "https://recovery-fixture.workos.test"

// recIDs are the identifiers the phases share (no credential, no secret).
type recIDs struct {
	Login                                                        string // the sign-in fixture Organization (G016 A2), created once by Seed
	OrgA, OrgB, A1, A2, B1, P, Q                                 string
	Users                                                        map[string]string // fixture key -> user UUID
	Subjects                                                     map[string]string // fixture key -> WorkOS subject
	ObjA1, ObjA2, ObjA, ObjB1, TombA1, TombB1, Deleting, Pending string
	DigestA1, DigestA2, DigestA, DigestB1                        string
	TMid                                                         string
	RemovedBinding, DowngradedBinding                            map[string]string
	PostDone                                                     bool
}

// setupRecovery is setupObjects without the namespace-row reset and with the fixed issuer. login is the sign-in fixture
// Organization the phases share (G016 A2: every itOrigin sign-in is bound to it and admits into it). Seed (ids nil)
// creates it before any identity use and records it in RECOVERY_IDS; every later phase reuses that restored
// Organization from ids and never creates tenant state of its own for the fixture (a missing value fails closed).
func setupRecovery(t *testing.T, ids *recIDs) *objEnv {
	t.Helper()
	ctx := context.Background()
	cfg, err := migrate.LoadConfig(os.Getenv)
	if err != nil {
		t.Fatalf("integration settings: %v", err)
	}
	e := &itEnv{cfg: cfg, pw: os.Getenv("DB_RUNTIME_PASSWORD")}
	canaryFile, logFile := os.Getenv("TENANCY_CANARY_FILE"), os.Getenv("TENANCY_LOG_FILE")
	if e.pw == "" || canaryFile == "" || logFile == "" {
		t.Fatal("integration settings: DB_RUNTIME_PASSWORD, TENANCY_CANARY_FILE and TENANCY_LOG_FILE are required")
	}
	b, err := os.ReadFile(canaryFile)
	if err != nil {
		t.Fatalf("canary file: %v", err)
	}
	if e.canaries = strings.Fields(string(b)); len(e.canaries) < 3 {
		t.Fatal("canary file: at least three canaries are required")
	}
	r, err := migrate.Open(ctx, cfg, migrations.FS)
	if err != nil {
		t.Fatalf("open migrator: %v", err)
	}
	var out bytes.Buffer
	if err := r.Up(ctx, &out); err != nil || !strings.HasSuffix(out.String(), "current_version\t22\n") {
		t.Fatalf("migrate up: %v %q", err, out.String())
	}
	_ = r.Close()
	e.setup, e.setupDB = itSetup(t, cfg)
	e.pool = e.runtimePool(t, nil)
	e.fake = identitytest.New(t, itClient)
	e.fake.Issuer = recIssuer
	e.clk = newITClock()
	e.fake.SetNow(e.clk.Now)
	e.model = tenancy.NewService(e.setupDB, nil)
	if ids == nil {
		e.loginOrg = e.org(t, "Recovery sign-in fixture")
	} else if ids.Login == "" {
		t.Fatal("RECOVERY_IDS: no sign-in fixture Organization")
	} else if e.loginOrg = mustOrg(t, ids.Login); zero(e.loginOrg) {
		t.Fatal("RECOVERY_IDS: no sign-in fixture Organization")
	}
	e.id = e.identity(t, e.clk, e.pool)
	e.logs = &itSyncBuf{}
	e.hook = &hookDB{pool: e.pool}
	e.svc = NewService(e.hook, e.id, observability.NewLogger(e.logs, slog.LevelInfo, nil))
	t.Cleanup(func() {
		f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err == nil {
			_, err = f.WriteString(e.logs.String())
			_ = f.Close()
		}
		if err != nil {
			t.Errorf("write the access log: %v", err)
		}
	})
	o := &objEnv{itEnv: e}
	o.scfg, o.ocfg = objectSettings(t)
	if o.store, err = objectstore.New(o.scfg, os.Getenv("APP_ENV")); err != nil {
		t.Fatalf("object store: %v", err)
	}
	o.spy = &objSpy{inner: o.store}
	o.osvc = NewServiceWithObjects(o.hook, o.id, o.svc.logger, o.spy, o.ocfg)
	return o
}

// recSignIn completes one real G004 sign-in of the fixed subject and returns the credential.
func (e *objEnv) recSignIn(t *testing.T, sub, name string) string {
	t.Helper()
	return e.signInWith(t, e.id, sub, name, nil, "", "")
}

func (e *objEnv) recUser(t *testing.T, cred string) *itUser {
	t.Helper()
	info, err := e.id.Session(context.Background(), cred, itOrigin)
	if err != nil {
		t.Fatalf("session: %v", err)
	}
	u := &itUser{cred: cred, session: info.SessionID}
	if u.id, err = tenancy.ParseUserID(info.UserID); err != nil {
		t.Fatal(err)
	}
	return u
}

// recBegin starts one request scope of svc for cred at org's own origin (G016 A2: cred is the credential issued for
// org; any other credential is refused, never substituted); the scope is ended at cleanup.
func recBegin(t *testing.T, svc *Service, cred string, org tenancy.OrganizationID) (context.Context, func()) {
	t.Helper()
	rctx, end, err := svc.BeginRequest(context.Background(), cred, itOriginOf(org))
	if err != nil {
		t.Fatalf("BeginRequest: %v", err)
	}
	t.Cleanup(end)
	return rctx, end
}

// recOrgSignIn completes one real G004 sign-in of subject sub at org's own origin (G016 A2) and returns the credential.
func (e *objEnv) recOrgSignIn(t *testing.T, org tenancy.OrganizationID, sub, name string) string {
	t.Helper()
	return e.signInAt(t, e.id, org, sub, name, nil, "", "")
}

// recPut uploads n bytes of a seeded payload through the facade as cred (its credential for org) into org (and Group
// g) and returns its id and digest.
func (e *objEnv) recPut(t *testing.T, cred string, org tenancy.OrganizationID, g tenancy.GroupID, seed string) (string, string) {
	t.Helper()
	rctx, end := recBegin(t, e.osvc, cred, org)
	defer end()
	tc := resolve(t, e.osvc, rctx, org, tenancy.GroupID{})
	data := payload(4096, seed)
	info, err := e.osvc.PutObject(rctx, tc, PutObjectRequest{Group: g, Size: int64(len(data)), Body: io.NopCloser(bytes.NewReader(data))})
	if err != nil {
		t.Fatalf("put %s: %v", seed, err)
	}
	sum := sha256.Sum256(data)
	return info.ID.String(), hex.EncodeToString(sum[:])
}

func (e *objEnv) recDelete(t *testing.T, cred string, org tenancy.OrganizationID, id string) DeleteResult {
	t.Helper()
	rctx, end := recBegin(t, e.osvc, cred, org) // G016 A2: cred is the credential for org
	defer end()
	tc := resolve(t, e.osvc, rctx, org, tenancy.GroupID{})
	oid, err := tenancy.ParseObjectID(id)
	if err != nil {
		t.Fatal(err)
	}
	res, err := e.osvc.DeleteObject(rctx, tc, oid)
	if err != nil {
		t.Fatalf("delete: %v", err)
	}
	return res
}

func recOldDir(t *testing.T) string {
	t.Helper()
	d := os.Getenv("RECOVERY_OLD_DIR")
	if d == "" {
		t.Fatal("integration settings: RECOVERY_OLD_DIR is required")
	}
	return d
}

func mustOrg(t *testing.T, s string) tenancy.OrganizationID {
	t.Helper()
	o, err := tenancy.ParseOrganizationID(s)
	if err != nil {
		t.Fatal(err)
	}
	return o
}

func mustGroup(t *testing.T, s string) tenancy.GroupID {
	t.Helper()
	g, err := tenancy.ParseGroupID(s)
	if err != nil {
		t.Fatal(err)
	}
	return g
}

func recLoadIDs(t *testing.T) recIDs {
	t.Helper()
	var ids recIDs
	if err := json.Unmarshal([]byte(os.Getenv("RECOVERY_IDS")), &ids); err != nil || ids.OrgA == "" {
		t.Fatalf("RECOVERY_IDS: %v", err)
	}
	return ids
}

// TestRecoverySeed (RD-S seed): the §2.2 fixture.
func TestRecoverySeed(t *testing.T) {
	e := setupRecovery(t, nil)
	ctx := context.Background()
	ids := recIDs{Login: e.loginOrg.String(), Users: map[string]string{}, Subjects: map[string]string{}}
	creds := map[string]string{}
	users := map[string]*itUser{}
	for _, k := range []string{"ownerA", "memberA1", "gadminA2", "memberA2", "ownerB", "cross", "disabled0", "disable1"} {
		sub := "user_01RECOVERY" + strings.ToUpper(k)
		creds[k] = e.recSignIn(t, sub, "Recovery "+k)
		users[k] = e.recUser(t, creds[k])
		ids.Users[k], ids.Subjects[k] = users[k].id.String(), sub
	}
	orgA, orgB := e.org(t, "Recovery A "+e.canaries[0]), e.org(t, "Recovery B")
	a1, a2 := e.group(t, orgA, users["ownerA"], "Recovery A1", "rec-a1"), e.group(t, orgA, users["ownerA"], "Recovery A2", "rec-a2")
	b1 := e.group(t, orgB, users["ownerB"], "Recovery B1", "rec-b1")
	ids.OrgA, ids.OrgB, ids.A1, ids.A2, ids.B1 = orgA.String(), orgB.String(), a1.ID.String(), a2.ID.String(), b1.ID.String()
	p, q := e.profile(t, orgA, "Recovery P"), e.profile(t, orgB, "Recovery Q")
	e.link(t, orgA, a1.ID, p.ID)
	e.link(t, orgA, a2.ID, p.ID)
	e.link(t, orgB, b1.ID, q.ID)
	ids.P, ids.Q = p.ID.String(), q.ID.String()
	e.grantOrg(t, orgA, users["ownerA"], tenancy.OrgOwner)
	e.grantOrg(t, orgB, users["ownerB"], tenancy.OrgOwner)
	for _, g := range []struct {
		u   string
		org tenancy.OrganizationID
		grp tenancy.GroupID
		r   tenancy.GroupRole
	}{{"memberA1", orgA, a1.ID, tenancy.GroupMember}, {"gadminA2", orgA, a2.ID, tenancy.GroupAdmin}, {"memberA2", orgA, a2.ID, tenancy.GroupMember},
		{"cross", orgA, a1.ID, tenancy.GroupMember}, {"cross", orgB, b1.ID, tenancy.GroupMember}} {
		e.grantOrg(t, g.org, users[g.u], tenancy.OrgMember)
		e.grantGroup(t, g.org, g.grp, users[g.u], g.r)
	}
	// G016 A2: each acting subject's credential for the Organization it acts in, issued at that Organization's origin
	// after its grants (a member there, so nothing is admitted); the same display name, so nothing is renamed.
	at := map[string]string{}
	for _, a := range []struct {
		k   string
		org tenancy.OrganizationID
	}{{"memberA1", orgA}, {"memberA2", orgA}, {"ownerA", orgA}, {"ownerB", orgB}, {"cross", orgB}} {
		at[a.k] = e.recOrgSignIn(t, a.org, ids.Subjects[a.k], "Recovery "+a.k)
	}
	// Objects through the facade (known bytes and digests), tombstones by real deletes.
	ids.ObjA1, ids.DigestA1 = e.recPut(t, at["memberA1"], orgA, a1.ID, "rec-a1")
	ids.ObjA2, ids.DigestA2 = e.recPut(t, at["memberA2"], orgA, a2.ID, "rec-a2")
	ids.ObjA, ids.DigestA = e.recPut(t, at["ownerA"], orgA, tenancy.GroupID{}, "rec-a")
	ids.ObjB1, ids.DigestB1 = e.recPut(t, at["ownerB"], orgB, b1.ID, "rec-b1")
	ids.TombA1, _ = e.recPut(t, at["memberA1"], orgA, a1.ID, "rec-tomb-a1")
	ids.TombB1, _ = e.recPut(t, at["ownerB"], orgB, b1.ID, "rec-tomb-b1")
	if r := e.recDelete(t, at["memberA1"], orgA, ids.TombA1); r.Outcome != DeleteOutcomeDeleted {
		t.Fatalf("tombstone A1: %v", r.Outcome)
	}
	if r := e.recDelete(t, at["ownerB"], orgB, ids.TombB1); r.Outcome != DeleteOutcomeDeleted {
		t.Fatalf("tombstone B1: %v", r.Outcome)
	}
	// A deleting row with bytes present (the provider-delete failure seam) and an inert pending row without bytes.
	del, pend := tenancy.ObjectID(newUUID16(t)), tenancy.ObjectID(newUUID16(t))
	e.seed(t, objFixture{id: del, org: orgA, group: a1.ID, creator: users["memberA1"].id, state: tenancy.ObjectDeleting, ct: "text/plain", data: payload(128, "rec-deleting")})
	e.seed(t, objFixture{id: pend, org: orgA, group: a1.ID, creator: users["memberA1"].id, state: tenancy.ObjectPending, ct: "text/plain"})
	ids.Deleting, ids.Pending = del.String(), pend.String()
	// Authentication before the checkpoint: disabled0 disabled; one session revoked by logout.
	if err := e.id.DisableUser(ctx, users["disabled0"].id.String()); err != nil {
		t.Fatal(err)
	}
	revoked := e.recSignIn(t, ids.Subjects["cross"], "Recovery cross again")
	e.logout(t, revoked, identity.ScopeCurrent)
	// T_mid: older history before; sign-ins, denials and URL issuance after.
	time.Sleep(20 * time.Millisecond)
	var tmid time.Time
	if err := e.setup.QueryRow(ctx, `SELECT clock_timestamp()`).Scan(&tmid); err != nil {
		t.Fatal(err)
	}
	ids.TMid = tmid.UTC().Format(retTimeLayout)
	time.Sleep(20 * time.Millisecond)
	old := map[string]string{"revoked": revoked}
	old["memberA2"] = e.recSignIn(t, ids.Subjects["memberA2"], "Recovery memberA2 live")
	// A provider logout reference left unreturned.
	lo := e.recSignIn(t, ids.Subjects["ownerB"], "Recovery ownerB provider")
	out, err := e.id.Logout(ctx, identity.LogoutInput{Scope: identity.ScopeProvider, CSRF: e.csrf(t, lo), Origin: itOrigin, Credential: lo})
	if err != nil || out.LogoutRef == "" {
		t.Fatalf("provider logout: %v", err)
	}
	old["logoutRef"] = out.LogoutRef
	// A pending login transaction (started, never completed).
	lin, err := e.id.Login(ctx, identity.LoginInput{Origin: itOrigin, ReturnTo: "/account"})
	if err != nil {
		t.Fatal(err)
	}
	u, _ := url.Parse(lin.AuthorizationURL)
	old["loginState"], old["loginBinding"] = u.Query().Get("state"), lin.Binding
	// G062 SUP-11 (Plan r5 §2 recovery fold): a live root signup credential with a live continuation generation, seeded
	// before the checkpoint (setup side); the restore invalidation must expire both. Only its digest is handed over.
	sig := sha256.Sum256([]byte("recovery-signup/" + ids.Users["memberA2"]))
	sigCSRF, sigRef := sha256.Sum256([]byte("recovery-signup-csrf")), sha256.Sum256([]byte("recovery-signup-ref"))
	sigIntent := sha256.Sum256([]byte("recovery-signup-intent"))
	if _, err := e.setup.Exec(ctx, `INSERT INTO app.signup_sessions (id, user_id, origin, credential_hash, csrf_token, provider_session_id,
  expires_at, intent_hash, organization_id, provisioned_at, target_origin, continuation_generation, continuation_expires_at, ref_hash)
VALUES (gen_random_uuid(), $1::uuid, 'https://localhost:543', $2, $3, 'session_recovery_signup', now() + interval '29 minutes', $4,
  $5::uuid, now(), $6, 1, now() + interval '29 minutes', $7)`, ids.Users["memberA2"], sig[:], sigCSRF[:], sigIntent[:], ids.OrgA, itOrigin,
		sigRef[:]); err != nil {
		t.Fatalf("signup fixture: %v", err)
	}
	old["signupDigest"] = hex.EncodeToString(sig[:])
	// A denial and a signed URL issuance after T_mid.
	rctx, end := recBegin(t, e.osvc, at["ownerB"], orgB) // G016 A2: ownerB's own (B) scope; the A refusal is kept
	if _, err := e.osvc.Resolve(rctx, ResolveRequest{OrganizationID: orgA.String()}); err == nil {
		t.Fatal("ownerB resolved Organization A")
	}
	end()
	rctx, end = recBegin(t, e.osvc, at["cross"], orgB)
	tc := resolve(t, e.osvc, rctx, orgB, tenancy.GroupID{})
	if _, err := e.osvc.SignedGetURL(rctx, tc, SignedGetURLRequest{Object: mustObj(t, ids.ObjB1), DefaultTTL: true}); err != nil {
		t.Fatalf("sign: %v", err)
	}
	end()
	b, _ := json.Marshal(old)
	if err := os.WriteFile(filepath.Join(recOldDir(t), "old-auth.json"), b, 0o600); err != nil {
		t.Fatalf("old authentication handoff: %v", err)
	}
	j, _ := json.Marshal(ids)
	fmt.Println("RECOVERY-SEED " + string(j))
}

func mustObj(t *testing.T, s string) tenancy.ObjectID {
	t.Helper()
	o, err := tenancy.ParseObjectID(s)
	if err != nil {
		t.Fatal(err)
	}
	return o
}

func newUUID16(t *testing.T) [16]byte {
	t.Helper()
	id, err := newObjectID()
	if err != nil {
		t.Fatal(err)
	}
	return id
}

// TestRecoveryPost (RD-P): the six post-checkpoint restriction kinds, each through its real path.
func TestRecoveryPost(t *testing.T) {
	ids := recLoadIDs(t)
	e := setupRecovery(t, &ids)
	ctx := context.Background()
	orgA, a1, a2 := mustOrg(t, ids.OrgA), mustGroup(t, ids.A1), mustGroup(t, ids.A2)
	uid := func(k string) tenancy.UserID {
		u, err := tenancy.ParseUserID(ids.Users[k])
		if err != nil {
			t.Fatal(err)
		}
		return u
	}
	if err := e.model.RemoveGroupMembership(ctx, orgA, a1, uid("memberA1")); err != nil {
		t.Fatalf("remove: %v", err)
	}
	if _, err := e.model.ChangeGroupRole(ctx, orgA, a2, uid("gadminA2"), tenancy.GroupMember); err != nil {
		t.Fatalf("downgrade: %v", err)
	}
	p, err := tenancy.ParseBusinessProfileID(ids.P)
	if err != nil {
		t.Fatal(err)
	}
	if err := e.model.UnlinkGroupBusinessProfile(ctx, orgA, a2, p); err != nil {
		t.Fatalf("unlink: %v", err)
	}
	if err := e.id.DisableUser(ctx, ids.Users["disable1"]); err != nil {
		t.Fatalf("disable: %v", err)
	}
	owner := e.recOrgSignIn(t, orgA, ids.Subjects["ownerA"], "Recovery ownerA post") // G016 A2: issued for A
	rctx, end := recBegin(t, e.osvc, owner, orgA)
	tc := resolve(t, e.osvc, rctx, orgA, tenancy.GroupID{})
	if _, err := e.osvc.SetOrganizationPolicy(rctx, tc, tenancy.OrganizationPolicy{GroupAdminsEditBusinessProfiles: false}); err != nil {
		t.Fatalf("policy: %v", err)
	}
	end()
	if r := e.recDelete(t, owner, orgA, ids.ObjA); r.Outcome != DeleteOutcomeDeleted {
		t.Fatalf("delete A object: %v", r.Outcome)
	}
	ids.RemovedBinding = map[string]string{"organization": ids.OrgA, "subject": ids.Users["memberA1"], "scope_type": "group", "scope_id": ids.A1, "role": "group_member"}
	ids.DowngradedBinding = map[string]string{"organization": ids.OrgA, "subject": ids.Users["gadminA2"], "scope_type": "group", "scope_id": ids.A2,
		"from_role": "group_admin", "to_role": "group_member"}
	ids.PostDone = true
	j, _ := json.Marshal(ids)
	fmt.Println("RECOVERY-POST " + string(j))
}

// recDenied asserts err is an access denial (any code).
func recDenied(t *testing.T, what string, err error) {
	t.Helper()
	var ae *Error
	if err == nil || !errors.As(err, &ae) {
		t.Errorf("%s: %v, want an access denial", what, err)
	}
}

// TestRecoveryVerify (RD-G V1-V9) on the restored target.
func TestRecoveryVerify(t *testing.T) {
	ids := recLoadIDs(t)
	cpEnd, err := time.Parse(retTimeLayout, os.Getenv("RECOVERY_CHECKPOINT_END"))
	if err != nil {
		t.Fatalf("RECOVERY_CHECKPOINT_END: %v", err)
	}
	e := setupRecovery(t, &ids)
	ctx := context.Background()
	var verifyStart time.Time
	if err := e.setup.QueryRow(ctx, `SELECT clock_timestamp()`).Scan(&verifyStart); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(recOldDir(t), "old-auth.json"))
	if err != nil {
		t.Fatalf("old authentication handoff: %v", err)
	}
	var old map[string]string
	if err := json.Unmarshal(b, &old); err != nil {
		t.Fatal(err)
	}
	orgA, orgB, a1, a2, b1 := mustOrg(t, ids.OrgA), mustOrg(t, ids.OrgB), mustGroup(t, ids.A1), mustGroup(t, ids.A2), mustGroup(t, ids.B1)
	one := func(sql string, args ...any) string {
		t.Helper()
		var s string
		if err := e.setup.QueryRow(ctx, sql, args...).Scan(&s); err != nil {
			t.Fatalf("%s: %v", sql, err)
		}
		return s
	}
	// G062 SUP-11: the restored signup row is read here, before any subtest or identity call: every login start sweeps
	// the rows expired by the database clock (signup_sessions_expire), so the invalidated row may be gone by V6.
	sig, err := hex.DecodeString(old["signupDigest"])
	if err != nil || len(sig) != 32 {
		t.Fatalf("signup digest handoff: %v", err)
	}
	const signupState = `SELECT coalesce((SELECT format('%s|%s|%s', revoked_at IS NULL, expires_at <= $2, continuation_expires_at <= $2)
FROM app.signup_sessions WHERE credential_hash = $1), 'swept')`
	if got := one(signupState, sig, verifyStart); got != "t|t|t" {
		t.Errorf("restored signup row (unrevoked|expired|generation expired): %q, want t|t|t", got)
	}

	t.Run("V1 restored relationships", func(t *testing.T) {
		got := one(`SELECT string_agg(format('%s:%s:%s', u.k, b.scope_type, b.role), ',' ORDER BY u.k, b.scope_type, b.role)
FROM app.role_bindings b JOIN (SELECT key AS k, value::uuid AS id FROM json_each_text($1::json)) u ON u.id = b.subject_user_id
WHERE b.organization_id IN ($2::uuid, $3::uuid)`, mustJSON(ids.Users), ids.OrgA, ids.OrgB)
		want := "cross:group:group_member,cross:group:group_member,cross:organization:org_member,cross:organization:org_member," +
			"gadminA2:group:group_member,gadminA2:organization:org_member,memberA1:organization:org_member,memberA2:group:group_member," +
			"memberA2:organization:org_member,ownerA:group:group_owner,ownerA:group:group_owner,ownerA:organization:org_owner," +
			"ownerB:group:group_owner,ownerB:organization:org_owner"
		if got != want {
			t.Errorf("bindings %q, want %q", got, want)
		}
		if got := one(`SELECT string_agg(group_id::text, ',' ORDER BY group_id) FROM app.group_business_profiles WHERE business_profile_id = $1::uuid`, ids.P); got != ids.A1 {
			t.Errorf("P links %q, want A1 only", got)
		}
		if got := one(`SELECT string_agg(status, ',' ORDER BY id) FROM app.users WHERE id = ANY($1::uuid[])`,
			[]string{ids.Users["disabled0"], ids.Users["disable1"]}); got != "disabled,disabled" {
			t.Errorf("disabled users %q", got)
		}
		if got := one(`SELECT group_admins_edit_business_profiles::text FROM app.organizations WHERE id = $1::uuid`, ids.OrgA); got != "false" {
			t.Errorf("A policy flag %s", got)
		}
	})

	// G016 A2: a fresh credential is issued for one Organization (at its origin); get runs in a request scope of cred at
	// org's origin, so cred must be the credential for org.
	fresh := func(k string, org tenancy.OrganizationID) string {
		return e.recOrgSignIn(t, org, ids.Subjects[k], "Recovery "+k+" restored")
	}
	get := func(t *testing.T, cred string, org tenancy.OrganizationID, g tenancy.GroupID, obj string) ([]byte, error) {
		t.Helper()
		rctx, end := recBegin(t, e.osvc, cred, org)
		defer end()
		tc, err := e.osvc.Resolve(rctx, ResolveRequest{OrganizationID: org.String(), GroupID: groupString(g)})
		if err != nil {
			return nil, err
		}
		_, rc, err := e.osvc.GetObject(rctx, tc, mustObj(t, obj))
		if err != nil {
			return nil, err
		}
		defer rc.Close()
		return io.ReadAll(rc)
	}
	digest := func(b []byte) string { s := sha256.Sum256(b); return hex.EncodeToString(s[:]) }
	ownerB, cross, crossB, ownerA := fresh("ownerB", orgB), fresh("cross", orgA), fresh("cross", orgB), fresh("ownerA", orgA)

	t.Run("V2 runtime positives incl. signed fetch and binding", func(t *testing.T) {
		if b, err := get(t, ownerB, orgB, tenancy.GroupID{}, ids.ObjB1); err != nil || digest(b) != ids.DigestB1 {
			t.Errorf("ownerB B1 object: %v", err)
		}
		if b, err := get(t, cross, orgA, a1, ids.ObjA1); err != nil || digest(b) != ids.DigestA1 {
			t.Errorf("cross A1 object: %v", err)
		}
		if b, err := get(t, ownerA, orgA, tenancy.GroupID{}, ids.ObjA2); err != nil || digest(b) != ids.DigestA2 {
			t.Errorf("ownerA A2 object: %v", err)
		}
		rctx, end := recBegin(t, e.osvc, ownerB, orgB)
		defer end()
		tc := resolve(t, e.osvc, rctx, orgB, tenancy.GroupID{})
		su, err := e.osvc.SignedGetURL(rctx, tc, SignedGetURLRequest{Object: mustObj(t, ids.ObjB1), DefaultTTL: true})
		if err != nil {
			t.Fatalf("sign: %v", err)
		}
		resp, err := http.Get(su.url)
		if err != nil {
			t.Fatalf("signed fetch: %v", objRedact(err.Error()))
		}
		body, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if resp.StatusCode != 200 || digest(body) != ids.DigestB1 || resp.Header.Get("Content-Disposition") != "attachment" {
			t.Errorf("signed fetch %d", resp.StatusCode)
		}
	})

	t.Run("V3 cross-Organization and cross-Group negatives", func(t *testing.T) {
		rctx, end := recBegin(t, e.osvc, ownerB, orgB) // ownerB's own (B) scope; the A refusal is kept
		_, err := e.osvc.Resolve(rctx, ResolveRequest{OrganizationID: ids.OrgA})
		end()
		recDenied(t, "ownerB resolves A", err)
		memberA2 := fresh("memberA2", orgA)
		_, err = get(t, memberA2, orgA, a2, ids.ObjA1)
		recDenied(t, "memberA2 reads an A1 object (shared profile P is not transitive)", err)
		_, err = get(t, memberA2, orgA, a1, ids.ObjA1)
		recDenied(t, "memberA2 selects A1", err)
		_, err = get(t, crossB, orgB, b1, ids.ObjA1)
		recDenied(t, "an A1 object through B", err)
	})

	t.Run("V4 audit history and refusals", func(t *testing.T) {
		rctx, end := recBegin(t, e.osvc, ownerA, orgA)
		tc := resolve(t, e.osvc, rctx, orgA, tenancy.GroupID{})
		page, err := e.osvc.ReadAuditEvents(rctx, tc, AuditReadRequest{Since: cpEnd.Add(-24 * time.Hour), Until: verifyStart.Add(time.Minute), Limit: 100})
		end()
		if err != nil || len(page.Events) == 0 {
			t.Fatalf("owner audit read: %d events (%v)", len(page.Events), err)
		}
		rt := e.runtimePool(t, nil)
		for _, s := range []string{`UPDATE audit.events SET outcome = 'failed'`, `DELETE FROM audit.events`, `TRUNCATE audit.events`,
			`INSERT INTO audit.events (action, outcome, organization_id, resource_type, resource_id) VALUES ('organization.created', 'success', NULL, 'organization', gen_random_uuid())`} {
			if _, err := rt.Exec(ctx, s); retState(err) != "42501" {
				t.Errorf("runtime %q: %v", s, err)
			}
		}
	})

	t.Run("V5 privilege refusals", func(t *testing.T) {
		rt := e.runtimePool(t, nil)
		for _, s := range []string{`SET ROLE ` + e.cfg.OwnerRole(), `CREATE TABLE app.recovery_probe (id int)`, `ALTER TABLE app.users ADD COLUMN probe int`} {
			if _, err := rt.Exec(ctx, s); retState(err) != "42501" {
				t.Errorf("runtime %q: %v", s, err)
			}
		}
	})

	t.Run("V6 old authentication refused", func(t *testing.T) {
		if _, err := e.id.Session(ctx, old["memberA2"], itOrigin); !errors.Is(err, identity.ErrUnauthenticated) {
			t.Errorf("old session: %v", err)
		}
		if _, err := e.id.Session(ctx, old["revoked"], itOrigin); !errors.Is(err, identity.ErrUnauthenticated) {
			t.Errorf("revoked session: %v", err)
		}
		if got := one(`SELECT count(*)::text FROM app.sessions WHERE revoked_at IS NOT NULL AND user_id = $1::uuid`, ids.Users["cross"]); got == "0" {
			t.Error("the revoked session lost its revocation")
		}
		cb, err := e.id.Callback(ctx, identity.CallbackInput{Origin: itOrigin, RawQuery: "code=code_stale&state=" + url.QueryEscape(old["loginState"]),
			Binding: old["loginBinding"]})
		if err == nil && cb.Credential != "" {
			t.Error("the pending login callback completed after the restore")
		}
		if got, _ := e.id.ProviderLogoutReturn(ctx, old["logoutRef"]); got != identity.OutcomeLocalRevoked {
			t.Errorf("logout reference: %q", got)
		}
		// G062 SUP-11: the restored signup credential and its continuation generation were expired by the invalidation
		// (not revoked: no forged revocation), so the runtime bootstrap reader refuses the credential. The binding row
		// state is read before the subtests; after the sign-ins above the row is unchanged or already swept.
		var resolved int
		if err := e.pool.QueryRow(ctx, `SELECT count(*)::int FROM app.auth_resolve_signup($1, now())`, sig).Scan(&resolved); err != nil || resolved != 0 {
			t.Errorf("restored signup credential resolves: %d (%v)", resolved, err)
		}
		if got := one(signupState, sig, verifyStart); got != "t|t|t" && got != "swept" {
			t.Errorf("restored signup row after sign-ins: %q, want t|t|t or swept", got)
		}
		if _, err := e.audCallback(t, e.id, ids.Subjects["disabled0"], fmt.Sprintf("session_01RC%08d", itSeq.Add(1)), "code", ""); err == nil {
			t.Error("disabled0 signed in")
		}
	})

	t.Run("V7 a fresh synthetic login works", func(t *testing.T) {
		u := e.recUser(t, e.recSignIn(t, fmt.Sprintf("user_01RECOVERYNEW%06d", itSeq.Add(1)), "Recovery newcomer"))
		if u.id.String() == "" {
			t.Error("no user")
		}
	})

	t.Run("V8 replayed restrictions deny their old authority", func(t *testing.T) {
		memberA1 := fresh("memberA1", orgA)
		_, err := get(t, memberA1, orgA, a1, ids.ObjA1)
		recDenied(t, "removed memberA1 selects A1", err)
		gadmin := fresh("gadminA2", orgA)
		rctx, end := recBegin(t, e.osvc, gadmin, orgA)
		tc, err := e.osvc.Resolve(rctx, ResolveRequest{OrganizationID: ids.OrgA, GroupID: ids.A2})
		if err == nil {
			_, err = e.osvc.RenameGroup(rctx, tc, a2, "Recovery A2 renamed")
		}
		end()
		recDenied(t, "downgraded gadminA2 renames A2", err)
		memberA2 := fresh("memberA2", orgA)
		rctx, end = recBegin(t, e.osvc, memberA2, orgA)
		tc, err = e.osvc.Resolve(rctx, ResolveRequest{OrganizationID: ids.OrgA, GroupID: ids.A2})
		if err == nil {
			p, _ := tenancy.ParseBusinessProfileID(ids.P)
			_, err = e.osvc.GetBusinessProfile(rctx, tc, p)
		}
		end()
		recDenied(t, "P through the unlinked A2", err)
		if _, err := e.audCallback(t, e.id, ids.Subjects["disable1"], fmt.Sprintf("session_01RC%08d", itSeq.Add(1)), "code", ""); err == nil {
			t.Error("disable1 signed in")
		}
		rctx, end = recBegin(t, e.osvc, ownerA, orgA)
		tc = resolve(t, e.osvc, rctx, orgA, tenancy.GroupID{})
		pol, err := e.osvc.GetOrganizationPolicy(rctx, tc)
		if err != nil || pol.GroupAdminsEditBusinessProfiles {
			t.Errorf("A policy: %+v %v", pol, err)
		}
		_, _, err = e.osvc.GetObject(rctx, tc, mustObj(t, ids.ObjA))
		recDenied(t, "the deleted object is readable", err)
		_, err = e.osvc.SignedGetURL(rctx, tc, SignedGetURLRequest{Object: mustObj(t, ids.ObjA), DefaultTTL: true})
		recDenied(t, "the deleted object is signable", err)
		end()
		k, _ := objectstore.NewKey(ids.OrgA, ids.ObjA)
		if e.present(t, k) {
			t.Error("the deleted object's bytes are present")
		}
		if b, err := get(t, crossB, orgB, b1, ids.ObjB1); err != nil || digest(b) != ids.DigestB1 {
			t.Errorf("cross's B1 access changed: %v", err)
		}
		if one(`SELECT count(*)::text FROM app.business_profiles WHERE id = $1::uuid`, ids.P) != "1" ||
			one(`SELECT count(*)::text FROM app.users WHERE id = $1::uuid`, ids.Users["memberA1"]) != "1" {
			t.Error("P or the removed user's global row is gone")
		}
	})

	t.Run("V9 replay events and attribution", func(t *testing.T) {
		got := one(`SELECT coalesce(string_agg(format('%s:%s:%s', action, actor_kind, CASE WHEN actor_user_id = $3::uuid THEN 'ownerA' ELSE coalesce(actor_user_id::text, '-') END),
  ',' ORDER BY action COLLATE "C"), '') FROM audit.events WHERE occurred_at > $1 AND occurred_at < $2`, cpEnd, verifyStart, ids.Users["ownerA"])
		want := "business_profile.unlinked:unknown:-,membership.removed:unknown:-,membership.role_changed:unknown:-,object.delete_started:user:ownerA," +
			"object.deleted:user:ownerA,organization.policy_changed:unknown:-,revocation.unaudited:unknown:-"
		if got != want {
			t.Errorf("events between the checkpoint and verification %q, want %q", got, want)
		}
		if one(`SELECT count(*)::text FROM audit.events WHERE action = 'user.disabled' AND subject_user_id = $1::uuid AND occurred_at > $2`, ids.Users["disable1"], cpEnd) != "0" {
			t.Error("a user.disabled event is claimed for disable1")
		}
		if one(`SELECT count(*)::text FROM audit.events WHERE action = 'revocation.unaudited' AND subject_user_id = $1::uuid AND resource_id = $1::uuid
  AND reason = 'disabled' AND actor_kind = 'unknown' AND occurred_at > $2`, ids.Users["disable1"], cpEnd) != "1" {
			t.Error("the revocation.unaudited record for disable1 is missing or misshaped")
		}
	})
}

func groupString(g tenancy.GroupID) string {
	if zero(g) {
		return ""
	}
	return g.String()
}

func mustJSON(v any) string { b, _ := json.Marshal(v); return string(b) }

// TestRecoveryRepair (MD-R): step 1 repairs the seeded deleting row (a provider fault keeps it deletion_pending, a retry
// completes it), deletes the inert pending row, and deletes a ready object without a ledger entry (tombstone T, printed);
// step 2 is the authorized repeat delete of the restored seed tombstone after bytes were planted under it.
func TestRecoveryRepair(t *testing.T) {
	ids := recLoadIDs(t)
	e := setupRecovery(t, &ids)
	orgA := mustOrg(t, ids.OrgA)
	owner := e.recOrgSignIn(t, orgA, ids.Subjects["ownerA"], "Recovery ownerA repair") // G016 A2: issued for A
	switch os.Getenv("RECOVERY_REPAIR_STEP") {
	case "1":
		e.spy.set(func(op string, _ objectstore.Key) error {
			if op == "delete" {
				return objectstore.ErrUnavailable
			}
			return nil
		})
		if r := e.recDelete(t, owner, orgA, ids.Deleting); r.Outcome != DeleteOutcomeDeletionPending {
			t.Errorf("deleting row under a provider fault: %v", r.Outcome)
		}
		e.spy.set(nil)
		if r := e.recDelete(t, owner, orgA, ids.Deleting); r.Outcome != DeleteOutcomeDeleted {
			t.Errorf("deleting row retry: %v", r.Outcome)
		}
		if r := e.recDelete(t, owner, orgA, ids.Pending); r.Outcome != DeleteOutcomeDeleted {
			t.Errorf("inert pending row: %v", r.Outcome)
		}
		if r := e.recDelete(t, owner, orgA, ids.ObjA2); r.Outcome != DeleteOutcomeDeleted {
			t.Errorf("ready object: %v", r.Outcome)
		}
		// The object cutoff for MD-O, then one more deletion after it (a recent tombstone the purge must keep).
		time.Sleep(20 * time.Millisecond)
		var cut time.Time
		if err := e.setup.QueryRow(context.Background(), `SELECT clock_timestamp()`).Scan(&cut); err != nil {
			t.Fatal(err)
		}
		time.Sleep(20 * time.Millisecond)
		if r := e.recDelete(t, owner, orgA, ids.ObjA1); r.Outcome != DeleteOutcomeDeleted {
			t.Errorf("recent object: %v", r.Outcome)
		}
		fmt.Println("RECOVERY-REPAIR " + mustJSON(map[string]string{"tombstone": ids.ObjA2, "cut": cut.UTC().Format(retTimeLayout)}))
	case "2":
		k, _ := objectstore.NewKey(ids.OrgA, ids.TombA1)
		if !e.present(t, k) {
			t.Fatal("no planted bytes under the seed tombstone")
		}
		if r := e.recDelete(t, owner, orgA, ids.TombA1); r.Outcome != DeleteOutcomeDeleted {
			t.Errorf("repeat delete of the tombstone: %v", r.Outcome)
		}
		if e.present(t, k) {
			t.Error("planted bytes survived the authorized repeat delete")
		}
	default:
		t.Fatal("RECOVERY_REPAIR_STEP must be 1 or 2")
	}
}
