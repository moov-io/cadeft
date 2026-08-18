//go:build tools

package cadeft

// Pin golang.org/x/net to a patched release so Dependabot can resolve
// GHSA-5cv4-jp36-h3mw (HTML parser DoS in versions before v0.55.0).
import _ "golang.org/x/net/idna"
