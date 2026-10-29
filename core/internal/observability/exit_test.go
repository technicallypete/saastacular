package observability

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"testing"

	"core/internal/config"
)

// loadConfig runs config.Load against env and records every variable name it
// reads, through both the environment and the secret provider.
func loadConfig(env map[string]string) (map[string]bool, error) {
	read := map[string]bool{}
	get := func(k string) string { read[k] = true; return env[k] }
	_, err := config.Load(get, config.EnvSecrets{Getenv: get})
	return read, err
}

func exitRecord(t *testing.T, err error) (map[string]any, string) {
	t.Helper()
	var buf bytes.Buffer
	LogExit(&buf, err)
	out := buf.String()
	if strings.Count(out, "\n") != 1 || !strings.HasSuffix(out, "\n") {
		t.Fatalf("want exactly one line, got %q", out)
	}
	recs := decodeRecords(t, out)
	if len(recs) != 1 {
		t.Fatalf("want one record: %q", out)
	}
	r := recs[0]
	assertEnvelope(t, r)
	if r["event"] != "core.exit" || r["level"] != "error" {
		t.Fatalf("record = %v", r)
	}
	want := []string{"error_type", "event", "level", "service", "time"}
	if r["error_type"] == "config_invalid" {
		want = append(want, "config_keys")
		sort.Strings(want)
	}
	if got := sortedKeys(r); !slices.Equal(got, want) {
		t.Fatalf("keys = %v, want %v", got, want)
	}
	return r, out
}

func configKeysOf(t *testing.T, r map[string]any) []string {
	t.Helper()
	raw, ok := r["config_keys"].([]any)
	if !ok {
		t.Fatalf("config_keys is not an array: %v", r["config_keys"])
	}
	keys := make([]string, 0, len(raw))
	for _, v := range raw {
		s, ok := v.(string)
		if !ok || !configKeys[s] {
			t.Fatalf("config key %v not allow-listed", v)
		}
		keys = append(keys, s)
	}
	if !slices.IsSorted(keys) || len(slices.Compact(slices.Clone(keys))) != len(keys) {
		t.Fatalf("config_keys not sorted and unique: %v", keys)
	}
	return keys
}

func TestLogExitCategories(t *testing.T) {
	cn := newCanary()
	upper := strings.ToUpper(cn)

	_, emptyErr := loadConfig(map[string]string{})
	if emptyErr == nil {
		t.Fatal("config.Load accepted an empty environment")
	}
	busy, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer busy.Close()
	_, inUseErr := net.Listen("tcp", busy.Addr().String())
	if inUseErr == nil {
		t.Fatal("second listen on a bound address succeeded")
	}

	for _, tc := range []struct {
		name string
		err  error
		want string
	}{
		{"config.Load with an empty environment", emptyErr, "config_invalid"},
		{"config error with canary text", errors.New("invalid configuration: PRODUCT_NAME is required; " + cn + " " + upper + "_KEY"), "config_invalid"},
		{"wrapped config error", fmt.Errorf("startup %s: %w", cn, emptyErr), "config_invalid"},
		{"database pool", errors.New("cannot configure database pool"), "database_pool_invalid"},
		{"wrapped database pool", fmt.Errorf("%s: %w", cn, errors.New("cannot configure database pool")), "database_pool_invalid"},
		{"listen address in use", inUseErr, "listen_failed"},
		{"listen with canary host", &net.OpError{Op: "listen", Net: "tcp", Err: errors.New("lookup bad-" + cn + ".invalid")}, "listen_failed"},
		{"wrapped listen", fmt.Errorf("%s: %w", cn, &net.OpError{Op: "listen", Net: "tcp", Err: errors.New(cn)}), "listen_failed"},
		{"accept", &net.OpError{Op: "accept", Net: "tcp", Err: errors.New(cn)}, "serve_failed"},
		{"shutdown deadline", context.DeadlineExceeded, "shutdown_timeout"},
		{"wrapped shutdown deadline", fmt.Errorf("%s: %w", cn, context.DeadlineExceeded), "shutdown_timeout"},
		{"unknown", errors.New(cn), "unknown"},
		{"wrapped unknown", fmt.Errorf("x %s: %w", cn, errors.New(upper)), "unknown"},
		{"pool text is not a prefix match", errors.New("cannot configure database pool " + cn), "unknown"},
		// G071 (Plan r4 S3(e), S14; U4): the fixed default-tenant lookup error, bare or wrapped, is its own category;
		// a near miss is not.
		{"default tenant lookup", errors.New("default tenant lookup unavailable"), "default_tenant_unavailable"},
		{"wrapped default tenant lookup", fmt.Errorf("%s: %w", cn, errors.New("default tenant lookup unavailable")), "default_tenant_unavailable"},
		{"default tenant text is not a prefix match", errors.New("default tenant lookup unavailable " + cn), "unknown"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r, out := exitRecord(t, tc.err)
			if r["error_type"] != tc.want {
				t.Fatalf("error_type = %v, want %s", r["error_type"], tc.want)
			}
			if tc.want == "config_invalid" {
				if keys := configKeysOf(t, r); len(keys) == 0 {
					t.Fatal("config_keys is empty")
				}
			}
			if strings.Contains(out, cn) || strings.Contains(out, upper) || strings.Contains(out, "127.0.0.1") {
				t.Fatalf("error text leaked: %s", out)
			}
			if strings.Contains(out, tc.err.Error()) {
				t.Fatalf("err.Error() written: %s", out)
			}
		})
	}

	// G071 (U4): neither DEFAULT_TENANT_SLUG nor PLATFORM_ROOT_PAGE is ever a config_keys member (both are never a
	// configuration error, and the allow-list stays exactly the names config.Load and config.LoadIdentity read).
	for _, name := range []string{"DEFAULT_TENANT_SLUG", "PLATFORM_ROOT_PAGE"} {
		if configKeys[name] {
			t.Fatalf("%s is allow-listed for config_keys", name)
		}
		r, out := exitRecord(t, errors.New("invalid configuration: "+name+" must be set; DB_PORT is required"))
		if got := configKeysOf(t, r); !slices.Equal(got, []string{"DB_PORT"}) || strings.Contains(out, name) {
			t.Fatalf("config_keys = %v (%s must never be reported): %s", got, name, out)
		}
	}
	r, out := exitRecord(t, errors.New("default tenant lookup unavailable"))
	if r["error_type"] != "default_tenant_unavailable" || r["config_keys"] != nil || strings.Contains(out, "lookup") {
		t.Fatalf("default_tenant_unavailable record = %s", out)
	}
}

func TestLogExitConfigKeysFromMessage(t *testing.T) {
	r, _ := exitRecord(t, errors.New("invalid configuration: secret REDIS_PASSWORD is not set; DB_PORT must be a port number; DB_PORT is required; NOT_A_KEY is required"))
	if got := configKeysOf(t, r); !slices.Equal(got, []string{"DB_PORT", "REDIS_PASSWORD"}) {
		t.Fatalf("config_keys = %v", got)
	}
	// Identity names are exact: a prefix, near miss or unknown IDENTITY_ name is never reported.
	r, _ = exitRecord(t, errors.New("invalid configuration: IDENTITY_ is required; IDENTITY_WORKOS is required; "+
		"IDENTITY_UNKNOWN is required; IDENTITY_WORKOS_API_KEY_FILES is required; XIDENTITY_PROVIDER must be workos; "+
		"IDENTITY_WORKOS_API_KEY_FILE cannot be read"))
	if got := configKeysOf(t, r); !slices.Equal(got, []string{"IDENTITY_WORKOS_API_KEY_FILE"}) {
		t.Fatalf("config_keys = %v", got)
	}
	r, out := exitRecord(t, errors.New("invalid configuration: "))
	if got := configKeysOf(t, r); len(got) != 0 || !strings.Contains(out, `"config_keys":[]`) {
		t.Fatalf("empty config_keys must be an empty array: %s", out)
	}
}

// TestConfigKeysAllowListMatchesConfigLoad keeps the allow-list in sync with
// config.Load and config.LoadIdentity: the names the two loaders read equal
// the allow-list, and every variable name in Load's messages is allow-listed
// and reported.
func TestConfigKeysAllowListMatchesConfigLoad(t *testing.T) {
	read, emptyErr := loadConfig(map[string]string{})
	if emptyErr == nil {
		t.Fatal("config.Load accepted an empty environment")
	}
	idRead, err := loadIdentity(t, "test", validIdentityEnv())
	if err != nil {
		t.Fatalf("config.LoadIdentity rejected a valid configuration: %v", err)
	}
	var readNames, allowed []string
	for k := range read {
		readNames = append(readNames, k)
	}
	for k := range idRead {
		if read[k] {
			t.Fatalf("config.Load and config.LoadIdentity both read %s", k)
		}
		readNames = append(readNames, k)
	}
	for k := range configKeys {
		allowed = append(allowed, k)
	}
	sort.Strings(readNames)
	sort.Strings(allowed)
	if !slices.Equal(readNames, allowed) {
		t.Fatalf("allow-list drifted from config.Load and config.LoadIdentity:\nread    %v\nallowed %v", readNames, allowed)
	}

	invalid := map[string]string{
		"PRODUCT_NAME": "x", "PLATFORM_BASE_DOMAIN": "Not_A_Host", "CUSTOM_DOMAIN_CNAME_TARGET": "single",
		"DATABASE_NAME": "Bad-Name", "APP_ENV": "staging", "DB_HOST": "h", "DB_PORT": "0", "DB_USER": "u",
		"REDIS_HOST": "h", "REDIS_PORT": "x", "OBJECT_STORE_FORCE_PATH_STYLE": "maybe",
		"INTELLIGENCE_SERVICE_TOKEN": "t",
	}
	withURL := map[string]string{"INTELLIGENCE_URL": "http://intelligence:8000"}

	for _, env := range []map[string]string{{}, invalid, withURL} {
		_, err := loadConfig(env)
		if err == nil {
			t.Fatalf("config.Load accepted %v", env)
		}
		msg := err.Error()
		if !strings.HasPrefix(msg, configErrorPrefix) {
			t.Fatalf("config.Load message lost its prefix: %q", msg)
		}
		var names []string
		for _, n := range configKeyRe.FindAllString(strings.TrimPrefix(msg, configErrorPrefix), -1) {
			if !configKeys[n] {
				t.Fatalf("config.Load names %q, which is not allow-listed", n)
			}
			if !slices.Contains(names, n) {
				names = append(names, n)
			}
		}
		sort.Strings(names)
		r, _ := exitRecord(t, err)
		if got := configKeysOf(t, r); !slices.Equal(got, names) {
			t.Fatalf("config_keys = %v, want %v", got, names)
		}
	}
}

// identityKeys are the twelve identity inputs (G004), listed independently of
// config.LoadIdentity and of the allow-list.
var identityKeys = []string{
	"IDENTITY_COOKIE_PROFILE", "IDENTITY_JWT_ALGORITHMS", "IDENTITY_MAX_AUTH_AGE", "IDENTITY_PROVIDER",
	"IDENTITY_PUBLIC_ORIGINS", "IDENTITY_SESSION_ABSOLUTE", "IDENTITY_SESSION_IDLE", "IDENTITY_WORKOS_API_BASE_URL",
	"IDENTITY_WORKOS_API_KEY", "IDENTITY_WORKOS_API_KEY_FILE", "IDENTITY_WORKOS_CLIENT_ID", "IDENTITY_WORKOS_ISSUER",
}

const (
	exitTestClient = "client_01EXITTESTCLIENT00000000"
	exitTestIssuer = "https://api.workos.com/user_management/client_01EXITTESTCLIENT00000000"
	exitTestKey    = "sk_test_" + "exitconfigfixture" // assembled, so no key-shaped literal is committed
)

// baseConfigEnv is a configuration config.Load accepts for appEnv.
func baseConfigEnv(appEnv string) map[string]string {
	return map[string]string{
		"PRODUCT_NAME": "Sample Product", "PLATFORM_BASE_DOMAIN": "sample.example.test",
		"CUSTOM_DOMAIN_CNAME_TARGET": "custom.sample.example.test", "DATABASE_NAME": "sample_app", "APP_ENV": appEnv,
		"DB_HOST": "postgres", "DB_PORT": "5432", "DB_USER": "sample_app_runtime", "DB_PASSWORD": "db-secret-value",
		"REDIS_HOST": "redis", "REDIS_PORT": "6379", "REDIS_PASSWORD": "redis-secret-value",
		"OBJECT_STORE_ENDPOINT": "https://minio:9000", "OBJECT_STORE_REGION": "us-east-1",
		"OBJECT_STORE_BUCKET": "sample-objects", "OBJECT_STORE_FORCE_PATH_STYLE": "true",
		"OBJECT_STORE_ACCESS_KEY": "core-service", "OBJECT_STORE_SECRET_KEY": "object-secret-value",
	}
}

// validIdentityEnv is a complete identity configuration that sets every
// identity input except the API key file, which excludes the inline key.
func validIdentityEnv() map[string]string {
	return map[string]string{
		"IDENTITY_PROVIDER": "workos", "IDENTITY_WORKOS_CLIENT_ID": exitTestClient,
		"IDENTITY_WORKOS_ISSUER": exitTestIssuer, "IDENTITY_WORKOS_API_KEY": exitTestKey,
		"IDENTITY_WORKOS_API_BASE_URL": "https://api.workos.com", "IDENTITY_PUBLIC_ORIGINS": "https://org-a.localhost:543",
		"IDENTITY_SESSION_IDLE": "1h", "IDENTITY_SESSION_ABSOLUTE": "8h", "IDENTITY_JWT_ALGORITHMS": "RS256",
		"IDENTITY_MAX_AUTH_AGE": "10m", "IDENTITY_COOKIE_PROFILE": "https",
	}
}

// loadIdentity runs config.Load and then config.LoadIdentity, as cmd/server
// does, with a valid base configuration for appEnv overlaid by identity. It
// records every variable name LoadIdentity reads; an empty identity value
// leaves the variable unset.
func loadIdentity(t *testing.T, appEnv string, identity map[string]string) (map[string]bool, error) {
	t.Helper()
	env := baseConfigEnv(appEnv)
	for k, v := range identity {
		if v != "" {
			env[k] = v
		}
	}
	get := func(k string) string { return env[k] }
	cfg, err := config.Load(get, config.EnvSecrets{Getenv: get})
	if err != nil {
		t.Fatalf("config.Load rejected the base configuration: %v", err)
	}
	read := map[string]bool{}
	_, err = config.LoadIdentity(func(k string) string { read[k] = true; return env[k] }, cfg.AppEnv)
	return read, err
}

func readNames(read map[string]bool) []string {
	names := make([]string, 0, len(read))
	for k := range read {
		names = append(names, k)
	}
	sort.Strings(names)
	return names
}

// TestLogExitIdentityConfigKeys exercises config.LoadIdentity behind a valid
// base configuration, through its unconfigured, configured and partially
// configured paths. Every path reads exactly the twelve identity inputs; every
// rejection yields a config_invalid record naming exactly the expected inputs,
// never a value, file name or message text; and together the rejections name
// all twelve.
func TestLogExitIdentityConfigKeys(t *testing.T) {
	cn := newCanary()
	dir := t.TempDir()
	keyFile := filepath.Join(dir, "key-"+cn)
	if err := os.WriteFile(keyFile, []byte("sk_test_"+cn+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	emptyFile := filepath.Join(dir, "empty-"+cn)
	if err := os.WriteFile(emptyFile, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	with := func(over map[string]string) map[string]string {
		env := validIdentityEnv()
		for k, v := range over {
			env[k] = v
		}
		return env
	}

	for _, tc := range []struct {
		name     string
		appEnv   string
		identity map[string]string
	}{
		{"unconfigured", "development", map[string]string{}},
		{"configured", "test", validIdentityEnv()},
		{"configured with a key file", "production",
			with(map[string]string{"IDENTITY_WORKOS_API_KEY": "", "IDENTITY_WORKOS_API_KEY_FILE": keyFile})},
	} {
		t.Run("accepts "+tc.name, func(t *testing.T) {
			read, err := loadIdentity(t, tc.appEnv, tc.identity)
			if err != nil {
				t.Fatalf("config.LoadIdentity rejected it: %v", err)
			}
			if got := readNames(read); !slices.Equal(got, identityKeys) {
				t.Fatalf("read %v, want %v", got, identityKeys)
			}
		})
	}

	named := map[string]bool{}
	for _, tc := range []struct {
		name     string
		appEnv   string
		identity map[string]string
		want     []string
	}{
		{"partial: client only", "development",
			with(map[string]string{"IDENTITY_WORKOS_ISSUER": "", "IDENTITY_WORKOS_API_KEY": ""}),
			[]string{"IDENTITY_WORKOS_API_KEY", "IDENTITY_WORKOS_ISSUER"}},
		{"partial: no client, malformed issuer", "development",
			with(map[string]string{"IDENTITY_WORKOS_CLIENT_ID": "", "IDENTITY_WORKOS_ISSUER": "http://" + cn + ".invalid/x"}),
			[]string{"IDENTITY_WORKOS_CLIENT_ID", "IDENTITY_WORKOS_ISSUER"}},
		{"provider and both key sources", "test",
			with(map[string]string{"IDENTITY_PROVIDER": "oidc-" + cn, "IDENTITY_WORKOS_API_KEY_FILE": keyFile}),
			[]string{"IDENTITY_PROVIDER", "IDENTITY_WORKOS_API_KEY", "IDENTITY_WORKOS_API_KEY_FILE"}},
		{"unreadable key file", "test",
			with(map[string]string{"IDENTITY_WORKOS_API_KEY": "", "IDENTITY_WORKOS_API_KEY_FILE": filepath.Join(dir, "missing-"+cn)}),
			[]string{"IDENTITY_WORKOS_API_KEY", "IDENTITY_WORKOS_API_KEY_FILE"}},
		{"empty key file", "test",
			with(map[string]string{"IDENTITY_WORKOS_API_KEY": "", "IDENTITY_WORKOS_API_KEY_FILE": emptyFile}),
			[]string{"IDENTITY_WORKOS_API_KEY", "IDENTITY_WORKOS_API_KEY_FILE"}},
		{"split inline key", "test",
			with(map[string]string{"IDENTITY_WORKOS_API_KEY": "sk_test_" + cn + " " + cn}),
			[]string{"IDENTITY_WORKOS_API_KEY"}},
		{"origins, lifetimes, algorithms", "development",
			with(map[string]string{"IDENTITY_PUBLIC_ORIGINS": "http://" + cn + ".example", "IDENTITY_SESSION_IDLE": "30h",
				"IDENTITY_SESSION_ABSOLUTE": "25h", "IDENTITY_MAX_AUTH_AGE": "-" + cn, "IDENTITY_JWT_ALGORITHMS": "HS256"}),
			[]string{"APP_ENV", "IDENTITY_JWT_ALGORITHMS", "IDENTITY_MAX_AUTH_AGE", "IDENTITY_PUBLIC_ORIGINS",
				"IDENTITY_SESSION_ABSOLUTE", "IDENTITY_SESSION_IDLE"}},
		{"production refusals", "production",
			with(map[string]string{"IDENTITY_COOKIE_PROFILE": "loopback-http", "IDENTITY_PUBLIC_ORIGINS": "http://localhost:3100",
				"IDENTITY_WORKOS_API_BASE_URL": "https://" + cn + ".example"}),
			[]string{"APP_ENV", "IDENTITY_COOKIE_PROFILE", "IDENTITY_WORKOS_API_BASE_URL"}},
		{"unconfigured with settings", "production",
			map[string]string{"IDENTITY_SESSION_ABSOLUTE": "8h", "IDENTITY_COOKIE_PROFILE": cn},
			[]string{"IDENTITY_COOKIE_PROFILE", "IDENTITY_SESSION_ABSOLUTE", "IDENTITY_WORKOS_CLIENT_ID"}},
	} {
		t.Run("rejects "+tc.name, func(t *testing.T) {
			read, err := loadIdentity(t, tc.appEnv, tc.identity)
			if err == nil {
				t.Fatal("config.LoadIdentity accepted it")
			}
			if got := readNames(read); !slices.Equal(got, identityKeys) {
				t.Fatalf("read %v, want %v", got, identityKeys)
			}
			msg := err.Error()
			if !strings.HasPrefix(msg, configErrorPrefix) {
				t.Fatalf("config.LoadIdentity message lost its prefix: %q", msg)
			}
			for _, n := range configKeyRe.FindAllString(msg, -1) {
				if (read[n] || n == "APP_ENV") && !slices.Contains(tc.want, n) {
					t.Fatalf("message names %s, which the case does not expect: %q", n, msg)
				}
			}
			r, out := exitRecord(t, err)
			if r["error_type"] != "config_invalid" {
				t.Fatalf("error_type = %v", r["error_type"])
			}
			got := configKeysOf(t, r)
			if !slices.Equal(got, tc.want) {
				t.Fatalf("config_keys = %v, want %v", got, tc.want)
			}
			for _, k := range got {
				named[k] = true
			}
			if strings.Contains(out, cn) || strings.Contains(out, strings.ToUpper(cn)) || strings.Contains(out, dir) ||
				strings.Contains(out, "sk_test_") || strings.Contains(out, msg) {
				t.Fatalf("value, file name or message written: %s", out)
			}
		})
	}
	for _, k := range identityKeys {
		if !named[k] {
			t.Errorf("no rejection named %s", k)
		}
	}
}

func TestLogExitRecordIsValidJSONWithTypes(t *testing.T) {
	var buf bytes.Buffer
	LogExit(&buf, errors.New("anything"))
	var m map[string]json.RawMessage
	if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if string(m["error_type"]) != `"unknown"` {
		t.Fatalf("error_type = %s", m["error_type"])
	}
}

func TestListenPort(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	for _, tc := range []struct {
		name string
		addr net.Addr
		want int
	}{
		{"TCPAddr", &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 8180}, 8180},
		{"bound listener", ln.Addr(), ln.Addr().(*net.TCPAddr).Port},
		{"UnixAddr", &net.UnixAddr{Name: "/tmp/core.sock", Net: "unix"}, 0},
		{"UDPAddr", &net.UDPAddr{Port: 8180}, 0},
		{"nil", nil, 0},
		{"typed nil TCPAddr", (*net.TCPAddr)(nil), 0},
	} {
		if got := ListenPort(tc.addr); got != tc.want {
			t.Errorf("%s: ListenPort = %d, want %d", tc.name, got, tc.want)
		}
	}
	if ListenPort(ln.Addr()) == 0 {
		t.Fatal("bound port is 0")
	}
}
