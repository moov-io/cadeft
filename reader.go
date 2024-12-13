package cadeft

import (
	"bufio"
	"errors"
	"fmt"
	"io"
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
		line, err := normalize(r.scanner.Text())
		if err != nil {
			return File{}, fmt.Errorf("failed to read line: %w", err)
		}
		recordType := line[:1]
		if recordType == string(HeaderRecord) {
			if err := r.parseARecord(line); err != nil {
				return File{}, fmt.Errorf("failed to parse header: %w", err)
			}
		} else if isTxnRecord(recordType) {
			if err := r.parseTxnRecord(line); err != nil {
				return File{}, fmt.Errorf("failed to parse txn: %w", err)
			}
		} else if recordType == string(FooterRecord) {
			if err := r.parseZRecord(line); err != nil {
				return File{}, fmt.Errorf("failed to parse footer: %w", err)
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
		return fmt.Errorf("failed to parse file header: %w", err)
	}
	r.File.Header = fHeader
	return nil
}

func (r *Reader) parseTxnRecord(data string) error {
	if len(data[commonRecordDataLength:])%segmentLength != 0 {
		return fmt.Errorf("record length is not valid multiple of 260, partial txn: %d", len(data[commonRecordDataLength:]))
	}
	rawTxnSegment := data[commonRecordDataLength:]
	numSegments := len(rawTxnSegment) / segmentLength
	// make this into a function
	recType, err := parseRecordType(data[:1])
	if err != nil {
		return fmt.Errorf("failed to parse transaction: %w", err)
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
				return fmt.Errorf("failed to parse debit transaction: %w", err)
			}
			r.File.Txns = append(r.File.Txns, &debit)
		case CreditRecord:
			credit := Credit{}
			if err := credit.Parse(rawTxnSegment[startIdx:endIdx]); err != nil {
				return fmt.Errorf("failed to parse credit transaction: %w", err)
			}
			r.File.Txns = append(r.File.Txns, &credit)
		case ReturnDebitRecord:
			debitReturn := DebitReturn{}
			if err := debitReturn.Parse(rawTxnSegment[startIdx:endIdx]); err != nil {
				return fmt.Errorf("failed to parse debit return transaction: %w", err)
			}
			r.File.Txns = append(r.File.Txns, &debitReturn)
		case ReturnCreditRecord:
			creditReturns := CreditReturn{}
			if err := creditReturns.Parse(rawTxnSegment[startIdx:endIdx]); err != nil {
				return fmt.Errorf("failed to parse credit return transaction: %w", err)
			}
			r.File.Txns = append(r.File.Txns, &creditReturns)
		case CreditReverseRecord:
			creditReverse := CreditReverse{}
			if err := creditReverse.Parse(rawTxnSegment[startIdx:endIdx]); err != nil {
				return fmt.Errorf("failed to parse credit reverse transaction: %w", err)
			}
			r.File.Txns = append(r.File.Txns, &creditReverse)
		case DebitReverseRecord:
			debitReverseRecord := DebitReverse{}
			if err := debitReverseRecord.Parse(rawTxnSegment[startIdx:endIdx]); err != nil {
				return fmt.Errorf("failed to parse debit reverse transaction: %w", err)
			}
			r.File.Txns = append(r.File.Txns, &debitReverseRecord)
		case HeaderRecord, FooterRecord, NoticeOfChangeRecord, NoticeOfChangeHeader, NoticeOfChangeFooter:
			return fmt.Errorf("unexpected %s record", recType)
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
		return fmt.Errorf("failed to parse file footer: %w", err)
	}
	r.File.Footer = footer
	return nil
}
