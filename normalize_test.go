package cadeft

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNormalize_PreservesAccents(t *testing.T) {
	in := "Nom Émetteur — accents survive"
	out, err := normalize(in)
	require.NoError(t, err)
	require.True(t, strings.Contains(out, "Émetteur"), "normalize stripped accents: got %q", out)
}
