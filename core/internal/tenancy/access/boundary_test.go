package access

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

// U-ACCESS-BOUNDARY (G007 Plan r1 §6.1, §3.7 bypass control): the access package's imports, its absence of SQL and
// HTTP, its exact exported surface, the core-wide production importer boundary of tenancy and access, and correlation
// read only for logging.

// The production files of the package (Plan r1 §8; G009 Plan r1 §2 adds the three storage files; G011 adds audit.go;
// G014 adds mcp.go; G016 adds domains.go; G019 adds invitations.go, membership.go, ownership.go and retirement.go for the
// membership, ownership, invitation and retirement operations; G058 adds enrollment.go for the enrollment admission;
// G020 adds labels.go, the guarded label operations, and labels_targets.go, the closed label target registry).
// Sorted, as the comparison is against the sorted names.
var accessFiles = []string{"access.go", "audit.go", "authority.go", "branding.go", "context.go", "domains.go", "enrollment.go", "errors.go", "facade.go", "invitations.go", "labels.go", "labels_targets.go",
	"log.go", "mcp.go", "membership.go", "objects.go", "objects_io.go", "objects_policy.go", "onboarding.go", "ownership.go", "policy.go",
	"retirement.go", "tx.go"}

// Imports a production file may use.
var accessImports = map[string]bool{
	"context": true, "errors": true, "log/slog": true, "sync/atomic": true, "time": true,
	"github.com/jackc/pgx/v5": true, "github.com/jackc/pgx/v5/pgconn": true,
	"core/internal/identity": true, "core/internal/tenancy": true, "core/internal/observability": true,
	// G009: the owned object store and the streaming, digest and identifier primitives of the storage operations.
	"core/internal/objectstore": true, "crypto/rand": true, "crypto/sha256": true, "hash": true, "io": true, "strings": true,
	"sync": true,
	// G011: the durable audit boundary (its SQL lives there, never here).
	"core/internal/audit": true,
	// G016: the domain model and lifecycle, its generated query package (the SQL lives there) and the hostname contract.
	"core/internal/domains": true, "core/internal/domains/db": true, "core/internal/hostname": true, "encoding/base32": true,
	// G016 deviation #3 closure: the opaque claim fence drawn by crypto/rand.Int over [0, math.MaxInt64).
	"math": true, "math/big": true,
	// G017: the pure branding validation (no I/O, no SQL), the hex logo revision and the in-memory normalized logo body.
	"core/internal/branding": true, "encoding/hex": true, "bytes": true,
}

// The pinned exported surface: BeginRequest, Resolve, the eight guarded operations and the context accessors, plus the
// types and error codes their signatures need. No evaluator, allow result, capability check or capability value is
// exported; the request scope and its key are unexported.
var (
	// G014 (Plan r8 §9): the MCP connection operations (BeginVerifiedRequest, activation, the context check, the context
	// read, the Group page and selection, setup create and list, and the resource_mismatch record) and their types.
	// G015 (Plan r4 §5.1, D-2): the opaque ToolRequirement, its four constructors and the names-only AuthorizeCatalog
	// discovery; no evaluator, allow result or per-target check (no AuthorizeTool).
	pinnedExportedDecls = []string{
		"const AcceptConsumed", "const CodeConflict", "const CodeContextDenied", "const CodeContrastFailed",
		"const CodeDomainRetryExhausted", "const CodeDomainStateInvalid", "const CodeForbidden",
		"const CodeGroupNotSelectable", "const CodeHostTaken", "const CodeInternal", "const CodeInvalidInput",
		"const CodeNotFound", "const CodeUnauthenticated", "const CodeUnavailable", "const DecisionApprove", "const DecisionDeny", "const DeleteOutcomeDeleted",
		"const DeleteOutcomeDeletionPending", "const EnrollmentStatusAdmitted", "const EnrollmentStatusNone", "const MaxRetirementBlockers", "const MemberFilterOwnerCandidates",
		"func CodeOf", "func NewService", "func NewServiceWithObjects", "func ParseBusinessProfileID", "func ParseGroupID",
		"func RequireBinding", "func RequireCapability", "func RequireContext", "func RequireMembership",
		"type AcceptOutcome", "type AuditCursor", "type AuditEvent", "type AuditPage", "type AuditReadRequest",
		"type BrandingLogo", "type BrandingUpdate", "type BrandingView", "type BusinessContext", "type BusinessProfileID",
		"type BusinessProfileSummary", "type Candidate", "type CandidatePage", "type Capability", "type Code",
		"type ConnectionContext", "type DeleteOutcome", "type DeleteResult", "type DomainClaim", "type DomainView",
		"type EnrollmentAdmission", "type EnrollmentPublic", "type EnrollmentRequestPage", "type EnrollmentRequestView", "type EnrollmentStatus", "type Error", "type Facts", "type GroupAuthorityView", "type GroupID", "type GroupOnboarding", "type GroupPage", "type Invitation",
		"type InvitationPage", "type IssuedInvitation",
		// G020 (Plan r2 §4.1): the label facade's selector, views, page and delete result.
		"type LabelAssignmentView", "type LabelDefinitionPage", "type LabelDefinitionView", "type LabelDeleteResult",
		"type LabelTargetSelector", "type Member", "type MemberPage", "type ObjectConfig",
		"type ObjectInfo", "type OnboardingGroupPage", "type ProfilePage", "type PublicBrandingView", "type PublicLogo",
		"type PutObjectRequest", "type Reason", "type Recipient", "type ResolveRequest", "type RetirementBlockers",
		"type RetirementProgress",
		"type Scope", "type ScopeRef", "type SelectableGroup", "type SelfMember", "type Service",
		"type SignedGetURLRequest", "type SignedURL", "type Target", "type TargetKind", "type TenantContext",
		"type ToolRequirement",
		// G058 slice 1b: the enrollment policy write's open-readiness refusal (addendum §A.5).
		"var ErrEnrollmentNotReady",
	}
	pinnedExportedMethods = []string{
		"Error.Error", "Error.Is", "Service.AbandonEnrollment", "Service.AcceptGroupInvitation", "Service.ActivateConnection", "Service.AdmitAtCallback",
		"Service.AdmitInvitationAtCallback", "Service.AdmitOpen", "Service.AdmitRequest", "Service.AssignLabel", "Service.AuthorizeCatalog",
		"Service.AuthorizeConnection", "Service.BeginDomainCheck", "Service.BeginRequest", "Service.BeginVerifiedRequest",
		"Service.ChangeGroupSlug", "Service.ChangeMemberRole", "Service.ClaimDomain",
		"Service.CompleteGroupOnboardingWithNewProfile", "Service.CompleteGroupOnboardingWithProfile",
		"Service.ConsumeOrganizationInvitation", "Service.CreateGroup", "Service.CreateLabelDefinition", "Service.CreateMCPConnection",
		"Service.DecideEnrollmentRequest", "Service.DeleteBrandingUpload", "Service.DeleteLabelDefinition", "Service.DeleteObject", "Service.DemoteOwner",
		"Service.EnrollmentStatus", "Service.GetBranding",
		"Service.GetBusinessProfile", "Service.GetBusinessProfileView", "Service.GetConnectionContext", "Service.GetEnrollmentPolicy", "Service.GetGroup",
		"Service.GetGroupAuthority", "Service.GetGroupOnboarding", "Service.GetInvitation", "Service.GetObject", "Service.GetOrganization",
		"Service.GetOrganizationPolicy", "Service.IssueInvitation", "Service.LeaveScope", "Service.ListBusinessProfiles",
		"Service.ListDomains", "Service.ListEnrollmentRequests", "Service.ListGroups", "Service.ListInvitationCandidates", "Service.ListInvitations",
		"Service.ListLabelAssignments", "Service.ListLabelDefinitions",
		"Service.ListMCPConnections", "Service.ListMembers", "Service.ListMyInvitations", "Service.ListSelectableGroups",
		"Service.ObjectMetadata", "Service.PromoteToOwner", "Service.PublicBranding", "Service.PublicBrandingLogo", "Service.PublicEnrollment",
		"Service.PutObject", "Service.ReadAuditEvents", "Service.RecordDomainCheck", "Service.RecordResourceMismatch",
		"Service.RecordWithdrawal", "Service.RemoveBrandingLogo", "Service.RemoveDomain", "Service.RemoveMember",
		"Service.RenameGroup", "Service.RenameMCPConnection", "Service.RequestEnrollment", "Service.ResetBranding", "Service.Resolve", "Service.RetireUser", "Service.RetirementEligibility",
		"Service.RetryWithdrawal", "Service.RevokeEnrollmentApproval", "Service.RevokeInvitation", "Service.RotateChallenge", "Service.SelectConnectionGroup",
		"Service.SelfMembership", "Service.SetBrandingLogo", "Service.SetGroupManagementOverride", "Service.SetOrganizationPolicy",
		"Service.SetOrganizationSlug", "Service.SignedGetURL", "Service.SuggestGroupSlug", "Service.TransferOwnership", "Service.UnassignLabel",
		"Service.UpdateBranding", "Service.UpdateBusinessProfile", "Service.UpdateEnrollmentPolicy", "Service.UpdateLabelDefinition", "SignedURL.GoString", "SignedURL.LogValue",
		"SignedURL.String", "SignedURL.URL", "TenantContext.Group", "TenantContext.Organization", "TenantContext.Subject",
	}
)

func accessProductionFiles(t *testing.T) map[string]*ast.File {
	t.Helper()
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	fset := token.NewFileSet()
	files := map[string]*ast.File{}
	var names []string
	for _, e := range entries {
		n := e.Name()
		if e.IsDir() || !strings.HasSuffix(n, ".go") || strings.HasSuffix(n, "_test.go") {
			continue
		}
		f, err := parser.ParseFile(fset, n, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		files[n] = f
		names = append(names, n)
	}
	slices.Sort(names)
	if strings.Join(names, ",") != strings.Join(accessFiles, ",") {
		t.Fatalf("production files: %v, want %v", names, accessFiles)
	}
	return files
}

func recvName(e ast.Expr) string {
	if s, ok := e.(*ast.StarExpr); ok {
		e = s.X
	}
	if id, ok := e.(*ast.Ident); ok {
		return id.Name
	}
	return ""
}

func TestAccessBoundary(t *testing.T) {
	files := accessProductionFiles(t)
	sqlLike := regexp.MustCompile(`\bapp\.|\b(SELECT|INSERT INTO|UPDATE|DELETE FROM|FOR SHARE|FOR NO KEY UPDATE)\b`)
	var decls, methods []string
	for name, f := range files {
		for _, imp := range f.Imports {
			p, _ := strconv.Unquote(imp.Path.Value)
			if !accessImports[p] {
				t.Errorf("%s imports %s", name, p)
			}
			if p == "core/internal/observability" && name != "log.go" {
				t.Errorf("%s imports observability: correlation is read only for logging (log.go)", name)
			}
		}
		ast.Inspect(f, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.BasicLit:
				if x.Kind == token.STRING && sqlLike.MatchString(x.Value) {
					t.Errorf("%s holds an SQL literal: %s", name, x.Value)
				}
			case *ast.SelectorExpr:
				if id, ok := x.X.(*ast.Ident); ok && id.Name == "observability" &&
					(name != "log.go" || (x.Sel.Name != "RequestID" && x.Sel.Name != "TraceID")) {
					t.Errorf("%s uses observability.%s", name, x.Sel.Name)
				}
			}
			return true
		})
		for _, d := range f.Decls {
			switch x := d.(type) {
			case *ast.FuncDecl:
				if x.Name.Name == "ServeHTTP" {
					t.Errorf("%s declares ServeHTTP", name)
				}
				if !x.Name.IsExported() {
					continue
				}
				if x.Recv == nil {
					decls = append(decls, "func "+x.Name.Name)
				} else if r := recvName(x.Recv.List[0].Type); ast.IsExported(r) {
					// Methods of unexported types (the transaction wrappers) are not reachable from another package.
					methods = append(methods, r+"."+x.Name.Name)
				}
			case *ast.GenDecl:
				for _, s := range x.Specs {
					switch sp := s.(type) {
					case *ast.TypeSpec:
						if sp.Name.IsExported() {
							decls = append(decls, "type "+sp.Name.Name)
						}
						if sp.Name.Name == "TenantContext" || sp.Name.Name == "Service" {
							st := sp.Type.(*ast.StructType)
							for _, fd := range st.Fields.List {
								for _, n := range fd.Names {
									if n.IsExported() {
										t.Errorf("%s has an exported field %s", sp.Name.Name, n.Name)
									}
								}
							}
						}
					case *ast.ValueSpec:
						for _, n := range sp.Names {
							if n.IsExported() {
								decls = append(decls, x.Tok.String()+" "+n.Name)
							}
						}
					}
				}
			}
		}
	}
	slices.Sort(decls)
	slices.Sort(methods)
	if strings.Join(decls, ",") != strings.Join(pinnedExportedDecls, ",") {
		t.Errorf("exported declarations:\n got %v\nwant %v", decls, pinnedExportedDecls)
	}
	if strings.Join(methods, ",") != strings.Join(pinnedExportedMethods, ",") {
		t.Errorf("exported methods:\n got %v\nwant %v", methods, pinnedExportedMethods)
	}
	attestCallers(t)
}

// attestCallers is G019's verified-address pin (Plan r2 addendum §2): no non-test Go file of core outside
// core/internal/identity/ names AttestVerifiedAddress, so provider address evidence is produced only by identity.
func attestCallers(t *testing.T) {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err != nil {
		t.Fatalf("core module root not found at %s: %v", root, err)
	}
	identityFiles := 0
	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == "testdata" || strings.HasPrefix(d.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if !strings.Contains(string(b), "AttestVerifiedAddress") {
			return nil
		}
		rel, _ := filepath.Rel(root, path)
		if rel = filepath.ToSlash(rel); strings.HasPrefix(rel, "internal/identity/") {
			identityFiles++
			return nil
		}
		t.Errorf("%s names AttestVerifiedAddress (only core/internal/identity/ may; G019 r2 §2)", rel)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if identityFiles == 0 {
		t.Error("AttestVerifiedAddress not found under core/internal/identity/: the pin is vacuous")
	}
}

// The production importer boundary (Plan r1 §3.7): across every non-test Go file of core, only this package imports
// core/internal/tenancy, and only core/internal/mcp imports core/internal/tenancy/access (G014 Plan r8 §7; nothing did in
// G007). B-MCP-SOURCE pins the server's linkage.
func TestProductionImporters(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err != nil {
		t.Fatalf("core module root not found at %s: %v", root, err)
	}
	fset := token.NewFileSet()
	n := 0
	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == "testdata" || strings.HasPrefix(d.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		f, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}
		n++
		rel, _ := filepath.Rel(root, path)
		dir := filepath.ToSlash(filepath.Dir(rel))
		for _, imp := range f.Imports {
			p, _ := strconv.Unquote(imp.Path.Value)
			switch {
			case p == "core/internal/tenancy/access" && dir != "internal/mcp" && dir != "internal/domains/api" && dir != "internal/branding/api" && dir != "internal/onboarding/api" && dir != "internal/membership/api" && dir != "internal/enrollment/api" && dir != "internal/labels/api":
				t.Errorf("%s imports the access package (only core/internal/mcp, core/internal/domains/api, core/internal/branding/api, core/internal/onboarding/api, core/internal/membership/api, core/internal/enrollment/api and core/internal/labels/api may; G014, G016, G017, G018, G019, G058, G020)", rel)
			case p == "core/internal/labels/api" && dir != "cmd/server":
				t.Errorf("%s imports core/internal/labels/api (only cmd/server may; G020)", rel)
			case p == "core/internal/enrollment/api" && dir != "cmd/server":
				t.Errorf("%s imports core/internal/enrollment/api (only cmd/server may; G058)", rel)
			case p == "core/internal/onboarding/api" && dir != "cmd/server":
				t.Errorf("%s imports core/internal/onboarding/api (only cmd/server may; G018)", rel)
			case p == "core/internal/membership/api" && dir != "cmd/server":
				t.Errorf("%s imports core/internal/membership/api (only cmd/server may; G019)", rel)
			case p == "core/internal/branding/api" && dir != "cmd/server":
				t.Errorf("%s imports core/internal/branding/api (only cmd/server may; G017)", rel)
			case p == "core/internal/branding" && dir != "internal/tenancy/access":
				t.Errorf("%s imports core/internal/branding (only the access package may; G017)", rel)
			case strings.HasPrefix(p, "core/internal/domains") && (strings.HasPrefix(dir, "internal/identity") || dir == "internal/tenancy"):
				t.Errorf("%s imports %s (identity receives the resolver as an interface it declares itself; G016)", rel, p)
			case strings.HasPrefix(dir, "internal/domains") && dir != "internal/domains/api" &&
				(p == "core/internal/tenancy/access" || strings.HasPrefix(p, "core/internal/identity")):
				t.Errorf("%s imports %s (the domains package never imports access or identity; G016)", rel, p)
			case p == "core/internal/domains/api" && dir != "cmd/server":
				t.Errorf("%s imports core/internal/domains/api (only cmd/server may; G016)", rel)
			case p == "core/internal/tenancy" && dir != "internal/tenancy/access":
				t.Errorf("%s imports core/internal/tenancy (only the access package may)", rel)
			case p == "core/internal/objectstore" && dir != "internal/tenancy/access" && dir != "internal/branding/api":
				t.Errorf("%s imports core/internal/objectstore (only the access package and, to build the store from the loaded settings, core/internal/branding/api may; G009, G017)", rel)
			case strings.HasPrefix(p, "github.com/aws/") && dir != "internal/objectstore" && dir != "internal/health" && dir != "cmd/server":
				t.Errorf("%s imports %s (only objectstore, health and cmd/server may use the AWS SDK; G009)", rel, p)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if n < 30 {
		t.Fatalf("walked only %d production files under %s", n, root)
	}
}

// --- U-RLS-BOUNDARY (G008 Plan r3 §8.2) ------------------------------------------------------------------------------

// rlsCoreFiles parses every non-test Go file of core outside testdata, keyed by its slash path relative to core.
func rlsCoreFiles(t *testing.T) (*token.FileSet, map[string]*ast.File) {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	fset := token.NewFileSet()
	files := map[string]*ast.File{}
	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == "testdata" || strings.HasPrefix(d.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		f, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(root, path)
		files[filepath.ToSlash(rel)] = f
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(files) < 30 {
		t.Fatalf("parsed only %d production files", len(files))
	}
	return fset, files
}

// rlsConsts evaluates the string constants of one file (literals and + concatenations of them).
func rlsConsts(f *ast.File) map[string]string {
	out := map[string]string{}
	var eval func(e ast.Expr) (string, bool)
	eval = func(e ast.Expr) (string, bool) {
		switch x := e.(type) {
		case *ast.BasicLit:
			if x.Kind != token.STRING {
				return "", false
			}
			s, err := strconv.Unquote(x.Value)
			return s, err == nil
		case *ast.BinaryExpr:
			a, ok1 := eval(x.X)
			b, ok2 := eval(x.Y)
			return a + b, ok1 && ok2 && x.Op == token.ADD
		case *ast.Ident:
			s, ok := out[x.Name]
			return s, ok
		case *ast.ParenExpr:
			return eval(x.X)
		}
		return "", false
	}
	for _, d := range f.Decls {
		g, ok := d.(*ast.GenDecl)
		if !ok || g.Tok != token.CONST {
			continue
		}
		for _, s := range g.Specs {
			vs := s.(*ast.ValueSpec)
			for i, n := range vs.Names {
				if i < len(vs.Values) {
					if v, ok := eval(vs.Values[i]); ok {
						out[n.Name] = v
					}
				}
			}
		}
	}
	return out
}

// funcDecl returns the named top-level function or method of f.
func funcDecl(f *ast.File, name string) *ast.FuncDecl {
	for _, d := range f.Decls {
		if fd, ok := d.(*ast.FuncDecl); ok && fd.Name.Name == name {
			return fd
		}
	}
	return nil
}

// calledNames lists, in source order, the called function or method names in n (the selector's name for x.f()).
func calledNames(n ast.Node) []string {
	var out []string
	ast.Inspect(n, func(n ast.Node) bool {
		if c, ok := n.(*ast.CallExpr); ok {
			switch fn := c.Fun.(type) {
			case *ast.Ident:
				out = append(out, fn.Name)
			case *ast.SelectorExpr:
				out = append(out, fn.Sel.Name)
			}
		}
		return true
	})
	return out
}

var rlsProtectedTable = regexp.MustCompile(`\bapp\.(users|sessions|identity_principals|login_transactions)\b`)

// The identity store functions that issue statements on the four protected identity tables, and how each one is
// contextualized. "init" functions call an initializer before their first protected statement; "callee" functions run
// only inside a caller's initialized transaction (issue: signIn after its initializer; liveSession: the facade's
// tenant transaction, F10); "reader" functions use only the E1-E3 reader calls.
var identityProtectedFuncs = map[string]string{
	"insertLogin": "init", "consumeLogin": "init", "resolve": "init", "signIn": "init", "createMapping": "init",
	"revokeSession": "init", "revokeAll": "init", "recordProviderLogout": "init", "consumeLogoutRef": "init",
	"DisableUser": "init", "EnableUser": "init", "issue": "callee", "liveSession": "callee", "lookupUser": "reader",
	"principalSubject": "reader",
	// G062: the root signup callback's T1 (the mapping and the active-status check, then the signup row).
	"signupT1": "init",
}

// The exact initializer calls per identity function. Two initializations in one transaction occur only at the two
// pinned sites: createMapping (F2's savepoint, then the winner after ROLLBACK TO SAVEPOINT) and insertLogin (F3's
// EMPTY then LOGIN; G016: insertTransaction and consumeTransaction delegate to insertLogin and consumeLogin, which carry
// the tenant-host admission and consumption check). signIn's two calls are exclusive branches (continuity, mapped subject).
var identityInitCalls = map[string]string{
	"insertLogin": "initEmpty,initLogin", "consumeLogin": "initLogin", "resolve": "initIdentity",
	"signIn": "initIdentity,initIdentity", "createMapping": "initIdentity,initIdentity", "revokeSession": "initIdentity",
	"revokeAll": "initIdentity", "recordProviderLogout": "initIdentity", "consumeLogoutRef": "initIdentity",
	"DisableUser": "initIdentity", "EnableUser": "initIdentity",
	// G011: the bounded independent recorder (an attributed subject's IDENTITY, or EMPTY for an unattributed event).
	"recordAuditEvents": "initIdentity,initEmpty",
	// G014 (mcp_store.go): the MCP binding read (TENANT-ORG of the mapped subject in the registry Organization) and the
	// M5 renewal (IDENTITY).
	"readMCPBinding": "initTenantOrg", "renewMCPSession": "initIdentity",
	// G062: T1 (an existing mapped subject's IDENTITY; a new principal's goes through createMapping), and the two signup
	// shape helpers: the IDENTITY shape of a resolved signup credential's user and the LOGIN shape of a continuation hop.
	// Neither helper names a protected identity table.
	"signupT1": "initIdentity", "signupIdentityTx": "initIdentity", "loginTx": "initLogin",
}

// G014: the functions of internal/identity/mcp_store.go that issue statements on the protected identity tables. The
// binding read and the renewal open their own initialized transactions; the facade helpers (callee) run only inside the
// access facade's tenant-shaped transaction, whose context statement is first.
var identityMCPProtectedFuncs = map[string]string{
	"readMCPBinding": "init", "renewMCPSession": "init", "InsertMCPConnectionTx": "callee", "ListMCPConnectionsTx": "callee",
	"UserDisplayNameTx": "callee",
}

// G019 F-01 (refinement e): the separate per-file inventory of internal/identity/retire.go. Only the finalization
// transaction issues statements on the protected identity tables (users FOR UPDATE, then DisableUser's two UPDATEs),
// after its IDENTITY initializer; the exported wrappers and the guarded check view issue none. The store.go and
// mcp_store.go maps above are not widened.
var (
	identityRetireProtectedFuncs = map[string]string{"finalizeRetirement": "init"}
	identityRetireInitCalls      = map[string]string{"finalizeRetirement": "initIdentity"}
)

func TestRLSBoundary(t *testing.T) {
	fset, files := rlsCoreFiles(t)
	tenancyStore, identityStore := files["internal/tenancy/store.go"], files["internal/identity/store.go"]
	if tenancyStore == nil || identityStore == nil {
		t.Fatal("tenancy or identity store.go not found")
	}
	tc, ic := rlsConsts(tenancyStore), rlsConsts(identityStore)

	t.Run("one context statement, transaction-local", func(t *testing.T) {
		a, b := tc["sqlInitContext"], ic["sqlInitContext"]
		if a == "" || a != b {
			t.Fatalf("sqlInitContext differs between tenancy and identity (or is missing)")
		}
		// G011: 7 set_config calls, all transaction-local: the 4 guarded app.* fields, the 2 unguarded audit correlation
		// settings, and the constant reset of the revocation gap marker.
		local := regexp.MustCompile(`set_config\('app\.[a-z_]+', \$[1-4], true\)`)
		corr := regexp.MustCompile(`set_config\('audit\.(request_id|trace_id)', \$[56], true\)`)
		if n := strings.Count(a, "set_config("); n != 7 || len(local.FindAllString(a, -1)) != 4 || len(corr.FindAllString(a, -1)) != 2 ||
			strings.Count(a, "set_config('audit.revocation_gap', '', true)") != 1 {
			t.Errorf("sqlInitContext: %d set_config calls, want 7 (4 app, 2 correlation, 1 constant gap reset), all transaction-local (true)", n)
		}
		for _, f := range []string{"audit.request_id", "audit.trace_id", "audit.revocation_gap"} {
			if strings.Count(a, "'"+f+"'") != 1 {
				t.Errorf("sqlInitContext must set %s exactly once and never guard it (it confers no authority)", f)
			}
		}
		for _, f := range []string{"app.user_id", "app.organization_id", "app.group_id", "app.login_binding"} {
			if strings.Count(a, "'"+f+"'") != 2 {
				t.Errorf("sqlInitContext must set and guard %s exactly once each", f)
			}
		}
		// set_config appears nowhere else in core's production Go.
		for path, f := range files {
			consts := rlsConsts(f)
			ast.Inspect(f, func(n ast.Node) bool {
				lit, ok := n.(*ast.BasicLit)
				if !ok || lit.Kind != token.STRING || !strings.Contains(lit.Value, "set_config") {
					return true
				}
				s, _ := strconv.Unquote(lit.Value)
				if (path != "internal/tenancy/store.go" && path != "internal/identity/store.go") || consts["sqlInitContext"] != s {
					t.Errorf("%s: set_config outside the pinned sqlInitContext (%s)", path, fset.Position(lit.Pos()))
				}
				return true
			})
		}
	})

	t.Run("forbidden production SQL", func(t *testing.T) {
		bad := regexp.MustCompile(`(?i)\bSET\s+(LOCAL\s+|SESSION\s+)?app\.|\bRESET\s+(ALL|ROLE|SESSION|app\.)|row_security|\bSET\s+(LOCAL\s+|SESSION\s+)?ROLE\s+[^=\s]|BYPASSRLS|SECURITY\s+DEFINER`)
		for path, f := range files {
			if strings.HasPrefix(path, "internal/migrate/") {
				continue // T1: the migration runner assumes the owner role by design (docs/foundation/MIGRATIONS.md)
			}
			ast.Inspect(f, func(n ast.Node) bool {
				if lit, ok := n.(*ast.BasicLit); ok && lit.Kind == token.STRING && bad.MatchString(lit.Value) {
					t.Errorf("%s holds forbidden SQL: %s", path, fset.Position(lit.Pos()))
				}
				return true
			})
		}
		// SECURITY DEFINER appears only in 00006, and in 00010 (G014) only as the E1 reader its owner replaces with the
		// browser-kind filter (CREATE OR REPLACE of the existing 00006 function; no new definer function).
		root, _ := filepath.Abs(filepath.Join("..", "..", "..", "migrations"))
		entries, err := os.ReadDir(root)
		if err != nil {
			t.Fatal(err)
		}
		found := false
		for _, e := range entries {
			if !strings.HasSuffix(e.Name(), ".sql") {
				continue
			}
			b, err := os.ReadFile(filepath.Join(root, e.Name()))
			if err != nil {
				t.Fatal(err)
			}
			if code := regexp.MustCompile(`--[^\n]*`).ReplaceAllString(string(b), " "); strings.Contains(code, "SECURITY DEFINER") {
				replacedE1 := e.Name() == "00010_mcp_connections.sql" && strings.Count(code, "SECURITY DEFINER") == 2 &&
					strings.Count(code, "CREATE OR REPLACE FUNCTION app.auth_resolve_session(p_hash bytea, p_at timestamptz)") == 2 &&
					!regexp.MustCompile(`(?i)\bCREATE\s+FUNCTION\b[^$]*SECURITY\s+DEFINER`).MatchString(code)
				// G016: 00011 declares exactly the two auth-reader functions (hostname_state, is_organization_member), created
				// by their owner through the session's SET chain (deviation #3 removed app.hostname_generation).
				g016Readers := e.Name() == "00011_organization_domains.sql" && strings.Count(code, "SECURITY DEFINER") == 2 &&
					len(regexp.MustCompile(`CREATE FUNCTION app\.(hostname_state|is_organization_member)\(`).FindAllString(code, -1)) == 2
				// G016 security amendment: 00012 replaces the two existing auth-reader functions only (the bound session
				// reader replaces the 2-argument one; hostname_state reads proven rows), and its down restores both.
				g016Binding := e.Name() == "00012_session_binding_and_claim_exclusivity.sql" && strings.Count(code, "SECURITY DEFINER") == 4 &&
					strings.Count(code, "CREATE FUNCTION app.auth_resolve_session(") == 2 &&
					strings.Count(code, "CREATE OR REPLACE FUNCTION app.hostname_state(") == 2 &&
					len(regexp.MustCompile(`(?i)\bCREATE\s+(OR\s+REPLACE\s+)?FUNCTION\b`).FindAllString(code, -1)) == 4
				// G017: 00013 declares exactly the auth reader's public branding projection, created by its owner through the
				// session's SET chain (Plan r2 §4.6); its down drops it.
				g017Branding := e.Name() == "00013_organization_branding.sql" && strings.Count(code, "SECURITY DEFINER") == 1 &&
					len(regexp.MustCompile(`(?i)\bCREATE\s+(OR\s+REPLACE\s+)?FUNCTION\s+app\.public_branding\(`).FindAllString(code, -1)) == 1
				// G018: 00014 declares exactly the policy reader's profile-eligibility function app.rls_profile_eligible
				// (Plan r5 §4.3); its down drops it.
				g018Eligible := e.Name() == "00014_group_onboarding.sql" && strings.Count(code, "SECURITY DEFINER") == 1 &&
					len(regexp.MustCompile(`(?i)\bCREATE\s+(OR\s+REPLACE\s+)?FUNCTION\s+app\.rls_profile_eligible\(`).FindAllString(code, -1)) == 1
				// G019: 00015 declares exactly the five policy-reader functions and the three claim-writer bodies (Plan r1 §4.3,
				// r2 addendum D-14a; Result addendum A1-01 adds rls_own_organizations); its down drops them.
				g019Membership := e.Name() == "00015_membership_invitations.sql" && strings.Count(code, "SECURITY DEFINER") == 8
				for _, fn := range []string{"rls_owner_count", "rls_dependent_group_bindings", "rls_sole_owner_scopes", "rls_own_organizations",
					"rls_group_invitation_candidates", "rls_user_enabled_lock", "invitation_claim_group", "invitation_claim_organization"} {
					g019Membership = g019Membership && strings.Count(code, "CREATE FUNCTION app."+fn+"(") == 1
				}
				// G058 slice 1a: 00016 declared six of its eight allowed functions (Plan r3 §4.3; Codex F1): the policy
				// reader's rls_enrollment_mode, the claim writer's rls_invitation_pending, enrollment_lock_organization and
				// enrollment_admit, the auth reader's auth_resolve_enrollment and the owner-owned departure trigger body; its
				// down drops them. Slice 1b adds three owner-owned SECURITY INVOKER functions in schema app
				// (organizations_enrollment_guard, enrollment_readiness_set, enrollment_history_backfill). The request slice
				// reaches the allowance of eight DEFINER functions (the claim writer's enrollment_request and the owner-owned
				// decision trigger body) and adds the owner-owned INVOKER state machine enrollment_requests_immutable: 12 app
				// functions, 8 SECURITY DEFINER, 4 SECURITY INVOKER.
				g058Enrollment := e.Name() == "00016_organization_enrollment.sql" && strings.Count(code, "SECURITY DEFINER") == 8 &&
					strings.Count(code, "SECURITY INVOKER") == 4 &&
					len(regexp.MustCompile(`(?i)\bCREATE\s+(OR\s+REPLACE\s+)?FUNCTION\s+app\.`).FindAllString(code, -1)) == 12
				for _, fn := range []string{"rls_enrollment_mode", "rls_invitation_pending", "enrollment_lock_organization", "enrollment_admit",
					"auth_resolve_enrollment", "enrollment_restrict_on_departure", "organizations_enrollment_guard", "enrollment_readiness_set",
					"enrollment_history_backfill", "enrollment_request", "enrollment_restrict_on_decision", "enrollment_requests_immutable"} {
					g058Enrollment = g058Enrollment && strings.Count(code, "CREATE FUNCTION app."+fn+"(") == 1
				}
				// G060: 00017 replaces (CREATE OR REPLACE, identity/owner/ACL kept) exactly the two policy-reader bodies and the
				// two claim-writer bodies whose Group branch reads the override; its down restores the version-16 bodies. No new
				// SECURITY DEFINER function (the override guard is an owner-owned SECURITY INVOKER trigger body).
				g060Authority := e.Name() == "00017_group_authority.sql" && strings.Count(code, "SECURITY DEFINER") == 8 &&
					!regexp.MustCompile(`(?i)\bCREATE\s+FUNCTION\b[^$]*SECURITY\s+DEFINER`).MatchString(code)
				for _, fn := range []string{"rls_owner_count", "rls_group_invitation_candidates", "rls_user_enabled_lock", "invitation_claim_group"} {
					g060Authority = g060Authority && strings.Count(code, "CREATE OR REPLACE FUNCTION app."+fn+"(") == 2
				}
				// G062: 00018 declares exactly its allowance of four DEFINER bodies (Plan r5 §3.4): the provision writer's
				// signup_provision_organization, signup_challenge_issue and signup_challenge_complete and the auth reader's
				// auth_resolve_signup (plus one postcondition message naming the inventory); its down drops them.
				g062Signup := e.Name() == "00018_platform_signup.sql" && strings.Count(code, "SECURITY DEFINER") == 5 &&
					strings.Count(code, "SECURITY DEFINER inventory differs") == 1 &&
					len(regexp.MustCompile(`(?i)\bCREATE\s+(OR\s+REPLACE\s+)?FUNCTION\s+app\.`).FindAllString(code, -1)) == 4
				for _, fn := range []string{"signup_provision_organization", "signup_challenge_issue", "signup_challenge_complete", "auth_resolve_signup"} {
					g062Signup = g062Signup && strings.Count(code, "CREATE FUNCTION app."+fn+"(") == 1
				}
				// G057: 00021 declares exactly five DEFINER bodies (Plan r5 §3.4): its up replaces the auth reader's hostname_state
				// and the provision writer's signup_provision_organization and adds the auth reader's signup_slug_taken; its down
				// restores the version-20 hostname_state and signup_provision_organization and drops signup_slug_taken.
				g057Platform := e.Name() == "00021_platform_labels.sql" && strings.Count(code, "SECURITY DEFINER") == 5 &&
					len(regexp.MustCompile(`(?i)\bCREATE\s+(OR\s+REPLACE\s+)?FUNCTION\s+app\.`).FindAllString(code, -1)) == 5
				for fn, n := range map[string]int{"hostname_state": 2, "signup_slug_taken": 1, "signup_provision_organization": 2} {
					g057Platform = g057Platform &&
						len(regexp.MustCompile(`(?i)\bCREATE\s+(OR\s+REPLACE\s+)?FUNCTION\s+app\.`+fn+`\(`).FindAllString(code, -1)) == n
				}
				// G071: 00022 declares exactly one DEFINER body (Plan r4 §3, S2): the auth reader's new
				// app.organization_id_by_slug, created by its owner; its down drops it.
				g071DefaultTenant := e.Name() == "00022_default_tenant.sql" && strings.Count(code, "SECURITY DEFINER") == 1 &&
					strings.Count(code, "CREATE FUNCTION app.") == 1 &&
					len(regexp.MustCompile(`(?i)\bCREATE\s+(OR\s+REPLACE\s+)?FUNCTION\s+app\.`).FindAllString(code, -1)) == 1 &&
					strings.Count(code, "CREATE FUNCTION app.organization_id_by_slug(") == 1
				if e.Name() != "00006_row_level_security.sql" && !replacedE1 && !g016Readers && !g016Binding && !g017Branding && !g018Eligible &&
					!g019Membership && !g058Enrollment && !g060Authority && !g062Signup && !g057Platform && !g071DefaultTenant {
					t.Errorf("%s declares a SECURITY DEFINER function", e.Name())
				}
				found = true
			}
		}
		if !found {
			t.Error("00006_row_level_security.sql declares no SECURITY DEFINER function")
		}
	})

	t.Run("facade transactions initialize the context first", func(t *testing.T) {
		ict := funcDecl(files["internal/tenancy/access/tx.go"], "inContextTx")
		if ict == nil {
			t.Fatal("inContextTx not found")
		}
		calls := strings.Join(calledNames(ict.Body), ",")
		if !strings.HasPrefix(calls, "inTx,InitContext,") || !strings.HasSuffix(calls, ",fn") {
			t.Errorf("inContextTx calls %s: want inTx, then InitContext before fn", calls)
		}
		// inTx is called only by inContextTx, and only inTx opens a transaction.
		n := 0
		for path, f := range files {
			if !strings.HasPrefix(path, "internal/tenancy/access/") {
				continue
			}
			for _, d := range f.Decls {
				fd, ok := d.(*ast.FuncDecl)
				if !ok || fd.Body == nil {
					continue
				}
				for _, c := range calledNames(fd.Body) {
					switch {
					case c == "inTx" && fd.Name.Name != "inContextTx":
						t.Errorf("%s: %s calls inTx directly", path, fd.Name.Name)
					case c == "BeginTx" && fd.Name.Name != "inTx":
						t.Errorf("%s: %s opens a transaction outside inTx", path, fd.Name.Name)
					case c == "inContextTx":
						n++
					}
				}
			}
		}
		if n != 19 {
			t.Errorf("%d inContextTx call sites, want 19 (G063's setup rename; resolve, guarded, manageableProfiles, G009's storageTx, G011's auditTx and "+
				"G014's six: the MCP snapshot, the selection decision and write, activation, setup create and setup list; G018's list operations; "+
				"G019's two IDENTITY-shape calls, ConsumeOrganizationInvitation and RetirementEligibility, and the trusted "+
				"RetireUser's two: the IDENTITY-shape preflight and the per-Organization departure; G058's two IDENTITY-shape "+
				"calls, admitInvitation and enrollmentTx (the open and request admissions and, request slice, the applicant's "+
				"request, status read and abandonment))", n)
		}
	})

	t.Run("identity statements run in an initialized transaction", func(t *testing.T) {
		mcpStore, retire := files["internal/identity/mcp_store.go"], files["internal/identity/retire.go"]
		if mcpStore == nil || retire == nil {
			t.Fatal("internal/identity/mcp_store.go or retire.go not found")
		}
		callers := map[string][]string{}
		// G014: the same rules for mcp_store.go, with its own constants and pinned functions. initContext and
		// initContextShape are the shared primitive under the named initializers, not initializers themselves.
		// G019 F-01: retire.go with its own pinned functions and initializer calls.
		for _, src := range []struct {
			file   *ast.File
			consts map[string]string
			pinned map[string]string
			inits  map[string]string
		}{{identityStore, ic, identityProtectedFuncs, identityInitCalls}, {mcpStore, rlsConsts(mcpStore), identityMCPProtectedFuncs, identityInitCalls},
			{retire, rlsConsts(retire), identityRetireProtectedFuncs, identityRetireInitCalls}} {
			seen := map[string]bool{}
			for _, d := range src.file.Decls {
				fd, ok := d.(*ast.FuncDecl)
				if !ok || fd.Body == nil {
					continue
				}
				name := fd.Name.Name
				for _, c := range calledNames(fd.Body) {
					callers[c] = append(callers[c], name)
				}
				firstProtected, firstInit := token.NoPos, token.NoPos
				var inits []string
				ast.Inspect(fd.Body, func(n ast.Node) bool {
					switch x := n.(type) {
					case *ast.BasicLit:
						if s, _ := strconv.Unquote(x.Value); x.Kind == token.STRING && rlsProtectedTable.MatchString(s) && firstProtected == token.NoPos {
							firstProtected = x.Pos()
						}
					case *ast.Ident:
						if s, ok := src.consts[x.Name]; ok && rlsProtectedTable.MatchString(s) && firstProtected == token.NoPos {
							firstProtected = x.Pos()
						}
					case *ast.CallExpr:
						if id, ok := x.Fun.(*ast.Ident); ok && strings.HasPrefix(id.Name, "init") && id.Name != "initContext" &&
							id.Name != "initContextShape" {
							inits = append(inits, id.Name)
							if firstInit == token.NoPos {
								firstInit = x.Pos()
							}
						}
					}
					return true
				})
				if want, ok := src.inits[name]; ok || len(inits) > 0 {
					if got := strings.Join(inits, ","); got != want {
						t.Errorf("%s initializes %q, want %q", name, got, want)
					}
				}
				if firstProtected == token.NoPos {
					continue
				}
				seen[name] = true
				switch src.pinned[name] {
				case "init":
					if firstInit == token.NoPos || firstInit > firstProtected {
						t.Errorf("%s issues a protected statement before its initializer", name)
					}
				case "callee", "reader":
				default:
					t.Errorf("%s issues a statement on a protected identity table outside the pinned functions", name)
				}
			}
			for name, kind := range src.pinned {
				if kind != "reader" && !seen[name] {
					t.Errorf("%s: pinned but issues no protected statement", name)
				}
			}
		}
		for _, c := range []string{"sqlAuthResolveSession", "sqlAuthPrincipalSubject", "sqlAuthLogoutRef"} {
			if !strings.Contains(ic[c], "app.auth_") || rlsProtectedTable.MatchString(ic[c]) {
				t.Errorf("%s is not a reader call: %s", c, ic[c])
			}
		}
		if got := strings.Join(callers["issue"], ","); got != "signIn,signIn" {
			t.Errorf("issue is called by %q, want signIn (continuity and mapped paths)", got)
		}
		if got := strings.Join(callers["createMapping"], ","); got != "signIn,signupT1" {
			t.Errorf("createMapping is called by %q, want signIn and (G062) signupT1", got)
		}
		// liveSession is reached only through the two liveness methods (the caller's tenant transaction).
		for path, f := range files {
			for _, c := range calledNames(f) {
				if c == "liveSession" && path != "internal/identity/service.go" {
					t.Errorf("%s calls liveSession", path)
				}
			}
		}
	})

	t.Run("expired-login cleanup only in login start after EMPTY", func(t *testing.T) {
		if ic["sqlExpireLoginTransactions"] != "DELETE FROM app.login_transactions" {
			t.Fatalf("the cleanup statement is %q, want exactly DELETE FROM app.login_transactions", ic["sqlExpireLoginTransactions"])
		}
		deletes := 0
		del := regexp.MustCompile(`DELETE\s+FROM\s+app\.login_transactions`)
		for path, f := range files {
			ast.Inspect(f, func(n ast.Node) bool {
				if id, ok := n.(*ast.Ident); ok && id.Name == "sqlExpireLoginTransactions" && path != "internal/identity/store.go" {
					t.Errorf("%s references the cleanup statement", path)
				}
				if lit, ok := n.(*ast.BasicLit); ok && lit.Kind == token.STRING {
					if s, _ := strconv.Unquote(lit.Value); del.MatchString(s) {
						deletes++
					}
				}
				return true
			})
		}
		if deletes != 2 {
			t.Errorf("%d DELETE statements on login_transactions, want 2 (the cleanup and the callback consume)", deletes)
		}
		// G016: between EMPTY and the cleanup, admit runs the tenant-host resolution (one reader statement). G058: the
		// expired enrollment-session cleanup Exec follows the login cleanup, still in the EMPTY shape before initLogin. G062:
		// the expired signup-credential cleanup Exec follows it, still EMPTY.
		calls := strings.Join(calledNames(funcDecl(identityStore, "insertLogin").Body), ",")
		if !strings.Contains(calls, "initEmpty,admit,Exec,unavailable,Exec,unavailable,Exec,unavailable,initLogin,Exec") {
			t.Errorf("insertLogin calls %s: want initEmpty, admit, the login cleanup Exec, the enrollment cleanup Exec, the signup "+
				"cleanup Exec, initLogin, the insert Exec", calls)
		}
		if ic["sqlExpireSignupSessions"] != "DELETE FROM app.signup_sessions" {
			t.Errorf("the signup cleanup statement is %q, want exactly DELETE FROM app.signup_sessions", ic["sqlExpireSignupSessions"])
		}
		var users []string
		for _, d := range identityStore.Decls {
			fd, ok := d.(*ast.FuncDecl)
			if !ok || fd.Body == nil {
				continue
			}
			ast.Inspect(fd.Body, func(n ast.Node) bool {
				if id, ok := n.(*ast.Ident); ok && id.Name == "sqlExpireLoginTransactions" {
					users = append(users, fd.Name.Name)
				}
				return true
			})
		}
		if strings.Join(users, ",") != "insertLogin" {
			t.Errorf("the cleanup statement is used by %v, want insertLogin only", users)
		}
	})

	t.Run("no bootstrap API accepts a user UUID", func(t *testing.T) {
		var got []string
		for path, f := range files {
			if !strings.HasPrefix(path, "internal/identity/") || strings.Contains(path, "/workos/") || strings.Contains(path, "/identitytest/") {
				continue
			}
			for _, d := range f.Decls {
				fd, ok := d.(*ast.FuncDecl)
				if !ok || !fd.Name.IsExported() {
					continue
				}
				for _, p := range fd.Type.Params.List {
					for _, n := range p.Names {
						if strings.EqualFold(n.Name, "userID") || strings.EqualFold(n.Name, "user") {
							r := ""
							if fd.Recv != nil {
								r = recvName(fd.Recv.List[0].Type) + "."
							}
							got = append(got, r+fd.Name.Name)
						}
					}
				}
			}
		}
		slices.Sort(got)
		// G014: the MCP binding read and the setup revoke take the subject a verified Connect token mapped through E2 (or
		// the resolved browser session); the three facade helpers take the facade's resolved subject. None creates a user.
		// G017 D-6: Membership reads the membership fact of an already resolved session's user for that session's own
		// bound Organization (the session route asks it only after its origin/Organization binding check); it creates or
		// changes nothing.
		// G019 F-01: the trusted, route-less retirement finalization of an operator-named subject (no user is created).
		// G063: the rename's two facade helpers (the target session read and the name write) take the facade's resolved
		// subject; neither names a protected identity table (identityMCPProtectedFuncs is unchanged).
		want := "Service.CheckLiveSession,Service.DisableUser,Service.EnableUser,Service.FinalizeRetirement,Service.InsertMCPConnectionTx," +
			"Service.ListMCPConnectionsTx,Service.LockLiveSession,Service.Membership,Service.ReadMCPConnection," +
			"Service.ReadMCPConnectionSessionTx,Service.RenameMCPConnectionTx,Service.RevokeOwnMCPConnection," +
			"Service.UserDisplayNameTx,Store.DisableUser,Store.EnableUser,Store.FinalizeRetirement"
		if strings.Join(got, ",") != want {
			t.Errorf("exported identity functions taking a user: %v, want %s (the trusted route-less lifecycle, E7, the "+
				"liveness checks of an already resolved session and the G014 MCP functions of an already mapped subject)", got, want)
		}
	})
}
