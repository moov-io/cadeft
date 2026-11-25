package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/moov-io/cadeft"
)

func main() {
	if err := run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr); err != nil {
		if err == flag.ErrHelp {
			os.Exit(0)
		}
		log.Fatal(err)
	}
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("cadeft", flag.ContinueOnError)
	fs.SetOutput(stderr)

	var mode string
	var fileName string
	fs.StringVar(&mode, "mode", "", "define the usage of the parser either build or parse")
	fs.StringVar(&fileName, "file", "", "EFT file to parse or JSON file to build")
	validate := fs.Bool("validate", false, "apply validation to the file when building or parsing")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if mode != "parse" && mode != "build" {
		return fmt.Errorf("invalid mode flag value, can only use parse or build")
	}

	if mode == "parse" {
		eftFile, err := parseFile(fileName, stdin)
		if err != nil {
			return err
		}
		if *validate {
			if err := eftFile.Validate(); err != nil {
				return err
			}
		}
		output, err := json.Marshal(eftFile)
		if err != nil {
			return err
		}
		fmt.Fprintf(stdout, "%s", output)
		return nil
	}

	if fileName == "" {
		return fmt.Errorf("-file is required when -mode is build")
	}
	file, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	var eftFile cadeft.File
	if err := json.NewDecoder(file).Decode(&eftFile); err != nil {
		return err
	}
	if *validate {
		if err := eftFile.Validate(); err != nil {
			return err
		}
	}
	s, err := eftFile.Create()
	if err != nil {
		return err
	}
	fmt.Fprintf(stdout, "%s", s)
	return nil
}

func parseFile(fileName string, stdin io.Reader) (cadeft.File, error) {
	if fileName != "" {
		file, err := os.Open(fileName)
		if err != nil {
			return cadeft.File{}, fmt.Errorf("failed to open file %s: %w", fileName, err)
		}
		defer file.Close()
		eftFile, err := cadeft.NewReader(file).ReadFile()
		if err != nil {
			return cadeft.File{}, fmt.Errorf("failed to parse file: %w", err)
		}
		return eftFile, nil
	}

	raw, err := io.ReadAll(stdin)
	if err != nil {
		return cadeft.File{}, err
	}
	if len(raw) == 0 {
		return cadeft.File{}, nil
	}
	eftFile, err := cadeft.NewReader(strings.NewReader(string(raw))).ReadFile()
	if err != nil {
		return cadeft.File{}, err
	}
	return eftFile, nil
}
