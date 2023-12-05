package cadeft

import (
	"fmt"
	"strings"
)

// RecordHeader represnts the common field that begins every Transaction line in an EFT file.
// The contents of this field are the RecordTpye the row count and originator id.
// This field is mostly internal and used for writing the file.
type RecordHeader struct {
	RecordType      RecordType `json:"type" validate:"rec_type"`
	OriginatorID    string     `json:"originator_id" validate:"required,eft_alpha,len=10"`
	FileCreationNum int64      `json:"file_creation_number" validate:"required,max=9999"`
	recordCount     int64      `json:"-"`
}

func (rh *RecordHeader) parse(line string) error {
	var err error
	if rh.RecordType, err = convertRecordType(line[:1]); err != nil {
		return fmt.Errorf("faield to parse RecordHeader: %w", err)
	}

	if rh.recordCount, err = parseNum(line[1:10]); err != nil {
		return fmt.Errorf("failed to parse RecordCount: %w", err)
	}

	rh.OriginatorID = strings.TrimSpace(line[10:20])
	if rh.FileCreationNum, err = parseNum(line[20:24]); err != nil {
		return fmt.Errorf("failed to parse FileCreationNum: %w", err)
	}

	return nil
}

func (rh RecordHeader) buildRecordHeader() (string, error) {
	var sb strings.Builder
	sb.WriteString(abreviateStringToLength(string(rh.RecordType), 1))
	sb.WriteString(convertNumToZeroPaddedString(rh.recordCount, 9))
	sb.WriteString(padNumericStringWithZeros(rh.OriginatorID, 10))
	sb.WriteString(convertNumToZeroPaddedString(rh.FileCreationNum, 4))
	return sb.String(), nil
}
