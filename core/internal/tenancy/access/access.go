// Package access is core's trusted tenant context and shared policy evaluation (G007; docs/foundation/AUTHORIZATION.md).
//
// A request owner calls BeginRequest once per request with the opaque session credential, defers the returned end
// function, and resolves an Organization (and optional Group) context with Resolve. The guarded facade then authorizes
// and acts in one database transaction per call: it rechecks the session, the user and every relied-on fact, evaluates
// one closed, deny-by-default capability matrix (Spec R3) and the one Organization policy flag, and writes only on an
// allow, ordered against every conflicting writer by row locks. Nothing here is an authority snapshot: a context binds
// a request to its authenticated subject and confers nothing by itself.
//
// The package contains no SQL (tenancy and identity own their statements), exports no evaluator, allow result or
// capability check, and has no HTTP, MCP or CLI surface. In G007 nothing imports it: future adapters own transport
// concerns such as origin and CSRF checks, request concurrency and deadlines.
package access

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5"

	"core/internal/identity"
	"core/internal/objectstore"
	"core/internal/tenancy"
)

// sessions is the identity boundary the facade uses; *identity.Service satisfies it. Session resolves and renews a
// credential at its bound origin (G004; G016 security amendment); the two liveness checks run inside the facade's
// transaction.
type sessions interface {
	Session(ctx context.Context, credential, origin string) (identity.SessionInfo, error)
	LockLiveSession(ctx context.Context, tx pgx.Tx, sessionID, userID string) error
	CheckLiveSession(ctx context.Context, tx pgx.Tx, sessionID, userID string) error
}

// Service is the trusted context resolver and guarded facade. Its database pool connects as the runtime role.
type Service struct {
	db     tenancy.DB
	id     sessions
	logger *slog.Logger

	// G009: the owned object store and its bounds. objects is nil when the Service was built without storage (or with
	// an invalid ObjectConfig); every storage operation is then unavailable (objects_unconfigured).
	objects objectstore.Store
	ocfg    ObjectConfig
	// hooks are inert test seams: nil in every production construction (U-OBJ-SEAM).
	hooks *objectHooks

	// G014: the MCP connection helpers of identity (nil when the identity boundary does not provide them: every MCP
	// operation is then refused) and their inert ordering hooks (nil in every production construction).
	mcp      mcpIdentity
	mcpHooks *mcpHooks
	// G016: inert domain-lifecycle ordering hooks (nil in every production construction); domains.go.
	domainHooks *domainHooks
}

// NewService returns a Service over the runtime-role pool db and the accepted identity service. A nil logger discards
// the access.decision records. It has no object store: every storage operation is unavailable (objects_unconfigured).
func NewService(db tenancy.DB, id *identity.Service, logger *slog.Logger) *Service {
	return newService(db, id, logger)
}

// NewServiceWithObjects is NewService with the guarded storage operations (G009; docs/foundation/OBJECT-STORAGE.md)
// over the owned object store and its bounds. A nil store or a cfg outside its bounds (ObjectConfig) is refused: the
// Service is built without storage, so every storage operation fails closed as unavailable (objects_unconfigured).
func NewServiceWithObjects(db tenancy.DB, id *identity.Service, logger *slog.Logger, store objectstore.Store, cfg ObjectConfig) *Service {
	s := newService(db, id, logger)
	if store != nil && cfg.valid() {
		s.objects, s.ocfg = store, cfg
	}
	return s
}

func newService(db tenancy.DB, id sessions, logger *slog.Logger) *Service {
	if logger == nil {
		logger = slog.New(slog.DiscardHandler)
	}
	s := &Service{db: db, id: id, logger: logger}
	if m, ok := id.(mcpIdentity); ok {
		s.mcp = m
	}
	return s
}
