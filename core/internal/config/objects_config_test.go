package config

import (
	"strings"
	"testing"
	"time"
)

// U-OBJ-CONFIG (G009 Plan r1 §4.7, §5.1): the optional storage settings' defaults, bounds and variable-name-only
// problems, and the endpoint transport rules. The same transport table runs against objectstore.New (constructor
// re-enforcement) in core/internal/objectstore. config_test.go is unchanged.

func objectEnv(appEnv string, extra map[string]string) map[string]string {
	env := baseEnv()
	env["APP_ENV"] = appEnv
	for k, v := range extra {
		if v == "<unset>" {
			delete(env, k)
			continue
		}
		env[k] = v
	}
	return env
}

func TestObjectSettingsDefaults(t *testing.T) {
	c, err := load(objectEnv("test", nil))
	if err != nil {
		t.Fatal(err)
	}
	o := c.ObjectStore
	if o.SigningEndpoint != "" || o.MaxObjectBytes != 26214400 || o.SignedURLDefaultTTL != 60*time.Second ||
		o.SignedURLMaxTTL != 300*time.Second || o.TransferTimeout != 120*time.Second {
		t.Fatalf("defaults: %+v", o)
	}
	c, err = load(objectEnv("test", map[string]string{
		"OBJECT_STORE_SIGNING_ENDPOINT": "http://127.0.0.1:9100", "OBJECT_STORE_MAX_OBJECT_BYTES": "1073741824",
		"OBJECT_STORE_SIGNED_URL_DEFAULT_TTL": "300", "OBJECT_STORE_SIGNED_URL_MAX_TTL": "300", "OBJECT_STORE_TRANSFER_TIMEOUT": "900",
	}))
	if err != nil {
		t.Fatal(err)
	}
	o = c.ObjectStore
	if o.SigningEndpoint != "http://127.0.0.1:9100" || o.MaxObjectBytes != 1073741824 || o.SignedURLDefaultTTL != 300*time.Second ||
		o.SignedURLMaxTTL != 300*time.Second || o.TransferTimeout != 900*time.Second {
		t.Fatalf("upper bounds: %+v", o)
	}
	c, err = load(objectEnv("test", map[string]string{
		"OBJECT_STORE_MAX_OBJECT_BYTES": "1", "OBJECT_STORE_SIGNED_URL_DEFAULT_TTL": "1", "OBJECT_STORE_SIGNED_URL_MAX_TTL": "1",
		"OBJECT_STORE_TRANSFER_TIMEOUT": "1",
	}))
	if err != nil || c.ObjectStore.MaxObjectBytes != 1 || c.ObjectStore.SignedURLMaxTTL != time.Second || c.ObjectStore.TransferTimeout != time.Second {
		t.Fatalf("lower bounds: %+v %v", c.ObjectStore, err)
	}
}

func TestObjectSettingsBounds(t *testing.T) {
	for _, c := range []struct {
		name, value string
	}{
		{"OBJECT_STORE_MAX_OBJECT_BYTES", "0"}, {"OBJECT_STORE_MAX_OBJECT_BYTES", "1073741825"}, {"OBJECT_STORE_MAX_OBJECT_BYTES", "-1"},
		{"OBJECT_STORE_MAX_OBJECT_BYTES", "25MiB"}, {"OBJECT_STORE_MAX_OBJECT_BYTES", "+5"}, {"OBJECT_STORE_MAX_OBJECT_BYTES", "05"},
		{"OBJECT_STORE_SIGNED_URL_DEFAULT_TTL", "0"}, {"OBJECT_STORE_SIGNED_URL_DEFAULT_TTL", "301"}, {"OBJECT_STORE_SIGNED_URL_DEFAULT_TTL", "60s"},
		{"OBJECT_STORE_SIGNED_URL_MAX_TTL", "0"}, {"OBJECT_STORE_SIGNED_URL_MAX_TTL", "301"}, {"OBJECT_STORE_SIGNED_URL_MAX_TTL", "9223372036854775808"},
		{"OBJECT_STORE_TRANSFER_TIMEOUT", "0"}, {"OBJECT_STORE_TRANSFER_TIMEOUT", "901"}, {"OBJECT_STORE_TRANSFER_TIMEOUT", "2m"},
	} {
		secret := "Cx9objectcanary" + c.value
		env := objectEnv("test", map[string]string{c.name: c.value, "OBJECT_STORE_SECRET_KEY": secret})
		_, err := load(env)
		if err == nil || !strings.Contains(err.Error(), c.name) {
			t.Errorf("%s=%q: %v", c.name, c.value, err)
			continue
		}
		if strings.Contains(err.Error(), c.value) && c.value != "0" && c.value != "1" {
			t.Errorf("%s=%q: the problem echoes the value: %v", c.name, c.value, err)
		}
		if strings.Contains(err.Error(), secret) {
			t.Errorf("%s: the problem leaks a secret", c.name)
		}
	}
	_, err := load(objectEnv("test", map[string]string{"OBJECT_STORE_SIGNED_URL_DEFAULT_TTL": "120", "OBJECT_STORE_SIGNED_URL_MAX_TTL": "60"}))
	if err == nil || !strings.Contains(err.Error(), "OBJECT_STORE_SIGNED_URL_DEFAULT_TTL must not exceed OBJECT_STORE_SIGNED_URL_MAX_TTL") {
		t.Errorf("default above max: %v", err)
	}
}

// objectTransportCases is the endpoint transport table (Plan r1 §4.7, correction 2): production needs https for both
// endpoints; an http data endpoint with an https signing endpoint is refused everywhere; http for both is fine in
// development and test. want is "" for acceptance, otherwise a variable the problem must name.
var objectTransportCases = []struct {
	name, appEnv, data, signing, want string
}{
	{"production https data and https signing", "production", "https://objects.example.net", "https://signing.example.net", ""},
	{"production https data, default signing", "production", "https://objects.example.net", "", ""},
	{"production http data refused", "production", "http://minio:9000", "", "OBJECT_STORE_ENDPOINT must be https with APP_ENV=production"},
	{"production http signing refused", "production", "https://objects.example.net", "http://127.0.0.1:9100", "OBJECT_STORE_SIGNING_ENDPOINT must be https with APP_ENV=production"},
	{"production mixed refused", "production", "http://minio:9000", "https://signing.example.net", "OBJECT_STORE_SIGNING_ENDPOINT must not be https while OBJECT_STORE_ENDPOINT is http"},
	{"development mixed refused", "development", "http://minio:9000", "https://signing.example.net", "OBJECT_STORE_SIGNING_ENDPOINT must not be https while OBJECT_STORE_ENDPOINT is http"},
	{"test mixed refused", "test", "http://minio:9000", "https://signing.example.net", "OBJECT_STORE_SIGNING_ENDPOINT must not be https while OBJECT_STORE_ENDPOINT is http"},
	{"development http both", "development", "http://minio:9000", "http://127.0.0.1:9100", ""},
	{"test http both", "test", "http://minio.invalid:9000", "", ""},
	{"test https data, http signing", "test", "https://objects.example.net", "http://127.0.0.1:9100", ""},
	{"data endpoint with a path", "test", "http://minio:9000/bucket", "", "OBJECT_STORE_ENDPOINT must be an absolute http or https URL"},
	{"data endpoint with userinfo", "test", "http://user:pw@minio:9000", "", "OBJECT_STORE_ENDPOINT must be an absolute http or https URL"},
	{"data endpoint with a query", "test", "http://minio:9000?x=1", "", "OBJECT_STORE_ENDPOINT must be an absolute http or https URL"},
	{"data endpoint with a fragment", "test", "http://minio:9000#f", "", "OBJECT_STORE_ENDPOINT must be an absolute http or https URL"},
	{"data endpoint trailing slash", "test", "http://minio:9000/", "", "OBJECT_STORE_ENDPOINT must be an absolute http or https URL"},
	{"data endpoint other scheme", "test", "ftp://minio:9000", "", "OBJECT_STORE_ENDPOINT must be an absolute http or https URL"},
	{"data endpoint relative", "test", "minio:9000", "", "OBJECT_STORE_ENDPOINT must be an absolute http or https URL"},
	{"data endpoint bad port", "test", "http://minio:99999", "", "OBJECT_STORE_ENDPOINT must be an absolute http or https URL"},
	{"data endpoint empty port", "test", "http://minio:", "", "OBJECT_STORE_ENDPOINT must be an absolute http or https URL"},
	{"signing endpoint with a path", "test", "http://minio:9000", "http://127.0.0.1:9100/x", "OBJECT_STORE_SIGNING_ENDPOINT must be an absolute http or https URL"},
	{"signing endpoint relative", "test", "http://minio:9000", "127.0.0.1:9100", "OBJECT_STORE_SIGNING_ENDPOINT must be an absolute http or https URL"},
}

func TestObjectTransportRules(t *testing.T) {
	for _, c := range objectTransportCases {
		extra := map[string]string{"OBJECT_STORE_ENDPOINT": c.data}
		if c.signing != "" {
			extra["OBJECT_STORE_SIGNING_ENDPOINT"] = c.signing
		}
		_, err := load(objectEnv(c.appEnv, extra))
		switch {
		case c.want == "" && err != nil:
			t.Errorf("%s: refused: %v", c.name, err)
		case c.want != "" && (err == nil || !strings.Contains(err.Error(), c.want)):
			t.Errorf("%s: %v, want a problem containing %q", c.name, err, c.want)
		}
		if err != nil && (strings.Contains(err.Error(), "minio") || strings.Contains(err.Error(), "example.net") || strings.Contains(err.Error(), "127.0.0.1")) {
			t.Errorf("%s: the problem echoes an endpoint: %v", c.name, err)
		}
	}
}
