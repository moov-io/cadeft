package cadeft

import (
	"fmt"
)

type ParseError struct {
	Err error
	Msg string
}

func (p *ParseError) Error() string {
	if p.Err != nil {
		return fmt.Errorf("%s: %w", p.Msg, p.Err).Error()
	} else {
		return p.Msg
	}
}

func (p *ParseError) UnWrap() error {
	return p.Err
}

func NewParseError(err error, msg string) error {
	return &ParseError{
		Err: err,
		Msg: msg,
	}
}

type ValidationError struct {
	Err error
}

func (v *ValidationError) Error() string {
	return v.Err.Error()

}

func (v *ValidationError) UnWrap() error {
	return v.Err
}

func NewValidationError(err error) error {
	return &ValidationError{
		Err: err,
	}
}

var (
	// parse errors
	ErrInvalidRecordLength = fmt.Errorf("transaction record is not 240 characters")
	// Validation errors
	ErrMissingOriginatorId                   = fmt.Errorf("missing originator ID")
	ErrMissingCreationDate                   = fmt.Errorf("missing creation date")
	ErrInvalidRecordType                     = fmt.Errorf("invalid record type")
	ErrInvalidCurrencyCode                   = fmt.Errorf("invalid currency code")
	ErrInvalidOriginatorIdLength             = fmt.Errorf("invalid originator ID length")
	ErrInvalidOriginatorId                   = fmt.Errorf("invalid originator ID not alpha numeric")
	ErrInvalidFileCreationNum                = fmt.Errorf("invalid file creation number")
	ErrInvalidDestinationDataCenterNo        = fmt.Errorf("invalid destination data center")
	ErrInvalidDirectClearerCommunicationArea = fmt.Errorf("invalid direct clearer communication area")
	ErrScanParseError                        = fmt.Errorf("failed to parse txn")
)
