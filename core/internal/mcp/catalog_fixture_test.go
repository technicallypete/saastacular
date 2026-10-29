package mcp

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"core/internal/tenancy/access"
)

// G015 test-only capability-mapped tools (Plan r4 §6.1). They exist only in this _test.go file, use canonical
// capabilities only, have closed schemas, and perform no effect and no check of their own: each counts its invocations.
// Production never sets a handler's fixture override (TestToolCatalog); each override is private to its handler.

// The fixture tools and their capabilities.
const (
	fixtureOrgPolicy     = "test.org_policy"
	fixtureGroupUpdate   = "test.group_update"
	fixtureProfileUpdate = "test.profile_update"
	fixtureAuditRead     = "test.audit_read"
)

var fixtureCapabilities = map[string]string{
	fixtureOrgPolicy:     "organization.manage_policy",
	fixtureGroupUpdate:   "group.update",
	fixtureProfileUpdate: "business_profile.profile.update",
	fixtureAuditRead:     "audit.read",
}

// fixtureNames are the fixture tools in sorted order.
var fixtureNames = []string{fixtureAuditRead, fixtureGroupUpdate, fixtureOrgPolicy, fixtureProfileUpdate}

// fixtureCounters counts the invocations of each fixture tool.
type fixtureCounters map[string]*atomic.Int64

func (f fixtureCounters) get(name string) int64 { return f[name].Load() }

// G020 (Plan r2 §4.3, T10; D-1 decided): three test-only capability-mapped label tools on the same fixture path, mapped
// to the three label capabilities. They prove finite label-capability eligibility, current-context refusal and hidden
// direct-call behaviour only; they are never a production label operation (the production catalog stays the five
// control tools, TestToolCatalog) and are published only by useLabelFixtures, never with the four G015 tools.
const (
	fixtureLabelsListOrganization = "labels.list_organization"
	fixtureLabelsListGroup        = "labels.list_group"
	fixtureLabelsReadGroup        = "labels.read_group"
)

var labelFixtureCapabilities = map[string]string{
	fixtureLabelsListOrganization: "label.organization.manage",
	fixtureLabelsListGroup:        "label.group.manage",
	fixtureLabelsReadGroup:        "label.group.read",
}

// labelFixtureNames are the label fixture tools in sorted order.
var labelFixtureNames = []string{fixtureLabelsListGroup, fixtureLabelsListOrganization, fixtureLabelsReadGroup}

// fixtureEntries returns the four fixture entries and their counters.
func fixtureEntries(t testing.TB) ([]catalogEntry, fixtureCounters) {
	t.Helper()
	return fixtureEntriesOf(t, fixtureNames, fixtureCapabilities)
}

// labelFixtureEntries returns the three label fixture entries and their counters.
func labelFixtureEntries(t testing.TB) ([]catalogEntry, fixtureCounters) {
	t.Helper()
	return fixtureEntriesOf(t, labelFixtureNames, labelFixtureCapabilities)
}

// fixtureEntriesOf returns one counting, effect-free entry per name, mapped to its canonical capability.
func fixtureEntriesOf(t testing.TB, names []string, capabilities map[string]string) ([]catalogEntry, fixtureCounters) {
	t.Helper()
	counters := fixtureCounters{}
	var entries []catalogEntry
	for _, name := range names {
		req, err := access.RequireCapability(capabilities[name])
		if err != nil {
			t.Fatalf("fixture %s: %v", name, err)
		}
		n := &atomic.Int64{}
		counters[name] = n
		entries = append(entries, catalogEntry{name: name, req: req, input: emptySchema,
			output: object(map[string]any{"invoked": map[string]any{"const": true}}, "invoked"),
			handler: func(tools) sdk.ToolHandler {
				return func(context.Context, *sdk.CallToolRequest) (*sdk.CallToolResult, error) {
					n.Add(1)
					return result(map[string]any{"invoked": true}), nil
				}
			}})
	}
	return entries, counters
}

// fixtureCatalog returns a new immutable catalog of the production entries plus the fixture tools, and its counters.
func fixtureCatalog(t testing.TB) (*catalog, fixtureCounters) {
	t.Helper()
	extra, counters := fixtureEntries(t)
	c, err := catalogSpec(append(productionEntries(), extra...))
	if err != nil {
		t.Fatalf("fixture catalog: %v", err)
	}
	return &c, counters
}

// labelFixtureCatalog returns a new immutable catalog of the production entries plus the label fixture tools.
func labelFixtureCatalog(t testing.TB) (*catalog, fixtureCounters) {
	t.Helper()
	extra, counters := labelFixtureEntries(t)
	c, err := catalogSpec(append(productionEntries(), extra...))
	if err != nil {
		t.Fatalf("label fixture catalog: %v", err)
	}
	return &c, counters
}

// useFixtures makes h, and only h, serve the production catalog plus the fixture tools until the test ends or
// clearFixtures(h). It publishes a new catalog with one atomic store, so it is safe while h serves requests.
func useFixtures(t testing.TB, h *Handler) fixtureCounters {
	t.Helper()
	c, counters := fixtureCatalog(t)
	h.fixture.Store(c)
	t.Cleanup(func() { clearFixtures(h) })
	return counters
}

// useLabelFixtures makes h, and only h, serve the production catalog plus the three label fixture tools (G020 T10)
// until the test ends or clearFixtures(h); one atomic store, like useFixtures.
func useLabelFixtures(t testing.TB, h *Handler) fixtureCounters {
	t.Helper()
	c, counters := labelFixtureCatalog(t)
	h.fixture.Store(c)
	t.Cleanup(func() { clearFixtures(h) })
	return counters
}

// clearFixtures restores the production catalog of h only.
func clearFixtures(h *Handler) { h.fixture.Store(nil) }

// TestCatalogFixtureIsolation (Duo D2): a fixture override is private to its handler. While handler A's override is
// published and cleared concurrently with reads (run with -race where the toolchain supports it), handler B always
// serves exactly the production catalog and A serves either the production or the whole fixture catalog, never a mix;
// clearing A leaves B's own override in place.
func TestCatalogFixtureIsolation(t *testing.T) {
	a, b := testHandler(t, &stubVerifier{}), testHandler(t, &stubVerifier{})
	fc, _ := fixtureCatalog(t)
	if c := b.catalog(); len(c.capability) != 0 || len(c.control) != len(ToolNames) {
		t.Fatalf("handler B before: %d control, %d capability", len(c.control), len(c.capability))
	}
	var wg sync.WaitGroup
	var bad atomic.Int64
	stop := make(chan struct{})
	for range 4 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stop:
					return
				default:
				}
				if c := b.catalog(); len(c.capability) != 0 || len(c.reqs) != 0 {
					bad.Add(1)
				}
				c := a.catalog()
				production := len(c.capability) == 0 && len(c.reqs) == 0
				fixture := len(c.capability) == len(fixtureNames) && len(c.reqs) == len(fixtureNames)
				if !production && !fixture {
					bad.Add(1)
				}
			}
		}()
	}
	for i := range 2000 {
		if i%2 == 0 {
			a.fixture.Store(fc)
		} else {
			clearFixtures(a)
		}
	}
	close(stop)
	wg.Wait()
	if n := bad.Load(); n != 0 {
		t.Fatalf("%d reads saw a foreign or partial catalog", n)
	}
	useFixtures(t, a)
	useFixtures(t, b)
	clearFixtures(a)
	if c := a.catalog(); len(c.capability) != 0 {
		t.Fatalf("handler A after its reset: %d capability entries", len(c.capability))
	}
	if c := b.catalog(); len(c.capability) != len(fixtureNames) {
		t.Fatalf("handler B after A's reset: %d capability entries", len(c.capability))
	}
}
