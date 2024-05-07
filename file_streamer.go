package cadeft

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/hashicorp/go-multierror"
)

// FileStreamer is used for stream parsing an EFT file. Instead of reading the whole file FileStreamer attempts to read segments of transactions line by line.
// The caller is responsible for handling/ignoring any errors that are encountered. FileStreamer stores the state of the parser and hence is not safe for concurrency usage.
// Main usage is via ScanTxn which attempts to parse a segment of the file, return a Transaction or an error and move the file pointer along.
type FileStreamer struct {
	r              io.ReadSeeker
	scanner        *bufio.Scanner
	lineContents   string
	numTxnsPerLine int
	currentTxn     int
	currentLine    int
}

func NewFileStream(in io.ReadSeeker) FileStreamer {
	return FileStreamer{
		r:       in,
		scanner: bufio.NewScanner(in),
	}
}

// GetHeader scans the file for a A record and attempts to parse the record and return a FileHeader, an error is returned if parsing fails.
// the file pointer is then reset to the beginning of the file.
func (fs FileStreamer) GetHeader() (*FileHeader, error) {
	scanner := bufio.NewScanner(fs.r)
	defer func() {
		_, _ = fs.r.Seek(0, io.SeekStart)
	}()
	// Frst line of the file should be the header
	success := scanner.Scan()
	if !success {
		return nil, fmt.Errorf("failed to scan for file header: %w", scanner.Err())
	}

	line := scanner.Text()
	if len(line) == 0 {
		return nil, fmt.Errorf("file header is empty")
	}

	recType, err := parseRecordType(string(line[0]))
	if err != nil {
		return nil, fmt.Errorf("file header not found: %w", err)
	}
	if recType != HeaderRecord {
		return nil, fmt.Errorf("first record in file is not a header record")
	}
	header := &FileHeader{}
	err = header.parse(line)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file header: %w", err)
	}
	return header, nil
}

// GetFooter attempts to seek to the end of the file in search of a Z record. If no footer record is found an error is returned.
// Once scanning is complete the file pointer is reset to the beginning of the file.
func (fs FileStreamer) GetFooter() (*FileFooter, error) {
	scanner := bufio.NewScanner(fs.r)
	defer func() {
		_, _ = fs.r.Seek(0, io.SeekStart)
	}()
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) < 1 {
			return nil, fmt.Errorf("line too short to determine record type")
		}
		recType := line[:1]
		if recType == string(FooterRecord) {
			ff := &FileFooter{}
			if err := ff.Parse(line); err != nil {
				return nil, fmt.Errorf("failed to parse file footer: %w", err)
			}
			return ff, nil
		}
	}
	return nil, fmt.Errorf("failed to find footer record")
}

// ScanTxn parses transaction records (D, C, I, J, E and F logical records) one at a time. Upon successfully parsing a transaction segment a Transaction struct is returned otherwise a non nil error is returned in
// both cases the file pointer is moved to the next transaction in the file. When parsing is complete  (ether file has ended or a footer record is encountered) an io.EOF error is returned, the caller can use this to
// terminate parsing of the file. If ScanTxn returns an instance of ScanParseError the error can be ignored by the caller. ScanTxn() will return EOF if a Footer record is encountered or if a line is empty
func (fs *FileStreamer) ScanTxn() (Transaction, error) {
	// forward the scanner to the first txn record
	if fs.currentLine == 0 {
		if !fs.scanner.Scan() {
			if fs.scanner.Err() != nil {
				return nil, fmt.Errorf("failed forward reader to txn record: %w", fs.scanner.Err())
			}
			return nil, io.EOF

		}
		if len(fs.scanner.Text()) == 0 || !isHeaderRecordType(string(fs.scanner.Text()[0])) {
			return nil, fmt.Errorf("first line in file is not a header record")
		}
		fs.currentLine++
	}

	// read a new line
	if fs.currentTxn == 0 && fs.numTxnsPerLine == 0 {
		if !fs.scanner.Scan() {
			if fs.scanner.Err() != nil {
				return nil, fmt.Errorf("failed to read transactions: %w", fs.scanner.Err())
			}
			return nil, io.EOF
		}
		fs.currentLine++

		line, err := normalize(strings.TrimSpace(fs.scanner.Text()))
		if err != nil {
			return nil, fmt.Errorf("failed to read transaction line: %w", err)
		}

		fs.lineContents = line

		if len(fs.lineContents) == 0 || isFooterRecord(string(fs.lineContents[0])) {
			return nil, io.EOF
		}

		if len(fs.lineContents[commonRecordDataLength:])%segmentLength != 0 {
			return nil, fmt.Errorf("txn record at line %d is not of correct length", fs.currentLine)
		}

		fs.numTxnsPerLine = len(fs.lineContents[commonRecordDataLength:]) / segmentLength

	}

	recordType, err := parseRecordType(fs.lineContents[:1])
	if err != nil {
		fs.currentTxn, fs.numTxnsPerLine = 0, 0
		return nil, fmt.Errorf("unrecognized record type at line %d: %w", fs.currentLine, err)
	}

	if fs.currentTxn == (fs.numTxnsPerLine - 1) {
		defer fs.reset()
	}
	defer fs.incrementTxnCount()

	// determine starting index and ending index from currentTxn
	txnsSegment := fs.lineContents[commonRecordDataLength:]
	startIdx := segmentLength * fs.currentTxn
	endIdx := segmentLength * (fs.currentTxn + 1)

	var txn Transaction
	switch recordType {
	case DebitRecord:
		d := Debit{}
		err = d.Parse(txnsSegment[startIdx:endIdx])
		if err != nil {
			return nil, NewParseError(err, "")
		}
		txn = &d
	case CreditRecord:
		c := Credit{}
		if err := c.Parse(txnsSegment[startIdx:endIdx]); err != nil {
			return nil, newStreamParseError(err, string(DebitRecord), fs.currentTxn, fs.currentLine)
		}
		txn = &c
	case ReturnDebitRecord:
		dr := DebitReturn{}
		if err := dr.Parse(txnsSegment[startIdx:endIdx]); err != nil {
			return nil, newStreamParseError(err, string(ReturnDebitRecord), fs.currentTxn, fs.currentLine)
		}
		txn = &dr
	case ReturnCreditRecord:
		cr := CreditReturn{}
		if err := cr.Parse(txnsSegment[startIdx:endIdx]); err != nil {
			return nil, newStreamParseError(err, string(ReturnCreditRecord), fs.currentTxn, fs.currentLine)
		}
		txn = &cr
	case CreditReverseRecord:
		cr := CreditReverse{}
		if err := cr.Parse(txnsSegment[startIdx:endIdx]); err != nil {
			return nil, newStreamParseError(err, string(CreditReverseRecord), fs.currentTxn, fs.currentLine)
		}
		txn = &cr
	case DebitReverseRecord:
		dr := DebitReverse{}
		if err := dr.Parse(txnsSegment[startIdx:endIdx]); err != nil {
			return nil, newStreamParseError(err, string(DebitReverseRecord), fs.currentTxn, fs.currentLine)
		}
		txn = &dr
	case HeaderRecord, NoticeOfChangeRecord, NoticeOfChangeHeader, NoticeOfChangeFooter, FooterRecord:
		return nil, fmt.Errorf("invalid record type: %v", recordType)
	}
	return txn, nil
}

func (fs *FileStreamer) incrementTxnCount() {
	fs.currentTxn++
}

func (fs *FileStreamer) reset() {
	fs.currentTxn = 0
	fs.numTxnsPerLine = 0
}

func newStreamParseError(err error, recordType string, txnNum, line int) error {
	return multierror.Append(
		err,
		fmt.Errorf("parse error for record %s number %d line %d", recordType, txnNum, line),
		ErrScanParseError,
	)
}
