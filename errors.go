package cadeft

import (
	"errors"
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
	ErrInvalidRecordLength = errors.New("transaction record is not 240 characters")
	// Validation errors
	ErrMissingOriginatorId                   = errors.New("missing originator ID")
	ErrMissingCreationDate                   = errors.New("missing creation date")
	ErrInvalidRecordType                     = errors.New("invalid record type")
	ErrInvalidCurrencyCode                   = errors.New("invalid currency code")
	ErrInvalidOriginatorIdLength             = errors.New("invalid originator ID length")
	ErrInvalidOriginatorId                   = errors.New("invalid originator ID not alpha numeric")
	ErrInvalidFileCreationNum                = errors.New("invalid file creation number")
	ErrInvalidDestinationDataCenterNo        = errors.New("invalid destination data center")
	ErrInvalidDirectClearerCommunicationArea = errors.New("invalid direct clearer communication area")
	ErrScanParseError                        = errors.New("failed to parse txn")
)
