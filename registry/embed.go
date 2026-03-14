// Package registrydata embeds the bundled skill registry JSON.
// It is a thin data-only package; use internal/registry for logic.
package registrydata

import _ "embed"

//go:embed skills.json
var Skills []byte
