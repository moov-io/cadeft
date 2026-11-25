# cadeft

[![Moov Banner Logo](https://user-images.githubusercontent.com/20115216/104214617-885b3c80-53ec-11eb-8ce0-9fc745fb5bfc.png)](https://github.com/moov-io)

<p align="center">
  <a href="https://slack.moov.io/">Community</a>
  ·
  <a href="https://moov.io/blog/">Blog</a>
  <br>
</p>

[![GoDoc](https://pkg.go.dev/badge/github.com/moov-io/cadeft.svg)](https://pkg.go.dev/github.com/moov-io/cadeft)
[![Build Status](https://github.com/moov-io/cadeft/workflows/Go/badge.svg)](https://github.com/moov-io/cadeft/actions)
[![Coverage Status](https://codecov.io/gh/moov-io/cadeft/branch/master/graph/badge.svg)](https://codecov.io/gh/moov-io/cadeft)
[![Go Report Card](https://goreportcard.com/badge/github.com/moov-io/cadeft)](https://goreportcard.com/report/github.com/moov-io/cadeft)
[![Repo Size](https://img.shields.io/github/languages/code-size/moov-io/cadeft?label=project%20size)](https://github.com/moov-io/cadeft)
[![Apache 2 License](https://img.shields.io/badge/license-Apache2-blue.svg)](https://github.com/moov-io/cadeft/blob/master/LICENSE)
[![GitHub Stars](https://img.shields.io/github/stars/moov-io/cadeft)](https://github.com/moov-io/cadeft)
[![Twitter](https://img.shields.io/twitter/follow/moov?style=social)](https://twitter.com/moov?lang=en)

Moov's mission is to give developers an easy way to create and integrate bank processing into their own software products. Our open source projects are each focused on solving a single responsibility in financial services and designed around performance, scalability, and ease of use.

## Description
`cadeft` is a Go library for parsing and generating Electronic Funds Transfer (EFT) files adhering to the [Payments Canada 005 specification](https://www.payments.ca/sites/default/files/standard005eng.pdf).

## Installation

This project uses [Go Modules](https://go.dev/blog/using-go-modules) and Go v1.26 or newer. See [Golang's install instructions](https://golang.org/doc/install) for help setting up Go.

```
go get github.com/moov-io/cadeft@latest
```

```go
import "github.com/moov-io/cadeft"
```

A complete write, read and stream walkthrough is in [example_test.go](example_test.go). The sample file at [sample_files/CO14821.txt](sample_files/CO14821.txt) is a useful input for the Reader and FileStreamer examples.

## File fields

These rules come from the library validation tags and constructors. The full field list is in [Payments Canada Standard 005](https://www.payments.ca/sites/default/files/standard005eng.pdf).

- **Originator ID:** exactly 10 characters (`len=10`).
- **File creation number:** required integer from 1 to 9999. It is written as 4 digits.
- **Destination data centre number:** integer from 0 to 99999. It is written as 5 digits in the header (`NewFileHeader`).
- **Currency:** `CAD` or `USD`.
- **Transaction type:** a 3-digit numeric code from Standard 005 (`TxnType`). The library accepts any 3-digit value and does not list the codes. The examples use `"450"`.
- **Amount:** an integer with no decimal point, written as 10 digits. Standard 005 expresses amounts in cents (420 is $4.20).
- **Institution ID:** up to 9 numeric digits. **Account number:** up to 12 characters. **Item trace number:** up to 22 numeric digits.

`File.Create()` writes a footer from the transactions when you do not set `File.Footer`. `Reader.ReadFile()` parses the header, transactions and footer, but it does not compare footer totals to the transactions. Call `File.Validate()` to check field rules and footer totals.

`NewTransaction` can also build return and reversal records (`I`, `J`, `E`, `F`) when you pass the matching `RecordType` and an original item trace number. Optional fields use `cadeft.WithCrossRefNo`, `WithUserID`, `WithSundryInfo`, `WithSettlementCode`, `WithStoredTransactionType` and `WithInvalidDataElementID`. The header accepts `cadeft.WithDirectClearerCommunicationArea`.

The JSON shape used by the CLI has three top-level keys: `file_header`, `transactions` and `file_footer`. Transaction objects include a `type` field (`C`, `D`, `I`, `J`, `E` or `F`) and the amount, names and account fields.

### Reading EFT files
There are 2 options to read files either `cadeft.Reader` or `cadeft.FileStreamer`
#### `cadeft.Reader`

Reader will attempt to read all transactions from an EFT file and return a populated `cadeft.File` struct or a collection of errors encountered during parsing.
```go
file, err := os.Open("sample_files/CO14821.txt")
if err != nil {
  return err
}
defer file.Close()

// Create a new reader passing an io.Reader
reader := cadeft.NewReader(file)

// attempt to read the file
eftFile, err := reader.ReadFile()
if err != nil {
  return err
}

// print out all transfers or handle the file
for _, txn := range eftFile.Txns {
  fmt.Printf("%+v", txn)
}
```

#### `cadeft.FileStreamer`
`cadeft.FileStreamer` will read one transaction from a file at a time or return an error. Consecutive calls to `ScanTxn()` will read the next transaction or return an error. `FileStreamer` will keep state of the parser's position and return new transactions every call. This allows the caller to either ignore errors that have surfaced when parsing/validating a transaction and construct their own array of `cadeft.Transaction` structs. You can also call `Validate()` on a `cadeft.Transaction` struct which will validate all fields against the Payments Canada 005 Spec.

`NewFileStream` takes an `io.ReadSeeker` (for example `*os.File` or `*strings.Reader`).
```go
file, err := os.Open("sample_files/CO14821.txt")
if err != nil {
  return err
}
defer file.Close()

// instantiate a new FileStreamer
fileStreamer := cadeft.NewFileStream(file)

// start reading the file
for {
  // every iteration a new transaction is returned or an error
  txn, err := fileStreamer.ScanTxn()
  if err != nil {
    // an io.EOF is returned when parsing is complete
    if err == io.EOF {
      break
    }

    // handle the parse error as you want
    var parseErr *cadeft.ParseError
    if ok := errors.As(err, &parseErr); ok {
      log.Printf("encountered parse error when processing incoming file: %v", parseErr)
      continue
    } else {
      log.Printf("fatal error when streaming transaction from file: %v", err)
      return err
    }
  }

  // validate that the parsed txn is valid
  err = txn.Validate()
  if err != nil {
    return err
  }
}
```

NOTE: Because `ScanTxn` keeps track of the parser's state it is not concurrency-safe. If you want to incorporate some level of concurrency, make sure the call to `ScanTxn()` is outside of a go routine like so:
```go
file, err := os.Open("sample_files/CO14821.txt")
if err != nil {
  return err
}
defer file.Close()

fileStreamer := cadeft.NewFileStream(file)

wg := sync.WaitGroup{}
batchSize := 10
guard := make(chan struct{}, batchSize)

for {
  txn, err := fileStreamer.ScanTxn()
  if err != nil {
    if err == io.EOF {
      break
    }
    // handle the error any which way you want
    continue
  }

  // read transactions into memory and concurrently handle each txn
  guard <- struct{}{}
  wg.Add(1)
  go func(txn cadeft.Transaction) {
    defer func() {
      wg.Done()
      <-guard
    }()
    // do your thing...
    _ = txn
  }(txn)
}
wg.Wait()
```

## Write EFT Files
To write an EFT file construct a `cadeft.File` struct with the appropriate `Header`, `Footer` and `[]Transactions`. You can confirm that your file is valid by calling `File.Validate()` or `Validate()` on the `Header`, `Footer` and each individual `Transaction`.
```go
// create the file header (originator ID must be exactly 10 characters)
header := cadeft.NewFileHeader("1234567890", 1, cadeft.Ptr(time.Now()), int64(610), "CAD")

// create a transaction, in this case it's a credit transaction
txn := cadeft.NewTransaction(
  cadeft.CreditRecord, // record type
  "450", // transaction type
  420, // amount in cents
  cadeft.Ptr(time.Now()), // date
  "123456789", // institution ID
  "12345", // account number
  "222222222222222", // item trace number
  "payor name", // originator short name
  "payee name", // payor/payee name
  "payor long name", // originator long name
  "987654321", // return institution ID
  "54321", // return account number
  "", // original item trace number (only used for returns)
  cadeft.WithCrossRefNo("0000100000024"), // cross ref number (optional)
)

// create a new cadeft.File instance with header and txn
file := cadeft.NewFile(header, cadeft.Transactions{txn})
if err := file.Validate(); err != nil {
  // file does not adhere to the 005 spec
  return err
}

// serialize the file into payments canada 005 spec
serializedFile, err := file.Create()
if err != nil {
  // error when writing the file
  return err
}

// write the serialized string to a file or print it out
fmt.Printf("%s", serializedFile)
```

## Command line

The `cmd` package is a small tool that converts an EFT file to JSON and JSON back to an EFT file. From a clone of this repository:

```
go build -o cadeft ./cmd
```

`go install github.com/moov-io/cadeft/cmd@latest` also works. The installed binary is named `cmd` (the directory name).

```
# parse an EFT file and print JSON
./cadeft -mode parse -file sample_files/CO14821.txt

# parse and validate header, transaction and footer fields (including footer totals)
./cadeft -mode parse -validate -file sample_files/CO14821.txt

# build an EFT file from JSON and write it to standard output
./cadeft -mode build -validate -file file.json
```

`-mode` must be `parse` or `build`. In parse mode, omit `-file` to read the EFT data from standard input. In build mode, `-file` is required and must point to a JSON file with `file_header`, `transactions` and optional `file_footer` keys. `-validate` runs `File.Validate()` in both parse and build modes.

## Project status

cadeft is currently being developed for use in multiple production environments. The library was developed in large part by [Synctera](https://synctera.com/). Please star the project if you are interested in its progress. Let us know if you encounter any bugs/unclear documentation or have feature suggestions by opening up an issue or pull request. Thanks!

## Getting help

| channel                                                    | info                                                                                                                                    |
|------------------------------------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------|
| [Project Documentation](https://pkg.go.dev/github.com/moov-io/cadeft) | Package documentation is on pkg.go.dev.                                                                                          |
| Twitter [@moov](https://twitter.com/moov)                  | You can follow Moov.io's Twitter feed to get updates on our project(s). You can also tweet us questions or just share blogs or stories. |
| [GitHub Issue](https://github.com/moov-io/cadeft/issues)   | If you are able to reproduce a problem please open a GitHub Issue under the specific project that caused the error.                     |
| [moov-io slack](https://slack.moov.io/)                    | Join our slack channel (`#cadeft`) to have an interactive discussion about the development of the project.                              |

## Supported and tested platforms

- 64-bit Linux (Ubuntu, Debian), macOS, and Windows

## Contributing

Yes please! Please review our [Contributing guide](https://github.com/moov-io/ach/blob/master/CONTRIBUTING.md) and [Code of Conduct](https://github.com/moov-io/ach/blob/master/CODE_OF_CONDUCT.md) to get started!

This project uses [Go Modules](https://go.dev/blog/using-go-modules) and Go v1.26 or newer. See [Golang's install instructions](https://golang.org/doc/install) for help setting up Go. You can download the source code and we offer [tagged and released versions](https://github.com/moov-io/cadeft/releases/latest) as well. We highly recommend you use a tagged release for production.

### Releasing

To make a release of cadeft simply open a pull request with `CHANGELOG.md` updated with the next version number and details. You'll also need to push the tag (i.e. `git push origin v1.0.0`) to origin in order for CI to make the release. See [GitHub Releases](https://github.com/moov-io/cadeft/releases) for versions after v0.1.0.

### Testing

We maintain a comprehensive suite of unit tests and recommend table-driven testing when a particular function warrants several very similar test cases. Run all tests with `go test ./...`. Current overall coverage can be found on [Codecov](https://app.codecov.io/gh/moov-io/cadeft/).

## License

Apache License 2.0 - See [LICENSE](LICENSE) for details.
