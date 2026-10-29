// Command identitysetup prepares one run of the identity test environment (G004 Plan r5 §8.2, §8.3; test only,
// never production). It writes, into an empty directory (the overlay's project-owned volume):
//
//	ca.pem                 the run CA certificate (public; for a dedicated manual-test browser profile)
//	proxy/tls.crt          one leaf certificate for the test hosts, then the CA certificate
//	proxy/tls.key          the leaf key (0600; read by the proxy's root master process)
//	spki.txt               base64 SHA-256 of the leaf's SubjectPublicKeyInfo (Chromium's exact key pin)
//	emulator/signing.pem   the emulator's per-run RSA signing key
//	emulator/seed.yaml     the seed template rendered with the run's API key, passwords and canaries
//	core/api-key           the per-run emulator API key (0600, owned by core's user)
//	browser/users.json     the seeded users' emails, passwords, names and canaries, and the emulator API key, for the
//	                       browser tests' driver (user administration through the emulator's API)
//	mcp/connect-jwks.json  the public key set of the per-run Connect test signer (G014 C10): core's MCP test trust
//	                       anchor (MCP_CONNECT_JWKS_FILE, accepted only with APP_ENV=test)
//	mcp/connect-signing.pem  the Connect test signer's RSA key (0600), read only by the B-MCP-E2E driver
//	secrets.txt            every generated secret and canary, one per line, for the harness's secret registry
//
//	identitysetup --out DIR --template PATH --hosts HOST[,HOST...] [--core-owner UID:GID]
//
// The template may contain only {{API_KEY}}, {{PASSWORD:name}} and {{CANARY:name}} placeholders (name:
// [a-z][a-z0-9-]*). A directory that already holds a complete run (".complete") is left unchanged, so a
// restarted stack keeps its keys. Nothing generated is printed. Exit status: 0 done; 1 failure; 2 usage.
package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"flag"
	"fmt"
	"io"
	"math/big"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
)

const (
	exitOK    = 0
	exitFail  = 1
	exitUsage = 2

	certLifetime = 72 * time.Hour
	completeFile = ".complete"
)

var (
	placeholderRe = regexp.MustCompile(`\{\{([A-Z_]+)(?::([a-z][a-z0-9-]*))?\}\}`)
	hostRe        = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?(?:\.[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?)*$`)
	ownerRe       = regexp.MustCompile(`^([0-9]{1,10}):([0-9]{1,10})$`)
	userBlockRe   = regexp.MustCompile(`(?m)^  - email: (.+)$`)
)

func main() {
	os.Exit(run(os.Args[1:], os.Stderr))
}

// run parses arguments and performs one setup. Diagnostics name files and steps only, never values.
func run(args []string, stderr io.Writer) int {
	fs := flag.NewFlagSet("identitysetup", flag.ContinueOnError)
	fs.SetOutput(stderr)
	out := fs.String("out", "", "output directory (empty, or holding a complete run)")
	tmpl := fs.String("template", "", "emulator seed template")
	hosts := fs.String("hosts", "", "comma-separated test host names for the leaf certificate")
	owner := fs.String("core-owner", "", "UID:GID that owns core/api-key (default: unchanged)")
	if err := fs.Parse(args); err != nil || fs.NArg() != 0 || *out == "" || *tmpl == "" || *hosts == "" {
		fmt.Fprintln(stderr, "usage: identitysetup --out DIR --template PATH --hosts HOST[,HOST...] [--core-owner UID:GID]")
		return exitUsage
	}
	names := strings.Split(*hosts, ",")
	for _, h := range names {
		if !hostRe.MatchString(h) {
			fmt.Fprintln(stderr, "identitysetup: --hosts must list lower-case DNS names")
			return exitUsage
		}
	}
	uid, gid := -1, -1
	if *owner != "" {
		m := ownerRe.FindStringSubmatch(*owner)
		if m == nil {
			fmt.Fprintln(stderr, "identitysetup: --core-owner must be UID:GID")
			return exitUsage
		}
		uid, _ = strconv.Atoi(m[1])
		gid, _ = strconv.Atoi(m[2])
	}
	if _, err := os.Stat(filepath.Join(*out, completeFile)); err == nil {
		fmt.Fprintln(stderr, "identitysetup: complete run present; unchanged")
		return exitOK
	}
	template, err := os.ReadFile(*tmpl)
	if err != nil {
		fmt.Fprintln(stderr, "identitysetup: cannot read the template")
		return exitFail
	}
	if err := setup(*out, string(template), names, uid, gid); err != nil {
		fmt.Fprintf(stderr, "identitysetup: %v\n", err)
		return exitFail
	}
	return exitOK
}

// secretHex returns n random bytes as lower-case hex.
func secretHex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", errors.New("random source failed")
	}
	return hex.EncodeToString(b), nil
}

// generated holds one run's values; secrets lists every value the harness must register.
type generated struct {
	apiKey    string
	passwords map[string]string
	canaries  map[string]string
	secrets   []string
}

func (g *generated) value(kind, name string) (string, error) {
	var m map[string]string
	var prefix string
	var n int
	switch kind {
	case "API_KEY":
		if name != "" {
			return "", errors.New("template: API_KEY takes no name")
		}
		return g.apiKey, nil
	case "PASSWORD":
		m, prefix, n = g.passwords, "Pw-", 18
	case "CANARY":
		m, prefix, n = g.canaries, "Cx", 12
	default:
		return "", fmt.Errorf("template: unknown placeholder %s", kind)
	}
	if name == "" {
		return "", fmt.Errorf("template: %s needs a name", kind)
	}
	if v, ok := m[name]; ok {
		return v, nil
	}
	h, err := secretHex(n)
	if err != nil {
		return "", err
	}
	m[name] = prefix + h
	g.secrets = append(g.secrets, m[name])
	return m[name], nil
}

// render replaces every placeholder; anything left that looks like a placeholder is an error.
func render(template string, g *generated) (string, error) {
	var firstErr error
	outText := placeholderRe.ReplaceAllStringFunc(template, func(p string) string {
		m := placeholderRe.FindStringSubmatch(p)
		v, err := g.value(m[1], m[2])
		if err != nil && firstErr == nil {
			firstErr = err
		}
		return v
	})
	if firstErr != nil {
		return "", firstErr
	}
	if strings.Contains(outText, "{{") || strings.Contains(outText, "}}") {
		return "", errors.New("template: malformed placeholder")
	}
	return outText, nil
}

type seededUser struct {
	Name      string `json:"name"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type usersFile struct {
	Users          []seededUser      `json:"users"`
	Canaries       map[string]string `json:"canaries"`
	EmulatorAPIKey string            `json:"emulator_api_key"`
}

// seedUsers reads the rendered seed's users (a fixed two-space list layout of email, password, first_name and
// last_name keys) so the browser tests learn exactly what the emulator was seeded with.
func seedUsers(rendered string, g *generated) ([]seededUser, error) {
	var users []seededUser
	idx := userBlockRe.FindAllStringSubmatchIndex(rendered, -1)
	for i, loc := range idx {
		end := len(rendered)
		if i+1 < len(idx) {
			end = idx[i+1][0]
		}
		block := rendered[loc[0]:end]
		u := seededUser{Email: unquote(rendered[loc[2]:loc[3]])}
		for _, line := range strings.Split(block, "\n") {
			k, v, ok := strings.Cut(strings.TrimSpace(line), ": ")
			if !ok {
				continue
			}
			switch k {
			case "password":
				u.Password = unquote(v)
			case "first_name":
				u.FirstName = unquote(v)
			case "last_name":
				u.LastName = unquote(v)
			}
		}
		for name, pw := range g.passwords {
			if pw == u.Password {
				u.Name = name
			}
		}
		if u.Name == "" {
			return nil, errors.New("template: every user needs a {{PASSWORD:name}} password")
		}
		users = append(users, u)
	}
	if len(users) == 0 {
		return nil, errors.New("template: no users")
	}
	return users, nil
}

func unquote(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 && (s[0] == '"' && s[len(s)-1] == '"' || s[0] == '\'' && s[len(s)-1] == '\'') {
		return s[1 : len(s)-1]
	}
	return s
}

func pemBytes(kind string, der []byte) []byte {
	return pem.EncodeToMemory(&pem.Block{Type: kind, Bytes: der})
}

// issue creates the run CA and one leaf for hosts. It returns the CA PEM, the leaf chain PEM, the leaf key PEM
// and the leaf's SPKI pin.
func issue(hosts []string, now time.Time) (caPEM, chainPEM, keyPEM []byte, pin string, err error) {
	serial := func() (*big.Int, error) { return rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 127)) }
	caKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, nil, "", errors.New("CA key generation failed")
	}
	tag, err := secretHex(4)
	if err != nil {
		return nil, nil, nil, "", err
	}
	caSerial, err := serial()
	if err != nil {
		return nil, nil, nil, "", errors.New("serial generation failed")
	}
	ca := &x509.Certificate{
		SerialNumber:          caSerial,
		Subject:               pkix.Name{CommonName: "identity test run CA " + tag},
		NotBefore:             now.Add(-time.Hour),
		NotAfter:              now.Add(certLifetime),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLenZero:        true,
	}
	caDER, err := x509.CreateCertificate(rand.Reader, ca, ca, &caKey.PublicKey, caKey)
	if err != nil {
		return nil, nil, nil, "", errors.New("CA certificate creation failed")
	}
	caCert, err := x509.ParseCertificate(caDER)
	if err != nil {
		return nil, nil, nil, "", errors.New("CA certificate parse failed")
	}
	leafKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, nil, "", errors.New("leaf key generation failed")
	}
	leafSerial, err := serial()
	if err != nil {
		return nil, nil, nil, "", errors.New("serial generation failed")
	}
	leaf := &x509.Certificate{
		SerialNumber: leafSerial,
		Subject:      pkix.Name{CommonName: hosts[0]},
		DNSNames:     hosts,
		NotBefore:    now.Add(-time.Hour),
		NotAfter:     now.Add(certLifetime),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}
	leafDER, err := x509.CreateCertificate(rand.Reader, leaf, caCert, &leafKey.PublicKey, caKey)
	if err != nil {
		return nil, nil, nil, "", errors.New("leaf certificate creation failed")
	}
	leafKeyDER, err := x509.MarshalPKCS8PrivateKey(leafKey)
	if err != nil {
		return nil, nil, nil, "", errors.New("leaf key encoding failed")
	}
	spki, err := x509.MarshalPKIXPublicKey(&leafKey.PublicKey)
	if err != nil {
		return nil, nil, nil, "", errors.New("leaf public key encoding failed")
	}
	sum := sha256.Sum256(spki)
	caPEM = pemBytes("CERTIFICATE", caDER)
	chainPEM = append(pemBytes("CERTIFICATE", leafDER), caPEM...)
	return caPEM, chainPEM, pemBytes("PRIVATE KEY", leafKeyDER), base64.StdEncoding.EncodeToString(sum[:]), nil
}

// ConnectKeyID is the key ID of the per-run Connect test signer in mcp/connect-jwks.json.
const ConnectKeyID = "mcp-connect-test"

// connectSigner creates the per-run Connect test signer (G014 C10): an RSA key for RS256 and its public key set.
func connectSigner() (jwks, keyPEM []byte, err error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, errors.New("Connect signer key generation failed")
	}
	der, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return nil, nil, errors.New("Connect signer key encoding failed")
	}
	b64 := base64.RawURLEncoding.EncodeToString
	set := map[string][]map[string]string{"keys": {{
		"kty": "RSA", "use": "sig", "alg": "RS256", "kid": ConnectKeyID,
		"n": b64(key.N.Bytes()), "e": b64(big.NewInt(int64(key.E)).Bytes()),
	}}}
	jwks, err = json.Marshal(set)
	if err != nil {
		return nil, nil, errors.New("Connect key set encoding failed")
	}
	return append(jwks, '\n'), pemBytes("PRIVATE KEY", der), nil
}

type file struct {
	path string
	data []byte
	mode os.FileMode
}

func setup(out, template string, hosts []string, uid, gid int) error {
	if entries, err := os.ReadDir(out); err == nil && len(entries) > 0 {
		return errors.New("output directory is not empty and holds no complete run")
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return errors.New("cannot read the output directory")
	}
	apiHex, err := secretHex(32)
	if err != nil {
		return err
	}
	g := &generated{apiKey: "sk_test_" + apiHex, passwords: map[string]string{}, canaries: map[string]string{}}
	g.secrets = append(g.secrets, g.apiKey)
	seed, err := render(template, g)
	if err != nil {
		return err
	}
	if !strings.Contains(seed, g.apiKey) {
		return errors.New("template: no {{API_KEY}}")
	}
	users, err := seedUsers(seed, g)
	if err != nil {
		return err
	}
	usersJSON, err := json.MarshalIndent(usersFile{Users: users, Canaries: g.canaries, EmulatorAPIKey: g.apiKey}, "", "  ")
	if err != nil {
		return errors.New("users encoding failed")
	}
	caPEM, chainPEM, keyPEM, pin, err := issue(hosts, time.Now())
	if err != nil {
		return err
	}
	signing, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return errors.New("signing key generation failed")
	}
	signingDER, err := x509.MarshalPKCS8PrivateKey(signing)
	if err != nil {
		return errors.New("signing key encoding failed")
	}
	jwks, connectPEM, err := connectSigner()
	if err != nil {
		return err
	}
	secrets := slices.Clone(g.secrets)
	slices.Sort(secrets)
	files := []file{
		{"ca.pem", caPEM, 0o644},
		{"proxy/tls.crt", chainPEM, 0o644},
		{"proxy/tls.key", keyPEM, 0o600},
		{"spki.txt", []byte(pin + "\n"), 0o644},
		{"emulator/signing.pem", pemBytes("PRIVATE KEY", signingDER), 0o644},
		{"emulator/seed.yaml", []byte(seed), 0o644},
		{"core/api-key", []byte(g.apiKey + "\n"), 0o600},
		{"browser/users.json", append(usersJSON, '\n'), 0o644},
		{"mcp/connect-jwks.json", jwks, 0o644},
		{"mcp/connect-signing.pem", connectPEM, 0o600},
		{"secrets.txt", []byte(strings.Join(secrets, "\n") + "\n"), 0o600},
	}
	for _, f := range files {
		p := filepath.Join(out, f.path)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			return fmt.Errorf("cannot create the directory for %s", f.path)
		}
		if err := os.WriteFile(p, f.data, f.mode); err != nil {
			return fmt.Errorf("cannot write %s", f.path)
		}
		if err := os.Chmod(p, f.mode); err != nil {
			return fmt.Errorf("cannot set the mode of %s", f.path)
		}
	}
	if uid >= 0 {
		for _, p := range []string{"core", "core/api-key"} {
			if err := os.Chown(filepath.Join(out, p), uid, gid); err != nil {
				return fmt.Errorf("cannot set the owner of %s", p)
			}
		}
	}
	if err := os.WriteFile(filepath.Join(out, completeFile), nil, 0o644); err != nil {
		return errors.New("cannot mark the run complete")
	}
	return nil
}
