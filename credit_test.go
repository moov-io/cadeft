package cadeft

import (
	"testing"
	"time"

	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
)

func TestParseCreditTxn(t *testing.T) {
	type testCase struct {
		in          string
		expectedTxn Credit
		expectErr   error
	}
	r := require.New(t)
	cases := map[string]testCase{
		"regular txn": {
			in: "45000000040420231370612148211010000003036121006100001000000003000BANK OF MONTREAD1-1-OCC ZERO                 BANK OF MONTREA               00000001101140               0001012618989899                                            00000000000",
			expectedTxn: Credit{
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
					RecordType:            CreditRecord,
				},
				DateFundsAvailable:  lo.ToPtr(time.Date(2023, time.May, 17, 0, 0, 0, 0, time.UTC)),
				PayeeAccountNo:      "101000000303",
				PayeeName:           "D1-1-OCC ZERO",
				ReturnInstitutionID: "000101261",
				ReturnAccountNo:     "8989899",
			},
		},
		"empty": {
			in:          "",
			expectedTxn: Credit{},
			expectErr: &ParseError{
				Err: ErrInvalidRecordLength,
			},
		},
		"failed to parse amount": {
			in:          "450000000404aa231370612148211010000003036121006100001000000003000BANK OF MONTREAD1-1-OCC ZERO                 BANK OF MONTREA               00000001101140               0001012618989899                                            00000000000",
			expectedTxn: Credit{},
			expectErr:   &ParseError{},
		},
	}
	for name := range cases {
		tc := cases[name]

		t.Run(name, func(t *testing.T) {
			var credit Credit
			err := credit.Parse(tc.in)
			if tc.expectErr != nil {
				r.ErrorAs(err, &tc.expectErr)
			} else {
				r.Equal(tc.expectedTxn, credit)
			}

		})
	}
}

func TestBuildCreditTxn(t *testing.T) {
	type testCase struct {
		in             Credit
		expectedOutput string
	}
	r := require.New(t)
	date := time.Date(2023, 8, 29, 0, 0, 0, 0, time.UTC)
	cases := map[string]testCase{
		"happy path": {
			in:             NewCredit("400", 999, &date, "123456789", "123456789012", "", "SHORT-NAME", "RECEIVER NAME", "LONG-NAME", "123456789", "210987654321", WithUserID("54321"), WithCrossRefNo("123"), WithSettlementCode("01")),
			expectedOutput: "40000000009990232411234567891234567890120000000000000000000000000SHORT-NAME     RECEIVER NAME                 LONG-NAME                     54321     123                123456789210987654321                                     0100000000000",
		},
		"empty": {
			in:             Credit{},
			expectedOutput: "0000000000000000000000000000            0000000000000000000000000                                                                                                        000000000                                                   00000000000",
		},
		"non ascii characters": {
			in:             NewCredit("400", 999, &date, "123456789", "123456789012", "", "śhôrt-ñàmè", "réçëîvér ńámê", "LÖŃG-Ñämė", "123456789", "210987654321", WithUserID("54321"), WithCrossRefNo("123"), WithSettlementCode("01")),
			expectedOutput: "40000000009990232411234567891234567890120000000000000000000000000short-name     receiver name                 LONG-Name                     54321     123                123456789210987654321                                     0100000000000",
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			out, err := tc.in.Build()
			t.Logf("%+v", out)
			r.NoError(err)
			r.Equal(tc.expectedOutput, out)
		})
	}
}

func TestValidateCreditTxn(t *testing.T) {
	r := require.New(t)
	type testCase struct {
		c       Credit
		noError bool
	}
	date := time.Date(2023, 8, 29, 0, 0, 0, 0, time.UTC)
	cases := map[string]testCase{
		"happy path": {
			c: NewCredit(
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
			c: NewCredit("", 0, &date, "", "", "", "", "", "", "", ""),
		},
		"validate invalid amount": {
			c: NewCredit("450", 99999999999, &date, "", "", "", "", "", "", "", ""),
		},
		"validate invalid date funds available": {
			c: NewCredit("450", 100, nil, "", "", "", "", "", "", "", ""),
		},
		"validate invalid institution id": {
			c: NewCredit("450", 100, &date, "abc", "", "", "", "", "", "", ""),
		},
		"validate institution id too long": {
			c: NewCredit("450", 100, &date, "1234567889", "", "", "", "", "", "", ""),
		},
		"validate missing institution id": {
			c: NewCredit("450", 100, &date, "", "", "", "", "", "", "", ""),
		},
		"validate payee account number too long": {
			c: NewCredit("450", 100, &date, "123456789", "123456789abchedfg", "", "", "", "", "", ""),
		},
		"validate missing payor account number": {
			c: NewCredit("450", 100, &date, "123456789", "", "", "", "", "", "", ""),
		},
		"validate missing item trace number": {
			c: NewCredit("450", 100, &date, "123456789", "12345", "", "", "", "", "", ""),
		},
		"validate invlaid item trace number": {
			c: NewCredit("450", 100, &date, "123456789", "12345", "fvhdijasf", "", "", "", "", ""),
		},
		"validate missing short name": {
			c: NewCredit("450", 100, &date, "123456789", "12345", "1234", "", "", "", "", ""),
		},
		"validate invalid short name": {
			c: NewCredit("450", 100, &date, "123456789", "12345", "1234", "THIS IS NOT A SHORT NAME AT ALL", "", "", "", ""),
		},
		"validate missing payee name": {
			c: NewCredit("450", 100, &date, "123456789", "12345", "1234", "Short name", "", "", "", ""),
		},
		"validate invalid payee name": {
			c: NewCredit("450", 100, &date, "123456789", "12345", "1234", "Short name", "THIS PAYOR NAME IS TOO LONG AHHHHHH", "", "", ""),
		},
		"validate missing long name": {
			c: NewCredit("450", 100, &date, "123456789", "12345", "1234", "Short-Name", "Receiver name", "", "", ""),
		},
		"validate invalid long name": {
			c: NewCredit("450", 100, &date, "123456789", "12345", "1234", "Short-Name", "Receiver name", "THIS LONG NAME IS TOO LONG AHHHHH", "", ""),
		},
		"validate missing return institution id": {
			c: NewCredit("450", 100, &date, "123456789", "12345", "1234", "Short-Name", "Receiver name", "Long-Name", "", ""),
		},
		"validate invalid return institution id": {
			c: NewCredit("450", 100, &date, "123456789", "12345", "1234", "Short-Name", "Receiver name", "Long-Name", "abc", ""),
		},
		"validate invalid return account number": {
			c: NewCredit("450", 100, &date, "123456789", "12345", "1234", "Short-Name", "Receiver name", "Long-Name", "54321", "123456789abchedfg"),
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			err := tc.c.Validate()
			if tc.noError {
				r.NoError(err)
			} else {
				r.NotNil(err)
			}
		})
	}
}
