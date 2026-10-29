//go:build integration

// B-OBJ-IT (G009 Plan r1 §5.2): TestObjects against real PostgreSQL (the attested runtime login) and the private MinIO
// of the migrations suite (the scoped credential), with the independent oracle (objects_expected_test.go) driving the
// read, write and RLS matrices. Every denial is paired with an authorized positive, and both systems' owned state is
// inspected. Each test records the rows and keys it creates and removes exactly those.
//
// Required environment, beside setupIT's: OBJECT_STORE_ENDPOINT, OBJECT_STORE_REGION, OBJECT_STORE_BUCKET,
// OBJECT_STORE_FORCE_PATH_STYLE, OBJECT_STORE_ACCESS_KEY and OBJECT_STORE_SECRET_KEY (the scoped credential; MinIO root
// credentials must be absent), read through config.Load like the server.
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
	"net/http"
	"net/url"
	"os"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"core/internal/config"
	"core/internal/objectstore"
	"core/internal/tenancy"
)

// --- environment -------------------------------------------------------------------------------------------------------

// objSpy wraps the real store: it counts provider calls and lets fault and ordering tests intercept each call.
type objSpy struct {
	inner objectstore.Store
	calls atomic.Int64
	mu    sync.Mutex
	// before, when set, runs before the named call ("put", "get", "delete", "head", "sign"); a non-nil error replaces
	// the call's result without reaching the provider.
	before func(op string, key objectstore.Key) error
	// putOverride, when set, replaces Put entirely.
	putOverride func(ctx context.Context, key objectstore.Key, body io.Reader, size int64, ct string, b objectstore.Binding) error
}

func (s *objSpy) hook(op string, key objectstore.Key) error {
	s.calls.Add(1)
	s.mu.Lock()
	f := s.before
	s.mu.Unlock()
	if f != nil {
		return f(op, key)
	}
	return nil
}

func (s *objSpy) set(before func(op string, key objectstore.Key) error) {
	s.mu.Lock()
	s.before = before
	s.mu.Unlock()
}

func (s *objSpy) Put(ctx context.Context, key objectstore.Key, body io.Reader, size int64, ct string, b objectstore.Binding) error {
	if err := s.hook("put", key); err != nil {
		return err
	}
	s.mu.Lock()
	o := s.putOverride
	s.mu.Unlock()
	if o != nil {
		return o(ctx, key, body, size, ct, b)
	}
	return s.inner.Put(ctx, key, body, size, ct, b)
}

func (s *objSpy) Get(ctx context.Context, key objectstore.Key) (objectstore.Info, io.ReadCloser, error) {
	if err := s.hook("get", key); err != nil {
		return objectstore.Info{}, nil, err
	}
	return s.inner.Get(ctx, key)
}

func (s *objSpy) Delete(ctx context.Context, key objectstore.Key) error {
	if err := s.hook("delete", key); err != nil {
		return err
	}
	return s.inner.Delete(ctx, key)
}

func (s *objSpy) Metadata(ctx context.Context, key objectstore.Key) (objectstore.Info, error) {
	if err := s.hook("head", key); err != nil {
		return objectstore.Info{}, err
	}
	return s.inner.Metadata(ctx, key)
}

func (s *objSpy) SignedGetURL(ctx context.Context, key objectstore.Key, ttl time.Duration, ct string) (string, error) {
	if err := s.hook("sign", key); err != nil {
		return "", err
	}
	return s.inner.SignedGetURL(ctx, key, ttl, ct)
}

type objEnv struct {
	*itEnv
	scfg  objectstore.Config
	ocfg  ObjectConfig
	store *objectstore.S3 // the scoped credential: fixtures, assertions of provider state
	spy   *objSpy
	osvc  *Service
}

// objectSettings reads the storage settings exactly as the server does, through config.Load: the object settings
// come from the environment, the unrelated required settings are fixed placeholders.
func objectSettings(t *testing.T) (objectstore.Config, ObjectConfig) {
	t.Helper()
	fixed := map[string]string{"PRODUCT_NAME": "Objects Probe", "PLATFORM_BASE_DOMAIN": "objects.example.test",
		"CUSTOM_DOMAIN_CNAME_TARGET": "custom.objects.example.test", "DATABASE_NAME": "objects_probe", "DB_HOST": "postgres",
		"DB_PORT": "5432", "DB_USER": "objects_probe_runtime", "DB_PASSWORD": "unused", "REDIS_HOST": "redis", "REDIS_PORT": "6379",
		"REDIS_PASSWORD": "unused"}
	get := func(k string) string {
		if strings.HasPrefix(k, "OBJECT_STORE_") || k == "APP_ENV" {
			return os.Getenv(k)
		}
		return fixed[k]
	}
	c, err := config.Load(get, config.EnvSecrets{Getenv: get})
	if err != nil {
		t.Fatalf("object store settings: %v", err)
	}
	o := c.ObjectStore
	return objectstore.Config{Endpoint: o.Endpoint, SigningEndpoint: o.SigningEndpoint, Region: o.Region, Bucket: o.Bucket,
			ForcePathStyle: o.ForcePathStyle, AccessKey: o.AccessKey, SecretKey: o.SecretKey},
		ObjectConfig{MaxObjectBytes: o.MaxObjectBytes, DefaultTTL: o.SignedURLDefaultTTL, MaxTTL: o.SignedURLMaxTTL, TransferTimeout: o.TransferTimeout}
}

func setupObjects(t *testing.T) *objEnv {
	t.Helper()
	e := &objEnv{itEnv: setupIT(t)}
	e.scfg, e.ocfg = objectSettings(t)
	var err error
	if e.store, err = objectstore.New(e.scfg, os.Getenv("APP_ENV")); err != nil {
		t.Fatalf("object store: %v", err)
	}
	e.spy = &objSpy{inner: e.store}
	e.osvc = NewServiceWithObjects(e.hook, e.id, e.svc.logger, e.spy, e.ocfg)
	if e.osvc.objects == nil {
		t.Fatal("NewServiceWithObjects refused the loaded configuration")
	}
	return e
}

// objService builds another storage service over db and store (its own hooks), sharing the identity and the log.
func (e *objEnv) objService(db tenancy.DB, store objectstore.Store, cfg ObjectConfig) *Service {
	return NewServiceWithObjects(db, e.id, e.svc.logger, store, cfg)
}

// --- object fixtures -----------------------------------------------------------------------------------------------------

// objFixture is one seeded object: its identity and scope, its state and its bytes.
type objFixture struct {
	id      tenancy.ObjectID
	org     tenancy.OrganizationID
	group   tenancy.GroupID
	creator tenancy.UserID
	state   tenancy.ObjectState
	ct      string
	data    []byte
}

func (o objFixture) key(t *testing.T) objectstore.Key {
	t.Helper()
	k, err := objectstore.NewKey(o.org.String(), o.id.String())
	if err != nil {
		t.Fatal(err)
	}
	return k
}

func (o objFixture) binding() objectstore.Binding {
	b := objectstore.Binding{Object: o.id.String(), Organization: o.org.String()}
	if !zero(o.group) {
		b.Group = o.group.String()
	}
	return b
}

func nullUUID[T ~[16]byte](id T) *string {
	if zero(id) {
		return nil
	}
	s := formatUUIDString(id)
	return &s
}

func formatUUIDString[T ~[16]byte](id T) string { return tenancy.ObjectID(id).String() }

// seed inserts the object's row as the attested superuser fixture (row-level security and the UPDATE trigger do not
// apply to a superuser INSERT) and uploads its bytes with the scoped credential when the state carries bytes.
func (e *objEnv) seed(t *testing.T, o objFixture) {
	t.Helper()
	ctx := context.Background()
	var sha []byte
	if o.state != tenancy.ObjectPending {
		s := sha256.Sum256(o.data)
		sha = s[:]
	}
	var readyAt, deletingAt, deletedAt, deletedBy any
	now := time.Now().UTC()
	switch o.state {
	case tenancy.ObjectReady:
		readyAt = now
	case tenancy.ObjectDeleting:
		deletingAt, deletedBy = now, o.creator.String()
	case tenancy.ObjectDeleted:
		deletingAt, deletedAt, deletedBy = now, now, o.creator.String()
	}
	if _, err := e.setup.Exec(ctx, `INSERT INTO app.stored_objects (id, organization_id, group_id, creator_user_id, state, byte_size,
  content_type, sha256, ready_at, deleting_at, deleted_at, deleted_by)
VALUES ($1::text::uuid, $2::text::uuid, $3::text::uuid, $4::text::uuid, $5, $6, $7, $8, $9, $10, $11, $12::text::uuid)`,
		o.id.String(), o.org.String(), nullUUID(o.group), o.creator.String(), string(o.state), int64(len(o.data)), o.ct, sha,
		readyAt, deletingAt, deletedAt, deletedBy); err != nil {
		t.Fatalf("seed object: %v", err)
	}
	if o.state == tenancy.ObjectReady || o.state == tenancy.ObjectDeleting {
		if err := e.store.Put(ctx, o.key(t), bytes.NewReader(o.data), int64(len(o.data)), o.ct, o.binding()); err != nil {
			t.Fatalf("seed bytes: %v", err)
		}
	}
}

// unseed removes the object's row (superuser fixture) and its key (scoped credential).
func (e *objEnv) unseed(t *testing.T, org tenancy.OrganizationID, id tenancy.ObjectID) {
	t.Helper()
	ctx := context.Background()
	if _, err := e.setup.Exec(ctx, `DELETE FROM app.stored_objects WHERE id = $1::text::uuid`, id.String()); err != nil {
		t.Fatalf("unseed row: %v", err)
	}
	k, err := objectstore.NewKey(org.String(), id.String())
	if err != nil {
		t.Fatal(err)
	}
	if err := e.store.Delete(ctx, k); err != nil {
		t.Fatalf("unseed key: %v", err)
	}
}

// present reports whether the provider holds the key.
func (e *objEnv) present(t *testing.T, k objectstore.Key) bool {
	t.Helper()
	_, err := e.store.Metadata(context.Background(), k)
	switch {
	case err == nil:
		return true
	case errors.Is(err, objectstore.ErrNotFound):
		return false
	}
	t.Fatalf("head: %v", err)
	return false
}

// rowState returns the stored row of an object as text ("-" when absent), read by the superuser fixture.
func (e *objEnv) rowState(t *testing.T, id tenancy.ObjectID) string {
	t.Helper()
	var s string
	err := e.setup.QueryRow(context.Background(), `SELECT coalesce((SELECT o::text FROM app.stored_objects o WHERE o.id = $1::text::uuid), '-')`,
		id.String()).Scan(&s)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

// objectRows hashes every stored-object row of the given Organizations.
func (e *objEnv) objectRows(t *testing.T, orgs ...tenancy.OrganizationID) string {
	t.Helper()
	ids := make([]string, len(orgs))
	for i, o := range orgs {
		ids[i] = o.String()
	}
	var s string
	if err := e.setup.QueryRow(context.Background(), `SELECT count(*) || ':' || coalesce(md5(string_agg(o::text, E'\n' ORDER BY o.id)), '-')
FROM app.stored_objects o WHERE o.organization_id::text = ANY ($1::text[])`, ids).Scan(&s); err != nil {
		t.Fatal(err)
	}
	return s
}

func payload(n int, seed string) []byte {
	out := make([]byte, n)
	h := sha256.Sum256([]byte(seed))
	for i := range out {
		if i%32 == 0 && i > 0 {
			h = sha256.Sum256(h[:])
		}
		out[i] = h[i%32]
	}
	return out
}

// --- the oracle world ----------------------------------------------------------------------------------------------------

// objWorld instantiates the independent oracle's fixture: real Organizations, Groups, profiles and users (signed in
// through the real G004 flow), their current bindings, and the objects (rows and bytes) under the oracle's object
// UUIDs. The oracle's Organization, Group and user UUIDs are mapped to the real ones.
type objWorld struct {
	e      *objEnv
	fx     oeFixture
	orgs   map[string]tenancy.OrganizationID
	groups map[string]tenancy.GroupID
	users  map[string]*itUser
	objs   map[string]objFixture
	missID tenancy.ObjectID
}

func (e *objEnv) objWorld(t *testing.T) *objWorld {
	t.Helper()
	fx := objectsExpectedFixture()
	w := &objWorld{e: e, fx: fx, orgs: map[string]tenancy.OrganizationID{}, groups: map[string]tenancy.GroupID{},
		users: map[string]*itUser{}, objs: map[string]objFixture{}}
	c := e.canaries
	for i, o := range fx.Organizations {
		w.orgs[o.Key] = e.org(t, "Objects "+o.Key+" "+c[i%len(c)])
	}
	creator := e.newUser(t) // creates the Groups (and so is their group_owner); not an oracle actor
	for _, g := range fx.Groups {
		w.groups[g.Key] = e.group(t, w.orgs[g.OrganizationKey], creator, "Objects "+g.Key, "objects-"+strings.ToLower(g.Key)).ID
	}
	for _, p := range fx.Profiles {
		bp := e.profile(t, w.orgs[p.OrganizationKey], "Objects "+p.Key)
		for _, g := range p.GroupKeys {
			e.link(t, w.orgs[p.OrganizationKey], w.groups[g], bp.ID)
		}
	}
	for _, u := range fx.Users {
		iu := e.newUser(t)
		w.users[u.Key] = iu
		for _, b := range u.Bindings {
			w.grant(t, iu, b)
		}
	}
	// G016 A2: one prepared credential per (actor, Organization) pair the service matrix resolves with a standing
	// credential (active actors with a live session, plus the subject-mismatch fault's other subject), and the pairs of
	// the direct subtests below. A membership-negative actor is admitted only for the issuance (see bind). Actors whose
	// session or user state the case varies sign in afresh in prepare, as before.
	bound := map[[2]string]bool{}
	bindPair := func(user, org string) {
		if k := [2]string{user, org}; !bound[k] {
			bound[k] = true
			e.bind(t, w.users[user], w.orgs[org])
		}
	}
	for _, c := range objectsExpectedCases() {
		if c.DBRole != oeDBNone || (c.Ctx.Kind != oeCtxOrganization && c.Ctx.Kind != oeCtxGroup) {
			continue
		}
		if u := w.user(c.Actor); u.State == oeUserActive && u.Session == oeSessionLive {
			bindPair(c.Actor, c.Ctx.OrganizationKey)
		}
		if c.Ctx.Fault == oeFaultSubjectMismatch {
			bindPair(mismatchOther(c.Actor), c.Ctx.OrganizationKey)
		}
	}
	for _, u := range []string{"aOwner", "aAdmin", "g1MemberX"} {
		bindPair(u, "orgA")
	}
	for _, o := range fx.Objects {
		id := mustParse[tenancy.ObjectID](o.ID)
		if !o.Stored {
			w.missID = id
			continue
		}
		f := objFixture{id: id, org: w.orgs[o.OrganizationKey], creator: w.users[o.CreatorKey].id, ct: "application/octet-stream",
			data: payload(1024+len(o.Key), o.Key+c[0])}
		if o.GroupKey != "" {
			f.group = w.groups[o.GroupKey]
		}
		switch o.State {
		case oeStatePending:
			f.state = tenancy.ObjectPending
		case oeStateReady:
			f.state = tenancy.ObjectReady
		case oeStateDeleting:
			f.state = tenancy.ObjectDeleting
		}
		w.objs[o.Key] = f
		e.seed(t, f)
	}
	t.Cleanup(func() {
		for _, f := range w.objs {
			e.unseed(t, f.org, f.id)
		}
	})
	return w
}

func (w *objWorld) grant(t *testing.T, u *itUser, b oeBinding) {
	t.Helper()
	if b.Scope == oeScopeOrganization {
		w.e.grantOrg(t, w.orgs[b.ScopeKey], u, tenancy.OrganizationRole(b.Role))
		return
	}
	g := w.groupFixture(b.ScopeKey)
	w.e.grantGroup(t, w.orgs[g.OrganizationKey], w.groups[b.ScopeKey], u, tenancy.GroupRole(b.Role))
}

func (w *objWorld) revoke(t *testing.T, u *itUser, b oeBinding) {
	t.Helper()
	ctx := context.Background()
	if b.Scope == oeScopeOrganization {
		if err := w.e.model.RemoveOrganizationMembership(ctx, w.orgs[b.ScopeKey], u.id); err != nil {
			t.Fatalf("remove membership: %v", err)
		}
		return
	}
	g := w.groupFixture(b.ScopeKey)
	if err := w.e.model.RemoveGroupMembership(ctx, w.orgs[g.OrganizationKey], w.groups[b.ScopeKey], u.id); err != nil {
		t.Fatalf("remove group membership: %v", err)
	}
}

// mismatchOther is the other subject whose context the subject-mismatch fault (B5/B6) substitutes for actor's.
func mismatchOther(actor string) string {
	if actor == "g1MemberY" {
		return "g1MemberX"
	}
	return "g1MemberY"
}

func (w *objWorld) groupFixture(key string) oeGroup {
	for _, g := range w.fx.Groups {
		if g.Key == key {
			return g
		}
	}
	panic("unknown oracle Group " + key)
}

func (w *objWorld) user(key string) oeUser {
	for _, u := range w.fx.Users {
		if u.Key == key {
			return u
		}
	}
	panic("unknown oracle user " + key)
}

// prepared is a case's request state: the request context, the TenantContext (nil when none is given), an outcome
// already decided while establishing the context (Resolve's), and the cleanup.
type prepared struct {
	rctx    context.Context
	tc      *TenantContext
	decided Code
	cleanup []func()
}

func (p *prepared) done() {
	for i := len(p.cleanup) - 1; i >= 0; i-- {
		p.cleanup[i]()
	}
}

// prepare establishes the case's request and context through the real services: a fresh sign-in for actors whose
// session or user state is varied, the stale basis (the former bindings) granted for Resolve and revoked before the
// operation, the binding-check faults B1-B6, and the session revocation, expiry or user disablement after Resolve.
func (w *objWorld) prepare(t *testing.T, svc *Service, c oeCase) *prepared {
	t.Helper()
	e := w.e
	ctx := context.Background()
	u := w.user(c.Actor)
	iu := w.users[c.Actor]
	// G016 A2: a case that resolves an Organization runs in a request scope of the actor's credential for that
	// Organization, at its origin: the prepared one (objWorld) or, for a varied session or user state, a fresh sign-in
	// there (as relogin before). A case without a context keeps the sign-in fixture credential (it resolves nothing).
	live := u.State == oeUserActive && u.Session == oeSessionLive
	cred, origin := iu.cred, itOrigin
	var org tenancy.OrganizationID
	if c.Ctx.Kind == oeCtxOrganization || c.Ctx.Kind == oeCtxGroup {
		org = w.orgs[c.Ctx.OrganizationKey]
		origin = itOriginOf(org)
		if live {
			cred = iu.in(t, org)
		} else {
			cred = e.signInAt(t, e.id, org, iu.sub, "Again", nil, "", "")
		}
	} else if !live {
		cred = e.relogin(t, iu)
	}
	p := &prepared{}
	rctx, end, err := svc.BeginRequest(ctx, cred, origin)
	if err != nil {
		t.Fatalf("%s: BeginRequest: %v", c.ID, err)
	}
	p.cleanup = append(p.cleanup, end)
	p.rctx = rctx
	if c.Ctx.Kind == oeCtxNone {
		return p
	}
	if c.Ctx.Stale {
		for _, b := range u.FormerBindings {
			w.grant(t, iu, b)
		}
	}
	resolveIn := func(svc *Service, rctx context.Context) (*TenantContext, error) {
		req := ResolveRequest{OrganizationID: w.orgs[c.Ctx.OrganizationKey].String()}
		if c.Ctx.GroupKey != "" {
			req.GroupID = w.groups[c.Ctx.GroupKey].String()
		}
		return svc.Resolve(rctx, req)
	}
	tc, err := resolveIn(svc, rctx)
	if c.Ctx.Stale {
		for _, b := range u.FormerBindings {
			w.revoke(t, iu, b)
		}
	}
	if err != nil {
		p.decided = CodeOf(err)
		return p
	}
	p.tc = tc
	switch c.Ctx.Fault {
	case oeFaultForged:
		p.tc = &TenantContext{}
	case oeFaultForeignService:
		other, _ := e.service(e.id)
		orctx, oend, err := other.BeginRequest(ctx, cred, origin)
		if err != nil {
			t.Fatal(err)
		}
		p.cleanup = append(p.cleanup, oend)
		if p.tc, err = resolveIn(other, orctx); err != nil {
			t.Fatal(err)
		}
	case oeFaultScopeMissing:
		p.rctx = context.Background()
	case oeFaultScopeEnded:
		end()
	case oeFaultScopeMismatch:
		orctx, oend, err := svc.BeginRequest(ctx, cred, origin)
		if err != nil {
			t.Fatal(err)
		}
		p.cleanup = append(p.cleanup, oend)
		if p.tc, err = resolveIn(svc, orctx); err != nil {
			t.Fatal(err)
		}
	case oeFaultSubjectMismatch:
		other := w.users[mismatchOther(c.Actor)]
		orctx, oend, err := svc.BeginRequest(ctx, other.in(t, org), origin) // G016 A2: the other subject's own scope for org
		if err != nil {
			t.Fatal(err)
		}
		p.cleanup = append(p.cleanup, oend)
		if p.tc, err = resolveIn(svc, orctx); err != nil {
			t.Fatal(err)
		}
	}
	switch {
	case u.Session == oeSessionRevoked:
		e.logout(t, cred, "current")
	case u.Session == oeSessionExpired:
		if _, err := e.setup.Exec(ctx, `UPDATE app.sessions SET idle_expires_at = now() - interval '1 minute' WHERE id = $1::text::uuid`,
			scopeOf(rctx).session); err != nil {
			t.Fatal(err)
		}
	case u.State == oeUserDisabled:
		if err := e.id.DisableUser(ctx, iu.id.String()); err != nil {
			t.Fatal(err)
		}
		p.cleanup = append(p.cleanup, func() {
			if err := e.id.EnableUser(ctx, iu.id.String()); err != nil {
				t.Errorf("enable: %v", err)
			}
		})
	}
	return p
}

// providerState is the provider presence of every seeded key plus the stored-object rows of both Organizations.
func (w *objWorld) state(t *testing.T) string {
	t.Helper()
	var b strings.Builder
	keys := make([]string, 0, len(w.objs))
	for k := range w.objs {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	for _, k := range keys {
		fmt.Fprintf(&b, "%s=%v;", k, w.e.present(t, w.objs[k].key(t)))
	}
	b.WriteString(w.e.objectRows(t, w.orgs["orgA"], w.orgs["orgB"]))
	return b.String()
}

// objectSelector maps a case's object key or raw selector onto the API's typed selector; a raw string the API
// boundary's parser refuses is invalid_input before the service is called (only canonical UUIDs reach it). The all-zero
// UUID is passed as the zero selector, so the service's own refusal is exercised.
func (w *objWorld) objectSelector(c oeCase) (tenancy.ObjectID, bool) {
	if c.Object != "" {
		if f, ok := w.objs[c.Object]; ok {
			return f.id, true
		}
		return w.missID, true
	}
	if c.RawObjectID == zeroUUID {
		return tenancy.ObjectID{}, true
	}
	id, err := tenancy.ParseObjectID(c.RawObjectID)
	return id, err == nil
}

// runServiceCase runs one oracle service case and returns the observed outcome.
func (w *objWorld) runServiceCase(t *testing.T, c oeCase) oeOutcome {
	t.Helper()
	e := w.e
	p := w.prepare(t, e.osvc, c)
	defer p.done()
	if p.decided != "" {
		return oeOutcome(p.decided)
	}
	before := w.state(t)
	calls := e.spy.calls.Load()
	var err error
	disclosed := false
	switch c.Action {
	case oeActMetadata, oeActGet, oeActSign:
		id, ok := w.objectSelector(c)
		if !ok {
			return oeInvalidInput
		}
		switch c.Action {
		case oeActMetadata:
			var info ObjectInfo
			info, err = e.osvc.ObjectMetadata(p.rctx, p.tc, id)
			disclosed = info != (ObjectInfo{})
		case oeActGet:
			var info ObjectInfo
			var rc io.ReadCloser
			info, rc, err = e.osvc.GetObject(p.rctx, p.tc, id)
			disclosed = info != (ObjectInfo{}) || rc != nil
			if rc != nil {
				got, rerr := io.ReadAll(rc)
				_ = rc.Close()
				if f := w.byID(id); err == nil && (rerr != nil || !bytes.Equal(got, f.data)) {
					t.Errorf("%s: get returned other bytes (%v)", c.ID, rerr)
				}
			}
		case oeActSign:
			req := SignedGetURLRequest{Object: id, DefaultTTL: !c.TTL.Set, TTLSeconds: c.TTL.Seconds}
			var u SignedURL
			u, err = e.osvc.SignedGetURL(p.rctx, p.tc, req)
			disclosed = u.URL() != ""
			if err == nil && int64(u.TTL/time.Second) != c.ExpectTTLSeconds {
				t.Errorf("%s: issued TTL %v, oracle %d s", c.ID, u.TTL, c.ExpectTTLSeconds)
			}
		}
	case oeActPut:
		var group tenancy.GroupID
		switch {
		case c.PutGroupKey != "":
			group = w.groups[c.PutGroupKey]
		case c.RawGroupID == zeroUUID:
			return w.putZeroGroup(t, p, c)
		case c.RawGroupID != "":
			g, perr := tenancy.ParseGroupID(c.RawGroupID)
			if perr != nil {
				return oeInvalidInput
			}
			group = g
		}
		data := payload(333, c.ID)
		var info ObjectInfo
		info, err = e.osvc.PutObject(p.rctx, p.tc, PutObjectRequest{Group: group, Size: int64(len(data)), Body: io.NopCloser(bytes.NewReader(data))})
		if err == nil {
			defer e.unseed(t, info.Organization, info.ID)
			if info.Group != group || info.Size != int64(len(data)) || info.SHA256 != sha256.Sum256(data) {
				t.Errorf("%s: put info %+v", c.ID, info)
			}
		}
	case oeActDelete:
		id, ok := w.objectSelector(c)
		if !ok {
			return oeInvalidInput
		}
		var res DeleteResult
		res, err = e.osvc.DeleteObject(p.rctx, p.tc, id)
		if err == nil {
			f := w.byID(id)
			defer func() { e.unseed(t, f.org, f.id); e.seed(t, f) }()
			if res.Outcome != DeleteOutcomeDeleted || e.present(t, f.key(t)) || !strings.Contains(e.rowState(t, f.id), ",deleted,") {
				t.Errorf("%s: delete %s; provider present %v; row %s", c.ID, res.Outcome, e.present(t, f.key(t)), itH(e.rowState(t, f.id)))
			}
		}
	default:
		t.Fatalf("%s: not a service case", c.ID)
	}
	if err == nil {
		return oeAllow
	}
	if n := e.spy.calls.Load() - calls; n != 0 {
		t.Errorf("%s: denial made %d provider calls", c.ID, n)
	}
	if after := w.state(t); after != before {
		t.Errorf("%s: denial changed the database or provider state", c.ID)
	}
	if disclosed {
		t.Errorf("%s: denial disclosed the object", c.ID)
	}
	return oeOutcome(CodeOf(err))
}

// putZeroGroup: the all-zero Group selector. The typed API has no Group selector string: a zero GroupID means an
// Organization object, so the canonical refusal of an all-zero selector is the API boundary's parser (invalid_id).
func (w *objWorld) putZeroGroup(t *testing.T, _ *prepared, _ oeCase) oeOutcome {
	t.Helper()
	if _, err := tenancy.ParseGroupID(zeroUUID); tenancy.CodeOf(err) != tenancy.CodeInvalidID {
		t.Fatalf("the all-zero Group selector parses: %v", err)
	}
	return oeInvalidInput
}

func (w *objWorld) byID(id tenancy.ObjectID) objFixture {
	for _, f := range w.objs {
		if f.id == id {
			return f
		}
	}
	return objFixture{}
}

// --- RLS cases -------------------------------------------------------------------------------------------------------------

// rlsConn returns the case's database login: the attested runtime login or the owner (the migrator in the owner role,
// under FORCE).
func (e *objEnv) rlsOwner(t *testing.T) *pgx.Conn {
	t.Helper()
	c := itConnect(t, e.cfg, e.cfg.MigratorRole(), e.cfg.Password)
	if _, err := c.Exec(context.Background(), "SET ROLE "+pgx.Identifier{e.cfg.OwnerRole()}.Sanitize()); err != nil {
		t.Fatal(err)
	}
	var cur string
	if err := c.QueryRow(context.Background(), `SELECT current_user::text`).Scan(&cur); err != nil || cur != e.cfg.OwnerRole() {
		t.Fatalf("owner connection acts as %s (%v)", cur, err)
	}
	return c
}

func (w *objWorld) rlsSettings(c oeCase) (user, org, group, binding *string) {
	switch c.Ctx.Kind {
	case oeCtxNone:
		return nil, nil, nil, nil
	case oeCtxRaw:
		r := c.Ctx.Raw
		opt := func(s string) *string {
			if s == "" {
				return nil
			}
			return &s
		}
		return opt(r.UserID), opt(r.OrganizationID), opt(r.GroupID), opt(r.LoginBinding)
	}
	user, org = ptr(w.users[c.Actor].id.String()), ptr(w.orgs[c.Ctx.OrganizationKey].String())
	if c.Ctx.GroupKey != "" {
		group = ptr(w.groups[c.Ctx.GroupKey].String())
	}
	return user, org, group, nil
}

// runRLSCase runs one oracle RLS case in a rolled-back transaction and reports the observed outcome (oeAllow, an RLS
// denial outcome, or a SQLSTATE outcome), with the mechanism for the record.
func (w *objWorld) runRLSCase(t *testing.T, conn *pgx.Conn, c oeCase) (oeOutcome, string) {
	t.Helper()
	ctx := context.Background()
	tx, err := conn.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	user, org, group, binding := w.rlsSettings(c)
	if err := rlsSet(ctx, tx, user, org, group, binding); err != nil {
		t.Fatal(err)
	}
	keyOf := map[string]string{}
	for k, f := range w.objs {
		keyOf[f.id.String()] = k
	}
	switch c.Action {
	case oeActRLSSelect:
		rows, err := tx.Query(ctx, `SELECT id::text FROM app.stored_objects`)
		var seen []string
		if err == nil {
			for rows.Next() {
				var id string
				if err = rows.Scan(&id); err != nil {
					break
				}
				if k, ok := keyOf[id]; ok {
					seen = append(seen, k)
				}
			}
			rows.Close()
			if err == nil {
				err = rows.Err()
			}
		}
		if s := pgState(err); s != "" {
			return oeOutcome("sqlstate_" + s), s
		}
		if err != nil {
			t.Fatalf("%s: %v", c.ID, err)
		}
		for _, k := range c.MustNotSee {
			if slices.Contains(seen, k) {
				return oeOutcome("visible:" + k), "select"
			}
		}
		for _, k := range c.MustSee {
			if !slices.Contains(seen, k) {
				return oeOutcome("invisible:" + k), "select"
			}
		}
		if c.Expect == oeRLSNoRows {
			if len(seen) == 0 {
				return oeRLSNoRows, "no rows"
			}
			return oeOutcome("rows"), "select"
		}
		return oeAllow, "select"
	case oeActRLSInsert:
		cols, vals := "id, organization_id, group_id, creator_user_id, byte_size, content_type", "gen_random_uuid(), $1::text::uuid, $2::text::uuid, $3::text::uuid, 0, 'application/octet-stream'"
		args := []any{w.orgs[c.Row.OrganizationKey].String(), nullUUID(w.groupOrZero(c.Row.GroupKey)), w.users[c.Row.CreatorKey].id.String()}
		if c.Row.State != oeStatePending {
			cols += ", state, sha256, ready_at, deleting_at, deleted_by"
			vals += ", $4, decode(repeat('ab', 32), 'hex'), now(), CASE WHEN $4 = 'deleting' THEN now() END, CASE WHEN $4 = 'deleting' THEN $3::text::uuid END"
			args = append(args, string(c.Row.State))
		}
		st, n := rlsStmt(t, tx, `INSERT INTO app.stored_objects (`+cols+`) VALUES (`+vals+`)`, args...)
		switch {
		case st == "" && n == 1:
			return oeAllow, "inserted"
		case st == "42501" && c.Row.State == oeStatePending && c.Expect == oeRLSWithCheck:
			return oeRLSWithCheck, "42501"
		case st != "" && c.Row.State != oeStatePending:
			return oeRLSStateRejected, st
		}
		return oeOutcome("insert:" + st), st
	case oeActRLSTransition, oeActRLSImmutable:
		f := w.objs[c.Object]
		before := w.e.rowState(t, f.id)
		var sql string
		var args []any
		if c.Action == oeActRLSTransition {
			sql, args = `UPDATE app.stored_objects SET state = $2 WHERE id = $1::text::uuid`, []any{f.id.String(), string(c.ToState)}
			if c.ToState == oeStateReady {
				sql = `UPDATE app.stored_objects SET state = $2, sha256 = decode(repeat('cd', 32), 'hex') WHERE id = $1::text::uuid`
			}
		} else {
			col, val := w.immutable(c)
			sql, args = `UPDATE app.stored_objects SET `+col+` = $2 WHERE id = $1::text::uuid`, []any{f.id.String(), val}
		}
		st, n := rlsStmt(t, tx, sql, args...)
		if w.e.rowState(t, f.id) != before {
			t.Errorf("%s: the row changed outside the case's rolled-back savepoint", c.ID)
		}
		switch {
		case st == "" && n == 1:
			return oeAllow, "updated"
		case st == "" && n == 0 && c.Expect == oeRLSZeroRows:
			return oeRLSZeroRows, "0 rows"
		case c.Expect == oeRLSDeniedNoChange && ((st == "" && n == 0) || st == "42501" || st == "23000"):
			return oeRLSDeniedNoChange, fmt.Sprintf("%s/%d", st, n)
		case c.Expect == oeRLSImmutable && (st == "42501" || st == "23000"):
			return oeRLSImmutable, st
		case c.Expect == oeRLSImmutableGenerated && c.Column == "object_key" && (st == "428C9" || st == "42501" || st == "23000"):
			// Codex ruling a466be31 (oracle r2): only the generated storage key's non-DEFAULT assignment may be 428C9.
			return oeRLSImmutableGenerated, st
		}
		return oeOutcome(fmt.Sprintf("update:%s/%d", st, n)), st
	case oeActRLSDelete:
		st, _ := rlsStmt(t, tx, `DELETE FROM app.stored_objects WHERE id = $1::text::uuid`, w.objs[c.Object].id.String())
		if st == "42501" {
			return oeSQLPrivilege, st
		}
		return oeOutcome("delete:" + st), st
	}
	t.Fatalf("%s: not an RLS case", c.ID)
	return "", ""
}

// guardTransitions (S-review finding 1; 00007's creator-only finalize and same-state refusal): updates the lifecycle
// scope of row-level security admits but the guard must refuse with stored_objects_transition, each leaving the row
// unchanged, beside positive controls on the same rows. Every case runs in a rolled-back runtime transaction; prep is an
// allowed transition run first in that transaction. It returns the number of refusals and controls checked.
// G060 (Plan r0 A10): the Organization admin's two former guard cases are now refused by row-level security first (a
// Group row is outside its USING under either setting): the UPDATE affects no row and the stored row is unchanged
// (hidden; counted apart from the guard refusals).
func (w *objWorld) guardTransitions(t *testing.T, rt *pgx.Conn) (refused, controls, hidden int) {
	t.Helper()
	ctx := context.Background()
	const (
		finalize = `UPDATE app.stored_objects SET state = 'ready', sha256 = decode(repeat('cd', 32), 'hex') WHERE id = $1::text::uuid`
		rewrite  = `UPDATE app.stored_objects SET sha256 = decode(repeat('ef', 32), 'hex') WHERE id = $1::text::uuid`
		reready  = `UPDATE app.stored_objects SET state = 'ready' WHERE id = $1::text::uuid`
		tomb     = `UPDATE app.stored_objects SET state = 'deleted' WHERE id = $1::text::uuid`
		undelete = `UPDATE app.stored_objects SET state = 'deleting' WHERE id = $1::text::uuid`
	)
	for _, c := range []struct {
		name, actor, group, object, prep, sql string
		allow                                 bool
	}{
		{"the creator finalizes its pending row (control)", "g1MemberX", "g1", "oG1Pending", "", finalize, true},
		{"the g1 admin finalizes g1MemberX's pending row", "g1Admin", "g1", "oG1Pending", "", finalize, false},
		{"the Organization admin finalizes g1MemberX's pending row", "aAdmin", "", "oG1Pending", "", finalize, false},
		{"the creator rewrites a ready row's sha256", "g1MemberX", "g1", "oG1X", "", rewrite, false},
		{"the creator updates a ready row to ready", "g1MemberX", "g1", "oG1X", "", reready, false},
		{"the g1 admin tombstones a deleting row (control)", "g1Admin", "g1", "oG1Deleting", "", tomb, true},
		{"the g1 admin updates a deleted row to deleted", "g1Admin", "g1", "oG1Deleting", tomb, tomb, false},
		{"the g1 admin updates a deleted row to deleting", "g1Admin", "g1", "oG1Deleting", tomb, undelete, false},
		{"the Organization admin rewrites a deleted row's sha256", "aAdmin", "", "oG1Deleting", tomb, rewrite, false},
	} {
		id := w.objs[c.object].id.String()
		before := w.e.rowState(t, w.objs[c.object].id)
		tx, err := rt.Begin(ctx)
		if err != nil {
			t.Fatal(err)
		}
		var group *string
		if c.group != "" {
			group = ptr(w.groups[c.group].String())
		}
		if err := rlsSet(ctx, tx, ptr(w.users[c.actor].id.String()), ptr(w.orgs["orgA"].String()), group, nil); err != nil {
			t.Fatal(err)
		}
		if c.actor == "aAdmin" { // G060: the Organization admin reaches no Group row (see above)
			tag, err := tx.Exec(ctx, c.sql, id)
			if err != nil || tag.RowsAffected() != 0 {
				t.Errorf("guard %s: %v (%d rows), want no row reached (row-level security)", c.name, err, tag.RowsAffected())
			} else {
				hidden++
			}
			_ = tx.Rollback(ctx)
			if w.e.rowState(t, w.objs[c.object].id) != before {
				t.Errorf("guard %s: the stored row changed", c.name)
			}
			continue
		}
		if c.prep != "" {
			if tag, err := tx.Exec(ctx, c.prep, id); err != nil || tag.RowsAffected() != 1 {
				t.Fatalf("guard %s: prep: %v (%d rows)", c.name, err, tag.RowsAffected())
			}
		}
		row := func() string {
			var s string
			if err := tx.QueryRow(ctx, `SELECT o::text FROM app.stored_objects o WHERE id = $1::text::uuid`, id).Scan(&s); err != nil {
				t.Fatalf("guard %s: row: %v", c.name, err)
			}
			return s
		}
		was := row()
		sp, err := tx.Begin(ctx)
		if err != nil {
			t.Fatal(err)
		}
		tag, err := sp.Exec(ctx, c.sql, id)
		_ = sp.Rollback(ctx)
		var pg *pgconn.PgError
		switch {
		case c.allow:
			controls++
			if err != nil || tag.RowsAffected() != 1 {
				t.Errorf("guard %s: %v (%d rows), want the update admitted", c.name, err, tag.RowsAffected())
			}
		case !errors.As(err, &pg) || pg.Code != "23000" || pg.ConstraintName != "stored_objects_transition":
			t.Errorf("guard %s: %v, want 23000 stored_objects_transition", c.name, err)
		default:
			refused++
			if row() != was {
				t.Errorf("guard %s: the row changed after the refusal", c.name)
			}
		}
		_ = tx.Rollback(ctx)
		if w.e.rowState(t, w.objs[c.object].id) != before {
			t.Errorf("guard %s: the stored row changed", c.name)
		}
	}
	return refused, controls, hidden
}

func (w *objWorld) groupOrZero(key string) tenancy.GroupID {
	if key == "" {
		return tenancy.GroupID{}
	}
	return w.groups[key]
}

// immutable maps the oracle's logical column and new value onto the stored column and a value.
func (w *objWorld) immutable(c oeCase) (string, any) {
	h := sha256.Sum256([]byte(c.ID))
	fresh := "11111111-2222-4333-8444-" + hex.EncodeToString(h[:6])
	switch c.Column {
	case "id":
		return "id", fresh
	case "organization_id":
		return "organization_id", w.orgs[c.NewValue].String()
	case "group_id":
		return "group_id", w.groups[c.NewValue].String()
	case "creator_id":
		return "creator_user_id", w.users[c.NewValue].id.String()
	case "object_key":
		return "storage_key", "objects/" + fresh + "/" + fresh
	}
	panic("unknown oracle column " + c.Column)
}

// --- TestObjects -----------------------------------------------------------------------------------------------------------

func TestObjects(t *testing.T) {
	e := setupObjects(t)
	ctx := context.Background()

	t.Run("OBJ-ATTEST runtime login and scoped credential", func(t *testing.T) {
		conn, err := e.pool.Acquire(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if err := e.rlsAttest(conn); err != nil {
			t.Fatal(err)
		}
		conn.Release()
		if err := e.rlsAttest(e.setup); err == nil {
			t.Fatal("the privileged setup connection passed the runtime attestation (the privileged-assertion negative)")
		}
		for _, k := range []string{"MINIO_ROOT_USER", "MINIO_ROOT_PASSWORD"} {
			if os.Getenv(k) != "" {
				t.Fatalf("%s is present in the test environment", k)
			}
		}
		if e.scfg.AccessKey == "" || e.scfg.SecretKey == "" {
			t.Fatal("no scoped credential")
		}
	})

	w := e.objWorld(t)

	t.Run("OBJ-READ and OBJ-WRITE: the oracle service matrix", func(t *testing.T) {
		n, dis := 0, 0
		for _, c := range objectsExpectedCases() {
			if c.DBRole != oeDBNone {
				continue
			}
			n++
			if got := w.runServiceCase(t, c); got != c.Expect {
				dis++
				t.Errorf("ORACLE-DISAGREEMENT %s (%s %s): observed %s, oracle %s (%s)", c.ID, c.Action, c.Actor, got, c.Expect, c.Why)
			}
		}
		fmt.Printf("OBJ-ORACLE-SERVICE {\"cases\":%d,\"disagreements\":%d}\n", n, dis)
	})

	t.Run("OBJ-RLS: the oracle RLS matrix, guard transitions, FORCE owner, poisoned context", func(t *testing.T) {
		rt, ow := e.rlsRuntime(t), e.rlsOwner(t)
		n, dis := 0, 0
		for _, c := range objectsExpectedCases() {
			var conn *pgx.Conn
			switch c.DBRole {
			case oeDBRuntime:
				conn = rt
			case oeDBOwner:
				conn = ow
			default:
				continue
			}
			n++
			got, mech := w.runRLSCase(t, conn, c)
			if got != c.Expect {
				dis++
				t.Errorf("ORACLE-DISAGREEMENT %s (%s %s %s): observed %s (%s), oracle %s (%s)", c.ID, c.Action, c.DBRole, c.Actor, got, mech, c.Expect, c.Why)
			}
		}
		fmt.Printf("OBJ-ORACLE-RLS {\"cases\":%d,\"disagreements\":%d}\n", n, dis)
		// The guard's own refusals beyond the oracle (S-review finding 1).
		refused, controls, hidden := w.guardTransitions(t, rt)
		fmt.Printf("OBJ-GUARD-TRANSITIONS {\"refused\":%d,\"controls\":%d,\"hidden\":%d}\n", refused, controls, hidden)
		if refused != 5 || controls != 2 || hidden != 2 {
			t.Errorf("guard transitions: %d refusals, %d controls and %d hidden checked, want 5, 2 and 2 (G060)", refused, controls, hidden)
		}
		// A poisoned pooled connection (a session-level context value) is refused by the context statement and
		// discarded; the next call on a fresh connection succeeds.
		pool := e.singlePool(t)
		one := e.objService(&hookDB{pool: pool}, e.spy, e.ocfg)
		owner := w.users["aOwner"]
		rctx, end := beginAs(t, one, ctx, owner, w.orgs["orgA"]) // G016 A2
		defer end()
		tc := resolve(t, one, rctx, w.orgs["orgA"], tenancy.GroupID{})
		conn, err := pool.Acquire(ctx)
		if err != nil {
			t.Fatal(err)
		}
		pid := conn.Conn().PgConn().PID()
		if _, err := conn.Exec(ctx, `SELECT set_config('app.user_id', $1, false)`, owner.id.String()); err != nil {
			t.Fatal(err)
		}
		conn.Release()
		calls := e.spy.calls.Load()
		if _, err := one.ObjectMetadata(rctx, tc, w.objs["oAOrg1"].id); CodeOf(err) != CodeUnavailable || e.lastStorage(t)["reason"] != "store_unavailable" ||
			e.spy.calls.Load() != calls {
			t.Errorf("poisoned connection: %v %v", err, e.lastStorage(t))
		}
		conn, err = pool.Acquire(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if conn.Conn().PgConn().PID() == pid {
			t.Error("the poisoned connection was reused")
		}
		conn.Release()
		if _, err := one.ObjectMetadata(rctx, tc, w.objs["oAOrg1"].id); err != nil {
			t.Errorf("after the poisoned connection: %v", err)
		}
	})

	t.Run("OBJ-RT round trips", func(t *testing.T) {
		owner, member := w.users["aOwner"], w.users["g1MemberX"]
		for _, c := range []struct {
			name  string
			user  *itUser
			group tenancy.GroupID
			size  int
			ct    string
		}{
			{"organization object", owner, tenancy.GroupID{}, 4096, "text/plain"},
			{"group object", member, w.groups["g1"], 70000, "image/png"},
			{"zero bytes", member, w.groups["g1"], 0, ""},
			{"exactly the maximum", owner, tenancy.GroupID{}, int(e.ocfg.MaxObjectBytes), "application/octet-stream"},
		} {
			rctx, end := beginAs(t, e.osvc, ctx, c.user, w.orgs["orgA"]) // G016 A2
			tc := resolve(t, e.osvc, rctx, w.orgs["orgA"], tenancy.GroupID{})
			data := payload(c.size, c.name+e.canaries[0])
			calls := e.spy.calls.Load()
			info, err := e.osvc.PutObject(rctx, tc, PutObjectRequest{Group: c.group, ContentType: c.ct, Size: int64(len(data)), Body: io.NopCloser(bytes.NewReader(data))})
			if err != nil {
				t.Fatalf("%s: put: %v", c.name, err)
			}
			wantCT := c.ct
			if wantCT == "" {
				wantCT = tenancy.DefaultContentType
			}
			if info.Size != int64(len(data)) || info.SHA256 != sha256.Sum256(data) || info.Group != c.group || info.ContentType != wantCT ||
				info.ReadyAt.IsZero() || info.Organization != w.orgs["orgA"] {
				t.Errorf("%s: info %+v", c.name, info)
			}
			if n := e.spy.calls.Load() - calls; n != 2 {
				t.Errorf("%s: %d provider calls for a put, want Put and Head", c.name, n)
			}
			key, _ := objectstore.NewKey(info.Organization.String(), info.ID.String())
			pi, err := e.store.Metadata(ctx, key)
			wantB := objectstore.Binding{Object: info.ID.String(), Organization: info.Organization.String()}
			if !zero(c.group) {
				wantB.Group = c.group.String()
			}
			if err != nil || pi.Size != int64(len(data)) || pi.Binding != wantB {
				t.Errorf("%s: provider %+v %v", c.name, pi, err)
			}
			if !strings.Contains(e.rowState(t, info.ID), key.String()) {
				t.Errorf("%s: the row's generated key differs from the derived key", c.name)
			}
			md, err := e.osvc.ObjectMetadata(rctx, tc, info.ID)
			if err != nil || md != info {
				t.Errorf("%s: metadata %+v %v", c.name, md, err)
			}
			gi, rc, err := e.osvc.GetObject(rctx, tc, info.ID)
			if err != nil {
				t.Fatalf("%s: get: %v", c.name, err)
			}
			got, err := io.ReadAll(rc)
			_ = rc.Close()
			if err != nil || !bytes.Equal(got, data) || gi != info {
				t.Errorf("%s: get %d bytes, %v", c.name, len(got), err)
			}
			e.unseed(t, info.Organization, info.ID)
			end()
		}
		// Max+1 is rejected before the provider.
		rctx, end := beginAs(t, e.osvc, ctx, owner, w.orgs["orgA"]) // G016 A2
		defer end()
		tc := resolve(t, e.osvc, rctx, w.orgs["orgA"], tenancy.GroupID{})
		calls, rows := e.spy.calls.Load(), e.objectRows(t, w.orgs["orgA"])
		_, err := e.osvc.PutObject(rctx, tc, PutObjectRequest{Size: e.ocfg.MaxObjectBytes + 1, Body: io.NopCloser(strings.NewReader("x"))})
		if CodeOf(err) != CodeInvalidInput || e.spy.calls.Load() != calls || e.objectRows(t, w.orgs["orgA"]) != rows {
			t.Errorf("max+1: %v", err)
		}
	})

	t.Run("OBJ-SCOPE scope discovery is constant in the binding count", func(t *testing.T) {
		org, g1 := w.orgs["orgA"], w.groups["g1"]
		s1, s50 := e.newUser(t), e.newUser(t)
		for _, u := range []*itUser{s1, s50} {
			e.grantOrg(t, org, u, tenancy.OrgMember)
			e.grantGroup(t, org, g1, u, tenancy.GroupMember)
		}
		creator := e.newUser(t)
		for i := 0; i < 49; i++ {
			g := e.group(t, org, creator, fmt.Sprintf("Scope %02d", i), fmt.Sprintf("objects-scope-%02d", i))
			e.grantGroup(t, org, g.ID, s50, tenancy.GroupMember)
		}
		e.bind(t, s1, org) // G016 A2: before any measured window
		e.bind(t, s50, org)
		obj := w.objs["oG1X"]
		measure := func(u *itUser) (string, []string) {
			h := &hookDB{pool: e.pool}
			svc := e.objService(h, e.spy, e.ocfg)
			rctx, end := beginAs(t, svc, ctx, u, org) // G016 A2
			defer end()
			tc := resolve(t, svc, rctx, org, tenancy.GroupID{})
			h.reset()
			h.capture = true
			if _, err := svc.ObjectMetadata(rctx, tc, obj.id); err != nil {
				t.Fatalf("metadata: %v", err)
			}
			_, steps := h.counts()
			var groupArgs []string
			for _, st := range h.capturedStmts() {
				if st.step == "GB" || st.step == "G" {
					groupArgs = append(groupArgs, fmt.Sprintf("%s:%d", st.step, len(st.args[len(st.args)-1].([]pgtype.UUID))))
				}
			}
			return strings.Join(steps, ","), groupArgs
		}
		a, ga := measure(s1)
		b, gb := measure(s50)
		if a != b || strings.Join(ga, ",") != strings.Join(gb, ",") {
			t.Errorf("s1 %s %v, s50 %s %v", a, ga, b, gb)
		}
		for _, g := range ga {
			if !strings.HasSuffix(g, ":1") {
				t.Errorf("a fact read names %s Groups, want only the owning Group", g)
			}
		}
		// The fact rows read for the decision are equal for both subjects.
		rows := func(u *itUser) string {
			tx, err := e.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = tx.Rollback(ctx) }()
			if err := tenancy.InitContext(ctx, tx, u.id, org, tenancy.GroupID{}); err != nil {
				t.Fatal(err)
			}
			tc := &TenantContext{subject: u.id, organization: org}
			f, err := tenancy.ReadFacts(ctx, tx, objectFactQuery(tc, g1, false))
			if err != nil {
				t.Fatal(err)
			}
			return fmt.Sprintf("org=%d groupBindings=%d groups=%d", len(f.OrganizationBindings), len(f.GroupBindings), len(f.Groups))
		}
		if r1, r50 := rows(s1), rows(s50); r1 != r50 || r1 != "org=1 groupBindings=1 groups=1" {
			t.Errorf("fact rows s1 %s, s50 %s", r1, r50)
		}
		fmt.Printf("OBJ-SCOPE {\"s1\":%q,\"s50\":%q}\n", a, b)
	})

	t.Run("OBJ-SCHEMA constraints, generated key and forged binding", func(t *testing.T) {
		owner := w.users["aOwner"]
		org, orgB, gb := w.orgs["orgA"], w.orgs["orgB"], w.groups["gb"]
		for _, c := range []struct{ name, sql, state string }{
			{"zero id", `INSERT INTO app.stored_objects (id, organization_id, creator_user_id, byte_size, content_type) VALUES ('` + zeroUUID + `', '` + org.String() + `', '` + owner.id.String() + `', 0, 'a/b')`, "23514"},
			{"cross-Organization Group", `INSERT INTO app.stored_objects (id, organization_id, group_id, creator_user_id, byte_size, content_type) VALUES (gen_random_uuid(), '` + org.String() + `', '` + gb.String() + `', '` + owner.id.String() + `', 0, 'a/b')`, "23503"},
			{"ready without digest", `INSERT INTO app.stored_objects (id, organization_id, creator_user_id, byte_size, content_type, state, ready_at) VALUES (gen_random_uuid(), '` + org.String() + `', '` + owner.id.String() + `', 0, 'a/b', 'ready', now())`, "23514"},
			{"oversize", `INSERT INTO app.stored_objects (id, organization_id, creator_user_id, byte_size, content_type) VALUES (gen_random_uuid(), '` + org.String() + `', '` + owner.id.String() + `', 1073741825, 'a/b')`, "23514"},
			{"content type parameters", `INSERT INTO app.stored_objects (id, organization_id, creator_user_id, byte_size, content_type) VALUES (gen_random_uuid(), '` + org.String() + `', '` + owner.id.String() + `', 0, 'text/plain; x=y')`, "23514"},
			{"written key", `INSERT INTO app.stored_objects (id, organization_id, creator_user_id, byte_size, content_type, storage_key) VALUES (gen_random_uuid(), '` + org.String() + `', '` + owner.id.String() + `', 0, 'a/b', 'objects/x/y')`, "428C9"},
			{"duplicate id", `INSERT INTO app.stored_objects (id, organization_id, creator_user_id, byte_size, content_type) VALUES ('` + w.objs["oAOrg1"].id.String() + `', '` + orgB.String() + `', '` + owner.id.String() + `', 0, 'a/b')`, "23505"},
		} {
			before := e.rowState(t, w.objs["oAOrg1"].id)
			tx, err := e.setup.Begin(ctx)
			if err != nil {
				t.Fatal(err)
			}
			_, err = tx.Exec(ctx, c.sql)
			_ = tx.Rollback(ctx)
			if pgState(err) != c.state {
				t.Errorf("%s: %v, want %s", c.name, err, c.state)
			}
			if e.rowState(t, w.objs["oAOrg1"].id) != before || !e.present(t, w.objs["oAOrg1"].key(t)) {
				t.Errorf("%s: the original row or bytes changed", c.name)
			}
		}
		// A forged binding (the provider object overwritten under the same key with another binding, through the
		// scoped credential) makes a guarded Get internal, with nothing returned.
		f := w.objs["oAOrg1"]
		forged := f.binding()
		forged.Object = w.objs["oAOrgFormer"].id.String()
		if err := e.store.Put(ctx, f.key(t), bytes.NewReader(f.data), int64(len(f.data)), f.ct, forged); err != nil {
			t.Fatal(err)
		}
		rctx, end := beginAs(t, e.osvc, ctx, owner, org) // G016 A2
		defer end()
		tc := resolve(t, e.osvc, rctx, org, tenancy.GroupID{})
		info, rc, err := e.osvc.GetObject(rctx, tc, f.id)
		if CodeOf(err) != CodeInternal || rc != nil || info != (ObjectInfo{}) || e.lastStorage(t)["reason"] != "binding_mismatch" {
			t.Errorf("forged binding: %v", err)
		}
		e.unseed(t, f.org, f.id)
		e.seed(t, f)
	})

	t.Run("OBJ-DEL physical deletion, tombstone and idempotent repeat", func(t *testing.T) {
		admin := w.users["aAdmin"]
		rctx, end := beginAs(t, e.osvc, ctx, admin, w.orgs["orgA"]) // G016 A2
		defer end()
		tc := resolve(t, e.osvc, rctx, w.orgs["orgA"], tenancy.GroupID{})
		f := w.objs["oAOrg1"]
		res, err := e.osvc.DeleteObject(rctx, tc, f.id)
		if err != nil || res.Outcome != DeleteOutcomeDeleted || e.present(t, f.key(t)) {
			t.Fatalf("delete: %+v %v", res, err)
		}
		row := e.rowState(t, f.id)
		if !strings.Contains(row, ",deleted,") || !strings.Contains(row, admin.id.String()) {
			t.Errorf("tombstone %s", itH(row))
		}
		if rec := e.lastStorage(t); rec["physical"] != "confirmed" || rec["outcome"] != "allow" {
			t.Errorf("record %v", rec)
		}
		res, err = e.osvc.DeleteObject(rctx, tc, f.id)
		if err != nil || res.Outcome != DeleteOutcomeDeleted {
			t.Errorf("repeat delete: %+v %v", res, err)
		}
		for name, op := range map[string]func() error{
			"metadata": func() error { _, err := e.osvc.ObjectMetadata(rctx, tc, f.id); return err },
			"get":      func() error { _, _, err := e.osvc.GetObject(rctx, tc, f.id); return err },
			"sign": func() error {
				_, err := e.osvc.SignedGetURL(rctx, tc, SignedGetURLRequest{Object: f.id, DefaultTTL: true})
				return err
			},
		} {
			if CodeOf(op()) != CodeNotFound {
				t.Errorf("%s after deletion: want not_found", name)
			}
		}
		for _, id := range []tenancy.ObjectID{w.missID, w.objs["oBOrg"].id} {
			if _, err := e.osvc.DeleteObject(rctx, tc, id); CodeOf(err) != CodeNotFound {
				t.Errorf("missing or foreign delete: %v", err)
			}
		}
		e.unseed(t, f.org, f.id)
		e.seed(t, f)
	})

	t.Run("OBJ-SIGN real fetch, expiry, alteration and loopback binding", func(t *testing.T) {
		member := w.users["g1MemberX"]
		rctx, end := beginAs(t, e.osvc, ctx, member, w.orgs["orgA"]) // G016 A2
		defer end()
		tc := resolve(t, e.osvc, rctx, w.orgs["orgA"], w.groups["g1"])
		f := w.objs["oG1X"]
		client := &http.Client{Timeout: 10 * time.Second}
		fetch := func(method, raw, host string) (int, []byte, http.Header) {
			t.Helper()
			req, err := http.NewRequest(method, raw, nil)
			if err != nil {
				t.Fatal(err)
			}
			if host != "" {
				req.Host = host
			}
			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("fetch: %s", objRedact(err.Error()))
			}
			defer resp.Body.Close()
			b, _ := io.ReadAll(resp.Body)
			return resp.StatusCode, b, resp.Header
		}
		for _, req := range []SignedGetURLRequest{{Object: f.id, DefaultTTL: true}, {Object: f.id, TTLSeconds: 300}} {
			u, err := e.osvc.SignedGetURL(rctx, tc, req)
			if err != nil {
				t.Fatal(err)
			}
			st, body, hdr := fetch(http.MethodGet, u.URL(), "")
			if st != 200 || !bytes.Equal(body, f.data) || hdr.Get("Content-Disposition") != "attachment" ||
				hdr.Get("Content-Type") != f.ct || hdr.Get("Cache-Control") != "no-store" {
				t.Errorf("signed fetch: %d, %d bytes, %v", st, len(body), hdr)
			}
			if st, _, _ := fetch(http.MethodGet, u.URL(), ""); st != 200 {
				t.Errorf("reuse within the TTL: %d", st)
			}
			if time.Until(u.ExpiresAt) > u.TTL || u.TTL != map[bool]time.Duration{true: 60 * time.Second, false: 300 * time.Second}[req.DefaultTTL] {
				t.Errorf("expiry %v ttl %v", time.Until(u.ExpiresAt), u.TTL)
			}
			pu, _ := url.Parse(u.URL())
			alter := func(fn func(q url.Values, p *url.URL)) string {
				c := *pu
				q := c.Query()
				fn(q, &c)
				c.RawQuery = q.Encode()
				return c.String()
			}
			for name, raw := range map[string]string{
				"key": alter(func(_ url.Values, p *url.URL) {
					p.Path = strings.Replace(p.Path, f.id.String(), w.objs["oG1Y"].id.String(), 1)
				}),
				"signature": alter(func(q url.Values, _ *url.URL) { q.Set("X-Amz-Signature", strings.Repeat("0", 64)) }),
				"expiry":    alter(func(q url.Values, _ *url.URL) { q.Set("X-Amz-Expires", "600") }),
				"override":  alter(func(q url.Values, _ *url.URL) { q.Set("response-content-type", "text/html") }),
				"anonymous": alter(func(q url.Values, p *url.URL) {
					for k := range q {
						q.Del(k)
					}
				}),
			} {
				if st, _, _ := fetch(http.MethodGet, raw, ""); st != 403 {
					t.Errorf("altered %s: %d, want 403", name, st)
				}
			}
			if st, _, _ := fetch(http.MethodPut, u.URL(), ""); st != 403 {
				t.Errorf("altered method: %d, want 403", st)
			}
			if st, _, _ := fetch(http.MethodGet, u.URL(), "attacker.example:9000"); st != 403 {
				t.Errorf("altered host: %d, want 403", st)
			}
		}
		for _, bad := range []SignedGetURLRequest{{Object: f.id, TTLSeconds: 0}, {Object: f.id, TTLSeconds: -1}, {Object: f.id, TTLSeconds: 301},
			{Object: f.id, TTLSeconds: 60, DefaultTTL: true}} {
			if _, err := e.osvc.SignedGetURL(rctx, tc, bad); CodeOf(err) != CodeInvalidInput {
				t.Errorf("TTL %+v: %v", bad, err)
			}
		}
		// Real expiry: a 2 s URL works, then fails after it (the one bounded wait).
		u, err := e.osvc.SignedGetURL(rctx, tc, SignedGetURLRequest{Object: f.id, TTLSeconds: 2})
		if err != nil {
			t.Fatal(err)
		}
		if st, _, _ := fetch(http.MethodGet, u.URL(), ""); st != 200 {
			t.Errorf("before expiry: %d", st)
		}
		time.Sleep(time.Until(u.ExpiresAt) + 1500*time.Millisecond)
		if st, _, _ := fetch(http.MethodGet, u.URL(), ""); st != 403 {
			t.Errorf("after expiry: %d", st)
		}
		// Loopback binding: a URL signed for http://127.0.0.1:<port> is valid only with exactly that Host.
		loop := e.scfg
		loop.SigningEndpoint = "http://127.0.0.1:19999"
		ls, err := objectstore.New(loop, os.Getenv("APP_ENV"))
		if err != nil {
			t.Fatal(err)
		}
		raw, err := ls.SignedGetURL(ctx, f.key(t), time.Minute, f.ct)
		if err != nil {
			t.Fatal(err)
		}
		lu, _ := url.Parse(raw)
		data, _ := url.Parse(e.scfg.Endpoint)
		lu.Host = data.Host
		if st, body, _ := fetch(http.MethodGet, lu.String(), "127.0.0.1:19999"); st != 200 || !bytes.Equal(body, f.data) {
			t.Errorf("loopback URL with its signed Host: %d", st)
		}
		for _, host := range []string{"127.0.0.1:19998", "localhost:19999", data.Host} {
			if st, _, _ := fetch(http.MethodGet, lu.String(), host); st != 403 {
				t.Errorf("loopback URL with Host %s: %d, want 403", host, st)
			}
		}
		if strings.Contains(e.logs.String(), u.URL()) || strings.Contains(e.logs.String(), "X-Amz-Signature") {
			t.Error("a signed URL reached the log")
		}
	})

	t.Run("OBJ-ADMIN the scoped credential administers nothing; anonymous access is denied", func(t *testing.T) {
		client := s3.New(s3.Options{BaseEndpoint: aws.String(e.scfg.Endpoint), Region: e.scfg.Region, UsePathStyle: true,
			Credentials: credentials.NewStaticCredentialsProvider(e.scfg.AccessKey, e.scfg.SecretKey, ""), Retryer: aws.NopRetryer{}})
		status := func(err error) string {
			var re interface{ HTTPStatusCode() int }
			if errors.As(err, &re) {
				return fmt.Sprint(re.HTTPStatusCode())
			}
			if err == nil {
				return "ok"
			}
			return "error"
		}
		policy := `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":"*","Action":["s3:GetObject"],"Resource":["arn:aws:s3:::` + e.scfg.Bucket + `/*"]}]}`
		for name, err := range map[string]error{
			"CreateBucket": func() error {
				_, err := client.CreateBucket(ctx, &s3.CreateBucketInput{Bucket: aws.String("g009-other-bucket")})
				return err
			}(),
			"DeleteBucket": func() error {
				_, err := client.DeleteBucket(ctx, &s3.DeleteBucketInput{Bucket: aws.String(e.scfg.Bucket)})
				return err
			}(),
			"PutBucketPolicy": func() error {
				_, err := client.PutBucketPolicy(ctx, &s3.PutBucketPolicyInput{Bucket: aws.String(e.scfg.Bucket), Policy: aws.String(policy)})
				return err
			}(),
			"GetBucketPolicy": func() error {
				_, err := client.GetBucketPolicy(ctx, &s3.GetBucketPolicyInput{Bucket: aws.String(e.scfg.Bucket)})
				return err
			}(),
		} {
			if s := status(err); s != "403" {
				t.Errorf("scoped %s: %s, want 403", name, s)
			}
		}
		// The MinIO admin API with a SigV4-signed request of the scoped credential.
		req, _ := http.NewRequest(http.MethodGet, strings.TrimRight(e.scfg.Endpoint, "/")+"/minio/admin/v3/info", nil)
		empty := sha256.Sum256(nil)
		req.Header.Set("X-Amz-Content-Sha256", hex.EncodeToString(empty[:]))
		if err := v4.NewSigner().SignHTTP(ctx, aws.Credentials{AccessKeyID: e.scfg.AccessKey, SecretAccessKey: e.scfg.SecretKey},
			req, hex.EncodeToString(empty[:]), "s3", e.scfg.Region, time.Now()); err != nil {
			t.Fatal(err)
		}
		// G063 repair: every raw probe dials its own connection. http.DefaultClient shares DefaultTransport's keep-alive
		// pool with OBJ-SIGN's fetches. A post-merge qualification run failed with `Put ".../objects/anonymous": EOF`,
		// a plain EOF that is neither a denial nor proof of one. Unproven hypothesis for its cause: a pooled connection
		// the server had since closed, which net/http does not replay for a PUT (a body, no Idempotency-Key) but does
		// replay silently for a GET. A fresh connection that ends without a response is still a failure, not a denial.
		raw := &http.Client{Timeout: 10 * time.Second, Transport: &http.Transport{Proxy: http.ProxyFromEnvironment, DisableKeepAlives: true}}
		denied := func(name string, r *http.Request) {
			t.Helper()
			resp, err := raw.Do(r)
			if err != nil {
				t.Fatalf("%s: %s", name, objRedact(err.Error()))
			}
			_, _ = io.Copy(io.Discard, resp.Body)
			_ = resp.Body.Close()
			if resp.StatusCode != 403 {
				t.Errorf("%s: %d, want 403", name, resp.StatusCode)
			}
		}
		denied("scoped admin API", req)
		f := w.objs["oAOrg1"]
		base := strings.TrimRight(e.scfg.Endpoint, "/") + "/" + e.scfg.Bucket
		for _, p := range []struct {
			name string
			req  *http.Request
		}{
			{"anonymous list", must(http.NewRequest(http.MethodGet, base+"/", nil))},
			{"anonymous get", must(http.NewRequest(http.MethodGet, base+"/"+f.key(t).String(), nil))},
			{"anonymous put", must(http.NewRequest(http.MethodPut, base+"/objects/anonymous", strings.NewReader("x")))},
		} {
			denied(p.name, p.req)
		}
		// The denied anonymous PUT wrote nothing: the scoped credential (s3:GetObject on the bucket's keys) finds no
		// such key.
		_, err := client.HeadObject(ctx, &s3.HeadObjectInput{Bucket: aws.String(e.scfg.Bucket), Key: aws.String("objects/anonymous")})
		if s := status(err); s != "404" {
			t.Errorf("after the denied anonymous put, objects/anonymous: %s, want 404", s)
		}
	})

	t.Run("OBJ-LOG records catalogued, no canary or secret", func(t *testing.T) {
		logs := e.logs.String()
		if !strings.Contains(logs, `"msg":"storage.operation"`) && !strings.Contains(logs, `"event":"storage.operation"`) {
			t.Fatal("no storage.operation record")
		}
		for _, s := range []string{e.scfg.SecretKey, "X-Amz-Signature", "X-Amz-Credential", "objects/" + w.orgs["orgA"].String(), "image/png",
			"application/octet-stream", string(w.objs["oAOrg1"].data[:16])} {
			if strings.Contains(logs, s) {
				t.Errorf("the log carries %s", itH(s))
			}
		}
		for _, line := range strings.Split(strings.TrimSpace(logs), "\n") {
			if strings.Contains(line, "storage.operation") && strings.Contains(line, "access.decision") {
				t.Errorf("a storage call also wrote access.decision")
			}
		}
	})
}

// singlePool is a runtime pool of exactly one connection (deterministic connection reuse).
func (e *objEnv) singlePool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	pc, err := pgxpool.ParseConfig("")
	if err != nil {
		t.Fatal(err)
	}
	pc.ConnConfig.Host, pc.ConnConfig.Port, pc.ConnConfig.Database = e.cfg.Host, uint16(e.cfg.Port), e.cfg.Database
	pc.ConnConfig.User, pc.ConnConfig.Password = e.cfg.RuntimeRole(), e.pw
	pc.ConnConfig.ConnectTimeout = 5 * time.Second
	pc.ConnConfig.TLSConfig, pc.ConnConfig.Fallbacks = nil, nil
	pc.MaxConns = 1
	p, err := pgxpool.NewWithConfig(context.Background(), pc)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(p.Close)
	return p
}

// lastStorage returns the last storage.operation record.
func (e *objEnv) lastStorage(t *testing.T) map[string]any {
	t.Helper()
	lines := strings.Split(strings.TrimSpace(e.logs.String()), "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		if strings.Contains(lines[i], "storage.operation") {
			var rec map[string]any
			if err := json.Unmarshal([]byte(lines[i]), &rec); err != nil {
				t.Fatal(err)
			}
			return rec
		}
	}
	t.Fatal("no storage.operation record")
	return nil
}

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

// objRedact removes a signed query from an error text before it is printed.
func objRedact(s string) string {
	if i := strings.Index(s, "?"); i >= 0 {
		return s[:i] + "?[redacted]"
	}
	return s
}

// --- OBJ-SIGN-HOST (Plan r1 §4.5, §5.5; correction 4) ---------------------------------------------------------------------

// TestObjectsSignHost is the issuing half of OBJ-SIGN-HOST: it uploads a real guarded object, issues a signed URL for
// the migrations project's published MinIO endpoint (OBJ_SIGN_HOST_ENDPOINT, http://127.0.0.1:<published port>), writes
// the URL and the expected body digest with mode 0600 into OBJ_SIGN_HOST_DIR (a tmpfs, never a host bind mount), and
// waits up to 60 s for the harness to fetch the unchanged URL through the host port and create OBJ_SIGN_HOST_DIR/done.
// The URL is never printed or logged.
func TestObjectsSignHost(t *testing.T) {
	endpoint, dir := os.Getenv("OBJ_SIGN_HOST_ENDPOINT"), os.Getenv("OBJ_SIGN_HOST_DIR")
	if endpoint == "" || dir == "" {
		t.Fatal("OBJ_SIGN_HOST_ENDPOINT and OBJ_SIGN_HOST_DIR are required")
	}
	e := setupObjects(t)
	ctx := context.Background()
	org := e.org(t, "Objects Sign Host "+e.canaries[0])
	owner := e.newUser(t)
	e.grantOrg(t, org, owner, tenancy.OrgOwner)
	e.bind(t, owner, org) // G016 A2
	rctx, end := beginAs(t, e.osvc, ctx, owner, org)
	defer end()
	tc := resolve(t, e.osvc, rctx, org, tenancy.GroupID{})
	data := payload(5000, "sign-host"+e.canaries[1])
	info, err := e.osvc.PutObject(rctx, tc, PutObjectRequest{ContentType: "text/plain", Size: int64(len(data)), Body: io.NopCloser(bytes.NewReader(data))})
	if err != nil {
		t.Fatal(err)
	}
	defer e.unseed(t, info.Organization, info.ID)
	hostCfg := e.scfg
	hostCfg.SigningEndpoint = endpoint
	hs, err := objectstore.New(hostCfg, os.Getenv("APP_ENV"))
	if err != nil {
		t.Fatalf("signing store: %v", err)
	}
	svc := e.objService(e.hook, hs, e.ocfg)
	rctx2, end2 := beginAs(t, svc, ctx, owner, org) // G016 A2
	defer end2()
	u, err := svc.SignedGetURL(rctx2, resolve(t, svc, rctx2, org, tenancy.GroupID{}), SignedGetURLRequest{Object: info.ID, DefaultTTL: true})
	if err != nil {
		t.Fatal(err)
	}
	pu, err := url.Parse(u.URL())
	if err != nil || pu.Scheme+"://"+pu.Host != endpoint {
		t.Fatalf("the signed URL does not name the published endpoint")
	}
	sum := sha256.Sum256(data)
	if err := os.WriteFile(dir+"/url", []byte(u.URL()+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dir+"/sha256", []byte(hex.EncodeToString(sum[:])+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	fmt.Println("OBJ-SIGN-HOST-READY")
	deadline := time.Now().Add(60 * time.Second)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(dir + "/done"); err == nil {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
	_ = os.Remove(dir + "/url")
	if _, err := os.Stat(dir + "/done"); err != nil {
		t.Fatal("the harness did not finish the published-host fetch within 60 s")
	}
	if strings.Contains(e.logs.String(), pu.RawQuery) {
		t.Error("the signed URL reached the log")
	}
}

// --- persistence (B-OBJ-MIGRATION restart) ---------------------------------------------------------------------------------

// TestObjectsPersistSeed uploads a guarded object and prints its identifiers and digest (OBJ-PERSIST); the harness then
// restarts PostgreSQL and MinIO and runs TestObjectsPersistVerify with OBJ_PERSIST_IDS.
func TestObjectsPersistSeed(t *testing.T) {
	e := setupObjects(t)
	ctx := context.Background()
	org := e.org(t, "Objects Persist "+e.canaries[0])
	owner := e.newUser(t)
	e.grantOrg(t, org, owner, tenancy.OrgOwner)
	e.bind(t, owner, org) // G016 A2
	rctx, end := beginAs(t, e.osvc, ctx, owner, org)
	defer end()
	tc := resolve(t, e.osvc, rctx, org, tenancy.GroupID{})
	data := payload(9000, "persist"+e.canaries[2])
	info, err := e.osvc.PutObject(rctx, tc, PutObjectRequest{Size: int64(len(data)), Body: io.NopCloser(bytes.NewReader(data))})
	if err != nil {
		t.Fatal(err)
	}
	b, _ := json.Marshal(map[string]string{"org": org.String(), "object": info.ID.String(), "sha256": hex.EncodeToString(info.SHA256[:])})
	fmt.Println("OBJ-PERSIST " + string(b))
}

// TestObjectsPersistVerify: after the restart, a guarded Get by a newly authorized reader returns identical bytes; then
// the object is removed.
func TestObjectsPersistVerify(t *testing.T) {
	var ids map[string]string
	if err := json.Unmarshal([]byte(os.Getenv("OBJ_PERSIST_IDS")), &ids); err != nil || ids["object"] == "" {
		t.Fatalf("OBJ_PERSIST_IDS: %v", err)
	}
	e := setupObjects(t)
	ctx := context.Background()
	org, err := tenancy.ParseOrganizationID(ids["org"])
	if err != nil {
		t.Fatal(err)
	}
	id, err := tenancy.ParseObjectID(ids["object"])
	if err != nil {
		t.Fatal(err)
	}
	defer e.unseed(t, org, id)
	reader := e.newUser(t)
	e.grantOrg(t, org, reader, tenancy.OrgMember)
	e.bind(t, reader, org) // G016 A2
	rctx, end := beginAs(t, e.osvc, ctx, reader, org)
	defer end()
	tc := resolve(t, e.osvc, rctx, org, tenancy.GroupID{})
	info, rc, err := e.osvc.GetObject(rctx, tc, id)
	if err != nil {
		t.Fatalf("guarded get after restart: %v", err)
	}
	got, err := io.ReadAll(rc)
	_ = rc.Close()
	sum := sha256.Sum256(got)
	if err != nil || hex.EncodeToString(sum[:]) != ids["sha256"] || info.SHA256 != sum {
		t.Fatalf("bytes after restart differ (%v)", err)
	}
}
