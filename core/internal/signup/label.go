package signup

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"

	"core/internal/hostname"
)

// The generated platform address (G057 Spec "Friendly generated tenant addresses", owner amendment 2026-10-29, which
// supersedes the 128-bit label of the 2026-10-23 amendment for new Organizations): every new Organization receives a
// generated, stable, random adjective-noun-four-digit DNS label such as gusty-flowerpot-2643 as its public platform
// hostname <label>.<PLATFORM_BASE_DOMAIN>; the friendly slug is metadata and never routes. The label is drawn by trusted
// core code only, never by a client, and is a public routing identifier, not a credential or an authorization factor.
// Labels issued before the amendment (26 lowercase base32 characters) stay valid and keep resolving: the stored
// platform_label grammar and the hostname classifier are unchanged.

// LabelBytes is the CSPRNG input of one generated platform label: one byte each for the adjective and the noun (the
// low 6 bits index a 64-word list, an exact power of two, so the choice is uniform) and four bytes for the digits.
const LabelBytes = 6

// labelWords is the size of each word list; it must stay 64 so the masked byte index is uniform.
const labelWords = 64

// labelCombinations is the number of distinct labels: 64 adjectives x 64 nouns x 10000 digit groups (40,960,000).
const labelCombinations = labelWords * labelWords * 10000

// maxLabelDraws bounds the label draws of one Provision call: the first draw plus one shared budget of 3 regenerations
// across reserved-label rejections and label_taken outcomes. Exhaustion is signup_unavailable.
const maxLabelDraws = 4

// The curated lists are small, generic and lowercase [a-z] only (DNS-safe), with no offensive, product-specific or
// reserved-host terms. A change keeps each list at exactly labelWords distinct words.
var labelAdjectives = [labelWords]string{
	"amber", "azure", "bold", "brave", "bright", "brisk", "calm", "clever",
	"cobalt", "cosmic", "crisp", "curious", "daring", "dusty", "eager", "early",
	"fancy", "fluffy", "gentle", "gifted", "golden", "grand", "gusty", "happy",
	"hardy", "hasty", "honest", "jolly", "keen", "kind", "lively", "lucky",
	"lunar", "mellow", "merry", "mighty", "misty", "modest", "nimble", "noble",
	"patient", "plucky", "polite", "proud", "quick", "quiet", "rapid", "rosy",
	"royal", "rustic", "shiny", "silent", "silver", "smooth", "snowy", "solar",
	"speedy", "steady", "sunny", "swift", "tidy", "vivid", "warm", "witty",
}

var labelNouns = [labelWords]string{
	"acorn", "anchor", "badger", "basket", "beacon", "biscuit", "bridge", "button",
	"canyon", "castle", "cedar", "cherry", "cloud", "clover", "comet", "compass",
	"cricket", "daisy", "delta", "ember", "falcon", "feather", "ferry", "flowerpot",
	"forest", "garden", "glacier", "harbor", "hazel", "heron", "island", "jungle",
	"kettle", "lantern", "lemon", "maple", "meadow", "marble", "meteor", "orchard",
	"otter", "panda", "pebble", "pepper", "pillow", "planet", "pocket", "pond",
	"prairie", "puffin", "quartz", "rabbit", "river", "robin", "saddle", "sparrow",
	"spruce", "summit", "thistle", "tulip", "valley", "violet", "walnut", "willow",
}

// GenerateLabel draws one platform label from r: exactly LabelBytes bytes become <adjective>-<noun>-<dddd> (the digits
// zero-padded to four). The digit group is a 32-bit value modulo 10000, a bias below one part in ten thousand of a public
// identifier that DB uniqueness, not secrecy, protects. A short or failing reader is an error (never a shorter label).
func GenerateLabel(r io.Reader) (string, error) {
	b := make([]byte, LabelBytes)
	if _, err := io.ReadFull(r, b); err != nil {
		return "", err
	}
	digits := binary.BigEndian.Uint32(b[2:]) % 10000
	return fmt.Sprintf("%s-%s-%04d", labelAdjectives[b[0]&(labelWords-1)], labelNouns[b[1]&(labelWords-1)], digits), nil
}

// labelReader is the label entropy source: crypto/rand unless a test injected Config.LabelRand.
func (s *Service) labelReader() io.Reader {
	if s.cfg.LabelRand != nil {
		return s.cfg.LabelRand
	}
	return rand.Reader
}

// labelIssuable reports whether a drawn label may be issued under the active configuration: the slug grammar and not a
// reserved first label (the floor, PLATFORM_RESERVED_LABELS and the first label of the CNAME target and auth host under
// the platform domain), the predicate Classify uses, so an issued label is never answered as a reserved host. The SQL
// CHECK holds only the static floor; configuration-specific values are enforced here.
func (s *Service) labelIssuable(label string) bool {
	return hostname.SlugGrammar(label) && !hostname.IsReservedLabel(label, s.cfg.Hostname)
}
