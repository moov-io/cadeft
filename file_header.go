package cadeft

import (
	"strings"
	"time"

	"github.com/pkg/errors"
)

// FileHeader represents Logical Record Type A in the EFT 005 spec.
// Every EFT file should have a header entry which contains information about the originator, currency and creation dates.
// the supported currency should be CAD or USD.
type FileHeader struct {
	RecordHeader
	CreationDate                   *time.Time `json:"creation_date" validate:"required"`
	DestinationDataCenterNo        int64      `json:"destination_data_center" validate:"min=0,max=99999,numeric"`
	DirectClearerCommunicationArea string     `json:"communication_area" validate:"max=20"`
	CurrencyCode                   string     `json:"currency_code" validate:"required,eft_cur,len=3"`
}

func NewFileHeader(
	originatorId string,
	fileCreationNum int64,
	creationDate *time.Time,
	destinationDataCenterNo int64,
	currencyCode string,
	opts ...HeaderOpts) *FileHeader {
	fh := &FileHeader{
		RecordHeader: RecordHeader{
			RecordType:      HeaderRecord,
			recordCount:     1,
			OriginatorID:    originatorId,
			FileCreationNum: fileCreationNum,
		},
		CreationDate:            creationDate,
		DestinationDataCenterNo: destinationDataCenterNo,
		CurrencyCode:            currencyCode,
	}
	for _, o := range opts {
		o(fh)
	}
	return fh
}

type HeaderOpts func(*FileHeader)

// This option will add an optional direct clearer communication area field to the header
func WithDirectClearerCommunicationArea(comm string) HeaderOpts {
	return func(fh *FileHeader) {
		fh.DirectClearerCommunicationArea = comm
	}
}

func (fh *FileHeader) parse(line string) error {
	var err error
	recordHeader := RecordHeader{}
	if len(line) < aRecordMinLength {
		return errors.New("invalid header record length")
	}
	err = recordHeader.parse(line)
	if err != nil {
		return errors.Wrap(err, "failed to parse record header")
	}
	fh.RecordHeader = recordHeader
	creationDate, err := parseDate(line[24:30])
	if err != nil {
		return errors.Wrap(err, "failed to parse creation date for file header")
	}
	fh.CreationDate = &creationDate
	fh.DestinationDataCenterNo, err = parseNum(line[30:35])
	if err != nil {
		return errors.Wrap(err, "failed to parse destination data center for file header")
	}
	fh.DirectClearerCommunicationArea = strings.TrimSpace(line[35:55])

	fh.CurrencyCode = strings.TrimSpace(line[55:58])
	return nil
}

// Validate checks whether the fields of a FileHeader struct contain the correct fields that are required when writing/reading an EFT file.
// The validation check can be found on Section D of EFT standard 005.
func (fh FileHeader) Validate() error {
	if err := eftValidator.Struct(&fh); err != nil {
		return err
	}
	return nil
}

func (fh FileHeader) buildHeader(currRecordCount int) (string, error) {
	var sb strings.Builder

	rh, err := fh.RecordHeader.buildRecordHeader()
	if err != nil {
		return "", errors.Wrap(err, "failed to write record header for file header")
	}
	sb.WriteString(rh)
	if fh.CreationDate != nil {
		sb.WriteString(convertTimestampToEftDate(*fh.CreationDate))
	} else {
		sb.WriteString(convertNumToZeroPaddedString(0, 6))
	}
	sb.WriteString(convertNumToZeroPaddedString(fh.DestinationDataCenterNo, 5))

	// add buffering of 20 empty characters Reserved Customer-Direct Clearer Communication area
	if fh.DirectClearerCommunicationArea == "" {
		sb.WriteString(createFillerString(20))
	} else {
		sb.WriteString(padStringWithBlanks(fh.DirectClearerCommunicationArea, 20))
	}

	sb.WriteString(padStringWithBlanks(fh.CurrencyCode, 3))
	sb.WriteString(createFillerString(1406))
	return sb.String(), nil
}

func isHeaderRecordType(s string) bool {
	switch s {
	case string(HeaderRecord), string(NoticeOfChangeHeader):
		return true
	default:
		return false
	}
}
