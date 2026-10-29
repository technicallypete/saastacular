//go:build integration

// G017 B-BRANDING-IT (Plan r2 §1.3, §1.4, §6.2): TestBranding against the access integration environment (real
// PostgreSQL at version 13, the attested runtime role, the real guarded facade). Slice 1: BRD-1 owner/admin update and
// reset, BRD-2 member/Group-owner/foreign denial with access.denied and zero effects, BRD-3 label normalization and
// invalid label input, BRD-4 invalid colour input and contrast refusal, BRD-5 stale versions, BRD-6 concurrent writes,
// BRD-7 the audit rows (one organization.branding_changed per write, no values) and the no-op paths. Slice 2 (real
// MinIO, setupObjects): BRD-8 PNG/JPEG normalization and publication, BRD-9 refused images with zero effects, BRD-10
// the logo publication stages P/A/R/C on the existing objectHooks seam and the hookDB commit modes (plus a test-local
// held-COMMIT wrapper for the in-flight case), BRD-11 the public logo read and the auth reader's visibility, BRD-12 the
// statement counts (PG-008), BRD-14 the guarded delete and purge of the current logo.
package access

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"core/internal/identity"
	"core/internal/objectstore"
	"core/internal/tenancy"
)

func (e *itEnv) brandingWorld(t *testing.T) domWorld {
	t.Helper()
	return e.domWorld(t)
}

// brandingEvents lists the Organization's organization.branding_changed rows as outcome/reason/resource pairs, oldest
// first.
func (e *itEnv) brandingEvents(t *testing.T, org tenancy.OrganizationID) []string {
	t.Helper()
	rows, err := e.setup.Query(context.Background(), `SELECT outcome || '/' || coalesce(reason, '') || '/' || coalesce(resource_type, '') || '/' ||
  (resource_id = organization_id)::text FROM audit.events WHERE action = 'organization.branding_changed' AND organization_id = $1::text::uuid
ORDER BY occurred_at, id`, org.String())
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			t.Fatal(err)
		}
		out = append(out, s)
	}
	return out
}

func strp(s string) *string { return &s }

func TestBranding(t *testing.T) {
	e := setupObjects(t)
	ctx := context.Background()
	w := e.brandingWorld(t)
	octx, otc := e.as(t, w.owner, w.org)
	actx, atc := e.as(t, w.admin, w.org)
	// White primary-foreground on blue keeps 4.5:1 in every T8 state; every other entry inherits the kit defaults, which
	// pass through their G069 kit-baseline floors (owner waiver, 2026-10-26). G069 S23 retyped the earlier blue #1d4ed8
	// (80 % hover over the page about 4.42; Codex P0 disposition, 2026-10-26); the floor would now admit it too, and the
	// retyped value is kept.
	valid := map[string]string{"primary": "#1E40AF"}

	t.Run("BRD-1 owner and admin update and reset", func(t *testing.T) {
		v, err := e.svc.GetBranding(octx, otc)
		if err != nil || v.Version != 0 || v.Label != nil || len(v.Light) != 0 || v.Logo != nil {
			t.Fatalf("initial view: %+v %v", v, err)
		}
		v, err = e.svc.UpdateBranding(octx, otc, BrandingUpdate{ExpectedVersion: 0, Label: strp("  Acme <Tools> "), Light: valid, Dark: map[string]string{}})
		if err != nil || v.Version != 1 || v.Label == nil || *v.Label != "Acme <Tools>" || v.Light["primary"] != "#1e40af" {
			t.Fatalf("owner first update: %+v %v", v, err)
		}
		v, err = e.svc.UpdateBranding(actx, atc, BrandingUpdate{ExpectedVersion: 1, Label: nil, Light: valid, Dark: valid})
		if err != nil || v.Version != 2 || v.Label != nil || v.Dark["primary"] != "#1e40af" {
			t.Fatalf("admin update: %+v %v", v, err)
		}
		if g, err := e.svc.GetBranding(actx, atc); err != nil || g.Version != 2 || g.Dark["primary"] != "#1e40af" {
			t.Fatalf("admin read: %+v %v", g, err)
		}
		v, err = e.svc.ResetBranding(actx, atc, 2)
		if err != nil || v.Version != 3 || v.Label != nil || len(v.Light) != 0 || len(v.Dark) != 0 || v.Logo != nil {
			t.Fatalf("admin reset: %+v %v", v, err)
		}
		v, err = e.svc.UpdateBranding(octx, otc, BrandingUpdate{ExpectedVersion: 3, Label: strp("Acme"), Light: valid})
		if err != nil || v.Version != 4 {
			t.Fatalf("owner update after reset: %+v %v", v, err)
		}
	})

	t.Run("BRD-2 member, Group owner and foreign denial with zero effects", func(t *testing.T) {
		cur, err := e.svc.GetBranding(octx, otc)
		if err != nil {
			t.Fatal(err)
		}
		before := e.part(t, "organization_branding", w.org) + e.part(t, "organization_branding", w.other)
		denied := func() int {
			var n int
			if err := e.setup.QueryRow(ctx, `SELECT count(*) FROM audit.events WHERE action = 'access.denied' AND operation LIKE 'branding.%'
AND organization_id = $1::text::uuid`, w.org.String()).Scan(&n); err != nil {
				t.Fatal(err)
			}
			return n
		}
		n0 := denied()
		for _, d := range []struct {
			name string
			u    *itUser
		}{{"member", w.member}, {"group owner", w.gowner}} {
			dctx, dtc := e.as(t, d.u, w.org)
			wantCode(t, d.name+" get", func() error { _, err := e.svc.GetBranding(dctx, dtc); return err }(), CodeForbidden)
			wantCode(t, d.name+" update", func() error {
				_, err := e.svc.UpdateBranding(dctx, dtc, BrandingUpdate{ExpectedVersion: cur.Version, Label: strp("Mine"), Light: valid})
				return err
			}(), CodeForbidden)
			wantCode(t, d.name+" reset", func() error { _, err := e.svc.ResetBranding(dctx, dtc, cur.Version); return err }(), CodeForbidden)
			wantCode(t, d.name+" remove logo", func() error { _, err := e.svc.RemoveBrandingLogo(dctx, dtc, cur.Version); return err }(), CodeForbidden)
		}
		if n1 := denied(); n1 != n0+8 {
			t.Errorf("access.denied branding rows %d -> %d, want +8", n0, n1)
		}
		// A foreign Organization's owner cannot resolve this Organization's context; in its own context it sees only its
		// own (absent) branding.
		rctx, _ := beginAs(t, e.svc, ctx, w.foreign, w.org)
		if _, err := e.svc.Resolve(rctx, ResolveRequest{OrganizationID: w.org.String()}); err == nil {
			t.Error("a non-member resolved the Organization context")
		}
		fctx, ftc := e.as(t, w.foreign, w.other)
		if v, err := e.svc.GetBranding(fctx, ftc); err != nil || v.Version != 0 || v.Label != nil {
			t.Errorf("foreign own view: %+v %v", v, err)
		}
		if got := e.part(t, "organization_branding", w.org) + e.part(t, "organization_branding", w.other); got != before {
			t.Fatal("a denied operation changed branding rows")
		}
	})

	t.Run("BRD-3 label normalization and invalid labels commit nothing", func(t *testing.T) {
		cur, err := e.svc.GetBranding(octx, otc)
		if err != nil {
			t.Fatal(err)
		}
		// 80 decomposed "e + combining acute" pairs (160 code points) normalize to 80 precomposed code points: accepted and
		// stored composed (the accepted trim + NFC contract, through PostgreSQL's normalize).
		decomposed := " " + strings.Repeat("e\u0301", 80) + " "
		v, err := e.svc.UpdateBranding(octx, otc, BrandingUpdate{ExpectedVersion: cur.Version, Label: &decomposed, Light: valid})
		if err != nil || v.Label == nil || *v.Label != strings.Repeat("\u00e9", 80) {
			t.Fatalf("decomposed label: %+v %v", v, err)
		}
		before := e.part(t, "organization_branding", w.org)
		for name, label := range map[string]string{
			"control":          "Acme\x00Tools",
			"c1 control":       "Acme\u0085Tools",
			"invalid utf-8":    "Acme\xffTools",
			"81 code points":   strings.Repeat("a", 81),
			"81 after NFC":     strings.Repeat("e\u0301", 81),
			"escape":           "Acme\x1b[31m",
			"embedded newline": "Acme\nTools",
			"delete character": "Acme\x7f",
			"81 emoji":         strings.Repeat("\U0001F600", 81),
		} {
			_, err := e.svc.UpdateBranding(octx, otc, BrandingUpdate{ExpectedVersion: v.Version, Label: &label, Light: valid})
			wantCode(t, name, err, CodeInvalidInput)
			if ae, ok := err.(*Error); !ok || ae.Field != "display_label" {
				t.Errorf("%s: field %v", name, err)
			}
		}
		if e.part(t, "organization_branding", w.org) != before {
			t.Fatal("a refused label changed the branding row")
		}
		// Whitespace only is "no label".
		v, err = e.svc.UpdateBranding(octx, otc, BrandingUpdate{ExpectedVersion: v.Version, Label: strp(" \t "), Light: valid})
		if err != nil || v.Label != nil {
			t.Fatalf("whitespace label: %+v %v", v, err)
		}
	})

	t.Run("BRD-4 invalid colours and contrast refusal commit nothing", func(t *testing.T) {
		cur, err := e.svc.GetBranding(octx, otc)
		if err != nil {
			t.Fatal(err)
		}
		before := e.part(t, "organization_branding", w.org)
		many := map[string]string{}
		for i := 0; i < 16; i++ {
			many[fmt.Sprintf("token-%d", i)] = "#000000"
		}
		for name, c := range map[string]struct {
			light, dark map[string]string
			code        Code
			field       string
		}{
			"unknown key":       {map[string]string{"overlay": "#000000"}, nil, CodeInvalidInput, "light"},
			"short hex":         {map[string]string{"primary": "#fff"}, nil, CodeInvalidInput, "light"},
			"named colour":      {nil, map[string]string{"primary": "red"}, CodeInvalidInput, "dark"},
			"bad digit":         {map[string]string{"primary": "#12345g"}, nil, CodeInvalidInput, "light"},
			"css injection":     {map[string]string{"primary": "#000000;}body{"}, nil, CodeInvalidInput, "light"},
			"too many keys":     {many, nil, CodeInvalidInput, "light"},
			"dark background":   {map[string]string{"background": "#000000"}, nil, CodeContrastFailed, "T1:light"},
			"dark primary loss": {nil, map[string]string{"primary": "#ffffff", "primary-foreground": "#ffffff"}, CodeContrastFailed, "T8:dark"},
		} {
			_, err := e.svc.UpdateBranding(octx, otc, BrandingUpdate{ExpectedVersion: cur.Version, Light: c.light, Dark: c.dark})
			wantCode(t, name, err, c.code)
			if ae, ok := err.(*Error); !ok || ae.Field != c.field {
				t.Errorf("%s: %v, want field %q", name, err, c.field)
			}
		}
		wantCode(t, "negative version", func() error {
			_, err := e.svc.UpdateBranding(octx, otc, BrandingUpdate{ExpectedVersion: -1, Light: valid})
			return err
		}(), CodeInvalidInput)
		if e.part(t, "organization_branding", w.org) != before {
			t.Fatal("a refused colour configuration changed the branding row")
		}
	})

	t.Run("BRD-5 stale versions are conflict with zero effects", func(t *testing.T) {
		cur, err := e.svc.GetBranding(octx, otc)
		if err != nil {
			t.Fatal(err)
		}
		before := e.part(t, "organization_branding", w.org)
		for _, stale := range []int64{0, cur.Version - 1, cur.Version + 1} {
			wantCode(t, fmt.Sprintf("update at %d", stale), func() error {
				_, err := e.svc.UpdateBranding(octx, otc, BrandingUpdate{ExpectedVersion: stale, Label: strp("Stale"), Light: valid})
				return err
			}(), CodeConflict)
			wantCode(t, fmt.Sprintf("reset at %d", stale), func() error { _, err := e.svc.ResetBranding(octx, otc, stale); return err }(), CodeConflict)
			wantCode(t, fmt.Sprintf("remove logo at %d", stale), func() error {
				_, err := e.svc.RemoveBrandingLogo(octx, otc, stale)
				return err
			}(), CodeConflict)
		}
		if e.part(t, "organization_branding", w.org) != before {
			t.Fatal("a stale write changed the branding row")
		}
		// With no row, only version 0 is current.
		fctx, ftc := e.as(t, w.foreign, w.other)
		wantCode(t, "reset of no row at 1", func() error { _, err := e.svc.ResetBranding(fctx, ftc, 1); return err }(), CodeConflict)
	})

	t.Run("BRD-6 concurrent writes: one wins, the other is conflict", func(t *testing.T) {
		// One request scope per writer (as DOM-2): the owner and the admin race at the same expected version.
		race := func(org tenancy.OrganizationID, users []*itUser, expected int64) []Code {
			type scope struct {
				ctx context.Context
				tc  *TenantContext
			}
			scopes := make([]scope, len(users))
			for i, u := range users {
				scopes[i].ctx, scopes[i].tc = e.as(t, u, org)
			}
			var wg sync.WaitGroup
			codes := make([]Code, len(users))
			for i := range codes {
				wg.Add(1)
				go func() {
					defer wg.Done()
					_, err := e.svc.UpdateBranding(scopes[i].ctx, scopes[i].tc, BrandingUpdate{ExpectedVersion: expected,
						Label: strp(fmt.Sprintf("Racer %d", i)), Light: valid})
					codes[i] = CodeOf(err)
				}()
			}
			wg.Wait()
			return codes
		}
		cur, err := e.svc.GetBranding(octx, otc)
		if err != nil {
			t.Fatal(err)
		}
		if c := race(w.org, []*itUser{w.owner, w.admin}, cur.Version); !((c[0] == "" && c[1] == CodeConflict) || (c[0] == CodeConflict && c[1] == "")) {
			t.Errorf("concurrent updates: %v", c)
		}
		if v, err := e.svc.GetBranding(octx, otc); err != nil || v.Version != cur.Version+1 {
			t.Errorf("after the update race: %+v %v", v, err)
		}
		// Two concurrent first writes of an Organization without a row: the primary key decides.
		org := e.org(t, "Branding Race")
		e.grantOrg(t, org, w.owner, tenancy.OrgOwner)
		e.grantOrg(t, org, w.admin, tenancy.OrgAdmin)
		e.bind(t, w.owner, org)
		e.bind(t, w.admin, org)
		rctx, rtc := e.as(t, w.owner, org)
		if c := race(org, []*itUser{w.owner, w.admin}, 0); !((c[0] == "" && c[1] == CodeConflict) || (c[0] == CodeConflict && c[1] == "")) {
			t.Errorf("concurrent first writes: %v", c)
		}
		if v, err := e.svc.GetBranding(rctx, rtc); err != nil || v.Version != 1 {
			t.Errorf("after the insert race: %+v %v", v, err)
		}
	})

	t.Run("BRD-7 one audit row per write without values; no-op paths write nothing", func(t *testing.T) {
		org := e.org(t, "Branding Audit")
		e.grantOrg(t, org, w.owner, tenancy.OrgOwner)
		e.bind(t, w.owner, org)
		bctx, btc := e.as(t, w.owner, org)
		// No row: reset and logo removal at version 0 are successful no-ops (no row, no event).
		if v, err := e.svc.ResetBranding(bctx, btc, 0); err != nil || v.Version != 0 {
			t.Fatalf("reset of no row: %+v %v", v, err)
		}
		if v, err := e.svc.RemoveBrandingLogo(bctx, btc, 0); err != nil || v.Version != 0 {
			t.Fatalf("remove logo of no row: %+v %v", v, err)
		}
		if got := e.part(t, "organization_branding", org); got != "-" {
			t.Fatalf("a no-op created a row: %s", got)
		}
		v, err := e.svc.UpdateBranding(bctx, btc, BrandingUpdate{Label: strp("Audit Co"), Light: valid})
		if err != nil {
			t.Fatal(err)
		}
		// Logo removal without a reference keeps the version.
		if r, err := e.svc.RemoveBrandingLogo(bctx, btc, v.Version); err != nil || r.Version != v.Version {
			t.Fatalf("remove logo without a reference: %+v %v", r, err)
		}
		if v, err = e.svc.UpdateBranding(bctx, btc, BrandingUpdate{ExpectedVersion: v.Version, Label: strp("Audit Co 2")}); err != nil {
			t.Fatal(err)
		}
		if v, err = e.svc.ResetBranding(bctx, btc, v.Version); err != nil || v.Version != 3 {
			t.Fatalf("reset: %+v %v", v, err)
		}
		// Reset of an already-default row writes nothing either.
		if r, err := e.svc.ResetBranding(bctx, btc, v.Version); err != nil || r.Version != v.Version {
			t.Fatalf("second reset: %+v %v", r, err)
		}
		want := []string{"success/updated/organization/true", "success/updated/organization/true", "success/reset/organization/true"}
		if got := e.brandingEvents(t, org); strings.Join(got, ",") != strings.Join(want, ",") {
			t.Fatalf("branding events: %v, want %v", got, want)
		}
		// No value is recorded: the label and the colours appear in no audit column.
		var n int
		if err := e.setup.QueryRow(ctx, `SELECT count(*) FROM audit.events e WHERE e.organization_id = $1::text::uuid
AND (e::text LIKE '%Audit Co%' OR e::text LIKE '%1e40af%')`, org.String()).Scan(&n); err != nil || n != 0 {
			t.Errorf("audit rows carrying branding values: %d (%v)", n, err)
		}
	})

	// BRD-13 (AC7, C-1; Plan r2 §1.5): the existing session route answers membership for the origin's Organization,
	// re-read on every call; removing the binding in-test flips it to "none" on the next request, and the credential stays
	// unauthenticated at another Organization's origin (the indicator never makes a foreign-host credential usable).
	t.Run("BRD-13 the session answer carries the current membership of the origin's Organization", func(t *testing.T) {
		h := identity.NewHandler(e.id)
		session := func(cred, origin string) (int, map[string]any) {
			req := httptest.NewRequest(http.MethodGet, "/internal/v1/auth/session?origin="+url.QueryEscape(origin), nil)
			req.Header.Set("Authorization", "Session "+cred)
			rec := httptest.NewRecorder()
			h.Session(rec, req)
			var b map[string]any
			if err := json.Unmarshal(rec.Body.Bytes(), &b); err != nil {
				t.Fatalf("session body: %v", err)
			}
			return rec.Code, b
		}
		u := e.newUser(t)
		e.grantOrg(t, w.org, u, tenancy.OrgMember)
		e.bind(t, u, w.org)
		origin := e.itOriginFor(w.org)
		if code, b := session(u.in(t, w.org), origin); code != http.StatusOK || b["membership"] != "member" || b["user_id"] != u.id.String() {
			t.Fatalf("member: %d %v", code, b)
		}
		if _, err := e.setupDB.Exec(ctx, `DELETE FROM app.role_bindings WHERE subject_user_id = $1::text::uuid AND organization_id = $2::text::uuid
AND scope_type = 'organization' AND source = 'membership'`, u.id.String(), w.org.String()); err != nil {
			t.Fatal(err)
		}
		if code, b := session(u.in(t, w.org), origin); code != http.StatusOK || b["membership"] != "none" || b["user_id"] != u.id.String() {
			t.Fatalf("binding removed: %d %v", code, b)
		}
		if code, b := session(u.in(t, w.org), e.itOriginFor(w.other)); code != http.StatusUnauthorized || b["error"] != "unauthenticated" || b["membership"] != nil {
			t.Fatalf("foreign origin: %d %v", code, b)
		}
		if member, err := e.id.Membership(ctx, u.id.String(), w.org.String()); err != nil || member {
			t.Fatalf("membership read after removal: %v %v", member, err)
		}
		if member, err := e.id.Membership(ctx, w.owner.id.String(), w.org.String()); err != nil || !member {
			t.Fatalf("owner membership: %v %v", member, err)
		}
	})

	e.brandingLogoCases(t, w)
}

// --- slice 2: logo publication, public read, footprint, current-logo deletion -------------------------------------------

// brdImage is a w×h test image with varied pixels (and alpha), encoded as PNG at level or as JPEG at quality q (q > 0).
func brdImage(t *testing.T, w, h int, level png.CompressionLevel, q int) []byte {
	t.Helper()
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	for i := range img.Pix {
		img.Pix[i] = byte(i*7 + i/5)
	}
	var b bytes.Buffer
	var err error
	if q > 0 {
		err = jpeg.Encode(&b, img, &jpeg.Options{Quality: q})
	} else {
		err = (&png.Encoder{CompressionLevel: level}).Encode(&b, img)
	}
	if err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}

// brdOrg is a fresh Organization owned by u; its objects (rows and bytes) are removed when the test ends.
func (e *objEnv) brdOrg(t *testing.T, u *itUser, name string) tenancy.OrganizationID {
	t.Helper()
	org := e.org(t, name)
	e.grantOrg(t, org, u, tenancy.OrgOwner)
	e.bind(t, u, org)
	t.Cleanup(func() { e.cleanupOrg(t, org) })
	return org
}

// brdScope is a request scope of u's credential for org on svc and its resolved context.
func brdScope(t *testing.T, svc *Service, u *itUser, org tenancy.OrganizationID) (context.Context, *TenantContext) {
	t.Helper()
	rctx, end := beginAs(t, svc, context.Background(), u, org)
	t.Cleanup(end)
	return rctx, resolve(t, svc, rctx, org, tenancy.GroupID{})
}

func brdSetLogo(svc *Service, ctx context.Context, tc *TenantContext, version int64, data []byte) (BrandingView, error) {
	return svc.SetBrandingLogo(ctx, tc, version, bytes.NewReader(data), int64(len(data)))
}

func brdErr(err error) Error {
	var ae *Error
	if errors.As(err, &ae) {
		return *ae
	}
	return Error{Code: "not-an-access-error"}
}

func brdIDs(us []BrandingLogo) string {
	ids := make([]string, 0, len(us))
	for _, u := range us {
		ids = append(ids, u.ObjectID+":"+u.State)
	}
	return strings.Join(ids, ",")
}

// brdOneCommit makes the next COMMIT of h use mode m and every later one normal.
func brdOneCommit(h *hookDB, m commitMode) {
	h.setCommit(m)
	h.on("before", "COMMIT", func() { h.setCommit(commitNormal) })
}

// brdPool is a single-statement reader over the runtime pool (the server's pool shape) counting transactions and
// statements; the anonymous public read needs QueryRow, which hookDB does not offer.
type brdPool struct {
	*pgxpool.Pool
	mu              sync.Mutex
	begins, queries int
}

func (p *brdPool) BeginTx(ctx context.Context, o pgx.TxOptions) (pgx.Tx, error) {
	p.mu.Lock()
	p.begins++
	p.mu.Unlock()
	return p.Pool.BeginTx(ctx, o)
}

func (p *brdPool) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	p.mu.Lock()
	p.queries++
	p.mu.Unlock()
	return p.Pool.QueryRow(ctx, sql, args...)
}

func (p *brdPool) take() (begins, queries int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	b, q := p.begins, p.queries
	p.begins, p.queries = 0, 0
	return b, q
}

// brdHeld wraps a hookDB for the in-flight case: once armed, the next COMMIT is held — it reports a lost reply
// (io.ErrUnexpectedEOF) without committing, the transaction stays open (its deferred Rollback is a no-op while held),
// and release commits it.
type brdHeld struct {
	*hookDB
	mu    sync.Mutex
	armed bool
	held  pgx.Tx
}

type brdHeldTx struct {
	pgx.Tx
	d    *brdHeld
	held bool
}

func (d *brdHeld) BeginTx(ctx context.Context, o pgx.TxOptions) (pgx.Tx, error) {
	tx, err := d.hookDB.BeginTx(ctx, o)
	if err != nil {
		return nil, err
	}
	return &brdHeldTx{Tx: tx, d: d}, nil
}

func (t *brdHeldTx) Commit(ctx context.Context) error {
	t.d.mu.Lock()
	armed := t.d.armed
	if armed {
		t.d.armed, t.d.held, t.held = false, t.Tx, true
	}
	t.d.mu.Unlock()
	if armed {
		return io.ErrUnexpectedEOF
	}
	return t.Tx.Commit(ctx)
}

func (t *brdHeldTx) Rollback(ctx context.Context) error {
	if t.held {
		return nil
	}
	return t.Tx.Rollback(ctx)
}

func (d *brdHeld) arm() {
	d.mu.Lock()
	d.armed = true
	d.mu.Unlock()
}

func (d *brdHeld) release(ctx context.Context) error {
	d.mu.Lock()
	tx := d.held
	d.held = nil
	d.mu.Unlock()
	if tx == nil {
		return errors.New("no held transaction")
	}
	return tx.Commit(ctx)
}

func (e *objEnv) brandingLogoCases(t *testing.T, w domWorld) {
	ctx := context.Background()
	pool := &brdPool{Pool: e.pool}
	psvc := e.objService(pool, e.store, e.ocfg) // the public read (single statements on the pool)

	t.Run("BRD-8 PNG and JPEG are normalized, stored and published", func(t *testing.T) {
		org := e.brdOrg(t, w.owner, "Branding Logo")
		octx, otc := brdScope(t, e.osvc, w.owner, org)
		var first string
		for i, c := range []struct {
			name, ct, format string
			data             []byte
		}{
			{"png", "image/png", "png", brdImage(t, 40, 24, png.BestSpeed, 0)},
			{"jpeg", "image/jpeg", "jpeg", brdImage(t, 40, 24, 0, 50)},
		} {
			v, err := brdSetLogo(e.osvc, octx, otc, int64(i), c.data)
			if err != nil || v.Version != int64(i+1) || v.Logo == nil || v.Logo.State != logoReady || v.Logo.ContentType != c.ct {
				t.Fatalf("%s: %+v %v", c.name, v, err)
			}
			k, err := objectstore.NewKey(org.String(), v.Logo.ObjectID)
			if err != nil {
				t.Fatal(err)
			}
			info, rc, err := e.store.Get(ctx, k)
			if err != nil {
				t.Fatalf("%s: stored bytes: %v", c.name, err)
			}
			data, err := io.ReadAll(rc)
			_ = rc.Close()
			sum := sha256.Sum256(data)
			if err != nil || bytes.Equal(data, c.data) || hex.EncodeToString(sum[:6]) != v.Logo.SHA12 || int64(len(data)) != v.Logo.ByteSize ||
				info.Binding.Organization != org.String() || info.Binding.Group != "" {
				t.Errorf("%s: stored %d bytes (input %d), sha12 %s/%s, binding %+v (%v)", c.name, len(data), len(c.data),
					hex.EncodeToString(sum[:6]), v.Logo.SHA12, info.Binding, err)
			}
			cfg, format, err := image.DecodeConfig(bytes.NewReader(data))
			if err != nil || cfg.Width != 40 || cfg.Height != 24 || format != c.format {
				t.Errorf("%s: normalized %s %dx%d (%v)", c.name, format, cfg.Width, cfg.Height, err)
			}
			g, err := e.osvc.GetBranding(octx, otc)
			if err != nil || g.Logo == nil || g.Logo.ObjectID != v.Logo.ObjectID || g.Logo.SHA12 != v.Logo.SHA12 {
				t.Fatalf("%s: view %+v %v", c.name, g, err)
			}
			// The replaced logo stays private: unreferenced, ready, listed for the owner, never public.
			if i == 0 {
				first = v.Logo.ObjectID
				if len(g.Uploads) != 0 || g.UploadsTotal != 0 {
					t.Errorf("uploads after the first logo: %s", brdIDs(g.Uploads))
				}
			} else if brdIDs(g.Uploads) != first+":ready" || g.UploadsTotal != 1 {
				t.Errorf("uploads after the replacement: %s (%d)", brdIDs(g.Uploads), g.UploadsTotal)
			}
			p, err := psvc.PublicBranding(ctx, org.String())
			if err != nil || p.Revision != v.Version || p.Logo == nil || p.Logo.V != v.Logo.SHA12 || p.Logo.ContentType != c.ct {
				t.Errorf("%s: public %+v %v", c.name, p, err)
			}
		}
		want := "success/logo_set/organization/true,success/logo_set/organization/true"
		if got := strings.Join(e.brandingEvents(t, org), ","); got != want {
			t.Errorf("events %s, want %s", got, want)
		}
	})

	t.Run("BRD-9 refused images store nothing", func(t *testing.T) {
		org := e.brdOrg(t, w.owner, "Branding Refused")
		octx, otc := brdScope(t, e.osvc, w.owner, org)
		valid := brdImage(t, 32, 32, png.DefaultCompression, 0)
		anyLogo := map[string]bool{"logo_type": true, "logo_malformed": true, "logo_dimensions": true}
		for _, c := range []struct {
			name  string
			data  []byte
			size  int64
			field string // "" = any image refusal
		}{
			{"svg", []byte(`<svg xmlns="http://www.w3.org/2000/svg" width="32" height="32"/>`), -1, "logo_type"},
			{"html", []byte(`<!DOCTYPE html><html><body>x</body></html>`), -1, "logo_type"},
			{"gif", []byte("GIF89a\x20\x00\x20\x00\x00\x00\x00;"), -1, "logo_type"},
			{"truncated png", valid[:len(valid)/2], -1, ""},
			{"bytes after IEND", append(bytes.Clone(valid), []byte("<script>")...), -1, "logo_malformed"},
			{"15 px", brdImage(t, 15, 15, png.DefaultCompression, 0), -1, "logo_dimensions"},
			{"over 512 KiB", make([]byte, 524289), -1, "logo_too_large"},
			{"declared size mismatch", valid, int64(len(valid) + 1), "logo_size"},
		} {
			size := c.size
			if size < 0 {
				size = int64(len(c.data))
			}
			_, err := e.osvc.SetBrandingLogo(octx, otc, 0, bytes.NewReader(c.data), size)
			ae := brdErr(err)
			if ae.Code != CodeInvalidInput || (c.field != "" && ae.Field != c.field) || (c.field == "" && !anyLogo[ae.Field]) || ae.Object != "" {
				t.Errorf("%s: %+v, want invalid_input/%s", c.name, ae, c.field)
			}
		}
		if got := e.objectsOf(t, org); len(got) != 0 {
			t.Errorf("objects after refused images: %v", got)
		}
		if g, err := e.osvc.GetBranding(octx, otc); err != nil || g.Version != 0 || g.Logo != nil || len(g.Uploads) != 0 {
			t.Errorf("view after refused images: %+v %v", g, err)
		}
		if got := e.part(t, "organization_branding", org); got != "-" || len(e.brandingEvents(t, org)) != 0 {
			t.Errorf("refused images wrote a row (%s) or events", got)
		}
		// D-1: a plain member's malformed image is refused by authorization before any decoding (forbidden, not
		// logo_invalid), and stores nothing.
		mctx, mtc := brdScope(t, e.osvc, w.member, w.org)
		objects := len(e.objectsOf(t, w.org))
		for _, data := range [][]byte{valid[:len(valid)/2], append(bytes.Clone(valid), []byte("<script>")...)} {
			_, err := e.osvc.SetBrandingLogo(mctx, mtc, 0, bytes.NewReader(data), int64(len(data)))
			if ae := brdErr(err); ae.Code != CodeForbidden || ae.Field != "" {
				t.Errorf("member malformed image: %+v, want forbidden", ae)
			}
		}
		if got := len(e.objectsOf(t, w.org)); got != objects {
			t.Errorf("member malformed images changed objects %d -> %d", objects, got)
		}
	})

	logo := brdImage(t, 48, 48, png.BestSpeed, 0)
	type fcase struct {
		f    *faultSvc
		org  tenancy.OrganizationID
		fctx context.Context
		ftc  *TenantContext
		octx context.Context // the plain storage facade (concurrent writes, final reads)
		otc  *TenantContext
	}
	newCase := func(t *testing.T, name string) fcase {
		c := fcase{f: e.faultService(e.ocfg), org: e.brdOrg(t, w.owner, name)}
		c.fctx, c.ftc = brdScope(t, c.f.svc, w.owner, c.org)
		c.octx, c.otc = brdScope(t, e.osvc, w.owner, c.org)
		return c
	}
	stageOf := func(want string, fn func()) func(op, stage string) {
		return func(op, stage string) {
			if op == opSetBrandingLogo && stage == want {
				fn()
			}
		}
	}
	concurrent := func(t *testing.T, c fcase, expected int64) func() {
		return func() {
			if _, err := e.osvc.UpdateBranding(c.octx, c.otc, BrandingUpdate{ExpectedVersion: expected, Label: strp("Concurrent")}); err != nil {
				t.Errorf("concurrent update: %v", err)
			}
		}
	}
	// after returns the published view and the Organization's object states.
	after := func(t *testing.T, c fcase) (BrandingView, map[tenancy.ObjectID]string) {
		t.Helper()
		g, err := e.osvc.GetBranding(c.octx, c.otc)
		if err != nil {
			t.Fatalf("view: %v", err)
		}
		return g, e.objectsOf(t, c.org)
	}
	stateOf := func(states map[tenancy.ObjectID]string, id string) string {
		oid, err := tenancy.ParseObjectID(id)
		if err != nil {
			return "bad-id"
		}
		if s, ok := states[oid]; ok {
			return s
		}
		return logoAbsent
	}

	t.Run("BRD-10 logo publication stages report the true outcome", func(t *testing.T) {
		t.Run("crash after P: the upload is listed, nothing published", func(t *testing.T) {
			c := newCase(t, "Branding P crash")
			c.f.hook.crash = func(op, stage string) bool { return op == opSetBrandingLogo && stage == stageLogoUploaded }
			_, err := brdSetLogo(c.f.svc, c.fctx, c.ftc, 0, logo)
			g, states := after(t, c)
			if ae := brdErr(err); ae.Code != CodeUnavailable || g.Version != 0 || g.Logo != nil || len(g.Uploads) != 1 ||
				g.Uploads[0].State != logoReady || len(states) != 1 {
				t.Errorf("%+v; view %+v; objects %v", ae, g, states)
			}
		})
		t.Run("definitive conflict at A: compensated and reported", func(t *testing.T) {
			c := newCase(t, "Branding A conflict")
			c.f.hook.stage = stageOf(stageLogoUploaded, concurrent(t, c, 0))
			_, err := brdSetLogo(c.f.svc, c.fctx, c.ftc, 0, logo)
			ae := brdErr(err)
			g, states := after(t, c)
			if ae.Code != CodeConflict || ae.Field != fieldCompensated || ae.ObjectState != logoDeleted || ae.Object == "" ||
				stateOf(states, ae.Object) != logoDeleted || g.Version != 1 || g.Logo != nil || len(g.Uploads) != 0 {
				t.Errorf("%+v; view %+v; objects %v", ae, g, states)
			}
		})
		t.Run("unknown at A, committed: R returns the view, nothing deleted", func(t *testing.T) {
			c := newCase(t, "Branding A lost reply")
			c.f.hook.stage = stageOf(stageLogoAssociate, func() { brdOneCommit(c.f.h, commitLostReply) })
			v, err := brdSetLogo(c.f.svc, c.fctx, c.ftc, 0, logo)
			g, states := after(t, c)
			if err != nil || v.Version != 1 || v.Logo == nil || g.Logo == nil || g.Logo.ObjectID != v.Logo.ObjectID ||
				stateOf(states, v.Logo.ObjectID) != logoReady || len(states) != 1 {
				t.Errorf("%+v %v; view %+v; objects %v", v, err, g, states)
			}
		})
		t.Run("unknown at A, rolled back: retained and reported, then the owner deletes it", func(t *testing.T) {
			c := newCase(t, "Branding A no commit")
			c.f.hook.stage = stageOf(stageLogoAssociate, func() { brdOneCommit(c.f.h, commitLostNoCommit) })
			_, err := brdSetLogo(c.f.svc, c.fctx, c.ftc, 0, logo)
			ae := brdErr(err)
			g, states := after(t, c)
			if ae.Code != CodeUnavailable || ae.Field != fieldPublicationUnknown || ae.ObjectState != logoReady || ae.Object == "" ||
				stateOf(states, ae.Object) != logoReady || g.Version != 0 || g.Logo != nil || brdIDs(g.Uploads) != ae.Object+":ready" {
				t.Fatalf("%+v; view %+v; objects %v", ae, g, states)
			}
			res, err := e.osvc.DeleteBrandingUpload(c.octx, c.otc, g.Version, ae.Object)
			if _, states = after(t, c); err != nil || res.Outcome != DeleteOutcomeDeleted || stateOf(states, ae.Object) != logoDeleted {
				t.Errorf("guarded delete: %+v %v; objects %v", res, err, states)
			}
		})
		t.Run("unknown at A while A is in flight: retained, then published", func(t *testing.T) {
			c := newCase(t, "Branding A in flight")
			d := &brdHeld{hookDB: &hookDB{pool: e.pool}}
			svc := e.objService(d, e.store, e.ocfg)
			svc.hooks = &objectHooks{stage: stageOf(stageLogoAssociate, d.arm)}
			hctx, htc := brdScope(t, svc, w.owner, c.org)
			cctx, cancel := context.WithTimeout(hctx, 60*time.Second)
			_, err := brdSetLogo(svc, cctx, htc, 0, logo)
			cancel()
			ae := brdErr(err)
			// R ran before A's COMMIT was sent: nothing references the object yet, so the outcome is unknown and the
			// object is retained (never deleted while it may still be published).
			if ae.Code != CodeUnavailable || ae.Field != fieldPublicationUnknown || ae.ObjectState != logoReady || ae.Object == "" {
				t.Errorf("in flight: %+v", ae)
			}
			if err := d.release(ctx); err != nil {
				t.Fatalf("release the held COMMIT: %v", err)
			}
			g, states := after(t, c)
			if g.Version != 1 || g.Logo == nil || g.Logo.ObjectID != ae.Object || g.Logo.State != logoReady || stateOf(states, ae.Object) != logoReady {
				t.Errorf("after the held COMMIT: %+v; objects %v", g, states)
			}
		})
		t.Run("unknown at A with an advanced version: retained and reported", func(t *testing.T) {
			c := newCase(t, "Branding R advanced")
			c.f.hook.stage = func(op, stage string) {
				switch {
				case op != opSetBrandingLogo:
				case stage == stageLogoAssociate:
					brdOneCommit(c.f.h, commitLostNoCommit)
				case stage == stageLogoReport:
					concurrent(t, c, 0)()
				}
			}
			_, err := brdSetLogo(c.f.svc, c.fctx, c.ftc, 0, logo)
			ae := brdErr(err)
			g, states := after(t, c)
			if ae.Field != fieldPublicationUnknown || ae.ObjectState != logoReady || g.Version != 1 || g.Logo != nil ||
				brdIDs(g.Uploads) != ae.Object+":ready" || stateOf(states, ae.Object) != logoReady {
				t.Errorf("%+v; view %+v; objects %v", ae, g, states)
			}
		})
		t.Run("R's read failing: unread, retained", func(t *testing.T) {
			c := newCase(t, "Branding R unread")
			c.f.hook.stage = func(op, stage string) {
				switch {
				case op != opSetBrandingLogo:
				case stage == stageLogoAssociate:
					brdOneCommit(c.f.h, commitLostNoCommit)
				case stage == stageLogoReport:
					c.f.h.setFail("C")
				}
			}
			_, err := brdSetLogo(c.f.svc, c.fctx, c.ftc, 0, logo)
			c.f.h.setFail("")
			ae := brdErr(err)
			g, states := after(t, c)
			if ae.Code != CodeUnavailable || ae.Field != fieldPublicationUnknown || ae.ObjectState != logoUnread || ae.Object == "" ||
				stateOf(states, ae.Object) != logoReady || g.Logo != nil {
				t.Errorf("%+v; view %+v; objects %v", ae, g, states)
			}
		})
		t.Run("compensation failure: the actual state is reported and the upload listed", func(t *testing.T) {
			c := newCase(t, "Branding C failure")
			c.f.hook.stage = stageOf(stageLogoUploaded, concurrent(t, c, 0))
			c.f.hook.crash = func(op, stage string) bool { return op == opSetBrandingLogo && stage == stageLogoCompensate }
			_, err := brdSetLogo(c.f.svc, c.fctx, c.ftc, 0, logo)
			ae := brdErr(err)
			g, states := after(t, c)
			if ae.Code != CodeConflict || ae.Field != fieldCompensationFailed || ae.ObjectState != logoReady ||
				stateOf(states, ae.Object) != logoReady || g.Logo != nil || brdIDs(g.Uploads) != ae.Object+":ready" {
				t.Errorf("%+v; view %+v; objects %v", ae, g, states)
			}
		})
		t.Run("cancellation between P and A: nothing published, the outcome is truthful", func(t *testing.T) {
			c := newCase(t, "Branding P cancel")
			cctx, cancel := context.WithCancel(c.fctx)
			defer cancel()
			c.f.hook.stage = stageOf(stageLogoUploaded, cancel)
			_, err := brdSetLogo(c.f.svc, cctx, c.ftc, 0, logo)
			ae := brdErr(err)
			g, states := after(t, c)
			got := stateOf(states, ae.Object)
			ok := ae.Object != "" && g.Version == 0 && g.Logo == nil
			switch ae.Field {
			case fieldCompensated:
				ok = ok && got == logoDeleted
			case fieldCompensationPending:
				ok = ok && got == logoDeleting
			case fieldCompensationFailed, fieldPublicationUnknown:
				ok = ok && got == logoReady && (ae.ObjectState == logoReady || ae.ObjectState == logoUnread) &&
					brdIDs(g.Uploads) == ae.Object+":ready"
			default:
				ok = false
			}
			if !ok {
				t.Errorf("%+v; view %+v; objects %v", ae, g, states)
			}
			t.Logf("BRD-10 cancel outcome: %s/%s", ae.Field, ae.ObjectState)
		})
	})

	t.Run("BRD-11 the public logo is the current ready logo of the resolved Organization only", func(t *testing.T) {
		orgA, orgB := e.brdOrg(t, w.owner, "Branding Public A"), e.brdOrg(t, w.owner, "Branding Public B")
		actx, atc := brdScope(t, e.osvc, w.owner, orgA)
		bctx, btc := brdScope(t, e.osvc, w.owner, orgB)
		a1, err := brdSetLogo(e.osvc, actx, atc, 0, brdImage(t, 20, 20, png.BestSpeed, 0))
		if err != nil {
			t.Fatal(err)
		}
		a2, err := brdSetLogo(e.osvc, actx, atc, 1, brdImage(t, 24, 24, png.BestSpeed, 0))
		if err != nil {
			t.Fatal(err)
		}
		b1, err := brdSetLogo(e.osvc, bctx, btc, 0, brdImage(t, 30, 30, 0, 60))
		if err != nil {
			t.Fatal(err)
		}
		other, err := e.osvc.PutObject(actx, atc, PutObjectRequest{ContentType: "image/png", Size: 3, Body: io.NopCloser(strings.NewReader("abc"))})
		if err != nil {
			t.Fatal(err)
		}
		ct, data, err := psvc.PublicBrandingLogo(ctx, orgA.String(), a2.Logo.SHA12)
		sum := sha256.Sum256(data)
		if err != nil || ct != "image/png" || hex.EncodeToString(sum[:6]) != a2.Logo.SHA12 {
			t.Errorf("current logo: %s %d bytes %v", ct, len(data), err)
		}
		for name, v := range map[string]string{"replaced (unreferenced) logo": a1.Logo.SHA12, "B's logo through A's host": b1.Logo.SHA12,
			"not a revision": "zzzzzzzzzzzz", "empty": ""} {
			if _, _, err := psvc.PublicBrandingLogo(ctx, orgA.String(), v); CodeOf(err) != CodeNotFound {
				t.Errorf("%s: %v, want not_found", name, err)
			}
		}
		if _, _, err := psvc.PublicBrandingLogo(ctx, orgB.String(), b1.Logo.SHA12); err != nil {
			t.Errorf("B's logo through B's host: %v", err)
		}
		// B's logo object under A's runtime context is denied as before G017.
		bID, _ := tenancy.ParseObjectID(b1.Logo.ObjectID)
		if _, rc, err := e.osvc.GetObject(actx, atc, bID); err == nil {
			_ = rc.Close()
			t.Error("B's logo object read under A's context")
		}
		// As the auth reader: only current logos are visible, any Organization's public projection is readable.
		tx, err := e.setup.Begin(ctx)
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = tx.Rollback(ctx) }()
		if _, err := tx.Exec(ctx, "SET LOCAL ROLE "+pgx.Identifier{e.cfg.Database + "_auth_reader"}.Sanitize()); err != nil {
			t.Fatal(err)
		}
		for id, want := range map[string]int{a2.Logo.ObjectID: 1, b1.Logo.ObjectID: 1, a1.Logo.ObjectID: 0, other.ID.String(): 0} {
			var n int
			if err := tx.QueryRow(ctx, `SELECT count(id) FROM app.stored_objects WHERE id = $1::text::uuid`, id).Scan(&n); err != nil || n != want {
				t.Errorf("auth reader sees %s: %d (%v), want %d", id, n, err, want)
			}
		}
		var rev int64
		if err := tx.QueryRow(ctx, `SELECT revision FROM app.public_branding($1::text::uuid)`, orgB.String()).Scan(&rev); err != nil || rev != b1.Version {
			t.Errorf("auth reader public_branding(B): %d (%v)", rev, err)
		}
	})

	t.Run("BRD-12 statement counts", func(t *testing.T) {
		org := e.brdOrg(t, w.owner, "Branding Footprint")
		octx, otc := brdScope(t, e.osvc, w.owner, org)
		count := func(op string, fn func() error) (int, int) {
			e.hook.reset()
			if err := fn(); err != nil {
				t.Fatalf("%s: %v", op, err)
			}
			b, stmts := e.hook.counts()
			t.Logf(`BRD-12 {"operation":%q,"transactions":%d,"statements":%d}`, op, b, len(stmts))
			return b, len(stmts)
		}
		var v BrandingView
		if b, _ := count("update", func() (err error) {
			v, err = e.osvc.UpdateBranding(octx, otc, BrandingUpdate{Label: strp("Footprint"), Light: map[string]string{"primary": "#1E40AF"}})
			return err
		}); b != 1 {
			t.Errorf("update: %d transactions, want 1", b)
		}
		count("set_logo", func() (err error) { v, err = brdSetLogo(e.osvc, octx, otc, v.Version, logo); return err })
		if b, _ := count("get", func() error { _, err := e.osvc.GetBranding(octx, otc); return err }); b != 1 {
			t.Errorf("get: %d transactions, want 1", b)
		}
		pool.take()
		if _, err := psvc.PublicBranding(ctx, org.String()); err != nil {
			t.Fatal(err)
		}
		pb, pq := pool.take()
		t.Logf(`BRD-12 {"operation":"public","transactions":%d,"statements":%d}`, pb, pq)
		if _, _, err := psvc.PublicBrandingLogo(ctx, org.String(), v.Logo.SHA12); err != nil {
			t.Fatal(err)
		}
		lb, lq := pool.take()
		t.Logf(`BRD-12 {"operation":"public_logo","transactions":%d,"statements":%d,"object_reads":1}`, lb, lq)
		if pb != 0 || pq != 1 || lb != 0 || lq != 1 {
			t.Errorf("public reads: %d/%d and %d/%d transactions/statements, want 0/1 each", pb, pq, lb, lq)
		}
	})

	t.Run("BRD-14 current-logo deletion and purge fall back; removal clears", func(t *testing.T) {
		org := e.brdOrg(t, w.owner, "Branding Delete")
		octx, otc := brdScope(t, e.osvc, w.owner, org)
		v, err := brdSetLogo(e.osvc, octx, otc, 0, logo)
		if err != nil {
			t.Fatal(err)
		}
		id, _ := tenancy.ParseObjectID(v.Logo.ObjectID)
		if res, err := e.osvc.DeleteObject(octx, otc, id); err != nil || (res.Outcome != DeleteOutcomeDeleted && res.Outcome != DeleteOutcomeDeletionPending) {
			t.Fatalf("guarded delete of the current logo: %+v %v", res, err)
		}
		g, err := e.osvc.GetBranding(octx, otc)
		if err != nil || g.Version != v.Version || g.Logo == nil || g.Logo.ObjectID != v.Logo.ObjectID || g.Logo.SHA12 != "" ||
			(g.Logo.State != logoDeleted && g.Logo.State != logoDeleting) {
			t.Errorf("view after the delete: %+v %v", g.Logo, err)
		}
		if p, err := psvc.PublicBranding(ctx, org.String()); err != nil || p.Logo != nil || p.Revision != v.Version {
			t.Errorf("public after the delete: %+v %v", p, err)
		}
		if _, _, err := psvc.PublicBrandingLogo(ctx, org.String(), v.Logo.SHA12); CodeOf(err) != CodeNotFound {
			t.Errorf("deleted logo bytes: %v, want not_found", err)
		}
		e.unseed(t, org, id) // a simulated purge of the row (the fixture role; G009's purge is not run here)
		if g, err = e.osvc.GetBranding(octx, otc); err != nil || g.Logo == nil || g.Logo.State != logoAbsent {
			t.Errorf("view after the purge: %+v %v", g.Logo, err)
		}
		r, err := e.osvc.RemoveBrandingLogo(octx, otc, g.Version)
		if err != nil || r.Logo != nil || r.Version != g.Version+1 {
			t.Errorf("removal: %+v %v", r, err)
		}
		want := "success/logo_set/organization/true,success/logo_removed/organization/true"
		if got := strings.Join(e.brandingEvents(t, org), ","); got != want {
			t.Errorf("events %s, want %s", got, want)
		}
	})
}
