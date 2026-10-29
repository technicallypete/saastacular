// Command readiness evaluates encryption and transport readiness from non-secret configuration (G012 Plan r1 §6.2).
// It reads KEY=VALUE lines on stdin (the runtime transport names, PG* if present, READINESS_GATEWAY_ENABLED and
// READINESS_EVIDENCE_<ITEM> references), prints a JSON report and exits 0 (configuration-ready), 3 (blocked) or 2
// (invalid input). Secret-named input is refused. The result is configuration-only: external items are declared, not
// verified, and it is never a deployed-platform attestation. Run it in the network-free core check image; it is in no
// runtime image and serves no route.
//
//	readiness <config.env
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"core/internal/config"
)

func main() {
	os.Exit(run(os.Stdin, os.Stdout, os.Stderr, config.CurrentCapabilities()))
}

func run(stdin io.Reader, stdout, stderr io.Writer, caps config.Capabilities) int {
	in, err := config.ParseReadinessInput(stdin)
	if err != nil {
		fmt.Fprintln(stderr, "readiness:", err)
		return 2
	}
	rep, err := config.AssessReadiness(in, caps)
	if err != nil {
		fmt.Fprintln(stderr, "readiness:", err)
		return 2
	}
	enc := json.NewEncoder(stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(rep); err != nil {
		fmt.Fprintln(stderr, "readiness: cannot write the report")
		return 2
	}
	if rep.Result == config.ReadinessReady {
		return 0
	}
	return 3
}
