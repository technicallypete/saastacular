package conformance

import (
	"bytes"
	"encoding/json"
	"strings"
	"time"
)

// Tool is an allowed tool and its highest permitted level.
type Tool struct {
	Name  string `json:"name"`
	Level string `json:"level"`
}

// ContextRef is a context reference core has already filtered.
type ContextRef struct {
	Kind string `json:"kind"`
	Ref  string `json:"ref"`
}

// Message is one input message.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Input is the envelope input.
type Input struct {
	Messages []Message `json:"messages"`
}

// Envelope is the execution envelope as a typed Go producer would build it. It is a test-only
// producer for the peer exchange (Plan §8.5); no production client exists.
type Envelope struct {
	RequestID        string       `json:"request_id"`
	TraceID          string       `json:"trace_id"`
	Capability       string       `json:"capability"`
	AuthorizationRef string       `json:"authorization_ref"`
	OrganizationID   string       `json:"organization_id"`
	GroupID          string       `json:"group_id,omitempty"`
	BusinessEntityID string       `json:"business_entity_id,omitempty"`
	UserID           string       `json:"user_id"`
	AllowedTools     []Tool       `json:"allowed_tools"`
	Deadline         string       `json:"deadline"`
	ContextRefs      []ContextRef `json:"context_refs"`
	CallbackGrantRef string       `json:"callback_grant_ref,omitempty"`
	Input            Input        `json:"input"`
}

// ProduceEnvelopes builds the three peer-exchange envelopes: minimal (required fields only),
// full (every optional field) and boundary (maximum lengths and a far-future deadline with an
// offset). Every deadline is in the future relative to now. All values are synthetic.
func ProduceEnvelopes(now time.Time) []Envelope {
	utc := func(d time.Duration) string { return now.Add(d).UTC().Format("2006-01-02T15:04:05.000Z") }
	minimal := Envelope{
		RequestID:        "go-peer-minimal-0001",
		TraceID:          "4bf92f3577b34da6a3ce929d0e0e4736",
		Capability:       "chat",
		AuthorizationRef: "authz-go-peer-minimal",
		OrganizationID:   "6f1c2a1e-6d3b-4d0e-9a51-0c9f5a7e2b10",
		UserID:           "0d7e6b3c-2a41-4f58-8c7d-5b2e9f1a3c40",
		AllowedTools:     []Tool{},
		Deadline:         utc(15 * time.Minute),
		ContextRefs:      []ContextRef{},
		Input:            Input{Messages: []Message{{Role: "user", Content: "Summarize the synthetic note."}}},
	}
	full := Envelope{
		RequestID:        "go-peer-full-0002",
		TraceID:          "a3ce929d0e0e47364bf92f3577b34da6",
		Capability:       "chat",
		AuthorizationRef: "authz-go-peer-full",
		OrganizationID:   "6f1c2a1e-6d3b-4d0e-9a51-0c9f5a7e2b10",
		GroupID:          "9b8a7c6d-5e4f-4a3b-8c2d-1e0f9a8b7c60",
		BusinessEntityID: "3c4d5e6f-7a8b-4c9d-8e0f-1a2b3c4d5e70",
		UserID:           "0d7e6b3c-2a41-4f58-8c7d-5b2e9f1a3c40",
		AllowedTools: []Tool{
			{Name: "notes.search", Level: "read"},
			{Name: "notes.draft", Level: "propose"},
			{Name: "notes.save", Level: "commit"},
		},
		Deadline: utc(15 * time.Minute),
		ContextRefs: []ContextRef{
			{Kind: "document", Ref: "doc-synthetic-0001"},
			{Kind: "note.section", Ref: "note-synthetic-0002#s1"},
		},
		CallbackGrantRef: "grant-go-peer-full",
		Input: Input{Messages: []Message{
			{Role: "system", Content: "You are a synthetic test assistant."},
			{Role: "user", Content: "Draft a short synthetic summary."},
			{Role: "assistant", Content: "Here is a synthetic draft."},
		}},
	}
	boundary := Envelope{
		RequestID:        "go.peer_boundary-" + strings.Repeat("b", 47),
		TraceID:          "0123456789abcdef0123456789abcdef",
		Capability:       "chat",
		AuthorizationRef: strings.Repeat("r", 512),
		OrganizationID:   "6F1C2A1E-6D3B-4D0E-9A51-0C9F5A7E2B10",
		UserID:           "0D7E6B3C-2A41-4F58-8C7D-5B2E9F1A3C40",
		AllowedTools:     []Tool{{Name: strings.Repeat("t", 128), Level: "read"}},
		Deadline:         now.Add(24 * time.Hour).In(time.FixedZone("", 2*3600)).Format("2006-01-02T15:04:05.000-07:00"),
		ContextRefs:      []ContextRef{{Kind: "k" + strings.Repeat("x", 63), Ref: strings.Repeat("f", 256)}},
		CallbackGrantRef: strings.Repeat("g", 512),
		Input:            Input{Messages: []Message{{Role: "user", Content: "Grüße – boundary ✓"}}},
	}
	return []Envelope{minimal, full, boundary}
}

// EncodeNDJSON writes one compact JSON document per line, each LF-terminated.
func EncodeNDJSON(envelopes []Envelope) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	for _, e := range envelopes {
		if err := enc.Encode(e); err != nil {
			return nil, err
		}
	}
	return buf.Bytes(), nil
}
