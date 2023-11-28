package cadeft

import (
	"testing"
	"time"

	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
)

func TestParseDebitTxn(t *testing.T) {
	type testCase struct {
		in          string
		expectedTxn Debit
		expectErr   error
	}
	r := require.New(t)
	cases := map[string]testCase{
		"regular txn": {
			in: "45000000040420231370612148211010000003036121006100001000000003000BANK OF MONTREAD1-1-OCC ZERO                 BANK OF MONTREA               00000001101140               0001012618989899                                            00000000000",
			expectedTxn: Debit{
				BaseTxn: BaseTxn{
					TxnType:               TransactionType("450"),
					Amount:                int64(4042),
					ItemTraceNo:           "6121006100001000000003",
					InstitutionID:         "061214821",
					StoredTransactionType: "000",
					OriginatorShortName:   "BANK OF MONTREA",
					OriginatorLongName:    "BANK OF MONTREA",
					UserID:                "0000000110",
					CrossRefNo:            "1140",
					SundryInfo:            "",
					SettlementCode:        "",
					RecordType:            DebitRecord,
				},
				DueDate:             lo.ToPtr(time.Date(2023, time.May, 17, 0, 0, 0, 0, time.UTC)),
				PayorAccountNo:      "101000000303",
				PayorName:           "D1-1-OCC ZERO",
				ReturnInstitutionID: "000101261",
				ReturnAccountNo:     "8989899",
			},
		},
		"empty": {
			in:          "",
			expectedTxn: Debit{},
			expectErr:   &ParseError{},
		},
		"parse error invlaid amount": {
			in:          "45000000040aa0231370612148211010000003036121006100001000000003000BANK OF MONTREAD1-1-OCC ZERO                 BANK OF MONTREA               00000001101140               0001012618989899                                            00000000000",
			expectedTxn: Debit{},
			expectErr:   &ParseError{},
		},
	}
	for name := range cases {
		tc := cases[name]

		t.Run(name, func(t *testing.T) {
			var debit Debit
			err := debit.Parse(tc.in)
			if tc.expectErr != nil {
				r.ErrorAs(err, &tc.expectErr)
				t.Logf("%v", tc.expectErr)
			} else {
				r.Equal(tc.expectedTxn, debit)
			}
		})
	}
}

func TestBuildDebitTxn(t *testing.T) {
	type testCase struct {
		in             Debit
		expectedOutput string
	}
	r := require.New(t)
	date := time.Date(2023, 8, 29, 0, 0, 0, 0, time.UTC)
	cases := map[string]testCase{
		"happy path": {
			in:             NewDebit("400", 999, &date, "123456789", "123456789012", "", "SHORT-NAME", "RECEIVER NAME", "LONG-NAME", "123456789", "210987654321", WithUserID("54321"), WithCrossRefNo("123"), WithSettlementCode("01")),
			expectedOutput: "40000000009990232411234567891234567890120000000000000000000000000SHORT-NAME     RECEIVER NAME                 LONG-NAME                     54321     123                123456789210987654321                                     0100000000000",
		},
		"empty debit txn": {
			in:             Debit{},
			expectedOutput: "0000000000000000000000000000            0000000000000000000000000                                                                                                        000000000                                                   00000000000",
		},
		"non ascii characters": {
			in:             NewDebit("400", 999, &date, "123456789", "123456789012", "", "śhôrt-ñàmè", "réçëîvér ńámê", "LÖŃG-Ñämė", "123456789", "210987654321", WithUserID("54321"), WithCrossRefNo("123"), WithSettlementCode("01")),
			expectedOutput: "40000000009990232411234567891234567890120000000000000000000000000short-name     receiver name                 LONG-Name                     54321     123                123456789210987654321                                     0100000000000",
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

func TestValidateDebit(t *testing.T) {
	r := require.New(t)
	type testCase struct {
		d       Debit
		noError bool
	}
	date := time.Date(2023, 8, 29, 0, 0, 0, 0, time.UTC)
	cases := map[string]testCase{
		"happy path": {
			d: NewDebit(
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
				WithStoredTransactionType("900"),
				WithCrossRefNo("123"),
				WithUserID("54321"),
			),
			noError: true,
		},
		"validate missing transaction type": {
			d: NewDebit("", 0, &date, "", "", "", "", "", "", "", ""),
		},
		"validate invalid amount": {
			d: NewDebit("450", 99999999999, &date, "", "", "", "", "", "", "", ""),
		},
		"validate invalid due date": {
			d: NewDebit("450", 100, nil, "", "", "", "", "", "", "", ""),
		},
		"validate invalid institution id": {
			d: NewDebit("450", 100, &date, "abc", "", "", "", "", "", "", ""),
		},
		"validate institution id too long": {
			d: NewDebit("450", 100, &date, "1234567889", "", "", "", "", "", "", ""),
		},
		"validate missing institution id": {
			d: NewDebit("450", 100, &date, "", "", "", "", "", "", "", ""),
		},
		"validate payor account number too long": {
			d: NewDebit("450", 100, &date, "123456789", "123456789abchedfg", "", "", "", "", "", ""),
		},
		"validate missing payor account number": {
			d: NewDebit("450", 100, &date, "123456789", "", "", "", "", "", "", ""),
		},
		"validate missing item trace number": {
			d: NewDebit("450", 100, &date, "123456789", "12345", "", "", "", "", "", ""),
		},
		"validate invlaid item trace number": {
			d: NewDebit("450", 100, &date, "123456789", "12345", "fvhdijasf", "", "", "", "", ""),
		},
		"validate missing short name": {
			d: NewDebit("450", 100, &date, "123456789", "12345", "1234", "", "", "", "", ""),
		},
		"validate invalid short name": {
			d: NewDebit("450", 100, &date, "123456789", "12345", "1234", "THIS IS NOT A SHORT NAME AT ALL", "", "", "", ""),
		},
		"validate missing payor name": {
			d: NewDebit("450", 100, &date, "123456789", "12345", "1234", "Short name", "", "", "", ""),
		},
		"validate invalid payor name": {
			d: NewDebit("450", 100, &date, "123456789", "12345", "1234", "Short name", "THIS PAYOR NAME IS TOO LONG AHHHHHH", "", "", ""),
		},
		"validate missing long name": {
			d: NewDebit("450", 100, &date, "123456789", "12345", "1234", "Short-Name", "Receiver name", "", "", ""),
		},
		"validate invalid long name": {
			d: NewDebit("450", 100, &date, "123456789", "12345", "1234", "Short-Name", "Receiver name", "THIS LONG NAME IS TOO LONG AHHHHH", "", ""),
		},
		"validate missing return institution id": {
			d: NewDebit("450", 100, &date, "123456789", "12345", "1234", "Short-Name", "Receiver name", "Long-Name", "", ""),
		},
		"validate invalid return institution id": {
			d: NewDebit("450", 100, &date, "123456789", "12345", "1234", "Short-Name", "Receiver name", "Long-Name", "abc", ""),
		},
		"validate invalid return account number": {
			d: NewDebit("450", 100, &date, "123456789", "12345", "1234", "Short-Name", "Receiver name", "Long-Name", "54321", "123456789abchedfg"),
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			err := tc.d.Validate()
			if tc.noError {
				r.NoError(err)
			} else {
				r.Error(err)
			}
		})
	}
}
