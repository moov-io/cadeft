# cadeft

[![Moov Banner Logo](https://user-images.githubusercontent.com/20115216/104214617-885b3c80-53ec-11eb-8ce0-9fc745fb5bfc.png)](https://github.com/moov-io)

<p align="center">
  <a href="https://slack.moov.io/">Community</a>
  ·
  <a href="https://moov.io/blog/">Blog</a>
  <br>
</p>

[![GoDoc](https://godoc.org/github.com/moov-io/cadeft?status.svg)](https://godoc.org/github.com/moov-io/cadeft)
[![Build Status](https://github.com/moov-io/cadeft/workflows/Go/badge.svg)](https://github.com/moov-io/cadeft/actions)
[![Coverage Status](https://codecov.io/gh/moov-io/cadeft/branch/master/graph/badge.svg)](https://codecov.io/gh/moov-io/cadeft)
[![Go Report Card](https://goreportcard.com/badge/github.com/moov-io/cadeft)](https://goreportcard.com/report/github.com/moov-io/cadeft)
[![Repo Size](https://img.shields.io/github/languages/code-size/moov-io/cadeft?label=project%20size)](https://github.com/moov-io/cadeft)
[![Apache 2 License](https://img.shields.io/badge/license-Apache2-blue.svg)](https://raw.githubusercontent.com/moov-io/ach/master/LICENSE)
[![Slack Channel](https://slack.moov.io/badge.svg?bg=e01563&fgColor=fffff)](https://slack.moov.io/)
[![GitHub Stars](https://img.shields.io/github/stars/moov-io/cadeft)](https://github.com/moov-io/cadeft)
[![Twitter](https://img.shields.io/twitter/follow/moov?style=social)](https://twitter.com/moov?lang=en)

Moov's mission is to give developers an easy way to create and integrate bank processing into their own software products. Our open source projects are each focused on solving a single responsibility in financial services and designed around performance, scalability, and ease of use.

## Description
`cadeft` is a parser library designed for parsing and generating Electronic Funds Transfer (EFT) files adhereing to the [Payments Canada 005 specification](https://www.payments.ca/sites/default/files/standard005eng.pdf). 

### Reading EFT files
There are 2 options to read files either `cadeft.Reader` or `cadeft.FileStreamer`
#### `cadeft.Reader`

Reader will attempt to read all transactions from an EFT file and return a populated `cadeft.File` struct or a collection of errors encountered during parsing.
```
file, err := os.Open("./eft_file.txt")
if err != nil {
  return err
}

// Create a new reader passing an io.Reader
reader := cadeft.NewReader(file)

// attempt to read the file
eftFile, err := reader.ReadFile()
if err != nil {
  return err
}

// print out all transers or handle the file
for _, txn := range eftFile.Txns {
  fmt.Printf("%+v", txn)
}
```

#### `cadeft.FileStreamer`
`cadeft.FileStreamer` will read one transaction from a file at a time or return an error. Consecutive calls to `ScanTxn()` will read the next transaction or return an error. `FileStreamer` will keep state of the parser's position and return new transactions every call. This allows the caller to either ignore errors that have surfaced when parsing/validating a transaction and construct their own array of `cadeft.Transaction` structs. You can also call `Validate()` on a `cadeft.Transaction` struct which will validate all fields against the Payments Canada 005 Spec. 
```
file, err := os.Open("./eft_file.txt")
if err != nil {
  return err
}

// instantiate a new FileStreamer
stramer := cadeft.NewFileStreamer(file)

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
				log.Err(parseErr).Msg("encountered parse error when processing incoming file")
				continue
			} else {
				log.Err(err).Msg("fatal error when streaming transaction from file")
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

NOTE: Because `ScanTxn` keeps track of the parser's state it is not concurrency-safe if you want to incorporate some level of concurrency make sure the call to `ScanTxn()` is outside of a go routine like so:
```
...
fileStreamer := cadeft.NewFileStreamer(file)

wg := sync.WaitGroup{}
batchSize := 10
guard := make(chan struct{}, batchSize)

for {
  txn, err := fileStreamer.ScanTxn()
  if err != nil {
    // handle the error any which way you want
  }

  // read transactions into memory and concurrently handle each txn
  guard <- struct{}{}
  wg.Add(1)
  go func() {
    defer func() {
      wg.Done()
      <-guard
    }()
    // do your thing...
  }
  wg.Wait()
}
```

## Write EFT Files
To write an EFT file construct a `cadeft.File` struct with the appropriate `Header`, `Footer` and `[]Transactions`. You can confirm that your file is valid by calling `File.Validate()` or `Validate()` on the `Header`, `Footer` and each individual `Transaction`.
```
// create the file header
header := cadeft.NewFileHeader("123456789", 1, time.Now(), int64(610), "CAD")

// create a transaction, in this case it's a credit transaction
txn := cadeft.NewTransaction(
  cadeft.CreditRecord, // recod type
  "450", // transaction code
  420, // amount
  Ptr(time.Now()), // date
  "123456789", // institution ID
  "12345", // account number
  "222222222222222", // item trace nnumber
  "payor name", // payor/payee name
  "payor long name", // payor/payee long name
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


## Project status

cadeft is currently being developed for use in multiple production environments. The library was developed in large part by [Synctera](https://synctera.com/). Please star the project if you are interested in its progress. Let us know if you encounter any bugs/unclear documentation or have feature suggestions by opening up an issue or pull request. Thanks!

## Getting help

| channel                                                    | info                                                                                                                                    |
|------------------------------------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------|
| [Project Documentation](https://moov-io.github.io/cadeft/) | Our project documentation available online.                                                                                             |
| Twitter [@moov](https://twitter.com/moov)                  | You can follow Moov.io's Twitter feed to get updates on our project(s). You can also tweet us questions or just share blogs or stories. |
| [GitHub Issue](https://github.com/moov-io/cadeft/issues)   | If you are able to reproduce a problem please open a GitHub Issue under the specific project that caused the error.                     |
| [moov-io slack](https://slack.moov.io/)                    | Join our slack channel (`#cadeft`) to have an interactive discussion about the development of the project.                              |

## Supported and tested platforms

- 64-bit Linux (Ubuntu, Debian), macOS, and Windows

## Contributing

Yes please! Please review our [Contributing guide](CONTRIBUTING.md) and [Code of Conduct](https://github.com/moov-io/ach/blob/master/CODE_OF_CONDUCT.md) to get started!

This project uses [Go Modules](https://go.dev/blog/using-go-modules) and Go v1.20 or newer. See [Golang's install instructions](https://golang.org/doc/install) for help setting up Go. You can download the source code and we offer [tagged and released versions](https://github.com/moov-io/cadeft/releases/latest) as well. We highly recommend you use a tagged release for production.

### Releasing

To make a release of cadeft simply open a pull request with `CHANGELOG.md` and `version.go` updated with the next version number and details. You'll also need to push the tag (i.e. `git push origin v1.0.0`) to origin in order for CI to make the release.

### Testing

We maintain a comprehensive suite of unit tests and recommend table-driven testing when a particular function warrants several very similar test cases. After starting the services with Docker Compose run all tests with `go test ./...`. Current overall coverage can be found on [Codecov](https://app.codecov.io/gh/moov-io/cadeft/).

## License

Apache License 2.0 - See [LICENSE](LICENSE) for details.
