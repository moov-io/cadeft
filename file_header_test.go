package cadeft

import (
	"strings"
	"testing"
	"time"

	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
)

func TestParseARecord(t *testing.T) {
	type testCase struct {
		input         string
		expected      FileHeader
		expectReadErr bool
		validationErr bool
	}

	cases := map[string]testCase{
		"Valid header record": {
			input: "A0000000010000000610000102313861210hello               CAD",
			expected: FileHeader{
				RecordHeader: RecordHeader{
					RecordType:      HeaderRecord,
					recordCount:     1,
					OriginatorID:    "0000000610",
					FileCreationNum: 1,
				},
				CreationDate:                   lo.ToPtr(time.Date(2023, time.May, 18, 0, 0, 0, 0, time.UTC)),
				DestinationDataCenterNo:        61210,
				DirectClearerCommunicationArea: "hello",
				CurrencyCode:                   "CAD",
			},
		},
		"Invalid header record length": {
			input:         "C123",
			expectReadErr: true,
		},
		"Invalid header record type": {
			input:         "C0000000010000000610000002313861210                    CAD",
			validationErr: true,
		},
		"Missing originator id": {
			input:         "A000000001          000002313861210                    CAD",
			validationErr: true,
		},
		"Missing creation date": {
			input:         "A00000000100000006100001      61210hello               CAD",
			validationErr: true,
		},
		"Invalid file creation num": {
			input:         "A0000000010000000610000002313861210                    CAD",
			validationErr: true,
		},
		"Invalid currency code": {
			input:         "A00000000100000006100001023138612210                    AAA",
			validationErr: true,
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			r := require.New(t)
			header := FileHeader{}
			err := header.parse(tc.input)
			if tc.expectReadErr {
				r.Error(err)
				return
			}
			err = header.Validate()
			if tc.validationErr {
				r.Error(err)
			} else {
				r.NoError(err)
				r.Equal(tc.expected.RecordType, header.RecordType)
				r.Equal(tc.expected.recordCount, header.recordCount)
				r.Equal(tc.expected.OriginatorID, header.OriginatorID)
				r.Equal(tc.expected.FileCreationNum, header.FileCreationNum)
				r.Equal(tc.expected.CreationDate, header.CreationDate)
				r.Equal(tc.expected.DestinationDataCenterNo, header.DestinationDataCenterNo)
				r.Equal(tc.expected.DirectClearerCommunicationArea, header.DirectClearerCommunicationArea)
				r.Equal(tc.expected.CurrencyCode, header.CurrencyCode)
			}
		})
	}
}

func TestCreateHeaderRecord(t *testing.T) {
	type testCase struct {
		header                   *FileHeader
		noError                  bool
		expectedSerializedOutput string
		debug                    bool
	}
	date := time.Date(2023, 8, 29, 0, 0, 0, 0, time.UTC)
	cases := map[string]testCase{
		"valid header": {
			header:                   NewFileHeader("0000000610", 1, &date, 61210, "CAD"),
			expectedSerializedOutput: "A0000000010000000610000102324161210                    CAD" + strings.Repeat(" ", 1406),
			noError:                  true,
		},
		"missing originator id": {
			header: NewFileHeader("", 1, &date, 1, "CAD", WithDirectClearerCommunicationArea("Samplecomm")),
			debug:  true,
		},
		"invalid file creation num": {
			header: NewFileHeader("0000000001", 0, &date, 1, "CAD", WithDirectClearerCommunicationArea("Samplecomm")),
		},
		"missing file creation date": {
			header: NewFileHeader("0000000001", 1, nil, 1, "CAD", WithDirectClearerCommunicationArea("Samplecomm")),
		},
		"invalid currency code": {
			header: NewFileHeader("0000000001", 1, &date, 1, "ABC", WithDirectClearerCommunicationArea("Samplecomm")),
		},
		"originator id not correct length": {
			header: NewFileHeader("000", 1, &date, 1, "CAD", WithDirectClearerCommunicationArea("Samplecomm")),
		},
		"originator id not alphanumeric": {
			header: NewFileHeader("000000000!", 1, &date, 1, "CAD", WithDirectClearerCommunicationArea("Samplecomm")),
		},
		"direct clearer communcation area too long": {
			header: NewFileHeader("0000000001", 1, &date, 0, "CAD", WithDirectClearerCommunicationArea("aaaaaaaaaaaaaaaaaaaaaaaaaaa")),
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			r := require.New(t)
			err := tc.header.Validate()
			if tc.noError {
				r.NoError(err)
				output, buildErr := tc.header.buildHeader(0)
				r.NoError(buildErr)
				r.Equal(tc.expectedSerializedOutput, output)
			} else {
				r.Error(err)
			}
		})
	}

}
