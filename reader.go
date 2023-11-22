package cadeft

import (
	"bufio"
	"io"

	"github.com/pkg/errors"
)

type Reader struct {
	File    File
	scanner *bufio.Scanner
}

func NewReader(in io.Reader) *Reader {
	return &Reader{
		scanner: bufio.NewScanner(in),
	}
}

// ReadFile will attempt to read the whole EFT file according to the 005 spec from payments canada.
// If no errors are encountered a populated File object is returned that contains the Header, Transactions and Footer.
// Use the FileStreamer object to be able ignore errors and proceed parsing the file.
func (r *Reader) ReadFile() (File, error) {
	for r.scanner.Scan() {
		line := r.scanner.Text()
		recordType := line[:1]
		if recordType == string(HeaderRecord) {
			if err := r.parseARecord(line); err != nil {
				return File{}, errors.Wrap(err, "failed to parse header")
			}
		} else if isTxnRecord(recordType) {
			if err := r.parseTxnRecord(line); err != nil {
				return File{}, errors.Wrap(err, "failed to parse txn")
			}
		} else if recordType == string(FooterRecord) {
			if err := r.parseZRecord(line); err != nil {
				return File{}, errors.Wrap(err, "failed to parse footer")
			}
		}
	}
	return r.File, nil
}

func (r *Reader) parseARecord(data string) error {
	if len(data) < aRecordMinLength {
		return errors.New("record type A is not required length")
	}
	fHeader := &FileHeader{}
	if err := fHeader.parse(data); err != nil {
		return errors.Wrap(err, "failed to parse file header")
	}
	r.File.Header = fHeader
	return nil
}

func (r *Reader) parseTxnRecord(data string) error {
	if len(data[commonRecordDataLength:])%segmentLength != 0 {
		return errors.Errorf("record length is not valid multiple of 260, partial txn: %d", len(data[commonRecordDataLength:]))
	}
	rawTxnSegment := data[commonRecordDataLength:]
	numSegments := len(rawTxnSegment) / segmentLength
	// make this into a function
	recType, err := parseRecordType(data[:1])
	if err != nil {
		return errors.Wrap(err, "failed to parse transaction")
	}

	startIdx := 0
	endIdx := segmentLength
	for i := 0; i < numSegments; i++ {
		if isFillerString(rawTxnSegment[startIdx:endIdx]) {
			continue
		}
		switch recType {
		case DebitRecord:
			debit := Debit{}
			if err := debit.Parse(rawTxnSegment[startIdx:endIdx]); err != nil {
				return errors.Wrap(err, "failed to parse debit transaction")
			}
			r.File.Txns = append(r.File.Txns, &debit)
		case CreditRecord:
			credit := Credit{}
			if err := credit.Parse(rawTxnSegment[startIdx:endIdx]); err != nil {
				return errors.Wrap(err, "failed to parse credit transaction")
			}
			r.File.Txns = append(r.File.Txns, &credit)
		case ReturnDebitRecord:
			debitReturn := DebitReturn{}
			if err := debitReturn.Parse(rawTxnSegment[startIdx:endIdx]); err != nil {
				return errors.Wrap(err, "failed to parse debit return transaction")
			}
			r.File.Txns = append(r.File.Txns, &debitReturn)
		case ReturnCreditRecord:
			creditReturns := CreditReturn{}
			if err := creditReturns.Parse(rawTxnSegment[startIdx:endIdx]); err != nil {
				return errors.Wrap(err, "failed to parse credit return transaction")
			}
			r.File.Txns = append(r.File.Txns, &creditReturns)
		}
		startIdx = endIdx
		endIdx += segmentLength
	}
	return nil
}

func (r *Reader) parseZRecord(data string) error {
	if len(data) < zRecordMinLength {
		return errors.New("z record does not contain minimum amount of data")
	}

	footer := &FileFooter{}
	if err := footer.Parse(data); err != nil {
		return errors.Wrap(err, "failed to parse file footer")
	}
	r.File.Footer = footer
	return nil
}
