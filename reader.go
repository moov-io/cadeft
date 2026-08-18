package cadeft

import (
	"bufio"
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
	// Allow CPA lines longer than bufio's default 64 KiB buffer. CPA Std
	// 005 lines are at most 1,464 chars; in UTF-8 with French chars they
	// can stretch a bit past that. 1 MiB is plenty.
	r.scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	for r.scanner.Scan() {
		line, err := normalize(r.scanner.Text())
		if err != nil {
			return File{}, fmt.Errorf("failed to read line: %w", err)
		}
		if line == "" {
			continue
		}
		recordType := string([]rune(line)[:1])
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
	if len([]rune(data)) < aRecordMinLength {
		return fmt.Errorf("record type A is not required length")
	}
	fHeader := &FileHeader{}
	if err := fHeader.parse(data); err != nil {
		return fmt.Errorf("failed to parse file header: %w", err)
	}
	r.File.Header = fHeader
	return nil
}

func (r *Reader) parseTxnRecord(data string) error {
	rs := []rune(data)
	if len(rs) < commonRecordDataLength {
		return fmt.Errorf("txn record shorter than common header length %d: got %d", commonRecordDataLength, len(rs))
	}
	if len(rs[commonRecordDataLength:])%segmentLength != 0 {
		return fmt.Errorf("record length is not valid multiple of %d, partial txn: %d", segmentLength, len(rs[commonRecordDataLength:]))
	}
	body := rs[commonRecordDataLength:]
	numSegments := len(body) / segmentLength

	recType, err := parseRecordType(string(rs[:1]))
	if err != nil {
		return fmt.Errorf("failed to parse transaction: %w", err)
	}

	for i := 0; i < numSegments; i++ {
		seg := string(body[i*segmentLength : (i+1)*segmentLength])
		if isFillerString(seg) {
			continue
		}
		switch recType {
		case DebitRecord:
			debit := Debit{}
			if err := debit.Parse(seg); err != nil {
				return fmt.Errorf("failed to parse debit transaction: %w", err)
			}
			r.File.Txns = append(r.File.Txns, &debit)
		case CreditRecord:
			credit := Credit{}
			if err := credit.Parse(seg); err != nil {
				return fmt.Errorf("failed to parse credit transaction: %w", err)
			}
			r.File.Txns = append(r.File.Txns, &credit)
		case ReturnDebitRecord:
			debitReturn := DebitReturn{}
			if err := debitReturn.Parse(seg); err != nil {
				return fmt.Errorf("failed to parse debit return transaction: %w", err)
			}
			r.File.Txns = append(r.File.Txns, &debitReturn)
		case ReturnCreditRecord:
			creditReturns := CreditReturn{}
			if err := creditReturns.Parse(seg); err != nil {
				return fmt.Errorf("failed to parse credit return transaction: %w", err)
			}
			r.File.Txns = append(r.File.Txns, &creditReturns)
		case CreditReverseRecord:
			creditReverse := CreditReverse{}
			if err := creditReverse.Parse(seg); err != nil {
				return fmt.Errorf("failed to parse credit reverse transaction: %w", err)
			}
			r.File.Txns = append(r.File.Txns, &creditReverse)
		case DebitReverseRecord:
			debitReverseRecord := DebitReverse{}
			if err := debitReverseRecord.Parse(seg); err != nil {
				return fmt.Errorf("failed to parse debit reverse transaction: %w", err)
			}
			r.File.Txns = append(r.File.Txns, &debitReverseRecord)
		case HeaderRecord, FooterRecord, NoticeOfChangeRecord, NoticeOfChangeHeader, NoticeOfChangeFooter:
			return fmt.Errorf("unexpected %s record", recType)
		}
	}
	return nil
}

func (r *Reader) parseZRecord(data string) error {
	if len([]rune(data)) < zRecordMinLength {
		return fmt.Errorf("z record does not contain minimum amount of data")
	}

	footer := &FileFooter{}
	if err := footer.Parse(data); err != nil {
		return fmt.Errorf("failed to parse file footer: %w", err)
	}
	r.File.Footer = footer
	return nil
}
