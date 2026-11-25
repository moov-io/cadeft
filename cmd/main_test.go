package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/moov-io/cadeft"
	"github.com/stretchr/testify/require"
)

func testCreditFile(t *testing.T, amount int64) cadeft.File {
	t.Helper()
	date := time.Date(2026, 11, 25, 0, 0, 0, 0, time.UTC)
	header := cadeft.NewFileHeader("1234567890", 1, cadeft.Ptr(date), 610, "CAD")
	txn := cadeft.NewTransaction(
		cadeft.CreditRecord,
		"450",
		amount,
		cadeft.Ptr(date),
		"123456789",
		"12345",
		"222222222222222",
		"PAYROLL CO",
		"JANE EMPLOYEE",
		"PAYROLL COMPANY INC",
		"987654321",
		"54321",
		"",
	)
	file := cadeft.NewFile(header, cadeft.Transactions{txn})
	require.NoError(t, file.Validate())
	return file
}

func writeTemp(t *testing.T, name, contents string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	require.NoError(t, os.WriteFile(path, []byte(contents), 0o600))
	return path
}

func TestRunInvalidMode(t *testing.T) {
	err := run([]string{"-mode", "convert"}, strings.NewReader(""), &bytes.Buffer{}, &bytes.Buffer{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid mode")
}

func TestRunBuildRequiresFile(t *testing.T) {
	err := run([]string{"-mode", "build"}, strings.NewReader(""), &bytes.Buffer{}, &bytes.Buffer{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "-file is required")
}

func TestRunParseValidate(t *testing.T) {
	file := testCreditFile(t, 420)
	raw, err := file.Create()
	require.NoError(t, err)
	path := writeTemp(t, "valid.eft", raw)

	var out bytes.Buffer
	require.NoError(t, run([]string{"-mode", "parse", "-validate", "-file", path}, strings.NewReader(""), &out, &bytes.Buffer{}))

	var parsed cadeft.File
	require.NoError(t, json.Unmarshal(out.Bytes(), &parsed))
	require.Len(t, parsed.Txns, 1)
	require.Equal(t, int64(420), parsed.Txns[0].GetAmount())
}

func TestRunParseValidateRejectsBadFooter(t *testing.T) {
	file := testCreditFile(t, 420)
	file.Footer = &cadeft.FileFooter{
		RecordHeader: cadeft.RecordHeader{
			RecordType:      cadeft.FooterRecord,
			OriginatorID:    "1234567890",
			FileCreationNum: 1,
		},
		TotalValueOfCredit: 1,
		TotalCountOfCredit: 1,
	}
	raw, err := file.Create()
	require.NoError(t, err)
	path := writeTemp(t, "bad-footer.eft", raw)

	var out bytes.Buffer
	err = run([]string{"-mode", "parse", "-file", path}, strings.NewReader(""), &out, &bytes.Buffer{})
	require.NoError(t, err, "parse without -validate must accept a footer that does not match the transactions")

	out.Reset()
	err = run([]string{"-mode", "parse", "-validate", "-file", path}, strings.NewReader(""), &out, &bytes.Buffer{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "footer total value of credit")
}

func TestRunBuildValidate(t *testing.T) {
	file := testCreditFile(t, 420)
	jsonPath := writeTemp(t, "file.json", mustJSON(t, file))

	var out bytes.Buffer
	require.NoError(t, run([]string{"-mode", "build", "-validate", "-file", jsonPath}, strings.NewReader(""), &out, &bytes.Buffer{}))
	require.Contains(t, out.String(), "A")
	require.Contains(t, out.String(), "C")
	require.Contains(t, out.String(), "Z")
}

func TestValidateFlagHelp(t *testing.T) {
	var stderr bytes.Buffer
	err := run([]string{"-h"}, strings.NewReader(""), &bytes.Buffer{}, &stderr)
	require.ErrorIs(t, err, flag.ErrHelp)
	require.Contains(t, stderr.String(), "apply validation to the file when building or parsing")
	require.NotContains(t, stderr.String(), "valdidation")
}

func mustJSON(t *testing.T, v any) string {
	t.Helper()
	b, err := json.Marshal(v)
	require.NoError(t, err)
	return string(b)
}
