package cadeft

import (
	"strings"
	"time"

	"github.com/pkg/errors"
)

// Credit represents Logical Record Type C according to the EFT standard 005
type Credit struct {
	BaseTxn
	DateFundsAvailable  *time.Time `json:"date_funds_available" validate:"required"`
	PayeeAccountNo      string     `json:"payee_account_no" validate:"required,max=12,numeric"`
	PayeeName           string     `json:"payee_name" validate:"required,max=30"`
	ReturnInstitutionID string     `json:"return_institution_id" validate:"required,max=9,numeric"`
	ReturnAccountNo     string     `json:"return_account_no" validate:"required,max=12,eft_alpha"`
}

func NewCredit(
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
	opts ...BaseTxnOpt,
) Credit {
	base := BaseTxn{
		TxnType:               txnType,
		Amount:                amount,
		ItemTraceNo:           itemTraceNo,
		InstitutionID:         institutionID,
		OriginatorShortName:   originatorShortName,
		OriginatorLongName:    originatorLongName,
		RecordType:            CreditRecord,
		StoredTransactionType: "000",
	}
	for _, o := range opts {
		o(&base)
	}
	return Credit{
		BaseTxn:             base,
		DateFundsAvailable:  dateFundsAvailable,
		PayeeAccountNo:      payeeAccountNo,
		PayeeName:           payeeName,
		ReturnInstitutionID: returnInstitutionID,
		ReturnAccountNo:     returnAccountNo,
	}
}

// Parse takes in a serialized transaction segment and populates a Credit struct containing the relevant data.
// The data passed in should be of length 240, the transaction length associated with the EFT file spec.
func (c *Credit) Parse(data string) error {
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
	// filler at 205:227
	c.SettlementCode = strings.TrimSpace(data[227:229])
	c.RecordType = CreditRecord
	return nil
}

// Build serializes a Credit into a 240 length string that adheres to the EFT standard 005 standard.
// Numeric fields are padded with zeros to the left and alphanumeric fields are padded with spaces to the right
// any missing fields are filled with 0's or blanks
func (c Credit) Build() (string, error) {
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
		return "", errors.Wrap(err, "failed to format originator short name")
	}
	sb.WriteString(shortName)
	payeeName, err := formatName(c.PayeeName, 30)
	if err != nil {
		return "", errors.Wrap(err, "failed to format payee name")
	}
	sb.WriteString(payeeName)
	longName, err := formatName(c.OriginatorLongName, 30)
	if err != nil {
		return "", errors.Wrap(err, "failed to format originator long name")
	}
	sb.WriteString(longName)
	sb.WriteString(abreviateStringToLength(c.UserID, 10))
	sb.WriteString(abreviateStringToLength(c.CrossRefNo, 19))
	sb.WriteString(padNumericStringWithZeros(c.ReturnInstitutionID, 9))
	sb.WriteString(abreviateStringToLength(c.ReturnAccountNo, 12))
	sb.WriteString(abreviateStringToLength(c.SundryInfo, 15))
	sb.WriteString(createFillerString(22))
	sb.WriteString(abreviateStringToLength(c.SettlementCode, 2))
	sb.WriteString(padNumericStringWithZeros("", 11))
	return sb.String(), nil
}

// Validate checks whether the fields of a Credit struct contain the correct fields that are required when writing/reading an EFT file.
// The validation check can be found on Section D of EFT standard 005.
func (c Credit) Validate() error {
	if err := eftValidator.Struct(&c); err != nil {
		return err
	}
	return nil
}

func (c Credit) GetType() RecordType {
	return CreditRecord
}

func (c Credit) GetAmount() int64 {
	return c.Amount
}

func (c Credit) GetBaseTxn() BaseTxn {
	return c.BaseTxn
}
func (c Credit) GetAccountNo() string {
	return c.PayeeAccountNo
}
func (c Credit) GetDate() *time.Time {
	return c.DateFundsAvailable
}
func (c Credit) GetName() string {
	return c.PayeeName
}
func (c Credit) GetReturnInstitutionID() string {
	return c.ReturnInstitutionID
}
func (c Credit) GetReturnAccountNo() string {
	return c.ReturnAccountNo
}
func (c Credit) GetOriginalInstitutionID() string {
	return ""
}
func (c Credit) GetOriginalAccountNo() string {
	return ""
}
func (c Credit) GetOriginalItemTraceNo() string {
	return ""
}
