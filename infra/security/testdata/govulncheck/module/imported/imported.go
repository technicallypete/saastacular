// Package imported imports encoding/csv without calling the GO-2099-0002 symbol (csv.NewReader): an imported finding.
package imported

import "encoding/csv"

// ErrRef refers to the package without calling the advisory symbol.
var ErrRef = csv.ErrFieldCount
