package cadeft

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

func parseNum(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

func convertRecordType(recType string) (RecordType, error) {
	switch recType {
	case "A":
		return HeaderRecord, nil
	case "C":
		return CreditRecord, nil
	case "D":
		return DebitRecord, nil
	case "Z":
		return FooterRecord, nil
	default:
		return "", errors.New(fmt.Sprintf("unrecognized record type %s", recType))
	}
}

func parseDate(date string) (time.Time, error) {
	if len(date) != 6 {
		return time.Time{}, errors.New("date string is not valid length")
	}
	year, err := strconv.Atoi(fmt.Sprintf("20%s", date[1:3]))
	if err != nil {
		return time.Time{}, errors.Wrap(err, "failed to convert year")
	}
	daysSinceJan1, err := strconv.Atoi(date[3:])
	if err != nil {
		return time.Time{}, errors.Wrap(err, "failed to convert days since jan 1st")
	}
	return time.Date(year, time.January, 0, 0, 0, 0, 0, time.UTC).AddDate(0, 0, daysSinceJan1), nil

}

// convert an interger to a numeric string with 0's padded to the left. eg 12 would be "00012" if required length is 5
func convertNumToZeroPaddedString(in int64, requiredLength int) string {
	return fmt.Sprintf("%0*d", requiredLength, in)
}

// default length of date string in eft format is 6 0yyddd
// this function assumes that the year will always be in the 20's (2010, 2020, 2050 etc.)
func convertTimestampToEftDate(in time.Time) string {
	year := in.Year() % 100
	day := in.YearDay()

	return fmt.Sprintf("0%d%d", year, day)
}

func isFillerString(s string) bool {
	for _, c := range s {
		if c != ' ' {
			return false
		}
	}
	return true
}

func createFillerString(length int) string {
	return strings.Repeat(" ", length)
}

// pads input string with blanks until string is of requiredLength length.
func padStringWithBlanks(s string, requiredLength int) string {
	if len(s) >= requiredLength {
		return s
	}
	blankStr := fmt.Sprintf("%-*s", requiredLength-len(s), " ")
	return fmt.Sprintf("%s%s", s, blankStr)
}

func abreviateStringToLength(s string, reqLength int) string {
	if len(s) > reqLength {
		return s[:reqLength]
	}
	// fill with blanks
	return padStringWithBlanks(s, reqLength)
}

func padNumericStringWithZeros(s string, reqLength int) string {
	return fmt.Sprintf("%0*s", reqLength, s)
}

func isTxnRecord(t string) bool {
	switch t {
	case "D", "C", "E", "F", "I", "J":
		return true
	default:
		return false
	}
}

func formatName(s string, reqLen int) (string, error) {
	normal, err := normalize(s)
	if err != nil {
		return "", errors.Wrap(err, "failed to normalize string")
	}
	formatted := abreviateStringToLength(normal, reqLen)
	return formatted, nil

}
