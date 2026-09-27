// Package baseline applies the shared response policy from baseline/policy.json.
//
// The JSON file is the contract shared with the language-agnostic template and
// the Node oracle tests under tests/. This package parses that policy, preserves
// CSP directive order, and sets headers and cookie serialisation rules for every
// HTTP reply. Do not invent a weaker header set for convenience.
//
// See docs/frontend-security.md and docs/frontend-performance.md.
package baseline
