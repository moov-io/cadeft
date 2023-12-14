package cadeft

import (
	"fmt"
	"strings"
)

// FileFooter represents either Logical Record Type Z or V in the EFT 005 standard spec.
// FileFooter fields are populated from the Transaction entries and their counts and values.
// When building a file unless you really need to add a custom FileFooter allow the library to generate it for you via File.Create()
type FileFooter struct {
	RecordHeader
	TotalValueOfDebit    int64 `json:"total_value_debit"`
	TotalCountOfDebit    int64 `json:"total_count_debit"`
	TotalValueOfCredit   int64 `json:"total_value_credit"`
	TotalCountOfCredit   int64 `json:"total_count_credit"`
	TotalValueOfERecords int64 `json:"total_value_reverse_debit"`
	TotalCountOfERecords int64 `json:"total_count_reverse_debit"`
	TotalValueOfFRecords int64 `json:"total_value_reverse_credit"`
	TotalCountOfFRecords int64 `json:"total_count_reverse_credit"`
}

func NewFileFooter(recordHeader RecordHeader, txns []Transaction) *FileFooter {
	totalValDebit, totalCountDebit := getTotalValueAndCount(DebitRecord, txns)
	totalValDebitReturns, totalCountDebitReturns := getTotalValueAndCount(ReturnDebitRecord, txns)
	totalValCredit, totalCountCredit := getTotalValueAndCount(CreditRecord, txns)
	totalValCreditReturns, totalCountCreditReturns := getTotalValueAndCount(ReturnCreditRecord, txns)
	totalValCreditReversal, totalCountCreditReversal := getTotalValueAndCount(CreditReverseRecord, txns)
	totalValDebitReversal, totalCountDebitReversal := getTotalValueAndCount(DebitReverseRecord, txns)
	totalValDebit += totalValDebitReturns
	totalCountDebit += totalCountDebitReturns
	totalValCredit += totalValCreditReturns
	totalCountCredit += totalCountCreditReturns
	return &FileFooter{
		RecordHeader:         recordHeader,
		TotalValueOfDebit:    totalValDebit,
		TotalCountOfDebit:    totalCountDebit,
		TotalValueOfCredit:   totalValCredit,
		TotalCountOfCredit:   totalCountCredit,
		TotalValueOfERecords: totalValCreditReversal,
		TotalCountOfERecords: totalCountCreditReversal,
		TotalValueOfFRecords: totalValDebitReversal,
		TotalCountOfFRecords: totalCountDebitReversal,
	}
}

// Parse will take in a serialized footer record of type Z and parse the amounts into a FileFooter struct
func (ff *FileFooter) Parse(line string) error {
	var err error
	if len(line) < zRecordMinLength {
		return fmt.Errorf("footer is too short")
	}

	recordHeader := RecordHeader{}
	err = recordHeader.parse(line)
	if err != nil {
		return fmt.Errorf("failed to parse record header for footer: %w", err)
	}
	ff.RecordHeader = recordHeader
	ff.TotalValueOfDebit, err = parseNum(line[24:38])
	if err != nil {
		return fmt.Errorf("failed to parse total value of debit: %w", err)
	}
	ff.TotalCountOfDebit, err = parseNum(line[38:46])
	if err != nil {
		return fmt.Errorf("failed to parse total count of debit: %w", err)
	}
	ff.TotalValueOfCredit, err = parseNum(line[46:60])
	if err != nil {
		return fmt.Errorf("failed to parse total value of credit: %w", err)
	}
	ff.TotalCountOfCredit, err = parseNum(line[60:68])
	if err != nil {
		return fmt.Errorf("failed to parse total count of credit: %w", err)
	}
	valERecordSegment := line[68:82]
	if !isFillerString(valERecordSegment) {
		if ff.TotalValueOfERecords, err = parseNum(valERecordSegment); err != nil {
			return fmt.Errorf("failed to parse total value of E records: %w", err)
		}
	}

	numERecordSegment := line[82:90]
	if !isFillerString(numERecordSegment) {
		if ff.TotalCountOfERecords, err = parseNum(numERecordSegment); err != nil {
			return fmt.Errorf("failed to parse total count of E records: %w", err)
		}
	}

	valFRecordsSegment := line[90:104]
	if !isFillerString(valFRecordsSegment) {
		if ff.TotalValueOfFRecords, err = parseNum(line[90:104]); err != nil {
			return fmt.Errorf("failed to parse total value of F records: %w", err)
		}
	}

	numFRecordSegment := line[104:112]
	if !isFillerString(numFRecordSegment) {
		if ff.TotalCountOfFRecords, err = parseNum(line[104:112]); err != nil {
			return fmt.Errorf("failed to parse totoal count of F records: %w", err)
		}
	}

	return nil
}

// Build takes the FileFooter struct, constructs a serialized string and returns it for writing
func (ff FileFooter) Build() (string, error) {
	var sb strings.Builder
	serializedHeader, err := ff.RecordHeader.buildRecordHeader()
	if err != nil {
		return "", fmt.Errorf("failed to write record header: %w", err)
	}
	sb.WriteString(serializedHeader)
	sb.WriteString(convertNumToZeroPaddedString(ff.TotalValueOfDebit, 14))
	sb.WriteString(convertNumToZeroPaddedString(ff.TotalCountOfDebit, 8))
	sb.WriteString(convertNumToZeroPaddedString(ff.TotalValueOfCredit, 14))
	sb.WriteString(convertNumToZeroPaddedString(ff.TotalCountOfCredit, 8))
	sb.WriteString(convertNumToZeroPaddedString(ff.TotalValueOfERecords, 14))
	sb.WriteString(convertNumToZeroPaddedString(ff.TotalCountOfERecords, 8))
	sb.WriteString(convertNumToZeroPaddedString(ff.TotalValueOfFRecords, 14))
	sb.WriteString(convertNumToZeroPaddedString(ff.TotalCountOfFRecords, 8))
	sb.WriteString(createFillerString(1352))
	return sb.String(), nil
}

func (ff FileFooter) GetType() RecordType {
	return FooterRecord
}

func isFooterRecord(s string) bool {
	switch s {
	case string(FooterRecord), string(NoticeOfChangeFooter):
		return true
	default:
		return false
	}
}
