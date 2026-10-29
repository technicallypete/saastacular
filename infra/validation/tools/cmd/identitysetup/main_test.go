package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

const testTemplate = `# test seed
apiKeys:
  "{{API_KEY}}":
    environment: test
users:
  - email: alice@identity.test
    password: "{{PASSWORD:alice}}"
    first_name: Alice
    last_name: Example
    email_verified: true
  - email: canary-{{CANARY:email}}@identity.test
    password: "{{PASSWORD:hostile}}"
    first_name: "<img src=x onerror=alert(1)>"
    last_name: "{{CANARY:display}}"
    email_verified: true
`

var hosts = "org-a.localhost,org-b.localhost,attacker.localhost"

func writeTemplate(t *testing.T, body string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "seed.yaml")
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func runSetup(t *testing.T, out, tmpl string, extra ...string) (int, string) {
	t.Helper()
	var stderr bytes.Buffer
	args := append([]string{"--out", out, "--template", tmpl, "--hosts", hosts}, extra...)
	return run(args, &stderr), stderr.String()
}

func read(t *testing.T, dir, name string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(dir, name))
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func mode(t *testing.T, dir, name string) os.FileMode {
	t.Helper()
	st, err := os.Stat(filepath.Join(dir, name))
	if err != nil {
		t.Fatal(err)
	}
	return st.Mode().Perm()
}

func TestUsage(t *testing.T) {
	tmpl := writeTemplate(t, testTemplate)
	for _, args := range [][]string{
		{},
		{"--out", t.TempDir(), "--template", tmpl},
		{"--out", t.TempDir(), "--hosts", hosts},
		{"--out", t.TempDir(), "--template", tmpl, "--hosts", "Org-A.localhost"},
		{"--out", t.TempDir(), "--template", tmpl, "--hosts", "org-a.localhost,"},
		{"--out", t.TempDir(), "--template", tmpl, "--hosts", hosts, "--core-owner", "nonroot"},
		{"--out", t.TempDir(), "--template", tmpl, "--hosts", hosts, "extra"},
	} {
		var stderr bytes.Buffer
		if got := run(args, &stderr); got != exitUsage {
			t.Errorf("%v: exit %d, want %d", args, got, exitUsage)
		}
	}
}

func TestSetupWritesOneRun(t *testing.T) {
	out := filepath.Join(t.TempDir(), "identity")
	code, stderr := runSetup(t, out, writeTemplate(t, testTemplate))
	if code != exitOK {
		t.Fatalf("exit %d: %s", code, stderr)
	}
	want := map[string]os.FileMode{
		"ca.pem": 0o644, "proxy/tls.crt": 0o644, "proxy/tls.key": 0o600, "spki.txt": 0o644,
		"emulator/signing.pem": 0o644, "emulator/seed.yaml": 0o644, "core/api-key": 0o600,
		"browser/users.json": 0o644, "secrets.txt": 0o600,
		"mcp/connect-jwks.json": 0o644, "mcp/connect-signing.pem": 0o600,
	}
	for name, m := range want {
		if got := mode(t, out, name); got != m {
			t.Errorf("%s: mode %o, want %o", name, got, m)
		}
	}

	// The leaf chains to the run CA for every host, is a server certificate, and the pin is its key's SPKI.
	caBlock, _ := pem.Decode(read(t, out, "ca.pem"))
	ca, err := x509.ParseCertificate(caBlock.Bytes)
	if err != nil || !ca.IsCA {
		t.Fatalf("CA: %v", err)
	}
	leafBlock, rest := pem.Decode(read(t, out, "proxy/tls.crt"))
	leaf, err := x509.ParseCertificate(leafBlock.Bytes)
	if err != nil {
		t.Fatal(err)
	}
	if chainCA, _ := pem.Decode(rest); chainCA == nil || !bytes.Equal(chainCA.Bytes, caBlock.Bytes) {
		t.Error("the chain file must end with the CA certificate")
	}
	roots := x509.NewCertPool()
	roots.AddCert(ca)
	for _, h := range strings.Split(hosts, ",") {
		if _, err := leaf.Verify(x509.VerifyOptions{DNSName: h, Roots: roots, KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}}); err != nil {
			t.Errorf("leaf for %s: %v", h, err)
		}
	}
	if _, err := leaf.Verify(x509.VerifyOptions{DNSName: "evil.localhost", Roots: roots}); err == nil {
		t.Error("leaf verified for an unlisted host")
	}
	keyBlock, _ := pem.Decode(read(t, out, "proxy/tls.key"))
	key, err := x509.ParsePKCS8PrivateKey(keyBlock.Bytes)
	if err != nil {
		t.Fatal(err)
	}
	if !key.(*ecdsa.PrivateKey).PublicKey.Equal(leaf.PublicKey) {
		t.Error("tls.key does not match the leaf")
	}
	sum := sha256.Sum256(leaf.RawSubjectPublicKeyInfo)
	if got := strings.TrimSpace(string(read(t, out, "spki.txt"))); got != base64.StdEncoding.EncodeToString(sum[:]) {
		t.Error("spki.txt is not the leaf's SPKI pin")
	}

	signBlock, _ := pem.Decode(read(t, out, "emulator/signing.pem"))
	sk, err := x509.ParsePKCS8PrivateKey(signBlock.Bytes)
	if err != nil || sk.(*rsa.PrivateKey).N.BitLen() != 2048 {
		t.Fatalf("signing key: %v", err)
	}

	// G014 C10: the Connect test signer's key set holds exactly its public key, distinct from the emulator's key.
	connectBlock, _ := pem.Decode(read(t, out, "mcp/connect-signing.pem"))
	ck, err := x509.ParsePKCS8PrivateKey(connectBlock.Bytes)
	if err != nil || ck.(*rsa.PrivateKey).N.BitLen() != 2048 || ck.(*rsa.PrivateKey).Equal(sk) {
		t.Fatalf("Connect signer key: %v", err)
	}
	var set struct {
		Keys []map[string]string `json:"keys"`
	}
	if err := json.Unmarshal(read(t, out, "mcp/connect-jwks.json"), &set); err != nil || len(set.Keys) != 1 {
		t.Fatalf("Connect key set: %v", err)
	}
	k := set.Keys[0]
	pub := ck.(*rsa.PrivateKey).PublicKey
	if k["kty"] != "RSA" || k["alg"] != "RS256" || k["use"] != "sig" || k["kid"] != ConnectKeyID || len(k) != 6 ||
		k["n"] != base64.RawURLEncoding.EncodeToString(pub.N.Bytes()) || k["e"] != "AQAB" {
		t.Errorf("Connect key set entry: %v", k)
	}

	apiKey := strings.TrimSpace(string(read(t, out, "core/api-key")))
	if !regexp.MustCompile(`^sk_test_[0-9a-f]{64}$`).MatchString(apiKey) {
		t.Error("api-key shape")
	}
	seed := string(read(t, out, "emulator/seed.yaml"))
	if !strings.Contains(seed, `"`+apiKey+`":`) || strings.Contains(seed, "{{") {
		t.Error("seed not rendered with the run's key")
	}

	var users usersFile
	if err := json.Unmarshal(read(t, out, "browser/users.json"), &users); err != nil {
		t.Fatal(err)
	}
	if len(users.Users) != 2 || users.Users[0].Name != "alice" || users.Users[1].Name != "hostile" {
		t.Fatalf("users: %+v", users.Users)
	}
	for _, u := range users.Users {
		if !strings.HasPrefix(u.Password, "Pw-") || !strings.Contains(seed, `password: "`+u.Password+`"`) || !strings.Contains(seed, "email: "+u.Email) {
			t.Errorf("user %s does not match the seed", u.Name)
		}
	}
	if users.Users[1].FirstName != "<img src=x onerror=alert(1)>" || users.Users[1].LastName != users.Canaries["display"] {
		t.Errorf("hostile user names: %q %q", users.Users[1].FirstName, users.Users[1].LastName)
	}
	if users.EmulatorAPIKey != apiKey {
		t.Error("users.json lacks the run's emulator API key")
	}
	if users.Users[1].Email != "canary-"+users.Canaries["email"]+"@identity.test" {
		t.Error("email canary not rendered into the email")
	}

	secrets := strings.Fields(string(read(t, out, "secrets.txt")))
	for _, v := range []string{apiKey, users.Users[0].Password, users.Users[1].Password, users.Canaries["display"], users.Canaries["email"]} {
		if !strings.Contains("\n"+strings.Join(secrets, "\n")+"\n", "\n"+v+"\n") {
			t.Errorf("secrets.txt lacks a generated value")
		}
	}
	if len(secrets) != 5 {
		t.Errorf("secrets.txt has %d values, want 5", len(secrets))
	}
	for _, v := range secrets {
		if strings.Contains(stderr, v) {
			t.Error("a generated value reached stderr")
		}
	}
}

func TestSetupIsIdempotentAndFresh(t *testing.T) {
	tmpl := writeTemplate(t, testTemplate)
	out := filepath.Join(t.TempDir(), "identity")
	if code, msg := runSetup(t, out, tmpl); code != exitOK {
		t.Fatal(msg)
	}
	first := read(t, out, "secrets.txt")
	ca := read(t, out, "ca.pem")
	if code, msg := runSetup(t, out, tmpl); code != exitOK || !strings.Contains(msg, "unchanged") {
		t.Fatalf("second run: %d %s", code, msg)
	}
	if !bytes.Equal(first, read(t, out, "secrets.txt")) || !bytes.Equal(ca, read(t, out, "ca.pem")) {
		t.Error("a complete run was changed")
	}
	other := filepath.Join(t.TempDir(), "identity")
	if code, msg := runSetup(t, other, tmpl); code != exitOK {
		t.Fatal(msg)
	}
	if bytes.Equal(first, read(t, other, "secrets.txt")) || bytes.Equal(ca, read(t, other, "ca.pem")) {
		t.Error("two runs share generated values")
	}
}

func TestSetupRefusals(t *testing.T) {
	// A non-empty directory without a complete run is never overwritten.
	dirty := t.TempDir()
	if err := os.WriteFile(filepath.Join(dirty, "keep"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if code, _ := runSetup(t, dirty, writeTemplate(t, testTemplate)); code != exitFail {
		t.Errorf("dirty directory: exit %d", code)
	}
	if entries, _ := os.ReadDir(dirty); len(entries) != 1 {
		t.Error("files written into a dirty directory")
	}
	for name, body := range map[string]string{
		"unknown placeholder": testTemplate + "x: {{TOKEN:x}}\n",
		"nameless password":   strings.Replace(testTemplate, "{{PASSWORD:alice}}", "{{PASSWORD}}", 1),
		"named api key":       strings.Replace(testTemplate, "{{API_KEY}}", "{{API_KEY:x}}", 1),
		"malformed":           testTemplate + "x: {{canary}}\n",
		"no api key":          strings.Replace(testTemplate, `"{{API_KEY}}"`, `"fixed"`, 1),
		"user without name":   strings.Replace(testTemplate, `"{{PASSWORD:alice}}"`, `"fixed"`, 1),
		"no users":            "apiKeys:\n  \"{{API_KEY}}\":\n    environment: test\n",
	} {
		out := filepath.Join(t.TempDir(), "identity")
		if code, _ := runSetup(t, out, writeTemplate(t, body)); code != exitFail {
			t.Errorf("%s: exit %d", name, code)
		}
		if _, err := os.Stat(out); err == nil {
			t.Errorf("%s: output written", name)
		}
	}
	if code, _ := runSetup(t, filepath.Join(t.TempDir(), "x"), filepath.Join(t.TempDir(), "absent.yaml")); code != exitFail {
		t.Error("missing template accepted")
	}
}

func TestCoreOwner(t *testing.T) {
	out := filepath.Join(t.TempDir(), "identity")
	uid, gid := os.Getuid(), os.Getgid()
	code, msg := runSetup(t, out, writeTemplate(t, testTemplate), "--core-owner", strconv.Itoa(uid)+":"+strconv.Itoa(gid))
	if code != exitOK {
		t.Fatal(msg)
	}
	if mode(t, out, "core/api-key") != 0o600 {
		t.Error("api-key mode")
	}
}
