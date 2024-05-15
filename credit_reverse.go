package cadeft

import (
	"fmt"
	"strings"
	"time"
)

// CreditReverse represents Logical Record Type E according to the EFT standard 005
type CreditReverse struct {
	BaseTxn
	DateFundsAvailable  *time.Time `json:"date_funds_available" validate:"required"`
	PayeeAccountNo      string     `json:"payee_account_no" validate:"required,max=12,numeric"`
	PayeeName           string     `json:"payee_name" validate:"required,max=30"`
	ReturnInstitutionID string     `json:"return_institution_id" validate:"required,max=9,numeric"`
	ReturnAccountNo     string     `json:"return_account_no" validate:"required,max=12,eft_alpha"`
	OriginalItemTraceNo string     `json:"original_item_trace_no" validate:"required,eft_num,max=22"`
}

func NewCreditReverse(
	txnType TransactionType,
	amount int64,
	dateFundsAvailable *time.Time,
	institutionID string,
	payeeAccountNo string,
	itemTraceNo string,
	originatorShortName string,
	payeeName string,
	originatorLongName string,
	returnInstitutionID string,
	returnAccountNo string,
	originalItemTraceNo string,
	opts ...BaseTxnOpt) CreditReverse {
	base := BaseTxn{
		TxnType:               txnType,
		Amount:                amount,
		ItemTraceNo:           itemTraceNo,
		InstitutionID:         institutionID,
		OriginatorShortName:   originatorShortName,
		OriginatorLongName:    originatorLongName,
		RecordType:            CreditReverseRecord,
		StoredTransactionType: "000",
	}
	for _, o := range opts {
		o(&base)
	}
	return CreditReverse{
		BaseTxn:             base,
		DateFundsAvailable:  dateFundsAvailable,
		PayeeAccountNo:      payeeAccountNo,
		PayeeName:           payeeName,
		ReturnInstitutionID: returnInstitutionID,
		ReturnAccountNo:     returnAccountNo,
		OriginalItemTraceNo: originalItemTraceNo,
	}
}

// Parse takes in a serialized transaction segment and populates a CreditReverse struct containing the relevant data.
// The data passed in should be of length 240, the transaction length associated with the EFT file spec.
func (c *CreditReverse) Parse(data string) error {
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
	c.ReturnInstitutionID = strings.TrimSpace(data[169:178])
	c.ReturnAccountNo = strings.TrimSpace(data[178:190])
	c.SundryInfo = strings.TrimSpace(data[190:205])
	c.OriginalItemTraceNo = strings.TrimSpace(data[205:227])
	c.SettlementCode = strings.TrimSpace(data[227:229])
	c.RecordType = CreditReverseRecord
	return nil
}

// Build serializes a CreditReverse into a 240 length string that adheres to the EFT standard 005 standard.
// Numeric fields are padded with zeros to the left and alphanumeric fields are padded with spaces to the right
// any missing fields are filled with 0's or blanks
func (c CreditReverse) Build() (string, error) {
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
	sb.WriteString(padNumericStringWithZeros(c.ReturnInstitutionID, 9))
	sb.WriteString(abreviateStringToLength(c.ReturnAccountNo, 12))
	sb.WriteString(abreviateStringToLength(c.SundryInfo, 15))
	sb.WriteString(padNumericStringWithZeros(c.OriginalItemTraceNo, 22))
	sb.WriteString(abreviateStringToLength(c.SettlementCode, 2))
	sb.WriteString(padNumericStringWithTrailingZeros(c.InvalidDataElementID, 11))
	return sb.String(), nil
}

func (c CreditReverse) Validate() error {
	if err := eftValidator.Struct(&c); err != nil {
		return err
	}
	return nil
}

func (c CreditReverse) GetType() RecordType {
	return CreditReverseRecord
}

func (c CreditReverse) GetAmount() int64 {
	return c.Amount
}

func (c CreditReverse) GetBaseTxn() BaseTxn {
	return c.BaseTxn
}
func (c CreditReverse) GetAccountNo() string {
	return c.PayeeAccountNo
}
func (c CreditReverse) GetDate() *time.Time {
	return c.DateFundsAvailable
}
func (c CreditReverse) GetName() string {
	return c.PayeeName
}
func (c CreditReverse) GetReturnInstitutionID() string {
	return c.ReturnInstitutionID
}
func (c CreditReverse) GetReturnAccountNo() string {
	return c.ReturnAccountNo
}
func (c CreditReverse) GetOriginalInstitutionID() string {
	return ""
}
func (c CreditReverse) GetOriginalAccountNo() string {
	return ""
}
func (c CreditReverse) GetOriginalItemTraceNo() string {
	return c.OriginalItemTraceNo
}
