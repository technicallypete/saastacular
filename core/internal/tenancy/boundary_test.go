package tenancy

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"
)

// The pinned exported Service surface (G005 Plan r4 §4.2 plus G006 Plan r2 §4.2): exactly these 29 methods, plus G019's
// CreateOrganizationWithOwner (Plan r1 §1.4 C2; the seven fixture-only primitives moved to fixture.go keep their names).
var pinnedServiceMethods = []string{
	"ChangeGroupRole", "ChangeGroupSlug", "ChangeOrganizationRole", "CreateBusinessProfile", "CreateGroup", "CreateOrganization",
	"CreateOrganizationWithOwner", "GetBusinessProfile", "GetGroup", "GetGroupBusinessProfile", "GetGroupBySlug", "GetGroupMembership", "GetOrganization",
	"GetOrganizationMembership", "GrantGroupRole", "GrantOrganizationRole", "LinkGroupBusinessProfile",
	"ListBusinessProfileAccessPaths", "ListGroupMemberships", "ListGroups", "ListGroupsForBusinessProfile",
	"ListOrganizationMemberships", "ListUserBusinessProfileAccessPaths", "RemoveGroupMembership", "RemoveOrganizationMembership",
	"RenameGroup", "RenameOrganization", "SuggestGroupSlug", "UnlinkGroupBusinessProfile", "UpdateBusinessProfile",
}

// The pinned exported package-level functions (G005, G006 and G007 Plan r1 §3.7): the parsers, validators, CodeOf,
// ReservedSlugs and NewService, plus G007's ReadFacts and the four transaction-scoped writes, G008's InitContext and
// G009's stored-object parser, validator, key derivation and six transaction-scoped primitives (objects.go). Nothing
// else outside the Service surface is callable from another package. G017 adds the seven transaction-scoped branding
// primitives and the public projection read (branding.go). G018 adds the onboarding primitives (onboarding.go): the three writers, the two list pages and the slug suggestion.
// G019 adds the membership primitives (membership.go: ChangeRoleTx, RemoveBindingTx, OwnerCount (G060 amendment01 drops SubjectBindingsLock),
// MembersPage, InvitationCandidatesPage) and the invitation helpers (invitations.go: ParseInvitationID,
// InvitationRoleValid, NormalizeAddress, NewInvitationToken, ValidInvitationToken, InvitationDigest, DigestEqual) with
// the invitation writers and readers (InsertInvitationTx, RevokeInvitationTx, ClaimInvitationTx, InvitationsPage,
// MyInvitationsPage, PendingInvitationCount), and, for the access operations, InvitationTx (the invitation read) and
// SoleOwnerScopes (the retirement-eligibility reader), and OwnOrganizationsPage (the trusted-retirement enumeration,
// Result addendum A1-01). G058 adds the enrollment admission primitive (enrollment.go: AdmitEnrollmentTx) and, slice 1b,
// the enrollment policy parsers and primitives (ParseEnrollmentMode, ParseEnrollmentPolicy, ReadEnrollmentPolicyTx,
// SetEnrollmentModeTx), and, request slice, the request primitives (AdmitRequestTx, RequestEnrollmentTx,
// EnrollmentRequestTx, DecideEnrollmentRequestTx, RevokeEnrollmentApprovalTx, EnrollmentRequestsPage, EnrollmentStateTx,
// AbandonEnrollmentTx) with their id parser and state validator (ParseEnrollmentRequestID, RequestStatusValid), and, HTTP
// slice, the public mode read (ReadPublicEnrollmentMode). G020 adds the label primitives (labels.go: the id and visibility
// parsers, the name validator, the definition read, lock, page, insert, update and two-statement delete, and the
// assignment read by target, insert and delete).
var pinnedPackageFunctions = []string{
	"AbandonEnrollmentTx", "AdmitEnrollmentTx", "AdmitRequestTx", "ChangeGroupSlugTx", "ChangeRoleTx", "ClaimInvitationTx",
	"CodeOf", "CreateBusinessProfileTx", "CreateGroupTx", "DecideEnrollmentRequestTx", "DeleteLabelAssignmentTx",
	"DeleteLabelDefinitionAssignmentsTx", "DeleteLabelDefinitionTx", "DigestEqual", "EnrollmentRequestTx", "EnrollmentRequestsPage",
	"EnrollmentStateTx", "FinalizeObjectTx", "GroupsPage", "InitContext", "InsertBrandingTx", "InsertInvitationTx",
	"InsertLabelAssignmentTx", "InsertLabelDefinitionTx", "InsertObjectTx", "InvitationCandidatesPage", "InvitationDigest",
	"InvitationRoleValid", "InvitationTx", "InvitationsPage", "LabelAssignmentsByTarget", "LabelDefinitionTx",
	"LabelDefinitionsPage", "LinkGroupBusinessProfileTx", "ListLogoUploadsTx", "LockBrandingTx", "LockLabelDefinitionTx",
	"LockObjectTx", "LogRetirement", "MarkObjectDeletedTx", "MarkObjectDeletingTx", "MembersPage", "MyInvitationsPage",
	"NewInvitationToken", "NewService", "NormalizeAddress", "NormalizeLabelTx", "ObjectStorageKey", "OwnOrganizationsPage",
	"OwnerCount", "ParseAssociationID", "ParseBindingID", "ParseBusinessProfileID", "ParseEnrollmentMode", "ParseEnrollmentPolicy",
	"ParseEnrollmentRequestID", "ParseGroupID", "ParseInvitationID", "ParseLabelDefinitionID", "ParseLabelVisibility",
	"ParseObjectID", "ParseOrganizationID", "ParseUserID", "PendingInvitationCount", "ProfilesPage", "ReadBrandingTx",
	"ReadEnrollmentPolicyTx", "ReadFacts", "ReadObjectTx", "ReadPublicBranding", "ReadPublicEnrollmentMode", "RemoveBindingTx",
	"RenameGroupTx", "RequestEnrollmentTx", "RequestStatusValid", "ReservedSlugs", "RevokeEnrollmentApprovalTx",
	"RevokeInvitationTx", "SelectableGroupsPage", "SetEnrollmentModeTx", "SetGroupManagementOverrideTx", "SetOrganizationPolicyTx",
	"ShareLogoObjectTx", "SoleOwnerScopes", "SuggestGroupSlugTx", "UpdateBrandingTx", "UpdateBusinessProfileTx",
	"UpdateLabelDefinitionTx", "ValidInvitationToken", "ValidateBusinessContext", "ValidateContentType", "ValidateLabelName",
	"ValidateName", "ValidateSlug", "ValidateWebsite",
}

// lockingStatements are the store.go statements that take row locks (G007 Plan r1 §4.1): the fact reads in their
// locking forms and Unlink's Group precheck. No other statement carries a locking clause.
var lockingStatements = []string{
	"sqlFactAllGroupBindingsLock", "sqlFactGroupBindingsLock", "sqlFactGroupsLock", "sqlFactOrganizationBindingLock",
	"sqlFactOrganizationExclusive", "sqlFactOrganizationShare", "sqlFactTargetGroupLock", "sqlLockGroupForUnlink",
	"sqlLockLabelDefinition",  // G020 (labels.go, Plan r2 §1.3): the label definition row after the facts (FOR NO KEY UPDATE: update, delete)
	"sqlLockOwnerStatus",      // G019 F-01 (a): the bootstrap owner's active-status lock (FOR SHARE)
	"sqlShareLabelDefinition", // G020 (labels.go): the label definition row after the facts (FOR SHARE: assign, unassign)
	"sqlSubjectBindingsLock",  // G019: the Subject fact (FOR NO KEY UPDATE); G060 amendment01 drops the departure lock
}

// ownerStatusLock is the one locking read without an Organization predicate (G019 F-01 refinement a): the Organization
// does not exist yet when CreateOrganizationWithOwner locks its owner's users row, pinned by exact text.
const ownerStatusLock = "SELECT status FROM app.users WHERE id = $1 FOR SHARE"

// Standard-library imports no production file may use (G006 Plan r2 §4.1), besides net/http and its subpackages.
var deniedStdImports = map[string]bool{"crypto/tls": true, "os/exec": true, "html/template": true, "text/template": true}

// Non-standard-library imports a production file may use.
var allowedModuleImports = map[string]bool{
	"github.com/jackc/pgx/v5": true, "github.com/jackc/pgx/v5/pgconn": true, "github.com/jackc/pgx/v5/pgtype": true,
	"core/internal/observability": true,
	// G019 (correction-01 B-1(ii)): domain NFC normalization of invitation addresses.
	"golang.org/x/text/unicode/norm": true,
}

// productionFiles parses every non-test Go file of the package.
func productionFiles(t *testing.T) (*token.FileSet, map[string]*ast.File) {
	t.Helper()
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	fset := token.NewFileSet()
	files := map[string]*ast.File{}
	for _, e := range entries {
		n := e.Name()
		if e.IsDir() || !strings.HasSuffix(n, ".go") || strings.HasSuffix(n, "_test.go") {
			continue
		}
		f, err := parser.ParseFile(fset, n, nil, parser.ParseComments)
		if err != nil {
			t.Fatal(err)
		}
		files[n] = f
	}
	if len(files) != 20 {
		t.Fatalf("production files: %d, want the 7 of G005 Plan r4 §4.1, the 3 of G006 Plan r2 §4.1, facts.go of G007 Plan r1 §3.7, "+
			"objects.go of G009 Plan r1 §2, mcp.go of G014 Plan r8 §8, branding.go of G017 Plan r2 §5.1, onboarding.go of G018 Plan r5 §3, "+
			"membership.go, invitations.go and fixture.go of G019 Plan r1 §3, enrollment.go of G058 Plan r3 §4.3 and labels.go of G020 Plan r2 §2", len(files))
	}
	return fset, files
}

// U-BOUNDARY: imports stay inside the allow-list (never net/http, the MCP SDK, identity, WorkOS or configuration);
// there is no HTTP handler; the exported Service surface is exactly the pinned 29 methods and the exported
// package-level functions exactly the pinned list.
func TestPackageBoundary(t *testing.T) {
	_, files := productionFiles(t)
	var methods, funcs []string
	for name, f := range files {
		for _, imp := range f.Imports {
			p, _ := strconv.Unquote(imp.Path.Value)
			std := !strings.Contains(strings.SplitN(p, "/", 2)[0], ".") && !strings.HasPrefix(p, "core/")
			if p == "net/http" || strings.HasPrefix(p, "net/http/") || deniedStdImports[p] || (!std && !allowedModuleImports[p]) {
				t.Errorf("%s imports %s", name, p)
			}
		}
		for _, d := range f.Decls {
			fn, ok := d.(*ast.FuncDecl)
			if !ok {
				continue
			}
			if fn.Name.Name == "ServeHTTP" {
				t.Errorf("%s declares ServeHTTP", name)
			}
			if fn.Recv == nil && fn.Name.IsExported() {
				funcs = append(funcs, fn.Name.Name)
			}
			if fn.Recv == nil || !fn.Name.IsExported() {
				continue
			}
			if recv := fn.Recv.List[0].Type; isServiceRecv(recv) {
				methods = append(methods, fn.Name.Name)
			}
		}
	}
	slices.Sort(methods)
	if strings.Join(methods, ",") != strings.Join(pinnedServiceMethods, ",") {
		t.Fatalf("exported Service methods:\n got %v\nwant %v", methods, pinnedServiceMethods)
	}
	slices.Sort(funcs)
	if strings.Join(funcs, ",") != strings.Join(pinnedPackageFunctions, ",") {
		t.Fatalf("exported package-level functions:\n got %v\nwant %v", funcs, pinnedPackageFunctions)
	}
}

func isServiceRecv(e ast.Expr) bool {
	if s, ok := e.(*ast.StarExpr); ok {
		e = s.X
	}
	id, ok := e.(*ast.Ident)
	return ok && id.Name == "Service"
}

// U-BOUNDARY (SQL): writes target only organizations, groups, role_bindings, business_profiles and
// group_business_profiles; a role value is written only to role_bindings; no statement writes a projection; membership
// and access-path reads use the projections, never role_bindings; every read or change of a tenant-owned relation
// carries the explicit Organization predicate.
func TestSQLBoundary(t *testing.T) {
	_, files := productionFiles(t)
	stmts := constStrings(t, files["store.go"])
	// G020: labels.go's label statements are store statements too (TestLabelStatements pins their shape).
	for name, sql := range constStrings(t, files["labels.go"]) {
		if _, dup := stmts[name]; dup {
			t.Fatalf("labels.go redeclares %s", name)
		}
		stmts[name] = sql
	}
	write := regexp.MustCompile(`^(INSERT INTO|UPDATE|DELETE FROM) app\.([a-z_]+)`)
	var n, writes int
	for name, sql := range stmts {
		if !strings.HasPrefix(name, "sql") {
			continue
		}
		n++
		body, _, _ := strings.Cut(sql, " RETURNING ")
		if m := write.FindStringSubmatch(sql); m != nil {
			writes++
			target := m[2]
			if !writeTargets[target] {
				t.Errorf("%s writes %s", name, target)
			}
			// G019: an invitation's role is the offered role (admin/member, CHECK-pinned), not a grant; only the claim
			// bodies turn it into a binding.
			if regexp.MustCompile(`\brole\b`).MatchString(body) && target != "role_bindings" && target != "invitations" {
				t.Errorf("%s writes a role outside role_bindings", name)
			}
			if target != "organizations" && !strings.HasPrefix(m[1], "INSERT") && !strings.Contains(body, "organization_id = $1") {
				t.Errorf("%s changes %s without the Organization predicate", name, target)
			}
			continue
		}
		if !strings.HasPrefix(sql, "SELECT ") {
			t.Errorf("%s is neither a write nor a read: %s", name, sql)
		}
		if strings.Contains(sql, "app.role_bindings") && !allowedDirectRead(name, sql) {
			t.Errorf("%s reads role_bindings directly; membership reads use the projections", name)
		}
		for _, rel := range []string{"app.groups", "app.organization_memberships", "app.group_memberships", "app.business_profiles",
			"app.group_business_profiles", "app.business_profile_access_paths"} {
			if strings.Contains(sql, rel) && !strings.Contains(sql, "organization_id = $1") {
				t.Errorf("%s reads %s without the Organization predicate", name, rel)
			}
		}
	}
	if n != 97 || writes != 30 {
		t.Fatalf("statements: %d (%d writes), want 97 (30): G020 adds the eleven label statements of labels.go (six writes: the definition insert, update and its two-statement delete, the assignment insert and delete; five reads: the definition read and its two locks, the definition page and the assignments by target); G060 adds the owner-only override write sqlSetGroupManagementOverride and (amendment01) drops the Organization-departure lock sqlSubjectGroupBindingsLock (the guard cascades it); G058 adds the enrollment admission call sqlEnrollmentAdmit and (slice 1b) the policy read sqlReadEnrollmentPolicy and the mode write sqlSetEnrollmentMode and (request slice) the request call, the request read, the decision and revocation writes, the request page, the applicant's latest request, its credential state and its abandonment write and (HTTP slice) the public mode read sqlPublicEnrollmentMode; G008 adds the read-only context statement sqlInitContext, G014 the "+
			"two selectable-Group page reads, G018 the two keyed onboarding inserts and the two onboarding list pages, G019 the "+
			"subject lock, the two id-keyed binding writes, the owner count, the two member pages, the candidate page, the "+
			"invitation insert and revoke, the invitation read and existence check, the two invitation pages, the pending count "+
			"and the two claim calls, the sole-owner scope read, the own-Organizations page and (F-01 a) the bootstrap owner "+
			"status lock", n, writes)
	}
	if stmts["sqlLockOwnerStatus"] != ownerStatusLock {
		t.Errorf("sqlLockOwnerStatus differs from its pinned text: %q", stmts["sqlLockOwnerStatus"])
	}
	// G007: a locking read is a SELECT of a pinned statement, takes FOR SHARE or FOR NO KEY UPDATE only, as its last
	// clause, and names its Organization (organization_id = $1, or id = $1 for the Organization row itself).
	lockClause := regexp.MustCompile(`\bFOR (UPDATE|NO KEY UPDATE|SHARE|KEY SHARE)\b`)
	var locking []string
	for name, sql := range stmts {
		if !strings.HasPrefix(name, "sql") || !lockClause.MatchString(sql) {
			continue
		}
		locking = append(locking, name)
		m := lockClause.FindStringSubmatch(sql)
		if !strings.HasPrefix(sql, "SELECT ") || (m[1] != "SHARE" && m[1] != "NO KEY UPDATE") || !strings.HasSuffix(sql, " FOR "+m[1]) {
			t.Errorf("%s is not a plain locking read (FOR SHARE or FOR NO KEY UPDATE last): %s", name, sql)
		}
		if !strings.Contains(sql, "organization_id = $1") && !strings.Contains(sql, "FROM app.organizations WHERE id = $1") &&
			(name != "sqlLockOwnerStatus" || sql != ownerStatusLock) {
			t.Errorf("%s locks without the Organization predicate", name)
		}
	}
	slices.Sort(locking)
	if strings.Join(locking, ",") != strings.Join(lockingStatements, ",") {
		t.Fatalf("locking statements:\n got %v\nwant %v", locking, lockingStatements)
	}
	// The three membership-fact locks exist with exactly their pinned text (Codex review 432d5f17 §1).
	for name, want := range membershipLocks {
		if stmts[name] != want {
			t.Errorf("%s differs from its pinned text:\n got %s\nwant %s", name, stmts[name], want)
		}
	}
	// Every statement lives in store.go: no other production file holds a string literal naming a schema object.
	for name, f := range files {
		if name == "store.go" || name == "objects.go" || name == "branding.go" || name == "labels.go" {
			continue // labels.go holds the G020 label statements (counted above; TestLabelStatements); objects.go holds the G009 stored-object statements (TestObjectSQLBoundary), branding.go the G017 branding statements (TestBrandingSQLBoundary)
		}
		ast.Inspect(f, func(n ast.Node) bool {
			if lit, ok := n.(*ast.BasicLit); ok && lit.Kind == token.STRING && !logOperationNames[lit.Value] &&
				regexp.MustCompile(`(?i)\bapp\.|\b(SELECT|INSERT|UPDATE|DELETE)\b`).MatchString(lit.Value) {
				t.Errorf("%s holds SQL outside store.go: %s", name, lit.Value)
			}
			return true
		})
	}
	for name, sql := range stmts {
		if (strings.Contains(sql, "_memberships") || strings.Contains(sql, "_access_paths")) && write.MatchString(sql) {
			t.Errorf("%s writes a projection", name)
		}
	}
}

// membershipLocks are the only statements that may read role_bindings directly (Plan r1 §11 fallback; Codex review
// 432d5f17 §1): the locking forms of the membership facts, pinned by name and exact text, because a row lock through a
// SELECT-only projection is refused. Each keeps the Organization and subject predicates, the projection's scope_type and
// membership-source filters, the candidate filter and order where applicable, and ends in FOR SHARE.
var membershipLocks = map[string]string{
	"sqlFactOrganizationBindingLock": "SELECT organization_id, role FROM app.role_bindings WHERE organization_id = $1 AND subject_user_id = $2 AND scope_type = 'organization' AND source = 'membership' FOR SHARE",
	"sqlFactGroupBindingsLock":       "SELECT organization_id, group_id, role FROM app.role_bindings WHERE organization_id = $1 AND subject_user_id = $2 AND scope_type = 'group' AND source = 'membership' AND group_id = ANY ($3) ORDER BY group_id FOR SHARE",
	"sqlFactAllGroupBindingsLock":    "SELECT organization_id, group_id, role FROM app.role_bindings WHERE organization_id = $1 AND subject_user_id = $2 AND scope_type = 'group' AND source = 'membership' ORDER BY group_id FOR SHARE",
}

// subjectLocks are G019's direct role_bindings reads (Plan r1 §1.4): the Subject fact, pinned by name and exact text
// (G060 amendment01 removed the Organization-departure lock: the authority guard cascades the Group bindings). It reads
// the runtime's eight granted columns only (never created_at), keeps the Organization, subject, scope and
// membership-source predicates, orders by id and ends in FOR NO KEY UPDATE (admitted by role_bindings_update's USING).
var subjectLocks = map[string]string{
	"sqlSubjectBindingsLock": "SELECT id, organization_id, subject_user_id, scope_type, scope_id, group_id, role, source FROM app.role_bindings " +
		"WHERE organization_id = $1 AND subject_user_id = ANY ($2) AND scope_type = $3 AND scope_id = $4 AND source = 'membership' " +
		"ORDER BY id FOR NO KEY UPDATE",
}

// allowedDirectRead reports whether a statement may read role_bindings directly: only a pinned membership lock or a
// pinned G019 subject lock with exactly its pinned text.
func allowedDirectRead(name, sql string) bool {
	if want, ok := membershipLocks[name]; ok {
		return sql == want
	}
	want, ok := subjectLocks[name]
	return ok && sql == want
}

// G019: the subject-lock exception's negative controls.
func TestSubjectLockException(t *testing.T) {
	_, files := productionFiles(t)
	stmts := constStrings(t, files["store.go"])
	for name, want := range subjectLocks {
		if stmts[name] != want {
			t.Errorf("%s differs from its pinned text:\n got %s\nwant %s", name, stmts[name], want)
		}
		if strings.Contains(want, "created_at") {
			t.Errorf("%s reads the ungranted created_at", name)
		}
	}
	s := subjectLocks["sqlSubjectBindingsLock"]
	for what, c := range map[string]struct{ name, sql string }{
		"unpinned name":         {"sqlListAllBindings", s},
		"missing tenant filter": {"sqlSubjectBindingsLock", strings.Replace(s, "organization_id = $1 AND ", "", 1)},
		"missing source filter": {"sqlSubjectBindingsLock", strings.Replace(s, " AND source = 'membership'", "", 1)},
		"missing scope filter":  {"sqlSubjectBindingsLock", strings.Replace(s, " AND scope_type = $3 AND scope_id = $4", "", 1)},
		"FOR UPDATE":            {"sqlSubjectBindingsLock", strings.Replace(s, "FOR NO KEY UPDATE", "FOR UPDATE", 1)},
		"no lock":               {"sqlSubjectBindingsLock", strings.TrimSuffix(s, " FOR NO KEY UPDATE")},
		"created_at":            {"sqlSubjectBindingsLock", strings.Replace(s, "role, source", "role, source, created_at", 1)},
	} {
		if allowedDirectRead(c.name, c.sql) {
			t.Errorf("%s: accepted by the exception", what)
		}
	}
}

// fixturePrimitives are the fixture-only Service methods of fixture.go (G019 Plan r1 §1.4 C2, Q4).
var fixturePrimitives = map[string]bool{
	"CreateOrganization": true, "GrantOrganizationRole": true, "ChangeOrganizationRole": true, "RemoveOrganizationMembership": true,
	"GrantGroupRole": true, "ChangeGroupRole": true, "RemoveGroupMembership": true,
}

// fixtureCallerFiles is the recorded call-site inventory of the fixture-only primitives (paths relative to the core
// module root): test files only. A production caller, or an unrecorded test caller, fails TestTenancyFixturePrimitives.
var fixtureCallerFiles = []string{
	"internal/mcp/env_integration_test.go",
	"internal/mcp/mcp_integration_test.go",
	"internal/migrate/business_profile_integration_test.go",
	"internal/migrate/policy_integration_test.go",
	"internal/migrate/tenancy_integration_test.go",
	"internal/tenancy/access/attack_integration_test.go",
	"internal/tenancy/access/audit_integration_test.go",
	"internal/tenancy/access/integration_test.go",
	"internal/tenancy/access/mcp_integration_test.go",
	"internal/tenancy/access/objects_integration_test.go",
	"internal/tenancy/access/objects_ordering_integration_test.go",
	"internal/tenancy/access/ordering_integration_test.go",
	"internal/tenancy/access/recovery_integration_test.go",
	"internal/tenancy/access/rls_ordering_integration_test.go",
	"internal/tenancy/business_profile_integration_test.go",
	"internal/tenancy/integration_test.go",
	"internal/tenancy/tenancy_test.go",
}

// TestTenancyFixturePrimitives (G019 Plan r1 §1.4 C2): the ownerless CreateOrganization and the unguarded
// Grant*/Change*/Remove* writers are declared in fixture.go only, and across the core module they are called from the
// recorded test files only — never from a production file.
func TestTenancyFixturePrimitives(t *testing.T) {
	_, files := productionFiles(t)
	for name, f := range files {
		for _, d := range f.Decls {
			if fn, ok := d.(*ast.FuncDecl); ok && fn.Recv != nil && fixturePrimitives[fn.Name.Name] && name != "fixture.go" {
				t.Errorf("%s declares the fixture-only %s", name, fn.Name.Name)
			}
		}
	}
	root := "../.."
	var callers []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}
		f, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			return err
		}
		called := false
		ast.Inspect(f, func(n ast.Node) bool {
			if c, ok := n.(*ast.CallExpr); ok {
				if sel, ok := c.Fun.(*ast.SelectorExpr); ok && fixturePrimitives[sel.Sel.Name] {
					called = true
				}
			}
			return true
		})
		rel, _ := filepath.Rel(root, path)
		rel = filepath.ToSlash(rel)
		switch {
		case !called || rel == "internal/tenancy/fixture.go":
		case !strings.HasSuffix(rel, "_test.go"):
			t.Errorf("production file %s calls a fixture-only tenancy primitive", rel)
		default:
			callers = append(callers, rel)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	slices.Sort(callers)
	if strings.Join(callers, ",") != strings.Join(fixtureCallerFiles, ",") {
		t.Fatalf("fixture-primitive test callers:\n got %v\nwant %v", callers, fixtureCallerFiles)
	}
}

// The exception's negative controls: another statement name, a missing tenant, subject, scope or source filter, another
// scope, a changed or missing lock mode, a missing candidate filter or order, and other columns all fail it.
func TestMembershipLockException(t *testing.T) {
	for name, sql := range membershipLocks {
		if !allowedDirectRead(name, sql) {
			t.Fatalf("%s: the pinned text is refused", name)
		}
	}
	org, grp := membershipLocks["sqlFactOrganizationBindingLock"], membershipLocks["sqlFactGroupBindingsLock"]
	for what, c := range map[string]struct{ name, sql string }{
		"another statement name":   {"sqlFactOrganizationBinding", org},
		"unpinned name":            {"sqlListAllBindings", org},
		"missing tenant filter":    {"sqlFactOrganizationBindingLock", strings.Replace(org, "organization_id = $1 AND ", "", 1)},
		"missing subject filter":   {"sqlFactOrganizationBindingLock", strings.Replace(org, " AND subject_user_id = $2", "", 1)},
		"missing scope filter":     {"sqlFactOrganizationBindingLock", strings.Replace(org, " AND scope_type = 'organization'", "", 1)},
		"missing source filter":    {"sqlFactOrganizationBindingLock", strings.Replace(org, " AND source = 'membership'", "", 1)},
		"another scope":            {"sqlFactOrganizationBindingLock", strings.Replace(org, "'organization'", "'group'", 1)},
		"FOR UPDATE":               {"sqlFactOrganizationBindingLock", strings.Replace(org, "FOR SHARE", "FOR UPDATE", 1)},
		"FOR NO KEY UPDATE":        {"sqlFactGroupBindingsLock", strings.Replace(grp, "FOR SHARE", "FOR NO KEY UPDATE", 1)},
		"no lock":                  {"sqlFactGroupBindingsLock", strings.TrimSuffix(grp, " FOR SHARE")},
		"missing candidate filter": {"sqlFactGroupBindingsLock", strings.Replace(grp, " AND group_id = ANY ($3)", "", 1)},
		"missing order":            {"sqlFactGroupBindingsLock", strings.Replace(grp, " ORDER BY group_id", "", 1)},
		"other columns":            {"sqlFactGroupBindingsLock", strings.Replace(grp, "SELECT organization_id, group_id, role", "SELECT *", 1)},
	} {
		if allowedDirectRead(c.name, c.sql) {
			t.Errorf("%s: accepted by the exception", what)
		}
	}
}

// writeTargets are the relations a statement may write (G005 and G006; G019 adds invitations: the keyed insert and the
// revocation, the only two runtime writes of app.invitations). G058's request slice adds enrollment_requests (the
// manager's decision and revocation; the applicant's insert happens inside the enrollment_request body) and
// enrollment_sessions (the applicant's abandonment of its own credential). G020 adds label_definitions and
// label_assignments (labels.go).
var writeTargets = map[string]bool{"organizations": true, "groups": true, "role_bindings": true, "business_profiles": true,
	"group_business_profiles": true, "invitations": true, "enrollment_requests": true, "enrollment_sessions": true,
	"label_definitions": true, "label_assignments": true}

// logOperationNames are the catalogued log operation literals of log.go that contain an SQL keyword as a word
// (business_profile.update); they are event values, not SQL (G006 Plan r2 §4.6).
var logOperationNames = map[string]bool{`"business_profile.update"`: true}

// constStrings evaluates the package-level string constants of one file (literals and + concatenations of them).
func constStrings(t *testing.T, f *ast.File) map[string]string {
	t.Helper()
	exprs := map[string]ast.Expr{}
	for _, d := range f.Decls {
		g, ok := d.(*ast.GenDecl)
		if !ok || g.Tok != token.CONST {
			continue
		}
		for _, s := range g.Specs {
			vs := s.(*ast.ValueSpec)
			for i, n := range vs.Names {
				if lit, ok := vs.Values[i].(*ast.BasicLit); ok && lit.Kind != token.STRING {
					continue // G009: a numeric constant (MaxObjectBytes in objects.go) is not a statement
				}
				exprs[n.Name] = vs.Values[i]
			}
		}
	}
	var eval func(ast.Expr) string
	eval = func(e ast.Expr) string {
		switch x := e.(type) {
		case *ast.BasicLit:
			s, err := strconv.Unquote(x.Value)
			if err != nil {
				t.Fatal(err)
			}
			return s
		case *ast.Ident:
			v, ok := exprs[x.Name]
			if !ok {
				t.Fatalf("unknown constant %s", x.Name)
			}
			return eval(v)
		case *ast.BinaryExpr:
			if x.Op != token.ADD {
				t.Fatalf("unsupported operator %s", x.Op)
			}
			return eval(x.X) + eval(x.Y)
		case *ast.ParenExpr:
			return eval(x.X)
		}
		t.Fatalf("unsupported constant expression %T", e)
		return ""
	}
	out := map[string]string{}
	for n, e := range exprs {
		out[n] = eval(e)
	}
	return out
}

// objectStatements are the G009 stored-object statements of objects.go (Plan r1 §2, §4.4): one insert, the plain read,
// its FOR UPDATE lock and the three lifecycle updates. They write app.stored_objects only, name the Organization as
// organization_id = $1, never write the generated key, the timestamps or deleted_by, and the lock is the only locking
// statement (FOR UPDATE, last).
var objectStatements = []string{
	"sqlFinalizeObject", "sqlInsertObject", "sqlLockObject", "sqlMarkObjectDeleted", "sqlMarkObjectDeleting", "sqlSelectObject",
}

func TestObjectSQLBoundary(t *testing.T) {
	_, files := productionFiles(t)
	stmts := constStrings(t, files["objects.go"])
	var names []string
	for name, sql := range stmts {
		if !strings.HasPrefix(name, "sql") {
			continue
		}
		names = append(names, name)
		if !strings.Contains(sql, "app.stored_objects") || regexp.MustCompile(`app\.[a-z_]+`).FindAllString(sql, -1)[0] != "app.stored_objects" ||
			len(regexp.MustCompile(`app\.[a-z_]+`).FindAllString(sql, -1)) != 1 {
			t.Errorf("%s names another relation: %s", name, sql)
		}
		if !strings.HasPrefix(sql, "INSERT INTO app.stored_objects (organization_id, ") && !strings.Contains(sql, "WHERE organization_id = $1 AND id = $2") {
			t.Errorf("%s lacks the Organization and object predicate: %s", name, sql)
		}
		body, _, _ := strings.Cut(sql, " RETURNING ")
		if strings.HasPrefix(sql, "UPDATE ") {
			set, _, _ := strings.Cut(strings.TrimPrefix(body, "UPDATE app.stored_objects SET "), " WHERE ")
			for _, a := range strings.Split(set, ", ") {
				if c, _, _ := strings.Cut(a, " = "); c != "state" && c != "sha256" {
					t.Errorf("%s writes %s", name, c)
				}
			}
		}
		if strings.HasPrefix(sql, "INSERT ") && !strings.HasPrefix(body, "INSERT INTO app.stored_objects (organization_id, id, group_id, creator_user_id, byte_size, content_type) VALUES ") {
			t.Errorf("%s inserts other columns: %s", name, body)
		}
		if regexp.MustCompile(`\bDELETE\b|\bstorage_key =|\b(ready_at|deleting_at|deleted_at|deleted_by|created_at) =`).MatchString(body) {
			t.Errorf("%s deletes or writes a database-owned column: %s", name, body)
		}
		lock := regexp.MustCompile(`\bFOR (UPDATE|NO KEY UPDATE|SHARE|KEY SHARE)\b`)
		if lock.MatchString(sql) && (name != "sqlLockObject" || !strings.HasSuffix(sql, " FOR UPDATE")) {
			t.Errorf("%s takes an unpinned row lock: %s", name, sql)
		}
	}
	slices.Sort(names)
	if strings.Join(names, ",") != strings.Join(objectStatements, ",") {
		t.Fatalf("stored-object statements:\n got %v\nwant %v", names, objectStatements)
	}
	if stmts["sqlLockObject"] != stmts["sqlSelectObject"]+" FOR UPDATE" {
		t.Error("sqlLockObject is not sqlSelectObject FOR UPDATE")
	}
	// No other file of the package names the stored-object relation (errors.go maps its constraint names only), except
	// branding.go's pinned logo statements (TestBrandingSQLBoundary).
	for name, f := range files {
		if name == "objects.go" || name == "branding.go" {
			continue
		}
		ast.Inspect(f, func(n ast.Node) bool {
			if lit, ok := n.(*ast.BasicLit); ok && lit.Kind == token.STRING && strings.Contains(lit.Value, "app.stored_objects") {
				t.Errorf("%s names app.stored_objects", name)
			}
			return true
		})
	}
}

// brandingStatements are the G017 branding statements of branding.go (Plan r2 §5.1): the plain read, its FOR NO KEY
// UPDATE lock, the version-1 insert, the version-guarded update and the relation-free NFC normalization name
// app.organization_branding only (the live logo state is read through objects.go's sqlSelectObject), name the
// Organization as organization_id = $1 and never write organization_id after insert; the two logo statements read
// app.stored_objects only (the new object's FOR SHARE lock and the bounded unreferenced-upload listing), and the public
// read calls the auth reader's app.public_branding projection only. The two locks are the only locking statements.
var brandingStatements = []string{
	"sqlInsertBranding", "sqlListLogoUploads", "sqlLockBranding", "sqlNormalizeLabel", "sqlPublicBranding", "sqlSelectBranding",
	"sqlShareLogoObject", "sqlUpdateBranding",
}

// brandingLogoStatements pins the exact text of the statements outside app.organization_branding.
var brandingLogoStatements = map[string]string{
	"sqlShareLogoObject": "SELECT id, organization_id, group_id, creator_user_id, storage_key, state, byte_size, content_type, sha256, " +
		"created_at, ready_at, deleting_at, deleted_at, deleted_by FROM app.stored_objects WHERE organization_id = $1 AND id = $2 FOR SHARE",
	"sqlListLogoUploads": "SELECT id, content_type, byte_size, state, sha256, count(*) OVER () FROM app.stored_objects " +
		"WHERE organization_id = $1 AND group_id IS NULL AND content_type IN ('image/png', 'image/jpeg') " +
		"AND state IN ('ready', 'deleting') AND id IS DISTINCT FROM $2 ORDER BY created_at, id LIMIT 100",
	"sqlPublicBranding": "SELECT revision, display_label, light, dark, logo_sha256, logo_content_type, logo_size, logo_key " +
		"FROM app.public_branding($1)",
}

func TestBrandingSQLBoundary(t *testing.T) {
	_, files := productionFiles(t)
	stmts := constStrings(t, files["branding.go"])
	var names []string
	relation := regexp.MustCompile(`app\.[a-z_]+`)
	for name, sql := range stmts {
		if !strings.HasPrefix(name, "sql") {
			continue
		}
		names = append(names, name)
		rels := relation.FindAllString(sql, -1)
		if name == "sqlNormalizeLabel" {
			if sql != "SELECT normalize($1::text, NFC)" {
				t.Errorf("sqlNormalizeLabel differs from its pinned text: %s", sql)
			}
			continue
		}
		if want, ok := brandingLogoStatements[name]; ok {
			if sql != want {
				t.Errorf("%s differs from its pinned text:\n got %s\nwant %s", name, sql, want)
			}
			continue
		}
		if len(rels) != 1 || rels[0] != "app.organization_branding" {
			t.Errorf("%s names another relation: %s", name, sql)
		}
		if !strings.HasPrefix(sql, "INSERT INTO app.organization_branding (organization_id, ") && !strings.Contains(sql, "WHERE organization_id = $1") {
			t.Errorf("%s lacks the Organization predicate: %s", name, sql)
		}
		body, _, _ := strings.Cut(sql, " RETURNING ")
		if strings.HasPrefix(sql, "UPDATE ") {
			if !strings.HasSuffix(body, " WHERE organization_id = $1 AND version = $2") {
				t.Errorf("%s is not version-guarded: %s", name, body)
			}
			set, _, _ := strings.Cut(strings.TrimPrefix(body, "UPDATE app.organization_branding SET "), " WHERE ")
			for _, a := range strings.Split(set, ", ") {
				switch c, _, _ := strings.Cut(a, " = "); c {
				case "display_label", "logo_object_id", "light", "dark", "version", "updated_at", "updated_by":
				default:
					t.Errorf("%s writes %s", name, c)
				}
			}
		}
		if strings.HasPrefix(sql, "INSERT ") && !strings.HasPrefix(body, "INSERT INTO app.organization_branding (organization_id, display_label, logo_object_id, light, dark, version, updated_by) VALUES ($1, $2, $3, $4, $5, 1, $6)") {
			t.Errorf("%s inserts other columns or another version: %s", name, body)
		}
		if regexp.MustCompile(`\bDELETE\b`).MatchString(sql) {
			t.Errorf("%s deletes: %s", name, sql)
		}
		lock := regexp.MustCompile(`\bFOR (UPDATE|NO KEY UPDATE|SHARE|KEY SHARE)\b`)
		if lock.MatchString(sql) && (name != "sqlLockBranding" || !strings.HasSuffix(sql, " FOR NO KEY UPDATE")) {
			t.Errorf("%s takes an unpinned row lock: %s", name, sql)
		}
	}
	slices.Sort(names)
	if strings.Join(names, ",") != strings.Join(brandingStatements, ",") {
		t.Fatalf("branding statements:\n got %v\nwant %v", names, brandingStatements)
	}
	if stmts["sqlLockBranding"] != stmts["sqlSelectBranding"]+" FOR NO KEY UPDATE" {
		t.Error("sqlLockBranding is not sqlSelectBranding FOR NO KEY UPDATE")
	}
	// No other file of the package names the branding relation.
	for name, f := range files {
		if name == "branding.go" {
			continue
		}
		ast.Inspect(f, func(n ast.Node) bool {
			if lit, ok := n.(*ast.BasicLit); ok && lit.Kind == token.STRING && strings.Contains(lit.Value, "app.organization_branding") {
				t.Errorf("%s names app.organization_branding", name)
			}
			return true
		})
	}
}
