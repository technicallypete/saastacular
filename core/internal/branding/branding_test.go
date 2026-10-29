package branding

import (
	"fmt"
	"math"
	"strings"
	"testing"
)

func field(t *testing.T, err error) string {
	t.Helper()
	e, ok := AsError(err)
	if !ok {
		t.Fatalf("not a branding error: %v", err)
	}
	return e.Code + "/" + e.Field
}

// U-BRD-1: the label rule.
func TestLabel(t *testing.T) {
	s := func(v string) *string { return &v }
	for _, c := range []struct {
		in   *string
		want *string
		bad  bool
	}{
		{nil, nil, false}, {s("   "), nil, false}, {s("  Acme Tools "), s("Acme Tools"), false},
		{s(strings.Repeat("é", 80)), s(strings.Repeat("é", 80)), false},
		{s(strings.Repeat("é", 81)), nil, true}, {s("a\x00b"), nil, true}, {s("a\u0085b"), nil, true},
		{s("a\x7fb"), nil, true}, {s("\xff"), nil, true}, {s("<script>"), s("<script>"), false},
	} {
		got, err := ValidateLabel(c.in)
		if c.bad {
			if field(t, err) != "invalid_input/display_label" {
				t.Errorf("%q: %v", *c.in, err)
			}
			continue
		}
		if err != nil || (got == nil) != (c.want == nil) || (got != nil && *got != *c.want) {
			t.Errorf("label %v -> %v, %v", c.in, got, err)
		}
	}
	// PrepareLabel (before normalization) trims and refuses control characters but does not bound the length: 80
	// decomposed "e + combining acute" pairs are 160 code points before NFC and 80 after (BRD-3 proves the composition
	// through the database's normalize).
	decomposed := " " + strings.Repeat("e\u0301", 80) + "\t"
	if got, err := PrepareLabel(&decomposed); err != nil || got == nil || *got != strings.Repeat("e\u0301", 80) {
		t.Errorf("PrepareLabel(decomposed) = %v, %v", got, err)
	}
	if _, err := PrepareLabel(s("a\x1bb")); field(t, err) != "invalid_input/display_label" {
		t.Errorf("PrepareLabel(control) = %v", err)
	}
}

// U-BRD-2: the override grammar and injection negatives.
func TestOverrideGrammar(t *testing.T) {
	got, err := ValidateOverrides("light", map[string]string{"primary": "#0062D6"})
	if err != nil || got["primary"] != "#0062d6" {
		t.Fatalf("uppercase is normalized: %v %v", got, err)
	}
	for _, in := range []map[string]string{
		{"overlay": "#000000"}, {"ink": "#000000"}, {"primary": "#0062d"}, {"primary": "0062d6"}, {"primary": "#0062d6;"},
		{"primary": "red"}, {"primary": "#0062d6 !important"}, {"primary": "url(x)"}, {"primary": "#0062d6ff"},
		{"--x": "#000000"}, {"Primary": "#000000"},
	} {
		if _, err := ValidateOverrides("dark", in); err == nil || field(t, err) != "invalid_input/dark" {
			t.Errorf("%v accepted: %v", in, err)
		}
	}
	if len(Tokens) != 15 {
		t.Fatalf("%d tokens, want 15", len(Tokens))
	}
}

// U-BRD-3: the platform defaults (the adopted kit's exact values) pass CheckContrast in both appearances because every
// entry's requirement is its kit-baseline floor (the unit precondition); the default ratios of the r3 states are pinned
// (G069 owner kit-fidelity direction, 2026-10-26). An entry pinned below its threshold is OWNER-WAIVED: the kit misses
// it, the floor admits exactly the kit's own ratio, and the outcome is a recorded waiver, not a threshold pass.
func TestDefaultsPass(t *testing.T) {
	for _, a := range []Appearance{Light, Dark} {
		if err := CheckContrast(a, nil); err != nil {
			t.Errorf("%s defaults: %v", a, err)
		}
	}
	for _, c := range []struct {
		name     string
		a        Appearance
		p        pair
		min, max float64
	}{
		{"T4 muted-foreground on muted, about 4.4014 (OWNER-WAIVED, below 4.5; the first pair the kit misses)", Light,
			text("T4", solid("muted-foreground"), solid("muted")), 4.35, 4.45},
		{"T8 white on primary at rest, about 4.2097 (OWNER-WAIVED, below 4.5)", Light,
			text("T8", solid("primary-foreground"), solid("primary")), 4.16, 4.26},
		{"T8 white on the 80 % primary hover over background, about 3.1768 (OWNER-WAIVED, below 4.5)", Light,
			text("T8", solid("primary-foreground"), over("primary", 0.8, "background")), 3.13, 3.23},
		{"T8 white on primary at rest, about 4.2097 (OWNER-WAIVED, below 4.5)", Dark,
			text("T8", solid("primary-foreground"), solid("primary")), 4.16, 4.26},
		{"T9 secondary-foreground on the 5 % OKLCH foreground hover mix, about 13.8236", Light,
			text("T9", solid("secondary-foreground"), mixed("secondary", "foreground", 0.05)), 13.70, 13.95},
		{"T10 destructive text on its 20 % hover tint, about 3.3168 (OWNER-WAIVED, below 4.5)", Light,
			text("T10", solid("destructive"), over("destructive", 0.2, "background")), 3.27, 3.37},
		{"T10 destructive text on its 30 % hover tint over popover, about 3.7297 (OWNER-WAIVED, below 4.5)", Dark,
			text("T10", solid("destructive"), over("destructive", 0.3, "popover")), 3.68, 3.78},
		{"T12 foreground on the 50 % input fill, about 17.6298", Light,
			text("T12", solid("foreground"), over("input", 0.5, "background")), 17.50, 17.75},
		{"T12 foreground on the 50 % input fill over popover, about 14.8800", Dark,
			text("T12", solid("foreground"), over("input", 0.5, "popover")), 14.75, 15.00},
		{"N3 input on background, about 1.2622 (OWNER-WAIVED, below 3)", Light,
			nonText("N3", solid("input"), solid("background")), 1.22, 1.30},
		{"N3 input on popover, about 1.3158 (OWNER-WAIVED, below 3)", Dark,
			nonText("N3", solid("input"), solid("popover")), 1.27, 1.36},
	} {
		pinned := false
		for _, q := range pairs {
			pinned = pinned || (q.id == c.p.id && q.fg == c.p.fg && q.bg == c.p.bg && q.applies(c.a))
		}
		if !pinned {
			t.Errorf("%s %s is not an evaluated pair", c.a, c.name)
		}
		if r := c.p.ratio(Effective(c.a, nil)); r < c.min || r > c.max {
			t.Errorf("%s %s = %.4f, want [%.2f, %.2f]", c.a, c.name, r, c.min, c.max)
		}
	}
	// The OWNER-WAIVED inventory: exactly these entries sit below their threshold on the kit defaults, each with the
	// kit's own ratio as its floor; every other entry keeps its full threshold. A new shortfall (a kit or pair change)
	// fails here instead of being admitted silently.
	waived := map[Appearance][]string{
		Light: {
			"T4 muted-foreground on muted",
			"T8 primary-foreground on primary",
			"T8 primary-foreground on primary/80% over background",
			"T8 primary-foreground on primary/80% over popover",
			"T10 destructive on destructive/10% over background",
			"T10 destructive on destructive/20% over background",
			"T10 destructive on destructive/10% over popover",
			"T10 destructive on destructive/20% over popover",
			"N3 input on background",
			"N3 input on popover",
		},
		Dark: {
			"T8 primary-foreground on primary",
			"T10 destructive on destructive/30% over background",
			"T10 destructive on destructive/30% over popover",
			"N3 input on background",
			"N3 input on popover",
		},
	}
	for _, a := range []Appearance{Light, Dark} {
		base := Effective(a, nil)
		var got []string
		for i, p := range pairs {
			if !p.applies(a) {
				continue
			}
			r := p.ratio(base)
			if floors[a][i] != math.Min(p.threshold, r) {
				t.Errorf("%s %s %s floor %.4f, want min(%.1f, kit %.4f)", a, p.id, describe(p), floors[a][i], p.threshold, r)
			}
			if r < p.threshold {
				got = append(got, p.id+" "+describe(p))
			} else if floors[a][i] != p.threshold {
				t.Errorf("%s %s %s: the kit meets it, floor %.4f != threshold", a, p.id, describe(p), floors[a][i])
			}
		}
		if strings.Join(got, "|") != strings.Join(waived[a], "|") {
			t.Errorf("%s OWNER-WAIVED entries = %q, want %q", a, got, waived[a])
		}
	}
	// S24: the CtaButton's fixed white label on the platform primary (--vx-blue #0075ff, both appearances) at rest,
	// hover:brightness-110 and active:brightness-90 (the filter scales the whole element, label included). It is a
	// platform contract, not a tenant pair, because the CTA renders only under the platform palette. OWNER-WAIVED: the
	// kit's white on #0075ff is about 4.2097 at rest, 3.7713 at x1.10 and 4.0165 at x0.90, all below 4.5; pinned so a
	// change is visible, never reported as a 4.5:1 pass.
	white := rgb("#ffffff")
	for _, a := range []Appearance{Light, Dark} {
		primary := rgb(Effective(a, nil)["primary"])
		for _, c := range []struct{ f, min, max float64 }{{1, 4.16, 4.26}, {1.10, 3.72, 3.82}, {0.90, 3.97, 4.07}} {
			if r := Contrast(scale(white, c.f), scale(primary, c.f)); r < c.min || r > c.max {
				t.Errorf("%s CTA white on primary x%.2f = %.4f, want [%.2f, %.2f] (OWNER-WAIVED)", a, c.f, r, c.min, c.max)
			}
		}
	}
}

// U-BRD-4: the effective palette and the r3 rejection edges (alpha states instead of brightness factors).
func TestContrastEdges(t *testing.T) {
	// The foreground override no longer fans out to text roles: the effective palette is exactly the 15 tokens, and a
	// Dark override leaves Light untouched.
	if eff := Effective(Dark, Overrides{"foreground": "#dddddd"}); len(eff) != len(Tokens) || eff["foreground"] != "#dddddd" {
		t.Errorf("dark foreground override: %v", eff)
	}
	for _, a := range []Appearance{Light, Dark} {
		eff := Effective(a, nil)
		for _, k := range Tokens {
			if !colorGrammar.MatchString(eff[k]) {
				t.Errorf("%s platform %s = %q", a, k, eff[k])
			}
		}
		if len(eff) != len(Tokens) {
			t.Errorf("%s platform has %d entries, want %d", a, len(eff), len(Tokens))
		}
	}
	if eff := Effective(Light, nil); eff["foreground"] != "#0c0a09" {
		t.Errorf("light foreground: %v", eff["foreground"])
	}
	// Refusals under the kit-baseline floor. The kit meets T1, T4 (Dark), T5, T9, T12, N2 and the Dark T10 20 % tints,
	// and the T5, T9, T12 and Dark T10 rows fail at about 4.34-4.47, above every waived text floor (the lowest is 3.1768),
	// so they prove the full threshold still applies where the kit meets it (no global relaxed threshold). The
	// below-floor rows fail an OWNER-WAIVED entry by falling under the kit's own ratio, in both appearances.
	for _, c := range []struct {
		name string
		a    Appearance
		o    Overrides
		want string
	}{
		{"light background only", Light, Overrides{"background": "#101010"}, "contrast_failed/T1:light"},
		{"dark foreground only, too dim", Dark, Overrides{"foreground": "#333333"}, "contrast_failed/T1:dark"},
		{"below floor: grey primary keeps 4.5:1 at rest, its 80 % hover (about 3.1356) is under the kit's 3.1768", Light,
			Overrides{"primary": "#767676"}, "contrast_failed/T8:light"},
		{"below floor: lighter primary at rest (about 3.3147) is under the kit's 4.2097", Light,
			Overrides{"primary": "#3d8bff"}, "contrast_failed/T8:light"},
		{"below floor: lighter primary at rest (about 3.3147) is under the kit's 4.2097", Dark,
			Overrides{"primary": "#3d8bff"}, "contrast_failed/T8:dark"},
		{"below floor: muted-foreground on muted (about 4.3380) is under the kit's 4.4014, on background 4.5:1 holds",
			Light, Overrides{"muted-foreground": "#7d6e68"}, "contrast_failed/T4:light"},
		{"below floor: lighter destructive on its 10 % tint (about 3.3610) is under the kit's 3.9917", Light,
			Overrides{"destructive": "#ff0000"}, "contrast_failed/T10:light"},
		{"below floor: the 30 % tint over popover (about 3.7069) is under the kit's 3.7297, the 20 % tints keep 4.5:1",
			Dark, Overrides{"destructive": "#be907c"}, "contrast_failed/T10:dark"},
		{"kit meets: destructive on its 20 % tint (about 4.3432) still needs 4.5:1", Dark,
			Overrides{"destructive": "#c07070"}, "contrast_failed/T10:dark"},
		{"ring equal to muted", Light, Overrides{"ring": "#f3f1f1"}, "contrast_failed/N2:light"},
		{"accent passes on background, fails on muted", Light, Overrides{"accent": "#6f6f6f", "background": "#ffffff"}, "contrast_failed/T5:light"},
		{"muted-foreground within 3:1 of popover", Dark, Overrides{"muted-foreground": "#3a3a3a"}, "contrast_failed/T4:dark"},
		{"accent inherits the synced Dark muted", Dark, Overrides{"accent": "#8a8a8a"}, "contrast_failed/T5:dark"},
		{"mid-grey foreground on a mid-grey input fill", Light,
			Overrides{"foreground": "#404040", "input": "#555555"}, "contrast_failed/T12:light"},
		{"secondary passes at rest, its OKLCH 5 % foreground hover mix loses the label", Light,
			Overrides{"secondary": "#7a7a7a", "secondary-foreground": "#000000"}, "contrast_failed/T9:light"},
		{"saturated secondary hover mix passes clipped to sRGB, fails unclipped (wide gamut)", Dark,
			Overrides{"foreground": "#55ffff", "secondary": "#aa22ff", "secondary-foreground": "#ffffff"},
			"contrast_failed/T9:dark"},
	} {
		err := CheckContrast(c.a, c.o)
		if err == nil || field(t, err) != c.want {
			t.Errorf("%s: %v, want %s", c.name, err, c.want)
		}
	}
	// Accepted overrides that inherit the synced grounds (the first was refused on the G017 Light background), and the
	// kit-baseline floor's accepted cases: no or empty overrides, every token explicitly at its kit default, and tenant
	// colours that match or beat the kit on its OWNER-WAIVED entries.
	explicit := func(a Appearance) Overrides {
		o := Overrides{}
		for k, v := range platform[a] {
			o[k] = v
		}
		return o
	}
	for _, c := range []struct {
		name string
		a    Appearance
		o    Overrides
	}{
		{"input on the synced Light background", Light, Overrides{"input": "#949494"}},
		{"sparse Dark-only popover-foreground", Dark, Overrides{"popover-foreground": "#bdbdbd"}},
		{"grey primary passes at rest and at the 80 % hover", Light, Overrides{"primary": "#525252"}},
		{"the T12 twin: mid-grey foreground on the platform input fill", Light, Overrides{"foreground": "#404040"}},
		{"no overrides", Light, nil},
		{"no overrides", Dark, nil},
		{"empty overrides", Light, Overrides{}},
		{"empty overrides", Dark, Overrides{}},
		{"every token explicitly at its kit default", Light, explicit(Light)},
		{"every token explicitly at its kit default", Dark, explicit(Dark)},
		{"the kit Dark destructive explicitly (its 30 % tints are OWNER-WAIVED)", Dark, Overrides{"destructive": "#ff6467"}},
		{"darker primary than the kit (white about 5.6481 at rest, 3.9187 at the Light 80 % hover)", Light,
			Overrides{"primary": "#0062d6"}},
		{"darker primary than the kit (white about 5.6481 at rest, 7.3052 at the 80 % hover over popover)", Dark,
			Overrides{"primary": "#0062d6"}},
		{"primary 80 % hover about 3.5131: below 4.5 but above the kit's 3.1768 (OWNER-WAIVED entry)", Light,
			Overrides{"primary": "#2f6fb8", "primary-foreground": "#ffffff"}},
		{"darker destructive: 20 % tint about 3.8580, below 4.5 but above the kit's 3.3168 (OWNER-WAIVED entry)", Light,
			Overrides{"destructive": "#d2000a"}},
	} {
		if err := CheckContrast(c.a, c.o); err != nil {
			t.Errorf("%s %s: %v, want accepted", c.a, c.name, err)
		}
	}
	// The floors are per entry and per appearance: a waived sibling never lowers an entry the kit meets. Light T4 on
	// muted is waived (about 4.4014) while T4 on background keeps 4.5; Dark T10 at 30 % over background is waived (about
	// 4.3358) while 20 % keeps 4.5; the Light T8 hover is waived (about 3.1768) while the Dark T8 hover keeps 4.5.
	floorOf := func(a Appearance, p pair) float64 {
		for i, q := range pairs {
			if q == p && q.applies(a) {
				return floors[a][i]
			}
		}
		t.Fatalf("%s %s %s is not an evaluated pair", a, p.id, describe(p))
		return 0
	}
	hover := text("T8", solid("primary-foreground"), over("primary", 0.8, "background"))
	for _, c := range []struct {
		name     string
		a        Appearance
		p        pair
		min, max float64
	}{
		{"T4 on background", Light, text("T4", solid("muted-foreground"), solid("background")), 4.5, 4.5},
		{"T4 on muted", Light, text("T4", solid("muted-foreground"), solid("muted")), 4.35, 4.45},
		{"T10 on its 20 % tint", Dark, onlyIn(Dark, text("T10", solid("destructive"), over("destructive", 0.2, "background"))), 4.5, 4.5},
		{"T10 on its 30 % tint", Dark, onlyIn(Dark, text("T10", solid("destructive"), over("destructive", 0.3, "background"))), 4.29, 4.39},
		{"T8 80 % hover", Light, hover, 3.13, 3.23},
		{"T8 80 % hover", Dark, hover, 4.5, 4.5},
	} {
		if f := floorOf(c.a, c.p); f < c.min || f > c.max {
			t.Errorf("%s %s floor = %.4f, want [%.2f, %.2f]", c.a, c.name, f, c.min, c.max)
		}
	}
	// T8 below the floor: white on grey #767676 keeps 4.5:1 at rest (about 4.5422) but its 80 % hover over the white page
	// (about 3.1356) is under the kit's own hover (about 3.1768); #2f6fb8 at the hover (about 3.5131) is above it.
	fg, page := rgb("#ffffff"), rgb("#ffffff")
	g := rgb("#767676")
	if Contrast(fg, g) < 4.5 || Contrast(fg, blend(g, page, 0.8)) >= floorOf(Light, hover) {
		t.Errorf("T8 floor edge: rest %.4f, hover %.4f, floor %.4f", Contrast(fg, g), Contrast(fg, blend(g, page, 0.8)),
			floorOf(Light, hover))
	}
	if r := Contrast(fg, blend(rgb("#2f6fb8"), page, 0.8)); r < floorOf(Light, hover) || r >= 4.5 {
		t.Errorf("T8 waived-entry edge: #2f6fb8 hover %.4f", r)
	}
	// The accepted grey-primary edge (S23 retype of #696969): white on #525252 is about 7.8139 at rest and about 4.6340
	// at the 80 % hover over the white page.
	if r := Contrast(fg, blend(rgb("#525252"), page, 0.8)); r < 4.60 || r > 4.67 {
		t.Errorf("grey primary 80 %% hover = %.4f, want about 4.6340", r)
	}
	// T10: #d2000a as text keeps 4.5:1 on its 10 % tint (about 4.6718) but not on its 20 % hover tint (about 3.8580),
	// which is above the kit's 20 % floor (about 3.3168): accepted on an OWNER-WAIVED entry. #ff0000 on its 10 % tint
	// (about 3.3610) is under the kit's 10 % floor (about 3.9917).
	d, red := rgb("#d2000a"), rgb("#ff0000")
	tint20 := onlyIn(Light, text("T10", solid("destructive"), over("destructive", 0.2, "background")))
	tint10 := onlyIn(Light, text("T10", solid("destructive"), over("destructive", 0.1, "background")))
	if Contrast(d, blend(d, page, 0.1)) < 4.5 || Contrast(d, blend(d, page, 0.2)) >= 4.5 ||
		Contrast(d, blend(d, page, 0.2)) < floorOf(Light, tint20) || Contrast(red, blend(red, page, 0.1)) >= floorOf(Light, tint10) {
		t.Errorf("T10 edge: 10 %% %.4f, 20 %% %.4f (floor %.4f), red 10 %% %.4f (floor %.4f)", Contrast(d, blend(d, page, 0.1)),
			Contrast(d, blend(d, page, 0.2)), floorOf(Light, tint20), Contrast(red, blend(red, page, 0.1)), floorOf(Light, tint10))
	}
	// T9 (Codex P0 disposition, 2026-10-26): the secondary hover is color-mix(in oklch, var(--secondary),
	// var(--foreground) 5%), not an sRGB blend. On the Light default foreground #0c0a09, black on #7a7a7a keeps 4.5:1 at
	// rest (about 4.8924) but not on the OKLCH hover mix (about 4.4723); the sRGB 5 % blend (about 4.5199) would admit
	// it. The grey's hue is powerless, so the mix takes the foreground's hue. No pair fades text over a mix.
	t9 := text("T9", solid("secondary-foreground"), mixed("secondary", "foreground", 0.05))
	evaluated := false
	for _, q := range pairs {
		evaluated = evaluated || q == t9
		if q.bg.mix && q.fg.alpha != 1 {
			t.Errorf("%s %s fades text over a mix", q.id, describe(q))
		}
	}
	if !evaluated {
		t.Error("the T9 OKLCH hover mix is not an evaluated pair")
	}
	grey, black := rgb("#7a7a7a"), rgb("#000000")
	if r := Contrast(black, grey); r < 4.87 || r > 4.91 {
		t.Errorf("T9 edge at rest = %.4f, want about 4.8924", r)
	}
	greyEdge := Effective(Light, Overrides{"secondary": "#7a7a7a", "secondary-foreground": "#000000"})
	if r := t9.ratio(greyEdge); r < 4.46 || r > 4.48 {
		t.Errorf("T9 edge OKLCH hover = %.4f, want about 4.4723", r)
	}
	if r := Contrast(black, blend(rgb("#0c0a09"), grey, 0.05)); r < 4.51 || r > 4.53 {
		t.Errorf("T9 edge sRGB blend = %.4f, want about 4.5199 (the model the OKLCH mix replaces)", r)
	}
	if h := oklch(grey)[2]; !math.IsNaN(h) {
		t.Errorf("grey hue = %v, want powerless", h)
	}
	// The wide-gamut bound: #aa22ff mixed 5 % toward a #55ffff foreground leaves sRGB (linear blue about 1.108). White
	// keeps about 4.5763 on the clipped mix but only about 4.4265 unclipped, and the pair takes the lower.
	lin := mixOKLCH(rgb("#aa22ff"), rgb("#55ffff"), 0.05)
	white := luminance(rgb("#ffffff"))
	c, w := ratioOf(white, luminance(clipped(lin))), ratioOf(white, wideLuminance(lin))
	wide := t9.ratio(Effective(Dark, Overrides{"foreground": "#55ffff", "secondary": "#aa22ff",
		"secondary-foreground": "#ffffff"}))
	if c < 4.56 || c > 4.59 || w < 4.41 || w > 4.44 || math.Abs(wide-w) > 1e-12 {
		t.Errorf("T9 wide-gamut edge: clipped %.4f (about 4.5763), unclipped %.4f (about 4.4265), pair %.4f", c, w, wide)
	}
}

// The seeded custom-host row (infra/identity/mcp-seed.sql:113-114), byte-identical to the browser fixtureBranded
// (web/tests/fixtures/token-scopes/fixtures.ts:91-126), with the S23 / D-M3a (i) test-input retype the Codex P0
// disposition (2026-10-26) approved: Light primary #0a5a4d and destructive #9a2c25, Dark destructive #ff9c8f.
var seedLight = Overrides{
	"background": "#f1f5f4", "foreground": "#13201e", "primary": "#0a5a4d", "primary-foreground": "#ffffff",
	"accent": "#7a4a00", "muted-foreground": "#4a5d59", "border": "#c9d6d3", "popover": "#ffffff",
	"popover-foreground": "#13201e", "muted": "#e6eeec", "secondary": "#3b5a7a", "secondary-foreground": "#ffffff",
	"destructive": "#9a2c25", "input": "#4a5d59", "ring": "#0f6b5c",
}

var seedDark = Overrides{
	"background": "#0b1514", "foreground": "#e2ecea", "primary": "#3fb8a3", "primary-foreground": "#06201b",
	"accent": "#f0c674", "muted-foreground": "#93a8a4", "border": "#24332f", "popover": "#101d1b",
	"popover-foreground": "#e2ecea", "muted": "#0f1b19", "secondary": "#8fb3d9", "secondary-foreground": "#0b1514",
	"destructive": "#ff9c8f", "input": "#93a8a4", "ring": "#f0c674",
}

// The pre-S23 seed values, kept as historical P0.3 evidence: Light primary #0f6b5c (T8 80 % hover about 4.27) and
// destructive #a8322a (T10 20 % tint about 4.42), Dark destructive #ff7a6b (T10 30 % tint about 4.35 / 4.04).
var (
	seedLightPreS23 = with(with(seedLight, "primary", "#0f6b5c"), "destructive", "#a8322a")
	seedDarkPreS23  = with(seedDark, "destructive", "#ff7a6b")
)

// with returns a copy of o with k set to v.
func with(o Overrides, k, v string) Overrides {
	out := Overrides{}
	for key, val := range o {
		out[key] = val
	}
	out[k] = v
	return out
}

// describe names a pair's layers for the P0.3 log, e.g. "primary-foreground on primary/80% over background" or
// "secondary-foreground on secondary mixed 5% foreground in oklch".
func describe(p pair) string {
	name := func(l layer) string {
		switch {
		case l.mix:
			return fmt.Sprintf("%s mixed %.0f%% %s in oklch", l.under, l.alpha*100, l.token)
		case l.alpha == 1:
			return l.token
		case l.under == "":
			return fmt.Sprintf("%s/%.0f%%", l.token, l.alpha*100)
		default:
			return fmt.Sprintf("%s/%.0f%% over %s", l.token, l.alpha*100, l.under)
		}
	}
	return name(p.fg) + " on " + name(p.bg)
}

func verdictText(v string) string {
	if v == "" {
		return "accepted"
	}
	return v
}

// P0.3 (Plan r1 §4, S23; A11): every existing accepted or refused saved-branding fixture runs through all r3 pairs of
// its appearance. Each pair's ratio is logged (go test -v; lines "P0.3 | fixture | appearance | pair | ratio |
// threshold | ok/FAIL") for the compatibility table, and the verdict must equal the recorded expectation: the Plan's
// pre-computed refusals (§0 item 15) and, for every other fixture, its existing accepted or refused contract. A
// different verdict is an S23 stop: it fails with the ratios and needs a recorded resolution proposal (D-M3a) before
// the expectation or the fixture changes; it is never silently re-expected. The P0.3 run of 2026-10-26 found three
// more refusals (seed Dark, sparse-destructive Dark, BRD-1 Light); the Codex P0 disposition approved D-M3a (i), so each
// retyped fixture keeps its old value as a refused row and gains its new value as an accepted row.
//
// G069 kit-baseline floor (owner kit-fidelity direction and waiver, 2026-10-26): each entry is now checked against its
// floor (the lower of its threshold and the kit defaults' own ratio), logged as a further column with the status
// ok / WAIVED (below the threshold, at or above the kit's floor) / FAIL (below the floor). Six refused rows whose only
// failures fall on OWNER-WAIVED entries at or above the kit's own ratios are re-expected as accepted by that approved
// design (Codex, 2026-10-26), each naming its former verdict; every other verdict, including the first failing pair of
// each refusal, is unchanged.
func TestSavedFixtureCompatibility(t *testing.T) {
	const pre = "pre-computed by Plan r1 §0 item 15 / S23"
	const found = "pre-S23 value, refused in the P0.3 run of 2026-10-26; historical evidence"
	const retyped = "S23 / D-M3a (i) retyped test input, Codex P0 disposition 2026-10-26"
	const kept = "existing contract (Plan r1 A11: no other listed fixture fails)"
	const floored = "G069 kit-baseline floor (owner waiver, 2026-10-26): accepted on OWNER-WAIVED entries, formerly "
	for _, c := range []struct {
		name string
		a    Appearance
		o    Overrides
		want string
		why  string
	}{
		{"seed row = fixtureBranded Light, pre-S23", Light, seedLightPreS23, "", floored + "T8:light (" + pre +
			"): 80 % hover about 4.27 and T10 20 % tint about 4.42, above the kit's 3.1768 and 3.3168"},
		{"seed row = fixtureBranded Light", Light, seedLight, "", retyped},
		{"seed row = fixtureBranded Dark, pre-S23", Dark, seedDarkPreS23, "", floored + "T10:dark (" + found +
			"): T10 30 % tint about 4.35 / 4.04, above the kit's 4.3358 / 3.7297"},
		{"seed row = fixtureBranded Dark", Dark, seedDark, "", retyped},
		{"fixtureGreyPrimary Light, pre-S23 #696969", Light, Overrides{"primary": "#696969", "primary-foreground": "#ffffff"},
			"", floored + "T8:light (" + pre + "): 80 % hover about 3.60, above the kit's 3.1768"},
		{"fixtureGreyPrimary Light #525252", Light, Overrides{"primary": "#525252", "primary-foreground": "#ffffff"}, "",
			retyped + ", 80 % hover about 4.63"},
		{"fixtureDarkSparseDestructive Dark, pre-S23 #c07070", Dark, Overrides{"destructive": "#c07070"},
			"contrast_failed/T10:dark", found + ", T10 20 % tint about 4.34 (the kit meets that entry: 4.5 still applies)"},
		{"fixtureDarkSparseDestructive Dark #e8a6a6", Dark, Overrides{"destructive": "#e8a6a6"}, "", retyped},
		{"BRD-1 valid Light, pre-S23 #1d4ed8", Light, Overrides{"primary": "#1d4ed8"}, "",
			floored + "T8:light (" + found + "): 80 % hover about 4.42, above the kit's 3.1768"},
		{"BRD-1 valid Dark, pre-S23 #1d4ed8", Dark, Overrides{"primary": "#1d4ed8"}, "", kept},
		{"BRD-1 valid #1e40af Light", Light, Overrides{"primary": "#1e40af"}, "", retyped + ", 80 % hover about 5.30"},
		{"BRD-1 valid #1e40af Dark", Dark, Overrides{"primary": "#1e40af"}, "", retyped},
		{"BRD-4 dark background Light", Light, Overrides{"background": "#000000"}, "contrast_failed/T1:light", kept},
		{"BRD-4 dark primary loss Dark", Dark, Overrides{"primary": "#ffffff", "primary-foreground": "#ffffff"},
			"contrast_failed/T8:dark", kept},
		{"E-BRD-5 seed with primary-foreground #0f6b5c Light", Light, with(seedLight, "primary-foreground", "#0f6b5c"),
			"contrast_failed/T8:light", kept},
		{"U-BRD-4 light background only", Light, Overrides{"background": "#101010"}, "contrast_failed/T1:light", kept},
		{"U-BRD-4 dark foreground only, too dim", Dark, Overrides{"foreground": "#333333"}, "contrast_failed/T1:dark", kept},
		{"U-BRD-4 primary hover loses the label", Light, Overrides{"primary": "#2f6fb8", "primary-foreground": "#ffffff"},
			"", floored + "T8:light (" + kept + "): 80 % hover about 3.51, above the kit's 3.1768"},
		{"U-BRD-4 ring equal to muted", Light, Overrides{"ring": "#f3f1f1"}, "contrast_failed/N2:light", kept},
		{"U-BRD-4 accent fails on muted", Light, Overrides{"accent": "#6f6f6f", "background": "#ffffff"},
			"contrast_failed/T5:light", kept},
		{"U-BRD-4 muted-foreground within 3:1 of popover", Dark, Overrides{"muted-foreground": "#3a3a3a"},
			"contrast_failed/T4:dark", kept},
		{"U-BRD-4 accent on the synced Dark muted", Dark, Overrides{"accent": "#8a8a8a"}, "contrast_failed/T5:dark", kept},
		{"U-BRD-4 input on the synced Light background", Light, Overrides{"input": "#949494"}, "", kept},
		{"U-BRD-4 sparse Dark-only popover-foreground", Dark, Overrides{"popover-foreground": "#bdbdbd"}, "", kept},
		{"U-BRD-4 grey primary #696969", Light, Overrides{"primary": "#696969"}, "",
			floored + "T8:light (" + pre + "): 80 % hover about 3.60, above the kit's 3.1768"},
		{"U-BRD-5 accent #6b4500 Light", Light, Overrides{"accent": "#6b4500"}, "", kept},
		{"U-BRD-5 white Dark background", Dark, Overrides{"background": "#ffffff"}, "contrast_failed/T1:dark", kept},
	} {
		t.Run(c.name, func(t *testing.T) {
			o, err := ValidateOverrides(string(c.a), c.o)
			if err != nil {
				t.Fatalf("fixture grammar: %v", err)
			}
			eff := Effective(c.a, o)
			first, firstField := "", ""
			for i, p := range pairs {
				if !p.applies(c.a) {
					continue
				}
				r, floor, status := p.ratio(eff), floors[c.a][i], "ok"
				switch {
				case r < floor:
					status = "FAIL"
					if first == "" {
						first = fmt.Sprintf("%s %s = %.4f < floor %.4f", p.id, describe(p), r, floor)
						firstField = CodeContrastFailed + "/" + p.id + ":" + string(c.a)
					}
				case r < p.threshold:
					status = "WAIVED"
				}
				t.Logf("P0.3 | %s | %s | %s %s | %.4f | %.1f | floor %.4f | %s", c.name, c.a, p.id, describe(p), r,
					p.threshold, floor, status)
			}
			got := ""
			if err := CheckContrast(c.a, o); err != nil {
				got = field(t, err)
			}
			if got != firstField {
				t.Errorf("CheckContrast %s disagrees with the logged ratios (first failure %q)", verdictText(got), first)
			}
			t.Logf("P0.3 verdict | %s | %s | %s | expected %s (%s)", c.name, c.a, verdictText(got), verdictText(c.want), c.why)
			if got != c.want {
				t.Errorf("S23 stop: %s %s verdict %s, want %s (%s); first failing pair: %s. Record a Plan r1 D-M3a "+
					"resolution proposal before changing the fixture or this expectation.",
					c.name, c.a, verdictText(got), verdictText(c.want), c.why, first)
			}
		})
	}
}

// U-BRD-5: Validate composes label, grammar and both appearances; a failure never echoes values.
func TestValidate(t *testing.T) {
	l := "Acme"
	// A darker accent than the kit's #8a5a00 keeps T5 and N2; the inherited kit grounds pass through their floors (on the
	// kit defaults this save was refused at T4:light before the G069 floor).
	gotL, lo, do, err := Validate(&l, map[string]string{"accent": "#6B4500"}, map[string]string{})
	if err != nil || *gotL != "Acme" || lo["accent"] != "#6b4500" || len(do) != 0 {
		t.Fatalf("valid config: %v %v %v %v", gotL, lo, do, err)
	}
	_, _, _, err = Validate(nil, nil, map[string]string{"background": "#ffffff"})
	if err == nil || strings.Contains(err.Error(), "#ffffff") || field(t, err) != "contrast_failed/T1:dark" {
		t.Errorf("dark white background: %v", err)
	}
	// G069 kit-baseline floor: an unrelated branding update (label only, no colours) saves on the unchanged kit.
	if gotL, lo, do, err := Validate(&l, nil, nil); err != nil || *gotL != "Acme" || len(lo) != 0 || len(do) != 0 {
		t.Errorf("label-only update: %v %v %v %v", gotL, lo, do, err)
	}
	// Every token explicitly at its kit default (uppercase input, stored lowercase) saves in both appearances.
	upper := func(a Appearance) map[string]string {
		m := map[string]string{}
		for k, v := range platform[a] {
			m[k] = strings.ToUpper(v)
		}
		return m
	}
	if _, lo, do, err := Validate(nil, upper(Light), upper(Dark)); err != nil || lo["primary"] != "#0075ff" ||
		do["destructive"] != "#ff6467" || len(lo) != len(Tokens) || len(do) != len(Tokens) {
		t.Errorf("explicit kit defaults: %v %v %v", lo, do, err)
	}
	// A darker primary than the kit's #0075ff saves in both appearances.
	if _, _, _, err := Validate(nil, map[string]string{"primary": "#0062d6"}, map[string]string{"primary": "#0062d6"}); err != nil {
		t.Errorf("darker primary: %v", err)
	}
	// Below-floor tenant colours are refused with the closed field, never the value: Light is checked before Dark.
	for _, c := range []struct {
		name        string
		light, dark map[string]string
		value, want string
	}{
		{"light grey primary hover under the kit's", map[string]string{"primary": "#767676"}, nil, "767676",
			"contrast_failed/T8:light"},
		{"dark lighter primary under the kit's", nil, map[string]string{"primary": "#3D8BFF"}, "3d8bff",
			"contrast_failed/T8:dark"},
		{"both below floor: Light reported first", map[string]string{"primary": "#767676"},
			map[string]string{"primary": "#3d8bff"}, "767676", "contrast_failed/T8:light"},
	} {
		_, _, _, err := Validate(nil, c.light, c.dark)
		if err == nil || strings.Contains(strings.ToLower(err.Error()), c.value) || field(t, err) != c.want {
			t.Errorf("%s: %v, want %s", c.name, err, c.want)
		}
	}
}
