// Command server is the core service. It owns domain authority, policy and
// the public edge; it serves liveness, readiness and the internal identity
// contract (G004): sign-in, callback, session resolution and sign-out under
// /internal/v1/auth/. When MCP is configured (G014) it also serves the
// tenant-bound MCP connections: the per-connection endpoint and its
// protected-resource metadata on tenant hosts, and the internal setup routes
// under /internal/v1/mcp/. G016: tenant hosts resolve through Organization
// slugs and verified custom domains (the durable resolver); the internal
// domain lifecycle, host resolution and tls-ask routes live under
// /internal/v1/domains/ and /internal/v1/hosts/. G017: Organization branding
// (the public projection, the logo bytes and the guarded management routes)
// lives under /internal/v1/branding, over the owned object store. G018: Group
// creation and Business Profile onboarding live under /internal/v1/groups and
// /internal/v1/business-profiles. G019: membership, ownership and Group
// invitation administration and the own retirement eligibility live under
// /internal/v1/members, /internal/v1/ownership/transfer,
// /internal/v1/invitations and /internal/v1/account/retirement-eligibility.
// G058: Organization enrollment (the public projection, the applicant's own
// enrollment credential and the manager's policy and requests) lives under
// /internal/v1/enrollment. G062: platform root signup (the root signup
// credential's status, slug suggestion, constrained provisioning, the root
// proof and the tenant continuation hops) lives under /internal/v1/signup;
// the platform root origin enables the signup-intent login and the
// root-origin callback limiter of the auth routes. G020: generic labels (the
// Organization and Group label definitions and Organization-owned labels on
// Groups) live under /internal/v1/labels and /internal/v1/groups/{id}/labels.
// G071: the optional default-tenant root alias (DEFAULT_TENANT_SLUG, selected
// and pinned once at startup) and the multi-tenant root presentation
// (PLATFORM_ROOT_PAGE) reach web through the host resolution answer.
//
// "server healthcheck [--url URL]" is the container healthcheck.
package main

import (
	"context"
	"errors"
	"flag"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/jackc/pgx/v5/pgxpool"

	brandingapi "core/internal/branding/api"
	"core/internal/config"
	"core/internal/domains"
	domainsapi "core/internal/domains/api"
	enrollmentapi "core/internal/enrollment/api"
	"core/internal/health"
	"core/internal/identity"
	"core/internal/identity/workos"
	labelsapi "core/internal/labels/api"
	"core/internal/mcp"
	membershipapi "core/internal/membership/api"
	"core/internal/observability"
	onboardingapi "core/internal/onboarding/api"
	"core/internal/signup"
	signupapi "core/internal/signup/api"
)

// version is the build version reported as MCP server info (no release versioning yet).
const version = "0.0.0"

func main() {
	if len(os.Args) > 1 && os.Args[1] == "healthcheck" {
		os.Exit(healthcheck(os.Args[2:]))
	}
	if err := run(); err != nil {
		observability.LogExit(os.Stderr, err)
		os.Exit(1)
	}
}

func healthcheck(args []string) int {
	fs := flag.NewFlagSet("healthcheck", flag.ContinueOnError)
	url := fs.String("url", "http://127.0.0.1:8180/healthz", "URL that must answer 200")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(*url)
	if err != nil {
		return 1
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 1
	}
	return 0
}

func run() error {
	cfg, err := config.Load(os.Getenv, config.EnvSecrets{Getenv: os.Getenv})
	if err != nil {
		return err
	}
	idCfg, err := config.LoadIdentity(os.Getenv, cfg.AppEnv)
	if err != nil {
		return err
	}
	mcpCfg, err := config.LoadMCP(os.Getenv, cfg.AppEnv, idCfg)
	if err != nil {
		return err
	}
	domCfg, err := config.LoadDomains(os.Getenv, cfg, idCfg)
	if err != nil {
		return err
	}
	logger := observability.NewLogger(os.Stdout, slog.LevelInfo, append(cfg.SecretValues(), idCfg.SecretValues()...))

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	poolCfg, err := pgxpool.ParseConfig("")
	if err != nil {
		return err
	}
	poolCfg.ConnConfig.Host = cfg.DB.Host
	poolCfg.ConnConfig.Port = uint16(cfg.DB.Port)
	poolCfg.ConnConfig.Database = cfg.DB.Name
	poolCfg.ConnConfig.User = cfg.DB.User
	poolCfg.ConnConfig.Password = cfg.DB.Password
	poolCfg.ConnConfig.ConnectTimeout = 2 * time.Second
	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return errors.New("cannot configure database pool")
	}
	defer pool.Close()

	s3Client := s3.New(s3.Options{
		BaseEndpoint: aws.String(cfg.ObjectStore.Endpoint),
		Region:       cfg.ObjectStore.Region,
		UsePathStyle: cfg.ObjectStore.ForcePathStyle,
		Credentials:  credentials.NewStaticCredentialsProvider(cfg.ObjectStore.AccessKey, cfg.ObjectStore.SecretKey, ""),
	})

	checks := []health.Check{
		health.Postgres(pool),
		health.Redis(net.JoinHostPort(cfg.Redis.Host, strconv.Itoa(cfg.Redis.Port)), cfg.Redis.Password),
		health.ObjectStore(s3Client, cfg.ObjectStore.Bucket),
	}
	if cfg.Intelligence.Enabled() {
		checks = append(checks, health.Intelligence(cfg.Intelligence.URL))
	}
	h := health.Handler{ProductName: cfg.ProductName, Checks: checks}

	// G071 (Plan r4 S3, S15): the multi-tenant root presentation and the optional default-tenant root alias, each
	// selected once from configuration before any handler is built. An invalid presentation warns and falls back to
	// signup_redirect; an unknown or invalid slug warns and stays inactive; a failed lookup stops startup
	// (default_tenant_unavailable), never treated as absence.
	rootPage, invalidRootPage := domains.RootPresentation(domCfg.RootPage)
	if invalidRootPage {
		domains.LogInvalidRootPage(ctx, logger)
	}
	domCfg.RootPage = rootPage
	defaultTenant, err := domains.SelectDefaultTenant(ctx, pool, domCfg.Hostnames(), domCfg.DefaultTenantSlug)
	if err != nil {
		return err
	}
	defaultTenant.Log(ctx, logger)

	// Identity: session resolution and local sign-out need only PostgreSQL; provider operations need a configured
	// WorkOS client (otherwise identity_unconfigured).
	// G016: the durable hostname resolver (no cache: invalidation is the commit) and the domain adapters. G071: the
	// exact platform root aliases the selected default tenant, pinned for the life of the process.
	resolver := domains.NewResolver(pool, domCfg.Hostnames()).WithDefaultTenant(defaultTenant.OrganizationID)
	var checker domains.Checker = domains.NewSystemChecker()
	if domCfg.DNSFixture != nil {
		fc, err := domains.ParseFixtureChecker(domCfg.DNSFixture)
		if err != nil {
			return errors.New("invalid configuration: CUSTOM_DOMAIN_DNS_FIXTURE_FILE")
		}
		checker = fc
	}
	var tlsAdapter domains.TLSAdapter = domains.EdgeProbe{Addr: domCfg.TLSEdgeAddr, Timeout: domCfg.TLSTimeout}
	if domCfg.TLSAdapter == "fixture" {
		fa, err := domains.ParseFixtureAdapter(domCfg.TLSFixture)
		if err != nil {
			return errors.New("invalid configuration: DOMAIN_TLS_FIXTURE_FILE")
		}
		tlsAdapter = fa
	}

	// G062 (Plan r5 §1.2, §4.1): the exact platform root origin, from configuration only. It is the only origin a
	// signup-intent login starts and completes on, and setting it here also arms the identity HTTP adapter's root login
	// start and root-origin callback limiters (§4.5 (i)-(ii)).
	rootOrigin := domCfg.RootOrigin()
	idService := identity.Config{
		Issuer: idCfg.Issuer, ClientID: idCfg.ClientID, Origins: identity.NewOrigins(idCfg.PublicOrigins),
		Hosts: originResolver{resolver}, AuthOrigin: domCfg.AuthOrigin(), RootOrigin: rootOrigin,
		Idle: idCfg.SessionIdle, Max: idCfg.SessionMax, MaxAuthAge: idCfg.MaxAuthAge, Logger: logger,
	}
	if idCfg.Configured() {
		adapter, err := workos.New(workos.Config{
			APIKey: idCfg.APIKey, ClientID: idCfg.ClientID, Issuer: idCfg.Issuer, BaseURL: idCfg.APIBaseURL, Algorithms: idCfg.Algorithms,
		})
		if err != nil {
			return errors.New("invalid configuration: IDENTITY_WORKOS_CLIENT_ID")
		}
		idService.Provider = adapter
	}
	idSvc := identity.NewService(identity.NewStore(pool), idService)
	idSvc.LogNamespaceAtStartup(ctx)
	auth := identity.NewHandler(idSvc)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", h.Liveness)
	mux.HandleFunc("GET /readyz", h.Readiness)
	mux.HandleFunc(identity.RouteLogin, auth.Login)
	mux.HandleFunc(identity.RouteCallback, auth.Callback)
	mux.HandleFunc(identity.RouteSession, auth.Session)
	mux.HandleFunc(identity.RouteLogout, auth.Logout)
	mux.HandleFunc(identity.RouteProviderReturn, auth.ProviderReturn)
	mux.HandleFunc(identity.RouteRelay, auth.Relay)

	dh, err := domainsapi.New(domainsapi.Config{Pool: pool, Identity: idSvc, Resolver: resolver, Checker: checker, TLS: tlsAdapter,
		Domains: domCfg, Logger: logger})
	if err != nil {
		return errors.New("invalid configuration: PLATFORM_BASE_DOMAIN")
	}
	dh.Register(mux)

	// G017: the branding routes, with their own access facade over the owned object store built from the loaded settings.
	bh, err := brandingapi.New(brandingapi.Config{Pool: pool, Identity: idSvc, Resolver: resolver, ObjectStore: cfg.ObjectStore,
		AppEnv: cfg.AppEnv, Logger: logger})
	if err != nil {
		return errors.New("invalid configuration: OBJECT_STORE_ENDPOINT")
	}
	bh.Register(mux)

	// G018: the Group onboarding routes, with their own access facade (every route one guarded operation).
	oh, err := onboardingapi.New(onboardingapi.Config{Pool: pool, Identity: idSvc, Resolver: resolver, Logger: logger})
	if err != nil {
		return errors.New("invalid configuration: onboarding")
	}
	oh.Register(mux)

	// G019: the membership, ownership and Group invitation routes and the own retirement eligibility, with their own
	// access facade (every route one guarded operation; no retire or Organization-invitation consumption route).
	membh, err := membershipapi.New(membershipapi.Config{Pool: pool, Identity: idSvc, Resolver: resolver, Logger: logger})
	if err != nil {
		return errors.New("invalid configuration: membership")
	}
	membh.Register(mux)

	// G020: the generic label routes (definitions in the Organization and Group scopes; Organization-owned labels on
	// Groups), with their own access facade (every route one guarded operation; no Group-owned assignment route).
	labh, err := labelsapi.New(labelsapi.Config{Pool: pool, Identity: idSvc, Resolver: resolver, Logger: logger})
	if err != nil {
		return errors.New("invalid configuration: labels")
	}
	labh.Register(mux)

	// G058: the callback-time enrollment admission, composed with its own access facade and bound into identity once
	// before serving (a failed binding stops startup; unbound, every admission is identity_unavailable), and the
	// enrollment routes under /internal/v1/enrollment (public projection, the applicant's own credential, the manager's
	// policy and requests).
	enh, err := enrollmentapi.New(enrollmentapi.Config{Pool: pool, Identity: idSvc, Resolver: resolver, Logger: logger})
	if err != nil {
		return errors.New("invalid configuration: enrollment")
	}
	if err := enh.Bind(); err != nil {
		return errors.New("invalid configuration: enrollment")
	}
	enh.Register(mux)

	// G062: the platform root signup routes under /internal/v1/signup (the signup service takes the user only from a
	// resolved root signup credential and every database context from identity; its tenant targets come from the
	// platform domain configuration).
	suh, err := signupapi.New(signupapi.Config{Service: signup.NewService(signup.Config{Identity: idSvc, Hostname: domCfg.Hostnames(),
		RootOrigin: rootOrigin}), Logger: logger})
	if err != nil {
		return errors.New("invalid configuration: signup")
	}
	suh.Register(mux)

	// MCP (G014): mounted only when configured; otherwise no MCP route exists.
	var handler http.Handler = mux
	if mcpCfg.Configured() {
		verifier, err := workos.NewConnectValidator(workos.ConnectConfig{Issuer: mcpCfg.Issuer, Algorithms: mcpCfg.Algorithms, JWKS: mcpCfg.TestJWKS})
		if err != nil {
			return errors.New("invalid configuration: MCP_CONNECT_ISSUER")
		}
		mh, err := mcp.New(mcp.Config{
			Issuer: mcpCfg.Issuer, Hosts: resolver, Verifier: verifier, Identity: idSvc, Pool: pool,
			Lifetimes:      identity.MCPLifetimes{Pending: mcpCfg.PendingTTL, Idle: mcpCfg.Idle, Absolute: mcpCfg.Absolute},
			AllowedOrigins: mcpCfg.AllowedOrigins, Title: cfg.ProductName, Version: version, Logger: logger,
		})
		if err != nil {
			return errors.New("invalid configuration: MCP_CONNECT_ISSUER")
		}
		mux.HandleFunc(mcp.RouteConnectionsList, mh.List)
		mux.HandleFunc(mcp.RouteConnectionsCreate, mh.Create)
		mux.HandleFunc(mcp.RouteConnectionsRevoke, mh.Revoke)
		mux.HandleFunc(mcp.RouteConnectionsRename, mh.Rename)
		handler = mh.Wrap(mux)
	}

	srv := &http.Server{
		Addr:              cfg.ListenAddr,
		Handler:           observability.Middleware(logger, handler),
		ReadHeaderTimeout: 5 * time.Second,
		ErrorLog:          observability.ServerErrorLog(logger),
	}
	ln, err := net.Listen("tcp", cfg.ListenAddr)
	if err != nil {
		return err
	}
	errCh := make(chan error, 1)
	go func() { errCh <- srv.Serve(ln) }()
	logger.Info("core listening", slog.Int("port", observability.ListenPort(ln.Addr())), slog.String("app_env", cfg.AppEnv))

	select {
	case <-ctx.Done():
	case err := <-errCh:
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	}
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return srv.Shutdown(shutdownCtx)
}

// originResolver adapts the durable resolver to the interface identity declares itself (identity never imports the
// domains package; G016 Plan r4 §5.0, Codex constraint 7): one statement on the caller's querier, exact origins only.
type originResolver struct{ r *domains.Resolver }

func (o originResolver) ResolveOrigin(ctx context.Context, q identity.Querier, origin string) (identity.HostResolution, error) {
	host, ok := strings.CutPrefix(origin, "https://")
	if !ok || host == "" || strings.ContainsAny(host, "/?#@") {
		return identity.HostResolution{}, nil
	}
	res, err := o.r.ResolveWith(ctx, q, host)
	if err != nil {
		return identity.HostResolution{}, err
	}
	if !res.Routing || res.Host.Origin() != origin {
		return identity.HostResolution{}, nil
	}
	return identity.HostResolution{Kind: string(res.Kind), OrganizationID: res.OrganizationID, Routing: true}, nil
}
