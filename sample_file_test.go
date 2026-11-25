package cadeft_test

import (
	"os"
	"testing"

	"github.com/moov-io/cadeft"
	"github.com/stretchr/testify/require"
)

func TestSampleFileCO14821(t *testing.T) {
	f, err := os.Open("sample_files/CO14821.txt")
	require.NoError(t, err)
	defer f.Close()

	eftFile, err := cadeft.NewReader(f).ReadFile()
	require.NoError(t, err)
	require.NotNil(t, eftFile.Header)
	require.NotNil(t, eftFile.Footer)
	require.Len(t, eftFile.Txns, 11)
	require.Equal(t, int64(160146), eftFile.Footer.TotalValueOfCredit)
	require.Equal(t, int64(6), eftFile.Footer.TotalCountOfCredit)
	require.Equal(t, int64(20165), eftFile.Footer.TotalValueOfDebit)
	require.Equal(t, int64(5), eftFile.Footer.TotalCountOfDebit)
	require.NoError(t, eftFile.Validate())
}
