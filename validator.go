package cadeft

import (
	"fmt"
	"regexp"

	"github.com/go-playground/validator/v10"
)

const (
	eftAlphaKey    = "eft_alpha"
	eftNumericKey  = "eft_num"
	eftRecTypeKey  = "rec_type"
	eftCurrencyKey = "eft_cur"
)

var alphaRegex = regexp.MustCompile(`^[\w\-\s]+$`)
var numericRegex = regexp.MustCompile(`^[\d]+$`)

var eftValidator = validator.New(validator.WithRequiredStructEnabled())

// EftAlphaRegex checks if the value is a valid string that matches an EFT alphanumeric field
func eftAlphaRegex(fl validator.FieldLevel) bool {
	if fl.Field().IsZero() {
		return true
	}

	return alphaRegex.MatchString(fl.Field().String())
}

func eftNumericRegex(fl validator.FieldLevel) bool {
	if fl.Field().IsZero() {
		return true
	}
	return numericRegex.MatchString(fl.Field().String())
}

func eftRecType(fl validator.FieldLevel) bool {
	switch fl.Field().String() {
	case string(HeaderRecord), string(CreditRecord), string(DebitRecord), string(CreditReverseRecord), string(DebitReverseRecord), string(ReturnCreditRecord), string(ReturnDebitRecord), string(NoticeOfChangeRecord), string(NoticeOfChangeHeader), string(NoticeOfChangeFooter):
		return true
	default:
		return false
	}
}

func eftCurrency(fl validator.FieldLevel) bool {
	switch fl.Field().String() {
	case "CAD", "USD":
		return true
	default:
		return false
	}
}

func c(e error) {
	if e != nil {
		panic(fmt.Sprintf("failed to register validators: %s", e)) //nolint:forbidigo
	}
}

func init() {
	c(eftValidator.RegisterValidation(eftAlphaKey, eftAlphaRegex))
	c(eftValidator.RegisterValidation(eftNumericKey, eftNumericRegex))
	c(eftValidator.RegisterValidation(eftRecTypeKey, eftRecType))
	c(eftValidator.RegisterValidation(eftCurrencyKey, eftCurrency))
}
