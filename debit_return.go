package cadeft

import (
	"fmt"
	"strings"
	"time"
)

// DebitReturn represents Logical Record Type J according to the EFT standard 005
type DebitReturn struct {
	BaseTxn
	DueDate               *time.Time `json:"due_date" validate:"required"`
	PayorAccountNo        string     `json:"payor_account_no" validate:"required,max=12,numeric"`
	PayorName             string     `json:"payor_name" validate:"required,max=30"`
	OriginalInstitutionID string     `json:"original_institution_id" validate:"required,max=9,numeric"`
	OriginalAccountNo     string     `json:"original_account_no" validate:"required,max=12,eft_alpha"`
	OriginalItemTraceNo   string     `json:"original_item_trace_no" validate:"required,eft_num,max=22"`
}

func NewDebitReturn(
	txnType TransactionType,
	amount int64,
	dueDate *time.Time,
	institutionID string,
	payorAccountNo string,
	itemTraceNo string,
	originatorShortName string,
	payorName string,
	originatorLongName string,
	originalInstitutionID string,
	originalAccountNo string,
	originalItemTraceNo string,
	opts ...BaseTxnOpt,
) DebitReturn {
	base := BaseTxn{
		TxnType:               txnType,
		Amount:                amount,
		ItemTraceNo:           itemTraceNo,
		InstitutionID:         institutionID,
		OriginatorShortName:   originatorShortName,
		OriginatorLongName:    originatorLongName,
		RecordType:            ReturnDebitRecord,
		StoredTransactionType: "000",
	}
	for _, o := range opts {
		o(&base)
	}
	return DebitReturn{
		BaseTxn:               base,
		DueDate:               dueDate,
		PayorAccountNo:        payorAccountNo,
		PayorName:             payorName,
		OriginalInstitutionID: originalInstitutionID,
		OriginalAccountNo:     originalAccountNo,
		OriginalItemTraceNo:   originalItemTraceNo,
	}
}

// Build serializes a DebitReturn into a 240 length string that adheres to the EFT standard 005 standard.
// Numeric fields are padded with zeros to the left and alphanumeric fields are padded with spaces to the right
// any missing fields are filled with 0's or blanks
func (d DebitReturn) Build() (string, error) {
	var sb strings.Builder
	sb.Grow(240)
	sb.WriteString(padNumericStringWithZeros(string(d.TxnType), 3))
	sb.WriteString(convertNumToZeroPaddedString(d.Amount, 10))
	if d.DueDate != nil {
		sb.WriteString(padNumericStringWithZeros(convertTimestampToEftDate(*d.DueDate), 6))
	} else {
		sb.WriteString(padNumericStringWithZeros("", 6))
	}
	sb.WriteString(padNumericStringWithZeros(d.InstitutionID, 9))
	sb.WriteString(abreviateStringToLength(d.PayorAccountNo, 12))
	sb.WriteString(padNumericStringWithZeros(d.ItemTraceNo, 22))
	sb.WriteString(padNumericStringWithZeros(string(d.StoredTransactionType), 3))
	shortName, err := formatName(d.OriginatorShortName, 15)
	if err != nil {
		return "", fmt.Errorf("failed to format originator short name: %w", err)
	}
	sb.WriteString(shortName)
	payorName, err := formatName(d.PayorName, 30)
	if err != nil {
		return "", fmt.Errorf("failed to format payor name: %w", err)
	}
	sb.WriteString(payorName)
	longName, err := formatName(d.OriginatorLongName, 30)
	if err != nil {
		return "", fmt.Errorf("failed to format originator long name: %w", err)
	}
	sb.WriteString(longName)
	sb.WriteString(abreviateStringToLength(d.UserID, 10))
	sb.WriteString(abreviateStringToLength(d.CrossRefNo, 19))
	sb.WriteString(padNumericStringWithZeros(d.OriginalInstitutionID, 9))
	sb.WriteString(abreviateStringToLength(d.OriginalAccountNo, 12))
	sb.WriteString(abreviateStringToLength(d.SundryInfo, 15))
	sb.WriteString(padNumericStringWithZeros(d.OriginalItemTraceNo, 22))
	sb.WriteString(abreviateStringToLength(d.SettlementCode, 2))
	sb.WriteString(padNumericStringWithTrailingZeros(d.InvalidDataElementID, 11))
	return sb.String(), nil
}

// Parse takes in a serialized transaction segment and populates a DebitReturn struct containing the relevant data.
// The data passed in should be of length 240, the transaction length associated with the EFT file spec.
func (d *DebitReturn) Parse(data string) error {
	var err error
	if len(data) != segmentLength {
		return NewParseError(ErrInvalidRecordLength, "")
	}
	d.TxnType = TransactionType(data[:3])
	d.Amount, err = parseNum(data[3:13])
	if err != nil {
		return NewParseError(err, "failed to parse amount")
	}
	dueDate, err := parseDate(data[13:19])
	if err != nil {
		return NewParseError(err, "failed to parse due date")
	}
	d.DueDate = &dueDate
	d.InstitutionID = data[19:28]
	d.PayorAccountNo = strings.TrimSpace(data[28:40])
	d.ItemTraceNo = data[40:62]
	d.StoredTransactionType = TransactionType(data[62:65])
	d.OriginatorShortName = strings.TrimSpace(data[65:80])
	d.PayorName = strings.TrimSpace(data[80:110])
	d.OriginatorLongName = strings.TrimSpace(data[110:140])
	d.UserID = strings.TrimSpace(data[140:150])
	d.CrossRefNo = strings.TrimSpace(data[150:169])
	d.OriginalInstitutionID = strings.TrimSpace(data[169:178])
	d.OriginalAccountNo = strings.TrimSpace(data[178:190])
	d.SundryInfo = strings.TrimSpace(data[190:205])
	d.OriginalItemTraceNo = strings.TrimSpace(data[205:227])
	d.SettlementCode = strings.TrimSpace(data[227:229])
	d.InvalidDataElementID = strings.TrimSpace(data[229:240])
	d.RecordType = ReturnDebitRecord
	return nil
}

// Validate checks whether the fields of a DebitReturn struct contain the correct fields that are required when writing/reading an EFT file.
// The validation check can be found on Section D of EFT standard 005.
func (d DebitReturn) Validate() error {
	if err := eftValidator.Struct(&d); err != nil {
		return err
	}
	return nil
}

func (d DebitReturn) GetType() RecordType {
	return ReturnDebitRecord
}

func (d DebitReturn) GetAmount() int64 {
	return d.Amount
}

func (dr DebitReturn) GetBaseTxn() BaseTxn {
	return dr.BaseTxn
}
func (dr DebitReturn) GetAccountNo() string {
	return dr.PayorAccountNo
}
func (dr DebitReturn) GetDate() *time.Time {
	return dr.DueDate
}
func (dr DebitReturn) GetName() string {
	return dr.PayorName
}
func (dr DebitReturn) GetReturnInstitutionID() string {
	return ""

}
func (dr DebitReturn) GetReturnAccountNo() string {
	return ""
}
func (dr DebitReturn) GetOriginalInstitutionID() string {
	return dr.OriginalInstitutionID
}
func (dr DebitReturn) GetOriginalAccountNo() string {
	return dr.OriginalAccountNo
}
func (dr DebitReturn) GetOriginalItemTraceNo() string {
	return dr.OriginalItemTraceNo
}
