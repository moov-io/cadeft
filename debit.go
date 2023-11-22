package cadeft

import (
	"strings"
	"time"

	"github.com/pkg/errors"
)

// Debit represents Logical Record Type D according to the EFT standard 005
type Debit struct {
	BaseTxn
	DueDate             *time.Time `json:"due_date" validate:"required"`
	PayorAccountNo      string     `json:"payor_account_no" validate:"required,max=12,numeric"`
	PayorName           string     `json:"payor_name" validate:"required,max=30"`
	ReturnInstitutionID string     `json:"return_institution_id" validate:"required,max=9,numeric"`
	ReturnAccountNo     string     `json:"return_account_no" validate:"required,max=12,eft_alpha"`
}

func NewDebit(
	txnType TransactionType,
	amount int64,
	dueDate *time.Time,
	institutionID string,
	payorAccountNo string,
	itemTraceNo string,
	originatorShortName string,
	payorName string,
	originatorLongName string,
	returnInstitutionID string,
	returnAccountNo string,
	opts ...BaseTxnOpt,
) Debit {
	base := BaseTxn{
		TxnType:               txnType,
		Amount:                amount,
		ItemTraceNo:           itemTraceNo,
		InstitutionID:         institutionID,
		OriginatorShortName:   originatorShortName,
		OriginatorLongName:    originatorLongName,
		RecordType:            DebitRecord,
		StoredTransactionType: "000",
	}
	for _, o := range opts {
		o(&base)
	}
	return Debit{
		BaseTxn:             base,
		DueDate:             dueDate,
		PayorAccountNo:      payorAccountNo,
		PayorName:           payorName,
		ReturnInstitutionID: returnInstitutionID,
		ReturnAccountNo:     returnAccountNo,
	}
}

// Build serializes a Debit into a 240 length string that adheres to the EFT standard 005 standard.
// Numeric fields are padded with zeros to the left and alphanumeric fields are padded with spaces to the right
// any missing fields are filled with 0's or blanks
func (d Debit) Build() (string, error) {
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
		return "", errors.Wrap(err, "failed to format originator short name")
	}
	sb.WriteString(shortName)
	payorName, err := formatName(d.PayorName, 30)
	if err != nil {
		return "", errors.Wrap(err, "failed to format payor name")
	}
	sb.WriteString(payorName)
	longName, err := formatName(d.OriginatorLongName, 30)
	if err != nil {
		return "", errors.Wrap(err, "failed to format originator long name")
	}
	sb.WriteString(longName)
	sb.WriteString(abreviateStringToLength(d.UserID, 10))
	sb.WriteString(abreviateStringToLength(d.CrossRefNo, 19))
	sb.WriteString(padNumericStringWithZeros(d.ReturnInstitutionID, 9))
	sb.WriteString(abreviateStringToLength(d.ReturnAccountNo, 12))
	sb.WriteString(abreviateStringToLength(d.SundryInfo, 15))
	sb.WriteString(createFillerString(22))
	sb.WriteString(abreviateStringToLength(d.SettlementCode, 2))
	sb.WriteString(padNumericStringWithZeros("", 11))
	return sb.String(), nil
}

// Parse takes in a serialized transaction segment and populates a Debit struct containing the relevant data.
// The data passed in should be of length 240, the transaction length associated with the EFT file spec.
func (d *Debit) Parse(data string) error {
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
	d.ReturnInstitutionID = strings.TrimSpace(data[169:178])
	d.ReturnAccountNo = strings.TrimSpace(data[178:190])
	d.SundryInfo = strings.TrimSpace(data[190:205])
	// filler at 205:227
	d.SettlementCode = strings.TrimSpace(data[227:229])
	d.RecordType = DebitRecord
	return nil
}

// Validate checks whether the fields of a Debit struct contain the correct fields that are required when writing/reading an EFT file.
// The validation check can be found on Section D of EFT standard 005.
func (d Debit) Validate() error {
	if err := eftValidator.Struct(&d); err != nil {
		return err
	}
	return nil
}

func (d Debit) GetType() RecordType {
	return DebitRecord
}

func (d Debit) GetAmount() int64 {
	return d.Amount
}
func (d Debit) GetBaseTxn() BaseTxn {
	return d.BaseTxn
}
func (d Debit) GetAccountNo() string {
	return d.PayorAccountNo
}
func (d Debit) GetDate() *time.Time {
	return d.DueDate
}
func (d Debit) GetName() string {
	return d.PayorName
}
func (d Debit) GetReturnInstitutionID() string {
	return d.ReturnInstitutionID
}
func (d Debit) GetReturnAccountNo() string {
	return d.ReturnAccountNo
}
func (d Debit) GetOriginalInstitutionID() string {
	return ""
}
func (d Debit) GetOriginalAccountNo() string {
	return ""
}
func (d Debit) GetOriginalItemTraceNo() string {
	return ""
}
