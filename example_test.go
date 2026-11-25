package cadeft_test

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/moov-io/cadeft"
)

func Example_writeFile() {
	date := time.Date(2026, 11, 25, 0, 0, 0, 0, time.UTC)
	header := cadeft.NewFileHeader("1234567890", 1, cadeft.Ptr(date), int64(610), "CAD")
	txn := cadeft.NewTransaction(
		cadeft.CreditRecord, // record type
		"450",               // transaction type
		420,                 // amount in cents
		cadeft.Ptr(date),    // date
		"123456789",         // institution ID
		"12345",             // account number
		"222222222222222",   // item trace number
		"payor name",        // originator short name
		"payee name",        // payor/payee name
		"payor long name",   // originator long name
		"987654321",         // return institution ID
		"54321",             // return account number
		"",                  // original item trace number (only used for returns)
		cadeft.WithCrossRefNo("0000100000024"),
	)

	file := cadeft.NewFile(header, cadeft.Transactions{txn})
	if err := file.Validate(); err != nil {
		log.Fatal(err)
	}

	serializedFile, err := file.Create()
	if err != nil {
		log.Fatal(err)
	}

	for _, line := range strings.Split(serializedFile, "\n") {
		if line == "" {
			continue
		}
		fmt.Println(line[:1], len(line))
	}
	// Output:
	// A 1464
	// C 1464
	// Z 1464
}

func Example_readFile() {
	f, err := os.Open("sample_files/CO14821.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	eftFile, err := cadeft.NewReader(f).ReadFile()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(len(eftFile.Txns))
	// Output:
	// 11
}

func Example_streamFile() {
	f, err := os.Open("sample_files/CO14821.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	fileStreamer := cadeft.NewFileStream(f)
	count := 0
	for {
		txn, err := fileStreamer.ScanTxn()
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatal(err)
		}
		if err := txn.Validate(); err != nil {
			log.Fatal(err)
		}
		count++
	}
	fmt.Println(count)
	// Output:
	// 11
}
