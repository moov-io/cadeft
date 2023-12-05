package cadeft

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/samber/lo"
)

const (
	maxLineLength          = 1464
	commonRecordDataLength = 24
	segmentLength          = 240
	zRecordMinLength       = 112
	aRecordMinLength       = 58
	maxTxnsPerRecord       = 6
)

type RecordType string
type DCSign string
type TransactionType string

const (
	HeaderRecord         RecordType = "A"
	CreditRecord         RecordType = "C"
	DebitRecord          RecordType = "D"
	CreditReverseRecord  RecordType = "E"
	DebitReverseRecord   RecordType = "F"
	ReturnCreditRecord   RecordType = "I"
	ReturnDebitRecord    RecordType = "J"
	NoticeOfChangeRecord RecordType = "S"
	NoticeOfChangeHeader RecordType = "U"
	NoticeOfChangeFooter RecordType = "V"
	FooterRecord         RecordType = "Z"
)

type Transaction interface {
	Parse(string) error
	Build() (string, error)
	Validate() error
	GetType() RecordType
	GetAmount() int64
	GetBaseTxn() BaseTxn
	GetAccountNo() string
	GetDate() *time.Time
	GetName() string
	GetReturnInstitutionID() string
	GetReturnAccountNo() string
	GetOriginalInstitutionID() string
	GetOriginalAccountNo() string
	GetOriginalItemTraceNo() string
}

type Transactions []Transaction
type File struct {
	Header *FileHeader  `json:"file_header,omitempty"`
	Txns   Transactions `json:"transactions,omitempty"`
	Footer *FileFooter  `json:"file_footer,omitempty"`
}

func NewFile(header *FileHeader, txns []Transaction) File {
	return File{
		Header: header,
		Txns:   txns,
	}
}

// Create returns a serialized EFT file as a string or an error.
// The serialized file will adhere to the EFT 005 payments canada specification but depending on individual fields the file may be rejected.
// Use the Validate function to catch any validation errors. Make sure to add the appropriate FileHeader and Transactions via NewFile before calling Create.
func (f *File) Create() (string, error) {
	// 1. run validation checks
	var sb strings.Builder
	currentLine := 1
	if f.Header == nil {
		return "", fmt.Errorf("file header is missing")
	}
	serializedHeader, err := f.Header.buildHeader(currentLine)
	if err != nil {
		return "", err
	}
	sb.WriteString(serializedHeader)
	currentLine++

	serializedTxns, err := f.buildTransactions(f.Header.RecordHeader, &currentLine)
	if err != nil {
		return "", err
	}
	sb.WriteString(serializedTxns)

	// if the user provides a footer use that otherwise create a new one
	if f.Footer != nil {
		f.Footer.recordCount = int64(currentLine)
		serializedFooter, err := f.Footer.Build()
		if err != nil {
			return "", fmt.Errorf("failed to build footer: %w", err)
		}
		sb.WriteString(serializedFooter)
	} else {
		// Build the footer
		footerRecordHeader := RecordHeader{
			RecordType:      FooterRecord,
			recordCount:     int64(currentLine),
			OriginatorID:    f.Header.OriginatorID,
			FileCreationNum: f.Header.FileCreationNum,
		}
		footer := NewFileFooter(footerRecordHeader, f.Txns)
		if f.Footer == nil {
			f.Footer = footer
		}
		serializedFooter, err := footer.Build()
		if err != nil {
			return "", err
		}
		sb.WriteString(serializedFooter)
	}
	return sb.String(), nil
}

// Returns all debit transactions or D records
func (f File) GetAllDebitTxns() []Debit {
	txns := make([]Debit, 0)
	for _, t := range f.Txns {
		if txn, ok := t.(*Debit); ok {
			txns = append(txns, *txn)
		}
	}
	return txns
}

// Returns all credit transactions or C records
func (f File) GetAllCredits() []Credit {
	txns := make([]Credit, 0)
	for _, t := range f.Txns {
		if txn, ok := t.(*Credit); ok {
			txns = append(txns, *txn)
		}
	}
	return txns
}

// Returns all debit return transactions or J records
func (f File) GetAllDebitReturns() []DebitReturn {
	txns := make([]DebitReturn, 0)
	for _, t := range f.Txns {
		if txn, ok := t.(*DebitReturn); ok {
			txns = append(txns, *txn)
		}
	}
	return txns
}

// Returns all credit return transactions or I records
func (f File) GetAllCreditReturns() []CreditReturn {
	txns := make([]CreditReturn, 0)
	for _, t := range f.Txns {
		if txn, ok := t.(*CreditReturn); ok {
			txns = append(txns, *txn)
		}
	}
	return txns
}

func (f File) buildTransactions(recordHeader RecordHeader, currentLine *int) (string, error) {
	recTypeToTxs := map[RecordType][]Transaction{
		DebitRecord:         make([]Transaction, 0),
		CreditRecord:        make([]Transaction, 0),
		CreditReverseRecord: make([]Transaction, 0),
		DebitReverseRecord:  make([]Transaction, 0),
		ReturnCreditRecord:  make([]Transaction, 0),
		ReturnDebitRecord:   make([]Transaction, 0),
	}
	for idx, t := range f.Txns {
		switch t.GetType() {
		case DebitRecord:
			recTypeToTxs[DebitRecord] = append(recTypeToTxs[DebitRecord], t)
		case CreditRecord:
			recTypeToTxs[CreditRecord] = append(recTypeToTxs[CreditRecord], t)
		case CreditReverseRecord:
			recTypeToTxs[CreditReverseRecord] = append(recTypeToTxs[CreditReverseRecord], t)
		case DebitReverseRecord:
			recTypeToTxs[DebitReverseRecord] = append(recTypeToTxs[DebitReverseRecord], t)
		case ReturnCreditRecord:
			recTypeToTxs[ReturnCreditRecord] = append(recTypeToTxs[ReturnCreditRecord], t)
		case ReturnDebitRecord:
			recTypeToTxs[ReturnDebitRecord] = append(recTypeToTxs[ReturnDebitRecord], t)
		case HeaderRecord, NoticeOfChangeRecord, NoticeOfChangeHeader, NoticeOfChangeFooter, FooterRecord:
			return "", fmt.Errorf("transaction[%d] has unexpected record type: %v", idx, t.GetType())
		}
	}
	var sb strings.Builder
	for recType, entries := range recTypeToTxs {
		recordHeader.RecordType = recType
		entriesStr, err := f.buildTxnEntries(entries, recordHeader, currentLine)
		if err != nil {
			return "", fmt.Errorf("failed to build txn entries: %w", err)
		}
		if len(entriesStr) > 0 {
			sb.WriteString(padStringWithBlanks(entriesStr, maxLineLength))
		}
	}
	sb.WriteString("\n")
	return sb.String(), nil
}

func (f File) buildTxnEntries(txns []Transaction, recordHeader RecordHeader, currentLine *int) (string, error) {
	var sb strings.Builder
	var txnSb strings.Builder
	recordHeader.recordCount = int64(*currentLine)
	recordHeaderStr, err := recordHeader.buildRecordHeader()
	if err != nil {
		return "", fmt.Errorf("failed to create Txn RecordHeader: %w", err)
	}
	// go through all the transactions and each line should contain at most 6 txns
	for count, txn := range txns {
		txnStr, err := txn.Build()
		if err != nil {
			return "", fmt.Errorf("failed to build transaction: %w", err)
		}
		txnSb.WriteString(txnStr)
		if (count+1)%maxTxnsPerRecord == 0 {
			//flush into sb
			sb.WriteString("\n")
			sb.WriteString(recordHeaderStr)
			// pad txn with blanks to adhere to length requirement of MAX_LINE_LENGTH
			blankPad := createFillerString(maxLineLength - txnSb.Len() - len(recordHeaderStr))
			txnSb.WriteString(blankPad)
			sb.WriteString(txnSb.String())
			txnSb.Reset()
			*currentLine++
			recordHeader.recordCount = int64(*currentLine)
			recordHeaderStr, err = recordHeader.buildRecordHeader()
			if err != nil {
				return "", fmt.Errorf("failed to create Txn RecordHeader: %w", err)
			}
		}
	}

	//left over debit txns txnSb should be populated
	if txnSb.Len() > 0 {
		sb.WriteString("\n")
		sb.WriteString(recordHeaderStr)
		sb.WriteString(txnSb.String())
		// write the filler to make the length of the line MAX_LINE_LENGTH
		filler := createFillerString(maxLineLength - txnSb.Len() - len(recordHeaderStr))
		sb.WriteString(filler)
		*currentLine++
	}

	return sb.String(), nil
}

// Validate runs validation on the entire file starting from the FileHeader then every Taransaction.
// Any error that is encountered will be appended to a multierror and returned to the caller.
func (f File) Validate() error {
	var err error
	headerErr := f.Header.Validate()
	if headerErr != nil {
		err = multierror.Append(err, headerErr)
	}
	for i, t := range f.Txns {
		txnErr := t.Validate()
		if txnErr != nil {
			err = multierror.Append(err, fmt.Errorf("faild to validate txn %d: %w", i, txnErr))
		}
	}
	return err
}

func NewTransaction(
	recordType RecordType,
	txnType TransactionType,
	amount int64,
	date *time.Time,
	institutionID string,
	payorPayeeAccountNo string,
	itemTraceNo string,
	originatorShortName string,
	payorPayeeName string,
	originatorLongName string,
	originalOrReturnInstitutionID string,
	originalOrReturnAccountNo string,
	originalItemTraceNo string,
	opts ...BaseTxnOpt,
) Transaction {
	var txn Transaction
	switch recordType {
	case DebitRecord:
		txn = lo.ToPtr(NewDebit(txnType, amount, date, institutionID, payorPayeeAccountNo, itemTraceNo, originatorShortName, payorPayeeName, originatorLongName, originalOrReturnInstitutionID, originalOrReturnAccountNo, opts...))
	case CreditRecord:
		txn = lo.ToPtr(NewCredit(txnType, amount, date, institutionID, payorPayeeAccountNo, itemTraceNo, originatorShortName, payorPayeeName, originatorLongName, originalOrReturnInstitutionID, originalOrReturnAccountNo, opts...))
	case ReturnDebitRecord:
		txn = lo.ToPtr(NewDebitReturn(txnType, amount, date, institutionID, payorPayeeAccountNo, itemTraceNo, originatorShortName, payorPayeeName, originatorLongName, originalOrReturnInstitutionID, originalOrReturnAccountNo, originalItemTraceNo, opts...))
	case ReturnCreditRecord:
		txn = lo.ToPtr(NewCreditReturn(txnType, amount, date, institutionID, payorPayeeAccountNo, itemTraceNo, originatorShortName, payorPayeeName, originatorLongName, originalOrReturnInstitutionID, originalOrReturnAccountNo, originalItemTraceNo, opts...))
	case CreditReverseRecord:
		txn = lo.ToPtr(NewCreditReverse(txnType, amount, date, institutionID, payorPayeeAccountNo, itemTraceNo, originatorShortName, payorPayeeName, originatorLongName, originalOrReturnInstitutionID, originalOrReturnAccountNo, originalItemTraceNo, opts...))
	case DebitReverseRecord:
		txn = lo.ToPtr(NewDebitReverse(txnType, amount, date, institutionID, payorPayeeAccountNo, itemTraceNo, originatorShortName, payorPayeeName, originatorLongName, originalOrReturnInstitutionID, originalOrReturnAccountNo, originalItemTraceNo, opts...))
	case HeaderRecord, NoticeOfChangeRecord, NoticeOfChangeHeader, NoticeOfChangeFooter, FooterRecord:
		return nil
	}
	return txn
}

func (f *Transactions) UnmarshalJSON(data []byte) error {
	var txns []interface{}

	if err := json.Unmarshal(data, &txns); err != nil {
		return err
	}
	for _, t := range txns {

		tMap, ok := t.(map[string]interface{})
		if !ok {
			return fmt.Errorf("failed to unmarshall transaction")
		}
		rawRecordType, ok := tMap["type"]
		if !ok {
			return fmt.Errorf("failed to unmarshall transaction")
		}
		recordType, ok := rawRecordType.(string)
		if !ok {
			return fmt.Errorf("failed to unmarshall transaction")
		}

		jsonStr, err := json.Marshal(tMap)
		if err != nil {
			return err
		}

		switch recordType {
		case string(DebitRecord):

			var d Debit
			err = json.Unmarshal(jsonStr, &d)
			if err != nil {
				return err
			}
			*f = append(*f, &d)
		case string(CreditRecord):
			var c Credit
			err = json.Unmarshal(jsonStr, &c)
			if err != nil {
				return err
			}
			*f = append(*f, &c)
		case string(ReturnCreditRecord):
			var cr CreditReturn
			err = json.Unmarshal(jsonStr, &cr)
			if err != nil {
				return err
			}
			*f = append(*f, &cr)
		case string(ReturnDebitRecord):
			var dr DebitReturn
			err = json.Unmarshal(jsonStr, &dr)
			if err != nil {
				return err
			}
			*f = append(*f, &dr)
		}
	}
	return nil
}

func getTotalValueAndCount(recordType RecordType, txns []Transaction) (int64, int64) {
	var totalValue int64
	var totalCount int64
	for _, t := range txns {
		if t.GetType() == recordType {
			totalValue += t.GetAmount()
			totalCount++
		}
	}
	return totalValue, totalCount
}

func parseRecordType(t string) (RecordType, error) {
	switch t {
	case "A":
		return HeaderRecord, nil
	case "D":
		return DebitRecord, nil
	case "C":
		return CreditRecord, nil
	case "E":
		return CreditReverseRecord, nil
	case "F":
		return DebitReverseRecord, nil
	case "I":
		return ReturnCreditRecord, nil
	case "J":
		return ReturnDebitRecord, nil
	case "Z":
		return FooterRecord, nil
	default:
		return "", fmt.Errorf(fmt.Sprintf("unrecognized record type: %s", t))
	}
}
