// Package called calls the symbol of the synthetic advisory GO-2099-0001 (strings.ToUpper): a called finding.
package called

import "strings"

// Do upper-cases s.
func Do(s string) string { return strings.ToUpper(s) }
