package cadeft

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseFileFooter(t *testing.T) {
	r := require.New(t)
	type testCase struct {
		in       string
		expected FileFooter
	}
	cases := map[string]testCase{
		"valid footer": {
			in: "Z000000004000000061000010000000002016500000005000000000722200000000600000000000000000000000000000000000000000000",
			expected: FileFooter{
				RecordHeader: RecordHeader{
					RecordType:      FooterRecord,
					OriginatorID:    "0000000610",
					FileCreationNum: 1,
					recordCount:     4,
				},
				TotalValueOfDebit:  20165,
				TotalCountOfDebit:  5,
				TotalValueOfCredit: 72220,
				TotalCountOfCredit: 6,
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			ff := FileFooter{}
			err := ff.Parse(tc.in)
			r.NoError(err)
			r.Equal(tc.expected, ff)
		})
	}
}

func TestBuildFileFooter(t *testing.T) {
	r := require.New(t)
	simpleDebitTxn := Debit{BaseTxn: BaseTxn{Amount: 1000}}
	simpleDebitReturnTxn := DebitReturn{BaseTxn: BaseTxn{Amount: 1000}}
	simpleCreditTxn := Credit{BaseTxn: BaseTxn{Amount: 1000}}
	simpleCreditReturn := CreditReturn{BaseTxn: BaseTxn{Amount: 1000}}
	txns := []Transaction{&simpleDebitTxn, &simpleCreditTxn, &simpleDebitReturnTxn, &simpleCreditReturn}
	recordHeader := RecordHeader{
		RecordType:      FooterRecord,
		OriginatorID:    "0000000610",
		FileCreationNum: 1,
	}
	ff := NewFileFooter(recordHeader, txns)
	r.Equal(int64(2000), ff.TotalValueOfCredit)
	r.Equal(int64(2000), ff.TotalValueOfDebit)
	r.Equal(int64(2), ff.TotalCountOfCredit)
	r.Equal(int64(2), ff.TotalCountOfDebit)
	s, err := ff.Build()
	r.NoError(err)
	expectedFile := "Z000000000000000061000010000000000200000000002000000000020000000000200000000000000000000000000000000000000000000" + strings.Repeat(" ", 1352)
	r.Equal(expectedFile, s)
}
