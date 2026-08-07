package cadeft

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReadFile_ShortTxnLineNoPanic(t *testing.T) {
	inputs := []string{"", "C", "D", "A", "Z", "Cshort"}
	for _, in := range inputs {
		require.NotPanics(t, func() {
			r := NewReader(strings.NewReader(in))
			_, _ = r.ReadFile()
		}, "input %q", in)
	}
}
