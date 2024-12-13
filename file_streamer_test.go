package cadeft

import (
	"io"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestStreamGetHeader(t *testing.T) {
	type testCase struct {
		file           string
		expectedHeader FileHeader
		expectErr      bool
	}
	r := require.New(t)

	cases := map[string]testCase{
		"happy path": {
			file: "A0000000010000000610000102313861210hello               CAD",
			expectedHeader: FileHeader{
				RecordHeader: RecordHeader{
					RecordType:      HeaderRecord,
					recordCount:     1,
					OriginatorID:    "0000000610",
					FileCreationNum: 1,
				},
				CreationDate:                   Ptr(time.Date(2023, time.May, 18, 0, 0, 0, 0, time.UTC)),
				DestinationDataCenterNo:        61210,
				DirectClearerCommunicationArea: "hello",
				CurrencyCode:                   "CAD",
			},
		},
		"invalid record type": {
			file:      "C0000000010000000610000102313861210hello               CAD",
			expectErr: true,
		},
		"partial file": {
			file:      "abc",
			expectErr: true,
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			file := strings.NewReader(tc.file)
			stream := NewFileStream(file)
			header, err := stream.GetHeader()
			if !tc.expectErr {
				r.NoError(err)
				r.NotNil(header)
				r.Equal(tc.expectedHeader, *header)
			} else {
				r.Error(err)
			}

		})
	}
}

func TestStreamGetFooter(t *testing.T) {
	r := require.New(t)
	type testCase struct {
		file           string
		expectedFooter FileFooter
		expectErr      bool
	}
	cases := map[string]testCase{
		"happy path": {
			file: "Z000000004000000061000010000000002016500000005000000000722200000000600000000000000000000000000000000000000000000",
			expectedFooter: FileFooter{
				RecordHeader: RecordHeader{
					RecordType:      FooterRecord,
					recordCount:     4,
					OriginatorID:    "0000000610",
					FileCreationNum: 1,
				},
				TotalValueOfDebit:    20165,
				TotalCountOfDebit:    5,
				TotalValueOfCredit:   72220,
				TotalCountOfCredit:   6,
				TotalValueOfERecords: 0,
				TotalCountOfERecords: 0,
				TotalValueOfFRecords: 0,
				TotalCountOfFRecords: 0,
			},
		},
		"missing footer": {
			file:      "A0000000010000000610000102313861210hello               CAD",
			expectErr: true,
		},
		"footer too small": {
			file:      "Z12355",
			expectErr: true,
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			file := strings.NewReader(tc.file)
			streamer := NewFileStream(file)
			footer, err := streamer.GetFooter()
			if !tc.expectErr {
				r.NoError(err)
				r.NotNil(footer)
				r.Equal(tc.expectedFooter, *footer)
			} else {
				r.Error(err)
			}

		})
	}
}

func TestStreamTxns(t *testing.T) {
	r := require.New(t)
	// test simple case of one txn per line
	validFile := `A0000000010000000001000102327512345                    CAD
D0000000030000000001000140000000010000232759876543211234        0000000000000000012345000Short name     payor name                    my long name                                               1234567891111111                                            00000000000
Z000000004000000000100010000000000700000000007000000000000000000000000000000000000000000000000000000000000000000`
	reader := strings.NewReader(validFile)

	fs1 := NewFileStream(reader)
	date := time.Date(2023, 10, 2, 0, 0, 0, 0, time.UTC)
	txn := Ptr(NewDebit("400", 1000, &date, "987654321", "1234", "0000000000000000012345", "Short name", "payor name", "my long name", "123456789", "1111111"))
	parsedTxn, err := fs1.ScanTxn()
	r.NoError(err)
	r.Equal(txn, parsedTxn)

	// test many transactions across multiple lines
	rawCreditAndDebitFile := `A0000000010000000001000102327512345                    CAD
C00000000200000000010001450000000100002327512345678912345       0000000000000012313213000short name     payee name                    someone                                                    00000123112345                                              00000000000450000000100002327512345678912345       0000000000000012313213000short name     payee name                    someone                                                    00000123112345                                              00000000000450000000100002327512345678912345       0000000000000012313213000short name     payee name                    someone                                                    00000123112345                                              00000000000450000000100002327512345678912345       0000000000000012313213000short name     payee name                    someone                                                    00000123112345                                              00000000000450000000100002327512345678912345       0000000000000012313213000short name     payee name                    someone                                                    00000123112345                                              00000000000450000000100002327512345678912345       0000000000000012313213000short name     payee name                    someone                                                    00000123112345                                              00000000000
C00000000300000000010001450000000100002327512345678912345       0000000000000012313213000short name     payee name                    someone                                                    00000123112345                                              00000000000                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                
D0000000040000000001000140000000010000232759876543211234        0000000000000000012345000Short name     payor name                    my long name                                               1234567891111111                                            0000000000040000000010000232759876543211234        0000000000000000012345000Short name     payor name                    my long name                                               1234567891111111                                            0000000000040000000010000232759876543211234        0000000000000000012345000Short name     payor name                    my long name                                               1234567891111111                                            0000000000040000000010000232759876543211234        0000000000000000012345000Short name     payor name                    my long name                                               1234567891111111                                            0000000000040000000010000232759876543211234        0000000000000000012345000Short name     payor name                    my long name                                               1234567891111111                                            0000000000040000000010000232759876543211234        0000000000000000012345000Short name     payor name                    my long name                                               1234567891111111                                            00000000000
D0000000050000000001000140000000010000232759876543211234        0000000000000000012345000Short name     payor name                    my long name                                               1234567891111111                                            00000000000                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                
Z000000006000000000100010000000001400000000007000000000140000000000700000000000000000000000000000000000000000000`
	reader2 := strings.NewReader(rawCreditAndDebitFile)
	fs2 := NewFileStream(reader2)
	var txns []Transaction
	complexTxns := append(txns,
		Ptr(NewCredit("450", 1000, &date, "123456789", "12345", "0000000000000012313213", "short name", "payee name", "someone", "000001231", "12345")),
		Ptr(NewCredit("450", 1000, &date, "123456789", "12345", "0000000000000012313213", "short name", "payee name", "someone", "000001231", "12345")),
		Ptr(NewCredit("450", 1000, &date, "123456789", "12345", "0000000000000012313213", "short name", "payee name", "someone", "000001231", "12345")),
		Ptr(NewCredit("450", 1000, &date, "123456789", "12345", "0000000000000012313213", "short name", "payee name", "someone", "000001231", "12345")),
		Ptr(NewCredit("450", 1000, &date, "123456789", "12345", "0000000000000012313213", "short name", "payee name", "someone", "000001231", "12345")),
		Ptr(NewCredit("450", 1000, &date, "123456789", "12345", "0000000000000012313213", "short name", "payee name", "someone", "000001231", "12345")),
		Ptr(NewCredit("450", 1000, &date, "123456789", "12345", "0000000000000012313213", "short name", "payee name", "someone", "000001231", "12345")),
		Ptr(NewDebit("400", 1000, &date, "987654321", "1234", "0000000000000000012345", "Short name", "payor name", "my long name", "123456789", "1111111")),
		Ptr(NewDebit("400", 1000, &date, "987654321", "1234", "0000000000000000012345", "Short name", "payor name", "my long name", "123456789", "1111111")),
		Ptr(NewDebit("400", 1000, &date, "987654321", "1234", "0000000000000000012345", "Short name", "payor name", "my long name", "123456789", "1111111")),
		Ptr(NewDebit("400", 1000, &date, "987654321", "1234", "0000000000000000012345", "Short name", "payor name", "my long name", "123456789", "1111111")),
		Ptr(NewDebit("400", 1000, &date, "987654321", "1234", "0000000000000000012345", "Short name", "payor name", "my long name", "123456789", "1111111")),
		Ptr(NewDebit("400", 1000, &date, "987654321", "1234", "0000000000000000012345", "Short name", "payor name", "my long name", "123456789", "1111111")),
		Ptr(NewDebit("400", 1000, &date, "987654321", "1234", "0000000000000000012345", "Short name", "payor name", "my long name", "123456789", "1111111")))
	for _, txn := range complexTxns {
		parsedTxn, err := fs2.ScanTxn()
		r.NoError(err)
		r.Equal(txn, parsedTxn)
	}

	// Test two error lines and one valid txn, the vaid txn should still get picked up
	invalidFile := `A0000000010000000001000102327512345                    CAD
T0000000030000000001000140000000010000232759876543211234        0000000000000000012345000Short name     payor name                    my long name                                               1234567891111111                                            00000000000
invlaisd dsadjkldsajdkslaj jdsla
D0000000030000000001000140000000010000232759876543211234        0000000000000000012345000Short name     payor name                    my long name                                               1234567891111111                                            00000000000
Z000000004000000000100010000000000700000000007000000000000000000000000000000000000000000000000000000000000000000`

	reader3 := strings.NewReader(invalidFile)
	fs3 := NewFileStream(reader3)
	// Find the valid txn in the invalid file
	for err != io.EOF {
		parsedTxn, err = fs3.ScanTxn()
		if err != nil {
			r.NotErrorIs(err, ErrScanParseError)
		} else {
			r.Equal(txn, parsedTxn)
		}
	}

	// Test if a transaction in the middle of a line is not parsable the other txns still get parsed
	invlaidFile2 := `A0000000010000000001000102327512345                    CAD
	D0000000030000000001000140000000010000232759876543211234        0000000000000000012345000Short name     payor name                    my long name                                               1234567891111111                                            00000000000400000000abcd0232759876543211234        0000000000000000012345000Short name     payor name                    my long name                                               1234567891111111                                            0000000000040000000010000232759876543211234        0000000000000000012345000Short name     payor name                    my long name                                               1234567891111111                                            00000000000
	C00000000400000000010001450000000100002327512345678912345       0000000000000012313213000short name     payee name                    someone                                                    00000123112345                                              00000000000                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                
	Z000000004000000000100010000000000700000000007000000000000000000000000000000000000000000000000000000000000000000`
	reader4 := strings.NewReader(invlaidFile2)
	fs4 := NewFileStream(reader4)
	expectedTxns := []Transaction{
		Ptr(NewDebit("400", 1000, &date, "987654321", "1234", "0000000000000000012345", "Short name", "payor name", "my long name", "123456789", "1111111")),
		Ptr(NewDebit("400", 1000, &date, "987654321", "1234", "0000000000000000012345", "Short name", "payor name", "my long name", "123456789", "1111111")),
		Ptr(NewCredit("450", 1000, &date, "123456789", "12345", "0000000000000012313213", "short name", "payee name", "someone", "000001231", "12345")),
	}

	i := 0
	currentTxnIdx := 0
	parsedTxn, err = fs4.ScanTxn()
	for err != io.EOF {
		if i == 1 {
			r.Error(err)
			var perr *ParseError
			r.ErrorAs(err, &perr)

		} else {
			r.NoError(err)
			r.Equal(expectedTxns[currentTxnIdx], parsedTxn)
			currentTxnIdx++
		}
		i++
		parsedTxn, err = fs4.ScanTxn()
	}

	parseErrorFile := `A0000000010000000001000102327512345                    CAD
	D00000000300000000010001400000000asdfg232749876543211234        0000000000000000012345000Short name     payor name                    my long name                                               1234567891111111                                            0000000000040000000010000232759876543211234        0000000000000000012345000Short name     payor name                    my long name                                               1234567891111111                                            00000000000
	`
	fs5 := NewFileStream(strings.NewReader(parseErrorFile))
	for i := 0; i < 3; i++ {
		txn, err := fs5.ScanTxn()
		if i == 0 {
			var perr *ParseError
			r.ErrorAs(err, &perr)
		} else if i == 1 {
			r.NoError(err)
			r.Equal(Ptr(NewDebit("400", 1000, &date, "987654321", "1234", "0000000000000000012345", "Short name", "payor name", "my long name", "123456789", "1111111")), txn)
		} else {
			r.ErrorIs(err, io.EOF)
		}
	}
}
