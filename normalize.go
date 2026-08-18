package cadeft

import "golang.org/x/text/unicode/norm"

// normalize returns the NFC-canonical form of in. Previous versions of
// this function aborted on non-ASCII input; that's gone — the byte-indexed
// fixed-width parsers in reader.go and the per-record Parse methods are
// rune-indexed (Spendbase fork 2026-05-04), so French / accented Latin
// flows through unchanged. NFC is applied so equivalent forms (precomposed
// "É" U+00C9 vs. decomposed "E"+U+0301) compare equal downstream.
func normalize(in string) (string, error) {
	return norm.NFC.String(in), nil
}
