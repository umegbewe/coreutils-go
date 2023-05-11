package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
)

func main() {
	decodeFlag := flag.Bool("d", false, "decode data")
	wrapFlag := flag.Int("w", 76, "wrap encoded lines after COLS character (default 76, 0 to disable wrapping)")
	inputFlag := flag.String("i", "", "input file (default: stdin)")
	outputFlag := flag.String("o", "", "output file (default: stdout)")
	flag.Parse()

	var inputData []byte
	var err error

	if *inputFlag != "" {
		inputData, err = ioutil.ReadFile(*inputFlag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading input file: %v\n", err)
			os.Exit(1)
		}
	} else {
		inputData, err = ioutil.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading from stdin: %v\n", err)
			os.Exit(1)
		}
	}

	var outputData []byte

	if *decodeFlag {
		outputData, err = base64.StdEncoding.DecodeString(string(inputData))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error decoding base64 data: %v\n", err)
			os.Exit(1)
		}
	} else {
		encoded := base64.StdEncoding.EncodeToString(inputData)
		if *wrapFlag > 0 {
			outputData = wrap(encoded, *wrapFlag)
		} else {
			outputData = []byte(encoded)
		}
	}

	if *outputFlag != "" {
		err = ioutil.WriteFile(*outputFlag, outputData, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing output file: %v\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Print(string(outputData))
	}
}

func wrap(text string, lineWidth int) []byte {
	var wrapped []byte

	lineStart := 0
	for i := lineWidth; i < len(text); i += lineWidth {
		wrapped = append(wrapped, text[lineStart:i]...)
		wrapped = append(wrapped, '\n')
		lineStart = i
	}

	wrapped = append(wrapped, text[lineStart:]...)
	wrapped = append(wrapped, '\n')

	return wrapped
}
