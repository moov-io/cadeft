package cadeft

import (
	"unicode"

	"github.com/pkg/errors"
	"golang.org/x/text/unicode/norm"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
)

// normalizer removes diacritical characters and replaces them with their ASCII representation
var normalizer = transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)

func normalize(in string) (string, error) {
	s, _, err := transform.String(normalizer, in)
	if err != nil {
		return "", err
	}
	for i := 0; i < len(s); i++ {
		if s[i] > unicode.MaxASCII {
			return "", errors.Errorf("failed to normalize rune %c", s[i])
		}
	}
	return s, nil
}
