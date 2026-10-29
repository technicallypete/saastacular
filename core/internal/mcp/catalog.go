package mcp

import (
	"errors"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"core/internal/tenancy/access"
)

// G015: the closed catalog spec and requirement-aware registration (docs/foundation/MCP.md "Discovery"; Plan r4 §1,
// §5.1). Every tool declares exactly one access.ToolRequirement. The binding, membership and context requirements are
// the G014 control tools' own handler checks: those tools are registered on every request's server and no check runs
// in front of them, so logout, the explicit clear and the read tools keep their G014 answers. A capability-mapped tool
// is registered on a request's server only when access.AuthorizeCatalog finds it eligible on that request's own
// snapshot; an unregistered name is then the SDK's unknown tool, so a hidden tool and an absent one answer alike.
// Production declares no capability-mapped tool. Nothing is cached: the catalog is fixed at compile time and the
// eligible set is computed per request.

// catalogEntry is one tool of the spec: its name, its one requirement, its input and output schemas and its handler
// bound to the request's tools.
type catalogEntry struct {
	name          string
	req           access.ToolRequirement
	input, output map[string]any
	handler       func(t tools) sdk.ToolHandler
}

// capabilityMapped reports whether the entry is registered only when eligible (any requirement but the three control
// requirements).
func (e catalogEntry) capabilityMapped() bool {
	switch e.req {
	case access.RequireBinding(), access.RequireMembership(), access.RequireContext():
		return false
	}
	return true
}

// catalog is a validated spec: the control entries registered on every request and the capability entries with their
// requirements (the AuthorizeCatalog input).
type catalog struct {
	control    []catalogEntry
	capability map[string]catalogEntry
	reqs       map[string]access.ToolRequirement
}

// errCatalog is a refused catalog spec: an empty or repeated name, a zero (unknown) requirement, a missing input schema
// or one whose type is not object, or a missing handler. It checks input-schema shape only: output schemas and schema
// closedness are not validated here (the fixed production schemas are pinned by TestToolCatalog).
var errCatalog = errors.New("mcp: invalid tool catalog")

// catalogSpec validates entries (names, requirements, input-schema shape and handlers only; see errCatalog) and splits
// them by requirement.
func catalogSpec(entries []catalogEntry) (catalog, error) {
	c := catalog{capability: map[string]catalogEntry{}, reqs: map[string]access.ToolRequirement{}}
	seen := map[string]bool{}
	for _, e := range entries {
		if e.name == "" || seen[e.name] || e.req == (access.ToolRequirement{}) || e.input == nil || e.input["type"] != "object" ||
			e.handler == nil {
			return catalog{}, errCatalog
		}
		seen[e.name] = true
		if e.capabilityMapped() {
			c.capability[e.name] = e
			c.reqs[e.name] = e.req
		} else {
			c.control = append(c.control, e)
		}
	}
	return c, nil
}

// productionRequirements are the requirements of the five control tools (Plan r4 §1): logout the exact binding only
// (the termination exception), switch_group the Organization membership (the explicit clear after a lost selected
// Group; a select rechecks both Groups in its guarded write), the read tools membership and a valid selection.
var productionRequirements = map[string]access.ToolRequirement{
	ToolGetCurrentUser:    access.RequireContext(),
	ToolGetCurrentContext: access.RequireContext(),
	ToolListGroups:        access.RequireContext(),
	ToolSwitchGroup:       access.RequireMembership(),
	ToolLogout:            access.RequireBinding(),
}

// productionEntries returns the production spec: exactly the five control tools, in ToolNames order.
func productionEntries() []catalogEntry {
	entries := make([]catalogEntry, 0, len(ToolNames))
	for _, name := range ToolNames {
		in, out, _ := ToolSchemas(name)
		entries = append(entries, catalogEntry{name: name, req: productionRequirements[name], input: in, output: out,
			handler: func(t tools) sdk.ToolHandler { return t.handler(name) }})
	}
	return entries
}

// productionCatalog is validated when the package initializes, before any server is built: an invalid production spec
// stops the process instead of serving.
var productionCatalog = mustCatalog(productionEntries())

func mustCatalog(entries []catalogEntry) catalog {
	c, err := catalogSpec(entries)
	if err != nil {
		panic(err)
	}
	return c
}

// catalog returns the spec this handler serves: productionCatalog, unless a test published an override on this handler
// (Handler.fixture, catalog_fixture_test.go). The override is private to its handler, never set outside tests, and
// read with one atomic load, so a production request takes no lock and shares no state with other handlers.
func (h *Handler) catalog() catalog {
	if c := h.fixture.Load(); c != nil {
		return *c
	}
	return productionCatalog
}

// registerControl adds every control entry to s (no statement; each handler runs its own G014 check).
func (c catalog) registerControl(s *sdk.Server, t tools) {
	for _, e := range c.control {
		s.AddTool(&sdk.Tool{Name: e.name, InputSchema: e.input, OutputSchema: e.output}, e.handler(t))
	}
}

// registerEligible adds the capability entries named eligible by AuthorizeCatalog to s; other names are ignored.
func (c catalog) registerEligible(s *sdk.Server, t tools, names []string) {
	for _, n := range names {
		if e, ok := c.capability[n]; ok {
			s.AddTool(&sdk.Tool{Name: e.name, InputSchema: e.input, OutputSchema: e.output}, e.handler(t))
		}
	}
}

// isControl reports whether name is a control entry, registered before any request runs.
func (c catalog) isControl(name string) bool {
	for _, e := range c.control {
		if e.name == name {
			return true
		}
	}
	return false
}
