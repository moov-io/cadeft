package cadeft_test

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/moov-io/cadeft"
)

func FuzzReader(f *testing.F) {
	populateCorpus(f)

	f.Fuzz(func(t *testing.T, contents string) {
		if len(contents) > 1<<20 {
			t.Skip()
		}

		r := cadeft.NewReader(strings.NewReader(contents))
		_, _ = r.ReadFile()

		// Streamer path — more tolerant of errors; needs ReadSeeker
		stream := cadeft.NewFileStream(strings.NewReader(contents))
		_, _ = stream.GetHeader()
		for i := 0; i < 100; i++ {
			txn, err := stream.ScanTxn()
			if err != nil {
				if err == io.EOF {
					break
				}
				// continue scanning after parse errors if possible
				if txn == nil {
					break
				}
			}
			if txn == nil {
				break
			}
		}
		_, _ = stream.GetFooter()
	})
}

func populateCorpus(f *testing.F) {
	f.Helper()

	f.Add("")
	f.Add("A0000000010000000610000102313861210                    CAD\n")

	_ = filepath.Walk("sample_files", func(path string, info fs.FileInfo, err error) error {
		if err != nil || info == nil || info.IsDir() {
			return nil
		}
		bs, err := os.ReadFile(path)
		if err != nil || len(bs) > 512*1024 {
			return nil
		}
		f.Add(string(bs))
		return nil
	})
}
