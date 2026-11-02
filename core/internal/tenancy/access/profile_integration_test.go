//go:build integration

// G074 B-PROFILE-IT (Plan r3.1 §5, §8 L1): TestProfile against the access integration environment with the private
// MinIO of setupObjects (the scoped credential as infra/minio/init.sh provisions it). EP-1 slice: PRO-AV-0, the
// adapter round trip of the identity-owned avatars/<user>/<avatar> key family (A4): Put with the user binding, Head and
// Get read it back exactly, Delete removes it, and the objects/ family keeps its Organization binding; a binding of the
// other owner kind and a presign of an avatar key never reach the provider. The name and read cases on proWorld:
// PRO-1 create, change and clear of the explicit name (the effective name, E1 and the DTO field pins), PRO-2 the same
// global row from two tenant contexts, PRO-3 stale versions and input refusals before any transaction, PRO-4
// concurrent identical and conflicting edits, PRO-5 raw runtime writes as another subject on both tables, PRO-8 the
// trigger-owned audit rows and no-op silence (with PRO-1), PRO-9 the attributed denials, PRO-10 the statement counts
// and the picker's predicate calls bounded by its page, PRO-11 the relationship matrix at the facade with one not_found
// identity, PRO-12 revocation read in one session, PRO-13 the gated picker (R1) and the member pages, and PRO-14 the
// resolved P3 against a real sign-in provider rename. The avatar lifecycle (PRO-6, PRO-7, PRO-AV-1..12) is
// profile_avatar_integration_test.go's (Plan r3.1 §8 L2), run through proAvatarCases on the same environment and world.
package access

import (
	"bytes"
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"core/internal/objectstore"
	"core/internal/tenancy"
)

// proAvatarCases, when profile_avatar_integration_test.go (Plan r3.1 §8 L2) sets it from an init function, runs the
// avatar cases (PRO-6, PRO-7, PRO-AV-1..12) inside TestProfile on the same environment and world, after the name cases.
var proAvatarCases func(t *testing.T, e *objEnv, w *proWorld)

// proWorld is TestProfile's fixture (Plan r3.1 §5 "Fixed-actor worlds": the identity seed's roles on fresh users).
// Organization A: carol (org_owner), frank (org_admin), alice, bob, dave, dana, hostile, leaver and renamer
// (org_member); Group research (bob group_owner, dave group_member) and design (dana group_owner). Organization B:
// bianca (org_owner) and alice (org_member). Organization C: frank (org_admin), cyrus (org_member) and Group cg, in which
// dave holds a Group-only binding. own creates the Groups (each creator's owner binding) and holds no Organization
// binding; it is never a viewer or a subject.
type proWorld struct {
	A, B, C                  tenancy.OrganizationID
	research, design, cg     tenancy.Group
	carol, frank, alice, bob *itUser
	dave, dana, hostile      *itUser
	bianca, cyrus, leaver    *itUser
	renamer, own             *itUser
}

func (e *itEnv) proWorld(t *testing.T) *proWorld {
	t.Helper()
	c := e.canaries
	w := &proWorld{}
	w.A, w.B, w.C = e.org(t, "Profiles A "+c[0]), e.org(t, "Profiles B "+c[1]), e.org(t, "Profiles C "+c[2])
	for _, u := range []**itUser{&w.carol, &w.frank, &w.alice, &w.bob, &w.dave, &w.dana, &w.hostile, &w.bianca, &w.cyrus, &w.leaver,
		&w.renamer, &w.own} {
		*u = e.newUser(t)
	}
	w.research = e.group(t, w.A, w.own, "Research "+c[1], "profiles-research")
	w.design = e.group(t, w.A, w.own, "Design "+c[2], "profiles-design")
	w.cg = e.group(t, w.C, w.own, "Group C "+c[0], "profiles-c")
	e.grantOrg(t, w.A, w.carol, tenancy.OrgOwner)
	e.grantOrg(t, w.A, w.frank, tenancy.OrgAdmin)
	for _, u := range []*itUser{w.alice, w.bob, w.dave, w.dana, w.hostile, w.leaver, w.renamer} {
		e.grantOrg(t, w.A, u, tenancy.OrgMember)
	}
	e.grantGroup(t, w.A, w.research.ID, w.bob, tenancy.GroupOwner)
	e.grantGroup(t, w.A, w.research.ID, w.dave, tenancy.GroupMember)
	e.grantGroup(t, w.A, w.design.ID, w.dana, tenancy.GroupOwner)
	e.grantOrg(t, w.B, w.bianca, tenancy.OrgOwner)
	e.grantOrg(t, w.B, w.alice, tenancy.OrgMember)
	e.grantOrg(t, w.C, w.frank, tenancy.OrgAdmin)
	e.grantOrg(t, w.C, w.cyrus, tenancy.OrgMember)
	e.grantGroup(t, w.C, w.cg.ID, w.dave, tenancy.GroupMember)
	for _, u := range []*itUser{w.carol, w.frank, w.alice, w.bob, w.dave, w.dana, w.hostile, w.leaver, w.renamer} {
		e.bind(t, u, w.A)
	}
	e.bind(t, w.alice, w.B)
	e.bind(t, w.bianca, w.B)
	e.bind(t, w.frank, w.C)
	return w
}

// proIn resolves u's context in org with Group g selected (zero: Organization-only) in a new request scope.
func (e *itEnv) proIn(t *testing.T, u *itUser, org tenancy.OrganizationID, g tenancy.GroupID) (context.Context, *TenantContext) {
	t.Helper()
	rctx, _ := beginAs(t, e.svc, context.Background(), u, org)
	return rctx, resolve(t, e.svc, rctx, org, g)
}

// proProvider is u's provider-seeded app.users.display_name (attested superuser read).
func (e *itEnv) proProvider(t *testing.T, u *itUser) string {
	t.Helper()
	var s string
	if err := e.setup.QueryRow(context.Background(), `SELECT coalesce(display_name, '') FROM app.users WHERE id = $1::text::uuid`,
		u.id.String()).Scan(&s); err != nil {
		t.Fatalf("provider name: %v", err)
	}
	return s
}

// proRow is u's account_profiles row as text ("" without one; attested superuser read).
func (e *itEnv) proRow(t *testing.T, u *itUser) string {
	t.Helper()
	var s string
	err := e.setup.QueryRow(context.Background(), `SELECT p::text FROM app.account_profiles p WHERE user_id = $1::text::uuid`, u.id.String()).Scan(&s)
	if errors.Is(err, pgx.ErrNoRows) {
		return ""
	}
	if err != nil {
		t.Fatalf("profile row: %v", err)
	}
	return s
}

// proAudit counts u's profile.updated rows with reason as the trigger owns them (u as actor, subject and resource,
// resource type account_profile, no Organization or Group), and every profile.updated row naming u at all.
func (e *itEnv) proAudit(t *testing.T, u *itUser, reason string) (exact, all int) {
	t.Helper()
	if err := e.setup.QueryRow(context.Background(), `SELECT
  count(*) FILTER (WHERE reason = $2 AND actor_user_id = $1::text::uuid AND subject_user_id = $1::text::uuid
    AND resource_type = 'account_profile' AND resource_id = $1::text::uuid AND organization_id IS NULL AND group_id IS NULL),
  count(*)
FROM audit.events WHERE action = 'profile.updated' AND (subject_user_id = $1::text::uuid OR resource_id = $1::text::uuid)`,
		u.id.String(), reason).Scan(&exact, &all); err != nil {
		t.Fatalf("profile audit rows: %v", err)
	}
	return exact, all
}

// proDenials counts the access.denied rows of actor for op with reason, in org (nil: global), without a resource.
func (e *itEnv) proDenials(t *testing.T, actor *itUser, op, reason string, org *tenancy.OrganizationID) int {
	t.Helper()
	var o *string
	if org != nil {
		s := org.String()
		o = &s
	}
	var n int
	if err := e.setup.QueryRow(context.Background(), `SELECT count(*) FROM audit.events WHERE action = 'access.denied'
  AND actor_user_id = $1::text::uuid AND operation = $2 AND reason = $3 AND organization_id IS NOT DISTINCT FROM $4::text::uuid
  AND resource_type IS NULL AND resource_id IS NULL`, actor.id.String(), op, reason, o).Scan(&n); err != nil {
		t.Fatalf("denials: %v", err)
	}
	return n
}

// proFields lists the exported field names of v's struct type in declaration order.
func proFields(v any) string {
	rt := reflect.TypeOf(v)
	var out []string
	for i := range rt.NumField() {
		out = append(out, rt.Field(i).Name)
	}
	return strings.Join(out, ",")
}

// proBoth runs f1 and f2 concurrently from a common start and returns their errors.
func proBoth(f1, f2 func() error) [2]error {
	var wg sync.WaitGroup
	var out [2]error
	start := make(chan struct{})
	for i, f := range []func() error{f1, f2} {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			out[i] = f()
		}()
	}
	close(start)
	wg.Wait()
	return out
}

func proStr(s string) *string { return &s }

func TestProfile(t *testing.T) {
	e := setupObjects(t)
	ctx := context.Background()
	w := e.proWorld(t)
	var aliceName string // alice's explicit name after PRO-2 (used by PRO-12/PRO-13)

	t.Run("PRO-AV-0 the avatars key family round trip on the scoped credential (A4)", func(t *testing.T) {
		user, object, org := rlsUUID(t), rlsUUID(t), rlsUUID(t)
		key, err := objectstore.NewAvatarKey(user, object)
		if err != nil || key.Family() != objectstore.FamilyAvatars || key.String() != "avatars/"+user+"/"+object {
			t.Fatalf("avatar key: %v %v", key, err)
		}
		body := make([]byte, 4096)
		if _, err := rand.Read(body); err != nil {
			t.Fatal(err)
		}
		bind := objectstore.Binding{Object: object, User: user}
		if err := e.store.Put(ctx, key, bytes.NewReader(body), int64(len(body)), "image/png", bind); err != nil {
			t.Fatalf("avatar Put on the scoped credential: %v", err)
		}
		defer func() { _ = e.store.Delete(context.Background(), key) }()
		info, err := e.store.Metadata(ctx, key)
		if err != nil || info.Size != int64(len(body)) || info.Binding != bind {
			t.Fatalf("avatar Head: size %d binding %+v, %v; want %d %+v", info.Size, info.Binding, err, len(body), bind)
		}
		info, rc, err := e.store.Get(ctx, key)
		if err != nil {
			t.Fatalf("avatar Get: %v", err)
		}
		got, rerr := io.ReadAll(rc)
		_ = rc.Close()
		if rerr != nil || info.Size != int64(len(body)) || info.Binding != bind || !bytes.Equal(got, body) {
			t.Fatalf("avatar Get: %d bytes (equal %t) binding %+v, %v", len(got), bytes.Equal(got, body), info.Binding, rerr)
		}
		if err := e.store.Delete(ctx, key); err != nil {
			t.Fatalf("avatar Delete: %v", err)
		}
		if _, err := e.store.Metadata(ctx, key); !errors.Is(err, objectstore.ErrNotFound) {
			t.Fatalf("avatar Head after Delete: %v, want not_found", err)
		}

		// The objects/ family is unchanged: the Organization binding, no user-id.
		okey, err := objectstore.NewKey(org, object)
		if err != nil || okey.Family() != objectstore.FamilyObjects {
			t.Fatalf("object key: %v %v", okey, err)
		}
		obind := objectstore.Binding{Object: object, Organization: org}
		if err := e.store.Put(ctx, okey, bytes.NewReader(body[:16]), 16, "text/plain", obind); err != nil {
			t.Fatalf("object Put: %v", err)
		}
		defer func() { _ = e.store.Delete(context.Background(), okey) }()
		if info, err := e.store.Metadata(ctx, okey); err != nil || info.Size != 16 || info.Binding != obind {
			t.Fatalf("object Head: %+v %v", info, err)
		}
		if err := e.store.Delete(ctx, okey); err != nil {
			t.Fatalf("object Delete: %v", err)
		}

		// Refusals before any request: the other owner kind on either family, a presign of an avatar key.
		for i, c := range []struct {
			key objectstore.Key
			b   objectstore.Binding
		}{
			{key, objectstore.Binding{Object: object, Organization: org}},
			{okey, objectstore.Binding{Object: object, User: user}},
		} {
			if err := e.store.Put(ctx, c.key, bytes.NewReader(body[:1]), 1, "image/png", c.b); !errors.Is(err, objectstore.ErrInvalid) {
				t.Errorf("mismatched binding %d: %v, want invalid", i, err)
			}
		}
		if raw, err := e.store.SignedGetURL(ctx, key, time.Minute, "image/png"); !errors.Is(err, objectstore.ErrInvalid) || raw != "" {
			t.Errorf("avatar presign: %v, want invalid and no URL", err)
		}
		for _, k := range []objectstore.Key{key, okey} {
			if _, err := e.store.Metadata(ctx, k); !errors.Is(err, objectstore.ErrNotFound) {
				t.Errorf("%s family left an object behind: %v", k.Family(), err)
			}
		}
		fmt.Println("PRO-AV-0 {\"round_trip\":true}")
	})

	t.Run("PRO-1 create, change and clear the explicit name; the effective name, E1 and the DTO fields (with PRO-8)", func(t *testing.T) {
		// The DTOs carry exactly the reviewed fields (Plan r3.1 §3.5 visibility table): never an email, provider
		// identity, status, session, Organization list, object id, key or digest.
		for _, c := range []struct {
			v    any
			want string
		}{
			{AccountProfileView{}, "UserID,Name,EffectiveName,Avatar,Version"},
			{AccountProfile{}, "UserID,Name,Avatar"},
			{AccountAvatar{}, "Version,ContentType"},
		} {
			if got := proFields(c.v); got != c.want {
				t.Errorf("%T fields %s, want %s", c.v, got, c.want)
			}
		}
		actx, atc := e.as(t, w.alice, w.A)
		provider := e.proProvider(t, w.alice)
		before, _ := e.proAudit(t, w.alice, "name_changed")
		v, err := e.svc.GetOwnAccountProfile(actx, atc)
		if err != nil || v.UserID != w.alice.id || v.Name != nil || v.EffectiveName != provider || v.Avatar != nil || v.Version != 0 {
			t.Fatalf("own profile without a row: name set %t, effective %s, avatar %v, version %d, %v", v.Name != nil, itH(v.EffectiveName),
				v.Avatar, v.Version, err)
		}
		if e.proRow(t, w.alice) != "" {
			t.Fatal("a read created a profile row")
		}
		session := func() string {
			t.Helper()
			info, err := e.id.Session(ctx, w.alice.in(t, w.A), itOriginOf(w.A))
			if err != nil {
				t.Fatalf("session: %v", err)
			}
			return info.DisplayName
		}
		steps := []struct {
			what     string
			expected int32
			name     *string
			explicit *string // the explicit name afterwards (nil: none)
			version  int32
			audited  int // name_changed rows this step adds
		}{
			{"first set (trimmed)", 0, proStr("  Alice First  "), proStr("Alice First"), 1, 1},
			{"identical repeat on a stale version", 0, proStr("Alice First"), proStr("Alice First"), 1, 0},
			{"change", 1, proStr("Alice Second"), proStr("Alice Second"), 2, 1},
			{"200 code points", 2, proStr(strings.Repeat("\u00e9", 200)), proStr(strings.Repeat("\u00e9", 200)), 3, 1},
			{"clear with nil", 3, nil, nil, 4, 1},
			{"clear again (no-op, stale)", 0, nil, nil, 4, 0},
			{"white space only is a clear (no-op)", 4, proStr(" \u00a0 "), nil, 4, 0},
			{"set again", 4, proStr("Alice Third"), proStr("Alice Third"), 5, 1},
		}
		for _, s := range steps {
			b, _ := e.proAudit(t, w.alice, "name_changed")
			v, err := e.svc.UpdateAccountProfileName(actx, atc, s.expected, s.name)
			effective := provider
			if s.explicit != nil {
				effective = *s.explicit
			}
			if err != nil || v.Version != s.version || v.UserID != w.alice.id || v.EffectiveName != effective || v.Avatar != nil ||
				(v.Name == nil) != (s.explicit == nil) || (v.Name != nil && *v.Name != *s.explicit) {
				t.Fatalf("%s: version %d (want %d), name set %t, effective %s (want %s), %v", s.what, v.Version, s.version, v.Name != nil,
					itH(v.EffectiveName), itH(effective), err)
			}
			if a, _ := e.proAudit(t, w.alice, "name_changed"); a-b != s.audited {
				t.Errorf("%s: %d name_changed rows, want %d", s.what, a-b, s.audited)
			}
			// E1: the session's display name is the effective name immediately.
			if got := session(); got != effective {
				t.Errorf("%s: session display name %s, want %s", s.what, itH(got), itH(effective))
			}
			if g, err := e.svc.GetOwnAccountProfile(actx, atc); err != nil || g.Version != v.Version || g.EffectiveName != v.EffectiveName {
				t.Errorf("%s: own read after the write: version %d, %v", s.what, g.Version, err)
			}
		}
		// PRO-8: one name_changed row per set, change or clear, attributed only to alice; no avatar reason; no name in any
		// audit row.
		if a, all := e.proAudit(t, w.alice, "name_changed"); a-before != 5 || all-before != 5 {
			t.Errorf("name_changed rows %d of %d profile.updated rows added, want 5 of 5", a-before, all-before)
		}
		var leaks int
		if err := e.setup.QueryRow(ctx, `SELECT count(*) FROM audit.events e WHERE position('Alice First' IN e::text) > 0
  OR position('Alice Second' IN e::text) > 0 OR position('Alice Third' IN e::text) > 0`).Scan(&leaks); err != nil || leaks != 0 {
			t.Errorf("%d audit rows carry a profile name (%v)", leaks, err)
		}
	})

	t.Run("PRO-2 one global row read and written from two tenant contexts", func(t *testing.T) {
		actx, atc := e.as(t, w.alice, w.A)
		bctx, btc := e.as(t, w.alice, w.B)
		va, err := e.svc.GetOwnAccountProfile(actx, atc)
		if err != nil {
			t.Fatal(err)
		}
		vb, err := e.svc.GetOwnAccountProfile(bctx, btc)
		if err != nil || vb.Version != va.Version || vb.EffectiveName != va.EffectiveName {
			t.Fatalf("host B reads version %d (host A %d), same name %t, %v", vb.Version, va.Version, vb.EffectiveName == va.EffectiveName, err)
		}
		// A write from host B with host A's version lands on the same row, visible from host A next.
		aliceName = "Alice Override"
		vb, err = e.svc.UpdateAccountProfileName(bctx, btc, va.Version, proStr(aliceName))
		if err != nil || vb.Version != va.Version+1 || vb.EffectiveName != aliceName {
			t.Fatalf("host B write: version %d, %v", vb.Version, err)
		}
		if va, err = e.svc.GetOwnAccountProfile(actx, atc); err != nil || va.Version != vb.Version || va.Name == nil || *va.Name != aliceName {
			t.Fatalf("host A after the host B write: version %d, %v", va.Version, err)
		}
		// Each tenant's owner reads the same effective name (alice holds a binding in both Organizations).
		for _, c := range []struct {
			viewer *itUser
			org    tenancy.OrganizationID
		}{{w.carol, w.A}, {w.bianca, w.B}} {
			vctx, vtc := e.as(t, c.viewer, c.org)
			if p, err := e.svc.GetAccountProfile(vctx, vtc, w.alice.id); err != nil || p.UserID != w.alice.id || p.Name != aliceName || p.Avatar != nil {
				t.Errorf("owner read from %s: name %s, %v", itH(c.org), itH(p.Name), err)
			}
		}
	})

	t.Run("PRO-3 stale versions and input refusals change nothing", func(t *testing.T) {
		actx, atc := e.as(t, w.alice, w.A)
		cur, err := e.svc.GetOwnAccountProfile(actx, atc)
		if err != nil {
			t.Fatal(err)
		}
		row := e.proRow(t, w.alice)
		_, all := e.proAudit(t, w.alice, "name_changed")
		_, err = e.svc.UpdateAccountProfileName(actx, atc, cur.Version-1, proStr("Alice Stale"))
		memWantCode(t, "stale version", err, CodeConflict, fieldVersionConflict)
		_, err = e.svc.UpdateAccountProfileName(actx, atc, cur.Version+1, proStr("Alice Ahead"))
		memWantCode(t, "future version", err, CodeConflict, fieldVersionConflict)
		_, err = e.svc.UpdateAccountProfileName(actx, atc, cur.Version-1, nil)
		memWantCode(t, "stale clear", err, CodeConflict, fieldVersionConflict)
		// Input refusals happen before any transaction.
		for _, c := range []struct {
			what     string
			expected int32
			name     *string
			field    string
		}{
			{"negative version", -1, proStr("Alice"), "expected_version"},
			{"line feed", cur.Version, proStr("Ali\nce"), "name"},
			{"control character", cur.Version, proStr("Ali\x07ce"), "name"},
			{"line separator", cur.Version, proStr("Ali\u2028ce"), "name"},
			{"paragraph separator", cur.Version, proStr("Ali\u2029ce"), "name"},
			{"invalid UTF-8", cur.Version, proStr("Ali\xffce"), "name"},
			{"201 code points", cur.Version, proStr(strings.Repeat("n", 201)), "name"},
		} {
			e.hook.reset()
			_, err := e.svc.UpdateAccountProfileName(actx, atc, c.expected, c.name)
			memWantCode(t, c.what, err, CodeInvalidInput, c.field)
			if begins, _ := e.hook.counts(); begins != 0 {
				t.Errorf("%s: %d transactions, want none", c.what, begins)
			}
		}
		_, err = e.svc.GetAccountProfile(actx, atc, tenancy.UserID{})
		memWantCode(t, "zero subject", err, CodeInvalidInput, string(tenancy.CodeInvalidID))
		if got := e.proRow(t, w.alice); got != row {
			t.Error("a refused write changed the profile row")
		}
		if _, a := e.proAudit(t, w.alice, "name_changed"); a != all {
			t.Errorf("refused writes added %d profile.updated rows", a-all)
		}
	})

	t.Run("PRO-4 concurrent identical edits are idempotent; concurrent different edits conflict", func(t *testing.T) {
		c1, tc1 := e.as(t, w.hostile, w.A)
		c2, tc2 := e.as(t, w.hostile, w.A)
		edit := func(c context.Context, tc *TenantContext, expected int32, name string) func() error {
			return func() error {
				_, err := e.svc.UpdateAccountProfileName(c, tc, expected, proStr(name))
				return err
			}
		}
		_, before := e.proAudit(t, w.hostile, "name_changed")
		// Two first edits of a user without a row, then two identical changes: both succeed each time, one row change.
		for _, s := range []struct {
			expected int32
			name     string
			version  int32
		}{{0, "Hostile Same", 1}, {1, "Hostile Twice", 2}} {
			errs := proBoth(edit(c1, tc1, s.expected, s.name), edit(c2, tc2, s.expected, s.name))
			if errs[0] != nil || errs[1] != nil {
				t.Fatalf("identical concurrent edits to version %d: %v, %v", s.version, errs[0], errs[1])
			}
			v, err := e.svc.GetOwnAccountProfile(c1, tc1)
			if err != nil || v.Version != s.version || v.Name == nil || *v.Name != s.name {
				t.Fatalf("after identical concurrent edits: version %d (want %d), %v", v.Version, s.version, err)
			}
		}
		// Two different values on the same version: exactly one wins, the other is the version conflict.
		errs := proBoth(edit(c1, tc1, 2, "Hostile Left"), edit(c2, tc2, 2, "Hostile Right"))
		won := 0
		for _, err := range errs {
			if err == nil {
				won++
				continue
			}
			memWantCode(t, "the losing concurrent edit", err, CodeConflict, fieldVersionConflict)
		}
		if won != 1 {
			t.Fatalf("%d concurrent different edits succeeded, want 1 (%v, %v)", won, errs[0], errs[1])
		}
		v, err := e.svc.GetOwnAccountProfile(c1, tc1)
		if err != nil || v.Version != 3 || v.Name == nil || (*v.Name != "Hostile Left" && *v.Name != "Hostile Right") {
			t.Fatalf("after the conflicting edits: version %d, %v", v.Version, err)
		}
		if a, all := e.proAudit(t, w.hostile, "name_changed"); all-before != 3 || a != all {
			t.Errorf("concurrent edits recorded %d profile.updated rows (%d attributed), want 3", all-before, a)
		}
	})

	t.Run("PRO-5 raw runtime writes as another subject are refused on both tables", func(t *testing.T) {
		for _, viewer := range []*itUser{w.bob, w.carol} { // unrelated to alice; alice's Organization owner (a reader)
			func() {
				tx, err := e.pool.Begin(ctx)
				if err != nil {
					t.Fatal(err)
				}
				defer func() { _ = tx.Rollback(ctx) }()
				if err := tenancy.InitContext(ctx, tx, viewer.id, w.A, tenancy.GroupID{}); err != nil {
					t.Fatal(err)
				}
				a, av := w.alice.id.String(), rlsUUID(t)
				for _, c := range []struct {
					what, sql, state string
					args             []any
					rows             int64
				}{
					{"update another user's profile", `UPDATE app.account_profiles SET name = 'X', version = version + 1 WHERE user_id = $1::text::uuid`, "", []any{a}, 0},
					{"insert a profile for another user", `INSERT INTO app.account_profiles (user_id, name) VALUES ($1::text::uuid, 'X')`, "42501", []any{a}, 0},
					{"delete a profile", `DELETE FROM app.account_profiles WHERE user_id = $1::text::uuid`, "42501", []any{a}, 0},
					{"insert another user's avatar row", `INSERT INTO app.account_avatars (id, user_id, byte_size, content_type) VALUES ($1::text::uuid, $2::text::uuid, 10, 'image/png')`, "42501", []any{av, a}, 0},
					{"update another user's avatar rows", `UPDATE app.account_avatars SET state = 'deleting' WHERE user_id = $1::text::uuid`, "", []any{a}, 0},
					{"read another user's avatar rows", `SELECT id FROM app.account_avatars WHERE user_id = $1::text::uuid`, "", []any{a}, 0},
				} {
					if st, _, n := rlsStmtC(t, tx, c.sql, c.args...); st != c.state || n != c.rows {
						t.Errorf("%s as %s: SQLSTATE %q rows %d, want %q %d", c.what, itH(viewer.id), st, n, c.state, c.rows)
					}
				}
				// The row-level read: carol (Organization owner) reads alice's profile row, bob does not.
				want := int64(0)
				if viewer == w.carol {
					want = 1
				}
				if st, n := rlsStmt(t, tx, `SELECT name FROM app.account_profiles WHERE user_id = $1::text::uuid`, a); st != "" || n != want {
					t.Errorf("profile row read as %s: SQLSTATE %q rows %d, want %d", itH(viewer.id), st, n, want)
				}
			}()
		}
	})

	t.Run("PRO-9 denials are attributed to the profile operations", func(t *testing.T) {
		// A subject the caller may not read: not_found, recorded as profile.get target_not_found in the Organization,
		// without the subject.
		actx, atc := e.as(t, w.alice, w.A)
		before := e.proDenials(t, w.alice, "profile.get", "target_not_found", &w.A)
		_, err := e.svc.GetAccountProfile(actx, atc, w.dave.id)
		memWantCode(t, "unrelated subject", err, CodeNotFound, "")
		if got := e.proDenials(t, w.alice, "profile.get", "target_not_found", &w.A); got != before+1 {
			t.Errorf("profile.get target_not_found denials %d -> %d, want one more", before, got)
		}
		// A session whose Organization membership was removed: context_denied organization_not_permitted on its next
		// guarded call, recorded globally per operation (G007 semantics).
		lctx, ltc := e.as(t, w.leaver, w.A)
		if err := e.model.RemoveOrganizationMembership(ctx, w.A, w.leaver.id); err != nil {
			t.Fatal(err)
		}
		for _, c := range []struct {
			op   string
			call func() error
		}{
			{"profile.get", func() error { _, err := e.svc.GetOwnAccountProfile(lctx, ltc); return err }},
			{"profile.update", func() error {
				_, err := e.svc.UpdateAccountProfileName(lctx, ltc, 0, proStr("Leaver"))
				return err
			}},
		} {
			b := e.proDenials(t, w.leaver, c.op, "organization_not_permitted", nil)
			memWantCode(t, c.op+" after the membership removal", c.call(), CodeContextDenied, "")
			if got := e.proDenials(t, w.leaver, c.op, "organization_not_permitted", nil); got != b+1 {
				t.Errorf("%s organization_not_permitted denials %d -> %d, want one more", c.op, b, got)
			}
		}
		if e.proRow(t, w.leaver) != "" {
			t.Error("the denied write created a profile row")
		}
	})

	t.Run("PRO-10 statement counts and the picker's predicate calls bounded by its page", func(t *testing.T) {
		count := func(op string, fn func() error) (int, int) {
			t.Helper()
			e.hook.reset()
			if err := fn(); err != nil {
				t.Fatalf("%s: %v", op, err)
			}
			b, stmts := e.hook.counts()
			t.Logf(`PRO-10 {"operation":%q,"transactions":%d,"statements":%d}`, op, b, len(stmts))
			return b, len(stmts)
		}
		dctx, dtc := e.as(t, w.dave, w.A)
		cur, err := e.svc.GetOwnAccountProfile(dctx, dtc)
		if err != nil {
			t.Fatal(err)
		}
		bOwn, sOwn := count("get_own", func() error { _, err := e.svc.GetOwnAccountProfile(dctx, dtc); return err })
		bGet, sGet := count("get", func() error { _, err := e.svc.GetAccountProfile(dctx, dtc, w.bob.id); return err })
		bSet, sSet := count("update", func() error {
			_, err := e.svc.UpdateAccountProfileName(dctx, dtc, cur.Version, proStr("Dave Counted"))
			return err
		})
		bNoop, sNoop := count("update_noop", func() error {
			_, err := e.svc.UpdateAccountProfileName(dctx, dtc, cur.Version, proStr("Dave Counted"))
			return err
		})
		if bOwn != 1 || bGet != 1 || bSet != 1 || bNoop != 1 {
			t.Errorf("transactions get_own %d, get %d, update %d, update_noop %d; want one each", bOwn, bGet, bSet, bNoop)
		}
		// One profile statement for either read; the first write is the insert, a no-op is the lock and the read only.
		if sGet != sOwn {
			t.Errorf("get %d statements, get_own %d: want the same guard plus one profile statement", sGet, sOwn)
		}
		if cur.Version == 0 && sSet != sNoop+1 {
			t.Errorf("first update %d statements, no-op update %d: want exactly the insert more", sSet, sNoop)
		}
		e.hook.reset()

		// The picker's predicate runs once per emitted row (at most its page, never per candidate): the policy reader's
		// own policy reads account_profiles inside the body, so only the CASE consults app.rls_profile_visible.
		tx, err := e.setup.Begin(ctx)
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = tx.Rollback(ctx) }()
		calls := func() int {
			t.Helper()
			var n int
			if err := tx.QueryRow(ctx, `SELECT coalesce(sum(calls), 0)::int FROM pg_stat_xact_user_functions
WHERE funcid = 'app.rls_profile_visible(uuid)'::regprocedure`).Scan(&n); err != nil {
				t.Fatalf("function statistics: %v", err)
			}
			return n
		}
		page := func(limit int) int {
			t.Helper()
			for _, s := range []string{"SET LOCAL ROLE " + pgx.Identifier{e.cfg.RuntimeRole()}.Sanitize()} {
				if _, err := tx.Exec(ctx, s); err != nil {
					t.Fatal(err)
				}
			}
			if err := rlsSet(ctx, tx, proStr(w.bob.id.String()), proStr(w.A.String()), proStr(""), nil); err != nil {
				t.Fatal(err)
			}
			var n int
			if err := tx.QueryRow(ctx, `SELECT count(*)::int FROM app.rls_group_invitation_candidates($1::text::uuid, NULL, $2)`,
				w.research.ID.String(), limit).Scan(&n); err != nil {
				t.Fatalf("invitation candidates: %v", err)
			}
			if _, err := tx.Exec(ctx, `RESET ROLE`); err != nil {
				t.Fatal(err)
			}
			return n
		}
		if _, err := tx.Exec(ctx, `SET LOCAL track_functions = 'pl'`); err != nil {
			t.Fatal(err)
		}
		c0 := calls()
		one := page(1)
		c1 := calls()
		all := page(101)
		c2 := calls()
		if one != 1 || all < 2 || all > 101 || c1-c0 != one || c2-c1 != all {
			t.Errorf("picker pages: %d rows with %d predicate calls (page 1), %d rows with %d calls (page 101); want one call per emitted row",
				one, c1-c0, all, c2-c1)
		}
		t.Logf(`PRO-10 {"picker_rows":%d,"predicate_calls":%d}`, all, c2-c1)
	})

	t.Run("PRO-11 the relationship matrix at the facade with one not_found identity", func(t *testing.T) {
		unknown, err := tenancy.ParseUserID(rlsUUID(t))
		if err != nil {
			t.Fatal(err)
		}
		var notFound *Error
		for _, c := range []struct {
			name    string
			viewer  *itUser
			org     tenancy.OrganizationID
			group   tenancy.GroupID
			subject tenancy.UserID
			visible bool
		}{
			{"P1 self", w.alice, w.A, tenancy.GroupID{}, w.alice.id, true},
			{"P1 self from host B", w.alice, w.B, tenancy.GroupID{}, w.alice.id, true},
			{"P2 Organization admin", w.frank, w.A, tenancy.GroupID{}, w.alice.id, true},
			{"P2 Organization owner", w.carol, w.A, tenancy.GroupID{}, w.dave.id, true},
			{"P3 Group peer", w.dave, w.A, tenancy.GroupID{}, w.bob.id, true},
			{"P3 Group peer, reverse", w.bob, w.A, tenancy.GroupID{}, w.dave.id, true},
			{"P3 Group peer, research selected", w.dave, w.A, w.research.ID, w.bob.id, true},
			{"P4 admin of C, Group-only binding in C", w.frank, w.C, tenancy.GroupID{}, w.dave.id, true},
			{"N1 no shared Group", w.alice, w.A, tenancy.GroupID{}, w.dave.id, false},
			{"N1 no shared Group, Group owner", w.alice, w.A, tenancy.GroupID{}, w.bob.id, false},
			{"N1 different Groups", w.dave, w.A, tenancy.GroupID{}, w.dana.id, false},
			{"N1 plain member, Organization owner", w.dave, w.A, w.research.ID, w.carol.id, false},
			{"N2 foreign subject", w.hostile, w.A, tenancy.GroupID{}, w.bianca.id, false},
			{"N2 foreign subject of C", w.alice, w.A, tenancy.GroupID{}, w.cyrus.id, false},
			{"N2 admin of C, subject outside C", w.frank, w.C, tenancy.GroupID{}, w.alice.id, false},
			{"N2 owner of B, A-only subject", w.bianca, w.B, tenancy.GroupID{}, w.dave.id, false},
			{"unknown user", w.carol, w.A, tenancy.GroupID{}, unknown, false},
		} {
			vctx, vtc := e.proIn(t, c.viewer, c.org, c.group)
			p, err := e.svc.GetAccountProfile(vctx, vtc, c.subject)
			if c.visible {
				if err != nil || p.UserID != c.subject || p.Name == "" {
					t.Errorf("%s: %v, want the profile", c.name, err)
				}
				continue
			}
			var ae *Error
			if !errors.As(err, &ae) || ae.Code != CodeNotFound || p != (AccountProfile{}) {
				t.Errorf("%s: %v, want not_found and nothing", c.name, err)
				continue
			}
			if notFound == nil {
				notFound = ae
			} else if *ae != *notFound {
				t.Errorf("%s: not_found %+v differs from %+v", c.name, *ae, *notFound)
			}
		}
	})

	t.Run("PRO-12 and PRO-13 revocation in one session, the gated picker (R1) and the member pages", func(t *testing.T) {
		if aliceName == "" {
			t.Skip("PRO-2 did not set alice's explicit name")
		}
		providerAlice := e.proProvider(t, w.alice)
		actx, atc := e.as(t, w.alice, w.A)
		bctx, btc := e.as(t, w.bob, w.A)
		fctx, ftc := e.as(t, w.frank, w.A)
		cctx, ctc := e.as(t, w.carol, w.A)
		// picker returns the research candidates' ids and alice's name as issuer viewer sees it.
		picker := func(what string, c context.Context, tc *TenantContext) ([]tenancy.UserID, string) {
			t.Helper()
			pg, err := e.svc.ListInvitationCandidates(c, tc, w.research.ID, "", 0)
			if err != nil {
				t.Fatalf("%s picker: %v", what, err)
			}
			var ids []tenancy.UserID
			name := ""
			for _, cand := range pg.Candidates {
				ids = append(ids, cand.UserID)
				if cand.UserID == w.alice.id && cand.DisplayName != nil {
					name = *cand.DisplayName
				}
			}
			return ids, name
		}
		if got := proFields(Candidate{}); got != "UserID,DisplayName" {
			t.Errorf("Candidate fields %s: the picker carries no avatar field", got)
		}
		// R1: bob (the research owner, a plain Organization member, unrelated to alice) sees alice listed with her
		// provider name and may not read her profile; frank (Organization admin through the override) sees the explicit
		// name for the same candidate set.
		bobIDs, name := picker("bob", bctx, btc)
		if name != providerAlice {
			t.Errorf("unrelated issuer sees alice's name %s, want the provider name %s", itH(name), itH(providerAlice))
		}
		_, err := e.svc.GetAccountProfile(bctx, btc, w.alice.id)
		memWantCode(t, "unrelated issuer reads alice", err, CodeNotFound, "")
		e.withOverride(t, w.carol, w.A)
		frankIDs, name := picker("frank", fctx, ftc)
		if name != aliceName || !reflect.DeepEqual(frankIDs, bobIDs) {
			t.Errorf("Organization-admin issuer: alice's name %s (want %s), same candidates %t", itH(name), itH(aliceName), reflect.DeepEqual(frankIDs, bobIDs))
		}

		// bob and alice share Group X: the picker shows the explicit name and the profile reads both ways (PRO-12 P),
		// in sessions resolved before the grant.
		x := e.group(t, w.A, w.own, "Profiles X "+e.canaries[0], "profiles-x")
		e.grantGroup(t, w.A, x.ID, w.bob, tenancy.GroupOwner)
		e.grantGroup(t, w.A, x.ID, w.alice, tenancy.GroupMember)
		ids, name := picker("bob sharing X", bctx, btc)
		if name != aliceName || !reflect.DeepEqual(ids, bobIDs) {
			t.Errorf("issuer sharing a Group: alice's name %s (want %s), same candidates %t", itH(name), itH(aliceName), reflect.DeepEqual(ids, bobIDs))
		}
		if p, err := e.svc.GetAccountProfile(bctx, btc, w.alice.id); err != nil || p.Name != aliceName {
			t.Errorf("bob reads alice through X: %s, %v", itH(p.Name), err)
		}
		if p, err := e.svc.GetAccountProfile(actx, atc, w.bob.id); err != nil || p.UserID != w.bob.id {
			t.Errorf("alice reads bob through X: %v", err)
		}
		// The member pages: the effective name where the caller may read the profile, and the current avatar's version
		// only where one exists (dana's is seeded below).
		members := func(what string, c context.Context, tc *TenantContext, sc Scope) map[tenancy.UserID]Member {
			t.Helper()
			pg, err := e.svc.ListMembers(c, tc, sc, "", 0, "")
			if err != nil {
				t.Fatalf("%s members: %v", what, err)
			}
			out := map[tenancy.UserID]Member{}
			for _, m := range pg.Members {
				out[m.UserID] = m
			}
			return out
		}
		xID := x.ID
		if m, ok := members("bob in X", bctx, btc, Scope{Kind: tenancy.ScopeGroup, Group: &xID})[w.alice.id]; !ok || m.DisplayName == nil ||
			*m.DisplayName != aliceName || m.AvatarVersion != nil {
			t.Errorf("X member page: alice listed %t, explicit name %t, avatar version %v", ok, m.DisplayName != nil && *m.DisplayName == aliceName,
				m.AvatarVersion)
		}

		// Revocation (PRO-12 N3): the same sessions read nothing after the shared Group binding is removed, and the picker
		// is back to the provider name.
		if err := e.model.RemoveGroupMembership(ctx, w.A, x.ID, w.alice.id); err != nil {
			t.Fatal(err)
		}
		_, err = e.svc.GetAccountProfile(bctx, btc, w.alice.id)
		memWantCode(t, "bob reads alice after the removal", err, CodeNotFound, "")
		_, err = e.svc.GetAccountProfile(actx, atc, w.bob.id)
		memWantCode(t, "alice reads bob after the removal", err, CodeNotFound, "")
		if _, name := picker("bob after the removal", bctx, btc); name != providerAlice {
			t.Errorf("after the removal the issuer sees alice's name %s, want the provider name", itH(name))
		}

		// A current ready avatar (seeded for dana as her own write would leave it: the avatar row ready, the reference at
		// version 1) is visible as its version on the Organization member page and as {version, type} through the
		// projection; members without one carry none.
		av := rlsUUID(t)
		func() {
			tx, err := e.setup.Begin(ctx)
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = tx.Rollback(ctx) }()
			if err := rlsSet(ctx, tx, proStr(w.dana.id.String()), proStr(""), proStr(""), nil); err != nil {
				t.Fatal(err)
			}
			for _, s := range []struct {
				sql  string
				args []any
			}{
				{`INSERT INTO app.account_avatars (id, user_id, state, byte_size, content_type, sha256, ready_at)
VALUES ($1::text::uuid, $2::text::uuid, 'ready', 10, 'image/png', sha256('pro-13'::bytea), now())`, []any{av, w.dana.id.String()}},
				{`INSERT INTO app.account_profiles (user_id, avatar_object_id, avatar_version) VALUES ($2::text::uuid, $1::text::uuid, 1)`,
					[]any{av, w.dana.id.String()}},
			} {
				if _, err := tx.Exec(ctx, s.sql, s.args...); err != nil {
					t.Fatalf("avatar fixture: %v", err)
				}
			}
			if err := tx.Commit(ctx); err != nil {
				t.Fatal(err)
			}
		}()
		org := members("carol", cctx, ctc, Scope{Kind: tenancy.ScopeOrganization})
		for id, m := range org {
			switch {
			case id == w.dana.id:
				if m.AvatarVersion == nil || *m.AvatarVersion != 1 {
					t.Errorf("dana's avatar version on the member page: %v, want 1", m.AvatarVersion)
				}
			case m.AvatarVersion != nil:
				t.Errorf("member %s carries avatar version %d without an avatar", itH(id), *m.AvatarVersion)
			}
		}
		if m := org[w.alice.id]; m.DisplayName == nil || *m.DisplayName != aliceName {
			t.Error("the Organization member page does not show alice's explicit name to her owner")
		}
		if m := org[w.bob.id]; m.DisplayName == nil || *m.DisplayName != e.proProvider(t, w.bob) {
			t.Error("the Organization member page does not show bob's provider name (no profile row)")
		}
		if p, err := e.svc.GetAccountProfile(cctx, ctc, w.dana.id); err != nil || p.Avatar == nil || p.Avatar.Version != 1 ||
			p.Avatar.ContentType != "image/png" {
			t.Errorf("projection of dana's avatar: %+v, %v", p.Avatar, err)
		}
		dctx, dtc := e.as(t, w.dana, w.A)
		if v, err := e.svc.GetOwnAccountProfile(dctx, dtc); err != nil || v.Avatar == nil || v.Avatar.Version != 1 || v.Version != 1 || v.Name != nil {
			t.Errorf("dana's own view with the avatar: %+v version %d, %v", v.Avatar, v.Version, err)
		}
		// An unrelated reader still sees nothing of it.
		_, err = e.svc.GetAccountProfile(bctx, btc, w.dana.id)
		memWantCode(t, "unrelated reader of the avatar owner", err, CodeNotFound, "")
	})

	t.Run("PRO-14 an explicit name survives a provider rename; clearing returns the current provider name (P3)", func(t *testing.T) {
		rctx, rtc := e.as(t, w.renamer, w.A)
		cctx, ctc := e.as(t, w.carol, w.A)
		provider0 := e.proProvider(t, w.renamer)
		v, err := e.svc.UpdateAccountProfileName(rctx, rtc, 0, proStr("Explicit Renamer"))
		if err != nil || v.EffectiveName != "Explicit Renamer" {
			t.Fatalf("explicit name: %v", err)
		}
		row := e.proRow(t, w.renamer)
		_, audited := e.proAudit(t, w.renamer, "name_changed")
		// The provider renames the user: a real sign-in with another name, whose identity refresh (identity/store.go
		// issue) writes app.users.display_name only.
		cred := e.signInAt(t, e.id, w.A, w.renamer.sub, "Renamed", nil, "", "")
		provider1 := e.proProvider(t, w.renamer)
		if provider1 == provider0 {
			t.Fatal("the sign-in did not change the provider name")
		}
		if e.proRow(t, w.renamer) != row {
			t.Error("the sign-in refresh wrote the profile row")
		}
		if _, a := e.proAudit(t, w.renamer, "name_changed"); a != audited {
			t.Error("the sign-in refresh recorded a profile change")
		}
		effective := func(what, want string) {
			t.Helper()
			if g, err := e.svc.GetOwnAccountProfile(rctx, rtc); err != nil || g.EffectiveName != want {
				t.Errorf("%s: own effective name %s, want %s (%v)", what, itH(g.EffectiveName), itH(want), err)
			}
			if p, err := e.svc.GetAccountProfile(cctx, ctc, w.renamer.id); err != nil || p.Name != want {
				t.Errorf("%s: owner reads %s, want %s (%v)", what, itH(p.Name), itH(want), err)
			}
			info, err := e.id.Session(ctx, cred, e.originOf(cred))
			if err != nil || info.DisplayName != want {
				t.Errorf("%s: session display name %s, want %s (%v)", what, itH(info.DisplayName), itH(want), err)
			}
		}
		effective("after the provider rename", "Explicit Renamer")
		if v, err = e.svc.UpdateAccountProfileName(rctx, rtc, v.Version, nil); err != nil || v.Name != nil || v.EffectiveName != provider1 {
			t.Fatalf("clear: effective %s, want the current provider name %s (%v)", itH(v.EffectiveName), itH(provider1), err)
		}
		effective("after the clear", provider1)
	})

	if proAvatarCases != nil {
		proAvatarCases(t, e, w)
	}
}
