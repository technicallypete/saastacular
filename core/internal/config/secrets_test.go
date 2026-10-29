package config

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// The four secrets Load acquires through the SecretProvider (G012 Plan r1 §2.1).
var loadSecretNames = []string{"DB_PASSWORD", "REDIS_PASSWORD", "OBJECT_STORE_SECRET_KEY", "INTELLIGENCE_SERVICE_TOKEN"}

// spyProvider is a test-only replacement SecretProvider: values come from a map, failures from errs, and every call
// is counted.
type spyProvider struct {
	values map[string]string
	errs   map[string]error
	calls  map[string]int
}

func (s *spyProvider) Secret(name string) (string, error) {
	if s.calls == nil {
		s.calls = map[string]int{}
	}
	s.calls[name]++
	if err, ok := s.errs[name]; ok {
		return "", err
	}
	return s.values[name], nil
}

func canary(t *testing.T) string {
	t.Helper()
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		t.Fatal(err)
	}
	return "Cx" + hex.EncodeToString(b)
}

// envWithoutSecrets is baseEnv minus every secret, so only a replacement provider can supply them.
func envWithoutSecrets() map[string]string {
	env := baseEnv()
	for _, n := range loadSecretNames {
		delete(env, n)
	}
	return env
}

func providerValues(t *testing.T) map[string]string {
	return map[string]string{
		"DB_PASSWORD": canary(t), "REDIS_PASSWORD": canary(t), "OBJECT_STORE_SECRET_KEY": canary(t),
		"INTELLIGENCE_SERVICE_TOKEN": canary(t),
	}
}

// U-SEC-1: EnvSecrets succeeds through Load (TestLoadValid) and reports a missing value by name only.
func TestEnvSecretsMissingIsNotSet(t *testing.T) {
	_, err := EnvSecrets{Getenv: func(string) string { return "" }}.Secret("DB_PASSWORD")
	if !errors.Is(err, ErrSecretNotSet) {
		t.Fatalf("EnvSecrets missing value: want ErrSecretNotSet, got %v", err)
	}
	env := baseEnv()
	delete(env, "REDIS_PASSWORD")
	_, err = load(env)
	if err == nil || !strings.Contains(err.Error(), "REDIS_PASSWORD is not set") {
		t.Fatalf("want REDIS_PASSWORD is not set, got %v", err)
	}
}

// U-SEC-2: a replacement provider supplies all four Load secrets while the environment holds none, through the same
// loader; EnvSecrets over that environment fails.
func TestReplacementProviderThroughLoad(t *testing.T) {
	env := envWithoutSecrets()
	env["INTELLIGENCE_URL"] = "http://intelligence:8000"
	get := func(k string) string { return env[k] }
	vals := providerValues(t)
	c, err := Load(get, &spyProvider{values: vals})
	if err != nil {
		t.Fatalf("replacement provider: %v", err)
	}
	if c.DB.Password != vals["DB_PASSWORD"] || c.Redis.Password != vals["REDIS_PASSWORD"] ||
		c.ObjectStore.SecretKey != vals["OBJECT_STORE_SECRET_KEY"] || c.Intelligence.Token != vals["INTELLIGENCE_SERVICE_TOKEN"] {
		t.Fatal("values did not come from the replacement provider")
	}
	want := []string{vals["DB_PASSWORD"], vals["REDIS_PASSWORD"], vals["OBJECT_STORE_SECRET_KEY"], vals["INTELLIGENCE_SERVICE_TOKEN"]}
	if !reflect.DeepEqual(c.SecretValues(), want) {
		t.Fatal("SecretValues does not list the provider values")
	}
	if _, err := Load(get, EnvSecrets{Getenv: get}); err == nil {
		t.Fatal("EnvSecrets over an environment without secrets must fail")
	}
}

// U-SEC-3: every provider outcome maps to exactly one closed problem text; provider error text (carrying a canary)
// never reaches the error; Load returns the zero Config.
func TestSecretOutcomesAreClosedAndUndisclosed(t *testing.T) {
	c := canary(t)
	other := providerValues(t)
	cases := []struct {
		name  string
		value string
		err   error
		want  string
	}{
		{"wrapped not set", "", fmt.Errorf("vault path /kv/%s: %w", c, ErrSecretNotSet), "DB_PASSWORD is not set"},
		{"empty without error", "", nil, "DB_PASSWORD is not set"},
		{"whitespace only", " \t ", nil, "DB_PASSWORD is empty"},
		{"NUL", "ab\x00cd", nil, "DB_PASSWORD is invalid"},
		{"CR", "abc\r", nil, "DB_PASSWORD is invalid"},
		{"LF", "ab\ncd", nil, "DB_PASSWORD is invalid"},
		{"denied", "", fmt.Errorf("policy %s refused: %w", c, ErrSecretDenied), "DB_PASSWORD is denied by the secret provider"},
		{"ambiguous", "", fmt.Errorf("two versions %s: %w", c, ErrSecretAmbiguous), "DB_PASSWORD is ambiguous"},
		{"unavailable", "", fmt.Errorf("dial %s: %w", c, ErrSecretUnavailable), "DB_PASSWORD is unavailable from the secret provider"},
		{"other error", "", errors.New("backend said " + c), "DB_PASSWORD is unavailable from the secret provider"},
		{"other type", "", &os.PathError{Op: "open", Path: "/run/" + c, Err: fs.ErrPermission}, "DB_PASSWORD is unavailable from the secret provider"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			env := envWithoutSecrets()
			get := func(k string) string { return env[k] }
			vals := map[string]string{"REDIS_PASSWORD": other["REDIS_PASSWORD"], "OBJECT_STORE_SECRET_KEY": other["OBJECT_STORE_SECRET_KEY"]}
			p := &spyProvider{values: vals, errs: map[string]error{}}
			if tc.err != nil {
				p.errs["DB_PASSWORD"] = tc.err
			} else {
				p.values["DB_PASSWORD"] = tc.value
			}
			cfg, err := Load(get, p)
			if err == nil {
				t.Fatal("expected an error")
			}
			if !reflect.DeepEqual(cfg, Config{}) {
				t.Fatal("a failed Load must return the zero Config")
			}
			msg := err.Error()
			if msg != "invalid configuration: "+tc.want {
				t.Fatalf("got %q, want exactly %q", msg, "invalid configuration: "+tc.want)
			}
			for _, leak := range []string{c, other["REDIS_PASSWORD"], other["OBJECT_STORE_SECRET_KEY"], "vault", "policy", "backend", "/run/"} {
				if strings.Contains(msg, leak) {
					t.Fatalf("error discloses provider or secret text: %q", msg)
				}
			}
		})
	}
	if _, problem := resolveSecret(nil, "DB_PASSWORD"); problem != "DB_PASSWORD is unavailable from the secret provider" {
		t.Fatalf("nil provider: %q", problem)
	}
}

// A valid value is returned byte-exact, never trimmed.
func TestSecretValueIsByteExact(t *testing.T) {
	v := " " + canary(t) + "\t"
	got, problem := resolveSecret(&spyProvider{values: map[string]string{"X": v}}, "X")
	if problem != "" || got != v {
		t.Fatalf("value altered or refused: problem %q", problem)
	}
}

// U-SEC-4: with INTELLIGENCE_URL unset the token is never requested; with it set an unavailable provider fails.
func TestIntelligenceTokenOnlyWhenEnabled(t *testing.T) {
	env := envWithoutSecrets()
	get := func(k string) string { return env[k] }
	p := &spyProvider{values: providerValues(t)}
	if _, err := Load(get, p); err != nil {
		t.Fatalf("disabled intelligence: %v", err)
	}
	if p.calls["INTELLIGENCE_SERVICE_TOKEN"] != 0 {
		t.Fatal("the intelligence token was requested although INTELLIGENCE_URL is unset")
	}
	env["INTELLIGENCE_URL"] = "http://intelligence:8000"
	p = &spyProvider{values: providerValues(t), errs: map[string]error{"INTELLIGENCE_SERVICE_TOKEN": ErrSecretUnavailable}}
	if _, err := Load(get, p); err == nil || !strings.Contains(err.Error(), "INTELLIGENCE_SERVICE_TOKEN is unavailable from the secret provider") {
		t.Fatalf("enabled intelligence with an unavailable provider: %v", err)
	}
}

// U-SEC-5: exactly one provider call per required name.
func TestOneProviderCallPerName(t *testing.T) {
	env := envWithoutSecrets()
	env["INTELLIGENCE_URL"] = "http://intelligence:8000"
	get := func(k string) string { return env[k] }
	p := &spyProvider{values: providerValues(t)}
	if _, err := Load(get, p); err != nil {
		t.Fatal(err)
	}
	want := map[string]int{}
	for _, n := range loadSecretNames {
		want[n] = 1
	}
	if !reflect.DeepEqual(p.calls, want) {
		t.Fatalf("provider calls %v, want %v", p.calls, want)
	}
}

// U-SEC-6: acquisition stays at startup in config.Load. No non-test core file outside this package and the server
// main references the provider boundary or calls .Secret(; inside this package only resolveSecret calls it.
func TestSecretAcquisitionBoundary(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err != nil {
		t.Fatalf("core module root not found at %s", root)
	}
	inPkg, resolveCalls, loadCalls := 0, map[string]int{}, map[string]int{}
	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if n := d.Name(); n == "testdata" || n == "vendor" || strings.HasPrefix(n, ".") && path != root {
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
		src, rel := string(b), strings.TrimPrefix(path, root+string(filepath.Separator))
		rel = filepath.ToSlash(rel)
		switch {
		case strings.HasPrefix(rel, "internal/config/"):
			n := strings.Count(src, ".Secret(")
			inPkg += n
			if n > 0 && rel != "internal/config/secrets.go" {
				t.Errorf("%s calls .Secret( outside resolveSecret", rel)
			}
			if c := strings.Count(src, "resolveSecret(") - strings.Count(src, "func resolveSecret("); c > 0 {
				resolveCalls[rel] += c
			}
		case rel == "cmd/server/main.go":
			if strings.Contains(src, ".Secret(") {
				t.Errorf("%s calls .Secret( directly", rel)
			}
			loadCalls[rel] += strings.Count(src, "config.Load(")
		default:
			for _, sym := range []string{"SecretProvider", "EnvSecrets", ".Secret(", "config.Load("} {
				if strings.Contains(src, sym) {
					t.Errorf("%s references %s outside the configuration boundary", rel, sym)
				}
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if inPkg != 1 {
		t.Fatalf("want exactly one .Secret( call in the config package (resolveSecret), found %d", inPkg)
	}
	// resolveSecret is called only by Load (config.go), and only the server's startup calls Load.
	if !reflect.DeepEqual(resolveCalls, map[string]int{"internal/config/config.go": 1}) {
		t.Fatalf("resolveSecret call sites %v, want exactly one in internal/config/config.go", resolveCalls)
	}
	if !reflect.DeepEqual(loadCalls, map[string]int{"cmd/server/main.go": 1}) {
		t.Fatalf("config.Load call sites %v, want exactly one in cmd/server/main.go", loadCalls)
	}
}
