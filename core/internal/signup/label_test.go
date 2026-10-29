package signup

import (
	"bytes"
	"context"
	"crypto/rand"
	"errors"
	"io"
	"regexp"
	"strings"
	"testing"

	"core/internal/hostname"
	"core/internal/identity"
)

// labelSeq is a deterministic label reader: each Read hands out the next block (the last block repeats) and counts the
// draws (GenerateLabel reads exactly one LabelBytes block per draw).
type labelSeq struct {
	blocks [][]byte
	draws  int
}

func (r *labelSeq) Read(p []byte) (int, error) {
	b := r.blocks[min(r.draws, len(r.blocks)-1)]
	r.draws++
	return copy(p, b), nil
}

// failingReader is an entropy source that always fails.
type failingReader struct{}

func (failingReader) Read([]byte) (int, error) { return 0, errors.New("entropy unavailable") }

func block(fill byte) []byte { return bytes.Repeat([]byte{fill}, LabelBytes) }

// labelOf is the label GenerateLabel draws from one block.
func labelOf(t *testing.T, b []byte) string {
	t.Helper()
	l, err := GenerateLabel(bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}
	return l
}

// generatedShape is the shape of a generated platform label: <adjective>-<noun>-<four digits>.
var generatedShape = regexp.MustCompile(`^[a-z]+-[a-z]+-[0-9]{4}$`)

// legacyLabel is a label issued before the 2026-10-29 amendment: 26 lowercase base32 characters (128 bits).
const legacyLabel = "mfrggzdfmztwq2lknnwg23tpoa"

// TestGenerateLabel (T1): exactly LabelBytes = 6 bytes from the reader become <adjective>-<noun>-<dddd> with a masked
// word index and zero-padded digits; fixed vectors including leading zeros and the owner's example; a short or failing
// reader is an error, never a shorter label; 10k crypto/rand draws are well-formed, inside the slug grammar and the
// DNS-label bound, and (the space being 40,960,000) almost always distinct.
func TestGenerateLabel(t *testing.T) {
	if LabelBytes != 6 || labelWords != 64 || labelCombinations != 40_960_000 || maxLabelDraws != 4 {
		t.Fatalf("LabelBytes %d, labelWords %d, labelCombinations %d, maxLabelDraws %d", LabelBytes, labelWords, labelCombinations, maxLabelDraws)
	}
	for _, c := range []struct {
		in   []byte
		want string
	}{
		{[]byte("abcdefghijklmnop"), "mellow-lemon-3942"},
		{block(0x00), "amber-acorn-0000"},
		{block(0xff), "witty-willow-7295"},
		{block(0x11), "fluffy-daisy-1153"},
		{[]byte{0, 0, 0, 0, 0, 7}, "amber-acorn-0007"},
		{[]byte{22, 23, 0, 0, 0x0a, 0x53}, "gusty-flowerpot-2643"},
		{[]byte{0x40 | 22, 0x80 | 23, 0, 0, 0x0a, 0x53}, "gusty-flowerpot-2643"},
	} {
		if got := labelOf(t, c.in); got != c.want {
			t.Errorf("GenerateLabel(%x) = %q, want %q", c.in, got, c.want)
		}
	}
	// exactly one LabelBytes block is consumed per draw
	src := bytes.NewReader(append([]byte("abcdef"), block(0x00)...))
	if l, err := GenerateLabel(src); err != nil || l != "mellow-lemon-3942" || src.Len() != LabelBytes {
		t.Errorf("first draw %q %v, %d bytes left", l, err, src.Len())
	}
	for name, r := range map[string]io.Reader{
		"short":   io.LimitReader(rand.Reader, LabelBytes-1),
		"empty":   bytes.NewReader(nil),
		"failing": failingReader{},
	} {
		if l, err := GenerateLabel(r); err == nil || l != "" {
			t.Errorf("%s reader: %q %v, want an error and no label", name, l, err)
		}
	}
	seen := make(map[string]bool, 10000)
	dups := 0
	for i := 0; i < 10000; i++ {
		l, err := GenerateLabel(rand.Reader)
		if err != nil {
			t.Fatal(err)
		}
		if !generatedShape.MatchString(l) || len(l) > 63 || !hostname.SlugGrammar(l) {
			t.Fatalf("draw %d: %q is not an adjective-noun-four-digit label in the slug grammar", i, l)
		}
		if seen[l] {
			dups++
		}
		seen[l] = true
	}
	// the birthday expectation over 40.96M labels is about 1.2 repeats in 10k draws; 20 only if the draw is not uniform
	if dups > 20 {
		t.Errorf("%d repeated labels in 10000 draws", dups)
	}
}

// TestLabelWordLists: each list holds exactly labelWords distinct, non-empty, lowercase [a-z] words of 3-10 letters (so
// every label is DNS-safe and well inside 63 characters), the lists share no word, and no word is a reserved host label.
// Every word pair forms an issuable label under a plain configuration.
func TestLabelWordLists(t *testing.T) {
	word := regexp.MustCompile(`^[a-z]{3,10}$`)
	cfg := hostname.Config{BaseDomain: "localhost", Port: "543"}
	seen := map[string]string{}
	for name, list := range map[string][]string{"adjective": labelAdjectives[:], "noun": labelNouns[:]} {
		if len(list) != labelWords {
			t.Fatalf("%s list has %d words, want %d", name, len(list), labelWords)
		}
		for _, w := range list {
			if !word.MatchString(w) {
				t.Errorf("%s %q is not 3-10 lowercase letters", name, w)
			}
			if prev, dup := seen[w]; dup {
				t.Errorf("%s %q repeats a %s word", name, w, prev)
			}
			seen[w] = name
			if hostname.IsReservedLabel(w, cfg) {
				t.Errorf("%s %q is a reserved host label", name, w)
			}
		}
	}
	s := NewService(Config{Hostname: cfg})
	for _, a := range labelAdjectives {
		for _, n := range labelNouns {
			if l := a + "-" + n + "-0000"; !s.labelIssuable(l) {
				t.Fatalf("%q is not issuable", l)
			}
		}
	}
}

// TestLegacyLabelStillResolves: a label issued before the amendment (26 lowercase base32 characters) keeps its grammar
// and its platform-host classification, exactly as a generated friendly label does; the old shape is not what the
// generator makes now.
func TestLegacyLabelStillResolves(t *testing.T) {
	cfg := hostname.Config{BaseDomain: "localhost", Port: "543"}
	friendly := labelOf(t, []byte{22, 23, 0, 0, 0x0a, 0x53})
	for _, l := range []string{legacyLabel, friendly, "org-a", "alpha-agency"} {
		if !hostname.SlugGrammar(l) {
			t.Errorf("%q is outside the slug grammar", l)
		}
		c := hostname.Classify(hostname.Host{Name: l + ".localhost"}, cfg)
		if c.Kind != hostname.KindPlatform || c.Slug != l {
			t.Errorf("%q classified %+v, want the platform label itself", l, c)
		}
	}
	if generatedShape.MatchString(legacyLabel) || !generatedShape.MatchString(friendly) {
		t.Errorf("the generator shape must match the friendly label only (%q, %q)", legacyLabel, friendly)
	}
}

// stmtCall is one provisioning statement the double answered.
type stmtCall struct{ label, target string }

// labelService is a unit service with a deterministic label reader and a provisioning statement double: answer
// decides each call's row; the identity pool is nil, so any real statement would fail rather than pass.
func labelService(cfg hostname.Config, r io.Reader, answer func(n int, label, target string) provisionRow) (*Service, *[]stmtCall) {
	id := identity.NewService(identity.NewStore(nil), identity.Config{RootOrigin: "https://localhost:543"})
	s := NewService(Config{Identity: id, Hostname: cfg, RootOrigin: "https://localhost:543", LabelRand: r})
	calls := &[]stmtCall{}
	s.provisionStmt = func(label, target string) (provisionRow, error) {
		*calls = append(*calls, stmtCall{label, target})
		return answer(len(*calls), label, target), nil
	}
	return s, calls
}

func created(_ int, _, target string) provisionRow {
	return provisionRow{Outcome: OutcomeCreated, OrganizationID: "org-id", Slug: "acme", Generation: 1, TargetOrigin: target}
}

func labelTaken(int, string, string) provisionRow { return provisionRow{Outcome: outcomeLabelTaken} }

// TestLabelReservedUnderConfiguration (T1, Plan r5 §3.2 a-e): a deterministic reader yields a fixed generated label
// v first. (a) v in PLATFORM_RESERVED_LABELS, (b) v the first label of CUSTOM_DOMAIN_CNAME_TARGET and (c) of
// PLATFORM_AUTH_HOST under the platform domain: v is redrawn without any statement and the second draw is issued; (d)
// v on every draw: signup_unavailable after 4 draws with the store never called; (e) one reserved draw plus three
// label_taken outcomes exhaust the same budget. Also: a label_taken outcome is retried with a fresh label, an
// already_created answer keeps the stored origin whatever label the retry drew, and a failing reader never reaches SQL.
func TestLabelReservedUnderConfiguration(t *testing.T) {
	ctx := context.Background()
	var zero identity.SignupCredentialInfo
	bv, bw, bx, by := []byte("abcdefghijklmnop"), block(0x00), block(0x11), block(0x22)
	v, w, x, y := labelOf(t, bv), labelOf(t, bw), labelOf(t, bx), labelOf(t, by)
	if !generatedShape.MatchString(v) || v == w || w == x || x == y {
		t.Fatalf("vectors %q %q %q %q", v, w, x, y)
	}
	base := hostname.Config{BaseDomain: "localhost", Port: "543"}
	with := func(f func(*hostname.Config)) hostname.Config {
		c := base
		f(&c)
		return c
	}
	configs := map[string]hostname.Config{
		"(a) PLATFORM_RESERVED_LABELS": with(func(c *hostname.Config) { c.ReservedLabels = []string{v} }),
		"(b) CNAME target label":       with(func(c *hostname.Config) { c.CNAMETarget = v + ".localhost" }),
		"(c) auth host label":          with(func(c *hostname.Config) { c.AuthHost = v + ".localhost" }),
	}
	for name, cfg := range configs {
		t.Run(name, func(t *testing.T) {
			if !hostname.IsReservedLabel(v, cfg) || hostname.Classify(hostname.Host{Name: v + ".localhost"}, cfg).Kind != hostname.KindReserved {
				t.Fatalf("%q is not reserved under the configuration", v)
			}
			r := &labelSeq{blocks: [][]byte{bv, bw}}
			s, calls := labelService(cfg, r, created)
			p, err := s.Provision(ctx, zero, "Acme", "acme")
			if err != nil || p.Outcome != OutcomeCreated {
				t.Fatalf("provision: %+v %v", p, err)
			}
			want := "https://" + w + ".localhost:543"
			if len(*calls) != 1 || (*calls)[0] != (stmtCall{w, want}) || r.draws != 2 {
				t.Errorf("statements %+v after %d draws; want one for %q", *calls, r.draws, w)
			}
			if !strings.HasPrefix(p.Continue, want+"/signup/continue?ref=") || strings.Contains(p.Continue, v) {
				t.Errorf("continue %q, want the second draw's origin %s", p.Continue, want)
			}
		})
	}
	reserved := configs["(a) PLATFORM_RESERVED_LABELS"]

	t.Run("(d) every draw reserved: unavailable, store never called", func(t *testing.T) {
		r := &labelSeq{blocks: [][]byte{bv}}
		s, calls := labelService(reserved, r, created)
		_, err := s.Provision(ctx, zero, "Acme", "acme")
		if CodeOf(err) != CodeUnavailable || len(*calls) != 0 || r.draws != maxLabelDraws {
			t.Errorf("got %v with %d statements after %d draws; want signup_unavailable, 0, %d", err, len(*calls), r.draws, maxLabelDraws)
		}
		// without the double the nil-pool identity would answer unauthenticated: the store is never reached either way
		s.provisionStmt = nil
		r.draws = 0
		if _, err := s.Provision(ctx, zero, "Acme", "acme"); CodeOf(err) != CodeUnavailable || r.draws != maxLabelDraws {
			t.Errorf("without the double: %v after %d draws", err, r.draws)
		}
	})

	t.Run("(e) one reserved draw and three label_taken exhaust the shared budget", func(t *testing.T) {
		r := &labelSeq{blocks: [][]byte{bv, bw, bx, by}}
		s, calls := labelService(reserved, r, labelTaken)
		_, err := s.Provision(ctx, zero, "Acme", "acme")
		if CodeOf(err) != CodeUnavailable || r.draws != maxLabelDraws {
			t.Fatalf("got %v after %d draws; want signup_unavailable after %d", err, r.draws, maxLabelDraws)
		}
		if len(*calls) != 3 || (*calls)[0].label != w || (*calls)[1].label != x || (*calls)[2].label != y {
			t.Errorf("statements %+v; want label_taken for %q, %q, %q", *calls, w, x, y)
		}
	})

	t.Run("label_taken is retried with a fresh label", func(t *testing.T) {
		r := &labelSeq{blocks: [][]byte{bw, bx}}
		s, calls := labelService(base, r, func(n int, label, target string) provisionRow {
			if n == 1 {
				return labelTaken(n, label, target)
			}
			return created(n, label, target)
		})
		p, err := s.Provision(ctx, zero, "Acme", "acme")
		want := "https://" + x + ".localhost:543"
		if err != nil || p.Outcome != OutcomeCreated || len(*calls) != 2 || (*calls)[1] != (stmtCall{x, want}) ||
			!strings.HasPrefix(p.Continue, want+"/signup/continue?ref=") {
			t.Errorf("retry: %+v %v, statements %+v", p, err, *calls)
		}
	})

	t.Run("already_created keeps the stored origin", func(t *testing.T) {
		const stored = "https://mfrggzdfmztwq2lknnwg23tpoa.localhost:543"
		r := &labelSeq{blocks: [][]byte{bw}}
		s, _ := labelService(base, r, func(int, string, string) provisionRow {
			return provisionRow{Outcome: OutcomeAlreadyCreated, OrganizationID: "org-id", Slug: "acme", Generation: 2, TargetOrigin: stored}
		})
		p, err := s.Provision(ctx, zero, "Acme", "acme")
		if err != nil || p.Outcome != OutcomeAlreadyCreated || !strings.HasPrefix(p.Continue, stored+"/signup/continue?ref=") ||
			strings.Contains(p.Continue, w) {
			t.Errorf("already_created: %+v %v; want the stored origin, not the drawn %q", p, err, w)
		}
		// a row without a stored origin fails closed
		s, _ = labelService(base, &labelSeq{blocks: [][]byte{bw}}, func(int, string, string) provisionRow {
			return provisionRow{Outcome: OutcomeAlreadyCreated, OrganizationID: "org-id", Slug: "acme", Generation: 2}
		})
		if _, err := s.Provision(ctx, zero, "Acme", "acme"); CodeOf(err) != CodeUnavailable {
			t.Errorf("no stored origin: %v", err)
		}
	})

	t.Run("a failing reader never reaches SQL", func(t *testing.T) {
		s, calls := labelService(base, failingReader{}, created)
		if _, err := s.Provision(ctx, zero, "Acme", "acme"); CodeOf(err) != CodeUnavailable || len(*calls) != 0 {
			t.Errorf("failing reader: %v with %d statements", err, len(*calls))
		}
	})
}

// TestLabelEqualToSlugIsRedrawn (T1): a generated label equal to the submitted slug is another redraw condition inside the
// shared four-draw budget, before any statement (web refuses a tenant origin whose host equals the Organization's own
// slug, so committing it would strand a created tenant). A first equal draw is replaced by the second draw; four equal
// draws exhaust to signup_unavailable with the store never called, so nothing is retained.
func TestLabelEqualToSlugIsRedrawn(t *testing.T) {
	ctx := context.Background()
	var zero identity.SignupCredentialInfo
	base := hostname.Config{BaseDomain: "localhost", Port: "543"}
	bw, bx := block(0x00), block(0x11)
	w, x := labelOf(t, bw), labelOf(t, bx)
	if !ValidSlug(w, base) || w == x {
		t.Fatalf("vectors %q %q", w, x)
	}

	t.Run("a first equal draw is redrawn and the second is issued", func(t *testing.T) {
		r := &labelSeq{blocks: [][]byte{bw, bx}}
		s, calls := labelService(base, r, created)
		p, err := s.Provision(ctx, zero, "Acme", w)
		want := "https://" + x + ".localhost:543"
		if err != nil || p.Outcome != OutcomeCreated || len(*calls) != 1 || (*calls)[0] != (stmtCall{x, want}) || r.draws != 2 {
			t.Fatalf("provision: %+v %v, statements %+v after %d draws; want one for %q", p, err, *calls, r.draws, x)
		}
		if !strings.HasPrefix(p.Continue, want+"/signup/continue?ref=") || strings.Contains(p.Continue, "//"+w+".") {
			t.Errorf("continue %q, want the second draw's origin %s", p.Continue, want)
		}
	})

	t.Run("four equal draws exhaust with nothing retained", func(t *testing.T) {
		r := &labelSeq{blocks: [][]byte{bw}}
		s, calls := labelService(base, r, created)
		_, err := s.Provision(ctx, zero, "Acme", w)
		if CodeOf(err) != CodeUnavailable || len(*calls) != 0 || r.draws != maxLabelDraws {
			t.Errorf("got %v with %d statements after %d draws; want signup_unavailable, 0, %d", err, len(*calls), r.draws, maxLabelDraws)
		}
	})

	t.Run("the equality shares the budget with label_taken", func(t *testing.T) {
		r := &labelSeq{blocks: [][]byte{bw, bx, block(0x22), block(0x33)}}
		s, calls := labelService(base, r, labelTaken)
		_, err := s.Provision(ctx, zero, "Acme", w)
		if CodeOf(err) != CodeUnavailable || r.draws != maxLabelDraws || len(*calls) != 3 {
			t.Errorf("got %v after %d draws with %d statements; want signup_unavailable after %d with 3", err, r.draws, len(*calls), maxLabelDraws)
		}
	})
}
