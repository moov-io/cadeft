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
	var mode string
	var fileName string
	flag.StringVar(&mode, "mode", "", "define the usage of the parser either build or parse")
	flag.StringVar(&fileName, "file", "", "eft file to parse")
	validate := flag.Bool("validate", false, "apply valdidation to file when building or parsing")
	flag.Parse()

	if mode != "parse" && mode != "build" {
		log.Fatal("invalid mode flag value, can only use parse or build")
	}

	if mode == "parse" {
		// open file pointed to by file
		var eftFile cadeft.File
		if fileName != "" {
			file, err := os.Open(fileName)
			if err != nil {
				log.Fatal(fmt.Errorf("failed to open file %s: %w", fileName, err))
			}
			reader := cadeft.NewReader(file)
			eftFile, err = reader.ReadFile()
			if err != nil {
				log.Fatal(fmt.Errorf("failed to parse file: %w", err))
			}
		} else {
			stdin, err := io.ReadAll(os.Stdin)
			if err != nil {
				log.Fatal(err)
			}

			rawFile := string(stdin)
			if len(stdin) == 0 {
				return
			}
			reader := cadeft.NewReader(strings.NewReader(rawFile))
			eftFile, err = reader.ReadFile()
			if err != nil {
				log.Fatal(err)
			}
		}
		output, err := json.Marshal(eftFile)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s", output)
	} else if mode == "build" {
		if fileName != "" {
			file, err := os.Open(fileName)
			if err != nil {
				log.Fatal(err)
			}
			var eftFile cadeft.File
			if err := json.NewDecoder(file).Decode(&eftFile); err != nil {
				log.Fatal(err)
			}
			if *validate {
				if err := eftFile.Validate(); err != nil {
					log.Fatal(err)
				}
			}
			s, err := eftFile.Create()
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("%s", s)
		}
	}
}
