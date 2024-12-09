package cadeft

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestParseDebitReverse(t *testing.T) {
	r := require.New(t)
	type testCase struct {
		in          string
		expectedTxn DebitReverse
		expectErr   any
	}
	cases := map[string]testCase{
		"regular txn": {
			in: "45000000040420231370612148211010000003036121006100001000000003000BANK OF MONTREAD1-1-OCC ZERO                 BANK OF MONTREA               00000001101140               0001012618989899                                            00000000000",
			expectedTxn: DebitReverse{
				BaseTxn: BaseTxn{
					TxnType:               TransactionType("450"),
					Amount:                int64(4042),
					InstitutionID:         "061214821",
					ItemTraceNo:           "6121006100001000000003",
					StoredTransactionType: "000",
					OriginatorShortName:   "BANK OF MONTREA",
					OriginatorLongName:    "BANK OF MONTREA",
					UserID:                "0000000110",
					CrossRefNo:            "1140",
					SundryInfo:            "",
					SettlementCode:        "",
					RecordType:            DebitReverseRecord,
				},
				DueDate:             Ptr(time.Date(2023, time.May, 17, 0, 0, 0, 0, time.UTC)),
				PayorAccountNo:      "101000000303",
				PayorName:           "D1-1-OCC ZERO",
				ReturnInstitutionID: "000101261",
				ReturnAccountNo:     "8989899",
			},
		},
		"empty": {
			in:          "",
			expectedTxn: DebitReverse{},
			expectErr: &ParseError{
				Err: ErrInvalidRecordLength,
			},
		},
		"failed to parse amount": {
			in:          "450000000404aa231370612148211010000003036121006100001000000003000BANK OF MONTREAD1-1-OCC ZERO                 BANK OF MONTREA               00000001101140               0001012618989899                                            00000000000",
			expectedTxn: DebitReverse{},
			expectErr:   &ParseError{},
		},
	}
	for name := range cases {
		tc := cases[name]

		t.Run(name, func(t *testing.T) {
			var cr DebitReverse
			err := cr.Parse(tc.in)
			if tc.expectErr != nil {
				r.ErrorAs(err, &tc.expectErr)
			} else {
				r.Equal(tc.expectedTxn, cr)
			}
		})
	}
}

func TestBuildDebitReverse(t *testing.T) {
	type testCase struct {
		in             DebitReverse
		expectedOutput string
	}
	r := require.New(t)
	date := time.Date(2023, 8, 29, 0, 0, 0, 0, time.UTC)
	cases := map[string]testCase{
		"happy path": {
			in:             NewDebitReverse("400", 999, &date, "123456789", "123456789012", "0000000000000000000000", "SHORT-NAME", "RECEIVER NAME", "LONG-NAME", "123456789", "210987654321", "040201", WithUserID("54321"), WithCrossRefNo("123"), WithSettlementCode("01"), WithInvalidDataElementID("19")),
			expectedOutput: "40000000009990232411234567891234567890120000000000000000000000000SHORT-NAME     RECEIVER NAME                 LONG-NAME                     54321     123                123456789210987654321               00000000000000000402010119000000000",
		},
		"default fields": {
			in:             DebitReverse{},
			expectedOutput: "0000000000000000000000000000            0000000000000000000000000                                                                                                        000000000                           0000000000000000000000  00000000000",
		},
		"non-ascii characters": {
			in:             NewDebitReverse("400", 999, &date, "123456789", "123456789012", "0000000000000000000000", "śhôrt-ñàmè", "réçëîvér ńámê", "LÖŃG-Ñämė", "123456789", "210987654321", "040201", WithUserID("54321"), WithCrossRefNo("123"), WithSettlementCode("01")),
			expectedOutput: "40000000009990232411234567891234567890120000000000000000000000000short-name     receiver name                 LONG-Name                     54321     123                123456789210987654321               00000000000000000402010100000000000",
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			out, err := tc.in.Build()
			r.NoError(err)
			r.Equal(tc.expectedOutput, out)
		})
	}
}

func TestValidateDebitReverse(t *testing.T) {
	r := require.New(t)
	type testCase struct {
		c       DebitReverse
		noError bool
	}
	date := time.Date(2023, 8, 29, 0, 0, 0, 0, time.UTC)
	cases := map[string]testCase{
		"happy path": {
			c: NewDebitReverse(
				"400",
				999,
				&date,
				"123456789",
				"123456789012",
				"2222222222222222222222",
				"SHORT-NAME",
				"RECEIVER NAME",
				"LONG-NAME",
				"987654321",
				"210987654321",
				"54321",
				WithStoredTransactionType("900"),
				WithCrossRefNo("123"),
				WithUserID("54321"),
			),
			noError: true,
		},
		"validate missing transaction type": {
			c: NewDebitReverse("", 0, &date, "", "", "", "", "", "", "", "", ""),
		},
		"validate invalid amount": {
			c: NewDebitReverse("450", 99999999999, &date, "", "", "", "", "", "", "", "", ""),
		},
		"validate invalid due date": {
			c: NewDebitReverse("450", 100, nil, "", "", "", "", "", "", "", "", ""),
		},
		"validate invalid institution id": {
			c: NewDebitReverse("450", 100, &date, "abc", "", "", "", "", "", "", "", ""),
		},
		"validate institution id too long": {
			c: NewDebitReverse("450", 100, &date, "1234567889", "", "", "", "", "", "", "", ""),
		},
		"validate missing institution id": {
			c: NewDebitReverse("450", 100, &date, "", "", "", "", "", "", "", "", ""),
		},
		"validate payee account number too long": {
			c: NewDebitReverse("450", 100, &date, "123456789", "123456789abchedfg", "", "", "", "", "", "", ""),
		},
		"validate missing payee account number": {
			c: NewDebitReverse("450", 100, &date, "123456789", "", "", "", "", "", "", "", ""),
		},
		"validate missing item trace number": {
			c: NewDebitReverse("450", 100, &date, "123456789", "12345", "", "", "", "", "", "", ""),
		},
		"validate invlaid item trace number": {
			c: NewDebitReverse("450", 100, &date, "123456789", "12345", "fvhdijasf", "", "", "", "", "", ""),
		},
		"validate missing short name and long name": {
			c: NewDebitReverse("450", 100, &date, "123456789", "12345", "1234", "", "", "", "", "", ""),
		},
		"validate invalid short name": {
			c: NewDebitReverse("450", 100, &date, "123456789", "12345", "1234", "THIS IS NOT A SHORT NAME AT ALL", "", "", "", "", ""),
		},
		"validate missing payee name": {
			c: NewDebitReverse("450", 100, &date, "123456789", "12345", "1234", "Short name", "", "", "", "", ""),
		},
		"validate invalid payee name": {
			c: NewDebitReverse("450", 100, &date, "123456789", "12345", "1234", "Short name", "THIS PAYEE NAME IS TOO LONG AHHHHHH", "", "", "", ""),
		},
		"validate invalid long name": {
			c: NewDebitReverse("450", 100, &date, "123456789", "12345", "1234", "Short-Name", "Receiver name", "THIS LONG NAME IS TOO LONG AHHHHH", "", "", ""),
		},
		"validate missing return institution id": {
			c: NewDebitReverse("450", 100, &date, "123456789", "12345", "1234", "Short-Name", "Receiver name", "Long-Name", "", "", ""),
		},
		"validate invalid return institution id": {
			c: NewDebitReverse("450", 100, &date, "123456789", "12345", "1234", "Short-Name", "Receiver name", "Long-Name", "abc", "", ""),
		},
		"validate invalid return account number": {
			c: NewDebitReverse("450", 100, &date, "123456789", "12345", "1234", "Short-Name", "Receiver name", "Long-Name", "54321", "123456789abchedfg", ""),
		},
		"validate malformed invalid data element id": {
			c: NewDebitReverse("450", 100, &date, "123456789", "12345", "1234", "Short-Name", "Receiver name", "Long-Name", "54321", "123456789abchedfg", "", WithInvalidDataElementID("abc123")),
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			err := tc.c.Validate()
			if tc.noError {
				r.NoError(err)
			} else {
				r.Error(err)
			}
		})
	}
}
