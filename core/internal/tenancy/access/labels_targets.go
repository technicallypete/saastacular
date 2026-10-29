package access

import (
	"context"

	"core/internal/tenancy"
)

// The closed typed label target registry (G020 Plan r2 §1.4; Spec §1 "closed typed core target-validation contract").
// A label assignment names its target by kind and id. Production registers exactly one kind, group, and admits it only
// for Organization-owned definitions: production storage (label_assignments_target_shape, 00020) holds no Group-owned
// assignment. Selector parsing and kind/owner-scope admission run before any SQL, so an unknown, unregistered or
// ambiguous selector, or a Group-owned definition aimed at a Group, is refused without touching storage.
//
// Registration contract for a future target type T (owned and reviewed by the Goal that introduces T; TENANCY.md):
//
//  1. a core adapter registered here: resolve reads the target from authoritative storage inside the operation's
//     transaction, under the invoker's row-level security (never a bypassing role, so a target the caller cannot see
//     resolves as absent and the facade's not_found/forbidden split reveals nothing), and returns its owning Group (the
//     caller supplies nothing but the id); readable decides current read
//     authority on the resolved target from the operation's facts; admits names the definition owner scopes it accepts
//     (a Group-owned target: the group scope only);
//  2. a migration that widens label_assignments_target_shape with
//     (owner_scope_type = 'group' AND target_type = 'T' AND target_group_id = owner_scope_id);
//  3. a per-type, NULL-skipping existence-and-ownership key: a stored generated column
//     target_T_id uuid GENERATED ALWAYS AS (CASE WHEN target_type = 'T' THEN target_id END) STORED and the foreign key
//     (organization_id, target_T_id, target_group_id) -> T (organization_id, id, group_id), ON UPDATE RESTRICT and ON
//     DELETE RESTRICT or CASCADE (T carries UNIQUE (organization_id, id, group_id)); under MATCH SIMPLE it constrains only
//     T rows and binds the target's existence and its actual owning Group to the definition's owning Group;
//  4. a disjunct of label_assignments_select requiring the target row to be visible to the invoker under T's own
//     row-level security with the row's ownership facts (EXISTS (SELECT 1 FROM T t WHERE t.organization_id =
//     organization_id AND t.id = target_id AND t.group_id = target_group_id)), so that readable and storage agree.
//
// Core statements never name a per-type key column. Until T exists, production SQL carries no per-type
// target-visibility predicate, and nothing here is evidence of a live Group-owned target.

// labelTargetKind is a registered target kind of a label assignment (the target_type column).
type labelTargetKind string

// labelTargetGroup is the one production kind: a Group of the active Organization (Organization-owned definitions only).
const labelTargetGroup labelTargetKind = "group"

// labelTarget is a parsed target selector: the kind and the target id. Group is the target's owning Group, set only
// from the adapter's resolve in the operation's transaction, never from the caller.
type labelTarget struct {
	Kind  labelTargetKind
	ID    [16]byte
	Group tenancy.GroupID
}

// labelTargetAdapter is the per-kind contract of the registry (item 1 above).
type labelTargetAdapter interface {
	// admits reports whether definitions of the owner scope may target this kind.
	admits(owner tenancy.ScopeType) bool
	// resolve reads the target inside the operation's transaction from authoritative storage (or the facts the
	// transaction already locked), under the invoker's row-level security, and returns its owning Group and whether it
	// exists in Organization org.
	resolve(ctx context.Context, tx *trackedTx, f Facts, org tenancy.OrganizationID, id [16]byte) (owning tenancy.GroupID, ok bool, err error)
	// readable decides current read authority on the resolved target for the caller (target-member visibility).
	readable(sel contextSelection, f Facts, owning tenancy.GroupID) bool
}

// labelTargets is the closed production registry: exactly {group}.
var labelTargets = map[labelTargetKind]labelTargetAdapter{labelTargetGroup: groupTargetAdapter{}}

// labelTargetHooks is an inert test seam: additional kinds a test registers (never group). It is consulted only for a
// kind the production registry does not hold.
type labelTargetHooks struct {
	targets map[labelTargetKind]labelTargetAdapter
}

// labelHooks is nil in every production path (the objectHooks/mcpHooks precedent): production startup registers
// nothing, so the registry stays {group}.
var labelHooks *labelTargetHooks

// labelTargetAdapterFor returns the adapter of a kind: the production registry first, then (tests only) labelHooks.
func labelTargetAdapterFor(k labelTargetKind) (labelTargetAdapter, bool) {
	if a, ok := labelTargets[k]; ok {
		return a, true
	}
	if labelHooks != nil {
		a, ok := labelHooks.targets[k]
		return a, ok
	}
	return nil, false
}

// groupTargetAdapter is the production group kind: the target is the Group itself, already read (and, on a write,
// locked FOR NO KEY UPDATE) by the operation's facts; readable is group.read on it.
type groupTargetAdapter struct{}

func (groupTargetAdapter) admits(owner tenancy.ScopeType) bool {
	return owner == tenancy.ScopeOrganization
}

func (groupTargetAdapter) resolve(_ context.Context, _ *trackedTx, f Facts, org tenancy.OrganizationID, id [16]byte) (tenancy.GroupID, bool, error) {
	g := tenancy.GroupID(id)
	if !hasGroup(f, org, g) {
		return tenancy.GroupID{}, false, nil
	}
	return g, true, nil
}

func (groupTargetAdapter) readable(sel contextSelection, f Facts, owning tenancy.GroupID) bool {
	ok, _ := decide(capGroupRead, sel, f, Target{Kind: kindGroup, Organization: sel.Organization, Group: owning})
	return ok
}

// labelSelector is the caller's untrusted target selector: either group (a Group id) or kind and id (another kind),
// never both. Every field is wire text.
type labelSelector struct {
	group string
	kind  string
	id    string
}

// parseLabelTarget parses a selector before any SQL: two target forms at once is target_ambiguous; no target, or a
// kind no registry holds, is target_unsupported; a malformed id is invalid_id.
func parseLabelTarget(s labelSelector) (labelTarget, *failure) {
	if s.group != "" && (s.kind != "" || s.id != "") {
		return labelTarget{}, labelInputFailure(fieldTargetAmbiguous)
	}
	kind, id := labelTargetKind(s.kind), s.id
	if s.group != "" {
		kind, id = labelTargetGroup, s.group
	}
	if kind == "" {
		return labelTarget{}, labelInputFailure(fieldTargetUnsupported)
	}
	if _, ok := labelTargetAdapterFor(kind); !ok {
		return labelTarget{}, labelInputFailure(fieldTargetUnsupported)
	}
	parsed, err := tenancy.ParseGroupID(id) // the canonical UUID parser; the id is any kind's target id
	if err != nil {
		return labelTarget{}, inputFailure(err, reasonInputInvalid)
	}
	return labelTarget{Kind: kind, ID: [16]byte(parsed)}, nil
}

// admitLabelTarget checks, before any SQL, that a definition of the owner scope may target the parsed kind: a
// Group-owned definition targeting a Group (or an Organization-owned one targeting a Group-owned kind) is
// target_unsupported. It returns the kind's adapter.
func admitLabelTarget(owner tenancy.ScopeType, t labelTarget) (labelTargetAdapter, *failure) {
	a, ok := labelTargetAdapterFor(t.Kind)
	if !ok || !a.admits(owner) {
		return nil, labelInputFailure(fieldTargetUnsupported)
	}
	return a, nil
}
