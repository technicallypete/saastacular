package branding

import (
	"math"
	"strconv"
)

// Appearance is Light or Dark.
type Appearance string

// The appearances.
const (
	Light Appearance = "light"
	Dark  Appearance = "dark"
)

// platform are the app's effective rendered defaults for the 15 tokens: exactly the adopted kit's
// web/src/styles/theme.css values, which the app renders unaltered (G069 owner kit-fidelity direction, 2026-10-26: the
// kit exactly, its contrast shortfalls waived; no app-owned colour correction remains). Saved accent paints the kit's
// --vx-gold (D-M2) and saved primary its --vx-blue (D-M1).
//
// Kit oklch values are converted to sRGB hex (OKLab -> LMS -> linear sRGB -> sRGB transfer, each channel clamped to
// [0, 1] and rounded to 8 bits; the 0-255 channels before rounding are listed):
//
//	oklch(1 0 0)                 = 255.00 255.00 255.00 = #ffffff  Light background, popover
//	oklch(0.147 0.004 49.3)      =  11.97  10.04   9.06 = #0c0a09  Light foreground, popover-foreground; Dark background
//	oklch(0.96 0.002 17.2)       = 243.01 241.14 241.14 = #f3f1f1  Light muted, secondary
//	oklch(0.214 0.009 43.1)      =  28.98  23.94  21.92 = #1d1816  Light secondary-foreground; Dark popover
//	oklch(0.922 0.005 34.3)      = 232.39 228.01 226.91 = #e8e4e3  Light border, input
//	oklch(0.547 0.021 43.1)      = 123.93 109.03 103.06 = #7c6d67  Light muted-foreground
//	oklch(0.577 0.245 27.325)    = 231.15   0.00  10.76 = #e7000b  Light destructive (green clamped)
//	oklch(0.986 0.002 67.8)      = 251.28 250.11 248.94 = #fbfaf9  Dark foreground, popover- and secondary-foreground
//	oklch(0.268 0.011 36.5)      =  43.00  36.16  34.21 = #2b2422  Dark muted, secondary
//	oklch(0.714 0.014 41.2)      = 170.93 160.04 156.08 = #aba09c  Dark muted-foreground
//	oklch(0.704 0.191 22.216)    = 255.00  99.74 102.98 = #ff6467  Dark destructive (red clamped)
//
// A saved colour is opaque #rrggbb, so the kit's two translucent Dark values are represented as they paint over the
// Dark background (source-over in sRGB: 255 x alpha + channel x (1 - alpha), on the unrounded background channels):
//
//	oklch(1 0 0 / 10%) = 36.28 34.53 33.66 = #242322  Dark border
//	oklch(1 0 0 / 15%) = 48.43 46.78 45.95 = #302f2e  Dark input
//
// The kit's --vx-blue is #0075ff in both appearances (primary), its --vx-gold (#8a5a00 Light, #f8ce8a Dark) is both
// accent and --ring, and --primary-foreground is #ffffff.
var platform = map[Appearance]map[string]string{
	Light: {
		"background": "#ffffff", "foreground": "#0c0a09", "popover": "#ffffff", "popover-foreground": "#0c0a09",
		"accent": "#8a5a00", "ring": "#8a5a00", "border": "#e8e4e3", "muted": "#f3f1f1",
		"secondary": "#f3f1f1", "secondary-foreground": "#1d1816", "primary-foreground": "#ffffff",
		"primary": "#0075ff", "muted-foreground": "#7c6d67", "input": "#e8e4e3", "destructive": "#e7000b",
	},
	Dark: {
		"background": "#0c0a09", "foreground": "#fbfaf9", "popover": "#1d1816", "popover-foreground": "#fbfaf9",
		"accent": "#f8ce8a", "ring": "#f8ce8a", "muted": "#2b2422", "muted-foreground": "#aba09c",
		"secondary": "#2b2422", "secondary-foreground": "#fbfaf9", "primary-foreground": "#ffffff",
		"primary": "#0075ff", "border": "#242322", "input": "#302f2e", "destructive": "#ff6467",
	},
}

// Effective returns one appearance's effective palette: override ?? platform for each of the 15 tokens. G069 withdrew
// the internal text roles (ink, text-2, text-3): the kit paints secondary text as foreground at 80 %, which the pairs
// composite themselves, so a foreground override no longer fans out.
func Effective(a Appearance, o Overrides) map[string]string {
	eff := map[string]string{}
	for k, v := range platform[a] {
		eff[k] = v
	}
	for k, v := range o {
		eff[k] = v
	}
	return eff
}

// layer is one painted colour: token at alpha over the opaque token under it (alpha 1: the token alone). A foreground
// layer with alpha below 1 composites over its pair's resolved background instead, so its under is empty. A mix layer
// is under mixed with token at weight alpha in OKLCH (CSS color-mix), not alpha-composited.
type layer struct {
	token string
	alpha float64
	under string
	mix   bool
}

func solid(token string) layer { return layer{token: token, alpha: 1} }

// over is token at alpha over under: the kit's bg-<token>/<alpha> utilities.
func over(token string, alpha float64, under string) layer {
	return layer{token: token, alpha: alpha, under: under}
}

// faded is text in token at alpha over the pair's background: the kit's text-<token>/<alpha> utilities.
func faded(token string, alpha float64) layer { return layer{token: token, alpha: alpha} }

// mixed is under mixed with token at weight in OKLCH: the kit's secondary hover, color-mix(in oklch,
// var(--secondary), var(--foreground) 5%), which mixOKLCH models. Only a background is ever mixed.
func mixed(under, token string, weight float64) layer {
	return layer{token: token, alpha: weight, under: under, mix: true}
}

// pair is one pinned contrast pair (Plan r1 §2.6, "Core palette delta r3"): fg on bg at threshold, evaluated in only
// that appearance when only is set. The whole-element brightness factors of G067 are withdrawn: the kit Button
// variants the app uses paint their states with alpha, which the layers model.
type pair struct {
	id        string
	fg, bg    layer
	threshold float64
	only      Appearance
}

func text(id string, fg, bg layer) pair    { return pair{id, fg, bg, 4.5, ""} }
func nonText(id string, fg, bg layer) pair { return pair{id, fg, bg, 3, ""} }

func onlyIn(a Appearance, p pair) pair {
	p.only = a
	return p
}

// pairs are T1-T10 and T12 (text, 4.5:1) and N2-N4 (non-text, 3:1), in ID order so the first failing pair is the one
// reported. Each entry's threshold is capped by the kit baseline (floors): an entry the platform defaults miss requires
// only the kit's own ratio. T8 is the default Button (hover:bg-primary/80), T9 the secondary Button (a 5 % foreground
// mix in OKLCH on hover), T10 the destructive Button and error text (text-destructive on bg-destructive/10 and /20 in
// Light, /20 and /30 in Dark), T12 the kit field fill (bg-input/50; Dark outline hover bg-input/30); the outline and
// ghost hovers are foreground on muted (T3). No pair carries the fixed-white CtaButton label: it renders only under the
// platform palette (S24), so TestDefaultsPass pins it (OWNER-WAIVED, below 4.5:1 on the kit's #0075ff).
var pairs = []pair{
	text("T1", solid("foreground"), solid("background")),
	text("T2", solid("foreground"), solid("background")), text("T2", faded("foreground", 0.8), solid("background")),
	text("T3", solid("foreground"), solid("muted")), text("T3", faded("foreground", 0.8), solid("muted")),
	text("T4", solid("muted-foreground"), solid("background")), text("T4", solid("muted-foreground"), solid("muted")),
	text("T5", solid("accent"), solid("background")), text("T5", solid("accent"), solid("muted")),
	text("T5", solid("accent"), solid("popover")),
	text("T6", solid("popover-foreground"), solid("popover")), text("T6", solid("foreground"), solid("popover")),
	text("T6", faded("foreground", 0.8), solid("popover")),
	text("T7", solid("foreground"), solid("border")),
	text("T8", solid("primary-foreground"), solid("primary")),
	text("T8", solid("primary-foreground"), over("primary", 0.8, "background")),
	text("T8", solid("primary-foreground"), over("primary", 0.8, "popover")),
	text("T9", solid("secondary-foreground"), solid("secondary")),
	text("T9", solid("secondary-foreground"), mixed("secondary", "foreground", 0.05)),
	onlyIn(Light, text("T10", solid("destructive"), over("destructive", 0.1, "background"))),
	onlyIn(Light, text("T10", solid("destructive"), over("destructive", 0.2, "background"))),
	onlyIn(Light, text("T10", solid("destructive"), over("destructive", 0.1, "popover"))),
	onlyIn(Light, text("T10", solid("destructive"), over("destructive", 0.2, "popover"))),
	onlyIn(Dark, text("T10", solid("destructive"), over("destructive", 0.2, "background"))),
	onlyIn(Dark, text("T10", solid("destructive"), over("destructive", 0.3, "background"))),
	onlyIn(Dark, text("T10", solid("destructive"), over("destructive", 0.2, "popover"))),
	onlyIn(Dark, text("T10", solid("destructive"), over("destructive", 0.3, "popover"))),
	text("T10", solid("destructive"), solid("background")),
	text("T12", solid("foreground"), over("input", 0.5, "background")),
	text("T12", solid("foreground"), over("input", 0.5, "popover")),
	onlyIn(Dark, text("T12", solid("foreground"), over("input", 0.3, "background"))),
	nonText("N2", solid("ring"), solid("background")), nonText("N2", solid("ring"), solid("muted")),
	nonText("N2", solid("ring"), solid("popover")),
	nonText("N3", solid("input"), solid("background")), nonText("N3", solid("input"), solid("popover")),
	nonText("N4", solid("muted-foreground"), solid("popover")),
}

// applies reports whether the pair is evaluated in appearance a.
func (p pair) applies(a Appearance) bool { return p.only == "" || p.only == a }

// ratio is the pair's contrast on an effective palette: the background is painted first (over its under token when
// translucent, or mixed into it), then the foreground over that result. A mix can fall outside sRGB: an sRGB surface
// paints it clipped and a wide-gamut surface unclipped, so the pair takes the lower of the two ratios (no pair fades
// text over a mix; TestContrastEdges pins that).
func (p pair) ratio(eff map[string]string) float64 {
	if p.bg.mix {
		lin := mixOKLCH(rgb(eff[p.bg.under]), rgb(eff[p.bg.token]), p.bg.alpha)
		fg := luminance(rgb(eff[p.fg.token]))
		return math.Min(ratioOf(fg, luminance(clipped(lin))), ratioOf(fg, wideLuminance(lin)))
	}
	bg := rgb(eff[p.bg.token])
	if p.bg.alpha < 1 {
		bg = blend(bg, rgb(eff[p.bg.under]), p.bg.alpha)
	}
	fg := rgb(eff[p.fg.token])
	if p.fg.alpha < 1 {
		fg = blend(fg, bg, p.fg.alpha)
	}
	return Contrast(fg, bg)
}

// floors[a][i] is the contrast pairs[i] must keep in appearance a: the lower of its threshold and its own ratio on the
// platform kit defaults (Effective(a, nil), no tenant overrides), computed once with the same ratio evaluator. The
// owner waived the adopted kit's contrast shortfalls (G069 kit-fidelity direction, 2026-10-26): an entry the kit meets
// keeps its full threshold, and an entry the kit misses requires at least the kit's own ratio, so a tenant colour may
// match or beat the kit but never fall below it. Each entry, state and appearance has its own floor; a waiver never
// lowers another's. A below-threshold baseline is a waived outcome, never a threshold pass. An unchanged or explicitly
// default palette evaluates the same inputs through the same code, so its ratios equal the floors exactly.
var floors = kitFloors()

func kitFloors() map[Appearance][]float64 {
	out := map[Appearance][]float64{}
	for _, a := range []Appearance{Light, Dark} {
		base := Effective(a, nil)
		out[a] = make([]float64, len(pairs))
		for i, p := range pairs {
			out[a][i] = math.Min(p.threshold, p.ratio(base))
		}
	}
	return out
}

// CheckContrast evaluates every pinned pair on one appearance's effective palette against its kit-baseline floor. The
// first failing pair is contrast_failed with the closed field "<pair id>:<appearance>"; values never appear.
func CheckContrast(a Appearance, o Overrides) error {
	eff, floor := Effective(a, o), floors[a]
	for i, p := range pairs {
		if p.applies(a) && p.ratio(eff) < floor[i] {
			return &Error{Code: CodeContrastFailed, Field: p.id + ":" + string(a)}
		}
	}
	return nil
}

// rgb parses a validated #rrggbb value into sRGB channels 0-255.
func rgb(hex string) [3]float64 {
	var c [3]float64
	for i := range 3 {
		v, _ := strconv.ParseUint(hex[1+2*i:3+2*i], 16, 8)
		c[i] = float64(v)
	}
	return c
}

// blend is source-over compositing of fg at alpha over an opaque bg, per sRGB channel: what a browser paints for an
// alpha utility (bg-primary/80, text-foreground/80). It is not a colour mix: the kit's secondary hover is
// color-mix(in oklch, ...), which mixOKLCH models (Codex P0 disposition, 2026-10-26).
func blend(fg, bg [3]float64, alpha float64) [3]float64 {
	for i := range fg {
		fg[i] = fg[i]*alpha + bg[i]*(1-alpha)
	}
	return fg
}

// achromatic is CSS Color 4's OKLab-to-OKLCH epsilon: at or below this chroma the hue is powerless (missing). Hex
// greys convert to a chroma near 1e-8.
const achromatic = 0.000004

// oklch converts sRGB channels 0-255 to OKLCH (lightness, chroma, hue in degrees; the hue is NaN when powerless):
// linear sRGB to LMS to OKLab with Ottosson's matrices, which define CSS Color 4's OKLab.
func oklch(c [3]float64) [3]float64 {
	lr, lg, lb := linear(c[0]), linear(c[1]), linear(c[2])
	l := math.Cbrt(0.4122214708*lr + 0.5363325363*lg + 0.0514459929*lb)
	m := math.Cbrt(0.2119034982*lr + 0.6806995451*lg + 0.1073969566*lb)
	s := math.Cbrt(0.0883024619*lr + 0.2817188376*lg + 0.6299787005*lb)
	lightness := 0.2104542553*l + 0.7936177850*m - 0.0040720468*s
	a := 1.9779984951*l - 2.4285922050*m + 0.4505937099*s
	b := 0.0259040371*l + 0.7827717662*m - 0.8086757660*s
	chroma, hue := math.Hypot(a, b), math.NaN()
	if chroma > achromatic {
		hue = math.Mod(math.Atan2(b, a)*180/math.Pi+360, 360)
	}
	return [3]float64{lightness, chroma, hue}
}

// mixOKLCH is CSS Color 4 color-mix(in oklch, base, add weight) of two opaque colours, returned as unclipped linear
// sRGB: lightness and chroma interpolate linearly and the hue along the shorter arc; a powerless hue is missing and
// takes the other colour's (both missing: 0, where the chroma is about 0 and the hue moves nothing).
func mixOKLCH(base, add [3]float64, weight float64) [3]float64 {
	p, q := oklch(base), oklch(add)
	switch {
	case math.IsNaN(p[2]) && math.IsNaN(q[2]):
		p[2], q[2] = 0, 0
	case math.IsNaN(p[2]):
		p[2] = q[2]
	case math.IsNaN(q[2]):
		q[2] = p[2]
	}
	d := q[2] - p[2]
	if d > 180 {
		d -= 360
	} else if d < -180 {
		d += 360
	}
	lightness := p[0]*(1-weight) + q[0]*weight
	chroma := p[1]*(1-weight) + q[1]*weight
	hue := (p[2] + d*weight) * math.Pi / 180
	a, b := chroma*math.Cos(hue), chroma*math.Sin(hue)
	l := lightness + 0.3963377774*a + 0.2158037573*b
	m := lightness - 0.1055613458*a - 0.0638541728*b
	s := lightness - 0.0894841775*a - 1.2914855480*b
	l, m, s = l*l*l, m*m*m, s*s*s
	return [3]float64{
		4.0767416621*l - 3.3077115913*m + 0.2309699292*s,
		-1.2684380046*l + 2.6097574011*m - 0.3413193965*s,
		-0.0041960863*l - 0.7034186147*m + 1.7076147010*s,
	}
}

// clipped is unclipped linear sRGB as an sRGB surface paints it: each channel clipped to [0, 1] (browsers clip an
// out-of-gamut colour to the surface; the kit's oklch values above are converted the same way) and encoded to 0-255.
func clipped(lin [3]float64) [3]float64 {
	var c [3]float64
	for i, v := range lin {
		v = math.Min(1, math.Max(0, v))
		if v <= 0.0031308 {
			c[i] = 255 * 12.92 * v
		} else {
			c[i] = 255 * (1.055*math.Pow(v, 1/2.4) - 0.055)
		}
	}
	return c
}

// wideLuminance is the relative luminance of unclipped linear sRGB: the colour as a wide-gamut surface paints it.
func wideLuminance(lin [3]float64) float64 {
	return math.Min(1, math.Max(0, 0.2126*lin[0]+0.7152*lin[1]+0.0722*lin[2]))
}

// scale models the CSS brightness(f) filter on the whole rendered element: each sRGB channel times f, clamped to
// [0, 255] (the filter's linear component transfer). G069 keeps it only for the platform-default CtaButton contract
// (S24: hover:brightness-110, active:brightness-90), which TestDefaultsPass pins; no tenant pair uses it.
func scale(c [3]float64, f float64) [3]float64 {
	for i := range c {
		c[i] = math.Min(255, math.Max(0, c[i]*f))
	}
	return c
}

// linear is one sRGB channel 0-255 decoded to linear light 0-1.
func linear(v float64) float64 {
	v /= 255
	if v <= 0.04045 {
		return v / 12.92
	}
	return math.Pow((v+0.055)/1.055, 2.4)
}

// luminance is the WCAG 2.x relative luminance.
func luminance(c [3]float64) float64 {
	return 0.2126*linear(c[0]) + 0.7152*linear(c[1]) + 0.0722*linear(c[2])
}

// Contrast is the WCAG 2.x contrast ratio of two colours.
func Contrast(a, b [3]float64) float64 { return ratioOf(luminance(a), luminance(b)) }

// ratioOf is the WCAG 2.x contrast ratio of two relative luminances.
func ratioOf(la, lb float64) float64 {
	if la < lb {
		la, lb = lb, la
	}
	return (la + 0.05) / (lb + 0.05)
}
