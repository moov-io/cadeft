package cadeft

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCreateFile(t *testing.T) {
	r := require.New(t)
	type testCase struct {
		file           File
		expectedOutput string
	}
	date := time.Date(2023, 10, 2, 0, 0, 0, 0, time.UTC)
	header := NewFileHeader("0000000001", 1, &date, 12345, "CAD")
	var debitTxns []Transaction
	debitTxns = append(debitTxns,
		Ptr(NewDebit("400", 1000, &date, "987654321", "1234", "12345", "Short name", "payor name", "my long name", "123456789", "1111111")),
		Ptr(NewDebit("400", 1000, &date, "987654321", "1234", "12345", "Short name", "payor name", "my long name", "123456789", "1111111")),
		Ptr(NewDebit("400", 1000, &date, "987654321", "1234", "12345", "Short name", "payor name", "my long name", "123456789", "1111111")),
		Ptr(NewDebit("400", 1000, &date, "987654321", "1234", "12345", "Short name", "payor name", "my long name", "123456789", "1111111")),
		Ptr(NewDebit("400", 1000, &date, "987654321", "1234", "12345", "Short name", "payor name", "my long name", "123456789", "1111111")),
		Ptr(NewDebit("400", 1000, &date, "987654321", "1234", "12345", "Short name", "payor name", "my long name", "123456789", "1111111")),
		Ptr(NewDebit("400", 1000, &date, "987654321", "1234", "12345", "Short name", "payor name", "my long name", "123456789", "1111111")))
	debitFile := NewFile(header, debitTxns)

	var creditTxns []Transaction
	creditTxns = append(creditTxns,
		Ptr(NewCredit("450", 1000, &date, "123456789", "12345", "12313213", "short name", "payee name", "someone", "1231", "12345")),
		Ptr(NewCredit("450", 1000, &date, "123456789", "12345", "12313213", "short name", "payee name", "someone", "1231", "12345")),
		Ptr(NewCredit("450", 1000, &date, "123456789", "12345", "12313213", "short name", "payee name", "someone", "1231", "12345")),
		Ptr(NewCredit("450", 1000, &date, "123456789", "12345", "12313213", "short name", "payee name", "someone", "1231", "12345")),
		Ptr(NewCredit("450", 1000, &date, "123456789", "12345", "12313213", "short name", "payee name", "someone", "1231", "12345")),
		Ptr(NewCredit("450", 1000, &date, "123456789", "12345", "12313213", "short name", "payee name", "someone", "1231", "12345")),
		Ptr(NewCredit("450", 1000, &date, "123456789", "12345", "12313213", "short name", "payee name", "someone", "1231", "12345")),
	)
	creditFile := NewFile(header, creditTxns)

	var creditReturns []Transaction
	creditReturns = append(creditReturns,
		Ptr(NewCreditReturn("900", 1000, &date, "12345", "1234567", "2222222222222222222222", "sender", "receiver", "sender full name", "54321", "7654321", "3333333333333333333333")),
		Ptr(NewCreditReturn("900", 1000, &date, "12345", "1234567", "2222222222222222222222", "sender", "receiver", "sender full name", "54321", "7654321", "3333333333333333333333")),
		Ptr(NewCreditReturn("900", 1000, &date, "12345", "1234567", "2222222222222222222222", "sender", "receiver", "sender full name", "54321", "7654321", "3333333333333333333333")),
		Ptr(NewCreditReturn("900", 1000, &date, "12345", "1234567", "2222222222222222222222", "sender", "receiver", "sender full name", "54321", "7654321", "3333333333333333333333")),
		Ptr(NewCreditReturn("900", 1000, &date, "12345", "1234567", "2222222222222222222222", "sender", "receiver", "sender full name", "54321", "7654321", "3333333333333333333333")),
		Ptr(NewCreditReturn("900", 1000, &date, "12345", "1234567", "2222222222222222222222", "sender", "receiver", "sender full name", "54321", "7654321", "3333333333333333333333")),
		Ptr(NewCreditReturn("900", 1000, &date, "12345", "1234567", "2222222222222222222222", "sender", "receiver", "sender full name", "54321", "7654321", "3333333333333333333333")),
		Ptr(NewCreditReturn("900", 1000, &date, "12345", "1234567", "2222222222222222222222", "sender", "receiver", "sender full name", "54321", "7654321", "3333333333333333333333")),
	)
	creditReturnsFile := NewFile(header, creditReturns)
	var debitReturns []Transaction
	debitReturns = append(debitReturns,
		Ptr(NewDebitReturn("900", 1000, &date, "12345", "123", "2222222222222222222222", "someone", "payor", "some long name", "54321", "123211", "4444444444444444444444")),
		Ptr(NewDebitReturn("900", 1000, &date, "12345", "123", "2222222222222222222222", "someone", "payor", "some long name", "54321", "123211", "4444444444444444444444")),
		Ptr(NewDebitReturn("900", 1000, &date, "12345", "123", "2222222222222222222222", "someone", "payor", "some long name", "54321", "123211", "4444444444444444444444")),
		Ptr(NewDebitReturn("900", 1000, &date, "12345", "123", "2222222222222222222222", "someone", "payor", "some long name", "54321", "123211", "4444444444444444444444")),
		Ptr(NewDebitReturn("900", 1000, &date, "12345", "123", "2222222222222222222222", "someone", "payor", "some long name", "54321", "123211", "4444444444444444444444")),
		Ptr(NewDebitReturn("900", 1000, &date, "12345", "123", "2222222222222222222222", "someone", "payor", "some long name", "54321", "123211", "4444444444444444444444")),
		Ptr(NewDebitReturn("900", 1000, &date, "12345", "123", "2222222222222222222222", "someone", "payor", "some long name", "54321", "123211", "4444444444444444444444")),
	)
	debitReturnsFile := NewFile(header, debitReturns)

	var creditReversals []Transaction
	creditReversals = append(creditReversals,
		Ptr(NewCreditReverse("900", 1000, &date, "12345", "1234567", "2222222222222222222222", "sender", "receiver", "sender full name", "54321", "7654321", "3333333333333333333333")),
		Ptr(NewCreditReverse("900", 1000, &date, "12345", "1234567", "2222222222222222222222", "sender", "receiver", "sender full name", "54321", "7654321", "3333333333333333333333")),
		Ptr(NewCreditReverse("900", 1000, &date, "12345", "1234567", "2222222222222222222222", "sender", "receiver", "sender full name", "54321", "7654321", "3333333333333333333333")),
		Ptr(NewCreditReverse("900", 1000, &date, "12345", "1234567", "2222222222222222222222", "sender", "receiver", "sender full name", "54321", "7654321", "3333333333333333333333")),
		Ptr(NewCreditReverse("900", 1000, &date, "12345", "1234567", "2222222222222222222222", "sender", "receiver", "sender full name", "54321", "7654321", "3333333333333333333333")),
		Ptr(NewCreditReverse("900", 1000, &date, "12345", "1234567", "2222222222222222222222", "sender", "receiver", "sender full name", "54321", "7654321", "3333333333333333333333")),
		Ptr(NewCreditReverse("900", 1000, &date, "12345", "1234567", "2222222222222222222222", "sender", "receiver", "sender full name", "54321", "7654321", "3333333333333333333333")),
	)
	creditReversalsFile := NewFile(header, creditReversals)

	var debitReversals []Transaction
	debitReversals = append(debitReversals,
		Ptr(NewDebitReverse("900", 1000, &date, "12345", "123", "2222222222222222222222", "someone", "payor", "some long name", "54321", "123211", "4444444444444444444444")),
		Ptr(NewDebitReverse("900", 1000, &date, "12345", "123", "2222222222222222222222", "someone", "payor", "some long name", "54321", "123211", "4444444444444444444444")),
		Ptr(NewDebitReverse("900", 1000, &date, "12345", "123", "2222222222222222222222", "someone", "payor", "some long name", "54321", "123211", "4444444444444444444444")),
		Ptr(NewDebitReverse("900", 1000, &date, "12345", "123", "2222222222222222222222", "someone", "payor", "some long name", "54321", "123211", "4444444444444444444444")),
		Ptr(NewDebitReverse("900", 1000, &date, "12345", "123", "2222222222222222222222", "someone", "payor", "some long name", "54321", "123211", "4444444444444444444444")),
		Ptr(NewDebitReverse("900", 1000, &date, "12345", "123", "2222222222222222222222", "someone", "payor", "some long name", "54321", "123211", "4444444444444444444444")),
		Ptr(NewDebitReverse("900", 1000, &date, "12345", "123", "2222222222222222222222", "someone", "payor", "some long name", "54321", "123211", "4444444444444444444444")),
	)
	debitReversalFile := NewFile(header, debitReversals)

	cases := map[string]testCase{
		"Debits only file": {
			file: debitFile,
			expectedOutput: fmt.Sprintf(`A0000000010000000001000102327512345                    CAD%s
D0000000020000000001000140000000010000232759876543211234        0000000000000000012345000Short name     payor name                    my long name                                               1234567891111111                                            0000000000040000000010000232759876543211234        0000000000000000012345000Short name     payor name                    my long name                                               1234567891111111                                            0000000000040000000010000232759876543211234        0000000000000000012345000Short name     payor name                    my long name                                               1234567891111111                                            0000000000040000000010000232759876543211234        0000000000000000012345000Short name     payor name                    my long name                                               1234567891111111                                            0000000000040000000010000232759876543211234        0000000000000000012345000Short name     payor name                    my long name                                               1234567891111111                                            0000000000040000000010000232759876543211234        0000000000000000012345000Short name     payor name                    my long name                                               1234567891111111                                            00000000000
D0000000030000000001000140000000010000232759876543211234        0000000000000000012345000Short name     payor name                    my long name                                               1234567891111111                                            00000000000                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                
Z000000004000000000100010000000000700000000007000000000000000000000000000000000000000000000000000000000000000000%s`, strings.Repeat(" ", 1406), strings.Repeat(" ", 1352)),
		},
		"credits only file": {
			file: creditFile,
			expectedOutput: fmt.Sprintf(`A0000000010000000001000102327512345                    CAD%s
C00000000200000000010001450000000100002327512345678912345       0000000000000012313213000short name     payee name                    someone                                                    00000123112345                                              00000000000450000000100002327512345678912345       0000000000000012313213000short name     payee name                    someone                                                    00000123112345                                              00000000000450000000100002327512345678912345       0000000000000012313213000short name     payee name                    someone                                                    00000123112345                                              00000000000450000000100002327512345678912345       0000000000000012313213000short name     payee name                    someone                                                    00000123112345                                              00000000000450000000100002327512345678912345       0000000000000012313213000short name     payee name                    someone                                                    00000123112345                                              00000000000450000000100002327512345678912345       0000000000000012313213000short name     payee name                    someone                                                    00000123112345                                              00000000000
C00000000300000000010001450000000100002327512345678912345       0000000000000012313213000short name     payee name                    someone                                                    00000123112345                                              00000000000                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                
Z000000004000000000100010000000000000000000000000000000070000000000700000000000000000000000000000000000000000000%s`, strings.Repeat(" ", 1406), strings.Repeat(" ", 1352)),
		},
		"credit returns only file": {
			file: creditReturnsFile,
			expectedOutput: fmt.Sprintf(`A0000000010000000001000102327512345                    CAD%s
I0000000020000000001000190000000010000232750000123451234567     2222222222222222222222000sender         receiver                      sender full name                                           0000543217654321                    3333333333333333333333  0000000000090000000010000232750000123451234567     2222222222222222222222000sender         receiver                      sender full name                                           0000543217654321                    3333333333333333333333  0000000000090000000010000232750000123451234567     2222222222222222222222000sender         receiver                      sender full name                                           0000543217654321                    3333333333333333333333  0000000000090000000010000232750000123451234567     2222222222222222222222000sender         receiver                      sender full name                                           0000543217654321                    3333333333333333333333  0000000000090000000010000232750000123451234567     2222222222222222222222000sender         receiver                      sender full name                                           0000543217654321                    3333333333333333333333  0000000000090000000010000232750000123451234567     2222222222222222222222000sender         receiver                      sender full name                                           0000543217654321                    3333333333333333333333  00000000000
I0000000030000000001000190000000010000232750000123451234567     2222222222222222222222000sender         receiver                      sender full name                                           0000543217654321                    3333333333333333333333  0000000000090000000010000232750000123451234567     2222222222222222222222000sender         receiver                      sender full name                                           0000543217654321                    3333333333333333333333  00000000000                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                
Z000000004000000000100010000000000000000000000000000000080000000000800000000000000000000000000000000000000000000%s`, strings.Repeat(" ", 1406), strings.Repeat(" ", 1352)),
		},
		"debit returns only file": {
			file: debitReturnsFile,
			expectedOutput: fmt.Sprintf(`A0000000010000000001000102327512345                    CAD%s
J000000002000000000100019000000001000023275000012345123         2222222222222222222222000someone        payor                         some long name                                             000054321123211                     4444444444444444444444  000000000009000000001000023275000012345123         2222222222222222222222000someone        payor                         some long name                                             000054321123211                     4444444444444444444444  000000000009000000001000023275000012345123         2222222222222222222222000someone        payor                         some long name                                             000054321123211                     4444444444444444444444  000000000009000000001000023275000012345123         2222222222222222222222000someone        payor                         some long name                                             000054321123211                     4444444444444444444444  000000000009000000001000023275000012345123         2222222222222222222222000someone        payor                         some long name                                             000054321123211                     4444444444444444444444  000000000009000000001000023275000012345123         2222222222222222222222000someone        payor                         some long name                                             000054321123211                     4444444444444444444444  00000000000
J000000003000000000100019000000001000023275000012345123         2222222222222222222222000someone        payor                         some long name                                             000054321123211                     4444444444444444444444  00000000000                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                
Z000000004000000000100010000000000700000000007000000000000000000000000000000000000000000000000000000000000000000%s`, strings.Repeat(" ", 1406), strings.Repeat(" ", 1352)),
		},
		"credit reversals file": {
			file: creditReversalsFile,
			expectedOutput: fmt.Sprintf(`A0000000010000000001000102327512345                    CAD%s
E0000000020000000001000190000000010000232750000123451234567     2222222222222222222222000sender         receiver                      sender full name                                           0000543217654321                    3333333333333333333333  0000000000090000000010000232750000123451234567     2222222222222222222222000sender         receiver                      sender full name                                           0000543217654321                    3333333333333333333333  0000000000090000000010000232750000123451234567     2222222222222222222222000sender         receiver                      sender full name                                           0000543217654321                    3333333333333333333333  0000000000090000000010000232750000123451234567     2222222222222222222222000sender         receiver                      sender full name                                           0000543217654321                    3333333333333333333333  0000000000090000000010000232750000123451234567     2222222222222222222222000sender         receiver                      sender full name                                           0000543217654321                    3333333333333333333333  0000000000090000000010000232750000123451234567     2222222222222222222222000sender         receiver                      sender full name                                           0000543217654321                    3333333333333333333333  00000000000
E0000000030000000001000190000000010000232750000123451234567     2222222222222222222222000sender         receiver                      sender full name                                           0000543217654321                    3333333333333333333333  00000000000                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                
Z000000004000000000100010000000000000000000000000000000000000000000000000000007000000000070000000000000000000000%s`, strings.Repeat(" ", 1406), strings.Repeat(" ", 1352)),
		},
		"debit reversals file": {
			file: debitReversalFile,
			expectedOutput: fmt.Sprintf(`A0000000010000000001000102327512345                    CAD%s
F000000002000000000100019000000001000023275000012345123         2222222222222222222222000someone        payor                         some long name                                             000054321123211                     4444444444444444444444  000000000009000000001000023275000012345123         2222222222222222222222000someone        payor                         some long name                                             000054321123211                     4444444444444444444444  000000000009000000001000023275000012345123         2222222222222222222222000someone        payor                         some long name                                             000054321123211                     4444444444444444444444  000000000009000000001000023275000012345123         2222222222222222222222000someone        payor                         some long name                                             000054321123211                     4444444444444444444444  000000000009000000001000023275000012345123         2222222222222222222222000someone        payor                         some long name                                             000054321123211                     4444444444444444444444  000000000009000000001000023275000012345123         2222222222222222222222000someone        payor                         some long name                                             000054321123211                     4444444444444444444444  00000000000
F000000003000000000100019000000001000023275000012345123         2222222222222222222222000someone        payor                         some long name                                             000054321123211                     4444444444444444444444  00000000000                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                
Z000000004000000000100010000000000000000000000000000000000000000000000000000000000000000000000000000700000000007%s`, strings.Repeat(" ", 1406), strings.Repeat(" ", 1352)),
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			s, err := tc.file.Create()
			r.NoError(err)
			r.NoError(tc.file.Validate())
			r.Equal(tc.expectedOutput, s)
		})
	}
}

func TestReadFile(t *testing.T) {
	r := require.New(t)
	type testCase struct {
		in   string
		file File
	}
	date := time.Date(2023, 10, 2, 0, 0, 0, 0, time.UTC)
	header := NewFileHeader("0000000001", 1, &date, 12345, "CAD")
	rawDebitFile := `A0000000010000000001000102327512345                    CAD
D0000000020000000001000140000000010000232759876543211234        0000000000000000012345000Short name     payor name                    my long name                                               1234567891111111                                            0000000000040000000010000232759876543211234        0000000000000000012345000Short name     payor name                    my long name                                               1234567891111111                                            0000000000040000000010000232759876543211234        0000000000000000012345000Short name     payor name                    my long name                                               1234567891111111                                            0000000000040000000010000232759876543211234        0000000000000000012345000Short name     payor name                    my long name                                               1234567891111111                                            0000000000040000000010000232759876543211234        0000000000000000012345000Short name     payor name                    my long name                                               1234567891111111                                            0000000000040000000010000232759876543211234        0000000000000000012345000Short name     payor name                    my long name                                               1234567891111111                                            00000000000
D0000000030000000001000140000000010000232759876543211234        0000000000000000012345000Short name     payor name                    my long name                                               1234567891111111                                            00000000000                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                
Z000000004000000000100010000000000700000000007000000000000000000000000000000000000000000000000000000000000000000`
	var debitTxns []Transaction
	debitTxns = append(debitTxns,
		Ptr(NewDebit("400", 1000, &date, "987654321", "1234", "0000000000000000012345", "Short name", "payor name", "my long name", "123456789", "1111111")),
		Ptr(NewDebit("400", 1000, &date, "987654321", "1234", "0000000000000000012345", "Short name", "payor name", "my long name", "123456789", "1111111")),
		Ptr(NewDebit("400", 1000, &date, "987654321", "1234", "0000000000000000012345", "Short name", "payor name", "my long name", "123456789", "1111111")),
		Ptr(NewDebit("400", 1000, &date, "987654321", "1234", "0000000000000000012345", "Short name", "payor name", "my long name", "123456789", "1111111")),
		Ptr(NewDebit("400", 1000, &date, "987654321", "1234", "0000000000000000012345", "Short name", "payor name", "my long name", "123456789", "1111111")),
		Ptr(NewDebit("400", 1000, &date, "987654321", "1234", "0000000000000000012345", "Short name", "payor name", "my long name", "123456789", "1111111")),
		Ptr(NewDebit("400", 1000, &date, "987654321", "1234", "0000000000000000012345", "Short name", "payor name", "my long name", "123456789", "1111111")))
	debitFile := NewFile(header, debitTxns)
	debitFooter := FileFooter{
		RecordHeader: RecordHeader{
			RecordType:      FooterRecord,
			OriginatorID:    "0000000001",
			FileCreationNum: 1,
			recordCount:     4,
		},
		TotalValueOfDebit: 7000,
		TotalCountOfDebit: 7,
	}
	debitFile.Footer = &debitFooter

	rawCreditsFile := `A0000000010000000001000102327512345                    CAD
C00000000200000000010001450000000100002327512345678912345       0000000000000012313213000short name     payee name                    someone                                                    00000123112345                                              00000000000450000000100002327512345678912345       0000000000000012313213000short name     payee name                    someone                                                    00000123112345                                              00000000000450000000100002327512345678912345       0000000000000012313213000short name     payee name                    someone                                                    00000123112345                                              00000000000450000000100002327512345678912345       0000000000000012313213000short name     payee name                    someone                                                    00000123112345                                              00000000000450000000100002327512345678912345       0000000000000012313213000short name     payee name                    someone                                                    00000123112345                                              00000000000450000000100002327512345678912345       0000000000000012313213000short name     payee name                    someone                                                    00000123112345                                              00000000000
C00000000300000000010001450000000100002327512345678912345       0000000000000012313213000short name     payee name                    someone                                                    00000123112345                                              00000000000                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                
Z000000004000000000100010000000000000000000000000000000070000000000700000000000000000000000000000000000000000000`
	var creditTxns []Transaction
	creditTxns = append(creditTxns,
		Ptr(NewCredit("450", 1000, &date, "123456789", "12345", "0000000000000012313213", "short name", "payee name", "someone", "000001231", "12345")),
		Ptr(NewCredit("450", 1000, &date, "123456789", "12345", "0000000000000012313213", "short name", "payee name", "someone", "000001231", "12345")),
		Ptr(NewCredit("450", 1000, &date, "123456789", "12345", "0000000000000012313213", "short name", "payee name", "someone", "000001231", "12345")),
		Ptr(NewCredit("450", 1000, &date, "123456789", "12345", "0000000000000012313213", "short name", "payee name", "someone", "000001231", "12345")),
		Ptr(NewCredit("450", 1000, &date, "123456789", "12345", "0000000000000012313213", "short name", "payee name", "someone", "000001231", "12345")),
		Ptr(NewCredit("450", 1000, &date, "123456789", "12345", "0000000000000012313213", "short name", "payee name", "someone", "000001231", "12345")),
		Ptr(NewCredit("450", 1000, &date, "123456789", "12345", "0000000000000012313213", "short name", "payee name", "someone", "000001231", "12345")),
	)
	creditFile := NewFile(header, creditTxns)
	creditFooter := FileFooter{
		RecordHeader: RecordHeader{
			RecordType:      FooterRecord,
			OriginatorID:    "0000000001",
			FileCreationNum: 1,
			recordCount:     4,
		},
		TotalValueOfCredit: 7000,
		TotalCountOfCredit: 7,
	}
	creditFile.Footer = &creditFooter

	rawCreditAndDebitFile := `A0000000010000000001000102327512345                    CAD
C00000000200000000010001450000000100002327512345678912345       0000000000000012313213000short name     payee name                    someone                                                    00000123112345                                              00000000000450000000100002327512345678912345       0000000000000012313213000short name     payee name                    someone                                                    00000123112345                                              00000000000450000000100002327512345678912345       0000000000000012313213000short name     payee name                    someone                                                    00000123112345                                              00000000000450000000100002327512345678912345       0000000000000012313213000short name     payee name                    someone                                                    00000123112345                                              00000000000450000000100002327512345678912345       0000000000000012313213000short name     payee name                    someone                                                    00000123112345                                              00000000000450000000100002327512345678912345       0000000000000012313213000short name     payee name                    someone                                                    00000123112345                                              00000000000
C00000000300000000010001450000000100002327512345678912345       0000000000000012313213000short name     payee name                    someone                                                    00000123112345                                              00000000000                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                
D0000000040000000001000140000000010000232759876543211234        0000000000000000012345000Short name     payor name                    my long name                                               1234567891111111                                            0000000000040000000010000232759876543211234        0000000000000000012345000Short name     payor name                    my long name                                               1234567891111111                                            0000000000040000000010000232759876543211234        0000000000000000012345000Short name     payor name                    my long name                                               1234567891111111                                            0000000000040000000010000232759876543211234        0000000000000000012345000Short name     payor name                    my long name                                               1234567891111111                                            0000000000040000000010000232759876543211234        0000000000000000012345000Short name     payor name                    my long name                                               1234567891111111                                            0000000000040000000010000232759876543211234        0000000000000000012345000Short name     payor name                    my long name                                               1234567891111111                                            00000000000
D0000000050000000001000140000000010000232759876543211234        0000000000000000012345000Short name     payor name                    my long name                                               1234567891111111                                            00000000000                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                
Z000000006000000000100010000000001400000000007000000000140000000000700000000000000000000000000000000000000000000`
	creditAndDebit := append(creditTxns, debitTxns...)
	creditsAndDebitFile := NewFile(header, creditAndDebit)
	creditsAndDebitsFooter := FileFooter{
		RecordHeader: RecordHeader{
			RecordType:      FooterRecord,
			OriginatorID:    "0000000001",
			FileCreationNum: 1,
			recordCount:     6,
		},
		TotalValueOfDebit:  14000,
		TotalCountOfDebit:  7,
		TotalValueOfCredit: 14000,
		TotalCountOfCredit: 7,
	}
	creditsAndDebitFile.Footer = &creditsAndDebitsFooter

	cases := map[string]testCase{
		"debit file": {
			in:   rawDebitFile,
			file: debitFile,
		},
		"credit file": {
			in:   rawCreditsFile,
			file: creditFile,
		},
		"credits and debits file": {
			in:   rawCreditAndDebitFile,
			file: creditsAndDebitFile,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			freader := NewReader(strings.NewReader(tc.in))
			f, err := freader.ReadFile()
			r.NoError(err)
			r.NoError(f.Validate())
			r.Equal(tc.file.Header, f.Header)
			r.Equal(tc.file.Footer, f.Footer)
			for i := 0; i < len(tc.file.Txns); i++ {
				r.Equal(tc.file.Txns[i], f.Txns[i])
			}
		})
	}
}
