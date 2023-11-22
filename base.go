package cadeft

// BaseTxn represents the common fields of every Transaction record D, C, E, F, I and J
type BaseTxn struct {
	TxnType               TransactionType `json:"txn_type" validate:"required,numeric,max=3"`
	Amount                int64           `json:"amount" validate:"required,max=9999999999"`
	ItemTraceNo           string          `json:"item_trace_no" validate:"eft_num,max=22"`
	InstitutionID         string          `json:"institution_id" validate:"required,eft_num,max=9"`
	StoredTransactionType TransactionType `json:"stored_txn_type" validate:"eft_num,max=3"`
	OriginatorShortName   string          `json:"short_name" validate:"max=15"`
	OriginatorLongName    string          `json:"long_name" validate:"required,max=30"`
	UserID                string          `json:"user_id" validate:"eft_alpha,max=10"`
	CrossRefNo            string          `json:"cross_ref_no" validate:"eft_alpha,max=19"`
	SundryInfo            string          `json:"sundry_info" validate:"eft_alpha,max=15"`
	SettlementCode        string          `json:"settlement_code" validate:"eft_alpha,max=2"`
	InvalidDataElementID  string          `json:"invalid_data_element_id" validate:"eft_num,max=11"`
	RecordType            RecordType      `json:"type" validate:"rec_type"`
}
type BaseTxnOpt func(d *BaseTxn)

// WithUserID sets the Originating DirectClearer's User's ID, field 14
func WithUserID(s string) BaseTxnOpt {
	return func(d *BaseTxn) {
		d.UserID = s
	}
}

// WithCrossRefNo sets the Originator'sCross Reference No. field 15
func WithCrossRefNo(s string) BaseTxnOpt {
	return func(d *BaseTxn) {
		d.CrossRefNo = s
	}
}

// WithSundryInfo sets the Originator's Sundry Information, field 18
func WithSundryInfo(s string) BaseTxnOpt {
	return func(d *BaseTxn) {
		d.SundryInfo = s
	}
}

// WithSettlementCode sets the Originator-Direct Clearer Settlement Code, field 20
func WithSettlementCode(s string) BaseTxnOpt {
	return func(d *BaseTxn) {
		d.SettlementCode = s
	}
}

// WithStoredTransactionType sets the Stored Transaction Type, field 10
// mainly used for returns and reversals (J, I, E and F records)
func WithStoredTransactionType(s string) BaseTxnOpt {
	return func(d *BaseTxn) {
		d.StoredTransactionType = TransactionType(s)
	}
}

// WithInvalidDataElementID sets the Invalid Data Element No. field 21
// mainly used for returns to indicate which fields are incorrect
func WithInvalidDataElementID(s string) BaseTxnOpt {
	return func(d *BaseTxn) {
		d.InvalidDataElementID = s
	}
}

// GetAmount returns the amount being transacted
func (b BaseTxn) GetAmount() int64 {
	return b.Amount
}
