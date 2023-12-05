package cadeft

import (
	"fmt"
	"strings"
	"time"
)

// CreditReturn represents Logical Record Type I according to the EFT standard 005
type CreditReturn struct {
	BaseTxn
	DateFundsAvailable    *time.Time `json:"date_funds_available" validate:"required"`
	PayeeAccountNo        string     `json:"payee_account_no" validate:"required,max=12,numeric"`
	PayeeName             string     `json:"payee_name" validate:"required,max=30"`
	OriginalInstitutionID string     `json:"original_institution_id" validate:"required,max=9,numeric"`
	OriginalAccountNo     string     `json:"original_account_no" validate:"required,max=12,eft_alpha"`
	OriginalItemTraceNo   string     `json:"original_item_trace_no" validate:"required,eft_num,max=22"`
}

func NewCreditReturn(
	txnType TransactionType,
	amount int64,
	dateFundsAvailable *time.Time,
	institutionID string,
	payeeAccountNo string,
	itemTraceNo string,
	originatorShortName string,
	payeeName string,
	originatorLongName string,
	originalInstitutionID string,
	originalAccountNo string,
	originalItemTraceNo string,
	opts ...BaseTxnOpt) CreditReturn {
	base := BaseTxn{
		TxnType:               txnType,
		Amount:                amount,
		ItemTraceNo:           itemTraceNo,
		InstitutionID:         institutionID,
		OriginatorShortName:   originatorShortName,
		OriginatorLongName:    originatorLongName,
		RecordType:            ReturnCreditRecord,
		StoredTransactionType: "000",
	}
	for _, o := range opts {
		o(&base)
	}
	return CreditReturn{
		BaseTxn:               base,
		DateFundsAvailable:    dateFundsAvailable,
		PayeeAccountNo:        payeeAccountNo,
		PayeeName:             payeeName,
		OriginalInstitutionID: originalInstitutionID,
		OriginalAccountNo:     originalAccountNo,
		OriginalItemTraceNo:   originalItemTraceNo,
	}
}

// Build serializes a CreditReturn into a 240 length string that adheres to the EFT standard 005 standard.
// Numeric fields are padded with zeros to the left and alphanumeric fields are padded with spaces to the right
// any missing fields are filled with 0's or blanks
func (c CreditReturn) Build() (string, error) {
	var sb strings.Builder
	sb.Grow(240)
	sb.WriteString(padNumericStringWithZeros(string(c.TxnType), 3))
	sb.WriteString(convertNumToZeroPaddedString(c.Amount, 10))
	if c.DateFundsAvailable != nil {
		sb.WriteString(padNumericStringWithZeros(convertTimestampToEftDate(*c.DateFundsAvailable), 6))
	} else {
		sb.WriteString(padNumericStringWithZeros("", 6))
	}
	sb.WriteString(padNumericStringWithZeros(c.InstitutionID, 9))
	sb.WriteString(abreviateStringToLength(c.PayeeAccountNo, 12))
	sb.WriteString(padNumericStringWithZeros(c.ItemTraceNo, 22))
	sb.WriteString(padNumericStringWithZeros(string(c.StoredTransactionType), 3))
	shortName, err := formatName(c.OriginatorShortName, 15)
	if err != nil {
		return "", fmt.Errorf("failed to format originator short name: %w", err)
	}
	sb.WriteString(shortName)
	payeeName, err := formatName(c.PayeeName, 30)
	if err != nil {
		return "", fmt.Errorf("failed to format payee name: %w", err)
	}
	sb.WriteString(payeeName)
	longName, err := formatName(c.OriginatorLongName, 30)
	if err != nil {
		return "", fmt.Errorf("failed to format originator long name: %w", err)
	}
	sb.WriteString(longName)
	sb.WriteString(abreviateStringToLength(c.UserID, 10))
	sb.WriteString(abreviateStringToLength(c.CrossRefNo, 19))
	sb.WriteString(padNumericStringWithZeros(c.OriginalInstitutionID, 9))
	sb.WriteString(abreviateStringToLength(c.OriginalAccountNo, 12))
	sb.WriteString(abreviateStringToLength(c.SundryInfo, 15))
	sb.WriteString(padNumericStringWithZeros(c.OriginalItemTraceNo, 22))
	sb.WriteString(abreviateStringToLength(c.SettlementCode, 2))
	sb.WriteString(padNumericStringWithZeros(c.InvalidDataElementID, 11))
	return sb.String(), nil
}

// Parse takes in a serialized transaction segment and populates a CreditReturn struct containing the relevant data.
// The data passed in should be of length 240, the transaction length associated with the EFT file spec.
func (c *CreditReturn) Parse(data string) error {
	var err error
	if len(data) != segmentLength {
		return NewParseError(ErrInvalidRecordLength, "")
	}
	c.TxnType = TransactionType(data[:3])
	c.Amount, err = parseNum(data[3:13])
	if err != nil {
		return NewParseError(err, "failed to parse amount")
	}
	dateFundsAvail, err := parseDate(data[13:19])
	if err != nil {
		return NewParseError(err, "failed to parse date funds available")
	}
	c.DateFundsAvailable = &dateFundsAvail
	c.InstitutionID = data[19:28]
	c.PayeeAccountNo = strings.TrimSpace(data[28:40])
	c.ItemTraceNo = data[40:62]
	c.StoredTransactionType = TransactionType(data[62:65])
	c.OriginatorShortName = strings.TrimSpace(data[65:80])
	c.PayeeName = strings.TrimSpace(data[80:110])
	c.OriginatorLongName = strings.TrimSpace(data[110:140])
	c.UserID = strings.TrimSpace(data[140:150])
	c.CrossRefNo = strings.TrimSpace(data[150:169])
	c.OriginalInstitutionID = strings.TrimSpace(data[169:178])
	c.OriginalAccountNo = strings.TrimSpace(data[178:190])
	c.SundryInfo = strings.TrimSpace(data[190:205])
	c.OriginalItemTraceNo = strings.TrimSpace(data[205:227])
	c.SettlementCode = strings.TrimSpace(data[227:229])
	c.InvalidDataElementID = strings.TrimSpace(data[229:240])
	c.RecordType = ReturnCreditRecord
	return nil
}

// Validate checks whether the fields of a CreditReturn struct contain the correct fields that are required when writing/reading an EFT file.
// The validation check can be found on Section D of EFT standard 005.
func (c CreditReturn) Validate() error {
	if err := eftValidator.Struct(&c); err != nil {
		return err
	}
	return nil
}

func (c CreditReturn) GetType() RecordType {
	return ReturnCreditRecord
}

func (c CreditReturn) GetAmount() int64 {
	return c.Amount
}
func (c CreditReturn) GetBaseTxn() BaseTxn {
	return c.BaseTxn
}
func (c CreditReturn) GetAccountNo() string {
	return c.PayeeAccountNo
}
func (c CreditReturn) GetDate() *time.Time {
	return c.DateFundsAvailable
}
func (c CreditReturn) GetName() string {
	return c.PayeeName
}
func (c CreditReturn) GetReturnInstitutionID() string {
	return ""
}
func (c CreditReturn) GetReturnAccountNo() string {
	return ""
}
func (c CreditReturn) GetOriginalInstitutionID() string {
	return c.OriginalInstitutionID
}
func (c CreditReturn) GetOriginalAccountNo() string {
	return c.OriginalAccountNo
}
func (c CreditReturn) GetOriginalItemTraceNo() string {
	return c.OriginalItemTraceNo
}
